package abilities

import (
	"context"
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/events"
	rulebook "github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockCharacterService for testing
type mockCharacterService struct {
	characters map[string]*character.Character
}

func (m *mockCharacterService) UpdateEquipment(char *character.Character) error {
	return nil
}

func (m *mockCharacterService) GetByID(id string) (*character.Character, error) {
	if char, ok := m.characters[id]; ok {
		return char, nil
	}
	return nil, nil
}

func TestViciousMockeryHandler_Execute(t *testing.T) {
	// Create event bus
	eventBus := events.NewEventBus()

	// Create handler
	handler := NewViciousMockeryHandler(eventBus)
	handler.SetDiceRoller(&mockDiceRoller{result: 3}) // 1d4 = 3

	// Create a bard character
	bard := &character.Character{
		ID:    "bard-123",
		Name:  "Scanlan",
		Level: 3,
		Class: &rulebook.Class{Key: "bard"},
		Attributes: map[shared.Attribute]*character.AbilityScore{
			shared.AttributeCharisma: {Score: 16, Bonus: 3},
		},
	}

	// Create ability
	ability := &shared.ActiveAbility{
		Key:           "vicious_mockery",
		IsActive:      false,
		UsesRemaining: -1, // Cantrip
	}

	// Test successful cast
	t.Run("successful cast", func(t *testing.T) {
		input := &Input{
			TargetID:    "goblin-123",
			EncounterID: "enc-123",
		}

		// Set up mock character service
		handler.SetCharacterService(&mockCharacterService{
			characters: map[string]*character.Character{
				"goblin-123": {ID: "goblin-123", Name: "Goblin"},
			},
		})

		result, err := handler.Execute(context.Background(), bard, ability, input)

		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Contains(t, result.Message, "cutting words deal")
		assert.Contains(t, result.Message, "psychic damage")
		assert.Equal(t, -1, result.UsesRemaining) // Unlimited
	})

	// Test non-bard
	t.Run("non-bard cannot use", func(t *testing.T) {
		fighter := &character.Character{
			ID:    "fighter-123",
			Name:  "Grog",
			Level: 3,
			Class: &rulebook.Class{Key: "fighter"},
		}

		input := &Input{
			TargetID: "goblin-123",
		}

		result, err := handler.Execute(context.Background(), fighter, ability, input)

		require.NoError(t, err)
		assert.False(t, result.Success)
		assert.Equal(t, "Only bards can use Vicious Mockery", result.Message)
	})

	// Test no target
	t.Run("requires target", func(t *testing.T) {
		input := &Input{
			TargetID: "", // No target
		}

		result, err := handler.Execute(context.Background(), bard, ability, input)

		require.NoError(t, err)
		assert.False(t, result.Success)
		assert.Contains(t, result.Message, "requires exactly one target")
	})
}
