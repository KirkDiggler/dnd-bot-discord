//go:build integration
// +build integration

package monster_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/monster"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMonsterService_GetMonstersByCR_Integration(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Create real D&D client
	client, err := dnd5e.New(&dnd5e.Config{
		HttpClient: &http.Client{},
	})
	require.NoError(t, err)

	// Create monster service with real client
	service := monster.NewService(&monster.ServiceConfig{
		DNDClient: client,
	})

	ctx := context.Background()

	// Test getting CR 0.25 monsters (should include goblins)
	monsters, err := service.GetMonstersByCR(ctx, 0.25, 0.25)
	require.NoError(t, err)
	assert.NotEmpty(t, monsters)

	// Check if we got some expected monsters
	foundGoblin := false
	for _, m := range monsters {
		if m.Key == "goblin" {
			foundGoblin = true
			break
		}
	}
	assert.True(t, foundGoblin, "Expected to find goblin in CR 0.25 monsters")

	// Test getting CR 1 monsters
	monsters, err = service.GetMonstersByCR(ctx, 1, 1)
	require.NoError(t, err)
	assert.NotEmpty(t, monsters)

	// Check if we got some expected monsters
	foundBugbear := false
	for _, m := range monsters {
		if m.Key == "bugbear" {
			foundBugbear = true
			break
		}
	}
	assert.True(t, foundBugbear, "Expected to find bugbear in CR 1 monsters")
}

func TestMonsterService_GetRandomMonsters_Integration(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Create real D&D client
	client, err := dnd5e.New(&dnd5e.Config{
		HttpClient: &http.Client{},
	})
	require.NoError(t, err)

	// Create monster service with real client
	service := monster.NewService(&monster.ServiceConfig{
		DNDClient: client,
	})

	ctx := context.Background()

	// Test getting random easy monsters
	monsters, err := service.GetRandomMonsters(ctx, "easy", 3)
	require.NoError(t, err)
	assert.Len(t, monsters, 3)

	// All should be low CR monsters
	for _, m := range monsters {
		assert.NotEmpty(t, m.Key)
		assert.NotEmpty(t, m.Name)
	}

	// Test getting random hard monsters
	monsters, err = service.GetRandomMonsters(ctx, "hard", 5)
	require.NoError(t, err)
	assert.Len(t, monsters, 5)
}
