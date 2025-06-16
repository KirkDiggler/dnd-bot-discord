package character

import (
	"context"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	characterService "github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
)

// ProficiencyChoicesHandler handles proficiency selection
type ProficiencyChoicesHandler struct {
	characterService characterService.Service
}

// ProficiencyChoicesHandlerConfig holds configuration
type ProficiencyChoicesHandlerConfig struct {
	CharacterService characterService.Service
}

// NewProficiencyChoicesHandler creates a new handler
func NewProficiencyChoicesHandler(cfg *ProficiencyChoicesHandlerConfig) *ProficiencyChoicesHandler {
	return &ProficiencyChoicesHandler{
		characterService: cfg.CharacterService,
	}
}

// ProficiencyChoicesRequest represents the request
type ProficiencyChoicesRequest struct {
	Session     *discordgo.Session
	Interaction *discordgo.InteractionCreate
	RaceKey     string
	ClassKey    string
}

// Handle processes proficiency choices
func (h *ProficiencyChoicesHandler) Handle(req *ProficiencyChoicesRequest) error {
	// Update the message
	err := req.Session.InteractionRespond(req.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content: "Loading proficiency choices...",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to acknowledge interaction: %w", err)
	}
	
	// Use character service to resolve choices
	choices, err := h.characterService.ResolveChoices(context.Background(), &characterService.ResolveChoicesInput{
		RaceKey:  req.RaceKey,
		ClassKey: req.ClassKey,
	})
	if err != nil {
		return h.respondWithError(req, fmt.Sprintf("Failed to load proficiency choices: %v", err))
	}

	// Get race and class details for display
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
		Title:       "Create New Character - Proficiencies",
		Description: fmt.Sprintf("**Race:** %s\n**Class:** %s\n\nChoose your character's proficiencies.", race.Name, class.Name),
		Color:       0x5865F2,
		Fields:      []*discordgo.MessageEmbedField{},
	}

	// Show stubbed ability scores
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "📊 Ability Scores",
		Value:  "**STR:** 15 (+2)\n**DEX:** 14 (+2)\n**CON:** 13 (+1)\n**INT:** 12 (+1)\n**WIS:** 10 (+0)\n**CHA:** 8 (-1)",
		Inline: true,
	})

	// Show class proficiencies (automatic)
	if len(class.Proficiencies) > 0 {
		profStrings := []string{}
		for _, prof := range class.Proficiencies {
			profStrings = append(profStrings, prof.Name)
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("✅ %s Proficiencies", class.Name),
			Value:  strings.Join(profStrings, "\n"),
			Inline: true,
		})
	}

	// Show proficiency choices from service
	hasChoices := len(choices.ProficiencyChoices) > 0
	
	for _, choice := range choices.ProficiencyChoices {
		choiceDesc := fmt.Sprintf("Choose %d from %d options", choice.Choose, len(choice.Options))
		if choice.Description != "" {
			choiceDesc = choice.Description
		}
		
		// Show choice type icon
		icon := "🎯"
		if strings.Contains(choice.ID, "race") || strings.Contains(strings.ToLower(choice.Name), "racial") {
			icon = "🏃"
		}
		
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%s %s", icon, choice.Name),
			Value:  choiceDesc,
			Inline: false,
		})
		
		fmt.Printf("DEBUG: Proficiency choice: %s (%d options)\n", choice.Name, len(choice.Options))
	}

	// Progress
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "Progress",
		Value:  "✅ Step 1: Race\n✅ Step 2: Class\n✅ Step 3: Abilities\n⏳ Step 4: Proficiencies\n⏳ Step 5: Details",
		Inline: false,
	})

	// Components
	components := []discordgo.MessageComponent{}

	if hasChoices {
		// Add a button to start choosing proficiencies
		components = append(components, discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Choose Proficiencies",
					Style:    discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("character_create:select_proficiencies:%s:%s", req.RaceKey, req.ClassKey),
					Emoji: &discordgo.ComponentEmoji{
						Name: "🎯",
					},
				},
			},
		})
		embed.Footer = &discordgo.MessageEmbedFooter{
			Text: "Click to select your bonus proficiencies",
		}
	} else {
		// No choices needed, go straight to next step
		components = append(components, discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Continue",
					Style:    discordgo.SuccessButton,
					CustomID: fmt.Sprintf("character_create:character_details:%s:%s", req.RaceKey, req.ClassKey),
					Emoji: &discordgo.ComponentEmoji{
						Name: "➡️",
					},
				},
			},
		})
		embed.Footer = &discordgo.MessageEmbedFooter{
			Text: "No additional proficiency choices needed",
		}
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
func (h *ProficiencyChoicesHandler) respondWithError(req *ProficiencyChoicesRequest, message string) error {
	content := fmt.Sprintf("❌ %s", message)
	_, err := req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
		Content: &content,
		Embeds:  &[]*discordgo.MessageEmbed{},
	})
	return err
}