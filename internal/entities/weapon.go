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
	var bonus int
	if w.WeaponRange == "Ranged" {
		bonus = char.Attributes[AttributeDexterity].Bonus
	} else if w.WeaponRange == "Melee" {

		bonus = char.Attributes[AttributeStrength].Bonus
	}

	// TODO: check proficiency
	if w.IsTwoHanded() {
		if w.TwoHandedDamage == nil {
			return attack.RollAttack(bonus, bonus, w.Damage)
		}

		return attack.RollAttack(bonus, bonus, w.TwoHandedDamage)
	}

	return attack.RollAttack(bonus, bonus, w.Damage)
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
	return "Weapon"
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
