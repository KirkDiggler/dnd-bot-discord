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

func TestPassiveFeaturesIntegration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockRepo := mockcharrepo.NewMockRepository(ctrl)
	mockDndClient := mockdnd5e.NewMockClient(ctrl)

	svc := character.NewService(&character.ServiceConfig{
		Repository: mockRepo,
		DNDClient:  mockDndClient,
	})

	t.Run("Elf character gets Perception proficiency from Keen Senses", func(t *testing.T) {
		characterID := "char-elf-123"
		userID := "user-123"
		realmID := "realm-123"

		elf := &entities.Character{
			ID:               characterID,
			OwnerID:          userID,
			RealmID:          realmID,
			Name:             "Legolas",
			Status:           entities.CharacterStatusDraft,
			Level:            1,
			HitDie:           6,
			MaxHitPoints:     8,
			CurrentHitPoints: 8,
			AbilityRolls: []entities.AbilityRoll{
				{ID: "roll1", Value: 16},
				{ID: "roll2", Value: 14},
				{ID: "roll3", Value: 14},
				{ID: "roll4", Value: 12},
				{ID: "roll5", Value: 10},
				{ID: "roll6", Value: 8},
			},
			AbilityAssignments: map[string]string{
				"STR": "roll4", // 12
				"DEX": "roll1", // 16
				"CON": "roll3", // 14
				"INT": "roll2", // 14
				"WIS": "roll5", // 10
				"CHA": "roll6", // 8
			},
			Race: &entities.Race{
				Key:  "elf",
				Name: "Elf",
			},
			Class: &entities.Class{
				Key:    "ranger",
				Name:   "Ranger",
				HitDie: 10,
			},
			Proficiencies: make(map[entities.ProficiencyType][]*entities.Proficiency),
		}

		// Mock repository Get
		mockRepo.EXPECT().
			Get(ctx, characterID).
			Return(elf, nil)

		// Mock repository Update
		mockRepo.EXPECT().
			Update(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, char *entities.Character) error {
				// Verify that passive features were applied
				assert.NotNil(t, char.Features)

				// Check that elf racial features exist
				hasKeenSenses := false
				hasDarkvision := false
				for _, feat := range char.Features {
					if feat.Key == "keen_senses" {
						hasKeenSenses = true
					}
					if feat.Key == "darkvision" {
						hasDarkvision = true
					}
				}
				assert.True(t, hasKeenSenses, "Should have Keen Senses feature")
				assert.True(t, hasDarkvision, "Should have Darkvision feature")

				// Most importantly: check that Perception proficiency was granted
				assert.True(t, char.HasSkillProficiency("skill-perception"),
					"Should have Perception proficiency from Keen Senses passive effect")

				return nil
			})

		// Finalize the character
		result, err := svc.FinalizeDraftCharacter(ctx, characterID)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, entities.CharacterStatusActive, result.Status)
	})

	t.Run("Multiple racial features apply correctly", func(t *testing.T) {
		characterID := "char-dwarf-456"
		userID := "user-456"
		realmID := "realm-456"

		dwarf := &entities.Character{
			ID:      characterID,
			OwnerID: userID,
			RealmID: realmID,
			Name:    "Gimli",
			Status:  entities.CharacterStatusDraft,
			Level:   1,
			Attributes: map[entities.Attribute]*entities.AbilityScore{
				entities.AttributeStrength:     {Score: 16, Bonus: 3},
				entities.AttributeDexterity:    {Score: 12, Bonus: 1},
				entities.AttributeConstitution: {Score: 16, Bonus: 3},
				entities.AttributeIntelligence: {Score: 10, Bonus: 0},
				entities.AttributeWisdom:       {Score: 14, Bonus: 2},
				entities.AttributeCharisma:     {Score: 8, Bonus: -1},
			},
			Race: &entities.Race{
				Key:  "dwarf",
				Name: "Dwarf",
			},
			Class: &entities.Class{
				Key:    "fighter",
				Name:   "Fighter",
				HitDie: 10,
			},
			Proficiencies: make(map[entities.ProficiencyType][]*entities.Proficiency),
		}

		// Mock repository Get
		mockRepo.EXPECT().
			Get(ctx, characterID).
			Return(dwarf, nil)

		// Mock repository Update
		mockRepo.EXPECT().
			Update(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, char *entities.Character) error {
				// Verify that dwarf racial features exist
				hasDarkvision := false
				hasDwarvenResilience := false
				hasStonecunning := false
				for _, feat := range char.Features {
					switch feat.Key {
					case "darkvision":
						hasDarkvision = true
					case "dwarven_resilience":
						hasDwarvenResilience = true
					case "stonecunning":
						hasStonecunning = true
					}
				}
				assert.True(t, hasDarkvision, "Should have Darkvision feature")
				assert.True(t, hasDwarvenResilience, "Should have Dwarven Resilience feature")
				assert.True(t, hasStonecunning, "Should have Stonecunning feature")

				return nil
			})

		// Finalize the character
		result, err := svc.FinalizeDraftCharacter(ctx, characterID)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, entities.CharacterStatusActive, result.Status)
	})

	t.Run("UpdateDraftCharacter applies passive effects when race changes", func(t *testing.T) {
		characterID := "char-change-789"
		userID := "user-789"
		realmID := "realm-789"

		char := &entities.Character{
			ID:      characterID,
			OwnerID: userID,
			RealmID: realmID,
			Name:    "Changeling",
			Status:  entities.CharacterStatusDraft,
			Level:   1,
			Attributes: map[entities.Attribute]*entities.AbilityScore{
				entities.AttributeStrength:     {Score: 10, Bonus: 0},
				entities.AttributeDexterity:    {Score: 16, Bonus: 3},
				entities.AttributeConstitution: {Score: 14, Bonus: 2},
				entities.AttributeIntelligence: {Score: 12, Bonus: 1},
				entities.AttributeWisdom:       {Score: 13, Bonus: 1},
				entities.AttributeCharisma:     {Score: 8, Bonus: -1},
			},
			Race: &entities.Race{
				Key:  "human",
				Name: "Human",
			},
			Proficiencies: make(map[entities.ProficiencyType][]*entities.Proficiency),
		}

		// Mock repository Get
		mockRepo.EXPECT().
			Get(ctx, characterID).
			Return(char, nil)

		// Mock getting elf race
		elfRace := &entities.Race{
			Key:  "elf",
			Name: "Elf",
		}
		mockDndClient.EXPECT().
			GetRace("elf").
			Return(elfRace, nil)

		// Mock repository Update
		mockRepo.EXPECT().
			Update(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, updated *entities.Character) error {
				// Verify elf racial features were applied
				hasKeenSenses := false
				for _, feat := range updated.Features {
					if feat.Key == "keen_senses" {
						hasKeenSenses = true
					}
				}
				assert.True(t, hasKeenSenses, "Should have Keen Senses after race change")

				// Verify passive effect was applied
				assert.True(t, updated.HasSkillProficiency("skill-perception"),
					"Should have Perception proficiency from new race")

				return nil
			})

		// Update to elf race
		raceKey := "elf"
		updates := &character.UpdateDraftInput{
			RaceKey: &raceKey,
		}

		result, err := svc.UpdateDraftCharacter(ctx, characterID, updates)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "elf", result.Race.Key)
	})
}
