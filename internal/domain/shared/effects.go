package shared

// DurationType represents how effect duration is measured
type DurationType string

const (
	DurationTypeRounds    DurationType = "rounds"
	DurationTypeMinutes   DurationType = "minutes"
	DurationTypeHours     DurationType = "hours"
	DurationTypeUntilRest DurationType = "until_rest"
	DurationTypePermanent DurationType = "permanent"
)

// ModifierType represents what an effect modifies
type ModifierType string

const (
	ModifierTypeDamageBonus      ModifierType = "damage_bonus"
	ModifierTypeDamageResistance ModifierType = "damage_resistance"
	ModifierTypeDamageImmunity   ModifierType = "damage_immunity"
	ModifierTypeACBonus          ModifierType = "ac_bonus"
	ModifierTypeAttackBonus      ModifierType = "attack_bonus"
	ModifierTypeSavingThrowBonus ModifierType = "saving_throw_bonus"
	ModifierTypeSkillBonus       ModifierType = "skill_bonus"
	ModifierTypeAdvantage        ModifierType = "advantage"
	ModifierTypeDisadvantage     ModifierType = "disadvantage"
	ModifierTypeSpeed            ModifierType = "speed"
	ModifierTypeCondition        ModifierType = "condition"
)

// Modifier represents a specific effect modification
type Modifier struct {
	Type        ModifierType `json:"type"`
	Value       int          `json:"value"`        // Amount of bonus/penalty
	DamageTypes []string     `json:"damage_types"` // For resistances/vulnerabilities
	SkillTypes  []string     `json:"skill_types"`  // For advantage/disadvantage on skills
	Condition   string       `json:"condition"`    // For condition effects (poisoned, etc)
}

// ActiveEffect represents a temporary effect on a character
type ActiveEffect struct {
	ID                    string       `json:"id"`
	Name                  string       `json:"name"`
	Description           string       `json:"description"`
	Source                string       `json:"source"`    // Spell/Feature that created it
	SourceID              string       `json:"source_id"` // ID of caster/user
	Duration              int          `json:"duration"`  // Amount remaining
	DurationType          DurationType `json:"duration_type"`
	Modifiers             []Modifier   `json:"modifiers"`
	RequiresConcentration bool         `json:"requires_concentration"`
}

// TickDuration decrements the duration and returns true if expired
func (e *ActiveEffect) TickDuration() bool {
	if e.DurationType != DurationTypeRounds || e.Duration <= 0 {
		return false
	}

	e.Duration--
	return e.Duration <= 0
}

// IsExpired checks if the effect should be removed
func (e *ActiveEffect) IsExpired() bool {
	if e.DurationType == DurationTypePermanent {
		return false
	}
	if e.DurationType == DurationTypeRounds && e.Duration <= 0 {
		return true
	}
	return false
}

// GetDamageBonus calculates damage bonus for a damage type
func (e *ActiveEffect) GetDamageBonus(damageType string) int {
	for _, mod := range e.Modifiers {
		if mod.Type != ModifierTypeDamageBonus {
			continue
		}
		// If no specific damage types, applies to all
		if len(mod.DamageTypes) == 0 {
			return mod.Value
		}
		// Check if this damage type is affected
		for _, dt := range mod.DamageTypes {
			if dt == damageType || dt == "all" {
				return mod.Value
			}
		}
	}
	return 0
}

// HasResistance checks if effect provides resistance to damage type
func (e *ActiveEffect) HasResistance(damageType string) bool {
	for _, mod := range e.Modifiers {
		if mod.Type != ModifierTypeDamageResistance {
			continue
		}
		for _, dt := range mod.DamageTypes {
			if dt == damageType || dt == "all" {
				return true
			}
		}
	}
	return false
}

// GetACBonus returns any AC bonus from this effect
func (e *ActiveEffect) GetACBonus() int {
	for _, mod := range e.Modifiers {
		if mod.Type == ModifierTypeACBonus {
			return mod.Value
		}
	}
	return 0
}
