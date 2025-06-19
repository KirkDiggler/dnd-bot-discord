package entities

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities/attack"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities/damage"
)

type Weapon struct {
	Base            BasicEquipment   `json:"base"`
	Damage          *damage.Damage   `json:"damage"`
	Range           int              `json:"range"`
	WeaponCategory  string           `json:"weapon_category"`
	WeaponRange     string           `json:"weapon_range"`
	CategoryRange   string           `json:"category_range"`
	Properties      []*ReferenceItem `json:"properties"`
	TwoHandedDamage *damage.Damage   `json:"two_handed_damage"`
}

func (w *Weapon) Attack(char *Character) (*attack.Result, error) {
	var abilityBonus int

	switch w.WeaponRange {
	case "Ranged":
		abilityBonus = char.Attributes[AttributeDexterity].Bonus
	case "Melee":
		abilityBonus = char.Attributes[AttributeStrength].Bonus
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
			return attack.RollAttack(attackBonus, damageBonus, w.Damage)
		}

		return attack.RollAttack(attackBonus, damageBonus, w.TwoHandedDamage)
	}

	return attack.RollAttack(attackBonus, damageBonus, w.Damage)
}

func (w *Weapon) IsRanged() bool {
	return w.WeaponRange == "Ranged"
}

func (w *Weapon) IsMelee() bool {
	return w.WeaponRange == "Melee"
}

func (w *Weapon) IsSimple() bool {
	return w.hasProperty("simple")

}

func (w *Weapon) IsTwoHanded() bool {
	return w.hasProperty("two-handed")
}

func (w *Weapon) hasProperty(prop string) bool {
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

func (w *Weapon) GetSlot() Slot {
	for _, p := range w.Properties {
		if p.Key == "two-handed" {
			return SlotTwoHanded
		}
	}

	return SlotMainHand
}
