package character_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	mockdnd "github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e/mock"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	mockdraft "github.com/KirkDiggler/dnd-bot-discord/internal/repositories/character_draft/mock"
	mockchar "github.com/KirkDiggler/dnd-bot-discord/internal/repositories/characters/mock"
)

// Test shows how draft repository would integrate with character service
func TestCharacterService_WithDraftRepository(t *testing.T) {
	// This test demonstrates the contract between service and repositories
	// Without modifying the actual service implementation yet

	setup := func(t *testing.T) (
		*mockchar.MockRepository,
		*mockdraft.MockRepository,
		*mockdnd.MockClient,
		context.Context,
	) {
		ctrl := gomock.NewController(t)
		t.Cleanup(ctrl.Finish)

		return mockchar.NewMockRepository(ctrl),
			mockdraft.NewMockRepository(ctrl),
			mockdnd.NewMockClient(ctrl),
			context.Background()
	}

	t.Run("GetOrCreateDraftCharacter workflow", func(t *testing.T) {
		mockCharRepo, mockDraftRepo, _, ctx := setup(t)
		userID := "user-123"
		realmID := "realm-456"

		// Current implementation: checks character repo for draft status
		mockCharRepo.EXPECT().
			GetByOwnerAndRealm(ctx, userID, realmID).
			Return([]*character.Character{}, nil) // No characters

		// Future: would check draft repo instead
		mockDraftRepo.EXPECT().
			GetByOwnerAndRealm(ctx, userID, realmID).
			Return(nil, nil) // No draft exists

		// Future: create draft instead of character
		mockDraftRepo.EXPECT().
			Create(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, draft *character.CharacterDraft) error {
				// Validate draft structure
				assert.Equal(t, userID, draft.OwnerID)
				assert.NotNil(t, draft.Character)
				assert.Equal(t, userID, draft.Character.OwnerID)
				assert.Equal(t, realmID, draft.Character.RealmID)
				assert.Equal(t, shared.CharacterStatusDraft, draft.Character.Status)

				// Draft should have flow state
				assert.NotNil(t, draft.FlowState)
				assert.Equal(t, "race", draft.FlowState.CurrentStepID)
				assert.Empty(t, draft.FlowState.CompletedSteps)

				return nil
			})

		// Simulate the workflow
		// 1. Check for existing drafts in character repo (legacy)
		chars, err := mockCharRepo.GetByOwnerAndRealm(ctx, userID, realmID)
		require.NoError(t, err)
		assert.Empty(t, chars)

		// 2. Check draft repo (new approach)
		existingDraft, err := mockDraftRepo.GetByOwnerAndRealm(ctx, userID, realmID)
		require.NoError(t, err)
		assert.Nil(t, existingDraft)

		// 3. Create new draft
		newDraft := &character.CharacterDraft{
			ID:      "draft-123",
			OwnerID: userID,
			Character: &character.Character{
				ID:      "char-123",
				OwnerID: userID,
				RealmID: realmID,
				Status:  shared.CharacterStatusDraft,
			},
			FlowState: &character.FlowState{
				CurrentStepID:  "race",
				AllSteps:       []string{"race", "class", "abilities", "equipment", "name"},
				CompletedSteps: []string{},
			},
		}
		err = mockDraftRepo.Create(ctx, newDraft)
		require.NoError(t, err)
	})

	t.Run("UpdateDraftCharacter workflow", func(t *testing.T) {
		_, mockDraftRepo, mockDND, ctx := setup(t)
		charID := "char-123"

		existingDraft := &character.CharacterDraft{
			ID:      "draft-123",
			OwnerID: "user-123",
			Character: &character.Character{
				ID:      charID,
				OwnerID: "user-123",
				Status:  shared.CharacterStatusDraft,
			},
			FlowState: &character.FlowState{
				CurrentStepID:  "race",
				CompletedSteps: []string{},
			},
		}

		// Current: service would get character from char repo
		// Future: get draft by character ID
		mockDraftRepo.EXPECT().
			Get(ctx, gomock.Any()).
			Return(existingDraft, nil)

		// Service would fetch race data
		mockDND.EXPECT().
			GetRace("human").
			Return(&rulebook.Race{Key: "human", Name: "Human"}, nil)

		// Update the draft
		mockDraftRepo.EXPECT().
			Update(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, draft *character.CharacterDraft) error {
				assert.Equal(t, "human", draft.Character.Race.Key)
				assert.Contains(t, draft.FlowState.CompletedSteps, "race")
				assert.Equal(t, "class", draft.FlowState.CurrentStepID)
				return nil
			})

		// Simulate update workflow
		draft, err := mockDraftRepo.Get(ctx, "draft-123")
		require.NoError(t, err)

		// Apply updates
		race, err := mockDND.GetRace("human")
		require.NoError(t, err)
		draft.Character.Race = race
		draft.FlowState.CompletedSteps = append(draft.FlowState.CompletedSteps, "race")
		draft.FlowState.CurrentStepID = "class"

		err = mockDraftRepo.Update(ctx, draft)
		require.NoError(t, err)
	})

	t.Run("FinalizeDraftCharacter workflow", func(t *testing.T) {
		mockCharRepo, mockDraftRepo, _, ctx := setup(t)
		draftID := "draft-123"
		charID := "char-123"

		completedDraft := &character.CharacterDraft{
			ID:      draftID,
			OwnerID: "user-123",
			Character: &character.Character{
				ID:     charID,
				Name:   "Gandalf",
				Race:   &rulebook.Race{Key: "human", Name: "Human"},
				Class:  &rulebook.Class{Key: "wizard", Name: "Wizard"},
				Level:  1,
				Status: shared.CharacterStatusDraft,
				// All required fields populated
			},
			FlowState: &character.FlowState{
				CurrentStepID:  "complete",
				CompletedSteps: []string{"race", "class", "abilities", "equipment", "name"},
			},
		}

		// Get the completed draft
		mockDraftRepo.EXPECT().
			Get(ctx, draftID).
			Return(completedDraft, nil)

		// Create active character
		mockCharRepo.EXPECT().
			Create(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, char *character.Character) error {
				// Verify finalized character
				assert.Equal(t, charID, char.ID)
				assert.Equal(t, shared.CharacterStatusActive, char.Status)
				assert.Equal(t, "Gandalf", char.Name)
				assert.NotNil(t, char.Race)
				assert.NotNil(t, char.Class)

				// Character should not have flow state - that stays with draft

				return nil
			})

		// Delete the draft
		mockDraftRepo.EXPECT().
			Delete(ctx, draftID).
			Return(nil)

		// Simulate finalization
		draft, err := mockDraftRepo.Get(ctx, draftID)
		require.NoError(t, err)

		// Extract and finalize character
		finalChar := draft.Character
		finalChar.Status = shared.CharacterStatusActive

		err = mockCharRepo.Create(ctx, finalChar)
		require.NoError(t, err)

		err = mockDraftRepo.Delete(ctx, draftID)
		require.NoError(t, err)
	})

	t.Run("list characters excludes drafts when using draft repo", func(t *testing.T) {
		mockCharRepo, _, _, ctx := setup(t)
		userID := "user-123"

		// Only return active characters
		activeChars := []*character.Character{
			{
				ID:     "char-1",
				Name:   "Aragorn",
				Status: shared.CharacterStatusActive,
			},
			{
				ID:     "char-2",
				Name:   "Legolas",
				Status: shared.CharacterStatusActive,
			},
		}

		mockCharRepo.EXPECT().
			GetByOwner(ctx, userID).
			Return(activeChars, nil)

		// List should only show active characters
		chars, err := mockCharRepo.GetByOwner(ctx, userID)
		require.NoError(t, err)
		assert.Len(t, chars, 2)

		// All returned characters should be active
		for _, char := range chars {
			assert.Equal(t, shared.CharacterStatusActive, char.Status)
		}
	})
}
