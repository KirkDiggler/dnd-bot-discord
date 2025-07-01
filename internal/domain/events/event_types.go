package events

// EventType represents different game events that can occur
type EventType int

const (
	// Attack sequence events
	BeforeAttackRoll EventType = iota
	OnAttackRoll
	AfterAttackRoll
	BeforeHit
	OnHit
	AfterHit

	// Damage events
	BeforeDamageRoll
	OnDamageRoll
	AfterDamageRoll
	BeforeTakeDamage
	OnTakeDamage
	AfterTakeDamage

	// Saving throws
	BeforeSavingThrow
	OnSavingThrow
	AfterSavingThrow

	// Ability checks
	BeforeAbilityCheck
	OnAbilityCheck
	AfterAbilityCheck

	// Turn management
	OnTurnStart
	OnTurnEnd

	// Status effects
	OnStatusApplied
	OnStatusRemoved

	// Rest events
	OnShortRest
	OnLongRest
)

// String returns the string representation of an EventType
func (e EventType) String() string {
	names := []string{
		"BeforeAttackRoll",
		"OnAttackRoll",
		"AfterAttackRoll",
		"BeforeHit",
		"OnHit",
		"AfterHit",
		"BeforeDamageRoll",
		"OnDamageRoll",
		"AfterDamageRoll",
		"BeforeTakeDamage",
		"OnTakeDamage",
		"AfterTakeDamage",
		"BeforeSavingThrow",
		"OnSavingThrow",
		"AfterSavingThrow",
		"BeforeAbilityCheck",
		"OnAbilityCheck",
		"AfterAbilityCheck",
		"OnTurnStart",
		"OnTurnEnd",
		"OnStatusApplied",
		"OnStatusRemoved",
		"OnShortRest",
		"OnLongRest",
	}

	if int(e) < len(names) {
		return names[e]
	}
	return "Unknown"
}
