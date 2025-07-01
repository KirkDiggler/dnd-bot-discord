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

// TODO: This file contains hardcoded D&D 5e ability handlers that should be moved to the ruleset
// This is a temporary solution to avoid circular dependencies while refactoring

// handleRage handles the Barbarian's Rage ability
func (s *service) handleRage(char *character.Character, ability *shared.ActiveAbility, result *UseAbilityResult, encounterID string) *UseAbilityResult {
	// Check if EventBus is available
	if s.eventBus == nil {
		// Fallback to old system if EventBus not available
		rageEffect := effects.BuildRageEffect(char.Level)
		err := char.AddStatusEffect(rageEffect)
		if err != nil {
			log.Printf("Failed to add rage effect: %v", err)
			result.Success = false
			result.Message = "Failed to enter rage"
			ability.UsesRemaining++
			return result
		}
		ability.IsActive = true
		ability.Duration = 10
		result.Message = fmt.Sprintf("You enter a rage! +%d damage to melee attacks, resistance to physical damage", 2+(char.Level-1)/8)
		result.EffectApplied = true
		result.EffectID = rageEffect.ID
		result.EffectName = rageEffect.Name
		result.Duration = 10
		return result
	}

	// Get current turn count from encounter
	currentTurn := 0
	if encounterID != "" && s.encounterService != nil {
		if encounter, err := s.encounterService.GetEncounter(context.Background(), encounterID); err == nil {
			// Calculate total turns (rounds * combatants + current turn index)
			if len(encounter.TurnOrder) > 0 {
				currentTurn = (encounter.Round-1)*len(encounter.TurnOrder) + encounter.Turn
			}
			log.Printf("=== RAGE ACTIVATION INFO ===")
			log.Printf("Encounter ID: %s", encounterID)
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
	s.eventBus.Subscribe(events.OnDamageRoll, rageListener)
	s.eventBus.Subscribe(events.BeforeTakeDamage, rageListener)

	// Track the active listener
	s.activeListeners[char.ID] = rageListener

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
	ability.IsActive = true
	ability.Duration = 10

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
	log.Printf("Rage uses remaining: %d/%d", ability.UsesRemaining, ability.UsesMax)
	log.Printf("Rage is active: %v", ability.IsActive)
	log.Printf("Event listeners registered for OnDamageRoll and BeforeTakeDamage")

	result.Message = fmt.Sprintf("You enter a rage! %s damage to melee attacks, resistance to physical damage", damageBonus)
	result.EffectApplied = true
	result.EffectID = rageListener.ID()
	result.EffectName = "Rage"
	result.Duration = 10

	return result
}

// handleSecondWind handles the Fighter's Second Wind ability
func (s *service) handleSecondWind(char *character.Character, result *UseAbilityResult) *UseAbilityResult {
	// Roll 1d10 + fighter level
	rollResult, err := s.diceRoller.Roll(1, 10, char.Level)
	if err != nil {
		result.Success = false
		result.Message = "Failed to roll healing"
		return result
	}

	resources := char.GetResources()
	healingDone := resources.HP.Heal(rollResult.Total)
	char.CurrentHitPoints = resources.HP.Current

	result.Message = fmt.Sprintf("Second Wind heals you for %d HP (rolled %d)", healingDone, rollResult.Total)
	result.HealingDone = healingDone
	result.TargetNewHP = resources.HP.Current

	return result
}

// handleBardicInspiration handles the Bard's Bardic Inspiration ability
func (s *service) handleBardicInspiration(char *character.Character, targetID string, ability *shared.ActiveAbility, result *UseAbilityResult) *UseAbilityResult {
	if targetID == "" {
		result.Success = false
		result.Message = "Bardic Inspiration requires a target"
		// Restore the use since we didn't actually use it
		ability.UsesRemaining++
		return result
	}

	// Create inspiration effect (would need to be tracked on target)
	result.Message = "You inspire your ally with a d6 Bardic Inspiration die"
	result.EffectApplied = true
	result.EffectName = "Bardic Inspiration (d6)"
	result.Duration = ability.Duration

	return result
}

// handleLayOnHands handles the Paladin's Lay on Hands ability
func (s *service) handleLayOnHands(char *character.Character, targetID string, healAmount int, result *UseAbilityResult) *UseAbilityResult {
	resources := char.GetResources()
	ability := resources.Abilities["lay-on-hands"]

	// Validate heal amount
	if healAmount <= 0 {
		result.Success = false
		result.Message = "Invalid heal amount"
		return result
	}

	if healAmount > ability.UsesRemaining {
		result.Success = false
		result.Message = fmt.Sprintf("Not enough healing pool remaining (%d/%d)", ability.UsesRemaining, ability.UsesMax)
		return result
	}

	// For now, assume self-healing if no target specified
	if targetID == "" || targetID == char.ID {
		// Self heal
		healingDone := resources.HP.Heal(healAmount)
		char.CurrentHitPoints = resources.HP.Current

		// Deduct from pool
		ability.UsesRemaining -= healAmount

		result.Message = fmt.Sprintf("Lay on Hands heals you for %d HP (%d points remaining)", healingDone, ability.UsesRemaining)
		result.HealingDone = healingDone
		result.TargetNewHP = resources.HP.Current
	} else {
		// Healing another target would require encounter context
		ability.UsesRemaining -= healAmount
		result.Message = fmt.Sprintf("Lay on Hands heals target for %d HP (%d points remaining)", healAmount, ability.UsesRemaining)
		result.HealingDone = healAmount
	}

	result.UsesRemaining = ability.UsesRemaining

	return result
}

// handleDivineSense handles the Paladin's Divine Sense ability
func (s *service) handleDivineSense(char *character.Character, result *UseAbilityResult) *UseAbilityResult {
	result.Message = "You open your awareness to detect celestials, fiends, and undead within 60 feet"
	result.EffectApplied = true
	result.EffectName = "Divine Sense"
	result.Duration = 1 // Until end of next turn

	return result
}
