package equipment

import "github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"

type ArmorCategory string

const (
	ArmorCategoryLight   ArmorCategory = "light"
	ArmorCategoryMedium  ArmorCategory = "medium"
	ArmorCategoryHeavy   ArmorCategory = "heavy"
	ArmorCategoryShield  ArmorCategory = "shield"
	ArmorCategoryUnknown ArmorCategory = ""
)

type ArmorClass struct {
	Base     int  `json:"armor_class"`
	DexBonus bool `json:"dex_bonus"`
	MaxBonus int  `json:"max_bonus"`
}

type Armor struct {
	Base          BasicEquipment `json:"base"`
	ArmorCategory ArmorCategory  `json:"armor_category"`
	ArmorClass    *ArmorClass    `json:"armor_class"`
	StrMin        int            `json:"str_minimum"`
	StealthDisadvantage bool           `json:"stealth_disadvantage"`
}

func (e *Armor) GetEquipmentType() EquipmentType {
	return EquipmentTypeArmor
}

func (e *Armor) GetName() string {
	return e.Base.Name
}

func (e *Armor) GetKey() string {
	return e.Base.Key
}

func (e *Armor) GetSlot() character.Slot {
	if e.ArmorCategory == ArmorCategoryShield {
		return character.SlotOffHand
	}

	return character.SlotBody
}
