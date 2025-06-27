package entities_test

import (
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHPResource_Damage(t *testing.T) {
	tests := []struct {
		name           string
		hp             entities.HPResource
		damage         int
		expectedHP     entities.HPResource
		expectedDamage int
	}{
		{
			name: "damage absorbed by temp HP",
			hp: entities.HPResource{
				Current:   10,
				Max:       10,
				Temporary: 5,
			},
			damage: 3,
			expectedHP: entities.HPResource{
				Current:   10,
				Max:       10,
				Temporary: 2,
			},
			expectedDamage: 3,
		},
		{
			name: "damage exceeds temp HP",
			hp: entities.HPResource{
				Current:   10,
				Max:       10,
				Temporary: 2,
			},
			damage: 5,
			expectedHP: entities.HPResource{
				Current:   7,
				Max:       10,
				Temporary: 0,
			},
			expectedDamage: 5,
		},
		{
			name: "damage reduces to 0",
			hp: entities.HPResource{
				Current:   3,
				Max:       10,
				Temporary: 0,
			},
			damage: 5,
			expectedHP: entities.HPResource{
				Current:   0,
				Max:       10,
				Temporary: 0,
			},
			expectedDamage: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hp := tt.hp
			actualDamage := hp.Damage(tt.damage)
			assert.Equal(t, tt.expectedHP, hp)
			assert.Equal(t, tt.expectedDamage, actualDamage)
		})
	}
}

func TestHPResource_Heal(t *testing.T) {
	tests := []struct {
		name           string
		hp             entities.HPResource
		healing        int
		expectedHP     entities.HPResource
		expectedHealed int
	}{
		{
			name: "heal partial damage",
			hp: entities.HPResource{
				Current: 5,
				Max:     10,
			},
			healing: 3,
			expectedHP: entities.HPResource{
				Current: 8,
				Max:     10,
			},
			expectedHealed: 3,
		},
		{
			name: "heal exceeds max",
			hp: entities.HPResource{
				Current: 8,
				Max:     10,
			},
			healing: 5,
			expectedHP: entities.HPResource{
				Current: 10,
				Max:     10,
			},
			expectedHealed: 2,
		},
		{
			name: "already at max HP",
			hp: entities.HPResource{
				Current: 10,
				Max:     10,
			},
			healing: 5,
			expectedHP: entities.HPResource{
				Current: 10,
				Max:     10,
			},
			expectedHealed: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hp := tt.hp
			actualHealed := hp.Heal(tt.healing)
			assert.Equal(t, tt.expectedHP, hp)
			assert.Equal(t, tt.expectedHealed, actualHealed)
		})
	}
}

func TestCharacterResources_Initialize(t *testing.T) {
	t.Run("initialize cleric resources", func(t *testing.T) {
		class := testutils.CreateTestClass("cleric", "Cleric", 8)
		resources := &entities.CharacterResources{}
		resources.Initialize(class, 1)

		// Check HP initialization
		assert.Equal(t, 8, resources.HP.Max)
		assert.Equal(t, 8, resources.HP.Current)

		// Check hit dice
		assert.Equal(t, 8, resources.HitDice.DiceType)
		assert.Equal(t, 1, resources.HitDice.Max)
		assert.Equal(t, 1, resources.HitDice.Remaining)

		// Check spell slots
		require.NotNil(t, resources.SpellSlots)
		assert.Equal(t, 2, resources.SpellSlots[1].Max)
		assert.Equal(t, 2, resources.SpellSlots[1].Remaining)
	})

	t.Run("initialize fighter resources", func(t *testing.T) {
		class := testutils.CreateTestClass("fighter", "Fighter", 10)
		resources := &entities.CharacterResources{}
		resources.Initialize(class, 1)

		// Check HP initialization
		assert.Equal(t, 10, resources.HP.Max)
		assert.Equal(t, 10, resources.HP.Current)

		// Check no spell slots
		assert.Empty(t, resources.SpellSlots)
	})
}

