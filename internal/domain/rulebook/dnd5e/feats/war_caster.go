package feats

import (
	"context"
	"log"

	"github.com/KirkDiggler/dnd-bot-discord/internal/adapters/rpgtoolkit"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	rpgevents "github.com/KirkDiggler/rpg-toolkit/events"
)

// WarCasterFeat implements the War Caster feat
// - Advantage on Constitution saves to maintain concentration
// - Can perform somatic components even with weapons/shield in hands
// - Can cast a spell as opportunity attack (instead of melee)
type WarCasterFeat struct {
	BaseFeat
}

// NewWarCasterFeat creates a new War Caster feat
func NewWarCasterFeat() Feat {
	return &WarCasterFeat{
		BaseFeat: BaseFeat{
			key:  "war_caster",
			name: "War Caster",
			description: "You have practiced casting spells in combat. You gain advantage on Constitution " +
				"saves to maintain concentration, can cast with weapons/shield in hand, and can cast spells " +
				"as opportunity attacks.",
			prerequisites: []Prerequisite{
				{
					Type:        "spellcasting",
					Requirement: "The ability to cast at least one spell",
					Check: func(char *character.Character) bool {
						// Check if character has any spellcasting ability
						// This would need to check class features or spell lists
						// For now, check if character has spell slots
						if char.Resources != nil && char.Resources.SpellSlots != nil {
							return len(char.Resources.SpellSlots) > 0
						}
						return false
					},
				},
			},
		},
	}
}

// RegisterHandlers registers event handlers for War Caster
func (f *WarCasterFeat) RegisterHandlers(bus *rpgevents.Bus, char *character.Character) {
	if bus == nil {
		return
	}

	// Subscribe to concentration save events
	bus.SubscribeFunc(rpgevents.EventBeforeSavingThrow, 20, func(ctx context.Context, event rpgevents.Event) error {
		// Check if this is our character making a save
		actor, ok := rpgtoolkit.ExtractCharacter(event.Source())
		if !ok || actor == nil || actor.ID != char.ID {
			return nil
		}

		// Check if character has War Caster feat
		hasWarCaster := false
		for _, feature := range actor.Features {
			if feature.Key == "war_caster" && feature.Type == "feat" {
				hasWarCaster = true
				break
			}
		}

		if !hasWarCaster {
			return nil
		}

		// Check if this is a concentration save
		saveType, _ := rpgtoolkit.GetStringContext(event, "save_type")
		saveReason, _ := rpgtoolkit.GetStringContext(event, "save_reason")

		if saveType == "constitution" && saveReason == "concentration" {
			// Grant advantage on the save
			event.Context().Set("has_advantage", true)
			event.Context().Set("advantage_source", "War Caster feat")

			log.Printf("[WAR CASTER] %s gains advantage on concentration save", actor.Name)
		}

		return nil
	})

	// Subscribe to opportunity attack events
	bus.SubscribeFunc("dndbot.on_opportunity_attack", 30, func(ctx context.Context, event rpgevents.Event) error {
		// Check if this is our character making an opportunity attack
		actor, ok := rpgtoolkit.ExtractCharacter(event.Source())
		if !ok || actor == nil || actor.ID != char.ID {
			return nil
		}

		// Check if character has War Caster feat
		hasWarCaster := false
		for _, feature := range actor.Features {
			if feature.Key == "war_caster" && feature.Type == "feat" {
				hasWarCaster = true
				break
			}
		}

		if !hasWarCaster {
			return nil
		}

		// Check if character has spells available
		hasSpells := false
		if actor.Resources != nil && actor.Resources.SpellSlots != nil {
			for _, slot := range actor.Resources.SpellSlots {
				if slot.Remaining > 0 {
					hasSpells = true
					break
				}
			}
		}

		if hasSpells {
			// Emit event to allow spell casting as opportunity attack
			spellOppEvent := rpgevents.NewGameEvent(
				"dndbot.spell_opportunity_attack",
				rpgtoolkit.WrapCharacter(actor),
				event.Target(),
			)
			spellOppEvent.Context().Set("source", "War Caster")

			if err := bus.Publish(ctx, spellOppEvent); err != nil {
				log.Printf("[WAR CASTER] Failed to publish spell opportunity event: %v", err)
			}

			log.Printf("[WAR CASTER] %s can cast a spell as opportunity attack", actor.Name)
		}

		return nil
	})

	// Subscribe to spell component check events
	bus.SubscribeFunc("dndbot.on_spell_component_check", 10, func(ctx context.Context, event rpgevents.Event) error {
		// Check if this is our character casting a spell
		actor, ok := rpgtoolkit.ExtractCharacter(event.Source())
		if !ok || actor == nil || actor.ID != char.ID {
			return nil
		}

		// Check if character has War Caster feat
		hasWarCaster := false
		for _, feature := range actor.Features {
			if feature.Key == "war_caster" && feature.Type == "feat" {
				hasWarCaster = true
				break
			}
		}

		if !hasWarCaster {
			return nil
		}

		// Check if the issue is somatic components with hands full
		componentType, _ := rpgtoolkit.GetStringContext(event, "component_type")
		handsFull, _ := rpgtoolkit.GetBoolContext(event, "hands_full")

		if componentType == "somatic" && handsFull {
			// War Caster allows casting with hands full
			event.Context().Set("component_satisfied", true)
			event.Context().Set("satisfied_by", "War Caster feat")

			log.Printf("[WAR CASTER] %s can perform somatic components with hands full", actor.Name)
		}

		return nil
	})
}
