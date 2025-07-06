package combat

import (
	"fmt"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/combat"
	"strings"
)

// BuildInitiativeDisplay creates a formatted initiative order display
func BuildInitiativeDisplay(enc *combat.Encounter) string {
	var sb strings.Builder

	// Use ANSI color codes for better visual distinction
	sb.WriteString("```ansi\n")

	// Header with color
	sb.WriteString("\u001b[1;36m") // Bold cyan for header
	sb.WriteString("Init │ Name                  │ HP       │ AC  │ Prof\n")
	sb.WriteString("─────┼─────────────────────┼────────┼─────┼─────\n")
	sb.WriteString("\u001b[0m") // Reset color

	// Show combatants in initiative order
	for i, id := range enc.TurnOrder {
		c, exists := enc.Combatants[id]
		if !exists || !c.IsActive {
			continue
		}

		// Current turn indicator
		if i == enc.Turn {
			sb.WriteString("\u001b[1;33m▶ ") // Bold yellow for current turn
		} else {
			sb.WriteString("  ")
		}

		// Initiative score
		sb.WriteString(fmt.Sprintf("%2d", c.Initiative))
		sb.WriteString(" │ ")

		// Name with type coloring
		if c.Type == combat.CombatantTypeMonster {
			sb.WriteString("\u001b[1;31m") // Bold red for monsters
		} else {
			sb.WriteString("\u001b[1;32m") // Bold green for players
		}

		nameDisplay := c.Name
		if c.Type == combat.CombatantTypePlayer && c.Class != "" {
			nameDisplay = fmt.Sprintf("%s (%s)", c.Name, c.Class)
		}
		if len(nameDisplay) > 21 {
			nameDisplay = nameDisplay[:18] + "..."
		}
		sb.WriteString(fmt.Sprintf("%-21s", nameDisplay))
		sb.WriteString("\u001b[0m") // Reset color

		sb.WriteString(" │ ")

		// HP with color based on health
		hpPercent := float64(c.CurrentHP) / float64(c.MaxHP)
		if hpPercent > 0.5 {
			sb.WriteString("\u001b[32m") // Green
		} else if hpPercent > 0.25 {
			sb.WriteString("\u001b[33m") // Yellow
		} else if c.CurrentHP > 0 {
			sb.WriteString("\u001b[31m") // Red
		} else {
			sb.WriteString("\u001b[90m") // Gray for dead
		}

		hpDisplay := fmt.Sprintf("%3d/%-3d", c.CurrentHP, c.MaxHP)
		sb.WriteString(hpDisplay)
		sb.WriteString("\u001b[0m") // Reset color

		sb.WriteString(" │ ")

		// AC
		sb.WriteString(fmt.Sprintf("%2d", c.AC))

		// End turn indicator if current
		if i == enc.Turn {
			sb.WriteString("\u001b[0m") // Ensure color reset
		}

		sb.WriteString("\n")
	}

	sb.WriteString("```")
	return sb.String()
}

// BuildCombatSummaryDisplay creates a summary of the current combat state
func BuildCombatSummaryDisplay(enc *combat.Encounter) string {
	var sb strings.Builder

	// Count active combatants by type
	var activeMonsters, activePlayers int
	for _, c := range enc.Combatants {
		if c.IsActive {
			if c.Type == combat.CombatantTypeMonster {
				activeMonsters++
			} else {
				activePlayers++
			}
		}
	}

	sb.WriteString(fmt.Sprintf("⚔️ **Round %d** | ", enc.Round))
	sb.WriteString(fmt.Sprintf("🛡️ Players: %d | ", activePlayers))
	sb.WriteString(fmt.Sprintf("🐉 Monsters: %d", activeMonsters))

	// Add current turn info
	if current := enc.GetCurrentCombatant(); current != nil {
		sb.WriteString(fmt.Sprintf("\n🎯 **%s's turn**", current.Name))
	}

	return sb.String()
}
