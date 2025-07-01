package events

// EventType represents the type of game event
type EventType int

const (
	// Combat Events
	BeforeAttackRoll EventType = iota
	OnAttackRoll
	AfterAttackRoll
	BeforeHit
	OnHit
	BeforeDamageRoll
	OnDamageRoll
	AfterDamageRoll
	BeforeTakeDamage
	OnTakeDamage
	AfterTakeDamage

	// Ability Check Events
	BeforeAbilityCheck
	OnAbilityCheck
	AfterAbilityCheck

	// Saving Throw Events
	BeforeSavingThrow
	OnSavingThrow
	AfterSavingThrow

	// Spell Events
	BeforeSpellCast
	OnSpellCast
	AfterSpellCast

	// Movement Events
	BeforeMove
	OnMove
	AfterMove

	// Resource Events
	OnShortRest
	OnLongRest
	OnTurnStart
	OnTurnEnd
)

// String returns the string representation of the event type
func (e EventType) String() string {
	names := [...]string{
		"BeforeAttackRoll",
		"OnAttackRoll",
		"AfterAttackRoll",
		"BeforeHit",
		"OnHit",
		"BeforeDamageRoll",
		"OnDamageRoll",
		"AfterDamageRoll",
		"BeforeTakeDamage",
		"OnTakeDamage",
		"AfterTakeDamage",
		"BeforeAbilityCheck",
		"OnAbilityCheck",
		"AfterAbilityCheck",
		"BeforeSavingThrow",
		"OnSavingThrow",
		"AfterSavingThrow",
		"BeforeSpellCast",
		"OnSpellCast",
		"AfterSpellCast",
		"BeforeMove",
		"OnMove",
		"AfterMove",
		"OnShortRest",
		"OnLongRest",
		"OnTurnStart",
		"OnTurnEnd",
	}
	if e < BeforeAttackRoll || int(e) >= len(names) {
		return "Unknown"
	}
	return names[e]
}
