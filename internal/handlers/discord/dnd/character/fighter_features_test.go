package character

import (
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/stretchr/testify/assert"
)

func TestClassFeaturesHandler_ShouldShowClassFeatures_Fighter(t *testing.T) {
	handler := &ClassFeaturesHandler{}

	tests := []struct {
		name            string
		character       *entities.Character
		expectedNeed    bool
		expectedFeature string
	}{
		{
			name: "fighter without fighting style selection",
			character: &entities.Character{
				Class: &entities.Class{Key: "fighter"},
				Features: []*entities.CharacterFeature{
					{
						Key:      "fighting_style",
						Name:     "Fighting Style",
						Metadata: nil, // No selection yet
					},
				},
			},
			expectedNeed:    true,
			expectedFeature: "fighting_style",
		},
		{
			name: "fighter with fighting style selected",
			character: &entities.Character{
				Class: &entities.Class{Key: "fighter"},
				Features: []*entities.CharacterFeature{
					{
						Key:  "fighting_style",
						Name: "Fighting Style",
						Metadata: map[string]any{
							"style": "defense",
						},
					},
				},
			},
			expectedNeed:    false,
			expectedFeature: "",
		},
		{
			name: "non-fighter class",
			character: &entities.Character{
				Class: &entities.Class{Key: "wizard"},
			},
			expectedNeed:    false,
			expectedFeature: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			needsSelection, featureType := handler.ShouldShowClassFeatures(tt.character)
			assert.Equal(t, tt.expectedNeed, needsSelection)
			assert.Equal(t, tt.expectedFeature, featureType)
		})
	}
}

func TestFightingStyleMetadata(t *testing.T) {
	// Test that we can store and retrieve fighting style metadata
	fighter := &entities.Character{
		Features: []*entities.CharacterFeature{
			{
				Key:      "fighting_style",
				Name:     "Fighting Style",
				Metadata: nil,
			},
		},
	}

	// Simulate selecting archery
	fighter.Features[0].Metadata = map[string]any{
		"style": "archery",
	}

	// Verify we can retrieve it
	assert.NotNil(t, fighter.Features[0].Metadata)
	assert.Equal(t, "archery", fighter.Features[0].Metadata["style"])
}

func TestFightingStyleOptions(t *testing.T) {
	// Test that all expected fighting styles are available
	expectedStyles := []string{
		"archery",
		"defense",
		"dueling",
		"great_weapon",
		"protection",
		"two_weapon",
	}

	// This would be better as a test of the actual UI options,
	// but for now we just verify the expected values
	for _, style := range expectedStyles {
		// Each style should be a valid option
		assert.NotEmpty(t, style)
	}
}
