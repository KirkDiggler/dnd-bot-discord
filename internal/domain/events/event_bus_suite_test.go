package events_test

import (
	"errors"
	"sync"
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/events"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/events/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type EventBusTestSuite struct {
	suite.Suite
	ctrl         *gomock.Controller
	eventBus     *events.EventBus
	mockListener *mockevents.MockEventListener
	actor        *character.Character
}

func (s *EventBusTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.eventBus = events.NewEventBus()
	s.mockListener = mockevents.NewMockEventListener(s.ctrl)
	s.actor = &character.Character{ID: "test-actor"}
}

func (s *EventBusTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func TestEventBusSuite(t *testing.T) {
	suite.Run(t, new(EventBusTestSuite))
}

// Subscribe and Emit Tests

func (s *EventBusTestSuite) TestSubscribeAndEmit() {
	// Setup
	event := events.NewGameEvent(events.BeforeAttackRoll, s.actor)

	s.mockListener.EXPECT().Priority().Return(100).AnyTimes()
	s.mockListener.EXPECT().HandleEvent(event).Return(nil)

	// Execute
	s.eventBus.Subscribe(events.BeforeAttackRoll, s.mockListener)
	err := s.eventBus.Emit(event)

	// Assert
	s.NoError(err)
}

func (s *EventBusTestSuite) TestEmitWithNoListeners() {
	// Setup
	event := events.NewGameEvent(events.OnLongRest, s.actor)

	// Execute - no listeners subscribed
	err := s.eventBus.Emit(event)

	// Assert
	s.NoError(err)
}

func (s *EventBusTestSuite) TestEmitNilEvent() {
	// Execute
	err := s.eventBus.Emit(nil)

	// Assert
	s.Error(err)
	s.Contains(err.Error(), "cannot emit nil event")
}

func (s *EventBusTestSuite) TestListenerError() {
	// Setup
	event := events.NewGameEvent(events.OnSpellCast, s.actor)
	expectedErr := errors.New("listener error")

	s.mockListener.EXPECT().Priority().Return(100).AnyTimes()
	s.mockListener.EXPECT().HandleEvent(event).Return(expectedErr)

	// Execute
	s.eventBus.Subscribe(events.OnSpellCast, s.mockListener)
	err := s.eventBus.Emit(event)

	// Assert
	s.Error(err)
	s.Contains(err.Error(), "listener error")
}

// Priority Tests

func (s *EventBusTestSuite) TestPriorityOrdering() {
	// Setup
	event := events.NewGameEvent(events.OnHit, s.actor)
	callOrder := make([]int, 0)
	mu := sync.Mutex{}

	// Create three mock listeners with different priorities
	listener1 := mockevents.NewMockEventListener(s.ctrl)
	listener2 := mockevents.NewMockEventListener(s.ctrl)
	listener3 := mockevents.NewMockEventListener(s.ctrl)

	// Set expectations with priority ordering
	listener1.EXPECT().Priority().Return(300).AnyTimes()
	listener1.EXPECT().HandleEvent(event).DoAndReturn(func(*events.GameEvent) error {
		mu.Lock()
		callOrder = append(callOrder, 300)
		mu.Unlock()
		return nil
	})

	listener2.EXPECT().Priority().Return(100).AnyTimes()
	listener2.EXPECT().HandleEvent(event).DoAndReturn(func(*events.GameEvent) error {
		mu.Lock()
		callOrder = append(callOrder, 100)
		mu.Unlock()
		return nil
	})

	listener3.EXPECT().Priority().Return(200).AnyTimes()
	listener3.EXPECT().HandleEvent(event).DoAndReturn(func(*events.GameEvent) error {
		mu.Lock()
		callOrder = append(callOrder, 200)
		mu.Unlock()
		return nil
	})

	// Subscribe in random order
	s.eventBus.Subscribe(events.OnHit, listener1)
	s.eventBus.Subscribe(events.OnHit, listener2)
	s.eventBus.Subscribe(events.OnHit, listener3)

	// Execute
	err := s.eventBus.Emit(event)

	// Assert
	s.NoError(err)
	s.Equal([]int{100, 200, 300}, callOrder)
}

// Cancellation Tests

func (s *EventBusTestSuite) TestEventCancellation() {
	// Setup
	event := events.NewGameEvent(events.BeforeTakeDamage, s.actor)

	// First listener cancels the event
	listener1 := mockevents.NewMockEventListener(s.ctrl)
	listener1.EXPECT().Priority().Return(100).AnyTimes()
	listener1.EXPECT().HandleEvent(event).DoAndReturn(func(e *events.GameEvent) error {
		e.Cancel()
		return nil
	})

	// Second listener should not be called
	listener2 := mockevents.NewMockEventListener(s.ctrl)
	listener2.EXPECT().Priority().Return(200).AnyTimes()
	// No expectation for HandleEvent - it should not be called

	// Subscribe both
	s.eventBus.Subscribe(events.BeforeTakeDamage, listener1)
	s.eventBus.Subscribe(events.BeforeTakeDamage, listener2)

	// Execute
	err := s.eventBus.Emit(event)

	// Assert
	s.NoError(err)
	s.True(event.IsCancelled())
}

// Unsubscribe Tests

func (s *EventBusTestSuite) TestUnsubscribe() {
	// Setup
	event := events.NewGameEvent(events.OnTurnStart, s.actor)

	listener1 := mockevents.NewMockEventListener(s.ctrl)
	listener2 := mockevents.NewMockEventListener(s.ctrl)

	// Both listeners should be called the first time
	listener1.EXPECT().Priority().Return(100).AnyTimes()
	listener1.EXPECT().HandleEvent(event).Return(nil).Times(1)

	listener2.EXPECT().Priority().Return(200).AnyTimes()
	listener2.EXPECT().HandleEvent(event).Return(nil).Times(2) // Called twice

	// Subscribe both
	s.eventBus.Subscribe(events.OnTurnStart, listener1)
	s.eventBus.Subscribe(events.OnTurnStart, listener2)

	// First emit - both called
	err := s.eventBus.Emit(event)
	s.NoError(err)

	// Unsubscribe listener1
	s.eventBus.Unsubscribe(events.OnTurnStart, listener1)

	// Second emit - only listener2 called
	err = s.eventBus.Emit(event)
	s.NoError(err)
}

// Utility Methods Tests

func (s *EventBusTestSuite) TestListenerCount() {
	// Initially no listeners
	s.Equal(0, s.eventBus.ListenerCount(events.OnHit))

	// Add listeners
	listener1 := mockevents.NewMockEventListener(s.ctrl)
	listener2 := mockevents.NewMockEventListener(s.ctrl)

	s.eventBus.Subscribe(events.OnHit, listener1)
	s.Equal(1, s.eventBus.ListenerCount(events.OnHit))

	s.eventBus.Subscribe(events.OnHit, listener2)
	s.Equal(2, s.eventBus.ListenerCount(events.OnHit))

	// Other event types should have 0
	s.Equal(0, s.eventBus.ListenerCount(events.OnDamageRoll))
}

func (s *EventBusTestSuite) TestClear() {
	// Setup - add multiple listeners
	listener := mockevents.NewMockEventListener(s.ctrl)

	s.eventBus.Subscribe(events.BeforeAttackRoll, listener)
	s.eventBus.Subscribe(events.OnDamageRoll, listener)
	s.eventBus.Subscribe(events.AfterTakeDamage, listener)

	s.Equal(3, s.eventBus.TotalListenerCount())

	// Execute
	s.eventBus.Clear()

	// Assert
	s.Equal(0, s.eventBus.TotalListenerCount())
}

// Concurrency Tests

func (s *EventBusTestSuite) TestConcurrentOperations() {
	// For concurrency testing, we'll use simple listeners instead of mocks
	// to avoid complex expectation setup in concurrent scenarios
	goroutines := 10

	var wg sync.WaitGroup
	wg.Add(goroutines * 3) // 3 operations per goroutine

	// Track created listeners for cleanup
	listeners := make([]events.EventListener, 0, goroutines)
	mu := sync.Mutex{}

	// Concurrent subscribes
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			listener := &testListener{priority: id, handleFunc: func(e *events.GameEvent) error { return nil }}
			mu.Lock()
			listeners = append(listeners, listener)
			mu.Unlock()
			s.eventBus.Subscribe(events.OnAttackRoll, listener)
		}(i)
	}

	// Concurrent emits
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			event := events.NewGameEvent(events.OnAttackRoll, s.actor)
			_ = s.eventBus.Emit(event) //nolint:errcheck
		}(i)
	}

	// Concurrent unsubscribes
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			// Create a new listener to unsubscribe (simulating removal of non-existent listener)
			listener := &testListener{priority: id}
			s.eventBus.Unsubscribe(events.OnAttackRoll, listener)
		}(i)
	}

	// Wait for all goroutines
	wg.Wait()

	// If we get here without panicking, thread safety is working
	s.True(true)
}

// testListener is a simple implementation for concurrency tests
type testListener struct {
	priority   int
	handleFunc func(*events.GameEvent) error
}

func (t *testListener) HandleEvent(event *events.GameEvent) error {
	if t.handleFunc != nil {
		return t.handleFunc(event)
	}
	return nil
}

func (t *testListener) Priority() int {
	return t.priority
}
