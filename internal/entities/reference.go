package entities

type ReferenceType string

const (
	ReferenceTypeAbilityScore   ReferenceType = "ability-score"
	ReferenceTypeEquipment      ReferenceType = "equipment"
	ReferenceTypeProficiency    ReferenceType = "proficiency"
	ReferenceTypeLanguage       ReferenceType = "language"
	ReferenceTypeSkill          ReferenceType = "skill"
	ReferenceTypeWeaponProperty ReferenceType = "weapon-properties"
	ReferenceTypeUnset          ReferenceType = ""
)

type ReferenceItem struct {
	Key  string        `json:"key"`
	Name string        `json:"name"`
	Type ReferenceType `json:"type"`
}
