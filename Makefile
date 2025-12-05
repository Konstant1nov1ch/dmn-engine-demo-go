.PHONY: build test run lint docker-build docker-run clean help db-up db-down

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=dmn-engine
BINARY_PATH=bin/$(BINARY_NAME)

# Environment file
ENV_FILE?=dev.env

# Load env file if exists
ifneq (,$(wildcard $(ENV_FILE)))
    include $(ENV_FILE)
    export
endif

# Database (fallback if not in env file)
DATABASE_URL?=postgres://dmn:dmn@localhost:5432/dmn?sslmode=disable

# Default target
all: build

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

## build: Build the application
build:
	$(GOBUILD) -o $(BINARY_PATH) ./cmd/server

## run: Run the application (requires PostgreSQL)
run:
	$(GOCMD) run ./cmd/server

## run-debug: Run with debug logging
run-debug:
	LOG_LEVEL=debug $(GOCMD) run ./cmd/server

## test: Run all tests
test:
	$(GOTEST) -v -race ./...

## test-short: Run short tests only
test-short:
	$(GOTEST) -v -short ./...

## test-coverage: Run tests with coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## lint: Run linter
lint:
	golangci-lint run

## fmt: Format code
fmt:
	$(GOCMD) fmt ./...

## mod-tidy: Tidy go modules
mod-tidy:
	$(GOMOD) tidy

## mod-download: Download go modules
mod-download:
	$(GOMOD) download

## db-up: Start PostgreSQL with docker-compose
db-up:
	docker-compose up -d postgres
	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 3
	@echo "PostgreSQL is ready at localhost:5432"

## db-down: Stop PostgreSQL
db-down:
	docker-compose down

## db-logs: Show PostgreSQL logs
db-logs:
	docker-compose logs -f postgres

## db-psql: Connect to PostgreSQL with psql
db-psql:
	docker-compose exec postgres psql -U dmn -d dmn

## docker-build: Build docker image
docker-build:
	docker build -t dmn-engine:latest -f deployments/docker/Dockerfile .

## clean: Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

## install-tools: Install development tools
install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

## demo: Deploy a sample DMN and test
demo:
	@echo "Deploying sample DMN..."
	curl -X POST http://localhost:8080/api/v1/definitions \
		-H "Content-Type: application/xml" \
		--data-binary @testdata/dmn/simple_decision.dmn
	@echo ""
	@echo ""
	@echo "Listing definitions..."
	curl http://localhost:8080/api/v1/definitions
	@echo ""
