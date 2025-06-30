package characters

import (
	"encoding/json"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/equipment"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEquipmentDataMarshaling(t *testing.T) {
	// Test weapon marshaling
	weapon := &equipment.Weapon{
		Base: equipment.BasicEquipment{
			Key:  "longsword",
			Name: "Longsword",
		},
		WeaponCategory: "Martial",
		WeaponRange:    "Melee",
	}

	// Convert to EquipmentData
	data, err := equipmentToData(weapon)
	require.NoError(t, err)
	assert.Equal(t, "weapon", data.Type)

	// Marshal to JSON
	jsonData, err := json.Marshal(data)
	require.NoError(t, err)

	// Unmarshal back
	var unmarshaledData EquipmentData
	err = json.Unmarshal(jsonData, &unmarshaledData)
	require.NoError(t, err)

	// Convert back to Equipment
	equipment, err := dataToEquipment(unmarshaledData)
	require.NoError(t, err)

	// Verify it's still a weapon
	weaponBack, ok := equipment.(*equipment.Weapon)
	require.True(t, ok)
	assert.Equal(t, "longsword", weaponBack.Base.Key)
	assert.Equal(t, "Longsword", weaponBack.Base.Name)
}

func TestEquipmentDataArmor(t *testing.T) {
	// Test armor marshaling
	armor := &equipment.Armor{
		Base: equipment.BasicEquipment{
			Key:  "chainmail",
			Name: "Chainmail",
		},
		ArmorCategory: "Heavy",
		ArmorClass: &equipment.ArmorClass{
			Base: 16,
		},
	}

	// Convert to EquipmentData
	data, err := equipmentToData(armor)
	require.NoError(t, err)
	assert.Equal(t, "armor", data.Type)

	// Marshal to JSON
	jsonData, err := json.Marshal(data)
	require.NoError(t, err)

	// Unmarshal back
	var unmarshaledData EquipmentData
	err = json.Unmarshal(jsonData, &unmarshaledData)
	require.NoError(t, err)

	// Convert back to Equipment
	equipment, err := dataToEquipment(unmarshaledData)
	require.NoError(t, err)

	// Verify it's still armor
	armorBack, ok := equipment.(*equipment.Armor)
	require.True(t, ok)
	assert.Equal(t, "chainmail", armorBack.Base.Key)
	assert.Equal(t, "Chainmail", armorBack.Base.Name)
	assert.Equal(t, 16, armorBack.ArmorClass.Base)
}
