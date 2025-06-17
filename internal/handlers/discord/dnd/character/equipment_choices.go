package character

import (
	"context"
	"fmt"
	"strings"

	characterService "github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/bwmarrin/discordgo"
)

// EquipmentChoicesHandler handles the equipment selection overview
type EquipmentChoicesHandler struct {
	characterService characterService.Service
}

// EquipmentChoicesHandlerConfig holds configuration
type EquipmentChoicesHandlerConfig struct {
	CharacterService characterService.Service
}

// NewEquipmentChoicesHandler creates a new handler
func NewEquipmentChoicesHandler(cfg *EquipmentChoicesHandlerConfig) *EquipmentChoicesHandler {
	return &EquipmentChoicesHandler{
		characterService: cfg.CharacterService,
	}
}

// EquipmentChoicesRequest represents the request
type EquipmentChoicesRequest struct {
	Session     *discordgo.Session
	Interaction *discordgo.InteractionCreate
	RaceKey     string
	ClassKey    string
}

// Handle processes equipment choices overview
func (h *EquipmentChoicesHandler) Handle(req *EquipmentChoicesRequest) error {
	// Update the message first
	err := req.Session.InteractionRespond(req.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content: "Loading equipment choices...",
		},
	})
	if err != nil {
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

	// Create embed
	embed := &discordgo.MessageEmbed{
		Title:       "Create New Character - Equipment",
		Description: fmt.Sprintf("**Race:** %s\n**Class:** %s\n\nSelect your starting equipment.", race.Name, class.Name),
		Color:       0x5865F2,
		Fields:      []*discordgo.MessageEmbedField{},
	}

	// Show fixed starting equipment
	if class.StartingEquipment != nil && len(class.StartingEquipment) > 0 {
		fixedItems := []string{}
		for _, se := range class.StartingEquipment {
			if se != nil && se.Equipment != nil {
				if se.Quantity > 1 {
					fixedItems = append(fixedItems, fmt.Sprintf("%dx %s", se.Quantity, se.Equipment.Name))
				} else {
					fixedItems = append(fixedItems, se.Equipment.Name)
				}
			}
		}
		if len(fixedItems) > 0 {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "📦 Starting Equipment",
				Value:  strings.Join(fixedItems, "\n"),
				Inline: false,
			})
		}
	}

	// Show equipment choices
	if len(choices.EquipmentChoices) > 0 {
		choiceDescriptions := []string{}
		for i, choice := range choices.EquipmentChoices {
			choiceDescriptions = append(choiceDescriptions, fmt.Sprintf("%d. **%s** - Choose %d", i+1, choice.Name, choice.Choose))
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "⚔️ Equipment Choices",
			Value:  strings.Join(choiceDescriptions, "\n"),
			Inline: false,
		})
	} else {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "⚔️ Equipment",
			Value:  "No equipment choices available - you'll receive standard equipment for your class.",
			Inline: false,
		})
	}

	// Progress
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "Progress",
		Value:  "✅ Step 1: Race\n✅ Step 2: Class\n✅ Step 3: Abilities\n✅ Step 4: Proficiencies\n⏳ Step 5: Equipment\n⏳ Step 6: Details",
		Inline: false,
	})

	embed.Footer = &discordgo.MessageEmbedFooter{
		Text: "Click 'Select Equipment' to choose your starting gear",
	}

	// Create components
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Select Equipment",
					Style:    discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("character_create:select_equipment:%s:%s", req.RaceKey, req.ClassKey),
					Emoji: &discordgo.ComponentEmoji{
						Name: "⚔️",
					},
					Disabled: len(choices.EquipmentChoices) == 0,
				},
				discordgo.Button{
					Label:    "Skip to Details",
					Style:    discordgo.SecondaryButton,
					CustomID: fmt.Sprintf("character_create:character_details:%s:%s", req.RaceKey, req.ClassKey),
					Emoji: &discordgo.ComponentEmoji{
						Name: "⏭️",
					},
					Disabled: len(choices.EquipmentChoices) > 0, // Only allow skip if no choices
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
func (h *EquipmentChoicesHandler) respondWithError(req *EquipmentChoicesRequest, message string) error {
	content := fmt.Sprintf("❌ %s", message)
	_, err := req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
		Content: &content,
		Embeds:  &[]*discordgo.MessageEmbed{},
	})
	return err
}
