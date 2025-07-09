package character

import (
	"context"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
)

// buildWizardSteps creates wizard-specific steps
func (b *FlowBuilderImpl) buildWizardSteps(ctx context.Context, char *character.Character) []character.CreationStep {
	var steps []character.CreationStep

	// For now, return empty until we add spell selection step types
	// TODO: Add spell selection when StepTypeSpellSelection is added to domain

	// Future steps would include:
	// 1. Cantrip selection (3 cantrips at level 1)
	// 2. Spell selection (6 1st-level spells for spellbook)
	// 3. Arcane tradition selection (at level 2)

	return steps
}
