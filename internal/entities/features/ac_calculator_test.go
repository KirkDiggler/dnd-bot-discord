package features_test

import (
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities/features"
	"github.com/stretchr/testify/assert"
)

func TestCalculateAC_LeatherArmorWithDex(t *testing.T) {
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
	}

	// Test 1: Leather armor WITH ArmorClass data
	leatherWithData := &entities.Armor{
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
	char.EquippedSlots[entities.SlotBody] = leatherWithData

	ac := features.CalculateAC(char)
	t.Logf("With ArmorClass data - BaseAC: 11, DexMod: 4, Total AC: %d", ac)
	assert.Equal(t, 15, ac, "AC with leather armor (with data) and +4 DEX should be 15")

	// Test 2: Leather armor WITHOUT ArmorClass data (fallback)
	leatherWithoutData := &entities.Armor{
		Base: entities.BasicEquipment{
			Key:  "leather-armor",
			Name: "Leather Armor",
		},
		ArmorCategory: entities.ArmorCategoryLight,
		ArmorClass:    nil, // No AC data
	}
	char.EquippedSlots[entities.SlotBody] = leatherWithoutData

	ac = features.CalculateAC(char)
	assert.Equal(t, 15, ac, "AC with leather armor (fallback) and +4 DEX should be 15")
}

func TestCalculateAC_MediumArmorLimitsDex(t *testing.T) {
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
	}

	// Test hide armor (medium) WITHOUT ArmorClass data (fallback)
	hideArmor := &entities.Armor{
		Base: entities.BasicEquipment{
			Key:  "hide-armor",
			Name: "Hide Armor",
		},
		ArmorCategory: entities.ArmorCategoryMedium,
		ArmorClass:    nil,
	}
	char.EquippedSlots[entities.SlotBody] = hideArmor

	ac := features.CalculateAC(char)
	assert.Equal(t, 14, ac, "AC with hide armor should be 14 (12 base + 2 max DEX)")

	// Test scale mail (medium) WITHOUT ArmorClass data
	scaleMail := &entities.Armor{
		Base: entities.BasicEquipment{
			Key:  "scale-mail",
			Name: "Scale Mail",
		},
		ArmorCategory: entities.ArmorCategoryMedium,
		ArmorClass:    nil,
	}
	char.EquippedSlots[entities.SlotBody] = scaleMail

	ac = features.CalculateAC(char)
	assert.Equal(t, 16, ac, "AC with scale mail should be 16 (14 base + 2 max DEX)")
}

func TestCalculateAC_ChainMailIgnoresDex(t *testing.T) {
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
	}

	// Chain mail WITHOUT ArmorClass data (fallback)
	chainMail := &entities.Armor{
		Base: entities.BasicEquipment{
			Key:  "chain-mail",
			Name: "Chain Mail",
		},
		ArmorCategory: entities.ArmorCategoryHeavy,
		ArmorClass:    nil,
	}
	char.EquippedSlots[entities.SlotBody] = chainMail

	ac := features.CalculateAC(char)
	assert.Equal(t, 16, ac, "AC with chain mail should be 16 (ignoring DEX)")
}
