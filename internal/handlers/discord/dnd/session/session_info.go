package session

import (
	"context"
	"fmt"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/session"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/services"
	"github.com/bwmarrin/discordgo"
)

type InfoRequest struct {
	Session     *discordgo.Session
	Interaction *discordgo.InteractionCreate
}

type InfoHandler struct {
	services *services.Provider
}

func NewInfoHandler(serviceProvider *services.Provider) *InfoHandler {
	return &InfoHandler{
		services: serviceProvider,
	}
}

func (h *InfoHandler) Handle(req *InfoRequest) error {
	// Defer acknowledge the interaction
	err := req.Session.InteractionRespond(req.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		return fmt.Errorf("failed to acknowledge interaction: %w", err)
	}

	// Get active sessions for the user
	sessions, err := h.services.SessionService.ListActiveUserSessions(
		context.Background(),
		req.Interaction.Member.User.ID,
	)
	if err != nil {
		content := fmt.Sprintf("❌ Failed to retrieve your active sessions: %v", err)
		_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return err
	}

	if len(sessions) == 0 {
		content := "📝 You're not part of any active sessions. Use `/dnd session create` to start one or `/dnd session join` with an invite code!"
		_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return err
	}

	// If user has only one active session, show that one
	var session *session.Session
	if len(sessions) == 1 {
		session = sessions[0]
	} else {
		// Find the most recently active session
		session = sessions[0]
		for _, s := range sessions[1:] {
			if s.LastActive.After(session.LastActive) {
				session = s
			}
		}
	}

	// Create detailed embed
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("🎲 %s", session.Name),
		Description: session.Description,
		Color:       h.getColorForStatus(session.Status),
		Fields:      []*discordgo.MessageEmbedField{},
	}

	// Add status and DM info
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "📋 Status",
		Value:  h.getStatusDisplay(session.Status),
		Inline: true,
	}, &discordgo.MessageEmbedField{
		Name:   "🎲 Dungeon Master",
		Value:  fmt.Sprintf("<@%s>", session.DMID),
		Inline: true,
	}, &discordgo.MessageEmbedField{
		Name:   "👥 Players",
		Value:  fmt.Sprintf("%d/%d", len(session.GetActivePlayers()), session.Settings.MaxPlayers),
		Inline: true,
	})

	// Add invite code if user is DM and session is planning
	member, exists := session.Members[req.Interaction.Member.User.ID]
	if exists && member.Role == session.SessionRoleDM && session.Status == session.SessionStatusPlanning {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "🔑 Invite Code",
			Value:  fmt.Sprintf("```%s```", session.InviteCode),
			Inline: false,
		})
	}

	// Add player list
	if players := session.GetActivePlayers(); len(players) > 0 {
		var playerList strings.Builder
		for i, player := range players {
			if i > 0 {
				playerList.WriteString("\n")
			}
			playerList.WriteString(fmt.Sprintf("<@%s>", player.UserID))
			if player.CharacterID != "" {
				// Get character name if possible
				if char, getCharErr := h.services.CharacterService.GetByID(player.CharacterID); getCharErr == nil {
					playerList.WriteString(fmt.Sprintf(" - %s", char.Name))
				} else {
					playerList.WriteString(" - Character Selected")
				}
			} else {
				playerList.WriteString(" - *No character*")
			}
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "🎭 Party Members",
			Value:  playerList.String(),
			Inline: false,
		})
	}

	// Add session timing info
	if session.Status == session.SessionStatusActive && session.StartedAt != nil {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "⏱️ Started",
			Value:  fmt.Sprintf("<t:%d:R>", session.StartedAt.Unix()),
			Inline: true,
		})
	}

	// Add settings info
	settingsInfo := []string{}
	if session.Settings.AllowSpectators {
		settingsInfo = append(settingsInfo, "👁️ Spectators allowed")
	}
	if session.Settings.RequireInvite {
		settingsInfo = append(settingsInfo, "🔑 Invite required")
	}
	if session.Settings.AllowLateJoin {
		settingsInfo = append(settingsInfo, "🚪 Late join allowed")
	}
	if session.Settings.AutoEndAfterHours > 0 {
		settingsInfo = append(settingsInfo, fmt.Sprintf("⏰ Auto-end after %d hours", session.Settings.AutoEndAfterHours))
	}

	if len(settingsInfo) > 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "⚙️ Settings",
			Value:  strings.Join(settingsInfo, "\n"),
			Inline: false,
		})
	}

	// Add footer
	embed.Footer = &discordgo.MessageEmbedFooter{
		Text: fmt.Sprintf("Session ID: %s | Last active: %s", session.ID, session.LastActive.Format("Jan 2, 3:04 PM")),
	}

	// Add action buttons based on user role and session status
	components := h.getActionButtons(session, member)

	// Send the response
	_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
	})
	return err
}

