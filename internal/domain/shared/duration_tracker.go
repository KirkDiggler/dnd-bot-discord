package shared

// DurationTracker tracks ability durations
type DurationTracker interface {
	// CharacterID returns the character this tracker is for
	CharacterID() string

	// AbilityKey returns the ability being tracked
	AbilityKey() string

	// EffectName returns the effect name for UI
	EffectName() string

	// IsExpired checks if the duration has expired
	IsExpired(currentTurn, numCombatants int) bool

	// GetRemainingRounds returns remaining rounds
	GetRemainingRounds(currentTurn, numCombatants int) int

	// OnExpire is called when the duration expires
	OnExpire()
}
