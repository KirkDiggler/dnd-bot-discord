package combat

import (
	"encoding/json"
	"fmt"
	"time"
)

// EncounterStatus represents the current state of an encounter
type EncounterStatus string

const (
	EncounterStatusSetup     EncounterStatus = "setup"     // Setting up the encounter
	EncounterStatusRolling   EncounterStatus = "rolling"   // Rolling initiative
	EncounterStatusActive    EncounterStatus = "active"    // Combat in progress
	EncounterStatusCompleted EncounterStatus = "completed" // Encounter finished
)

// CombatantType represents the type of combatant
type CombatantType string

const (
	CombatantTypePlayer  CombatantType = "player"
	CombatantTypeMonster CombatantType = "monster"
	CombatantTypeNPC     CombatantType = "npc"
)

// Encounter represents a combat encounter in a session
type Encounter struct {
	ID          string                `json:"id"`
	SessionID   string                `json:"session_id"`
	MessageID   string                `json:"message_id"` // Discord message ID for the encounter
	ChannelID   string                `json:"channel_id"` // Discord channel ID
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Status      EncounterStatus       `json:"status"`
	Round       int                   `json:"round"`      // Current round number
	Turn        int                   `json:"turn"`       // Current turn index
	Combatants  map[string]*Combatant `json:"combatants"` // ID -> Combatant
	TurnOrder   []string              `json:"turn_order"` // Ordered list of combatant IDs
	CreatedAt   time.Time             `json:"created_at"`
	StartedAt   *time.Time            `json:"started_at"`
	EndedAt     *time.Time            `json:"ended_at"`
	CreatedBy   string                `json:"created_by"` // User ID who created the encounter
	CombatLog   []string              `json:"combat_log"` // History of combat actions
}

// Combatant represents a participant in combat
type Combatant struct {
	ID              string        `json:"id"`
	Name            string        `json:"name"`
	Type            CombatantType `json:"type"`
	Initiative      int           `json:"initiative"`
	InitiativeBonus int           `json:"initiative_bonus"`
	CurrentHP       int           `json:"current_hp"`
	MaxHP           int           `json:"max_hp"`
	TempHP          int           `json:"temp_hp"`
	AC              int           `json:"ac"`
	Speed           int           `json:"speed"`
	Conditions      []string      `json:"conditions"` // Poisoned, Stunned, etc.
	IsActive        bool          `json:"is_active"`  // Still in combat
	HasActed        bool          `json:"has_acted"`  // Has taken turn this round

	// For players
	PlayerID    string `json:"player_id,omitempty"`
	CharacterID string `json:"character_id,omitempty"`
	Class       string `json:"class,omitempty"` // Character class (Fighter, Wizard, etc.)
	Race        string `json:"race,omitempty"`  // Character race

	// For monsters
	MonsterRef string           `json:"monster_ref,omitempty"` // D&D API reference
	CR         float64          `json:"cr,omitempty"`          // Challenge Rating
	XP         int              `json:"xp,omitempty"`          // Experience Points
	Abilities  map[string]int   `json:"abilities,omitempty"`   // STR, DEX, etc.
	Actions    []*MonsterAction `json:"actions,omitempty"`     // Available actions
}

// IsAlive returns true if the combatant has more than 0 HP
func (c *Combatant) IsAlive() bool {
	return c.CurrentHP > 0
}

// CanAct returns true if the combatant is alive and has actions available
func (c *Combatant) CanAct() bool {
	return c.IsAlive() && len(c.Actions) > 0
}

// NewEncounter creates a new encounter
func NewEncounter(id, sessionID, channelID, name, createdBy string) *Encounter {
	return &Encounter{
		ID:         id,
		SessionID:  sessionID,
		ChannelID:  channelID,
		Name:       name,
		Status:     EncounterStatusSetup,
		Round:      0,
		Turn:       0,
		Combatants: make(map[string]*Combatant),
		TurnOrder:  []string{},
		CreatedAt:  time.Now(),
		CreatedBy:  createdBy,
		CombatLog:  []string{},
	}
}

// AddCombatant adds a new combatant to the encounter
func (e *Encounter) AddCombatant(combatant *Combatant) {
	e.Combatants[combatant.ID] = combatant
}

// RemoveCombatant removes a combatant from the encounter
func (e *Encounter) RemoveCombatant(id string) {
	delete(e.Combatants, id)
	// Remove from turn order
	newOrder := []string{}
	for _, cid := range e.TurnOrder {
		if cid != id {
			newOrder = append(newOrder, cid)
		}
	}
	e.TurnOrder = newOrder
}

// Start begins the encounter after initiative is rolled
func (e *Encounter) Start() bool {
	if e.Status != EncounterStatusRolling || len(e.TurnOrder) == 0 {
		return false
	}

	now := time.Now()
	e.Status = EncounterStatusActive
	e.StartedAt = &now
	e.Round = 1
	e.Turn = 0

	// Check if combat should end before starting
	if shouldEnd, _ := e.CheckCombatEnd(); shouldEnd {
		e.End()
		return false
	}

	// Skip dead combatants at the start
	for e.Turn < len(e.TurnOrder) {
		if combatant, exists := e.Combatants[e.TurnOrder[e.Turn]]; exists && combatant.IsActive && combatant.CurrentHP > 0 {
			// Found an active, alive combatant
			break
		}
		// Skip this dead/inactive combatant
		e.Turn++
	}

	return true
}

