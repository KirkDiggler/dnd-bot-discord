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

// TestFullCharacterCreationFlow tests the complete character creation process
func TestFullCharacterCreationFlow(t *testing.T) {
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

	t.Run("Complete wizard creation flow", func(t *testing.T) {
		userID := "test-wizard-user"
		realmID := "test-realm"

		// Track steps completed
		stepsCompleted := []string{}

		// Step 1: Create draft character
		char, err := svc.GetOrCreateDraftCharacter(ctx, userID, realmID)
		require.NoError(t, err)
		t.Logf("✅ Created character: %s", char.ID)

		// Step 2: Select race (Human)
		step, err := flowService.GetNextStep(ctx, char.ID)
		require.NoError(t, err)
		assert.Equal(t, character.StepTypeRaceSelection, step.Type)
		t.Logf("📍 Current step: %s", step.Type)

		result := &character.CreationStepResult{
			StepType:   character.StepTypeRaceSelection,
			Selections: []string{"human"},
		}
		nextStep, err := flowService.ProcessStepResult(ctx, char.ID, result)
		require.NoError(t, err)
		stepsCompleted = append(stepsCompleted, "race")
		t.Logf("✅ Selected race: human, next step: %s", nextStep.Type)

		// Step 3: Select class (Wizard)
		assert.Equal(t, character.StepTypeClassSelection, nextStep.Type)
		result = &character.CreationStepResult{
			StepType:   character.StepTypeClassSelection,
			Selections: []string{"wizard"},
		}
		nextStep, err = flowService.ProcessStepResult(ctx, char.ID, result)
		require.NoError(t, err)
		stepsCompleted = append(stepsCompleted, "class")
		t.Logf("✅ Selected class: wizard, next step: %s", nextStep.Type)

		// Step 4: Roll ability scores
		assert.Equal(t, character.StepTypeAbilityScores, nextStep.Type)
		// Simulate rolling abilities
		char, err = svc.GetCharacter(ctx, char.ID)
		require.NoError(t, err)
		char.AbilityRolls = []character.AbilityRoll{
			{ID: "0", Value: 15},
			{ID: "1", Value: 14},
			{ID: "2", Value: 13},
			{ID: "3", Value: 12},
			{ID: "4", Value: 10},
			{ID: "5", Value: 8},
		}
		err = svc.UpdateEquipment(char)
		require.NoError(t, err)

		// Process ability score step to move forward
		result = &character.CreationStepResult{
			StepType: character.StepTypeAbilityScores,
		}
		nextStep, err = flowService.ProcessStepResult(ctx, char.ID, result)
		require.NoError(t, err)
		stepsCompleted = append(stepsCompleted, "ability_scores")
		t.Logf("✅ Rolled abilities, next step: %s", nextStep.Type)

		// Step 5: Assign abilities
		assert.Equal(t, character.StepTypeAbilityAssignment, nextStep.Type)
		// Wizard prioritizes INT
		updateInput := &charService.UpdateDraftInput{
			AbilityAssignments: map[string]string{
				"STR": "5", // 8
				"DEX": "1", // 14
				"CON": "2", // 13
				"INT": "0", // 15
				"WIS": "3", // 12
				"CHA": "4", // 10
			},
		}
		_, err = svc.UpdateDraftCharacter(ctx, char.ID, updateInput)
		require.NoError(t, err)

		result = &character.CreationStepResult{
			StepType: character.StepTypeAbilityAssignment,
		}
		nextStep, err = flowService.ProcessStepResult(ctx, char.ID, result)
		require.NoError(t, err)
		stepsCompleted = append(stepsCompleted, "ability_assignment")
		t.Logf("✅ Assigned abilities, next step: %s", nextStep.Type)

		// Step 6: Select cantrips
		assert.Equal(t, character.StepTypeCantripsSelection, nextStep.Type)
		assert.NotEmpty(t, nextStep.Options, "Should have cantrip options")

		// Select 3 cantrips (wizard requirement)
		selectedCantrips := []string{
			nextStep.Options[0].Key,
			nextStep.Options[1].Key,
			nextStep.Options[2].Key,
		}
		result = &character.CreationStepResult{
			StepType:   character.StepTypeCantripsSelection,
			Selections: selectedCantrips,
		}
		nextStep, err = flowService.ProcessStepResult(ctx, char.ID, result)
		require.NoError(t, err)
		stepsCompleted = append(stepsCompleted, "cantrips")
		t.Logf("✅ Selected cantrips: %v, next step: %s", selectedCantrips, nextStep.Type)

		// Verify cantrips were saved
		char, err = svc.GetCharacter(ctx, char.ID)
		require.NoError(t, err)
		require.NotNil(t, char.Spells)
		assert.Len(t, char.Spells.Cantrips, 3)

		// Step 7: Select spells (should be spellbook selection for wizard)
		assert.Equal(t, character.StepTypeSpellbookSelection, nextStep.Type)
		assert.NotEmpty(t, nextStep.Options, "Should have spell options")

		// Select 6 spells (wizard requirement)
		selectedSpells := []string{}
		for i := 0; i < 6 && i < len(nextStep.Options); i++ {
			selectedSpells = append(selectedSpells, nextStep.Options[i].Key)
		}
		result = &character.CreationStepResult{
			StepType:   character.StepTypeSpellbookSelection,
			Selections: selectedSpells,
		}
		nextStep, err = flowService.ProcessStepResult(ctx, char.ID, result)
		require.NoError(t, err)
		stepsCompleted = append(stepsCompleted, "spells")
		t.Logf("✅ Selected spells: %v, next step: %s", selectedSpells, nextStep.Type)

		// Verify spells were saved
		char, err = svc.GetCharacter(ctx, char.ID)
		require.NoError(t, err)
		require.NotNil(t, char.Spells)
		assert.Len(t, char.Spells.KnownSpells, 6)

		// Step 8: Proficiencies (this is where you're getting stuck)
		assert.Equal(t, character.StepTypeProficiencySelection, nextStep.Type)
		t.Logf("🔍 Proficiency step: %+v", nextStep)

		// Check if there are any options to select
		if len(nextStep.Options) > 0 {
			// Select some proficiencies if available
			proficiencySelections := []string{}
			if len(nextStep.Options) > 0 {
				proficiencySelections = append(proficiencySelections, nextStep.Options[0].Key)
			}

			result = &character.CreationStepResult{
				StepType:   character.StepTypeProficiencySelection,
				Selections: proficiencySelections,
			}
			nextStep, err = flowService.ProcessStepResult(ctx, char.ID, result)
			require.NoError(t, err)
			stepsCompleted = append(stepsCompleted, "proficiencies")
			t.Logf("✅ Selected proficiencies: %v, next step: %s", proficiencySelections, nextStep.Type)
		} else {
			t.Logf("⚠️  No proficiency options available, this might be why it's stuck")
			// Try to process with empty selections
			result = &character.CreationStepResult{
				StepType:   character.StepTypeProficiencySelection,
				Selections: []string{},
			}
			nextStep, err = flowService.ProcessStepResult(ctx, char.ID, result)
			if err != nil {
				t.Logf("❌ Error processing empty proficiencies: %v", err)
			} else {
				stepsCompleted = append(stepsCompleted, "proficiencies")
				t.Logf("✅ Processed empty proficiencies, next step: %s", nextStep.Type)
			}
		}

		// Step 9: Equipment selection
		if nextStep.Type == character.StepTypeEquipmentSelection {
			t.Logf("📍 Equipment step reached")
			// For now, just process with empty selections
			result = &character.CreationStepResult{
				StepType:   character.StepTypeEquipmentSelection,
				Selections: []string{},
			}
			nextStep, err = flowService.ProcessStepResult(ctx, char.ID, result)
			require.NoError(t, err)
			stepsCompleted = append(stepsCompleted, "equipment")
			t.Logf("✅ Processed equipment, next step: %s", nextStep.Type)
		}

		// Step 10: Character details (name and finalize)
		if nextStep.Type == character.StepTypeCharacterDetails {
			t.Logf("📍 Character details step reached")

			// Set character name
			updateInput := &charService.UpdateDraftInput{
				Name: stringPtr("Gandalf the Test Wizard"),
			}
			_, err = svc.UpdateDraftCharacter(ctx, char.ID, updateInput)
			require.NoError(t, err)

			// Finalize the character
			finalizedChar, finalizeErr := svc.FinalizeDraftCharacter(ctx, char.ID)
			require.NoError(t, finalizeErr)
			stepsCompleted = append(stepsCompleted, "character_details")
			t.Logf("✅ Finalized character: %s", finalizedChar.Name)

			// Verify character is complete
			isComplete, completeErr := flowService.IsCreationComplete(ctx, char.ID)
			require.NoError(t, completeErr)
			assert.True(t, isComplete, "Character creation should be complete")
		}

		// Log summary
		t.Logf("\n📊 Character Creation Summary:")
		t.Logf("   Steps completed: %v", stepsCompleted)
		t.Logf("   Final step reached: %s", nextStep.Type)

		// Get final character state
		finalChar, err := svc.GetCharacter(ctx, char.ID)
		require.NoError(t, err)
		t.Logf("   Character: %s (Level %d %s %s)",
			finalChar.Name, finalChar.Level, finalChar.Race.Name, finalChar.Class.Name)
		if finalChar.Spells != nil {
			t.Logf("   Cantrips: %d, Spells: %d",
				len(finalChar.Spells.Cantrips), len(finalChar.Spells.KnownSpells))
		}
	})

	// Additional test for other classes
	t.Run("Fighter creation flow - no spell steps", func(t *testing.T) {
		userID := "test-fighter-user"
		realmID := "test-realm"

		// Create character and select fighter
		char, err := svc.GetOrCreateDraftCharacter(ctx, userID, realmID)
		require.NoError(t, err)

		// Race
		_, err = flowService.ProcessStepResult(ctx, char.ID, &character.CreationStepResult{
			StepType:   character.StepTypeRaceSelection,
			Selections: []string{"human"},
		})
		require.NoError(t, err)

		// Class - Fighter
		_, err = flowService.ProcessStepResult(ctx, char.ID, &character.CreationStepResult{
			StepType:   character.StepTypeClassSelection,
			Selections: []string{"fighter"},
		})
		require.NoError(t, err)

		// Abilities
		char, err = svc.GetCharacter(ctx, char.ID)
		require.NoError(t, err)
		char.AbilityRolls = []character.AbilityRoll{
			{ID: "0", Value: 15},
			{ID: "1", Value: 14},
			{ID: "2", Value: 13},
			{ID: "3", Value: 12},
			{ID: "4", Value: 10},
			{ID: "5", Value: 8},
		}
		err = svc.UpdateEquipment(char)
		require.NoError(t, err)

		_, err = flowService.ProcessStepResult(ctx, char.ID, &character.CreationStepResult{
			StepType: character.StepTypeAbilityScores,
		})
		require.NoError(t, err)

		// Assign abilities (fighter prioritizes STR)
		_, err = svc.UpdateDraftCharacter(ctx, char.ID, &charService.UpdateDraftInput{
			AbilityAssignments: map[string]string{
				"STR": "0", // 15
				"DEX": "2", // 13
				"CON": "1", // 14
				"INT": "4", // 10
				"WIS": "3", // 12
				"CHA": "5", // 8
			},
		})
		require.NoError(t, err)

		nextStep, err := flowService.ProcessStepResult(ctx, char.ID, &character.CreationStepResult{
			StepType: character.StepTypeAbilityAssignment,
		})
		require.NoError(t, err)

		// Fighter should have fighting style selection
		if nextStep.Type == character.StepTypeFightingStyleSelection {
			t.Logf("✅ Fighter has fighting style selection step")
			// Select a fighting style if available
			if len(nextStep.Options) > 0 {
				_, err = flowService.ProcessStepResult(ctx, char.ID, &character.CreationStepResult{
					StepType:   character.StepTypeFightingStyleSelection,
					Selections: []string{nextStep.Options[0].Key},
				})
				require.NoError(t, err)
			}
		}

		// Should NOT have spell steps
		allSteps, err := flowService.GetProgressSteps(ctx, char.ID)
		require.NoError(t, err)

		hasSpellSteps := false
		for _, step := range allSteps {
			if step.Step.Type == character.StepTypeCantripsSelection ||
				step.Step.Type == character.StepTypeSpellbookSelection ||
				step.Step.Type == character.StepTypeSpellSelection {
				hasSpellSteps = true
				break
			}
		}
		assert.False(t, hasSpellSteps, "Fighter should not have spell selection steps")
	})
}
