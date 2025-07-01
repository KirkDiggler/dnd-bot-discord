//go:build integration
// +build integration

package dnd5e_test

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
	"net/http"
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_ListEquipment_Integration(t *testing.T) {
	// This test requires network access to the D&D 5e API
	client, err := dnd5e.New(&dnd5e.Config{
		HttpClient: http.DefaultClient,
	})
	require.NoError(t, err)

	equipment, err := client.ListEquipment()
	require.NoError(t, err)

	// The API should have many equipment items
	assert.NotEmpty(t, equipment, "ListEquipment should return equipment")
	assert.Greater(t, len(equipment), 50, "API should have many equipment items")

	// Verify some equipment properties
	foundWeapon := false
	foundArmor := false
	for _, equip := range equipment {
		assert.NotEmpty(t, equip.GetKey(), "Equipment should have a key")
		assert.NotEmpty(t, equip.GetName(), "Equipment should have a name")

		if equip.GetEquipmentType() == "Weapon" {
			foundWeapon = true
		}
		if equip.GetEquipmentType() == "Armor" {
			foundArmor = true
		}
	}

	assert.True(t, foundWeapon, "Should find at least one weapon")
	assert.True(t, foundArmor, "Should find at least one armor")
}

func TestClient_ListMonstersByCR_Integration(t *testing.T) {
	// This test requires network access to the D&D 5e API
	client, err := dnd5e.New(&dnd5e.Config{
		HttpClient: http.DefaultClient,
	})
	require.NoError(t, err)

	// Test getting low CR monsters (0 to 1)
	lowCRMonsters, err := client.ListMonstersByCR(0, 1)
	require.NoError(t, err)
	assert.NotEmpty(t, lowCRMonsters, "Should find low CR monsters")

	// Verify all returned monsters are in the CR range
	for _, monster := range lowCRMonsters {
		assert.LessOrEqual(t, monster.ChallengeRating, float32(1), "Monster CR should be <= 1")
		assert.GreaterOrEqual(t, monster.ChallengeRating, float32(0), "Monster CR should be >= 0")
		assert.NotEmpty(t, monster.Name, "Monster should have a name")
		assert.NotEmpty(t, monster.Key, "Monster should have a key")
	}

	// Test getting a single CR value (more efficient for testing)
	cr2Monsters, err := client.ListMonstersByCR(2, 2)
	require.NoError(t, err)
	assert.NotEmpty(t, cr2Monsters, "Should find CR 2 monsters")

	// Verify all returned monsters have exactly CR 2
	for _, monster := range cr2Monsters {
		assert.Equal(t, float32(2), monster.ChallengeRating, "Monster CR should be exactly 2")
	}
}

func TestClient_GetClassFeatures_Integration(t *testing.T) {
	// This test requires network access to the D&D 5e API
	client, err := dnd5e.New(&dnd5e.Config{
		HttpClient: http.DefaultClient,
	})
	require.NoError(t, err)

	// Test getting fighter features at level 1
	features, err := client.GetClassFeatures("fighter", 1)
	require.NoError(t, err)
	assert.NotEmpty(t, features, "Fighter should have features at level 1")

	// Verify feature properties
	for _, feature := range features {
		assert.NotEmpty(t, feature.Key, "Feature should have a key")
		assert.NotEmpty(t, feature.Name, "Feature should have a name")
		assert.Equal(t, rulebook.FeatureTypeClass, feature.Type)
		assert.Equal(t, 1, feature.Level)
		assert.Equal(t, "fighter", feature.Source)
	}

	// Test getting wizard features at level 1
	wizardFeatures, err := client.GetClassFeatures("wizard", 1)
	require.NoError(t, err)
	assert.NotEmpty(t, wizardFeatures, "Wizard should have features at level 1")

	// Different classes should have different features
	assert.NotEqual(t, features[0].Key, wizardFeatures[0].Key, "Fighter and wizard should have different features")
}
