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

// GreatWeaponMasterFeat implements the Great Weapon Master feat
// - On critical hit or reducing a creature to 0 HP, can make one melee weapon attack as a bonus action
// - Before making a melee attack with a heavy weapon, can choose to take -5 to hit for +10 damage
type GreatWeaponMasterFeat struct {
	BaseFeat
}

// NewGreatWeaponMasterFeat creates a new Great Weapon Master feat
func NewGreatWeaponMasterFeat() Feat {
	return &GreatWeaponMasterFeat{
		BaseFeat: BaseFeat{
			key:  "great_weapon_master",
			name: "Great Weapon Master",
			description: "You've learned to put the weight of a weapon to your advantage. " +
				"You can take -5 to attack rolls to gain +10 damage with heavy weapons.",
			prerequisites: []Prerequisite{
				// No prerequisites for Great Weapon Master
			},
		},
	}
}

// RegisterHandlers registers event handlers for Great Weapon Master
func (f *GreatWeaponMasterFeat) RegisterHandlers(bus *rpgevents.Bus, char *character.Character) {
	if bus == nil {
		return
	}

	// Track if GWM power attack is active for this character
	gwmActive := false

	// Subscribe to before attack roll events to apply -5/+10
	bus.SubscribeFunc(rpgevents.EventBeforeAttackRoll, 30, func(ctx context.Context, event rpgevents.Event) error {
		// Check if this is our character attacking
		actor, ok := rpgtoolkit.ExtractCharacter(event.Source())
		if !ok || actor == nil || actor.ID != char.ID {
			return nil
		}

		// Check if character has GWM feat
		hasGWM := false
		for _, feature := range actor.Features {
			if feature.Key == "great_weapon_master" && feature.Type == "feat" {
				hasGWM = true
				break
			}
		}

		if !hasGWM {
			return nil
		}

		// Check if using a heavy weapon
		weaponName, _ := rpgtoolkit.GetStringContext(event, rpgtoolkit.ContextWeapon)
		isHeavyWeapon := false

		// Check equipped weapons for heavy property
		if actor.EquippedSlots != nil {
			// Check main hand
			if item := actor.EquippedSlots[shared.SlotMainHand]; item != nil {
				if weapon, ok := item.(*equipment.Weapon); ok && weapon.Base.Name == weaponName {
					// Check if weapon has heavy property
					if weapon.IsHeavy() {
						isHeavyWeapon = true
					}
				}
			}

			// Check two-handed slot
			if !isHeavyWeapon {
				if item := actor.EquippedSlots[shared.SlotTwoHanded]; item != nil {
					if weapon, ok := item.(*equipment.Weapon); ok && weapon.Base.Name == weaponName {
						if weapon.IsHeavy() {
							isHeavyWeapon = true
						}
					}
				}
			}
		}

		if !isHeavyWeapon {
			return nil
		}

		// For now, we'll always apply GWM when using a heavy weapon
		// In a full implementation, this would be a player choice
		if gwmActive || true { // Always active for now
			// Apply -5 to attack roll
			currentBonus, _ := rpgtoolkit.GetIntContext(event, rpgtoolkit.ContextAttackBonus)
			event.Context().Set(rpgtoolkit.ContextAttackBonus, currentBonus-5)
			event.Context().Set("gwm_active", true)

			log.Printf("[GWM] %s takes -5 to attack for Great Weapon Master", actor.Name)
		}

		return nil
	})

	// Subscribe to damage roll events to apply +10 damage
	bus.SubscribeFunc(rpgevents.EventOnDamageRoll, 40, func(ctx context.Context, event rpgevents.Event) error {
		// Check if this is our character dealing damage
		actor, ok := rpgtoolkit.ExtractCharacter(event.Source())
		if !ok || actor == nil || actor.ID != char.ID {
			return nil
		}

		// Check if GWM was active for this attack
		gwmActive, _ := rpgtoolkit.GetBoolContext(event, "gwm_active")
		if !gwmActive {
			return nil
		}

		// Add +10 damage
		currentDamage, _ := rpgtoolkit.GetIntContext(event, rpgtoolkit.ContextDamage)
		event.Context().Set(rpgtoolkit.ContextDamage, currentDamage+10)
		event.Context().Set("damage_bonus_source", "Great Weapon Master")

		log.Printf("[GWM] %s gains +10 damage from Great Weapon Master", actor.Name)

		return nil
	})

	// Subscribe to after damage events to check for bonus action attack
	bus.SubscribeFunc(rpgevents.EventAfterDamageRoll, 60, func(ctx context.Context, event rpgevents.Event) error {
		// Check if this is our character dealing damage
		actor, ok := rpgtoolkit.ExtractCharacter(event.Source())
		if !ok || actor == nil || actor.ID != char.ID {
			return nil
		}

		// Check if it was a critical hit
		isCritical, _ := rpgtoolkit.GetBoolContext(event, rpgtoolkit.ContextIsCritical)

		// Check if target was reduced to 0 HP
		targetKilled := false
		if target, ok := rpgtoolkit.ExtractCharacter(event.Target()); ok && target != nil {
			// Check if damage reduced target to 0 HP
			damage, _ := rpgtoolkit.GetIntContext(event, rpgtoolkit.ContextDamage)
			if target.CurrentHitPoints <= damage {
				targetKilled = true
			}
		}

		if isCritical || targetKilled {
			// Grant bonus action attack
			// This would need to be communicated to the UI layer
			log.Printf("[GWM] %s can make a bonus action attack from Great Weapon Master", actor.Name)

			// Emit a custom event that the UI can listen for
			bonusActionEvent := rpgevents.NewGameEvent(
				"dndbot.bonus_action_granted",
				rpgtoolkit.WrapCharacter(actor),
				nil,
			)
			bonusActionEvent.Context().Set("source", "Great Weapon Master")
			bonusActionEvent.Context().Set("action_type", "melee_attack")

			if err := bus.Publish(ctx, bonusActionEvent); err != nil {
				log.Printf("[GWM] Failed to publish bonus action event: %v", err)
			}
		}

		return nil
	})
}
