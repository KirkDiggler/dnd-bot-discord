package character

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	characterService "github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
)

// CharacterDetailsHandler handles character name and final details
type CharacterDetailsHandler struct {
	characterService characterService.Service
}

// CharacterDetailsHandlerConfig holds configuration
type CharacterDetailsHandlerConfig struct {
	CharacterService characterService.Service
}

// NewCharacterDetailsHandler creates a new handler
func NewCharacterDetailsHandler(cfg *CharacterDetailsHandlerConfig) *CharacterDetailsHandler {
	return &CharacterDetailsHandler{
		characterService: cfg.CharacterService,
	}
}

// CharacterDetailsRequest represents the request
type CharacterDetailsRequest struct {
	Session     *discordgo.Session
	Interaction *discordgo.InteractionCreate
	RaceKey     string
	ClassKey    string
}

// Handle processes character details input
func (h *CharacterDetailsHandler) Handle(req *CharacterDetailsRequest) error {
	// Update the message
	err := req.Session.InteractionRespond(req.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content: "Loading character details...",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to acknowledge interaction: %w", err)
	}

	// Get race and class for display
	race, err := h.characterService.GetRace(context.Background(), req.RaceKey)
	if err != nil {
		return h.respondWithError(req, "Failed to fetch race details.")
	}

	class, err := h.characterService.GetClass(context.Background(), req.ClassKey)
	if err != nil {
		return h.respondWithError(req, "Failed to fetch class details.")
	}

	// Create embed
	embed := &discordgo.MessageEmbed{
		Title:       "Character Details",
		Description: fmt.Sprintf("**Race:** %s\n**Class:** %s\n\nAlmost done! Please provide your character's name.", race.Name, class.Name),
		Color:       0x5865F2,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "📊 Summary",
				Value:  "**Abilities:** Assigned\n**Proficiencies:** Selected\n**Equipment:** Standard starting gear",
				Inline: false,
			},
			{
				Name:   "Progress",
				Value:  "✅ Step 1: Race\n✅ Step 2: Class\n✅ Step 3: Abilities\n✅ Step 4: Proficiencies\n⏳ Step 5: Name",
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Click the button below to enter your character name",
		},
	}

	// Create button to trigger modal
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Enter Character Name",
					Style:    discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("character_create:name_modal:%s:%s", req.RaceKey, req.ClassKey),
					Emoji: &discordgo.ComponentEmoji{
						Name: "✏️",
					},
				},
			},
		},
	}

	// Update message
	_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
		Content:    &[]string{""}[0],
	})

	return err
}

// respondWithError updates the message with an error
func (h *CharacterDetailsHandler) respondWithError(req *CharacterDetailsRequest, message string) error {
	content := fmt.Sprintf("❌ %s", message)
	_, err := req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
		Content: &content,
		Embeds:  &[]*discordgo.MessageEmbed{},
	})
	return err
}