package session

import (
	"context"
	"fmt"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/services"
	"github.com/bwmarrin/discordgo"
)

type JoinRequest struct {
	Session     *discordgo.Session
	Interaction *discordgo.InteractionCreate
	InviteCode  string
}

type JoinHandler struct {
	services *services.Provider
}

func NewJoinHandler(servicesProvider *services.Provider) *JoinHandler {
	return &JoinHandler{
		services: servicesProvider,
	}
}

func (h *JoinHandler) Handle(req *JoinRequest) error {
	// Defer acknowledge the interaction
	err := req.Session.InteractionRespond(req.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		return fmt.Errorf("failed to acknowledge interaction: %w", err)
	}

	// Validate invite code
	if strings.TrimSpace(req.InviteCode) == "" {
		content := "❌ Invite code is required! Ask your DM for the session invite code."
		_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return err
	}

	// Join the session
	member, err := h.services.SessionService.JoinSessionByCode(
		context.Background(),
		strings.ToUpper(strings.TrimSpace(req.InviteCode)),
		req.Interaction.Member.User.ID,
	)
	if err != nil {
		content := fmt.Sprintf("❌ Failed to join session: %v", err)
		_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return err
	}

	// Get the session details
	session, err := h.services.SessionService.GetSessionByInviteCode(
		context.Background(),
		strings.ToUpper(strings.TrimSpace(req.InviteCode)),
	)
	if err != nil {
		content := "✅ Successfully joined the session!"
		_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return err
	}

	// Create success embed
	embed := &discordgo.MessageEmbed{
		Title:       "🎉 Joined Session!",
		Description: fmt.Sprintf("You've successfully joined **%s**!", session.Name),
		Color:       0x2ecc71, // Green
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "📋 Session Info",
				Value:  fmt.Sprintf("**Status:** %s\n**DM:** <@%s>\n**Players:** %d/%d", session.Status, session.DMID, len(session.GetActivePlayers()), session.Settings.MaxPlayers),
				Inline: true,
			},
			{
				Name:   "👤 Your Role",
				Value:  string(member.Role),
				Inline: true,
			},
		},
	}

	if session.Description != "" {
		embed.Fields = append([]*discordgo.MessageEmbedField{
			{
				Name:   "📝 Description",
				Value:  session.Description,
				Inline: false,
			},
		}, embed.Fields...)
	}

	// Add action buttons
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Select Character",
					Style:    discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("session_manage:select_character:%s", session.ID),
					Emoji: &discordgo.ComponentEmoji{
						Name: "🎭",
					},
				},
				discordgo.Button{
					Label:    "Leave Session",
					Style:    discordgo.DangerButton,
					CustomID: fmt.Sprintf("session_manage:leave:%s", session.ID),
					Emoji: &discordgo.ComponentEmoji{
						Name: "🚪",
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
