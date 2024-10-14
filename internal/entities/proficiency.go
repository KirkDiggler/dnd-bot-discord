package entities

type ProficiencyType string

const (
	ProficiencyTypeArmor       ProficiencyType = "Armor"
	ProficiencyTypeWeapon      ProficiencyType = "Weapon"
	ProficiencyTypeTool        ProficiencyType = "Tool"
	ProficiencyTypeSavingThrow ProficiencyType = "Saving-Throw"
	ProficiencyTypeSkill       ProficiencyType = "Skill"
	ProficiencyTypeInstrument  ProficiencyType = "Instrument"
	ProficiencyTypeUnknown     ProficiencyType = "Other"
)

var ProficiencyTypes = []ProficiencyType{
	ProficiencyTypeArmor,
	ProficiencyTypeWeapon,
	ProficiencyTypeTool,
	ProficiencyTypeSavingThrow,
	ProficiencyTypeSkill,
	ProficiencyTypeInstrument,
	ProficiencyTypeUnknown,
}

type Proficiency struct {
	Key  string          `json:"key"`
	Name string          `json:"name"`
	Type ProficiencyType `json:"type"`
}

func (p *Proficiency) String() string {
	return p.Name
}
