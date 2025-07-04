package rpgtoolkit

import (
	"context"
	"fmt"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/events"
	rpgevents "github.com/KirkDiggler/rpg-toolkit/events"
)

// EventBusAdapter wraps rpg-toolkit's event bus to provide compatibility
// with the Discord bot's existing event system
type EventBusAdapter struct {
	rpgBus *rpgevents.Bus
	// Maps Discord bot event types to rpg-toolkit event types
	eventTypeMap map[events.EventType]string
	// Maps rpg-toolkit event types back to Discord bot types
	reverseMap map[string]events.EventType
	// Track listeners for unsubscribe support
	listenerMap map[events.EventListener]map[events.EventType][]func(context.Context, rpgevents.Event) error
}

// NewEventBusAdapter creates a new adapter wrapping an rpg-toolkit event bus
func NewEventBusAdapter() *EventBusAdapter {
	adapter := &EventBusAdapter{
		rpgBus:       rpgevents.NewBus(),
		eventTypeMap: make(map[events.EventType]string),
		reverseMap:   make(map[string]events.EventType),
		listenerMap:  make(map[events.EventListener]map[events.EventType][]func(context.Context, rpgevents.Event) error),
	}

	// Set up event type mappings
	adapter.setupEventMappings()

	return adapter
}

// setupEventMappings configures the bidirectional mapping between event systems
func (a *EventBusAdapter) setupEventMappings() {
	mappings := map[events.EventType]string{
		events.BeforeAttackRoll:  rpgevents.EventBeforeAttackRoll,
		events.OnAttackRoll:      rpgevents.EventOnAttackRoll,
		events.AfterAttackRoll:   rpgevents.EventAfterAttackRoll,
		events.BeforeDamageRoll:  rpgevents.EventBeforeDamageRoll,
		events.OnDamageRoll:      rpgevents.EventOnDamageRoll,
		events.AfterDamageRoll:   rpgevents.EventAfterDamageRoll,
		events.BeforeTakeDamage:  rpgevents.EventBeforeTakeDamage,
		events.OnTakeDamage:      rpgevents.EventOnTakeDamage,
		events.AfterTakeDamage:   rpgevents.EventAfterTakeDamage,
		events.BeforeSavingThrow: rpgevents.EventBeforeSavingThrow,
		events.OnSavingThrow:     rpgevents.EventOnSavingThrow,
		events.AfterSavingThrow:  rpgevents.EventAfterSavingThrow,
		events.OnTurnStart:       rpgevents.EventOnTurnStart,
		events.OnTurnEnd:         rpgevents.EventOnTurnEnd,
		events.OnStatusApplied:   rpgevents.EventOnConditionApplied,
		events.OnStatusRemoved:   rpgevents.EventOnConditionRemoved,
		events.OnShortRest:       rpgevents.EventOnShortRest,
		events.OnLongRest:        rpgevents.EventOnLongRest,
		events.OnSpellCast:       rpgevents.EventOnSpellCast,
		events.OnSpellDamage:     rpgevents.EventOnSpellDamage,
	}

	for discordType, rpgType := range mappings {
		a.eventTypeMap[discordType] = rpgType
		a.reverseMap[rpgType] = discordType
	}
}

// Emit publishes an event to the rpg-toolkit event bus
func (a *EventBusAdapter) Emit(event *events.GameEvent) error {
	rpgEventType, ok := a.eventTypeMap[event.Type]
	if !ok {
		return fmt.Errorf("unknown event type: %v", event.Type)
	}

	// Convert the Discord event to an rpg-toolkit event
	gameEvent := a.createGameEvent(rpgEventType, event)

	return a.rpgBus.Publish(context.Background(), gameEvent)
}

// Subscribe registers a listener for a specific event type
func (a *EventBusAdapter) Subscribe(eventType events.EventType, listener events.EventListener) {
	rpgEventType, ok := a.eventTypeMap[eventType]
	if !ok {
		// Log warning about unknown event type
		return
	}

	// Wrap the listener to convert rpg-toolkit events back to Discord format
	wrappedHandler := a.wrapHandler(listener, eventType)

	// Track the handler for unsubscribe
	if a.listenerMap[listener] == nil {
		a.listenerMap[listener] = make(map[events.EventType][]func(context.Context, rpgevents.Event) error)
	}
	a.listenerMap[listener][eventType] = append(a.listenerMap[listener][eventType], wrappedHandler)

	// Use the listener's priority
	a.rpgBus.SubscribeFunc(rpgEventType, listener.Priority(), wrappedHandler)
}

