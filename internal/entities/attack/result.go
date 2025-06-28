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
}

func (r *Result) String() string {
	return fmt.Sprintf("attack: %d, type: %s, damage: %d", r.AttackRoll, r.AttackType, r.DamageRoll)
}

func RollAttack(attackBonus, damageBonus int, dmg *damage.Damage) (*Result, error) {
	attackResult, err := dice.Roll(1, 20, 0)
	if err != nil {
		return nil, err
	}

	dmgResult, err := dice.Roll(dmg.DiceCount, dmg.DiceSize, 0)
	if err != nil {
		return nil, err
	}
	dmgValue := dmgResult.Total
	attackRoll := attackResult.Total
	allRolls := make([]int, 0, len(dmgResult.Rolls))
	allRolls = append(allRolls, dmgResult.Rolls...)

	switch attackResult.Total {
	case 20:
		critDmg, err := dice.Roll(dmg.DiceCount, dmg.DiceSize, 0)
		if err != nil {
			return nil, err
		}

		dmgValue += critDmg.Total
		attackRoll += attackBonus
		// Add critical damage rolls
		allRolls = append(allRolls, critDmg.Rolls...)
	case 1:
		attackRoll = 0
	default:
		attackRoll += attackBonus

	}

	return &Result{
		AttackRoll:     attackRoll,
		AttackType:     dmg.DamageType,
		DamageRoll:     damageBonus + dmgValue,
		AttackResult:   attackResult,
		DamageResult:   dmgResult,
		WeaponDamage:   dmg,
		AllDamageRolls: allRolls,
	}, nil
}
