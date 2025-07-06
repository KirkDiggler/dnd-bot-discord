package rpgtoolkit

import (
	"fmt"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/rpg-toolkit/core"
	rpgevents "github.com/KirkDiggler/rpg-toolkit/events"
)

// EventHandler wraps a function to handle rpg-toolkit events
type EventHandler = rpgevents.HandlerFunc

// CharacterEntity wraps a D&D character to implement core.Entity
type CharacterEntity struct {
	Character *character.Character
}

// GetID returns the character's ID
func (ce *CharacterEntity) GetID() string {
	if ce.Character == nil {
		return ""
	}
	return ce.Character.ID
}

// Name returns the character's name
func (ce *CharacterEntity) Name() string {
	if ce.Character == nil {
		return ""
	}
	return ce.Character.Name
}

// GetType returns the entity type
func (ce *CharacterEntity) GetType() string {
	return "character"
}

// GetAttribute implements core.Entity
func (ce *CharacterEntity) GetAttribute(key string) (interface{}, bool) {
	if ce.Character == nil {
		return nil, false
	}

	switch key {
	case "level":
		return ce.Character.Level, true
	case "hp":
		return ce.Character.CurrentHitPoints, true
	case "max_hp":
		return ce.Character.MaxHitPoints, true
	case "ac":
		return ce.Character.AC, true
	case "class":
		if ce.Character.Class != nil {
			return ce.Character.Class.Name, true
		}
		return nil, false
	default:
		return nil, false
	}
}

// SetAttribute implements core.Entity
func (ce *CharacterEntity) SetAttribute(key string, value interface{}) error {
	if ce.Character == nil {
		return fmt.Errorf("character is nil")
	}

	switch key {
	case "hp":
		if hp, ok := value.(int); ok {
			ce.Character.CurrentHitPoints = hp
			return nil
		}
		return fmt.Errorf("hp must be int")
	default:
		return fmt.Errorf("attribute %s is read-only or unknown", key)
	}
}

// WrapCharacter creates a CharacterEntity from a character
func WrapCharacter(char *character.Character) core.Entity {
	if char == nil {
		return nil
	}
	return &CharacterEntity{Character: char}
}

// ExtractCharacter gets the character from an entity if it's a CharacterEntity
func ExtractCharacter(entity core.Entity) (*character.Character, bool) {
	if entity == nil {
		return nil, false
	}
	if ce, ok := entity.(*CharacterEntity); ok {
		return ce.Character, true
	}
	return nil, false
}

// EventListenerAdapter adapts old EventListener interface to rpg-toolkit handler
type EventListenerAdapter struct {
	eventType string
	priority  int
	handler   EventHandler
}

// NewEventListenerAdapter creates a new adapter for an event handler
func NewEventListenerAdapter(eventType string, priority int, handler EventHandler) *EventListenerAdapter {
	return &EventListenerAdapter{
		eventType: eventType,
		priority:  priority,
		handler:   handler,
	}
}

// Subscribe registers this listener with an rpg-toolkit event bus
func (ela *EventListenerAdapter) Subscribe(bus *rpgevents.Bus) string {
	return bus.SubscribeFunc(ela.eventType, ela.priority, ela.handler)
}

// Common event context keys
const (
	ContextWeapon          = "weapon"
	ContextDamage          = "damage"
	ContextDamageType      = "damage_type"
	ContextAttackBonus     = "attack_bonus"
	ContextTotalAttack     = "total_attack"
	ContextSaveType        = "save_type"
	ContextSaveBonus       = "save_bonus"
	ContextTotalSave       = "total_save"
	ContextHasAdvantage    = "has_advantage"
	ContextHasDisadvantage = "has_disadvantage"
	ContextIsCritical      = "is_critical"
	ContextSpellLevel      = "spell_level"
	ContextDC              = "dc"
	ContextAbility         = "ability"
	ContextRoundNumber     = "round_number"
)

// GetIntContext safely extracts an int from event context
func GetIntContext(event rpgevents.Event, key string) (int, bool) {
	if val, ok := event.Context().Get(key); ok {
		if intVal, ok := val.(int); ok {
			return intVal, true
		}
	}
	return 0, false
}

// GetStringContext safely extracts a string from event context
func GetStringContext(event rpgevents.Event, key string) (string, bool) {
	if val, ok := event.Context().Get(key); ok {
		if strVal, ok := val.(string); ok {
			return strVal, true
		}
	}
	return "", false
}

// GetBoolContext safely extracts a bool from event context
func GetBoolContext(event rpgevents.Event, key string) (value, exists bool) {
	if val, ok := event.Context().Get(key); ok {
		if boolVal, ok := val.(bool); ok {
			return boolVal, true
		}
	}
	return false, false
}
