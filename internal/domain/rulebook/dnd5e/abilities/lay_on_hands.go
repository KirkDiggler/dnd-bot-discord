package abilities

import (
	"context"
	"fmt"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
)

// LayOnHandsHandler implements the paladin's lay on hands ability
type LayOnHandsHandler struct{}

// NewLayOnHandsHandler creates a new lay on hands handler
func NewLayOnHandsHandler() *LayOnHandsHandler {
	return &LayOnHandsHandler{}
}

// Key returns the ability key
func (l *LayOnHandsHandler) Key() string {
	return "lay-on-hands"
}

// Execute performs lay on hands healing
func (l *LayOnHandsHandler) Execute(ctx context.Context, char *character.Character, ability *shared.ActiveAbility, input *Input) (*Result, error) {
	resources := char.GetResources()

	// Validate heal amount
	if input.Value <= 0 {
		return &Result{
			Success:       false,
			Message:       "Invalid heal amount",
			UsesRemaining: ability.UsesRemaining,
		}, nil
	}

	if input.Value > ability.UsesRemaining {
		return &Result{
			Success:       false,
			Message:       fmt.Sprintf("Not enough healing pool remaining (%d/%d)", ability.UsesRemaining, ability.UsesMax),
			UsesRemaining: ability.UsesRemaining,
		}, nil
	}

	// For now, assume self-healing if no target specified
	if input.TargetID == "" || input.TargetID == char.ID {
		// Self heal
		healingDone := resources.HP.Heal(input.Value)
		char.CurrentHitPoints = resources.HP.Current

		// Deduct from pool
		ability.UsesRemaining -= input.Value

		return &Result{
			Success:       true,
			Message:       fmt.Sprintf("Lay on Hands heals you for %d HP (%d points remaining)", healingDone, ability.UsesRemaining),
			HealingDone:   healingDone,
			TargetNewHP:   resources.HP.Current,
			UsesRemaining: ability.UsesRemaining,
		}, nil
	}

	// Healing another target would require encounter context
	ability.UsesRemaining -= input.Value
	return &Result{
		Success:       true,
		Message:       fmt.Sprintf("Lay on Hands heals target for %d HP (%d points remaining)", input.Value, ability.UsesRemaining),
		HealingDone:   input.Value,
		UsesRemaining: ability.UsesRemaining,
	}, nil
}
