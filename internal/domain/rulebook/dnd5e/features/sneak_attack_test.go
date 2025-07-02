package features

import (
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/events"
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

func TestSneakAttackModifier_Condition(t *testing.T) {
	charID := "rogue-123"
	modifier := NewSneakAttackModifier(charID, 5, &mockDiceRoller{result: 10})

	tests := []struct {
		name     string
		event    *events.GameEvent
		expected bool
	}{
		{
			name: "valid sneak attack with advantage and finesse weapon",
			event: events.NewGameEvent(events.OnDamageRoll).
				WithActor(&character.Character{ID: charID}).
				WithContext("weapon_key", "rapier").
				WithContext("weapon_has_finesse", true).
				WithContext("has_advantage", true),
			expected: true,
		},
		{
			name: "valid sneak attack with ranged weapon and ally adjacent",
			event: events.NewGameEvent(events.OnDamageRoll).
				WithActor(&character.Character{ID: charID}).
				WithContext("weapon_key", "shortbow").
				WithContext("weapon_type", "ranged").
				WithContext("ally_adjacent", true),
			expected: true,
		},
		{
			name: "invalid - no advantage or ally",
			event: events.NewGameEvent(events.OnDamageRoll).
				WithActor(&character.Character{ID: charID}).
				WithContext("weapon_key", "rapier").
				WithContext("weapon_has_finesse", true).
				WithContext("has_advantage", false).
				WithContext("ally_adjacent", false),
			expected: false,
		},
		{
			name: "invalid - weapon not finesse or ranged",
			event: events.NewGameEvent(events.OnDamageRoll).
				WithActor(&character.Character{ID: charID}).
				WithContext("weapon_key", "greataxe").
				WithContext("weapon_has_finesse", false).
				WithContext("weapon_type", "melee").
				WithContext("has_advantage", true),
			expected: false,
		},
		{
			name: "invalid - advantage and disadvantage cancel",
			event: events.NewGameEvent(events.OnDamageRoll).
				WithActor(&character.Character{ID: charID}).
				WithContext("weapon_key", "rapier").
				WithContext("weapon_has_finesse", true).
				WithContext("has_advantage", true).
				WithContext("has_disadvantage", true),
			expected: false,
		},
		{
			name: "invalid - different character",
			event: events.NewGameEvent(events.OnDamageRoll).
				WithActor(&character.Character{ID: "different-char"}).
				WithContext("weapon_key", "rapier").
				WithContext("weapon_has_finesse", true).
				WithContext("has_advantage", true),
			expected: false,
		},
		{
			name: "turn start event for correct character",
			event: events.NewGameEvent(events.OnTurnStart).
				WithActor(&character.Character{ID: charID}),
			expected: true,
		},
		{
			name: "turn start event for different character",
			event: events.NewGameEvent(events.OnTurnStart).
				WithActor(&character.Character{ID: "different-char"}),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := modifier.Condition(tt.event)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSneakAttackModifier_Apply(t *testing.T) {
	charID := "rogue-123"
	modifier := NewSneakAttackModifier(charID, 5, &mockDiceRoller{result: 15}) // 3d6 = 15

	t.Run("applies sneak attack damage", func(t *testing.T) {
		event := events.NewGameEvent(events.OnDamageRoll).
			WithActor(&character.Character{ID: charID}).
			WithContext("damage", 10).
			WithContext("is_critical", false)

		err := modifier.Apply(event)
		require.NoError(t, err)

		// Check damage was increased
		damage, _ := event.GetIntContext("damage")
		assert.Equal(t, 25, damage) // 10 + 15

		// Check sneak attack info was added
		sneakDamage, _ := event.GetIntContext("sneak_attack_damage")
		assert.Equal(t, 15, sneakDamage)

		sneakDice, _ := event.GetStringContext("sneak_attack_dice")
		assert.Equal(t, "3d6", sneakDice)

		// Should be marked as used
		assert.True(t, modifier.usedThisTurn)
	})

	t.Run("doubles dice on critical", func(t *testing.T) {
		mod := NewSneakAttackModifier(charID, 5, &mockDiceRoller{result: 30}) // 6d6 = 30

		event := events.NewGameEvent(events.OnDamageRoll).
			WithActor(&character.Character{ID: charID}).
			WithContext("damage", 10).
			WithContext("is_critical", true)

		err := mod.Apply(event)
		require.NoError(t, err)

		damage, _ := event.GetIntContext("damage")
		assert.Equal(t, 40, damage) // 10 + 30

		sneakDice, _ := event.GetStringContext("sneak_attack_dice")
		assert.Equal(t, "6d6", sneakDice)
	})

	t.Run("resets on turn start", func(t *testing.T) {
		modifier.usedThisTurn = true

		event := events.NewGameEvent(events.OnTurnStart).
			WithActor(&character.Character{ID: charID})

		err := modifier.Apply(event)
		require.NoError(t, err)

		assert.False(t, modifier.usedThisTurn)
	})
}

func TestSneakAttackModifier_OnlyOncePerTurn(t *testing.T) {
	charID := "rogue-123"
	modifier := NewSneakAttackModifier(charID, 5, &mockDiceRoller{result: 15})

	// First attack
	event1 := events.NewGameEvent(events.OnDamageRoll).
		WithActor(&character.Character{ID: charID}).
		WithContext("weapon_key", "rapier").
		WithContext("weapon_has_finesse", true).
		WithContext("has_advantage", true)

	// Should apply to first attack
	assert.True(t, modifier.Condition(event1))

	// Apply the damage
	event1.WithContext("damage", 10)
	err := modifier.Apply(event1)
	require.NoError(t, err)

	// Second attack same turn
	event2 := events.NewGameEvent(events.OnDamageRoll).
		WithActor(&character.Character{ID: charID}).
		WithContext("weapon_key", "rapier").
		WithContext("weapon_has_finesse", true).
		WithContext("has_advantage", true)

	// Should not apply to second attack
	assert.False(t, modifier.Condition(event2))
}

func TestSneakAttackListener(t *testing.T) {
	listener := NewSneakAttackListener("rogue-123", 5, 0, &mockDiceRoller{result: 15})

	assert.Equal(t, "sneak_attack_rogue-123", listener.ID())
	assert.Equal(t, 90, listener.Priority())

	// Test permanent duration
	duration := listener.Duration()
	assert.IsType(t, &events.PermanentDuration{}, duration)
}
