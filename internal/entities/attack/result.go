package attack

import (
	"fmt"

	"github.com/KirkDiggler/dnd-bot-discord/internal/dice"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities/damage"
)

type Result struct {
	AttackRoll   int
	AttackType   damage.Type
	DamageRoll   int
	AttackResult *dice.RollResult
	DamageResult *dice.RollResult
	// Weapon damage info for proper display
	WeaponDamage *damage.Damage
	// All damage dice rolls (including crits)
	AllDamageRolls []int
	// Reroll information for Great Weapon Fighting display
	RerollInfo []DieReroll
	// Weapon key used for this attack (for action economy tracking)
	WeaponKey string
}

// DieReroll tracks information about rerolled dice for display
type DieReroll struct {
	OriginalRoll int // The original roll that was rerolled
	NewRoll      int // The new roll after rerolling
	Position     int // Position in the damage roll sequence (0-based)
}

func (r *Result) String() string {
	return fmt.Sprintf("attack: %d, type: %s, damage: %d", r.AttackRoll, r.AttackType, r.DamageRoll)
}

// RollAttackWithFightingStyle rolls an attack with fighting style modifications
func RollAttackWithFightingStyle(roller dice.Roller, attackBonus, damageBonus int, dmg *damage.Damage, fightingStyle string) (*Result, error) {
	attackResult, err := roller.Roll(1, 20, 0)
	if err != nil {
		return nil, err
	}

	dmgValue := 0
	attackRoll := attackResult.Total
	allRolls := make([]int, 0, dmg.DiceCount*2) // Extra space for potential crits
	var rerollInfo []DieReroll

	// Roll damage dice with potential rerolls for Great Weapon Fighting
	if fightingStyle == "great_weapon" {
		dmgValue, allRolls, rerollInfo = rollDamageWithGreatWeaponFighting(roller, dmg.DiceCount, dmg.DiceSize)
	} else {
		// Standard damage roll
		dmgResult, err := roller.Roll(dmg.DiceCount, dmg.DiceSize, 0)
		if err != nil {
			return nil, err
		}
		dmgValue = dmgResult.Total
		allRolls = append(allRolls, dmgResult.Rolls...)
	}

	// Always add attack bonus to the roll
	attackRoll += attackBonus

	// Handle critical hit (natural 20)
	if attackResult.IsCrit {
		if fightingStyle == "great_weapon" {
			critValue, critRolls, critRerolls := rollDamageWithGreatWeaponFighting(roller, dmg.DiceCount, dmg.DiceSize)
			dmgValue += critValue
			allRolls = append(allRolls, critRolls...)
			// Adjust reroll positions for crit dice
			for _, reroll := range critRerolls {
				reroll.Position += len(allRolls) - len(critRolls)
				rerollInfo = append(rerollInfo, reroll)
			}
		} else {
			critResult, err := roller.Roll(dmg.DiceCount, dmg.DiceSize, 0)
			if err != nil {
				return nil, err
			}
			dmgValue += critResult.Total
			allRolls = append(allRolls, critResult.Rolls...)
		}
	}

	return &Result{
		AttackRoll:     attackRoll,
		AttackType:     dmg.DamageType,
		DamageRoll:     damageBonus + dmgValue,
		AttackResult:   attackResult,
		DamageResult:   &dice.RollResult{Total: dmgValue, Rolls: allRolls},
		WeaponDamage:   dmg,
		AllDamageRolls: allRolls,
		RerollInfo:     rerollInfo,
	}, nil
}

// rollDamageWithGreatWeaponFighting handles rerolling 1s and 2s once per die
func rollDamageWithGreatWeaponFighting(roller dice.Roller, diceCount, diceSize int) (totalDamage int, finalRolls []int, rerollInfo []DieReroll) {
	// First, roll all dice at once to get the initial results
	initialResult, err := roller.Roll(diceCount, diceSize, 0)
	if err != nil {
		// Fallback - return minimum damage if rolling fails
		finalRolls = make([]int, diceCount)
		for i := range finalRolls {
			finalRolls[i] = 1
		}
		totalDamage = diceCount
		return
	}

	// Process each die from the initial roll
	for i, roll := range initialResult.Rolls {
		// Check if we need to reroll (1 or 2 on any die)
		if roll <= 2 {
			// Reroll this die once
			rerollResult, err := roller.Roll(1, diceSize, 0)
			if err != nil {
				// If reroll fails, keep original
				finalRolls = append(finalRolls, roll)
				totalDamage += roll
				continue
			}

			newRoll := rerollResult.Rolls[0]
			finalRolls = append(finalRolls, newRoll)
			totalDamage += newRoll

			// Track the reroll for display
			rerollInfo = append(rerollInfo, DieReroll{
				OriginalRoll: roll,
				NewRoll:      newRoll,
				Position:     i,
			})
		} else {
			// Keep original roll
			finalRolls = append(finalRolls, roll)
			totalDamage += roll
		}
	}

	return
}

func RollAttack(roller dice.Roller, attackBonus, damageBonus int, dmg *damage.Damage) (*Result, error) {
	// Call the new function with no fighting style
	return RollAttackWithFightingStyle(roller, attackBonus, damageBonus, dmg, "")
}
