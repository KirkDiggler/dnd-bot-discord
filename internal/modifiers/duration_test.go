package modifiers_test

import (
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/events"
	"github.com/KirkDiggler/dnd-bot-discord/internal/modifiers"
	"github.com/stretchr/testify/assert"
)

// testEvent implements events.Event for testing
type testEvent struct {
	eventType events.EventType
}

func (e *testEvent) GetType() events.EventType       { return e.eventType }
func (e *testEvent) GetActor() *entities.Character   { return nil }
func (e *testEvent) GetTarget() *entities.Character  { return nil }
func (e *testEvent) IsCancelled() bool               { return false }
func (e *testEvent) Cancel()                         {}
func (e *testEvent) Accept(v events.ModifierVisitor) {}

func TestPermanentDuration(t *testing.T) {
	d := &modifiers.PermanentDuration{}

	// Should never expire
	assert.False(t, d.IsExpired())

	// Events should not affect it
	event := &testEvent{eventType: events.EventTypeOnTurnEnd}
	d.OnEventOccurred(event)
	assert.False(t, d.IsExpired())

	assert.Equal(t, "permanent", d.String())
}

func TestRoundsDuration(t *testing.T) {
	d := modifiers.NewRoundsDuration(3)

	// Should not be expired initially
	assert.False(t, d.IsExpired())
	assert.Equal(t, "3 rounds remaining", d.String())

	// Simulate turn end events
	turnEndEvent := &testEvent{eventType: events.EventTypeOnTurnEnd}

	// After 1 turn
	d.OnEventOccurred(turnEndEvent)
	assert.False(t, d.IsExpired())
	assert.Equal(t, "2 rounds remaining", d.String())

	// After 2 turns
	d.OnEventOccurred(turnEndEvent)
	assert.False(t, d.IsExpired())
	assert.Equal(t, "1 rounds remaining", d.String())

	// After 3 turns - should expire
	d.OnEventOccurred(turnEndEvent)
	assert.True(t, d.IsExpired())
	assert.Equal(t, "0 rounds remaining", d.String())

	// Other events should not affect it
	otherEvent := &testEvent{eventType: events.EventTypeBeforeAttackRoll}
	d = modifiers.NewRoundsDuration(1)
	d.OnEventOccurred(otherEvent)
	assert.False(t, d.IsExpired()) // Should still have 1 round
}

func TestUntilRestDuration(t *testing.T) {
	d := &modifiers.UntilRestDuration{}

	// Should not be expired initially
	assert.False(t, d.IsExpired())
	assert.Equal(t, "until rest", d.String())

	// Other events should not expire it
	turnEndEvent := &testEvent{eventType: events.EventTypeOnTurnEnd}
	d.OnEventOccurred(turnEndEvent)
	assert.False(t, d.IsExpired())

	// Short rest should expire it
	shortRestEvent := &testEvent{eventType: events.EventTypeOnShortRest}
	d.OnEventOccurred(shortRestEvent)
	assert.True(t, d.IsExpired())

	// Test long rest as well
	d = &modifiers.UntilRestDuration{}
	longRestEvent := &testEvent{eventType: events.EventTypeOnLongRest}
	d.OnEventOccurred(longRestEvent)
	assert.True(t, d.IsExpired())
}

func TestConcentrationDuration(t *testing.T) {
	d := modifiers.NewConcentrationDuration(10)

	// Should not be expired initially
	assert.False(t, d.IsExpired())
	assert.Equal(t, "concentration (10 rounds)", d.String())

	// Turn end should decrement rounds
	turnEndEvent := &testEvent{eventType: events.EventTypeOnTurnEnd}
	d.OnEventOccurred(turnEndEvent)
	assert.False(t, d.IsExpired())
	assert.Equal(t, "concentration (9 rounds)", d.String())

	// Breaking concentration should expire it
	d.Break()
	assert.True(t, d.IsExpired())
	assert.Equal(t, "concentration broken", d.String())

	// Test expiration by round count
	d = modifiers.NewConcentrationDuration(2)
	d.OnEventOccurred(turnEndEvent)
	d.OnEventOccurred(turnEndEvent)
	assert.True(t, d.IsExpired())
}
