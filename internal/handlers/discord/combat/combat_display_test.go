package combat

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/combat"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildInitiativeDisplay(t *testing.T) {
	// Create test encounter
	enc := &combat.Encounter{
		ID:     "test-enc",
		Name:   "Test Combat",
		Round:  2,
		Turn:   1, // Second combatant's turn
		Status: combat.EncounterStatusActive,
		Combatants: map[string]*combat.Combatant{
			"c1": {
				ID:         "c1",
				Name:       "Grunk",
				Class:      "Barbarian",
				Type:       combat.CombatantTypePlayer,
				Initiative: 18,
				AC:         16,
				MaxHP:      50,
				CurrentHP:  45,
				IsActive:   true,
			},
			"c2": {
				ID:         "c2",
				Name:       "Goblin",
				Type:       combat.CombatantTypeMonster,
				Initiative: 15,
				AC:         13,
				MaxHP:      12,
				CurrentHP:  3,
				IsActive:   true,
			},
			"c3": {
				ID:         "c3",
				Name:       "Thorin",
				Class:      "Fighter",
				Type:       combat.CombatantTypePlayer,
				Initiative: 10,
				AC:         18,
				MaxHP:      40,
				CurrentHP:  40,
				IsActive:   true,
			},
		},
		TurnOrder: []string{"c1", "c2", "c3"},
	}

	display := BuildInitiativeDisplay(enc)

	// Check that it contains expected elements
	assert.Contains(t, display, "```ansi")
	assert.Contains(t, display, "Init")
	assert.Contains(t, display, "Name")
	assert.Contains(t, display, "HP")
	assert.Contains(t, display, "AC")

	// Check for combatants
	assert.Contains(t, display, "Grunk (Barbarian)")
	assert.Contains(t, display, "Goblin")
	assert.Contains(t, display, "Thorin (Fighter)")

	// Check that current turn indicator is present
	assert.Contains(t, display, "▶")

	// Verify color codes are present (ANSI escape sequences)
	assert.Contains(t, display, "\u001b[")
}

func TestBuildCombatSummaryDisplay(t *testing.T) {
	enc := &combat.Encounter{
		Round: 3,
		Combatants: map[string]*combat.Combatant{
			"c1": {
				Name:     "Player1",
				Type:     combat.CombatantTypePlayer,
				IsActive: true,
			},
			"c2": {
				Name:     "Player2",
				Type:     combat.CombatantTypePlayer,
				IsActive: true,
			},
			"c3": {
				Name:     "Goblin",
				Type:     combat.CombatantTypeMonster,
				IsActive: true,
			},
			"c4": {
				Name:     "Orc",
				Type:     combat.CombatantTypeMonster,
				IsActive: false, // Dead
			},
		},
		TurnOrder: []string{"c1", "c2", "c3"},
		Turn:      0,
	}

	summary := BuildCombatSummaryDisplay(enc)

	assert.Contains(t, summary, "Round 3")
	assert.Contains(t, summary, "Players: 2")
	assert.Contains(t, summary, "Monsters: 1") // Only 1 active
	assert.Contains(t, summary, "Player1's turn")
}

func TestInitiativeDisplay_ColorCoding(t *testing.T) {
	enc := &combat.Encounter{
		Combatants: map[string]*combat.Combatant{
			"c1": {
				ID:         "c1",
				Name:       "HealthyPlayer",
				Type:       combat.CombatantTypePlayer,
				Initiative: 20,
				AC:         15,
				MaxHP:      40,
				CurrentHP:  40, // Full health
				IsActive:   true,
			},
			"c2": {
				ID:         "c2",
				Name:       "HurtPlayer",
				Type:       combat.CombatantTypePlayer,
				Initiative: 15,
				AC:         14,
				MaxHP:      40,
				CurrentHP:  15, // Less than 50%
				IsActive:   true,
			},
			"c3": {
				ID:         "c3",
				Name:       "CriticalMonster",
				Type:       combat.CombatantTypeMonster,
				Initiative: 10,
				AC:         12,
				MaxHP:      20,
				CurrentHP:  3, // Less than 25%
				IsActive:   true,
			},
		},
		TurnOrder: []string{"c1", "c2", "c3"},
		Turn:      0,
	}

	display := BuildInitiativeDisplay(enc)

	// Split into lines for easier testing
	lines := strings.Split(display, "\n")

	// Find each combatant's line and check for appropriate color codes
	for _, line := range lines {
		if strings.Contains(line, "HealthyPlayer") {
			assert.Contains(t, line, "\u001b[32m") // Green for healthy
		} else if strings.Contains(line, "HurtPlayer") {
			assert.Contains(t, line, "\u001b[33m") // Yellow for hurt
		} else if strings.Contains(line, "CriticalMonster") {
			assert.Contains(t, line, "\u001b[31m") // Red for critical
		}
	}
}
