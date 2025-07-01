package events_test

import (
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/events"
	"github.com/stretchr/testify/suite"
)

type DurationSuite struct {
	suite.Suite
}

func TestDurationSuite(t *testing.T) {
	suite.Run(t, new(DurationSuite))
}

func (s *DurationSuite) TestPermanentDuration() {
	duration := events.PermanentDuration{}

	// Should never expire
	event := events.NewGameEvent(events.OnTurnStart)
	s.False(duration.IsExpired(event))

	event = events.NewGameEvent(events.OnLongRest)
	s.False(duration.IsExpired(event))

	s.Equal("Permanent", duration.Description())
}

func (s *DurationSuite) TestRoundsDuration() {
	duration := &events.RoundsDuration{
		Rounds:    3,
		StartTurn: 5,
	}

	// Not a turn event
	event := events.NewGameEvent(events.BeforeAttackRoll)
	s.False(duration.IsExpired(event))

	// Still within duration
	event = events.NewGameEvent(events.OnTurnStart).
		WithContext("turn_count", 7) // 2 turns passed
	s.False(duration.IsExpired(event))

	// Exactly at expiration
	event = events.NewGameEvent(events.OnTurnStart).
		WithContext("turn_count", 8) // 3 turns passed
	s.True(duration.IsExpired(event))

	// Past expiration
	event = events.NewGameEvent(events.OnTurnStart).
		WithContext("turn_count", 10)
	s.True(duration.IsExpired(event))

	s.Equal("3 rounds", duration.Description())
}

func (s *DurationSuite) TestEncounterDuration() {
	duration := events.EncounterDuration{}

	// Normal events don't expire it
	event := events.NewGameEvent(events.OnTurnStart)
	s.False(duration.IsExpired(event))

	// Wrong status type
	event = events.NewGameEvent(events.OnStatusRemoved).
		WithContext("status", "rage")
	s.False(duration.IsExpired(event))

	// Combat ended
	event = events.NewGameEvent(events.OnStatusRemoved).
		WithContext("status", "combat_ended")
	s.True(duration.IsExpired(event))

	s.Equal("Until end of encounter", duration.Description())
}

func (s *DurationSuite) TestConcentrationDuration() {
	duration := &events.ConcentrationDuration{
		SpellName: "Bless",
	}

	// Normal events don't expire it
	event := events.NewGameEvent(events.OnTurnStart)
	s.False(duration.IsExpired(event))

	// Wrong status type
	event = events.NewGameEvent(events.OnStatusRemoved).
		WithContext("status", "poisoned")
	s.False(duration.IsExpired(event))

	// Concentration broken
	event = events.NewGameEvent(events.OnStatusRemoved).
		WithContext("status", "concentration_broken")
	s.True(duration.IsExpired(event))

	s.Equal("Concentration (Bless)", duration.Description())
}

func (s *DurationSuite) TestShortRestDuration() {
	duration := events.ShortRestDuration{}

	// Normal events don't expire it
	event := events.NewGameEvent(events.OnTurnStart)
	s.False(duration.IsExpired(event))

	// Short rest expires it
	event = events.NewGameEvent(events.OnShortRest)
	s.True(duration.IsExpired(event))

	// Long rest also expires it
	event = events.NewGameEvent(events.OnLongRest)
	s.True(duration.IsExpired(event))

	s.Equal("Until short rest", duration.Description())
}

func (s *DurationSuite) TestLongRestDuration() {
	duration := events.LongRestDuration{}

	// Normal events don't expire it
	event := events.NewGameEvent(events.OnTurnStart)
	s.False(duration.IsExpired(event))

	// Short rest doesn't expire it
	event = events.NewGameEvent(events.OnShortRest)
	s.False(duration.IsExpired(event))

	// Long rest expires it
	event = events.NewGameEvent(events.OnLongRest)
	s.True(duration.IsExpired(event))

	s.Equal("Until long rest", duration.Description())
}

func (s *DurationSuite) TestUntilDamagedDuration() {
	duration := events.UntilDamagedDuration{}

	// Normal events don't expire it
	event := events.NewGameEvent(events.OnTurnStart)
	s.False(duration.IsExpired(event))

	// Taking damage expires it
	event = events.NewGameEvent(events.OnTakeDamage)
	s.True(duration.IsExpired(event))

	s.Equal("Until damaged", duration.Description())
}
