package events_test

import (
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/events"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/events/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type GameEventTestSuite struct {
	suite.Suite
	ctrl         *gomock.Controller
	actor        *character.Character
	target       *character.Character
	mockModifier *mockevents.MockModifier
}

func (s *GameEventTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.actor = &character.Character{ID: "actor1", Name: "Thorin"}
	s.target = &character.Character{ID: "target1", Name: "Goblin"}
	s.mockModifier = mockevents.NewMockModifier(s.ctrl)
}

func (s *GameEventTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func TestGameEventSuite(t *testing.T) {
	suite.Run(t, new(GameEventTestSuite))
}

// Constructor Tests

func (s *GameEventTestSuite) TestNewGameEvent() {
	// Execute
	event := events.NewGameEvent(events.BeforeAttackRoll, s.actor)

	// Assert
	s.NotNil(event)
	s.Equal(events.BeforeAttackRoll, event.Type)
	s.Equal(s.actor, event.Actor)
	s.Nil(event.Target)
	s.NotNil(event.Context)
	s.NotNil(event.Modifiers)
	s.False(event.Cancelled)
}

// Builder Pattern Tests

func (s *GameEventTestSuite) TestWithTarget() {
	// Execute
	event := events.NewGameEvent(events.OnAttackRoll, s.actor).
		WithTarget(s.target)

	// Assert
	s.Equal(s.target, event.Target)
}

func (s *GameEventTestSuite) TestWithContext() {
	// Execute
	event := events.NewGameEvent(events.OnDamageRoll, s.actor).
		WithContext("weapon", "longsword").
		WithContext("damageType", "slashing").
		WithContext("damageAmount", 8).
		WithContext("criticalHit", true)

	// Assert
	s.Equal("longsword", event.Context["weapon"])
	s.Equal("slashing", event.Context["damageType"])
	s.Equal(8, event.Context["damageAmount"])
	s.Equal(true, event.Context["criticalHit"])
}

func (s *GameEventTestSuite) TestChainedBuilders() {
	// Execute - test method chaining
	event := events.NewGameEvent(events.BeforeHit, s.actor).
		WithTarget(s.target).
		WithContext("attackRoll", 15).
		WithContext("advantage", true)

	// Assert
	s.Equal(s.actor, event.Actor)
	s.Equal(s.target, event.Target)
	s.Equal(15, event.Context["attackRoll"])
	s.Equal(true, event.Context["advantage"])
}

// Cancellation Tests

func (s *GameEventTestSuite) TestCancellation() {
	// Setup
	event := events.NewGameEvent(events.BeforeTakeDamage, s.actor)

	// Assert initial state
	s.False(event.IsCancelled())

	// Execute
	event.Cancel()

	// Assert cancelled state
	s.True(event.IsCancelled())
}

// Modifier Tests

func (s *GameEventTestSuite) TestAddModifier() {
	// Setup
	event := events.NewGameEvent(events.OnAttackRoll, s.actor)

	// Initial state
	s.Len(event.Modifiers, 0)

	// Execute
	event.AddModifier(s.mockModifier)

	// Assert
	s.Len(event.Modifiers, 1)
	s.Equal(s.mockModifier, event.Modifiers[0])
}

func (s *GameEventTestSuite) TestAddMultipleModifiers() {
	// Setup
	event := events.NewGameEvent(events.OnDamageRoll, s.actor)
	modifier1 := mockevents.NewMockModifier(s.ctrl)
	modifier2 := mockevents.NewMockModifier(s.ctrl)

	// Execute
	event.AddModifier(modifier1)
	event.AddModifier(modifier2)
	event.AddModifier(s.mockModifier)

	// Assert
	s.Len(event.Modifiers, 3)
	s.Equal(modifier1, event.Modifiers[0])
	s.Equal(modifier2, event.Modifiers[1])
	s.Equal(s.mockModifier, event.Modifiers[2])
}

// Context Getter Tests

func (s *GameEventTestSuite) TestGetContext() {
	// Setup
	event := events.NewGameEvent(events.OnHit, s.actor).
		WithContext("attackBonus", 5).
		WithContext("weaponName", "shortsword")

	// Test existing value
	val, exists := event.GetContext("attackBonus")
	s.True(exists)
	s.Equal(5, val)

	// Test another existing value
	val, exists = event.GetContext("weaponName")
	s.True(exists)
	s.Equal("shortsword", val)

	// Test non-existing value
	val, exists = event.GetContext("nonexistent")
	s.False(exists)
	s.Nil(val)
}

func (s *GameEventTestSuite) TestGetIntContext() {
	// Setup
	event := events.NewGameEvent(events.OnDamageRoll, s.actor).
		WithContext("damage", 10).
		WithContext("notAnInt", "string").
		WithContext("nilValue", nil)

	// Test valid int
	val, ok := event.GetIntContext("damage")
	s.True(ok)
	s.Equal(10, val)

	// Test non-existent key
	val, ok = event.GetIntContext("missing")
	s.False(ok)
	s.Equal(0, val)

	// Test wrong type
	val, ok = event.GetIntContext("notAnInt")
	s.False(ok)
	s.Equal(0, val)

	// Test nil value
	val, ok = event.GetIntContext("nilValue")
	s.False(ok)
	s.Equal(0, val)
}

func (s *GameEventTestSuite) TestGetBoolContext() {
	// Setup
	event := events.NewGameEvent(events.BeforeAttackRoll, s.actor).
		WithContext("advantage", true).
		WithContext("disadvantage", false).
		WithContext("notABool", 123).
		WithContext("stringBool", "true")

	// Test valid bool (true)
	val, ok := event.GetBoolContext("advantage")
	s.True(ok)
	s.True(val)

	// Test valid bool (false)
	val, ok = event.GetBoolContext("disadvantage")
	s.True(ok)
	s.False(val)

	// Test non-existent key
	val, ok = event.GetBoolContext("missing")
	s.False(ok)
	s.False(val)

	// Test wrong type (int)
	val, ok = event.GetBoolContext("notABool")
	s.False(ok)
	s.False(val)

	// Test wrong type (string)
	val, ok = event.GetBoolContext("stringBool")
	s.False(ok)
	s.False(val)
}

func (s *GameEventTestSuite) TestGetStringContext() {
	// Setup
	event := events.NewGameEvent(events.OnSpellCast, s.actor).
		WithContext("spellName", "fireball").
		WithContext("spellLevel", 3).
		WithContext("emptyString", "")

	// Test valid string
	val, ok := event.GetStringContext("spellName")
	s.True(ok)
	s.Equal("fireball", val)

	// Test empty string (should still be valid)
	val, ok = event.GetStringContext("emptyString")
	s.True(ok)
	s.Equal("", val)

	// Test non-existent key
	val, ok = event.GetStringContext("missing")
	s.False(ok)
	s.Equal("", val)

	// Test wrong type
	val, ok = event.GetStringContext("spellLevel")
	s.False(ok)
	s.Equal("", val)
}

// Complex Context Tests

func (s *GameEventTestSuite) TestComplexContextTypes() {
	// Setup - test various types that might be stored in context
	type CustomStruct struct {
		Value int
		Name  string
	}

	customObj := &CustomStruct{Value: 42, Name: "test"}

	event := events.NewGameEvent(events.OnMove, s.actor).
		WithContext("intSlice", []int{1, 2, 3}).
		WithContext("stringSlice", []string{"a", "b", "c"}).
		WithContext("mapData", map[string]int{"x": 10, "y": 20}).
		WithContext("customStruct", customObj)

	// Test slice retrieval
	val, exists := event.GetContext("intSlice")
	s.True(exists)
	intSlice, ok := val.([]int)
	s.True(ok)
	s.Equal([]int{1, 2, 3}, intSlice)

	// Test map retrieval
	val, exists = event.GetContext("mapData")
	s.True(exists)
	mapData, ok := val.(map[string]int)
	s.True(ok)
	s.Equal(10, mapData["x"])
	s.Equal(20, mapData["y"])

	// Test custom struct retrieval
	val, exists = event.GetContext("customStruct")
	s.True(exists)
	retrieved, ok := val.(*CustomStruct)
	s.True(ok)
	s.Equal(42, retrieved.Value)
	s.Equal("test", retrieved.Name)
}
