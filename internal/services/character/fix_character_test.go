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

func TestFixCharacterAttributes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mockdnd5e.NewMockClient(ctrl)
	mockRepo := mockcharrepo.NewMockRepository(ctrl)

	svc := character.NewService(&character.ServiceConfig{
		DNDClient:  mockClient,
		Repository: mockRepo,
	})

	ctx := context.Background()
	characterID := "broken_char"

	// Character with the bug: has AbilityAssignments but no Attributes
	brokenChar := &character2.Character{
		ID:      characterID,
		Name:    "BrokenMonk",
		OwnerID: "user_123",
		RealmID: "realm_123",
		Status:  shared.CharacterStatusActive,
		Level:   1,
		Race: &rulebook.Race{
			Key:  "elf",
			Name: "Elf",
			AbilityBonuses: []*character2.AbilityBonus{
				{Attribute: shared.AttributeDexterity, Bonus: 2},
			},
		},
		Class: &rulebook.Class{
			Key:    "monk",
			Name:   "Monk",
			HitDie: 8,
		},
		AbilityRolls: []character2.AbilityRoll{
			{ID: "roll_1", Value: 15},
			{ID: "roll_2", Value: 14},
			{ID: "roll_3", Value: 13},
			{ID: "roll_4", Value: 12},
			{ID: "roll_5", Value: 11},
			{ID: "roll_6", Value: 10},
		},
		AbilityAssignments: map[string]string{
			"STR": "roll_3",
			"DEX": "roll_2",
			"CON": "roll_4",
			"INT": "roll_1",
			"WIS": "roll_5",
			"CHA": "roll_6",
		},
		Attributes:    map[shared.Attribute]*character2.AbilityScore{}, // Empty!
		Proficiencies: make(map[rulebook.ProficiencyType][]*rulebook.Proficiency),
		Inventory:     make(map[equipment.EquipmentType][]equipment.Equipment),
		EquippedSlots: make(map[shared.Slot]equipment.Equipment),
	}

	// Mock repository calls
	mockRepo.EXPECT().Get(ctx, characterID).Return(brokenChar, nil)

	// Expect the fixed character to be saved
	mockRepo.EXPECT().Update(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, char *character2.Character) error {
		// Verify the character was fixed
		assert.NotEmpty(t, char.Attributes, "Character should have attributes after fix")
		assert.Len(t, char.Attributes, 6, "Should have all 6 ability scores")

		// Verify specific scores with racial bonuses
		assert.Equal(t, 16, char.Attributes[shared.AttributeDexterity].Score, "DEX should be 14 + 2 racial")
		assert.Equal(t, 15, char.Attributes[shared.AttributeIntelligence].Score, "INT should be 15")

		// Verify HP was calculated
		assert.Equal(t, 9, char.MaxHitPoints, "HP should be 8 (hit die) + 1 (CON mod)")

		// Verify AC was calculated
		assert.Equal(t, 13, char.AC, "AC should be 10 + 3 (DEX mod)")

		return nil
	})

	// Mock getting class features for AC calculation
	mockClient.EXPECT().GetClassFeatures("monk", 1).Return([]*rulebook.CharacterFeature{
		{
			Name: "Unarmored Defense",
			Type: rulebook.FeatureTypeClass,
		},
	}, nil).AnyTimes()

	// Call the fix method
	fixedChar, err := svc.FixCharacterAttributes(ctx, characterID)
	require.NoError(t, err)
	require.NotNil(t, fixedChar)

	// Verify the character is now complete
	assert.True(t, fixedChar.IsComplete(), "Fixed character should be complete")
	assert.NotEmpty(t, fixedChar.Attributes, "Fixed character should have attributes")
}

func TestFixCharacterAttributes_AlreadyHasAttributes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mockdnd5e.NewMockClient(ctrl)
	mockRepo := mockcharrepo.NewMockRepository(ctrl)

	svc := character.NewService(&character.ServiceConfig{
		DNDClient:  mockClient,
		Repository: mockRepo,
	})

	ctx := context.Background()
	characterID := "good_char"

	// Character that doesn't need fixing
	goodChar := &character2.Character{
		ID:     characterID,
		Name:   "GoodChar",
		Status: shared.CharacterStatusActive,
		Attributes: map[shared.Attribute]*character2.AbilityScore{
			shared.AttributeStrength: {Score: 15, Bonus: 2},
		},
	}

	// Mock repository - should only get, not update
	mockRepo.EXPECT().Get(ctx, characterID).Return(goodChar, nil)
	// No Update expected since character doesn't need fixing

	// Call the fix method
	fixedChar, err := svc.FixCharacterAttributes(ctx, characterID)
	require.NoError(t, err)
	require.NotNil(t, fixedChar)

	// Should return the same character unchanged
	assert.Equal(t, goodChar, fixedChar)
}
