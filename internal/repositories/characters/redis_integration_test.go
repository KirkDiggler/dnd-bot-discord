//go:build integration
// +build integration

package characters_test

import (
	"context"
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/repositories/characters"
	"github.com/KirkDiggler/dnd-bot-discord/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisRepository_Integration(t *testing.T) {
	// This test requires Redis to be running
	client := testutils.CreateTestRedisClientOrSkip(t)

	repo := characters.NewRedisRepository(&characters.RedisRepoConfig{
		Client: client,
	})

	ctx := context.Background()

	t.Run("create and retrieve character", func(t *testing.T) {
		// Create a test character
		char := testutils.CreateTestCharacter("test-char-1", "user-123", "realm-456", "Aragorn")

		// Create the character
		err := repo.Create(ctx, char)
		require.NoError(t, err)

		// Retrieve the character
		retrieved, err := repo.Get(ctx, char.ID)
		require.NoError(t, err)
		assert.NotNil(t, retrieved)

		// Verify all fields are preserved
		assert.Equal(t, char.ID, retrieved.ID)
		assert.Equal(t, char.Name, retrieved.Name)
		assert.Equal(t, char.OwnerID, retrieved.OwnerID)
		assert.Equal(t, char.RealmID, retrieved.RealmID)
		assert.NotNil(t, retrieved.Race)
		assert.Equal(t, char.Race.Key, retrieved.Race.Key)
		assert.NotNil(t, retrieved.Class)
		assert.Equal(t, char.Class.Key, retrieved.Class.Key)
		assert.Len(t, retrieved.Attributes, 6)
		assert.Equal(t, char.MaxHitPoints, retrieved.MaxHitPoints)
		assert.Equal(t, char.AC, retrieved.AC)
	})

	t.Run("create duplicate character fails", func(t *testing.T) {
		char := testutils.CreateTestCharacter("test-char-2", "user-123", "realm-456", "Legolas")

		// Create the character
		err := repo.Create(ctx, char)
		require.NoError(t, err)

		// Try to create again
		err = repo.Create(ctx, char)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("update character", func(t *testing.T) {
		char := testutils.CreateTestCharacter("test-char-3", "user-123", "realm-456", "Gimli")

		// Create the character
		err := repo.Create(ctx, char)
		require.NoError(t, err)

		// Update the character
		char.Name = "Gimli Son of Gloin"
		char.Level = 5
		char.CurrentHitPoints = 20
		char.MaxHitPoints = 45

		err = repo.Update(ctx, char)
		require.NoError(t, err)

		// Retrieve and verify
		updated, err := repo.Get(ctx, char.ID)
		require.NoError(t, err)
		assert.Equal(t, "Gimli Son of Gloin", updated.Name)
		assert.Equal(t, 5, updated.Level)
		assert.Equal(t, 20, updated.CurrentHitPoints)
		assert.Equal(t, 45, updated.MaxHitPoints)
	})

	t.Run("get by owner", func(t *testing.T) {
		ownerID := "owner-test-1"

		// Create multiple characters for the owner
		char1 := testutils.CreateTestCharacter("owner-char-1", ownerID, "realm-456", "Frodo")
		char2 := testutils.CreateTestCharacter("owner-char-2", ownerID, "realm-456", "Sam")
		char3 := testutils.CreateTestCharacter("owner-char-3", "other-owner", "realm-456", "Merry")

		require.NoError(t, repo.Create(ctx, char1))
		require.NoError(t, repo.Create(ctx, char2))
		require.NoError(t, repo.Create(ctx, char3))

		// Get characters by owner
		chars, err := repo.GetByOwner(ctx, ownerID)
		require.NoError(t, err)
		assert.Len(t, chars, 2)

		// Verify we got the right characters
		names := []string{chars[0].Name, chars[1].Name}
		assert.Contains(t, names, "Frodo")
		assert.Contains(t, names, "Sam")
	})

	t.Run("get by owner and realm", func(t *testing.T) {
		ownerID := "owner-test-2"
		realm1 := "realm-1"
		realm2 := "realm-2"

		// Create characters in different realms
		char1 := testutils.CreateTestCharacter("realm-char-1", ownerID, realm1, "Gandalf")
		char2 := testutils.CreateTestCharacter("realm-char-2", ownerID, realm1, "Saruman")
		char3 := testutils.CreateTestCharacter("realm-char-3", ownerID, realm2, "Radagast")

		require.NoError(t, repo.Create(ctx, char1))
		require.NoError(t, repo.Create(ctx, char2))
		require.NoError(t, repo.Create(ctx, char3))

		// Get characters in realm1
		chars, err := repo.GetByOwnerAndRealm(ctx, ownerID, realm1)
		require.NoError(t, err)
		assert.Len(t, chars, 2)

		// Verify we got the right characters
		for _, char := range chars {
			assert.Equal(t, realm1, char.RealmID)
		}
	})

	t.Run("delete character", func(t *testing.T) {
		char := testutils.CreateTestCharacter("test-char-delete", "user-123", "realm-456", "Boromir")

		// Create the character
		err := repo.Create(ctx, char)
		require.NoError(t, err)

		// Verify it exists
		retrieved, err := repo.Get(ctx, char.ID)
		require.NoError(t, err)
		assert.NotNil(t, retrieved)

		// Delete the character
		err = repo.Delete(ctx, char.ID)
		require.NoError(t, err)

		// Verify it's gone
		_, err = repo.Get(ctx, char.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")

		// Verify it's removed from owner index
		chars, err := repo.GetByOwner(ctx, char.OwnerID)
		require.NoError(t, err)
		for _, c := range chars {
			assert.NotEqual(t, char.ID, c.ID)
		}
	})
}
