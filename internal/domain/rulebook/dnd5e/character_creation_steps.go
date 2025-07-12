package rulebook

// CharacterCreationSteps defines all character creation step types used by D&D 5e
// This is the contract between the rulebook and UI layers - the UI must provide
// handlers for each of these step types.
// These values must match the CreationStepType constants in the character package.
var CharacterCreationSteps = []string{
	"race_selection",
	"class_selection",
	"ability_scores",
	"ability_assignment",
	"skill_selection",
	"language_selection",
	"proficiency_selection",
	"equipment_selection",
	"character_details",

	// Class-specific steps
	"fighting_style_selection",
	"divine_domain_selection",
	"favored_enemy_selection",
	"natural_explorer_selection",

	// Spellcaster steps
	"cantrips_selection",
	"spell_selection",

	// Final step
	"complete",
}
