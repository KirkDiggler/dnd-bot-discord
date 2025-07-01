package character_test

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStepOrder(t *testing.T) {
	expectedOrder := []character.CreateStep{
		character.SelectRaceStep,
		character.SelectClassStep,
		character.SelectAbilityScoresStep,
		character.SelectProficienciesStep,
		character.SelectEquipmentStep,
		character.SelectFeaturesStep,
		character.EnterNameStep,
	}

	assert.Equal(t, expectedOrder, character.StepOrder[:7], "First 7 steps should match expected order")
}

func TestAllStepsCompleted(t *testing.T) {
	tests := []struct {
		name           string
		completedSteps character.CreateStep
		expected       bool
	}{
		{
			name:           "no steps completed",
			completedSteps: 0,
			expected:       false,
		},
		{
			name: "all implemented steps completed",
			completedSteps: character.SelectRaceStep | character.SelectClassStep |
				character.SelectAbilityScoresStep | character.SelectProficienciesStep |
				character.SelectEquipmentStep | character.SelectFeaturesStep | character.EnterNameStep,
			expected: true,
		},
		{
			name: "missing name step",
			completedSteps: character.SelectRaceStep | character.SelectClassStep |
				character.SelectAbilityScoresStep | character.SelectProficienciesStep |
				character.SelectEquipmentStep | character.SelectFeaturesStep,
			expected: false,
		},
		{
			name: "includes unimplemented steps",
			completedSteps: character.SelectRaceStep | character.SelectClassStep |
				character.SelectAbilityScoresStep | character.SelectProficienciesStep |
				character.SelectEquipmentStep | character.SelectFeaturesStep |
				character.EnterNameStep | character.SelectBackgroundStep | character.SelectAlignmentStep,
			expected: true, // Should still be true as we have all implemented steps
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			draft := &character.CharacterDraft{
				CompletedSteps: tt.completedSteps,
			}
			assert.Equal(t, tt.expected, draft.AllStepsCompleted())
		})
	}
}

func TestStepDependencies(t *testing.T) {
	// Verify race affects features
	assert.Contains(t, character.StepDependencies[character.SelectRaceStep], character.SelectFeaturesStep)

	// Verify class affects features
	assert.Contains(t, character.StepDependencies[character.SelectClassStep], character.SelectFeaturesStep)

	// Verify background affects features
	assert.Contains(t, character.StepDependencies[character.SelectBackgroundStep], character.SelectFeaturesStep)
}
