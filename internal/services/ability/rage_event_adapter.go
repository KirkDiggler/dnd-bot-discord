package ability

import (
	"fmt"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/events"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/events/modifiers"
)

// RageEventAdapter demonstrates how rage would work with the event system
// This is a proof of concept - in a full implementation, this would be integrated
// into the ability service
type RageEventAdapter struct {
	eventBus *events.EventBus
	// Track active rage modifiers by character ID
	activeRage map[string]*rageListener
}

// NewRageEventAdapter creates a new rage event adapter
func NewRageEventAdapter(eventBus *events.EventBus) *RageEventAdapter {
	return &RageEventAdapter{
		eventBus:   eventBus,
		activeRage: make(map[string]*rageListener),
	}
}

// ActivateRage activates rage for a character using the event system
func (r *RageEventAdapter) ActivateRage(char *character.Character) error {
	// Check if already raging
	if _, exists := r.activeRage[char.ID]; exists {
		return fmt.Errorf("character is already raging")
	}

	// Create rage modifier
	rageModifier := modifiers.NewRageModifier(char.Level)

	// Create listener that filters events for this character
	listener := &rageListener{
		characterID: char.ID,
		modifier:    rageModifier,
	}

	// Subscribe to relevant events
	r.eventBus.Subscribe(events.OnDamageRoll, listener)
	r.eventBus.Subscribe(events.BeforeTakeDamage, listener)

	// Store the listener so we can unsubscribe later
	r.activeRage[char.ID] = listener

	// Emit status applied event
	statusEvent := events.NewGameEvent(events.OnStatusApplied).
		WithActor(char).
		WithContext("status", "Rage").
		WithContext("duration", "10 rounds")

	return r.eventBus.Emit(statusEvent)
}

// DeactivateRage removes rage from a character
func (r *RageEventAdapter) DeactivateRage(char *character.Character) error {
	listener, exists := r.activeRage[char.ID]
	if !exists {
		return fmt.Errorf("character is not raging")
	}

	// Unsubscribe from events
	r.eventBus.Unsubscribe(events.OnDamageRoll, listener)
	r.eventBus.Unsubscribe(events.BeforeTakeDamage, listener)

	// Remove from tracking
	delete(r.activeRage, char.ID)

	// Emit status removed event
	statusEvent := events.NewGameEvent(events.OnStatusRemoved).
		WithActor(char).
		WithContext("status", "Rage")

	return r.eventBus.Emit(statusEvent)
}

// rageListener wraps a rage modifier to filter by character
type rageListener struct {
	characterID string
	modifier    *modifiers.RageModifier
}

func (rl *rageListener) HandleEvent(event *events.GameEvent) error {
	// Only apply to events for this character
	if event.Actor == nil || event.Actor.ID != rl.characterID {
		// For damage taken, check target instead
		if event.Type == events.BeforeTakeDamage && event.Target != nil && event.Target.ID == rl.characterID {
			// This is damage to our character
		} else {
			return nil
		}
	}

	// Check if modifier condition is met
	if !rl.modifier.Condition(event) {
		return nil
	}

	// Apply the modifier
	return rl.modifier.Apply(event)
}

func (rl *rageListener) Priority() int {
	return rl.modifier.Priority()
}
