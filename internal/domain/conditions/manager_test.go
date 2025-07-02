package conditions

import (
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_AddCondition(t *testing.T) {
	eventBus := events.NewEventBus()
	manager := NewManager("test-char-1", eventBus)

	t.Run("add poisoned condition", func(t *testing.T) {
		condition, err := manager.AddCondition(Poisoned, "Giant Spider", DurationRounds, 3)
		require.NoError(t, err)
		assert.NotNil(t, condition)
		assert.Equal(t, Poisoned, condition.Type)
		assert.Equal(t, "Giant Spider", condition.Source)
		assert.Equal(t, DurationRounds, condition.DurationType)
		assert.Equal(t, 3, condition.Duration)
		assert.Equal(t, 3, condition.Remaining)
	})

	t.Run("conditions don't stack by default", func(t *testing.T) {
		// Add poisoned again
		condition2, err := manager.AddCondition(Poisoned, "Another Spider", DurationRounds, 5)
		require.NoError(t, err)

		// Should return existing condition but with updated duration
		conditions := manager.GetConditions()
		assert.Len(t, conditions, 1)
		assert.Equal(t, 5, condition2.Remaining) // Duration refreshed to higher value
	})

	t.Run("exhaustion stacks", func(t *testing.T) {
		// Add exhaustion
		ex1, err := manager.AddCondition(Exhaustion, "Forced March", DurationPermanent, 0)
		require.NoError(t, err)
		assert.Equal(t, 1, ex1.Level)

		// Add exhaustion again
		ex2, err := manager.AddCondition(Exhaustion, "Starvation", DurationPermanent, 0)
		require.NoError(t, err)
		assert.Equal(t, 2, ex2.Level)

		// Should still be one condition but at level 2
		conditions := manager.GetConditions()
		exhaustionCount := 0
		for _, c := range conditions {
			if c.Type == Exhaustion {
				exhaustionCount++
				assert.Equal(t, 2, c.Level)
			}
		}
		assert.Equal(t, 1, exhaustionCount)
	})
}

func TestManager_RemoveCondition(t *testing.T) {
	eventBus := events.NewEventBus()
	manager := NewManager("test-char-2", eventBus)

	// Add a condition
	condition, err := manager.AddCondition(Stunned, "Mind Flayer", DurationRounds, 1)
	require.NoError(t, err)

	// Remove it
	err = manager.RemoveCondition(condition.ID)
	assert.NoError(t, err)

	// Verify it's gone
	assert.False(t, manager.HasCondition(Stunned))
	assert.Len(t, manager.GetConditions(), 0)
}

func TestManager_ProcessTurnStart(t *testing.T) {
	eventBus := events.NewEventBus()
	manager := NewManager("test-char-3", eventBus)

	// Add condition that lasts 2 turns
	_, err := manager.AddCondition(DisadvantageOnNextAttack, "spell", DurationTurns, 2)
	require.NoError(t, err)

	// Process first turn
	manager.ProcessTurnStart()
	conditions := manager.GetConditions()
	require.Len(t, conditions, 1)
	assert.Equal(t, 1, conditions[0].Remaining)

	// Process second turn
	manager.ProcessTurnStart()
	conditions = manager.GetConditions()
	require.Len(t, conditions, 1)
	assert.Equal(t, 0, conditions[0].Remaining)

	// Process third turn - should be removed
	manager.ProcessTurnStart()
	assert.Len(t, manager.GetConditions(), 0)
}

func TestManager_ProcessRoundEnd(t *testing.T) {
	eventBus := events.NewEventBus()
	manager := NewManager("test-char-4", eventBus)

	// Add condition that lasts 2 rounds
	_, err := manager.AddCondition(Rage, "Barbarian", DurationRounds, 2)
	require.NoError(t, err)

	// Process first round end
	manager.ProcessRoundEnd()
	conditions := manager.GetConditions()
	require.Len(t, conditions, 1)
	assert.Equal(t, 1, conditions[0].Remaining)

	// Process second round end
	manager.ProcessRoundEnd()
	conditions = manager.GetConditions()
	require.Len(t, conditions, 1)
	assert.Equal(t, 0, conditions[0].Remaining)

	// Process third round end - should be removed
	manager.ProcessRoundEnd()
	assert.Len(t, manager.GetConditions(), 0)
}

func TestManager_ProcessDamage(t *testing.T) {
	eventBus := events.NewEventBus()
	manager := NewManager("test-char-5", eventBus)

	// Add condition that ends on damage
	_, err := manager.AddCondition(DisadvantageOnNextAttack, "spell", DurationUntilDamaged, 0)
	require.NoError(t, err)

	// Take damage
	manager.ProcessDamage(5)

	// Condition should be removed
	assert.False(t, manager.HasCondition(DisadvantageOnNextAttack))
	assert.Len(t, manager.GetConditions(), 0)
}

func TestManager_GetActiveEffects(t *testing.T) {
	eventBus := events.NewEventBus()
	manager := NewManager("test-char-6", eventBus)

	// Add multiple conditions
	_, err := manager.AddCondition(Poisoned, "source1", DurationRounds, 3)
	require.NoError(t, err)
	_, err = manager.AddCondition(Prone, "source2", DurationPermanent, 0)
	require.NoError(t, err)

	// Get combined effects
	effects := manager.GetActiveEffects()

	// Poisoned gives disadvantage on attacks
	assert.True(t, effects.AttackDisadvantage)
	// Poisoned gives disadvantage on all saves
	assert.True(t, effects.SaveDisadvantage["all"])
	// Prone reduces speed
	assert.Equal(t, 0.5, effects.SpeedMultiplier)
	// Prone causes fall prone
	assert.True(t, effects.FallProne)
}

func TestGetStandardEffects(t *testing.T) {
	tests := []struct {
		condition ConditionType
		check     func(*Effect)
	}{
		{
			condition: Blinded,
			check: func(e *Effect) {
				assert.True(t, e.AttackDisadvantage)
				assert.True(t, e.DefenseAdvantage)
			},
		},
		{
			condition: Paralyzed,
			check: func(e *Effect) {
				assert.True(t, e.Incapacitated)
				assert.True(t, e.CantMove)
				assert.True(t, e.CantSpeak)
				assert.True(t, e.DefenseAdvantage)
				assert.True(t, e.SaveAutoFail["STR"])
				assert.True(t, e.SaveAutoFail["DEX"])
			},
		},
		{
			condition: Stunned,
			check: func(e *Effect) {
				assert.True(t, e.Incapacitated)
				assert.True(t, e.CantAct)
				assert.True(t, e.CantReact)
				assert.True(t, e.DefenseAdvantage)
			},
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.condition), func(t *testing.T) {
			effects := GetStandardEffects(tt.condition)
			tt.check(effects)
		})
	}
}
