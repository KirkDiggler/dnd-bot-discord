package events_test

import (
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/events"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/events/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type DurationTestSuite struct {
	suite.Suite
	ctrl  *gomock.Controller
	actor *character.Character
}

func (s *DurationTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.actor = &character.Character{ID: "test-actor"}
}

func (s *DurationTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func TestDurationSuite(t *testing.T) {
	suite.Run(t, new(DurationTestSuite))
}

// PermanentDuration Tests

func (s *DurationTestSuite) TestPermanentDurationNeverExpires() {
	// Setup
	duration := &events.PermanentDuration{}

	// Assert - should never expire
	s.False(duration.IsExpired())

	// Send various events
	eventTypes := []events.EventType{
		events.OnTurnEnd,
		events.OnLongRest,
		events.OnShortRest,
		events.OnTakeDamage,
	}

	for _, eventType := range eventTypes {
		event := events.NewGameEvent(eventType, s.actor)
		duration.OnEventOccurred(event)

		// Still not expired
		s.False(duration.IsExpired())
	}
}

// TurnDuration Tests

func (s *DurationTestSuite) TestTurnDurationExpiresAfterTurns() {
	// Setup - 3 turn duration
	duration := events.NewTurnDuration(3, false)

	// Initial state
	s.False(duration.IsExpired())

	// First turn
	event := events.NewGameEvent(events.OnTurnEnd, s.actor)
	duration.OnEventOccurred(event)
	s.False(duration.IsExpired())

	// Second turn
	duration.OnEventOccurred(event)
	s.False(duration.IsExpired())

	// Third turn - should expire
	duration.OnEventOccurred(event)
	s.True(duration.IsExpired())
}

func (s *DurationTestSuite) TestTurnDurationOnlyCountsTurnEndEvents() {
	// Setup
	duration := events.NewTurnDuration(2, false)

	// Send non-turn-end events
	attackEvent := events.NewGameEvent(events.OnAttackRoll, s.actor)
	duration.OnEventOccurred(attackEvent)
	s.False(duration.IsExpired())

	damageEvent := events.NewGameEvent(events.OnDamageRoll, s.actor)
	duration.OnEventOccurred(damageEvent)
	s.False(duration.IsExpired())

	// Only turn end events should count
	turnEndEvent := events.NewGameEvent(events.OnTurnEnd, s.actor)
	duration.OnEventOccurred(turnEndEvent)
	s.False(duration.IsExpired()) // Still 1 turn left

	duration.OnEventOccurred(turnEndEvent)
	s.True(duration.IsExpired()) // Now expired
}

func (s *DurationTestSuite) TestTurnDurationWithCombatExtension() {
	// Setup - 2 turn duration that extends on combat
	duration := events.NewTurnDuration(2, true)

	// First turn
	turnEnd := events.NewGameEvent(events.OnTurnEnd, s.actor)
	duration.OnEventOccurred(turnEnd)
	s.False(duration.IsExpired())

	// Attack resets the counter
	attack := events.NewGameEvent(events.OnAttackRoll, s.actor)
	duration.OnEventOccurred(attack)

	// Now need 2 more turns to expire
	duration.OnEventOccurred(turnEnd)
	s.False(duration.IsExpired())

	duration.OnEventOccurred(turnEnd)
	s.True(duration.IsExpired())
}

func (s *DurationTestSuite) TestTurnDurationDamageRollAlsoExtends() {
	// Setup
	duration := events.NewTurnDuration(1, true)

	// Damage roll should reset counter
	damage := events.NewGameEvent(events.OnDamageRoll, s.actor)
	duration.OnEventOccurred(damage)

	// Still need 1 turn to expire
	turnEnd := events.NewGameEvent(events.OnTurnEnd, s.actor)
	duration.OnEventOccurred(turnEnd)
	s.True(duration.IsExpired())
}

// UntilEventDuration Tests

func (s *DurationTestSuite) TestUntilEventDurationExpiresOnTargetEvent() {
	// Setup - expires on long rest
	duration := events.NewUntilEventDuration(events.OnLongRest)

	// Initial state
	s.False(duration.IsExpired())

	// Other events don't expire it
	turnEnd := events.NewGameEvent(events.OnTurnEnd, s.actor)
	duration.OnEventOccurred(turnEnd)
	s.False(duration.IsExpired())

	shortRest := events.NewGameEvent(events.OnShortRest, s.actor)
	duration.OnEventOccurred(shortRest)
	s.False(duration.IsExpired())

	// Target event expires it
	longRest := events.NewGameEvent(events.OnLongRest, s.actor)
	duration.OnEventOccurred(longRest)
	s.True(duration.IsExpired())
}

func (s *DurationTestSuite) TestUntilEventDurationRemainsExpired() {
	// Setup
	duration := events.NewUntilEventDuration(events.OnShortRest)

	// Trigger expiration
	shortRest := events.NewGameEvent(events.OnShortRest, s.actor)
	duration.OnEventOccurred(shortRest)
	s.True(duration.IsExpired())

	// Should remain expired even after more events
	turnEnd := events.NewGameEvent(events.OnTurnEnd, s.actor)
	duration.OnEventOccurred(turnEnd)
	s.True(duration.IsExpired())
}

// ConcentrationDuration Tests

func (s *DurationTestSuite) TestConcentrationDurationWithBaseDuration() {
	// Setup - concentration with 5 turn base
	baseDuration := events.NewTurnDuration(5, false)
	concentration := events.NewConcentrationDuration(baseDuration)

	// Initial state
	s.False(concentration.IsExpired())

	// Pass through turn events
	turnEnd := events.NewGameEvent(events.OnTurnEnd, s.actor)
	for i := 0; i < 4; i++ {
		concentration.OnEventOccurred(turnEnd)
		s.False(concentration.IsExpired())
	}

	// 5th turn expires base duration
	concentration.OnEventOccurred(turnEnd)
	s.True(concentration.IsExpired())
}

func (s *DurationTestSuite) TestConcentrationDurationTakeDamageEvent() {
	// Setup
	baseDuration := events.NewTurnDuration(10, false)
	concentration := events.NewConcentrationDuration(baseDuration)
	caster := &character.Character{ID: "caster"}

	// Take damage event (as target)
	damageEvent := events.NewGameEvent(events.OnTakeDamage, nil).
		WithTarget(caster)

	// This should be noticed by concentration (though not broken without saves)
	concentration.OnEventOccurred(damageEvent)

	// Without concentration save implementation, it shouldn't break
	s.False(concentration.IsExpired())
}

func (s *DurationTestSuite) TestConcentrationDurationWithExpiredBase() {
	// Setup - use a mock base duration that's already expired
	mockDuration := mockevents.NewMockModifierDuration(s.ctrl)
	mockDuration.EXPECT().IsExpired().Return(true).AnyTimes()
	mockDuration.EXPECT().OnEventOccurred(gomock.Any()).AnyTimes()

	concentration := events.NewConcentrationDuration(mockDuration)

	// Should be expired because base is expired
	s.True(concentration.IsExpired())
}

func (s *DurationTestSuite) TestConcentrationDurationPassesEventsToBase() {
	// Setup - use mock to verify events are passed through
	mockDuration := mockevents.NewMockModifierDuration(s.ctrl)
	concentration := events.NewConcentrationDuration(mockDuration)

	// Create various events
	gameEvents := []*events.GameEvent{
		events.NewGameEvent(events.OnTurnEnd, s.actor),
		events.NewGameEvent(events.OnAttackRoll, s.actor),
		events.NewGameEvent(events.OnDamageRoll, s.actor),
	}

	// Expect all events to be passed to base duration
	for _, event := range gameEvents {
		mockDuration.EXPECT().OnEventOccurred(event)
		mockDuration.EXPECT().IsExpired().Return(false)

		concentration.OnEventOccurred(event)
		s.False(concentration.IsExpired())
	}
}
