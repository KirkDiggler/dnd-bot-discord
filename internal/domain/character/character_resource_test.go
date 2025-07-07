package character

import (
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"github.com/stretchr/testify/assert"
)

func TestCharacterResources_LongRest_ClearsActiveAbilitiesAndEffects(t *testing.T) {
	// Create a character with an active rage and effects
	resources := &CharacterResources{
		HP: shared.HPResource{
			Current: 15,
			Max:     30,
		},
		Abilities: map[string]*shared.ActiveAbility{
			shared.AbilityKeyRage: {
				Key:           shared.AbilityKeyRage,
				Name:          "Rage",
				UsesMax:       3,
				UsesRemaining: 1,
				IsActive:      true,
				Duration:      5, // 5 rounds remaining
			},
			"second-wind": {
				Key:           "second-wind",
				Name:          "Second Wind",
				UsesMax:       1,
				UsesRemaining: 0,
				IsActive:      false,
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
				Name:                  "Shield of Faith",
				DurationType:          shared.DurationTypeMinutes,
				Duration:              10,
				RequiresConcentration: true,
			},
		},
		HitDice: shared.HitDiceResource{
			DiceType:  10,
			Max:       3,
			Remaining: 1,
		},
	}

	// Perform long rest
	resources.LongRest()

	// Assert HP is restored
	assert.Equal(t, 30, resources.HP.Current, "HP should be restored to max")
	assert.Equal(t, 0, resources.HP.Temporary, "Temporary HP should be cleared")

	// Assert abilities are restored and deactivated
	rage := resources.Abilities[shared.AbilityKeyRage]
	assert.Equal(t, 3, rage.UsesRemaining, "Rage uses should be restored to max")
	assert.False(t, rage.IsActive, "Rage should be deactivated")
	assert.Equal(t, 0, rage.Duration, "Rage duration should be reset")

	secondWind := resources.Abilities["second-wind"]
	assert.Equal(t, 1, secondWind.UsesRemaining, "Second Wind uses should be restored")

	// Assert ALL active effects are cleared
	assert.Empty(t, resources.ActiveEffects, "All active effects should be cleared after long rest")

	// Assert hit dice are restored (half of max, minimum 1)
	assert.Equal(t, 2, resources.HitDice.Remaining, "Should restore half hit dice (1 + 1 = 2)")
}

func TestCharacterResources_LongRest_RestoresSpellSlots(t *testing.T) {
	resources := &CharacterResources{
		HP: shared.HPResource{
			Current: 10,
			Max:     20,
		},
		SpellSlots: map[int]shared.SpellSlotInfo{
			1: {Max: 3, Remaining: 0, Source: "spellcasting"},
			2: {Max: 2, Remaining: 1, Source: "spellcasting"},
		},
	}

	resources.LongRest()

	// Assert spell slots are restored
	assert.Equal(t, 3, resources.SpellSlots[1].Remaining, "Level 1 spell slots should be restored")
	assert.Equal(t, 2, resources.SpellSlots[2].Remaining, "Level 2 spell slots should be restored")
}

func TestCharacterResources_Initialize(t *testing.T) {
	// Test with a barbarian class
	barbarian := &rulebook.Class{
		Key:    "barbarian",
		Name:   "Barbarian",
		HitDie: 12,
	}

	resources := &CharacterResources{}
	resources.Initialize(barbarian, 3)

	assert.Equal(t, 12, resources.HP.Current)
	assert.Equal(t, 12, resources.HP.Max)
	assert.Equal(t, 12, resources.HitDice.DiceType)
	assert.Equal(t, 3, resources.HitDice.Max)
	assert.Equal(t, 3, resources.HitDice.Remaining)
	assert.NotNil(t, resources.Abilities)
	assert.Empty(t, resources.SpellSlots) // Barbarians don't have spell slots
}

func TestCharacterResources_ShortRest(t *testing.T) {
	resources := &CharacterResources{
		Abilities: map[string]*shared.ActiveAbility{
			"action-surge": {
				Key:           "action-surge",
				RestType:      shared.RestTypeShort,
				UsesMax:       1,
				UsesRemaining: 0,
			},
			shared.AbilityKeyRage: {
				Key:           shared.AbilityKeyRage,
				RestType:      shared.RestTypeLong,
				UsesMax:       3,
				UsesRemaining: 1,
			},
		},
		SpellSlots: map[int]shared.SpellSlotInfo{
			1: {Max: 2, Remaining: 0, Source: "pact_magic"},   // Warlock slots
			2: {Max: 2, Remaining: 1, Source: "spellcasting"}, // Regular slots
		},
	}

	resources.ShortRest()

	// Assert short rest abilities are restored
	assert.Equal(t, 1, resources.Abilities["action-surge"].UsesRemaining, "Action Surge should be restored on short rest")
	assert.Equal(t, 1, resources.Abilities[shared.AbilityKeyRage].UsesRemaining, "Rage should not be restored on short rest")

	// Assert only pact magic slots are restored
	assert.Equal(t, 2, resources.SpellSlots[1].Remaining, "Pact magic slots should be restored")
	assert.Equal(t, 1, resources.SpellSlots[2].Remaining, "Regular spell slots should not be restored")
}
