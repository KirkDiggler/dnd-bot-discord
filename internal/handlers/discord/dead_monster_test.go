package discord

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/combat"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestDeadMonsterShouldNotAct verifies that monsters with 0 HP don't take turns
func TestDeadMonsterShouldNotAct(t *testing.T) {
	tests := []struct {
		name      string
		combatant *combat.Combatant
		shouldAct bool
		reason    string
	}{
		{
			name: "Alive monster should act",
			combatant: &combat.Combatant{
				Name:      "Goblin",
				Type:      combat.CombatantTypeMonster,
				CurrentHP: 7,
				MaxHP:     7,
				Actions: []*combat.MonsterAction{
					{Name: "Scimitar", AttackBonus: 4},
				},
			},
			shouldAct: true,
			reason:    "Monster with HP > 0 and actions should be able to act",
		},
		{
			name: "Dead monster should not act",
			combatant: &combat.Combatant{
				Name:      "Skeleton",
				Type:      combat.CombatantTypeMonster,
				CurrentHP: 0,
				MaxHP:     13,
				Actions: []*combat.MonsterAction{
					{Name: "Shortsword", AttackBonus: 4},
				},
			},
			shouldAct: false,
			reason:    "Monster with 0 HP should not be able to act",
		},
		{
			name: "Monster with negative HP should not act",
			combatant: &combat.Combatant{
				Name:      "Zombie",
				Type:      combat.CombatantTypeMonster,
				CurrentHP: -5,
				MaxHP:     22,
				Actions: []*combat.MonsterAction{
					{Name: "Slam", AttackBonus: 3},
				},
			},
			shouldAct: false,
			reason:    "Monster with negative HP should not be able to act",
		},
		{
			name: "Monster at 1 HP should still act",
			combatant: &combat.Combatant{
				Name:      "Orc",
				Type:      combat.CombatantTypeMonster,
				CurrentHP: 1,
				MaxHP:     15,
				Actions: []*combat.MonsterAction{
					{Name: "Greataxe", AttackBonus: 5},
				},
			},
			shouldAct: true,
			reason:    "Monster with 1 HP is still alive and should act",
		},
		{
			name: "Player combatant check (should not affect logic)",
			combatant: &combat.Combatant{
				Name:      "Gandalf",
				Type:      combat.CombatantTypePlayer,
				CurrentHP: 0,
				MaxHP:     50,
			},
			shouldAct: false,
			reason:    "This test is for monster logic, player combatants handled separately",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Check the condition we use in the handler
			canAct := false
			if tt.combatant != nil && tt.combatant.Type == combat.CombatantTypeMonster {
				canAct = tt.combatant.CanAct()
			}

			assert.Equal(t, tt.shouldAct, canAct, tt.reason)
		})
	}
}

// TestDeadMonsterInEncounter tests that dead monsters are properly handled in encounter flow
func TestDeadMonsterInEncounter(t *testing.T) {
	// Create a test encounter
	encounter := &combat.Encounter{
		ID:     "test-encounter",
		Status: combat.EncounterStatusActive,
		Round:  1,
		Turn:   0,
		Combatants: map[string]*combat.Combatant{
			"skeleton-1": {
				ID:        "skeleton-1",
				Name:      "Skeleton",
				Type:      combat.CombatantTypeMonster,
				CurrentHP: 0, // Dead
				MaxHP:     13,
				IsActive:  true,
				Actions: []*combat.MonsterAction{
					{Name: "Shortsword", AttackBonus: 4},
				},
			},
			"player-1": {
				ID:        "player-1",
				Name:      "Hero",
				Type:      combat.CombatantTypePlayer,
				CurrentHP: 25,
				MaxHP:     30,
				IsActive:  true,
			},
		},
		TurnOrder: []string{"skeleton-1", "player-1"},
	}

	// Get current combatant
	current := encounter.GetCurrentCombatant()
	assert.NotNil(t, current)
	assert.Equal(t, "skeleton-1", current.ID)

	// Verify dead monster check
	shouldProcessTurn := current.Type == combat.CombatantTypeMonster && current.CanAct()

	assert.False(t, shouldProcessTurn, "Dead skeleton should not process turn")

	// Advance turn
	encounter.NextTurn()

	// Now should be player's turn
	current = encounter.GetCurrentCombatant()
	assert.NotNil(t, current)
	assert.Equal(t, "player-1", current.ID)
	assert.Equal(t, combat.CombatantTypePlayer, current.Type)
}
