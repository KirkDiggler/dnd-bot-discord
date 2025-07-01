package abilities

import (
	"context"
	"fmt"

	"github.com/KirkDiggler/dnd-bot-discord/internal/dice"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
)

// SecondWindHandler implements the fighter's second wind ability
type SecondWindHandler struct {
	diceRoller dice.Roller
}

// NewSecondWindHandler creates a new second wind handler
func NewSecondWindHandler(diceRoller dice.Roller) *SecondWindHandler {
	return &SecondWindHandler{
		diceRoller: diceRoller,
	}
}

// Key returns the ability key
func (s *SecondWindHandler) Key() string {
	return "second-wind"
}

// Execute performs the second wind healing
func (s *SecondWindHandler) Execute(ctx context.Context, char *character.Character, ability *shared.ActiveAbility, input *Input) (*Result, error) {
	result := &Result{
		Success:       true,
		UsesRemaining: ability.UsesRemaining,
	}

	// Roll 1d10 + fighter level
	rollResult, err := s.diceRoller.Roll(1, 10, char.Level)
	if err != nil {
		ability.UsesRemaining++ // Restore the use
		result.Success = false
		result.Message = "Failed to roll healing"
		return result, nil
	}

	resources := char.GetResources()
	healingDone := resources.HP.Heal(rollResult.Total)
	char.CurrentHitPoints = resources.HP.Current

	result.Message = fmt.Sprintf("Second Wind heals you for %d HP (rolled %d)", healingDone, rollResult.Total)
	result.HealingDone = healingDone
	result.TargetNewHP = resources.HP.Current

	return result, nil
}
