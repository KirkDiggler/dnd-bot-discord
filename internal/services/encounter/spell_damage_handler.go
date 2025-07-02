package encounter

import (
	"context"
	"log"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/conditions"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/events"
)

// SpellDamageHandler handles spell damage events
type SpellDamageHandler struct {
	service Service
}

// NewSpellDamageHandler creates a new spell damage handler
func NewSpellDamageHandler(service Service) *SpellDamageHandler {
	return &SpellDamageHandler{
		service: service,
	}
}

// HandleEvent processes spell damage events
//
//nolint:errcheck // False positive - we properly check the AddCondition error
func (h *SpellDamageHandler) HandleEvent(event *events.GameEvent) error {
	// Only handle spell damage events
	if event.Type != events.OnSpellDamage {
		return nil
	}

	// Get damage amount
	damage, exists := event.GetIntContext(events.ContextDamage)
	if !exists || damage <= 0 {
		return nil
	}

	// Get target ID
	targetID, exists := event.GetStringContext(events.ContextTargetID)
	if !exists && event.Target != nil {
		targetID = event.Target.ID
	}

	if targetID == "" {
		log.Printf("SpellDamageHandler: No target ID for spell damage")
		return nil
	}

	// Get encounter ID (would need to be added to event context)
	encounterID, exists := event.GetStringContext(events.ContextEncounterID)
	if !exists {
		log.Printf("SpellDamageHandler: No encounter ID in spell damage event")
		return nil
	}

	// Apply the damage
	// TODO: Need to get userID from event context - for now use system
	userID := "system"
	if uid, exists := event.GetStringContext(events.ContextUserID); exists {
		userID = uid
	}

	err := h.service.ApplyDamage(context.Background(), encounterID, targetID, userID, damage)

	if err != nil {
		log.Printf("SpellDamageHandler: Failed to apply spell damage: %v", err)
		return err
	}

	// Get spell name for logging
	spellName, _ := event.GetStringContext(events.ContextSpellName)
	damageType, _ := event.GetStringContext(events.ContextDamageType)

	log.Printf("SpellDamageHandler: Applied %d %s damage from %s to target %s",
		damage, damageType, spellName, targetID)

	// Handle spell-specific effects
	if spellName == "Vicious Mockery" {
		// Apply disadvantage on next attack using the condition service
		// The targetID can be either a character ID or a monster combatant ID
		if h.service.(*service).conditionService != nil {
			_, err := h.service.(*service).conditionService.AddCondition(
				targetID,
				conditions.DisadvantageOnNextAttack,
				spellName,
				conditions.DurationEndOfNextTurn,
				1,
			)
			if err != nil {
				log.Printf("Failed to apply disadvantage condition from %s: %v", spellName, err)
			} else {
				log.Printf("Applied disadvantage on next attack to %s from %s", targetID, spellName)
			}
		}
	}

	return nil
}

// Priority returns the handler priority
func (h *SpellDamageHandler) Priority() int {
	return 100
}
