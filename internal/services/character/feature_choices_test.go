package character_test

import (
	"context"
	"testing"

	mockdnd5e "github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e/mock"
	charDomain "github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
	mockcharrepo "github.com/KirkDiggler/dnd-bot-discord/internal/repositories/characters/mock"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGetPendingFeatureChoices(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockRepo := mockcharrepo.NewMockRepository(ctrl)
	mockDndClient := mockdnd5e.NewMockClient(ctrl)

	svc := character.NewService(&character.ServiceConfig{
		Repository: mockRepo,
		DNDClient:  mockDndClient,
	})

	t.Run("Fighter with no fighting style selected", func(t *testing.T) {
		// Create a fighter character with fighting_style feature but no selection
		characterID := "fighter-123"
		fighterChar := &charDomain.Character{
			ID:    characterID,
			Name:  "Test Fighter",
			Level: 1,
			Class: &rulebook.Class{
				Key:  "fighter",
				Name: "Fighter",
			},
			Features: []*rulebook.CharacterFeature{
				{
					Key:      "fighting_style",
					Name:     "Fighting Style",
					Type:     rulebook.FeatureTypeClass,
					Metadata: nil, // No style selected yet
				},
			},
		}

		// Mock repository Get
		mockRepo.EXPECT().
			Get(ctx, characterID).
			Return(fighterChar, nil)

		// Get pending choices
		choices, err := svc.GetPendingFeatureChoices(ctx, characterID)
		require.NoError(t, err)
		require.Len(t, choices, 1)

		// Verify fighting style choice
		choice := choices[0]
		assert.Equal(t, rulebook.FeatureChoiceTypeFightingStyle, choice.Type)
		assert.Equal(t, "fighting_style", choice.FeatureKey)
		assert.Equal(t, "Fighting Style", choice.Name)
		assert.Greater(t, len(choice.Options), 0)

		// Verify some options exist
		optionKeys := make(map[string]bool)
		for _, opt := range choice.Options {
			optionKeys[opt.Key] = true
		}
		assert.True(t, optionKeys["archery"])
		assert.True(t, optionKeys["defense"])
		assert.True(t, optionKeys["dueling"])
	})

	t.Run("Fighter with fighting style already selected", func(t *testing.T) {
		// Create a fighter character with fighting style already selected
		characterID := "fighter-456"
		fighterChar := &charDomain.Character{
			ID:    characterID,
			Name:  "Test Fighter",
			Level: 1,
			Class: &rulebook.Class{
				Key:  "fighter",
				Name: "Fighter",
			},
			Features: []*rulebook.CharacterFeature{
				{
					Key:  "fighting_style",
					Name: "Fighting Style",
					Type: rulebook.FeatureTypeClass,
					Metadata: map[string]any{
						"style": "defense", // Already selected
					},
				},
			},
		}

		// Mock repository Get
		mockRepo.EXPECT().
			Get(ctx, characterID).
			Return(fighterChar, nil)

		// Get pending choices
		choices, err := svc.GetPendingFeatureChoices(ctx, characterID)
		require.NoError(t, err)
		assert.Empty(t, choices, "Should have no pending choices when style is already selected")
	})

	t.Run("Cleric with no divine domain selected", func(t *testing.T) {
		// Create a cleric character with divine_domain feature but no selection
		characterID := "cleric-123"
		clericChar := &charDomain.Character{
			ID:    characterID,
			Name:  "Test Cleric",
			Level: 1,
			Class: &rulebook.Class{
				Key:  "cleric",
				Name: "Cleric",
			},
			Features: []*rulebook.CharacterFeature{
				{
					Key:      "divine_domain",
					Name:     "Divine Domain",
					Type:     rulebook.FeatureTypeClass,
					Metadata: nil, // No domain selected yet
				},
			},
		}

		// Mock repository Get
		mockRepo.EXPECT().
			Get(ctx, characterID).
			Return(clericChar, nil)

		// Get pending choices
		choices, err := svc.GetPendingFeatureChoices(ctx, characterID)
		require.NoError(t, err)
		require.Len(t, choices, 1)

		// Verify divine domain choice
		choice := choices[0]
		assert.Equal(t, rulebook.FeatureChoiceTypeDivineDomain, choice.Type)
		assert.Equal(t, "divine_domain", choice.FeatureKey)
		assert.Equal(t, "Divine Domain", choice.Name)
		assert.Len(t, choice.Options, 7) // 7 PHB domains

		// Verify all domains are present
		domainKeys := make(map[string]bool)
		for _, opt := range choice.Options {
			domainKeys[opt.Key] = true
		}
		assert.True(t, domainKeys["knowledge"])
		assert.True(t, domainKeys["life"])
		assert.True(t, domainKeys["light"])
		assert.True(t, domainKeys["nature"])
		assert.True(t, domainKeys["tempest"])
		assert.True(t, domainKeys["trickery"])
		assert.True(t, domainKeys["war"])
	})

	t.Run("Ranger with both features missing", func(t *testing.T) {
		// Create a ranger character with neither feature selected
		characterID := "ranger-123"
		rangerChar := &charDomain.Character{
			ID:    characterID,
			Name:  "Test Ranger",
			Level: 1,
			Class: &rulebook.Class{
				Key:  "ranger",
				Name: "Ranger",
			},
			Features: []*rulebook.CharacterFeature{
				{
					Key:      "favored_enemy",
					Name:     "Favored Enemy",
					Type:     rulebook.FeatureTypeClass,
					Metadata: nil, // No enemy selected
				},
				{
					Key:      "natural_explorer",
					Name:     "Natural Explorer",
					Type:     rulebook.FeatureTypeClass,
					Metadata: nil, // No terrain selected
				},
			},
		}

		// Mock repository Get
		mockRepo.EXPECT().
			Get(ctx, characterID).
			Return(rangerChar, nil)

		// Get pending choices
		choices, err := svc.GetPendingFeatureChoices(ctx, characterID)
		require.NoError(t, err)
		assert.Len(t, choices, 2, "Ranger should have 2 pending choices")

		// Check both choices are present
		choiceTypes := make(map[rulebook.FeatureChoiceType]bool)
		for _, choice := range choices {
			choiceTypes[choice.Type] = true
		}
		assert.True(t, choiceTypes[rulebook.FeatureChoiceTypeFavoredEnemy])
		assert.True(t, choiceTypes[rulebook.FeatureChoiceTypeNaturalExplorer])
	})

	t.Run("Character with no class", func(t *testing.T) {
		// Create a character with no class
		characterID := "no-class-123"
		char := &charDomain.Character{
			ID:    characterID,
			Name:  "Test Character",
			Level: 1,
			Class: nil, // No class
		}

		// Mock repository Get
		mockRepo.EXPECT().
			Get(ctx, characterID).
			Return(char, nil)

		// Get pending choices
		choices, err := svc.GetPendingFeatureChoices(ctx, characterID)
		require.NoError(t, err)
		assert.Empty(t, choices, "Should have no choices without a class")
	})

	t.Run("Invalid character ID", func(t *testing.T) {
		// Test with empty character ID
		choices, err := svc.GetPendingFeatureChoices(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, choices)
	})
}
