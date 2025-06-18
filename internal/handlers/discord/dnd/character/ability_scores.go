package character

import (
	"context"
	"fmt"
	"strings"

	characterService "github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/bwmarrin/discordgo"
)

// AbilityScoresHandler handles the ability score generation
type AbilityScoresHandler struct {
	characterService characterService.Service
}

// AbilityScoresHandlerConfig holds configuration
type AbilityScoresHandlerConfig struct {
	CharacterService characterService.Service
}

// NewAbilityScoresHandler creates a new handler
func NewAbilityScoresHandler(cfg *AbilityScoresHandlerConfig) *AbilityScoresHandler {
	return &AbilityScoresHandler{
		characterService: cfg.CharacterService,
	}
}

// AbilityScoresRequest represents the request
type AbilityScoresRequest struct {
	Session     *discordgo.Session
	Interaction *discordgo.InteractionCreate
	RaceKey     string
	ClassKey    string
}

// Handle processes ability score generation
func (h *AbilityScoresHandler) Handle(req *AbilityScoresRequest) error {
	// Update the message
	err := req.Session.InteractionRespond(req.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content: "Choose how to roll your ability scores...",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to acknowledge interaction: %w", err)
	}

	// Get race and class for context
	race, err := h.characterService.GetRace(context.Background(), req.RaceKey)
	if err != nil {
		return h.respondWithError(req, "Failed to fetch race details.")
	}

	class, err := h.characterService.GetClass(context.Background(), req.ClassKey)
	if err != nil {
		return h.respondWithError(req, "Failed to fetch class details.")
	}

	// Show roll mode selection
	embed := &discordgo.MessageEmbed{
		Title:       "Create New Character - Roll Ability Scores",
		Description: fmt.Sprintf("**Race:** %s\n**Class:** %s\n\n**Choose how to roll your ability scores:**", race.Name, class.Name),
		Color:       0x5865F2,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "🎲 Roll All at Once",
				Value:  "Roll all 6 ability scores immediately using 4d6 drop lowest.",
				Inline: false,
			},
			{
				Name:   "🎯 Roll Individually",
				Value:  "Roll each ability score one at a time, seeing the dice details for each roll.",
				Inline: false,
			},
			{
				Name:   "Progress",
				Value:  "✅ Step 1: Race\n✅ Step 2: Class\n⏳ Step 3: Abilities\n⏳ Step 4: Details",
				Inline: false,
			},
		},
	}

	// Show racial bonuses
	if len(race.AbilityBonuses) > 0 {
		bonuses := []string{}
		for _, bonus := range race.AbilityBonuses {
			if bonus.Bonus > 0 {
				bonuses = append(bonuses, fmt.Sprintf("%s +%d", bonus.Attribute, bonus.Bonus))
			}
		}
		if len(bonuses) > 0 {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "🏃 Racial Bonuses",
				Value:  strings.Join(bonuses, "\n") + "\n*(Applied after assignment)*",
				Inline: true,
			})
		}
	}

	embed.Footer = &discordgo.MessageEmbedFooter{
		Text: "Choose your preferred rolling method",
	}

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Roll All at Once",
					Style:    discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("character_create:roll_all:%s:%s", req.RaceKey, req.ClassKey),
					Emoji: &discordgo.ComponentEmoji{
						Name: "🎲",
					},
				},
				discordgo.Button{
					Label:    "Roll Individually",
					Style:    discordgo.SecondaryButton,
					CustomID: fmt.Sprintf("character_create:roll_individual:%s:%s", req.RaceKey, req.ClassKey),
					Emoji: &discordgo.ComponentEmoji{
						Name: "🎯",
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
func (h *AbilityScoresHandler) respondWithError(req *AbilityScoresRequest, message string) error {
	content := fmt.Sprintf("❌ %s", message)
	_, err := req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
		Content: &content,
		Embeds:  &[]*discordgo.MessageEmbed{},
	})
	return err
}
