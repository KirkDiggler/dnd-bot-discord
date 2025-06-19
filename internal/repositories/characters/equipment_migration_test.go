package characters

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities/damage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDataToEquipmentWithMigration(t *testing.T) {
	// Create test weapon data
	weaponJSON, err := json.Marshal(&entities.Weapon{
		Base: entities.BasicEquipment{
			Key:    "greataxe",
			Name:   "Greataxe",
			Weight: 7,
		},
		WeaponCategory: "martial",
		WeaponRange:    "melee",
		Damage: &damage.Damage{
			DiceCount:  1,
			DiceSize:   12,
			DamageType: damage.TypeSlashing,
		},
	})
	require.NoError(t, err)

	tests := []struct {
		name         string
		data         EquipmentData
		expectedType string
		validate     func(t *testing.T, eq entities.Equipment)
	}{
		{
			name: "weapon with correct type",
			data: EquipmentData{
				Type:      "weapon",
				Equipment: weaponJSON,
			},
			expectedType: "*entities.Weapon",
			validate: func(t *testing.T, eq entities.Equipment) {
				weapon, ok := eq.(*entities.Weapon)
				require.True(t, ok)
				assert.Equal(t, "greataxe", weapon.GetKey())
				assert.Equal(t, "Greataxe", weapon.GetName())
				assert.Equal(t, "martial", weapon.WeaponCategory)
			},
		},
		{
			name: "weapon with empty type (legacy)",
			data: EquipmentData{
				Type:      "",
				Equipment: weaponJSON,
			},
			expectedType: "*entities.Weapon",
			validate: func(t *testing.T, eq entities.Equipment) {
				weapon, ok := eq.(*entities.Weapon)
				require.True(t, ok)
				assert.Equal(t, "greataxe", weapon.GetKey())
				assert.Equal(t, "martial", weapon.WeaponCategory)
			},
		},
		{
			name: "weapon with unknown type",
			data: EquipmentData{
				Type:      "unknown",
				Equipment: weaponJSON,
			},
			expectedType: "*entities.Weapon",
			validate: func(t *testing.T, eq entities.Equipment) {
				weapon, ok := eq.(*entities.Weapon)
				require.True(t, ok)
				assert.Equal(t, "greataxe", weapon.GetKey())
			},
		},
		{
			name: "weapon with uppercase type",
			data: EquipmentData{
				Type:      "WEAPON",
				Equipment: weaponJSON,
			},
			expectedType: "*entities.Weapon",
			validate: func(t *testing.T, eq entities.Equipment) {
				weapon, ok := eq.(*entities.Weapon)
				require.True(t, ok)
				assert.Equal(t, "greataxe", weapon.GetKey())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eq, err := DataToEquipmentWithMigration(tt.data)
			require.NoError(t, err)
			require.NotNil(t, eq)

			// Check type
			actualType := fmt.Sprintf("%T", eq)
			assert.Equal(t, tt.expectedType, actualType)

			// Run validation
			if tt.validate != nil {
				tt.validate(t, eq)
			}
		})
	}
}

func TestEquipmentNilHandling(t *testing.T) {
	// This test verifies that nil equipment is properly handled
	// in the toCharacterData function
	
	char := &entities.Character{
		ID:      "test-char",
		OwnerID: "test-user",
		RealmID: "test-realm",
		Name:    "Test Character",
		EquippedSlots: map[entities.Slot]entities.Equipment{
			entities.SlotMainHand:  nil, // Nil equipment
			entities.SlotOffHand:   nil,
			entities.SlotTwoHanded: &entities.Weapon{
				Base: entities.BasicEquipment{
					Key:  "longsword",
					Name: "Longsword",
				},
				WeaponCategory: "martial",
			},
		},
	}

	repo := &redisRepo{}
	data, err := repo.toCharacterData(char)
	require.NoError(t, err)

	// Verify that nil equipment slots are not included in the data
	assert.Len(t, data.EquippedSlots, 1, "Only non-nil equipment should be included")
	_, hasMainHand := data.EquippedSlots[entities.SlotMainHand]
	assert.False(t, hasMainHand, "Nil MainHand slot should not be included")
	_, hasTwoHanded := data.EquippedSlots[entities.SlotTwoHanded]
	assert.True(t, hasTwoHanded, "Non-nil TwoHanded slot should be included")
}