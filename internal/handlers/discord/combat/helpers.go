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

// isEphemeralInteraction checks if an interaction originated from an ephemeral message
func isEphemeralInteraction(i *discordgo.InteractionCreate) bool {
	return i.Message != nil && i.Message.Flags&discordgo.MessageFlagsEphemeral != 0
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

// DiscordSession interface for Discord operations (for testing)
type DiscordSession interface {
	ChannelMessageEditComplex(edit *discordgo.MessageEdit, options ...discordgo.RequestOption) (*discordgo.Message, error)
}

// updateSharedCombatMessage updates the main shared combat message if MessageID is stored
func updateSharedCombatMessage(s DiscordSession, encounterID, messageID, channelID string, embed *discordgo.MessageEmbed, components []discordgo.MessageComponent) error {
	if messageID == "" || channelID == "" {
		log.Printf("WARNING: Cannot update shared combat message - MessageID=%s, ChannelID=%s", messageID, channelID)
		return nil // Not an error, just can't update
	}

	log.Printf("Updating shared combat message: EncounterID=%s, MessageID=%s, ChannelID=%s", encounterID, messageID, channelID)

	_, err := s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		ID:         messageID,
		Channel:    channelID,
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
	})

	if err != nil {
		log.Printf("Failed to update shared combat message: %v", err)
		// Don't return error - combat should continue even if message update fails
	}

	return nil
}
