package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/redis/go-redis/v9"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: debug-session <session-id>")
		os.Exit(1)
	}

	sessionID := os.Args[1]
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

	// Get session data
	data, err := client.Get(ctx, fmt.Sprintf("session:%s", sessionID)).Result()
	if err != nil {
		log.Fatalf("Failed to get session: %v", err)
	}

	// Parse the session
	var session entities.Session
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		log.Fatalf("Failed to parse session: %v", err)
	}

	fmt.Printf("Session ID: %s\n", session.ID)
	fmt.Printf("Name: %s\n", session.Name)
	fmt.Printf("Creator: %s\n", session.CreatorID)
	fmt.Printf("Channel: %s\n", session.ChannelID)
	fmt.Printf("Status: %s\n", session.Status)
	fmt.Printf("Members: %d\n", len(session.Members))

	for userID, member := range session.Members {
		fmt.Printf("  %s: %s (character: %s)\n", userID, member.Role, member.CharacterID)
	}

	fmt.Printf("Metadata: %d items\n", len(session.Metadata))
	for key, value := range session.Metadata {
		fmt.Printf("  %s: %v\n", key, value)
	}
}
