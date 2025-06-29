package character

import (
	"context"
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/repositories/characters/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestActualFinalizeDraftCharacterPreservesMetadata(t *testing.T) {
	// Test the ACTUAL FinalizeDraftCharacter method to see if it preserves metadata

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mockcharacters.NewMockRepository(ctrl)
	service := NewService(&ServiceConfig{
		Repository: mockRepo,
	})

	t.Run("Real FinalizeDraftCharacter preserves fighting style metadata", func(t *testing.T) {
		// Create draft character with fighting style metadata (like from logs)
		draftChar := &entities.Character{
			ID:     "char_real_test",
			Name:   "Draft Character",
			Status: entities.CharacterStatusDraft,
			Race:   &entities.Race{Key: "human", Name: "Human"},
			Class:  &entities.Class{Key: "fighter", Name: "Fighter", HitDie: 10},
			Level:  1,
			Features: []*entities.CharacterFeature{
				{
					Key:      "human_versatility",
					Name:     "Versatility",
					Type:     entities.FeatureTypeRacial,
					Level:    0,
					Source:   "Human",
					Metadata: map[string]any{},
				},
				{
					Key:    "fighting_style",
					Name:   "Fighting Style",
					Type:   entities.FeatureTypeClass,
					Level:  1,
					Source: "Fighter",
					Metadata: map[string]any{
						"style": "dueling", // This should be preserved!
					},
				},
				{
					Key:      "second_wind",
					Name:     "Second Wind",
					Type:     entities.FeatureTypeClass,
					Level:    1,
					Source:   "Fighter",
					Metadata: map[string]any{},
				},
			},
		}

		// Mock the repository calls
		mockRepo.EXPECT().
			Get(gomock.Any(), "char_real_test").
			Return(draftChar, nil)

		// The finalized character should be saved back
		var savedChar *entities.Character
		mockRepo.EXPECT().
			Update(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, char *entities.Character) {
				savedChar = char
			}).
			Return(nil)

		// Call the actual FinalizeDraftCharacter method
		finalizedChar, err := service.FinalizeDraftCharacter(context.Background(), "char_real_test")
		require.NoError(t, err)
		require.NotNil(t, finalizedChar)

		// Check that the character status was updated
		assert.Equal(t, entities.CharacterStatusActive, finalizedChar.Status)

		// Most importantly: check if fighting style metadata is preserved
		fightingStyleFeature := findFeatureByKey(finalizedChar.Features, "fighting_style")
		require.NotNil(t, fightingStyleFeature, "Fighting style feature should exist")

		if fightingStyleFeature.Metadata == nil {
			t.Errorf("❌ METADATA LOST: Fighting style metadata is nil")
		} else if len(fightingStyleFeature.Metadata) == 0 {
			t.Errorf("❌ METADATA LOST: Fighting style metadata is empty: %v", fightingStyleFeature.Metadata)
		} else if style, ok := fightingStyleFeature.Metadata["style"]; !ok {
			t.Errorf("❌ STYLE KEY LOST: Fighting style metadata missing 'style' key: %v", fightingStyleFeature.Metadata)
		} else if style != "dueling" {
			t.Errorf("❌ STYLE VALUE WRONG: Expected 'dueling', got '%v'", style)
		} else {
			t.Logf("✅ SUCCESS: Fighting style metadata preserved: %v", fightingStyleFeature.Metadata)
		}

		// Verify that the character was saved with the correct metadata
		require.NotNil(t, savedChar, "Character should have been saved")
		savedFightingStyle := findFeatureByKey(savedChar.Features, "fighting_style")
		require.NotNil(t, savedFightingStyle)

		if len(savedFightingStyle.Metadata) > 0 {
			t.Logf("✅ SAVED CORRECTLY: Saved character has metadata: %v", savedFightingStyle.Metadata)
		} else {
			t.Errorf("❌ SAVE ISSUE: Saved character has empty metadata: %v", savedFightingStyle.Metadata)
		}
	})
}
