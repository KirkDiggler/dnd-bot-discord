package features

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/equipment"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
)

// CalculateAC calculates AC based on class features, armor, and abilities
func CalculateAC(char *character.Character) int {
	if char == nil {
		return 10
	}

	// Get ability modifiers
	dexMod := 0
	wisMod := 0
	conMod := 0

	if dexScore, exists := char.Attributes[shared.AttributeDexterity]; exists && dexScore != nil {
		dexMod = dexScore.Bonus
	}
	if wisScore, exists := char.Attributes[shared.AttributeWisdom]; exists && wisScore != nil {
		wisMod = wisScore.Bonus
	}
	if conScore, exists := char.Attributes[shared.AttributeConstitution]; exists && conScore != nil {
		conMod = conScore.Bonus
	}

	// Check for armor
	hasArmor := false
	baseAC := 10

	// Check equipped slots for armor
	if char.EquippedSlots != nil {
		for slot, item := range char.EquippedSlots {
			if item != nil && slot == shared.SlotBody && item.GetEquipmentType() == equipment.EquipmentTypeArmor {
				hasArmor = true
				// Try to cast to Armor type to get AC values
				if armor, ok := item.(*equipment.Armor); ok && armor.ArmorClass != nil {
					baseAC = armor.ArmorClass.Base
					// Special handling for leather armor - D&D 5e API may incorrectly set DexBonus to false
					if item.GetKey() == "leather-armor" || item.GetKey() == "studded-leather-armor" {
						// Light armor always uses full DEX bonus
						// Keep dexMod as is
					} else if !armor.ArmorClass.DexBonus {
						dexMod = 0 // Heavy armor doesn't use DEX
					} else if armor.ArmorClass.MaxBonus > 0 {
						// Limit dex bonus for medium armor
						if dexMod > armor.ArmorClass.MaxBonus {
							dexMod = armor.ArmorClass.MaxBonus
						}
					}
					// Note: If we reach here with valid ArmorClass data, we use it
				} else {
					// Fallback for armor without proper AC data
					switch item.GetKey() {
					case "leather-armor":
						baseAC = 11
					case "studded-leather-armor":
						baseAC = 12
					case "hide-armor":
						baseAC = 12
						if dexMod > 2 {
							dexMod = 2 // Medium armor max +2 DEX
						}
					case "chain-shirt":
						baseAC = 13
						if dexMod > 2 {
							dexMod = 2
						}
					case "scale-mail":
						baseAC = 14
						if dexMod > 2 {
							dexMod = 2
						}
					case "breastplate":
						baseAC = 14
						if dexMod > 2 {
							dexMod = 2
						}
					case "half-plate", "half-plate-armor":
						baseAC = 15
						if dexMod > 2 {
							dexMod = 2
						}
					case "chain-mail":
						baseAC = 16
						dexMod = 0 // Heavy armor doesn't use DEX
					case "plate-armor", "plate":
						baseAC = 18
						dexMod = 0 // Heavy armor doesn't use DEX
					default:
						baseAC = 10
					}
				}
				break
			}
		}
	}

	// Get class features
	classFeatures := []rulebook.CharacterFeature{}
	if char.Class != nil {
		classFeatures = GetClassFeatures(char.Class.Key, char.Level)
	}

	// Calculate AC based on armor or unarmored defense
	var ac int
	if !hasArmor {
		// Check for Unarmored Defense
		if HasFeature(classFeatures, "unarmored_defense_monk") {
			// Monk Unarmored Defense
			ac = 10 + dexMod + wisMod
		} else if HasFeature(classFeatures, "unarmored_defense_barbarian") {
			// Barbarian Unarmored Defense
			ac = 10 + dexMod + conMod
		} else {
			// No armor and no unarmored defense
			ac = 10 + dexMod
		}
	} else {
		// Standard AC with armor
		ac = baseAC + dexMod
	}

	// Check for shield (works with any AC calculation including unarmored defense)
	hasShield := false
	if char.EquippedSlots != nil {
		for _, item := range char.EquippedSlots {
			if item != nil && item.GetKey() == "shield" {
				hasShield = true
				break
			}
		}
	}

	if hasShield {
		ac += 2
	}

	return ac
}

// CalculateInitiativeBonus calculates initiative modifier
func CalculateInitiativeBonus(char *character.Character) int {
	if char == nil {
		return 0
	}

	dexMod := 0
	if dexScore, exists := char.Attributes[shared.AttributeDexterity]; exists && dexScore != nil {
		dexMod = dexScore.Bonus
	}

	// Could add features that modify initiative here
	// For example, Alert feat adds +5

	return dexMod
}

// CalculateSpeed calculates movement speed including racial modifiers
func CalculateSpeed(char *character.Character) int {
	baseSpeed := 30 // Default for most races

	if char.Race != nil {
		switch char.Race.Key {
		case "dwarf":
			baseSpeed = 25
		case "halfling":
			baseSpeed = 25
		case "gnome":
			baseSpeed = 25
		case "wood-elf":
			baseSpeed = 35
		}
	}

	// Could add features that modify speed here
	// For example, Monk's Unarmored Movement

	return baseSpeed
}
