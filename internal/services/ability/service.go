package ability

import (
	"context"
	"fmt"
	"log"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/events"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e/features"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"

	"github.com/KirkDiggler/dnd-bot-discord/internal/dice"
	"github.com/KirkDiggler/dnd-bot-discord/internal/effects"
	dnderr "github.com/KirkDiggler/dnd-bot-discord/internal/errors"
	charService "github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	encounterService "github.com/KirkDiggler/dnd-bot-discord/internal/services/encounter"
)

type service struct {
	characterService charService.Service
	encounterService encounterService.Service
	diceRoller       dice.Roller
	eventBus         *events.EventBus
	activeListeners  map[string]events.EventListener // Track active listeners by character ID
}

// ServiceConfig holds configuration for the ability service
type ServiceConfig struct {
	CharacterService charService.Service
	EncounterService encounterService.Service
	DiceRoller       dice.Roller
	EventBus         *events.EventBus
}

// NewService creates a new ability service
func NewService(cfg *ServiceConfig) Service {
	if cfg.CharacterService == nil {
		panic("character service is required")
	}

	svc := &service{
		characterService: cfg.CharacterService,
		encounterService: cfg.EncounterService,
		diceRoller:       cfg.DiceRoller,
		eventBus:         cfg.EventBus,
		activeListeners:  make(map[string]events.EventListener),
	}

	if svc.diceRoller == nil {
		svc.diceRoller = dice.NewRandomRoller()
	}

	// Subscribe to turn start events to update effect durations
	if svc.eventBus != nil {
		svc.eventBus.Subscribe(events.OnTurnStart, &turnStartListener{service: svc})
	}

	return svc
}

// UseAbility executes an ability
func (s *service) UseAbility(ctx context.Context, input *UseAbilityInput) (*UseAbilityResult, error) {
	if input == nil {
		return nil, dnderr.InvalidArgument("input cannot be nil")
	}

	// Get the character
	char, err := s.characterService.GetByID(input.CharacterID)
	if err != nil {
		return nil, dnderr.Wrap(err, "failed to get character")
	}

	// Get character resources
	resources := char.GetResources()
	if resources == nil {
		return nil, dnderr.InvalidArgument("character has no resources")
	}

	// Get the ability
	ability, exists := resources.Abilities[input.AbilityKey]
	if !exists {
		return nil, dnderr.NotFound("ability not found")
	}

	// Check if ability can be used
	if !ability.CanUse() {
		return &UseAbilityResult{
			Success:       false,
			Message:       "No uses remaining",
			UsesRemaining: ability.UsesRemaining,
		}, nil
	}

	// Special handling for lay on hands which uses a pool
	if input.AbilityKey == "lay-on-hands" {
		// Don't use the standard Use() method for lay on hands
		result := &UseAbilityResult{
			Success:       true,
			UsesRemaining: ability.UsesRemaining,
		}
		return s.handleLayOnHands(char, input.TargetID, input.Value, result), nil
	}

	// Use the ability
	if !ability.Use() {
		return &UseAbilityResult{
			Success:       false,
			Message:       "Failed to use ability",
			UsesRemaining: ability.UsesRemaining,
		}, nil
	}

	// Update action economy based on ability type
	if char.Resources != nil {
		switch ability.ActionType {
		case shared.AbilityTypeAction:
			char.Resources.ActionEconomy.ActionUsed = true
			char.Resources.ActionEconomy.RecordAction("action", "ability", ability.Key)
		case shared.AbilityTypeBonusAction:
			char.Resources.ActionEconomy.BonusActionUsed = true
			char.Resources.ActionEconomy.RecordAction("bonus_action", "ability", ability.Key)
		case shared.AbilityTypeReaction:
			char.Resources.ActionEconomy.ReactionUsed = true
			char.Resources.ActionEconomy.RecordAction("reaction", "ability", ability.Key)
		case shared.AbilityTypeFree:
			// Free actions don't consume any action economy resources
			char.Resources.ActionEconomy.RecordAction("free", "ability", ability.Key)
		default:
			// Log unexpected action types for debugging
			log.Printf("Unexpected ability action type: %s for ability %s", ability.ActionType, ability.Key)
		}
	}

	// Handle ability effects based on key
	result := &UseAbilityResult{
		Success:       true,
		UsesRemaining: ability.UsesRemaining,
	}

	switch input.AbilityKey {
	case "rage":
		result = s.handleRage(char, ability, result, input.EncounterID)
	case "second-wind":
		result = s.handleSecondWind(char, result)
	case "bardic-inspiration":
		result = s.handleBardicInspiration(char, input.TargetID, ability, result)
	case "divine-sense":
		result = s.handleDivineSense(char, result)
	default:
		result.Message = fmt.Sprintf("Used %s", ability.Name)
	}

	// Save character state
	log.Printf("=== SAVING CHARACTER AFTER ABILITY USE ===")
	log.Printf("Character: %s", char.Name)
	if char.Resources != nil {
		log.Printf("Active effects before save: %d", len(char.Resources.ActiveEffects))
	}

	if err := s.characterService.UpdateEquipment(char); err != nil {
		log.Printf("Failed to save character state after ability use: %v", err)
	} else {
		log.Printf("Character saved successfully")
	}

	return result, nil
}

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

