package character_draft_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/repositories/character_draft"
)

func TestInMemoryRepository_Create(t *testing.T) {
	setup := func(t *testing.T) (character_draft.Repository, context.Context) {
		t.Helper()
		return character_draft.NewInMemoryRepository(), context.Background()
	}

	setupDraft := func(id, ownerID, realmID string) *character.CharacterDraft {
		return &character.CharacterDraft{
			ID:        id,
			OwnerID:   ownerID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Character: &character.Character{
				ID:      "char-" + id,
				OwnerID: ownerID,
				RealmID: realmID,
			},
			FlowState: &character.FlowState{
				CurrentStepID:  "race",
				AllSteps:       []string{"race", "class", "abilities"},
				CompletedSteps: []string{},
			},
		}
	}

	t.Run("creates new draft successfully", func(t *testing.T) {
		repo, ctx := setup(t)
		draft := setupDraft("123", "user-123", "realm-123")

		err := repo.Create(ctx, draft)
		require.NoError(t, err)

		// Verify we can retrieve it
		retrieved, err := repo.Get(ctx, draft.ID)
		require.NoError(t, err)
		assert.Equal(t, draft.ID, retrieved.ID)
		assert.Equal(t, draft.OwnerID, retrieved.OwnerID)
	})

	t.Run("returns error for nil draft", func(t *testing.T) {
		repo, ctx := setup(t)

		err := repo.Create(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "draft cannot be nil")
	})

	t.Run("returns error for empty ID", func(t *testing.T) {
		repo, ctx := setup(t)
		draft := &character.CharacterDraft{ID: ""}

		err := repo.Create(ctx, draft)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "draft ID is required")
	})

	t.Run("returns error for duplicate ID", func(t *testing.T) {
		repo, ctx := setup(t)
		draft := setupDraft("duplicate", "user-123", "realm-123")

		err := repo.Create(ctx, draft)
		require.NoError(t, err)

		// Try to create again with same ID
		err = repo.Create(ctx, draft)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})
}

func TestInMemoryRepository_Get(t *testing.T) {
	ctx := context.Background()
	repo := character_draft.NewInMemoryRepository()

	// Create a test draft
	draft := &character.CharacterDraft{
		ID:      "draft-123",
		OwnerID: "user-123",
		Character: &character.Character{
			ID:      "char-123",
			RealmID: "realm-123",
		},
	}
	require.NoError(t, repo.Create(ctx, draft))

	t.Run("retrieves existing draft", func(t *testing.T) {
		retrieved, err := repo.Get(ctx, "draft-123")
		require.NoError(t, err)
		assert.Equal(t, draft.ID, retrieved.ID)
		assert.Equal(t, draft.OwnerID, retrieved.OwnerID)
	})

	t.Run("returns error for non-existent draft", func(t *testing.T) {
		_, err := repo.Get(ctx, "non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("returns error for empty ID", func(t *testing.T) {
		_, err := repo.Get(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "draft ID is required")
	})
}

func TestInMemoryRepository_GetByOwnerAndRealm(t *testing.T) {
	ctx := context.Background()
	repo := character_draft.NewInMemoryRepository()

	// Create test drafts
	draft1 := &character.CharacterDraft{
		ID:      "draft-1",
		OwnerID: "user-123",
		Character: &character.Character{
			ID:      "char-1",
			RealmID: "realm-123",
		},
	}
	draft2 := &character.CharacterDraft{
		ID:      "draft-2",
		OwnerID: "user-456",
		Character: &character.Character{
			ID:      "char-2",
			RealmID: "realm-123",
		},
	}
	draft3 := &character.CharacterDraft{
		ID:      "draft-3",
		OwnerID: "user-123",
		Character: &character.Character{
			ID:      "char-3",
			RealmID: "realm-456",
		},
	}

	require.NoError(t, repo.Create(ctx, draft1))
	require.NoError(t, repo.Create(ctx, draft2))
	require.NoError(t, repo.Create(ctx, draft3))

	t.Run("finds draft by owner and realm", func(t *testing.T) {
		found, err := repo.GetByOwnerAndRealm(ctx, "user-123", "realm-123")
		require.NoError(t, err)
		assert.Equal(t, draft1.ID, found.ID)
	})

	t.Run("returns not found for non-existent combination", func(t *testing.T) {
		_, err := repo.GetByOwnerAndRealm(ctx, "user-999", "realm-999")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no draft found")
	})

	t.Run("returns error for empty owner ID", func(t *testing.T) {
		_, err := repo.GetByOwnerAndRealm(ctx, "", "realm-123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "owner ID is required")
	})
}

func TestInMemoryRepository_Update(t *testing.T) {
	ctx := context.Background()
	repo := character_draft.NewInMemoryRepository()

	// Create initial draft
	draft := &character.CharacterDraft{
		ID:      "draft-123",
		OwnerID: "user-123",
		Character: &character.Character{
			ID:   "char-123",
			Name: "Initial Name",
		},
		FlowState: &character.FlowState{
			CurrentStepID:  "race",
			CompletedSteps: []string{},
		},
	}
	require.NoError(t, repo.Create(ctx, draft))

	t.Run("updates existing draft", func(t *testing.T) {
		// Update the draft
		draft.Character.Name = "Updated Name"
		draft.FlowState.CurrentStepID = "class"
		draft.FlowState.CompletedSteps = []string{"race"}

		err := repo.Update(ctx, draft)
		require.NoError(t, err)

		// Verify updates
		retrieved, err := repo.Get(ctx, draft.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", retrieved.Character.Name)
		assert.Equal(t, "class", retrieved.FlowState.CurrentStepID)
		assert.Equal(t, []string{"race"}, retrieved.FlowState.CompletedSteps)
	})

	t.Run("returns error for non-existent draft", func(t *testing.T) {
		nonExistent := &character.CharacterDraft{
			ID: "non-existent",
		}
		err := repo.Update(ctx, nonExistent)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("returns error for nil draft", func(t *testing.T) {
		err := repo.Update(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "draft cannot be nil")
	})

	t.Run("returns error for empty ID", func(t *testing.T) {
		invalidDraft := &character.CharacterDraft{
			ID: "",
		}
		err := repo.Update(ctx, invalidDraft)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "draft ID is required")
	})
}

func TestInMemoryRepository_Delete(t *testing.T) {
	ctx := context.Background()
	repo := character_draft.NewInMemoryRepository()

	// Create a draft to delete
	draft := &character.CharacterDraft{
		ID:      "draft-to-delete",
		OwnerID: "user-123",
	}
	require.NoError(t, repo.Create(ctx, draft))

	t.Run("deletes existing draft", func(t *testing.T) {
		err := repo.Delete(ctx, draft.ID)
		require.NoError(t, err)

		// Verify it's gone
		_, err = repo.Get(ctx, draft.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("returns error for non-existent draft", func(t *testing.T) {
		err := repo.Delete(ctx, "non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("returns error for empty ID", func(t *testing.T) {
		err := repo.Delete(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "draft ID is required")
	})
}
