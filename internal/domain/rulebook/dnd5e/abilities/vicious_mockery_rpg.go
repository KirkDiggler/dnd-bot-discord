package abilities

import (
	"context"
	"log"

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

	// Check if this is a character with vicious mockery effect
	charAdapter, ok := attacker.(*rpgtoolkit.CharacterEntityAdapter)
	if !ok || charAdapter.Character == nil {
		return nil
	}

	char := charAdapter.Character

	// Check if character has vicious mockery effect
	hasViciousMockery := false
	if char.Resources != nil && char.Resources.ActiveEffects != nil {
		for _, effect := range char.Resources.ActiveEffects {
			if effect.Name == "Vicious Mockery Disadvantage" && !effect.IsExpired() {
				hasViciousMockery = true
				break
			}
		}
	}

	if hasViciousMockery {
		// Add disadvantage modifier to the event
		event.Context().AddModifier(rpgevents.NewModifier(
			"vicious_mockery",
			rpgevents.ModifierDisadvantage,
			rpgevents.IntValue(1), // 1 = has disadvantage
			100,                   // High priority to ensure it's applied
		))

		log.Printf("Applied disadvantage from Vicious Mockery to %s's attack", char.Name)

		// Remove the effect after it's used (it only affects the next attack)
		if char.Resources != nil && char.Resources.ActiveEffects != nil {
			newEffects := []*shared.ActiveEffect{}
			for _, effect := range char.Resources.ActiveEffects {
				if effect.Name != "Vicious Mockery Disadvantage" {
					newEffects = append(newEffects, effect)
				}
			}
			char.Resources.ActiveEffects = newEffects
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
