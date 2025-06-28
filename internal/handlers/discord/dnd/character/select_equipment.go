package character

import (
	"context"
	"fmt"
	"strings"

	characterService "github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/bwmarrin/discordgo"
)

// SelectEquipmentHandler handles individual equipment selection
type SelectEquipmentHandler struct {
	characterService characterService.Service
}

// SelectEquipmentHandlerConfig holds configuration
type SelectEquipmentHandlerConfig struct {
	CharacterService characterService.Service
}

// NewSelectEquipmentHandler creates a new handler
func NewSelectEquipmentHandler(cfg *SelectEquipmentHandlerConfig) *SelectEquipmentHandler {
	return &SelectEquipmentHandler{
		characterService: cfg.CharacterService,
	}
}

// SelectEquipmentRequest represents the request
type SelectEquipmentRequest struct {
	Session     *discordgo.Session
	Interaction *discordgo.InteractionCreate
	RaceKey     string
	ClassKey    string
	ChoiceIndex int
}

// Handle processes equipment selection
func (h *SelectEquipmentHandler) Handle(req *SelectEquipmentRequest) error {
	// For nested equipment flow, the interaction is already acknowledged
	// Try to update first, if that fails then this is the initial interaction
	content := "Loading equipment options..."
	err := req.Session.InteractionRespond(req.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
	if err != nil && strings.Contains(err.Error(), "already been acknowledged") {
		// Interaction already acknowledged, just edit instead
		_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		if err != nil {
			return fmt.Errorf("failed to update interaction: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to acknowledge interaction: %w", err)
	}

	// Get race and class details
	race, err := h.characterService.GetRace(context.Background(), req.RaceKey)
	if err != nil {
		return h.respondWithError(req, "Failed to fetch race details.")
	}

	class, err := h.characterService.GetClass(context.Background(), req.ClassKey)
	if err != nil {
		return h.respondWithError(req, "Failed to fetch class details.")
	}

	// Get equipment choices
	choices, err := h.characterService.ResolveChoices(context.Background(), &characterService.ResolveChoicesInput{
		RaceKey:  req.RaceKey,
		ClassKey: req.ClassKey,
	})
	if err != nil {
		return h.respondWithError(req, "Failed to resolve equipment choices.")
	}

	// Validate choice index
	if req.ChoiceIndex >= len(choices.EquipmentChoices) {
		// No more choices, continue to character details
		return h.continueToCharacterDetails(req)
	}

	choice := choices.EquipmentChoices[req.ChoiceIndex]

	// Create embed
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Equipment Choice %d of %d", req.ChoiceIndex+1, len(choices.EquipmentChoices)),
		Description: fmt.Sprintf("**Race:** %s\n**Class:** %s\n\n**%s**", race.Name, class.Name, choice.Name),
		Color:       0x5865F2,
		Fields:      []*discordgo.MessageEmbedField{},
	}

	// Add choice description if available
	if choice.Description != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "📋 Description",
			Value:  choice.Description,
			Inline: false,
		})
	}

	// Create select menu options
	var selectOptions []discordgo.SelectMenuOption
	for _, opt := range choice.Options {
		option := discordgo.SelectMenuOption{
			Label: opt.Name,
			Value: opt.Key,
		}

		// Add description based on equipment type
		if strings.Contains(opt.Key, "weapon") || strings.Contains(opt.Name, "sword") || strings.Contains(opt.Name, "axe") {
			option.Description = h.getWeaponDescription(opt.Key)
		} else if strings.Contains(opt.Key, "armor") || strings.Contains(opt.Name, "armor") {
			option.Description = h.getArmorDescription(opt.Key)
		} else if opt.Description != "" {
			option.Description = opt.Description
		}

		// If this is a nested choice with bundle items, enhance the description
		if strings.HasPrefix(opt.Key, "nested-") && len(opt.BundleItems) > 0 {
			// Show what's included in the bundle
			if opt.Description != "" {
				option.Description = opt.Description + " (includes shield)"
			} else {
				option.Description = "Includes shield"
			}
		}

		selectOptions = append(selectOptions, option)
	}

	// Progress
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "Progress",
		Value:  "✅ Step 1: Race\n✅ Step 2: Class\n✅ Step 3: Abilities\n✅ Step 4: Proficiencies\n⏳ Step 5: Equipment\n⏳ Step 6: Details",
		Inline: false,
	})

	embed.Footer = &discordgo.MessageEmbedFooter{
		Text: fmt.Sprintf("Select %d item(s) from the list", choice.Choose),
	}

	// Create components
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    fmt.Sprintf("character_create:confirm_equipment:%s:%s:%d", req.RaceKey, req.ClassKey, req.ChoiceIndex),
					Placeholder: fmt.Sprintf("Select %d equipment", choice.Choose),
					MinValues:   &choice.Choose,
					MaxValues:   choice.Choose,
					Options:     selectOptions,
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

