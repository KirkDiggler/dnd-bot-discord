package abilities

import (
	"context"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	abilityService "github.com/KirkDiggler/dnd-bot-discord/internal/services/ability"
)

// ServiceHandlerAdapter adapts a D&D 5e ability handler to the service's Handler interface
type ServiceHandlerAdapter struct {
	handler Handler
}

// NewServiceHandlerAdapter creates a new adapter
func NewServiceHandlerAdapter(handler Handler) abilityService.Handler {
	return &ServiceHandlerAdapter{handler: handler}
}

// Key returns the ability key
func (a *ServiceHandlerAdapter) Key() string {
	return a.handler.Key()
}

// Execute adapts the execution between interfaces
func (a *ServiceHandlerAdapter) Execute(ctx context.Context, char *character.Character, ability *shared.ActiveAbility, input *abilityService.HandlerInput) (*abilityService.HandlerResult, error) {
	// Convert service input to D&D 5e input
	dndInput := &Input{
		CharacterID: input.CharacterID,
		AbilityKey:  input.AbilityKey,
		TargetID:    input.TargetID,
		Value:       input.Value,
		EncounterID: input.EncounterID,
	}

	// Execute using D&D 5e handler
	dndResult, err := a.handler.Execute(ctx, char, ability, dndInput)
	if err != nil {
		return nil, err
	}

	// Convert D&D 5e result to service result
	return &abilityService.HandlerResult{
		Success:       dndResult.Success,
		Message:       dndResult.Message,
		UsesRemaining: dndResult.UsesRemaining,
		EffectApplied: dndResult.EffectApplied,
		EffectID:      dndResult.EffectID,
		EffectName:    dndResult.EffectName,
		Duration:      dndResult.Duration,
		HealingDone:   dndResult.HealingDone,
		DamageBonus:   dndResult.DamageBonus,
		TargetNewHP:   dndResult.TargetNewHP,
	}, nil
}
