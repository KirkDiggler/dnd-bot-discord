package events

import (
	"fmt"
	"sort"
	"sync"
)

// EventBus manages event listeners and dispatches events
type EventBus struct {
	listeners map[EventType][]EventListener
	mu        sync.RWMutex
}

// NewEventBus creates a new event bus
func NewEventBus() *EventBus {
	return &EventBus{
		listeners: make(map[EventType][]EventListener),
	}
}

// Subscribe adds a listener for a specific event type
func (eb *EventBus) Subscribe(eventType EventType, listener EventListener) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.listeners[eventType] = append(eb.listeners[eventType], listener)
}

// Unsubscribe removes a listener for a specific event type
func (eb *EventBus) Unsubscribe(eventType EventType, listener EventListener) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	listeners := eb.listeners[eventType]
	for i, l := range listeners {
		if l == listener {
			// Remove by swapping with last and truncating
			listeners[i] = listeners[len(listeners)-1]
			eb.listeners[eventType] = listeners[:len(listeners)-1]
			break
		}
	}
}

// Emit sends an event to all registered listeners
func (eb *EventBus) Emit(event *GameEvent) error {
	listeners := eb.getListeners(event.Type)

	// Sort by priority
	sort.Slice(listeners, func(i, j int) bool {
		return listeners[i].Priority() < listeners[j].Priority()
	})

	// Execute listeners
	for _, listener := range listeners {
		if event.IsCancelled() {
			break
		}

		if err := listener.HandleEvent(event); err != nil {
			return fmt.Errorf("listener error: %w", err)
		}
	}

	return nil
}

// getListeners returns a copy of listeners for thread safety
func (eb *EventBus) getListeners(eventType EventType) []EventListener {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	original := eb.listeners[eventType]
	if len(original) == 0 {
		return nil
	}

	// Make a copy to avoid race conditions
	listeners := make([]EventListener, len(original))
	copy(listeners, original)

	return listeners
}

// Clear removes all listeners
func (eb *EventBus) Clear() {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.listeners = make(map[EventType][]EventListener)
}

// ListenerCount returns the number of listeners for an event type
func (eb *EventBus) ListenerCount(eventType EventType) int {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	return len(eb.listeners[eventType])
}
