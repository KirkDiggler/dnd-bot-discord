package combat

import (
	"fmt"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/bwmarrin/discordgo"
)

// BuildInitiativeFields creates Discord embed fields for initiative order
func BuildInitiativeFields(enc *entities.Encounter) []*discordgo.MessageEmbedField {
	// Build a single table-style display
	var sb strings.Builder

	// Use ANSI code block for color support
	sb.WriteString("```ansi\n")
	sb.WriteString("Init│Name              │HP               │ AC\n")
	sb.WriteString("────┼──────────────────┼─────────────────┼───\n")

	for i, id := range enc.TurnOrder {
		c, exists := enc.Combatants[id]
		if !exists {
			continue
		}

		// Current turn indicator and initiative in fixed width
		if i == enc.Turn {
			sb.WriteString(fmt.Sprintf("▶%-2d", c.Initiative))
		} else {
			sb.WriteString(fmt.Sprintf(" %-2d", c.Initiative))
		}
		sb.WriteString(" │")

		// Name with icon (truncated if needed)
		icon := ""
		if c.CurrentHP == 0 {
			icon = "💀" // Dead indicator replaces type icon
		} else if c.Type == entities.CombatantTypePlayer {
			icon = getClassIcon(c.Class)
		} else {
			icon = "🐉" // Monster icon
		}

		name := c.Name
		maxNameLen := 13 // Reduced to fit better
		if len(name) > maxNameLen {
			name = name[:maxNameLen-3] + "..."
		}

		// Format: "icon name" padded to visual width of 16
		nameStr := fmt.Sprintf("%s %s", icon, name)

		// Some emojis have variation selectors (️) that make them wider
		// These need less padding to align properly
		visualWidth := 16
		if strings.Contains(icon, "️") {
			// Icons with variation selectors need adjustment
			visualWidth = 14
		}

		sb.WriteString(fmt.Sprintf("%-*s", visualWidth, nameStr))
		sb.WriteString(" │")

		// HP with visual bar and color coding
		percent := float64(c.CurrentHP) / float64(c.MaxHP)
		hpBar := getCompactHPBar(c.CurrentHP, c.MaxHP)
		hpText := fmt.Sprintf("%3d/%-3d", c.CurrentHP, c.MaxHP)

		// Apply color based on health
		if c.CurrentHP == 0 {
			sb.WriteString("\u001b[90m") // Gray for dead
		} else if percent > 0.5 {
			sb.WriteString("\u001b[32m") // Green
		} else if percent > 0.25 {
			sb.WriteString("\u001b[33m") // Yellow
		} else {
			sb.WriteString("\u001b[31m") // Red
		}

		// Write HP bar and text (total 17 chars: 8 bar + 1 space + 8 text)
		sb.WriteString(fmt.Sprintf("%-17s", hpBar+" "+hpText))
		sb.WriteString("\u001b[0m") // Reset color
		sb.WriteString(" │")

		// AC
		sb.WriteString(fmt.Sprintf("%2d", c.AC))

		sb.WriteString("\n")
	}

	sb.WriteString("```")

	return []*discordgo.MessageEmbedField{
		{
			Name:   "🎯 Initiative Order",
			Value:  sb.String(),
			Inline: false,
		},
	}
}

// BuildCompactInitiativeDisplay creates a compact single-line display for each combatant
func BuildCompactInitiativeDisplay(enc *entities.Encounter) string {
	var sb strings.Builder

	sb.WriteString("```css\n")
	sb.WriteString("[Initiative Order]\n")
	sb.WriteString("─────────────────────────────────────────\n")

	for i, id := range enc.TurnOrder {
		c, exists := enc.Combatants[id]
		if !exists {
			continue
		}

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

// getCompactHPBar returns a compact visual HP bar using single-width characters
func getCompactHPBar(current, maxHP int) string {
	if maxHP == 0 || current == 0 {
		return "████████" // All filled for dead
	}

	percent := float64(current) / float64(maxHP)
	filled := int(percent * 8) // 8 character bar

	bar := ""
	for i := 0; i < 8; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
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
