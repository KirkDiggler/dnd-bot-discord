package events

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
)

// EventType represents the type of game event
type EventType string

// Event is the base interface for all game events
type Event interface {
	GetType() EventType
	GetActor() *entities.Character
	GetTarget() *entities.Character
	IsCancelled() bool
	Cancel()
	Accept(visitor ModifierVisitor)
}

// BaseEvent provides common implementation for all events
type BaseEvent struct {
	Type      EventType
	Actor     *entities.Character
	Target    *entities.Character
	Cancelled bool
}

func (e *BaseEvent) GetType() EventType             { return e.Type }
func (e *BaseEvent) GetActor() *entities.Character  { return e.Actor }
func (e *BaseEvent) GetTarget() *entities.Character { return e.Target }
func (e *BaseEvent) IsCancelled() bool              { return e.Cancelled }
func (e *BaseEvent) Cancel()                        { e.Cancelled = true }

// ModifierVisitor defines the visitor pattern for applying modifiers to events
type ModifierVisitor interface {
	// Combat events
	VisitBeforeAttackRollEvent(*BeforeAttackRollEvent)
	VisitOnAttackRollEvent(*OnAttackRollEvent)
	VisitAfterAttackRollEvent(*AfterAttackRollEvent)
	VisitBeforeHitEvent(*BeforeHitEvent)
	VisitOnHitEvent(*OnHitEvent)
	VisitBeforeDamageRollEvent(*BeforeDamageRollEvent)
	VisitOnDamageRollEvent(*OnDamageRollEvent)
	VisitAfterDamageRollEvent(*AfterDamageRollEvent)
	VisitBeforeTakeDamageEvent(*BeforeTakeDamageEvent)
	VisitOnTakeDamageEvent(*OnTakeDamageEvent)
	VisitAfterTakeDamageEvent(*AfterTakeDamageEvent)

	// Ability events
	VisitBeforeAbilityCheckEvent(*BeforeAbilityCheckEvent)
	VisitOnAbilityCheckEvent(*OnAbilityCheckEvent)
	VisitAfterAbilityCheckEvent(*AfterAbilityCheckEvent)

	// Saving throw events
	VisitBeforeSavingThrowEvent(*BeforeSavingThrowEvent)
	VisitOnSavingThrowEvent(*OnSavingThrowEvent)
	VisitAfterSavingThrowEvent(*AfterSavingThrowEvent)
}

// BaseModifierVisitor provides default no-op implementations
type BaseModifierVisitor struct{}

// Combat event default implementations
func (v *BaseModifierVisitor) VisitBeforeAttackRollEvent(*BeforeAttackRollEvent) {}
func (v *BaseModifierVisitor) VisitOnAttackRollEvent(*OnAttackRollEvent)         {}
func (v *BaseModifierVisitor) VisitAfterAttackRollEvent(*AfterAttackRollEvent)   {}
func (v *BaseModifierVisitor) VisitBeforeHitEvent(*BeforeHitEvent)               {}
func (v *BaseModifierVisitor) VisitOnHitEvent(*OnHitEvent)                       {}
func (v *BaseModifierVisitor) VisitBeforeDamageRollEvent(*BeforeDamageRollEvent) {}
func (v *BaseModifierVisitor) VisitOnDamageRollEvent(*OnDamageRollEvent)         {}
func (v *BaseModifierVisitor) VisitAfterDamageRollEvent(*AfterDamageRollEvent)   {}
func (v *BaseModifierVisitor) VisitBeforeTakeDamageEvent(*BeforeTakeDamageEvent) {}
func (v *BaseModifierVisitor) VisitOnTakeDamageEvent(*OnTakeDamageEvent)         {}
func (v *BaseModifierVisitor) VisitAfterTakeDamageEvent(*AfterTakeDamageEvent)   {}

// Ability event default implementations
func (v *BaseModifierVisitor) VisitBeforeAbilityCheckEvent(*BeforeAbilityCheckEvent) {}
func (v *BaseModifierVisitor) VisitOnAbilityCheckEvent(*OnAbilityCheckEvent)         {}
func (v *BaseModifierVisitor) VisitAfterAbilityCheckEvent(*AfterAbilityCheckEvent)   {}

// Saving throw default implementations
func (v *BaseModifierVisitor) VisitBeforeSavingThrowEvent(*BeforeSavingThrowEvent) {}
func (v *BaseModifierVisitor) VisitOnSavingThrowEvent(*OnSavingThrowEvent)         {}
func (v *BaseModifierVisitor) VisitAfterSavingThrowEvent(*AfterSavingThrowEvent)   {}
