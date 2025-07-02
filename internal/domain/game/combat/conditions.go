package combat

import (
	"time"
)

// CombatantCondition represents a condition affecting a combatant (player or monster)
type CombatantCondition struct {
	Type        string    `json:"type"`        // e.g., "disadvantage_next_attack", "poisoned"
	Source      string    `json:"source"`      // What caused this condition
	SourceID    string    `json:"source_id"`   // ID of the caster/attacker
	Duration    string    `json:"duration"`    // "rounds", "turns", "until_damaged", etc.
	Remaining   int       `json:"remaining"`   // Remaining duration
	AppliedAt   time.Time `json:"applied_at"`  // When the condition was applied
	Description string    `json:"description"` // Human-readable description
}

// HasCondition checks if a combatant has a specific condition
func (c *Combatant) HasCondition(conditionType string) bool {
	for _, cond := range c.Conditions {
		if cond == conditionType {
			return true
		}
	}
	return false
}

// AddCondition adds a condition to the combatant if not already present
func (c *Combatant) AddCondition(condition string) {
	if !c.HasCondition(condition) {
		c.Conditions = append(c.Conditions, condition)
	}
}

// RemoveCondition removes a specific condition
func (c *Combatant) RemoveCondition(condition string) {
	newConditions := []string{}
	for _, cond := range c.Conditions {
		if cond != condition {
			newConditions = append(newConditions, cond)
		}
	}
	c.Conditions = newConditions
}

// ClearCondition removes all instances of a condition type
func (c *Combatant) ClearCondition(conditionType string) {
	c.RemoveCondition(conditionType)
}
