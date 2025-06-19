package encounter_test

import (
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/stretchr/testify/assert"
)

func TestDungeonCombatIntegration(t *testing.T) {
	// This test simulates the full flow of:
	// 1. Player joins dungeon with character selected
	// 2. Player enters combat room
	// 3. Player is added to encounter as player combatant
	// 4. Player does NOT appear in enemy list

	t.Run("Player with character enters dungeon combat", func(t *testing.T) {
		// Setup
		playerID := "player-123"
		characterID := "char-123"
		sessionID := "dungeon-session-123"

		// Create a dungeon session with player who has character selected
		_ = &entities.Session{
			ID:        sessionID,
			Name:      "Test Dungeon",
			CreatorID: "bot-123",
			DMID:      "bot-123",
			Status:    entities.SessionStatusActive,
			Members: map[string]*entities.SessionMember{
				playerID: {
					UserID:      playerID,
					Role:        entities.SessionRolePlayer,
					CharacterID: characterID, // Character is selected!
				},
				"bot-123": {
					UserID: "bot-123",
					Role:   entities.SessionRoleDM,
				},
			},
			Metadata: map[string]interface{}{
				"sessionType": "dungeon",
			},
		}

		// Create encounter for combat room
		encounter := entities.NewEncounter("enc-123", sessionID, "channel-123", "Guard Chamber", "bot-123")

		// Simulate adding monsters (what happens in enter_room.go)
		goblin := &entities.Combatant{
			ID:        "goblin-1",
			Name:      "Goblin",
			Type:      entities.CombatantTypeMonster,
			CurrentHP: 7,
			MaxHP:     7,
			AC:        15,
			IsActive:  true,
		}
		orc := &entities.Combatant{
			ID:        "orc-1",
			Name:      "Orc",
			Type:      entities.CombatantTypeMonster,
			CurrentHP: 15,
			MaxHP:     15,
			AC:        13,
			IsActive:  true,
		}
		encounter.AddCombatant(goblin)
		encounter.AddCombatant(orc)

		// Now simulate adding player (what SHOULD happen when character is selected)
		playerCombatant := &entities.Combatant{
			ID:              "player-comb-1",
			Name:            "Aragorn", // Player's character name
			Type:            entities.CombatantTypePlayer,
			PlayerID:        playerID,
			CharacterID:     characterID,
			CurrentHP:       50,
			MaxHP:           50,
			AC:              17,
			InitiativeBonus: 2,
			IsActive:        true,
		}
		encounter.AddCombatant(playerCombatant)

		// Verify encounter state
		assert.Len(t, encounter.Combatants, 3, "Should have 3 combatants total")

		// Verify we can filter monsters for enemy list
		var enemies []*entities.Combatant
		var players []*entities.Combatant

		for _, c := range encounter.Combatants {
			if c.Type == entities.CombatantTypeMonster {
				enemies = append(enemies, c)
			} else if c.Type == entities.CombatantTypePlayer {
				players = append(players, c)
			}
		}

		assert.Len(t, enemies, 2, "Should have 2 enemies")
		assert.Len(t, players, 1, "Should have 1 player")

		// Verify enemy list contains only monsters
		for _, enemy := range enemies {
			assert.Equal(t, entities.CombatantTypeMonster, enemy.Type)
			assert.Empty(t, enemy.PlayerID, "Monsters should not have PlayerID")
		}

		// Verify player is correctly identified
		assert.Equal(t, "Aragorn", players[0].Name)
		assert.Equal(t, playerID, players[0].PlayerID)
		assert.Equal(t, characterID, players[0].CharacterID)
	})

	t.Run("Player without character cannot participate in combat", func(t *testing.T) {
		playerID := "player-no-char"
		sessionID := "dungeon-session-456"

		// Session where player has NOT selected a character
		_ = &entities.Session{
			ID: sessionID,
			Members: map[string]*entities.SessionMember{
				playerID: {
					UserID:      playerID,
					Role:        entities.SessionRolePlayer,
					CharacterID: "", // No character selected!
				},
			},
		}

		// Create encounter
		encounter := entities.NewEncounter("enc-456", sessionID, "channel-456", "Dark Crypt", "bot-456")

		// Add monsters
		skeleton := &entities.Combatant{
			ID:   "skeleton-1",
			Name: "Skeleton",
			Type: entities.CombatantTypeMonster,
		}
		encounter.AddCombatant(skeleton)

		// Player should NOT be added to encounter since they have no character
		// In real code, this is checked in enter_room.go: if member.CharacterID != ""

		// Verify only monster in encounter
		assert.Len(t, encounter.Combatants, 1)
		assert.Equal(t, entities.CombatantTypeMonster, encounter.Combatants["skeleton-1"].Type)
	})
}

func TestPlayerMonsterNameCollision(t *testing.T) {
	// Specific test for the "Orc" issue - when a player might be confused with a monster

	encounter := entities.NewEncounter("enc-789", "session-789", "channel-789", "Mixed Battle", "dm-789")

	// Add Orc monster
	orcMonster := &entities.Combatant{
		ID:        "monster-orc",
		Name:      "Orc",
		Type:      entities.CombatantTypeMonster,
		CurrentHP: 15,
		MaxHP:     15,
		AC:        13,
		IsActive:  true,
	}
	encounter.AddCombatant(orcMonster)

	// Add player with unfortunate name "Orc"
	playerOrc := &entities.Combatant{
		ID:          "player-orc",
		Name:        "Orc", // Player named their character "Orc"!
		Type:        entities.CombatantTypePlayer,
		PlayerID:    "player-999",
		CharacterID: "char-orc",
		CurrentHP:   30,
		MaxHP:       30,
		AC:          14,
		IsActive:    true,
	}
	encounter.AddCombatant(playerOrc)

	// Test that we can distinguish them
	t.Run("Can identify player vs monster by Type", func(t *testing.T) {
		for _, combatant := range encounter.Combatants {
			if combatant.Type == entities.CombatantTypePlayer {
				assert.NotEmpty(t, combatant.PlayerID, "Players must have PlayerID")
				assert.NotEmpty(t, combatant.CharacterID, "Players must have CharacterID")
			} else if combatant.Type == entities.CombatantTypeMonster {
				assert.Empty(t, combatant.PlayerID, "Monsters must not have PlayerID")
				assert.Empty(t, combatant.CharacterID, "Monsters must not have CharacterID")
			}
		}
	})

	t.Run("Enemy list excludes player Orc", func(t *testing.T) {
		var enemyNames []string
		for _, c := range encounter.Combatants {
			if c.Type == entities.CombatantTypeMonster {
				enemyNames = append(enemyNames, c.Name)
			}
		}

		assert.Len(t, enemyNames, 1, "Should only have 1 enemy")
		assert.Contains(t, enemyNames, "Orc", "Monster Orc should be in enemy list")

		// But there are 2 total combatants named "Orc"
		orcCount := 0
		for _, c := range encounter.Combatants {
			if c.Name == "Orc" {
				orcCount++
			}
		}
		assert.Equal(t, 2, orcCount, "Should have 2 combatants named Orc")
	})
}
