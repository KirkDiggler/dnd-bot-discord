package dungeon

import (
	"context"
	"fmt"
	
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services"
	"github.com/bwmarrin/discordgo"
)

type JoinPartyHandler struct {
	services *services.Provider
}

func NewJoinPartyHandler(services *services.Provider) *JoinPartyHandler {
	return &JoinPartyHandler{
		services: services,
	}
}

func (h *JoinPartyHandler) HandleButton(s *discordgo.Session, i *discordgo.InteractionCreate, sessionID string) error {
	// Get user's active character
	chars, err := h.services.CharacterService.ListByOwner(i.Member.User.ID)
	if err != nil || len(chars) == 0 {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ You need a character to join! Use `/dnd character create`",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}
	
	// Find first active character
	var playerChar *entities.Character
	for _, char := range chars {
		if char.Status == entities.CharacterStatusActive {
			playerChar = char
			break
		}
	}
	
	if playerChar == nil {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ No active character found! Activate a character first.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}
	
	// Join the session
	_, err = h.services.SessionService.JoinSession(context.Background(), sessionID, i.Member.User.ID)
	if err != nil {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("❌ Failed to join party: %v", err),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}
	
	// Select character
	err = h.services.SessionService.SelectCharacter(context.Background(), sessionID, i.Member.User.ID, playerChar.ID)
	if err != nil {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("❌ Failed to select character: %v", err),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}
	
	// Success response
	embed := &discordgo.MessageEmbed{
		Title:       "🎉 Joined the Party!",
		Description: fmt.Sprintf("**%s** has joined the dungeon delve!", playerChar.Name),
		Color:       0x2ecc71, // Green
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Character",
				Value:  fmt.Sprintf("%s %s (Level %d)", playerChar.Race.Name, playerChar.Class.Name, playerChar.Level),
				Inline: true,
			},
			{
				Name:   "HP",
				Value:  fmt.Sprintf("%d/%d", playerChar.CurrentHitPoints, playerChar.MaxHitPoints),
				Inline: true,
			},
			{
				Name:   "AC",
				Value:  fmt.Sprintf("%d", playerChar.AC),
				Inline: true,
			},
		},
	}
	
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
}