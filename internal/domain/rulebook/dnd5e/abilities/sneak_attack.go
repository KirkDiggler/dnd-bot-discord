package abilities

import (
	"context"
	"fmt"
	"log"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/events"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e/features"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
)

// SneakAttackHandler implements the rogue sneak attack ability
type SneakAttackHandler struct {
	eventBus   events.Bus
	diceRoller interface {
		Roll(numDice, sides, modifier int) (struct{ Total int }, error)
	}
	encounterService interface {
		GetEncounter(ctx context.Context, id string) (*Encounter, error)
	}
	characterService interface {
		UpdateEquipment(char *character.Character) error
	}
	activeListeners map[string]events.EventListener // Track active sneak attack listeners
}

// NewSneakAttackHandler creates a new sneak attack handler
func NewSneakAttackHandler(eventBus events.Bus) *SneakAttackHandler {
	return &SneakAttackHandler{
		eventBus:        eventBus,
		activeListeners: make(map[string]events.EventListener),
	}
}

// SetDiceRoller sets the dice roller dependency
func (s *SneakAttackHandler) SetDiceRoller(roller interface{}) {
	if r, ok := roller.(interface {
		Roll(numDice, sides, modifier int) (struct{ Total int }, error)
	}); ok {
		s.diceRoller = r
	}
}

// SetEncounterService sets the encounter service dependency
func (s *SneakAttackHandler) SetEncounterService(service interface{}) {
	// Type assert to check if it has the method we need
	if svc, ok := service.(interface {
		GetEncounter(ctx context.Context, id string) (interface{}, error)
	}); ok {
		s.encounterService = &encounterServiceAdapter{service: svc}
	}
}

// SetCharacterService sets the character service dependency
func (s *SneakAttackHandler) SetCharacterService(service interface{}) {
	// Type assert to check if it has the method we need
	if svc, ok := service.(interface {
		UpdateEquipment(char *character.Character) error
	}); ok {
		s.characterService = svc
	}
}

// Key returns the ability key
func (s *SneakAttackHandler) Key() string {
	return "sneak_attack"
}

// Execute activates sneak attack tracking
func (s *SneakAttackHandler) Execute(ctx context.Context, char *character.Character, ability *shared.ActiveAbility, input *Input) (*Result, error) {
	result := &Result{
		Success: true,
		Message: "Sneak Attack tracking enabled",
	}

	// Check if character is a rogue
	if char.Class == nil || char.Class.Key != "rogue" {
		result.Success = false
		result.Message = "Only rogues can use Sneak Attack"
		return result, nil
	}

	// Check if EventBus is available
	if s.eventBus == nil {
		result.Success = false
		result.Message = "Event system not available"
		return result, nil
	}

	// Initialize character resources if needed
	if char.Resources == nil {
		char.Resources = &character.CharacterResources{}
	}

	// Get current turn count from encounter
	currentTurn := 0
	if input.EncounterID != "" && s.encounterService != nil {
		if encounter, err := s.encounterService.GetEncounter(ctx, input.EncounterID); err == nil {
			// Calculate total turns (rounds * combatants + current turn index)
			if len(encounter.TurnOrder) > 0 {
				currentTurn = (encounter.Round-1)*len(encounter.TurnOrder) + encounter.Turn
			}
			log.Printf("=== SNEAK ATTACK ACTIVATION INFO ===")
			log.Printf("Encounter ID: %s", input.EncounterID)
			log.Printf("Current Round: %d, Turn index: %d, Combatants: %d", encounter.Round, encounter.Turn, len(encounter.TurnOrder))
			log.Printf("Starting at turn count: %d", currentTurn)
		} else {
			log.Printf("Failed to get encounter for sneak attack activation: %v", err)
		}
	}

	// Check if we already have an active listener
	if existingListener, exists := s.activeListeners[char.ID]; exists {
		// Unsubscribe the old listener
		s.eventBus.Unsubscribe(events.OnDamageRoll, existingListener)
		s.eventBus.Unsubscribe(events.OnTurnStart, existingListener)
		delete(s.activeListeners, char.ID)
		log.Printf("Removed existing sneak attack listener for character %s", char.ID)
	}

	// Create sneak attack listener for event system
	sneakAttackListener := features.NewSneakAttackListener(char.ID, char.Level, currentTurn, s.diceRoller)

	// Subscribe sneak attack listener to relevant events
	s.eventBus.Subscribe(events.OnDamageRoll, sneakAttackListener)
	s.eventBus.Subscribe(events.OnTurnStart, sneakAttackListener)

	// Track the active listener
	s.activeListeners[char.ID] = sneakAttackListener

	// Mark ability as active
	ability.IsActive = true

	log.Printf("=== SNEAK ATTACK ACTIVATION (EVENT SYSTEM) ===")
	log.Printf("Character: %s (ID: %s)", char.Name, char.ID)
	log.Printf("Sneak attack listener registered with ID: %s", sneakAttackListener.ID())
	log.Printf("Sneak attack dice: %dd6", (char.Level+1)/2)
	log.Printf("Event listeners registered for OnDamageRoll and OnTurnStart")

	result.Message = fmt.Sprintf("Sneak Attack tracking enabled (%dd6 damage)", (char.Level+1)/2)
	result.EffectApplied = true
	result.EffectID = sneakAttackListener.ID()
	result.EffectName = "Sneak Attack"

	// Save character to persist the active state
	if s.characterService != nil {
		if err := s.characterService.UpdateEquipment(char); err != nil {
			log.Printf("Failed to save character after enabling sneak attack: %v", err)
		}
	}

	return result, nil
}

// GetActiveListener returns the active sneak attack listener for a character
func (s *SneakAttackHandler) GetActiveListener(characterID string) events.EventListener {
	return s.activeListeners[characterID]
}

// RemoveActiveListener removes the active sneak attack listener for a character
func (s *SneakAttackHandler) RemoveActiveListener(characterID string) {
	delete(s.activeListeners, characterID)
}
