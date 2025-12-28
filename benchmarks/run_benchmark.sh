#!/bin/bash

# Professional benchmark orchestration script using k6

set -e

BLUE='\033[0;34m'
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
RESULTS_DIR="$PROJECT_ROOT/benchmarks/results"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

mkdir -p "$RESULTS_DIR"

# Check if k6 is installed
if ! command -v k6 &> /dev/null; then
    echo -e "${RED}❌ k6 is not installed!${NC}"
    echo "Install it:"
    echo "  macOS: brew install k6"
    echo "  Linux: https://k6.io/docs/getting-started/installation/"
    exit 1
fi

NUM_DECISIONS=${1:-50}

echo -e "${BLUE}╔════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║         Professional Load Testing with k6                      ║${NC}"
echo -e "${BLUE}║         Decisions: ${NUM_DECISIONS}, Sequential Testing        ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${YELLOW}💡 Sequential testing: Systems tested one at a time${NC}"
echo -e "${YELLOW}   Phase 1: Camunda 7 (start → test → stop)${NC}"
echo -e "${YELLOW}   Phase 2: DMN Go (start → test → stop)${NC}"
echo -e "${YELLOW}   No resource contention, fair comparison${NC}"
echo ""
sleep 2

cd "$PROJECT_ROOT"

# Clean up any existing containers
echo -e "${YELLOW}Cleaning up any existing containers...${NC}"
cd deployments/camunda
docker-compose down -v --remove-orphans 2>/dev/null || true
cd ../..
docker-compose down -v --remove-orphans 2>/dev/null || true
echo -e "${GREEN}✅ Clean state${NC}"
echo ""

# Generate DMN tables
echo -e "${YELLOW}Generating ${NUM_DECISIONS} DMN decision tables...${NC}"
./scripts/generate_camunda_dmn.sh $NUM_DECISIONS
echo ""

# Phase 1: Camunda 7
echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}  Phase 1: Load Testing Camunda 7                               ${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}\n"

echo "Starting Camunda..."
cd deployments/camunda
docker-compose up -d
cd ../..

echo "Waiting for Camunda startup..."
for i in {1..180}; do
    if curl -s "http://localhost:8080/engine-rest/engine" > /dev/null 2>&1; then
        echo "Camunda ready! (after ${i}s)"
        break
    fi
    [ $((i % 10)) -eq 0 ] && echo "Still waiting... (${i}s)"
    sleep 1
done
sleep 5

