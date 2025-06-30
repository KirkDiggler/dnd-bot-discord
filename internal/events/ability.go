package events

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
)

// BeforeAbilityCheckEvent is emitted before an ability check is made
type BeforeAbilityCheckEvent struct {
	BaseEvent
	Ability      entities.Attribute // STR, DEX, etc.
	Skill        string             // Optional: athletics, acrobatics, etc.
	CheckBonus   int
	Advantage    bool
	Disadvantage bool
	DC           int // Difficulty Class if known
}

func (e *BeforeAbilityCheckEvent) Accept(v ModifierVisitor) {
	v.VisitBeforeAbilityCheckEvent(e)
}

// OnAbilityCheckEvent is emitted when the ability check is made
type OnAbilityCheckEvent struct {
	BaseEvent
	Ability    entities.Attribute
	Skill      string
	BaseRoll   int
	CheckBonus int
	TotalCheck int
	IsCritical bool // Natural 20
	IsFumble   bool // Natural 1
}

func (e *OnAbilityCheckEvent) Accept(v ModifierVisitor) {
	v.VisitOnAbilityCheckEvent(e)
}

// AfterAbilityCheckEvent is emitted after the ability check is complete
type AfterAbilityCheckEvent struct {
	BaseEvent
	Ability    entities.Attribute
	Skill      string
	TotalCheck int
	DC         int
	Success    bool
}

func (e *AfterAbilityCheckEvent) Accept(v ModifierVisitor) {
	v.VisitAfterAbilityCheckEvent(e)
}

// BeforeSavingThrowEvent is emitted before a saving throw is made
type BeforeSavingThrowEvent struct {
	BaseEvent
	Ability      entities.Attribute
	SaveBonus    int
	Advantage    bool
	Disadvantage bool
	DC           int
	Source       string // spell, effect, trap, etc.
}

func (e *BeforeSavingThrowEvent) Accept(v ModifierVisitor) {
	v.VisitBeforeSavingThrowEvent(e)
}

// OnSavingThrowEvent is emitted when the saving throw is made
type OnSavingThrowEvent struct {
	BaseEvent
	Ability    entities.Attribute
	BaseRoll   int
	SaveBonus  int
	TotalSave  int
	IsCritical bool // Natural 20
	IsFumble   bool // Natural 1
}

func (e *OnSavingThrowEvent) Accept(v ModifierVisitor) {
	v.VisitOnSavingThrowEvent(e)
}

// AfterSavingThrowEvent is emitted after the saving throw is complete
type AfterSavingThrowEvent struct {
	BaseEvent
	Ability   entities.Attribute
	TotalSave int
	DC        int
	Success   bool
	Source    string
}

func (e *AfterSavingThrowEvent) Accept(v ModifierVisitor) {
	v.VisitAfterSavingThrowEvent(e)
}
