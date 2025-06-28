package entities

import (
	"log"

	"github.com/KirkDiggler/dnd-bot-discord/internal/dice"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities/attack"
)

// CanSneakAttack checks if the character can use sneak attack with given conditions
func (c *Character) CanSneakAttack(weapon *Weapon, hasAdvantage, allyAdjacent, hasDisadvantage bool) bool {
	// Must be a rogue
	if c.Class == nil || c.Class.Key != "rogue" {
		return false
	}

	// Can't sneak attack if advantage and disadvantage cancel out
	if hasAdvantage && hasDisadvantage {
		return false
	}

	// Check weapon eligibility
	weaponEligible := false

	// Ranged weapons are always eligible
	if weapon.WeaponRange == "Ranged" {
		weaponEligible = true
	} else {
		// Melee weapons must have finesse property
		for _, prop := range weapon.Properties {
			if prop != nil && prop.Key == "finesse" {
				weaponEligible = true
				break
			}
		}
	}

	if !weaponEligible {
		return false
	}

	// Must have advantage OR an ally adjacent to target
	return hasAdvantage || allyAdjacent
}

// GetSneakAttackDice returns the number of d6 dice for sneak attack based on rogue level
func (c *Character) GetSneakAttackDice() int {
	if c.Class == nil || c.Class.Key != "rogue" {
		return 0
	}

	// Sneak attack damage: 1d6 per 2 rogue levels (rounded up)
	// Level 1-2: 1d6
	// Level 3-4: 2d6
	// Level 5-6: 3d6
	// etc...
	return (c.Level + 1) / 2
}

// ApplySneakAttack applies sneak attack damage if eligible
func (c *Character) ApplySneakAttack(ctx *CombatContext) int {
	// Check if already used this turn
	if c.Resources == nil {
		return 0
	}

	if c.Resources.SneakAttackUsedThisTurn {
		return 0
	}

	// Get number of dice
	diceCount := c.GetSneakAttackDice()
	if diceCount == 0 {
		return 0
	}

	// Double dice on critical
	if ctx.IsCritical {
		diceCount *= 2
	}

	// Roll sneak attack damage
	result, err := dice.Roll(diceCount, 6, 0)
	if err != nil {
		log.Printf("Error rolling sneak attack damage dice: %v", err)
		return 0
	}

	// Mark as used this turn
	c.Resources.SneakAttackUsedThisTurn = true

	return result.Total
}

// StartNewTurn resets per-turn resources
func (c *Character) StartNewTurn() {
	if c.Resources != nil {
		c.Resources.SneakAttackUsedThisTurn = false
	}
}

// CombatContext provides context for combat calculations
type CombatContext struct {
	AttackResult *attack.Result
	IsCritical   bool
}
