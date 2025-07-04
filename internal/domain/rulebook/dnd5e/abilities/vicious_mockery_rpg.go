package abilities

import (
	"context"

	"github.com/KirkDiggler/dnd-bot-discord/internal/adapters/rpgtoolkit"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	abilityService "github.com/KirkDiggler/dnd-bot-discord/internal/services/ability"
	rpgevents "github.com/KirkDiggler/rpg-toolkit/events"
)

// ViciousMockeryRPGListener listens for attack roll events and applies disadvantage
type ViciousMockeryRPGListener struct {
	rpgBus *rpgevents.Bus
}

// NewViciousMockeryRPGListener creates a new listener for vicious mockery effects
func NewViciousMockeryRPGListener(rpgBus *rpgevents.Bus) *ViciousMockeryRPGListener {
	listener := &ViciousMockeryRPGListener{
		rpgBus: rpgBus,
	}

	// Subscribe to attack roll events to apply disadvantage
	if rpgBus != nil {
		rpgBus.SubscribeFunc(rpgevents.EventBeforeAttackRoll, 10, listener.handleBeforeAttackRoll)
	}

	return listener
}

// handleBeforeAttackRoll checks if the attacker has vicious mockery disadvantage
func (v *ViciousMockeryRPGListener) handleBeforeAttackRoll(ctx context.Context, event rpgevents.Event) error {
	// Get the attacker
	attacker := event.Source()
	if attacker == nil {
		return nil
	}

	// Check if this is an entity with vicious mockery effect
	var hasViciousMockery bool

	switch adapter := attacker.(type) {
	case *rpgtoolkit.CharacterEntityAdapter:
		// Handle player characters
		if adapter.Character == nil {
			return nil
		}

		if adapter.Resources != nil && adapter.Resources.ActiveEffects != nil {
			for _, effect := range adapter.Resources.ActiveEffects {
				if effect.Name == "Vicious Mockery Disadvantage" && !effect.IsExpired() {
					hasViciousMockery = true
					break
				}
			}
		}

	case *rpgtoolkit.MonsterEntityAdapter:
		// Handle monsters
		if adapter.Combatant == nil {
			return nil
		}

		if adapter.ActiveEffects != nil {
			for _, effect := range adapter.ActiveEffects {
				if effect.Name == "Vicious Mockery Disadvantage" && !effect.IsExpired() {
					hasViciousMockery = true
					break
				}
			}
		}

	default:
		// Unknown entity type
		return nil
	}

	if hasViciousMockery {
		// Add disadvantage modifier to the event
		event.Context().AddModifier(rpgevents.NewModifier(
			"vicious_mockery",
			rpgevents.ModifierDisadvantage,
			rpgevents.IntValue(1), // 1 = has disadvantage
			100,                   // High priority to ensure it's applied
		))

		// Applied disadvantage from Vicious Mockery (removed excessive logging)

		// Remove the effect after it's used (it only affects the next attack)
		switch adapter := attacker.(type) {
		case *rpgtoolkit.CharacterEntityAdapter:
			if adapter.Resources != nil && adapter.Resources.ActiveEffects != nil {
				newEffects := []*shared.ActiveEffect{}
				for _, effect := range adapter.Resources.ActiveEffects {
					if effect.Name != "Vicious Mockery Disadvantage" {
						newEffects = append(newEffects, effect)
					}
				}
				adapter.Resources.ActiveEffects = newEffects
			}

		case *rpgtoolkit.MonsterEntityAdapter:
			if adapter.ActiveEffects != nil {
				newEffects := []*shared.ActiveEffect{}
				for _, effect := range adapter.ActiveEffects {
					if effect.Name != "Vicious Mockery Disadvantage" {
						newEffects = append(newEffects, effect)
					}
				}
				adapter.ActiveEffects = newEffects
			}
		}
	}

	return nil
}

// RegisterViciousMockeryHandler updates the registration to include RPG event listener
func RegisterViciousMockeryHandler(registry interface {
	RegisterHandler(handler abilityService.Handler)
}, cfg *RegistryConfig) {
	// Register the traditional handler
	viciousMockeryHandler := NewViciousMockeryHandler(cfg.EventBus)
	if cfg.DiceRoller != nil {
		viciousMockeryHandler.SetDiceRoller(cfg.DiceRoller)
	}
	if cfg.CharacterService != nil {
		viciousMockeryHandler.SetCharacterService(cfg.CharacterService)
	}
	registry.RegisterHandler(NewServiceHandlerAdapter(viciousMockeryHandler))

	// Also create the RPG event listener if we have an RPG event bus
	if cfg.RPGEventBus != nil {
		NewViciousMockeryRPGListener(cfg.RPGEventBus)
	}
}