// NextTurn advances to the next turn
func (e *Encounter) NextTurn() {
	if e.Status != EncounterStatusActive {
		return
	}

	// Mark current combatant as having acted
	if e.Turn < len(e.TurnOrder) {
		if combatant, exists := e.Combatants[e.TurnOrder[e.Turn]]; exists {
			combatant.HasActed = true
		}
	}

	// Advance turn
	e.Turn++

	// Check if combat should end before skipping dead combatants
	if shouldEnd, _ := e.CheckCombatEnd(); shouldEnd {
		e.End()
		return
	}

	// Skip dead combatants
	for e.Turn < len(e.TurnOrder) {
		if combatant, exists := e.Combatants[e.TurnOrder[e.Turn]]; exists && combatant.IsActive && combatant.CurrentHP > 0 {
			// Found an active, alive combatant
			break
		}
		// Skip this dead/inactive combatant
		e.Turn++
	}

	// Check if round is complete
	if e.Turn >= len(e.TurnOrder) {
		// Start new round
		e.Round++
		e.Turn = 0

		// Reset all combatants' HasActed flag
		for _, combatant := range e.Combatants {
			combatant.HasActed = false
		}

		// Check if combat should end after round reset
		if shouldEnd, _ := e.CheckCombatEnd(); shouldEnd {
			e.End()
			return
		}

		// Skip dead combatants at the start of the new round
		for e.Turn < len(e.TurnOrder) {
			if combatant, exists := e.Combatants[e.TurnOrder[e.Turn]]; exists && combatant.IsActive && combatant.CurrentHP > 0 {
				// Found an active, alive combatant
				break
			}
			// Skip this dead/inactive combatant
			e.Turn++
		}
	}
}

// NextRound advances to the next round
func (e *Encounter) NextRound() {
	// NextTurn handles round advancement when Turn >= len(TurnOrder)
	// This method is kept for backward compatibility but just delegates to NextTurn
	if e.Turn < len(e.TurnOrder) {
		// Force advancement to end of turn order to trigger new round
		e.Turn = len(e.TurnOrder)
	}
	e.NextTurn()
}

// GetCurrentCombatant returns the combatant whose turn it is
func (e *Encounter) GetCurrentCombatant() *Combatant {
	if e.Turn < len(e.TurnOrder) {
		return e.Combatants[e.TurnOrder[e.Turn]]
	}
	return nil
}

// End concludes the encounter
func (e *Encounter) End() {
	now := time.Now()
	e.Status = EncounterStatusCompleted
	e.EndedAt = &now
}

// IsPlayerTurn checks if it's a specific player's turn
func (e *Encounter) IsPlayerTurn(playerID string) bool {
	current := e.GetCurrentCombatant()
	return current != nil && current.PlayerID == playerID
}

// CanPlayerAct checks if a player can take an action
func (e *Encounter) CanPlayerAct(playerID string) bool {
	// DM can always act (for NPCs/monsters)
	if playerID == e.CreatedBy {
		return true
	}

	// Players can act on their turn
	return e.IsPlayerTurn(playerID)
}

// CheckCombatEnd checks if combat should end (all enemies or all players defeated)
func (e *Encounter) CheckCombatEnd() (shouldEnd, playersWon bool) {
	activeMonsters := 0
	activePlayers := 0

	for _, combatant := range e.Combatants {
		if combatant.IsActive {
			switch combatant.Type {
			case CombatantTypeMonster:
				activeMonsters++
			case CombatantTypePlayer:
				activePlayers++
			}
		}
	}

	// Combat ends if either side has no active combatants
	if activeMonsters == 0 && activePlayers > 0 {
		return true, true // Players won
	} else if activePlayers == 0 && activeMonsters > 0 {
		return true, false // Players lost
	}

	return false, false
}

// ApplyDamage applies damage to a combatant
func (c *Combatant) ApplyDamage(damage int) {
	// First reduce temp HP
	if c.TempHP > 0 {
		if damage <= c.TempHP {
			c.TempHP -= damage
			return
		}
		damage -= c.TempHP
		c.TempHP = 0
	}

	// Then reduce current HP
	c.CurrentHP -= damage
	if c.CurrentHP < 0 {
		c.CurrentHP = 0
		c.IsActive = false
	}
}

// Heal restores hit points to a combatant
func (c *Combatant) Heal(amount int) {
	c.CurrentHP += amount
	if c.CurrentHP > c.MaxHP {
		c.CurrentHP = c.MaxHP
	}

	// If they were at 0, they're back in the fight
	if c.CurrentHP > 0 && !c.IsActive {
		c.IsActive = true
	}
}

// AddTempHP adds temporary hit points
func (c *Combatant) AddTempHP(amount int) {
	// Temp HP doesn't stack, take the higher value
	if amount > c.TempHP {
		c.TempHP = amount
	}
}

// AddCombatLogEntry adds an entry to the combat log
func (e *Encounter) AddCombatLogEntry(entry string) {
	if e.CombatLog == nil {
		e.CombatLog = []string{}
	}
	// Prefix with round number
	logEntry := fmt.Sprintf("Round %d: %s", e.Round, entry)
	e.CombatLog = append(e.CombatLog, logEntry)

	// Keep only last 20 entries to prevent unbounded growth
	if len(e.CombatLog) > 20 {
		e.CombatLog = e.CombatLog[len(e.CombatLog)-20:]
	}
}

// IsRoundComplete checks if all active combatants have acted this round
func (e *Encounter) IsRoundComplete() bool {
	for _, combatant := range e.Combatants {
		if combatant.IsActive && !combatant.HasActed {
			return false
		}
	}
	return true
}

// MarshalJSON implements json.Marshaler
func (e *Encounter) MarshalJSON() ([]byte, error) {
	type Alias Encounter
	return json.Marshal((*Alias)(e))
}
