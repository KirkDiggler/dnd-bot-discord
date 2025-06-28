package features

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetClassFeatures_Ranger(t *testing.T) {
	tests := []struct {
		name     string
		level    int
		expected []string
	}{
		{
			name:  "Level 1 Ranger",
			level: 1,
			expected: []string{
				"favored_enemy",
				"natural_explorer",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			features := GetClassFeatures("ranger", tt.level)

			// Check we got the right number of features
			assert.Len(t, features, len(tt.expected), "Should have correct number of features")

			// Check each expected feature is present
			for _, expectedKey := range tt.expected {
				assert.True(t, HasFeature(features, expectedKey), "Should have feature: %s", expectedKey)
			}

			// Verify specific feature details
			for _, feature := range features {
				switch feature.Key {
				case "favored_enemy":
					assert.Equal(t, "Favored Enemy", feature.Name)
					assert.Contains(t, feature.Description, "significant experience studying")
					assert.Contains(t, feature.Description, "advantage on Wisdom (Survival) checks")
				case "natural_explorer":
					assert.Equal(t, "Natural Explorer", feature.Name)
					assert.Contains(t, feature.Description, "familiar with one type of natural environment")
					assert.Contains(t, feature.Description, "difficult terrain doesn't slow")
				}
			}
		})
	}
}

func TestRangerFeatureDescriptions(t *testing.T) {
	features := GetClassFeatures("ranger", 1)

	// Verify Favored Enemy includes all enemy types
	favoredEnemy := features[0]
	enemyTypes := []string{
		"aberrations", "beasts", "celestials", "constructs", "dragons",
		"elementals", "fey", "fiends", "giants", "monstrosities",
		"oozes", "plants", "undead", "humanoid",
	}

	for _, enemyType := range enemyTypes {
		assert.Contains(t, favoredEnemy.Description, enemyType,
			"Favored Enemy description should include %s", enemyType)
	}

	// Verify Natural Explorer includes all terrain types
	naturalExplorer := features[1]
	terrainTypes := []string{
		"arctic", "coast", "desert", "forest",
		"grassland", "mountain", "swamp", "Underdark",
	}

	for _, terrain := range terrainTypes {
		assert.Contains(t, naturalExplorer.Description, terrain,
			"Natural Explorer description should include %s", terrain)
	}
}
