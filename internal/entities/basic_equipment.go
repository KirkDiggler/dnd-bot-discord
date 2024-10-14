package entities

type BasicEquipment struct {
	Key    string  `json:"key"`
	Name   string  `json:"name"`
	Cost   *Cost   `json:"cost"`
	Weight float32 `json:"weight"`
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

func (e *BasicEquipment) GetSlot() Slot {
	return SlotNone
}
