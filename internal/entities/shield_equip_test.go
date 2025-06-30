package entities

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/damage"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/equipment"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCharacter_EquipShield(t *testing.T) {
	// Create a test character
	char := &character.Character{
		ID:   "test-char",
		Name: "Shield Bearer",
		Attributes: map[shared.Attribute]*character.AbilityScore{
			shared.AttributeDexterity: {Score: 14, Bonus: 2},
		},
		Inventory:     make(map[equipment.EquipmentType][]equipment.Equipment),
		EquippedSlots: make(map[character.Slot]equipment.Equipment),
	}

	// Add a shield to inventory
	shield := &equipment.Armor{
		Base: equipment.BasicEquipment{
			Key:  "shield",
			Name: "Shield",
		},
		ArmorCategory: equipment.ArmorCategoryShield,
		ArmorClass: &equipment.ArmorClass{
			Base:     2, // Standard shield provides +2 AC
			DexBonus: false,
		},
	}
	char.Inventory[equipment.EquipmentTypeArmor] = []equipment.Equipment{shield}

	// Test initial AC (should be 10 base, no DEX bonus without armor that allows it)
	char.calculateAC()
	assert.Equal(t, 10, char.AC, "Base AC should be 10 (no DEX bonus without armor)")

	// Equip the shield
	success := char.Equip("shield")
	require.True(t, success, "Should successfully equip shield")

	// Verify shield is in off-hand slot
	assert.NotNil(t, char.EquippedSlots[character.SlotOffHand], "Shield should be in off-hand slot")
	assert.Equal(t, "shield", char.EquippedSlots[character.SlotOffHand].GetKey(), "Shield key should match")

	// Verify AC increased by shield bonus (10 base + 2 shield)
	assert.Equal(t, 12, char.AC, "AC should be 10 base + 2 shield")
}

func TestCharacter_ShieldAndWeapon(t *testing.T) {
	// Create a test character
	char := &character.Character{
		ID:   "test-char",
		Name: "Sword and Board",
		Attributes: map[shared.Attribute]*character.AbilityScore{
			shared.AttributeDexterity: {Score: 14, Bonus: 2},
			shared.AttributeStrength:  {Score: 16, Bonus: 3},
		},
		Inventory:     make(map[equipment.EquipmentType][]equipment.Equipment),
		EquippedSlots: make(map[character.Slot]equipment.Equipment),
	}

	// Add a weapon and shield to inventory
	sword := &equipment.Weapon{
		Base: equipment.BasicEquipment{
			Key:  "longsword",
			Name: "Longsword",
		},
		WeaponCategory: "martial",
		WeaponRange:    "melee",
		Damage: &damage.Damage{
			DiceCount:  1,
			DiceSize:   8,
			Bonus:      0,
			DamageType: damage.TypeSlashing,
		},
	}
	shield := &equipment.Armor{
		Base: equipment.BasicEquipment{
			Key:  "shield",
			Name: "Shield",
		},
		ArmorCategory: equipment.ArmorCategoryShield,
		ArmorClass: &equipment.ArmorClass{
			Base:     2,
			DexBonus: false,
		},
	}

	char.Inventory[equipment.EquipmentTypeWeapon] = []equipment.Equipment{sword}
	char.Inventory[equipment.EquipmentTypeArmor] = []equipment.Equipment{shield}

	// Equip sword first
	success := char.Equip("longsword")
	require.True(t, success, "Should successfully equip longsword")
	assert.NotNil(t, char.EquippedSlots[character.SlotMainHand], "Sword should be in main hand")

	// Then equip shield
	success = char.Equip("shield")
	require.True(t, success, "Should successfully equip shield")
	assert.NotNil(t, char.EquippedSlots[character.SlotOffHand], "Shield should be in off-hand")

	// Both should be equipped
	assert.Equal(t, "longsword", char.EquippedSlots[character.SlotMainHand].GetKey())
	assert.Equal(t, "shield", char.EquippedSlots[character.SlotOffHand].GetKey())

	// AC should include shield bonus (10 base + 2 shield)
	assert.Equal(t, 12, char.AC, "AC should be 10 base + 2 shield")
}

func TestCharacter_ShieldPreventsOffhandWeapon(t *testing.T) {
	// Create a test character
	char := &character.Character{
		ID:            "test-char",
		Name:          "Conflicted Warrior",
		Inventory:     make(map[equipment.EquipmentType][]equipment.Equipment),
		EquippedSlots: make(map[character.Slot]equipment.Equipment),
		Attributes: map[shared.Attribute]*character.AbilityScore{
			shared.AttributeDexterity: {Score: 14, Bonus: 2},
		},
	}

	// Add two weapons and a shield
	sword := &equipment.Weapon{
		Base: equipment.BasicEquipment{
			Key:  "shortsword",
			Name: "Shortsword",
		},
		Properties: []*shared.ReferenceItem{{Key: "light", Name: "Light", Type: shared.ReferenceTypeWeaponProperty}},
	}
	dagger := &equipment.Weapon{
		Base: equipment.BasicEquipment{
			Key:  "dagger",
			Name: "Dagger",
		},
		Properties: []*shared.ReferenceItem{{Key: "light", Name: "Light", Type: shared.ReferenceTypeWeaponProperty}},
	}
	shield := &equipment.Armor{
		Base: equipment.BasicEquipment{
			Key:  "shield",
			Name: "Shield",
		},
		ArmorCategory: equipment.ArmorCategoryShield,
		ArmorClass:    &equipment.ArmorClass{Base: 2},
	}

	char.Inventory[equipment.EquipmentTypeWeapon] = []equipment.Equipment{sword, dagger}
	char.Inventory[equipment.EquipmentTypeArmor] = []equipment.Equipment{shield}

	// Equip main hand weapon
	char.Equip("shortsword")
	assert.NotNil(t, char.EquippedSlots[character.SlotMainHand])

	// Equip shield
	char.Equip("shield")
	assert.NotNil(t, char.EquippedSlots[character.SlotOffHand])
	assert.Equal(t, "shield", char.EquippedSlots[character.SlotOffHand].GetKey())

	// Try to equip dagger in off-hand (shield should be replaced)
	char.Equip("dagger")
	// Since dagger is a main hand weapon, it should move sword to off-hand,
	// replacing the shield
	assert.Equal(t, "dagger", char.EquippedSlots[character.SlotMainHand].GetKey())
	assert.Equal(t, "shortsword", char.EquippedSlots[character.SlotOffHand].GetKey())
}
