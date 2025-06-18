package session

import (
	"context"
	"fmt"
	"log"

	"github.com/KirkDiggler/dnd-bot-discord/internal/services"
	"github.com/bwmarrin/discordgo"
)

type StartRequest struct {
	Session     *discordgo.Session
	Interaction *discordgo.InteractionCreate
	SessionID   string
}

type StartHandler struct {
	services *services.Provider
}

func NewStartHandler(servicesProvider *services.Provider) *StartHandler {
	return &StartHandler{
		services: servicesProvider,
	}
}

func (h *StartHandler) Handle(req *StartRequest) error {
	// Defer acknowledge the interaction
	err := req.Session.InteractionRespond(req.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		return fmt.Errorf("failed to acknowledge interaction: %w", err)
	}

	// Start the session
	err = h.services.SessionService.StartSession(
		context.Background(),
		req.SessionID,
		req.Interaction.Member.User.ID,
	)
	if err != nil {
		content := fmt.Sprintf("❌ Failed to start session: %v", err)
		_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return err
	}

	// Get updated session details
	session, err := h.services.SessionService.GetSession(
		context.Background(),
		req.SessionID,
	)
	if err != nil {
		content := "✅ Session started successfully!"
		_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		if err != nil {
			log.Println("Failed to edit interaction response:", err)
		}
		return nil
	}

	// Create success embed
	embed := &discordgo.MessageEmbed{
		Title:       "🎮 Session Started!",
		Description: fmt.Sprintf("**%s** is now active and ready for play!", session.Name),
		Color:       0x2ecc71, // Green
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "📋 Session Info",
				Value:  fmt.Sprintf("**Status:** %s\n**DM:** <@%s>\n**Players:** %d/%d", session.Status, session.DMID, len(session.GetActivePlayers()), session.Settings.MaxPlayers),
				Inline: true,
			},
		},
	}

	// List active players
	if players := session.GetActivePlayers(); len(players) > 0 {
		playerList := ""
		for _, player := range players {
			playerList += fmt.Sprintf("<@%s>", player.UserID)
			if player.CharacterID != "" {
				playerList += " ✅"
			} else {
				playerList += " ⚠️"
			}
			playerList += "\n"
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "👥 Players",
			Value:  playerList,
			Inline: true,
		})
	}

	// Add footer
	embed.Footer = &discordgo.MessageEmbedFooter{
		Text: "⚠️ Players without characters are marked with warning",
	}

	// Add action buttons
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Pause Session",
					Style:    discordgo.SecondaryButton,
					CustomID: fmt.Sprintf("session_manage:pause:%s", session.ID),
					Emoji: &discordgo.ComponentEmoji{
						Name: "⏸️",
					},
				},
				discordgo.Button{
					Label:    "End Session",
					Style:    discordgo.DangerButton,
					CustomID: fmt.Sprintf("session_manage:end:%s", session.ID),
					Emoji: &discordgo.ComponentEmoji{
						Name: "🏁",
					},
				},
			},
		},
	}

	// Send the response
	_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
	})
	return err
}
