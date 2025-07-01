package ability

import (
	"context"
	"fmt"
	"log"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/events"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"

	"github.com/KirkDiggler/dnd-bot-discord/internal/dice"
	dnderr "github.com/KirkDiggler/dnd-bot-discord/internal/errors"
	charService "github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	encounterService "github.com/KirkDiggler/dnd-bot-discord/internal/services/encounter"
)

type service struct {
	characterService charService.Service
	encounterService encounterService.Service
	diceRoller       dice.Roller
	eventBus         *events.EventBus
	trackedAbilities map[string]*TrackedAbility // Track active abilities with duration by character ID
	executorRegistry *ExecutorRegistry
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
		trackedAbilities: make(map[string]*TrackedAbility),
		executorRegistry: NewExecutorRegistry(),
	}

	if svc.diceRoller == nil {
		svc.diceRoller = dice.NewRandomRoller()
	}

	// Subscribe to turn start events to update effect durations
	if svc.eventBus != nil {
		svc.eventBus.Subscribe(events.OnTurnStart, &turnStartListener{service: svc})
	}

	// Register D&D 5e ability executors
	// TODO: This should be moved to a rulebook registration system (#242)
	svc.executorRegistry.Register(newRageExecutor(svc))
	svc.executorRegistry.Register(newSecondWindExecutor(svc))
	svc.executorRegistry.Register(newBardicInspirationExecutor())
	svc.executorRegistry.Register(newLayOnHandsExecutor(svc))
	svc.executorRegistry.Register(newDivineSenseExecutor())

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

	// Check if we have a custom executor for this ability
	executor, hasExecutor := s.executorRegistry.Get(input.AbilityKey)

	// For abilities with custom executors that manage their own resources,
	// don't use the standard Use() method
	if !hasExecutor || input.AbilityKey != "lay-on-hands" {
		// Use the ability for standard abilities
		if !ability.Use() {
			return &UseAbilityResult{
				Success:       false,
				Message:       "Failed to use ability",
				UsesRemaining: ability.UsesRemaining,
			}, nil
		}
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

	// Execute using the registered executor if available
	if !hasExecutor {
		// If no specific executor, just mark as used with generic message
		result.Message = fmt.Sprintf("Used %s", ability.Name)
	} else {
		// Execute the ability using its specific executor
		executorResult, err := executor.Execute(ctx, char, ability, input)
		if err != nil {
			return nil, dnderr.Wrap(err, "failed to execute ability")
		}
		result = executorResult
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

// GetExecutorRegistry returns the executor registry for registration
func (s *service) GetExecutorRegistry() *ExecutorRegistry {
	return s.executorRegistry
}

// turnStartListener handles turn start events to update ability durations
type turnStartListener struct {
	service *service
}

func (t *turnStartListener) HandleEvent(event *events.GameEvent) error {
	if event.Actor == nil {
		return nil
	}

	// Check if this character has an active tracked ability
	if tracked, exists := t.service.trackedAbilities[event.Actor.ID]; exists {
		currentTurn, _ := event.GetIntContext("turn_count")       // TODO: Use constant
		numCombatants, _ := event.GetIntContext("num_combatants") // TODO: Use constant

		// Check if the tracked ability has expired
		if tracked.Tracker.IsExpired(currentTurn, numCombatants) {
			// Handle expiration
			tracked.Tracker.OnExpire(t.service.eventBus)
			delete(t.service.trackedAbilities, event.Actor.ID)

			// Remove effect from character
			activeEffects := event.Actor.GetActiveStatusEffects()
			for _, effect := range activeEffects {
				if effect.Name == tracked.Tracker.GetEffectName() {
					event.Actor.RemoveStatusEffect(effect.ID)
					break
				}
			}

			// Mark ability as inactive
			if resources := event.Actor.GetResources(); resources != nil {
				if ability, exists := resources.Abilities[tracked.Tracker.GetAbilityKey()]; exists {
					ability.IsActive = false
				}
			}

			// Sync effects to ensure persistence layer is updated
			event.Actor.SyncEffects()

			// Save character
			if t.service.characterService != nil {
				if err := t.service.characterService.UpdateEquipment(event.Actor); err != nil {
					log.Printf("Failed to save character after %s expiry: %v", tracked.Tracker.GetEffectName(), err)
				}
			}
		} else {
			// Update the UI effect duration
			remainingRounds := tracked.Tracker.GetRemainingRounds(currentTurn, numCombatants)
			activeEffects := event.Actor.GetActiveStatusEffects()
			for _, effect := range activeEffects {
				if effect.Name != tracked.Tracker.GetEffectName() {
					continue
				}
				effect.Duration.Rounds = remainingRounds
				log.Printf("=== %s DURATION UPDATE ===", tracked.Tracker.GetEffectName())
				log.Printf("Character: %s", event.Actor.Name)
				log.Printf("Remaining rounds: %d", remainingRounds)
				break
			}

			// Sync effects to ensure persistence layer is updated
			event.Actor.SyncEffects()

			// Save character to persist the updated duration
			if t.service.characterService != nil {
				if err := t.service.characterService.UpdateEquipment(event.Actor); err != nil {
					log.Printf("Failed to save character after %s duration update: %v", tracked.Tracker.GetEffectName(), err)
				} else {
					log.Printf("Character saved with updated %s duration", tracked.Tracker.GetEffectName())
				}
			}
		}
	}

	return nil
}

func (t *turnStartListener) Priority() int {
	return 200 // Lower priority, runs after other effects
}
