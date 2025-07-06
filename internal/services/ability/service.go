package ability

import (
	"context"
	"fmt"
	"log"

	"github.com/KirkDiggler/dnd-bot-discord/internal/dice"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	dnderr "github.com/KirkDiggler/dnd-bot-discord/internal/errors"
	charService "github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	rpgevents "github.com/KirkDiggler/rpg-toolkit/events"
)

// service is the implementation without any hardcoded abilities
type service struct {
	characterService charService.Service
	diceRoller       dice.Roller
	eventBus         *rpgevents.Bus
	registry         *HandlerRegistry
}

// ServiceConfig holds configuration for the ability service
type ServiceConfig struct {
	CharacterService charService.Service
	DiceRoller       dice.Roller
	EventBus         *rpgevents.Bus
}

// NewService creates a new ability service without any hardcoded abilities
func NewService(cfg *ServiceConfig) Service {
	if cfg.CharacterService == nil {
		panic("character service is required")
	}

	svc := &service{
		characterService: cfg.CharacterService,
		diceRoller:       cfg.DiceRoller,
		eventBus:         cfg.EventBus,
		registry:         NewHandlerRegistry(),
	}

	if svc.diceRoller == nil {
		svc.diceRoller = dice.NewRandomRoller()
	}

	return svc
}

// RegisterHandler allows external packages to register ability handlers
func (s *service) RegisterHandler(handler Handler) {
	s.registry.Register(handler)
}

// UseAbility executes an ability using the registered handler
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

	// Look up the handler for this ability
	handler, hasHandler := s.registry.Get(input.AbilityKey)
	if !hasHandler {
		// No specific handler, just use the ability and return generic message
		if !ability.Use() {
			return &UseAbilityResult{
				Success:       false,
				Message:       "Failed to use ability",
				UsesRemaining: ability.UsesRemaining,
			}, nil
		}

		s.updateActionEconomy(char, ability)

		// Save character state
		if updateErr := s.characterService.UpdateEquipment(char); updateErr != nil {
			log.Printf("Failed to save character state after ability use: %v", updateErr)
		}

		return &UseAbilityResult{
			Success:       true,
			Message:       fmt.Sprintf("Used %s", ability.Name),
			UsesRemaining: ability.UsesRemaining,
		}, nil
	}

	// Special handling for abilities that manage their own resources
	shouldUseStandardResource := input.AbilityKey != "lay-on-hands"

	if shouldUseStandardResource {
		// Use the ability for standard abilities
		if !ability.Use() {
			return &UseAbilityResult{
				Success:       false,
				Message:       "Failed to use ability",
				UsesRemaining: ability.UsesRemaining,
			}, nil
		}
	}

	// Update action economy
	s.updateActionEconomy(char, ability)

	// Convert input types
	handlerInput := &HandlerInput{
		CharacterID: input.CharacterID,
		AbilityKey:  input.AbilityKey,
		TargetID:    input.TargetID,
		Value:       input.Value,
		EncounterID: input.EncounterID,
	}

	// Execute the ability using the handler
	handlerResult, err := handler.Execute(ctx, char, ability, handlerInput)
	if err != nil {
		// Restore the use if execution failed
		if shouldUseStandardResource {
			ability.UsesRemaining++
		}
		return nil, dnderr.Wrap(err, "failed to execute ability")
	}

	// Convert result types
	result := &UseAbilityResult{
		Success:       handlerResult.Success,
		Message:       handlerResult.Message,
		EffectApplied: handlerResult.EffectApplied,
		UsesRemaining: handlerResult.UsesRemaining,
		HealingDone:   handlerResult.HealingDone,
		TargetNewHP:   handlerResult.TargetNewHP,
		EffectID:      handlerResult.EffectID,
		EffectName:    handlerResult.EffectName,
		Duration:      handlerResult.Duration,
	}

	// Save character state
	if updateErr := s.characterService.UpdateEquipment(char); updateErr != nil {
		log.Printf("Failed to save character state after ability use: %v", updateErr)
	}

	return result, nil
}

func (s *service) updateActionEconomy(char *character.Character, ability *shared.ActiveAbility) {
	if char.Resources == nil {
		return
	}

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
