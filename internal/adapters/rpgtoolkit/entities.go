package rpgtoolkit

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/combat"
	"github.com/KirkDiggler/rpg-toolkit/core"
)

// EntityAdapter provides a common interface for adapting Discord bot entities
// to rpg-toolkit's core.Entity interface
type EntityAdapter interface {
	core.Entity
}

// CharacterEntityAdapter adapts a Discord bot Character to rpg-toolkit's Entity interface
type CharacterEntityAdapter struct {
	*character.Character
}

// GetID returns the character's ID
func (c *CharacterEntityAdapter) GetID() string {
	if c.Character == nil {
		return ""
	}
	return c.ID
}

// GetType returns the entity type
func (c *CharacterEntityAdapter) GetType() string {
	return "character"
}

// MonsterEntityAdapter adapts a combat monster to rpg-toolkit's Entity interface
type MonsterEntityAdapter struct {
	*combat.Combatant
}

// GetID returns the monster's ID
func (m *MonsterEntityAdapter) GetID() string {
	if m.Combatant == nil {
		return ""
	}
	return m.ID
}

// GetType returns the entity type
func (m *MonsterEntityAdapter) GetType() string {
	return "monster"
}

// CreateEntityAdapter creates an appropriate adapter based on the input type
func CreateEntityAdapter(entity interface{}) EntityAdapter {
	switch e := entity.(type) {
	case *character.Character:
		return &CharacterEntityAdapter{Character: e}
	case character.Character:
		return &CharacterEntityAdapter{Character: &e}
	case *combat.Combatant:
		return &MonsterEntityAdapter{Combatant: e}
	case combat.Combatant:
		return &MonsterEntityAdapter{Combatant: &e}
	default:
		// For now, return nil for unknown types
		// In the future, we might want to handle more entity types
		return nil
	}
}
