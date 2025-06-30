package character

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/equipment"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestInventoryDisplaysShields verifies that shields appear in inventory listing
func TestInventoryDisplaysShields(t *testing.T) {
	// Create a character with a shield in inventory
	char := &character.Character{
		ID:   "test-char",
		Name: "Shield Bearer",
		Inventory: map[equipment.EquipmentType][]equipment.Equipment{
			equipment.EquipmentTypeArmor: {
				&equipment.Armor{
					Base: equipment.BasicEquipment{
						Key:  "shield",
						Name: "Shield",
					},
					ArmorCategory: equipment.ArmorCategoryShield,
					ArmorClass: &equipment.ArmorClass{
						Base: 2,
					},
				},
			},
		},
	}

	// Verify shield is in armor inventory
	armor, exists := char.Inventory[equipment.EquipmentTypeArmor]
	assert.True(t, exists, "Should have armor inventory")
	assert.Len(t, armor, 1, "Should have one armor item")

	shield := armor[0]
	assert.Equal(t, "shield", shield.GetKey())
	assert.Equal(t, "Shield", shield.GetName())
	assert.Equal(t, equipment.EquipmentTypeArmor, shield.GetEquipmentType())
}