// getWeaponDescription returns a description for common weapons
func (h *SelectEquipmentHandler) getWeaponDescription(key string) string {
	// Common weapon descriptions - in a real implementation, fetch from API
	weaponDescriptions := map[string]string{
		"longsword":      "1d8 slashing, versatile (1d10)",
		"shortsword":     "1d6 piercing, finesse, light",
		"battleaxe":      "1d8 slashing, versatile (1d10)",
		"handaxe":        "1d6 slashing, light, thrown (20/60)",
		"warhammer":      "1d8 bludgeoning, versatile (1d10)",
		"mace":           "1d6 bludgeoning",
		"greataxe":       "1d12 slashing, heavy, two-handed",
		"greatsword":     "2d6 slashing, heavy, two-handed",
		"rapier":         "1d8 piercing, finesse",
		"scimitar":       "1d6 slashing, finesse, light",
		"shortbow":       "1d6 piercing, range 80/320",
		"longbow":        "1d8 piercing, range 150/600",
		"light-crossbow": "1d8 piercing, range 80/320",
		"shield":         "+2 AC",
	}

	if desc, ok := weaponDescriptions[key]; ok {
		return desc
	}
	return ""
}

// getArmorDescription returns a description for common armor
func (h *SelectEquipmentHandler) getArmorDescription(key string) string {
	// Common armor descriptions
	armorDescriptions := map[string]string{
		"leather-armor":   "11 + Dex modifier",
		"scale-mail":      "14 + Dex (max 2)",
		"chain-mail":      "16 AC",
		"chain-shirt":     "13 + Dex (max 2)",
		"padded-armor":    "11 + Dex modifier",
		"studded-leather": "12 + Dex modifier",
		"hide-armor":      "12 + Dex (max 2)",
		"ring-mail":       "14 AC",
		"splint-armor":    "17 AC",
		"plate-armor":     "18 AC",
	}

	if desc, ok := armorDescriptions[key]; ok {
		return desc
	}
	return ""
}

// continueToCharacterDetails moves to the character details step
func (h *SelectEquipmentHandler) continueToCharacterDetails(req *SelectEquipmentRequest) error {
	detailsReq := &CharacterDetailsRequest{
		Session:     req.Session,
		Interaction: req.Interaction,
		RaceKey:     req.RaceKey,
		ClassKey:    req.ClassKey,
	}

	// Get the handler and call it
	handler := NewCharacterDetailsHandler(&CharacterDetailsHandlerConfig{
		CharacterService: h.characterService,
	})

	return handler.Handle(detailsReq)
}

// respondWithError updates the message with an error
func (h *SelectEquipmentHandler) respondWithError(req *SelectEquipmentRequest, message string) error {
	content := fmt.Sprintf("❌ %s", message)
	_, err := req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
		Content: &content,
		Embeds:  &[]*discordgo.MessageEmbed{},
	})
	return err
}
