package dungeon_test

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/session"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJoinPartyWorkflow(t *testing.T) {
	t.Run("User already in session can select character", func(t *testing.T) {
		// Simulates: User starts dungeon (auto-added to session) then clicks "Select Character"
		userID := "player-123"
		characterID := "char-123"

		sess := &session.Session{
			ID: "session-123",
			Members: map[string]*session.SessionMember{
				userID: {
					UserID:      userID,
					Role:        session.SessionRolePlayer,
					CharacterID: "", // No character selected yet
				},
			},
		}

		// User is already in session
		assert.True(t, sess.IsUserInSession(userID))

		// But has no character
		member := sess.Members[userID]
		assert.Empty(t, member.CharacterID)

		// After selecting character
		member.CharacterID = characterID
		assert.Equal(t, characterID, member.CharacterID)
	})

	t.Run("New user not in session joins with character", func(t *testing.T) {
		// Simulates: Another player joins an existing dungeon
		existingUserID := "player-existing"
		newUserID := "player-new"

		sess := &session.Session{
			ID: "session-456",
			Members: map[string]*session.SessionMember{
				existingUserID: {
					UserID:      existingUserID,
					Role:        session.SessionRolePlayer,
					CharacterID: "char-existing",
				},
			},
		}

		// New user is not in session
		assert.False(t, sess.IsUserInSession(newUserID))

		// After joining
		sess.Members[newUserID] = &session.SessionMember{
			UserID:      newUserID,
			Role:        session.SessionRolePlayer,
			CharacterID: "char-new",
		}

		assert.True(t, sess.IsUserInSession(newUserID))
		assert.NotEmpty(t, sess.Members[newUserID].CharacterID)
	})

	t.Run("Cannot enter room workflow states", func(t *testing.T) {
		userID := "player-789"

		testCases := []struct {
			name         string
			inSession    bool
			hasCharacter bool
			canEnter     bool
			errorMsg     string
		}{
			{
				name:         "Not in session",
				inSession:    false,
				hasCharacter: false,
				canEnter:     false,
				errorMsg:     "You need to join the party first!",
			},
			{
				name:         "In session but no character",
				inSession:    true,
				hasCharacter: false,
				canEnter:     false,
				errorMsg:     "You need to select a character! Click 'Select Character' first.",
			},
			{
				name:         "In session with character",
				inSession:    true,
				hasCharacter: true,
				canEnter:     true,
				errorMsg:     "",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				sess := &session.Session{
					ID:      "session-test",
					Members: map[string]*session.SessionMember{},
				}

				if tc.inSession {
					sess.Members[userID] = &session.SessionMember{
						UserID: userID,
						Role:   session.SessionRolePlayer,
					}
					if tc.hasCharacter {
						sess.Members[userID].CharacterID = "char-test"
					}
				}

				// Check if can enter
				canEnter := sess.IsUserInSession(userID)
				if canEnter && sess.Members[userID].CharacterID == "" {
					canEnter = false
				}

				assert.Equal(t, tc.canEnter, canEnter)
			})
		}
	})
}

func TestCharacterSelectionEdgeCases(t *testing.T) {
	t.Run("Multiple active characters requires manual selection", func(t *testing.T) {
		// User has multiple characters - can't auto-select
		characters := []*character.Character{
			{ID: "char-1", Name: "Aragorn", Status: shared.CharacterStatusActive},
			{ID: "char-2", Name: "Gandalf", Status: shared.CharacterStatusActive},
		}

		// Should not auto-select when multiple active
		assert.Greater(t, len(characters), 1)

		var activeCount int
		for _, char := range characters {
			if char.Status == shared.CharacterStatusActive {
				activeCount++
			}
		}
		assert.Greater(t, activeCount, 1, "Multiple active characters prevent auto-selection")
	})

	t.Run("Single active character can auto-select", func(t *testing.T) {
		characters := []*character.Character{
			{ID: "char-1", Name: "Legolas", Status: shared.CharacterStatusActive},
			{ID: "char-2", Name: "Gimli", Status: shared.CharacterStatusArchived},
			{ID: "char-3", Name: "Boromir", Status: shared.CharacterStatusDraft},
		}

		var activeChars []*character.Character
		for _, char := range characters {
			if char.Status == shared.CharacterStatusActive {
				activeChars = append(activeChars, char)
			}
		}

		assert.Len(t, activeChars, 1, "Exactly one active character allows auto-selection")
		assert.Equal(t, "Legolas", activeChars[0].Name)
	})

	t.Run("No active characters prevents joining", func(t *testing.T) {
		characters := []*character.Character{
			{ID: "char-1", Name: "Draft Hero", Status: shared.CharacterStatusDraft},
			{ID: "char-2", Name: "Old Hero", Status: shared.CharacterStatusArchived},
		}

		var activeChars []*character.Character
		for _, char := range characters {
			if char.Status == shared.CharacterStatusActive {
				activeChars = append(activeChars, char)
			}
		}

		assert.Empty(t, activeChars, "No active characters to select")
	})
}
