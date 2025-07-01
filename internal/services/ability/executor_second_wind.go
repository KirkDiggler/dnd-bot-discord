package ability

import (
	"context"
	"fmt"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
)

// secondWindExecutor implements the fighter's second wind ability
type secondWindExecutor struct {
	service *service
}

func newSecondWindExecutor(svc *service) Executor {
	return &secondWindExecutor{service: svc}
}

func (s *secondWindExecutor) Key() string {
	return "second-wind"
}

func (s *secondWindExecutor) Execute(ctx context.Context, char *character.Character, abilityDef *shared.ActiveAbility, input *UseAbilityInput) (*UseAbilityResult, error) {
	// Roll 1d10 + fighter level
	rollResult, err := s.service.diceRoller.Roll(1, 10, char.Level)
	if err != nil {
		abilityDef.UsesRemaining++ // Restore the use
		return &UseAbilityResult{
			Success:       false,
			Message:       "Failed to roll healing",
			UsesRemaining: abilityDef.UsesRemaining,
		}, nil
	}

	resources := char.GetResources()
	healingDone := resources.HP.Heal(rollResult.Total)
	char.CurrentHitPoints = resources.HP.Current

	return &UseAbilityResult{
		Success:       true,
		Message:       fmt.Sprintf("Second Wind heals you for %d HP (rolled %d)", healingDone, rollResult.Total),
		HealingDone:   healingDone,
		TargetNewHP:   resources.HP.Current,
		UsesRemaining: abilityDef.UsesRemaining,
	}, nil
}
