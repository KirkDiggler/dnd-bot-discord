package abilities

import (
	"context"
	"fmt"
	"log"
	"reflect"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/events"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e/features"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"github.com/KirkDiggler/dnd-bot-discord/internal/effects"
)

// RageHandler implements the barbarian rage ability
type RageHandler struct {
	eventBus         events.Bus
	encounterService interface {
		GetEncounter(ctx context.Context, id string) (*Encounter, error)
	}
	characterService interface {
		UpdateEquipment(char *character.Character) error
	}
	activeListeners map[string]events.EventListener // Track active rage listeners
}

// Encounter represents minimal encounter data needed
type Encounter struct {
	Round     int
	Turn      int
	TurnOrder []string
}

// NewRageHandler creates a new rage handler
func NewRageHandler(eventBus events.Bus) *RageHandler {
	return &RageHandler{
		eventBus:        eventBus,
		activeListeners: make(map[string]events.EventListener),
	}
}

// SetEncounterService sets the encounter service dependency
// The service should have a GetEncounter method that returns an object with Round, Turn, and TurnOrder fields
func (r *RageHandler) SetEncounterService(service interface{}) {
	// Type assert to check if it has the method we need
	if svc, ok := service.(interface {
		GetEncounter(ctx context.Context, id string) (interface{}, error)
	}); ok {
		r.encounterService = &encounterServiceAdapter{service: svc}
	}
}

// encounterServiceAdapter adapts the actual encounter service to our minimal interface
type encounterServiceAdapter struct {
	service interface {
		GetEncounter(ctx context.Context, id string) (interface{}, error)
	}
}

func (a *encounterServiceAdapter) GetEncounter(ctx context.Context, id string) (*Encounter, error) {
	enc, err := a.service.GetEncounter(ctx, id)
	if err != nil {
		return nil, err
	}

	// Use reflection to extract the fields we need
	val := reflect.ValueOf(enc)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("encounter is not a struct")
	}

	// Extract fields using reflection
	roundField := val.FieldByName("Round")
	turnField := val.FieldByName("Turn")
	turnOrderField := val.FieldByName("TurnOrder")

	if !roundField.IsValid() || !turnField.IsValid() || !turnOrderField.IsValid() {
		return nil, fmt.Errorf("encounter missing required fields")
	}

	round := int(roundField.Int())
	turn := int(turnField.Int())

	// TurnOrder is []string in combat.Encounter
	var turnOrder []string
	if turnOrderField.Kind() == reflect.Slice {
		for i := 0; i < turnOrderField.Len(); i++ {
			if id := turnOrderField.Index(i).String(); id != "" {
				turnOrder = append(turnOrder, id)
			}
		}
	}

	return &Encounter{
		Round:     round,
		Turn:      turn,
		TurnOrder: turnOrder,
	}, nil
}

// SetCharacterService sets the character service dependency
func (r *RageHandler) SetCharacterService(service interface{}) {
	// Type assert to check if it has the method we need
	if svc, ok := service.(interface {
		UpdateEquipment(char *character.Character) error
	}); ok {
		r.characterService = svc
	}
}

// Key returns the ability key
func (r *RageHandler) Key() string {
	return "rage"
}

// Execute activates rage
func (r *RageHandler) Execute(ctx context.Context, char *character.Character, ability *shared.ActiveAbility, input *Input) (*Result, error) {
	result := &Result{
		Success:       true,
		UsesRemaining: ability.UsesRemaining,
	}

	// Check if EventBus is available
	if r.eventBus == nil {
		// Fallback to old system if EventBus not available
		rageEffect := effects.BuildRageEffect(char.Level)
		err := char.AddStatusEffect(rageEffect)
		if err != nil {
			log.Printf("Failed to add rage effect: %v", err)
			ability.UsesRemaining++ // Restore the use
			result.Success = false
			result.Message = "Failed to enter rage"
			return result, nil
		}
		ability.IsActive = true
		ability.Duration = 10
		result.Message = fmt.Sprintf("You enter a rage! +%d damage to melee attacks, resistance to physical damage", r.calculateRageBonus(char.Level))
		result.EffectApplied = true
		result.EffectID = rageEffect.ID
		result.EffectName = rageEffect.Name
		result.Duration = 10
		return result, nil
	}

	// Get current turn count from encounter
	currentTurn := 0
	if input.EncounterID != "" && r.encounterService != nil {
		if encounter, err := r.encounterService.GetEncounter(ctx, input.EncounterID); err == nil {
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
	r.eventBus.Subscribe(events.OnDamageRoll, rageListener)
	r.eventBus.Subscribe(events.BeforeTakeDamage, rageListener)

	// Track the active listener
	r.activeListeners[char.ID] = rageListener

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
	damageBonus := r.calculateRageBonus(char.Level)
	damageBonusStr := fmt.Sprintf("+%d", damageBonus)

	log.Printf("=== RAGE ACTIVATION (EVENT SYSTEM) ===")
	log.Printf("Character: %s (ID: %s)", char.Name, char.ID)
	log.Printf("Rage listener registered with ID: %s", rageListener.ID())
	log.Printf("Rage uses remaining: %d/%d", ability.UsesRemaining, ability.UsesMax)
	log.Printf("Rage is active: %v", ability.IsActive)
	log.Printf("Event listeners registered for OnDamageRoll and BeforeTakeDamage")

	result.Message = fmt.Sprintf("You enter a rage! %s damage to melee attacks, resistance to physical damage", damageBonusStr)
	result.EffectApplied = true
	result.EffectID = rageListener.ID()
	result.EffectName = "Rage"
	result.Duration = 10
	result.DamageBonus = damageBonus

	return result, nil
}

func (r *RageHandler) calculateRageBonus(level int) int {
	if level >= 16 {
		return 4
	} else if level >= 9 {
		return 3
	}
	return 2
}

// GetActiveListener returns the active rage listener for a character
func (r *RageHandler) GetActiveListener(characterID string) events.EventListener {
	return r.activeListeners[characterID]
}

// RemoveActiveListener removes the active rage listener for a character
func (r *RageHandler) RemoveActiveListener(characterID string) {
	delete(r.activeListeners, characterID)
}
