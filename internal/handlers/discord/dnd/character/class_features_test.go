package character

import (
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	mockchar "github.com/KirkDiggler/dnd-bot-discord/internal/services/character/mock"
	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestClassFeaturesHandler_HandleFavoredEnemy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mockchar.NewMockService(ctrl)
	handler := NewClassFeaturesHandler(mockService)

	// Create a test ranger character
	char := &entities.Character{
		ID:   "test-ranger",
		Name: "Test Ranger",
		Class: &entities.Class{
			Key:  "ranger",
			Name: "Ranger",
		},
		Features: []*entities.CharacterFeature{
			{
				Key:      "favored_enemy",
				Name:     "Favored Enemy",
				Type:     entities.FeatureTypeClass,
				Source:   "Ranger",
				Metadata: nil,
			},
		},
	}

	// Mock the service calls
	mockService.EXPECT().GetByID("test-ranger").Return(char, nil)
	mockService.EXPECT().UpdateEquipment(char).Return(nil)

	// Create a test request
	req := &ClassFeaturesRequest{
		Session:     &discordgo.Session{},
		Interaction: &discordgo.InteractionCreate{},
		CharacterID: "test-ranger",
		FeatureType: "favored_enemy",
		Selection:   "orcs",
	}

	// Handle the request
	err := handler.Handle(req)
	assert.NoError(t, err)

	// Verify the favored enemy was set
	assert.NotNil(t, char.Features[0].Metadata)
	assert.Equal(t, "orcs", char.Features[0].Metadata["enemy_type"])
}

func TestClassFeaturesHandler_HandleNaturalExplorer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mockchar.NewMockService(ctrl)
	handler := NewClassFeaturesHandler(mockService)

	// Create a test ranger character
	char := &entities.Character{
		ID:   "test-ranger",
		Name: "Test Ranger",
		Class: &entities.Class{
			Key:  "ranger",
			Name: "Ranger",
		},
		Features: []*entities.CharacterFeature{
			{
				Key:      "natural_explorer",
				Name:     "Natural Explorer",
				Type:     entities.FeatureTypeClass,
				Source:   "Ranger",
				Metadata: nil,
			},
		},
	}

	// Mock the service calls
	mockService.EXPECT().GetByID("test-ranger").Return(char, nil)
	mockService.EXPECT().UpdateEquipment(char).Return(nil)

	// Create a test request
	req := &ClassFeaturesRequest{
		Session:     &discordgo.Session{},
		Interaction: &discordgo.InteractionCreate{},
		CharacterID: "test-ranger",
		FeatureType: "natural_explorer",
		Selection:   "forest",
	}

	// Handle the request
	err := handler.Handle(req)
	assert.NoError(t, err)

	// Verify the natural explorer terrain was set
	assert.NotNil(t, char.Features[0].Metadata)
	assert.Equal(t, "forest", char.Features[0].Metadata["terrain_type"])
}
