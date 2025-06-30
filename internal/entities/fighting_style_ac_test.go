package entities

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/equipment"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefenseFightingStyleACCalculation(t *testing.T) {
	// This test demonstrates that Defense fighting style (+1 AC when wearing armor)
	// is not currently implemented in the calculateAC method

	t.Run("defense fighting style should add +1 AC with armor", func(t *testing.T) {
		char := &character.Character{
			Name:  "Defender",
			Level: 1,
			Class: &rulebook.Class{Key: "fighter"},
			Attributes: map[character.Attribute]*character.AbilityScore{
				character.AttributeDexterity: {Score: 14, Bonus: 2},
			},
			Features: []*rulebook.CharacterFeature{
				{
					Key:  "fighting_style",
					Name: "Fighting Style",
					Metadata: map[string]any{
						"style": "defense",
					},
				},
			},
			EquippedSlots: make(map[character.Slot]equipment.Equipment),
		}

		// Test 1: No armor = no defense bonus
		char.calculateAC()
		assert.Equal(t, 10, char.AC, "Base AC without armor should be 10")

		// Test 2: Light armor (leather)
		leather := &equipment.Armor{
			Base: equipment.BasicEquipment{
				Key:  "leather",
				Name: "Leather Armor",
			},
			ArmorClass: &equipment.ArmorClass{
				Base:     11,
				DexBonus: true,
			},
			ArmorCategory: "light",
		}
		char.EquippedSlots[character.SlotBody] = leather
		char.calculateAC()
		// Expected: 11 (leather) + 2 (DEX) + 1 (defense) = 14
		// calculateAC DOES apply defense fighting style via applyFightingStyleAC
		assert.Equal(t, 14, char.AC, "AC with leather armor and defense fighting style")

		// Test 3: Heavy armor (plate)
		plate := &equipment.Armor{
			Base: equipment.BasicEquipment{
				Key:  "plate",
				Name: "Plate Armor",
			},
			ArmorClass: &equipment.ArmorClass{
				Base:     18,
				DexBonus: false,
			},
			ArmorCategory: "heavy",
		}
		char.EquippedSlots[character.SlotBody] = plate
		char.calculateAC()
		// Expected: 18 (plate) + 1 (defense) = 19
		// calculateAC DOES apply defense fighting style
		assert.Equal(t, 19, char.AC, "AC with plate armor and defense fighting style")
	})

	t.Run("proposed calculateAC implementation with fighting styles", func(t *testing.T) {
		// This shows how calculateAC could be enhanced to support fighting styles
		// and other AC-modifying features

		enhancedCalculateAC := func(c *character.Character) int {
			ac := 10

			// First, check for body armor which sets the base AC
			if bodyArmor := c.EquippedSlots[character.SlotBody]; bodyArmor != nil {
				if armor, ok := bodyArmor.(*equipment.Armor); ok && armor.ArmorClass != nil {
					ac = armor.ArmorClass.Base
					if armor.ArmorClass.DexBonus {
						ac += c.Attributes[character.AttributeDexterity].Bonus
					}
				}
			}

			// Then add bonuses from other armor pieces (like shields)
			for slot, e := range c.EquippedSlots {
				if e == nil || slot == character.SlotBody {
					continue
				}
				if armor, ok := e.(*equipment.Armor); ok && armor.ArmorClass != nil {
					ac += armor.ArmorClass.Base
				}
			}

			// Apply fighting style bonuses
			for _, feature := range c.Features {
				if feature.Key == "fighting_style" && feature.Metadata != nil {
					if style, ok := feature.Metadata["style"].(string); ok && style == "defense" {
						// Defense: +1 AC while wearing armor
						if c.EquippedSlots[character.SlotBody] != nil {
							ac += 1
						}
					}
				}
			}

			// Apply other AC modifying features
			// - Barbarian/Monk unarmored defense
			// - Magic items
			// - Spells (Shield of Faith, etc.)

			return ac
		}

		// Test the enhanced function
		char := &character.Character{
			Attributes: map[character.Attribute]*character.AbilityScore{
				character.AttributeDexterity: {Score: 14, Bonus: 2},
			},
			Features: []*rulebook.CharacterFeature{
				{
					Key:  "fighting_style",
					Name: "Fighting Style",
					Metadata: map[string]any{
						"style": "defense",
					},
				},
			},
			EquippedSlots: make(map[character.Slot]equipment.Equipment),
		}

		// With chain mail
		char.EquippedSlots[character.SlotBody] = &equipment.Armor{
			ArmorClass: &equipment.ArmorClass{
				Base:     16,
				DexBonus: false,
			},
		}

		ac := enhancedCalculateAC(char)
		assert.Equal(t, 17, ac, "Chain mail (16) + defense (1) = 17")

		// With shield too
		char.EquippedSlots[character.SlotOffHand] = &equipment.Armor{
			ArmorClass: &equipment.ArmorClass{
				Base: 2,
			},
			ArmorCategory: "shield",
		}

		ac = enhancedCalculateAC(char)
		assert.Equal(t, 19, ac, "Chain mail (16) + shield (2) + defense (1) = 19")
	})
}

