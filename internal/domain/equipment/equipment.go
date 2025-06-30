package equipment

import "github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"

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
	GetSlot() character.Slot
}
