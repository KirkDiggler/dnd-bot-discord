package character_test

import (
	"context"
	"testing"

	mockdnd5e "github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e/mock"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	mockcharrepo "github.com/KirkDiggler/dnd-bot-discord/internal/repositories/characters/mock"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestAbilityAssignmentFlow_Integration(t *testing.T) {
	t.Skip("Skipping test - needs full mock setup")

	// This test reproduces the exact bug: characters showing 0 attributes after creation
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mockdnd5e.NewMockClient(ctrl)
	mockRepo := mockcharrepo.NewMockRepository(ctrl)

	svc := character.NewService(&character.ServiceConfig{
		DNDClient:  mockClient,
		Repository: mockRepo,
	})

	ctx := context.Background()
	userID := "test-user"
	realmID := "test-realm"

	// Step 1: Get or create draft character
	// Mock expects no existing characters
	mockRepo.EXPECT().GetByOwnerAndRealm(ctx, userID, realmID).Return([]*entities.Character{}, nil)

	// Mock the creation of a new draft character
	mockRepo.EXPECT().Create(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, char *entities.Character) error {
		// Validate the character being created
		assert.Equal(t, userID, char.OwnerID)
		assert.Equal(t, realmID, char.RealmID)
		assert.Equal(t, entities.CharacterStatusDraft, char.Status)
		return nil
	})

	char, err := svc.GetOrCreateDraftCharacter(ctx, userID, realmID)
	require.NoError(t, err)
	require.NotNil(t, char)
	assert.Equal(t, entities.CharacterStatusDraft, char.Status)

	// Step 2: Update with race
	mockClient.EXPECT().GetRace("elf").Return(&entities.Race{
		Key:  "elf",
		Name: "Elf",
		AbilityBonuses: []*entities.AbilityBonus{
			{Attribute: entities.AttributeDexterity, Bonus: 2},
			{Attribute: entities.AttributeIntelligence, Bonus: 1},
		},
		Speed: 30,
	}, nil)

	// Mock Get and Update for race update
	mockRepo.EXPECT().Get(ctx, char.ID).Return(char, nil)
	mockRepo.EXPECT().Update(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, updated *entities.Character) error {
		char = updated
		return nil
	})

	raceKey := "elf"
	char, err = svc.UpdateDraftCharacter(ctx, char.ID, &character.UpdateDraftInput{
		RaceKey: &raceKey,
	})
	require.NoError(t, err)
	assert.Equal(t, "elf", char.Race.Key)

	// Step 3: Update with class
	mockClient.EXPECT().GetClass("wizard").Return(&entities.Class{
		Key:    "wizard",
		Name:   "Wizard",
		HitDie: 6,
	}, nil)

	classKey := "wizard"
	char, err = svc.UpdateDraftCharacter(ctx, char.ID, &character.UpdateDraftInput{
		ClassKey: &classKey,
	})
	require.NoError(t, err)
	assert.Equal(t, "wizard", char.Class.Key)

	// Step 4: Assign abilities (THIS IS WHERE THE BUG OCCURS)
	// Simulate the exact data structure from the Discord handler
	abilityRolls := []entities.AbilityRoll{
		{ID: "roll_1", Value: 15},
		{ID: "roll_2", Value: 14},
		{ID: "roll_3", Value: 13},
		{ID: "roll_4", Value: 12},
		{ID: "roll_5", Value: 11},
		{ID: "roll_6", Value: 10},
	}

	abilityAssignments := map[string]string{
		"STR": "roll_3", // 13
		"DEX": "roll_2", // 14 + 2 (racial) = 16
		"CON": "roll_4", // 12
		"INT": "roll_1", // 15 + 1 (racial) = 16
		"WIS": "roll_5", // 11
		"CHA": "roll_6", // 10
	}

	char, err = svc.UpdateDraftCharacter(ctx, char.ID, &character.UpdateDraftInput{
		AbilityRolls:       abilityRolls,
		AbilityAssignments: abilityAssignments,
	})
	require.NoError(t, err)

	// At this point, the character should have attributes
	t.Logf("After ability assignment - Attributes: %d, AbilityAssignments: %d, AbilityRolls: %d",
		len(char.Attributes), len(char.AbilityAssignments), len(char.AbilityRolls))

	// Step 5: Update name
	name := "TestWizard"
	char, err = svc.UpdateDraftCharacter(ctx, char.ID, &character.UpdateDraftInput{
		Name: &name,
	})
	require.NoError(t, err)

	// Step 6: Finalize character
	mockClient.EXPECT().ListClassFeatures("wizard", 1).Return(nil, nil).AnyTimes()

	finalChar, err := svc.FinalizeDraftCharacter(ctx, char.ID)
	require.NoError(t, err)
	require.NotNil(t, finalChar)

	// THIS IS THE BUG: Character should have attributes but doesn't
	assert.Equal(t, entities.CharacterStatusActive, finalChar.Status)
	assert.NotEmpty(t, finalChar.Attributes, "Character should have attributes after finalization")
	assert.Len(t, finalChar.Attributes, 6, "Character should have all 6 ability scores")

	// Verify the character is complete
	assert.True(t, finalChar.IsComplete(), "Character should be complete after finalization")

	// Verify specific ability scores with racial bonuses
	assert.Equal(t, 16, finalChar.Attributes[entities.AttributeDexterity].Score, "DEX should be 14 + 2 racial")
	assert.Equal(t, 16, finalChar.Attributes[entities.AttributeIntelligence].Score, "INT should be 15 + 1 racial")
	assert.Equal(t, 13, finalChar.Attributes[entities.AttributeStrength].Score, "STR should be 13")
}

