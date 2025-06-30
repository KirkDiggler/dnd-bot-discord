package character

// AbilityType represents the action cost of using an ability
type AbilityType string

const (
	AbilityTypeAction      AbilityType = "action"
	AbilityTypeBonusAction AbilityType = "bonus_action"
	AbilityTypeReaction    AbilityType = "reaction"
	AbilityTypeFree        AbilityType = "free" // No action required
)

// RestType represents when an ability's uses are restored
type RestType string

const (
	RestTypeShort RestType = "short"
	RestTypeLong  RestType = "long"
	RestTypeNone  RestType = "none" // Never restored (consumables)
)

// ActiveAbility represents an ability that can be used in combat
type ActiveAbility struct {
	Key           string      `json:"key"`
	Name          string      `json:"name"`
	Description   string      `json:"description"`
	FeatureKey    string      `json:"feature_key"` // Links to CharacterFeature
	ActionType    AbilityType `json:"action_type"`
	UsesMax       int         `json:"uses_max"` // -1 for unlimited
	UsesRemaining int         `json:"uses_remaining"`
	RestType      RestType    `json:"rest_type"`
	IsActive      bool        `json:"is_active"` // For toggle abilities like Rage
	Duration      int         `json:"duration"`  // Rounds remaining (-1 for unlimited)
}

// CanUse checks if the ability can be used
func (a *ActiveAbility) CanUse() bool {
	if a.UsesMax == -1 {
		return true // Unlimited uses
	}
	return a.UsesRemaining > 0
}

// Use decrements the uses remaining
func (a *ActiveAbility) Use() bool {
	if !a.CanUse() {
		return false
	}
	if a.UsesMax != -1 {
		a.UsesRemaining--
	}
	if a.Duration > 0 {
		a.IsActive = true
	}
	return true
}

// Deactivate turns off a toggle ability
func (a *ActiveAbility) Deactivate() {
	a.IsActive = false
	a.Duration = 0
}

// TickDuration decrements duration and deactivates if expired
func (a *ActiveAbility) TickDuration() {
	if a.Duration > 0 {
		a.Duration--
		if a.Duration == 0 {
			a.Deactivate()
		}
	}
}

// RestoreUses restores uses based on rest type
func (a *ActiveAbility) RestoreUses(restType RestType) {
	if a.RestType == RestTypeNone {
		return // Never restored
	}

	// Long rest restores everything
	if restType == RestTypeLong {
		a.UsesRemaining = a.UsesMax
		a.Deactivate() // End any active effects
		return
	}

	// Short rest only restores short rest abilities
	if restType == RestTypeShort && a.RestType == RestTypeShort {
		a.UsesRemaining = a.UsesMax
		a.Deactivate()
	}
}
