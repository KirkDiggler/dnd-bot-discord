package events

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
)

// GameEvent represents a game event that can be processed by listeners
type GameEvent struct {
	Type      EventType
	Actor     *character.Character
	Target    *character.Character
	Context   map[string]interface{}
	Modifiers []Modifier
	Cancelled bool
}

// NewGameEvent creates a new game event
func NewGameEvent(eventType EventType) *GameEvent {
	return &GameEvent{
		Type:    eventType,
		Context: make(map[string]interface{}),
	}
}

// Cancel marks the event as cancelled
func (e *GameEvent) Cancel() {
	e.Cancelled = true
}

// IsCancelled returns whether the event has been cancelled
func (e *GameEvent) IsCancelled() bool {
	return e.Cancelled
}

// WithActor sets the actor for the event
func (e *GameEvent) WithActor(actor *character.Character) *GameEvent {
	e.Actor = actor
	return e
}

// WithTarget sets the target for the event
func (e *GameEvent) WithTarget(target *character.Character) *GameEvent {
	e.Target = target
	return e
}

// WithContext adds a context value to the event
func (e *GameEvent) WithContext(key string, value interface{}) *GameEvent {
	e.Context[key] = value
	return e
}

// GetContext retrieves a context value
func (e *GameEvent) GetContext(key string) (interface{}, bool) {
	val, exists := e.Context[key]
	return val, exists
}

// GetIntContext retrieves an int context value
func (e *GameEvent) GetIntContext(key string) (value int, exists bool) {
	val, exists := e.Context[key]
	if !exists {
		return 0, false
	}
	intVal, ok := val.(int)
	return intVal, ok
}

// GetStringContext retrieves a string context value
func (e *GameEvent) GetStringContext(key string) (value string, exists bool) {
	val, exists := e.Context[key]
	if !exists {
		return "", false
	}
	strVal, ok := val.(string)
	return strVal, ok
}

// GetBoolContext retrieves a bool context value
func (e *GameEvent) GetBoolContext(key string) (value, exists bool) {
	val, exists := e.Context[key]
	if !exists {
		return false, false
	}
	boolVal, ok := val.(bool)
	return boolVal, ok
}
