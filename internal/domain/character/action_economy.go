package character

import "time"

// ActionEconomy tracks available actions for a character during their turn
type ActionEconomy struct {
	// Core actions
	ActionUsed      bool `json:"action_used"`
	BonusActionUsed bool `json:"bonus_action_used"`
	ReactionUsed    bool `json:"reaction_used"`
	MovementUsed    int  `json:"movement_used"`

	// What was done this turn (for triggering bonus actions)
	ActionsThisTurn []ActionRecord `json:"actions_this_turn"`

	// Available bonus actions based on triggers
	AvailableBonusActions []BonusActionOption `json:"available_bonus_actions"`
}

// ActionRecord tracks what actions were taken
type ActionRecord struct {
	Type      string    `json:"type"`       // "attack", "spell", "dash", etc.
	Subtype   string    `json:"subtype"`    // "unarmed_strike", "monk_weapon", etc.
	WeaponKey string    `json:"weapon_key"` // If attack, what weapon
	Timestamp time.Time `json:"timestamp"`
}

// BonusActionOption represents an available bonus action
type BonusActionOption struct {
	Key         string `json:"key"`         // "martial_arts_strike", "two_weapon_attack", etc.
	Name        string `json:"name"`        // Display name
	Description string `json:"description"` // What it does
	Source      string `json:"source"`      // "martial_arts", "cunning_action", etc.
	ActionType  string `json:"action_type"` // "attack", "dash", "hide", etc.
}

// Reset clears the action economy for a new turn
func (ae *ActionEconomy) Reset() {
	ae.ActionUsed = false
	ae.BonusActionUsed = false
	ae.MovementUsed = 0
	ae.ActionsThisTurn = []ActionRecord{}
	ae.AvailableBonusActions = []BonusActionOption{}
	// Note: ReactionUsed is NOT reset here - reactions reset at start of YOUR turn
}

// RecordAction adds an action to the history
func (ae *ActionEconomy) RecordAction(actionType, subtype, weaponKey string) {
	record := ActionRecord{
		Type:      actionType,
		Subtype:   subtype,
		WeaponKey: weaponKey,
		Timestamp: time.Now(),
	}
	ae.ActionsThisTurn = append(ae.ActionsThisTurn, record)

	// Mark action as used if it's a main action
	if actionType == "attack" || actionType == "spell" || actionType == "dash" ||
		actionType == "dodge" || actionType == "help" || actionType == "ready" {
		ae.ActionUsed = true
	}
}

// HasTakenAction checks if a specific action type was taken this turn
func (ae *ActionEconomy) HasTakenAction(actionType string) bool {
	for _, action := range ae.ActionsThisTurn {
		if action.Type == actionType {
			return true
		}
	}
	return false
}

// GetActionsByType returns all actions of a specific type taken this turn
func (ae *ActionEconomy) GetActionsByType(actionType string) []ActionRecord {
	var actions []ActionRecord
	for _, action := range ae.ActionsThisTurn {
		if action.Type == actionType {
			actions = append(actions, action)
		}
	}
	return actions
}