func TestCharacterResources_UseSpellSlot(t *testing.T) {
	resources := &entities.CharacterResources{
		SpellSlots: map[int]entities.SpellSlotInfo{
			1: {Max: 2, Remaining: 2},
			2: {Max: 1, Remaining: 1},
		},
	}

	// Use a 1st level slot
	success := resources.UseSpellSlot(1)
	assert.True(t, success)
	assert.Equal(t, 1, resources.SpellSlots[1].Remaining)

	// Use another 1st level slot
	success = resources.UseSpellSlot(1)
	assert.True(t, success)
	assert.Equal(t, 0, resources.SpellSlots[1].Remaining)

	// Try to use when none remaining
	success = resources.UseSpellSlot(1)
	assert.False(t, success)

	// Try to use non-existent level
	success = resources.UseSpellSlot(3)
	assert.False(t, success)
}

func TestCharacterResources_Rest(t *testing.T) {
	t.Run("short rest", func(t *testing.T) {
		resources := &entities.CharacterResources{
			HP: entities.HPResource{
				Current: 5,
				Max:     10,
			},
			Abilities: map[string]*entities.ActiveAbility{
				"second-wind": {
					RestType:      entities.RestTypeShort,
					UsesMax:       1,
					UsesRemaining: 0,
				},
				"rage": {
					RestType:      entities.RestTypeLong,
					UsesMax:       2,
					UsesRemaining: 0,
				},
			},
		}

		resources.ShortRest()

		// HP should not change
		assert.Equal(t, 5, resources.HP.Current)

		// Short rest ability should restore
		assert.Equal(t, 1, resources.Abilities["second-wind"].UsesRemaining)

		// Long rest ability should not restore
		assert.Equal(t, 0, resources.Abilities["rage"].UsesRemaining)
	})

	t.Run("long rest", func(t *testing.T) {
		resources := &entities.CharacterResources{
			HP: entities.HPResource{
				Current:   5,
				Max:       10,
				Temporary: 3,
			},
			HitDice: entities.HitDiceResource{
				DiceType:  8,
				Max:       4,
				Remaining: 1,
			},
			SpellSlots: map[int]entities.SpellSlotInfo{
				1: {Max: 2, Remaining: 0},
			},
			Abilities: map[string]*entities.ActiveAbility{
				"rage": {
					RestType:      entities.RestTypeLong,
					UsesMax:       2,
					UsesRemaining: 0,
					IsActive:      true,
					Duration:      5,
				},
			},
			ActiveEffects: []*entities.ActiveEffect{
				{
					Name:         "Test Effect",
					DurationType: entities.DurationTypeUntilRest,
				},
				{
					Name:         "Permanent Effect",
					DurationType: entities.DurationTypePermanent,
				},
			},
		}

		resources.LongRest()

		// HP should restore to max
		assert.Equal(t, 10, resources.HP.Current)
		assert.Equal(t, 0, resources.HP.Temporary)

		// Half hit dice should restore (minimum 1)
		assert.Equal(t, 3, resources.HitDice.Remaining) // 1 + (4/2)

		// Spell slots should restore
		assert.Equal(t, 2, resources.SpellSlots[1].Remaining)

		// Abilities should restore and deactivate
		assert.Equal(t, 2, resources.Abilities["rage"].UsesRemaining)
		assert.False(t, resources.Abilities["rage"].IsActive)
		assert.Equal(t, 0, resources.Abilities["rage"].Duration)

		// Until rest effects should be removed
		assert.Len(t, resources.ActiveEffects, 1)
		assert.Equal(t, "Permanent Effect", resources.ActiveEffects[0].Name)
	})
}

