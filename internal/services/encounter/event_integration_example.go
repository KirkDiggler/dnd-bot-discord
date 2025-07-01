package encounter

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/events"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/combat/attack"
)

// EventBasedDamageCalculator shows how damage calculation would work with events
// This is a proof of concept for migrating to event-driven architecture
type EventBasedDamageCalculator struct {
	eventBus *events.EventBus
}

// NewEventBasedDamageCalculator creates a new damage calculator
func NewEventBasedDamageCalculator(eventBus *events.EventBus) *EventBasedDamageCalculator {
	return &EventBasedDamageCalculator{
		eventBus: eventBus,
	}
}

// CalculateAttackDamage calculates damage using the event system
func (calc *EventBasedDamageCalculator) CalculateAttackDamage(
	attacker *character.Character,
	target *character.Character,
	attackResult *attack.Result,
	weaponType string,
) (int, error) {
	// Start with base damage from attack result
	baseDamage := attackResult.DamageRoll

	// Create damage roll event
	damageEvent := events.NewGameEvent(events.OnDamageRoll).
		WithActor(attacker).
		WithTarget(target).
		WithContext("damage", baseDamage).
		WithContext("attack_type", weaponType).
		WithContext("weapon", attackResult.WeaponKey).
		WithContext("is_critical", attackResult.AttackRoll == 20)

	// Emit event - all damage modifiers will be applied
	if err := calc.eventBus.Emit(damageEvent); err != nil {
		return baseDamage, err
	}

	// Get modified damage
	modifiedDamage, _ := damageEvent.GetIntContext("damage")

	// Now create a "before take damage" event for the target
	takeDamageEvent := events.NewGameEvent(events.BeforeTakeDamage).
		WithActor(attacker).
		WithTarget(target).
		WithContext("damage", modifiedDamage).
		WithContext("damage_type", "slashing") // Would be determined by weapon

	// Emit event - resistances/vulnerabilities will be applied
	if err := calc.eventBus.Emit(takeDamageEvent); err != nil {
		return modifiedDamage, err
	}

	// Get final damage after resistances
	finalDamage, _ := takeDamageEvent.GetIntContext("damage")

	return finalDamage, nil
}

// This example shows how the encounter service would use the event-based damage calculator
// In a real implementation, the eventBus would be injected into the service
