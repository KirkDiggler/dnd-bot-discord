package events

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/KirkDiggler/rpg-toolkit/core"
	rpgevents "github.com/KirkDiggler/rpg-toolkit/events"
)

// ToolkitBus directly uses rpg-toolkit's event bus
// This replaces the old event bus implementation
type ToolkitBus struct {
	bus *rpgevents.Bus
	mu  sync.RWMutex

	// Track subscriptions for ListenerCount and Unsubscribe support
	subscriptions map[EventType]map[EventListener]string // eventType -> listener -> subscriptionID
}

// GetRPGBus returns the underlying rpg-toolkit event bus for direct usage
func (tb *ToolkitBus) GetRPGBus() *rpgevents.Bus {
	return tb.bus
}

// NewToolkitBus creates a new event bus using rpg-toolkit directly
func NewToolkitBus() *ToolkitBus {
	return &ToolkitBus{
		bus:           rpgevents.NewBus(),
		subscriptions: make(map[EventType]map[EventListener]string),
	}
}

// Subscribe adds a listener for a specific event type
func (tb *ToolkitBus) Subscribe(eventType EventType, listener EventListener) {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// Map DND bot event type to toolkit event name
	toolkitEvent := mapEventType(eventType)

	// Create handler that converts toolkit events to DND bot format
	handler := func(ctx context.Context, e rpgevents.Event) error {
		// Convert toolkit event to GameEvent
		gameEvent := convertToGameEvent(e, eventType)
		if gameEvent == nil {
			return nil
		}

		// Call the original listener
		return listener.HandleEvent(gameEvent)
	}

	// Subscribe with priority from listener
	id := tb.bus.SubscribeFunc(toolkitEvent, listener.Priority(), handler)

	// Track subscription
	if tb.subscriptions[eventType] == nil {
		tb.subscriptions[eventType] = make(map[EventListener]string)
	}
	tb.subscriptions[eventType][listener] = id
}

// Unsubscribe removes a listener for a specific event type
func (tb *ToolkitBus) Unsubscribe(eventType EventType, listener EventListener) {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	if listeners, ok := tb.subscriptions[eventType]; ok {
		if id, ok := listeners[listener]; ok {
			// Ignore error as we're cleaning up our tracking anyway
			// The subscription might already be gone
			if err := tb.bus.Unsubscribe(id); err != nil {
				// Log but continue - we still want to clean up our tracking
				// This can happen if the subscription was already removed
				log.Printf("ToolkitBus.Unsubscribe: failed to unsubscribe %s: %v", id, err)
			}
			delete(listeners, listener)

			if len(listeners) == 0 {
				delete(tb.subscriptions, eventType)
			}
		}
	}
}

// Emit sends an event to all registered listeners
func (tb *ToolkitBus) Emit(event *GameEvent) error {
	// Map to toolkit event type
	toolkitEvent := mapEventType(event.Type)

	// Convert GameEvent to toolkit event
	tkEvent := convertToToolkitEvent(toolkitEvent, event)

	// Publish using toolkit
	return tb.bus.Publish(context.Background(), tkEvent)
}

// Clear removes all listeners
func (tb *ToolkitBus) Clear() {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// Unsubscribe all tracked subscriptions
	for _, listeners := range tb.subscriptions {
		for _, id := range listeners {
			// Best effort - ignore errors during cleanup
			if err := tb.bus.Unsubscribe(id); err != nil {
				// Already unsubscribed or other error - continue cleanup
				log.Printf("ToolkitBus.Clear: failed to unsubscribe %s: %v", id, err)
			}
		}
	}

	// Clear tracking
	tb.subscriptions = make(map[EventType]map[EventListener]string)
}

// ListenerCount returns the number of listeners for an event type
func (tb *ToolkitBus) ListenerCount(eventType EventType) int {
	tb.mu.RLock()
	defer tb.mu.RUnlock()

	if listeners, ok := tb.subscriptions[eventType]; ok {
		return len(listeners)
	}
	return 0
}

