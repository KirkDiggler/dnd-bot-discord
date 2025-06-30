package characters

import (
	"encoding/json"
	"fmt"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/damage"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/equipment"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDataToEquipmentWithMigration(t *testing.T) {
	// Create test weapon data
	weaponJSON, err := json.Marshal(&equipment.Weapon{
		Base: equipment.BasicEquipment{
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
		validate     func(t *testing.T, eq equipment.Equipment)
	}{
		{
			name: "weapon with correct type",
			data: EquipmentData{
				Type:      "weapon",
				Equipment: weaponJSON,
			},
			expectedType: "*entities.Weapon",
			validate: func(t *testing.T, eq equipment.Equipment) {
				weapon, ok := eq.(*equipment.Weapon)
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
			validate: func(t *testing.T, eq equipment.Equipment) {
				weapon, ok := eq.(*equipment.Weapon)
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
			validate: func(t *testing.T, eq equipment.Equipment) {
				weapon, ok := eq.(*equipment.Weapon)
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
			validate: func(t *testing.T, eq equipment.Equipment) {
				weapon, ok := eq.(*equipment.Weapon)
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

	char := &character.Character{
		ID:      "test-char",
		OwnerID: "test-user",
		RealmID: "test-realm",
		Name:    "Test Character",
		EquippedSlots: map[shared.Slot]equipment.Equipment{
			shared.SlotMainHand: nil, // Nil equipment
			shared.SlotOffHand:  nil,
			shared.SlotTwoHanded: &equipment.Weapon{
				Base: equipment.BasicEquipment{
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
	_, hasMainHand := data.EquippedSlots[shared.SlotMainHand]
	assert.False(t, hasMainHand, "Nil MainHand slot should not be included")
	_, hasTwoHanded := data.EquippedSlots[shared.SlotTwoHanded]
	assert.True(t, hasTwoHanded, "Non-nil TwoHanded slot should be included")
}
