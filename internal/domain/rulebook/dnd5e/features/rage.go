package features

import (
	"fmt"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/events"
)

// RageModifier implements the barbarian rage feature as an event modifier
type RageModifier struct {
	id          string
	characterID string
	level       int
	damageBonus int
}

// NewRageModifier creates a new rage modifier for a specific character
func NewRageModifier(characterID string, level int) *RageModifier {
	// Determine damage bonus based on level
	damageBonus := 2
	if level >= 16 {
		damageBonus = 4
	} else if level >= 9 {
		damageBonus = 3
	}

	return &RageModifier{
		id:          fmt.Sprintf("rage_%s", characterID),
		characterID: characterID,
		level:       level,
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
		Type:        "feature",
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
	// Only apply to the raging character's events
	if event.Type == events.OnDamageRoll {
		// Must be this character's attack
		if event.Actor == nil || event.Actor.ID != r.characterID {
			return false
		}

		// Rage damage bonus only applies to melee attacks
		attackType, exists := event.GetStringContext("attack_type")
		if !exists {
			return false
		}
		return strings.EqualFold(attackType, "melee")
	}

	// Check if this character is taking damage (for resistance)
	if event.Type == events.BeforeTakeDamage {
		// Must be damage to this character
		if event.Target == nil || event.Target.ID != r.characterID {
			return false
		}

		// Rage provides resistance to physical damage
		damageType, exists := event.GetStringContext("damage_type")
		if !exists {
			return false
		}
		return damageType == "bludgeoning" || damageType == "piercing" || damageType == "slashing"
	}

	// TODO: Add advantage on Strength checks and saves
	// if event.Type == events.BeforeAbilityCheck || event.Type == events.BeforeSavingThrow

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

// RageListener wraps a rage modifier to work with the event bus
type RageListener struct {
	modifier *RageModifier
}

// NewRageListener creates a new rage listener
func NewRageListener(characterID string, level int) *RageListener {
	return &RageListener{
		modifier: NewRageModifier(characterID, level),
	}
}

// HandleEvent processes events for rage
func (rl *RageListener) HandleEvent(event *events.GameEvent) error {
	// Check if modifier condition is met
	if !rl.modifier.Condition(event) {
		return nil
	}

	// Apply the modifier
	return rl.modifier.Apply(event)
}

// Priority returns the listener priority
func (rl *RageListener) Priority() int {
	return rl.modifier.Priority()
}

// ID returns the modifier ID for tracking
func (rl *RageListener) ID() string {
	return rl.modifier.ID()
}
