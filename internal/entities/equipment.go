package entities

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
	GetSlot() Slot
}
