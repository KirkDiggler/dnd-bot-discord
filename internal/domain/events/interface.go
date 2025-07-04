package events

// Bus is the interface for event bus implementations
type Bus interface {
	// Subscribe adds a listener for a specific event type
	Subscribe(eventType EventType, listener EventListener)

	// Unsubscribe removes a listener for a specific event type
	Unsubscribe(eventType EventType, listener EventListener)

	// Emit sends an event to all registered listeners
	Emit(event *GameEvent) error

	// Clear removes all listeners
	Clear()

	// ListenerCount returns the number of listeners for an event type
	ListenerCount(eventType EventType) int
}
