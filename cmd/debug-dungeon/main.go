package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	dungeonsRepo "github.com/KirkDiggler/dnd-bot-discord/internal/repositories/dungeons"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found")
	}

	// Parse command line arguments
	sessionID := flag.String("session", "", "Session ID to debug")
	flag.Parse()

	if *sessionID == "" {
		log.Fatal("Please provide a session ID with -session flag")
	}

	// Setup Redis
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379"
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("Failed to parse Redis URL: %v", err)
	}

	client := redis.NewClient(opt)
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	// Create repository
	repo := dungeonsRepo.NewRedisRepository(&dungeonsRepo.RedisRepoConfig{
		Client: client,
	})

	// Find active dungeon for this session
	dungeon, err := repo.GetBySession(ctx, *sessionID)
	if err != nil {
		log.Fatalf("Failed to get dungeon: %v", err)
	}

	if dungeon == nil {
		fmt.Printf("No active dungeon found for session %s\n", *sessionID)
		return
	}

	// Display dungeon information
	fmt.Printf("=== Active Dungeon ===\n")
	fmt.Printf("ID: %s\n", dungeon.ID)
	fmt.Printf("State: %s\n", dungeon.State)
	fmt.Printf("Room Number: %d\n", dungeon.RoomNumber)
	fmt.Printf("Party Size: %d\n", len(dungeon.Party))

	if dungeon.CurrentRoom != nil {
		fmt.Printf("Current Room: %s (%s)\n", dungeon.CurrentRoom.Name, dungeon.CurrentRoom.Type)
	}

	fmt.Printf("\nParty Members:\n")
	for j, member := range dungeon.Party {
		fmt.Printf("  %d. User: %s, Character: %s, Status: %s\n",
			j+1, member.UserID, member.CharacterID, member.Status)
	}

	fmt.Printf("\nState Checks:\n")
	fmt.Printf("  CanEnterRoom(): %v\n", dungeon.CanEnterRoom())
	fmt.Printf("  IsActive(): %v\n", dungeon.IsActive())

	// Debug the logic
	fmt.Printf("\nDebug Info:\n")
	fmt.Printf("  Party Size > 0: %v\n", len(dungeon.Party) > 0)
	fmt.Printf("  State is AwaitingParty: %v\n", dungeon.State == entities.DungeonStateAwaitingParty)
	fmt.Printf("  State is RoomReady: %v\n", dungeon.State == entities.DungeonStateRoomReady)

	if dungeon.Metadata != nil {
		fmt.Printf("\nMetadata:\n")
		for k, v := range dungeon.Metadata {
			fmt.Printf("  %s: %v\n", k, v)
		}
	}
}