func TestMonkUnarmoredDefenseAC(t *testing.T) {
	t.Run("monk unarmored defense calculation", func(t *testing.T) {
		// Monk AC = 10 + DEX modifier + WIS modifier (when not wearing armor)
		char := &character.Character{
			Name:  "Monk",
			Level: 1,
			Class: &rulebook.Class{Key: "monk"},
			Attributes: map[character.Attribute]*character.AbilityScore{
				character.AttributeDexterity: {Score: 16, Bonus: 3},
				character.AttributeWisdom:    {Score: 15, Bonus: 2},
			},
			Features: []*rulebook.CharacterFeature{
				{
					Key:         "unarmored_defense_monk",
					Name:        "Unarmored Defense",
					Description: "AC equals 10 + DEX modifier + WIS modifier",
				},
			},
			EquippedSlots: make(map[character.Slot]equipment.Equipment),
		}

		char.calculateAC()
		// Current implementation: Just 10 (doesn't handle monk unarmored defense)
		// Should be: 10 + 3 (DEX) + 2 (WIS) = 15
		assert.Equal(t, 10, char.AC, "Monk unarmored defense not implemented")

		// If wearing armor, should use armor AC instead
		char.EquippedSlots[character.SlotBody] = &equipment.Armor{
			ArmorClass: &equipment.ArmorClass{
				Base:     12,
				DexBonus: true,
			},
		}
		char.calculateAC()
		// Should use armor: 12 + 3 (DEX) = 15
		assert.Equal(t, 15, char.AC, "Should use armor AC when worn")
	})
}

func TestBarbarianUnarmoredDefenseAC(t *testing.T) {
	t.Run("barbarian unarmored defense calculation", func(t *testing.T) {
		// Barbarian AC = 10 + DEX modifier + CON modifier (when not wearing armor)
		// Can still use a shield
		char := &character.Character{
			Name:  "Barbarian",
			Level: 1,
			Class: &rulebook.Class{Key: "barbarian"},
			Attributes: map[character.Attribute]*character.AbilityScore{
				character.AttributeDexterity:    {Score: 14, Bonus: 2},
				character.AttributeConstitution: {Score: 16, Bonus: 3},
			},
			Features: []*rulebook.CharacterFeature{
				{
					Key:         "unarmored_defense_barbarian",
					Name:        "Unarmored Defense",
					Description: "AC equals 10 + DEX modifier + CON modifier",
				},
			},
			EquippedSlots: make(map[character.Slot]equipment.Equipment),
		}

		char.calculateAC()
		// Current: Just 10
		// Should be: 10 + 2 (DEX) + 3 (CON) = 15
		assert.Equal(t, 10, char.AC, "Barbarian unarmored defense not implemented")

		// With shield (barbarian can use shield with unarmored defense)
		char.EquippedSlots[character.SlotOffHand] = &equipment.Armor{
			ArmorClass: &equipment.ArmorClass{
				Base: 2,
			},
			ArmorCategory: "shield",
		}
		char.calculateAC()
		// Current: 12 (10 + 2 from shield)
		// Should be: 10 + 2 (DEX) + 3 (CON) + 2 (shield) = 17
		assert.Equal(t, 12, char.AC, "Shield AC added but unarmored defense not applied")
	})
}
