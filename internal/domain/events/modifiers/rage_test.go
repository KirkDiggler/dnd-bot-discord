package modifiers_test

import (
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/events"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/events/modifiers"
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
			rage := modifiers.NewRageModifier(tc.level)

			// Create a melee damage roll event
			event := events.NewGameEvent(events.OnDamageRoll).
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

func (s *RageModifierSuite) TestRageOnlyAppliesToMelee() {
	rage := modifiers.NewRageModifier(5)

	// Ranged attack should not get rage bonus
	event := events.NewGameEvent(events.OnDamageRoll).
		WithContext("attack_type", "ranged").
		WithContext("damage", 10)

	s.False(rage.Condition(event))
}

func (s *RageModifierSuite) TestRagePhysicalResistance() {
	rage := modifiers.NewRageModifier(5)

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

func (s *RageModifierSuite) TestRageMetadata() {
	rage := modifiers.NewRageModifier(5)

	s.Equal("rage_5", rage.ID())
	s.Equal(100, rage.Priority()) // Feature priority

	source := rage.Source()
	s.Equal("ability", source.Type)
	s.Equal("Rage", source.Name)

	duration := rage.Duration()
	s.IsType(&events.RoundsDuration{}, duration)
}
