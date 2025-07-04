package shared_test

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestActiveAbility_CanUse(t *testing.T) {
	tests := []struct {
		name     string
		ability  *shared.ActiveAbility
		expected bool
	}{
		{
			name: "ability with uses remaining",
			ability: &shared.ActiveAbility{
				UsesMax:       3,
				UsesRemaining: 2,
			},
			expected: true,
		},
		{
			name: "ability with no uses remaining",
			ability: &shared.ActiveAbility{
				UsesMax:       3,
				UsesRemaining: 0,
			},
			expected: false,
		},
		{
			name: "unlimited use ability",
			ability: &shared.ActiveAbility{
				UsesMax:       -1,
				UsesRemaining: 0,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.ability.CanUse())
		})
	}
}

func TestActiveAbility_Use(t *testing.T) {
	t.Run("use limited ability", func(t *testing.T) {
		ability := &shared.ActiveAbility{
			UsesMax:       3,
			UsesRemaining: 2,
			Duration:      10,
		}

		success := ability.Use()
		assert.True(t, success)
		assert.Equal(t, 1, ability.UsesRemaining)
		assert.True(t, ability.IsActive)
	})

	t.Run("use unlimited ability", func(t *testing.T) {
		ability := &shared.ActiveAbility{
			UsesMax:       -1,
			UsesRemaining: 0,
		}

		success := ability.Use()
		assert.True(t, success)
		assert.Equal(t, 0, ability.UsesRemaining) // Should not change
	})

	t.Run("cannot use depleted ability", func(t *testing.T) {
		ability := &shared.ActiveAbility{
			UsesMax:       1,
			UsesRemaining: 0,
		}

		success := ability.Use()
		assert.False(t, success)
	})
}

func TestActiveAbility_TickDuration(t *testing.T) {
	t.Run("tick active ability", func(t *testing.T) {
		ability := &shared.ActiveAbility{
			Duration: 3,
			IsActive: true,
		}

		ability.TickDuration()
		assert.Equal(t, 2, ability.Duration)
		assert.True(t, ability.IsActive)

		ability.TickDuration()
		assert.Equal(t, 1, ability.Duration)
		assert.True(t, ability.IsActive)

		ability.TickDuration()
		assert.Equal(t, 0, ability.Duration)
		assert.False(t, ability.IsActive) // Should deactivate when duration expires
	})

	t.Run("tick ability with no duration", func(t *testing.T) {
		ability := &shared.ActiveAbility{
			Duration: 0,
			IsActive: true,
		}

		ability.TickDuration()
		assert.Equal(t, 0, ability.Duration)
		assert.True(t, ability.IsActive) // Should not change
	})
}

func TestActiveAbility_RestoreUses(t *testing.T) {
	t.Run("long rest restores all abilities", func(t *testing.T) {
		abilities := []*shared.ActiveAbility{
			{
				RestType:      shared.RestTypeLong,
				UsesMax:       3,
				UsesRemaining: 0,
				IsActive:      true,
				Duration:      5,
			},
			{
				RestType:      shared.RestTypeShort,
				UsesMax:       2,
				UsesRemaining: 0,
			},
		}

		for _, ability := range abilities {
			ability.RestoreUses(shared.RestTypeLong)
			assert.Equal(t, ability.UsesMax, ability.UsesRemaining)
			assert.False(t, ability.IsActive)
			assert.Equal(t, 0, ability.Duration)
		}
	})

	t.Run("short rest only restores short rest abilities", func(t *testing.T) {
		longRestAbility := &shared.ActiveAbility{
			RestType:      shared.RestTypeLong,
			UsesMax:       3,
			UsesRemaining: 0,
		}

		shortRestAbility := &shared.ActiveAbility{
			RestType:      shared.RestTypeShort,
			UsesMax:       2,
			UsesRemaining: 0,
		}

		longRestAbility.RestoreUses(shared.RestTypeShort)
		shortRestAbility.RestoreUses(shared.RestTypeShort)

		assert.Equal(t, 0, longRestAbility.UsesRemaining)  // Should not restore
		assert.Equal(t, 2, shortRestAbility.UsesRemaining) // Should restore
	})

	t.Run("never restore abilities are not restored", func(t *testing.T) {
		ability := &shared.ActiveAbility{
			RestType:      shared.RestTypeNone,
			UsesMax:       1,
			UsesRemaining: 0,
		}

		ability.RestoreUses(shared.RestTypeLong)
		assert.Equal(t, 0, ability.UsesRemaining)
	})
}
