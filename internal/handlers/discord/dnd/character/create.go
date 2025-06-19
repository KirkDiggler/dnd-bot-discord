package character

import (
	"context"
	"fmt"

	characterService "github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/bwmarrin/discordgo"
)

// CreateHandler handles the /dnd character create command
type CreateHandler struct {
	characterService characterService.Service
}

// CreateHandlerConfig holds configuration for the create handler
type CreateHandlerConfig struct {
	CharacterService characterService.Service
}

// NewCreateHandler creates a new character creation handler
func NewCreateHandler(cfg *CreateHandlerConfig) *CreateHandler {
	return &CreateHandler{
		characterService: cfg.CharacterService,
	}
}

// CreateRequest represents a request to create a character
type CreateRequest struct {
	Session     *discordgo.Session
	Interaction *discordgo.InteractionCreate
}

// Handle processes the character create command
func (h *CreateHandler) Handle(req *CreateRequest) error {
	// Check if this is a button interaction (going back) or a command
	isUpdate := req.Interaction.Type == discordgo.InteractionMessageComponent

	if isUpdate {
		// This is a button click, update the message
		err := req.Session.InteractionRespond(req.Interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content: "Loading...",
			},
		})
		if err != nil {
			return fmt.Errorf("failed to acknowledge interaction: %w", err)
		}
	} else {
		// This is a slash command, defer the response
		err := req.Session.InteractionRespond(req.Interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags: discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			return fmt.Errorf("failed to defer response: %w", err)
		}
	}

	// Fetch available races from the API
	races, err := h.characterService.GetRaces(context.Background())
	if err != nil {
		return h.respondWithError(req, "Failed to fetch races. Please try again.")
	}

	// Build the race selection menu
	options := make([]discordgo.SelectMenuOption, 0, len(races))
	for _, race := range races {
		// For the list, we'll just show the name since we don't have speed yet
		options = append(options, discordgo.SelectMenuOption{
			Label: race.Name,
			Value: race.Key,
		})
	}

	// Create the select menu component
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    "character_create:race_select",
					Placeholder: "Choose your race",
					Options:     options,
				},
			},
		},
	}

	// Clear any existing ability rolls for a fresh start
	_, err = h.characterService.StartFreshCharacterCreation(
		context.Background(),
		req.Interaction.Member.User.ID,
		req.Interaction.GuildID,
	)
	if err != nil {
		// Non-fatal, continue anyway
		fmt.Printf("Failed to clear draft character rolls: %v\n", err)
	}

	// Create the embed
	embed := &discordgo.MessageEmbed{
		Title:       "Create New Character",
		Description: "Welcome to the D&D 5e Character Creator! Let's start by choosing your race.",
		Color:       0x5865F2, // Discord blurple
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Step 1: Race",
				Value: "Select your character's race from the dropdown below.",
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "D&D 5e Character Creator",
		},
	}

	// Send the response
	_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
		Content:    &[]string{""}[0], // Clear any loading message
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
	})

	return err
}

// respondWithError sends an error message as an ephemeral response
func (h *CreateHandler) respondWithError(req *CreateRequest, message string) error {
	_, err := req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
		Content: &message,
	})
	return err
}
