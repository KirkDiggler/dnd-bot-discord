package events

import (
	"fmt"
)

// PermanentDuration never expires
type PermanentDuration struct{}

func (d PermanentDuration) IsExpired(event *GameEvent) bool {
	return false
}

func (d PermanentDuration) Description() string {
	return "Permanent"
}

// RoundsDuration expires after a number of rounds
type RoundsDuration struct {
	Rounds    int
	StartTurn int
}

func (d *RoundsDuration) IsExpired(event *GameEvent) bool {
	if event.Type != OnTurnStart && event.Type != OnTurnEnd {
		return false
	}

	currentTurn, exists := event.GetIntContext("turn_count")
	if !exists {
		return false
	}

	return currentTurn >= d.StartTurn+d.Rounds
}

func (d *RoundsDuration) Description() string {
	return fmt.Sprintf("%d rounds", d.Rounds)
}

// EncounterDuration lasts until the end of combat
type EncounterDuration struct{}

func (d EncounterDuration) IsExpired(event *GameEvent) bool {
	if event.Type != OnStatusRemoved {
		return false
	}

	status, _ := event.GetStringContext("status")
	return status == "combat_ended"
}

func (d EncounterDuration) Description() string {
	return "Until end of encounter"
}

// ConcentrationDuration lasts until concentration is broken
type ConcentrationDuration struct {
	SpellName string
}

func (d *ConcentrationDuration) IsExpired(event *GameEvent) bool {
	if event.Type != OnStatusRemoved {
		return false
	}

	status, _ := event.GetStringContext("status")
	return status == "concentration_broken"
}

func (d *ConcentrationDuration) Description() string {
	return fmt.Sprintf("Concentration (%s)", d.SpellName)
}

// ShortRestDuration expires on a short rest
type ShortRestDuration struct{}

func (d ShortRestDuration) IsExpired(event *GameEvent) bool {
	return event.Type == OnShortRest || event.Type == OnLongRest
}

func (d ShortRestDuration) Description() string {
	return "Until short rest"
}

// LongRestDuration expires on a long rest
type LongRestDuration struct{}

func (d LongRestDuration) IsExpired(event *GameEvent) bool {
	return event.Type == OnLongRest
}

func (d LongRestDuration) Description() string {
	return "Until long rest"
}

// UntilDamagedDuration expires when the character takes damage
type UntilDamagedDuration struct{}

func (d UntilDamagedDuration) IsExpired(event *GameEvent) bool {
	return event.Type == OnTakeDamage
}

func (d UntilDamagedDuration) Description() string {
	return "Until damaged"
}
