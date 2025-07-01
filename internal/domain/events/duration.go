package events

// PermanentDuration represents a modifier that never expires
type PermanentDuration struct{}

func (p *PermanentDuration) IsExpired() bool {
	return false
}

func (p *PermanentDuration) OnEventOccurred(event *GameEvent) {}

// TurnDuration represents a modifier that lasts for a number of turns
type TurnDuration struct {
	turns          int
	extendOnCombat bool
	currentTurn    int
}

// NewTurnDuration creates a new turn-based duration
func NewTurnDuration(turns int, extendOnCombat bool) *TurnDuration {
	return &TurnDuration{
		turns:          turns,
		extendOnCombat: extendOnCombat,
		currentTurn:    0,
	}
}

func (t *TurnDuration) IsExpired() bool {
	return t.currentTurn >= t.turns
}

func (t *TurnDuration) OnEventOccurred(event *GameEvent) {
	if event.Type == OnTurnEnd && event.Actor != nil {
		t.currentTurn++
	}
	// If extendOnCombat is true and combat action occurred, reset the turn count
	if t.extendOnCombat {
		switch event.Type {
		case OnAttackRoll, OnDamageRoll:
			t.currentTurn = 0
		}
	}
}

// UntilEventDuration represents a modifier that lasts until a specific event occurs
type UntilEventDuration struct {
	targetEventType EventType
	expired         bool
}

// NewUntilEventDuration creates a duration that lasts until a specific event
func NewUntilEventDuration(eventType EventType) *UntilEventDuration {
	return &UntilEventDuration{
		targetEventType: eventType,
		expired:         false,
	}
}

func (u *UntilEventDuration) IsExpired() bool {
	return u.expired
}

func (u *UntilEventDuration) OnEventOccurred(event *GameEvent) {
	if event.Type == u.targetEventType {
		u.expired = true
	}
}

// ConcentrationDuration represents a modifier that requires concentration
type ConcentrationDuration struct {
	baseDuration ModifierDuration
	broken       bool
}

// NewConcentrationDuration creates a new concentration-based duration
func NewConcentrationDuration(baseDuration ModifierDuration) *ConcentrationDuration {
	return &ConcentrationDuration{
		baseDuration: baseDuration,
		broken:       false,
	}
}

func (c *ConcentrationDuration) IsExpired() bool {
	return c.broken || c.baseDuration.IsExpired()
}

func (c *ConcentrationDuration) OnEventOccurred(event *GameEvent) {
	// Pass event to base duration
	c.baseDuration.OnEventOccurred(event)

	// Check if concentration is broken
	// TODO: Implement concentration saves when OnTakeDamage occurs
	// This would trigger a Constitution saving throw to maintain concentration
}
