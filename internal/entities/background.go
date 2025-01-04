package entities

// Background represents a D&D 5e character background
type Background struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`

	// Starting proficiencies and choices
	SkillProficiencies    []*Proficiency `json:"skill_proficiencies"`
	ToolProficiencies     []*Proficiency `json:"tool_proficiencies"`
	LanguageOptions       *Choice        `json:"language_options"`
	StartingProficiencies []*Proficiency `json:"starting_proficiencies"`

	// Starting equipment
	StartingEquipment        []Equipment `json:"starting_equipment"`
	StartingEquipmentOptions []*Choice   `json:"starting_equipment_options"`

	// Feature that comes with the background
	Feature *Feature `json:"feature"`

	// Personality customization options
	PersonalityTraits []string `json:"personality_traits"`
	Ideals            []string `json:"ideals"`
	Bonds             []string `json:"bonds"`
	Flaws             []string `json:"flaws"`

	// Suggested characteristics
	SuggestedCharacteristics []string `json:"suggested_characteristics"`
}

// Feature represents a special feature granted by a background
type Feature struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
