package character

import (
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/stretchr/testify/assert"
)

// TestInventoryDisplaysShields verifies that shields appear in inventory listing
func TestInventoryDisplaysShields(t *testing.T) {
	// Create a character with a shield in inventory
	char := &entities.Character{
		ID:   "test-char",
		Name: "Shield Bearer",
		Inventory: map[entities.EquipmentType][]entities.Equipment{
			entities.EquipmentTypeArmor: {
				&entities.Armor{
					Base: entities.BasicEquipment{
						Key:  "shield",
						Name: "Shield",
					},
					ArmorCategory: entities.ArmorCategoryShield,
					ArmorClass: &entities.ArmorClass{
						Base: 2,
					},
				},
			},
		},
	}

	// Verify shield is in armor inventory
	armor, exists := char.Inventory[entities.EquipmentTypeArmor]
	assert.True(t, exists, "Should have armor inventory")
	assert.Len(t, armor, 1, "Should have one armor item")

	shield := armor[0]
	assert.Equal(t, "shield", shield.GetKey())
	assert.Equal(t, "Shield", shield.GetName())
	assert.Equal(t, entities.EquipmentTypeArmor, shield.GetEquipmentType())
}
