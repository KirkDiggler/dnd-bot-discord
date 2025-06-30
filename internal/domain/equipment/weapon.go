package equipment

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/damage"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/combat/attack"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
)

const (
	// WeaponKeyShortsword is the key for shortsword weapons
	WeaponKeyShortsword = "shortsword"
)

type Weapon struct {
	Base            BasicEquipment          `json:"base"`
	Damage          *damage.Damage          `json:"damage"`
	Range           int                     `json:"range"`
	WeaponCategory  string                  `json:"weapon_category"`
	WeaponRange     string                  `json:"weapon_range"`
	CategoryRange   string                  `json:"category_range"`
	Properties      []*shared.ReferenceItem `json:"properties"`
	TwoHandedDamage *damage.Damage          `json:"two_handed_damage"`
}

func (w *Weapon) Attack(char *character.Character) (*attack.Result, error) {
	var abilityBonus int

	switch w.WeaponRange {
	case "Ranged":
		abilityBonus = char.Attributes[shared.AttributeDexterity].Bonus
	case "Melee":
		abilityBonus = char.Attributes[shared.AttributeStrength].Bonus
	}

	// Check for weapon proficiency
	proficiencyBonus := 0
	if char.HasWeaponProficiency(w.Base.Key) {
		// Base proficiency bonus is +2 at level 1-4, +3 at 5-8, +4 at 9-12, etc.
		proficiencyBonus = 2 + ((char.Level - 1) / 4)
	}

	attackBonus := abilityBonus + proficiencyBonus
	damageBonus := abilityBonus // Only ability modifier applies to damage

	if w.IsTwoHanded() {
		if w.TwoHandedDamage == nil {
			return attack.RollAttack(char.getDiceRoller(), attackBonus, damageBonus, w.Damage)
		}

		return attack.RollAttack(char.getDiceRoller(), attackBonus, damageBonus, w.TwoHandedDamage)
	}

	return attack.RollAttack(char.getDiceRoller(), attackBonus, damageBonus, w.Damage)
}

func (w *Weapon) IsRanged() bool {
	return w.WeaponRange == "Ranged"
}

func (w *Weapon) IsMelee() bool {
	return w.WeaponRange == "Melee"
}

func (w *Weapon) IsSimple() bool {
	return w.HasProperty("simple")

}

func (w *Weapon) IsTwoHanded() bool {
	return w.HasProperty("two-handed")
}

func (w *Weapon) IsHeavy() bool {
	return w.HasProperty("heavy")
}

func (w *Weapon) IsFinesse() bool {
	return w.HasProperty("finesse")
}

// IsMonkWeapon returns true if this weapon can be used with monk Martial Arts
// Monk weapons are shortswords and any simple melee weapons that don't have
// the two-handed or heavy property
func (w *Weapon) IsMonkWeapon() bool {
	// Shortswords are always monk weapons
	if w.Base.Key == WeaponKeyShortsword {
		return true
	}

	// Simple melee weapons without two-handed or heavy properties
	if w.WeaponCategory == "Simple" && w.IsMelee() && !w.IsTwoHanded() && !w.IsHeavy() {
		return true
	}

	return false
}

// HasProperty checks if the weapon has a specific property
func (w *Weapon) HasProperty(prop string) bool {
	for _, p := range w.Properties {
		if p.Key == prop {
			return true
		}
	}

	return false
}

func (w *Weapon) GetEquipmentType() EquipmentType {
	return EquipmentTypeWeapon
}

func (w *Weapon) GetName() string {
	return w.Base.Name
}

func (w *Weapon) GetKey() string {
	return w.Base.Key
}

func (w *Weapon) GetSlot() character.Slot {
	for _, p := range w.Properties {
		if p.Key == "two-handed" {
			return character.SlotTwoHanded
		}
	}

	return character.SlotMainHand
}
