package features

import (
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/stretchr/testify/assert"
)

func TestUnarmoredDefenseWithShield(t *testing.T) {
	t.Run("barbarian unarmored defense + shield", func(t *testing.T) {
		// Create a barbarian with unarmored defense
		char := &entities.Character{
			Level: 5,
			Class: &entities.Class{Key: "barbarian"},
			Attributes: map[entities.Attribute]*entities.AbilityScore{
				entities.AttributeDexterity:    {Score: 14, Bonus: 2},
				entities.AttributeConstitution: {Score: 18, Bonus: 4},
			},
			EquippedSlots: make(map[entities.Slot]entities.Equipment),
		}

		// Base unarmored defense
		ac := CalculateAC(char)
		assert.Equal(t, 16, ac, "Base unarmored defense: 10 + 2 (DEX) + 4 (CON) = 16")

		// Add shield
		char.EquippedSlots[entities.SlotOffHand] = &entities.Armor{
			Base: entities.BasicEquipment{
				Key:  "shield",
				Name: "Shield",
			},
			ArmorClass: &entities.ArmorClass{
				Base: 2,
			},
			ArmorCategory: "shield",
		}

		ac = CalculateAC(char)
		assert.Equal(t, 18, ac, "Unarmored defense + shield: 10 + 2 (DEX) + 4 (CON) + 2 (shield) = 18")
	})

	t.Run("monk unarmored defense + shield", func(t *testing.T) {
		// Create a monk with unarmored defense
		char := &entities.Character{
			Level: 5,
			Class: &entities.Class{Key: "monk"},
			Attributes: map[entities.Attribute]*entities.AbilityScore{
				entities.AttributeDexterity: {Score: 18, Bonus: 4},
				entities.AttributeWisdom:    {Score: 16, Bonus: 3},
			},
			EquippedSlots: make(map[entities.Slot]entities.Equipment),
		}

		// Base unarmored defense
		ac := CalculateAC(char)
		assert.Equal(t, 17, ac, "Base unarmored defense: 10 + 4 (DEX) + 3 (WIS) = 17")

		// Add shield
		char.EquippedSlots[entities.SlotOffHand] = &entities.Armor{
			Base: entities.BasicEquipment{
				Key:  "shield",
				Name: "Shield",
			},
			ArmorClass: &entities.ArmorClass{
				Base: 2,
			},
			ArmorCategory: "shield",
		}

		ac = CalculateAC(char)
		assert.Equal(t, 19, ac, "Unarmored defense + shield: 10 + 4 (DEX) + 3 (WIS) + 2 (shield) = 19")
	})

	t.Run("normal character with armor + shield", func(t *testing.T) {
		// Create a fighter with regular armor
		char := &entities.Character{
			Level: 5,
			Class: &entities.Class{Key: "fighter"},
			Attributes: map[entities.Attribute]*entities.AbilityScore{
				entities.AttributeDexterity: {Score: 14, Bonus: 2},
			},
			EquippedSlots: make(map[entities.Slot]entities.Equipment),
		}

		// Equip chain mail
		char.EquippedSlots[entities.SlotBody] = &entities.Armor{
			Base: entities.BasicEquipment{
				Key:  "chain-mail",
				Name: "Chain Mail",
			},
			ArmorClass: &entities.ArmorClass{
				Base:     16,
				DexBonus: false,
			},
			ArmorCategory: "heavy",
		}

		ac := CalculateAC(char)
		assert.Equal(t, 16, ac, "Chain mail AC: 16 (no DEX)")

		// Add shield
		char.EquippedSlots[entities.SlotOffHand] = &entities.Armor{
			Base: entities.BasicEquipment{
				Key:  "shield",
				Name: "Shield",
			},
			ArmorClass: &entities.ArmorClass{
				Base: 2,
			},
			ArmorCategory: "shield",
		}

		ac = CalculateAC(char)
		assert.Equal(t, 18, ac, "Chain mail + shield: 16 + 2 = 18")
	})
}
