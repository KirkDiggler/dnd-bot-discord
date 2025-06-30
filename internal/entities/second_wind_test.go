package entities

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFighterSecondWind_UsesAction(t *testing.T) {
	// Create a fighter character
	fighter := &Character{
		Name:  "Test Fighter",
		Level: 3,
		Class: &Class{
			Key:    "fighter",
			Name:   "Fighter",
			HitDie: 10,
		},
		MaxHitPoints:     28,
		CurrentHitPoints: 15, // Damaged
		Features: []*CharacterFeature{
			{Key: "second_wind", Name: "Second Wind"},
			{Key: "fighting_style", Name: "Fighting Style"},
		},
	}

	// Initialize resources
	fighter.InitializeResources()

	// Verify Second Wind ability exists and uses an Action
	abilities := fighter.GetResources().Abilities
	require.NotNil(t, abilities)

	secondWind, exists := abilities["second-wind"]
	require.True(t, exists, "Second Wind ability should exist")
	require.NotNil(t, secondWind)

	// The key assertion: Second Wind should use an Action, not Bonus Action
	assert.Equal(t, AbilityTypeAction, secondWind.ActionType,
		"Second Wind should use an Action, not a Bonus Action")

	// Verify other properties
	assert.Equal(t, "second-wind", secondWind.Key)
	assert.Equal(t, "Second Wind", secondWind.Name)
	assert.Equal(t, 1, secondWind.UsesMax)
	assert.Equal(t, 1, secondWind.UsesRemaining)
	assert.Equal(t, RestTypeShort, secondWind.RestType)
	assert.Equal(t, 0, secondWind.Duration, "Second Wind is instant")
}

func TestFighterSecondWind_ActionEconomy(t *testing.T) {
	// Create a fighter
	fighter := &Character{
		Name:  "Test Fighter",
		Level: 1,
		Class: &Class{
			Key:    "fighter",
			Name:   "Fighter",
			HitDie: 10,
		},
		Resources: &CharacterResources{},
	}

	// Initialize resources
	fighter.InitializeResources()

	// Start a new turn
	fighter.StartNewTurn()

	// Fighter should have action available
	assert.True(t, fighter.HasActionAvailable(), "Fighter should have action available")
	assert.False(t, fighter.Resources.ActionEconomy.ActionUsed, "Action should not be used yet")

	// Using Second Wind should consume the action
	// Note: We're not actually calling the ability service here, just testing that
	// the ability is configured to use an action
	secondWind := fighter.Resources.Abilities["second-wind"]
	require.NotNil(t, secondWind)

	// Simulate using Second Wind (which requires an action)
	if secondWind.ActionType == AbilityTypeAction {
		fighter.Resources.ActionEconomy.ActionUsed = true
	}

	assert.True(t, fighter.Resources.ActionEconomy.ActionUsed, "Action should be used after Second Wind")
	assert.False(t, fighter.HasActionAvailable(), "Fighter should not have action available after Second Wind")
}

func TestFighterSecondWind_NotBonusAction(t *testing.T) {
	// This test explicitly verifies Second Wind is NOT a bonus action
	fighter := &Character{
		Name:  "Test Fighter",
		Level: 5,
		Class: &Class{
			Key:    "fighter",
			Name:   "Fighter",
			HitDie: 10,
		},
	}

	fighter.InitializeResources()

	secondWind := fighter.Resources.Abilities["second-wind"]
	require.NotNil(t, secondWind)

	// Explicitly check it's not a bonus action
	assert.NotEqual(t, AbilityTypeBonusAction, secondWind.ActionType,
		"Second Wind should NOT be a bonus action")
}
