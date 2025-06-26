package entities_test

import (
	"testing"
	"time"

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

func TestEncounter_NextTurn_SkipsDeadCombatants(t *testing.T) {
	// Create an encounter with some combatants
	enc := &entities.Encounter{
		ID:         "test-encounter",
		Status:     entities.EncounterStatusActive,
		Round:      1,
		Turn:       0,
		Combatants: make(map[string]*entities.Combatant),
		TurnOrder:  []string{"c1", "c2", "c3", "c4"},
		StartedAt:  &time.Time{},
	}

	// Add combatants - c2 and c3 are dead
	enc.Combatants["c1"] = &entities.Combatant{
		ID:        "c1",
		Name:      "Fighter",
		IsActive:  true,
		CurrentHP: 10,
		MaxHP:     10,
	}
	enc.Combatants["c2"] = &entities.Combatant{
		ID:        "c2",
		Name:      "Dead Goblin",
		IsActive:  true,
		CurrentHP: 0, // Dead
		MaxHP:     5,
	}
	enc.Combatants["c3"] = &entities.Combatant{
		ID:        "c3",
		Name:      "Inactive Orc",
		IsActive:  false, // Inactive
		CurrentHP: 10,
		MaxHP:     10,
	}
	enc.Combatants["c4"] = &entities.Combatant{
		ID:        "c4",
		Name:      "Wizard",
		IsActive:  true,
		CurrentHP: 8,
		MaxHP:     8,
	}

	// Start at turn 0 (Fighter)
	current := enc.GetCurrentCombatant()
	assert.Equal(t, "c1", current.ID)

	// Advance turn - should skip dead goblin and inactive orc, land on wizard
	enc.NextTurn()
	current = enc.GetCurrentCombatant()
	assert.Equal(t, "c4", current.ID)

	// Advance turn again - should go to next round and back to fighter
	enc.NextTurn()
	assert.Equal(t, 2, enc.Round)
	current = enc.GetCurrentCombatant()
	assert.Equal(t, "c1", current.ID)
}

func TestEncounter_NextRound_SkipsDeadCombatants(t *testing.T) {
	// Create an encounter where the first combatant is dead
	enc := &entities.Encounter{
		ID:         "test-encounter",
		Status:     entities.EncounterStatusActive,
		Round:      1,
		Turn:       2,
		Combatants: make(map[string]*entities.Combatant),
		TurnOrder:  []string{"c1", "c2", "c3"},
		StartedAt:  &time.Time{},
	}

	// c1 is dead, c2 is alive, c3 is alive
	enc.Combatants["c1"] = &entities.Combatant{
		ID:        "c1",
		Name:      "Dead Fighter",
		IsActive:  true,
		CurrentHP: 0, // Dead
		MaxHP:     10,
	}
	enc.Combatants["c2"] = &entities.Combatant{
		ID:        "c2",
		Name:      "Cleric",
		IsActive:  true,
		CurrentHP: 12,
		MaxHP:     12,
	}
	enc.Combatants["c3"] = &entities.Combatant{
		ID:        "c3",
		Name:      "Rogue",
		IsActive:  true,
		CurrentHP: 8,
		MaxHP:     8,
	}

	// Start new round
	enc.NextRound()

	// Should be round 2, turn should skip dead fighter and be on cleric
	assert.Equal(t, 2, enc.Round)
	assert.Equal(t, 1, enc.Turn)
	current := enc.GetCurrentCombatant()
	assert.Equal(t, "c2", current.ID)
}

func TestEncounter_Start_SkipsDeadCombatants(t *testing.T) {
	// Create an encounter in rolling status
	enc := &entities.Encounter{
		ID:         "test-encounter",
		Status:     entities.EncounterStatusRolling,
		Combatants: make(map[string]*entities.Combatant),
		TurnOrder:  []string{"c1", "c2"},
	}

	// First combatant is dead
	enc.Combatants["c1"] = &entities.Combatant{
		ID:        "c1",
		Name:      "Dead Zombie",
		IsActive:  true,
		CurrentHP: 0, // Dead
		MaxHP:     10,
	}
	enc.Combatants["c2"] = &entities.Combatant{
		ID:        "c2",
		Name:      "Living Paladin",
		IsActive:  true,
		CurrentHP: 15,
		MaxHP:     15,
	}

	// Start the encounter
	assert.True(t, enc.Start())

	// Should skip dead zombie and start with paladin
	assert.Equal(t, 1, enc.Turn)
	current := enc.GetCurrentCombatant()
	assert.Equal(t, "c2", current.ID)
}

func TestEncounter_AllCombatantsDead(t *testing.T) {
	// Edge case: all combatants are dead
	enc := &entities.Encounter{
		ID:         "test-encounter",
		Status:     entities.EncounterStatusActive,
		Round:      1,
		Turn:       0,
		Combatants: make(map[string]*entities.Combatant),
		TurnOrder:  []string{"c1", "c2"},
		StartedAt:  &time.Time{},
	}

	// Both combatants are dead
	enc.Combatants["c1"] = &entities.Combatant{
		ID:        "c1",
		Name:      "Dead Fighter",
		IsActive:  true,
		CurrentHP: 0,
		MaxHP:     10,
	}
	enc.Combatants["c2"] = &entities.Combatant{
		ID:        "c2",
		Name:      "Dead Goblin",
		IsActive:  true,
		CurrentHP: 0,
		MaxHP:     5,
	}

	// Advance turn - should handle gracefully
	enc.NextTurn()

	// Turn should advance past all dead combatants
	assert.GreaterOrEqual(t, enc.Turn, len(enc.TurnOrder))

	// Should have advanced to next round
	assert.Equal(t, 2, enc.Round)
}
