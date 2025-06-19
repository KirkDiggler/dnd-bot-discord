package encounters_test

import (
	"context"
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/repositories/encounters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemoryRepository_GetActiveBySession(t *testing.T) {
	ctx := context.Background()
	repo := encounters.NewInMemoryRepository()

	sessionID := "test-session-123"

	t.Run("Returns nil when no encounters exist", func(t *testing.T) {
		// Should return nil, nil when no active encounter exists
		encounter, err := repo.GetActiveBySession(ctx, sessionID)
		assert.NoError(t, err)
		assert.Nil(t, encounter)
	})

	t.Run("Returns active encounter when exists", func(t *testing.T) {
		// Create an active encounter
		activeEnc := entities.NewEncounter("enc-1", sessionID, "channel-1", "Test Encounter", "user-1")
		activeEnc.Status = entities.EncounterStatusActive
		err := repo.Create(ctx, activeEnc)
		require.NoError(t, err)

		// Should return the active encounter
		encounter, err := repo.GetActiveBySession(ctx, sessionID)
		require.NoError(t, err)
		require.NotNil(t, encounter)
		assert.Equal(t, activeEnc.ID, encounter.ID)
		assert.Equal(t, entities.EncounterStatusActive, encounter.Status)
	})

	t.Run("Returns setup encounter when exists", func(t *testing.T) {
		// Create a new session
		sessionID2 := "test-session-456"
		
		// Create a setup encounter
		setupEnc := entities.NewEncounter("enc-2", sessionID2, "channel-2", "Setup Encounter", "user-2")
		setupEnc.Status = entities.EncounterStatusSetup
		err := repo.Create(ctx, setupEnc)
		require.NoError(t, err)

		// Should return the setup encounter
		encounter, err := repo.GetActiveBySession(ctx, sessionID2)
		require.NoError(t, err)
		require.NotNil(t, encounter)
		assert.Equal(t, setupEnc.ID, encounter.ID)
		assert.Equal(t, entities.EncounterStatusSetup, encounter.Status)
	})

	t.Run("Returns nil when only completed encounters exist", func(t *testing.T) {
		// Create a new session
		sessionID3 := "test-session-789"
		
		// Create a completed encounter
		completedEnc := entities.NewEncounter("enc-3", sessionID3, "channel-3", "Completed Encounter", "user-3")
		completedEnc.Status = entities.EncounterStatusCompleted
		err := repo.Create(ctx, completedEnc)
		require.NoError(t, err)

		// Should return nil when only completed encounters exist
		encounter, err := repo.GetActiveBySession(ctx, sessionID3)
		assert.NoError(t, err)
		assert.Nil(t, encounter)
	})
}

func TestInMemoryRepository_Create(t *testing.T) {
	ctx := context.Background()
	repo := encounters.NewInMemoryRepository()

	t.Run("Successfully creates encounter", func(t *testing.T) {
		encounter := entities.NewEncounter("enc-1", "session-1", "channel-1", "Test Encounter", "user-1")
		err := repo.Create(ctx, encounter)
		require.NoError(t, err)

		// Verify it can be retrieved
		retrieved, err := repo.Get(ctx, encounter.ID)
		require.NoError(t, err)
		assert.Equal(t, encounter.ID, retrieved.ID)
		assert.Equal(t, encounter.Name, retrieved.Name)
	})

	t.Run("Indexes by session", func(t *testing.T) {
		sessionID := "session-2"
		enc1 := entities.NewEncounter("enc-2", sessionID, "channel-2", "Encounter 1", "user-2")
		enc2 := entities.NewEncounter("enc-3", sessionID, "channel-2", "Encounter 2", "user-2")

		err := repo.Create(ctx, enc1)
		require.NoError(t, err)
		err = repo.Create(ctx, enc2)
		require.NoError(t, err)

		// Get all encounters for session
		encounters, err := repo.GetBySession(ctx, sessionID)
		require.NoError(t, err)
		assert.Len(t, encounters, 2)
	})
}