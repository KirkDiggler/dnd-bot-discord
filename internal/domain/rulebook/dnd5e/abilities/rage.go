package abilities

import (
	"context"
	"fmt"
	"log"

	"github.com/KirkDiggler/dnd-bot-discord/internal/adapters/rpgtoolkit"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"github.com/KirkDiggler/dnd-bot-discord/internal/effects"
	charService "github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	rpgevents "github.com/KirkDiggler/rpg-toolkit/events"
)

// RageHandler handles the barbarian rage ability
type RageHandler struct {
	eventBus         *rpgevents.Bus
	characterService charService.Service
}

// NewRageHandler creates a new rage handler
func NewRageHandler(eventBus *rpgevents.Bus, characterService charService.Service) Handler {
	return &RageHandler{
		eventBus:         eventBus,
		characterService: characterService,
	}
}

// Key returns the ability key
func (h *RageHandler) Key() string {
	return shared.AbilityKeyRage
}

// Execute activates or deactivates rage
func (h *RageHandler) Execute(ctx context.Context, char *character.Character, ability *shared.ActiveAbility, input *Input) (*Result, error) {
	if char == nil || ability == nil {
		return nil, fmt.Errorf("character and ability cannot be nil")
	}

	// Check if rage is already active
	isActive := ability.IsActive

	if isActive {
		// Deactivate rage
		ability.IsActive = false
		ability.Duration = 0

		// Remove rage effects
		h.removeRageEffects(char)

		// Save character state
		if err := h.characterService.UpdateEquipment(char); err != nil {
			log.Printf("Failed to save character after deactivating rage: %v", err)
		}

		// Emit rage deactivated event
		if h.eventBus != nil {
			deactivateEvent := rpgevents.NewGameEvent(
				"dndbot.rage_deactivated",
				rpgtoolkit.WrapCharacter(char),
				nil,
			)
			if err := h.eventBus.Publish(ctx, deactivateEvent); err != nil {
				log.Printf("Failed to publish rage deactivated event: %v", err)
			}
		}

		return &Result{
			Success:       true,
			Message:       fmt.Sprintf("%s is no longer raging", char.Name),
			UsesRemaining: ability.UsesRemaining,
			EffectApplied: false,
		}, nil
	}

	// Check if character has uses remaining
	if ability.UsesRemaining <= 0 {
		return &Result{
			Success:       false,
			Message:       "No rage uses remaining. Take a long rest to recover.",
			UsesRemaining: 0,
		}, nil
	}

	// Activate rage
	ability.IsActive = true
	ability.Duration = 10

	// Add rage effects
	rageEffect := effects.BuildRageEffect(char.Level)
	if rageEffect != nil {
		// Convert StatusEffect to ActiveEffect
		activeEffect := &shared.ActiveEffect{
			ID:                    rageEffect.ID,
			Name:                  rageEffect.Name,
			Description:           rageEffect.Description,
			Source:                string(rageEffect.Source),
			SourceID:              rageEffect.SourceID,
			Duration:              rageEffect.Duration.Rounds,
			DurationType:          shared.DurationTypeRounds,
			Modifiers:             []shared.Modifier{},
			RequiresConcentration: rageEffect.Duration.Concentration,
		}

		// Convert modifiers
		for _, mod := range rageEffect.Modifiers {
			// Map damage bonus
			if mod.Target == effects.TargetDamage {
				activeEffect.Modifiers = append(activeEffect.Modifiers, shared.Modifier{
					Type:  shared.ModifierTypeDamageBonus,
					Value: h.getRageDamageBonus(char.Level),
				})
			}
			// Map resistances
			if mod.Target == effects.TargetResistance && mod.Value == "resistance" {
				activeEffect.Modifiers = append(activeEffect.Modifiers, shared.Modifier{
					Type:        shared.ModifierTypeDamageResistance,
					Value:       1, // Resistance is boolean, represented as 1
					DamageTypes: []string{mod.DamageType},
				})
			}
		}

		char.Resources.AddEffect(activeEffect)
	}

	// Save character state
	if err := h.characterService.UpdateEquipment(char); err != nil {
		log.Printf("Failed to save character after activating rage: %v", err)
	}

	// Register event handlers for rage benefits
	if h.eventBus != nil {
		h.registerRageHandlers(char)

		// Emit rage activated event
		activateEvent := rpgevents.NewGameEvent(
			"dndbot.rage_activated",
			rpgtoolkit.WrapCharacter(char),
			nil,
		)
		activateEvent.Context().Set("duration", 10) // 10 rounds
		activateEvent.Context().Set("damage_bonus", h.getRageDamageBonus(char.Level))

		if err := h.eventBus.Publish(ctx, activateEvent); err != nil {
			log.Printf("Failed to publish rage activated event: %v", err)
		}
	}

	return &Result{
		Success:       true,
		Message:       fmt.Sprintf("%s enters a rage! (+%d melee damage, resistance to physical damage)", char.Name, h.getRageDamageBonus(char.Level)),
		UsesRemaining: ability.UsesRemaining,
		EffectApplied: true,
		EffectName:    "Rage",
		Duration:      10, // 10 rounds
		DamageBonus:   h.getRageDamageBonus(char.Level),
	}, nil
}

