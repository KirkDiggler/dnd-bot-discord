//go:build integration
// +build integration

package ability

import (
	"context"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"testing"

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
	char := &character.Character{
		ID:               "test_123",
		OwnerID:          "player_123",
		RealmID:          "realm_123",
		Name:             "TestChar",
		Class:            &rulebook.Class{Key: "fighter", Name: "Fighter"},
		Level:            1,
		Status:           shared.CharacterStatusActive,
		CurrentHitPoints: 10,
		MaxHitPoints:     10,
		Resources: &shared.CharacterResources{
			Abilities: map[string]*shared.ActiveAbility{
				"test_ability": {
					Name:          "Test Ability",
					Key:           "test_ability",
					ActionType:    shared.AbilityTypeAction,
					UsesMax:       3,
					UsesRemaining: 3,
					RestType:      shared.RestTypeLong,
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
