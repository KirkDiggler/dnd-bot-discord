#!/bin/bash

# Script to run integration tests with Redis

echo "Starting integration test for character creation flow..."

# Check if Redis is running
if ! redis-cli ping > /dev/null 2>&1; then
    echo "Redis is not running. Starting test Redis container..."
    docker run -d --name dnd-bot-test-redis -p 6380:6379 redis:alpine
    sleep 2
    export REDIS_URL="redis://localhost:6380/0"
else
    export REDIS_URL="redis://localhost:6379/0"
fi

# Run the integration test
export INTEGRATION_TEST=true
go test -v ./internal/services/character -run TestCharacterCreationFlow_FullIntegration

# Clean up test Redis if we started it
if docker ps | grep -q dnd-bot-test-redis; then
    echo "Stopping test Redis container..."
    docker stop dnd-bot-test-redis
    docker rm dnd-bot-test-redis
fi