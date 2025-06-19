package entities_test

import (
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/stretchr/testify/assert"
)

func TestCheckCombatEnd(t *testing.T) {
	t.Run("Players win when all monsters defeated", func(t *testing.T) {
		encounter := entities.NewEncounter("enc-1", "session-1", "channel-1", "Test Combat", "dm-1")
		
		// Add player
		player := &entities.Combatant{
			ID:        "player-1",
			Name:      "Hero",
			Type:      entities.CombatantTypePlayer,
			CurrentHP: 10,
			MaxHP:     10,
			IsActive:  true,
		}
		encounter.AddCombatant(player)
		
		// Add defeated monster
		monster := &entities.Combatant{
			ID:        "monster-1",
			Name:      "Goblin",
			Type:      entities.CombatantTypeMonster,
			CurrentHP: 0,
			MaxHP:     7,
			IsActive:  false,
		}
		encounter.AddCombatant(monster)
		
		shouldEnd, playersWon := encounter.CheckCombatEnd()
		assert.True(t, shouldEnd)
		assert.True(t, playersWon)
	})
	
	t.Run("Players lose when all players defeated", func(t *testing.T) {
		encounter := entities.NewEncounter("enc-2", "session-1", "channel-1", "Test Combat", "dm-1")
		
		// Add defeated player
		player := &entities.Combatant{
			ID:        "player-1",
			Name:      "Hero",
			Type:      entities.CombatantTypePlayer,
			CurrentHP: 0,
			MaxHP:     10,
			IsActive:  false,
		}
		encounter.AddCombatant(player)
		
		// Add active monster
		monster := &entities.Combatant{
			ID:        "monster-1",
			Name:      "Orc",
			Type:      entities.CombatantTypeMonster,
			CurrentHP: 15,
			MaxHP:     15,
			IsActive:  true,
		}
		encounter.AddCombatant(monster)
		
		shouldEnd, playersWon := encounter.CheckCombatEnd()
		assert.True(t, shouldEnd)
		assert.False(t, playersWon)
	})
	
	t.Run("Combat continues when both sides have active combatants", func(t *testing.T) {
		encounter := entities.NewEncounter("enc-3", "session-1", "channel-1", "Test Combat", "dm-1")
		
		// Add active player
		player := &entities.Combatant{
			ID:        "player-1",
			Name:      "Hero",
			Type:      entities.CombatantTypePlayer,
			CurrentHP: 5,
			MaxHP:     10,
			IsActive:  true,
		}
		encounter.AddCombatant(player)
		
		// Add active monster
		monster := &entities.Combatant{
			ID:        "monster-1",
			Name:      "Skeleton",
			Type:      entities.CombatantTypeMonster,
			CurrentHP: 8,
			MaxHP:     13,
			IsActive:  true,
		}
		encounter.AddCombatant(monster)
		
		shouldEnd, playersWon := encounter.CheckCombatEnd()
		assert.False(t, shouldEnd)
		assert.False(t, playersWon)
	})
}

func TestCombatLog(t *testing.T) {
	t.Run("Adds entries with round number", func(t *testing.T) {
		encounter := entities.NewEncounter("enc-1", "session-1", "channel-1", "Test Combat", "dm-1")
		encounter.Round = 2
		
		encounter.AddCombatLogEntry("Goblin hit Hero for 5 damage")
		
		assert.Len(t, encounter.CombatLog, 1)
		assert.Equal(t, "Round 2: Goblin hit Hero for 5 damage", encounter.CombatLog[0])
	})
	
	t.Run("Maintains maximum of 20 entries", func(t *testing.T) {
		encounter := entities.NewEncounter("enc-2", "session-1", "channel-1", "Test Combat", "dm-1")
		
		// Add 25 entries
		for i := 0; i < 25; i++ {
			encounter.AddCombatLogEntry("Test action")
		}
		
		assert.Len(t, encounter.CombatLog, 20)
		// First 5 entries should have been removed
		assert.Equal(t, "Round 0: Test action", encounter.CombatLog[0])
	})
}