package character

import (
	"context"
	"fmt"
	"log"

	character2 "github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
	"github.com/bwmarrin/discordgo"
)

// ShowDivineDomainSelection displays the divine domain selection UI
func (h *ClassFeaturesHandler) ShowDivineDomainSelection(req *InteractionRequest) error {
	// Get the character
	char, err := h.characterService.GetByID(req.CharacterID)
	if err != nil {
		return fmt.Errorf("failed to get character: %w", err)
	}

	// Get pending feature choices to find the divine domain choice
	ctx := context.TODO()
	pendingChoices, err := h.characterService.GetPendingFeatureChoices(ctx, req.CharacterID)
	if err != nil {
		return fmt.Errorf("failed to get pending choices: %w", err)
	}

	// Find the divine domain choice
	var domainChoice *rulebook.FeatureChoice
	for _, choice := range pendingChoices {
		if choice.Type == rulebook.FeatureChoiceTypeDivineDomain {
			domainChoice = choice
			break
		}
	}

	if domainChoice == nil {
		return fmt.Errorf("no divine domain choice found for character")
	}

	// Convert rulebook options to Discord select menu options
	var domainOptions []discordgo.SelectMenuOption
	for _, option := range domainChoice.Options {
		domainOptions = append(domainOptions, discordgo.SelectMenuOption{
			Label:       option.Name,
			Value:       option.Key,
			Description: option.Description,
		})
	}

	// Create embed
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Choose Your %s", domainChoice.Name),
		Description: fmt.Sprintf("**%s**, %s", char.Name, domainChoice.Description),
		Color:       0xFFD700, // Gold for cleric
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Your Selection",
				Value:  "Your choice grants you domain spells and other features when you reach certain levels.",
				Inline: false,
			},
		},
	}

	// Add progress field
	classKey := ""
	if char.Class != nil {
		classKey = char.Class.Key
	}
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "Progress",
		Value:  BuildProgressValue(classKey, "class_features"),
		Inline: false,
	})

	// Create components
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    fmt.Sprintf("character_create:class_features:%s:divine_domain", char.ID),
					Placeholder: "Select your divine domain...",
					Options:     domainOptions,
				},
			},
		},
	}

	// Update the interaction response
	return req.Session.InteractionRespond(req.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
			Flags:      discordgo.MessageFlagsEphemeral,
		},
	})
}

// handleDivineDomain stores the cleric's divine domain selection
func (h *ClassFeaturesHandler) handleDivineDomain(req *ClassFeaturesRequest, char *character2.Character) error {
	// Get the divine domains to find the display name
	domains := rulebook.GetDivineDomains()
	var selectedName string
	for _, domain := range domains {
		if domain.Key == req.Selection {
			selectedName = domain.Name
			break
		}
	}

	// If we couldn't find the display name, use the key
	if selectedName == "" {
		selectedName = req.Selection
	}

	// Find the divine domain feature and update its metadata
	updateFeatureMetadata(char, "divine_domain", map[string]any{
		"domain":            req.Selection,
		"selection_display": selectedName,
	})
	log.Printf("Set divine domain for %s to %s", char.Name, selectedName)

	// TODO: Add domain spells to character's known spells when spellcasting is implemented

	// After saving domain selection, check if this requires additional steps
	// For Knowledge Domain, we need to trigger skill and language selection
	if req.Selection == "knowledge" {
		log.Printf("Knowledge Domain selected, flow service should handle next steps")
		// The flow service will handle showing the skill/language selection steps
		// when the character progresses to the next step
	}

	return nil
}
