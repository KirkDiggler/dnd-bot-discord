package combat

import (
	"fmt"
	"strings"
)

// RoundSummary tracks combat actions for a single round
type RoundSummary struct {
	Round         int
	PlayerActions map[string]*PlayerRoundInfo // key is player name
}

// PlayerRoundInfo tracks a player's combat stats for the round
type PlayerRoundInfo struct {
	Name         string
	DamageDealt  int
	DamageTaken  int
	AttacksMade  []AttackInfo
	AttacksRecvd []AttackInfo
}

// AttackInfo represents a single attack
type AttackInfo struct {
	AttackerName string
	TargetName   string
	Damage       int
	Hit          bool
	Critical     bool
	WeaponName   string
}

// NewRoundSummary creates a new round summary
func NewRoundSummary(round int) *RoundSummary {
	return &RoundSummary{
		Round:         round,
		PlayerActions: make(map[string]*PlayerRoundInfo),
	}
}

// RecordAttack records an attack in the round summary
func (rs *RoundSummary) RecordAttack(attack AttackInfo) {
	// Record for attacker (if player)
	if _, exists := rs.PlayerActions[attack.AttackerName]; !exists {
		rs.PlayerActions[attack.AttackerName] = &PlayerRoundInfo{
			Name:         attack.AttackerName,
			AttacksMade:  []AttackInfo{},
			AttacksRecvd: []AttackInfo{},
		}
	}

	// Record for target (if player)
	if _, exists := rs.PlayerActions[attack.TargetName]; !exists {
		rs.PlayerActions[attack.TargetName] = &PlayerRoundInfo{
			Name:         attack.TargetName,
			AttacksMade:  []AttackInfo{},
			AttacksRecvd: []AttackInfo{},
		}
	}

	// Update attacker stats
	if attacker := rs.PlayerActions[attack.AttackerName]; attacker != nil {
		attacker.AttacksMade = append(attacker.AttacksMade, attack)
		if attack.Hit {
			attacker.DamageDealt += attack.Damage
		}
	}

	// Update target stats
	if target := rs.PlayerActions[attack.TargetName]; target != nil {
		target.AttacksRecvd = append(target.AttacksRecvd, attack)
		if attack.Hit {
			target.DamageTaken += attack.Damage
		}
	}
}

// GetPlayerSummary returns a formatted summary for a specific player
func (rs *RoundSummary) GetPlayerSummary(playerName string) string {
	info, exists := rs.PlayerActions[playerName]
	if !exists {
		return ""
	}

	var sb strings.Builder

	// Your attacks
	if len(info.AttacksMade) > 0 {
		sb.WriteString("**Your Attacks:**\n")
		for _, atk := range info.AttacksMade {
			icon := "❌"
			if atk.Hit {
				if atk.Critical {
					icon = "💥"
				} else {
					icon = "⚔️"
				}
			}
			sb.WriteString(fmt.Sprintf("%s %s → **%s** ", icon, atk.WeaponName, atk.TargetName))
			if atk.Hit {
				sb.WriteString(fmt.Sprintf("🩸 **%d** damage", atk.Damage))
			} else {
				sb.WriteString("**MISS**")
			}
			sb.WriteString("\n")
		}
	}

	// Attacks against you
	if len(info.AttacksRecvd) > 0 {
		sb.WriteString("\n**Attacks Against You:**\n")
		for _, atk := range info.AttacksRecvd {
			icon := "🛡️"
			if atk.Hit {
				icon = "🩸"
			}
			sb.WriteString(fmt.Sprintf("%s **%s** → You ", icon, atk.AttackerName))
			if atk.Hit {
				sb.WriteString(fmt.Sprintf("**%d** damage", atk.Damage))
			} else {
				sb.WriteString("**MISS**")
			}
			sb.WriteString("\n")
		}
	}

	// Summary
	sb.WriteString("\n**Round Summary:**\n")
	sb.WriteString(fmt.Sprintf("💔 Damage Taken: **%d**\n", info.DamageTaken))
	sb.WriteString(fmt.Sprintf("⚔️ Damage Dealt: **%d**\n", info.DamageDealt))

	return sb.String()
}

// GetRoundOverview returns a general overview of the round
func (rs *RoundSummary) GetRoundOverview() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("**Round %d Overview:**\n", rs.Round))

	for _, player := range rs.PlayerActions {
		if player.DamageDealt > 0 || player.DamageTaken > 0 {
			sb.WriteString(fmt.Sprintf("• **%s**: ", player.Name))
			if player.DamageDealt > 0 {
				sb.WriteString(fmt.Sprintf("dealt **%d** ", player.DamageDealt))
			}
			if player.DamageTaken > 0 {
				if player.DamageDealt > 0 {
					sb.WriteString("| ")
				}
				sb.WriteString(fmt.Sprintf("took **%d**", player.DamageTaken))
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}
