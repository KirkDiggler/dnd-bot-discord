package events_test

import (
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/events"
)

// Simple listener implementation for benchmarks
type benchListener struct {
	priority int
}

func (b *benchListener) HandleEvent(event *events.GameEvent) error {
	// Simulate some work
	event.WithContext("processed", true)
	return nil
}

func (b *benchListener) Priority() int {
	return b.priority
}

func BenchmarkEventBus_Emit_SingleListener(b *testing.B) {
	bus := events.NewEventBus()
	listener := &benchListener{priority: 100}
	bus.Subscribe(events.OnAttackRoll, listener)

	actor := &character.Character{ID: "bench-actor"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		event := events.NewGameEvent(events.OnAttackRoll, actor)
		_ = bus.Emit(event) //nolint:errcheck
	}
}

func BenchmarkEventBus_Emit_MultipleListeners(b *testing.B) {
	bus := events.NewEventBus()

	// Add 10 listeners with different priorities
	for i := 0; i < 10; i++ {
		listener := &benchListener{priority: i * 10}
		bus.Subscribe(events.OnDamageRoll, listener)
	}

	actor := &character.Character{ID: "bench-actor"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		event := events.NewGameEvent(events.OnDamageRoll, actor)
		_ = bus.Emit(event) //nolint:errcheck
	}
}

func BenchmarkEventBus_Emit_NoListeners(b *testing.B) {
	bus := events.NewEventBus()
	actor := &character.Character{ID: "bench-actor"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		event := events.NewGameEvent(events.OnHit, actor)
		_ = bus.Emit(event) //nolint:errcheck
	}
}

func BenchmarkEventBus_Subscribe(b *testing.B) {
	bus := events.NewEventBus()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		listener := &benchListener{priority: i % 500}
		bus.Subscribe(events.EventType(i%27), listener) // Cycle through event types
	}
}

func BenchmarkEventBus_Concurrent_Emit(b *testing.B) {
	bus := events.NewEventBus()

	// Add some listeners
	for i := 0; i < 5; i++ {
		listener := &benchListener{priority: i * 10}
		bus.Subscribe(events.BeforeAttackRoll, listener)
	}

	actor := &character.Character{ID: "bench-actor"}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			event := events.NewGameEvent(events.BeforeAttackRoll, actor)
			_ = bus.Emit(event) //nolint:errcheck
		}
	})
}

func BenchmarkGameEvent_Creation(b *testing.B) {
	actor := &character.Character{ID: "bench-actor"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		event := events.NewGameEvent(events.OnAttackRoll, actor)
		_ = event // Prevent compiler optimization
	}
}

func BenchmarkGameEvent_WithContext(b *testing.B) {
	actor := &character.Character{ID: "bench-actor"}
	target := &character.Character{ID: "bench-target"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		event := events.NewGameEvent(events.OnDamageRoll, actor).
			WithTarget(target).
			WithContext("weapon", "longsword").
			WithContext("damageType", "slashing").
			WithContext("damageAmount", 8).
			WithContext("criticalHit", false).
			WithContext("attackBonus", 5)
		_ = event // Prevent compiler optimization
	}
}

func BenchmarkGameEvent_GetContext(b *testing.B) {
	actor := &character.Character{ID: "bench-actor"}
	event := events.NewGameEvent(events.OnHit, actor).
		WithContext("weapon", "shortsword").
		WithContext("damageType", "piercing").
		WithContext("damageAmount", 6).
		WithContext("attackBonus", 3).
		WithContext("proficiencyBonus", 2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		val, exists := event.GetContext("damageAmount")
		_ = val
		_ = exists
	}
}

func BenchmarkGameEvent_GetTypedContext(b *testing.B) {
	actor := &character.Character{ID: "bench-actor"}
	event := events.NewGameEvent(events.OnHit, actor).
		WithContext("damage", 10).
		WithContext("advantage", true).
		WithContext("weaponName", "longsword")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Mix of typed context retrievals
		intVal, _ := event.GetIntContext("damage")
		boolVal, _ := event.GetBoolContext("advantage")
		strVal, _ := event.GetStringContext("weaponName")

		_ = intVal
		_ = boolVal
		_ = strVal
	}
}
