package abilities

import (
	"context"
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/adapters/rpgtoolkit"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	rpgevents "github.com/KirkDiggler/rpg-toolkit/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestViciousMockeryRPGListener(t *testing.T) {
	t.Run("applies disadvantage to attack roll", func(t *testing.T) {
		// Create RPG event bus
		rpgBus := rpgevents.NewBus()

		// Create listener
		listener := NewViciousMockeryRPGListener(rpgBus)
		require.NotNil(t, listener)

		// Create a character with vicious mockery effect
		char := &character.Character{
			ID:   "test-char",
			Name: "Test Character",
			Resources: &character.CharacterResources{
				ActiveEffects: []*shared.ActiveEffect{
					{
						Name:         "Vicious Mockery Disadvantage",
						Description:  "Disadvantage on next attack roll",
						Source:       "Vicious Mockery",
						Duration:     1,
						DurationType: shared.DurationTypeRounds,
					},
				},
			},
		}

		// Create attack roll event
		attackEvent := rpgevents.NewGameEvent(
			rpgevents.EventBeforeAttackRoll,
			&rpgtoolkit.CharacterEntityAdapter{Character: char},
			nil,
		)

		// Emit the event
		err := rpgBus.Publish(context.Background(), attackEvent)
		require.NoError(t, err)

		// Check that disadvantage modifier was added
		modifiers := attackEvent.Context().Modifiers()
		require.Len(t, modifiers, 1)

		modifier := modifiers[0]
		assert.Equal(t, "vicious_mockery", modifier.Source())
		assert.Equal(t, rpgevents.ModifierDisadvantage, modifier.Type())
		assert.Equal(t, 1, modifier.ModifierValue().GetValue())

		// Check that the effect was removed after use
		assert.Empty(t, char.Resources.ActiveEffects)
	})

	t.Run("ignores characters without vicious mockery", func(t *testing.T) {
		// Create RPG event bus
		rpgBus := rpgevents.NewBus()

		// Create listener
		NewViciousMockeryRPGListener(rpgBus)

		// Create a character without vicious mockery effect
		char := &character.Character{
			ID:        "test-char",
			Name:      "Test Character",
			Resources: &character.CharacterResources{},
		}

		// Create attack roll event
		attackEvent := rpgevents.NewGameEvent(
			rpgevents.EventBeforeAttackRoll,
			&rpgtoolkit.CharacterEntityAdapter{Character: char},
			nil,
		)

		// Emit the event
		err := rpgBus.Publish(context.Background(), attackEvent)
		require.NoError(t, err)

		// Check that no modifiers were added
		modifiers := attackEvent.Context().Modifiers()
		assert.Empty(t, modifiers)
	})

	t.Run("only removes vicious mockery effect", func(t *testing.T) {
		// Create RPG event bus
		rpgBus := rpgevents.NewBus()

		// Create listener
		NewViciousMockeryRPGListener(rpgBus)

		// Create a character with multiple effects
		char := &character.Character{
			ID:   "test-char",
			Name: "Test Character",
			Resources: &character.CharacterResources{
				ActiveEffects: []*shared.ActiveEffect{
					{
						Name:         "Vicious Mockery Disadvantage",
						Description:  "Disadvantage on next attack roll",
						Source:       "Vicious Mockery",
						Duration:     1,
						DurationType: shared.DurationTypeRounds,
					},
					{
						Name:         "Bless",
						Description:  "+1d4 to attack rolls",
						Source:       "Bless",
						Duration:     10,
						DurationType: shared.DurationTypeMinutes,
					},
				},
			},
		}

		// Create attack roll event
		attackEvent := rpgevents.NewGameEvent(
			rpgevents.EventBeforeAttackRoll,
			&rpgtoolkit.CharacterEntityAdapter{Character: char},
			nil,
		)

		// Emit the event
		err := rpgBus.Publish(context.Background(), attackEvent)
		require.NoError(t, err)

		// Check that only vicious mockery was removed
		require.Len(t, char.Resources.ActiveEffects, 1)
		assert.Equal(t, "Bless", char.Resources.ActiveEffects[0].Name)
	})
}
