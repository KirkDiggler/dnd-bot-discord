.PHONY: help test test-unit test-integration test-all coverage build run clean lint install-tools

# Default target
help:
	@echo "Available targets:"
	@echo "  make test           - Run unit tests"
	@echo "  make test-unit      - Run unit tests only"
	@echo "  make test-integration - Run integration tests (requires Redis)"
	@echo "  make test-all       - Run all tests"
	@echo "  make coverage       - Run tests with coverage report"
	@echo "  make build          - Build the bot"
	@echo "  make run            - Run the bot"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make lint           - Run linters"
	@echo "  make fmt            - Format code and tidy modules"
	@echo "  make pre-commit     - Run all pre-commit checks (fmt, lint, test)"
	@echo "  make install-tools  - Install development tools"
	@echo ""
	@echo "GitHub workflow:"
	@echo "  make start-issue ISSUE=<num> - Start working on an issue"
	@echo "  make create-pr ISSUE=<num> TITLE=\"<title>\" - Create PR for an issue"

# Test targets
test: test-unit

test-unit:
	@echo "Running unit tests..."
	@go test ./... -short -race -v

test-unit-no-race:
	@echo "Running unit tests without race detector..."
	@CGO_ENABLED=0 go test ./... -short -v

test-integration:
	@echo "Running integration tests..."
	@go test ./... -tags=integration -race -v

test-integration-no-race:
	@echo "Running integration tests without race detector..."
	@CGO_ENABLED=0 go test ./... -tags=integration -v

test-all: test-unit test-integration

test-all-no-race: test-unit-no-race test-integration-no-race

# Coverage target
coverage:
	@echo "Running tests with coverage..."
	@go test ./... -short -race -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

coverage-integration:
	@echo "Running all tests with coverage..."
	@go test ./... -tags=integration -race -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Build target
build:
	@echo "Building bot..."
	@go build -o bin/dnd-bot ./cmd/bot

# Run target
run: build
	@echo "Running bot..."
	@./bin/dnd-bot

# Clean target
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@go clean -testcache

# Lint target
lint:
	@echo "Running linters..."
	@/home/kirk/go/bin/golangci-lint run ./...

# Install development tools
install-tools:
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/golang/mock/mockgen@latest
	@go install github.com/stretchr/testify@latest

# Generate mocks
generate-mocks:
	@echo "Generating mocks..."
	@PATH="/home/kirk/go/bin:$$PATH" go generate ./...

# Docker targets
docker-build:
	@echo "Building Docker image..."
	@docker build -t dnd-bot-discord:latest .

docker-run:
	@echo "Running bot in Docker..."
	@docker run --rm -it --env-file .env dnd-bot-discord:latest

# Redis for testing
redis-start:
	@echo "Starting Redis for testing..."
	@docker run -d --name dnd-bot-redis -p 6379:6379 redis:alpine

redis-stop:
	@echo "Stopping Redis..."
	@docker stop dnd-bot-redis || true
	@docker rm dnd-bot-redis || true

# Development workflow
dev-test: redis-start test-all redis-stop

# CI/CD simulation
ci: lint test-all

# Format targets
format fmt:
	@echo "Formatting Go code..."
	@go fmt ./...
	@go mod tidy

# Pre-commit checks
.PHONY: pre-commit
pre-commit:
	@echo "Running pre-commit checks..."
	@echo "→ Formatting code..."
	@go fmt ./...
	@echo "→ Tidying modules..."
	@go mod tidy
	@echo "→ Running linter..."
	@/home/kirk/go/bin/golangci-lint run ./...
	@echo "→ Running unit tests..."
	@if command -v gcc >/dev/null 2>&1; then \
		echo "  GCC detected, running with race detector..."; \
		go test ./... -short -race; \
	else \
		echo "  GCC not found, running without race detector..."; \
		CGO_ENABLED=0 go test ./... -short; \
	fi
	@echo "✓ All pre-commit checks passed!"

# GitHub workflow helpers
.PHONY: start-issue create-pr
start-issue:
	@if [ -z "$(ISSUE)" ]; then \
		echo "Usage: make start-issue ISSUE=<number>"; \
		echo "Example: make start-issue ISSUE=57"; \
		exit 1; \
	fi
	@./scripts/start-issue.sh $(ISSUE)

create-pr:
	@if [ -z "$(ISSUE)" ] || [ -z "$(TITLE)" ]; then \
		echo "Usage: make create-pr ISSUE=<number> TITLE=\"<title>\""; \
		echo "Example: make create-pr ISSUE=57 TITLE=\"Refactor attack logic\""; \
		exit 1; \
	fi
	@./scripts/create-pr.sh $(ISSUE) "$(TITLE)"
