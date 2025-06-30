package modifiers

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/events"
)

// SourceType defines where a modifier comes from
type SourceType string

const (
	SourceTypeFeat         SourceType = "feat"
	SourceTypeSpell        SourceType = "spell"
	SourceTypeItem         SourceType = "item"
	SourceTypeClassFeature SourceType = "class_feature"
	SourceTypeRacialTrait  SourceType = "racial_trait"
	SourceTypeStatusEffect SourceType = "status_effect"
)

// Source describes where a modifier originates
type Source struct {
	Type SourceType
	Name string
	ID   string
}

// Duration defines how long a modifier lasts
type Duration interface {
	IsExpired() bool
	OnEventOccurred(event events.Event)
	String() string
}

// Modifier represents a game mechanic that modifies events
type Modifier interface {
	// ID returns a unique identifier for debugging/logging
	ID() string

	// Source returns where this modifier comes from
	Source() Source

	// Priority determines order of application (lower = earlier)
	Priority() int

	// Duration returns how long this modifier lasts
	Duration() Duration

	// IsActive checks if this modifier should apply
	IsActive(character *entities.Character) bool

	// AsVisitor returns the modifier as a visitor for applying effects
	AsVisitor() events.ModifierVisitor
}

// BaseModifier provides common implementation for modifiers
type BaseModifier struct {
	events.BaseModifierVisitor // Embed to get default no-op implementations
	id                         string
	source                     Source
	priority                   int
	duration                   Duration
}

func NewBaseModifier(id string, source Source, priority int, duration Duration) BaseModifier {
	return BaseModifier{
		id:       id,
		source:   source,
		priority: priority,
		duration: duration,
	}
}

func (m *BaseModifier) ID() string         { return m.id }
func (m *BaseModifier) Source() Source     { return m.source }
func (m *BaseModifier) Priority() int      { return m.priority }
func (m *BaseModifier) Duration() Duration { return m.duration }

// IsActive checks if the modifier is still active
func (m *BaseModifier) IsActive(character *entities.Character) bool {
	return m.duration != nil && !m.duration.IsExpired()
}

// AsVisitor returns self as the visitor
func (m *BaseModifier) AsVisitor() events.ModifierVisitor {
	return m
}
