package events_test

import (
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/events"
)

// benchmarkListener is a simple listener for benchmarks
type benchmarkListener struct {
	priority int
}

func (b *benchmarkListener) HandleEvent(event *events.GameEvent) error {
	return nil
}

func (b *benchmarkListener) Priority() int {
	return b.priority
}

func BenchmarkEventBusEmit(b *testing.B) {
	bus := events.NewEventBus()

	// Add some listeners
	for i := 0; i < 10; i++ {
		listener := &benchmarkListener{priority: i}
		bus.Subscribe(events.BeforeAttackRoll, listener)
	}

	event := events.NewGameEvent(events.BeforeAttackRoll)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bus.Emit(event) //nolint:errcheck // benchmark
	}
}

func BenchmarkEventBusEmitSingleListener(b *testing.B) {
	bus := events.NewEventBus()
	listener := &benchmarkListener{priority: 10}
	bus.Subscribe(events.BeforeAttackRoll, listener)

	event := events.NewGameEvent(events.BeforeAttackRoll)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bus.Emit(event) //nolint:errcheck // benchmark
	}
}

func BenchmarkEventBusEmitNoListeners(b *testing.B) {
	bus := events.NewEventBus()
	event := events.NewGameEvent(events.BeforeAttackRoll)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bus.Emit(event) //nolint:errcheck // benchmark
	}
}

func BenchmarkEventBusSubscribe(b *testing.B) {
	bus := events.NewEventBus()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		listener := &benchmarkListener{priority: i}
		bus.Subscribe(events.BeforeAttackRoll, listener)
	}
}

func BenchmarkGameEventBuilder(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = events.NewGameEvent(events.BeforeAttackRoll).
			WithContext("weapon", "longsword").
			WithContext("advantage", true).
			WithContext("damage", 10)
	}
}
