package spells

import (
	"context"
	"fmt"
	"log"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/events"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
)

// ViciousMockeryHandler implements the vicious mockery cantrip
type ViciousMockeryHandler struct {
	eventBus   *events.EventBus
	diceRoller interface {
		Roll(numDice, sides, modifier int) (struct{ Total int }, error)
	}
	characterService interface {
		UpdateEquipment(char *character.Character) error
		GetByID(characterID string) (*character.Character, error)
	}
}

// NewViciousMockeryHandler creates a new vicious mockery handler
func NewViciousMockeryHandler(eventBus *events.EventBus) *ViciousMockeryHandler {
	return &ViciousMockeryHandler{
		eventBus: eventBus,
	}
}

// SetDiceRoller sets the dice roller dependency
func (v *ViciousMockeryHandler) SetDiceRoller(roller interface{}) {
	if r, ok := roller.(interface {
		Roll(numDice, sides, modifier int) (struct{ Total int }, error)
	}); ok {
		v.diceRoller = r
	}
}

// SetCharacterService sets the character service dependency
func (v *ViciousMockeryHandler) SetCharacterService(service interface{}) {
	if svc, ok := service.(interface {
		UpdateEquipment(char *character.Character) error
		GetByID(characterID string) (*character.Character, error)
	}); ok {
		v.characterService = svc
	}
}

// Key returns the spell key
func (v *ViciousMockeryHandler) Key() string {
	return "vicious_mockery"
}

// Execute casts vicious mockery
func (v *ViciousMockeryHandler) Execute(ctx context.Context, caster *character.Character, input *SpellInput) (*SpellResult, error) {
	result := &SpellResult{
		Success: true,
		Damage:  make(map[string]int),
		// Cantrips don't use spell slots
		SpellSlotUsed: 0,
	}

	// Validate single target
	if len(input.TargetIDs) != 1 {
		result.Success = false
		result.Message = "Vicious Mockery requires exactly one target"
		return result, nil
	}

	targetID := input.TargetIDs[0]

	// Get target - for now, we'll create a minimal target for monsters
	var target *character.Character
	if v.characterService != nil {
		var err error
		target, err = v.characterService.GetByID(targetID)
		if err != nil || target == nil {
			// For monsters, create a minimal character object for the event
			// In a full implementation, we'd have a proper way to handle monster targets
			target = &character.Character{
				ID:   targetID,
				Name: "Monster", // We don't have the name here
			}
		}
	}

	// Calculate save DC (8 + proficiency + CHA modifier)
	proficiencyBonus := caster.GetProficiencyBonus()
	chaBonus := 0
	if chaScore, exists := caster.Attributes[shared.AttributeCharisma]; exists && chaScore != nil {
		chaBonus = chaScore.Bonus
	}
	saveDC := 8 + proficiencyBonus + chaBonus

	// Emit spell cast event
	if v.eventBus != nil {
		spellEvent := events.NewGameEvent(events.OnSpellCast).
			WithActor(caster).
			WithTarget(target).
			WithContext(events.ContextSpellLevel, 0). // Cantrip
			WithContext("spell_name", "Vicious Mockery").
			WithContext(events.ContextSpellSchool, "enchantment").
			WithContext(events.ContextSpellSaveType, "WIS").
			WithContext(events.ContextSpellSaveDC, saveDC)

		if err := v.eventBus.Emit(spellEvent); err != nil {
			log.Printf("Failed to emit OnSpellCast event: %v", err)
		}
	}

	// TODO: Implement actual saving throw
	// For now, we'll assume the save fails
	saveFailed := true

	if saveFailed {
		// Calculate damage based on caster level
		// 1d4 at level 1, 2d4 at level 5, 3d4 at level 11, 4d4 at level 17
		diceCount := 1
		if caster.Level >= 17 {
			diceCount = 4
		} else if caster.Level >= 11 {
			diceCount = 3
		} else if caster.Level >= 5 {
			diceCount = 2
		}

		// Roll damage
		damage := 0
		if v.diceRoller != nil {
			rollResult, err := v.diceRoller.Roll(diceCount, 4, 0)
			if err != nil {
				return nil, fmt.Errorf("failed to roll damage: %w", err)
			}
			damage = rollResult.Total
		} else {
			// Average damage
			damage = diceCount * 2 // Average of 1d4 is 2.5, rounded down
		}

		// Emit damage event
		if v.eventBus != nil && target != nil {
			damageEvent := events.NewGameEvent(events.OnSpellDamage).
				WithActor(caster).
				WithTarget(target).
				WithContext(events.ContextDamage, damage).
				WithContext(events.ContextDamageType, "psychic").
				WithContext("spell_name", "Vicious Mockery").
				WithContext("damage_dice", fmt.Sprintf("%dd4", diceCount)).
				WithContext("encounter_id", input.EncounterID).
				WithContext(events.ContextTargetID, targetID)

			if err := v.eventBus.Emit(damageEvent); err != nil {
				log.Printf("Failed to emit OnSpellDamage event: %v", err)
			}

			// Check if damage was modified
			if modifiedDamage, exists := damageEvent.GetIntContext(events.ContextDamage); exists {
				damage = modifiedDamage
			}
		}

		result.Damage[targetID] = damage
		result.TotalDamage = damage
		result.Message = fmt.Sprintf("Your cutting words deal %d psychic damage! The target has disadvantage on their next attack.", damage)

		// TODO: Apply disadvantage effect to target's next attack
		// This would need to be tracked as a status effect

	} else {
		result.Message = "The target shrugs off your mockery!"
	}

	log.Printf("=== VICIOUS MOCKERY CAST ===")
	log.Printf("Caster: %s (Level %d)", caster.Name, caster.Level)
	log.Printf("Target: %s", targetID)
	log.Printf("Save DC: %d", saveDC)
	log.Printf("Save Failed: %v", saveFailed)
	if saveFailed {
		log.Printf("Damage: %d", result.TotalDamage)
	}

	return result, nil
}
