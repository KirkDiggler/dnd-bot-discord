package character_test

import (
	"context"
	"testing"

	mockdnd5e "github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e/mock"
	character2 "github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestAllClassesCanCreateFlow(t *testing.T) {
	// Test that all D&D 5e classes can generate a creation flow
	classes := []string{
		"barbarian",
		"bard",
		"cleric",
		"druid",
		"fighter",
		"monk",
		"paladin",
		"ranger",
		"rogue",
		"sorcerer",
		"warlock",
		"wizard",
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mockdnd5e.NewMockClient(ctrl)
	// Set up any needed expectations here - for this test we just need it to not panic
	mockClient.EXPECT().ListSpellsByClassAndLevel(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()

	flowBuilder := character.NewFlowBuilder(mockClient)
	ctx := context.Background()

	for _, className := range classes {
		t.Run(className, func(t *testing.T) {
			// Create a character with the class
			char := &character2.Character{
				ID:    "test-char",
				Race:  &rulebook.Race{Key: "human", Name: "Human"},
				Class: &rulebook.Class{Key: className, Name: className},
				Level: 1,
			}

			// Build the flow
			flow, err := flowBuilder.BuildFlow(ctx, char)
			require.NoError(t, err)
			require.NotNil(t, flow)

			// Verify basic flow properties
			assert.NotEmpty(t, flow.Steps, "Flow should have steps")

			// Find class-specific steps
			var hasClassStep bool
			var classStepTypes []string

			for _, step := range flow.Steps {
				switch step.Type {
				case character2.StepTypeCantripsSelection,
					character2.StepTypeSpellSelection,
					character2.StepTypeSpellbookSelection,
					character2.StepTypeSpellsKnownSelection,
					character2.StepTypeExpertiseSelection,
					character2.StepTypeFightingStyleSelection,
					character2.StepTypeDivineDomainSelection,
					character2.StepTypeFavoredEnemySelection,
					character2.StepTypeNaturalExplorerSelection,
					character2.StepTypeSubclassSelection:
					hasClassStep = true
					classStepTypes = append(classStepTypes, string(step.Type))
				}
			}

			// Log what steps each class has
			t.Logf("%s has steps: %v", className, classStepTypes)

			// Verify expected steps for each class
			switch className {
			case "bard":
				assert.True(t, hasClassStep, "Bard should have class-specific steps")
				assert.Contains(t, classStepTypes, "expertise_selection")
				assert.Contains(t, classStepTypes, "cantrips_selection")
				assert.Contains(t, classStepTypes, "spells_known_selection")
			case "cleric":
				assert.True(t, hasClassStep, "Cleric should have class-specific steps")
				assert.Contains(t, classStepTypes, "divine_domain_selection")
			case "druid":
				assert.True(t, hasClassStep, "Druid should have class-specific steps")
				assert.Contains(t, classStepTypes, "cantrips_selection")
			case "fighter":
				assert.True(t, hasClassStep, "Fighter should have class-specific steps")
				assert.Contains(t, classStepTypes, "fighting_style_selection")
			case "ranger":
				assert.True(t, hasClassStep, "Ranger should have class-specific steps")
				assert.Contains(t, classStepTypes, "favored_enemy_selection")
				assert.Contains(t, classStepTypes, "natural_explorer_selection")
			case "rogue":
				assert.True(t, hasClassStep, "Rogue should have class-specific steps")
				assert.Contains(t, classStepTypes, "expertise_selection")
			case "sorcerer":
				assert.True(t, hasClassStep, "Sorcerer should have class-specific steps")
				assert.Contains(t, classStepTypes, "subclass_selection")
				assert.Contains(t, classStepTypes, "cantrips_selection")
				assert.Contains(t, classStepTypes, "spells_known_selection")
			case "warlock":
				assert.True(t, hasClassStep, "Warlock should have class-specific steps")
				assert.Contains(t, classStepTypes, "subclass_selection")
				assert.Contains(t, classStepTypes, "cantrips_selection")
				assert.Contains(t, classStepTypes, "spells_known_selection")
			case "wizard":
				assert.True(t, hasClassStep, "Wizard should have class-specific steps")
				assert.Contains(t, classStepTypes, "cantrips_selection")
				assert.Contains(t, classStepTypes, "spellbook_selection")
			case "barbarian", "monk", "paladin":
				// These classes have no special level 1 choices
				assert.False(t, hasClassStep, "%s should not have class-specific steps at level 1", className)
			}
		})
	}
}

func TestSpellcasterCantripsCount(t *testing.T) {
	// Verify each spellcaster gets the correct number of cantrips
	testCases := []struct {
		class            string
		expectedCantrips int
	}{
		{"bard", 2},
		{"cleric", 3}, // Clerics get 3 cantrips
		{"druid", 2},
		{"sorcerer", 4},
		{"warlock", 2},
		{"wizard", 3},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mockdnd5e.NewMockClient(ctrl)
	mockClient.EXPECT().ListSpellsByClassAndLevel(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()

	flowBuilder := character.NewFlowBuilder(mockClient)
	ctx := context.Background()

	for _, tc := range testCases {
		t.Run(tc.class, func(t *testing.T) {
			char := &character2.Character{
				ID:    "test-char",
				Race:  &rulebook.Race{Key: "human", Name: "Human"},
				Class: &rulebook.Class{Key: tc.class, Name: tc.class},
				Level: 1,
			}

			flow, err := flowBuilder.BuildFlow(ctx, char)
			require.NoError(t, err)

			// Find cantrip selection step
			var cantripStep *character2.CreationStep
			for _, step := range flow.Steps {
				if step.Type == character2.StepTypeCantripsSelection {
					cantripStep = &step
					break
				}
			}

			if tc.class == "cleric" {
				// Clerics get cantrips automatically based on Wisdom, not through selection
				assert.Nil(t, cantripStep, "Cleric should not have cantrip selection step")
			} else {
				require.NotNil(t, cantripStep, "%s should have cantrip selection", tc.class)
				assert.Equal(t, tc.expectedCantrips, cantripStep.MinChoices,
					"%s should select %d cantrips", tc.class, tc.expectedCantrips)
				assert.Equal(t, tc.expectedCantrips, cantripStep.MaxChoices,
					"%s should select exactly %d cantrips", tc.class, tc.expectedCantrips)
			}
		})
	}
}
