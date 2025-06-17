.PHONY: run build test clean redis-up redis-down

# Run the bot
run:
	go run cmd/bot/main.go

# Start Redis using Docker Compose
redis-up:
	docker-compose up -d redis

# Stop Redis
redis-down:
	docker-compose down

# Run bot with Redis
run-with-redis: redis-up run

# Build the bot
build:
	go build -o bin/dnd-bot cmd/bot/main.go

# Run tests
test:
	go test ./... -v

# Run tests with coverage
test-coverage:
	go test -cover ./...

# Run unit tests only (fast)
test-unit:
	go test ./... -v -short

# Run integration tests (requires Docker)
test-integration:
	docker-compose -f docker-compose.test.yml up -d
	go test ./... -v -tags=integration
	docker-compose -f docker-compose.test.yml down

# Run all tests with coverage
test-all: test-unit test-integration
	go test ./... -coverprofile=coverage.out -tags=integration
	go tool cover -html=coverage.out

# Clean build artifacts
clean:
	rm -rf bin/

# Generate mocks
generate-mocks:
	go generate ./...

# Install dependencies
deps:
	go mod download
	go mod tidy

# Docker development commands
docker-up-debug:
	docker-compose --profile debug up -d

# Build for Raspberry Pi
build-pi:
	GOOS=linux GOARCH=arm64 go build -o bin/dnd-bot-arm64 cmd/bot/main.go
	GOOS=linux GOARCH=arm GOARM=7 go build -o bin/dnd-bot-armv7 cmd/bot/main.go

# Deploy to Raspberry Pi
deploy-pi:
	chmod +x scripts/deploy-to-pi.sh
	./scripts/deploy-to-pi.sh