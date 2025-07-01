package events_test

import (
	"errors"
	"sync"
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/events"
	"github.com/stretchr/testify/suite"
)

type EventBusSuite struct {
	suite.Suite
	bus *events.EventBus
}

func TestEventBusSuite(t *testing.T) {
	suite.Run(t, new(EventBusSuite))
}

func (s *EventBusSuite) SetupTest() {
	s.bus = events.NewEventBus()
}

// mockListener implements EventListener for testing
type mockListener struct {
	priority int
	handler  func(event *events.GameEvent) error
	called   bool
	mu       sync.Mutex
}

func (m *mockListener) HandleEvent(event *events.GameEvent) error {
	m.mu.Lock()
	m.called = true
	m.mu.Unlock()

	if m.handler != nil {
		return m.handler(event)
	}
	return nil
}

func (m *mockListener) Priority() int {
	return m.priority
}

func (m *mockListener) wasCalled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.called
}

func (s *EventBusSuite) TestSubscribeAndEmit() {
	listener := &mockListener{priority: 10}
	s.bus.Subscribe(events.BeforeAttackRoll, listener)

	event := events.NewGameEvent(events.BeforeAttackRoll)
	err := s.bus.Emit(event)

	s.NoError(err)
	s.True(listener.wasCalled())
}

func (s *EventBusSuite) TestUnsubscribe() {
	listener := &mockListener{priority: 10}
	s.bus.Subscribe(events.BeforeAttackRoll, listener)
	s.bus.Unsubscribe(events.BeforeAttackRoll, listener)

	event := events.NewGameEvent(events.BeforeAttackRoll)
	err := s.bus.Emit(event)

	s.NoError(err)
	s.False(listener.wasCalled())
}

func (s *EventBusSuite) TestPriorityOrdering() {
	var executionOrder []int
	mu := &sync.Mutex{}

	// Create listeners with different priorities
	listener1 := &mockListener{
		priority: 20,
		handler: func(event *events.GameEvent) error {
			mu.Lock()
			executionOrder = append(executionOrder, 20)
			mu.Unlock()
			return nil
		},
	}

	listener2 := &mockListener{
		priority: 10,
		handler: func(event *events.GameEvent) error {
			mu.Lock()
			executionOrder = append(executionOrder, 10)
			mu.Unlock()
			return nil
		},
	}

	listener3 := &mockListener{
		priority: 30,
		handler: func(event *events.GameEvent) error {
			mu.Lock()
			executionOrder = append(executionOrder, 30)
			mu.Unlock()
			return nil
		},
	}

	// Subscribe in random order
	s.bus.Subscribe(events.BeforeAttackRoll, listener3)
	s.bus.Subscribe(events.BeforeAttackRoll, listener1)
	s.bus.Subscribe(events.BeforeAttackRoll, listener2)

	event := events.NewGameEvent(events.BeforeAttackRoll)
	err := s.bus.Emit(event)

	s.NoError(err)
	s.Equal([]int{10, 20, 30}, executionOrder)
}

func (s *EventBusSuite) TestEventCancellation() {
	listener1 := &mockListener{
		priority: 10,
		handler: func(event *events.GameEvent) error {
			event.Cancel()
			return nil
		},
	}

	listener2 := &mockListener{
		priority: 20,
	}

	s.bus.Subscribe(events.BeforeAttackRoll, listener1)
	s.bus.Subscribe(events.BeforeAttackRoll, listener2)

	event := events.NewGameEvent(events.BeforeAttackRoll)
	err := s.bus.Emit(event)

	s.NoError(err)
	s.True(event.IsCancelled())
	s.True(listener1.wasCalled())
	s.False(listener2.wasCalled()) // Should not be called due to cancellation
}

func (s *EventBusSuite) TestListenerError() {
	expectedErr := errors.New("test error")
	listener := &mockListener{
		priority: 10,
		handler: func(event *events.GameEvent) error {
			return expectedErr
		},
	}

	s.bus.Subscribe(events.BeforeAttackRoll, listener)

	event := events.NewGameEvent(events.BeforeAttackRoll)
	err := s.bus.Emit(event)

	s.Error(err)
	s.Contains(err.Error(), "test error")
}

func (s *EventBusSuite) TestConcurrentAccess() {
	// Test concurrent subscribe/unsubscribe/emit
	var wg sync.WaitGroup

	// Spawn multiple goroutines
	for i := 0; i < 10; i++ {
		wg.Add(3)

		// Subscribe goroutine
		go func(id int) {
			defer wg.Done()
			listener := &mockListener{priority: id}
			s.bus.Subscribe(events.BeforeAttackRoll, listener)
		}(i)

		// Emit goroutine
		go func() {
			defer wg.Done()
			event := events.NewGameEvent(events.BeforeAttackRoll)
			_ = s.bus.Emit(event) //nolint:errcheck // test concurrent access
		}()

		// Clear goroutine
		go func() {
			defer wg.Done()
			if i%3 == 0 {
				s.bus.Clear()
			}
		}()
	}

	wg.Wait()
	// If we get here without deadlock or panic, the test passes
}

func (s *EventBusSuite) TestListenerCount() {
	s.Equal(0, s.bus.ListenerCount(events.BeforeAttackRoll))

	listener1 := &mockListener{priority: 10}
	listener2 := &mockListener{priority: 20}

	s.bus.Subscribe(events.BeforeAttackRoll, listener1)
	s.Equal(1, s.bus.ListenerCount(events.BeforeAttackRoll))

	s.bus.Subscribe(events.BeforeAttackRoll, listener2)
	s.Equal(2, s.bus.ListenerCount(events.BeforeAttackRoll))

	s.bus.Unsubscribe(events.BeforeAttackRoll, listener1)
	s.Equal(1, s.bus.ListenerCount(events.BeforeAttackRoll))
}

func (s *EventBusSuite) TestClear() {
	listener := &mockListener{priority: 10}
	s.bus.Subscribe(events.BeforeAttackRoll, listener)
	s.bus.Subscribe(events.OnAttackRoll, listener)

	s.bus.Clear()

	s.Equal(0, s.bus.ListenerCount(events.BeforeAttackRoll))
	s.Equal(0, s.bus.ListenerCount(events.OnAttackRoll))
}

func (s *EventBusSuite) TestMultipleEventTypes() {
	listener1 := &mockListener{priority: 10}
	listener2 := &mockListener{priority: 20}

	s.bus.Subscribe(events.BeforeAttackRoll, listener1)
	s.bus.Subscribe(events.OnDamageRoll, listener2)

	// Emit attack event
	attackEvent := events.NewGameEvent(events.BeforeAttackRoll)
	err := s.bus.Emit(attackEvent)
	s.NoError(err)
	s.True(listener1.wasCalled())
	s.False(listener2.wasCalled())

	// Reset
	listener1.called = false

	// Emit damage event
	damageEvent := events.NewGameEvent(events.OnDamageRoll)
	err = s.bus.Emit(damageEvent)
	s.NoError(err)
	s.False(listener1.wasCalled())
	s.True(listener2.wasCalled())
}
