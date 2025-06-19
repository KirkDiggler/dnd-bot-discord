package character

import (
	"fmt"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services"
	"github.com/bwmarrin/discordgo"
)

type ListRequest struct {
	Session     *discordgo.Session
	Interaction *discordgo.InteractionCreate
}

type ListHandler struct {
	services *services.Provider
}

func NewListHandler(serviceProvider *services.Provider) *ListHandler {
	return &ListHandler{
		services: serviceProvider,
	}
}

func (h *ListHandler) Handle(req *ListRequest) error {
	// Defer acknowledge the interaction
	err := req.Session.InteractionRespond(req.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		return fmt.Errorf("failed to acknowledge interaction: %w", err)
	}

	// Get user's characters
	characters, err := h.services.CharacterService.ListByOwner(req.Interaction.Member.User.ID)
	if err != nil {
		content := fmt.Sprintf("❌ Failed to retrieve your characters: %v", err)
		_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return err
	}

	// Build response
	if len(characters) == 0 {
		content := "📝 You don't have any characters yet. Use `/dnd character create` to create one!"
		_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return err
	}

	// Create embed with character list
	embed := &discordgo.MessageEmbed{
		Title:       "📚 Your Characters",
		Description: fmt.Sprintf("You have %d character(s):", len(characters)),
		Color:       0x3498db, // Blue color
		Fields:      make([]*discordgo.MessageEmbedField, 0, len(characters)),
	}

	// Group characters by status
	activeChars := make([]*entities.Character, 0)
	draftChars := make([]*entities.Character, 0)
	archivedChars := make([]*entities.Character, 0)

	for _, char := range characters {
		switch char.Status {
		case entities.CharacterStatusActive:
			activeChars = append(activeChars, char)
		case entities.CharacterStatusDraft:
			// Only show drafts that have meaningful progress (name or race/class selected)
			if char.Name != "" || char.Race != nil || char.Class != nil {
				draftChars = append(draftChars, char)
			}
		case entities.CharacterStatusArchived:
			archivedChars = append(archivedChars, char)
		}
	}

	// Add active characters
	if len(activeChars) > 0 {
		var sb strings.Builder
		for _, char := range activeChars {
			// Debug logging
			fmt.Printf("Character %s (ID: %s) - Race: %v, Class: %v\n",
				char.Name, char.ID, char.Race != nil, char.Class != nil)
			sb.WriteString(fmt.Sprintf("**%s** - %s (Level %d)\n",
				char.Name,
				char.GetDisplayInfo(),
				char.Level,
			))
			sb.WriteString(fmt.Sprintf("  HP: %d/%d | AC: %d | ID: `%s`\n\n",
				char.CurrentHitPoints,
				char.MaxHitPoints,
				char.AC,
				char.ID,
			))
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "✅ Active Characters",
			Value:  sb.String(),
			Inline: false,
		})
	}

	// Add draft characters
	if len(draftChars) > 0 {
		var sb strings.Builder
		for _, char := range draftChars {
			status := "Creating..."
			if char.Name != "" {
				status = char.Name
			} else if char.Race != nil && char.Class != nil {
				status = fmt.Sprintf("%s %s (unnamed)", char.Race.Name, char.Class.Name)
			} else if char.Race != nil {
				status = fmt.Sprintf("%s (selecting class)", char.Race.Name)
			}

			// Add progress indicator
			progress := ""
			if char.Race != nil {
				progress += "✓ Race "
			}
			if char.Class != nil {
				progress += "✓ Class "
			}
			if len(char.Attributes) > 0 {
				progress += "✓ Abilities "
			}
			if char.Name != "" {
				progress += "✓ Name"
			}

			sb.WriteString(fmt.Sprintf("**%s**\n", status))
			if progress != "" {
				sb.WriteString(fmt.Sprintf("  Progress: %s\n", progress))
			}
			sb.WriteString(fmt.Sprintf("  ID: `%s`\n\n", char.ID))
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "📝 Draft Characters",
			Value:  sb.String(),
			Inline: false,
		})
	}

	// Add archived characters
	if len(archivedChars) > 0 {
		var sb strings.Builder
		for _, char := range archivedChars {
			sb.WriteString(fmt.Sprintf("**%s** - %s | ID: `%s`\n",
				char.Name,
				char.GetDisplayInfo(),
				char.ID,
			))
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "🗄️ Archived Characters",
			Value:  sb.String(),
			Inline: false,
		})
	}

	// Add footer with helpful commands
	embed.Footer = &discordgo.MessageEmbedFooter{
		Text: "Use /dnd character show <id> to view details | /dnd character select <id> to set as active",
	}

	// Add components for quick actions
	components := []discordgo.MessageComponent{}

	// Add show/edit buttons for active characters
	if len(activeChars) > 0 {
		var buttons []discordgo.MessageComponent
		for i, char := range activeChars {
			// Limit to 5 characters per row
			if i >= 5 {
				break
			}
			buttons = append(buttons, discordgo.Button{
				Label:    char.Name,
				Style:    discordgo.SecondaryButton,
				CustomID: fmt.Sprintf("character:quickshow:%s", char.ID),
				Emoji: &discordgo.ComponentEmoji{
					Name: "👁️",
				},
			})
		}
		if len(buttons) > 0 {
			components = append(components, discordgo.ActionsRow{
				Components: buttons,
			})
		}

		// Add second row for edit buttons if we have any characters
		var editButtons []discordgo.MessageComponent
		for i, char := range activeChars {
			if i >= 5 {
				break
			}
			editButtons = append(editButtons, discordgo.Button{
				Label:    "Sheet " + char.Name,
				Style:    discordgo.PrimaryButton,
				CustomID: fmt.Sprintf("character:sheet_show:%s", char.ID),
				Emoji: &discordgo.ComponentEmoji{
					Name: "📋",
				},
			})
		}
		if len(editButtons) > 0 {
			components = append(components, discordgo.ActionsRow{
				Components: editButtons,
			})
		}
	}

	// Send the embed with components
	_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
	})
	return err
}
