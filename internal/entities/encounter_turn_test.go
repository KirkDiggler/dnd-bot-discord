package entities_test

import (
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/stretchr/testify/assert"
)

func TestNextTurn_SkipsInactiveCombatants(t *testing.T) {
	// Create encounter
	enc := entities.NewEncounter("test-id", "session-1", "channel-1", "Test Encounter", "dm-1")
	
	// Add combatants
	player1 := &entities.Combatant{
		ID:              "player-1",
		Name:            "Player 1",
		Type:            entities.CombatantTypePlayer,
		Initiative:      20,
		InitiativeBonus: 2,
		CurrentHP:       10,
		MaxHP:           10,
		AC:              15,
		IsActive:        true,
		PlayerID:        "user-1",
	}
	
	monster1 := &entities.Combatant{
		ID:              "monster-1",
		Name:            "Goblin 1",
		Type:            entities.CombatantTypeMonster,
		Initiative:      15,
		InitiativeBonus: 1,
		CurrentHP:       7,
		MaxHP:           7,
		AC:              13,
		IsActive:        true,
		Actions:         []*entities.MonsterAction{{Name: "Attack", AttackBonus: 3}},
	}
	
	monster2 := &entities.Combatant{
		ID:              "monster-2",
		Name:            "Goblin 2",
		Type:            entities.CombatantTypeMonster,
		Initiative:      10,
		InitiativeBonus: 1,
		CurrentHP:       0, // Dead
		MaxHP:           7,
		AC:              13,
		IsActive:        false, // Inactive
		Actions:         []*entities.MonsterAction{{Name: "Attack", AttackBonus: 3}},
	}
	
	player2 := &entities.Combatant{
		ID:              "player-2",
		Name:            "Player 2",
		Type:            entities.CombatantTypePlayer,
		Initiative:      5,
		InitiativeBonus: 0,
		CurrentHP:       12,
		MaxHP:           12,
		AC:              14,
		IsActive:        true,
		PlayerID:        "user-2",
	}
	
	enc.AddCombatant(player1)
	enc.AddCombatant(monster1)
	enc.AddCombatant(monster2)
	enc.AddCombatant(player2)
	
	// Set turn order based on initiative
	enc.TurnOrder = []string{"player-1", "monster-1", "monster-2", "player-2"}
	enc.Status = entities.EncounterStatusActive
	enc.Round = 1
	enc.Turn = 0
	
	// Test 1: First turn should be player 1
	current := enc.GetCurrentCombatant()
	assert.NotNil(t, current)
	assert.Equal(t, "player-1", current.ID)
	assert.False(t, enc.RoundPending)
	
	// Test 2: Advance to next turn - should be monster 1
	enc.NextTurn()
	current = enc.GetCurrentCombatant()
	assert.NotNil(t, current)
	assert.Equal(t, "monster-1", current.ID)
	assert.False(t, enc.RoundPending)
	
	// Test 3: Advance to next turn - should skip dead monster 2 and go to player 2
	enc.NextTurn()
	current = enc.GetCurrentCombatant()
	assert.NotNil(t, current)
	assert.Equal(t, "player-2", current.ID)
	assert.False(t, enc.RoundPending)
	
	// Test 4: Advance to next turn - should set round pending since all active have acted
	enc.NextTurn()
	assert.True(t, enc.RoundPending)
	current = enc.GetCurrentCombatant()
	assert.Nil(t, current) // No current combatant when round is pending
	
	// Test 5: Continue round - should go back to player 1
	success := enc.ContinueRound()
	assert.True(t, success)
	assert.False(t, enc.RoundPending)
	assert.Equal(t, 2, enc.Round)
	current = enc.GetCurrentCombatant()
	assert.NotNil(t, current)
	assert.Equal(t, "player-1", current.ID)
}

func TestNextTurn_AllInactiveCombatants(t *testing.T) {
	// Create encounter
	enc := entities.NewEncounter("test-id", "session-1", "channel-1", "Test Encounter", "dm-1")
	
	// Add all dead combatants
	monster1 := &entities.Combatant{
		ID:       "monster-1",
		Name:     "Dead Goblin 1",
		Type:     entities.CombatantTypeMonster,
		CurrentHP: 0,
		MaxHP:    7,
		IsActive: false,
	}
	
	monster2 := &entities.Combatant{
		ID:       "monster-2",
		Name:     "Dead Goblin 2",
		Type:     entities.CombatantTypeMonster,
		CurrentHP: 0,
		MaxHP:    7,
		IsActive: false,
	}
	
	enc.AddCombatant(monster1)
	enc.AddCombatant(monster2)
	enc.TurnOrder = []string{"monster-1", "monster-2"}
	enc.Status = entities.EncounterStatusActive
	enc.Round = 1
	enc.Turn = 0
	
	// Should immediately set round pending since no active combatants
	enc.NextTurn()
	assert.True(t, enc.RoundPending)
}