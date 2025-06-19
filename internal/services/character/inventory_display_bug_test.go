//go:build integration
// +build integration

package character_test

import (
	"context"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	charactersRepo "github.com/KirkDiggler/dnd-bot-discord/internal/repositories/characters"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInventoryDisplayBug reproduces the issue where inventory command shows empty
// even though character has equipment
func TestInventoryDisplayBug(t *testing.T) {
	// Set up Redis client
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379/0"
	}

	opts, err := redis.ParseURL(redisURL)
	require.NoError(t, err)

	client := redis.NewClient(opts)
	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("Failed to close Redis client: %v", err)
		}
	}()

	ctx := context.Background()
	_, err = client.Ping(ctx).Result()
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}

	defer func() { _ = client.FlushDB(ctx) }()

	repo := charactersRepo.NewRedisRepository(&charactersRepo.RedisRepoConfig{
		Client: client,
	})

	dndClient, err := dnd5e.New(&dnd5e.Config{
		HttpClient: &http.Client{},
	})
	require.NoError(t, err)

	svc := character.NewService(&character.ServiceConfig{
		DNDClient:  dndClient,
		Repository: repo,
	})

	t.Run("barbarian starting equipment types", func(t *testing.T) {
		// Create a barbarian like Standre
		draft, err := svc.GetOrCreateDraftCharacter(ctx, "test-barbarian", "test-realm")
		require.NoError(t, err)
		
		// Set race and class
		raceKey := "half-orc"
		classKey := "barbarian"
		_, err = svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
			RaceKey:  &raceKey,
			ClassKey: &classKey,
		})
		require.NoError(t, err)

		// Add typical barbarian starting equipment
		// Based on the logs, let's try common barbarian choices
		_, err = svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
			Equipment: []string{"greataxe"}, // Barbarian weapon choice
		})
		require.NoError(t, err)

		_, err = svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
			Equipment: []string{"handaxe"}, // Two handaxes is common barbarian choice
		})
		require.NoError(t, err)

		_, err = svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
			Equipment: []string{"explorers-pack"}, // Common pack choice
		})
		require.NoError(t, err)

		// Finalize
		finalChar, err := svc.FinalizeCharacterWithName(ctx, draft.ID, "Test Barbarian", raceKey, classKey)
		require.NoError(t, err)

		// Check equipment types
		log.Println("=== Equipment by Type ===")
		totalItems := 0
		weaponTypeCount := 0
		basicEquipmentCount := 0
		
		for eqType, items := range finalChar.Inventory {
			log.Printf("Type '%s': %d items", eqType, len(items))
			totalItems += len(items)
			
			if eqType == entities.EquipmentTypeWeapon {
				weaponTypeCount = len(items)
			}
			if eqType == "BasicEquipment" {
				basicEquipmentCount = len(items)
			}
			
			for _, item := range items {
				log.Printf("  - %s (key: %s, type from GetEquipmentType(): %s)", 
					item.GetName(), item.GetKey(), item.GetEquipmentType())
			}
		}

		log.Printf("Total items: %d", totalItems)
		log.Printf("Items with EquipmentTypeWeapon: %d", weaponTypeCount)
		log.Printf("Items with 'BasicEquipment' type: %d", basicEquipmentCount)

		// The bug: inventory command only shows EquipmentTypeWeapon
		// but weapons were being stored with type "Weapon" instead of "weapon"
		assert.Equal(t, 3, totalItems, "Should have 3 total items")
		
		// This is what the inventory handler checks
		assert.Greater(t, weaponTypeCount, 0, "Should have weapons with EquipmentTypeWeapon")
		
		// After fix: weapons should use the constant
		for _, item := range finalChar.Inventory[entities.EquipmentTypeWeapon] {
			assert.Equal(t, entities.EquipmentTypeWeapon, item.GetEquipmentType(), 
				"Weapon should return EquipmentTypeWeapon constant")
		}
	})

	t.Run("check what types D&D API returns", func(t *testing.T) {
		// Let's see what the API actually returns for these items
		items := []string{"greataxe", "handaxe", "explorers-pack"}
		
		for _, key := range items {
			equipment, err := dndClient.GetEquipment(key)
			if err != nil {
				log.Printf("Error getting %s: %v", key, err)
				continue
			}
			
			log.Printf("Equipment '%s': Type = %T, GetEquipmentType() = %s", 
				key, equipment, equipment.GetEquipmentType())
		}
	})
}