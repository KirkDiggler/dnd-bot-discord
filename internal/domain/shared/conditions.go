package shared

// ConditionType represents standard D&D 5e conditions
type ConditionType string

const (
	ConditionBlinded       ConditionType = "blinded"
	ConditionCharmed       ConditionType = "charmed"
	ConditionDeafened      ConditionType = "deafened"
	ConditionFrightened    ConditionType = "frightened"
	ConditionGrappled      ConditionType = "grappled"
	ConditionIncapacitated ConditionType = "incapacitated"
	ConditionInvisible     ConditionType = "invisible"
	ConditionParalyzed     ConditionType = "paralyzed"
	ConditionPetrified     ConditionType = "petrified"
	ConditionPoisoned      ConditionType = "poisoned"
	ConditionProne         ConditionType = "prone"
	ConditionRestrained    ConditionType = "restrained"
	ConditionStunned       ConditionType = "stunned"
	ConditionUnconscious   ConditionType = "unconscious"
	ConditionExhaustion    ConditionType = "exhaustion"
)

// Condition represents a status condition with its effects
type Condition struct {
	Type        ConditionType `json:"type"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Effects     []string      `json:"effects"` // List of mechanical effects
}

// StandardConditions defines all D&D 5e conditions and their effects
var StandardConditions = map[ConditionType]*Condition{
	ConditionBlinded: {
		Type:        ConditionBlinded,
		Name:        "Blinded",
		Description: "A blinded creature can't see and automatically fails any ability check that requires sight.",
		Effects: []string{
			"Attack rolls against the creature have advantage",
			"The creature's attack rolls have disadvantage",
		},
	},
	ConditionPoisoned: {
		Type:        ConditionPoisoned,
		Name:        "Poisoned",
		Description: "A poisoned creature has disadvantage on attack rolls and ability checks.",
		Effects: []string{
			"Disadvantage on attack rolls",
			"Disadvantage on ability checks",
		},
	},
	ConditionStunned: {
		Type:        ConditionStunned,
		Name:        "Stunned",
		Description: "A stunned creature is incapacitated, can't move, and can speak only falteringly.",
		Effects: []string{
			"Incapacitated (can't take actions or reactions)",
			"Can't move",
			"Can speak only falteringly",
			"Automatically fails Strength and Dexterity saving throws",
			"Attack rolls against the creature have advantage",
		},
	},
	ConditionProne: {
		Type:        ConditionProne,
		Name:        "Prone",
		Description: "A prone creature's only movement option is to crawl, unless it stands up.",
		Effects: []string{
			"Disadvantage on attack rolls",
			"Attack rolls against the creature have advantage if attacker is within 5 feet",
			"Attack rolls against the creature have disadvantage if attacker is farther than 5 feet",
			"Must spend half movement to stand up",
		},
	},
}

// CreateConditionEffect creates an ActiveEffect for a condition
func CreateConditionEffect(condition ConditionType, source string, duration int) *ActiveEffect {
	cond, exists := StandardConditions[condition]
	if !exists {
		return nil
	}

	effect := &ActiveEffect{
		Name:         cond.Name,
		Description:  cond.Description,
		Source:       source,
		Duration:     duration,
		DurationType: DurationTypeRounds,
		Modifiers:    []Modifier{},
	}

	// Add condition modifier
	effect.Modifiers = append(effect.Modifiers, Modifier{
		Type:      ModifierTypeCondition,
		Condition: string(condition),
	})

	// Add specific modifiers based on condition
	switch condition {
	case ConditionBlinded:
		// Attacks against have advantage
		effect.Modifiers = append(effect.Modifiers, Modifier{
			Type:  ModifierTypeDisadvantage,
			Value: 1,
		})
	case ConditionPoisoned:
		// Disadvantage on attacks and ability checks
		effect.Modifiers = append(effect.Modifiers, Modifier{
			Type:  ModifierTypeDisadvantage,
			Value: 1,
		})
	case ConditionStunned:
		// Multiple effects including advantage on attacks against
		effect.Modifiers = append(effect.Modifiers, Modifier{
			Type:  ModifierTypeCondition,
			Value: 1, // Incapacitated
		})
	}

	return effect
}
