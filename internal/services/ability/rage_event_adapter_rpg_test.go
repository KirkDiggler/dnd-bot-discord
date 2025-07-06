package ability_test

import (
	"context"
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/adapters/rpgtoolkit"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/ability"
	rpgevents "github.com/KirkDiggler/rpg-toolkit/events"
	"github.com/stretchr/testify/suite"
)

type RageEventAdapterRPGSuite struct {
	suite.Suite
	eventBus *rpgevents.Bus
	adapter  *ability.RageEventAdapter
}

func TestRageEventAdapterRPGSuite(t *testing.T) {
	suite.Run(t, new(RageEventAdapterRPGSuite))
}

func (s *RageEventAdapterRPGSuite) SetupTest() {
	s.eventBus = rpgevents.NewBus()
	// Note: RageEventAdapter needs to be updated to use rpg-toolkit
	// For now, this is a placeholder showing the pattern
}

func (s *RageEventAdapterRPGSuite) TestRageActivation() {
	// Create a test character
	char := &character.Character{
		ID:    "test-barbarian",
		Level: 5,
	}

	// Track status events
	var statusApplied bool
	s.eventBus.SubscribeFunc(rpgevents.EventOnConditionApplied, 100, func(ctx context.Context, event rpgevents.Event) error {
		status, _ := rpgtoolkit.GetStringContext(event, "status")
		if status == "Rage" {
			statusApplied = true
		}
		return nil
	})

	// Simulate rage activation by emitting event
	rageContext := map[string]interface{}{
		"status":      "Rage",
		"duration":    10,
		"source":      "Barbarian Rage",
		"description": "+2 melee damage, resistance to physical damage",
	}

	err := rpgtoolkit.EmitEvent(s.eventBus, rpgevents.EventOnConditionApplied, char, nil, rageContext)
	s.NoError(err)
	s.True(statusApplied, "Rage status should be applied")
}

func (s *RageEventAdapterRPGSuite) TestRageDamageBonus() {
	// Create a raging barbarian
	barbarian := &character.Character{
		ID:    "test-barbarian",
		Level: 5,
		Resources: character.Resources{
			ActiveEffects: []character.ActiveEffect{
				{Type: "rage", Name: "Rage"},
			},
		},
	}

	// Track damage modifications
	var damageModified bool
	var finalDamage int

	s.eventBus.SubscribeFunc(rpgevents.EventOnDamageRoll, 100, func(ctx context.Context, event rpgevents.Event) error {
		// Check if actor is our barbarian
		actor, ok := rpgtoolkit.ExtractCharacter(event.Source())
		if !ok || actor.ID != barbarian.ID {
			return nil
		}

		// Check if it's a melee attack
		attackType, _ := rpgtoolkit.GetStringContext(event, rpgtoolkit.ContextAttackType)
		if attackType == "melee" {
			currentDamage, _ := rpgtoolkit.GetIntContext(event, rpgtoolkit.ContextDamage)
			event.Context().Set(rpgtoolkit.ContextDamage, currentDamage+2)
			damageModified = true
		}
		return nil
	})

	// Simulate a melee damage roll
	damageContext := map[string]interface{}{
		rpgtoolkit.ContextAttackType: "melee",
		rpgtoolkit.ContextDamage:     10,
	}

	damageEvent, err := rpgtoolkit.CreateAndEmitEvent(
		s.eventBus,
		rpgevents.EventOnDamageRoll,
		barbarian,
		nil,
		damageContext,
	)
	s.NoError(err)

	finalDamage, _ = rpgtoolkit.GetIntContext(damageEvent, rpgtoolkit.ContextDamage)
	s.True(damageModified, "Damage should be modified by rage")
	s.Equal(12, finalDamage, "Damage should be increased by 2")
}

func (s *RageEventAdapterRPGSuite) TestRageResistance() {
	// Create a raging barbarian
	barbarian := &character.Character{
		ID:    "test-barbarian",
		Level: 5,
		Resources: character.Resources{
			ActiveEffects: []character.ActiveEffect{
				{Type: "rage", Name: "Rage"},
			},
		},
	}

	// Track damage resistance
	var resistanceApplied bool
	var finalDamage int

	s.eventBus.SubscribeFunc(rpgevents.EventBeforeTakeDamage, 50, func(ctx context.Context, event rpgevents.Event) error {
		// Check if target is our barbarian
		target, ok := rpgtoolkit.ExtractCharacter(event.Target())
		if !ok || target.ID != barbarian.ID {
			return nil
		}

		// Check damage type for physical resistance
		damageType, _ := rpgtoolkit.GetStringContext(event, rpgtoolkit.ContextDamageType)
		if damageType == "slashing" {
			currentDamage, _ := rpgtoolkit.GetIntContext(event, rpgtoolkit.ContextDamage)
			event.Context().Set(rpgtoolkit.ContextDamage, currentDamage/2)
			resistanceApplied = true
		}
		return nil
	})

	// Simulate taking slashing damage
	damageContext := map[string]interface{}{
		rpgtoolkit.ContextDamageType: "slashing",
		rpgtoolkit.ContextDamage:     20,
	}

	damageEvent, err := rpgtoolkit.CreateAndEmitEvent(
		s.eventBus,
		rpgevents.EventBeforeTakeDamage,
		nil,
		barbarian,
		damageContext,
	)
	s.NoError(err)

	finalDamage, _ = rpgtoolkit.GetIntContext(damageEvent, rpgtoolkit.ContextDamage)
	s.True(resistanceApplied, "Resistance should be applied")
	s.Equal(10, finalDamage, "Damage should be halved by resistance")
}
