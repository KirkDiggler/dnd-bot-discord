package encounter

import (
	"fmt"
	"log"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/equipment"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/events"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/combat"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
)

// ProficiencyHandler handles proficiency-related events for encounters
type ProficiencyHandler struct {
	service Service
}

// NewProficiencyHandler creates a new proficiency event handler
func NewProficiencyHandler(service Service) *ProficiencyHandler {
	return &ProficiencyHandler{
		service: service,
	}
}

// HandleEvent implements the EventListener interface
func (ph *ProficiencyHandler) HandleEvent(event *events.GameEvent) error {
	switch event.Type {
	case events.OnAttackRoll:
		return ph.handleAttackRollProficiency(event)
	case events.OnSavingThrow:
		return ph.handleSavingThrowProficiency(event)
	}
	return nil
}

// Priority returns the priority for this handler (lower executes first)
func (ph *ProficiencyHandler) Priority() int {
	return 50 // Mid-priority, runs after base calculations but before most modifiers
}

// handleAttackRollProficiency adds proficiency bonus to attack rolls
func (ph *ProficiencyHandler) handleAttackRollProficiency(event *events.GameEvent) error {
	// Get attacker from event
	attacker := event.Actor
	if attacker == nil {
		return nil
	}

	// Get weapon being used from context
	weaponName, ok := event.GetStringContext("weapon")
	if !ok || weaponName == "" {
		return nil
	}

	// Skip unarmed strikes for proficiency (everyone is proficient)
	if weaponName == "Unarmed Strike" {
		return nil
	}

	// Find weapon in character's equipment
	var weapon *equipment.Weapon
	if attacker.EquippedSlots != nil {
		// Check main hand
		if item := attacker.EquippedSlots[shared.SlotMainHand]; item != nil {
			if w, ok := item.(*equipment.Weapon); ok && w.Base.Name == weaponName {
				weapon = w
			}
		}
		// Check two-handed
		if weapon == nil {
			if item := attacker.EquippedSlots[shared.SlotTwoHanded]; item != nil {
				if w, ok := item.(*equipment.Weapon); ok && w.Base.Name == weaponName {
					weapon = w
				}
			}
		}
	}

	// Check if character is proficient
	if weapon != nil && attacker.HasWeaponProficiency(weapon.Base.Key) {
		// Calculate proficiency bonus based on level
		profBonus := getProficiencyBonus(attacker.Level)

		// Get current attack bonus and add proficiency
		currentBonus, _ := event.GetIntContext("attack_bonus")
		totalBonus := currentBonus + profBonus

		// Update the attack bonus in the event
		event.WithContext("attack_bonus", totalBonus)

		// Also update total attack if it exists
		if totalAttack, ok := event.GetIntContext("total_attack"); ok {
			event.WithContext("total_attack", totalAttack+profBonus)
		}

		log.Printf("Applied proficiency bonus +%d for %s's attack with %s (total bonus now: %d)",
			profBonus, attacker.Name, weapon.Base.Name, totalBonus)
	} else if weapon != nil {
		log.Printf("No proficiency bonus for %s's attack with %s (not proficient)",
			attacker.Name, weapon.Base.Name)
	}

	return nil
}

// handleSavingThrowProficiency adds proficiency bonus to saving throws
func (ph *ProficiencyHandler) handleSavingThrowProficiency(event *events.GameEvent) error {
	// Get target from event
	target := event.Target
	if target == nil {
		// Sometimes the character making the save is the actor
		target = event.Actor
		if target == nil {
			return nil
		}
	}

	// Get save type from context
	saveType, ok := event.GetStringContext("save_type")
	if !ok || saveType == "" {
		return nil
	}

	// Check if character is proficient in this saving throw
	profKey := fmt.Sprintf("saving-throw-%s", strings.ToLower(saveType))
	if target.Proficiencies != nil {
		// Check all proficiency types for saving throws
		for _, profs := range target.Proficiencies {
			for _, prof := range profs {
				if prof == nil || prof.Key != profKey {
					continue
				}
				// Calculate proficiency bonus based on level
				profBonus := getProficiencyBonus(target.Level)

				// Get current save bonus and add proficiency
				currentBonus, _ := event.GetIntContext("save_bonus")
				totalBonus := currentBonus + profBonus

				// Update the save bonus in the event
				event.WithContext("save_bonus", totalBonus)

				// Also update total save if it exists
				if totalSave, ok := event.GetIntContext("total_save"); ok {
					event.WithContext("total_save", totalSave+profBonus)
				}

				log.Printf("Applied proficiency bonus +%d for %s's %s saving throw",
					profBonus, target.Name, saveType)
				break
			}
		}
	}

	return nil
}

// getProficiencyBonus calculates D&D 5e proficiency bonus based on level
func getProficiencyBonus(level int) int {
	if level < 1 {
		level = 1
	}
	return 2 + ((level - 1) / 4)
}

// GetCombatantProficiencyIndicator returns a string indicator for proficiency status
func GetCombatantProficiencyIndicator(combatant *combat.Combatant, char *character.Character) string {
	if char == nil || char.EquippedSlots == nil {
		return ""
	}

	// Check main hand
	if item := char.EquippedSlots[shared.SlotMainHand]; item != nil {
		if weapon, ok := item.(*equipment.Weapon); ok {
			if char.HasWeaponProficiency(weapon.Base.Key) {
				return " PROF"
			}
			return " !PROF"
		}
	}

	// Check two-handed
	if item := char.EquippedSlots[shared.SlotTwoHanded]; item != nil {
		if weapon, ok := item.(*equipment.Weapon); ok {
			if char.HasWeaponProficiency(weapon.Base.Key) {
				return " PROF"
			}
			return " !PROF"
		}
	}

	return ""
}
