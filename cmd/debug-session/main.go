package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/session"

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

	// Test connection first
	if _, pingErr := client.Ping(ctx).Result(); pingErr != nil {
		log.Fatalf("Failed to connect to Redis: %v", pingErr)
	}
	defer func() {
		clientErr := client.Close()
		if clientErr != nil {
			log.Printf("Failed to close Redis connection: %v", clientErr)
		}
	}()

	// Get session data
	data, err := client.Get(ctx, fmt.Sprintf("session:%s", sessionID)).Result()
	if err != nil {
		log.Printf("Failed to get session: %v", err)
		return
	}

	// Parse the sess
	var sess session.Session
	if err := json.Unmarshal([]byte(data), &sess); err != nil {
		log.Printf("Failed to parse session: %v", err)
		return
	}

	fmt.Printf("Session ID: %s\n", sess.ID)
	fmt.Printf("Name: %s\n", sess.Name)
	fmt.Printf("Creator: %s\n", sess.CreatorID)
	fmt.Printf("Channel: %s\n", sess.ChannelID)
	fmt.Printf("Status: %s\n", sess.Status)
	fmt.Printf("Members: %d\n", len(sess.Members))

	for userID, member := range sess.Members {
		fmt.Printf("  %s: %s (character: %s)\n", userID, member.Role, member.CharacterID)
	}

	fmt.Printf("Metadata: %d items\n", len(sess.Metadata))
	for key, value := range sess.Metadata {
		fmt.Printf("  %s: %v\n", key, value)
	}
}
