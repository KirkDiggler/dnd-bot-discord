package features_test

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/equipment"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/features"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateAC_LeatherArmorWithDex(t *testing.T) {
	// Create a character with +4 DEX bonus
	char := &character.Character{
		ID:      "test-char",
		OwnerID: "test-owner",
		Name:    "Test Ranger",
		Level:   1,
		Attributes: map[shared.Attribute]*character.AbilityScore{
			shared.AttributeDexterity: {
				Score: 18, // +4 bonus
				Bonus: 4,
			},
		},
		EquippedSlots: make(map[shared.Slot]equipment.Equipment),
	}

	// Test 1: Leather armor WITH ArmorClass data
	leatherWithData := &equipment.Armor{
		Base: equipment.BasicEquipment{
			Key:  "leather-armor",
			Name: "Leather Armor",
		},
		ArmorCategory: equipment.ArmorCategoryLight,
		ArmorClass: &equipment.ArmorClass{
			Base:     11,
			DexBonus: true,
			MaxBonus: 0, // No max for light armor
		},
	}
	char.EquippedSlots[shared.SlotBody] = leatherWithData

	ac := features.CalculateAC(char)
	t.Logf("With ArmorClass data - BaseAC: 11, DexMod: 4, Total AC: %d", ac)
	assert.Equal(t, 15, ac, "AC with leather armor (with data) and +4 DEX should be 15")

	// Test 2: Leather armor WITHOUT ArmorClass data (fallback)
	leatherWithoutData := &equipment.Armor{
		Base: equipment.BasicEquipment{
			Key:  "leather-armor",
			Name: "Leather Armor",
		},
		ArmorCategory: equipment.ArmorCategoryLight,
		ArmorClass:    nil, // No AC data
	}
	char.EquippedSlots[shared.SlotBody] = leatherWithoutData

	ac = features.CalculateAC(char)
	assert.Equal(t, 15, ac, "AC with leather armor (fallback) and +4 DEX should be 15")
}

func TestCalculateAC_MediumArmorLimitsDex(t *testing.T) {
	// Create a character with +4 DEX bonus
	char := &character.Character{
		ID:      "test-char",
		OwnerID: "test-owner",
		Name:    "Test Fighter",
		Level:   1,
		Attributes: map[shared.Attribute]*character.AbilityScore{
			shared.AttributeDexterity: {
				Score: 18, // +4 bonus
				Bonus: 4,
			},
		},
		EquippedSlots: make(map[shared.Slot]equipment.Equipment),
	}

	// Test hide armor (medium) WITHOUT ArmorClass data (fallback)
	hideArmor := &equipment.Armor{
		Base: equipment.BasicEquipment{
			Key:  "hide-armor",
			Name: "Hide Armor",
		},
		ArmorCategory: equipment.ArmorCategoryMedium,
		ArmorClass:    nil,
	}
	char.EquippedSlots[shared.SlotBody] = hideArmor

	ac := features.CalculateAC(char)
	assert.Equal(t, 14, ac, "AC with hide armor should be 14 (12 base + 2 max DEX)")

	// Test scale mail (medium) WITHOUT ArmorClass data
	scaleMail := &equipment.Armor{
		Base: equipment.BasicEquipment{
			Key:  "scale-mail",
			Name: "Scale Mail",
		},
		ArmorCategory: equipment.ArmorCategoryMedium,
		ArmorClass:    nil,
	}
	char.EquippedSlots[shared.SlotBody] = scaleMail

	ac = features.CalculateAC(char)
	assert.Equal(t, 16, ac, "AC with scale mail should be 16 (14 base + 2 max DEX)")
}

func TestCalculateAC_ChainMailIgnoresDex(t *testing.T) {
	// Create a character with +4 DEX bonus
	char := &character.Character{
		ID:      "test-char",
		OwnerID: "test-owner",
		Name:    "Test Fighter",
		Level:   1,
		Attributes: map[shared.Attribute]*character.AbilityScore{
			shared.AttributeDexterity: {
				Score: 18, // +4 bonus
				Bonus: 4,
			},
		},
		EquippedSlots: make(map[shared.Slot]equipment.Equipment),
	}

	// Chain mail WITHOUT ArmorClass data (fallback)
	chainMail := &equipment.Armor{
		Base: equipment.BasicEquipment{
			Key:  "chain-mail",
			Name: "Chain Mail",
		},
		ArmorCategory: equipment.ArmorCategoryHeavy,
		ArmorClass:    nil,
	}
	char.EquippedSlots[shared.SlotBody] = chainMail

	ac := features.CalculateAC(char)
	assert.Equal(t, 16, ac, "AC with chain mail should be 16 (ignoring DEX)")
}
