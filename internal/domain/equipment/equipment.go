package equipment

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
)

type EquipmentType string

const (
	EquipmentTypeArmor   EquipmentType = "armor"
	EquipmentTypeWeapon  EquipmentType = "weapon"
	EquipmentTypeOther   EquipmentType = "other"
	EquipmentTypeUnknown EquipmentType = ""
)

type Equipment interface {
	GetEquipmentType() EquipmentType
	GetName() string
	GetKey() string
	GetSlot() shared.Slot
}
