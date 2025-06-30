package events

// Event type constants
const (
	// Combat Events
	EventTypeBeforeAttackRoll EventType = "before_attack_roll"
	EventTypeOnAttackRoll     EventType = "on_attack_roll"
	EventTypeAfterAttackRoll  EventType = "after_attack_roll"
	EventTypeBeforeHit        EventType = "before_hit"
	EventTypeOnHit            EventType = "on_hit"
	EventTypeBeforeDamageRoll EventType = "before_damage_roll"
	EventTypeOnDamageRoll     EventType = "on_damage_roll"
	EventTypeAfterDamageRoll  EventType = "after_damage_roll"
	EventTypeBeforeTakeDamage EventType = "before_take_damage"
	EventTypeOnTakeDamage     EventType = "on_take_damage"
	EventTypeAfterTakeDamage  EventType = "after_take_damage"

	// Ability Check Events
	EventTypeBeforeAbilityCheck EventType = "before_ability_check"
	EventTypeOnAbilityCheck     EventType = "on_ability_check"
	EventTypeAfterAbilityCheck  EventType = "after_ability_check"

	// Saving Throw Events
	EventTypeBeforeSavingThrow EventType = "before_saving_throw"
	EventTypeOnSavingThrow     EventType = "on_saving_throw"
	EventTypeAfterSavingThrow  EventType = "after_saving_throw"

	// Spell Events
	EventTypeBeforeSpellCast EventType = "before_spell_cast"
	EventTypeOnSpellCast     EventType = "on_spell_cast"
	EventTypeAfterSpellCast  EventType = "after_spell_cast"

	// Movement Events
	EventTypeBeforeMove EventType = "before_move"
	EventTypeOnMove     EventType = "on_move"
	EventTypeAfterMove  EventType = "after_move"

	// Resource Events
	EventTypeOnShortRest EventType = "on_short_rest"
	EventTypeOnLongRest  EventType = "on_long_rest"
	EventTypeOnTurnStart EventType = "on_turn_start"
	EventTypeOnTurnEnd   EventType = "on_turn_end"
)

// Priority levels for modifier application order
const (
	PriorityPreCalculation  = 0   // Set base values
	PriorityFeatures        = 100 // Class features, racial traits
	PriorityStatusEffects   = 200 // Conditions, spells
	PriorityEquipment       = 300 // Equipment modifiers
	PriorityTemporary       = 400 // Inspiration, guidance
	PriorityPostCalculation = 500 // Caps, limits
)
