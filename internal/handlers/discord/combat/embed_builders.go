package combat

import (
	"fmt"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/encounter"
	"github.com/bwmarrin/discordgo"
)

// buildCombatStatusEmbed creates a status embed with optional monster actions
// BuildCombatStatusEmbed creates the main combat status embed
func BuildCombatStatusEmbed(enc *entities.Encounter, monsterActions []*encounter.AttackResult) *discordgo.MessageEmbed {
	return BuildCombatStatusEmbedForPlayer(enc, monsterActions, "")
}

// BuildCombatStatusEmbedForPlayer creates a player-focused combat status embed
func BuildCombatStatusEmbedForPlayer(enc *entities.Encounter, monsterActions []*encounter.AttackResult, playerName string) *discordgo.MessageEmbed {
	current := enc.GetCurrentCombatant()

	embed := &discordgo.MessageEmbed{
		Title:  fmt.Sprintf("⚔️ Combat - Round %d", enc.Round),
		Color:  0x3498db,
		Fields: []*discordgo.MessageEmbedField{},
	}

	// Current turn
	if current != nil {
		embed.Description = fmt.Sprintf("**%s's turn** (HP: %d/%d)", current.Name, current.CurrentHP, current.MaxHP)
	}

	// Create round summary if we have actions
	if len(monsterActions) > 0 && playerName != "" {
		roundSummary := NewRoundSummary(enc.Round)

		// Record monster actions
		for _, ma := range monsterActions {
			roundSummary.RecordAttack(AttackInfo{
				AttackerName: ma.AttackerName,
				TargetName:   ma.TargetName,
				Damage:       ma.Damage,
				Hit:          ma.Hit,
				Critical:     ma.Critical,
				WeaponName:   ma.WeaponName,
			})
		}

		// Add player-focused summary
		playerSummary := roundSummary.GetPlayerSummary(playerName)
		if playerSummary != "" {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "⚔️ Your Combat Summary",
				Value:  playerSummary,
				Inline: false,
			})
		}
	} else if len(monsterActions) > 0 {
		// Fallback to old style for non-player-specific views
		for _, ma := range monsterActions {
			var value string
			if ma.Hit {
				value = fmt.Sprintf("Attacked %s for **%d** damage!", ma.TargetName, ma.Damage)
			} else {
				value = fmt.Sprintf("Missed %s!", ma.TargetName)
			}
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   fmt.Sprintf("🐉 %s", ma.AttackerName),
				Value:  value,
				Inline: false,
			})
		}
	}

	// Active combatants
	var enemies, allies strings.Builder
	for _, c := range enc.Combatants {
		if !c.IsActive {
			continue
		}

		hpBar := getHPBar(c.CurrentHP, c.MaxHP)

		if c.Type == entities.CombatantTypeMonster {
			line := fmt.Sprintf("%s **%s** - %d/%d HP\n", hpBar, c.Name, c.CurrentHP, c.MaxHP)
			enemies.WriteString(line)
		} else {
			// For players, show class icon, name, class, HP, and AC
			classIcon := getClassIcon(c.Class)
			classInfo := ""
			if c.Class != "" {
				classInfo = fmt.Sprintf(" (%s)", c.Class)
			}
			line := fmt.Sprintf("%s %s **%s**%s - %d/%d HP | AC %d\n",
				classIcon, hpBar, c.Name, classInfo, c.CurrentHP, c.MaxHP, c.AC)
			allies.WriteString(line)
		}
	}

	if enemies.Len() > 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "🐉 Enemies",
			Value:  enemies.String(),
			Inline: true,
		})
	}

	if allies.Len() > 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "🛡️ Allies",
			Value:  allies.String(),
			Inline: true,
		})
	}

	// Add combat history - last 5 entries
	if len(enc.CombatLog) > 0 {
		var history strings.Builder
		start := len(enc.CombatLog) - 5
		if start < 0 {
			start = 0
		}
		for i := start; i < len(enc.CombatLog); i++ {
			history.WriteString("• " + enc.CombatLog[i] + "\n")
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "📜 Recent Actions",
			Value:  history.String(),
			Inline: false,
		})
	}

	return embed
}

