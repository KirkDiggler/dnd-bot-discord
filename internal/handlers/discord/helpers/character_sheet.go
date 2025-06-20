package helpers

import (
	"fmt"
	"log"

	"github.com/KirkDiggler/dnd-bot-discord/internal/handlers/discord/dnd/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services"
	"github.com/bwmarrin/discordgo"
)

// ShowCharacterSheet is a helper function to display a character sheet
// It handles fetching the character, verifying ownership, and building the sheet
func ShowCharacterSheet(s *discordgo.Session, i *discordgo.InteractionCreate, characterID string, userID string, services *services.Provider, updateMessage bool) error {
	// Get the character
	char, err := services.CharacterService.GetByID(characterID)
	if err != nil {
		log.Printf("Error getting character %s: %v", characterID, err)
		return err
	}

	// Verify ownership
	if char.OwnerID != userID {
		log.Printf("User %s attempted to view character %s owned by %s", userID, characterID, char.OwnerID)
		return fmt.Errorf("unauthorized access to character")
	}

	// Build the sheet
	embed := character.BuildCharacterSheetEmbed(char)
	components := character.BuildCharacterSheetComponents(characterID)

	if updateMessage {
		// Update existing message
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Embeds:     []*discordgo.MessageEmbed{embed},
				Components: components,
			},
		})
	} else {
		// Send new ephemeral response
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds:     []*discordgo.MessageEmbed{embed},
				Components: components,
				Flags:      discordgo.MessageFlagsEphemeral,
			},
		})
	}
}
