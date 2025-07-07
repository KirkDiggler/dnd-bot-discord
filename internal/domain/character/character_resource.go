package character

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
)

// CharacterResources tracks all expendable resources
type CharacterResources struct {
	HP            shared.HPResource                `json:"hp"`
	SpellSlots    map[int]shared.SpellSlotInfo     `json:"spell_slots"` // level -> slot info
	Abilities     map[string]*shared.ActiveAbility `json:"abilities"`   // key -> ability
	ActiveEffects []*shared.ActiveEffect           `json:"active_effects"`
	HitDice       shared.HitDiceResource           `json:"hit_dice"`

	// Combat state tracking
	SneakAttackUsedThisTurn bool `json:"sneak_attack_used_this_turn"`

	// Action economy for current turn
	ActionEconomy shared.ActionEconomy `json:"action_economy"`
}

// Initialize sets up resources based on class and level
func (r *CharacterResources) Initialize(class *rulebook.Class, level int) {
	// Initialize HP
	r.HP = shared.HPResource{
		Current: class.HitDie, // Will be modified by CON later
		Max:     class.HitDie,
	}

	// Initialize hit dice
	r.HitDice = shared.HitDiceResource{
		DiceType:  class.HitDie,
		Max:       level,
		Remaining: level,
	}

	// Initialize abilities map
	if r.Abilities == nil {
		r.Abilities = make(map[string]*shared.ActiveAbility)
	}

	// Initialize spell slots if caster
	r.initializeSpellSlots(class, level)
}

// initializeSpellSlots sets up spell slots for caster classes
func (r *CharacterResources) initializeSpellSlots(class *rulebook.Class, level int) {
	r.SpellSlots = make(map[int]shared.SpellSlotInfo)

	// Level 1 full casters (Cleric, Druid, Sorcerer, Wizard, Bard)
	switch class.Key {
	case "cleric", "druid", "sorcerer", "wizard", "bard":
		r.SpellSlots[1] = shared.SpellSlotInfo{
			Max:       2,
			Remaining: 2,
			Source:    "spellcasting",
		}
	case "ranger", "paladin": // Half casters don't get slots until level 2
		if level >= 2 {
			r.SpellSlots[1] = shared.SpellSlotInfo{
				Max:       2,
				Remaining: 2,
				Source:    "spellcasting",
			}
		}
	case "warlock": // Pact magic - different progression
		r.SpellSlots[1] = shared.SpellSlotInfo{
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

	r.SpellSlots[level] = shared.SpellSlotInfo{
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
		ability.RestoreUses(shared.RestTypeShort)
	}

	// Only warlock (pact magic) spell slots restore on short rest
	for level, slot := range r.SpellSlots {
		if slot.Source == "pact_magic" {
			r.SpellSlots[level] = shared.SpellSlotInfo{
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

	// Restore all abilities (RestoreUses also deactivates active abilities)
	for _, ability := range r.Abilities {
		ability.RestoreUses(shared.RestTypeLong)
	}

	// Restore all spell slots
	for level, slot := range r.SpellSlots {
		r.SpellSlots[level] = shared.SpellSlotInfo{
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

	// Clear temporary effects but keep permanent ones
	var permanentEffects []*shared.ActiveEffect
	for _, effect := range r.ActiveEffects {
		if effect.DurationType == shared.DurationTypePermanent {
			permanentEffects = append(permanentEffects, effect)
		}
	}
	r.ActiveEffects = permanentEffects
}

// AddEffect adds a new active effect
func (r *CharacterResources) AddEffect(effect *shared.ActiveEffect) {
	// Remove existing concentration if needed
	if effect.RequiresConcentration {
		r.RemoveConcentrationEffects()
	}
	r.ActiveEffects = append(r.ActiveEffects, effect)
}

// RemoveConcentrationEffects removes all concentration effects
func (r *CharacterResources) RemoveConcentrationEffects() {
	var nonConcentration []*shared.ActiveEffect
	for _, effect := range r.ActiveEffects {
		if !effect.RequiresConcentration {
			nonConcentration = append(nonConcentration, effect)
		}
	}
	r.ActiveEffects = nonConcentration
}

// TickEffectDurations advances all effect durations by one round
func (r *CharacterResources) TickEffectDurations() {
	var activeEffects []*shared.ActiveEffect
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
