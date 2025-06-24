package combat

import (
	"fmt"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/encounter"
	"github.com/bwmarrin/discordgo"
)

// buildAttackResultEmbed creates an embed for attack results
func buildAttackResultEmbed(result *encounter.ExecuteAttackResult) *discordgo.MessageEmbed {
	attack := result.PlayerAttack
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("⚔️ %s attacks %s!", attack.AttackerName, attack.TargetName),
		Description: fmt.Sprintf("**%s**", attack.WeaponName),
		Color:       0xe74c3c,
		Fields:      []*discordgo.MessageEmbedField{},
	}

	// Attack roll
	hitText := "❌ **MISS!**"
	if attack.Critical {
		hitText = "🎆 **CRITICAL HIT!**"
	} else if attack.Hit {
		hitText = "✅ **HIT!**"
	}

	attackRoll := fmt.Sprintf("Roll: %v + %d = **%d** vs AC %d\n%s",
		attack.DiceRolls, attack.AttackBonus, attack.TotalAttack, attack.TargetAC, hitText)
	
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "🎲 Attack Roll",
		Value:  attackRoll,
		Inline: true,
	})

	// Damage if hit
	if attack.Hit {
		damageText := fmt.Sprintf("Roll: %v", attack.DamageRolls)
		if attack.DamageBonus != 0 {
			damageText += fmt.Sprintf(" + %d", attack.DamageBonus)
		}
		damageText += fmt.Sprintf(" = **%d** %s", attack.Damage, attack.DamageType)
		
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "💥 Damage",
			Value:  damageText,
			Inline: true,
		})

		// Target status
		targetStatus := fmt.Sprintf("%s: **%d HP**", attack.TargetName, attack.TargetNewHP)
		if attack.TargetDefeated {
			targetStatus += "\n💀 **DEFEATED!**"
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "🩸 Target Status",
			Value:  targetStatus,
			Inline: false,
		})
	}

	// Monster turns if any
	if len(result.MonsterAttacks) > 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  "━━━━━━━━━━━━━━━",
			Value: "**Monster Turns**",
			Inline: false,
		})

		for _, ma := range result.MonsterAttacks {
			var value string
			if ma.Hit {
				value = fmt.Sprintf("⚔️ Attacks %s with %s\n"+
					"Roll: %d vs AC %d - **HIT!**\n"+
					"💥 Damage: **%d** %s",
					ma.TargetName, ma.WeaponName, ma.TotalAttack, ma.TargetAC, ma.Damage, ma.DamageType)
			} else {
				value = fmt.Sprintf("⚔️ Attacks %s with %s\n"+
					"Roll: %d vs AC %d - **MISS!**",
					ma.TargetName, ma.WeaponName, ma.TotalAttack, ma.TargetAC)
			}
			
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   fmt.Sprintf("🐉 %s's Turn", ma.AttackerName),
				Value:  value,
				Inline: false,
			})
		}
	}

	// Victory check
	if result.CombatEnded && result.PlayersWon {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "🎉 Victory!",
			Value:  "All enemies have been defeated!",
			Inline: false,
		})
	}

	return embed
}

// buildCombatStatusEmbed creates a status embed with optional monster actions
func buildCombatStatusEmbed(enc *entities.Encounter, monsterActions []*encounter.AttackResult) *discordgo.MessageEmbed {
	current := enc.GetCurrentCombatant()
	
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("⚔️ Combat - Round %d", enc.Round),
		Color:       0x3498db,
		Fields:      []*discordgo.MessageEmbedField{},
	}

	// Current turn
	if current != nil {
		embed.Description = fmt.Sprintf("**%s's turn** (HP: %d/%d)", current.Name, current.CurrentHP, current.MaxHP)
	}

	// Monster actions if any
	if len(monsterActions) > 0 {
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
		line := fmt.Sprintf("%s **%s** - %d/%d HP\n", hpBar, c.Name, c.CurrentHP, c.MaxHP)
		
		if c.Type == entities.CombatantTypeMonster {
			enemies.WriteString(line)
		} else {
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
			status = fmt.Sprintf("%s %d/%d HP | AC %d", hpBar, c.CurrentHP, c.MaxHP, c.AC)
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
func buildCombatComponents(encounterID string, result *encounter.ExecuteAttackResult) []discordgo.MessageComponent {
	// Check if combat ended
	if result.CombatEnded {
		return []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "Combat Complete",
						Style:    discordgo.SuccessButton,
						Disabled: true,
						Emoji:    &discordgo.ComponentEmoji{Name: "🎉"},
					},
				},
			},
		}
	}

	// Normal combat buttons
	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Attack Again",
					Style:    discordgo.DangerButton,
					CustomID: fmt.Sprintf("combat:attack:%s", encounterID),
					Emoji:    &discordgo.ComponentEmoji{Name: "⚔️"},
					Disabled: !result.IsPlayerTurn,
				},
				discordgo.Button{
					Label:    "Next Turn",
					Style:    discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("combat:next_turn:%s", encounterID),
					Emoji:    &discordgo.ComponentEmoji{Name: "➡️"},
				},
				discordgo.Button{
					Label:    "View Status",
					Style:    discordgo.SecondaryButton,
					CustomID: fmt.Sprintf("combat:view:%s", encounterID),
					Emoji:    &discordgo.ComponentEmoji{Name: "📊"},
				},
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
func getHPBar(current, max int) string {
	if max == 0 {
		return "💀"
	}
	percent := float64(current) / float64(max)
	if percent > 0.5 {
		return "🟢"
	} else if percent > 0.25 {
		return "🟡"
	} else if current > 0 {
		return "🔴"
	}
	return "💀"
}