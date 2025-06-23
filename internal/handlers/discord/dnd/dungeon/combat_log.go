package dungeon

import (
	"fmt"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/bwmarrin/discordgo"
)

// CreateCombatLogMessage creates a public message to track combat progress
func CreateCombatLogMessage(s *discordgo.Session, channelID string, room *Room, enc *entities.Encounter) (*discordgo.Message, error) {
	embed := buildCombatLogEmbed(room, enc)
	components := buildCombatLogComponents(enc)

	msg, err := s.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Embed:      embed,
		Components: components,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create combat log message: %w", err)
	}

	return msg, nil
}

// UpdateCombatLogMessage updates the public combat log with new information
func UpdateCombatLogMessage(s *discordgo.Session, channelID, messageID string, room *Room, enc *entities.Encounter) error {
	embed := buildCombatLogEmbed(room, enc)
	components := buildCombatLogComponents(enc)

	_, err := s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		Channel:    channelID,
		ID:         messageID,
		Embed:      embed,
		Components: &components,
	})
	if err != nil {
		return fmt.Errorf("failed to update combat log message: %w", err)
	}

	return nil
}

// buildCombatLogEmbed creates the embed for the combat log
func buildCombatLogEmbed(room *Room, enc *entities.Encounter) *discordgo.MessageEmbed {
	// Determine color based on status
	color := 0xe74c3c // Red for active combat
	if enc.Status == entities.EncounterStatusCompleted {
		shouldEnd, playersWon := enc.CheckCombatEnd()
		if shouldEnd && playersWon {
			color = 0x2ecc71 // Green for victory
		} else if shouldEnd && !playersWon {
			color = 0x95a5a6 // Gray for defeat
		}
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("⚔️ %s - Combat Log", room.Name),
		Description: getCombatDescription(enc),
		Color:       color,
		Fields:      []*discordgo.MessageEmbedField{},
	}

	// Show combatants status
	var playerStatus strings.Builder
	var monsterStatus strings.Builder

	for _, combatant := range enc.Combatants {
		status := fmt.Sprintf("• **%s** ", combatant.Name)
		if combatant.IsActive {
			status += fmt.Sprintf("HP: %d/%d", combatant.CurrentHP, combatant.MaxHP)
		} else {
			status += "💀 Defeated"
		}
		status += "\n"

		if combatant.Type == entities.CombatantTypePlayer {
			playerStatus.WriteString(status)
		} else {
			monsterStatus.WriteString(status)
		}
	}

	if playerStatus.Len() > 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "🛡️ Party Status",
			Value:  playerStatus.String(),
			Inline: true,
		})
	}

	if monsterStatus.Len() > 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "🐉 Enemy Status",
			Value:  monsterStatus.String(),
			Inline: true,
		})
	}

	// Show current turn if active
	if enc.Status == entities.EncounterStatusActive {
		// Check if we're at the end of a round
		if enc.Turn >= len(enc.TurnOrder)-1 && len(enc.TurnOrder) > 0 {
			// Show round summary
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "🏁 Round Complete",
				Value:  fmt.Sprintf("**Round %d Complete!** Click continue to start Round %d.", enc.Round, enc.Round+1),
				Inline: false,
			})
			
			// Show round summary stats
			var roundSummary strings.Builder
			roundSummary.WriteString(fmt.Sprintf("📊 **Round %d Summary:**\n", enc.Round))
			
			// Count damage dealt this round
			knockouts := 0
			
			// Parse combat log for this round's actions
			for _, entry := range enc.CombatLog {
				if strings.HasPrefix(entry, fmt.Sprintf("Round %d:", enc.Round)) {
					if strings.Contains(entry, "damage") {
						// Simple damage parsing - could be improved
						if strings.Contains(entry, "defeated") {
							knockouts++
						}
					}
				}
			}
			
			if knockouts > 0 {
				roundSummary.WriteString(fmt.Sprintf("💀 Combatants defeated: %d\n", knockouts))
			}
			roundSummary.WriteString(fmt.Sprintf("⏱️ Turns taken: %d", len(enc.TurnOrder)))
			
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "📜 Round Details",
				Value:  roundSummary.String(),
				Inline: false,
			})
		} else if current := enc.GetCurrentCombatant(); current != nil {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "🎯 Current Turn",
				Value:  fmt.Sprintf("**%s's turn** (Round %d)", current.Name, enc.Round),
				Inline: false,
			})
		}
	}

	// Show recent combat actions (last 5)
	if len(enc.CombatLog) > 0 {
		var recentActions strings.Builder

		// Skip initiative rolls, focus on combat actions
		actionCount := 0
		for i := len(enc.CombatLog) - 1; i >= 0 && actionCount < 5; i-- {
			entry := enc.CombatLog[i]
			// Skip initiative-related entries
			if !strings.Contains(entry, "Rolling Initiative") && !strings.Contains(entry, "rolls initiative:") {
				recentActions.WriteString("⚔️ " + entry + "\n")
				actionCount++
			}
		}

		if recentActions.Len() > 0 {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "📜 Recent Actions",
				Value:  recentActions.String(),
				Inline: false,
			})
		}
	}

	// Add footer with round info
	embed.Footer = &discordgo.MessageEmbedFooter{
		Text: fmt.Sprintf("Round %d | %s", enc.Round, enc.Status),
	}

	return embed
}

