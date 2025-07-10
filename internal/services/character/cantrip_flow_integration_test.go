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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWizardCantripFlowIntegration tests the full flow for wizard cantrip selection
func TestWizardCantripFlowIntegration(t *testing.T) {
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

	t.Run("Wizard cantrip selection flow", func(t *testing.T) {
		userID := "test-user"
		realmID := "test-realm"

		// STEP 1: Create a draft character
		char, err := svc.GetOrCreateDraftCharacter(ctx, userID, realmID)
		require.NoError(t, err)
		t.Logf("✅ Created character: %s", char.ID)

		// STEP 2: Set race and class to get to cantrips
		_, err = svc.UpdateDraftCharacter(ctx, char.ID, &charService.UpdateDraftInput{
			RaceKey:  stringPtr("human"),
			ClassKey: stringPtr("wizard"),
		})
		require.NoError(t, err)
		t.Logf("✅ Set race and class")

		// STEP 3: Get the updated character
		char, err = svc.GetCharacter(ctx, char.ID)
		require.NoError(t, err)

		flow, err := flowBuilder.BuildFlow(ctx, char)
		require.NoError(t, err)
		t.Logf("📋 Flow has %d steps", len(flow.Steps))

		// Find the cantrips step
		var cantripStep *character.CreationStep
		for i, step := range flow.Steps {
			t.Logf("   Step %d: %s", i+1, step.Type)
			if step.Type == character.StepTypeCantripsSelection {
				cantripStep = &step
				t.Logf("🎯 Found cantrips step at position %d", i+1)
				break
			}
		}

		require.NotNil(t, cantripStep, "Should have a cantrips selection step for wizard")

		// STEP 4: Verify cantrips step has options from API
		assert.NotEmpty(t, cantripStep.Options, "Cantrips step should have options populated from D&D API")
		t.Logf("📚 Cantrips step has %d options", len(cantripStep.Options))

		if len(cantripStep.Options) > 0 {
			t.Logf("   First few cantrips: %s, %s",
				cantripStep.Options[0].Name,
				cantripStep.Options[min(1, len(cantripStep.Options)-1)].Name)
		}

		// STEP 5: Verify wizard needs 3 cantrips
		assert.Equal(t, 3, cantripStep.MinChoices, "Wizard should need 3 cantrips")
		assert.Equal(t, 3, cantripStep.MaxChoices, "Wizard should choose exactly 3 cantrips")

		// STEP 6: Select 3 cantrips
		require.GreaterOrEqual(t, len(cantripStep.Options), 3, "Need at least 3 cantrip options")

		selectedCantrips := []string{
			cantripStep.Options[0].Key,
			cantripStep.Options[1].Key,
			cantripStep.Options[2].Key,
		}

		result := &character.CreationStepResult{
			StepType:   character.StepTypeCantripsSelection,
			Selections: selectedCantrips,
		}

		// STEP 7: Process the cantrip selection
		t.Logf("🎯 About to process cantrip selection: %v", selectedCantrips)
		nextStep, err := flowService.ProcessStepResult(ctx, char.ID, result)
		require.NoError(t, err)
		t.Logf("✅ Processed cantrip selection, next step: %s", nextStep.Type)

		// STEP 8: Verify cantrips were saved to character
		char, err = svc.GetCharacter(ctx, char.ID)
		require.NoError(t, err)

		t.Logf("🔍 Character spells after processing: %+v", char.Spells)
		if char.Spells != nil {
			t.Logf("🔍 Character cantrips: %v", char.Spells.Cantrips)
		}

		require.NotNil(t, char.Spells, "Character should have spells after cantrip selection")
		assert.Len(t, char.Spells.Cantrips, 3, "Character should have exactly 3 cantrips")

		t.Logf("🎉 Character cantrips: %v", char.Spells.Cantrips)

		// Verify the exact cantrips we selected
		for _, expected := range selectedCantrips {
			assert.Contains(t, char.Spells.Cantrips, expected, "Selected cantrip should be saved")
		}

		// STEP 9: Verify we progressed past cantrips by checking current step

		// STEP 10: Verify we can get the next step without errors
		currentStep, err := flowService.GetCurrentStep(ctx, char.ID)
		require.NoError(t, err)
		assert.NotEqual(t, character.StepTypeCantripsSelection, currentStep.Type, "Should have moved past cantrips step")
		t.Logf("🚀 Successfully progressed to: %s", currentStep.Type)
	})
}

// Test other spellcaster classes too
func TestAllSpellcasterCantripFlows(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping comprehensive test in short mode")
	}

	spellcasters := map[string]int{
		"bard":     2, // 2 cantrips at level 1
		"druid":    2, // 2 cantrips at level 1
		"sorcerer": 4, // 4 cantrips at level 1
		"warlock":  2, // 2 cantrips at level 1
		"wizard":   3, // 3 cantrips at level 1
	}

	// Set up Redis and services (same as above)
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379/15"
	}

	redisOpts, err := redis.ParseURL(redisURL)
	require.NoError(t, err)
	client := redis.NewClient(redisOpts)

	ctx := context.Background()
	err = client.Ping(ctx).Err()
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}

	defer func() { _ = client.FlushDB(ctx) }()

	charRepo := charactersRepo.NewRedisRepository(&charactersRepo.RedisRepoConfig{
		Client: client,
	})

	draftRepo := characterdraft.NewInMemoryRepository()

	dndClient, err := dnd5e.New(&dnd5e.Config{
		HttpClient: &http.Client{},
	})
	require.NoError(t, err)

	acCalculator := calculators.NewDnD5eACCalculator()

	flowBuilder := charService.NewFlowBuilder(dndClient)

	svc := charService.NewService(&charService.ServiceConfig{
		DNDClient:       dndClient,
		Repository:      charRepo,
		DraftRepository: draftRepo,
		ACCalculator:    acCalculator,
	})

	for className, expectedCantrips := range spellcasters {
		t.Run(className, func(t *testing.T) {
			// Create character with this class
			userID := "test-user-" + className
			realmID := "test-realm"

			char, err := svc.GetOrCreateDraftCharacter(ctx, userID, realmID)
			require.NoError(t, err)

			_, err = svc.UpdateDraftCharacter(ctx, char.ID, &charService.UpdateDraftInput{
				RaceKey:  stringPtr("human"),
				ClassKey: stringPtr(className),
			})
			require.NoError(t, err)

			// Get flow and check cantrips
			char, err = svc.GetCharacter(ctx, char.ID)
			require.NoError(t, err)

			flow, err := flowBuilder.BuildFlow(ctx, char)
			require.NoError(t, err)

			// Find cantrips step
			var cantripStep *character.CreationStep
			for _, step := range flow.Steps {
				if step.Type == character.StepTypeCantripsSelection {
					cantripStep = &step
					break
				}
			}

			require.NotNil(t, cantripStep, "%s should have cantrips step", className)
			assert.Equal(t, expectedCantrips, cantripStep.MinChoices, "%s should need %d cantrips", className, expectedCantrips)
			assert.NotEmpty(t, cantripStep.Options, "%s cantrips should have options", className)

			t.Logf("✅ %s: %d cantrips required, %d options available",
				className, expectedCantrips, len(cantripStep.Options))
		})
	}
}
