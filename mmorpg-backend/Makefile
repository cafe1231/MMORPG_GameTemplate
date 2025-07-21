# MMORPG Backend Makefile

# Variables
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOLINT=golangci-lint
BINARY_NAME=mmorpg
DOCKER_COMPOSE=docker-compose

# Service names
SERVICES=gateway auth character world game

# Build flags
LDFLAGS=-ldflags "-s -w"

# Paths
CMD_PATH=./cmd
INTERNAL_PATH=./internal
PKG_PATH=./pkg

.PHONY: all build clean test coverage fmt lint help proto docker-up docker-down

# Default target
all: clean fmt lint test build

# Build all services
build:
	@echo "Building all services..."
	@for service in $(SERVICES); do \
		echo "Building $$service..."; \
		$(GOBUILD) $(LDFLAGS) -o bin/$$service $(CMD_PATH)/$$service/main.go || exit 1; \
	done
	@echo "Build complete!"

# Build specific service
build-%:
	@echo "Building $*..."
	$(GOBUILD) $(LDFLAGS) -o bin/$* $(CMD_PATH)/$*/main.go

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf bin/
	@echo "Clean complete!"

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v -race -cover ./...

# Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) -s -w .
	@echo "Format complete!"

# Run linter
lint:
	@echo "Running linter..."
	@if ! command -v $(GOLINT) &> /dev/null; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	$(GOLINT) run --timeout=5m ./...

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "Dependencies updated!"

# Generate protobuf code
proto:
	@echo "Generating protobuf code..."
	@if ! command -v protoc &> /dev/null; then \
		echo "Error: protoc is not installed. Please install protocol buffers compiler."; \
		exit 1; \
	fi
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		$(PKG_PATH)/proto/*.proto
	@echo "Protobuf generation complete!"

# Run specific service
run-%:
	@echo "Running $*..."
	$(GOCMD) run $(CMD_PATH)/$*/main.go

# Docker commands
docker-build:
	@echo "Building Docker images..."
	@for service in $(SERVICES); do \
		echo "Building $$service image..."; \
		docker build -f deployments/docker/Dockerfile.$$service -t mmorpg-$$service:latest . || exit 1; \
	done
	@echo "Docker build complete!"

docker-up:
	@echo "Starting Docker environment..."
	cd deployments/docker && $(DOCKER_COMPOSE) up -d
	@echo "Docker environment started!"

docker-down:
	@echo "Stopping Docker environment..."
	cd deployments/docker && $(DOCKER_COMPOSE) down
	@echo "Docker environment stopped!"

docker-logs:
	cd deployments/docker && $(DOCKER_COMPOSE) logs -f

# Database migrations
migrate-up:
	@echo "Running database migrations..."
	migrate -path ./migrations -database "postgresql://dev:dev@localhost:5432/mmorpg?sslmode=disable" up

migrate-down:
	@echo "Rolling back database migrations..."
	migrate -path ./migrations -database "postgresql://dev:dev@localhost:5432/mmorpg?sslmode=disable" down

migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir ./migrations -seq $$name

# Development helpers
dev-setup:
	@echo "Setting up development environment..."
	@make deps
	@make docker-up
	@echo "Waiting for services to start..."
	@sleep 5
	@make migrate-up
	@echo "Development environment ready!"

dev-reset:
	@echo "Resetting development environment..."
	@make docker-down
	@docker volume prune -f
	@make dev-setup

# Benchmarks
bench:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

# Security scan
security:
	@echo "Running security scan..."
	@if ! command -v gosec &> /dev/null; then \
		echo "Installing gosec..."; \
		go install github.com/securego/gosec/v2/cmd/gosec@latest; \
	fi
	gosec -fmt=json -out=security-report.json ./... || true
	@echo "Security scan complete! Check security-report.json"

# Help
help:
	@echo "MMORPG Backend Makefile Commands:"
	@echo ""
	@echo "  make all          - Clean, format, lint, test, and build"
	@echo "  make build        - Build all services"
	@echo "  make build-<svc>  - Build specific service (gateway, auth, etc.)"
	@echo "  make clean        - Remove build artifacts"
	@echo "  make test         - Run tests"
	@echo "  make coverage     - Run tests with coverage report"
	@echo "  make fmt          - Format code"
	@echo "  make lint         - Run linter"
	@echo "  make deps         - Download and tidy dependencies"
	@echo "  make proto        - Generate protobuf code"
	@echo "  make run-<svc>    - Run specific service"
	@echo "  make docker-build - Build Docker images"
	@echo "  make docker-up    - Start Docker environment"
	@echo "  make docker-down  - Stop Docker environment"
	@echo "  make docker-logs  - View Docker logs"
	@echo "  make migrate-up   - Run database migrations"
	@echo "  make migrate-down - Rollback database migrations"
	@echo "  make dev-setup    - Setup development environment"
	@echo "  make dev-reset    - Reset development environment"
	@echo "  make bench        - Run benchmarks"
	@echo "  make security     - Run security scan"
	@echo "  make help         - Show this help message"