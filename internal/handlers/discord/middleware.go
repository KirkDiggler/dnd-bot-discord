package discord

import (
	"fmt"
	"log"
	"runtime/debug"

	"github.com/bwmarrin/discordgo"
)

// RecoverMiddleware wraps handler functions to recover from panics
func RecoverMiddleware(handlerName string, handler func(*discordgo.Session, *discordgo.InteractionCreate)) func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		defer func() {
			if r := recover(); r != nil {
				// Log the panic with stack trace
				log.Printf("PANIC in %s handler: %v\nStack trace:\n%s", handlerName, r, debug.Stack())
				
				// Try to respond to the user if possible
				respondWithError(s, i, fmt.Sprintf("An unexpected error occurred: %v", r))
			}
		}()
		
		handler(s, i)
	}
}

// respondWithError attempts to send an error message to the user
func respondWithError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	// Try different response methods based on interaction state
	responses := []func() error{
		// Try responding if not yet responded
		func() error {
			return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("❌ %s", message),
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		},
		// Try editing if already responded
		func() error {
			_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &message,
			})
			return err
		},
		// Try followup if other methods fail
		func() error {
			_, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: fmt.Sprintf("❌ %s", message),
				Flags:   discordgo.MessageFlagsEphemeral,
			})
			return err
		},
	}
	
	for _, respond := range responses {
		if err := respond(); err == nil {
			return
		}
	}
	
	// If all methods fail, at least we logged the error
	log.Printf("Failed to send error response to user: %s", message)
}