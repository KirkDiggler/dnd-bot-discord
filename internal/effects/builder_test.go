package effects

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuilder(t *testing.T) {
	t.Run("creates basic effect", func(t *testing.T) {
		effect := NewBuilder("Test Effect").
			WithSource(SourceAbility, "test_ability").
			WithDescription("A test effect").
			WithDuration(DurationRounds, 5).
			AddModifier(TargetAC, "+2").
			Build()

		assert.Equal(t, "Test Effect", effect.Name)
		assert.Equal(t, SourceAbility, effect.Source)
		assert.Equal(t, "test_ability", effect.SourceID)
		assert.Equal(t, "A test effect", effect.Description)
		assert.Equal(t, DurationRounds, effect.Duration.Type)
		assert.Equal(t, 5, effect.Duration.Rounds)
		assert.Len(t, effect.Modifiers, 1)
		assert.Equal(t, TargetAC, effect.Modifiers[0].Target)
		assert.Equal(t, "+2", effect.Modifiers[0].Value)
	})

	t.Run("creates effect with conditions", func(t *testing.T) {
		effect := NewBuilder("Conditional Effect").
			WithSource(SourceFeature, "test_feature").
			AddCondition("enemy_type", "undead").
			AddModifierWithCondition(TargetDamage, "+1d6", "vs_enemy_type:undead").
			Build()

		assert.Len(t, effect.Conditions, 1)
		assert.Equal(t, "enemy_type", effect.Conditions[0].Type)
		assert.Equal(t, "undead", effect.Conditions[0].Value)
		assert.Equal(t, "vs_enemy_type:undead", effect.Modifiers[0].Condition)
	})
}

func TestBuildRageEffect(t *testing.T) {
	tests := []struct {
		level         int
		expectedBonus string
	}{
		{1, "+2"},
		{8, "+2"},
		{9, "+3"},
		{15, "+3"},
		{16, "+4"},
		{20, "+4"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("level %d", tt.level), func(t *testing.T) {
			effect := BuildRageEffect(tt.level)

			assert.Equal(t, "Rage", effect.Name)
			assert.Equal(t, SourceAbility, effect.Source)
			assert.Equal(t, "barbarian_rage", effect.SourceID)
			assert.Equal(t, DurationRounds, effect.Duration.Type)
			assert.Equal(t, 10, effect.Duration.Rounds)

			// Find damage modifier
			var damageModifier *Modifier
			for i := range effect.Modifiers {
				if effect.Modifiers[i].Target == TargetDamage {
					damageModifier = &effect.Modifiers[i]
					break
				}
			}

			assert.NotNil(t, damageModifier)
			assert.Equal(t, tt.expectedBonus, damageModifier.Value)
			assert.Equal(t, "melee_only", damageModifier.Condition)

			// Check resistances
			resistanceCount := 0
			for _, mod := range effect.Modifiers {
				if mod.Target == TargetResistance {
					resistanceCount++
					assert.Equal(t, "resistance", mod.Value)
				}
			}
			assert.Equal(t, 3, resistanceCount) // bludgeoning, piercing, slashing
		})
	}
}

func TestBuildFavoredEnemyEffect(t *testing.T) {
	effect := BuildFavoredEnemyEffect("orc")

	assert.Equal(t, "Favored Enemy", effect.Name)
	assert.Equal(t, SourceFeature, effect.Source)
	assert.Equal(t, DurationPermanent, effect.Duration.Type)

	// Check conditions
	assert.Len(t, effect.Conditions, 1)
	assert.Equal(t, "enemy_type", effect.Conditions[0].Type)
	assert.Equal(t, "orc", effect.Conditions[0].Value)

	// Check modifiers
	hasSkillAdvantage := false
	hasIntAdvantage := false
	for _, mod := range effect.Modifiers {
		if mod.Target == TargetSkillCheck && mod.SubTarget == "survival" {
			hasSkillAdvantage = true
			assert.Equal(t, "advantage", mod.Value)
			assert.Equal(t, "vs_enemy_type:orc", mod.Condition)
		}
		if mod.Target == TargetAbilityScore && mod.SubTarget == "intelligence" {
			hasIntAdvantage = true
			assert.Equal(t, "advantage", mod.Value)
		}
	}
	assert.True(t, hasSkillAdvantage)
	assert.True(t, hasIntAdvantage)
}

func TestBuildBlessEffect(t *testing.T) {
	effect := BuildBlessEffect()

	assert.Equal(t, "Bless", effect.Name)
	assert.Equal(t, SourceSpell, effect.Source)
	assert.Equal(t, "bless", effect.SourceID)
	assert.Equal(t, DurationRounds, effect.Duration.Type)
	assert.Equal(t, 10, effect.Duration.Rounds)
	assert.True(t, effect.Duration.Concentration)

	// Check modifiers
	hasAttackBonus := false
	hasSaveBonus := false
	for _, mod := range effect.Modifiers {
		if mod.Target == TargetAttackRoll {
			hasAttackBonus = true
			assert.Equal(t, "+1d4", mod.Value)
		}
		if mod.Target == TargetSavingThrow {
			hasSaveBonus = true
			assert.Equal(t, "+1d4", mod.Value)
		}
	}
	assert.True(t, hasAttackBonus)
	assert.True(t, hasSaveBonus)
}

func TestBuildMagicWeaponEffect(t *testing.T) {
	effect := BuildMagicWeaponEffect("longsword", 2)

	assert.Equal(t, "longsword +2", effect.Name)
	assert.Equal(t, SourceItem, effect.Source)
	assert.Equal(t, "longsword", effect.SourceID)
	assert.Equal(t, DurationWhileEquipped, effect.Duration.Type)

	// Check modifiers
	for _, mod := range effect.Modifiers {
		assert.Equal(t, "+2", mod.Value)
		assert.Equal(t, "with_weapon:longsword", mod.Condition)
		assert.True(t, mod.Target == TargetAttackRoll || mod.Target == TargetDamage)
	}
}
