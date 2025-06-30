package shared

// HPResource tracks hit points and temporary HP
type HPResource struct {
	Current   int `json:"current"`
	Max       int `json:"max"`
	Temporary int `json:"temporary"`
}

// Damage applies damage, using temp HP first
func (hp *HPResource) Damage(amount int) int {
	if amount <= 0 {
		return 0
	}

	originalAmount := amount

	// Apply to temporary HP first
	if hp.Temporary > 0 {
		if hp.Temporary >= amount {
			hp.Temporary -= amount
			return originalAmount // All absorbed by temp HP
		}
		amount -= hp.Temporary
		hp.Temporary = 0
	}

	// Apply remaining to current HP
	hp.Current -= amount
	if hp.Current < 0 {
		hp.Current = 0
	}

	return originalAmount // Return total damage dealt
}

// Heal restores hit points up to max
func (hp *HPResource) Heal(amount int) int {
	if amount <= 0 || hp.Current >= hp.Max {
		return 0
	}

	oldHP := hp.Current
	hp.Current += amount
	if hp.Current > hp.Max {
		hp.Current = hp.Max
	}

	return hp.Current - oldHP
}

// AddTemporaryHP adds temporary hit points (doesn't stack)
func (hp *HPResource) AddTemporaryHP(amount int) {
	if amount > hp.Temporary {
		hp.Temporary = amount
	}
}

// SpellSlotInfo tracks spell slots at a specific level
type SpellSlotInfo struct {
	Max       int    `json:"max"`
	Remaining int    `json:"remaining"`
	Source    string `json:"source"` // "spellcasting" or "pact_magic"
}

// HitDiceResource tracks hit dice for healing
type HitDiceResource struct {
	DiceType  int `json:"dice_type"` // d6, d8, d10, d12
	Max       int `json:"max"`       // Usually equals level
	Remaining int `json:"remaining"`
}
