package character

import (
	"context"
	"fmt"

	characterService "github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/bwmarrin/discordgo"
)

// ShowClassesHandler handles showing the class selection after race is chosen
type ShowClassesHandler struct {
	characterService characterService.Service
}

// ShowClassesHandlerConfig holds configuration for the show classes handler
type ShowClassesHandlerConfig struct {
	CharacterService characterService.Service
}

// NewShowClassesHandler creates a new show classes handler
func NewShowClassesHandler(cfg *ShowClassesHandlerConfig) *ShowClassesHandler {
	return &ShowClassesHandler{
		characterService: cfg.CharacterService,
	}
}

// ShowClassesRequest represents a request to show class options
type ShowClassesRequest struct {
	Session     *discordgo.Session
	Interaction *discordgo.InteractionCreate
	RaceKey     string
}

// Handle shows the class selection screen
func (h *ShowClassesHandler) Handle(req *ShowClassesRequest) error {
	// Update the message
	err := req.Session.InteractionRespond(req.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content: "Loading classes...",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to acknowledge interaction: %w", err)
	}

	// Fetch the race details again to show in the embed
	race, err := h.characterService.GetRace(context.Background(), req.RaceKey)
	if err != nil {
		return h.respondWithError(req, "Failed to fetch race details. Please try again.")
	}

	// Fetch available classes
	classes, err := h.characterService.GetClasses(context.Background())
	if err != nil {
		return h.respondWithError(req, "Failed to fetch classes. Please try again.")
	}

	// Build class selection menu
	options := make([]discordgo.SelectMenuOption, 0, len(classes))
	for _, class := range classes {
		options = append(options, discordgo.SelectMenuOption{
			Label: class.Name,
			Value: class.Key,
		})
	}

	// Get all races for the race dropdown
	races, err := h.characterService.GetRaces(context.Background())
	if err != nil {
		return h.respondWithError(req, "Failed to fetch races. Please try again.")
	}

	// Build race options with the selected one marked as default
	raceOptions := make([]discordgo.SelectMenuOption, 0, len(races))
	for _, r := range races {
		option := discordgo.SelectMenuOption{
			Label:   r.Name,
			Value:   r.Key,
			Default: r.Key == req.RaceKey,
		}
		raceOptions = append(raceOptions, option)
	}

	// Create embed
	embed := &discordgo.MessageEmbed{
		Title:       "Create New Character",
		Description: fmt.Sprintf("**Selected Race:** %s\n\nNow choose your class. Your class determines your role in the party and your primary abilities.", race.Name),
		Color:       0x5865F2,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Step 2: Class",
				Value: "Select your character's class from the dropdown below.",
			},
			{
				Name:   "Progress",
				Value:  "✅ Step 1: Race\n⏳ Step 2: Class\n⏳ Step 3: Abilities\n⏳ Step 4: Details",
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "D&D 5e Character Creator",
		},
	}

	// Create components with both race and class dropdowns
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    "character_create:race_select",
					Placeholder: "Change race",
					Options:     raceOptions,
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    fmt.Sprintf("character_create:class_select:%s", req.RaceKey),
					Placeholder: "Choose your class",
					Options:     options,
				},
			},
		},
	}

	// Update the message
	_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
		Content:    &[]string{""}[0], // Clear the loading message
	})

	return err
}

// respondWithError updates the message with an error
func (h *ShowClassesHandler) respondWithError(req *ShowClassesRequest, message string) error {
	content := fmt.Sprintf("❌ %s", message)
	_, err := req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
		Content: &content,
		Embeds:  &[]*discordgo.MessageEmbed{},
	})
	return err
}
