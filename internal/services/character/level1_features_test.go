package character_test

import (
	"context"
	character2 "github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/equipment"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"testing"

	mockdnd5e "github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e/mock"
	mockdraftrepo "github.com/KirkDiggler/dnd-bot-discord/internal/repositories/character_draft/mock"
	mockcharrepo "github.com/KirkDiggler/dnd-bot-discord/internal/repositories/characters/mock"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestLevel1Features(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockRepo := mockcharrepo.NewMockRepository(ctrl)
	mockDraftRepo := mockdraftrepo.NewMockRepository(ctrl)
	mockDndClient := mockdnd5e.NewMockClient(ctrl)

	svc := character.NewService(&character.ServiceConfig{
		Repository:      mockRepo,
		DraftRepository: mockDraftRepo,
		DNDClient:       mockDndClient,
	})

	t.Run("Barbarian Unarmored Defense", func(t *testing.T) {
		// Create a draft barbarian character
		characterID := "char-123"
		userID := "user-123"
		realmID := "realm-123"

		barbarian := &character2.Character{
			ID:               characterID,
			OwnerID:          userID,
			RealmID:          realmID,
			Name:             "Grog",
			Status:           shared.CharacterStatusDraft,
			Level:            1,
			HitDie:           12,
			MaxHitPoints:     15, // 12 + 3 CON
			CurrentHitPoints: 15,
			AbilityRolls: []character2.AbilityRoll{
				{ID: "roll1", Value: 18},
				{ID: "roll2", Value: 16},
				{ID: "roll3", Value: 14},
				{ID: "roll4", Value: 12},
				{ID: "roll5", Value: 10},
				{ID: "roll6", Value: 8},
			},
			AbilityAssignments: map[string]string{
				"STR": "roll1", // 18
				"DEX": "roll3", // 14
				"CON": "roll2", // 16
				"INT": "roll5", // 10
				"WIS": "roll4", // 12
				"CHA": "roll6", // 8
			},
			Race: &rulebook.Race{
				Key:  "human",
				Name: "Human",
			},
			Class: &rulebook.Class{
				Key:    "barbarian",
				Name:   "Barbarian",
				HitDie: 12,
			},
		}

		// Mock repository Get
		mockRepo.EXPECT().
			Get(ctx, characterID).
			Return(barbarian, nil)

		// Mock draft repository - no draft exists
		mockDraftRepo.EXPECT().
			GetByCharacterID(ctx, characterID).
			Return(nil, nil)

		// Mock repository Update
		mockRepo.EXPECT().
			Update(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, char *character2.Character) error {
				// Verify features were applied
				assert.NotNil(t, char.Features)

				// Check for barbarian features
				hasUnarmoredDefense := false
				hasRage := false
				for _, feat := range char.Features {
					if feat.Key == "unarmored_defense_barbarian" {
						hasUnarmoredDefense = true
						assert.Equal(t, "Unarmored Defense", feat.Name)
						assert.Equal(t, rulebook.FeatureTypeClass, feat.Type)
						assert.Equal(t, 1, feat.Level)
						assert.Equal(t, "Barbarian", feat.Source)
					}
					if feat.Key == "rage" {
						hasRage = true
						assert.Equal(t, "Rage", feat.Name)
					}
				}
				assert.True(t, hasUnarmoredDefense, "Should have Unarmored Defense")
				assert.True(t, hasRage, "Should have Rage")

				// Verify AC calculation (10 + DEX(2) + CON(3) = 15)
				assert.Equal(t, 15, char.AC, "AC should be 10 + DEX mod (2) + CON mod (3)")

				return nil
			})

		// Finalize the character
		result, err := svc.FinalizeDraftCharacter(ctx, characterID)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, shared.CharacterStatusActive, result.Status)
	})

	t.Run("Monk Unarmored Defense", func(t *testing.T) {
		// Create a draft monk character
		characterID := "char-456"
		userID := "user-456"
		realmID := "realm-456"

		monk := &character2.Character{
			ID:               characterID,
			OwnerID:          userID,
			RealmID:          realmID,
			Name:             "Kwai Chang",
			Status:           shared.CharacterStatusDraft,
			Level:            1,
			HitDie:           8,
			MaxHitPoints:     10, // 8 + 2 CON
			CurrentHitPoints: 10,
			AbilityRolls: []character2.AbilityRoll{
				{ID: "roll1", Value: 16},
				{ID: "roll2", Value: 16},
				{ID: "roll3", Value: 14},
				{ID: "roll4", Value: 12},
				{ID: "roll5", Value: 10},
				{ID: "roll6", Value: 8},
			},
			AbilityAssignments: map[string]string{
				"STR": "roll4", // 12
				"DEX": "roll1", // 16
				"CON": "roll3", // 14
				"INT": "roll5", // 10
				"WIS": "roll2", // 16
				"CHA": "roll6", // 8
			},
			Race: &rulebook.Race{
				Key:  "elf",
				Name: "Elf",
			},
			Class: &rulebook.Class{
				Key:    "monk",
				Name:   "Monk",
				HitDie: 8,
			},
		}

		// Mock repository Get
		mockRepo.EXPECT().
			Get(ctx, characterID).
			Return(monk, nil)

		// Mock draft repository - no draft exists
		mockDraftRepo.EXPECT().
			GetByCharacterID(ctx, characterID).
			Return(nil, nil)

		// Mock repository Update
		mockRepo.EXPECT().
			Update(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, char *character2.Character) error {
				// Verify features were applied
				assert.NotNil(t, char.Features)

				// Check for monk features
				hasUnarmoredDefense := false
				hasMartialArts := false
				for _, feat := range char.Features {
					if feat.Key == "unarmored_defense_monk" {
						hasUnarmoredDefense = true
						assert.Equal(t, "Unarmored Defense", feat.Name)
						assert.Equal(t, rulebook.FeatureTypeClass, feat.Type)
						assert.Equal(t, 1, feat.Level)
						assert.Equal(t, "Monk", feat.Source)
					}
					if feat.Key == "martial-arts" {
						hasMartialArts = true
						assert.Equal(t, "Martial Arts", feat.Name)
					}
				}
				assert.True(t, hasUnarmoredDefense, "Should have Unarmored Defense")
				assert.True(t, hasMartialArts, "Should have Martial Arts")

				// Also check for racial features
				hasDarkvision := false
				for _, feat := range char.Features {
					if feat.Key == "darkvision" {
						hasDarkvision = true
						assert.Equal(t, rulebook.FeatureTypeRacial, feat.Type)
						assert.Equal(t, "Elf", feat.Source)
					}
				}
				assert.True(t, hasDarkvision, "Should have Darkvision from Elf")

				// Verify AC calculation (10 + DEX(3) + WIS(3) = 16)
				assert.Equal(t, 16, char.AC, "AC should be 10 + DEX mod (3) + WIS mod (3)")

				return nil
			})

		// Finalize the character
		result, err := svc.FinalizeDraftCharacter(ctx, characterID)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, shared.CharacterStatusActive, result.Status)
	})

	t.Run("Fighter with Armor", func(t *testing.T) {
		// Create a draft fighter character with chain mail
		characterID := "char-789"
		userID := "user-789"
		realmID := "realm-789"

		fighter := &character2.Character{
			ID:               characterID,
			OwnerID:          userID,
			RealmID:          realmID,
			Name:             "Ser Arthur",
			Status:           shared.CharacterStatusDraft,
			Level:            1,
			HitDie:           10,
			MaxHitPoints:     12, // 10 + 2 CON
			CurrentHitPoints: 12,
			AbilityRolls: []character2.AbilityRoll{
				{ID: "roll1", Value: 16},
				{ID: "roll2", Value: 14},
				{ID: "roll3", Value: 14},
				{ID: "roll4", Value: 12},
				{ID: "roll5", Value: 10},
				{ID: "roll6", Value: 8},
			},
			AbilityAssignments: map[string]string{
				"STR": "roll1", // 16
				"DEX": "roll3", // 14
				"CON": "roll2", // 14
				"INT": "roll5", // 10
				"WIS": "roll4", // 12
				"CHA": "roll6", // 8
			},
			Race: &rulebook.Race{
				Key:  "human",
				Name: "Human",
			},
			Class: &rulebook.Class{
				Key:    "fighter",
				Name:   "Fighter",
				HitDie: 10,
			},
			EquippedSlots: map[shared.Slot]equipment.Equipment{
				shared.SlotBody: &equipment.Armor{
					Base: equipment.BasicEquipment{
						Key:  "chain-mail",
						Name: "Chain Mail",
					},
					ArmorCategory: equipment.ArmorCategoryHeavy,
					ArmorClass: &equipment.ArmorClass{
						Base:     16,
						DexBonus: false,
					},
				},
			},
		}

		// Mock repository Get
		mockRepo.EXPECT().
			Get(ctx, characterID).
			Return(fighter, nil)

		// Mock draft repository - no draft exists
		mockDraftRepo.EXPECT().
			GetByCharacterID(ctx, characterID).
			Return(nil, nil)

		// Mock repository Update
		mockRepo.EXPECT().
			Update(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, char *character2.Character) error {
				// Verify features were applied
				assert.NotNil(t, char.Features)

				// Check for fighter features
				hasFightingStyle := false
				hasSecondWind := false
				for _, feat := range char.Features {
					if feat.Key == "fighting_style" {
						hasFightingStyle = true
					}
					if feat.Key == "second_wind" {
						hasSecondWind = true
					}
				}
				assert.True(t, hasFightingStyle, "Should have Fighting Style")
				assert.True(t, hasSecondWind, "Should have Second Wind")

				// Verify AC calculation with armor (16 base, no DEX for heavy armor)
				assert.Equal(t, 16, char.AC, "AC should be 16 from chain mail")

				return nil
			})

		// Finalize the character
		result, err := svc.FinalizeDraftCharacter(ctx, characterID)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, shared.CharacterStatusActive, result.Status)
	})

	t.Run("UpdateDraftCharacter applies features when class changes", func(t *testing.T) {
		// Start with a character without a class
		characterID := "char-999"
		userID := "user-999"
		realmID := "realm-999"

		char := &character2.Character{
			ID:      characterID,
			OwnerID: userID,
			RealmID: realmID,
			Name:    "Changeling",
			Status:  shared.CharacterStatusDraft,
			Level:   1,
			Attributes: map[shared.Attribute]*character2.AbilityScore{
				shared.AttributeStrength:     {Score: 10, Bonus: 0},
				shared.AttributeDexterity:    {Score: 14, Bonus: 2},
				shared.AttributeConstitution: {Score: 16, Bonus: 3},
				shared.AttributeIntelligence: {Score: 12, Bonus: 1},
				shared.AttributeWisdom:       {Score: 13, Bonus: 1},
				shared.AttributeCharisma:     {Score: 8, Bonus: -1},
			},
			Race: &rulebook.Race{
				Key:  "human",
				Name: "Human",
			},
		}

		// Mock repository Get
		mockRepo.EXPECT().
			Get(ctx, characterID).
			Return(char, nil)

		// Mock draft repository - no draft exists
		mockDraftRepo.EXPECT().
			GetByCharacterID(ctx, characterID).
			Return(nil, nil)

		// Mock getting barbarian class
		barbarianClass := &rulebook.Class{
			Key:    "barbarian",
			Name:   "Barbarian",
			HitDie: 12,
		}
		mockDndClient.EXPECT().
			GetClass("barbarian").
			Return(barbarianClass, nil)

		// Mock repository Update
		mockRepo.EXPECT().
			Update(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, updated *character2.Character) error {
				// Verify barbarian features were applied
				hasUnarmoredDefense := false
				for _, feat := range updated.Features {
					if feat.Key == "unarmored_defense_barbarian" {
						hasUnarmoredDefense = true
					}
				}
				assert.True(t, hasUnarmoredDefense, "Should have Barbarian Unarmored Defense")

				// Verify AC was recalculated (10 + DEX(2) + CON(3) = 15)
				assert.Equal(t, 15, updated.AC, "AC should be recalculated with Unarmored Defense")

				return nil
			})

		// Update to barbarian class
		classKey := "barbarian"
		updates := &character.UpdateDraftInput{
			ClassKey: &classKey,
		}

		result, err := svc.UpdateDraftCharacter(ctx, characterID, updates)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "barbarian", result.Class.Key)
	})
}
