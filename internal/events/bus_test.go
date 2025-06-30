package events_test

import (
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities/damage"
	"github.com/KirkDiggler/dnd-bot-discord/internal/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventBus_SimpleAttackFlow(t *testing.T) {
	// Create event bus
	bus := events.NewBus()

	// Create a test character
	character := &entities.Character{
		ID:    "test-barbarian",
		Name:  "Grunk",
		Level: 3,
		Class: &entities.Class{Key: "barbarian"},
	}

	// Create a simple test modifier that always adds +2 damage
	testModifier := &testDamageModifier{
		bonusAmount: 2,
		priority:    100,
	}

	// Subscribe to damage events
	bus.Subscribe(events.EventTypeOnDamageRoll, testModifier)

	// Create a damage roll event
	weapon := &entities.Weapon{
		Base: entities.BasicEquipment{
			Key:  "greataxe",
			Name: "Greataxe",
		},
		WeaponRange: "Melee",
		Properties: []*entities.ReferenceItem{
			{Key: "two-handed"},
		},
	}

	damageEvent := &events.OnDamageRollEvent{
		BaseEvent: events.BaseEvent{
			Type:  events.EventTypeOnDamageRoll,
			Actor: character,
		},
		Weapon:      weapon,
		DamageType:  damage.TypeSlashing,
		BaseDamage:  10,
		DamageBonus: 3, // STR bonus
		TotalDamage: 13,
	}

	// Emit the event
	err := bus.Emit(damageEvent)
	require.NoError(t, err)

	// Verify rage bonus was applied
	expectedDamage := 13 + 2 // Original + rage bonus
	assert.Equal(t, expectedDamage, damageEvent.TotalDamage)
	assert.Equal(t, 5, damageEvent.DamageBonus) // STR + rage
}

func TestEventBus_Priority(t *testing.T) {
	bus := events.NewBus()

	// Track execution order
	var executionOrder []string

	// Create listeners with different priorities
	lowPriority := &testListener{
		id:       "low",
		priority: 300,
		handler: func(e events.Event) error {
			executionOrder = append(executionOrder, "low")
			return nil
		},
	}

	highPriority := &testListener{
		id:       "high",
		priority: 100,
		handler: func(e events.Event) error {
			executionOrder = append(executionOrder, "high")
			return nil
		},
	}

	mediumPriority := &testListener{
		id:       "medium",
		priority: 200,
		handler: func(e events.Event) error {
			executionOrder = append(executionOrder, "medium")
			return nil
		},
	}

	// Subscribe in random order
	bus.Subscribe(events.EventTypeBeforeAttackRoll, lowPriority)
	bus.Subscribe(events.EventTypeBeforeAttackRoll, highPriority)
	bus.Subscribe(events.EventTypeBeforeAttackRoll, mediumPriority)

	// Create and emit event
	event := &events.BeforeAttackRollEvent{
		BaseEvent: events.BaseEvent{
			Type: events.EventTypeBeforeAttackRoll,
		},
	}

	err := bus.Emit(event)
	require.NoError(t, err)

	// Verify execution order (lower priority number = earlier execution)
	assert.Equal(t, []string{"high", "medium", "low"}, executionOrder)
}

func TestEventBus_Cancellation(t *testing.T) {
	bus := events.NewBus()

	var firstExecuted, secondExecuted bool

	// First listener cancels the event
	first := &testListener{
		id:       "first",
		priority: 100,
		handler: func(e events.Event) error {
			firstExecuted = true
			e.Cancel()
			return nil
		},
	}

	// Second listener should not execute
	second := &testListener{
		id:       "second",
		priority: 200,
		handler: func(e events.Event) error {
			secondExecuted = true
			return nil
		},
	}

	bus.Subscribe(events.EventTypeBeforeHit, first)
	bus.Subscribe(events.EventTypeBeforeHit, second)

	event := &events.BeforeHitEvent{
		BaseEvent: events.BaseEvent{
			Type: events.EventTypeBeforeHit,
		},
		AttackRoll: 15,
		TargetAC:   12,
	}

	err := bus.Emit(event)
	require.NoError(t, err)

	assert.True(t, firstExecuted)
	assert.False(t, secondExecuted)
	assert.True(t, event.IsCancelled())
}

// Test helper: simple event listener
type testListener struct {
	id       string
	priority int
	handler  func(events.Event) error
}

func (l *testListener) ID() string                       { return l.id }
func (l *testListener) Priority() int                    { return l.priority }
func (l *testListener) HandleEvent(e events.Event) error { return l.handler(e) }

// Test helper: damage modifier that always adds bonus
type testDamageModifier struct {
	bonusAmount int
	priority    int
}

func (m *testDamageModifier) ID() string    { return "test-damage-modifier" }
func (m *testDamageModifier) Priority() int { return m.priority }
func (m *testDamageModifier) HandleEvent(e events.Event) error {
	if damageEvent, ok := e.(*events.OnDamageRollEvent); ok {
		damageEvent.DamageBonus += m.bonusAmount
		damageEvent.TotalDamage += m.bonusAmount
	}
	return nil
}
