package character

import (
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"github.com/stretchr/testify/assert"
)

func TestCharacter_RageReset_OnLongRest(t *testing.T) {
	// Simulate the issue: Character with active rage and rage effects
	char := &Character{
		Name:  "Grunk",
		Level: 5,
		Resources: &CharacterResources{
			HP: shared.HPResource{
				Current: 20,
				Max:     45,
			},
			Abilities: map[string]*shared.ActiveAbility{
				shared.AbilityKeyRage: {
					Key:           shared.AbilityKeyRage,
					Name:          "Rage",
					UsesMax:       3,
					UsesRemaining: 2,
					IsActive:      true,
					Duration:      8, // 8 rounds remaining
					RestType:      shared.RestTypeLong,
				},
			},
			ActiveEffects: []*shared.ActiveEffect{
				{
					ID:           "rage-effect-1",
					Name:         "Rage",
					Source:       "barbarian_rage",
					DurationType: shared.DurationTypeRounds,
					Duration:     8,
					Modifiers: []shared.Modifier{
						{
							Type:        shared.ModifierTypeDamageBonus,
							Value:       2,
							DamageTypes: []string{"melee"},
						},
						{
							Type:        shared.ModifierTypeDamageResistance,
							Value:       1,
							DamageTypes: []string{"bludgeoning", "piercing", "slashing"},
						},
					},
				},
			},
		},
	}

	// Verify rage is active before long rest
	assert.True(t, char.Resources.Abilities[shared.AbilityKeyRage].IsActive, "Rage should be active before long rest")
	assert.Equal(t, 8, char.Resources.Abilities[shared.AbilityKeyRage].Duration, "Rage should have duration before long rest")
	assert.Len(t, char.Resources.ActiveEffects, 1, "Should have rage effect before long rest")

	// Perform long rest (simulating entering dungeon room)
	char.Resources.LongRest()

	// Verify rage is completely reset
	rage := char.Resources.Abilities[shared.AbilityKeyRage]
	assert.False(t, rage.IsActive, "Rage should be deactivated after long rest")
	assert.Equal(t, 0, rage.Duration, "Rage duration should be reset after long rest")
	assert.Equal(t, 3, rage.UsesRemaining, "Rage uses should be restored to max after long rest")

	// Verify all effects are cleared
	assert.Empty(t, char.Resources.ActiveEffects, "All active effects should be cleared after long rest")

	// Verify HP is restored
	assert.Equal(t, 45, char.Resources.HP.Current, "HP should be restored to max after long rest")
}

func TestCharacter_RageReset_MultipleEffects(t *testing.T) {
	// Test with multiple effects to ensure ALL are cleared
	char := &Character{
		Resources: &CharacterResources{
			HP: shared.HPResource{
				Current: 10,
				Max:     30,
			},
			Abilities: map[string]*shared.ActiveAbility{
				shared.AbilityKeyRage: {
					Key:           shared.AbilityKeyRage,
					Name:          "Rage",
					UsesMax:       3,
					UsesRemaining: 1,
					IsActive:      true,
					Duration:      5,
					RestType:      shared.RestTypeLong,
				},
			},
			ActiveEffects: []*shared.ActiveEffect{
				{
					Name:         "Rage",
					DurationType: shared.DurationTypeRounds,
					Duration:     5,
				},
				{
					Name:         "Bless",
					DurationType: shared.DurationTypeRounds,
					Duration:     10,
				},
				{
					Name:                  "Haste",
					DurationType:          shared.DurationTypeRounds,
					Duration:              10,
					RequiresConcentration: true,
				},
				{
					Name:         "Mage Armor",
					DurationType: shared.DurationTypeHours,
					Duration:     8,
				},
			},
		},
	}

	// Verify multiple effects exist
	assert.Len(t, char.Resources.ActiveEffects, 4, "Should have 4 effects before long rest")

	// Perform long rest
	char.Resources.LongRest()

	// Verify all non-permanent effects are cleared
	assert.Empty(t, char.Resources.ActiveEffects, "All non-permanent effects should be cleared after long rest")
	assert.False(t, char.Resources.Abilities[shared.AbilityKeyRage].IsActive, "Rage should be deactivated")
}
