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

func TestAbilityAssignmentPersistence_SimulateDiscordFlow(t *testing.T) {
	// This test simulates the exact Discord flow to find where attributes are lost

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mockdnd5e.NewMockClient(ctrl)
	mockRepo := mockcharrepo.NewMockRepository(ctrl)

	svc := character.NewService(&character.ServiceConfig{
		DNDClient:  mockClient,
		Repository: mockRepo,
	})

	ctx := context.Background()
	userID := "discord_user_123"
	realmID := "discord_realm"
	characterID := "char_test"

	// Step 1: GetOrCreateDraftCharacter
	// This simulates what happens when user starts character creation

	// Mock the repository calls for GetOrCreateDraftCharacter
	mockRepo.EXPECT().GetByOwnerAndRealm(ctx, userID, realmID).Return([]*character2.Character{}, nil)
	mockRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil)

	draft, err := svc.GetOrCreateDraftCharacter(ctx, userID, realmID)
	require.NoError(t, err)
	assert.Equal(t, shared.CharacterStatusDraft, draft.Status)

	// Step 2: UpdateDraftCharacter with ability assignments
	// This simulates the assign_abilities handler
	charWithAssignments := &character2.Character{
		ID:      characterID,
		OwnerID: userID,
		RealmID: realmID,
		Name:    "Draft Character",
		Status:  shared.CharacterStatusDraft,
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
			{ID: "roll_1", Value: 14},
		},
		AbilityAssignments: map[string]string{
			"DEX": "roll_1",
		},
		// Key point: UpdateDraftCharacter should populate Attributes
		Attributes:    make(map[shared.Attribute]*character2.AbilityScore),
		Proficiencies: make(map[rulebook.ProficiencyType][]*rulebook.Proficiency),
		Inventory:     make(map[equipment.EquipmentType][]equipment.Equipment),
		EquippedSlots: make(map[character2.Slot]equipment.Equipment),
	}

	mockRepo.EXPECT().Get(ctx, characterID).Return(charWithAssignments, nil)

	// Here's the potential bug: UpdateDraftCharacter should convert assignments to attributes
	var updatedChar *character2.Character
	mockRepo.EXPECT().Update(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, char *character2.Character) error {
		updatedChar = char
		// Verify that UpdateDraftCharacter populated the Attributes
		assert.NotEmpty(t, char.Attributes, "UpdateDraftCharacter should populate Attributes from AbilityAssignments")
		return nil
	})

	updates := &character.UpdateDraftInput{
		AbilityRolls: []character2.AbilityRoll{
			{ID: "roll_1", Value: 14},
		},
		AbilityAssignments: map[string]string{
			"DEX": "roll_1",
		},
	}

	updated, err := svc.UpdateDraftCharacter(ctx, characterID, updates)
	require.NoError(t, err)

	// THIS IS KEY: UpdateDraftCharacter should populate Attributes immediately
	assert.NotEmpty(t, updated.Attributes, "Attributes should be populated after UpdateDraftCharacter")

	if updatedChar != nil {
		t.Logf("After UpdateDraftCharacter - Attributes: %d, AbilityAssignments: %d",
			len(updatedChar.Attributes), len(updatedChar.AbilityAssignments))
	}

	// Step 3: Verify what gets loaded later
	// This simulates checking if the character is complete
	mockRepo.EXPECT().Get(ctx, characterID).DoAndReturn(func(_ context.Context, id string) (*character2.Character, error) {
		// Return what was saved - this is where the bug might be
		// If Attributes weren't saved properly, they'd be empty here
		return updatedChar, nil
	}).AnyTimes()

	loadedChar, err := svc.GetCharacter(ctx, characterID)
	require.NoError(t, err)

	// Verify the loaded character has attributes
	assert.NotEmpty(t, loadedChar.Attributes, "Loaded character should have attributes")
	assert.True(t, loadedChar.IsComplete(), "Character should be complete")
}
