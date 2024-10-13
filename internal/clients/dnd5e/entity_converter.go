package dnd5e

import (
	"github.com/fadedpez/dnd5e-api"
	"your-username/your-repo/internal/entities"
)

func convertToInternalSpell(apiSpell *dnd5eapi.Spell) *entities.Spell {
	return &entities.Spell{
		Name:        apiSpell.Name,
		Description: apiSpell.Description,
		Level:       apiSpell.Level,
		// Add more fields as needed
	}
}

func convertToInternalMonster(apiMonster *dnd5eapi.Monster) *entities.Monster {
	return &entities.Monster{
		Name:        apiMonster.Name,
		Description: apiMonster.Description,
		HitPoints:   apiMonster.HitPoints,
		// Add more fields as needed
	}
}

// Add more conversion functions for other entity types as needed