// removeRageEffects removes all rage effects from the character
func (h *RageHandler) removeRageEffects(char *character.Character) {
	if char.Resources == nil {
		return
	}

	// Remove rage from active effects
	var newEffects []*shared.ActiveEffect
	for _, effect := range char.Resources.ActiveEffects {
		if effect.Name != "Rage" {
			newEffects = append(newEffects, effect)
		}
	}
	char.Resources.ActiveEffects = newEffects
}

// getRageDamageBonus returns the rage damage bonus based on level
func (h *RageHandler) getRageDamageBonus(level int) int {
	if level >= 16 {
		return 4
	} else if level >= 9 {
		return 3
	}
	return 2
}

// registerRageHandlers registers event handlers for rage benefits
func (h *RageHandler) registerRageHandlers(char *character.Character) {
	if h.eventBus == nil {
		return
	}

	// Subscribe to damage roll events to add rage bonus
	h.eventBus.SubscribeFunc(rpgevents.EventOnDamageRoll, 50, func(ctx context.Context, event rpgevents.Event) error {
		// Check if this is our character dealing damage
		actor, ok := rpgtoolkit.ExtractCharacter(event.Source())
		if !ok || actor == nil || actor.ID != char.ID {
			return nil
		}

		// Check if character is raging
		ability := char.Resources.Abilities[shared.AbilityKeyRage]
		if ability == nil || !ability.IsActive {
			return nil
		}

		// Check if it's a melee weapon attack
		weaponType, _ := rpgtoolkit.GetStringContext(event, "weapon_type")
		attackType, _ := rpgtoolkit.GetStringContext(event, "attack_type")

		if weaponType == "melee" || attackType == "melee" {
			// Add rage damage bonus
			currentDamage, _ := rpgtoolkit.GetIntContext(event, rpgtoolkit.ContextDamage)
			rageBonus := h.getRageDamageBonus(char.Level)
			event.Context().Set(rpgtoolkit.ContextDamage, currentDamage+rageBonus)
			event.Context().Set("rage_damage_applied", true)

			log.Printf("[RAGE] %s gains +%d damage from rage", char.Name, rageBonus)
		}

		return nil
	})

	// Subscribe to before take damage events for resistance
	h.eventBus.SubscribeFunc(rpgevents.EventBeforeTakeDamage, 30, func(ctx context.Context, event rpgevents.Event) error {
		// Check if this is our character taking damage
		target, ok := rpgtoolkit.ExtractCharacter(event.Target())
		if !ok || target == nil || target.ID != char.ID {
			return nil
		}

		// Check if character is raging
		ability := char.Resources.Abilities[shared.AbilityKeyRage]
		if ability == nil || !ability.IsActive {
			return nil
		}

		// Check damage type
		damageType, _ := rpgtoolkit.GetStringContext(event, "damage_type")
		if damageType == "bludgeoning" || damageType == "piercing" || damageType == "slashing" {
			// Apply resistance (half damage)
			currentDamage, _ := rpgtoolkit.GetIntContext(event, rpgtoolkit.ContextDamage)
			resistedDamage := currentDamage / 2
			event.Context().Set(rpgtoolkit.ContextDamage, resistedDamage)
			event.Context().Set("resistance_applied", true)
			event.Context().Set("resistance_source", "Rage")

			log.Printf("[RAGE] %s takes %d damage (reduced from %d by rage resistance)",
				target.Name, resistedDamage, currentDamage)
		}

		return nil
	})
}
