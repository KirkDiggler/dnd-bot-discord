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
		fmt.Println("Usage: fix-character <character-id>")
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
	defer client.Close()

	// Test connection
	if _, err := client.Ping(ctx).Result(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	// Create repository
	repo := charactersRepo.NewRedisRepository(&charactersRepo.RedisRepoConfig{
		Client: client,
	})

	// Get the character
	char, err := repo.Get(ctx, characterID)
	if err != nil {
		log.Fatalf("Failed to get character: %v", err)
	}

	log.Printf("Character: %s (ID: %s)", char.Name, char.ID)
	log.Printf("Status: %s, Level: %d", char.Status, char.Level)
	log.Printf("Race: %v, Class: %v", char.Race != nil, char.Class != nil)
	log.Printf("Attributes: %d, AbilityAssignments: %d, AbilityRolls: %d",
		len(char.Attributes), len(char.AbilityAssignments), len(char.AbilityRolls))

	// If character already has attributes, no fix needed
	if len(char.Attributes) > 0 {
		log.Println("Character already has attributes, no fix needed")
		return
	}

	// Create default ability scores for a level 1 character
	// Using standard array: 15, 14, 13, 12, 10, 8
	log.Println("Fixing character with standard ability array...")

	// Initialize attributes map
	char.Attributes = make(map[entities.Attribute]*entities.AbilityScore)

	// Assign scores based on class (optimized for each class)
	scores := map[entities.Attribute]int{}
	
	if char.Class != nil {
		switch char.Class.Key {
		case "monk":
			scores[entities.AttributeDexterity] = 15
			scores[entities.AttributeWisdom] = 14
			scores[entities.AttributeConstitution] = 13
			scores[entities.AttributeStrength] = 12
			scores[entities.AttributeIntelligence] = 10
			scores[entities.AttributeCharisma] = 8
		case "fighter":
			scores[entities.AttributeStrength] = 15
			scores[entities.AttributeConstitution] = 14
			scores[entities.AttributeDexterity] = 13
			scores[entities.AttributeWisdom] = 12
			scores[entities.AttributeCharisma] = 10
			scores[entities.AttributeIntelligence] = 8
		case "wizard":
			scores[entities.AttributeIntelligence] = 15
			scores[entities.AttributeConstitution] = 14
			scores[entities.AttributeDexterity] = 13
			scores[entities.AttributeWisdom] = 12
			scores[entities.AttributeCharisma] = 10
			scores[entities.AttributeStrength] = 8
		default:
			// Generic distribution
			scores[entities.AttributeStrength] = 15
			scores[entities.AttributeDexterity] = 14
			scores[entities.AttributeConstitution] = 13
			scores[entities.AttributeIntelligence] = 12
			scores[entities.AttributeWisdom] = 10
			scores[entities.AttributeCharisma] = 8
		}
	} else {
		// No class, use generic
		scores[entities.AttributeStrength] = 15
		scores[entities.AttributeDexterity] = 14
		scores[entities.AttributeConstitution] = 13
		scores[entities.AttributeIntelligence] = 12
		scores[entities.AttributeWisdom] = 10
		scores[entities.AttributeCharisma] = 8
	}

	// Apply scores and racial bonuses
	for attr, baseScore := range scores {
		score := baseScore
		
		// Apply racial bonuses
		if char.Race != nil {
			for _, bonus := range char.Race.AbilityBonuses {
				if bonus.Attribute == attr {
					score += bonus.Bonus
					log.Printf("Applied racial bonus to %s: +%d", attr, bonus.Bonus)
				}
			}
		}
		
		// Calculate modifier
		modifier := (score - 10) / 2
		
		// Create ability score
		char.Attributes[attr] = &entities.AbilityScore{
			Score: score,
			Bonus: modifier,
		}
		
		log.Printf("Set %s: %d (modifier: %+d)", attr, score, modifier)
	}

	// Calculate HP if not set
	if char.MaxHitPoints == 0 && char.Class != nil {
		conMod := 0
		if con, ok := char.Attributes[entities.AttributeConstitution]; ok && con != nil {
			conMod = con.Bonus
		}
		char.MaxHitPoints = char.Class.HitDie + conMod
		char.CurrentHitPoints = char.MaxHitPoints
		log.Printf("Calculated HP: %d", char.MaxHitPoints)
	}

	// Calculate AC if not set
	if char.AC == 0 {
		baseAC := 10
		dexMod := 0
		
		if dex, ok := char.Attributes[entities.AttributeDexterity]; ok && dex != nil {
			dexMod = dex.Bonus
		}
		
		// Basic AC calculation (can be improved with features)
		char.AC = baseAC + dexMod
		log.Printf("Calculated AC: %d", char.AC)
	}

	// Save the fixed character
	if err := repo.Update(ctx, char); err != nil {
		log.Fatalf("Failed to save fixed character: %v", err)
	}

	log.Println("Character fixed successfully!")
	
	// Verify the fix
	fixed, err := repo.Get(ctx, characterID)
	if err != nil {
		log.Fatalf("Failed to verify fix: %v", err)
	}
	
	log.Printf("Verification - Attributes: %d, HP: %d/%d, AC: %d",
		len(fixed.Attributes), fixed.CurrentHitPoints, fixed.MaxHitPoints, fixed.AC)
}