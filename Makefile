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
	@echo "  make install-tools  - Install development tools"

# Test targets
test: test-unit

test-unit:
	@echo "Running unit tests..."
	@go test ./... -short -race -v

test-integration:
	@echo "Running integration tests..."
	@go test ./... -tags=integration -race -v

test-all: test-unit test-integration

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
	@golangci-lint run ./...

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
	@gofmt -w .
	@go mod tidy

# Raspberry Pi deployment
PI_HOST = pi@10.0.0.129
PI_DIR = /home/pi/dnd-bot

.PHONY: deploy
deploy:
	@echo "🚀 Deploying to Raspberry Pi..."
	@echo "📦 Building for ARM64..."
	@GOOS=linux GOARCH=arm64 go build -o dnd-bot-arm64 ./cmd/bot
	@echo "📤 Copying files to Pi..."
	@ssh $(PI_HOST) "mkdir -p $(PI_DIR)"
	@scp dnd-bot-arm64 $(PI_HOST):$(PI_DIR)/dnd-bot
	@scp .env.production $(PI_HOST):$(PI_DIR)/.env 2>/dev/null || echo "⚠️  No .env.production found, using existing .env on Pi"
	@echo "🔄 Restarting service..."
	@ssh $(PI_HOST) "sudo systemctl restart dnd-bot" || echo "⚠️  Service not set up yet. Run 'make setup-pi' first"
	@rm dnd-bot-arm64
	@echo "✅ Deployment complete!"

.PHONY: setup-pi
setup-pi:
	@echo "🔧 Setting up systemd service on Pi..."
	@scp scripts/dnd-bot.service $(PI_HOST):/tmp/
	@ssh $(PI_HOST) "sudo mv /tmp/dnd-bot.service /etc/systemd/system/ && sudo systemctl daemon-reload && sudo systemctl enable dnd-bot"
	@echo "✅ Service installed and enabled!"

.PHONY: logs
logs:
	@echo "📜 Showing bot logs (Ctrl+C to exit)..."
	@ssh $(PI_HOST) "sudo journalctl -fu dnd-bot"

.PHONY: status
status:
	@ssh $(PI_HOST) "sudo systemctl status dnd-bot --no-pager"

.PHONY: stop
stop:
	@echo "🛑 Stopping bot..."
	@ssh $(PI_HOST) "sudo systemctl stop dnd-bot"

.PHONY: start
start:
	@echo "▶️  Starting bot..."
	@ssh $(PI_HOST) "sudo systemctl start dnd-bot"