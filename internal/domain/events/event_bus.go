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

// Subscribe adds a listener for specific event types
func (eb *EventBus) Subscribe(eventType EventType, listener EventListener) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.listeners[eventType] = append(eb.listeners[eventType], listener)
}

// Unsubscribe removes a listener for specific event types
func (eb *EventBus) Unsubscribe(eventType EventType, listener EventListener) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	listeners := eb.listeners[eventType]
	for i, l := range listeners {
		if l == listener {
			// Remove the listener by swapping with last and truncating
			listeners[i] = listeners[len(listeners)-1]
			eb.listeners[eventType] = listeners[:len(listeners)-1]
			break
		}
	}
}

// Emit fires an event to all registered listeners
func (eb *EventBus) Emit(event *GameEvent) error {
	if event == nil {
		return fmt.Errorf("cannot emit nil event")
	}

	listeners := eb.getListeners(event.Type)
	if len(listeners) == 0 {
		return nil
	}

	// Sort by priority
	sort.Slice(listeners, func(i, j int) bool {
		return listeners[i].Priority() < listeners[j].Priority()
	})

	// Execute listeners
	for _, listener := range listeners {
		if err := listener.HandleEvent(event); err != nil {
			return fmt.Errorf("error handling event %s: %w", event.Type, err)
		}
		if event.Cancelled {
			break
		}
	}

	return nil
}

// getListeners returns a copy of listeners for a specific event type
func (eb *EventBus) getListeners(eventType EventType) []EventListener {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	originalListeners := eb.listeners[eventType]
	if len(originalListeners) == 0 {
		return nil
	}

	// Create a copy to avoid race conditions
	listeners := make([]EventListener, len(originalListeners))
	copy(listeners, originalListeners)
	return listeners
}

// Clear removes all listeners
func (eb *EventBus) Clear() {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.listeners = make(map[EventType][]EventListener)
}

// ListenerCount returns the number of listeners for a specific event type
func (eb *EventBus) ListenerCount(eventType EventType) int {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	return len(eb.listeners[eventType])
}

// TotalListenerCount returns the total number of listeners across all event types
func (eb *EventBus) TotalListenerCount() int {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	total := 0
	for _, listeners := range eb.listeners {
		total += len(listeners)
	}
	return total
}
