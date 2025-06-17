package character

import (
	"fmt"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services"
	"github.com/bwmarrin/discordgo"
)

type ShowRequest struct {
	Session     *discordgo.Session
	Interaction *discordgo.InteractionCreate
	CharacterID string
}

type ShowHandler struct {
	services *services.Provider
}

func NewShowHandler(services *services.Provider) *ShowHandler {
	return &ShowHandler{
		services: services,
	}
}

func (h *ShowHandler) Handle(req *ShowRequest) error {
	// Defer acknowledge the interaction
	err := req.Session.InteractionRespond(req.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		return fmt.Errorf("failed to acknowledge interaction: %w", err)
	}

	// Get the character
	character, err := h.services.CharacterService.GetByID(req.CharacterID)
	if err != nil {
		content := fmt.Sprintf("❌ Character not found with ID: %s", req.CharacterID)
		_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return err
	}

	// Verify ownership
	if character.OwnerID != req.Interaction.Member.User.ID {
		content := "❌ You can only view your own characters!"
		_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return err
	}

	// Create detailed embed
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("🎭 %s", character.NameString()),
		Description: fmt.Sprintf("**Status:** %s", character.Status),
		Color:       getColorForStatus(character.Status),
		Fields:      make([]*discordgo.MessageEmbedField, 0),
	}

	// Basic info
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name: "📋 Basic Info",
		Value: fmt.Sprintf("**Level:** %d\n**Experience:** %d XP\n**Speed:** %d ft",
			character.Level,
			character.Experience,
			character.Speed,
		),
		Inline: true,
	})

	// Combat stats
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name: "⚔️ Combat Stats",
		Value: fmt.Sprintf("**HP:** %d/%d\n**AC:** %d\n**Hit Die:** d%d",
			character.CurrentHitPoints,
			character.MaxHitPoints,
			character.AC,
			character.HitDie,
		),
		Inline: true,
	})

	// Attributes
	if len(character.Attributes) > 0 {
		attrValue := ""
		for _, attr := range entities.Attributes {
			if character.Attributes[attr] != nil {
				attrValue += fmt.Sprintf("**%s:** %d (%+d)\n",
					attr.Short(),
					character.Attributes[attr].Score,
					character.Attributes[attr].Bonus,
				)
			}
		}
		if attrValue != "" {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "💪 Attributes",
				Value:  attrValue,
				Inline: true,
			})
		}
	}

	// Proficiencies
	if len(character.Proficiencies) > 0 {
		profValue := ""
		for profType, profs := range character.Proficiencies {
			if len(profs) > 0 {
				profValue += fmt.Sprintf("**%s:**\n", profType)
				for _, prof := range profs {
					profValue += fmt.Sprintf("• %s\n", prof.Name)
				}
			}
		}
		if profValue != "" {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "🎯 Proficiencies",
				Value:  profValue,
				Inline: false,
			})
		}
	}

	// Equipment
	if len(character.EquippedSlots) > 0 {
		equipValue := ""
		for slot, item := range character.EquippedSlots {
			if item != nil {
				equipValue += fmt.Sprintf("**%s:** %s\n", slot, item.GetName())
			}
		}
		if equipValue != "" {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "🛡️ Equipped Items",
				Value:  equipValue,
				Inline: false,
			})
		}
	}

	// Inventory summary
	if len(character.Inventory) > 0 {
		invCount := 0
		for _, items := range character.Inventory {
			invCount += len(items)
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "🎒 Inventory",
			Value:  fmt.Sprintf("Total items: %d", invCount),
			Inline: true,
		})
	}

	// Footer with ID
	embed.Footer = &discordgo.MessageEmbedFooter{
		Text: fmt.Sprintf("Character ID: %s", character.ID),
	}

	// Add action buttons based on status
	components := []discordgo.MessageComponent{}
	if character.Status == entities.CharacterStatusActive {
		components = append(components, discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Edit",
					Style:    discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("character_manage:edit:%s", character.ID),
					Emoji: &discordgo.ComponentEmoji{
						Name: "✏️",
					},
				},
				discordgo.Button{
					Label:    "Archive",
					Style:    discordgo.SecondaryButton,
					CustomID: fmt.Sprintf("character_manage:archive:%s", character.ID),
					Emoji: &discordgo.ComponentEmoji{
						Name: "🗄️",
					},
				},
				discordgo.Button{
					Label:    "Delete",
					Style:    discordgo.DangerButton,
					CustomID: fmt.Sprintf("character_manage:delete:%s", character.ID),
					Emoji: &discordgo.ComponentEmoji{
						Name: "🗑️",
					},
				},
			},
		})
	} else if character.Status == entities.CharacterStatusDraft {
		components = append(components, discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Continue Creating",
					Style:    discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("character_manage:continue:%s", character.ID),
					Emoji: &discordgo.ComponentEmoji{
						Name: "▶️",
					},
				},
				discordgo.Button{
					Label:    "Delete Draft",
					Style:    discordgo.DangerButton,
					CustomID: fmt.Sprintf("character_manage:delete:%s", character.ID),
					Emoji: &discordgo.ComponentEmoji{
						Name: "🗑️",
					},
				},
			},
		})
	} else if character.Status == entities.CharacterStatusArchived {
		components = append(components, discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Restore",
					Style:    discordgo.SuccessButton,
					CustomID: fmt.Sprintf("character_manage:restore:%s", character.ID),
					Emoji: &discordgo.ComponentEmoji{
						Name: "♻️",
					},
				},
				discordgo.Button{
					Label:    "Delete Permanently",
					Style:    discordgo.DangerButton,
					CustomID: fmt.Sprintf("character_manage:delete:%s", character.ID),
					Emoji: &discordgo.ComponentEmoji{
						Name: "🗑️",
					},
				},
			},
		})
	}

	// Send the embed with components
	_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
	})
	return err
}

func getColorForStatus(status entities.CharacterStatus) int {
	switch status {
	case entities.CharacterStatusActive:
		return 0x2ecc71 // Green
	case entities.CharacterStatusDraft:
		return 0xf39c12 // Orange
	case entities.CharacterStatusArchived:
		return 0x95a5a6 // Gray
	default:
		return 0x3498db // Blue
	}
}