package character

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
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
					"domain": "life",
				},
			},
		},
	}

	lines := buildFeatureSummary(cleric)

	// Check that the feature shows the selected domain
	foundDivineDomain := false
	for _, line := range lines {
		if strings.Contains(line, "Divine Domain (Life)") {
			foundDivineDomain = true
		}
	}

	if !foundDivineDomain {
		t.Errorf("Expected to find 'Divine Domain (Life)' in feature summary, got: %v", lines)
	}
}

func TestBuildFeatureSummary_ClericAllDomains(t *testing.T) {
	// Test all domain names are properly capitalized
	domains := map[string]string{
		"knowledge": "Knowledge",
		"life":      "Life",
		"light":     "Light",
		"nature":    "Nature",
		"tempest":   "Tempest",
		"trickery":  "Trickery",
		"war":       "War",
	}

	for domainKey, expectedDisplay := range domains {
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
						"domain": domainKey,
					},
				},
			},
		}

		lines := buildFeatureSummary(cleric)

		// Check that the feature shows the properly formatted domain
		expectedText := "Divine Domain (" + expectedDisplay + ")"
		foundDomain := false
		for _, line := range lines {
			if strings.Contains(line, expectedText) {
				foundDomain = true
				break
			}
		}

		if !foundDomain {
			t.Errorf("Expected to find '%s' in feature summary for domain '%s', got: %v",
				expectedText, domainKey, lines)
		}
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
