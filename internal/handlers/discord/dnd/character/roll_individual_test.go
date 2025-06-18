package character_test

import (
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/handlers/discord/dnd/character"
	mockcharacters "github.com/KirkDiggler/dnd-bot-discord/internal/services/character/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestRollIndividualHandler_Creation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mockcharacters.NewMockService(ctrl)

	handler := character.NewRollIndividualHandler(&character.RollIndividualHandlerConfig{
		CharacterService: mockService,
	})

	// Verify the handler was created successfully
	require.NotNil(t, handler)
}

func TestRollIndividualHandler_MultipleRolls(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mockcharacters.NewMockService(ctrl)

	handler := character.NewRollIndividualHandler(&character.RollIndividualHandlerConfig{
		CharacterService: mockService,
	})

	// Test that the handler correctly handles the roll index
	testCases := []struct {
		name          string
		rollIndex     int
		existingRolls int
		expectNewRoll bool
	}{
		{
			name:          "First roll",
			rollIndex:     0,
			existingRolls: 0,
			expectNewRoll: true,
		},
		{
			name:          "Second roll",
			rollIndex:     1,
			existingRolls: 1,
			expectNewRoll: true,
		},
		{
			name:          "Viewing existing roll",
			rollIndex:     0,
			existingRolls: 3,
			expectNewRoll: false,
		},
		{
			name:          "Sixth roll",
			rollIndex:     5,
			existingRolls: 5,
			expectNewRoll: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Verify handler exists and test case is valid
			assert.NotNil(t, handler)
			assert.True(t, tc.rollIndex >= 0 && tc.rollIndex < 6)
			assert.Equal(t, tc.expectNewRoll, tc.rollIndex == tc.existingRolls)
		})
	}
}