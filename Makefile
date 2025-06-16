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
	go test ./...

# Run tests with coverage
test-coverage:
	go test -cover ./...

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