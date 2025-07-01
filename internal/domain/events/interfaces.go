package events

//go:generate mockgen -destination=mock/mock_events.go -package=mockevents -source=interfaces.go

// EventListener handles game events
type EventListener interface {
	// HandleEvent processes a game event
	HandleEvent(event *GameEvent) error

	// Priority determines the order of execution (lower executes first)
	Priority() int
}

// ModifierSource represents where a modifier comes from
type ModifierSource struct {
	Type        string // "class", "race", "item", "spell", "condition"
	Name        string // "Barbarian", "Half-Orc", "Sword of Sharpness"
	Description string // Optional description
}

// ModifierDuration defines how long a modifier lasts
type ModifierDuration interface {
	// IsExpired checks if the duration has expired
	IsExpired(event *GameEvent) bool

	// Description returns a human-readable description
	Description() string
}

// Modifier represents a game modifier that can affect events
type Modifier interface {
	// ID returns a unique identifier for this modifier
	ID() string

	// Source returns information about where this modifier comes from
	Source() ModifierSource

	// Priority determines the order of execution (lower executes first)
	Priority() int

	// Condition checks if this modifier should apply to the event
	Condition(event *GameEvent) bool

	// Apply modifies the event
	Apply(event *GameEvent) error

	// Duration returns how long this modifier lasts
	Duration() ModifierDuration
}
