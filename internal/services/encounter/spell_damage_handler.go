package encounter

import (
	"context"
	"log"

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
		return nil // No target ID for spell damage
	}

	// Get encounter ID (would need to be added to event context)
	encounterID, exists := event.GetStringContext(events.ContextEncounterID)
	if !exists {
		return nil // No encounter ID in spell damage event
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

	// Spell damage applied successfully (removed excessive logging)

	return nil
}

// Priority returns the handler priority
func (h *SpellDamageHandler) Priority() int {
	return 100
}
