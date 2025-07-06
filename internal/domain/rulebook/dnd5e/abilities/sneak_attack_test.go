//go:build ignore
// +build ignore

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

// mockDiceRoller for testing
type mockDiceRoller struct {
	result int
}

func (m *mockDiceRoller) Roll(numDice, sides, modifier int) (struct{ Total int }, error) {
	return struct{ Total int }{Total: m.result}, nil
}

func TestSneakAttackHandler_Execute(t *testing.T) {
	tests := []struct {
		name           string
		character      *character.Character
		ability        *shared.ActiveAbility
		input          *Input
		expectSuccess  bool
		expectMessage  string
		expectListener bool
	}{
		{
			name: "successful activation for rogue",
			character: &character.Character{
				ID:    "rogue-123",
				Name:  "Sneaky",
				Level: 5,
				Class: &rulebook.Class{Key: "rogue"},
			},
			ability: &shared.ActiveAbility{
				Key:      "sneak_attack",
				IsActive: false,
			},
			input: &Input{
				EncounterID: "enc-123",
			},
			expectSuccess:  true,
			expectMessage:  "Sneak Attack tracking enabled (3d6 damage)",
			expectListener: true,
		},
		{
			name: "fails for non-rogue",
			character: &character.Character{
				ID:    "fighter-123",
				Name:  "Fighter",
				Level: 5,
				Class: &rulebook.Class{Key: "fighter"},
			},
			ability: &shared.ActiveAbility{
				Key:      "sneak_attack",
				IsActive: false,
			},
			input:          &Input{},
			expectSuccess:  false,
			expectMessage:  "Only rogues can use Sneak Attack",
			expectListener: false,
		},
		{
			name: "level 1 rogue gets 1d6",
			character: &character.Character{
				ID:    "rogue-456",
				Name:  "Newbie",
				Level: 1,
				Class: &rulebook.Class{Key: "rogue"},
			},
			ability: &shared.ActiveAbility{
				Key:      "sneak_attack",
				IsActive: false,
			},
			input:          &Input{},
			expectSuccess:  true,
			expectMessage:  "Sneak Attack tracking enabled (1d6 damage)",
			expectListener: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create event bus
			eventBus := events.NewEventBus()

			// Create handler
			handler := NewSneakAttackHandler(eventBus)
			handler.SetDiceRoller(&mockDiceRoller{result: 10})

			// Execute
			result, err := handler.Execute(context.Background(), tt.character, tt.ability, tt.input)

			// Verify
			require.NoError(t, err)
			assert.Equal(t, tt.expectSuccess, result.Success)
			assert.Equal(t, tt.expectMessage, result.Message)

			// Check if listener was registered
			if tt.expectListener {
				assert.NotNil(t, handler.GetActiveListener(tt.character.ID))
				assert.True(t, tt.ability.IsActive)
				assert.Equal(t, "Sneak Attack", result.EffectName)
			} else {
				assert.Nil(t, handler.GetActiveListener(tt.character.ID))
			}
		})
	}
}

func TestSneakAttackHandler_NoEventBus(t *testing.T) {
	// Create handler without event bus
	handler := NewSneakAttackHandler(nil)

	char := &character.Character{
		ID:    "rogue-123",
		Name:  "Sneaky",
		Level: 5,
		Class: &rulebook.Class{Key: "rogue"},
	}

	ability := &shared.ActiveAbility{
		Key: "sneak_attack",
	}

	result, err := handler.Execute(context.Background(), char, ability, &Input{})

	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, "Event system not available", result.Message)
}