func TestUpdateDraftCharacter_AbilityAssignmentConversion(t *testing.T) {
	// This test specifically verifies that UpdateDraftCharacter converts assignments to attributes
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mockdnd5e.NewMockClient(ctrl)
	mockRepo := mockcharrepo.NewMockRepository(ctrl)

	svc := character.NewService(&character.ServiceConfig{
		DNDClient:  mockClient,
		Repository: mockRepo,
	})

	ctx := context.Background()

	// Create a character with race already set
	char := &entities.Character{
		ID:      "test-char",
		OwnerID: "test-user",
		RealmID: "test-realm",
		Status:  entities.CharacterStatusDraft,
		Race: &entities.Race{
			Key:  "elf",
			Name: "Elf",
			AbilityBonuses: []*entities.AbilityBonus{
				{Attribute: entities.AttributeDexterity, Bonus: 2},
			},
		},
		Attributes: make(map[entities.Attribute]*entities.AbilityScore),
	}

	// Mock the initial create
	mockRepo.EXPECT().Create(ctx, char).Return(nil)
	err := mockRepo.Create(ctx, char)
	require.NoError(t, err)

	// Mock the Get call for UpdateDraftCharacter
	mockRepo.EXPECT().Get(ctx, char.ID).Return(char, nil)
	// Mock the Update call
	mockRepo.EXPECT().Update(ctx, gomock.Any()).Return(nil)

	// Update with ability assignments
	updated, err := svc.UpdateDraftCharacter(ctx, char.ID, &character.UpdateDraftInput{
		AbilityRolls: []entities.AbilityRoll{
			{ID: "roll_1", Value: 14},
		},
		AbilityAssignments: map[string]string{
			"DEX": "roll_1", // 14 + 2 racial = 16
		},
	})
	require.NoError(t, err)

	// Verify the conversion happened immediately
	assert.NotEmpty(t, updated.Attributes, "Attributes should be populated after assignment")
	assert.NotNil(t, updated.Attributes[entities.AttributeDexterity], "DEX attribute should exist")
	assert.Equal(t, 16, updated.Attributes[entities.AttributeDexterity].Score, "DEX should include racial bonus")
	assert.Equal(t, 3, updated.Attributes[entities.AttributeDexterity].Bonus, "DEX modifier should be (16-10)/2 = 3")
}
