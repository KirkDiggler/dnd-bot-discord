package characters

import (
	"encoding/json"
	"fmt"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/equipment"
	"strings"
)

// DataToEquipmentWithMigration converts EquipmentData back to Equipment interface
// with support for legacy data formats
func DataToEquipmentWithMigration(data EquipmentData) (equipment.Equipment, error) {
	// Normalize the type to lowercase for legacy compatibility
	normalizedType := strings.ToLower(data.Type)

	// Handle empty or unknown types
	if normalizedType == "" || normalizedType == "unknown" {
		// Try to detect type from the JSON structure
		var rawData map[string]interface{}
		if err := json.Unmarshal(data.Equipment, &rawData); err == nil {
			// Check for weapon-specific fields
			if _, hasWeaponCategory := rawData["weapon_category"]; hasWeaponCategory {
				normalizedType = "weapon"
			} else if _, hasArmorCategory := rawData["armor_category"]; hasArmorCategory {
				normalizedType = "armor"
			} else {
				normalizedType = "basic"
			}
		}
	}

	switch normalizedType {
	case "weapon":
		var weapon equipment.Weapon
		if err := json.Unmarshal(data.Equipment, &weapon); err != nil {
			return nil, fmt.Errorf("failed to unmarshal weapon: %w", err)
		}
		return &weapon, nil
	case "armor":
		var armor equipment.Armor
		if err := json.Unmarshal(data.Equipment, &armor); err != nil {
			return nil, fmt.Errorf("failed to unmarshal armor: %w", err)
		}
		return &armor, nil
	case "basic", "basicequipment":
		var basic equipment.BasicEquipment
		if err := json.Unmarshal(data.Equipment, &basic); err != nil {
			return nil, fmt.Errorf("failed to unmarshal basic equipment: %w", err)
		}
		return &basic, nil
	default:
		// Last resort: try basic equipment
		var basic equipment.BasicEquipment
		if err := json.Unmarshal(data.Equipment, &basic); err != nil {
			return nil, fmt.Errorf("unknown equipment type '%s' and failed to parse as basic: %w", data.Type, err)
		}
		return &basic, nil
	}
}
