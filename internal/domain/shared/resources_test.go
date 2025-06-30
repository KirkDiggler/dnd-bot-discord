package shared_test

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHPResource_Damage(t *testing.T) {
	tests := []struct {
		name           string
		hp             shared.HPResource
		damage         int
		expectedHP     shared.HPResource
		expectedDamage int
	}{
		{
			name: "damage absorbed by temp HP",
			hp: shared.HPResource{
				Current:   10,
				Max:       10,
				Temporary: 5,
			},
			damage: 3,
			expectedHP: shared.HPResource{
				Current:   10,
				Max:       10,
				Temporary: 2,
			},
			expectedDamage: 3,
		},
		{
			name: "damage exceeds temp HP",
			hp: shared.HPResource{
				Current:   10,
				Max:       10,
				Temporary: 2,
			},
			damage: 5,
			expectedHP: shared.HPResource{
				Current:   7,
				Max:       10,
				Temporary: 0,
			},
			expectedDamage: 5,
		},
		{
			name: "damage reduces to 0",
			hp: shared.HPResource{
				Current:   3,
				Max:       10,
				Temporary: 0,
			},
			damage: 5,
			expectedHP: shared.HPResource{
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
		hp             shared.HPResource
		healing        int
		expectedHP     shared.HPResource
		expectedHealed int
	}{
		{
			name: "heal partial damage",
			hp: shared.HPResource{
				Current: 5,
				Max:     10,
			},
			healing: 3,
			expectedHP: shared.HPResource{
				Current: 8,
				Max:     10,
			},
			expectedHealed: 3,
		},
		{
			name: "heal exceeds max",
			hp: shared.HPResource{
				Current: 8,
				Max:     10,
			},
			healing: 5,
			expectedHP: shared.HPResource{
				Current: 10,
				Max:     10,
			},
			expectedHealed: 2,
		},
		{
			name: "already at max HP",
			hp: shared.HPResource{
				Current: 10,
				Max:     10,
			},
			healing: 5,
			expectedHP: shared.HPResource{
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
		resources := &shared.CharacterResources{}
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
		resources := &shared.CharacterResources{}
		resources.Initialize(class, 1)

		// Check HP initialization
		assert.Equal(t, 10, resources.HP.Max)
		assert.Equal(t, 10, resources.HP.Current)

		// Check no spell slots
		assert.Empty(t, resources.SpellSlots)
	})

	t.Run("initialize ranger at level 1", func(t *testing.T) {
		class := testutils.CreateTestClass("ranger", "Ranger", 10)

		resources := &shared.CharacterResources{}
		resources.Initialize(class, 1)

		// Rangers don't get spell slots until level 2
		assert.Empty(t, resources.SpellSlots)
	})

	t.Run("initialize ranger at level 2", func(t *testing.T) {
		class := testutils.CreateTestClass("ranger", "Ranger", 10)

		resources := &shared.CharacterResources{}
		resources.Initialize(class, 2)

		// Rangers get spell slots at level 2
		require.Contains(t, resources.SpellSlots, 1)
		assert.Equal(t, 2, resources.SpellSlots[1].Max)
		assert.Equal(t, 2, resources.SpellSlots[1].Remaining)
		assert.Equal(t, "spellcasting", resources.SpellSlots[1].Source)
	})

	t.Run("initialize warlock resources", func(t *testing.T) {
		class := testutils.CreateTestClass("warlock", "Warlock", 8)

		resources := &shared.CharacterResources{}
		resources.Initialize(class, 1)

		// Warlocks get pact magic at level 1
		require.Contains(t, resources.SpellSlots, 1)
		assert.Equal(t, 1, resources.SpellSlots[1].Max)
		assert.Equal(t, 1, resources.SpellSlots[1].Remaining)
		assert.Equal(t, "pact_magic", resources.SpellSlots[1].Source)
	})
}

func TestCharacterResources_UseSpellSlot(t *testing.T) {
	resources := &shared.CharacterResources{
		SpellSlots: map[int]shared.SpellSlotInfo{
			1: {Max: 2, Remaining: 2, Source: "spellcasting"},
			2: {Max: 1, Remaining: 1, Source: "pact_magic"},
		},
	}

	// Use a 1st level slot
	success := resources.UseSpellSlot(1)
	assert.True(t, success)
	assert.Equal(t, 1, resources.SpellSlots[1].Remaining)
	assert.Equal(t, "spellcasting", resources.SpellSlots[1].Source) // Source preserved

	// Use another 1st level slot
	success = resources.UseSpellSlot(1)
	assert.True(t, success)
	assert.Equal(t, 0, resources.SpellSlots[1].Remaining)
	assert.Equal(t, "spellcasting", resources.SpellSlots[1].Source) // Source preserved

	// Try to use when none remaining
	success = resources.UseSpellSlot(1)
	assert.False(t, success)

	// Try to use non-existent level
	success = resources.UseSpellSlot(3)
	assert.False(t, success)

	// Verify pact magic source is preserved
	assert.True(t, resources.UseSpellSlot(2))
	assert.Equal(t, 0, resources.SpellSlots[2].Remaining)
	assert.Equal(t, "pact_magic", resources.SpellSlots[2].Source) // Source preserved
}

func TestCharacterResources_Rest(t *testing.T) {
	t.Run("short rest", func(t *testing.T) {
		resources := &shared.CharacterResources{
			HP: shared.HPResource{
				Current: 5,
				Max:     10,
			},
			SpellSlots: map[int]shared.SpellSlotInfo{
				1: {Max: 2, Remaining: 0, Source: "spellcasting"}, // Regular spell slots
				2: {Max: 1, Remaining: 0, Source: "pact_magic"},   // Warlock slots
			},
			Abilities: map[string]*shared.ActiveAbility{
				"second-wind": {
					RestType:      shared.RestTypeShort,
					UsesMax:       1,
					UsesRemaining: 0,
				},
				"rage": {
					RestType:      shared.RestTypeLong,
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

		// Regular spell slots should not restore on short rest
		assert.Equal(t, 0, resources.SpellSlots[1].Remaining)
		assert.Equal(t, "spellcasting", resources.SpellSlots[1].Source)

		// Warlock pact magic slots SHOULD restore on short rest
		assert.Equal(t, 1, resources.SpellSlots[2].Remaining)
		assert.Equal(t, "pact_magic", resources.SpellSlots[2].Source)
	})

	t.Run("long rest", func(t *testing.T) {
		resources := &shared.CharacterResources{
			HP: shared.HPResource{
				Current:   5,
				Max:       10,
				Temporary: 3,
			},
			HitDice: shared.HitDiceResource{
				DiceType:  8,
				Max:       4,
				Remaining: 1,
			},
			SpellSlots: map[int]shared.SpellSlotInfo{
				1: {Max: 2, Remaining: 0, Source: "spellcasting"},
				2: {Max: 1, Remaining: 0, Source: "pact_magic"},
			},
			Abilities: map[string]*shared.ActiveAbility{
				"rage": {
					RestType:      shared.RestTypeLong,
					UsesMax:       2,
					UsesRemaining: 0,
					IsActive:      true,
					Duration:      5,
				},
			},
			ActiveEffects: []*shared.ActiveEffect{
				{
					Name:         "Test Effect",
					DurationType: shared.DurationTypeUntilRest,
				},
				{
					Name:         "Permanent Effect",
					DurationType: shared.DurationTypePermanent,
				},
			},
		}

		resources.LongRest()

		// HP should restore to max
		assert.Equal(t, 10, resources.HP.Current)
		assert.Equal(t, 0, resources.HP.Temporary)

		// Half hit dice should restore (minimum 1)
		assert.Equal(t, 3, resources.HitDice.Remaining) // 1 + (4/2)

		// All spell slots should restore on long rest
		assert.Equal(t, 2, resources.SpellSlots[1].Remaining)
		assert.Equal(t, "spellcasting", resources.SpellSlots[1].Source)
		assert.Equal(t, 1, resources.SpellSlots[2].Remaining)
		assert.Equal(t, "pact_magic", resources.SpellSlots[2].Source)

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
		resources := &shared.CharacterResources{
			ActiveEffects: []*shared.ActiveEffect{
				{
					ID:                    "1",
					Name:                  "Shield of Faith",
					RequiresConcentration: true,
				},
			},
		}

		newEffect := &shared.ActiveEffect{
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
		resources := &shared.CharacterResources{
			ActiveEffects: []*shared.ActiveEffect{
				{
					Name:         "Shield",
					DurationType: shared.DurationTypeRounds,
					Duration:     2,
				},
				{
					Name:         "Bless",
					DurationType: shared.DurationTypeRounds,
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
		resources := &shared.CharacterResources{
			ActiveEffects: []*shared.ActiveEffect{
				{
					Name: "Shield of Faith",
					Modifiers: []shared.Modifier{
						{
							Type:  shared.ModifierTypeACBonus,
							Value: 2,
						},
					},
				},
				{
					Name: "Rage",
					Modifiers: []shared.Modifier{
						{
							Type:        shared.ModifierTypeDamageBonus,
							Value:       2,
							DamageTypes: []string{"melee"},
						},
						{
							Type:        shared.ModifierTypeDamageResistance,
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
		char := &character.Character{
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
		assert.Equal(t, shared.AbilityTypeBonusAction, rage.ActionType)
		assert.Equal(t, 2, rage.UsesMax)
		assert.Equal(t, 2, rage.UsesRemaining)
		assert.Equal(t, shared.RestTypeLong, rage.RestType)
		assert.Equal(t, 10, rage.Duration)
	})

	t.Run("fighter resources", func(t *testing.T) {
		char := &character.Character{
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
		assert.Equal(t, shared.AbilityTypeBonusAction, secondWind.ActionType)
		assert.Equal(t, 1, secondWind.UsesMax)
		assert.Equal(t, 1, secondWind.UsesRemaining)
		assert.Equal(t, shared.RestTypeShort, secondWind.RestType)
		assert.Equal(t, 0, secondWind.Duration) // Instant effect
	})

	t.Run("bard resources", func(t *testing.T) {
		char := &character.Character{
			Level:            1,
			MaxHitPoints:     8,
			CurrentHitPoints: 8,
			Class:            testutils.CreateTestClass("bard", "Bard", 8),
			Attributes: map[shared.Attribute]*character.AbilityScore{
				shared.AttributeCharisma: {Score: 16, Bonus: 3},
			},
		}

		char.InitializeResources()

		require.NotNil(t, char.Resources)

		// Check bardic inspiration
		bardicInspiration, exists := char.Resources.Abilities["bardic-inspiration"]
		require.True(t, exists)
		assert.Equal(t, "bardic-inspiration", bardicInspiration.Key)
		assert.Equal(t, shared.AbilityTypeBonusAction, bardicInspiration.ActionType)
		assert.Equal(t, 3, bardicInspiration.UsesMax) // Charisma modifier
		assert.Equal(t, 3, bardicInspiration.UsesRemaining)
		assert.Equal(t, shared.RestTypeLong, bardicInspiration.RestType)
		assert.Equal(t, 10, bardicInspiration.Duration) // 10 minutes
	})

	t.Run("paladin resources", func(t *testing.T) {
		char := &character.Character{
			Level:            1,
			MaxHitPoints:     10,
			CurrentHitPoints: 10,
			Class:            testutils.CreateTestClass("paladin", "Paladin", 10),
			Attributes: map[shared.Attribute]*character.AbilityScore{
				shared.AttributeCharisma: {Score: 14, Bonus: 2},
			},
		}

		char.InitializeResources()

		require.NotNil(t, char.Resources)

		// Check lay on hands
		layOnHands, exists := char.Resources.Abilities["lay-on-hands"]
		require.True(t, exists)
		assert.Equal(t, "lay-on-hands", layOnHands.Key)
		assert.Equal(t, shared.AbilityTypeAction, layOnHands.ActionType)
		assert.Equal(t, 5, layOnHands.UsesMax) // 5 HP per level
		assert.Equal(t, 5, layOnHands.UsesRemaining)
		assert.Equal(t, shared.RestTypeLong, layOnHands.RestType)
		assert.Equal(t, 0, layOnHands.Duration) // Instant

		// Check divine sense
		divineSense, exists := char.Resources.Abilities["divine-sense"]
		require.True(t, exists)
		assert.Equal(t, "divine-sense", divineSense.Key)
		assert.Equal(t, shared.AbilityTypeAction, divineSense.ActionType)
		assert.Equal(t, 3, divineSense.UsesMax) // 1 + Charisma modifier
		assert.Equal(t, 3, divineSense.UsesRemaining)
		assert.Equal(t, shared.RestTypeLong, divineSense.RestType)
		assert.Equal(t, 0, divineSense.Duration) // Until end of next turn
	})

	t.Run("bard with no charisma", func(t *testing.T) {
		char := &character.Character{
			Level:            1,
			MaxHitPoints:     8,
			CurrentHitPoints: 8,
			Class:            testutils.CreateTestClass("bard", "Bard", 8),
			// No attributes set
		}

		char.InitializeResources()

		require.NotNil(t, char.Resources)

		// Check bardic inspiration with 0 Charisma modifier
		bardicInspiration, exists := char.Resources.Abilities["bardic-inspiration"]
		require.True(t, exists)
		assert.Equal(t, 0, bardicInspiration.UsesMax) // 0 Charisma modifier
		assert.Equal(t, 0, bardicInspiration.UsesRemaining)
	})
}

func TestCharacter_GetResources(t *testing.T) {
	t.Run("lazy initialization", func(t *testing.T) {
		char := &character.Character{
			Level:            1,
			MaxHitPoints:     10,
			CurrentHitPoints: 10,
			Class:            testutils.CreateTestClass("fighter", "Fighter", 10),
		}

		// Resources should be nil initially
		assert.Nil(t, char.Resources)

		// GetResources should initialize and return resources
		resources := char.GetResources()
		require.NotNil(t, resources)
		assert.NotNil(t, char.Resources)

		// Should return same instance on subsequent calls
		resources2 := char.GetResources()
		assert.Same(t, resources, resources2)
	})

	t.Run("returns existing resources", func(t *testing.T) {
		char := &character.Character{
			Resources: &shared.CharacterResources{
				HP: shared.HPResource{
					Current: 5,
					Max:     10,
				},
			},
		}

		resources := char.GetResources()
		assert.NotNil(t, resources)
		assert.Equal(t, 5, resources.HP.Current)
		assert.Equal(t, 10, resources.HP.Max)
	})
}
