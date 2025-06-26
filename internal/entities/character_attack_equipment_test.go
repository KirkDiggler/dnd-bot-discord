package entities_test

import (
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities/damage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCharacterAttackWithBasicEquipment reproduces the issue where
// BasicEquipment is in the weapon slot instead of a Weapon
func TestCharacterAttackWithBasicEquipment(t *testing.T) {
	tests := []struct {
		name          string
		equipped      entities.Equipment
		expectResults bool
		description   string
	}{
		{
			name: "weapon in main hand",
			equipped: &entities.Weapon{
				Base: entities.BasicEquipment{
					Key:  "greataxe",
					Name: "Greataxe",
				},
				WeaponRange: "Melee",
				Damage: &damage.Damage{
					DiceCount: 1,
					DiceSize:  12,
				},
			},
			expectResults: true,
			description:   "Should attack successfully with weapon",
		},
		{
			name: "basic equipment in main hand",
			equipped: &entities.BasicEquipment{
				Key:  "greataxe",
				Name: "Greataxe",
			},
			expectResults: false,
			description:   "Should not attack with basic equipment",
		},
		{
			name:          "empty basic equipment",
			equipped:      &entities.BasicEquipment{},
			expectResults: false,
			description:   "Should not attack with empty equipment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create character with equipped item
			char := &entities.Character{
				Name:  "Test Fighter",
				Level: 1,
				Attributes: map[entities.Attribute]*entities.AbilityScore{
					entities.AttributeStrength: {Score: 16, Bonus: 3},
				},
				EquippedSlots: map[entities.Slot]entities.Equipment{
					entities.SlotMainHand: tt.equipped,
				},
			}

			// Try to attack
			results, err := char.Attack()

			if tt.expectResults {
				require.NoError(t, err, tt.description)
				assert.NotEmpty(t, results, "Should have attack results")
			} else {
				// After fix: BasicEquipment now falls back to improvised melee
				require.NoError(t, err, tt.description)
				assert.NotEmpty(t, results, "Should fall back to improvised melee")
				assert.Equal(t, damage.TypeBludgeoning, results[0].AttackType, "Should be bludgeoning damage")
			}
		})
	}
}

// TestCharacterAttackFallbackBehavior tests what should happen when
// equipped item is not a weapon
func TestCharacterAttackFallbackBehavior(t *testing.T) {
	char := &entities.Character{
		Name:  "Test Fighter",
		Level: 1,
		Attributes: map[entities.Attribute]*entities.AbilityScore{
			entities.AttributeStrength: {Score: 16, Bonus: 3},
		},
		EquippedSlots: map[entities.Slot]entities.Equipment{
			entities.SlotMainHand: &entities.BasicEquipment{
				Key:  "torch",
				Name: "Torch",
			},
		},
	}

	results, err := char.Attack()

	// Should fall back to improvised weapon attack
	require.NoError(t, err)
	assert.NotEmpty(t, results, "Should fall back to improvised/unarmed attack")
	
	// Verify it's an improvised attack
	assert.Len(t, results, 1)
	assert.Equal(t, damage.TypeBludgeoning, results[0].AttackType, "Improvised attacks deal bludgeoning damage")
}
