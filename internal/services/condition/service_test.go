package condition

import (
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/conditions"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testEventListener is a simple event listener for testing
type testEventListener struct {
	onEvent func(event *events.GameEvent) error
}

func (t *testEventListener) HandleEvent(event *events.GameEvent) error {
	if t.onEvent != nil {
		return t.onEvent(event)
	}
	return nil
}

func (t *testEventListener) Priority() int {
	return 100
}

func TestService_EventIntegration(t *testing.T) {
	// Create event bus
	eventBus := events.NewEventBus()

	// Create condition service
	svc := NewService(eventBus)

	// Test that conditions emit events
	t.Run("conditions emit events", func(t *testing.T) {
		// Set up event listener to verify events are emitted
		var appliedEvent *events.GameEvent
		var removedEvent *events.GameEvent

		listener := &testEventListener{
			onEvent: func(event *events.GameEvent) error {
				switch event.Type {
				case events.OnConditionApplied:
					appliedEvent = event
				case events.OnConditionRemoved:
					removedEvent = event
				}
				return nil
			},
		}

		eventBus.Subscribe(events.OnConditionApplied, listener)
		eventBus.Subscribe(events.OnConditionRemoved, listener)

		// Add condition - should emit OnConditionApplied
		cond, err := svc.AddCondition("char-1", conditions.Poisoned, "test", conditions.DurationRounds, 3)
		require.NoError(t, err)
		assert.NotNil(t, cond)

		// Verify event was emitted
		require.NotNil(t, appliedEvent)
		charID, _ := appliedEvent.GetStringContext(events.ContextCharacterID)
		assert.Equal(t, "char-1", charID)

		// Remove condition - should emit OnConditionRemoved
		err = svc.RemoveCondition("char-1", cond.ID)
		require.NoError(t, err)

		// Verify event was emitted
		require.NotNil(t, removedEvent)
		charID, _ = removedEvent.GetStringContext(events.ContextCharacterID)
		assert.Equal(t, "char-1", charID)
	})

	// Test direct service methods
	t.Run("direct service methods", func(t *testing.T) {
		// Add condition directly
		cond, err := svc.AddCondition("char-2", conditions.Stunned, "spell", conditions.DurationRounds, 2)
		require.NoError(t, err)
		assert.NotNil(t, cond)

		// Check condition exists
		assert.True(t, svc.HasCondition("char-2", conditions.Stunned))

		// Get active effects
		effects := svc.GetActiveEffects("char-2")
		assert.NotNil(t, effects)
		assert.True(t, effects.Incapacitated)

		// Process turn start
		svc.ProcessTurnStart("char-2")
		conds := svc.GetConditions("char-2")
		require.Len(t, conds, 1)
		assert.Equal(t, 2, conds[0].Remaining) // Still 2 because turn-based durations decrement at turn start

		// Process damage - add condition that ends on damage
		_, err = svc.AddCondition("char-2", conditions.DisadvantageOnNextAttack, "spell", conditions.DurationUntilDamaged, 0)
		require.NoError(t, err)
		assert.True(t, svc.HasCondition("char-2", conditions.DisadvantageOnNextAttack))

		// Apply damage
		svc.ProcessDamage("char-2", 5)
		assert.False(t, svc.HasCondition("char-2", conditions.DisadvantageOnNextAttack))
	})
}
