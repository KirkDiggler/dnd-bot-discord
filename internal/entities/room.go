package entities

type RoomStatus string

const (
	RoomStatusUnset    RoomStatus = ""
	RoomStatusActive   RoomStatus = "active"
	RoomStatusInactive RoomStatus = "inactive"
)

type Room struct {
	ID        string
	Status    RoomStatus
	Character *Character
	Monster   *Monster
}