func TestCharacterResources_Effects(t *testing.T) {
	t.Run("add concentration effect", func(t *testing.T) {
		resources := &entities.CharacterResources{
			ActiveEffects: []*entities.ActiveEffect{
				{
					ID:                    "1",
					Name:                  "Shield of Faith",
					RequiresConcentration: true,
				},
			},
		}

		newEffect := &entities.ActiveEffect{
			ID:                    "2",
			Name:                  "Bless",
			RequiresConcentration: true,
		}

		resources.AddEffect(newEffect)

		// Should only have the new concentration effect
		assert.Len(t, resources.ActiveEffects, 1)
		assert.Equal(t, "Bless", resources.ActiveEffects[0].Name)
	})

	t.Run("tick effect durations", func(t *testing.T) {
		resources := &entities.CharacterResources{
			ActiveEffects: []*entities.ActiveEffect{
				{
					Name:         "Shield",
					DurationType: entities.DurationTypeRounds,
					Duration:     2,
				},
				{
					Name:         "Bless",
					DurationType: entities.DurationTypeRounds,
					Duration:     1,
				},
			},
		}

		resources.TickEffectDurations()

		// Shield should still be active
		assert.Len(t, resources.ActiveEffects, 1)
		assert.Equal(t, "Shield", resources.ActiveEffects[0].Name)
		assert.Equal(t, 1, resources.ActiveEffects[0].Duration)
	})

	t.Run("calculate bonuses", func(t *testing.T) {
		resources := &entities.CharacterResources{
			ActiveEffects: []*entities.ActiveEffect{
				{
					Name: "Shield of Faith",
					Modifiers: []entities.Modifier{
						{
							Type:  entities.ModifierTypeACBonus,
							Value: 2,
						},
					},
				},
				{
					Name: "Rage",
					Modifiers: []entities.Modifier{
						{
							Type:        entities.ModifierTypeDamageBonus,
							Value:       2,
							DamageTypes: []string{"melee"},
						},
						{
							Type:        entities.ModifierTypeDamageResistance,
							DamageTypes: []string{"slashing", "piercing", "bludgeoning"},
						},
					},
				},
			},
		}

		assert.Equal(t, 2, resources.GetTotalACBonus())
		assert.Equal(t, 2, resources.GetTotalDamageBonus("melee"))
		assert.Equal(t, 0, resources.GetTotalDamageBonus("ranged"))
		assert.True(t, resources.HasResistance("slashing"))
		assert.False(t, resources.HasResistance("fire"))
	})
}

func TestCharacter_InitializeResources(t *testing.T) {
	t.Run("barbarian resources", func(t *testing.T) {
		char := &entities.Character{
			Level:            1,
			MaxHitPoints:     12,
			CurrentHitPoints: 12,
			Class:            testutils.CreateTestClass("barbarian", "Barbarian", 12),
		}

		char.InitializeResources()

		require.NotNil(t, char.Resources)
		assert.Equal(t, 12, char.Resources.HP.Max)
		assert.Equal(t, 12, char.Resources.HP.Current)

		// Check rage ability
		rage, exists := char.Resources.Abilities["rage"]
		require.True(t, exists)
		assert.Equal(t, "rage", rage.Key)
		assert.Equal(t, entities.AbilityTypeBonusAction, rage.ActionType)
		assert.Equal(t, 2, rage.UsesMax)
		assert.Equal(t, 2, rage.UsesRemaining)
		assert.Equal(t, entities.RestTypeLong, rage.RestType)
		assert.Equal(t, 10, rage.Duration)
	})

	t.Run("fighter resources", func(t *testing.T) {
		char := &entities.Character{
			Level:            1,
			MaxHitPoints:     10,
			CurrentHitPoints: 10,
			Class:            testutils.CreateTestClass("fighter", "Fighter", 10),
		}

		char.InitializeResources()

		require.NotNil(t, char.Resources)

		// Check second wind ability
		secondWind, exists := char.Resources.Abilities["second-wind"]
		require.True(t, exists)
		assert.Equal(t, "second-wind", secondWind.Key)
		assert.Equal(t, entities.AbilityTypeBonusAction, secondWind.ActionType)
		assert.Equal(t, 1, secondWind.UsesMax)
		assert.Equal(t, 1, secondWind.UsesRemaining)
		assert.Equal(t, entities.RestTypeShort, secondWind.RestType)
		assert.Equal(t, 0, secondWind.Duration) // Instant effect
	})
}
