package session

import (
	"context"
	"fmt"
	
	"github.com/KirkDiggler/dnd-bot-discord/internal/services"
	"github.com/bwmarrin/discordgo"
)

type LeaveRequest struct {
	Session     *discordgo.Session
	Interaction *discordgo.InteractionCreate
	SessionID   string // Optional specific session ID
}

type LeaveHandler struct {
	services *services.Provider
}

func NewLeaveHandler(services *services.Provider) *LeaveHandler {
	return &LeaveHandler{
		services: services,
	}
}

func (h *LeaveHandler) Handle(req *LeaveRequest) error {
	// If no specific session ID, leave all active sessions
	if req.SessionID == "" {
		activeSessions, err := h.services.SessionService.ListActiveUserSessions(context.Background(), req.Interaction.Member.User.ID)
		if err != nil || len(activeSessions) == 0 {
			return req.Session.InteractionRespond(req.Interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "❌ You are not in any active sessions.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		}
		
		// Leave all sessions
		leftCount := 0
		for _, sess := range activeSessions {
			err := h.services.SessionService.LeaveSession(context.Background(), sess.ID, req.Interaction.Member.User.ID)
			if err == nil {
				leftCount++
			}
		}
		
		return req.Session.InteractionRespond(req.Interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("✅ Left %d session(s).", leftCount),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}
	
	// Leave specific session
	err := h.services.SessionService.LeaveSession(context.Background(), req.SessionID, req.Interaction.Member.User.ID)
	if err != nil {
		return req.Session.InteractionRespond(req.Interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("❌ Failed to leave session: %v", err),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}
	
	return req.Session.InteractionRespond(req.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "✅ Successfully left the session.",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}