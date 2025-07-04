package spells

import (
	"context"
	"fmt"
	"log"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/events"
)

// MagicMissileHandler implements the magic missile spell
type MagicMissileHandler struct {
	eventBus   events.Bus
	diceRoller interface {
		Roll(numDice, sides, modifier int) (struct{ Total int }, error)
	}
	characterService interface {
		UpdateEquipment(char *character.Character) error
		GetByID(characterID string) (*character.Character, error)
	}
}

// NewMagicMissileHandler creates a new magic missile handler
func NewMagicMissileHandler(eventBus events.Bus) *MagicMissileHandler {
	return &MagicMissileHandler{
		eventBus: eventBus,
	}
}

// SetDiceRoller sets the dice roller dependency
func (m *MagicMissileHandler) SetDiceRoller(roller interface{}) {
	if r, ok := roller.(interface {
		Roll(numDice, sides, modifier int) (struct{ Total int }, error)
	}); ok {
		m.diceRoller = r
	}
}

// SetCharacterService sets the character service dependency
func (m *MagicMissileHandler) SetCharacterService(service interface{}) {
	if svc, ok := service.(interface {
		UpdateEquipment(char *character.Character) error
		GetByID(characterID string) (*character.Character, error)
	}); ok {
		m.characterService = svc
	}
}

// Key returns the spell key
func (m *MagicMissileHandler) Key() string {
	return "magic_missile"
}

// SpellInput contains input for spell execution
type SpellInput struct {
	SpellLevel  int      // Level at which spell is cast
	TargetIDs   []string // Target character IDs
	EncounterID string   // Current encounter
}

// SpellResult contains the result of spell execution
type SpellResult struct {
	Success       bool
	Message       string
	Damage        map[string]int // Damage per target
	TotalDamage   int
	SpellSlotUsed int
}

// Execute casts magic missile
func (m *MagicMissileHandler) Execute(ctx context.Context, caster *character.Character, input *SpellInput) (*SpellResult, error) {
	result := &SpellResult{
		Success:       true,
		Damage:        make(map[string]int),
		SpellSlotUsed: input.SpellLevel,
	}

	// Validate spell level (1-9)
	if input.SpellLevel < 1 || input.SpellLevel > 9 {
		result.Success = false
		result.Message = "Invalid spell level"
		return result, nil
	}

	// Check if caster has spell slots
	// TODO: Implement spell slot tracking
	// For now, we'll assume they have slots

	// Calculate number of missiles (3 + 1 per level above 1st)
	numMissiles := 3 + (input.SpellLevel - 1)

	// Each missile does 1d4+1 force damage
	damagePerMissile := 0
	if m.diceRoller != nil {
		rollResult, err := m.diceRoller.Roll(1, 4, 1)
		if err != nil {
			return nil, fmt.Errorf("failed to roll damage: %w", err)
		}
		damagePerMissile = rollResult.Total
	} else {
		damagePerMissile = 3 // Average of 1d4+1
	}

	// Emit spell cast event
	if m.eventBus != nil {
		spellEvent := events.NewGameEvent(events.OnSpellCast).
			WithActor(caster).
			WithContext(events.ContextSpellLevel, input.SpellLevel).
			WithContext(events.ContextSpellName, "Magic Missile").
			WithContext(events.ContextSpellSchool, "evocation")

		if err := m.eventBus.Emit(spellEvent); err != nil {
			log.Printf("Failed to emit OnSpellCast event: %v", err)
		}
	}

	// Distribute missiles among targets
	// For simplicity, distribute evenly with remainder going to first target
	if len(input.TargetIDs) == 0 {
		result.Success = false
		result.Message = "No targets specified"
		return result, nil
	}

	missilesPerTarget := numMissiles / len(input.TargetIDs)
	extraMissiles := numMissiles % len(input.TargetIDs)

	// Apply damage to each target
	for i, targetID := range input.TargetIDs {
		missiles := missilesPerTarget
		if i < extraMissiles {
			missiles++
		}

		totalDamage := missiles * damagePerMissile

		// Get target character
		var target *character.Character
		if m.characterService != nil {
			var err error
			target, err = m.characterService.GetByID(targetID)
			if err != nil {
				log.Printf("Failed to get target %s: %v", targetID, err)
				continue
			}
		}

		// Emit damage event for each target
		if m.eventBus != nil && target != nil {
			damageEvent := events.NewGameEvent(events.OnSpellDamage).
				WithActor(caster).
				WithTarget(target).
				WithContext(events.ContextDamage, totalDamage).
				WithContext(events.ContextDamageType, "force").
				WithContext(events.ContextSpellName, "Magic Missile").
				WithContext("num_missiles", missiles).
				WithContext(events.ContextEncounterID, input.EncounterID).
				WithContext(events.ContextTargetID, targetID)

			if err := m.eventBus.Emit(damageEvent); err != nil {
				log.Printf("Failed to emit OnSpellDamage event: %v", err)
			}

			// Check if damage was modified
			if modifiedDamage, exists := damageEvent.GetIntContext(events.ContextDamage); exists {
				totalDamage = modifiedDamage
			}
		}

		result.Damage[targetID] = totalDamage
		result.TotalDamage += totalDamage
	}

	// Format message
	if len(input.TargetIDs) == 1 {
		result.Message = fmt.Sprintf("Magic Missile strikes for %d force damage!", result.TotalDamage)
	} else {
		result.Message = fmt.Sprintf("Magic Missile strikes %d targets for a total of %d force damage!",
			len(input.TargetIDs), result.TotalDamage)
	}

	log.Printf("=== MAGIC MISSILE CAST ===")
	log.Printf("Caster: %s", caster.Name)
	log.Printf("Spell Level: %d", input.SpellLevel)
	log.Printf("Number of Missiles: %d", numMissiles)
	log.Printf("Damage per Missile: %d", damagePerMissile)
	log.Printf("Targets: %v", input.TargetIDs)
	log.Printf("Total Damage: %d", result.TotalDamage)

	return result, nil
}
