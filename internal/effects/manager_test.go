package effects

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_AddEffect(t *testing.T) {
	manager := NewManager()

	t.Run("adds simple effect", func(t *testing.T) {
		effect := &StatusEffect{
			ID:     "test_effect_1",
			Name:   "Test Effect",
			Source: SourceAbility,
		}

		err := manager.AddEffect(effect)
		require.NoError(t, err)

		active := manager.GetActiveEffects()
		assert.Len(t, active, 1)
		assert.Equal(t, "Test Effect", active[0].Name)
	})

	t.Run("rejects effect without ID", func(t *testing.T) {
		effect := &StatusEffect{
			Name:   "No ID Effect",
			Source: SourceAbility,
		}

		err := manager.AddEffect(effect)
		assert.Error(t, err)
	})

	t.Run("handles stacking replace", func(t *testing.T) {
		// Create new manager for this test
		replaceManager := NewManager()

		// Add first effect
		effect1 := &StatusEffect{
			ID:           "rage_1",
			Name:         "Rage",
			Source:       SourceAbility,
			StackingRule: StackingReplace,
		}
		require.NoError(t, replaceManager.AddEffect(effect1))

		// Add second effect with same name and source
		effect2 := &StatusEffect{
			ID:           "rage_2",
			Name:         "Rage",
			Source:       SourceAbility,
			StackingRule: StackingReplace,
		}
		require.NoError(t, replaceManager.AddEffect(effect2))

		active := replaceManager.GetActiveEffects()
		assert.Len(t, active, 1)
		assert.Equal(t, "rage_2", active[0].ID)
	})

	t.Run("handles stacking stack", func(t *testing.T) {
		manager = NewManager()

		// Add multiple stackable effects
		for i := 0; i < 3; i++ {
			effect := &StatusEffect{
				ID:           fmt.Sprintf("bless_%d", i),
				Name:         "Bless",
				Source:       SourceSpell,
				StackingRule: StackingStack,
			}
			require.NoError(t, manager.AddEffect(effect))
		}

		active := manager.GetActiveEffects()
		assert.Len(t, active, 3)
	})
}

func TestManager_RemoveEffect(t *testing.T) {
	manager := NewManager()

	effect := &StatusEffect{
		ID:     "test_effect",
		Name:   "Test Effect",
		Source: SourceAbility,
	}
	require.NoError(t, manager.AddEffect(effect))

	manager.RemoveEffect("test_effect")
	active := manager.GetActiveEffects()
	assert.Len(t, active, 0)
}

func TestManager_RemoveEffectsBySource(t *testing.T) {
	manager := NewManager()

	// Add effects from different sources
	spellEffect := &StatusEffect{
		ID:       "spell_1",
		Name:     "Bless",
		Source:   SourceSpell,
		SourceID: "bless_spell",
	}
	require.NoError(t, manager.AddEffect(spellEffect))

	abilityEffect := &StatusEffect{
		ID:       "ability_1",
		Name:     "Rage",
		Source:   SourceAbility,
		SourceID: "barbarian_rage",
	}
	require.NoError(t, manager.AddEffect(abilityEffect))

	// Remove all spell effects
	manager.RemoveEffectsBySource(SourceSpell, "bless_spell")

	active := manager.GetActiveEffects()
	assert.Len(t, active, 1)
	assert.Equal(t, "Rage", active[0].Name)
}

func TestManager_GetModifiers(t *testing.T) {
	manager := NewManager()

	t.Run("gets modifiers for target", func(t *testing.T) {
		effect := BuildRageEffect(1)
		require.NoError(t, manager.AddEffect(effect))

		// Get damage modifiers
		modifiers := manager.GetModifiers(TargetDamage, map[string]string{
			"attack_type": "melee",
		})

		assert.Len(t, modifiers, 1)
		assert.Equal(t, "+2", modifiers[0].Value)
	})

	t.Run("respects modifier conditions", func(t *testing.T) {
		effect := BuildRageEffect(1)
		require.NoError(t, manager.AddEffect(effect))

		// Get damage modifiers for ranged attack (should not apply)
		modifiers := manager.GetModifiers(TargetDamage, map[string]string{
			"attack_type": "ranged",
		})

		assert.Len(t, modifiers, 0)
	})

	t.Run("respects effect conditions", func(t *testing.T) {
		effect := BuildFavoredEnemyEffect("orc")
		require.NoError(t, manager.AddEffect(effect))

		// Get modifiers when fighting an orc
		modifiers := manager.GetModifiers(TargetSkillCheck, map[string]string{
			"enemy_type": "orc",
		})
		assert.Len(t, modifiers, 1)

		// Get modifiers when fighting a goblin
		modifiers = manager.GetModifiers(TargetSkillCheck, map[string]string{
			"enemy_type": "goblin",
		})
		assert.Len(t, modifiers, 0)
	})
}

func TestManager_Expiration(t *testing.T) {
	manager := NewManager()

	t.Run("instant effects expire immediately", func(t *testing.T) {
		effect := &StatusEffect{
			ID:       "instant",
			Name:     "Instant Effect",
			Duration: Duration{Type: DurationInstant},
		}
		require.NoError(t, manager.AddEffect(effect))

		// Sleep to ensure time passes
		time.Sleep(10 * time.Millisecond)

		active := manager.GetActiveEffects()
		assert.Len(t, active, 0)
	})

	t.Run("permanent effects don't expire", func(t *testing.T) {
		effect := &StatusEffect{
			ID:       "permanent",
			Name:     "Permanent Effect",
			Duration: Duration{Type: DurationPermanent},
		}
		require.NoError(t, manager.AddEffect(effect))

		active := manager.GetActiveEffects()
		assert.Len(t, active, 1)
	})
}

func TestManager_ClearConcentration(t *testing.T) {
	manager := NewManager()

	// Add concentration effect
	concentrationEffect := BuildBlessEffect()
	require.NoError(t, manager.AddEffect(concentrationEffect))

	// Add non-concentration effect
	normalEffect := BuildRageEffect(1)
	require.NoError(t, manager.AddEffect(normalEffect))

	// Clear concentration
	manager.ClearConcentration()

	active := manager.GetActiveEffects()
	assert.Len(t, active, 1)
	assert.Equal(t, "Rage", active[0].Name)
}

func TestCheckModifierCondition(t *testing.T) {
	tests := []struct {
		name       string
		condition  string
		conditions map[string]string
		expected   bool
	}{
		{
			name:      "melee only - melee attack",
			condition: "melee_only",
			conditions: map[string]string{
				"attack_type": "melee",
			},
			expected: true,
		},
		{
			name:      "melee only - ranged attack",
			condition: "melee_only",
			conditions: map[string]string{
				"attack_type": "ranged",
			},
			expected: false,
		},
		{
			name:      "vs enemy type - matching",
			condition: "vs_enemy_type:orc",
			conditions: map[string]string{
				"enemy_type": "orc",
			},
			expected: true,
		},
		{
			name:      "vs enemy type - not matching",
			condition: "vs_enemy_type:orc",
			conditions: map[string]string{
				"enemy_type": "goblin",
			},
			expected: false,
		},
		{
			name:      "with weapon - matching",
			condition: "with_weapon:longsword",
			conditions: map[string]string{
				"weapon": "longsword",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checkModifierCondition(tt.condition, tt.conditions)
			assert.Equal(t, tt.expected, result)
		})
	}
}
