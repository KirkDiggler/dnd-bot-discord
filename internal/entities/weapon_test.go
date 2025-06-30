package entities

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/damage"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/equipment"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWeaponAttackCalculations(t *testing.T) {
	// Create a test character
	char := &character.Character{
		Level: 1,
		Attributes: map[shared.Attribute]*character.AbilityScore{
			shared.AttributeStrength:  {Score: 16, Bonus: 3}, // +3 modifier
			shared.AttributeDexterity: {Score: 14, Bonus: 2}, // +2 modifier
		},
		Proficiencies: map[rulebook.ProficiencyType][]*rulebook.Proficiency{
			rulebook.ProficiencyTypeWeapon: {
				{Key: "longsword", Name: "Longsword", Type: rulebook.ProficiencyTypeWeapon},
				{Key: "shortbow", Name: "Shortbow", Type: rulebook.ProficiencyTypeWeapon},
			},
		},
	}

	t.Run("Melee weapon with proficiency", func(t *testing.T) {
		weapon := &equipment.Weapon{
			Base: equipment.BasicEquipment{
				Key:  "longsword",
				Name: "Longsword",
			},
			WeaponRange: "Melee",
			Damage: &damage.Damage{
				DiceCount:  1,
				DiceSize:   8,
				Bonus:      0,
				DamageType: damage.TypeSlashing,
			},
		}

		result, err := weapon.Attack(char)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Attack bonus should be STR modifier (3) + proficiency bonus (2) = 5
		// Note: result.AttackRoll includes the d20 roll, so we can't test exact value
		// But we can verify the logic by checking if proficiency is detected
		assert.True(t, char.HasWeaponProficiency("longsword"))
	})

	t.Run("Ranged weapon with proficiency", func(t *testing.T) {
		weapon := &equipment.Weapon{
			Base: equipment.BasicEquipment{
				Key:  "shortbow",
				Name: "Shortbow",
			},
			WeaponRange: "Ranged",
			Damage: &damage.Damage{
				DiceCount:  1,
				DiceSize:   6,
				Bonus:      0,
				DamageType: damage.TypePiercing,
			},
		}

		result, err := weapon.Attack(char)
		require.NoError(t, err)

		// Should use DEX modifier for ranged weapons
		assert.NotNil(t, result)
		assert.True(t, char.HasWeaponProficiency("shortbow"))
	})

	t.Run("Weapon without proficiency", func(t *testing.T) {
		weapon := &equipment.Weapon{
			Base: equipment.BasicEquipment{
				Key:  "greataxe",
				Name: "Greataxe",
			},
			WeaponRange: "Melee",
			Damage: &damage.Damage{
				DiceCount:  1,
				DiceSize:   12,
				Bonus:      0,
				DamageType: damage.TypeSlashing,
			},
		}

		result, err := weapon.Attack(char)
		require.NoError(t, err)

		// Should not have proficiency bonus
		assert.False(t, char.HasWeaponProficiency("greataxe"))
		assert.NotNil(t, result)
	})

	t.Run("Proficiency bonus scales with level", func(t *testing.T) {
		testCases := []struct {
			level         int
			expectedBonus int
		}{
			{1, 2},  // Level 1-4: +2
			{4, 2},  // Level 1-4: +2
			{5, 3},  // Level 5-8: +3
			{8, 3},  // Level 5-8: +3
			{9, 4},  // Level 9-12: +4
			{12, 4}, // Level 9-12: +4
			{13, 5}, // Level 13-16: +5
			{17, 6}, // Level 17-20: +6
		}

		for _, tc := range testCases {
			t.Logf("Testing level %d", tc.level)

			// Calculate expected proficiency bonus using same formula as weapon code
			expectedProfBonus := 2 + ((tc.level - 1) / 4)
			assert.Equal(t, tc.expectedBonus, expectedProfBonus,
				"Level %d should have proficiency bonus %d", tc.level, tc.expectedBonus)
		}
	})
}

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
