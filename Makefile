# DMN Engine Go - Makefile

.PHONY: help build run run-debug test db-up db-down clean docker-build camunda-up camunda-down benchmark

help: ## Show this help
	@echo "DMN Engine Go - Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary
	@echo "Building DMN Engine Go..."
	@mkdir -p bin
	@go build -o bin/dmn-engine ./cmd/server
	@echo "✅ Build complete: bin/dmn-engine"

run: build ## Build and run the server
	@echo "Starting DMN Engine Go..."
	@export $$(cat dev.env | xargs) && ./bin/dmn-engine

run-debug: build ## Run with debug logging
	@echo "Starting DMN Engine Go (DEBUG)..."
	@export LOG_LEVEL=debug && export $$(cat dev.env | xargs) && ./bin/dmn-engine

test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

db-up: ## Start PostgreSQL
	@echo "Starting PostgreSQL..."
	@docker-compose up -d postgres
	@echo "✅ PostgreSQL started on port 5432"

db-down: ## Stop PostgreSQL
	@echo "Stopping PostgreSQL..."
	@docker-compose down
	@echo "✅ PostgreSQL stopped"

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@echo "✅ Cleaned"

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -f deployments/docker/Dockerfile -t dmn-engine-go:latest .
	@echo "✅ Docker image built"

# Camunda benchmarking
camunda-up: ## Start Camunda 7 for benchmarking
	@echo "Starting Camunda 7..."
	@cd deployments/camunda && docker-compose up -d
	@echo "⏳ Waiting for Camunda to start (this takes ~30-60 seconds)..."
	@sleep 10
	@echo "Check status: docker-compose -f deployments/camunda/docker-compose.yml logs -f camunda"
	@echo "Camunda UI: http://localhost:8081/camunda/ (admin/admin)"

camunda-down: ## Stop Camunda 7
	@echo "Stopping Camunda 7..."
	@cd deployments/camunda && docker-compose down
	@echo "✅ Camunda stopped"

camunda-logs: ## Show Camunda logs
	@cd deployments/camunda && docker-compose logs -f camunda

# Benchmarking
benchmark: ## Run professional k6 load test (usage: make benchmark DECISIONS=50)
	@echo "🚀 Running professional k6 benchmark with $(or $(DECISIONS),50) decisions..."
	@./benchmarks/run_benchmark.sh $(or $(DECISIONS),50)

benchmark-quick: ## Quick benchmark with fewer decisions (10 decisions)
	@echo "⚡ Running quick benchmark with 10 decisions..."
	@./benchmarks/run_benchmark.sh 10

# Setup everything
setup-all: db-up ## Setup database
	@echo "✅ Database ready!"
	@echo ""
	@echo "Run benchmarks: make benchmark DECISIONS=50"

demo: ## Quick demo deployment
	@echo "Demo: Deploying sample DMN..."
	@curl -X POST http://localhost:8080/api/v1/definitions \
		-H "Content-Type: application/xml" \
		--data-binary @testdata/dmn/simple_decision.dmn
	@echo ""
	@echo "Demo: Evaluating decision..."
	@curl -X POST http://localhost:8080/api/v1/evaluate \
		-H "Content-Type: application/json" \
		-d '{"decisionKey":"eligibility","variables":{"age":25}}' | jq .
