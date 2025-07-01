package features_test

import (
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/events"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e/features"
	"github.com/stretchr/testify/suite"
)

type RageModifierSuite struct {
	suite.Suite
}

func TestRageModifierSuite(t *testing.T) {
	suite.Run(t, new(RageModifierSuite))
}

func (s *RageModifierSuite) TestRageDamageBonus() {
	testCases := []struct {
		name      string
		level     int
		wantBonus int
	}{
		{"Level 1", 1, 2},
		{"Level 8", 8, 2},
		{"Level 9", 9, 3},
		{"Level 15", 15, 3},
		{"Level 16", 16, 4},
		{"Level 20", 20, 4},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			characterID := "test-barbarian"
			rage := features.NewRageModifier(characterID, tc.level, 0)

			// Create character for event
			char := &character.Character{ID: characterID}

			// Create a melee damage roll event
			event := events.NewGameEvent(events.OnDamageRoll).
				WithActor(char).
				WithContext("attack_type", "melee").
				WithContext("damage", 10)

			// Check condition
			s.True(rage.Condition(event))

			// Apply modifier
			err := rage.Apply(event)
			s.NoError(err)

			// Check damage was increased
			damage, exists := event.GetIntContext("damage")
			s.True(exists)
			s.Equal(10+tc.wantBonus, damage)
		})
	}
}

func (s *RageModifierSuite) TestRageOnlyAppliesToCorrectCharacter() {
	characterID := "barbarian-1"
	rage := features.NewRageModifier(characterID, 5, 0)

	// Different character's attack should not get rage bonus
	otherChar := &character.Character{ID: "other-character"}
	event := events.NewGameEvent(events.OnDamageRoll).
		WithActor(otherChar).
		WithContext("attack_type", "melee").
		WithContext("damage", 10)

	s.False(rage.Condition(event))
}

func (s *RageModifierSuite) TestRageOnlyAppliesToMelee() {
	characterID := "test-barbarian"
	rage := features.NewRageModifier(characterID, 5, 0)
	char := &character.Character{ID: characterID}

	// Ranged attack should not get rage bonus
	event := events.NewGameEvent(events.OnDamageRoll).
		WithActor(char).
		WithContext("attack_type", "ranged").
		WithContext("damage", 10)

	s.False(rage.Condition(event))
}

func (s *RageModifierSuite) TestRagePhysicalResistance() {
	characterID := "test-barbarian"
	rage := features.NewRageModifier(characterID, 5, 0)
	char := &character.Character{ID: characterID}

	damageTypes := []struct {
		damageType   string
		shouldResist bool
	}{
		{"bludgeoning", true},
		{"piercing", true},
		{"slashing", true},
		{"fire", false},
		{"cold", false},
		{"psychic", false},
	}

	for _, dt := range damageTypes {
		s.Run(dt.damageType, func() {
			event := events.NewGameEvent(events.BeforeTakeDamage).
				WithTarget(char).
				WithContext("damage_type", dt.damageType).
				WithContext("damage", 20)

			if dt.shouldResist {
				s.True(rage.Condition(event))

				err := rage.Apply(event)
				s.NoError(err)

				// Check damage was halved
				damage, exists := event.GetIntContext("damage")
				s.True(exists)
				s.Equal(10, damage)
			} else {
				s.False(rage.Condition(event))
			}
		})
	}
}

func (s *RageModifierSuite) TestRageListener() {
	characterID := "test-barbarian"
	listener := features.NewRageListener(characterID, 5, 0)
	char := &character.Character{ID: characterID}

	// Test that listener properly wraps modifier
	event := events.NewGameEvent(events.OnDamageRoll).
		WithActor(char).
		WithContext("attack_type", "melee").
		WithContext("damage", 10)

	err := listener.HandleEvent(event)
	s.NoError(err)

	// Check damage was increased
	damage, exists := event.GetIntContext("damage")
	s.True(exists)
	s.Equal(12, damage) // 10 + 2 rage bonus

	// Check listener metadata
	s.Equal(100, listener.Priority())
	s.Equal("rage_test-barbarian", listener.ID())
}
