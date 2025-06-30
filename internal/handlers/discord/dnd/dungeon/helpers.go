package dungeon

import (
	"context"
	"fmt"
	session2 "github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/session"
	"log"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/session"
	"github.com/bwmarrin/discordgo"
)

// UpdateDungeonLobbyMessage updates the shared dungeon lobby message with current party members
func UpdateDungeonLobbyMessage(s *discordgo.Session, sessionService session.Service, characterService character.Service, sessionID, messageID, channelID string) error {
	// Get fresh session data
	sess, err := sessionService.GetSession(context.Background(), sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Get the original message to preserve its structure
	origMsg, err := s.ChannelMessage(channelID, messageID)
	if err != nil {
		return fmt.Errorf("failed to get original message: %w", err)
	}

	if len(origMsg.Embeds) == 0 {
		return fmt.Errorf("original message has no embeds")
	}

	// Build updated party list
	partyLines := buildPartyMembersList(sess, characterService)

	// Update the party field
	updatedEmbed := origMsg.Embeds[0]
	for idx, field := range updatedEmbed.Fields {
		if field.Name == "👥 Party" {
			updatedEmbed.Fields[idx].Value = strings.Join(partyLines, "\n")
			break
		}
	}

	// Edit the message
	_, err = s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		ID:         messageID,
		Channel:    channelID,
		Embeds:     &[]*discordgo.MessageEmbed{updatedEmbed},
		Components: &origMsg.Components,
	})

	if err != nil {
		return fmt.Errorf("failed to update message: %w", err)
	}

	log.Printf("Successfully updated dungeon lobby message %s", messageID)
	return nil
}

// buildPartyMembersList builds a list of party members with their characters
func buildPartyMembersList(sess *session2.Session, characterService character.Service) []string {
	var partyLines []string

	// TODO: Optimize with batch character fetching to avoid N+1 queries
	// For now, we fetch individually
	for userID, member := range sess.Members {
		if member.Role == session2.SessionRolePlayer {
			if member.CharacterID != "" {
				// Get character info
				char, err := characterService.GetByID(member.CharacterID)
				if err == nil && char != nil {
					partyLines = append(partyLines, fmt.Sprintf("<@%s> - %s", userID, char.Name))
				} else {
					partyLines = append(partyLines, fmt.Sprintf("<@%s> (character not found)", userID))
				}
			} else {
				partyLines = append(partyLines, fmt.Sprintf("<@%s> (no character selected)", userID))
			}
		}
	}

	return partyLines
}
