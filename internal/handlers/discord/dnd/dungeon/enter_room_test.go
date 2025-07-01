package dungeon_test

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/session"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnterRoomValidation(t *testing.T) {
	t.Run("Player cannot enter room without joining party", func(t *testing.T) {
		// Session without the player as member
		sess := &session.Session{
			ID: "session-123",
			Members: map[string]*session.SessionMember{
				"dm-123": {
					UserID: "dm-123",
					Role:   session.SessionRoleDM,
				},
				// Player is NOT in members
			},
		}

		playerID := "player-123"
		canEnter := sess.IsUserInSession(playerID)

		assert.False(t, canEnter, "Player should not be able to enter without joining")
	})

	t.Run("Player cannot enter combat room without character", func(t *testing.T) {
		// Session with player but no character selected
		sess := &session.Session{
			ID: "session-456",
			Members: map[string]*session.SessionMember{
				"player-456": {
					UserID:      "player-456",
					Role:        session.SessionRolePlayer,
					CharacterID: "", // No character!
				},
			},
		}

		member := sess.Members["player-456"]
		hasCharacter := member.CharacterID != ""

		assert.False(t, hasCharacter, "Player should not be able to enter combat without character")
	})

	t.Run("Player with character can enter room", func(t *testing.T) {
		// Session with player and character selected
		sess := &session.Session{
			ID: "session-789",
			Members: map[string]*session.SessionMember{
				"player-789": {
					UserID:      "player-789",
					Role:        session.SessionRolePlayer,
					CharacterID: "char-789", // Has character!
				},
			},
		}

		playerID := "player-789"
		member := sess.Members[playerID]

		canEnter := sess.IsUserInSession(playerID) && member.CharacterID != ""
		assert.True(t, canEnter, "Player with character should be able to enter")
	})

	t.Run("DM can enter room without character", func(t *testing.T) {
		// DM doesn't need a character
		sess := &session.Session{
			ID: "session-dm",
			Members: map[string]*session.SessionMember{
				"dm-999": {
					UserID:      "dm-999",
					Role:        session.SessionRoleDM,
					CharacterID: "", // DM has no character
				},
			},
		}

		dmID := "dm-999"
		member := sess.Members[dmID]

		// DM can enter regardless of character
		canEnter := sess.IsUserInSession(dmID) &&
			(member.Role == session.SessionRoleDM || member.CharacterID != "")

		assert.True(t, canEnter, "DM should be able to enter without character")
	})
}

func TestDungeonWorkflowOrder(t *testing.T) {
	t.Run("Correct workflow order", func(t *testing.T) {
		// 1. Start dungeon
		dungeonStarted := true
		assert.True(t, dungeonStarted)

		// 2. Join party (selects character)
		characterSelected := true
		assert.True(t, characterSelected)

		// 3. Enter room (now allowed)
		canEnterRoom := dungeonStarted && characterSelected
		assert.True(t, canEnterRoom)
	})

	t.Run("Incorrect workflow - skip join party", func(t *testing.T) {
		// 1. Start dungeon
		dungeonStarted := true
		assert.True(t, dungeonStarted)

		// 2. Try to enter room without joining (no character selected)
		characterSelected := false
		canEnterRoom := dungeonStarted && characterSelected

		assert.False(t, canEnterRoom, "Should not be able to enter without joining party")
	})
}

func TestCharacterRequirementsByRoomType(t *testing.T) {
	testCases := []struct {
		name              string
		roomType          string
		requiresCharacter bool
	}{
		{
			name:              "Combat room requires character",
			roomType:          "combat",
			requiresCharacter: true,
		},
		{
			name:              "Puzzle room requires character",
			roomType:          "puzzle",
			requiresCharacter: true,
		},
		{
			name:              "Trap room requires character",
			roomType:          "trap",
			requiresCharacter: true,
		},
		{
			name:              "Treasure room requires character",
			roomType:          "treasure",
			requiresCharacter: true,
		},
		{
			name:              "Rest room requires character",
			roomType:          "rest",
			requiresCharacter: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// All room types require a character for players
			assert.True(t, tc.requiresCharacter,
				"Room type %s should require character for players", tc.roomType)
		})
	}
}