// buildCombatLogComponents creates buttons for the combat log
func buildCombatLogComponents(enc *entities.Encounter) []discordgo.MessageComponent {
	// No buttons if combat is completed
	if enc.Status != entities.EncounterStatusActive {
		return nil
	}

	// Check if we need to show round continue button
	// This happens when we're at the end of the turn order (about to start new round)
	if enc.Turn >= len(enc.TurnOrder)-1 && len(enc.TurnOrder) > 0 {
		// Show continue button for next round
		return []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    fmt.Sprintf("Continue to Round %d", enc.Round+1),
						Style:    discordgo.SuccessButton,
						CustomID: fmt.Sprintf("encounter:continue_round:%s", enc.ID),
						Emoji:    &discordgo.ComponentEmoji{Name: "▶️"},
					},
				},
			},
		}
	}

	// Get current combatant
	current := enc.GetCurrentCombatant()
	if current == nil || current.Type != entities.CombatantTypePlayer {
		// No button if it's not a player's turn
		return nil
	}

	// Single "My Turn" button that opens the action controller
	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "🎮 Open My Actions",
					Style:    discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("encounter:my_turn:%s", enc.ID),
					Emoji:    &discordgo.ComponentEmoji{Name: "⚔️"},
				},
			},
		},
	}
}

// getCombatDescription returns appropriate description based on encounter status
func getCombatDescription(enc *entities.Encounter) string {
	switch enc.Status {
	case entities.EncounterStatusActive:
		return "⚔️ Battle rages on! Heroes clash with monsters in deadly combat!"
	case entities.EncounterStatusCompleted:
		shouldEnd, playersWon := enc.CheckCombatEnd()
		if shouldEnd && playersWon {
			return "🎉 **Victory!** The party has triumphed over their foes!"
		} else if shouldEnd && !playersWon {
			return "💀 **Defeat!** The party has fallen in battle..."
		}
		return "Combat has ended."
	default:
		return "Preparing for battle..."
	}
}

// CreateCombatEndMessage creates a summary message when combat ends
func CreateCombatEndMessage(s *discordgo.Session, channelID string, room *Room, enc *entities.Encounter, loot []*entities.Equipment) error {
	_, playersWon := enc.CheckCombatEnd()

	embed := &discordgo.MessageEmbed{
		Title:  fmt.Sprintf("⚔️ Combat Complete - %s", room.Name),
		Fields: []*discordgo.MessageEmbedField{},
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Combat ID: %s | Use /dnd encounter history to view full details", enc.ID),
		},
	}

	if playersWon {
		embed.Color = 0x2ecc71 // Green
		embed.Description = "🎉 **Victory!** The party has defeated all enemies!"

		// Show loot if any
		if len(loot) > 0 {
			var lootList strings.Builder
			for _, item := range loot {
				lootList.WriteString(fmt.Sprintf("• %s\n", (*item).GetName()))
			}
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "💰 Treasure Found",
				Value:  lootList.String(),
				Inline: false,
			})
		}

		// Show surviving party members
		var survivors strings.Builder
		survivorCount := 0
		for _, combatant := range enc.Combatants {
			if combatant.Type == entities.CombatantTypePlayer && combatant.IsActive {
				survivors.WriteString(fmt.Sprintf("• **%s** (%d/%d HP)\n", combatant.Name, combatant.CurrentHP, combatant.MaxHP))
				survivorCount++
			}
		}
		if survivorCount > 0 {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "🛡️ Survivors",
				Value:  survivors.String(),
				Inline: false,
			})
		}
	} else {
		embed.Color = 0xe74c3c // Red
		embed.Description = "💀 **Defeat!** The party has been overwhelmed..."

		// Show what defeated them
		var remainingEnemies strings.Builder
		for _, combatant := range enc.Combatants {
			if combatant.Type == entities.CombatantTypeMonster && combatant.IsActive {
				remainingEnemies.WriteString(fmt.Sprintf("• **%s** (%d HP remaining)\n", combatant.Name, combatant.CurrentHP))
			}
		}
		if remainingEnemies.Len() > 0 {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "🐉 Victorious Enemies",
				Value:  remainingEnemies.String(),
				Inline: false,
			})
		}
	}

	// Combat statistics
	var stats strings.Builder
	stats.WriteString(fmt.Sprintf("⏱️ Duration: %d rounds\n", enc.Round))
	stats.WriteString(fmt.Sprintf("⚔️ Total actions: %d\n", len(enc.CombatLog)))

	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "📊 Combat Statistics",
		Value:  stats.String(),
		Inline: false,
	})

	// Add button to view full history
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "View Full Combat History",
					Style:    discordgo.SecondaryButton,
					CustomID: fmt.Sprintf("encounter:history:%s", enc.ID),
					Emoji:    &discordgo.ComponentEmoji{Name: "📜"},
				},
			},
		},
	}

	_, err := s.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Embed:      embed,
		Components: components,
	})
	return err
}
