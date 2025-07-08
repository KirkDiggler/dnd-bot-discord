package rulebook

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
)

type StartingEquipment struct {
	Quantity  int                   `json:"quantity"`
	Equipment *shared.ReferenceItem `json:"equipment"`
}

type Class struct {
	Key                      string                  `json:"key"`
	Name                     string                  `json:"name"`
	HitDie                   int                     `json:"hit_die"`
	ProficiencyChoices       []*shared.Choice        `json:"proficiency_choices"`
	StartingEquipmentChoices []*shared.Choice        `json:"starting_equipment_choices"`
	Proficiencies            []*shared.ReferenceItem `json:"proficiencies"`
	StartingEquipment        []*StartingEquipment    `json:"starting_equipment"`
	PrimaryAbility           string                  `json:"primary_ability"`
}

// GetPrimaryAbility returns the primary ability for the class
// This could eventually be loaded from configuration or the D&D API
func (c *Class) GetPrimaryAbility() string {
	if c.PrimaryAbility != "" {
		return c.PrimaryAbility
	}

	// Fallback to hardcoded values for now
	switch c.Key {
	case "barbarian":
		return "Strength"
	case "bard":
		return "Charisma"
	case "cleric":
		return "Wisdom"
	case "druid":
		return "Wisdom"
	case "fighter":
		return "Strength or Dexterity"
	case "monk":
		return "Dexterity and Wisdom"
	case "paladin":
		return "Strength and Charisma"
	case "ranger":
		return "Dexterity and Wisdom"
	case "rogue":
		return "Dexterity"
	case "sorcerer":
		return "Charisma"
	case "warlock":
		return "Charisma"
	case "wizard":
		return "Intelligence"
	default:
		return ""
	}
}
