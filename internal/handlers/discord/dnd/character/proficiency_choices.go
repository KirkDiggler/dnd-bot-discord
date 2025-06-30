package character

import (
	"context"
	"fmt"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"strings"

	characterService "github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/bwmarrin/discordgo"
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

	// Get the draft character to show actual ability scores
	draftChar, err := h.characterService.GetOrCreateDraftCharacter(
		context.Background(),
		req.Interaction.Member.User.ID,
		req.Interaction.GuildID,
	)
	if err == nil && draftChar.Attributes != nil {
		// Build ability score display from actual character
		scoreLines := []string{}
		abilities := []struct {
			name string
			attr character.Attribute
		}{
			{"STR", character.AttributeStrength},
			{"DEX", character.AttributeDexterity},
			{"CON", character.AttributeConstitution},
			{"INT", character.AttributeIntelligence},
			{"WIS", character.AttributeWisdom},
			{"CHA", character.AttributeCharisma},
		}

		for _, ability := range abilities {
			if score, ok := draftChar.Attributes[ability.attr]; ok && score != nil {
				modifier := score.Bonus
				modStr := fmt.Sprintf("%+d", modifier)
				scoreLines = append(scoreLines, fmt.Sprintf("**%s:** %d (%s)", ability.name, score.Score, modStr))
			} else {
				scoreLines = append(scoreLines, fmt.Sprintf("**%s:** -", ability.name))
			}
		}

		if len(scoreLines) > 0 {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "📊 Ability Scores",
				Value:  strings.Join(scoreLines, "\n"),
				Inline: true,
			})
		}
	} else {
		// Fallback if we can't get character data
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "📊 Ability Scores",
			Value:  "*Ability scores will be shown here*",
			Inline: true,
		})
	}

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
		Value:  BuildProgressValue(req.ClassKey, "proficiencies"),
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
