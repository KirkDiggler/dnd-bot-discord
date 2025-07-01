package events_test

import (
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/events"
	"github.com/stretchr/testify/suite"
)

type GameEventSuite struct {
	suite.Suite
}

func TestGameEventSuite(t *testing.T) {
	suite.Run(t, new(GameEventSuite))
}

func (s *GameEventSuite) TestNewGameEvent() {
	event := events.NewGameEvent(events.BeforeAttackRoll)

	s.Equal(events.BeforeAttackRoll, event.Type)
	s.NotNil(event.Context)
	s.False(event.Cancelled)
	s.Nil(event.Actor)
	s.Nil(event.Target)
}

func (s *GameEventSuite) TestBuilderPattern() {
	actor := &character.Character{ID: "actor123"}
	target := &character.Character{ID: "target456"}

	event := events.NewGameEvent(events.BeforeAttackRoll).
		WithActor(actor).
		WithTarget(target).
		WithContext("weapon", "longsword").
		WithContext("advantage", true)

	s.Equal(actor, event.Actor)
	s.Equal(target, event.Target)

	weapon, exists := event.GetStringContext("weapon")
	s.True(exists)
	s.Equal("longsword", weapon)

	advantage, exists := event.GetBoolContext("advantage")
	s.True(exists)
	s.True(advantage)
}

func (s *GameEventSuite) TestCancellation() {
	event := events.NewGameEvent(events.BeforeAttackRoll)

	s.False(event.IsCancelled())

	event.Cancel()

	s.True(event.IsCancelled())
}

func (s *GameEventSuite) TestGetIntContext() {
	event := events.NewGameEvent(events.OnDamageRoll).
		WithContext("damage", 10).
		WithContext("notAnInt", "string")

	// Valid int
	damage, exists := event.GetIntContext("damage")
	s.True(exists)
	s.Equal(10, damage)

	// Missing key
	missing, exists := event.GetIntContext("missing")
	s.False(exists)
	s.Equal(0, missing)

	// Wrong type
	notInt, exists := event.GetIntContext("notAnInt")
	s.False(exists)
	s.Equal(0, notInt)
}

func (s *GameEventSuite) TestGetStringContext() {
	event := events.NewGameEvent(events.BeforeAttackRoll).
		WithContext("weapon", "longsword").
		WithContext("notAString", 123)

	// Valid string
	weapon, exists := event.GetStringContext("weapon")
	s.True(exists)
	s.Equal("longsword", weapon)

	// Missing key
	missing, exists := event.GetStringContext("missing")
	s.False(exists)
	s.Equal("", missing)

	// Wrong type
	notString, exists := event.GetStringContext("notAString")
	s.False(exists)
	s.Equal("", notString)
}

func (s *GameEventSuite) TestGetBoolContext() {
	event := events.NewGameEvent(events.BeforeAttackRoll).
		WithContext("advantage", true).
		WithContext("disadvantage", false).
		WithContext("notABool", "yes")

	// Valid true bool
	advantage, exists := event.GetBoolContext("advantage")
	s.True(exists)
	s.True(advantage)

	// Valid false bool
	disadvantage, exists := event.GetBoolContext("disadvantage")
	s.True(exists)
	s.False(disadvantage)

	// Missing key
	missing, exists := event.GetBoolContext("missing")
	s.False(exists)
	s.False(missing)

	// Wrong type
	notBool, exists := event.GetBoolContext("notABool")
	s.False(exists)
	s.False(notBool)
}

func (s *GameEventSuite) TestContextOverwrite() {
	event := events.NewGameEvent(events.BeforeAttackRoll).
		WithContext("key", "value1").
		WithContext("key", "value2")

	value, exists := event.GetStringContext("key")
	s.True(exists)
	s.Equal("value2", value)
}
