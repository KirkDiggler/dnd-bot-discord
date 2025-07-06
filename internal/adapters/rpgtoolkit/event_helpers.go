package rpgtoolkit

import (
	"context"
	"log"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/rpg-toolkit/core"
	rpgevents "github.com/KirkDiggler/rpg-toolkit/events"
)

// EmitEvent is a helper to emit events using rpg-toolkit style
func EmitEvent(bus *rpgevents.Bus, eventType string, actor, target *character.Character, contextData map[string]interface{}) error {
	if bus == nil {
		return nil
	}

	// Convert characters to entities
	var source, tgt core.Entity
	if actor != nil {
		source = WrapCharacter(actor)
	}
	if target != nil {
		tgt = WrapCharacter(target)
	}

	// Create the event
	event := rpgevents.NewGameEvent(eventType, source, tgt)

	// Add context data
	for k, v := range contextData {
		event.Context().Set(k, v)
	}

	// Publish the event
	return bus.Publish(context.Background(), event)
}

// CreateAndEmitEvent creates an event, emits it, and returns the event for checking cancellation
func CreateAndEmitEvent(bus *rpgevents.Bus, eventType string, actor, target *character.Character, contextData map[string]interface{}) (rpgevents.Event, error) {
	if bus == nil {
		return nil, nil
	}

	// Convert characters to entities
	var source, tgt core.Entity
	if actor != nil {
		source = WrapCharacter(actor)
	}
	if target != nil {
		tgt = WrapCharacter(target)
	}

	// Create the event
	event := rpgevents.NewGameEvent(eventType, source, tgt)

	// Add context data
	for k, v := range contextData {
		event.Context().Set(k, v)
	}

	// Publish the event
	err := bus.Publish(context.Background(), event)
	return event, err
}

// Event type mappings from old system to rpg-toolkit
var EventTypeMappings = map[string]string{
	"OnTurnStart":      rpgevents.EventOnTurnStart,
	"BeforeAttackRoll": rpgevents.EventBeforeAttackRoll,
	"OnAttackRoll":     rpgevents.EventOnAttackRoll,
	"AfterAttackRoll":  rpgevents.EventAfterAttackRoll,
	"OnDamageRoll":     rpgevents.EventOnDamageRoll,
	"BeforeTakeDamage": rpgevents.EventBeforeTakeDamage,
	"OnTakeDamage":     rpgevents.EventOnTakeDamage,
	"OnSpellDamage":    rpgevents.EventOnSpellDamage,
	"OnStatusApplied":  rpgevents.EventOnConditionApplied,
	"OnStatusRemoved":  rpgevents.EventOnConditionRemoved,
	"OnSavingThrow":    rpgevents.EventOnSavingThrow,
	"OnSpellCast":      rpgevents.EventOnSpellCast,
}

// GetEventType converts old event type to rpg-toolkit event type
func GetEventType(oldType string) string {
	if mapped, ok := EventTypeMappings[oldType]; ok {
		return mapped
	}
	// Default: use the old type with dndbot prefix
	return "dndbot." + oldType
}

// Additional context keys not in event_adapter.go
const (
	// Combat context keys
	ContextAttackType       = "attack_type"
	ContextTargetID         = "target_id"
	ContextWeaponKey        = "weapon_key"
	ContextWeaponType       = "weapon_type"
	ContextWeaponHasFinesse = "weapon_has_finesse"

	// Combat conditions
	ContextAllyAdjacent = "ally_adjacent"

	// Turn tracking
	ContextTurnCount     = "turn_count"
	ContextRound         = "round"
	ContextNumCombatants = "num_combatants"

	// Spell context keys
	ContextSpellName   = "spell_name"
	ContextSpellSaveDC = "spell_save_dc"

	// Status effect context
	ContextStatusType     = "status"
	ContextEffectDuration = "effect_duration"
	ContextEffectSource   = "effect_source"

	// Session and encounter context
	ContextEncounterID = "encounter_id"
	ContextUserID      = "user_id"

	// Sneak attack context
	ContextSneakAttackDamage = "sneak_attack_damage"
	ContextSneakAttackDice   = "sneak_attack_dice"
)

// CreateAndEmitEventWithEntities creates an event with Entity types, emits it, and returns the event
func CreateAndEmitEventWithEntities(bus *rpgevents.Bus, eventType string, actor, target core.Entity, contextData map[string]interface{}) (rpgevents.Event, error) {
	if bus == nil {
		return nil, nil
	}

	// Convert to core.Entity if needed
	var source, tgt core.Entity

	// Handle actor
	if actor != nil {
		// Already a core.Entity, just use it
		source = actor
	}

	// Handle target
	if target != nil {
		// Already a core.Entity, just use it
		tgt = target
	}

	// Create the event
	event := rpgevents.NewGameEvent(eventType, source, tgt)

	// Add context data
	for k, v := range contextData {
		event.Context().Set(k, v)
	}

	// Publish the event
	err := bus.Publish(context.Background(), event)
	return event, err
}

// LogEventError logs event emission errors
func LogEventError(eventType string, err error) {
	if err != nil {
		log.Printf("Failed to emit %s event: %v", eventType, err)
	}
}
