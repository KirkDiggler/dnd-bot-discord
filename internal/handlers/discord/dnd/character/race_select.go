package character

import (
	"context"
	"fmt"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	characterService "github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/bwmarrin/discordgo"
)

// RaceSelectHandler handles the race selection interaction
type RaceSelectHandler struct {
	characterService characterService.Service
}

// RaceSelectHandlerConfig holds configuration for the race select handler
type RaceSelectHandlerConfig struct {
	CharacterService characterService.Service
}

// NewRaceSelectHandler creates a new race selection handler
func NewRaceSelectHandler(cfg *RaceSelectHandlerConfig) *RaceSelectHandler {
	return &RaceSelectHandler{
		characterService: cfg.CharacterService,
	}
}

// RaceSelectRequest represents a race selection interaction
type RaceSelectRequest struct {
	Session     *discordgo.Session
	Interaction *discordgo.InteractionCreate
	RaceKey     string
}

// Handle processes the race selection
func (h *RaceSelectHandler) Handle(req *RaceSelectRequest) error {
	// Update the message instead of creating a new one
	err := req.Session.InteractionRespond(req.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content: "Loading race details...",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to acknowledge interaction: %w", err)
	}

	// Get the draft character for this user
	draftChar, err := h.characterService.GetOrCreateDraftCharacter(
		context.Background(),
		req.Interaction.Member.User.ID,
		req.Interaction.GuildID,
	)
	if err != nil {
		return h.respondWithError(req, "Failed to get character draft. Please try again.")
	}

	// Update the draft with the selected race
	updatedChar, err := h.characterService.UpdateDraftCharacter(context.Background(), draftChar.ID, &characterService.UpdateDraftInput{
		RaceKey: &req.RaceKey,
	})
	if err != nil {
		return h.respondWithError(req, "Failed to update character race. Please try again.")
	}
	// Use the updated character's race for display
	race := updatedChar.Race

	// Build the updated embed with race details
	embed := h.buildRaceDetailsEmbed(race)

	// Fetch all races to rebuild the dropdown
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

	// Create components with race dropdown (active) and Next button
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
				discordgo.Button{
					Label:    "Next: Choose Class",
					Style:    discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("character_create:show_classes:%s", req.RaceKey),
					Emoji: &discordgo.ComponentEmoji{
						Name: "▶️",
					},
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

// buildRaceDetailsEmbed creates an embed showing race details
func (h *RaceSelectHandler) buildRaceDetailsEmbed(race *entities.Race) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Title:       "Create New Character",
		Description: fmt.Sprintf("**Selected Race:** %s\n\nGreat choice! %s characters have unique abilities and traits.", race.Name, race.Name),
		Color:       0x5865F2,
		Fields:      []*discordgo.MessageEmbedField{},
	}

	// Add speed
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "⚡ Speed",
		Value:  fmt.Sprintf("%d feet", race.Speed),
		Inline: true,
	})

	// Add ability bonuses
	if len(race.AbilityBonuses) > 0 {
		bonuses := []string{}
		for _, bonus := range race.AbilityBonuses {
			if bonus.Bonus > 0 {
				bonuses = append(bonuses, fmt.Sprintf("%s +%d", bonus.Attribute, bonus.Bonus))
			}
		}
		if len(bonuses) > 0 {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "📊 Ability Bonuses",
				Value:  strings.Join(bonuses, ", "),
				Inline: true,
			})
		}
	}

	// Add starting proficiencies
	if len(race.StartingProficiencies) > 0 {
		profs := []string{}
		for _, prof := range race.StartingProficiencies {
			profs = append(profs, prof.Name)
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "🎯 Racial Proficiencies",
			Value:  strings.Join(profs, ", "),
			Inline: false,
		})
	}

	// Add step indicator
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "Progress",
		Value:  "✅ Step 1: Race\n⏳ Step 2: Class\n⏳ Step 3: Abilities\n⏳ Step 4: Details",
		Inline: false,
	})

	embed.Footer = &discordgo.MessageEmbedFooter{
		Text: "Click 'Next' to choose your class",
	}

	return embed
}

// respondWithError updates the message with an error
func (h *RaceSelectHandler) respondWithError(req *RaceSelectRequest, message string) error {
	content := fmt.Sprintf("❌ %s", message)
	_, err := req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
		Content: &content,
		Embeds:  &[]*discordgo.MessageEmbed{},
	})
	return err
}
