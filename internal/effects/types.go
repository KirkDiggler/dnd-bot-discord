package effects

import (
	"time"
)

// EffectSource represents where an effect comes from
type EffectSource string

const (
	SourceAbility   EffectSource = "ability"
	SourceSpell     EffectSource = "spell"
	SourceItem      EffectSource = "item"
	SourceCondition EffectSource = "condition"
	SourceFeature   EffectSource = "feature"
	SourceOther     EffectSource = "other"
)

// DurationType represents different duration types
type DurationType string

const (
	DurationPermanent     DurationType = "permanent"
	DurationRounds        DurationType = "rounds"
	DurationUntilRest     DurationType = "until_rest"
	DurationConcentration DurationType = "concentration"
	DurationWhileEquipped DurationType = "while_equipped"
	DurationInstant       DurationType = "instant"
)

// StackingRule defines how effects stack with each other
type StackingRule string

const (
	StackingReplace     StackingRule = "replace"      // New effect replaces old
	StackingStack       StackingRule = "stack"        // Effects add together
	StackingTakeHighest StackingRule = "take_highest" // Only highest applies
	StackingTakeLowest  StackingRule = "take_lowest"  // Only lowest applies
)

// ModifierTarget represents what an effect modifies
type ModifierTarget string

const (
	TargetAttackRoll    ModifierTarget = "attack_roll"
	TargetDamage        ModifierTarget = "damage"
	TargetAC            ModifierTarget = "ac"
	TargetAbilityScore  ModifierTarget = "ability_score"
	TargetSkillCheck    ModifierTarget = "skill_check"
	TargetSavingThrow   ModifierTarget = "saving_throw"
	TargetSpeed         ModifierTarget = "speed"
	TargetInitiative    ModifierTarget = "initiative"
	TargetHP            ModifierTarget = "hp"
	TargetMaxHP         ModifierTarget = "max_hp"
	TargetResistance    ModifierTarget = "resistance"
	TargetImmunity      ModifierTarget = "immunity"
	TargetVulnerability ModifierTarget = "vulnerability"
)

// Duration represents how long an effect lasts
type Duration struct {
	Type          DurationType
	Rounds        int       // For round-based durations
	EndTime       time.Time // For time-based durations
	Concentration bool      // Whether this requires concentration
}

// Condition represents when an effect applies
type Condition struct {
	Type       string            // "enemy_type", "weapon_type", "ability_check", etc.
	Value      string            // "orc", "melee", "strength", etc.
	Parameters map[string]string // Additional parameters
}

// Modifier represents a single modification an effect makes
type Modifier struct {
	Target      ModifierTarget
	Value       string // Can be "+2", "1d4", "advantage", "resistance", etc.
	Condition   string // "melee_only", "vs_enemy_type:orc", etc.
	SubTarget   string // For ability scores/skills: "strength", "athletics", etc.
	DamageType  string // For resistance/immunity/vulnerability
	Description string // Human-readable description
}

// StatusEffect represents any effect that modifies a character
type StatusEffect struct {
	ID           string
	Source       EffectSource
	SourceID     string // ID of the spell/item/ability that created this
	Name         string
	Description  string
	Duration     Duration
	Modifiers    []Modifier
	Conditions   []Condition // When the effect applies
	StackingRule StackingRule
	Active       bool // Whether the effect is currently active
	CreatedAt    time.Time
	ExpiresAt    *time.Time // Nil for permanent effects
}

// IsExpired checks if the effect has expired
func (e *StatusEffect) IsExpired() bool {
	if e.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*e.ExpiresAt)
}

// AppliesToCondition checks if the effect applies given a set of conditions
func (e *StatusEffect) AppliesToCondition(conditionType, conditionValue string) bool {
	// If no conditions, always applies
	if len(e.Conditions) == 0 {
		return true
	}

	// Check if any condition matches
	for _, cond := range e.Conditions {
		if cond.Type == conditionType && cond.Value == conditionValue {
			return true
		}
	}
	return false
}
