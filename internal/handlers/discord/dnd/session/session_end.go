package session

import (
	"context"
	"fmt"

	"github.com/KirkDiggler/dnd-bot-discord/internal/services"
	"github.com/bwmarrin/discordgo"
)

type EndRequest struct {
	Session     *discordgo.Session
	Interaction *discordgo.InteractionCreate
	SessionID   string
}

type EndHandler struct {
	services *services.Provider
}

func NewEndHandler(services *services.Provider) *EndHandler {
	return &EndHandler{
		services: services,
	}
}

func (h *EndHandler) Handle(req *EndRequest) error {
	// Defer acknowledge the interaction
	err := req.Session.InteractionRespond(req.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		return fmt.Errorf("failed to acknowledge interaction: %w", err)
	}

	// Get session details before ending
	session, err := h.services.SessionService.GetSession(
		context.Background(),
		req.SessionID,
	)
	if err != nil {
		content := fmt.Sprintf("❌ Failed to get session: %v", err)
		_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return err
	}

	// End the session
	err = h.services.SessionService.EndSession(
		context.Background(),
		req.SessionID,
		req.Interaction.Member.User.ID,
	)
	if err != nil {
		content := fmt.Sprintf("❌ Failed to end session: %v", err)
		_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return err
	}

	// Create summary embed
	embed := &discordgo.MessageEmbed{
		Title:       "🏁 Session Ended",
		Description: fmt.Sprintf("**%s** has been concluded.", session.Name),
		Color:       0xe74c3c, // Red
		Fields:      []*discordgo.MessageEmbedField{},
	}

	// Add session duration if it was started
	if session.StartedAt != nil {
		duration := session.EndedAt.Sub(*session.StartedAt)
		hours := int(duration.Hours())
		minutes := int(duration.Minutes()) % 60

		durationStr := ""
		if hours > 0 {
			durationStr = fmt.Sprintf("%d hours, %d minutes", hours, minutes)
		} else {
			durationStr = fmt.Sprintf("%d minutes", minutes)
		}

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "⏱️ Duration",
			Value:  durationStr,
			Inline: true,
		})
	}

	// Add player count
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "👥 Players",
		Value:  fmt.Sprintf("%d players participated", len(session.GetActivePlayers())),
		Inline: true,
	})

	// Add DM info
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "🎲 Dungeon Master",
		Value:  fmt.Sprintf("<@%s>", session.DMID),
		Inline: true,
	})

	// Add footer
	embed.Footer = &discordgo.MessageEmbedFooter{
		Text: fmt.Sprintf("Session ID: %s", session.ID),
	}

	// Send the response
	_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	})
	return err
}
