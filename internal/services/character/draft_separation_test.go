package character_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	inmemoryDraft "github.com/KirkDiggler/dnd-bot-discord/internal/repositories/character_draft"
	inmemoryChar "github.com/KirkDiggler/dnd-bot-discord/internal/repositories/characters"
)

// TestDraftSeparation demonstrates clean separation between Character and CharacterDraft
func TestDraftSeparation(t *testing.T) {
	ctx := context.Background()

	// Create repositories
	charRepo := inmemoryChar.NewInMemoryRepository()
	draftRepo := inmemoryDraft.NewInMemoryRepository()

	t.Run("draft and character are stored separately", func(t *testing.T) {
		// Create a character for draft
		char := &character.Character{
			ID:      "char-123",
			OwnerID: "user-456",
			RealmID: "realm-789",
			Name:    "",
			Status:  shared.CharacterStatusDraft,
			Level:   1,
		}

		// Create a draft wrapper
		draft := &character.CharacterDraft{
			ID:        "draft-123",
			OwnerID:   "user-456",
			Character: char,
			FlowState: &character.FlowState{
				CurrentStepID:  "race",
				AllSteps:       []string{"race", "class", "abilities", "name"},
				CompletedSteps: []string{},
			},
		}

		// Save draft
		err := draftRepo.Create(ctx, draft)
		require.NoError(t, err)

		// Character should not be in character repository yet
		_, err = charRepo.Get(ctx, char.ID)
		assert.Error(t, err, "character should not exist in character repo while in draft")

		// Draft should be retrievable
		retrieved, err := draftRepo.Get(ctx, draft.ID)
		require.NoError(t, err)
		assert.Equal(t, draft.ID, retrieved.ID)
		assert.Equal(t, char.ID, retrieved.Character.ID)
		assert.Equal(t, "race", retrieved.FlowState.CurrentStepID)
	})

	t.Run("finalization moves character from draft to active", func(t *testing.T) {
		// Create a completed draft
		char := &character.Character{
			ID:      "char-456",
			OwnerID: "user-456",
			RealmID: "realm-789",
			Name:    "Gandalf",
			Status:  shared.CharacterStatusDraft,
			Level:   1,
		}

		draft := &character.CharacterDraft{
			ID:        "draft-456",
			OwnerID:   "user-456",
			Character: char,
			FlowState: &character.FlowState{
				CurrentStepID:  "complete",
				AllSteps:       []string{"race", "class", "abilities", "name"},
				CompletedSteps: []string{"race", "class", "abilities", "name"},
			},
		}

		// Save draft
		err := draftRepo.Create(ctx, draft)
		require.NoError(t, err)

		// Simulate finalization
		// 1. Extract character and update status
		finalChar := draft.Character
		finalChar.Status = shared.CharacterStatusActive

		// 2. Save to character repository
		err = charRepo.Create(ctx, finalChar)
		require.NoError(t, err)

		// 3. Delete draft
		err = draftRepo.Delete(ctx, draft.ID)
		require.NoError(t, err)

		// Verify character is now active
		activeChar, err := charRepo.Get(ctx, char.ID)
		require.NoError(t, err)
		assert.Equal(t, shared.CharacterStatusActive, activeChar.Status)
		assert.Equal(t, "Gandalf", activeChar.Name)

		// Verify draft is deleted
		_, err = draftRepo.Get(ctx, draft.ID)
		assert.Error(t, err, "draft should be deleted after finalization")
	})

	t.Run("flow state is preserved in draft but not in character", func(t *testing.T) {
		// This test verifies that flow state stays with draft
		char := &character.Character{
			ID:      "char-789",
			OwnerID: "user-456",
			RealmID: "realm-789",
			Name:    "In Progress",
			Status:  shared.CharacterStatusDraft,
			Level:   1,
		}

		draft := &character.CharacterDraft{
			ID:        "draft-789",
			OwnerID:   "user-456",
			Character: char,
			FlowState: &character.FlowState{
				CurrentStepID:  "class",
				AllSteps:       []string{"race", "class", "abilities", "name"},
				CompletedSteps: []string{"race"},
				StepData: map[string]interface{}{
					"race_selected": "human",
				},
			},
		}

		// Save draft
		err := draftRepo.Create(ctx, draft)
		require.NoError(t, err)

		// Update flow state
		draft.FlowState.CompletedSteps = append(draft.FlowState.CompletedSteps, "class")
		draft.FlowState.CurrentStepID = "abilities"
		draft.FlowState.StepData["class_selected"] = "wizard"

		err = draftRepo.Update(ctx, draft)
		require.NoError(t, err)

		// Retrieve and verify
		updated, err := draftRepo.Get(ctx, draft.ID)
		require.NoError(t, err)

		// Flow state should be preserved
		assert.Equal(t, "abilities", updated.FlowState.CurrentStepID)
		assert.Contains(t, updated.FlowState.CompletedSteps, "race")
		assert.Contains(t, updated.FlowState.CompletedSteps, "class")
		assert.Equal(t, "wizard", updated.FlowState.StepData["class_selected"])

		// Character should not have flow state when finalized
		finalChar := updated.Character
		finalChar.Status = shared.CharacterStatusActive

		err = charRepo.Create(ctx, finalChar)
		require.NoError(t, err)

		// Retrieved character has no flow state
		activeChar, err := charRepo.Get(ctx, finalChar.ID)
		require.NoError(t, err)
		assert.Equal(t, shared.CharacterStatusActive, activeChar.Status)
		// Character struct doesn't have FlowState field - good separation!
	})
}
