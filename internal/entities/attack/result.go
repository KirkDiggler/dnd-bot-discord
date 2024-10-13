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
	switch attackResult.Total {
	case 20:
		critDmg, err := dice.Roll(dmg.DiceCount, dmg.DiceSize, 0)
		if err != nil {
			return nil, err
		}

		dmgValue = dmgValue + critDmg.Total
		attackRoll = attackRoll + attackBonus
	case 1:
		attackRoll = 0
	default:
		attackRoll = attackRoll + attackBonus

	}

	return &Result{
		AttackRoll:   attackRoll,
		AttackType:   dmg.DamageType,
		DamageRoll:   damageBonus + dmgValue,
		AttackResult: attackResult,
		DamageResult: dmgResult,
	}, nil
}
