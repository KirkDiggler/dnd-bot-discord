package ability

import (
	"context"
	"fmt"
	"log"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/events"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e/features"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"github.com/KirkDiggler/dnd-bot-discord/internal/effects"
)

// rageExecutor implements the barbarian rage ability executor
type rageExecutor struct {
	service *service // Reference back to service for dependencies
}

func newRageExecutor(svc *service) Executor {
	return &rageExecutor{service: svc}
}

func (r *rageExecutor) Key() string {
	return "rage"
}

func (r *rageExecutor) Execute(ctx context.Context, char *character.Character, abilityDef *shared.ActiveAbility, input *UseAbilityInput) (*UseAbilityResult, error) {
	// Check if EventBus is available
	if r.service.eventBus == nil {
		// Fallback to old system if EventBus not available
		rageEffect := effects.BuildRageEffect(char.Level)
		err := char.AddStatusEffect(rageEffect)
		if err != nil {
			log.Printf("Failed to add rage effect: %v", err)
			abilityDef.UsesRemaining++ // Restore the use
			return &UseAbilityResult{
				Success:       false,
				Message:       "Failed to enter rage",
				UsesRemaining: abilityDef.UsesRemaining,
			}, nil
		}
		abilityDef.IsActive = true
		abilityDef.Duration = 10
		return &UseAbilityResult{
			Success:       true,
			Message:       fmt.Sprintf("You enter a rage! +%d damage to melee attacks, resistance to physical damage", 2+(char.Level-1)/8),
			EffectApplied: true,
			EffectID:      rageEffect.ID,
			EffectName:    rageEffect.Name,
			Duration:      10,
			UsesRemaining: abilityDef.UsesRemaining,
		}, nil
	}

	// Get current turn count from encounter
	currentTurn := 0
	if input.EncounterID != "" && r.service.encounterService != nil {
		if encounter, err := r.service.encounterService.GetEncounter(ctx, input.EncounterID); err == nil {
			// Calculate total turns (rounds * combatants + current turn index)
			if len(encounter.TurnOrder) > 0 {
				currentTurn = (encounter.Round-1)*len(encounter.TurnOrder) + encounter.Turn
			}
			log.Printf("=== RAGE ACTIVATION INFO ===")
			log.Printf("Encounter ID: %s", input.EncounterID)
			log.Printf("Current Round: %d, Turn index: %d, Combatants: %d", encounter.Round, encounter.Turn, len(encounter.TurnOrder))
			log.Printf("Starting at turn count: %d", currentTurn)
			log.Printf("Rage lasts 10 rounds (with %d combatants, that's %d turns)", len(encounter.TurnOrder), 10*len(encounter.TurnOrder))
			log.Printf("Rage will expire after turn: %d", currentTurn+(10*len(encounter.TurnOrder)))
		} else {
			log.Printf("Failed to get encounter for rage activation: %v", err)
		}
	} else {
		log.Printf("No encounter ID provided for rage activation - duration tracking may not work")
	}

	// Create rage listener for event system with current turn
	rageListener := features.NewRageListener(char.ID, char.Level, currentTurn)

	// Subscribe rage listener to relevant events
	r.service.eventBus.Subscribe(events.OnDamageRoll, rageListener)
	r.service.eventBus.Subscribe(events.BeforeTakeDamage, rageListener)

	// Track the active listener
	r.service.activeListeners[char.ID] = rageListener

	// Also add rage to the old effect system for UI display
	// This ensures the character sheet shows rage as active
	rageEffect := effects.BuildRageEffect(char.Level)
	// Update the duration to match what we're tracking
	rageEffect.Duration = effects.Duration{
		Type:   effects.DurationRounds,
		Rounds: 10,
	}
	if err := char.AddStatusEffect(rageEffect); err != nil {
		log.Printf("Failed to add rage status effect for UI: %v", err)
	}

	// Mark ability as active
	abilityDef.IsActive = true
	abilityDef.Duration = 10

	// Calculate damage bonus based on level
	damageBonus := "+2"
	if char.Level >= 16 {
		damageBonus = "+4"
	} else if char.Level >= 9 {
		damageBonus = "+3"
	}

	log.Printf("=== RAGE ACTIVATION (EVENT SYSTEM) ===")
	log.Printf("Character: %s (ID: %s)", char.Name, char.ID)
	log.Printf("Rage listener registered with ID: %s", rageListener.ID())
	log.Printf("Rage uses remaining: %d/%d", abilityDef.UsesRemaining, abilityDef.UsesMax)
	log.Printf("Rage is active: %v", abilityDef.IsActive)
	log.Printf("Event listeners registered for OnDamageRoll and BeforeTakeDamage")

	return &UseAbilityResult{
		Success:       true,
		Message:       fmt.Sprintf("You enter a rage! %s damage to melee attacks, resistance to physical damage", damageBonus),
		EffectApplied: true,
		EffectID:      rageListener.ID(),
		EffectName:    "Rage",
		Duration:      10,
		UsesRemaining: abilityDef.UsesRemaining,
	}, nil
}
