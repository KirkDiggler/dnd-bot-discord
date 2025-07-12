//go:build integration
// +build integration

package character_test

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e/calculators"
	characterdraft "github.com/KirkDiggler/dnd-bot-discord/internal/repositories/character_draft"
	charactersRepo "github.com/KirkDiggler/dnd-bot-discord/internal/repositories/characters"
	charService "github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

// TestSpellSelectionRegression prevents regressions in spell selection
func TestSpellSelectionRegression(t *testing.T) {
	// Set up Redis client
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379/15" // Use test database
	}

	redisOpts, err := redis.ParseURL(redisURL)
	require.NoError(t, err)
	client := redis.NewClient(redisOpts)

	// Test Redis connection
	ctx := context.Background()
	err = client.Ping(ctx).Err()
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}

	// Clean up test data
	defer func() { _ = client.FlushDB(ctx) }()

	// Create repositories
	charRepo := charactersRepo.NewRedisRepository(&charactersRepo.RedisRepoConfig{
		Client: client,
	})

	draftRepo := characterdraft.NewInMemoryRepository()

	// Create real D&D API client
	dndClient, err := dnd5e.New(&dnd5e.Config{
		HttpClient: &http.Client{},
	})
	require.NoError(t, err)

	// Create AC calculator
	acCalculator := calculators.NewDnD5eACCalculator()

	// Create flow builder and services
	flowBuilder := charService.NewFlowBuilder(dndClient)

	// Create service
	svc := charService.NewService(&charService.ServiceConfig{
		DNDClient:       dndClient,
		Repository:      charRepo,
		DraftRepository: draftRepo,
		ACCalculator:    acCalculator,
	})

	flowService := charService.NewCreationFlowService(svc, flowBuilder)

	t.Run("spell_selection_persists_across_pages", func(t *testing.T) {
		// Create a wizard character
		char, err := svc.GetOrCreateDraftCharacter(ctx, "test-user", "test-realm")
		require.NoError(t, err)

		// Set race and class
		_, err = svc.UpdateDraftCharacter(ctx, char.ID, &charService.UpdateDraftInput{
			RaceKey:  stringPtr("human"),
			ClassKey: stringPtr("wizard"),
		})
		require.NoError(t, err)

		// Set ability scores
		char, err = svc.GetCharacter(ctx, char.ID)
		require.NoError(t, err)

		char.AbilityRolls = []character.AbilityRoll{
			{Value: 15}, {Value: 14}, {Value: 13},
			{Value: 12}, {Value: 10}, {Value: 8},
		}
		char.AbilityAssignments = map[string]string{
			"STR": "0", "DEX": "1", "CON": "2",
			"INT": "3", "WIS": "4", "CHA": "5",
		}
		err = svc.UpdateEquipment(char)
		require.NoError(t, err)

		// Process cantrip selection
		result := &character.CreationStepResult{
			StepType:   character.StepTypeCantripsSelection,
			Selections: []string{"fire-bolt", "mage-hand", "prestidigitation"},
		}

		_, err = flowService.ProcessStepResult(ctx, char.ID, result)
		require.NoError(t, err)

		// Verify cantrips persist after reload
		reloaded, err := svc.GetCharacter(ctx, char.ID)
		require.NoError(t, err)
		require.NotNil(t, reloaded.Spells)
		require.Equal(t, 3, len(reloaded.Spells.Cantrips))

		// Test that the flow advances to spell selection
		nextStep, err := flowService.GetNextStep(ctx, char.ID)
		require.NoError(t, err)
		require.Equal(t, character.StepTypeSpellbookSelection, nextStep.Type)
	})

	t.Run("cantrips_marked_as_complete", func(t *testing.T) {
		// Create another wizard
		char, err := svc.GetOrCreateDraftCharacter(ctx, "test-user-2", "test-realm")
		require.NoError(t, err)

		_, err = svc.UpdateDraftCharacter(ctx, char.ID, &charService.UpdateDraftInput{
			RaceKey:  stringPtr("elf"),
			ClassKey: stringPtr("wizard"),
		})
		require.NoError(t, err)

		// Set abilities
		char, err = svc.GetCharacter(ctx, char.ID)
		require.NoError(t, err)
		char.AbilityRolls = []character.AbilityRoll{
			{Value: 15}, {Value: 14}, {Value: 13},
			{Value: 12}, {Value: 10}, {Value: 8},
		}
		char.AbilityAssignments = map[string]string{
			"STR": "0", "DEX": "1", "CON": "2",
			"INT": "3", "WIS": "4", "CHA": "5",
		}
		err = svc.UpdateEquipment(char)
		require.NoError(t, err)

		// Get flow before cantrip selection
		flow, err := flowBuilder.BuildFlow(ctx, char)
		require.NoError(t, err)

		// Find cantrip step
		var cantripStep *character.CreationStep
		for _, step := range flow.Steps {
			if step.Type == character.StepTypeCantripsSelection {
				cantripStep = &step
				break
			}
		}
		require.NotNil(t, cantripStep)

		// Select exactly 3 cantrips (wizard requirement)
		result := &character.CreationStepResult{
			StepType:   character.StepTypeCantripsSelection,
			Selections: []string{"fire-bolt", "mage-hand", "ray-of-frost"},
		}

		nextStep, err := flowService.ProcessStepResult(ctx, char.ID, result)
		require.NoError(t, err)

		// Should advance past cantrips
		require.NotEqual(t, character.StepTypeCantripsSelection, nextStep.Type)

		// Double-check the character has the cantrips saved
		final, err := svc.GetCharacter(ctx, char.ID)
		require.NoError(t, err)
		require.NotNil(t, final.Spells)
		require.Equal(t, 3, len(final.Spells.Cantrips))
	})
}
