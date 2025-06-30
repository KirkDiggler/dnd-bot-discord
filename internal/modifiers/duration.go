package modifiers

import (
	"fmt"
	"github.com/KirkDiggler/dnd-bot-discord/internal/events"
)

// PermanentDuration never expires
type PermanentDuration struct{}

func (d *PermanentDuration) IsExpired() bool                    { return false }
func (d *PermanentDuration) OnEventOccurred(event events.Event) {}
func (d *PermanentDuration) String() string                     { return "permanent" }

// RoundsDuration lasts for a number of rounds
type RoundsDuration struct {
	Rounds    int
	remaining int
}

func NewRoundsDuration(rounds int) *RoundsDuration {
	return &RoundsDuration{
		Rounds:    rounds,
		remaining: rounds,
	}
}

func (d *RoundsDuration) IsExpired() bool { return d.remaining <= 0 }

func (d *RoundsDuration) OnEventOccurred(event events.Event) {
	// Decrement on turn end
	if event.GetType() == events.EventTypeOnTurnEnd {
		d.remaining--
	}
}

func (d *RoundsDuration) String() string {
	return fmt.Sprintf("%d rounds remaining", d.remaining)
}

// UntilRestDuration lasts until a rest
type UntilRestDuration struct {
	expired bool
}

func (d *UntilRestDuration) IsExpired() bool { return d.expired }

func (d *UntilRestDuration) OnEventOccurred(event events.Event) {
	if event.GetType() == events.EventTypeOnShortRest || event.GetType() == events.EventTypeOnLongRest {
		d.expired = true
	}
}

func (d *UntilRestDuration) String() string { return "until rest" }

// ConcentrationDuration lasts while concentration is maintained
type ConcentrationDuration struct {
	expired   bool
	rounds    int
	maxRounds int
}

func NewConcentrationDuration(maxRounds int) *ConcentrationDuration {
	return &ConcentrationDuration{
		maxRounds: maxRounds,
		rounds:    maxRounds,
	}
}

func (d *ConcentrationDuration) IsExpired() bool { return d.expired || d.rounds <= 0 }

func (d *ConcentrationDuration) OnEventOccurred(event events.Event) {
	switch event.GetType() {
	case events.EventTypeOnTurnEnd:
		d.rounds--
	case events.EventTypeAfterTakeDamage:
		// Concentration would require a saving throw here
		// For now, we'll handle this elsewhere
	}
}

func (d *ConcentrationDuration) Break() {
	d.expired = true
}

func (d *ConcentrationDuration) String() string {
	if d.expired {
		return "concentration broken"
	}
	return fmt.Sprintf("concentration (%d rounds)", d.rounds)
}
