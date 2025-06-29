package combat

import (
	"strings"
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildInitiativeFields_DeadCreatureAlignment(t *testing.T) {
	// Create test encounter with dead and alive creatures
	enc := &entities.Encounter{
		ID:        "test-encounter",
		Round:     1,
		Turn:      0,
		TurnOrder: []string{"player1", "goblin1", "goblin2"},
		Combatants: map[string]*entities.Combatant{
			"player1": {
				ID:         "player1",
				Name:       "Stanthony",
				Type:       entities.CombatantTypePlayer,
				Class:      "Fighter",
				Initiative: 20,
				CurrentHP:  9,
				MaxHP:      13,
				AC:         18,
				IsActive:   true,
			},
			"goblin1": {
				ID:         "goblin1",
				Name:       "Goblin",
				Type:       entities.CombatantTypeMonster,
				Initiative: 17,
				CurrentHP:  0, // Dead
				MaxHP:      7,
				AC:         15,
				IsActive:   true,
			},
			"goblin2": {
				ID:         "goblin2",
				Name:       "Skeleton",
				Type:       entities.CombatantTypeMonster,
				Initiative: 15,
				CurrentHP:  13,
				MaxHP:      13,
				AC:         13,
				IsActive:   true,
			},
		},
	}

	// Build initiative display
	fields := BuildInitiativeFields(enc)
	require.Len(t, fields, 1)

	display := fields[0].Value

	// Verify the display contains our combatants
	assert.Contains(t, display, "Stanthony")
	assert.Contains(t, display, "Goblin")
	assert.Contains(t, display, "Skeleton")

	// Extract lines to check formatting
	lines := strings.Split(display, "\n")

	// Find the dead goblin line
	var goblinLine string
	for _, line := range lines {
		if strings.Contains(line, "Goblin") && strings.Contains(line, "0/") {
			goblinLine = line
			break
		}
	}

	require.NotEmpty(t, goblinLine, "Should find dead goblin line")

	// Check that the dead goblin has 💀 in the name column, not after AC
	assert.Contains(t, goblinLine, "💀 Goblin")

	// Check that line ends with AC value and no trailing emoji
	// The line should end with the AC value (15) and not have 💀 after it
	trimmed := strings.TrimSpace(goblinLine)
	assert.Regexp(t, `15\s*$`, trimmed, "Line should end with AC value")
	assert.NotRegexp(t, `15\s*💀`, trimmed, "Skull emoji should not appear after AC")

	// Verify table alignment by checking all lines have consistent structure
	for _, line := range lines {
		if strings.Contains(line, "│") && !strings.Contains(line, "Init│") && !strings.Contains(line, "────") {
			// Count pipe characters - should be consistent
			pipeCount := strings.Count(line, "│")
			assert.Equal(t, 3, pipeCount, "Each data line should have exactly 3 pipe characters")
		}
	}
}

func TestBuildInitiativeFields_LowHealthWarning(t *testing.T) {
	// Test that low health creatures don't get extra indicators breaking alignment
	enc := &entities.Encounter{
		ID:        "test-encounter",
		Round:     1,
		Turn:      0,
		TurnOrder: []string{"player1"},
		Combatants: map[string]*entities.Combatant{
			"player1": {
				ID:         "player1",
				Name:       "LowHealthHero",
				Type:       entities.CombatantTypePlayer,
				Class:      "Fighter",
				Initiative: 20,
				CurrentHP:  2, // Very low health
				MaxHP:      20,
				AC:         15,
				IsActive:   true,
			},
		},
	}

	fields := BuildInitiativeFields(enc)
	require.Len(t, fields, 1)

	display := fields[0].Value

	// Verify no warning indicators appear after AC
	lines := strings.Split(display, "\n")
	for _, line := range lines {
		if strings.Contains(line, "LowHealthHero") {
			trimmed := strings.TrimSpace(line)
			// Should end with AC value, no extra symbols
			assert.Regexp(t, `15\s*$`, trimmed, "Line should end with AC value")
			assert.NotContains(t, trimmed, "❗", "Warning emoji should not appear in table")
		}
	}
}
