package attack

import (
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/dice"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities/damage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRollAttack_AlwaysAddsModifiers(t *testing.T) {
	t.Skip("Skipping until dice roller interface is available for mocking")

	// Test that attack modifiers are always added, even on natural 1
	weaponDamage := &damage.Damage{
		DiceCount:  1,
		DiceSize:   8,
		DamageType: damage.TypeSlashing,
	}

	attackBonus := 6 // +4 DEX + 2 proficiency
	damageBonus := 4 // +4 DEX

	// TODO: Once dice roller interface is available:
	// 1. Test natural 1 - should show as "1+6=7" not "1=1"
	// 2. Test natural 20 - should show crit with bonus
	// 3. Test normal rolls - should always include bonus

	// For now, just run a basic test to ensure the function exists
	// Need to provide a dice roller
	roller := dice.NewRandomRoller()
	result, err := RollAttack(roller, attackBonus, damageBonus, weaponDamage)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Greater(t, result.AttackRoll, 0)
}
