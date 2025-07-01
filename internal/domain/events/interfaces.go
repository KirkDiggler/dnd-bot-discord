package events

//go:generate mockgen -destination=mock/mock_event_listener.go -package=mockevents -source=interfaces.go EventListener
//go:generate mockgen -destination=mock/mock_modifier.go -package=mockevents -source=interfaces.go Modifier
//go:generate mockgen -destination=mock/mock_modifier_duration.go -package=mockevents -source=interfaces.go ModifierDuration

// EventListener represents an object that can handle game events
type EventListener interface {
	HandleEvent(event *GameEvent) error
	Priority() int
}

// Modifier represents a game modifier that can affect events
type Modifier interface {
	// Unique identifier for debugging/logging
	ID() string

	// Source of the modifier (feat, spell, item, etc)
	Source() ModifierSource

	// Priority determines order of application (lower = earlier)
	Priority() int

	// Condition determines if this modifier applies to the event
	Condition(event *GameEvent) bool

	// Apply the modifier to the event
	Apply(event *GameEvent) error

	// Duration/expiration logic
	Duration() ModifierDuration
}

// ModifierDuration represents how long a modifier lasts
type ModifierDuration interface {
	IsExpired() bool
	OnEventOccurred(event *GameEvent)
}
