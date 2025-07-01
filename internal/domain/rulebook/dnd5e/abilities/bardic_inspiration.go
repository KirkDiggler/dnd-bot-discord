package abilities

import (
	"context"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
)

// BardicInspirationHandler implements the bard's bardic inspiration ability
type BardicInspirationHandler struct{}

// NewBardicInspirationHandler creates a new bardic inspiration handler
func NewBardicInspirationHandler() *BardicInspirationHandler {
	return &BardicInspirationHandler{}
}

// Key returns the ability key
func (b *BardicInspirationHandler) Key() string {
	return "bardic-inspiration"
}

// Execute grants bardic inspiration to a target
func (b *BardicInspirationHandler) Execute(ctx context.Context, char *character.Character, ability *shared.ActiveAbility, input *Input) (*Result, error) {
	if input.TargetID == "" {
		// Restore the use since we didn't actually use it
		ability.UsesRemaining++
		return &Result{
			Success:       false,
			Message:       "Bardic Inspiration requires a target",
			UsesRemaining: ability.UsesRemaining,
		}, nil
	}

	// Create inspiration effect (would need to be tracked on target)
	return &Result{
		Success:       true,
		Message:       "You inspire your ally with a d6 Bardic Inspiration die",
		EffectApplied: true,
		EffectName:    "Bardic Inspiration (d6)",
		Duration:      ability.Duration,
		UsesRemaining: ability.UsesRemaining,
	}, nil
}
