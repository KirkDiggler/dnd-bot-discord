package equipment

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
)

type BasicEquipment struct {
	Key    string       `json:"key"`
	Name   string       `json:"name"`
	Cost   *shared.Cost `json:"cost"`
	Weight float32      `json:"weight"`
}

func (e *BasicEquipment) GetEquipmentType() EquipmentType {
	return "BasicEquipment"
}

func (e *BasicEquipment) GetName() string {
	return e.Name
}

func (e *BasicEquipment) GetKey() string {
	return e.Key
}

func (e *BasicEquipment) GetSlot() shared.Slot {
	return shared.SlotNone
}
