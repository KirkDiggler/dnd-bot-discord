//go:build integration
// +build integration

package character_test

import (
	"context"
	"os"
	"testing"

	mockdnd5e "github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e/mock"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	charactersRepo "github.com/KirkDiggler/dnd-bot-discord/internal/repositories/characters"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCharacterAbilityAssignment_RedisIntegration(t *testing.T) {
	// Skip if not running integration tests
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration test")
	}

	// Set up Redis client
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379/0"
	}

	opts, err := redis.ParseURL(redisURL)
	require.NoError(t, err)

	client := redis.NewClient(opts)
	defer client.Close()

	// Verify Redis is available
	ctx := context.Background()
	_, err = client.Ping(ctx).Result()
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}

	// Create repository
	repo := charactersRepo.NewRedisRepository(&charactersRepo.RedisRepoConfig{
		Client: client,
	})

	// Create mock D&D client
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mockdnd5e.NewMockClient(ctrl)

	// Set up mock expectations for the flow
	mockClient.EXPECT().GetRace("elf").Return(&entities.Race{
		Key:  "elf",
		Name: "Elf",
		AbilityBonuses: []*entities.AbilityBonus{
			{Attribute: entities.AttributeDexterity, Bonus: 2},
		},
		Speed: 30,
	}, nil).AnyTimes()

	mockClient.EXPECT().GetClass("monk").Return(&entities.Class{
		Key:    "monk",
		Name:   "Monk",
		HitDie: 8,
	}, nil).AnyTimes()

	mockClient.EXPECT().ListClassFeatures("monk", 1).Return([]*entities.CharacterFeature{
		{
			Name: "Unarmored Defense",
			Type: entities.FeatureTypeClass,
		},
	}, nil).AnyTimes()

	// Create service
	svc := character.NewService(&character.ServiceConfig{
		DNDClient:  mockClient,
		Repository: repo,
	})

	// Clean up any existing test data
	testUserID := "test_user_redis"
	testRealmID := "test_realm_redis"

	// Test the exact flow that's failing
	t.Run("AbilityAssignmentFlow", func(t *testing.T) {
		// Step 1: Create draft character
		draft, err := svc.GetOrCreateDraftCharacter(ctx, testUserID, testRealmID)
		require.NoError(t, err)
		require.NotNil(t, draft)

		// Step 2: Update with race
		raceKey := "elf"
		draft, err = svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
			RaceKey: &raceKey,
		})
		require.NoError(t, err)

		// Step 3: Update with class
		classKey := "monk"
		draft, err = svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
			ClassKey: &classKey,
		})
		require.NoError(t, err)

		// Step 4: Assign abilities
		abilityRolls := []entities.AbilityRoll{
			{ID: "roll_1", Value: 15},
			{ID: "roll_2", Value: 14},
			{ID: "roll_3", Value: 13},
			{ID: "roll_4", Value: 12},
			{ID: "roll_5", Value: 11},
			{ID: "roll_6", Value: 10},
		}

		abilityAssignments := map[string]string{
			"STR": "roll_3",
			"DEX": "roll_2",
			"CON": "roll_4",
			"INT": "roll_1",
			"WIS": "roll_5",
			"CHA": "roll_6",
		}

		draft, err = svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
			AbilityRolls:       abilityRolls,
			AbilityAssignments: abilityAssignments,
		})
		require.NoError(t, err)

		// Verify attributes are populated after assignment
		assert.NotEmpty(t, draft.Attributes, "Attributes should be populated after assignment")
		assert.Len(t, draft.Attributes, 6, "Should have all 6 ability scores")

		// Step 5: Update name
		name := "TestMonk"
		draft, err = svc.UpdateDraftCharacter(ctx, draft.ID, &character.UpdateDraftInput{
			Name: &name,
		})
		require.NoError(t, err)

		// Step 6: Finalize character
		final, err := svc.FinalizeDraftCharacter(ctx, draft.ID)
		require.NoError(t, err)

		// Verify final character has attributes
		assert.NotEmpty(t, final.Attributes, "Final character should have attributes")
		assert.Len(t, final.Attributes, 6, "Final character should have all 6 ability scores")

		// Step 7: Load character again (simulating Discord bot restart)
		loaded, err := svc.GetCharacter(ctx, final.ID)
		require.NoError(t, err)

		// THIS IS THE BUG: Loaded character should have attributes
		assert.NotEmpty(t, loaded.Attributes, "Loaded character should have attributes")
		assert.Len(t, loaded.Attributes, 6, "Loaded character should have all 6 ability scores")
		assert.True(t, loaded.IsComplete(), "Loaded character should be complete")

		// Verify specific values
		assert.Equal(t, 16, loaded.Attributes[entities.AttributeDexterity].Score, "DEX should be 14 + 2 racial")

		// Clean up
		err = repo.Delete(ctx, final.ID)
		require.NoError(t, err)
	})
}
