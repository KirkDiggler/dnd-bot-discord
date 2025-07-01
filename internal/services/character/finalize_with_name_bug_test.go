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

func TestFinalizeCharacterWithNameBug(t *testing.T) {
	t.Run("FinalizeCharacterWithName preserves fighting style metadata (BUG FIXED)", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mockcharrepo.NewMockRepository(ctrl)
		mockDNDClient := mockdnd5e.NewMockClient(ctrl)

		svc := character.NewService(&character.ServiceConfig{
			DNDClient:  mockDNDClient,
			Repository: mockRepo,
		})

		ctx := context.Background()
		characterID := "test-fighter-bug"

		// Create a draft character with fighting style metadata (like from logs)
		draftChar := &character2.Character{
			ID:      characterID,
			OwnerID: "user123",
			RealmID: "realm123",
			Name:    "Draft Character", // This will change to "Stanthony Hopkins"
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
					Metadata:    map[string]any{},
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
		t.Logf("✅ 18:01:46 Draft Character has fighting_style metadata=%v", fightingStyleBefore.Metadata)

		// Mock expectations for the UpdateDraftCharacter call (first call)
		mockRepo.EXPECT().Get(ctx, characterID).Return(draftChar, nil).Times(1)

		// Mock the GetClass call that happens during UpdateDraftCharacter
		mockDNDClient.EXPECT().GetClass("fighter").Return(&rulebook.Class{
			Key:    "fighter",
			Name:   "Fighter",
			HitDie: 10,
		}, nil).Times(1)

		// Mock the first Update call (from UpdateDraftCharacter with name change)
		// This should now preserve the metadata (bug fixed)
		mockRepo.EXPECT().Update(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, char *character2.Character) error {
			// Verify the fix: name changes but metadata is preserved
			t.Logf("UPDATE 1: Character name changed to '%s', checking metadata...", char.Name)
			fightingStyle := findFeatureByKey(char.Features, "fighting_style")
			if fightingStyle != nil && fightingStyle.Metadata != nil && len(fightingStyle.Metadata) > 0 {
				t.Logf("   ✅ UPDATE 1: Fighting style metadata preserved: %v", fightingStyle.Metadata)
			} else {
				t.Logf("   ❌ UPDATE 1: Fighting style metadata LOST - fix failed!")
			}
			// Update key fields to avoid copying mutex
			draftChar.Name = char.Name
			draftChar.Status = char.Status
			draftChar.Features = char.Features
			return nil
		}).Times(1)

		// Mock expectations for the FinalizeDraftCharacter call (second call)
		mockRepo.EXPECT().Get(ctx, characterID).Return(draftChar, nil).Times(1)

		// Mock the second Update call (from FinalizeDraftCharacter)
		mockRepo.EXPECT().Update(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, char *character2.Character) error {
			// Update key fields to avoid copying mutex
			draftChar.Name = char.Name
			draftChar.Status = char.Status
			draftChar.Features = char.Features
			return nil
		}).Times(1)

		// Call FinalizeCharacterWithName - metadata should now be preserved
		// Parameters: characterID, name, raceKey, classKey
		finalizedChar, err := svc.FinalizeCharacterWithName(ctx, characterID, "Stanthony Hopkins", "", "fighter")
		require.NoError(t, err)
		require.NotNil(t, finalizedChar)

		// Verify the fix worked
		assert.Equal(t, shared.CharacterStatusActive, finalizedChar.Status, "Character should be active")
		assert.Equal(t, "Stanthony Hopkins", finalizedChar.Name, "Character name should be updated")

		fightingStyleAfter := findFeatureByKey(finalizedChar.Features, "fighting_style")
		require.NotNil(t, fightingStyleAfter, "Fighting style feature should exist")

		// This is the critical test - metadata should be preserved (bug fixed)
		assert.NotNil(t, fightingStyleAfter.Metadata, "Fighting style metadata should be preserved")
		assert.NotEmpty(t, fightingStyleAfter.Metadata, "Fighting style metadata should not be empty")

		if style, ok := fightingStyleAfter.Metadata["style"]; ok && style == "dueling" {
			t.Logf("✅ 18:01:47 Stanthony Hopkins has fighting_style metadata=%v (BUG FIXED)", fightingStyleAfter.Metadata)
		} else {
			t.Errorf("❌ Fighting style metadata incorrect: expected 'dueling', got %v", fightingStyleAfter.Metadata)
		}
	})
}