// buildDetailedCombatEmbed creates a detailed view of combat
func buildDetailedCombatEmbed(enc *entities.Encounter) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Title:       enc.Name,
		Description: fmt.Sprintf("**Status:** %s | **Round:** %d", enc.Status, enc.Round),
		Color:       0x3498db,
		Fields:      []*discordgo.MessageEmbedField{},
	}

	// Turn order
	var turnOrder strings.Builder
	for i, id := range enc.TurnOrder {
		if c, exists := enc.Combatants[id]; exists && c.IsActive {
			prefix := "  "
			if i == enc.Turn {
				prefix = "▶️"
			}
			turnOrder.WriteString(fmt.Sprintf("%s %s (Init: %d)\n", prefix, c.Name, c.Initiative))
		}
	}

	if turnOrder.Len() > 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "📋 Initiative Order",
			Value:  turnOrder.String(),
			Inline: false,
		})
	}

	// All combatants with details
	var combatantList strings.Builder
	for _, c := range enc.Combatants {
		status := "💀 Defeated"
		if c.IsActive {
			hpBar := getHPBar(c.CurrentHP, c.MaxHP)
			if c.Type == entities.CombatantTypePlayer && c.Class != "" {
				classIcon := getClassIcon(c.Class)
				status = fmt.Sprintf("%s %s %d/%d HP | AC %d | %s", classIcon, hpBar, c.CurrentHP, c.MaxHP, c.AC, c.Class)
			} else {
				status = fmt.Sprintf("%s %d/%d HP | AC %d", hpBar, c.CurrentHP, c.MaxHP, c.AC)
			}
		}
		combatantList.WriteString(fmt.Sprintf("**%s**\n%s\n\n", c.Name, status))
	}

	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "⚔️ Combatants",
		Value:  combatantList.String(),
		Inline: false,
	})

	// Recent combat log
	if len(enc.CombatLog) > 0 {
		var recentLog strings.Builder
		start := len(enc.CombatLog) - 5
		if start < 0 {
			start = 0
		}
		for i := start; i < len(enc.CombatLog); i++ {
			recentLog.WriteString(enc.CombatLog[i] + "\n")
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "📜 Recent Actions",
			Value:  recentLog.String(),
			Inline: false,
		})
	}

	return embed
}

// buildCombatComponents creates appropriate buttons based on combat state
// BuildCombatComponents creates the combat UI components
func BuildCombatComponents(encounterID string, result *encounter.ExecuteAttackResult) []discordgo.MessageComponent {
	// Check if combat ended
	if result.CombatEnded {
		style := discordgo.SuccessButton
		emoji := "🎉"
		if !result.PlayersWon {
			style = discordgo.DangerButton
			emoji = "💀"
		}

		return []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "View History",
						Style:    discordgo.PrimaryButton,
						CustomID: fmt.Sprintf("combat:history:%s", encounterID),
						Emoji:    &discordgo.ComponentEmoji{Name: "📜"},
					},
					discordgo.Button{
						Label:    "Combat Summary",
						Style:    style,
						CustomID: fmt.Sprintf("combat:summary:%s", encounterID),
						Emoji:    &discordgo.ComponentEmoji{Name: emoji},
					},
				},
			},
		}
	}

	// Normal combat buttons - no Attack button on shared messages
	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Next Turn",
					Style:    discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("combat:next_turn:%s", encounterID),
					Emoji:    &discordgo.ComponentEmoji{Name: "➡️"},
				},
				discordgo.Button{
					Label:    "Get My Actions",
					Style:    discordgo.SuccessButton,
					CustomID: fmt.Sprintf("combat:my_actions:%s", encounterID),
					Emoji:    &discordgo.ComponentEmoji{Name: "🎯"},
				},
				discordgo.Button{
					Label:    "View Status",
					Style:    discordgo.SecondaryButton,
					CustomID: fmt.Sprintf("combat:view:%s", encounterID),
					Emoji:    &discordgo.ComponentEmoji{Name: "📊"},
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "History",
					Style:    discordgo.SecondaryButton,
					CustomID: fmt.Sprintf("combat:history:%s", encounterID),
					Emoji:    &discordgo.ComponentEmoji{Name: "📜"},
				},
			},
		},
	}
}

// getHPBar returns an emoji HP indicator
// getClassIcon returns an emoji icon for the character class
func getClassIcon(class string) string {
	switch class {
	case "Fighter":
		return "⚔️"
	case "Wizard":
		return "🧙"
	case "Cleric":
		return "✨"
	case "Rogue":
		return "🗡️"
	case "Ranger":
		return "🏹"
	case "Barbarian":
		return "🪓"
	case "Paladin":
		return "🛡️"
	case "Monk":
		return "👊"
	case "Warlock":
		return "🔮"
	case "Sorcerer":
		return "⚡"
	case "Druid":
		return "🌿"
	case "Bard":
		return "🎵"
	default:
		return "👤"
	}
}

func getHPBar(current, maxHP int) string {
	if maxHP == 0 {
		return "💀"
	}
	percent := float64(current) / float64(maxHP)
	if percent > 0.5 {
		return "🟢"
	} else if percent > 0.25 {
		return "🟡"
	} else if current > 0 {
		return "🔴"
	}
	return "💀"
}
