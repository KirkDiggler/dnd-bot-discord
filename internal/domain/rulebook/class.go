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
}
