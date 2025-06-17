package session

import (
	"context"
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

func NewListHandler(services *services.Provider) *ListHandler {
	return &ListHandler{
		services: services,
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

	// Get user's sessions
	sessions, err := h.services.SessionService.ListUserSessions(context.Background(), req.Interaction.Member.User.ID)
	if err != nil {
		content := fmt.Sprintf("❌ Failed to retrieve your sessions: %v", err)
		_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return err
	}

	// Build response
	if len(sessions) == 0 {
		content := "📝 You're not part of any sessions. Use `/dnd session create` to start one or ask for an invite code!"
		_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return err
	}

	// Create embed with session list
	embed := &discordgo.MessageEmbed{
		Title:       "🎲 Your Sessions",
		Description: fmt.Sprintf("You're part of %d session(s):", len(sessions)),
		Color:       0x3498db, // Blue color
		Fields:      make([]*discordgo.MessageEmbedField, 0),
	}

	// Group sessions by status
	activeSessions := make([]*entities.Session, 0)
	planningSessions := make([]*entities.Session, 0)
	pausedSessions := make([]*entities.Session, 0)
	endedSessions := make([]*entities.Session, 0)

	for _, session := range sessions {
		switch session.Status {
		case entities.SessionStatusActive:
			activeSessions = append(activeSessions, session)
		case entities.SessionStatusPlanning:
			planningSessions = append(planningSessions, session)
		case entities.SessionStatusPaused:
			pausedSessions = append(pausedSessions, session)
		case entities.SessionStatusEnded:
			endedSessions = append(endedSessions, session)
		}
	}

	// Add active sessions
	if len(activeSessions) > 0 {
		var sb strings.Builder
		for _, session := range activeSessions {
			role := h.getUserRole(session, req.Interaction.Member.User.ID)
			sb.WriteString(fmt.Sprintf("**%s** - %s\n", session.Name, role))
			sb.WriteString(fmt.Sprintf("  DM: <@%s> | Players: %d/%d\n", session.DMID, len(session.GetActivePlayers()), session.Settings.MaxPlayers))
			sb.WriteString(fmt.Sprintf("  ID: `%s`\n\n", session.ID))
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "🟢 Active Sessions",
			Value:  sb.String(),
			Inline: false,
		})
	}

	// Add planning sessions
	if len(planningSessions) > 0 {
		var sb strings.Builder
		for _, session := range planningSessions {
			role := h.getUserRole(session, req.Interaction.Member.User.ID)
			sb.WriteString(fmt.Sprintf("**%s** - %s\n", session.Name, role))
			sb.WriteString(fmt.Sprintf("  DM: <@%s> | Players: %d/%d\n", session.DMID, len(session.GetActivePlayers()), session.Settings.MaxPlayers))
			if role == "DM" {
				sb.WriteString(fmt.Sprintf("  Invite Code: `%s`\n", session.InviteCode))
			}
			sb.WriteString(fmt.Sprintf("  ID: `%s`\n\n", session.ID))
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "📋 Planning Sessions",
			Value:  sb.String(),
			Inline: false,
		})
	}

	// Add paused sessions
	if len(pausedSessions) > 0 {
		var sb strings.Builder
		for _, session := range pausedSessions {
			role := h.getUserRole(session, req.Interaction.Member.User.ID)
			sb.WriteString(fmt.Sprintf("**%s** - %s | ID: `%s`\n", session.Name, role, session.ID))
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "⏸️ Paused Sessions",
			Value:  sb.String(),
			Inline: false,
		})
	}

	// Add ended sessions (limit to recent 5)
	if len(endedSessions) > 0 {
		var sb strings.Builder
		limit := 5
		if len(endedSessions) < limit {
			limit = len(endedSessions)
		}
		for i := 0; i < limit; i++ {
			session := endedSessions[i]
			sb.WriteString(fmt.Sprintf("**%s** - Ended\n", session.Name))
		}
		if len(endedSessions) > limit {
			sb.WriteString(fmt.Sprintf("_...and %d more_\n", len(endedSessions)-limit))
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "🏁 Recent Ended Sessions",
			Value:  sb.String(),
			Inline: false,
		})
	}

	// Add footer with helpful commands
	embed.Footer = &discordgo.MessageEmbedFooter{
		Text: "Use /dnd session join <code> to join a session | /dnd session create to start a new one",
	}

	// Send the embed
	_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	})
	return err
}

func (h *ListHandler) getUserRole(session *entities.Session, userID string) string {
	if member, exists := session.Members[userID]; exists {
		switch member.Role {
		case entities.SessionRoleDM:
			return "DM"
		case entities.SessionRolePlayer:
			return "Player"
		case entities.SessionRoleSpectator:
			return "Spectator"
		}
	}
	return "Unknown"
}
