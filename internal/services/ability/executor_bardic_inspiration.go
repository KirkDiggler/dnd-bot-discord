package ability

import (
	"context"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
)

// bardicInspirationExecutor implements the bard's bardic inspiration ability
type bardicInspirationExecutor struct{}

func newBardicInspirationExecutor() Executor {
	return &bardicInspirationExecutor{}
}

func (b *bardicInspirationExecutor) Key() string {
	return "bardic-inspiration"
}

func (b *bardicInspirationExecutor) Execute(ctx context.Context, char *character.Character, abilityDef *shared.ActiveAbility, input *UseAbilityInput) (*UseAbilityResult, error) {
	if input.TargetID == "" {
		// Restore the use since we didn't actually use it
		abilityDef.UsesRemaining++
		return &UseAbilityResult{
			Success:       false,
			Message:       "Bardic Inspiration requires a target",
			UsesRemaining: abilityDef.UsesRemaining,
		}, nil
	}

	// Create inspiration effect (would need to be tracked on target)
	return &UseAbilityResult{
		Success:       true,
		Message:       "You inspire your ally with a d6 Bardic Inspiration die",
		EffectApplied: true,
		EffectName:    "Bardic Inspiration (d6)",
		Duration:      abilityDef.Duration,
		UsesRemaining: abilityDef.UsesRemaining,
	}, nil
}
