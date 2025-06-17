package features

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
)

// CalculateAC calculates AC based on class features, armor, and abilities
func CalculateAC(character *entities.Character) int {
	if character == nil {
		return 10
	}
	
	// Get ability modifiers
	dexMod := 0
	wisMod := 0
	conMod := 0
	
	if dexScore, exists := character.Attributes[entities.AttributeDexterity]; exists && dexScore != nil {
		dexMod = dexScore.Bonus
	}
	if wisScore, exists := character.Attributes[entities.AttributeWisdom]; exists && wisScore != nil {
		wisMod = wisScore.Bonus
	}
	if conScore, exists := character.Attributes[entities.AttributeConstitution]; exists && conScore != nil {
		conMod = conScore.Bonus
	}
	
	// Check for armor
	hasArmor := false
	baseAC := 10
	
	// Check equipped slots for armor
	if character.EquippedSlots != nil {
		for slot, item := range character.EquippedSlots {
			if item != nil && slot == entities.SlotBody && item.GetEquipmentType() == "Armor" {
				hasArmor = true
				// Try to cast to Armor type to get AC values
				if armor, ok := item.(*entities.Armor); ok && armor.ArmorClass != nil {
					baseAC = armor.ArmorClass.Base
					if !armor.ArmorClass.DexBonus {
						dexMod = 0 // Heavy armor doesn't use DEX
					} else if armor.ArmorClass.MaxBonus > 0 {
						// Limit dex bonus for medium armor
						if dexMod > armor.ArmorClass.MaxBonus {
							dexMod = armor.ArmorClass.MaxBonus
						}
					}
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
	classFeatures := []entities.CharacterFeature{}
	if character.Class != nil {
		classFeatures = GetClassFeatures(character.Class.Key, character.Level)
	}
	
	// Check for Unarmored Defense
	if !hasArmor {
		// Monk Unarmored Defense
		if HasFeature(classFeatures, "unarmored_defense_monk") {
			return 10 + dexMod + wisMod
		}
		
		// Barbarian Unarmored Defense
		if HasFeature(classFeatures, "unarmored_defense_barbarian") {
			return 10 + dexMod + conMod
		}
	}
	
	// Standard AC calculation
	ac := baseAC + dexMod
	
	// Check for shield
	hasShield := false
	if character.EquippedSlots != nil {
		for _, item := range character.EquippedSlots {
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
func CalculateInitiativeBonus(character *entities.Character) int {
	if character == nil {
		return 0
	}
	
	dexMod := 0
	if dexScore, exists := character.Attributes[entities.AttributeDexterity]; exists && dexScore != nil {
		dexMod = dexScore.Bonus
	}
	
	// Could add features that modify initiative here
	// For example, Alert feat adds +5
	
	return dexMod
}

// CalculateSpeed calculates movement speed including racial modifiers
func CalculateSpeed(character *entities.Character) int {
	baseSpeed := 30 // Default for most races
	
	if character.Race != nil {
		switch character.Race.Key {
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