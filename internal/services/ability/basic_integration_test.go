//go:build integration
// +build integration

package ability

import (
	"context"
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/repositories/characters"
	"github.com/KirkDiggler/dnd-bot-discord/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasicRedisIntegration(t *testing.T) {
	// Setup Redis
	redisClient := testutils.CreateTestRedisClientOrSkip(t)

	// Create repository
	charRepo := characters.NewRedis(redisClient)

	// Create a simple character
	char := &entities.Character{
		ID:               "test_123",
		OwnerID:          "player_123",
		RealmID:          "realm_123",
		Name:             "TestChar",
		Class:            &entities.Class{Key: "fighter", Name: "Fighter"},
		Level:            1,
		Status:           entities.CharacterStatusActive,
		CurrentHitPoints: 10,
		MaxHitPoints:     10,
		Resources: &entities.CharacterResources{
			Abilities: map[string]*entities.ActiveAbility{
				"test_ability": {
					Name:          "Test Ability",
					Key:           "test_ability",
					ActionType:    entities.AbilityTypeAction,
					UsesMax:       3,
					UsesRemaining: 3,
					RestType:      entities.RestTypeLong,
				},
			},
		},
	}

	ctx := context.Background()

	// Test Create
	err := charRepo.Create(ctx, char)
	require.NoError(t, err, "Failed to create character")

	// Test Get
	loadedChar, err := charRepo.Get(ctx, char.ID)
	require.NoError(t, err, "Failed to get character")

	assert.Equal(t, char.Name, loadedChar.Name)
	assert.Equal(t, char.Level, loadedChar.Level)
	assert.NotNil(t, loadedChar.Resources)
	assert.NotNil(t, loadedChar.Resources.Abilities["test_ability"])
	assert.Equal(t, 3, loadedChar.Resources.Abilities["test_ability"].UsesRemaining)

	// Test Update
	loadedChar.Resources.Abilities["test_ability"].UsesRemaining = 2
	err = charRepo.Update(ctx, loadedChar)
	require.NoError(t, err, "Failed to update character")

	// Verify update
	updatedChar, err := charRepo.Get(ctx, char.ID)
	require.NoError(t, err, "Failed to get updated character")
	assert.Equal(t, 2, updatedChar.Resources.Abilities["test_ability"].UsesRemaining)

	// Test persistence of Resources field
	assert.NotNil(t, updatedChar.Resources, "Resources should not be nil after save/load")
	assert.NotEmpty(t, updatedChar.Resources.Abilities, "Abilities should not be empty after save/load")
}
