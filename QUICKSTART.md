# Quick Start Guide

## Prerequisites

- Go 1.21+
- Docker & Docker Compose
- k6 (for benchmarking)

```bash
# Install k6
brew install k6  # macOS
# or
sudo apt install k6  # Linux
```

## 1. Start Services

```bash
# Start database
make db-up

# Run DMN Engine
make run
```

## 2. Test API

```bash
# Health check
curl http://localhost:8080/health

# Deploy a DMN decision
curl -X POST http://localhost:8080/api/v1/definitions \
  -H "Content-Type: application/xml" \
  --data-binary @testdata/dmn/simple_decision.dmn

# Evaluate decision
curl -X POST http://localhost:8080/api/v1/evaluate \
  -H "Content-Type: application/json" \
  -d '{
    "decisionKey": "eligibility",
    "variables": {"age": 25}
  }'
```

## 3. Run Benchmark

```bash
# Professional k6 benchmark
make benchmark DECISIONS=50

# Quick test (10 decisions)
make benchmark-quick

# Results in: benchmarks/results/
```

## 4. View Results

```bash
# Latest comparison
ls -lt benchmarks/results/ | head -5

# View comparison report
cat benchmarks/results/comparison_*.txt
```

## Development

```bash
# Build
make build

# Run tests
make test

# Lint
make lint

# Generate DMN test tables
./scripts/generate_camunda_dmn.sh 100
```

## Troubleshooting

### k6 not found

```bash
# macOS
brew install k6

# Linux
curl -L https://github.com/grafana/k6/releases/download/v0.47.0/k6-v0.47.0-linux-amd64.tar.gz | tar xvz
sudo cp k6-v0.47.0-linux-amd64/k6 /usr/local/bin/
```

### Port already in use

```bash
# Stop all services
make stop

# Or kill specific port
lsof -ti:8080 | xargs kill -9
```

### Database connection failed

```bash
# Reset database
make db-down
make db-up
```

## Next Steps

- Read [Architecture Documentation](docs/architecture/ARCHITECTURE.md)
- Check [Benchmark Guide](benchmarks/README.md)
- See [API Documentation](README.md#api-reference)
