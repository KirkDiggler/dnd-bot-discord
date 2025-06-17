package entities_test

import (
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/stretchr/testify/assert"
)

func TestStepOrder(t *testing.T) {
	expectedOrder := []entities.CreateStep{
		entities.SelectRaceStep,
		entities.SelectClassStep,
		entities.SelectAbilityScoresStep,
		entities.SelectProficienciesStep,
		entities.SelectEquipmentStep,
		entities.SelectFeaturesStep,
		entities.EnterNameStep,
	}

	assert.Equal(t, expectedOrder, entities.StepOrder[:7], "First 7 steps should match expected order")
}

func TestAllStepsCompleted(t *testing.T) {
	tests := []struct {
		name           string
		completedSteps entities.CreateStep
		expected       bool
	}{
		{
			name:           "no steps completed",
			completedSteps: 0,
			expected:       false,
		},
		{
			name: "all implemented steps completed",
			completedSteps: entities.SelectRaceStep | entities.SelectClassStep | 
				entities.SelectAbilityScoresStep | entities.SelectProficienciesStep |
				entities.SelectEquipmentStep | entities.SelectFeaturesStep | entities.EnterNameStep,
			expected: true,
		},
		{
			name: "missing name step",
			completedSteps: entities.SelectRaceStep | entities.SelectClassStep | 
				entities.SelectAbilityScoresStep | entities.SelectProficienciesStep |
				entities.SelectEquipmentStep | entities.SelectFeaturesStep,
			expected: false,
		},
		{
			name: "includes unimplemented steps",
			completedSteps: entities.SelectRaceStep | entities.SelectClassStep | 
				entities.SelectAbilityScoresStep | entities.SelectProficienciesStep |
				entities.SelectEquipmentStep | entities.SelectFeaturesStep | 
				entities.EnterNameStep | entities.SelectBackgroundStep | entities.SelectAlignmentStep,
			expected: true, // Should still be true as we have all implemented steps
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			draft := &entities.CharacterDraft{
				CompletedSteps: tt.completedSteps,
			}
			assert.Equal(t, tt.expected, draft.AllStepsCompleted())
		})
	}
}

func TestStepDependencies(t *testing.T) {
	// Verify race affects features
	assert.Contains(t, entities.StepDependencies[entities.SelectRaceStep], entities.SelectFeaturesStep)
	
	// Verify class affects features
	assert.Contains(t, entities.StepDependencies[entities.SelectClassStep], entities.SelectFeaturesStep)
	
	// Verify background affects features
	assert.Contains(t, entities.StepDependencies[entities.SelectBackgroundStep], entities.SelectFeaturesStep)
}