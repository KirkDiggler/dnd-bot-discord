package handlers

import "github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"

// dnd5eCreationStepTypes declares all character creation step types used by D&D 5e
// This serves as the authoritative list for UI validation
var dnd5eCreationStepTypes = []character.CreationStepType{
	// Core steps - all characters
	character.StepTypeRaceSelection,
	character.StepTypeClassSelection,
	character.StepTypeAbilityAssignment,
	character.StepTypeProficiencySelection,
	character.StepTypeEquipmentSelection,
	character.StepTypeCharacterDetails,

	// Class-specific features
	character.StepTypeDivineDomainSelection,    // Cleric
	character.StepTypeFightingStyleSelection,   // Fighter, Paladin, Ranger
	character.StepTypeFavoredEnemySelection,    // Ranger
	character.StepTypeNaturalExplorerSelection, // Ranger
	character.StepTypeSkillSelection,           // Various classes
	character.StepTypeLanguageSelection,        // Various classes
	character.StepTypeExpertiseSelection,       // Rogue, Bard

	// Spellcaster steps
	character.StepTypeCantripsSelection,    // Various casters
	character.StepTypeSpellSelection,       // Various casters
	character.StepTypeSpellbookSelection,   // Wizard
	character.StepTypeSpellsKnownSelection, // Sorcerer, Bard

	// Subclass steps (future implementation)
	character.StepTypeSubclassSelection,        // Generic subclass
	character.StepTypePatronSelection,          // Warlock
	character.StepTypeSorcerousOriginSelection, // Sorcerer
}
