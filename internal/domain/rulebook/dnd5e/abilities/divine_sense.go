package abilities

import (
	"context"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
)

// DivineSenseHandler implements the paladin's divine sense ability
type DivineSenseHandler struct{}

// NewDivineSenseHandler creates a new divine sense handler
func NewDivineSenseHandler() *DivineSenseHandler {
	return &DivineSenseHandler{}
}

// Key returns the ability key
func (d *DivineSenseHandler) Key() string {
	return "divine-sense"
}

// Execute activates divine sense
func (d *DivineSenseHandler) Execute(ctx context.Context, char *character.Character, ability *shared.ActiveAbility, input *Input) (*Result, error) {
	return &Result{
		Success:       true,
		Message:       "You open your awareness to detect celestials, fiends, and undead within 60 feet",
		EffectApplied: true,
		EffectName:    "Divine Sense",
		Duration:      1, // Until end of next turn
		UsesRemaining: ability.UsesRemaining,
	}, nil
}
