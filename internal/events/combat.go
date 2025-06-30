package events

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities/damage"
)

// BeforeAttackRollEvent is emitted before an attack roll is made
type BeforeAttackRollEvent struct {
	BaseEvent
	Weapon          entities.Equipment
	AttackType      string // melee, ranged, spell
	AttackBonus     int
	Advantage       bool
	Disadvantage    bool
	AbilityModifier entities.Attribute // STR, DEX, etc.
}

func (e *BeforeAttackRollEvent) Accept(v ModifierVisitor) {
	v.VisitBeforeAttackRollEvent(e)
}

// OnAttackRollEvent is emitted when the attack roll is made
type OnAttackRollEvent struct {
	BaseEvent
	Weapon      entities.Equipment
	AttackType  string
	BaseRoll    int  // The d20 roll
	AttackBonus int  // Total bonus being applied
	TotalAttack int  // BaseRoll + AttackBonus
	IsCritical  bool // Natural 20
	IsFumble    bool // Natural 1
}

func (e *OnAttackRollEvent) Accept(v ModifierVisitor) {
	v.VisitOnAttackRollEvent(e)
}

// AfterAttackRollEvent is emitted after the attack roll is complete
type AfterAttackRollEvent struct {
	BaseEvent
	Weapon      entities.Equipment
	AttackType  string
	TotalAttack int
	Hit         bool
}

func (e *AfterAttackRollEvent) Accept(v ModifierVisitor) {
	v.VisitAfterAttackRollEvent(e)
}

// BeforeHitEvent is emitted before checking if an attack hits
type BeforeHitEvent struct {
	BaseEvent
	AttackRoll int
	TargetAC   int
	ACBonuses  int // Additional AC from reactions, etc.
}

func (e *BeforeHitEvent) Accept(v ModifierVisitor) {
	v.VisitBeforeHitEvent(e)
}

// OnHitEvent is emitted when an attack hits
type OnHitEvent struct {
	BaseEvent
	Weapon     entities.Equipment
	AttackType string
	IsCritical bool
}

func (e *OnHitEvent) Accept(v ModifierVisitor) {
	v.VisitOnHitEvent(e)
}

// BeforeDamageRollEvent is emitted before rolling damage
type BeforeDamageRollEvent struct {
	BaseEvent
	Weapon      entities.Equipment
	DamageType  damage.Type
	DamageBonus int
	IsCritical  bool
	DamageDice  string // e.g., "1d8"
}

func (e *BeforeDamageRollEvent) Accept(v ModifierVisitor) {
	v.VisitBeforeDamageRollEvent(e)
}

// OnDamageRollEvent is emitted when damage is rolled
type OnDamageRollEvent struct {
	BaseEvent
	Weapon      entities.Equipment
	DamageType  damage.Type
	BaseDamage  int // Dice roll result
	DamageBonus int // Bonuses being applied
	TotalDamage int // BaseDamage + DamageBonus
	IsCritical  bool
}

func (e *OnDamageRollEvent) Accept(v ModifierVisitor) {
	v.VisitOnDamageRollEvent(e)
}

// AfterDamageRollEvent is emitted after damage is calculated
type AfterDamageRollEvent struct {
	BaseEvent
	Weapon      entities.Equipment
	DamageType  damage.Type
	TotalDamage int
}

func (e *AfterDamageRollEvent) Accept(v ModifierVisitor) {
	v.VisitAfterDamageRollEvent(e)
}

// BeforeTakeDamageEvent is emitted before a character takes damage
type BeforeTakeDamageEvent struct {
	BaseEvent
	DamageAmount    int
	DamageType      damage.Type
	Source          string // weapon, spell, effect, etc.
	Resistances     []damage.Type
	Immunities      []damage.Type
	Vulnerabilities []damage.Type
}

func (e *BeforeTakeDamageEvent) Accept(v ModifierVisitor) {
	v.VisitBeforeTakeDamageEvent(e)
}

// OnTakeDamageEvent is emitted when damage is applied
type OnTakeDamageEvent struct {
	BaseEvent
	OriginalDamage int
	FinalDamage    int
	DamageType     damage.Type
	WasResisted    bool
	WasImmune      bool
	WasVulnerable  bool
}

func (e *OnTakeDamageEvent) Accept(v ModifierVisitor) {
	v.VisitOnTakeDamageEvent(e)
}

// AfterTakeDamageEvent is emitted after damage is taken
type AfterTakeDamageEvent struct {
	BaseEvent
	DamageTaken int
	DamageType  damage.Type
	NewHP       int
	Unconscious bool
	Dead        bool
}

func (e *AfterTakeDamageEvent) Accept(v ModifierVisitor) {
	v.VisitAfterTakeDamageEvent(e)
}