// Unsubscribe removes a listener for a specific event type
func (a *EventBusAdapter) Unsubscribe(eventType events.EventType, listener events.EventListener) {
	// For now, we can't unsubscribe from rpg-toolkit bus
	// TODO: Implement when rpg-toolkit supports unsubscribe
	if handlers, ok := a.listenerMap[listener]; ok {
		delete(handlers, eventType)
		if len(handlers) == 0 {
			delete(a.listenerMap, listener)
		}
	}
}

// createGameEvent converts Discord bot event data to rpg-toolkit GameEvent
func (a *EventBusAdapter) createGameEvent(eventType string, data interface{}) rpgevents.Event {
	// Extract source and target based on the data type
	var source, target EntityAdapter

	switch d := data.(type) {
	case *events.GameEvent:
		if d.Actor != nil {
			source = &CharacterEntityAdapter{Character: d.Actor}
		}
		if d.Target != nil {
			target = &CharacterEntityAdapter{Character: d.Target}
		}

		event := rpgevents.NewGameEvent(eventType, source, target)

		// Copy context data
		for k, v := range d.Context {
			event.Context().Set(k, v)
		}

		// Handle cancellation
		if d.Cancelled {
			event.Cancel()
		}

		return event

	default:
		// For other data types, create a simple event
		event := rpgevents.NewGameEvent(eventType, nil, nil)

		// Store the entire data object in context
		event.Context().Set("data", data)

		return event
	}
}

// wrapHandler wraps a Discord bot event listener to work with rpg-toolkit events
func (a *EventBusAdapter) wrapHandler(listener events.EventListener, expectedType events.EventType) func(context.Context, rpgevents.Event) error {
	return func(ctx context.Context, e rpgevents.Event) error {
		// Convert rpg-toolkit event back to Discord bot format
		gameEvent := a.convertToGameEvent(e, expectedType)
		if gameEvent == nil {
			return nil
		}

		// Call the original listener
		return listener.HandleEvent(gameEvent)
	}
}

// convertToGameEvent converts an rpg-toolkit event to Discord bot GameEvent
func (a *EventBusAdapter) convertToGameEvent(rpgEvent rpgevents.Event, eventType events.EventType) *events.GameEvent {
	// Get the reverse mapped event type
	discordType, ok := a.reverseMap[rpgEvent.Type()]
	if !ok {
		discordType = eventType // fallback to expected type
	}

	// Create GameEvent with basic fields
	gameEvent := &events.GameEvent{
		Type:      discordType,
		Context:   make(map[string]interface{}),
		Cancelled: rpgEvent.IsCancelled(),
	}

	// Convert entities to characters if possible
	if source := rpgEvent.Source(); source != nil {
		if adapter, ok := source.(*CharacterEntityAdapter); ok {
			gameEvent.Actor = adapter.Character
		}
	}
	if target := rpgEvent.Target(); target != nil {
		if adapter, ok := target.(*CharacterEntityAdapter); ok {
			gameEvent.Target = adapter.Character
		}
	}

	// Copy context data - iterate through all context fields
	ctx := rpgEvent.Context()
	// TODO: The rpg-toolkit Context interface doesn't expose a way to iterate all keys
	// For now, copy all known fields that might be used
	knownFields := []string{
		"weapon", "damage", "damage_type", "target_id", "effect_name",
		"effect_duration", "effect_type", "encounter_id", "user_id",
		"spell_name", "caster_id", "damage_amount",
	}

	for _, field := range knownFields {
		if value, ok := ctx.Get(field); ok {
			gameEvent.Context[field] = value
		}
	}

	return gameEvent
}

// GetRPGBus returns the underlying rpg-toolkit event bus for direct access
func (a *EventBusAdapter) GetRPGBus() *rpgevents.Bus {
	return a.rpgBus
}

// Clear removes all listeners
func (a *EventBusAdapter) Clear() {
	// Create a new bus to clear all listeners
	a.rpgBus = rpgevents.NewBus()
	a.listenerMap = make(map[events.EventListener]map[events.EventType][]func(context.Context, rpgevents.Event) error)
}

// ListenerCount returns the number of listeners for an event type
func (a *EventBusAdapter) ListenerCount(eventType events.EventType) int {
	count := 0
	for _, handlers := range a.listenerMap {
		if _, ok := handlers[eventType]; ok {
			count += len(handlers[eventType])
		}
	}
	return count
}
