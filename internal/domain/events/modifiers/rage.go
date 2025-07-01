package modifiers

import (
	"fmt"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/events"
	"strings"
)

// RageModifier implements the barbarian rage effect using the event system
type RageModifier struct {
	id          string
	level       int
	roundsLeft  int
	damageBonus int
}

// NewRageModifier creates a new rage modifier
func NewRageModifier(level int) *RageModifier {
	// Determine damage bonus based on level
	damageBonus := 2
	if level >= 16 {
		damageBonus = 4
	} else if level >= 9 {
		damageBonus = 3
	}

	return &RageModifier{
		id:          fmt.Sprintf("rage_%d", level),
		level:       level,
		roundsLeft:  10,
		damageBonus: damageBonus,
	}
}

// ID returns the unique identifier for this modifier
func (r *RageModifier) ID() string {
	return r.id
}

// Source returns information about where this modifier comes from
func (r *RageModifier) Source() events.ModifierSource {
	return events.ModifierSource{
		Type:        "ability",
		Name:        "Rage",
		Description: "Barbarian class feature",
	}
}

// Priority returns the execution priority (100 = feature priority)
func (r *RageModifier) Priority() int {
	return 100
}

// Condition checks if this modifier should apply to the event
func (r *RageModifier) Condition(event *events.GameEvent) bool {
	// Check if this is a damage roll event
	if event.Type == events.OnDamageRoll {
		// Rage damage bonus only applies to melee attacks
		attackType, exists := event.GetStringContext("attack_type")
		if !exists {
			return false
		}
		return strings.EqualFold(attackType, "melee")
	}

	// Check if this is taking damage (for resistance)
	if event.Type == events.BeforeTakeDamage {
		// Rage provides resistance to physical damage
		damageType, exists := event.GetStringContext("damage_type")
		if !exists {
			return false
		}
		return damageType == "bludgeoning" || damageType == "piercing" || damageType == "slashing"
	}

	// TODO: Add advantage on Strength checks and saves

	return false
}

// Apply modifies the event based on rage effects
func (r *RageModifier) Apply(event *events.GameEvent) error {
	switch event.Type {
	case events.OnDamageRoll:
		// Add rage damage bonus
		currentDamage, exists := event.GetIntContext("damage")
		if !exists {
			return fmt.Errorf("no damage value in event context")
		}

		event.WithContext("damage", currentDamage+r.damageBonus)
		event.WithContext("damage_bonus_source", fmt.Sprintf("Rage (+%d)", r.damageBonus))

	case events.BeforeTakeDamage:
		// Apply resistance (half damage)
		currentDamage, exists := event.GetIntContext("damage")
		if !exists {
			return fmt.Errorf("no damage value in event context")
		}

		reducedDamage := currentDamage / 2
		event.WithContext("damage", reducedDamage)
		event.WithContext("resistance_applied", "Rage (physical damage resistance)")
	}

	return nil
}

// Duration returns how long this modifier lasts
func (r *RageModifier) Duration() events.ModifierDuration {
	return &events.RoundsDuration{
		Rounds:    10,
		StartTurn: 0, // TODO: Track actual start turn
	}
}
