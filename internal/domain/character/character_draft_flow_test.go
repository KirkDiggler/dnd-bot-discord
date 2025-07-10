package character_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
)

func TestCharacterDraft_WithFlowState(t *testing.T) {
	t.Run("draft tracks flow state separately from character", func(t *testing.T) {
		// Create a draft with flow state
		draft := &character.CharacterDraft{
			ID:        "draft-123",
			OwnerID:   "user-123",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Character: &character.Character{
				ID:      "char-123",
				OwnerID: "user-123",
				Name:    "Test Character",
			},
			FlowState: &character.FlowState{
				CurrentStepID:  "race",
				AllSteps:       []string{"race", "class", "abilities", "equipment", "name"},
				CompletedSteps: []string{},
				StepData:       map[string]any{},
				LastUpdated:    time.Now(),
			},
		}

		// Character is pure domain object without flow concerns
		assert.Equal(t, "Test Character", draft.Character.Name)

		// Draft should have flow state
		assert.NotNil(t, draft.FlowState)
		assert.Equal(t, "race", draft.FlowState.CurrentStepID)
	})

	t.Run("flow state methods work correctly", func(t *testing.T) {
		flowState := &character.FlowState{
			CurrentStepID:  "class",
			AllSteps:       []string{"race", "class", "abilities", "equipment", "name"},
			CompletedSteps: []string{"race"},
		}

		// Test IsStepCompleted
		assert.True(t, flowState.IsStepCompleted("race"))
		assert.False(t, flowState.IsStepCompleted("class"))
		assert.False(t, flowState.IsStepCompleted("abilities"))

		// Test GetStepIndex
		assert.Equal(t, 0, flowState.GetStepIndex("race"))
		assert.Equal(t, 1, flowState.GetStepIndex("class"))
		assert.Equal(t, -1, flowState.GetStepIndex("invalid"))

		// Test GetCurrentStepIndex
		assert.Equal(t, 1, flowState.GetCurrentStepIndex())

		// Test CanNavigateBack
		assert.True(t, flowState.CanNavigateBack())

		// Test CanNavigateForward
		assert.True(t, flowState.CanNavigateForward())
	})

	t.Run("navigation edge cases", func(t *testing.T) {
		// First step - can't go back
		flowState := &character.FlowState{
			CurrentStepID:  "race",
			AllSteps:       []string{"race", "class", "abilities"},
			CompletedSteps: []string{},
		}
		assert.False(t, flowState.CanNavigateBack())
		assert.True(t, flowState.CanNavigateForward())

		// Last step - can't go forward
		flowState.CurrentStepID = "abilities"
		assert.True(t, flowState.CanNavigateBack())
		assert.False(t, flowState.CanNavigateForward())

		// Invalid current step
		flowState.CurrentStepID = "invalid"
		assert.False(t, flowState.CanNavigateBack())
		assert.False(t, flowState.CanNavigateForward())
	})

	t.Run("step data storage", func(t *testing.T) {
		flowState := &character.FlowState{
			CurrentStepID:  "abilities",
			AllSteps:       []string{"race", "class", "abilities"},
			CompletedSteps: []string{"race", "class"},
			StepData: map[string]any{
				"race": map[string]string{
					"selected": "human",
				},
				"class": map[string]string{
					"selected": "wizard",
				},
				"abilities": map[string]int{
					"STR": 15,
					"DEX": 14,
					"CON": 13,
					"INT": 12,
					"WIS": 10,
					"CHA": 8,
				},
			},
		}

		// Verify we can store and retrieve step data
		raceData, ok := flowState.StepData["race"].(map[string]string)
		require.True(t, ok, "race data should be a map[string]string")
		assert.Equal(t, "human", raceData["selected"])

		classData, ok := flowState.StepData["class"].(map[string]string)
		require.True(t, ok, "class data should be a map[string]string")
		assert.Equal(t, "wizard", classData["selected"])

		abilityData, ok := flowState.StepData["abilities"].(map[string]int)
		require.True(t, ok, "ability data should be a map[string]int")
		assert.Equal(t, 15, abilityData["STR"])
	})
}

