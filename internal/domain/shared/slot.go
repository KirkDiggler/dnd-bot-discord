package shared

type Slot string

const (
	SlotMainHand  Slot = "main-hand"
	SlotOffHand   Slot = "off-hand"
	SlotTwoHanded Slot = "two-handed"
	SlotBody      Slot = "body"
	SlotNone      Slot = "none"
)
