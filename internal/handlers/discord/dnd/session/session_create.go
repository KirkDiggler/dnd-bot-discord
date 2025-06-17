package session

import (
	"context"
	"fmt"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/services"
	sessionService "github.com/KirkDiggler/dnd-bot-discord/internal/services/session"
	"github.com/bwmarrin/discordgo"
)

type CreateRequest struct {
	Session     *discordgo.Session
	Interaction *discordgo.InteractionCreate
	Name        string
	Description string
}

type CreateHandler struct {
	services *services.Provider
}

func NewCreateHandler(services *services.Provider) *CreateHandler {
	return &CreateHandler{
		services: services,
	}
}

func (h *CreateHandler) Handle(req *CreateRequest) error {
	// Defer acknowledge the interaction
	err := req.Session.InteractionRespond(req.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		return fmt.Errorf("failed to acknowledge interaction: %w", err)
	}

	// Validate session name
	if strings.TrimSpace(req.Name) == "" {
		content := "❌ Session name is required!"
		_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return err
	}

	// Create the session
	session, err := h.services.SessionService.CreateSession(context.Background(), &sessionService.CreateSessionInput{
		Name:        req.Name,
		Description: req.Description,
		RealmID:     req.Interaction.GuildID,
		ChannelID:   req.Interaction.ChannelID,
		CreatorID:   req.Interaction.Member.User.ID,
	})
	if err != nil {
		content := fmt.Sprintf("❌ Failed to create session: %v", err)
		_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return err
	}

	// Create success embed
	embed := &discordgo.MessageEmbed{
		Title:       "🎲 Session Created!",
		Description: fmt.Sprintf("**%s** has been created successfully!", session.Name),
		Color:       0x2ecc71, // Green
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "📋 Details",
				Value:  fmt.Sprintf("**Status:** %s\n**Max Players:** %d\n**DM:** <@%s>", session.Status, session.Settings.MaxPlayers, session.DMID),
				Inline: true,
			},
			{
				Name:   "🔑 Invite Code",
				Value:  fmt.Sprintf("```%s```", session.InviteCode),
				Inline: true,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Session ID: %s", session.ID),
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
					Label:    "Invite Players",
					Style:    discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("session_manage:invite:%s", session.ID),
					Emoji: &discordgo.ComponentEmoji{
						Name: "📨",
					},
				},
				discordgo.Button{
					Label:    "Start Session",
					Style:    discordgo.SuccessButton,
					CustomID: fmt.Sprintf("session_manage:start:%s", session.ID),
					Emoji: &discordgo.ComponentEmoji{
						Name: "▶️",
					},
				},
				discordgo.Button{
					Label:    "Session Settings",
					Style:    discordgo.SecondaryButton,
					CustomID: fmt.Sprintf("session_manage:settings:%s", session.ID),
					Emoji: &discordgo.ComponentEmoji{
						Name: "⚙️",
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