# Deploy DMN files
echo "Deploying DMN definitions to Camunda..."
DEPLOYED=0
for file in testdata/dmn/stress/*.dmn; do
    [ -f "$file" ] || continue
    curl -s -X POST "http://localhost:8080/engine-rest/deployment/create" \
        -F "deployment-name=$(basename $file)" \
        -F "data=@$file" > /dev/null 2>&1 && DEPLOYED=$((DEPLOYED + 1))
done
echo "Deployed: $DEPLOYED decisions"
echo ""

# Memory before test
CAMUNDA_MEMORY_BEFORE=$(docker stats camunda-7 --no-stream --format "{{.MemUsage}}" 2>/dev/null | awk '{print $1}')
echo "Memory before test: $CAMUNDA_MEMORY_BEFORE"

# Run k6 load test
echo -e "${GREEN}Running k6 load test...${NC}"
k6 run \
    --out json="$RESULTS_DIR/camunda_${TIMESTAMP}.json" \
    --summary-export="$RESULTS_DIR/camunda_summary_${TIMESTAMP}.json" \
    -e CAMUNDA_URL=http://localhost:8080 \
    benchmarks/k6/camunda_load_test.js || {
        echo -e "${YELLOW}⚠️  k6 test completed with warnings (threshold violations are OK)${NC}"
    }

# Memory after test
CAMUNDA_MEMORY_AFTER=$(docker stats camunda-7 --no-stream --format "{{.MemUsage}}" 2>/dev/null | awk '{print $1}')
echo "Memory after test: $CAMUNDA_MEMORY_AFTER"

echo -e "${GREEN}✅ Camunda load test complete${NC}"
echo "   Results: $RESULTS_DIR/camunda_${TIMESTAMP}.json"
echo ""

# Stop Camunda COMPLETELY
echo -e "${YELLOW}Stopping Camunda and ALL its containers...${NC}"
cd deployments/camunda
docker-compose down -v --remove-orphans
cd ../..
echo -e "${GREEN}✅ Camunda fully stopped${NC}"

# Verify all stopped
echo "Verifying no Camunda containers running..."
CAMUNDA_RUNNING=$(docker ps --filter "name=camunda" --format "{{.Names}}" | wc -l | tr -d ' ')
if [ "$CAMUNDA_RUNNING" -gt 0 ]; then
    echo -e "${RED}⚠️  Warning: Camunda containers still running!${NC}"
    docker ps --filter "name=camunda"
fi

# Wait for full cleanup
sleep 5
echo ""

# Phase 2: DMN Engine Go (in Docker container)
echo -e "\n${BLUE}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}  Phase 2: Load Testing DMN Engine Go (Docker container)        ${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}\n"

# Start ALL DMN Go services in Docker
echo "Starting DMN Engine Go in Docker container..."
docker-compose up -d --build
sleep 5

# Wait for DMN Go to be healthy
echo "Waiting for DMN Go to become healthy..."
for i in {1..60}; do
    HEALTH=$(docker inspect --format='{{.State.Health.Status}}' dmn-engine 2>/dev/null)
    if [ "$HEALTH" = "healthy" ]; then
        echo -e "${GREEN}DMN Go is healthy!${NC}"
        break
    fi
    echo "  Waiting... ($i/60) status: $HEALTH"
    sleep 2
done

# Verify it's running
if ! curl -s "http://localhost:8080/health" > /dev/null 2>&1; then
    echo -e "${RED}ERROR: DMN Go health check failed!${NC}"
    docker logs dmn-engine --tail 20
    exit 1
fi
echo ""

# Deploy DMN files
echo "Deploying DMN definitions to DMN Go..."
DEPLOYED=0
for file in testdata/dmn/stress/*.dmn; do
    [ -f "$file" ] || continue
    curl -s -X POST "http://localhost:8080/api/v1/definitions" \
        -H "Content-Type: application/xml" \
        --data-binary "@$file" > /dev/null 2>&1 && DEPLOYED=$((DEPLOYED + 1))
done
echo "Deployed: $DEPLOYED decisions"
echo ""

# Memory before test
DMN_GO_MEMORY_BEFORE=$(docker stats dmn-engine --no-stream --format "{{.MemUsage}}" 2>/dev/null | awk '{print $1}')
if [ -z "$DMN_GO_MEMORY_BEFORE" ] || [ "$DMN_GO_MEMORY_BEFORE" = "0B" ]; then
    DMN_GO_MEMORY_BEFORE="N/A"
fi
echo "Memory before test: $DMN_GO_MEMORY_BEFORE"

# Run k6 load test
echo -e "${GREEN}Running k6 load test...${NC}"
k6 run \
    --out json="$RESULTS_DIR/dmn_go_${TIMESTAMP}.json" \
    --summary-export="$RESULTS_DIR/dmn_go_summary_${TIMESTAMP}.json" \
    -e DMN_GO_URL=http://localhost:8080 \
    benchmarks/k6/dmn_go_load_test.js || {
        echo -e "${YELLOW}⚠️  k6 test completed with warnings (threshold violations are OK)${NC}"
    }

# Memory after test
DMN_GO_MEMORY_AFTER=$(docker stats dmn-engine --no-stream --format "{{.MemUsage}}" 2>/dev/null | awk '{print $1}')
if [ -z "$DMN_GO_MEMORY_AFTER" ] || [ "$DMN_GO_MEMORY_AFTER" = "0B" ]; then
    DMN_GO_MEMORY_AFTER="N/A"
fi
echo "Memory after test: $DMN_GO_MEMORY_AFTER"

echo -e "${GREEN}✅ DMN Go load test complete${NC}"
echo "   Results: $RESULTS_DIR/dmn_go_${TIMESTAMP}.json"
echo ""

# Stop DMN Go COMPLETELY
echo -e "${YELLOW}Stopping DMN Engine Go and ALL its containers...${NC}"
docker-compose down -v --remove-orphans
echo -e "${GREEN}✅ DMN Go fully stopped${NC}"

# Verify all stopped
echo "Verifying no DMN Go containers running..."
DMN_RUNNING=$(docker ps --filter "name=dmn" --format "{{.Names}}" | wc -l | tr -d ' ')
if [ "$DMN_RUNNING" -gt 0 ]; then
    echo -e "${RED}⚠️  Warning: DMN containers still running!${NC}"
    docker ps --filter "name=dmn"
fi

sleep 2
echo ""

# Generate comparison report
echo -e "\n${BLUE}╔════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║              GENERATING COMPARISON REPORT                       ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════════╝${NC}\n"

python3 benchmarks/analyze_results.py \
    "$RESULTS_DIR/camunda_summary_${TIMESTAMP}.json" \
    "$RESULTS_DIR/dmn_go_summary_${TIMESTAMP}.json" \
    > "$RESULTS_DIR/comparison_${TIMESTAMP}.txt"

cat "$RESULTS_DIR/comparison_${TIMESTAMP}.txt"

echo -e "\n${GREEN}✅ Benchmark complete!${NC}"
echo "   Results directory: $RESULTS_DIR"
echo "   Timestamp: $TIMESTAMP"
echo ""
echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}   Sequential Testing Summary                                  ${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
echo -e "${GREEN}✅ Phase 1: Camunda 7 tested in isolation${NC}"
echo -e "${GREEN}✅ Phase 2: DMN Go tested in isolation${NC}"
echo -e "${GREEN}✅ No resource contention${NC}"
echo -e "${GREEN}✅ Fair apples-to-apples comparison${NC}"
echo ""
echo "View results:"
echo "  cat $RESULTS_DIR/comparison_${TIMESTAMP}.txt"
echo ""

