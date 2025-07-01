package calculators

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/equipment"
	rulebook "github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e/features"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
)

// DnD5eACCalculator implements AC calculation following D&D 5e rules
type DnD5eACCalculator struct{}

// NewDnD5eACCalculator creates a new D&D 5e AC calculator
func NewDnD5eACCalculator() *DnD5eACCalculator {
	return &DnD5eACCalculator{}
}

// Calculate computes AC following D&D 5e rules
func (c *DnD5eACCalculator) Calculate(char *character.Character) int {
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
				// Check if it implements ACProvider interface
				if acProvider, ok := item.(equipment.ACProvider); ok {
					acBase := acProvider.GetACBase()
					if acBase < 0 {
						// No AC data, use fallback
						baseAC = c.getArmorACFallback(item.GetKey(), &dexMod)
					} else {
						baseAC = acBase
						// Special handling for leather armor - D&D 5e API may incorrectly set DexBonus to false
						if item.GetKey() == "leather-armor" || item.GetKey() == "studded-leather-armor" {
							// Light armor always uses full DEX bonus
							// Keep dexMod as is
						} else if !acProvider.UsesDexBonus() {
							dexMod = 0 // Heavy armor doesn't use DEX
						} else if maxBonus := acProvider.GetMaxDexBonus(); maxBonus > 0 {
							// Limit dex bonus for medium armor
							if dexMod > maxBonus {
								dexMod = maxBonus
							}
						}
					}
				} else {
					// Fallback for equipment that doesn't implement ACProvider
					baseAC = c.getArmorACFallback(item.GetKey(), &dexMod)
				}
				break
			}
		}
	}

	// Get class features
	classFeatures := []rulebook.CharacterFeature{}
	if char.Class != nil {
		classFeatures = features.GetClassFeatures(char.Class.Key, char.Level)
	}

	// Calculate AC based on armor or unarmored defense
	var ac int
	if !hasArmor {
		// Check for Unarmored Defense
		if features.HasFeature(classFeatures, "unarmored_defense_monk") {
			// Monk Unarmored Defense
			ac = 10 + dexMod + wisMod
		} else if features.HasFeature(classFeatures, "unarmored_defense_barbarian") {
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

	// Apply Defense fighting style (+1 AC while wearing armor)
	if hasArmor && c.hasDefenseFightingStyle(char) {
		ac++
	}

	// Apply any AC modifiers from active effects
	if char.Resources != nil && len(char.Resources.ActiveEffects) > 0 {
		for _, effect := range char.Resources.ActiveEffects {
			// Note: The effect interface has GetACBonus(), not GetACModifier()
			ac += effect.GetACBonus()
		}
	}

	return ac
}

// getArmorACFallback provides fallback AC values for armor without proper data
func (c *DnD5eACCalculator) getArmorACFallback(key string, dexMod *int) int {
	switch key {
	case "leather-armor":
		return 11
	case "studded-leather-armor":
		return 12
	case "hide-armor":
		if *dexMod > 2 {
			*dexMod = 2 // Medium armor max +2 DEX
		}
		return 12
	case "chain-shirt":
		if *dexMod > 2 {
			*dexMod = 2
		}
		return 13
	case "scale-mail":
		if *dexMod > 2 {
			*dexMod = 2
		}
		return 14
	case "breastplate":
		if *dexMod > 2 {
			*dexMod = 2
		}
		return 14
	case "half-plate", "half-plate-armor":
		if *dexMod > 2 {
			*dexMod = 2
		}
		return 15
	case "chain-mail":
		*dexMod = 0 // Heavy armor doesn't use DEX
		return 16
	case "plate-armor", "plate":
		*dexMod = 0 // Heavy armor doesn't use DEX
		return 18
	default:
		return 10
	}
}

// hasDefenseFightingStyle checks if the character has the Defense fighting style
func (c *DnD5eACCalculator) hasDefenseFightingStyle(char *character.Character) bool {
	if char.Features == nil {
		return false
	}

	for _, feature := range char.Features {
		if feature.Key == "fighting_style" && feature.Metadata != nil {
			// Check if the style is defense
			if styleValue, exists := feature.Metadata["style"]; exists {
				if style, isString := styleValue.(string); isString && style == "defense" {
					return true
				}
			}
		}
	}
	return false
}
