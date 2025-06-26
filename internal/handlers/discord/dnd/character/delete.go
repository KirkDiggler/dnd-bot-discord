package character

import (
	"fmt"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services"
	"github.com/bwmarrin/discordgo"
)

type DeleteRequest struct {
	Session     *discordgo.Session
	Interaction *discordgo.InteractionCreate
	CharacterID string // Optional - if provided, skip to confirmation
}

type DeleteHandler struct {
	services *services.Provider
}

func NewDeleteHandler(serviceProvider *services.Provider) *DeleteHandler {
	return &DeleteHandler{
		services: serviceProvider,
	}
}

func (h *DeleteHandler) Handle(req *DeleteRequest) error {
	// If a character ID is provided, show confirmation directly
	if req.CharacterID != "" {
		return h.showDeleteConfirmation(req, req.CharacterID)
	}

	// Otherwise, show character selection
	return h.showCharacterSelection(req)
}

func (h *DeleteHandler) showCharacterSelection(req *DeleteRequest) error {
	// Defer acknowledge the interaction with ephemeral flag
	err := req.Session.InteractionRespond(req.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to acknowledge interaction: %w", err)
	}

	// Get user's characters
	characters, err := h.services.CharacterService.ListByOwner(req.Interaction.Member.User.ID)
	if err != nil {
		content := fmt.Sprintf("Failed to retrieve your characters: %v", err)
		_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return err
	}

	// Check if user has any characters
	if len(characters) == 0 {
		content := "You don't have any characters to delete."
		_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return err
	}

	// Create embed
	embed := &discordgo.MessageEmbed{
		Title:       "Delete Character",
		Description: "Select a character to delete. This action cannot be undone!",
		Color:       0xe74c3c, // Red color for danger
	}

	// Build components based on number of characters
	var components []discordgo.MessageComponent

	if len(characters) <= 25 {
		// Use buttons for <= 25 characters
		components = h.buildButtonComponents(characters)
	} else {
		// Use select menu for > 25 characters
		components = h.buildSelectMenuComponents(characters)
	}

	// Add cancel button
	cancelRow := discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				Label:    "Cancel",
				Style:    discordgo.SecondaryButton,
				CustomID: "character:delete_cancel",
				Emoji: &discordgo.ComponentEmoji{
					Name: "❌",
				},
			},
		},
	}
	components = append(components, cancelRow)

	// Send the response
	_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
	})
	return err
}

func (h *DeleteHandler) buildButtonComponents(characters []*entities.Character) []discordgo.MessageComponent {
	var components []discordgo.MessageComponent
	var currentRow []discordgo.MessageComponent

	for i, char := range characters {
		// Create button label with name and status
		label := char.Name
		if char.Name == "" || char.Name == "Draft Character" {
			if char.Race != nil && char.Class != nil {
				label = fmt.Sprintf("%s %s", char.Race.Name, char.Class.Name)
			} else {
				label = "Unnamed Character"
			}
		}

		// Add status indicator
		statusEmoji := h.getStatusEmoji(char.Status)

		button := discordgo.Button{
			Label:    fmt.Sprintf("%s %s", statusEmoji, label),
			Style:    discordgo.DangerButton,
			CustomID: fmt.Sprintf("character:delete_select:%s", char.ID),
		}

		currentRow = append(currentRow, button)

		// Discord allows max 5 buttons per row
		if len(currentRow) == 5 || i == len(characters)-1 {
			components = append(components, discordgo.ActionsRow{
				Components: currentRow,
			})
			currentRow = []discordgo.MessageComponent{}
		}

		// Discord allows max 5 rows of buttons
		if len(components) >= 5 {
			break
		}
	}

	return components
}

func (h *DeleteHandler) buildSelectMenuComponents(characters []*entities.Character) []discordgo.MessageComponent {
	var options []discordgo.SelectMenuOption

	for _, char := range characters {
		// Create option label with name and status
		label := char.Name
		if char.Name == "" || char.Name == "Draft Character" {
			if char.Race != nil && char.Class != nil {
				label = fmt.Sprintf("%s %s", char.Race.Name, char.Class.Name)
			} else {
				label = "Unnamed Character"
			}
		}

		// Add description with level and status
		description := fmt.Sprintf("Level %d %s", char.Level, char.Status)
		if char.Race != nil && char.Class != nil {
			description = fmt.Sprintf("%s %s - %s", char.Race.Name, char.Class.Name, char.Status)
		}

		option := discordgo.SelectMenuOption{
			Label:       label,
			Value:       char.ID,
			Description: description,
			Emoji: &discordgo.ComponentEmoji{
				Name: h.getStatusEmoji(char.Status),
			},
		}
		options = append(options, option)

		// Discord allows max 25 options in a select menu
		if len(options) >= 25 {
			break
		}
	}

	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    "character:delete_select_menu",
					Placeholder: "Select a character to delete",
					Options:     options,
				},
			},
		},
	}
}

