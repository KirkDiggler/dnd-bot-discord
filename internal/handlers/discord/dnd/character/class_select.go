package character

import (
	"context"
	"fmt"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	characterService "github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/bwmarrin/discordgo"
)

// ClassSelectHandler handles the class selection interaction
type ClassSelectHandler struct {
	characterService characterService.Service
}

// ClassSelectHandlerConfig holds configuration for the class select handler
type ClassSelectHandlerConfig struct {
	CharacterService characterService.Service
}

// NewClassSelectHandler creates a new class selection handler
func NewClassSelectHandler(cfg *ClassSelectHandlerConfig) *ClassSelectHandler {
	return &ClassSelectHandler{
		characterService: cfg.CharacterService,
	}
}

// ClassSelectRequest represents a class selection interaction
type ClassSelectRequest struct {
	Session     *discordgo.Session
	Interaction *discordgo.InteractionCreate
	RaceKey     string
	ClassKey    string
}

// Handle processes the class selection
func (h *ClassSelectHandler) Handle(req *ClassSelectRequest) error {
	// Update the message
	err := req.Session.InteractionRespond(req.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content: "Loading class details...",
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

	// Update the draft with the selected class
	updatedChar, err := h.characterService.UpdateDraftCharacter(context.Background(), draftChar.ID, &characterService.UpdateDraftInput{
		ClassKey: &req.ClassKey,
	})
	if err != nil {
		return h.respondWithError(req, "Failed to update character class. Please try again.")
	}

	// Use the updated character for display
	race := updatedChar.Race
	class := updatedChar.Class

	// Build the summary embed
	embed := h.buildSummaryEmbed(race, class)

	// Get all races and classes for the dropdowns
	races, err := h.characterService.GetRaces(context.Background())
	if err != nil {
		return h.respondWithError(req, "Failed to fetch races. Please try again.")
	}

	classes, err := h.characterService.GetClasses(context.Background())
	if err != nil {
		return h.respondWithError(req, "Failed to fetch classes. Please try again.")
	}

	// Build race options
	raceOptions := make([]discordgo.SelectMenuOption, 0, len(races))
	for _, r := range races {
		option := discordgo.SelectMenuOption{
			Label:   r.Name,
			Value:   r.Key,
			Default: r.Key == req.RaceKey,
		}
		raceOptions = append(raceOptions, option)
	}

	// Build class options
	classOptions := make([]discordgo.SelectMenuOption, 0, len(classes))
	for _, c := range classes {
		option := discordgo.SelectMenuOption{
			Label:   c.Name,
			Value:   c.Key,
			Default: c.Key == req.ClassKey,
		}
		classOptions = append(classOptions, option)
	}

	// Create components
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    "character_create:race_select",
					Placeholder: race.Name,
					Options:     raceOptions,
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    fmt.Sprintf("character_create:class_select:%s", req.RaceKey),
					Placeholder: class.Name,
					Options:     classOptions,
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Next: Roll Abilities",
					Style:    discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("character_create:ability_scores:%s:%s", req.RaceKey, req.ClassKey),
					Emoji: &discordgo.ComponentEmoji{
						Name: "🎲",
					},
				},
			},
		},
	}

	// Update the message
	_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
		Content:    &[]string{""}[0], // Clear loading message
	})

	return err
}

// buildSummaryEmbed creates an embed showing race and class summary
func (h *ClassSelectHandler) buildSummaryEmbed(race *entities.Race, class *entities.Class) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Title:       "Create New Character",
		Description: fmt.Sprintf("**Race:** %s\n**Class:** %s\n\nExcellent choices! Let's review your selections.", race.Name, class.Name),
		Color:       0x5865F2,
		Fields:      []*discordgo.MessageEmbedField{},
	}

	// Race details
	raceDetails := []string{}
	raceDetails = append(raceDetails, fmt.Sprintf("Speed: %d feet", race.Speed))

	if len(race.AbilityBonuses) > 0 {
		bonuses := []string{}
		for _, bonus := range race.AbilityBonuses {
			if bonus.Bonus > 0 {
				bonuses = append(bonuses, fmt.Sprintf("%s +%d", bonus.Attribute, bonus.Bonus))
			}
		}
		if len(bonuses) > 0 {
			raceDetails = append(raceDetails, "Bonuses: "+strings.Join(bonuses, ", "))
		}
	}

	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("🏃 %s Traits", race.Name),
		Value:  strings.Join(raceDetails, "\n"),
		Inline: true,
	})

	// Class details
	classDetails := []string{}
	classDetails = append(classDetails, fmt.Sprintf("Hit Die: d%d", class.HitDie))

	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("⚔️ %s Features", class.Name),
		Value:  strings.Join(classDetails, "\n"),
		Inline: true,
	})

	// Progress indicator
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "Progress",
		Value:  "✅ Step 1: Race\n✅ Step 2: Class\n⏳ Step 3: Abilities\n⏳ Step 4: Details",
		Inline: false,
	})

	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "📊 Starting Hit Points",
		Value:  fmt.Sprintf("Base: %d (will add Constitution modifier)", class.HitDie),
		Inline: false,
	})

	embed.Footer = &discordgo.MessageEmbedFooter{
		Text: "Click 'Next' to roll your ability scores",
	}

	return embed
}

// respondWithError updates the message with an error
func (h *ClassSelectHandler) respondWithError(req *ClassSelectRequest, message string) error {
	content := fmt.Sprintf("❌ %s", message)
	_, err := req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
		Content: &content,
		Embeds:  &[]*discordgo.MessageEmbed{},
	})
	return err
}
