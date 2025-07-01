package events

import "github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"

// GameEvent represents a game event that can be processed by the event system
type GameEvent struct {
	Type      EventType
	Actor     *character.Character
	Target    *character.Character
	Context   map[string]interface{} // Flexible context data
	Modifiers []Modifier             // Collected modifiers
	Cancelled bool                   // Events can be cancelled
}

// NewGameEvent creates a new game event
func NewGameEvent(eventType EventType, actor *character.Character) *GameEvent {
	return &GameEvent{
		Type:      eventType,
		Actor:     actor,
		Context:   make(map[string]interface{}),
		Modifiers: make([]Modifier, 0),
	}
}

// WithTarget sets the target for the event
func (e *GameEvent) WithTarget(target *character.Character) *GameEvent {
	e.Target = target
	return e
}

// WithContext adds context data to the event
func (e *GameEvent) WithContext(key string, value interface{}) *GameEvent {
	e.Context[key] = value
	return e
}

// Cancel marks the event as cancelled
func (e *GameEvent) Cancel() {
	e.Cancelled = true
}

// IsCancelled returns whether the event has been cancelled
func (e *GameEvent) IsCancelled() bool {
	return e.Cancelled
}

// AddModifier adds a modifier to the event
func (e *GameEvent) AddModifier(mod Modifier) {
	e.Modifiers = append(e.Modifiers, mod)
}

// GetContext retrieves a value from the context
func (e *GameEvent) GetContext(key string) (interface{}, bool) {
	val, exists := e.Context[key]
	return val, exists
}

// GetIntContext retrieves an int value from the context
func (e *GameEvent) GetIntContext(key string) (int, bool) {
	val, exists := e.Context[key]
	if !exists {
		return 0, false
	}
	intVal, ok := val.(int)
	return intVal, ok
}

// GetBoolContext retrieves a bool value from the context
func (e *GameEvent) GetBoolContext(key string) (value, exists bool) {
	val, exists := e.Context[key]
	if !exists {
		return false, false
	}
	boolVal, ok := val.(bool)
	return boolVal, ok
}

// GetStringContext retrieves a string value from the context
func (e *GameEvent) GetStringContext(key string) (string, bool) {
	val, exists := e.Context[key]
	if !exists {
		return "", false
	}
	strVal, ok := val.(string)
	return strVal, ok
}