// GetAvailableAbilities returns all abilities a character can currently use
func (s *service) GetAvailableAbilities(ctx context.Context, characterID string) ([]*AvailableAbility, error) {
	// Get the character
	char, err := s.characterService.GetByID(characterID)
	if err != nil {
		return nil, dnderr.Wrap(err, "failed to get character")
	}

	resources := char.GetResources()
	if resources == nil || resources.Abilities == nil {
		return []*AvailableAbility{}, nil
	}

	var abilities []*AvailableAbility
	for _, ability := range resources.Abilities {
		available := &AvailableAbility{
			Ability:   ability,
			Available: ability.CanUse(),
		}

		if !available.Available {
			if ability.UsesRemaining == 0 {
				available.Reason = "No uses remaining"
			} else {
				available.Reason = "Cannot use ability"
			}
		}

		// Note: Duration for active abilities (like rage) is tracked in the turnStartListener
		// and updated on the status effect which is what the UI shows.
		// The ability definition always shows the max duration.

		abilities = append(abilities, available)
	}

	return abilities, nil
}

// ApplyAbilityEffects applies the effects of an ability (placeholder for future implementation)
func (s *service) ApplyAbilityEffects(ctx context.Context, input *ApplyEffectsInput) error {
	// This would handle applying effects in combat context
	// For now, it's a placeholder
	return nil
}

// turnStartListener handles turn start events to update ability durations
type turnStartListener struct {
	service *service
}

func (t *turnStartListener) HandleEvent(event *events.GameEvent) error {
	if event.Actor == nil {
		return nil
	}

	// Check if this character has active rage
	if listener, exists := t.service.activeListeners[event.Actor.ID]; exists {
		// Check if rage has expired
		if rageListener, ok := listener.(*features.RageListener); ok {
			duration := rageListener.Duration()
			if roundsDuration, ok := duration.(*events.RoundsDuration); ok {
				currentTurn, _ := event.GetIntContext("turn_count")
				numCombatants, _ := event.GetIntContext("num_combatants")

				// Calculate rounds elapsed (turns / combatants)
				turnsElapsed := currentTurn - roundsDuration.StartTurn
				roundsElapsed := 0
				if numCombatants > 0 {
					roundsElapsed = turnsElapsed / numCombatants
				}
				remainingRounds := roundsDuration.Rounds - roundsElapsed

				if remainingRounds <= 0 {
					// Rage has expired - unsubscribe and remove
					t.service.eventBus.Unsubscribe(events.OnDamageRoll, rageListener)
					t.service.eventBus.Unsubscribe(events.BeforeTakeDamage, rageListener)
					delete(t.service.activeListeners, event.Actor.ID)

					// Remove rage effect from character
					activeEffects := event.Actor.GetActiveStatusEffects()
					for _, effect := range activeEffects {
						if effect.Name == "Rage" {
							event.Actor.RemoveStatusEffect(effect.ID)
							break
						}
					}

					// Mark ability as inactive
					if resources := event.Actor.GetResources(); resources != nil {
						if rage, exists := resources.Abilities["rage"]; exists {
							rage.IsActive = false
						}
					}

					// Sync effects to ensure persistence layer is updated
					event.Actor.SyncEffects()

					// Save character
					if t.service.characterService != nil {
						if err := t.service.characterService.UpdateEquipment(event.Actor); err != nil {
							log.Printf("Failed to save character after rage expiry: %v", err)
						}
					}
				} else {
					// Update the UI effect duration
					activeEffects := event.Actor.GetActiveStatusEffects()
					for _, effect := range activeEffects {
						if effect.Name != "Rage" {
							continue
						}
						effect.Duration.Rounds = remainingRounds
						log.Printf("=== RAGE DURATION UPDATE ===")
						log.Printf("Character: %s", event.Actor.Name)
						log.Printf("Current turn: %d, Start turn: %d", currentTurn, roundsDuration.StartTurn)
						log.Printf("Turns elapsed: %d, Combatants: %d", turnsElapsed, numCombatants)
						log.Printf("Rounds elapsed: %d, Remaining rounds: %d", roundsElapsed, remainingRounds)
						break
					}

					// Sync effects to ensure persistence layer is updated
					event.Actor.SyncEffects()

					// Save character to persist the updated duration
					if t.service.characterService != nil {
						if err := t.service.characterService.UpdateEquipment(event.Actor); err != nil {
							log.Printf("Failed to save character after rage duration update: %v", err)
						} else {
							log.Printf("Character saved with updated rage duration")
						}
					}
				}
			}
		}
	}

	return nil
}

func (t *turnStartListener) Priority() int {
	return 200 // Lower priority, runs after other effects
}
