//go:build integration
// +build integration

package character_test

import (
	"context"
	character2 "github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e"
	charactersRepo "github.com/KirkDiggler/dnd-bot-discord/internal/repositories/characters"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCharacterCreationFlow_FullIntegration(t *testing.T) {

	// Set up Redis client
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379/15" // Use test database
	}

	opts, err := redis.ParseURL(redisURL)
	require.NoError(t, err)

	client := redis.NewClient(opts)
	defer func() {
		clientErr := client.Close()
		if clientErr != nil {
			log.Printf("Failed to close Redis client: %v", clientErr)
		}
	}()

	// Verify Redis is available
	ctx := context.Background()
	_, err = client.Ping(ctx).Result()
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}

	// Clean up test data
	defer func() { _ = client.FlushDB(ctx) }()

	// Create repository
	repo := charactersRepo.NewRedisRepository(&charactersRepo.RedisRepoConfig{
		Client: client,
	})

	// Create real D&D API client
	dndClient, err := dnd5e.New(&dnd5e.Config{
		HttpClient: &http.Client{}, // Create default HTTP client
	})
	require.NoError(t, err)

	// Create service
	svc := character.NewService(&character.ServiceConfig{
		DNDClient:  dndClient,
		Repository: repo,
	})

	userID := "test_user_123"
	realmID := "test_realm_456"

	t.Run("complete character creation flow - monk", func(t *testing.T) {
		// Add small delays between API calls to avoid overwhelming the D&D API
		const apiDelay = 100 * time.Millisecond

		// Step 1: Initial GetOrCreateDraftCharacter (simulating race selection)
		char1, err := svc.GetOrCreateDraftCharacter(ctx, userID, realmID)
		require.NoError(t, err)
		assert.Equal(t, character2.CharacterStatusDraft, char1.Status)
		assert.NotEmpty(t, char1.ID)

		originalID := char1.ID
		t.Logf("Created draft character: %s", originalID)

		// Step 2: Update with race (elf)
		time.Sleep(apiDelay) // Rate limit API calls
		raceUpdated, err := svc.UpdateDraftCharacter(ctx, originalID, &character.UpdateDraftInput{
			RaceKey: stringPtr("elf"),
		})
		require.NoError(t, err)
		assert.Equal(t, originalID, raceUpdated.ID)
		assert.NotNil(t, raceUpdated.Race)
		assert.Equal(t, "Elf", raceUpdated.Race.Name)

		// Verify GetOrCreateDraftCharacter returns same character
		char2, err := svc.GetOrCreateDraftCharacter(ctx, userID, realmID)
		require.NoError(t, err)
		assert.Equal(t, originalID, char2.ID, "Should return same draft character")

		// Step 3: Update with class (monk)
		time.Sleep(apiDelay) // Rate limit API calls
		classUpdated, err := svc.UpdateDraftCharacter(ctx, originalID, &character.UpdateDraftInput{
			ClassKey: stringPtr("monk"),
		})
		require.NoError(t, err)
		assert.Equal(t, originalID, classUpdated.ID)
		assert.NotNil(t, classUpdated.Class)
		assert.Equal(t, "Monk", classUpdated.Class.Name)

		// Step 4: Roll abilities and update
		rolls := []character2.AbilityRoll{
			{ID: "roll_1", Value: 16},
			{ID: "roll_2", Value: 15},
			{ID: "roll_3", Value: 14},
			{ID: "roll_4", Value: 13},
			{ID: "roll_5", Value: 12},
			{ID: "roll_6", Value: 10},
		}

		rollsUpdated, err := svc.UpdateDraftCharacter(ctx, originalID, &character.UpdateDraftInput{
			AbilityRolls: rolls,
		})
		require.NoError(t, err)
		assert.Equal(t, 6, len(rollsUpdated.AbilityRolls))

		// Step 5: Auto-assign abilities (monk priorities)
		assignments := map[string]string{
			"DEX": "roll_1", // 16
			"WIS": "roll_2", // 15
			"CON": "roll_3", // 14
			"STR": "roll_4", // 13
			"INT": "roll_5", // 12
			"CHA": "roll_6", // 10
		}

		assignUpdated, err := svc.UpdateDraftCharacter(ctx, originalID, &character.UpdateDraftInput{
			AbilityAssignments: assignments,
		})
		require.NoError(t, err)
		assert.Equal(t, 6, len(assignUpdated.Attributes))
		assert.Equal(t, 6, len(assignUpdated.AbilityAssignments))

		// Verify character still exists and has abilities
		char3, err := svc.GetOrCreateDraftCharacter(ctx, userID, realmID)
		require.NoError(t, err)
		assert.Equal(t, originalID, char3.ID, "Should still return same draft character")
		assert.Equal(t, 6, len(char3.Attributes), "Should maintain attributes")

		// Step 6: Add proficiencies
		profUpdated, err := svc.UpdateDraftCharacter(ctx, originalID, &character.UpdateDraftInput{
			Proficiencies: []string{"skill-acrobatics", "skill-insight"},
		})
		require.NoError(t, err)
		assert.Equal(t, originalID, profUpdated.ID)
		assert.Equal(t, 6, len(profUpdated.Attributes), "Should still have attributes after proficiency update")

		// Step 7: Add first equipment (shortsword)
		equip1Updated, err := svc.UpdateDraftCharacter(ctx, originalID, &character.UpdateDraftInput{
			Equipment: []string{"shortsword"},
		})
		require.NoError(t, err)
		assert.Equal(t, originalID, equip1Updated.ID)
		assert.Equal(t, 6, len(equip1Updated.Attributes), "Should still have attributes after first equipment")

		// THIS IS WHERE IT BREAKS IN PRODUCTION
		// Verify character still exists before second equipment
		char4, err := svc.GetOrCreateDraftCharacter(ctx, userID, realmID)
		require.NoError(t, err)
		assert.Equal(t, originalID, char4.ID, "Should STILL return same draft character before second equipment")

		// Step 8: Add second equipment (dungeoneers-pack) - THIS IS WHERE NEW CHARACTER GETS CREATED
		equip2Updated, err := svc.UpdateDraftCharacter(ctx, originalID, &character.UpdateDraftInput{
			Equipment: []string{"dungeoneers-pack"},
		})
		require.NoError(t, err)
		assert.Equal(t, originalID, equip2Updated.ID)
		assert.Equal(t, 6, len(equip2Updated.Attributes), "Should still have attributes after second equipment")

		// Verify character STILL exists after second equipment
		char5, err := svc.GetOrCreateDraftCharacter(ctx, userID, realmID)
		require.NoError(t, err)
		assert.Equal(t, originalID, char5.ID, "Should STILL return same draft character after second equipment")
		assert.Equal(t, character2.CharacterStatusDraft, char5.Status, "Should still be draft status")

		// Step 9: Finalize with name
		finalChar, err := svc.FinalizeCharacterWithName(ctx, originalID, "Test Monk", "elf", "monk")
		require.NoError(t, err)
		assert.Equal(t, originalID, finalChar.ID)
		assert.Equal(t, character2.CharacterStatusActive, finalChar.Status)
		assert.Equal(t, "Test Monk", finalChar.Name)
		assert.Equal(t, 6, len(finalChar.Attributes), "Should have attributes after finalization")
		assert.True(t, finalChar.IsComplete(), "Character should be complete")

		// Verify no draft exists after finalization
		char6, err := svc.GetOrCreateDraftCharacter(ctx, userID, realmID)
		require.NoError(t, err)
		assert.NotEqual(t, originalID, char6.ID, "Should create new draft after finalization")
	})
}

func stringPtr(s string) *string {
	return &s
}
