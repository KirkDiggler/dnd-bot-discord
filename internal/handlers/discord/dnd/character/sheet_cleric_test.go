package character

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	rulebook "github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
	"strings"
	"testing"
)

func TestBuildFeatureSummary_ClericWithDivineDomain(t *testing.T) {
	// Create a cleric character with divine domain selection
	cleric := &character.Character{
		Name:  "Test Cleric",
		Level: 1,
		Features: []*rulebook.CharacterFeature{
			{
				Key:         "divine_domain",
				Name:        "Divine Domain",
				Type:        rulebook.FeatureTypeClass,
				Description: "Choose one domain related to your deity.",
				Metadata: map[string]any{
					"domain":            "life",
					"selection_display": "Life Domain",
				},
			},
		},
	}

	lines := buildFeatureSummary(cleric)

	// Check that the feature shows the selected domain
	foundDivineDomain := false
	for _, line := range lines {
		if strings.Contains(line, "Divine Domain (Life Domain)") {
			foundDivineDomain = true
		}
	}

	if !foundDivineDomain {
		t.Errorf("Expected to find 'Divine Domain (Life Domain)' in feature summary, got: %v", lines)
	}
}

func TestBuildFeatureSummary_WithSelectionDisplay(t *testing.T) {
	// Test that the handler properly renders whatever is in selection_display
	char := &character.Character{
		Name:  "Test Character",
		Level: 1,
		Features: []*rulebook.CharacterFeature{
			{
				Key:         "some_feature",
				Name:        "Some Feature",
				Type:        rulebook.FeatureTypeClass,
				Description: "A feature with a selection",
				Metadata: map[string]any{
					"selection_key":     "some_key",
					"selection_display": "Display Name",
				},
			},
		},
	}

	lines := buildFeatureSummary(char)

	// Check that the feature shows the selection display
	foundFeature := false
	for _, line := range lines {
		if strings.Contains(line, "Some Feature (Display Name)") {
			foundFeature = true
			break
		}
	}

	if !foundFeature {
		t.Errorf("Expected to find 'Some Feature (Display Name)' in feature summary, got: %v", lines)
	}
}

func TestBuildFeatureSummary_ClericWithoutDomain(t *testing.T) {
	// Create a cleric character without domain metadata (shouldn't crash)
	cleric := &character.Character{
		Name:  "Test Cleric",
		Level: 1,
		Features: []*rulebook.CharacterFeature{
			{
				Key:         "divine_domain",
				Name:        "Divine Domain",
				Type:        rulebook.FeatureTypeClass,
				Description: "Choose one domain related to your deity.",
				// No metadata
			},
		},
	}

	lines := buildFeatureSummary(cleric)

	// Should still show the feature, just without the selection
	foundDivineDomain := false
	for _, line := range lines {
		if strings.Contains(line, "Divine Domain") && !strings.Contains(line, "(") {
			foundDivineDomain = true
		}
	}

	if !foundDivineDomain {
		t.Errorf("Expected to find 'Divine Domain' in feature summary, got: %v", lines)
	}
}
