package character

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e"
)

// ProficiencyChoicesHandler handles proficiency selection
type ProficiencyChoicesHandler struct {
	dndClient dnd5e.Client
}

// ProficiencyChoicesHandlerConfig holds configuration
type ProficiencyChoicesHandlerConfig struct {
	DNDClient dnd5e.Client
}

// NewProficiencyChoicesHandler creates a new handler
func NewProficiencyChoicesHandler(cfg *ProficiencyChoicesHandlerConfig) *ProficiencyChoicesHandler {
	return &ProficiencyChoicesHandler{
		dndClient: cfg.DNDClient,
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

	// Get race and class details
	race, err := h.dndClient.GetRace(req.RaceKey)
	if err != nil {
		return h.respondWithError(req, "Failed to fetch race details.")
	}

	class, err := h.dndClient.GetClass(req.ClassKey)
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

	// Show proficiency choices from class
	hasChoices := false
	fmt.Printf("DEBUG: %s has %d proficiency choices\n", class.Name, len(class.ProficiencyChoices))
	if len(class.ProficiencyChoices) > 0 {
		for i, choice := range class.ProficiencyChoices {
			if choice != nil && len(choice.Options) > 0 {
				hasChoices = true
				choiceDesc := fmt.Sprintf("Choose %d from %d options", choice.Count, len(choice.Options))
				embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
					Name:   fmt.Sprintf("🎯 %s", choice.Name),
					Value:  choiceDesc,
					Inline: false,
				})
				fmt.Printf("DEBUG: Choice %d: %s (%d options)\n", i, choice.Name, len(choice.Options))
			}
		}
	} else {
		fmt.Printf("DEBUG: No proficiency choices found for %s\n", class.Name)
	}

	// Show proficiency choices from race
	if race.StartingProficiencyOptions != nil && len(race.StartingProficiencyOptions.Options) > 0 {
		hasChoices = true
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("🏃 %s Bonus", race.Name),
			Value:  fmt.Sprintf("Choose %d proficiency", race.StartingProficiencyOptions.Count),
			Inline: false,
		})
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