func TestCharacterDraft_LegacyStepTracking(t *testing.T) {
	t.Run("legacy bitwise steps still work", func(t *testing.T) {
		draft := &character.CharacterDraft{
			ID:             "draft-123",
			CurrentStep:    character.SelectClassStep,
			CompletedSteps: character.SelectRaceStep,
		}

		// Legacy methods should still work
		assert.True(t, draft.IsStepCompleted(character.SelectRaceStep))
		assert.False(t, draft.IsStepCompleted(character.SelectClassStep))

		// Complete a step
		err := draft.CompleteStep(character.SelectClassStep)
		require.NoError(t, err)
		assert.True(t, draft.IsStepCompleted(character.SelectClassStep))
	})

	t.Run("can use both legacy and new flow state", func(t *testing.T) {
		draft := &character.CharacterDraft{
			ID:             "draft-123",
			CurrentStep:    character.SelectClassStep,
			CompletedSteps: character.SelectRaceStep,
			FlowState: &character.FlowState{
				CurrentStepID:  "class",
				AllSteps:       []string{"race", "class", "abilities"},
				CompletedSteps: []string{"race"},
			},
		}

		// Both tracking systems work independently
		assert.True(t, draft.IsStepCompleted(character.SelectRaceStep))
		assert.True(t, draft.FlowState.IsStepCompleted("race"))
	})
}

func TestCharacterDraft_Workflow(t *testing.T) {
	t.Run("typical creation workflow", func(t *testing.T) {
		// Start with empty draft
		draft := &character.CharacterDraft{
			ID:        "draft-123",
			OwnerID:   "user-123",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Character: &character.Character{
				ID:      "char-123",
				OwnerID: "user-123",
				RealmID: "realm-123",
			},
			FlowState: &character.FlowState{
				CurrentStepID:  "race",
				AllSteps:       []string{"race", "class", "abilities", "equipment", "name"},
				CompletedSteps: []string{},
				StepData:       map[string]any{},
			},
		}

		// Step 1: Select race
		draft.Character.Race = &rulebook.Race{Key: "human", Name: "Human"}
		draft.FlowState.CompletedSteps = append(draft.FlowState.CompletedSteps, "race")
		draft.FlowState.CurrentStepID = "class"
		draft.FlowState.StepData["race"] = map[string]string{"selected": "human"}

		// Step 2: Select class (wizard)
		draft.Character.Class = &rulebook.Class{Key: "wizard", Name: "Wizard"}
		draft.FlowState.CompletedSteps = append(draft.FlowState.CompletedSteps, "class")
		draft.FlowState.CurrentStepID = "abilities"
		draft.FlowState.StepData["class"] = map[string]string{"selected": "wizard"}

		// Now flow should inject wizard-specific steps
		wizardSteps := []string{"race", "class", "abilities", "cantrips", "spells", "equipment", "name"}
		draft.FlowState.AllSteps = wizardSteps

		// Verify state
		assert.Equal(t, 2, len(draft.FlowState.CompletedSteps))
		assert.Equal(t, "abilities", draft.FlowState.CurrentStepID)
		assert.Contains(t, draft.FlowState.AllSteps, "cantrips")
		assert.Contains(t, draft.FlowState.AllSteps, "spells")
	})

	t.Run("convert draft to character on completion", func(t *testing.T) {
		// Completed draft
		draft := &character.CharacterDraft{
			ID:      "draft-123",
			OwnerID: "user-123",
			Character: &character.Character{
				ID:      "char-123",
				OwnerID: "user-123",
				Name:    "Gandalf",
				Race:    &rulebook.Race{Key: "human", Name: "Human"},
				Class:   &rulebook.Class{Key: "wizard", Name: "Wizard"},
				// ... other fields populated
			},
			FlowState: &character.FlowState{
				CurrentStepID:  "complete",
				AllSteps:       []string{"race", "class", "abilities", "cantrips", "spells", "equipment", "name"},
				CompletedSteps: []string{"race", "class", "abilities", "cantrips", "spells", "equipment", "name"},
			},
		}

		// Extract the character (this would be done by service)
		finalCharacter := draft.Character

		// Character should have all data but no flow state
		assert.NotNil(t, finalCharacter)
		assert.Equal(t, "Gandalf", finalCharacter.Name)
		assert.Equal(t, "wizard", finalCharacter.Class.Key)

		// In real implementation, the service would:
		// 1. Set character status to "active"
		// 2. Save character to character repository
		// 3. Delete draft from draft repository
	})
}
