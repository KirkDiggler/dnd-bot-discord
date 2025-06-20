package entities_test

import (
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/stretchr/testify/assert"
)

func TestCharacter_EquipLeatherArmor_CalculatesAC(t *testing.T) {
	// Create a character with +4 DEX bonus
	char := &entities.Character{
		ID:      "test-char",
		OwnerID: "test-owner",
		Name:    "Test Ranger",
		Level:   1,
		Attributes: map[entities.Attribute]*entities.AbilityScore{
			entities.AttributeDexterity: {
				Score: 18, // +4 bonus
				Bonus: 4,
			},
		},
		EquippedSlots: make(map[entities.Slot]entities.Equipment),
		Inventory:     make(map[entities.EquipmentType][]entities.Equipment),
	}

	// Create leather armor
	leatherArmor := &entities.Armor{
		Base: entities.BasicEquipment{
			Key:  "leather-armor",
			Name: "Leather Armor",
		},
		ArmorCategory: entities.ArmorCategoryLight,
		ArmorClass: &entities.ArmorClass{
			Base:     11,
			DexBonus: true,
			MaxBonus: 0, // No max for light armor
		},
	}

	// Add armor to inventory
	char.Inventory[entities.EquipmentTypeArmor] = []entities.Equipment{leatherArmor}

	// Initial AC should be 10
	assert.Equal(t, 0, char.AC, "Initial AC should be 0 (not set)")

	// Equip the armor
	success := char.Equip("leather-armor")
	assert.True(t, success, "Should successfully equip leather armor")

	// Check if armor is actually equipped
	equippedArmor := char.EquippedSlots[entities.SlotBody]
	assert.NotNil(t, equippedArmor, "Armor should be equipped in body slot")

	// Debug print the actual armor
	if armor, ok := equippedArmor.(*entities.Armor); ok {
		t.Logf("Equipped armor: %s, AC base: %d, DexBonus: %v",
			armor.GetName(), armor.ArmorClass.Base, armor.ArmorClass.DexBonus)
	}

	// AC should be 11 (leather) + 4 (dex) = 15
	assert.Equal(t, 15, char.AC, "AC with leather armor and +4 DEX should be 15")
}

func TestCharacter_EquipChainMail_IgnoresDex(t *testing.T) {
	// Create a character with +4 DEX bonus
	char := &entities.Character{
		ID:      "test-char",
		OwnerID: "test-owner",
		Name:    "Test Fighter",
		Level:   1,
		Attributes: map[entities.Attribute]*entities.AbilityScore{
			entities.AttributeDexterity: {
				Score: 18, // +4 bonus
				Bonus: 4,
			},
		},
		EquippedSlots: make(map[entities.Slot]entities.Equipment),
		Inventory:     make(map[entities.EquipmentType][]entities.Equipment),
	}

	// Create chain mail (heavy armor)
	chainMail := &entities.Armor{
		Base: entities.BasicEquipment{
			Key:  "chain-mail",
			Name: "Chain Mail",
		},
		ArmorCategory: entities.ArmorCategoryHeavy,
		ArmorClass: &entities.ArmorClass{
			Base:     16,
			DexBonus: false, // Heavy armor doesn't use DEX
			MaxBonus: 0,
		},
	}

	// Add armor to inventory
	char.Inventory[entities.EquipmentTypeArmor] = []entities.Equipment{chainMail}

	// Equip the armor
	success := char.Equip("chain-mail")
	assert.True(t, success, "Should successfully equip chain mail")

	// AC should be 16 (no dex bonus for heavy armor)
	assert.Equal(t, 16, char.AC, "AC with chain mail should be 16 (ignoring DEX)")
}

func TestCharacter_EquipShield_AddsBonus(t *testing.T) {
	// Create a character with leather armor and shield
	char := &entities.Character{
		ID:      "test-char",
		OwnerID: "test-owner",
		Name:    "Test Fighter",
		Level:   1,
		Attributes: map[entities.Attribute]*entities.AbilityScore{
			entities.AttributeDexterity: {
				Score: 14, // +2 bonus
				Bonus: 2,
			},
		},
		EquippedSlots: make(map[entities.Slot]entities.Equipment),
		Inventory:     make(map[entities.EquipmentType][]entities.Equipment),
	}

	// Create leather armor
	leatherArmor := &entities.Armor{
		Base: entities.BasicEquipment{
			Key:  "leather-armor",
			Name: "Leather Armor",
		},
		ArmorCategory: entities.ArmorCategoryLight,
		ArmorClass: &entities.ArmorClass{
			Base:     11,
			DexBonus: true,
			MaxBonus: 0,
		},
	}

	// Create shield
	shield := &entities.Armor{
		Base: entities.BasicEquipment{
			Key:  "shield",
			Name: "Shield",
		},
		ArmorCategory: entities.ArmorCategoryShield,
		ArmorClass: &entities.ArmorClass{
			Base:     2,
			DexBonus: false,
			MaxBonus: 0,
		},
	}

	// Add to inventory
	char.Inventory[entities.EquipmentTypeArmor] = []entities.Equipment{leatherArmor, shield}

	// Equip leather armor first
	success := char.Equip("leather-armor")
	assert.True(t, success, "Should successfully equip leather armor")
	assert.Equal(t, 13, char.AC, "AC with leather armor and +2 DEX should be 13")

	// Equip shield
	success = char.Equip("shield")
	assert.True(t, success, "Should successfully equip shield")
	assert.Equal(t, 15, char.AC, "AC with leather armor, shield, and +2 DEX should be 15 (11+2+2)")
}
