package ability

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/events"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e/features"
)

// rageDurationTracker implements DurationTracker for rage ability
type rageDurationTracker struct {
	rageListener *features.RageListener
}

func newRageDurationTracker(listener *features.RageListener) DurationTracker {
	return &rageDurationTracker{
		rageListener: listener,
	}
}

func (r *rageDurationTracker) GetDuration() events.ModifierDuration {
	return r.rageListener.Duration()
}

func (r *rageDurationTracker) IsExpired(currentTurn, numCombatants int) bool {
	duration := r.GetDuration()
	if roundsDuration, ok := duration.(*events.RoundsDuration); ok {
		turnsElapsed := currentTurn - roundsDuration.StartTurn
		roundsElapsed := 0
		if numCombatants > 0 {
			roundsElapsed = turnsElapsed / numCombatants
		}
		return roundsElapsed >= roundsDuration.Rounds
	}
	return false
}

func (r *rageDurationTracker) GetRemainingRounds(currentTurn, numCombatants int) int {
	duration := r.GetDuration()
	if roundsDuration, ok := duration.(*events.RoundsDuration); ok {
		turnsElapsed := currentTurn - roundsDuration.StartTurn
		roundsElapsed := 0
		if numCombatants > 0 {
			roundsElapsed = turnsElapsed / numCombatants
		}
		remainingRounds := roundsDuration.Rounds - roundsElapsed
		if remainingRounds < 0 {
			return 0
		}
		return remainingRounds
	}
	return 0
}

func (r *rageDurationTracker) GetEffectName() string {
	return "Rage"
}

func (r *rageDurationTracker) GetAbilityKey() string {
	return "rage"
}

func (r *rageDurationTracker) OnExpire(eventBus *events.EventBus) {
	// Unsubscribe rage listener from events
	eventBus.Unsubscribe(events.OnDamageRoll, r.rageListener)
	eventBus.Unsubscribe(events.BeforeTakeDamage, r.rageListener)
}
