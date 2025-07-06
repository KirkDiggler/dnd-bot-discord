package events

import (
	"context"
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	rpgevents "github.com/KirkDiggler/rpg-toolkit/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockListener implements EventListener for testing
type MockListener struct {
	priority int
	called   bool
	event    *GameEvent
	err      error
}

func (m *MockListener) Priority() int {
	return m.priority
}

func (m *MockListener) HandleEvent(event *GameEvent) error {
	m.called = true
	m.event = event
	return m.err
}

// customMockListener allows custom handlers for testing
type customMockListener struct {
	priority int
	handler  func(*GameEvent) error
}

func (c *customMockListener) Priority() int {
	return c.priority
}

func (c *customMockListener) HandleEvent(event *GameEvent) error {
	if c.handler != nil {
		return c.handler(event)
	}
	return nil
}

func TestToolkitBus_Subscribe_And_Emit(t *testing.T) {
	bus := NewToolkitBus()

	// Create a mock listener
	listener := &MockListener{priority: 100}

	// Subscribe to attack roll events
	bus.Subscribe(OnAttackRoll, listener)

	// Create test characters
	actor := &character.Character{ID: "fighter-123", Name: "Fighter"}
	target := &character.Character{ID: "goblin-456", Name: "Goblin"}

	// Emit an attack roll event
	event := &GameEvent{
		Type:   OnAttackRoll,
		Actor:  actor,
		Target: target,
		Context: map[string]interface{}{
			"weapon": "longsword",
		},
	}

	err := bus.Emit(event)
	require.NoError(t, err)

	// Verify listener was called
	assert.True(t, listener.called)
	assert.NotNil(t, listener.event)
	assert.Equal(t, OnAttackRoll, listener.event.Type)
	assert.Equal(t, actor.ID, listener.event.Actor.ID)
	assert.Equal(t, target.ID, listener.event.Target.ID)
	assert.Equal(t, "longsword", listener.event.Context["weapon"])
}

func TestToolkitBus_Modifiers(t *testing.T) {
	bus := NewToolkitBus()

	// Create listeners that add modifiers
	attackListener := &MockListener{priority: 100}
	damageListener := &MockListener{priority: 100}

	// Subscribe to events
	bus.Subscribe(OnAttackRoll, attackListener)
	bus.Subscribe(OnDamageRoll, damageListener)

	// For attack rolls, we'll use the toolkit bus directly to add modifiers
	rpgBus := bus.GetRPGBus()
	rpgBus.SubscribeFunc("attack.roll", 50, func(_ context.Context, e rpgevents.Event) error {
		// Add proficiency bonus
		e.Context().AddModifier(rpgevents.NewModifier(
			"proficiency",
			rpgevents.ModifierAttackBonus,
			rpgevents.NewRawValue(3, "proficiency"),
			100,
		))
		return nil
	})

	// Emit attack roll
	attackEvent := &GameEvent{
		Type:    OnAttackRoll,
		Context: make(map[string]interface{}),
	}

	err := bus.Emit(attackEvent)
	require.NoError(t, err)

	// The modifier should be converted to context
	assert.True(t, attackListener.called)
	if bonus, ok := attackListener.event.Context["attack_bonus"]; ok {
		assert.Equal(t, 3, bonus)
	}
}

func TestToolkitBus_ListenerCount(t *testing.T) {
	bus := NewToolkitBus()

	// Initially no listeners
	assert.Equal(t, 0, bus.ListenerCount(OnAttackRoll))

	// Add a listener
	listener1 := &MockListener{priority: 100}
	bus.Subscribe(OnAttackRoll, listener1)
	assert.Equal(t, 1, bus.ListenerCount(OnAttackRoll))

	// Add another listener
	listener2 := &MockListener{priority: 50}
	bus.Subscribe(OnAttackRoll, listener2)
	assert.Equal(t, 2, bus.ListenerCount(OnAttackRoll))

	// Remove a listener
	bus.Unsubscribe(OnAttackRoll, listener1)
	assert.Equal(t, 1, bus.ListenerCount(OnAttackRoll))
}

func TestToolkitBus_Clear(t *testing.T) {
	bus := NewToolkitBus()

	// Add multiple listeners
	listener1 := &MockListener{priority: 100}
	listener2 := &MockListener{priority: 50}

	bus.Subscribe(OnAttackRoll, listener1)
	bus.Subscribe(OnDamageRoll, listener2)

	assert.Equal(t, 1, bus.ListenerCount(OnAttackRoll))
	assert.Equal(t, 1, bus.ListenerCount(OnDamageRoll))

	// Clear all listeners
	bus.Clear()

	assert.Equal(t, 0, bus.ListenerCount(OnAttackRoll))
	assert.Equal(t, 0, bus.ListenerCount(OnDamageRoll))
}

func TestToolkitBus_ClearAlsoClearsDirectSubscriptions(t *testing.T) {
	bus := NewToolkitBus()
	rpgBus := bus.GetRPGBus()

	// Track if handlers are called
	trackedHandlerCalled := false
	directHandlerCalled := false

	// Add a tracked subscription through ToolkitBus
	listener := &customMockListener{
		priority: 100,
		handler: func(event *GameEvent) error {
			trackedHandlerCalled = true
			return nil
		},
	}
	bus.Subscribe(OnAttackRoll, listener)

	// Add a direct subscription to the underlying rpg-toolkit bus
	rpgBus.SubscribeFunc(rpgevents.EventOnAttackRoll, 50, func(_ context.Context, e rpgevents.Event) error {
		directHandlerCalled = true
		return nil
	})

	// Emit an event - both handlers should be called
	event := &GameEvent{Type: OnAttackRoll, Context: make(map[string]interface{})}
	err := bus.Emit(event)
	require.NoError(t, err)

	assert.True(t, trackedHandlerCalled, "Tracked handler should be called")
	assert.True(t, directHandlerCalled, "Direct handler should be called")

	// Reset flags
	trackedHandlerCalled = false
	directHandlerCalled = false

	// Clear the bus
	bus.Clear()

	// Emit again - neither handler should be called
	err = bus.Emit(event)
	require.NoError(t, err)

	assert.False(t, trackedHandlerCalled, "Tracked handler should not be called after Clear")
	assert.False(t, directHandlerCalled, "Direct handler should not be called after Clear")
}

func TestToolkitBus_DirectToolkitUsage(t *testing.T) {
	// This test shows that new code can use the toolkit bus directly
	// alongside old DND bot event handlers
	bus := NewToolkitBus()
	rpgBus := bus.GetRPGBus()

	// Old style DND bot handler
	oldStyleCalled := false
	oldListener := &customMockListener{
		priority: 50,
		handler: func(event *GameEvent) error {
			oldStyleCalled = true
			// Old code gets attack bonus from context
			if bonus, ok := event.Context["attack_bonus"]; ok {
				assert.Equal(t, 3, bonus)
			}
			return nil
		},
	}
	bus.Subscribe(OnAttackRoll, oldListener)

	// New style toolkit handler that adds modifiers
	newStyleCalled := false
	rpgBus.SubscribeFunc(rpgevents.EventOnAttackRoll, 100, func(_ context.Context, e rpgevents.Event) error {
		newStyleCalled = true

		// New code uses modifier system
		e.Context().AddModifier(rpgevents.NewModifier(
			"proficiency",
			rpgevents.ModifierAttackBonus,
			rpgevents.NewRawValue(3, "proficiency"),
			100,
		))
		return nil
	})

	// Emit event through old interface
	event := &GameEvent{
		Type:    OnAttackRoll,
		Context: make(map[string]interface{}),
	}

	err := bus.Emit(event)
	require.NoError(t, err)

	// Both handlers should be called
	assert.True(t, oldStyleCalled, "Old style handler should be called")
	assert.True(t, newStyleCalled, "New style handler should be called")
}
