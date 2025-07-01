package exploration

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/combat"
)

type RoomStatus string

const (
	RoomStatusUnset    RoomStatus = ""
	RoomStatusActive   RoomStatus = "active"
	RoomStatusInactive RoomStatus = "inactive"
)

type Room struct {
	ID        string
	Status    RoomStatus
	Character *character.Character
	Monster   *combat.Monster
}
