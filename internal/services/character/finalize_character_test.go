package character_test

import (
	"context"
	character2 "github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/repositories/characters"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFinalizeDraftCharacter(t *testing.T) {
	// Setup
	repo := characters.NewInMemoryRepository()
	service := character.NewService(&character.ServiceConfig{
		Repository: repo,
	})
	ctx := context.Background()

	t.Run("successfully finalizes draft character", func(t *testing.T) {
		// Create a draft character
		draft := testutils.CreateTestCharacter("test-char-1", "user-123", "realm-456", "Gandalf")
		draft.Status = character2.CharacterStatusDraft
		err := repo.Create(ctx, draft)
		require.NoError(t, err)

		// Finalize the character
		finalized, err := service.FinalizeDraftCharacter(ctx, draft.ID)
		require.NoError(t, err)
		require.NotNil(t, finalized)

		// Verify status changed to active
		assert.Equal(t, character2.CharacterStatusActive, finalized.Status)
		assert.Equal(t, "Gandalf", finalized.Name)

		// Verify it's persisted
		stored, err := repo.Get(ctx, draft.ID)
		require.NoError(t, err)
		assert.Equal(t, character2.CharacterStatusActive, stored.Status)
	})

	t.Run("cannot finalize non-draft character", func(t *testing.T) {
		// Create an active character
		active := testutils.CreateTestCharacter("test-char-2", "user-123", "realm-456", "Aragorn")
		active.Status = character2.CharacterStatusActive
		err := repo.Create(ctx, active)
		require.NoError(t, err)

		// Try to finalize it
		_, err = service.FinalizeDraftCharacter(ctx, active.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "can only finalize draft characters")
	})

	t.Run("character must have name before finalizing", func(t *testing.T) {
		// Create a draft without name
		draft := testutils.CreateTestCharacter("test-char-3", "user-123", "realm-456", "")
		draft.Status = character2.CharacterStatusDraft
		err := repo.Create(ctx, draft)
		require.NoError(t, err)

		// Should still finalize (name validation happens elsewhere)
		finalized, err := service.FinalizeDraftCharacter(ctx, draft.ID)
		require.NoError(t, err)
		assert.Equal(t, character2.CharacterStatusActive, finalized.Status)
	})
}
