package character

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCharacterDraft_ResetStep(t *testing.T) {
	tests := []struct {
		name          string
		draft         *CharacterDraft
		stepToReset   CreateStep
		expectedError string
		verify        func(*testing.T, *CharacterDraft)
	}{
		{
			name: "reset uncompleted step returns error",
			draft: &CharacterDraft{
				CompletedSteps: SelectRaceStep | SelectClassStep,
				Character:      &Character{},
			},
			stepToReset:   SelectBackgroundStep,
			expectedError: "step 128 is not completed",
		},
		{
			name: "reset race step clears race and dependent steps",
			draft: &CharacterDraft{
				CompletedSteps: SelectRaceStep | SelectClassStep | SelectProficienciesStep | SelectAbilityScoresStep,
				Character: &Character{
					Race:  &rulebook.Race{Name: "Elf"},
					Class: &rulebook.Class{Name: "Fighter"},
				},
			},
			stepToReset: SelectRaceStep,
			verify: func(t *testing.T, d *CharacterDraft) {
				assert.Nil(t, d.Character.Race)
				assert.NotNil(t, d.Character.Class) // Class should remain
				// Verify dependent steps are uncompleted
				assert.False(t, d.IsStepCompleted(SelectRaceStep))
				assert.True(t, d.IsStepCompleted(SelectClassStep))
				assert.False(t, d.IsStepCompleted(SelectProficienciesStep))
				assert.False(t, d.IsStepCompleted(SelectAbilityScoresStep))
			},
		},
		{
			name: "reset class step clears class and dependent steps",
			draft: &CharacterDraft{
				CompletedSteps: SelectRaceStep | SelectClassStep | SelectProficienciesStep | SelectSkillsStep | SelectEquipmentStep,
				Character: &Character{
					Race:  &rulebook.Race{Name: "Elf"},
					Class: &rulebook.Class{Name: "Fighter"},
				},
			},
			stepToReset: SelectClassStep,
			verify: func(t *testing.T, d *CharacterDraft) {
				assert.NotNil(t, d.Character.Race) // Race should remain
				assert.Nil(t, d.Character.Class)
				// Verify dependent steps are uncompleted
				assert.True(t, d.IsStepCompleted(SelectRaceStep))
				assert.False(t, d.IsStepCompleted(SelectClassStep))
				assert.False(t, d.IsStepCompleted(SelectProficienciesStep))
				assert.False(t, d.IsStepCompleted(SelectSkillsStep))
				assert.False(t, d.IsStepCompleted(SelectEquipmentStep))
			},
		},
		{
			name: "reset background step maintains race and class but clears downstream data",
			draft: &CharacterDraft{
				CompletedSteps: SelectRaceStep | SelectClassStep | SelectBackgroundStep | SelectProficienciesStep | SelectSkillsStep,
				Character: &Character{
					Race:  &rulebook.Race{Name: "Elf"},
					Class: &rulebook.Class{Name: "Fighter"},
					Background: &rulebook.Background{
						Name: "Soldier",
						Feature: &rulebook.Feature{
							Name: "Military Rank",
						},
					},
					Proficiencies: map[rulebook.ProficiencyType][]*rulebook.Proficiency{
						"weapons": {{Name: "Longsword"}},
					},
				},
			},
			stepToReset: SelectBackgroundStep,
			verify: func(t *testing.T, d *CharacterDraft) {
				assert.NotNil(t, d.Character.Race)
				assert.NotNil(t, d.Character.Class)
				assert.Nil(t, d.Character.Background)
				assert.Empty(t, d.Character.Proficiencies)
				// Verify dependent steps are uncompleted
				assert.True(t, d.IsStepCompleted(SelectRaceStep))
				assert.True(t, d.IsStepCompleted(SelectClassStep))
				assert.False(t, d.IsStepCompleted(SelectBackgroundStep))
				assert.False(t, d.IsStepCompleted(SelectProficienciesStep))
				assert.False(t, d.IsStepCompleted(SelectSkillsStep))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.draft.ResetStep(tt.stepToReset)

			if tt.expectedError != "" {
				assert.EqualError(t, err, tt.expectedError)
				return
			}

			assert.NoError(t, err)
			if tt.verify != nil {
				tt.verify(t, tt.draft)
			}
		})
	}
}

func TestCharacterDraft_StepDependencies(t *testing.T) {
	t.Run("verify step dependencies are properly defined", func(t *testing.T) {
		// Race dependencies
		raceDeps := StepDependencies[SelectRaceStep]
		assert.Contains(t, raceDeps, SelectProficienciesStep)
		assert.Contains(t, raceDeps, SelectAbilityScoresStep)
		assert.Contains(t, raceDeps, SelectFeaturesStep)
		assert.Len(t, raceDeps, 3)

		// Class dependencies
		classDeps := StepDependencies[SelectClassStep]
		assert.Contains(t, classDeps, SelectProficienciesStep)
		assert.Contains(t, classDeps, SelectSkillsStep)
		assert.Contains(t, classDeps, SelectEquipmentStep)
		assert.Contains(t, classDeps, SelectFeaturesStep)
		assert.Len(t, classDeps, 4)

		// Background dependencies
		backgroundDeps := StepDependencies[SelectBackgroundStep]
		assert.Contains(t, backgroundDeps, SelectProficienciesStep)
		assert.Contains(t, backgroundDeps, SelectSkillsStep)
		assert.Contains(t, backgroundDeps, SelectEquipmentStep)
		assert.Contains(t, backgroundDeps, SelectFeaturesStep)
		assert.Len(t, backgroundDeps, 4)
	})
}
