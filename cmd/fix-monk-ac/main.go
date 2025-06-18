package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	charactersRepo "github.com/KirkDiggler/dnd-bot-discord/internal/repositories/characters"
	"github.com/redis/go-redis/v9"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: fix-monk-ac <character-id>")
		os.Exit(1)
	}

	characterID := os.Args[1]
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

	// Create repository
	repo := charactersRepo.NewRedisRepository(&charactersRepo.RedisRepoConfig{
		Client: client,
	})

	// Get the character
	char, err := repo.Get(ctx, characterID)
	if err != nil {
		log.Printf("Failed to get character: %v", err)
		return
	}

	log.Printf("Character: %s, Class: %s", char.Name, char.Class.Key)

	// Check if monk
	if char.Class.Key == "monk" {
		dexMod := 0
		wisMod := 0

		if dex, ok := char.Attributes[entities.AttributeDexterity]; ok && dex != nil {
			dexMod = dex.Bonus
		}
		if wis, ok := char.Attributes[entities.AttributeWisdom]; ok && wis != nil {
			wisMod = wis.Bonus
		}

		// Monk Unarmored Defense: 10 + DEX + WIS
		newAC := 10 + dexMod + wisMod
		log.Printf("Old AC: %d, New AC: %d (10 + %d DEX + %d WIS)", char.AC, newAC, dexMod, wisMod)

		char.AC = newAC

		if err := repo.Update(ctx, char); err != nil {
			log.Printf("Failed to save character: %v", err)
			return
		}

		log.Println("AC updated successfully!")
	} else {
		log.Println("Not a monk, no AC update needed")
	}
}
