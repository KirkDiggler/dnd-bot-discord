package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

func main() {
	ctx := context.Background()

	// Set up Redis
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379/0"
	}

	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("Failed to parse Redis URL: %v", err)
	}

	client := redis.NewClient(opts)
	defer client.Close()

	// Test connection
	if _, pingErr := client.Ping(ctx).Result(); pingErr != nil {
		log.Fatalf("Failed to connect to Redis: %v", pingErr)
	}

	// Find all session keys
	sessionKeys, err := client.Keys(ctx, "session:*").Result()
	if err != nil {
		log.Fatalf("Failed to get session keys: %v", err)
	}

	fmt.Printf("Found %d sessions:\n", len(sessionKeys))
	for _, key := range sessionKeys {
		// Get the session data
		data, getErr := client.Get(ctx, key).Result()
		if getErr != nil {
			fmt.Printf("  %s: ERROR - %v\n", key, getErr)
			continue
		}

		// Just show basic info
		fmt.Printf("  %s: %d bytes\n", key, len(data))
	}

	// Also find dungeon keys
	dungeonKeys, err := client.Keys(ctx, "dungeon:*").Result()
	if err != nil {
		log.Fatalf("Failed to get dungeon keys: %v", err)
	}

	fmt.Printf("\nFound %d dungeons:\n", len(dungeonKeys))
	for _, key := range dungeonKeys {
		fmt.Printf("  %s\n", key)
	}
}
