package exploration

import (
	"time"
)

// DungeonState represents the current state of a dungeon
type DungeonState string

const (
	DungeonStateAwaitingParty DungeonState = "awaiting_party"
	DungeonStateRoomReady     DungeonState = "room_ready"
	DungeonStateInProgress    DungeonState = "in_progress" // combat/puzzle/trap active
	DungeonStateRoomCleared   DungeonState = "room_cleared"
	DungeonStateComplete      DungeonState = "complete"
	DungeonStateFailed        DungeonState = "failed"
)

// PartyMember represents a player in the dungeon party
type PartyMember struct {
	UserID      string `json:"user_id"`
	CharacterID string `json:"character_id"`
	Status      string `json:"status"` // "active", "unconscious", "dead"
}

// RoomType represents different types of dungeon rooms
type RoomType string

const (
	RoomTypeCombat   RoomType = "combat"
	RoomTypePuzzle   RoomType = "puzzle"
	RoomTypeTreasure RoomType = "treasure"
	RoomTypeTrap     RoomType = "trap"
	RoomTypeRest     RoomType = "rest"
)

// DungeonRoom represents a room in the dungeon
type DungeonRoom struct {
	Type        RoomType `json:"type"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Completed   bool     `json:"completed"`
	Monsters    []string `json:"monsters,omitempty"`
	Treasure    []string `json:"treasure,omitempty"`
	Challenge   string   `json:"challenge"`
}

// Dungeon represents a dungeon instance
type Dungeon struct {
	ID            string        `json:"id"`
	SessionID     string        `json:"session_id"`
	State         DungeonState  `json:"state"`
	CurrentRoom   *DungeonRoom  `json:"current_room"`
	RoomNumber    int           `json:"room_number"`
	Difficulty    string        `json:"difficulty"`
	Party         []PartyMember `json:"party"`
	RoomsCleared  int           `json:"rooms_cleared"`
	LootCollected []string      `json:"loot_collected"`
	CreatedAt     time.Time     `json:"created_at"`
	CompletedAt   *time.Time    `json:"completed_at,omitempty"`
}

// CanEnterRoom checks if the party can enter the current room
func (d *Dungeon) CanEnterRoom() bool {
	// Players should always be able to enter if there's anyone in the party
	return len(d.Party) > 0 && d.IsActive()
}

// CanProceed checks if the party can move to the next room
func (d *Dungeon) CanProceed() bool {
	return d.State == DungeonStateRoomCleared
}

// IsActive checks if the dungeon is still active
func (d *Dungeon) IsActive() bool {
	return d.State != DungeonStateComplete && d.State != DungeonStateFailed
}

// GetActivePartyMembers returns all active party members
func (d *Dungeon) GetActivePartyMembers() []PartyMember {
	var active []PartyMember
	for _, member := range d.Party {
		if member.Status == "active" {
			active = append(active, member)
		}
	}
	return active
}
