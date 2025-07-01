package character_test

import (
	"context"
	character2 "github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"testing"

	mockdnd5e "github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e/mock"
	mockcharrepo "github.com/KirkDiggler/dnd-bot-discord/internal/repositories/characters/mock"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestFinalizeDraftCharacterPreservesMetadata(t *testing.T) {
	t.Run("Fighting style metadata is preserved during finalization", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mockcharrepo.NewMockRepository(ctrl)
		mockDNDClient := mockdnd5e.NewMockClient(ctrl)

		svc := character.NewService(&character.ServiceConfig{
			DNDClient:  mockDNDClient,
			Repository: mockRepo,
		})

		ctx := context.Background()
		characterID := "test-fighter-id"

		// Create a draft character with fighting style metadata
		draftChar := &character2.Character{
			ID:      characterID,
			OwnerID: "user123",
			RealmID: "realm123",
			Name:    "Draft Character",
			Status:  shared.CharacterStatusDraft,
			Level:   1,
			Race: &rulebook.Race{
				Key:   "human",
				Name:  "Human",
				Speed: 30,
			},
			Class: &rulebook.Class{
				Key:    "fighter",
				Name:   "Fighter",
				HitDie: 10,
			},
			Features: []*rulebook.CharacterFeature{
				{
					Key:         "fighting_style",
					Name:        "Fighting Style",
					Description: "You adopt a particular style of fighting as your specialty.",
					Type:        rulebook.FeatureTypeClass,
					Level:       1,
					Source:      "Fighter",
					Metadata: map[string]any{
						"style": "dueling", // Critical: User has selected dueling
					},
				},
				{
					Key:         "second_wind",
					Name:        "Second Wind",
					Description: "You have a limited well of stamina that you can draw on to protect yourself from harm.",
					Type:        rulebook.FeatureTypeClass,
					Level:       1,
					Source:      "Fighter",
				},
			},
			Attributes: map[shared.Attribute]*character2.AbilityScore{
				shared.AttributeStrength:     {Score: 16, Bonus: 3},
				shared.AttributeDexterity:    {Score: 14, Bonus: 2},
				shared.AttributeConstitution: {Score: 15, Bonus: 2},
				shared.AttributeIntelligence: {Score: 13, Bonus: 1},
				shared.AttributeWisdom:       {Score: 12, Bonus: 1},
				shared.AttributeCharisma:     {Score: 10, Bonus: 0},
			},
		}

		// Verify initial state
		fightingStyleBefore := findFeatureByKey(draftChar.Features, "fighting_style")
		require.NotNil(t, fightingStyleBefore)
		require.NotNil(t, fightingStyleBefore.Metadata)
		assert.Equal(t, "dueling", fightingStyleBefore.Metadata["style"])
		t.Logf("✅ Draft Character has fighting_style metadata=%v", fightingStyleBefore.Metadata)

		// Mock expectations
		mockRepo.EXPECT().Get(ctx, characterID).Return(draftChar, nil)
		mockRepo.EXPECT().Update(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, char *character2.Character) error {
			// This is where we'll capture the finalized character to verify the fix
			// Copy the character back to draftChar so we can verify it
			// Update key fields to avoid copying mutex
			draftChar.Name = char.Name
			draftChar.Status = char.Status
			draftChar.Features = char.Features
			return nil
		})

		// Call FinalizeDraftCharacter directly
		finalizedChar, err := svc.FinalizeDraftCharacter(ctx, characterID)
		require.NoError(t, err)
		require.NotNil(t, finalizedChar)

		// Verify the fix - metadata should be preserved!
		assert.Equal(t, shared.CharacterStatusActive, finalizedChar.Status, "Character should be active")

		fightingStyleAfter := findFeatureByKey(finalizedChar.Features, "fighting_style")
		require.NotNil(t, fightingStyleAfter, "Fighting style feature should exist")

		// This is the critical test - metadata should NOT be nil
		assert.NotNil(t, fightingStyleAfter.Metadata, "Fighting style metadata should be preserved")
		if fightingStyleAfter.Metadata != nil {
			assert.Equal(t, "dueling", fightingStyleAfter.Metadata["style"], "Fighting style selection should be preserved")
			t.Logf("✅ Stanthony Hopkins has fighting_style metadata=%v", fightingStyleAfter.Metadata)
		} else {
			t.Logf("❌ Stanthony Hopkins has fighting_style metadata=nil (BUG NOT FIXED)")
		}

		// Also verify that second wind is still there
		secondWind := findFeatureByKey(finalizedChar.Features, "second_wind")
		assert.NotNil(t, secondWind, "Second Wind should be preserved")
	})
}

// Helper function to find a feature by key
func findFeatureByKey(features []*rulebook.CharacterFeature, key string) *rulebook.CharacterFeature {
	for _, feature := range features {
		if feature.Key == key {
			return feature
		}
	}
	return nil
}
