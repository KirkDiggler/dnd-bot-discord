package ability_test

import (
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/events"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/ability"
	"github.com/stretchr/testify/suite"
)

type RageEventAdapterSuite struct {
	suite.Suite
	eventBus *events.EventBus
	adapter  *ability.RageEventAdapter
}

func TestRageEventAdapterSuite(t *testing.T) {
	suite.Run(t, new(RageEventAdapterSuite))
}

func (s *RageEventAdapterSuite) SetupTest() {
	s.eventBus = events.NewEventBus()
	s.adapter = ability.NewRageEventAdapter(s.eventBus)
}

func (s *RageEventAdapterSuite) TestRageActivation() {
	// Create a test character
	char := &character.Character{
		ID:    "test-barbarian",
		Level: 5,
	}

	// Track status events
	var statusApplied bool
	statusListener := &testListener{
		handler: func(event *events.GameEvent) error {
			status, _ := event.GetStringContext("status")
			if status == "Rage" {
				statusApplied = true
			}
			return nil
		},
	}
	s.eventBus.Subscribe(events.OnStatusApplied, statusListener)

	// Activate rage
	err := s.adapter.ActivateRage(char)
	s.NoError(err)
	s.True(statusApplied)

	// Try to activate again - should fail
	err = s.adapter.ActivateRage(char)
	s.Error(err)
}

func (s *RageEventAdapterSuite) TestRageDamageBonus() {
	// Create a level 5 barbarian
	char := &character.Character{
		ID:    "test-barbarian",
		Level: 5, // Should have +2 rage damage
	}

	// Activate rage
	err := s.adapter.ActivateRage(char)
	s.NoError(err)

	// Simulate a melee damage roll
	damageEvent := events.NewGameEvent(events.OnDamageRoll).
		WithActor(char).
		WithContext("damage", 10).
		WithContext("attack_type", "melee")

	// Emit the event
	err = s.eventBus.Emit(damageEvent)
	s.NoError(err)

	// Check that damage was increased by rage bonus
	finalDamage, exists := damageEvent.GetIntContext("damage")
	s.True(exists)
	s.Equal(12, finalDamage) // 10 + 2 rage bonus
}

func (s *RageEventAdapterSuite) TestRageResistance() {
	// Create a barbarian
	char := &character.Character{
		ID:    "test-barbarian",
		Level: 5,
	}

	// Activate rage
	err := s.adapter.ActivateRage(char)
	s.NoError(err)

	// Simulate taking physical damage
	damageEvent := events.NewGameEvent(events.BeforeTakeDamage).
		WithTarget(char).
		WithContext("damage", 20).
		WithContext("damage_type", "slashing")

	// Emit the event
	err = s.eventBus.Emit(damageEvent)
	s.NoError(err)

	// Check that damage was halved
	finalDamage, exists := damageEvent.GetIntContext("damage")
	s.True(exists)
	s.Equal(10, finalDamage) // 20 / 2 resistance
}

func (s *RageEventAdapterSuite) TestRageDeactivation() {
	// Create a barbarian
	char := &character.Character{
		ID:    "test-barbarian",
		Level: 5,
	}

	// Activate rage
	err := s.adapter.ActivateRage(char)
	s.NoError(err)

	// Track status removal
	var statusRemoved bool
	statusListener := &testListener{
		handler: func(event *events.GameEvent) error {
			status, _ := event.GetStringContext("status")
			if status == "Rage" {
				statusRemoved = true
			}
			return nil
		},
	}
	s.eventBus.Subscribe(events.OnStatusRemoved, statusListener)

	// Deactivate rage
	err = s.adapter.DeactivateRage(char)
	s.NoError(err)
	s.True(statusRemoved)

	// After deactivation, rage bonus should not apply
	damageEvent := events.NewGameEvent(events.OnDamageRoll).
		WithActor(char).
		WithContext("damage", 10).
		WithContext("attack_type", "melee")

	err = s.eventBus.Emit(damageEvent)
	s.NoError(err)

	// Damage should remain unchanged
	finalDamage, exists := damageEvent.GetIntContext("damage")
	s.True(exists)
	s.Equal(10, finalDamage) // No rage bonus
}

// testListener is a simple event listener for testing
type testListener struct {
	handler func(event *events.GameEvent) error
}

func (tl *testListener) HandleEvent(event *events.GameEvent) error {
	if tl.handler != nil {
		return tl.handler(event)
	}
	return nil
}

func (tl *testListener) Priority() int {
	return 1000 // Low priority
}
