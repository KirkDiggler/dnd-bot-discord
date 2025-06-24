package combat

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// parseCustomID splits a custom ID by colon
func parseCustomID(customID string) []string {
	return strings.Split(customID, ":")
}

// respondError sends an error response
func respondError(s *discordgo.Session, i *discordgo.InteractionCreate, message string, err error) error {
	content := fmt.Sprintf("❌ %s", message)
	if err != nil {
		content += fmt.Sprintf(": %v", err)
		log.Printf("Combat error - %s: %v", message, err)
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

// respondEditError edits a deferred response with an error
func respondEditError(s *discordgo.Session, i *discordgo.InteractionCreate, message string, err error) error {
	content := fmt.Sprintf("❌ %s", message)
	if err != nil {
		content += fmt.Sprintf(": %v", err)
		log.Printf("Combat error - %s: %v", message, err)
	}

	_, editErr := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	})
	return editErr
}