package conditions

import "time"

// ConditionType represents a type of condition
type ConditionType string

// Standard D&D 5e conditions
const (
	Blinded       ConditionType = "blinded"
	Charmed       ConditionType = "charmed"
	Deafened      ConditionType = "deafened"
	Frightened    ConditionType = "frightened"
	Grappled      ConditionType = "grappled"
	Incapacitated ConditionType = "incapacitated"
	Invisible     ConditionType = "invisible"
	Paralyzed     ConditionType = "paralyzed"
	Petrified     ConditionType = "petrified"
	Poisoned      ConditionType = "poisoned"
	Prone         ConditionType = "prone"
	Restrained    ConditionType = "restrained"
	Stunned       ConditionType = "stunned"
	Unconscious   ConditionType = "unconscious"
	Exhaustion    ConditionType = "exhaustion" // Has levels 1-6

	// Custom conditions
	DisadvantageOnNextAttack ConditionType = "disadvantage_next_attack"
	Concentration            ConditionType = "concentration"
	Rage                     ConditionType = "rage" // Already tracked elsewhere, but could unify
)

// DurationType defines how long a condition lasts
type DurationType string

const (
	DurationRounds        DurationType = "rounds"          // Lasts X rounds
	DurationTurns         DurationType = "turns"           // Lasts X turns
	DurationUntilRest     DurationType = "until_rest"      // Until short or long rest
	DurationUntilLongRest DurationType = "until_long_rest" // Until long rest only
	DurationConcentration DurationType = "concentration"   // Until concentration breaks
	DurationPermanent     DurationType = "permanent"       // Until removed by effect
	DurationInstant       DurationType = "instant"         // Applied and immediately removed
	DurationUntilDamaged  DurationType = "until_damaged"   // Until target takes damage
	DurationEndOfNextTurn DurationType = "end_next_turn"   // Until end of target's next turn
)

