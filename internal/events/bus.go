package events

import (
	"fmt"
	"log"
	"sort"
	"sync"
)

// EventListener processes events
type EventListener interface {
	HandleEvent(event Event) error
	Priority() int
	ID() string
}

// Bus manages event distribution
type Bus struct {
	listeners map[EventType][]EventListener
	mu        sync.RWMutex
}

// NewBus creates a new event bus
func NewBus() *Bus {
	return &Bus{
		listeners: make(map[EventType][]EventListener),
	}
}

// Subscribe adds a listener for specific event types
func (b *Bus) Subscribe(eventType EventType, listener EventListener) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.listeners[eventType] = append(b.listeners[eventType], listener)

	// Sort by priority
	sort.Slice(b.listeners[eventType], func(i, j int) bool {
		return b.listeners[eventType][i].Priority() < b.listeners[eventType][j].Priority()
	})

	log.Printf("EventBus: Subscribed listener %s to event %s with priority %d",
		listener.ID(), eventType, listener.Priority())
}

// Unsubscribe removes a listener
func (b *Bus) Unsubscribe(eventType EventType, listenerID string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	listeners := b.listeners[eventType]
	for i, l := range listeners {
		if l.ID() != listenerID {
			continue
		}
		// Remove by swapping with last and truncating
		listeners[i] = listeners[len(listeners)-1]
		b.listeners[eventType] = listeners[:len(listeners)-1]

		// Re-sort after removal
		sort.Slice(b.listeners[eventType], func(i, j int) bool {
			return b.listeners[eventType][i].Priority() < b.listeners[eventType][j].Priority()
		})

		log.Printf("EventBus: Unsubscribed listener %s from event %s", listenerID, eventType)
		return
	}
}

// Emit sends an event to all registered listeners
func (b *Bus) Emit(event Event) error {
	b.mu.RLock()
	listeners := make([]EventListener, len(b.listeners[event.GetType()]))
	copy(listeners, b.listeners[event.GetType()])
	b.mu.RUnlock()

	log.Printf("EventBus: Emitting event %s with %d listeners", event.GetType(), len(listeners))

	// Process listeners in priority order
	for _, listener := range listeners {
		if event.IsCancelled() {
			log.Printf("EventBus: Event %s cancelled, stopping propagation", event.GetType())
			break
		}

		if err := listener.HandleEvent(event); err != nil {
			return fmt.Errorf("listener %s failed: %w", listener.ID(), err)
		}
	}

	return nil
}

// Clear removes all listeners
func (b *Bus) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.listeners = make(map[EventType][]EventListener)
	log.Printf("EventBus: Cleared all listeners")
}
