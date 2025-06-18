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

func TestService_FinalizeDraftCharacter_ConvertsAbilityAssignments(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mockcharrepo.NewMockRepository(ctrl)
	mockDNDClient := mockdnd5e.NewMockClient(ctrl)

	svc := character.NewService(&character.ServiceConfig{
		DNDClient:  mockDNDClient,
		Repository: mockRepo,
	})

	ctx := context.Background()
	characterID := "char_123"

	// Create a draft character with AbilityAssignments but no Attributes
	draftChar := &entities.Character{
		ID:      characterID,
		Name:    "Test Character",
		OwnerID: "user_123",
		RealmID: "realm_123",
		Status:  entities.CharacterStatusDraft,
		Level:   1,
		Race: &entities.Race{
			Key:  "elf",
			Name: "Elf",
			AbilityBonuses: []*entities.AbilityBonus{
				{Attribute: entities.AttributeDexterity, Bonus: 2},
				{Attribute: entities.AttributeIntelligence, Bonus: 1},
			},
		},
		Class: &entities.Class{
			Key:    "wizard",
			Name:   "Wizard",
			HitDie: 6,
		},
		AbilityRolls: []entities.AbilityRoll{
			{ID: "roll_1", Value: 15},
			{ID: "roll_2", Value: 14},
			{ID: "roll_3", Value: 13},
			{ID: "roll_4", Value: 12},
			{ID: "roll_5", Value: 11},
			{ID: "roll_6", Value: 10},
		},
		AbilityAssignments: map[string]string{
			"STR": "roll_3", // 13
			"DEX": "roll_2", // 14 + 2 (racial) = 16
			"CON": "roll_4", // 12
			"INT": "roll_1", // 15 + 1 (racial) = 16
			"WIS": "roll_5", // 11
			"CHA": "roll_6", // 10
		},
		Attributes: make(map[entities.Attribute]*entities.AbilityScore), // Empty attributes
	}

	// Mock repository calls
	mockRepo.EXPECT().Get(ctx, characterID).Return(draftChar, nil)

	// Mock ListClassFeatures call (happens during finalization)
	mockDNDClient.EXPECT().ListClassFeatures("wizard", 1).Return(nil, nil)

	// Expect update with converted attributes
	mockRepo.EXPECT().Update(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, char *entities.Character) error {
		// Verify attributes were converted correctly
		assert.Equal(t, entities.CharacterStatusActive, char.Status)

		// Check STR (13, no racial bonus)
		assert.NotNil(t, char.Attributes[entities.AttributeStrength])
		assert.Equal(t, 13, char.Attributes[entities.AttributeStrength].Score)
		assert.Equal(t, 1, char.Attributes[entities.AttributeStrength].Bonus) // (13-10)/2 = 1

		// Check DEX (14 + 2 racial = 16)
		assert.NotNil(t, char.Attributes[entities.AttributeDexterity])
		assert.Equal(t, 16, char.Attributes[entities.AttributeDexterity].Score)
		assert.Equal(t, 3, char.Attributes[entities.AttributeDexterity].Bonus) // (16-10)/2 = 3

		// Check INT (15 + 1 racial = 16)
		assert.NotNil(t, char.Attributes[entities.AttributeIntelligence])
		assert.Equal(t, 16, char.Attributes[entities.AttributeIntelligence].Score)
		assert.Equal(t, 3, char.Attributes[entities.AttributeIntelligence].Bonus) // (16-10)/2 = 3

		// Check HP calculation (6 base + 1 con modifier)
		assert.Equal(t, 7, char.MaxHitPoints)
		assert.Equal(t, 7, char.CurrentHitPoints)

		// Check AC calculation (10 + 3 dex modifier)
		assert.Equal(t, 13, char.AC)

		return nil
	})

	// Execute
	result, err := svc.FinalizeDraftCharacter(ctx, characterID)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, entities.CharacterStatusActive, result.Status)
}

func TestService_FinalizeDraftCharacter_PreservesExistingAttributes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mockcharrepo.NewMockRepository(ctrl)
	mockDNDClient := mockdnd5e.NewMockClient(ctrl)

	svc := character.NewService(&character.ServiceConfig{
		DNDClient:  mockDNDClient,
		Repository: mockRepo,
	})

	ctx := context.Background()
	characterID := "char_456"

	// Create a draft character with existing Attributes (should not convert)
	draftChar := &entities.Character{
		ID:      characterID,
		Name:    "Test Character",
		OwnerID: "user_123",
		RealmID: "realm_123",
		Status:  entities.CharacterStatusDraft,
		Level:   1,
		Class: &entities.Class{
			Key:    "fighter",
			Name:   "Fighter",
			HitDie: 10,
		},
		Attributes: map[entities.Attribute]*entities.AbilityScore{
			entities.AttributeStrength:     {Score: 16, Bonus: 3},
			entities.AttributeDexterity:    {Score: 14, Bonus: 2},
			entities.AttributeConstitution: {Score: 15, Bonus: 2},
			entities.AttributeIntelligence: {Score: 10, Bonus: 0},
			entities.AttributeWisdom:       {Score: 12, Bonus: 1},
			entities.AttributeCharisma:     {Score: 8, Bonus: -1},
		},
		MaxHitPoints:     12, // Already calculated
		CurrentHitPoints: 12,
		AC:               12, // Already calculated
	}

	// Mock repository calls
	mockRepo.EXPECT().Get(ctx, characterID).Return(draftChar, nil)

	// Mock ListClassFeatures call (happens during finalization)
	mockDNDClient.EXPECT().ListClassFeatures("fighter", 1).Return(nil, nil)

	// Expect update without conversion
	mockRepo.EXPECT().Update(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, char *entities.Character) error {
		// Verify attributes were NOT changed
		assert.Equal(t, entities.CharacterStatusActive, char.Status)
		assert.Equal(t, 16, char.Attributes[entities.AttributeStrength].Score)
		assert.Equal(t, 3, char.Attributes[entities.AttributeStrength].Bonus)

		// HP and AC should remain the same
		assert.Equal(t, 12, char.MaxHitPoints)
		assert.Equal(t, 12, char.AC)

		return nil
	})

	// Execute
	result, err := svc.FinalizeDraftCharacter(ctx, characterID)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, entities.CharacterStatusActive, result.Status)
}