func (h *InfoHandler) getColorForStatus(status session.SessionStatus) int {
	switch status {
	case session.SessionStatusActive:
		return 0x2ecc71 // Green
	case session.SessionStatusPlanning:
		return 0x3498db // Blue
	case session.SessionStatusPaused:
		return 0xf39c12 // Orange
	case session.SessionStatusEnded:
		return 0x95a5a6 // Gray
	default:
		return 0x7289da // Discord blue
	}
}

func (h *InfoHandler) getStatusDisplay(status session.SessionStatus) string {
	switch status {
	case session.SessionStatusActive:
		return "🟢 Active"
	case session.SessionStatusPlanning:
		return "📋 Planning"
	case session.SessionStatusPaused:
		return "⏸️ Paused"
	case session.SessionStatusEnded:
		return "🏁 Ended"
	default:
		return string(status)
	}
}

func (h *InfoHandler) getActionButtons(session *session.Session, member *session.SessionMember) []discordgo.MessageComponent {
	if member == nil {
		return nil
	}

	components := []discordgo.MessageComponent{}
	buttons := []discordgo.MessageComponent{}

	// DM actions
	if member.Role == session.SessionRoleDM {
		switch session.Status {
		case session.SessionStatusPlanning:
			buttons = append(buttons, discordgo.Button{
				Label:    "Start Session",
				Style:    discordgo.SuccessButton,
				CustomID: fmt.Sprintf("session_manage:start:%s", session.ID),
				Emoji:    &discordgo.ComponentEmoji{Name: "▶️"},
			}, discordgo.Button{
				Label:    "Invite Players",
				Style:    discordgo.PrimaryButton,
				CustomID: fmt.Sprintf("session_manage:invite:%s", session.ID),
				Emoji:    &discordgo.ComponentEmoji{Name: "📨"},
			})
		case session.SessionStatusActive:
			buttons = append(buttons, discordgo.Button{
				Label:    "Pause Session",
				Style:    discordgo.SecondaryButton,
				CustomID: fmt.Sprintf("session_manage:pause:%s", session.ID),
				Emoji:    &discordgo.ComponentEmoji{Name: "⏸️"},
			}, discordgo.Button{
				Label:    "End Session",
				Style:    discordgo.DangerButton,
				CustomID: fmt.Sprintf("session_manage:end:%s", session.ID),
				Emoji:    &discordgo.ComponentEmoji{Name: "🏁"},
			})
		case session.SessionStatusPaused:
			buttons = append(buttons, discordgo.Button{
				Label:    "Resume Session",
				Style:    discordgo.SuccessButton,
				CustomID: fmt.Sprintf("session_manage:resume:%s", session.ID),
				Emoji:    &discordgo.ComponentEmoji{Name: "▶️"},
			}, discordgo.Button{
				Label:    "End Session",
				Style:    discordgo.DangerButton,
				CustomID: fmt.Sprintf("session_manage:end:%s", session.ID),
				Emoji:    &discordgo.ComponentEmoji{Name: "🏁"},
			})
		}

		// Session settings button (always available for DM)
		if session.Status != session.SessionStatusEnded {
			buttons = append(buttons, discordgo.Button{
				Label:    "Settings",
				Style:    discordgo.SecondaryButton,
				CustomID: fmt.Sprintf("session_manage:settings:%s", session.ID),
				Emoji:    &discordgo.ComponentEmoji{Name: "⚙️"},
			})
		}
	}

	// Player actions
	if member.Role == session.SessionRolePlayer && session.Status != session.SessionStatusEnded {
		if member.CharacterID == "" {
			buttons = append(buttons, discordgo.Button{
				Label:    "Select Character",
				Style:    discordgo.PrimaryButton,
				CustomID: fmt.Sprintf("session_manage:select_character:%s", session.ID),
				Emoji:    &discordgo.ComponentEmoji{Name: "🎭"},
			})
		}

		buttons = append(buttons, discordgo.Button{
			Label:    "Leave Session",
			Style:    discordgo.DangerButton,
			CustomID: fmt.Sprintf("session_manage:leave:%s", session.ID),
			Emoji:    &discordgo.ComponentEmoji{Name: "🚪"},
		})
	}

	if len(buttons) > 0 {
		components = append(components, discordgo.ActionsRow{
			Components: buttons,
		})
	}

	return components
}
