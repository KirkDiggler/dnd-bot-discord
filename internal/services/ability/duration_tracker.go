package ability

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/events"
)

// DurationTracker is an interface for tracking ability durations
type DurationTracker interface {
	// GetDuration returns the current duration
	GetDuration() events.ModifierDuration
	// IsExpired checks if the duration has expired based on current turn/round
	IsExpired(currentTurn, numCombatants int) bool
	// GetRemainingRounds calculates remaining rounds
	GetRemainingRounds(currentTurn, numCombatants int) int
	// GetEffectName returns the name of the effect for UI updates
	GetEffectName() string
	// GetAbilityKey returns the ability key
	GetAbilityKey() string
	// OnExpire handles cleanup when the duration expires
	OnExpire(eventBus *events.EventBus)
}

// TrackedAbility wraps an event listener with duration tracking
type TrackedAbility struct {
	CharacterID string
	Listener    events.EventListener
	Tracker     DurationTracker
}
