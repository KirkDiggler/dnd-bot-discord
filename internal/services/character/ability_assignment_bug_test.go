package character_test

import (
	"context"
	character2 "github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/equipment"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"testing"

	mockdnd5e "github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e/mock"
	mockcharrepo "github.com/KirkDiggler/dnd-bot-discord/internal/repositories/characters/mock"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestAbilityAssignmentBug_CharacterShowsZeroAttributes(t *testing.T) {
	// This test demonstrates the bug where characters show 0 attributes
	// even after ability assignment is complete

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mockdnd5e.NewMockClient(ctrl)
	mockRepo := mockcharrepo.NewMockRepository(ctrl)

	svc := character.NewService(&character.ServiceConfig{
		DNDClient:  mockClient,
		Repository: mockRepo,
	})

	ctx := context.Background()
	characterID := "char_123"

	// Character state after ability assignment but before finalization
	charAfterAssignment := &character2.Character{
		ID:      characterID,
		Name:    "NotGonnaWorkHere",
		OwnerID: "user_123",
		RealmID: "test_realm",
		Status:  shared.CharacterStatusDraft,
		Level:   1,
		Race: &rulebook.Race{
			Key:  "elf",
			Name: "Elf",
			AbilityBonuses: []*character2.AbilityBonus{
				{Attribute: shared.AttributeDexterity, Bonus: 2},
				{Attribute: shared.AttributeIntelligence, Bonus: 1},
			},
		},
		Class: &rulebook.Class{
			Key:    "monk",
			Name:   "Monk",
			HitDie: 8,
		},
		// THIS IS THE KEY: Character has AbilityAssignments but NO Attributes
		AbilityRolls: []character2.AbilityRoll{
			{ID: "roll_1", Value: 15},
			{ID: "roll_2", Value: 14},
			{ID: "roll_3", Value: 13},
			{ID: "roll_4", Value: 12},
			{ID: "roll_5", Value: 11},
			{ID: "roll_6", Value: 10},
		},
		AbilityAssignments: map[string]string{
			"STR": "roll_3", // Strength is roll 3 and has a score of 13
			"DEX": "roll_2", // Dexterity is roll 2 and has a score of 14 + 2 (racial) = 16
			"CON": "roll_4", // Constitution is roll 4 and has a score of 12
			"INT": "roll_1", // Intelligence is roll 1 and has a score of 15 + 1 (racial) = 16
			"WIS": "roll_5", // Wisdom is roll 5 and has a score of 11
			"CHA": "roll_6", // Charisma is roll 6 and has a score of 10
		},
		Attributes:    map[shared.Attribute]*character2.AbilityScore{}, // Empty!
		Proficiencies: make(map[rulebook.ProficiencyType][]*rulebook.Proficiency),
		Inventory:     make(map[equipment.EquipmentType][]equipment.Equipment),
		EquippedSlots: make(map[shared.Slot]equipment.Equipment),
		Features:      []*rulebook.CharacterFeature{},
	}

	// Mock the repository Get to return our test character
	mockRepo.EXPECT().Get(ctx, characterID).Return(charAfterAssignment, nil).Times(1)

	// Mock the Update call
	mockRepo.EXPECT().Update(ctx, gomock.Any()).Return(nil).Times(1)

	// Mock getting class features
	mockClient.EXPECT().GetClassFeatures("monk", 1).Return([]*rulebook.CharacterFeature{
		{
			Name: "Unarmored Defense",
			Type: rulebook.FeatureTypeClass,
		},
	}, nil).AnyTimes()

	// Call FinalizeDraftCharacter
	finalChar, err := svc.FinalizeDraftCharacter(ctx, characterID)
	require.NoError(t, err)
	require.NotNil(t, finalChar)

	// BUG VERIFICATION: Character should have attributes but doesn't
	t.Logf("Character after finalization - Name: %s, Attributes: %d, Status: %s",
		finalChar.Name, len(finalChar.Attributes), finalChar.Status)

	// These assertions would FAIL with the bug
	assert.NotEmpty(t, finalChar.Attributes, "BUG: Character has 0 attributes after finalization")
	assert.Len(t, finalChar.Attributes, 6, "BUG: Character missing ability scores")

	// Verify specific conversions with racial bonuses
	if len(finalChar.Attributes) > 0 {
		assert.Equal(t, 16, finalChar.Attributes[shared.AttributeDexterity].Score, "DEX should be 14 + 2 racial")
		assert.Equal(t, 16, finalChar.Attributes[shared.AttributeIntelligence].Score, "INT should be 15 + 1 racial")
	}

	// Verify the character shows as complete
	assert.True(t, finalChar.IsComplete(), "BUG: Character shows as incomplete due to missing ability scores")
}
