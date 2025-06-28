package ability

import (
	"context"
	"fmt"
	"log"

	"github.com/KirkDiggler/dnd-bot-discord/internal/dice"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	dnderr "github.com/KirkDiggler/dnd-bot-discord/internal/errors"
	"github.com/KirkDiggler/dnd-bot-discord/internal/interfaces"
	charService "github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	encounterService "github.com/KirkDiggler/dnd-bot-discord/internal/services/encounter"
)

type service struct {
	characterService charService.Service
	encounterService encounterService.Service
	diceRoller       interfaces.DiceRoller
}

// ServiceConfig holds configuration for the ability service
type ServiceConfig struct {
	CharacterService charService.Service
	EncounterService encounterService.Service
	DiceRoller       interfaces.DiceRoller
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
	}

	if svc.diceRoller == nil {
		svc.diceRoller = dice.NewRandomRoller()
	}

	return svc
}

// UseAbility executes an ability
func (s *service) UseAbility(ctx context.Context, input *UseAbilityInput) (*UseAbilityResult, error) {
	if input == nil {
		return nil, dnderr.InvalidArgument("input cannot be nil")
	}

	// Get the character
	character, err := s.characterService.GetByID(input.CharacterID)
	if err != nil {
		return nil, dnderr.Wrap(err, "failed to get character")
	}

	// Get character resources
	resources := character.GetResources()
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
		return s.handleLayOnHands(character, input.TargetID, input.Value, result), nil
	}

	// Use the ability
	if !ability.Use() {
		return &UseAbilityResult{
			Success:       false,
			Message:       "Failed to use ability",
			UsesRemaining: ability.UsesRemaining,
		}, nil
	}

	// Handle ability effects based on key
	result := &UseAbilityResult{
		Success:       true,
		UsesRemaining: ability.UsesRemaining,
	}

	switch input.AbilityKey {
	case "rage":
		result = s.handleRage(character, ability, result)
	case "second-wind":
		result = s.handleSecondWind(character, result)
	case "bardic-inspiration":
		result = s.handleBardicInspiration(character, input.TargetID, ability, result)
	case "divine-sense":
		result = s.handleDivineSense(character, result)
	default:
		result.Message = fmt.Sprintf("Used %s", ability.Name)
	}

	// Save character state
	log.Printf("=== SAVING CHARACTER AFTER ABILITY USE ===")
	log.Printf("Character: %s", character.Name)
	if character.Resources != nil {
		log.Printf("Active effects before save: %d", len(character.Resources.ActiveEffects))
	}

	if err := s.characterService.UpdateEquipment(character); err != nil {
		log.Printf("Failed to save character state after ability use: %v", err)
	} else {
		log.Printf("Character saved successfully")
	}

	return result, nil
}

// handleRage handles the Barbarian's Rage ability
func (s *service) handleRage(character *entities.Character, ability *entities.ActiveAbility, result *UseAbilityResult) *UseAbilityResult {
	// Create rage effect
	resources := character.GetResources()
	effect := &entities.ActiveEffect{
		ID:           fmt.Sprintf("rage_%s", character.ID),
		Name:         "Rage",
		Description:  "Advantage on Strength checks and saves, resistance to physical damage, +2 melee damage",
		Source:       "barbarian-rage",
		SourceID:     character.ID,
		Duration:     10, // 10 rounds
		DurationType: entities.DurationTypeRounds,
		Modifiers: []entities.Modifier{
			{
				Type:        entities.ModifierTypeDamageBonus,
				Value:       2,
				DamageTypes: []string{"melee"},
			},
			{
				Type:        entities.ModifierTypeDamageResistance,
				DamageTypes: []string{"bludgeoning", "piercing", "slashing"},
			},
		},
	}

	resources.AddEffect(effect)
	ability.IsActive = true
	ability.Duration = 10

	log.Printf("=== RAGE ACTIVATION DEBUG ===")
	log.Printf("Character: %s", character.Name)
	log.Printf("Effect added: %s (ID: %s)", effect.Name, effect.ID)
	log.Printf("Active effects after adding rage: %d", len(resources.ActiveEffects))
	log.Printf("Rage uses remaining: %d/%d", ability.UsesRemaining, ability.UsesMax)
	log.Printf("Rage is active: %v", ability.IsActive)

	result.Message = "You enter a rage! +2 damage to melee attacks, resistance to physical damage"
	result.EffectApplied = true
	result.EffectID = effect.ID
	result.EffectName = effect.Name
	result.Duration = effect.Duration

	return result
}

// handleSecondWind handles the Fighter's Second Wind ability
func (s *service) handleSecondWind(character *entities.Character, result *UseAbilityResult) *UseAbilityResult {
	// Roll 1d10 + fighter level
	total, _, err := s.diceRoller.Roll(1, 10, character.Level)
	if err != nil {
		result.Success = false
		result.Message = "Failed to roll healing"
		return result
	}

	resources := character.GetResources()
	healingDone := resources.HP.Heal(total)
	character.CurrentHitPoints = resources.HP.Current

	result.Message = fmt.Sprintf("Second Wind heals you for %d HP (rolled %d)", healingDone, total)
	result.HealingDone = healingDone
	result.TargetNewHP = resources.HP.Current

	return result
}

// handleBardicInspiration handles the Bard's Bardic Inspiration ability
func (s *service) handleBardicInspiration(character *entities.Character, targetID string, ability *entities.ActiveAbility, result *UseAbilityResult) *UseAbilityResult {
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
func (s *service) handleLayOnHands(character *entities.Character, targetID string, healAmount int, result *UseAbilityResult) *UseAbilityResult {
	resources := character.GetResources()
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
	if targetID == "" || targetID == character.ID {
		// Self heal
		healingDone := resources.HP.Heal(healAmount)
		character.CurrentHitPoints = resources.HP.Current

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
func (s *service) handleDivineSense(character *entities.Character, result *UseAbilityResult) *UseAbilityResult {
	result.Message = "You open your awareness to detect celestials, fiends, and undead within 60 feet"
	result.EffectApplied = true
	result.EffectName = "Divine Sense"
	result.Duration = 1 // Until end of next turn

	return result
}

// GetAvailableAbilities returns all abilities a character can currently use
func (s *service) GetAvailableAbilities(ctx context.Context, characterID string) ([]*AvailableAbility, error) {
	// Get the character
	character, err := s.characterService.GetByID(characterID)
	if err != nil {
		return nil, dnderr.Wrap(err, "failed to get character")
	}

	resources := character.GetResources()
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
