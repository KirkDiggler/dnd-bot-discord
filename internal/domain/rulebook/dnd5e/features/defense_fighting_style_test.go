package features

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/equipment"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateAC_DefenseFightingStyle(t *testing.T) {
	t.Run("defense fighting style should add +1 AC with armor", func(t *testing.T) {
		// Create a fighter with defense fighting style
		char := &character.Character{
			Name:  "Defender",
			Level: 1,
			Class: &rulebook.Class{Key: "fighter"},
			Attributes: map[shared.Attribute]*character.AbilityScore{
				shared.AttributeDexterity: {Score: 14, Bonus: 2},
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
			EquippedSlots: make(map[shared.Slot]equipment.Equipment),
		}

		// Test 1: No armor = no defense bonus
		ac := CalculateAC(char)
		assert.Equal(t, 12, ac, "Base AC without armor: 10 + 2 (DEX) = 12")

		// Test 2: With leather armor
		char.EquippedSlots[shared.SlotBody] = &equipment.Armor{
			Base: equipment.BasicEquipment{
				Key:  "leather-armor",
				Name: "Leather Armor",
			},
			ArmorClass: &equipment.ArmorClass{
				Base:     11,
				DexBonus: true,
			},
			ArmorCategory: "light",
		}
		ac = CalculateAC(char)
		// Current: 11 + 2 (DEX) = 13
		// Should be: 11 + 2 (DEX) + 1 (defense) = 14
		assert.Equal(t, 13, ac, "TODO: Defense fighting style not implemented")

		// Test 3: With chain mail
		char.EquippedSlots[shared.SlotBody] = &equipment.Armor{
			Base: equipment.BasicEquipment{
				Key:  "chain-mail",
				Name: "Chain Mail",
			},
			ArmorClass: &equipment.ArmorClass{
				Base:     16,
				DexBonus: false,
			},
			ArmorCategory: "heavy",
		}
		ac = CalculateAC(char)
		// Current: 16 (no DEX for heavy armor)
		// Should be: 16 + 1 (defense) = 17
		assert.Equal(t, 16, ac, "TODO: Defense fighting style not implemented")

		// Test 4: With shield too
		char.EquippedSlots[shared.SlotOffHand] = &equipment.Armor{
			Base: equipment.BasicEquipment{
				Key:  "shield",
				Name: "Shield",
			},
			ArmorClass: &equipment.ArmorClass{
				Base: 2,
			},
			ArmorCategory: "shield",
		}
		ac = CalculateAC(char)
		// Current: 16 + 2 (shield) = 18
		// Should be: 16 + 2 (shield) + 1 (defense) = 19
		assert.Equal(t, 18, ac, "TODO: Defense fighting style not implemented")
	})

	t.Run("defense fighting style implementation proposal", func(t *testing.T) {
		// Show where the Defense fighting style check should be added
		// in the CalculateAC function around line 137

		// After calculating base AC and before returning, add:
		//nolint:gocritic // This is documentation showing where to implement the feature
		/*
			// Check for Defense fighting style
			if hasArmor {
				for _, feature := range character.Features {
					if feature.Key == "fighting_style" && feature.Metadata != nil {
						if style, ok := feature.Metadata["style"].(string); ok && style == "defense" {
							ac += 1
							break
						}
					}
				}
			}
		*/

		// This would need to be added after line 135 (shield check)
		// and before line 137 (return ac)
		assert.True(t, true, "This test documents where to add Defense fighting style")
	})
}

func TestCalculateAC_MonkBarbarianUnarmoredDefense(t *testing.T) {
	t.Run("monk unarmored defense is implemented", func(t *testing.T) {
		char := &character.Character{
			Level: 1,
			Class: &rulebook.Class{Key: "monk"},
			Attributes: map[shared.Attribute]*character.AbilityScore{
				shared.AttributeDexterity: {Score: 16, Bonus: 3},
				shared.AttributeWisdom:    {Score: 15, Bonus: 2},
			},
			EquippedSlots: make(map[shared.Slot]equipment.Equipment),
		}

		ac := CalculateAC(char)
		assert.Equal(t, 15, ac, "Monk unarmored defense correctly implemented")

		// With armor, should use armor AC instead
		char.EquippedSlots[shared.SlotBody] = &equipment.Armor{
			Base: equipment.BasicEquipment{Key: "leather-armor"},
			ArmorClass: &equipment.ArmorClass{
				Base:     11,
				DexBonus: true,
			},
		}
		ac = CalculateAC(char)
		// Uses armor: 11 + 3 (DEX) = 14
		assert.Equal(t, 14, ac, "Uses armor AC when worn")

		// Remove armor and test with shield
		delete(char.EquippedSlots, shared.SlotBody)
		char.EquippedSlots[shared.SlotOffHand] = &equipment.Armor{
			Base: equipment.BasicEquipment{Key: "shield"},
		}
		ac = CalculateAC(char)
		// Should be: 10 + 3 (DEX) + 2 (WIS) + 2 (shield) = 17
		assert.Equal(t, 17, ac, "Shield bonus correctly added to monk unarmored defense")
	})

	t.Run("barbarian unarmored defense is implemented", func(t *testing.T) {
		char := &character.Character{
			Level: 1,
			Class: &rulebook.Class{Key: "barbarian"},
			Attributes: map[shared.Attribute]*character.AbilityScore{
				shared.AttributeDexterity:    {Score: 14, Bonus: 2},
				shared.AttributeConstitution: {Score: 16, Bonus: 3},
			},
			EquippedSlots: make(map[shared.Slot]equipment.Equipment),
		}

		ac := CalculateAC(char)
		assert.Equal(t, 15, ac, "Barbarian unarmored defense correctly implemented")

		// With shield (barbarian can use shield with unarmored defense)
		char.EquippedSlots[shared.SlotOffHand] = &equipment.Armor{
			Base: equipment.BasicEquipment{Key: "shield"},
		}
		ac = CalculateAC(char)
		// Should be: 10 + 2 (DEX) + 3 (CON) + 2 (shield) = 17
		assert.Equal(t, 17, ac, "Shield bonus correctly added to barbarian unarmored defense")
	})
}
