package effects

import (
	"fmt"
	"time"
)

// Builder helps create status effects
type Builder struct {
	effect *StatusEffect
}

// NewBuilder creates a new effect builder
func NewBuilder(name string) *Builder {
	return &Builder{
		effect: &StatusEffect{
			ID:           fmt.Sprintf("%s_%d", name, time.Now().UnixNano()),
			Name:         name,
			Active:       true,
			StackingRule: StackingReplace,
			Modifiers:    []Modifier{},
			Conditions:   []Condition{},
		},
	}
}

// WithSource sets the effect source
func (b *Builder) WithSource(source EffectSource, sourceID string) *Builder {
	b.effect.Source = source
	b.effect.SourceID = sourceID
	return b
}

// WithDescription adds a description
func (b *Builder) WithDescription(desc string) *Builder {
	b.effect.Description = desc
	return b
}

// WithDuration sets the duration
func (b *Builder) WithDuration(durationType DurationType, rounds int) *Builder {
	b.effect.Duration = Duration{
		Type:   durationType,
		Rounds: rounds,
	}
	return b
}

// WithConcentration marks this as requiring concentration
func (b *Builder) WithConcentration() *Builder {
	b.effect.Duration.Concentration = true
	return b
}

// WithStackingRule sets how this effect stacks
func (b *Builder) WithStackingRule(rule StackingRule) *Builder {
	b.effect.StackingRule = rule
	return b
}

// AddModifier adds a modifier to the effect
func (b *Builder) AddModifier(target ModifierTarget, value string) *Builder {
	b.effect.Modifiers = append(b.effect.Modifiers, Modifier{
		Target: target,
		Value:  value,
	})
	return b
}

// AddModifierWithCondition adds a conditional modifier
func (b *Builder) AddModifierWithCondition(target ModifierTarget, value, condition string) *Builder {
	b.effect.Modifiers = append(b.effect.Modifiers, Modifier{
		Target:    target,
		Value:     value,
		Condition: condition,
	})
	return b
}

// AddModifierWithDetails adds a detailed modifier
func (b *Builder) AddModifierWithDetails(target ModifierTarget, value, condition, subTarget, damageType, description string) *Builder {
	b.effect.Modifiers = append(b.effect.Modifiers, Modifier{
		Target:      target,
		Value:       value,
		Condition:   condition,
		SubTarget:   subTarget,
		DamageType:  damageType,
		Description: description,
	})
	return b
}

// AddCondition adds an application condition
func (b *Builder) AddCondition(condType, value string) *Builder {
	b.effect.Conditions = append(b.effect.Conditions, Condition{
		Type:  condType,
		Value: value,
	})
	return b
}

// Build returns the constructed effect
func (b *Builder) Build() *StatusEffect {
	return b.effect
}

// Common effect builders

// BuildRageEffect creates a barbarian rage effect
func BuildRageEffect(level int) *StatusEffect {
	damageBonus := "+2"
	if level >= 16 {
		damageBonus = "+4"
	} else if level >= 9 {
		damageBonus = "+3"
	}

	return NewBuilder("Rage").
		WithSource(SourceAbility, "barbarian_rage").
		WithDescription("You have advantage on Strength checks and Strength saving throws. Melee weapon attacks using Strength gain a damage bonus. You have resistance to bludgeoning, piercing, and slashing damage.").
		WithDuration(DurationRounds, 10).
		AddModifierWithCondition(TargetDamage, damageBonus, "melee_only").
		AddModifierWithDetails(TargetResistance, "resistance", "", "", "bludgeoning", "Resistance to bludgeoning damage").
		AddModifierWithDetails(TargetResistance, "resistance", "", "", "piercing", "Resistance to piercing damage").
		AddModifierWithDetails(TargetResistance, "resistance", "", "", "slashing", "Resistance to slashing damage").
		AddModifierWithDetails(TargetAbilityScore, "advantage", "", "strength", "", "Advantage on Strength checks").
		AddModifierWithDetails(TargetSavingThrow, "advantage", "", "strength", "", "Advantage on Strength saving throws").
		Build()
}

// BuildFavoredEnemyEffect creates a ranger's favored enemy effect
func BuildFavoredEnemyEffect(enemyType string) *StatusEffect {
	return NewBuilder("Favored Enemy").
		WithSource(SourceFeature, "ranger_favored_enemy").
		WithDescription(fmt.Sprintf("You have advantage on Wisdom (Survival) checks to track %s, as well as on Intelligence checks to recall information about them.", enemyType)).
		WithDuration(DurationPermanent, 0).
		AddModifierWithDetails(TargetSkillCheck, "advantage", fmt.Sprintf("vs_enemy_type:%s", enemyType), "survival", "", "Advantage on Survival checks").
		AddModifierWithDetails(TargetAbilityScore, "advantage", fmt.Sprintf("vs_enemy_type:%s", enemyType), "intelligence", "", "Advantage on Intelligence checks").
		AddCondition("enemy_type", enemyType).
		Build()
}

// BuildBlessEffect creates a bless spell effect
func BuildBlessEffect() *StatusEffect {
	return NewBuilder("Bless").
		WithSource(SourceSpell, "bless").
		WithDescription("You bless up to three creatures of your choice within range. Whenever a target makes an attack roll or a saving throw before the spell ends, the target can roll a d4 and add the number rolled to the attack roll or saving throw.").
		WithDuration(DurationRounds, 10).
		WithConcentration().
		AddModifier(TargetAttackRoll, "+1d4").
		AddModifier(TargetSavingThrow, "+1d4").
		Build()
}

// BuildShieldSpellEffect creates a shield spell effect
func BuildShieldSpellEffect() *StatusEffect {
	return NewBuilder("Shield").
		WithSource(SourceSpell, "shield").
		WithDescription("An invisible barrier of magical force appears and protects you. Until the start of your next turn, you have a +5 bonus to AC.").
		WithDuration(DurationRounds, 1).
		AddModifier(TargetAC, "+5").
		Build()
}

// BuildMagicWeaponEffect creates a +X magic weapon effect
func BuildMagicWeaponEffect(weaponName string, bonus int) *StatusEffect {
	bonusStr := fmt.Sprintf("+%d", bonus)
	return NewBuilder(fmt.Sprintf("%s %s", weaponName, bonusStr)).
		WithSource(SourceItem, weaponName).
		WithDescription(fmt.Sprintf("This magic weapon grants a %s bonus to attack and damage rolls.", bonusStr)).
		WithDuration(DurationWhileEquipped, 0).
		AddModifierWithCondition(TargetAttackRoll, bonusStr, fmt.Sprintf("with_weapon:%s", weaponName)).
		AddModifierWithCondition(TargetDamage, bonusStr, fmt.Sprintf("with_weapon:%s", weaponName)).
		Build()
}

// BuildPoisonedCondition creates a poisoned condition effect
func BuildPoisonedCondition() *StatusEffect {
	return NewBuilder("Poisoned").
		WithSource(SourceCondition, "poisoned").
		WithDescription("A poisoned creature has disadvantage on attack rolls and ability checks.").
		WithDuration(DurationUntilRest, 0).
		AddModifier(TargetAttackRoll, "disadvantage").
		AddModifier(TargetAbilityScore, "disadvantage").
		Build()
}
