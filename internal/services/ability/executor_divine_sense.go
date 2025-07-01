package ability

import (
	"context"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
)

// divineSenseExecutor implements the paladin's divine sense ability
type divineSenseExecutor struct{}

func newDivineSenseExecutor() Executor {
	return &divineSenseExecutor{}
}

func (d *divineSenseExecutor) Key() string {
	return "divine-sense"
}

func (d *divineSenseExecutor) Execute(ctx context.Context, char *character.Character, abilityDef *shared.ActiveAbility, input *UseAbilityInput) (*UseAbilityResult, error) {
	return &UseAbilityResult{
		Success:       true,
		Message:       "You open your awareness to detect celestials, fiends, and undead within 60 feet",
		EffectApplied: true,
		EffectName:    "Divine Sense",
		Duration:      1, // Until end of next turn
		UsesRemaining: abilityDef.UsesRemaining,
	}, nil
}
