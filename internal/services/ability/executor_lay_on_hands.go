package ability

import (
	"context"
	"fmt"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
)

// layOnHandsExecutor implements the paladin's lay on hands ability
type layOnHandsExecutor struct {
	service *service
}

func newLayOnHandsExecutor(svc *service) Executor {
	return &layOnHandsExecutor{service: svc}
}

func (l *layOnHandsExecutor) Key() string {
	return "lay-on-hands"
}

func (l *layOnHandsExecutor) Execute(ctx context.Context, char *character.Character, abilityDef *shared.ActiveAbility, input *UseAbilityInput) (*UseAbilityResult, error) {
	resources := char.GetResources()

	// Validate heal amount
	if input.Value <= 0 {
		return &UseAbilityResult{
			Success:       false,
			Message:       "Invalid heal amount",
			UsesRemaining: abilityDef.UsesRemaining,
		}, nil
	}

	if input.Value > abilityDef.UsesRemaining {
		return &UseAbilityResult{
			Success:       false,
			Message:       fmt.Sprintf("Not enough healing pool remaining (%d/%d)", abilityDef.UsesRemaining, abilityDef.UsesMax),
			UsesRemaining: abilityDef.UsesRemaining,
		}, nil
	}

	// For now, assume self-healing if no target specified
	if input.TargetID == "" || input.TargetID == char.ID {
		// Self heal
		healingDone := resources.HP.Heal(input.Value)
		char.CurrentHitPoints = resources.HP.Current

		// Deduct from pool - DON'T use the standard Use() method
		abilityDef.UsesRemaining -= input.Value

		return &UseAbilityResult{
			Success:       true,
			Message:       fmt.Sprintf("Lay on Hands heals you for %d HP (%d points remaining)", healingDone, abilityDef.UsesRemaining),
			HealingDone:   healingDone,
			TargetNewHP:   resources.HP.Current,
			UsesRemaining: abilityDef.UsesRemaining,
		}, nil
	} else {
		// Healing another target would require encounter context
		abilityDef.UsesRemaining -= input.Value
		return &UseAbilityResult{
			Success:       true,
			Message:       fmt.Sprintf("Lay on Hands heals target for %d HP (%d points remaining)", input.Value, abilityDef.UsesRemaining),
			HealingDone:   input.Value,
			UsesRemaining: abilityDef.UsesRemaining,
		}, nil
	}
}
