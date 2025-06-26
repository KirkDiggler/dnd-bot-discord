package combat

import (
	"context"
	"fmt"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/bwmarrin/discordgo"
)

// handleSummary shows a summary of the completed combat
func (h *Handler) handleSummary(s *discordgo.Session, i *discordgo.InteractionCreate, encounterID string) error {
	enc, err := h.encounterService.GetEncounter(context.Background(), encounterID)
	if err != nil {
		return respondError(s, i, "Failed to get encounter", err)
	}

	// Build summary embed
	embed := &discordgo.MessageEmbed{
		Title:  fmt.Sprintf("⚔️ Combat Summary - %s", enc.Name),
		Color:  0x9b59b6, // Purple
		Fields: []*discordgo.MessageEmbedField{},
	}

	// Determine outcome
	if enc.Status == entities.EncounterStatusCompleted {
		// Check who won based on remaining combatants
		playersAlive := 0
		monstersAlive := 0
		for _, c := range enc.Combatants {
			if c.IsActive && c.CurrentHP > 0 {
				if c.Type == entities.CombatantTypePlayer {
					playersAlive++
				} else if c.Type == entities.CombatantTypeMonster {
					monstersAlive++
				}
			}
		}

		if playersAlive > 0 && monstersAlive == 0 {
			embed.Description = "🎉 **VICTORY!** The party emerged triumphant!"
			embed.Color = 0x00ff00 // Green
		} else if playersAlive == 0 && monstersAlive > 0 {
			embed.Description = "💀 **DEFEAT!** The party has fallen..."
			embed.Color = 0xff0000 // Red
		} else {
			embed.Description = "⚔️ Combat has ended."
		}
	}

	// Combat duration
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "📊 Combat Statistics",
		Value:  fmt.Sprintf("**Rounds:** %d\n**Total Actions:** %d", enc.Round, len(enc.CombatLog)),
		Inline: true,
	})

	// Casualty report
	var casualties strings.Builder
	var survivors strings.Builder

	for _, c := range enc.Combatants {
		status := fmt.Sprintf("**%s** - ", c.Name)
		if c.CurrentHP <= 0 {
			status += "💀 Defeated\n"
			casualties.WriteString(status)
		} else {
			status += fmt.Sprintf("❤️ %d/%d HP\n", c.CurrentHP, c.MaxHP)
			survivors.WriteString(status)
		}
	}

	if casualties.Len() > 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "💀 Casualties",
			Value:  casualties.String(),
			Inline: true,
		})
	}

	if survivors.Len() > 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "🛡️ Survivors",
			Value:  survivors.String(),
			Inline: true,
		})
	}

	// TODO: Add damage dealt/taken statistics
	// TODO: Add loot summary when implemented
	// TODO: Add XP gained when implemented

	embed.Footer = &discordgo.MessageEmbedFooter{
		Text: "Use the History button to see the full combat log",
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
}