package rulebook

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
)

type Race struct {
	Key                        string                  `json:"key"`
	Name                       string                  `json:"name"`
	Speed                      int                     `json:"speed"`
	StartingProficiencyOptions *shared.Choice          `json:"proficiency_choices"`
	StartingProficiencies      []*shared.ReferenceItem `json:"proficiencies"`
	AbilityBonuses             []*shared.AbilityBonus  `json:"ability_bonuses"`
}
