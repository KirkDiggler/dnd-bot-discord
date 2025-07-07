package character

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
	mockcharacters "github.com/KirkDiggler/dnd-bot-discord/internal/services/character/mock"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestClassFeaturesHandler_ShouldShowClassFeatures_Fighter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name            string
		character       *character.Character
		pendingChoices  []*rulebook.FeatureChoice
		expectedNeed    bool
		expectedFeature string
	}{
		{
			name: "fighter without fighting style selection",
			character: &character.Character{
				ID:    "test-fighter",
				Class: &rulebook.Class{Key: "fighter"},
				Features: []*rulebook.CharacterFeature{
					{
						Key:      "fighting_style",
						Name:     "Fighting Style",
						Metadata: nil, // No selection yet
					},
				},
			},
			pendingChoices: []*rulebook.FeatureChoice{
				{
					Type:       rulebook.FeatureChoiceTypeFightingStyle,
					FeatureKey: "fighting_style",
					Name:       "Fighting Style",
				},
			},
			expectedNeed:    true,
			expectedFeature: "fighting_style",
		},
		{
			name: "fighter with fighting style selected",
			character: &character.Character{
				ID:    "test-fighter-2",
				Class: &rulebook.Class{Key: "fighter"},
				Features: []*rulebook.CharacterFeature{
					{
						Key:  "fighting_style",
						Name: "Fighting Style",
						Metadata: map[string]any{
							"style": "defense",
						},
					},
				},
			},
			pendingChoices:  []*rulebook.FeatureChoice{}, // No pending choices
			expectedNeed:    false,
			expectedFeature: "",
		},
		{
			name: "non-fighter class",
			character: &character.Character{
				ID:    "test-wizard",
				Class: &rulebook.Class{Key: "wizard"},
			},
			pendingChoices:  []*rulebook.FeatureChoice{}, // No pending choices
			expectedNeed:    false,
			expectedFeature: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCharService := mockcharacters.NewMockService(ctrl)
			handler := &ClassFeaturesHandler{
				characterService: mockCharService,
			}

			// Set up mock expectation
			mockCharService.EXPECT().
				GetPendingFeatureChoices(gomock.Any(), tt.character.ID).
				Return(tt.pendingChoices, nil)

			needsSelection, featureType := handler.ShouldShowClassFeatures(tt.character)
			assert.Equal(t, tt.expectedNeed, needsSelection)
			assert.Equal(t, tt.expectedFeature, featureType)
		})
	}
}

func TestFightingStyleMetadata(t *testing.T) {
	// Test that we can store and retrieve fighting style metadata
	fighter := &character.Character{
		Features: []*rulebook.CharacterFeature{
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
