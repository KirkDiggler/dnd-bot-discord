package entities

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/equipment"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCharacterEquipMethod(t *testing.T) {
	t.Run("Equip main hand weapon", func(t *testing.T) {
		char := &character.Character{
			Inventory: map[equipment.EquipmentType][]equipment.Equipment{
				equipment.EquipmentTypeWeapon: {
					&equipment.Weapon{
						Base: equipment.BasicEquipment{
							Key:  "longsword",
							Name: "Longsword",
						},
					},
				},
			},
		}

		success := char.Equip("longsword")
		assert.True(t, success)
		assert.NotNil(t, char.EquippedSlots[shared.SlotMainHand])
		assert.Equal(t, "longsword", char.EquippedSlots[shared.SlotMainHand].GetKey())
	})

	t.Run("Equip weapon not in inventory", func(t *testing.T) {
		char := &character.Character{
			Inventory: map[equipment.EquipmentType][]equipment.Equipment{
				equipment.EquipmentTypeWeapon: {},
			},
		}

		success := char.Equip("nonexistent-weapon")
		assert.False(t, success)
		assert.Nil(t, char.EquippedSlots[shared.SlotMainHand])
	})

	t.Run("Dual wielding weapons", func(t *testing.T) {
		mainHand := &equipment.Weapon{
			Base: equipment.BasicEquipment{Key: "shortsword1", Name: "Shortsword"},
		}
		offHand := &equipment.Weapon{
			Base: equipment.BasicEquipment{Key: "shortsword2", Name: "Shortsword"},
		}

		char := &character.Character{
			Inventory: map[equipment.EquipmentType][]equipment.Equipment{
				equipment.EquipmentTypeWeapon: {mainHand, offHand},
			},
		}

		// Equip main hand first
		success1 := char.Equip("shortsword1")
		assert.True(t, success1)
		assert.Equal(t, "shortsword1", char.EquippedSlots[shared.SlotMainHand].GetKey())

		// Equip second weapon - should move first to off hand
		success2 := char.Equip("shortsword2")
		assert.True(t, success2)
		assert.Equal(t, "shortsword2", char.EquippedSlots[shared.SlotMainHand].GetKey())
		assert.Equal(t, "shortsword1", char.EquippedSlots[shared.SlotOffHand].GetKey())
	})
}

func TestHasWeaponProficiency(t *testing.T) {
	char := &character.Character{
		Proficiencies: map[rulebook.ProficiencyType][]*rulebook.Proficiency{
			rulebook.ProficiencyTypeWeapon: {
				{Key: "longsword", Name: "Longsword", Type: rulebook.ProficiencyTypeWeapon},
				{Key: "shortbow", Name: "Shortbow", Type: rulebook.ProficiencyTypeWeapon},
			},
		},
	}

	t.Run("Has proficiency", func(t *testing.T) {
		assert.True(t, char.HasWeaponProficiency("longsword"))
		assert.True(t, char.HasWeaponProficiency("shortbow"))
	})

	t.Run("No proficiency", func(t *testing.T) {
		assert.False(t, char.HasWeaponProficiency("greataxe"))
		assert.False(t, char.HasWeaponProficiency("nonexistent"))
	})

	t.Run("No weapon proficiencies at all", func(t *testing.T) {
		charNoProficiencies := &character.Character{
			Proficiencies: map[rulebook.ProficiencyType][]*rulebook.Proficiency{},
		}
		assert.False(t, charNoProficiencies.HasWeaponProficiency("longsword"))
	})

	t.Run("Nil proficiencies map", func(t *testing.T) {
		charNilProficiencies := &character.Character{}
		assert.False(t, charNilProficiencies.HasWeaponProficiency("longsword"))
	})
}
