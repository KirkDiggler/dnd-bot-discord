package entities

type StartingEquipment struct {
	Quantity  int            `json:"quantity"`
	Equipment *ReferenceItem `json:"equipment"`
}

type Class struct {
	Key                      string               `json:"key"`
	Name                     string               `json:"name"`
	HitDie                   int                  `json:"hit_die"`
	ProficiencyChoices       []*Choice            `json:"proficiency_choices"`
	StartingEquipmentChoices []*Choice            `json:"starting_equipment_choices"`
	Proficiencies            []*ReferenceItem     `json:"proficiencies"`
	StartingEquipment        []*StartingEquipment `json:"starting_equipment"`
}
