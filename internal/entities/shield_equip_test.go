package entities

import (
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities/damage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCharacter_EquipShield(t *testing.T) {
	// Create a test character
	char := &Character{
		ID:   "test-char",
		Name: "Shield Bearer",
		Attributes: map[Attribute]*AbilityScore{
			AttributeDexterity: {Score: 14, Bonus: 2},
		},
		Inventory:     make(map[EquipmentType][]Equipment),
		EquippedSlots: make(map[Slot]Equipment),
	}

	// Add a shield to inventory
	shield := &Armor{
		Base: BasicEquipment{
			Key:  "shield",
			Name: "Shield",
		},
		ArmorCategory: ArmorCategoryShield,
		ArmorClass: &ArmorClass{
			Base:     2, // Standard shield provides +2 AC
			DexBonus: false,
		},
	}
	char.Inventory[EquipmentTypeArmor] = []Equipment{shield}

	// Test initial AC (should be 10 base, no DEX bonus without armor that allows it)
	char.calculateAC()
	assert.Equal(t, 10, char.AC, "Base AC should be 10 (no DEX bonus without armor)")

	// Equip the shield
	success := char.Equip("shield")
	require.True(t, success, "Should successfully equip shield")

	// Verify shield is in off-hand slot
	assert.NotNil(t, char.EquippedSlots[SlotOffHand], "Shield should be in off-hand slot")
	assert.Equal(t, "shield", char.EquippedSlots[SlotOffHand].GetKey(), "Shield key should match")

	// Verify AC increased by shield bonus (10 base + 2 shield)
	assert.Equal(t, 12, char.AC, "AC should be 10 base + 2 shield")
}

func TestCharacter_ShieldAndWeapon(t *testing.T) {
	// Create a test character
	char := &Character{
		ID:   "test-char",
		Name: "Sword and Board",
		Attributes: map[Attribute]*AbilityScore{
			AttributeDexterity: {Score: 14, Bonus: 2},
			AttributeStrength:  {Score: 16, Bonus: 3},
		},
		Inventory:     make(map[EquipmentType][]Equipment),
		EquippedSlots: make(map[Slot]Equipment),
	}

	// Add a weapon and shield to inventory
	sword := &Weapon{
		Base: BasicEquipment{
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
	shield := &Armor{
		Base: BasicEquipment{
			Key:  "shield",
			Name: "Shield",
		},
		ArmorCategory: ArmorCategoryShield,
		ArmorClass: &ArmorClass{
			Base:     2,
			DexBonus: false,
		},
	}

	char.Inventory[EquipmentTypeWeapon] = []Equipment{sword}
	char.Inventory[EquipmentTypeArmor] = []Equipment{shield}

	// Equip sword first
	success := char.Equip("longsword")
	require.True(t, success, "Should successfully equip longsword")
	assert.NotNil(t, char.EquippedSlots[SlotMainHand], "Sword should be in main hand")

	// Then equip shield
	success = char.Equip("shield")
	require.True(t, success, "Should successfully equip shield")
	assert.NotNil(t, char.EquippedSlots[SlotOffHand], "Shield should be in off-hand")

	// Both should be equipped
	assert.Equal(t, "longsword", char.EquippedSlots[SlotMainHand].GetKey())
	assert.Equal(t, "shield", char.EquippedSlots[SlotOffHand].GetKey())

	// AC should include shield bonus (10 base + 2 shield)
	assert.Equal(t, 12, char.AC, "AC should be 10 base + 2 shield")
}

func TestCharacter_ShieldPreventsOffhandWeapon(t *testing.T) {
	// Create a test character
	char := &Character{
		ID:            "test-char",
		Name:          "Conflicted Warrior",
		Inventory:     make(map[EquipmentType][]Equipment),
		EquippedSlots: make(map[Slot]Equipment),
		Attributes: map[Attribute]*AbilityScore{
			AttributeDexterity: {Score: 14, Bonus: 2},
		},
	}

	// Add two weapons and a shield
	sword := &Weapon{
		Base: BasicEquipment{
			Key:  "shortsword",
			Name: "Shortsword",
		},
		Properties: []*ReferenceItem{{Key: "light", Name: "Light", Type: ReferenceTypeWeaponProperty}},
	}
	dagger := &Weapon{
		Base: BasicEquipment{
			Key:  "dagger",
			Name: "Dagger",
		},
		Properties: []*ReferenceItem{{Key: "light", Name: "Light", Type: ReferenceTypeWeaponProperty}},
	}
	shield := &Armor{
		Base: BasicEquipment{
			Key:  "shield",
			Name: "Shield",
		},
		ArmorCategory: ArmorCategoryShield,
		ArmorClass:    &ArmorClass{Base: 2},
	}

	char.Inventory[EquipmentTypeWeapon] = []Equipment{sword, dagger}
	char.Inventory[EquipmentTypeArmor] = []Equipment{shield}

	// Equip main hand weapon
	char.Equip("shortsword")
	assert.NotNil(t, char.EquippedSlots[SlotMainHand])

	// Equip shield
	char.Equip("shield")
	assert.NotNil(t, char.EquippedSlots[SlotOffHand])
	assert.Equal(t, "shield", char.EquippedSlots[SlotOffHand].GetKey())

	// Try to equip dagger in off-hand (shield should be replaced)
	char.Equip("dagger")
	// Since dagger is a main hand weapon, it should move sword to off-hand,
	// replacing the shield
	assert.Equal(t, "dagger", char.EquippedSlots[SlotMainHand].GetKey())
	assert.Equal(t, "shortsword", char.EquippedSlots[SlotOffHand].GetKey())
}
