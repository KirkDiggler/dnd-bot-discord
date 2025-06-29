package entities

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

// CharacterResources tracks all expendable resources
type CharacterResources struct {
	HP            HPResource                `json:"hp"`
	SpellSlots    map[int]SpellSlotInfo     `json:"spell_slots"` // level -> slot info
	Abilities     map[string]*ActiveAbility `json:"abilities"`   // key -> ability
	ActiveEffects []*ActiveEffect           `json:"active_effects"`
	HitDice       HitDiceResource           `json:"hit_dice"`

	// Combat state tracking
	SneakAttackUsedThisTurn bool `json:"sneak_attack_used_this_turn"`

	// Action economy for current turn
	ActionEconomy ActionEconomy `json:"action_economy"`
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

// Initialize sets up resources based on class and level
func (r *CharacterResources) Initialize(class *Class, level int) {
	// Initialize HP
	r.HP = HPResource{
		Current: class.HitDie, // Will be modified by CON later
		Max:     class.HitDie,
	}

	// Initialize hit dice
	r.HitDice = HitDiceResource{
		DiceType:  class.HitDie,
		Max:       level,
		Remaining: level,
	}

	// Initialize abilities map
	if r.Abilities == nil {
		r.Abilities = make(map[string]*ActiveAbility)
	}

	// Initialize spell slots if caster
	r.initializeSpellSlots(class, level)
}

// initializeSpellSlots sets up spell slots for caster classes
func (r *CharacterResources) initializeSpellSlots(class *Class, level int) {
	r.SpellSlots = make(map[int]SpellSlotInfo)

	// Level 1 full casters (Cleric, Druid, Sorcerer, Wizard, Bard)
	switch class.Key {
	case "cleric", "druid", "sorcerer", "wizard", "bard":
		r.SpellSlots[1] = SpellSlotInfo{
			Max:       2,
			Remaining: 2,
			Source:    "spellcasting",
		}
	case "ranger", "paladin": // Half casters don't get slots until level 2
		if level >= 2 {
			r.SpellSlots[1] = SpellSlotInfo{
				Max:       2,
				Remaining: 2,
				Source:    "spellcasting",
			}
		}
	case "warlock": // Pact magic - different progression
		r.SpellSlots[1] = SpellSlotInfo{
			Max:       1,
			Remaining: 1,
			Source:    "pact_magic",
		}
	}
}

// UseSpellSlot consumes a spell slot of the given level
func (r *CharacterResources) UseSpellSlot(level int) bool {
	slot, exists := r.SpellSlots[level]
	if !exists || slot.Remaining <= 0 {
		return false
	}

	r.SpellSlots[level] = SpellSlotInfo{
		Max:       slot.Max,
		Remaining: slot.Remaining - 1,
		Source:    slot.Source,
	}
	return true
}

// ShortRest restores short rest resources
func (r *CharacterResources) ShortRest() {
	// Restore abilities that come back on short rest
	for _, ability := range r.Abilities {
		ability.RestoreUses(RestTypeShort)
	}

	// Only warlock (pact magic) spell slots restore on short rest
	for level, slot := range r.SpellSlots {
		if slot.Source == "pact_magic" {
			r.SpellSlots[level] = SpellSlotInfo{
				Max:       slot.Max,
				Remaining: slot.Max,
				Source:    slot.Source,
			}
		}
	}
}

// LongRest restores all resources
func (r *CharacterResources) LongRest() {
	// Restore HP to max
	r.HP.Current = r.HP.Max
	r.HP.Temporary = 0

	// Restore all abilities
	for _, ability := range r.Abilities {
		ability.RestoreUses(RestTypeLong)
	}

	// Restore all spell slots
	for level, slot := range r.SpellSlots {
		r.SpellSlots[level] = SpellSlotInfo{
			Max:       slot.Max,
			Remaining: slot.Max,
			Source:    slot.Source,
		}
	}

	// Restore half hit dice (minimum 1)
	restored := r.HitDice.Max / 2
	if restored < 1 {
		restored = 1
	}
	r.HitDice.Remaining += restored
	if r.HitDice.Remaining > r.HitDice.Max {
		r.HitDice.Remaining = r.HitDice.Max
	}

	// Remove effects that end on rest
	var persistentEffects []*ActiveEffect
	for _, effect := range r.ActiveEffects {
		if effect.DurationType != DurationTypeUntilRest {
			persistentEffects = append(persistentEffects, effect)
		}
	}
	r.ActiveEffects = persistentEffects
}

// AddEffect adds a new active effect
func (r *CharacterResources) AddEffect(effect *ActiveEffect) {
	// Remove existing concentration if needed
	if effect.RequiresConcentration {
		r.RemoveConcentrationEffects()
	}
	r.ActiveEffects = append(r.ActiveEffects, effect)
}

// RemoveConcentrationEffects removes all concentration effects
func (r *CharacterResources) RemoveConcentrationEffects() {
	var nonConcentration []*ActiveEffect
	for _, effect := range r.ActiveEffects {
		if !effect.RequiresConcentration {
			nonConcentration = append(nonConcentration, effect)
		}
	}
	r.ActiveEffects = nonConcentration
}

// TickEffectDurations advances all effect durations by one round
func (r *CharacterResources) TickEffectDurations() {
	var activeEffects []*ActiveEffect
	for _, effect := range r.ActiveEffects {
		if !effect.TickDuration() {
			activeEffects = append(activeEffects, effect)
		}
	}
	r.ActiveEffects = activeEffects

	// Also tick ability durations
	for _, ability := range r.Abilities {
		ability.TickDuration()
	}
}

// GetTotalACBonus calculates total AC bonus from all effects
func (r *CharacterResources) GetTotalACBonus() int {
	bonus := 0
	for _, effect := range r.ActiveEffects {
		bonus += effect.GetACBonus()
	}
	return bonus
}

// GetTotalDamageBonus calculates total damage bonus for a damage type
func (r *CharacterResources) GetTotalDamageBonus(damageType string) int {
	bonus := 0
	for _, effect := range r.ActiveEffects {
		bonus += effect.GetDamageBonus(damageType)
	}
	return bonus
}

// HasResistance checks if any effect provides resistance
func (r *CharacterResources) HasResistance(damageType string) bool {
	for _, effect := range r.ActiveEffects {
		if effect.HasResistance(damageType) {
			return true
		}
	}
	return false
}