// Condition represents an active condition on a character
type Condition struct {
	ID          string        `json:"id"`
	Type        ConditionType `json:"type"`
	Name        string        `json:"name"`        // Display name
	Description string        `json:"description"` // What it does
	Source      string        `json:"source"`      // What caused it (spell, ability, etc)
	SourceID    string        `json:"source_id"`   // ID of caster/attacker

	// Duration tracking
	DurationType DurationType `json:"duration_type"`
	Duration     int          `json:"duration"`  // Number of rounds/turns if applicable
	Remaining    int          `json:"remaining"` // Rounds/turns remaining
	Level        int          `json:"level"`     // For conditions with levels (exhaustion)
	SaveDC       int          `json:"save_dc"`   // DC to end condition (if applicable)
	SaveType     string       `json:"save_type"` // Ability for save (STR, DEX, etc)
	SaveEnd      bool         `json:"save_end"`  // Can save at end of turn to end?

	// Metadata
	AppliedAt time.Time              `json:"applied_at"`
	AppliedBy string                 `json:"applied_by"` // User ID who applied it
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Effect describes what a condition does
type Effect struct {
	// Combat effects
	AttackAdvantage     bool `json:"attack_advantage"`     // Advantage on attacks
	AttackDisadvantage  bool `json:"attack_disadvantage"`  // Disadvantage on attacks
	DefenseAdvantage    bool `json:"defense_advantage"`    // Attackers have advantage
	DefenseDisadvantage bool `json:"defense_disadvantage"` // Attackers have disadvantage

	// Movement effects
	SpeedMultiplier float64 `json:"speed_multiplier"` // 0 = can't move, 0.5 = half speed
	CantMove        bool    `json:"cant_move"`        // Completely prevents movement

	// Action effects
	CantAct   bool `json:"cant_act"`   // No actions
	CantReact bool `json:"cant_react"` // No reactions
	CantSpeak bool `json:"cant_speak"` // No verbal components

	// Save effects
	SaveAdvantage    map[string]bool `json:"save_advantage"`    // Advantage on saves by type
	SaveDisadvantage map[string]bool `json:"save_disadvantage"` // Disadvantage on saves
	SaveAutoFail     map[string]bool `json:"save_auto_fail"`    // Auto-fail saves

	// Damage effects
	Vulnerability map[string]bool `json:"vulnerability"` // Damage vulnerabilities
	Resistance    map[string]bool `json:"resistance"`    // Damage resistances
	Immunity      map[string]bool `json:"immunity"`      // Damage immunities

	// Other effects
	Incapacitated   bool `json:"incapacitated"`    // Can't take actions or reactions
	FallProne       bool `json:"fall_prone"`       // Falls prone when applied
	DropItems       bool `json:"drop_items"`       // Drops held items
	CantConcentrate bool `json:"cant_concentrate"` // Breaks & prevents concentration
}

// GetStandardEffects returns the standard effects for each condition type
func GetStandardEffects(conditionType ConditionType) *Effect {
	effects := map[ConditionType]*Effect{
		Blinded: {
			AttackDisadvantage: true,
			DefenseAdvantage:   true,
			CantMove:           false, // Can move but can't see
		},
		Charmed: {
			CantMove: false, // Specific to source
			// Note: Can't attack charmer, handled separately
		},
		Deafened: {
			// Fails perception checks based on hearing
		},
		Frightened: {
			AttackDisadvantage: true, // When source is in sight
			// Can't willingly move closer to source
		},
		Grappled: {
			SpeedMultiplier: 0, // Speed becomes 0
		},
		Incapacitated: {
			Incapacitated: true,
			CantAct:       true,
			CantReact:     true,
		},
		Invisible: {
			AttackAdvantage:     true,
			DefenseDisadvantage: true,
		},
		Paralyzed: {
			Incapacitated:    true,
			CantAct:          true,
			CantReact:        true,
			CantMove:         true,
			CantSpeak:        true,
			DefenseAdvantage: true, // Attacks have advantage
			SaveAutoFail:     map[string]bool{"STR": true, "DEX": true},
			// Note: Hits within 5 feet are crits
		},
		Petrified: {
			// Transformed to stone
			Resistance:       map[string]bool{"all": true}, // Resistance to all damage
			Immunity:         map[string]bool{"poison": true},
			CantAct:          true,
			CantReact:        true,
			CantMove:         true,
			CantSpeak:        true,
			DefenseAdvantage: true,
			SaveAutoFail:     map[string]bool{"STR": true, "DEX": true},
		},
		Poisoned: {
			AttackDisadvantage: true,
			SaveDisadvantage:   map[string]bool{"all": true}, // Disadvantage on all saves
		},
		Prone: {
			AttackDisadvantage: true, // Melee within 5 ft has advantage, ranged has disadvantage
			SpeedMultiplier:    0.5,  // Crawling is half speed
			FallProne:          true,
		},
		Restrained: {
			SpeedMultiplier:    0,
			AttackDisadvantage: true,
			DefenseAdvantage:   true,
			SaveDisadvantage:   map[string]bool{"DEX": true},
		},
		Stunned: {
			Incapacitated:    true,
			CantAct:          true,
			CantReact:        true,
			CantMove:         true,
			CantSpeak:        true,
			DefenseAdvantage: true,
			SaveAutoFail:     map[string]bool{"STR": true, "DEX": true},
		},
		Unconscious: {
			Incapacitated:    true,
			CantAct:          true,
			CantReact:        true,
			CantMove:         true,
			CantSpeak:        true,
			DropItems:        true,
			FallProne:        true,
			DefenseAdvantage: true,
			SaveAutoFail:     map[string]bool{"STR": true, "DEX": true},
			// Note: Hits within 5 feet are crits
		},
	}

	if effect, exists := effects[conditionType]; exists {
		return effect
	}
	return &Effect{} // Empty effect for unknown conditions
}
