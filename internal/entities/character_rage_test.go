package entities

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/damage"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/equipment"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/effects"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCharacter_RageDamageBonus(t *testing.T) {
	// Create a barbarian character
	char := &character.Character{
		ID:    "barbarian_1",
		Name:  "Grog",
		Level: 1,
		Class: &rulebook.Class{
			Key:  "barbarian",
			Name: "Barbarian",
		},
		Attributes: map[shared.Attribute]*character.AbilityScore{
			shared.AttributeStrength: {Score: 16, Bonus: 3},
		},
	}

	// Create a simple weapon
	weapon := &equipment.Weapon{
		Base: equipment.BasicEquipment{
			Key:  "handaxe",
			Name: "Handaxe",
		},
		WeaponCategory: "simple",
		WeaponRange:    "Melee",
		Damage: &damage.Damage{
			DiceCount:  1,
			DiceSize:   6,
			DamageType: damage.TypeSlashing,
		},
	}

	// Add weapon to inventory and equip it
	char.AddInventory(weapon)
	char.Equip("handaxe")

	// Test 1: Get base damage bonus without rage
	baseBonus, err := char.applyActiveEffectDamageBonus(3, "melee") // STR bonus of 3
	require.NoError(t, err)
	assert.Equal(t, 3, baseBonus, "Without rage, damage bonus should just be ability modifier")

	// Test 2: Add rage effect and check damage bonus
	rageEffect := effects.BuildRageEffect(1) // Level 1 barbarian
	err = char.AddStatusEffect(rageEffect)
	require.NoError(t, err)

	rageBonus, err := char.applyActiveEffectDamageBonus(3, "melee") // STR bonus of 3
	require.NoError(t, err)
	assert.Equal(t, 5, rageBonus, "With rage at level 1, damage bonus should be 3 (STR) + 2 (rage) = 5")

	// Test 3: Verify rage doesn't apply to ranged attacks
	rangedBonus, err := char.applyActiveEffectDamageBonus(3, "ranged")
	require.NoError(t, err)
	assert.Equal(t, 3, rangedBonus, "Rage should not apply to ranged attacks")

	// Test 4: Test higher level rage bonuses
	char.RemoveStatusEffect(rageEffect.ID)

	// Level 9 barbarian
	char.Level = 9
	rageEffect9 := effects.BuildRageEffect(9)
	err = char.AddStatusEffect(rageEffect9)
	require.NoError(t, err)

	rageBonus9, err := char.applyActiveEffectDamageBonus(3, "melee")
	require.NoError(t, err)
	assert.Equal(t, 6, rageBonus9, "With rage at level 9, damage bonus should be 3 (STR) + 3 (rage) = 6")

	// Level 16 barbarian
	char.RemoveStatusEffect(rageEffect9.ID)
	char.Level = 16
	rageEffect16 := effects.BuildRageEffect(16)
	err = char.AddStatusEffect(rageEffect16)
	require.NoError(t, err)

	rageBonus16, err := char.applyActiveEffectDamageBonus(3, "melee")
	require.NoError(t, err)
	assert.Equal(t, 7, rageBonus16, "With rage at level 16, damage bonus should be 3 (STR) + 4 (rage) = 7")
}

func TestCharacter_RageDamageResistance(t *testing.T) {
	// Create a barbarian character
	char := &character.Character{
		ID:    "barbarian_1",
		Name:  "Grog",
		Level: 1,
		Class: &rulebook.Class{
			Key:  "barbarian",
			Name: "Barbarian",
		},
	}

	// Test damage without rage
	normalDamage := char.ApplyDamageResistance(damage.TypeSlashing, 10)
	assert.Equal(t, 10, normalDamage, "Without rage, damage should not be reduced")

	// Add rage effect
	rageEffect := effects.BuildRageEffect(1)
	err := char.AddStatusEffect(rageEffect)
	require.NoError(t, err)

	// Test physical damage types with rage
	slashingDamage := char.ApplyDamageResistance(damage.TypeSlashing, 10)
	assert.Equal(t, 5, slashingDamage, "With rage, slashing damage should be halved")

	piercingDamage := char.ApplyDamageResistance(damage.TypePiercing, 10)
	assert.Equal(t, 5, piercingDamage, "With rage, piercing damage should be halved")

	bludgeoningDamage := char.ApplyDamageResistance(damage.TypeBludgeoning, 10)
	assert.Equal(t, 5, bludgeoningDamage, "With rage, bludgeoning damage should be halved")

	// Test non-physical damage is not resisted
	fireDamage := char.ApplyDamageResistance(damage.TypeFire, 10)
	assert.Equal(t, 10, fireDamage, "With rage, fire damage should not be reduced")

	psychicDamage := char.ApplyDamageResistance(damage.TypePsychic, 10)
	assert.Equal(t, 10, psychicDamage, "With rage, psychic damage should not be reduced")
}

func TestCharacter_MultipleStatusEffects(t *testing.T) {
	// Create a character
	char := &character.Character{
		ID:    "test_char",
		Name:  "Test",
		Level: 1,
	}

	// Add multiple effects
	rageEffect := effects.BuildRageEffect(1)
	blessEffect := effects.BuildBlessEffect()

	err := char.AddStatusEffect(rageEffect)
	require.NoError(t, err)

	err = char.AddStatusEffect(blessEffect)
	require.NoError(t, err)

	// Verify both effects are active
	activeEffects := char.GetActiveStatusEffects()
	assert.Len(t, activeEffects, 2, "Should have 2 active effects")

	// Verify attack modifiers from bless
	attackMods := char.GetAttackModifiers(map[string]string{})
	hasBlessing := false
	for _, mod := range attackMods {
		if mod.Value == "+1d4" {
			hasBlessing = true
			break
		}
	}
	assert.True(t, hasBlessing, "Should have +1d4 attack bonus from Bless")

	// Verify damage modifiers from rage
	damageMods := char.GetDamageModifiers(map[string]string{"attack_type": "melee"})
	hasRage := false
	for _, mod := range damageMods {
		if mod.Value == "+2" {
			hasRage = true
			break
		}
	}
	assert.True(t, hasRage, "Should have +2 damage bonus from Rage")
}
