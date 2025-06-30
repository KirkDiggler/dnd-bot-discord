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

func TestDiscordCharacterCreationFlow(t *testing.T) {
	// This test simulates the exact Discord flow to identify where attributes might be lost

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mockdnd5e.NewMockClient(ctrl)
	mockRepo := mockcharrepo.NewMockRepository(ctrl)

	svc := character.NewService(&character.ServiceConfig{
		DNDClient:  mockClient,
		Repository: mockRepo,
	})

	ctx := context.Background()
	userID := "discord_user"
	realmID := "discord_realm"
	characterID := "char_123"

	// Simulate the exact state after ability assignment
	charState := &character2.Character{
		ID:      characterID,
		OwnerID: userID,
		RealmID: realmID,
		Name:    "NotGonnaWorkHere",
		Status:  shared.CharacterStatusActive, // Already finalized
		Level:   1,
		Race: &rulebook.Race{
			Key:  "elf",
			Name: "Elf",
			AbilityBonuses: []*shared.AbilityBonus{
				{Attribute: shared.AttributeDexterity, Bonus: 2},
			},
		},
		Class: &rulebook.Class{
			Key:    "monk",
			Name:   "Monk",
			HitDie: 8,
		},
		// This is what might be happening - character was finalized but attributes weren't saved
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
		Attributes:       map[shared.Attribute]*character2.AbilityScore{}, // Empty!
		Proficiencies:    make(map[rulebook.ProficiencyType][]*rulebook.Proficiency),
		Inventory:        make(map[equipment.EquipmentType][]equipment.Equipment),
		EquippedSlots:    make(map[shared.Slot]equipment.Equipment),
		MaxHitPoints:     9,
		CurrentHitPoints: 9,
		AC:               13,
	}

	// Test 1: ListByOwner returns character with empty attributes
	mockRepo.EXPECT().GetByOwner(ctx, userID).Return([]*character2.Character{charState}, nil)

	chars, err := svc.ListByOwner(userID)
	require.NoError(t, err)
	require.Len(t, chars, 1)

	char := chars[0]
	t.Logf("Character from ListByOwner - Name: %s, Status: %s, Attributes: %d, AbilityAssignments: %d",
		char.Name, char.Status, len(char.Attributes), len(char.AbilityAssignments))

	// This reproduces the bug
	assert.Empty(t, char.Attributes, "Character has empty attributes")
	assert.NotEmpty(t, char.AbilityAssignments, "Character has ability assignments")
	assert.False(t, char.IsComplete(), "Character shows as incomplete due to missing attributes")

	// Test 2: What happens if we try to "fix" the character
	// The service doesn't have a method to fix already finalized characters
	// This might be the root cause - characters get finalized before attributes are properly converted
}
