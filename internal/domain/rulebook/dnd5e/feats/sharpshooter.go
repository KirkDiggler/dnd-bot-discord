package feats

import (
	"context"
	"log"

	"github.com/KirkDiggler/dnd-bot-discord/internal/adapters/rpgtoolkit"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/equipment"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	rpgevents "github.com/KirkDiggler/rpg-toolkit/events"
)

// SharpshooterFeat implements the Sharpshooter feat
// - Attacking at long range doesn't impose disadvantage on ranged weapon attack rolls
// - Ranged weapon attacks ignore half and three-quarters cover
// - Before making a ranged attack, can choose to take -5 to hit for +10 damage
type SharpshooterFeat struct {
	BaseFeat
}

// NewSharpshooterFeat creates a new Sharpshooter feat
func NewSharpshooterFeat() Feat {
	return &SharpshooterFeat{
		BaseFeat: BaseFeat{
			key:  "sharpshooter",
			name: "Sharpshooter",
			description: "You have mastered ranged weapons and can make shots that others find impossible. " +
				"You can take -5 to attack rolls to gain +10 damage with ranged weapons.",
			prerequisites: []Prerequisite{
				// No prerequisites for Sharpshooter
			},
		},
	}
}

// RegisterHandlers registers event handlers for Sharpshooter
func (f *SharpshooterFeat) RegisterHandlers(bus *rpgevents.Bus, char *character.Character) {
	if bus == nil {
		return
	}

	// Subscribe to before attack roll events to apply -5/+10
	bus.SubscribeFunc(rpgevents.EventBeforeAttackRoll, 30, func(ctx context.Context, event rpgevents.Event) error {
		// Check if this is our character attacking
		actor, ok := rpgtoolkit.ExtractCharacter(event.Source())
		if !ok || actor == nil || actor.ID != char.ID {
			return nil
		}

		// Check if character has Sharpshooter feat
		hasSharpshooter := false
		for _, feature := range actor.Features {
			if feature.Key == "sharpshooter" && feature.Type == "feat" {
				hasSharpshooter = true
				break
			}
		}

		if !hasSharpshooter {
			return nil
		}

		// Check if using a ranged weapon
		weaponName, _ := rpgtoolkit.GetStringContext(event, rpgtoolkit.ContextWeapon)
		isRangedWeapon := false

		// Check equipped weapons for ranged property
		if actor.EquippedSlots != nil {
			// Check main hand
			if item := actor.EquippedSlots[shared.SlotMainHand]; item != nil {
				if weapon, ok := item.(*equipment.Weapon); ok && weapon.Base.Name == weaponName {
					// Check if weapon is ranged
					isRangedWeapon = weapon.IsRanged()
				}
			}

			// Check off hand for thrown weapons
			if !isRangedWeapon {
				if item := actor.EquippedSlots[shared.SlotOffHand]; item != nil {
					if weapon, ok := item.(*equipment.Weapon); ok && weapon.Base.Name == weaponName {
						isRangedWeapon = weapon.IsRanged()
					}
				}
			}
		}

		if !isRangedWeapon {
			return nil
		}

		// For now, we'll always apply sharpshooter when using a ranged weapon
		// In a full implementation, this would be a player choice
		// Apply -5 to attack roll
		currentBonus, _ := rpgtoolkit.GetIntContext(event, rpgtoolkit.ContextAttackBonus)
		event.Context().Set(rpgtoolkit.ContextAttackBonus, currentBonus-5)
		event.Context().Set("sharpshooter_active", true)

		// Remove cover penalties
		event.Context().Set("ignore_cover", true)

		log.Printf("[SHARPSHOOTER] %s takes -5 to attack for Sharpshooter", actor.Name)

		return nil
	})

	// Subscribe to damage roll events to apply +10 damage
	bus.SubscribeFunc(rpgevents.EventOnDamageRoll, 40, func(ctx context.Context, event rpgevents.Event) error {
		// Check if this is our character dealing damage
		actor, ok := rpgtoolkit.ExtractCharacter(event.Source())
		if !ok || actor == nil || actor.ID != char.ID {
			return nil
		}

		// Check if sharpshooter was active for this attack
		sharpshooterActive, _ := rpgtoolkit.GetBoolContext(event, "sharpshooter_active")
		if !sharpshooterActive {
			return nil
		}

		// Add +10 damage
		currentDamage, _ := rpgtoolkit.GetIntContext(event, rpgtoolkit.ContextDamage)
		event.Context().Set(rpgtoolkit.ContextDamage, currentDamage+10)
		event.Context().Set("damage_bonus_source", "Sharpshooter")

		log.Printf("[SHARPSHOOTER] %s gains +10 damage from Sharpshooter", actor.Name)

		return nil
	})

	// Subscribe to attack disadvantage check events to remove long range disadvantage
	bus.SubscribeFunc("dndbot.on_attack_disadvantage_check", 20, func(ctx context.Context, event rpgevents.Event) error {
		// Check if this is our character attacking
		actor, ok := rpgtoolkit.ExtractCharacter(event.Source())
		if !ok || actor == nil || actor.ID != char.ID {
			return nil
		}

		// Check if character has Sharpshooter feat
		hasSharpshooter := false
		for _, feature := range actor.Features {
			if feature.Key == "sharpshooter" && feature.Type == "feat" {
				hasSharpshooter = true
				break
			}
		}

		if !hasSharpshooter {
			return nil
		}

		// Check if disadvantage is due to long range
		disadvantageReason, _ := rpgtoolkit.GetStringContext(event, "disadvantage_reason")
		if disadvantageReason == "long_range" {
			// Cancel the disadvantage
			event.Cancel()
			log.Printf("[SHARPSHOOTER] %s ignores long range disadvantage", actor.Name)
		}

		return nil
	})
}