// mapEventType converts DND bot event types to toolkit event names
func mapEventType(eventType EventType) string {
	// Direct mapping to toolkit events
	mappings := map[EventType]string{
		BeforeAttackRoll:  rpgevents.EventBeforeAttackRoll,
		OnAttackRoll:      rpgevents.EventOnAttackRoll,
		AfterAttackRoll:   rpgevents.EventAfterAttackRoll,
		BeforeDamageRoll:  rpgevents.EventBeforeDamageRoll,
		OnDamageRoll:      rpgevents.EventOnDamageRoll,
		AfterDamageRoll:   rpgevents.EventAfterDamageRoll,
		BeforeTakeDamage:  rpgevents.EventBeforeTakeDamage,
		OnTakeDamage:      rpgevents.EventOnTakeDamage,
		AfterTakeDamage:   rpgevents.EventAfterTakeDamage,
		BeforeSavingThrow: rpgevents.EventBeforeSavingThrow,
		OnSavingThrow:     rpgevents.EventOnSavingThrow,
		AfterSavingThrow:  rpgevents.EventAfterSavingThrow,
		OnTurnStart:       rpgevents.EventOnTurnStart,
		OnTurnEnd:         rpgevents.EventOnTurnEnd,
		OnStatusApplied:   rpgevents.EventOnConditionApplied,
		OnStatusRemoved:   rpgevents.EventOnConditionRemoved,
		OnShortRest:       rpgevents.EventOnShortRest,
		OnLongRest:        rpgevents.EventOnLongRest,
		OnSpellCast:       rpgevents.EventOnSpellCast,
		OnSpellDamage:     rpgevents.EventOnSpellDamage,
	}

	if tkEvent, ok := mappings[eventType]; ok {
		return tkEvent
	}

	// Default: use the event type string with dndbot prefix
	return fmt.Sprintf("dndbot.%s", eventType.String())
}

// convertToToolkitEvent converts a GameEvent to toolkit Event
func convertToToolkitEvent(eventType string, gameEvent *GameEvent) rpgevents.Event {
	// Create appropriate source and target entities
	var source, target core.Entity

	if gameEvent.Actor != nil {
		source = WrapCharacter(gameEvent.Actor)
	}
	if gameEvent.Target != nil {
		target = WrapCharacter(gameEvent.Target)
	}

	// Create toolkit event
	event := rpgevents.NewGameEvent(eventType, source, target)

	// Copy context data
	for k, v := range gameEvent.Context {
		event.Context().Set(k, v)
	}

	// Handle cancellation
	if gameEvent.Cancelled {
		event.Cancel()
	}

	return event
}

// convertToGameEvent converts a toolkit Event back to GameEvent
func convertToGameEvent(tkEvent rpgevents.Event, expectedType EventType) *GameEvent {
	gameEvent := &GameEvent{
		Type:      expectedType,
		Context:   make(map[string]interface{}),
		Cancelled: tkEvent.IsCancelled(),
	}

	// Extract entities
	if source := tkEvent.Source(); source != nil {
		if charEntity, ok := source.(*CharacterEntity); ok {
			gameEvent.Actor = charEntity.Character
		}
	}
	if target := tkEvent.Target(); target != nil {
		if charEntity, ok := target.(*CharacterEntity); ok {
			gameEvent.Target = charEntity.Character
		}
	}

	// Copy modifiers as attack/damage bonuses
	modifiers := tkEvent.Context().Modifiers()
	attackBonus := 0
	damageBonus := 0

	for _, mod := range modifiers {
		switch mod.Type() {
		case rpgevents.ModifierAttackBonus:
			attackBonus += mod.ModifierValue().GetValue()
		case rpgevents.ModifierDamageBonus:
			damageBonus += mod.ModifierValue().GetValue()
		}
	}

	if attackBonus != 0 {
		gameEvent.Context["attack_bonus"] = attackBonus
	}
	if damageBonus != 0 {
		gameEvent.Context["damage_bonus"] = damageBonus
	}

	// Copy other context fields we know about
	knownFields := []string{
		"weapon", "damage", "damage_type", "spell_level",
		"dc", "ability", "save_type", "has_advantage",
		"has_disadvantage", "is_critical", "round_number",
	}

	for _, field := range knownFields {
		if value, ok := tkEvent.Context().Get(field); ok {
			gameEvent.Context[field] = value
		}
	}

	return gameEvent
}
