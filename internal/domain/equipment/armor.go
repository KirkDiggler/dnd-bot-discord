package equipment

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
)

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
	Base                BasicEquipment `json:"base"`
	ArmorCategory       ArmorCategory  `json:"armor_category"`
	ArmorClass          *ArmorClass    `json:"armor_class"`
	StrMin              int            `json:"str_minimum"`
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

func (e *Armor) GetSlot() shared.Slot {
	if e.ArmorCategory == ArmorCategoryShield {
		return shared.SlotOffHand
	}

	return shared.SlotBody
}

// GetACBase returns the base AC for this armor
func (e *Armor) GetACBase() int {
	if e.ArmorClass != nil && e.ArmorClass.Base > 0 {
		return e.ArmorClass.Base
	}
	// Return -1 to indicate no AC data available
	return -1
}

// UsesDexBonus returns whether this armor uses DEX bonus
func (e *Armor) UsesDexBonus() bool {
	if e.ArmorClass != nil {
		return e.ArmorClass.DexBonus
	}
	// Default to true for no armor data
	return true
}

// GetMaxDexBonus returns the maximum DEX bonus (0 = unlimited)
func (e *Armor) GetMaxDexBonus() int {
	if e.ArmorClass != nil {
		return e.ArmorClass.MaxBonus
	}
	return 0
}
