package character

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
	"strings"
	"testing"
)

func TestBuildFeatureSummary_RangerWithSelections(t *testing.T) {
	// Create a ranger character with favored enemy and natural explorer selections
	ranger := &character.Character{
		Name:  "Test Ranger",
		Level: 1,
		Features: []*rulebook.CharacterFeature{
			{
				Key:         "favored_enemy",
				Name:        "Favored Enemy",
				Type:        rulebook.FeatureTypeClass,
				Description: "You have significant experience studying, tracking, hunting, and even talking to a certain type of enemy.",
				Metadata: map[string]any{
					"enemy_type": "undead",
				},
			},
			{
				Key:         "natural_explorer",
				Name:        "Natural Explorer",
				Type:        rulebook.FeatureTypeClass,
				Description: "You are particularly familiar with one type of natural environment and are adept at traveling and surviving in such regions.",
				Metadata: map[string]any{
					"terrain_type": "forest",
				},
			},
		},
	}

	lines := buildFeatureSummary(ranger)

	// Check that the features show the selected values
	foundFavoredEnemy := false
	foundNaturalExplorer := false

	for _, line := range lines {
		if strings.Contains(line, "Favored Enemy (Undead)") {
			foundFavoredEnemy = true
		}
		if strings.Contains(line, "Natural Explorer (Forest)") {
			foundNaturalExplorer = true
		}
	}

	if !foundFavoredEnemy {
		t.Errorf("Expected to find 'Favored Enemy (Undead)' in feature summary, got: %v", lines)
	}

	if !foundNaturalExplorer {
		t.Errorf("Expected to find 'Natural Explorer (forest)' in feature summary, got: %v", lines)
	}
}

func TestBuildFeatureSummary_RangerWithoutSelections(t *testing.T) {
	// Create a ranger character without metadata (shouldn't crash)
	ranger := &character.Character{
		Name:  "Test Ranger",
		Level: 1,
		Features: []*rulebook.CharacterFeature{
			{
				Key:         "favored_enemy",
				Name:        "Favored Enemy",
				Type:        rulebook.FeatureTypeClass,
				Description: "You have significant experience studying, tracking, hunting, and even talking to a certain type of enemy.",
				// No metadata
			},
		},
	}

	lines := buildFeatureSummary(ranger)

	// Should still show the feature, just without the selection
	foundFavoredEnemy := false
	for _, line := range lines {
		if strings.Contains(line, "Favored Enemy") && !strings.Contains(line, "(") {
			foundFavoredEnemy = true
		}
	}

	if !foundFavoredEnemy {
		t.Errorf("Expected to find 'Favored Enemy' in feature summary, got: %v", lines)
	}
}

func TestBuildFeatureSummary_RangerHumanoids(t *testing.T) {
	// Test the special case for humanoids
	ranger := &character.Character{
		Name:  "Test Ranger",
		Level: 1,
		Features: []*rulebook.CharacterFeature{
			{
				Key:         "favored_enemy",
				Name:        "Favored Enemy",
				Type:        rulebook.FeatureTypeClass,
				Description: "You have significant experience studying, tracking, hunting, and even talking to a certain type of enemy.",
				Metadata: map[string]any{
					"enemy_type": "humanoids",
				},
			},
		},
	}

	lines := buildFeatureSummary(ranger)

	// Check that humanoids shows as "Two Humanoid Races"
	foundCorrectDisplay := false
	for _, line := range lines {
		if strings.Contains(line, "Favored Enemy (Two Humanoid Races)") {
			foundCorrectDisplay = true
		}
	}

	if !foundCorrectDisplay {
		t.Errorf("Expected to find 'Favored Enemy (Two Humanoid Races)' in feature summary, got: %v", lines)
	}
}
