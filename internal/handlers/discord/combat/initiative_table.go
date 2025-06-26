package combat

import (
	"fmt"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/bwmarrin/discordgo"
)

// BuildInitiativeFields creates Discord embed fields for initiative order
func BuildInitiativeFields(enc *entities.Encounter) []*discordgo.MessageEmbedField {
	var fields []*discordgo.MessageEmbedField

	// Build two columns: Players and Monsters
	var playerLines, monsterLines strings.Builder

	for i, id := range enc.TurnOrder {
		if c, exists := enc.Combatants[id]; exists && c.IsActive {
			// Turn indicator
			turnMarker := ""
			if i == enc.Turn {
				turnMarker = "▶ "
			}

			// HP visual indicator
			hpIcon := getHPIcon(c.CurrentHP, c.MaxHP)

			// Name (truncated if needed)
			name := c.Name
			if len(name) > 15 {
				name = name[:12] + "..."
			}

			// Build the line
			line := fmt.Sprintf("%s`%2d` %s **%s**\n├─ %s HP: %d/%d | AC: %d\n",
				turnMarker,
				c.Initiative,
				hpIcon,
				name,
				getHPBar(c.CurrentHP, c.MaxHP),
				c.CurrentHP,
				c.MaxHP,
				c.AC,
			)

			// Add to appropriate column
			if c.Type == entities.CombatantTypePlayer {
				playerLines.WriteString(line)
				if c.Class != "" {
					playerLines.WriteString(fmt.Sprintf("└─ *%s*\n\n", c.Class))
				} else {
					playerLines.WriteString("\n")
				}
			} else {
				monsterLines.WriteString(line)
				monsterLines.WriteString("\n")
			}
		}
	}

	// Add player field if there are players
	if playerLines.Len() > 0 {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "🛡️ Party Members",
			Value:  playerLines.String(),
			Inline: true,
		})
	}

	// Add monster field if there are monsters
	if monsterLines.Len() > 0 {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "⚔️ Enemies",
			Value:  monsterLines.String(),
			Inline: true,
		})
	}

	return fields
}

// BuildCompactInitiativeDisplay creates a compact single-line display for each combatant
func BuildCompactInitiativeDisplay(enc *entities.Encounter) string {
	var sb strings.Builder

	sb.WriteString("```css\n")
	sb.WriteString("[Initiative Order]\n")
	sb.WriteString("─────────────────────────────────────────\n")

	for i, id := range enc.TurnOrder {
		if c, exists := enc.Combatants[id]; exists && c.IsActive {
			// Current turn indicator
			if i == enc.Turn {
				sb.WriteString("▶ ")
			} else {
				sb.WriteString("  ")
			}

			// Initiative
			sb.WriteString(fmt.Sprintf("[%2d] ", c.Initiative))

			// Name with type indicator
			typeIcon := "👤"
			if c.Type == entities.CombatantTypeMonster {
				typeIcon = "👹"
			}

			name := c.Name
			if len(name) > 12 {
				name = name[:10] + ".."
			}
			sb.WriteString(fmt.Sprintf("%s %-12s ", typeIcon, name))

			// HP bar
			sb.WriteString(fmt.Sprintf("HP[%s] ", getCompactHPBar(c.CurrentHP, c.MaxHP)))

			// AC
			sb.WriteString(fmt.Sprintf("AC:%2d", c.AC))

			// Status effects (future enhancement)
			// if c.HasConditions() {
			//     sb.WriteString(" [!]")
			// }

			sb.WriteString("\n")
		}
	}

	sb.WriteString("```")
	return sb.String()
}

// getHPIcon returns an emoji based on health percentage
func getHPIcon(current, maxHP int) string {
	if maxHP == 0 {
		return "💀"
	}
	percent := float64(current) / float64(maxHP)
	if percent > 0.75 {
		return "💚" // Healthy
	} else if percent > 0.5 {
		return "💛" // Good
	} else if percent > 0.25 {
		return "🧡" // Hurt
	} else if current > 0 {
		return "❤️" // Critical
	}
	return "💀" // Dead
}

// getCompactHPBar returns a compact visual HP bar
func getCompactHPBar(current, maxHP int) string {
	if maxHP == 0 {
		return "████" // All black for dead
	}

	percent := float64(current) / float64(maxHP)
	filled := int(percent * 8) // 8 character bar

	bar := ""
	for i := 0; i < 8; i++ {
		if i < filled {
			bar += "▰"
		} else {
			bar += "▱"
		}
	}

	return bar
}

// BuildDetailedCombatant creates a detailed view of a single combatant
// This could be shown when someone clicks a button or uses a command
func BuildDetailedCombatant(c *entities.Combatant) *discordgo.MessageEmbed {
	color := 0x3498db // Blue default
	if c.Type == entities.CombatantTypeMonster {
		color = 0xe74c3c // Red for monsters
	}

	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("%s %s", getHPIcon(c.CurrentHP, c.MaxHP), c.Name),
		Color: color,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "📊 Stats",
				Value:  fmt.Sprintf("**HP:** %d/%d\n**AC:** %d\n**Initiative:** %d", c.CurrentHP, c.MaxHP, c.AC, c.Initiative),
				Inline: true,
			},
		},
	}

	if c.Class != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "⚔️ Class",
			Value:  c.Class,
			Inline: true,
		})
	}

	// Add visual HP bar
	hpPercent := float64(c.CurrentHP) / float64(c.MaxHP)
	hpBar := ""
	for i := 0; i < 10; i++ {
		if float64(i)/10 < hpPercent {
			hpBar += "🟩"
		} else {
			hpBar += "⬜"
		}
	}

	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "💚 Health",
		Value:  hpBar,
		Inline: false,
	})

	return embed
}
