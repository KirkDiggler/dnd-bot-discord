package abilities

import (
	"context"
	"fmt"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/events"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e/spells"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
)

// ViciousMockeryHandler wraps the spell handler to work as an ability
type ViciousMockeryHandler struct {
	spellHandler *spells.ViciousMockeryHandler
}

// NewViciousMockeryHandler creates a new vicious mockery ability handler
func NewViciousMockeryHandler(eventBus events.Bus) *ViciousMockeryHandler {
	return &ViciousMockeryHandler{
		spellHandler: spells.NewViciousMockeryHandler(eventBus),
	}
}

// SetDiceRoller sets the dice roller dependency
func (v *ViciousMockeryHandler) SetDiceRoller(roller interface{}) {
	v.spellHandler.SetDiceRoller(roller)
}

// SetCharacterService sets the character service dependency
func (v *ViciousMockeryHandler) SetCharacterService(service interface{}) {
	v.spellHandler.SetCharacterService(service)
}

// Key returns the ability key
func (v *ViciousMockeryHandler) Key() string {
	return shared.AbilityKeyViciousMockery
}

// Execute casts vicious mockery as an ability
func (v *ViciousMockeryHandler) Execute(ctx context.Context, char *character.Character, ability *shared.ActiveAbility, input *Input) (*Result, error) {
	// Check if character is a bard
	if char.Class == nil || char.Class.Key != "bard" {
		return &Result{
			Success: false,
			Message: "Only bards can use Vicious Mockery",
		}, nil
	}

	// Convert ability input to spell input
	spellInput := &spells.SpellInput{
		SpellLevel:  0, // Cantrip
		TargetIDs:   []string{},
		EncounterID: input.EncounterID,
	}

	// Add target if provided
	if input.TargetID != "" {
		spellInput.TargetIDs = append(spellInput.TargetIDs, input.TargetID)
	}

	// Execute the spell
	spellResult, err := v.spellHandler.Execute(ctx, char, spellInput)
	if err != nil {
		return nil, fmt.Errorf("failed to cast vicious mockery: %w", err)
	}

	// Convert spell result to ability result
	result := &Result{
		Success:       spellResult.Success,
		Message:       spellResult.Message,
		EffectApplied: spellResult.Success && spellResult.TotalDamage > 0,
		EffectName:    "Vicious Mockery",
	}

	// Cantrips don't use resources
	result.UsesRemaining = -1 // Unlimited

	return result, nil
}