func (h *DeleteHandler) showDeleteConfirmation(req *DeleteRequest, characterID string) error {
	// Update the message to show confirmation
	err := req.Session.InteractionRespond(req.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content: "Loading character details...",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to acknowledge interaction: %w", err)
	}

	// Get the character
	character, err := h.services.CharacterService.GetByID(characterID)
	if err != nil {
		content := fmt.Sprintf("Character not found with ID: %s", characterID)
		_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return err
	}

	// Verify ownership
	if character.OwnerID != req.Interaction.Member.User.ID {
		content := "You can only delete your own characters!"
		_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return err
	}

	// Build confirmation embed
	embed := &discordgo.MessageEmbed{
		Title:       "⚠️ Confirm Character Deletion",
		Description: "Are you sure you want to delete this character? This action cannot be undone!",
		Color:       0xe74c3c, // Red color for danger
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Character",
				Value:  character.NameString(),
				Inline: true,
			},
			{
				Name:   "Status",
				Value:  string(character.Status),
				Inline: true,
			},
		},
	}

	// Add more details if available
	if character.Race != nil && character.Class != nil {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Details",
			Value:  fmt.Sprintf("%s %s (Level %d)", character.Race.Name, character.Class.Name, character.Level),
			Inline: false,
		})
	}

	// Add confirmation buttons
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Delete Forever",
					Style:    discordgo.DangerButton,
					CustomID: fmt.Sprintf("character:delete_confirm:%s", character.ID),
					Emoji: &discordgo.ComponentEmoji{
						Name: "🗑️",
					},
				},
				discordgo.Button{
					Label:    "Cancel",
					Style:    discordgo.SecondaryButton,
					CustomID: "character:delete_cancel",
					Emoji: &discordgo.ComponentEmoji{
						Name: "❌",
					},
				},
			},
		},
	}

	// Send the confirmation
	_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
		Content:    &[]string{""}[0], // Clear content
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
	})
	return err
}

// HandleDeleteConfirm handles the actual deletion after confirmation
func (h *DeleteHandler) HandleDeleteConfirm(req *DeleteRequest) error {
	// Update message to show deletion in progress
	err := req.Session.InteractionRespond(req.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content:    "Deleting character...",
			Embeds:     []*discordgo.MessageEmbed{},
			Components: []discordgo.MessageComponent{},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to acknowledge interaction: %w", err)
	}

	// Extract character ID from the custom ID
	parts := strings.Split(req.Interaction.MessageComponentData().CustomID, ":")
	if len(parts) < 3 {
		content := "Invalid character ID format"
		_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return err
	}
	characterID := parts[2]

	// Verify ownership one more time
	character, err := h.services.CharacterService.GetByID(characterID)
	if err != nil {
		content := fmt.Sprintf("Character not found with ID: %s", characterID)
		_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return err
	}

	if character.OwnerID != req.Interaction.Member.User.ID {
		content := "You can only delete your own characters!"
		_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return err
	}

	// Delete the character
	err = h.services.CharacterService.Delete(characterID)
	if err != nil {
		content := fmt.Sprintf("Failed to delete character: %v", err)
		_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return err
	}

	// Success message
	embed := &discordgo.MessageEmbed{
		Title:       "Character Deleted",
		Description: fmt.Sprintf("**%s** has been permanently deleted.", character.NameString()),
		Color:       0x2ecc71, // Green color for success
		Footer: &discordgo.MessageEmbedFooter{
			Text: "This action cannot be undone",
		},
	}

	_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
		Content:    &[]string{""}[0], // Clear content
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &[]discordgo.MessageComponent{},
	})
	return err
}

// HandleDeleteCancel handles cancellation of deletion
func (h *DeleteHandler) HandleDeleteCancel(req *DeleteRequest) error {
	// Update message to show cancellation
	err := req.Session.InteractionRespond(req.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content:    "Character deletion cancelled.",
			Embeds:     []*discordgo.MessageEmbed{},
			Components: []discordgo.MessageComponent{},
		},
	})
	return err
}

// HandleSelectMenu handles character selection from dropdown
func (h *DeleteHandler) HandleSelectMenu(req *DeleteRequest) error {
	// Get selected character ID from the interaction data
	if len(req.Interaction.MessageComponentData().Values) == 0 {
		return fmt.Errorf("no character selected")
	}

	characterID := req.Interaction.MessageComponentData().Values[0]
	return h.showDeleteConfirmation(req, characterID)
}

func (h *DeleteHandler) getStatusEmoji(status entities.CharacterStatus) string {
	switch status {
	case entities.CharacterStatusActive:
		return "✅"
	case entities.CharacterStatusDraft:
		return "📝"
	case entities.CharacterStatusArchived:
		return "🗄️"
	default:
		return "❓"
	}
}
