package events

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/rpg-toolkit/core"
)

// CharacterEntity wraps a DND bot character to implement rpg-toolkit's Entity interface
type CharacterEntity struct {
	Character *character.Character
}

// Ensure CharacterEntity implements core.Entity
var _ core.Entity = (*CharacterEntity)(nil)

// GetID returns the character's unique identifier
func (c *CharacterEntity) GetID() string {
	if c.Character == nil {
		return ""
	}
	return c.Character.ID
}

// GetType returns the entity type
func (c *CharacterEntity) GetType() string {
	return "character"
}

// WrapCharacter creates a CharacterEntity from a Character
func WrapCharacter(char *character.Character) *CharacterEntity {
	if char == nil {
		return nil
	}
	return &CharacterEntity{Character: char}
}
