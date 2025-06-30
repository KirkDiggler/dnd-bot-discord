package features

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/equipment"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnarmoredDefenseWithShield(t *testing.T) {
	t.Run("barbarian unarmored defense + shield", func(t *testing.T) {
		// Create a barbarian with unarmored defense
		char := &character.Character{
			Level: 5,
			Class: &rulebook.Class{Key: "barbarian"},
			Attributes: map[shared.Attribute]*character.AbilityScore{
				shared.AttributeDexterity:    {Score: 14, Bonus: 2},
				shared.AttributeConstitution: {Score: 18, Bonus: 4},
			},
			EquippedSlots: make(map[character.Slot]equipment.Equipment),
		}

		// Base unarmored defense
		ac := CalculateAC(char)
		assert.Equal(t, 16, ac, "Base unarmored defense: 10 + 2 (DEX) + 4 (CON) = 16")

		// Add shield
		char.EquippedSlots[character.SlotOffHand] = &equipment.Armor{
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
		assert.Equal(t, 18, ac, "Unarmored defense + shield: 10 + 2 (DEX) + 4 (CON) + 2 (shield) = 18")
	})

	t.Run("monk unarmored defense + shield", func(t *testing.T) {
		// Create a monk with unarmored defense
		char := &character.Character{
			Level: 5,
			Class: &rulebook.Class{Key: "monk"},
			Attributes: map[shared.Attribute]*character.AbilityScore{
				shared.AttributeDexterity: {Score: 18, Bonus: 4},
				shared.AttributeWisdom:    {Score: 16, Bonus: 3},
			},
			EquippedSlots: make(map[character.Slot]equipment.Equipment),
		}

		// Base unarmored defense
		ac := CalculateAC(char)
		assert.Equal(t, 17, ac, "Base unarmored defense: 10 + 4 (DEX) + 3 (WIS) = 17")

		// Add shield
		char.EquippedSlots[character.SlotOffHand] = &equipment.Armor{
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
		assert.Equal(t, 19, ac, "Unarmored defense + shield: 10 + 4 (DEX) + 3 (WIS) + 2 (shield) = 19")
	})

	t.Run("normal character with armor + shield", func(t *testing.T) {
		// Create a fighter with regular armor
		char := &character.Character{
			Level: 5,
			Class: &rulebook.Class{Key: "fighter"},
			Attributes: map[shared.Attribute]*character.AbilityScore{
				shared.AttributeDexterity: {Score: 14, Bonus: 2},
			},
			EquippedSlots: make(map[character.Slot]equipment.Equipment),
		}

		// Equip chain mail
		char.EquippedSlots[character.SlotBody] = &equipment.Armor{
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

		ac := CalculateAC(char)
		assert.Equal(t, 16, ac, "Chain mail AC: 16 (no DEX)")

		// Add shield
		char.EquippedSlots[character.SlotOffHand] = &equipment.Armor{
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
		assert.Equal(t, 18, ac, "Chain mail + shield: 16 + 2 = 18")
	})
}
