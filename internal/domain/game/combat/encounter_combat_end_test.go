package combat_test

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/combat"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckCombatEnd_PlayersWin(t *testing.T) {
	enc := &combat.Encounter{
		Combatants: map[string]*combat.Combatant{
			"player1": {
				ID:       "player1",
				Type:     combat.CombatantTypePlayer,
				IsActive: true,
			},
			"monster1": {
				ID:       "monster1",
				Type:     combat.CombatantTypeMonster,
				IsActive: false, // Defeated
			},
			"monster2": {
				ID:       "monster2",
				Type:     combat.CombatantTypeMonster,
				IsActive: false, // Defeated
			},
		},
	}

	shouldEnd, playersWon := enc.CheckCombatEnd()
	assert.True(t, shouldEnd)
	assert.True(t, playersWon)
}

func TestCheckCombatEnd_PlayersLose(t *testing.T) {
	enc := &combat.Encounter{
		Combatants: map[string]*combat.Combatant{
			"player1": {
				ID:       "player1",
				Type:     combat.CombatantTypePlayer,
				IsActive: false, // Defeated
			},
			"player2": {
				ID:       "player2",
				Type:     combat.CombatantTypePlayer,
				IsActive: false, // Defeated
			},
			"monster1": {
				ID:       "monster1",
				Type:     combat.CombatantTypeMonster,
				IsActive: true,
			},
		},
	}

	shouldEnd, playersWon := enc.CheckCombatEnd()
	assert.True(t, shouldEnd)
	assert.False(t, playersWon)
}

func TestCheckCombatEnd_CombatContinues(t *testing.T) {
	enc := &combat.Encounter{
		Combatants: map[string]*combat.Combatant{
			"player1": {
				ID:       "player1",
				Type:     combat.CombatantTypePlayer,
				IsActive: true,
			},
			"monster1": {
				ID:       "monster1",
				Type:     combat.CombatantTypeMonster,
				IsActive: true,
			},
		},
	}

	shouldEnd, playersWon := enc.CheckCombatEnd()
	assert.False(t, shouldEnd)
	assert.False(t, playersWon)
}

func TestCheckCombatEnd_MonsterOnlyBattle(t *testing.T) {
	enc := &combat.Encounter{
		Combatants: map[string]*combat.Combatant{
			"monster1": {
				ID:       "monster1",
				Type:     combat.CombatantTypeMonster,
				IsActive: false, // Defeated
			},
			"monster2": {
				ID:       "monster2",
				Type:     combat.CombatantTypeMonster,
				IsActive: true,
			},
		},
	}

	// In monster-only battles with some active and some defeated,
	// CheckCombatEnd returns true, false (as if players lost)
	shouldEnd, playersWon := enc.CheckCombatEnd()
	assert.True(t, shouldEnd)
	assert.False(t, playersWon)
}

func TestCheckCombatEnd_AllDefeated(t *testing.T) {
	enc := &combat.Encounter{
		Combatants: map[string]*combat.Combatant{
			"player1": {
				ID:       "player1",
				Type:     combat.CombatantTypePlayer,
				IsActive: false,
			},
			"monster1": {
				ID:       "monster1",
				Type:     combat.CombatantTypeMonster,
				IsActive: false,
			},
		},
	}

	// When everyone is defeated, combat doesn't end (edge case)
	shouldEnd, playersWon := enc.CheckCombatEnd()
	assert.False(t, shouldEnd)
	assert.False(t, playersWon)
}

func TestApplyDamage_Defeat(t *testing.T) {
	combatant := &combat.Combatant{
		ID:        "test",
		CurrentHP: 10,
		MaxHP:     20,
		IsActive:  true,
	}

	// Apply non-lethal damage
	combatant.ApplyDamage(5)
	assert.Equal(t, 5, combatant.CurrentHP)
	assert.True(t, combatant.IsActive)

	// Apply lethal damage
	combatant.ApplyDamage(10)
	assert.Equal(t, 0, combatant.CurrentHP)
	assert.False(t, combatant.IsActive)
}

func TestApplyDamage_TempHP(t *testing.T) {
	combatant := &combat.Combatant{
		ID:        "test",
		CurrentHP: 10,
		MaxHP:     20,
		TempHP:    5,
		IsActive:  true,
	}

	// Damage absorbed by temp HP
	combatant.ApplyDamage(3)
	assert.Equal(t, 10, combatant.CurrentHP)
	assert.Equal(t, 2, combatant.TempHP)

	// Damage exceeds temp HP
	combatant.ApplyDamage(7)
	assert.Equal(t, 5, combatant.CurrentHP)
	assert.Equal(t, 0, combatant.TempHP)
}
