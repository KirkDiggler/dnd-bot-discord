package discord

import (
	"context"
	"log"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/handlers/discord/dnd/dungeon"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/encounter"
	"github.com/bwmarrin/discordgo"
)

// CombatLogUpdater handles updating the public combat log message
type CombatLogUpdater struct {
	session          *discordgo.Session
	encounterService encounter.Service
}

// NewCombatLogUpdater creates a new combat log updater
func NewCombatLogUpdater(session *discordgo.Session, encounterService encounter.Service) *CombatLogUpdater {
	return &CombatLogUpdater{
		session:          session,
		encounterService: encounterService,
	}
}

// UpdateCombatLog updates the public combat log message for an encounter
func (u *CombatLogUpdater) UpdateCombatLog(ctx context.Context, encounterID string) error {
	// Get the encounter
	enc, err := u.encounterService.GetEncounter(ctx, encounterID)
	if err != nil {
		return err
	}

	// Skip if no message ID
	if enc.MessageID == "" || enc.ChannelID == "" {
		return nil
	}

	// For dungeon encounters, we need the room info
	// Since we don't have it here, we'll use a generic room
	room := &dungeon.Room{
		Name: "Combat Encounter",
		Type: dungeon.RoomTypeCombat,
	}

	// Update the message
	err = dungeon.UpdateCombatLogMessage(u.session, enc.ChannelID, enc.MessageID, room, enc)
	if err != nil {
		log.Printf("Failed to update combat log message: %v", err)
		// Don't return error - we don't want to fail the whole action
	}

	// Check if combat ended and send summary
	if enc.Status == entities.EncounterStatusCompleted {
		// Send combat end message
		err = dungeon.CreateCombatEndMessage(u.session, enc.ChannelID, room, enc, nil)
		if err != nil {
			log.Printf("Failed to create combat end message: %v", err)
		}
	}

	return nil
}
