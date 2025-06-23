package discord

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestProficiencyCustomIDParsing tests that the proficiency confirmation custom ID is parsed correctly
func TestProficiencyCustomIDParsing(t *testing.T) {
	tests := []struct {
		name          string
		customID      string
		expectedParts int
		shouldHandle  bool
	}{
		{
			name:          "Valid proficiency confirmation",
			customID:      "character_create:confirm_proficiency:human:rogue:class:0",
			expectedParts: 6,
			shouldHandle:  true,
		},
		{
			name:          "Too few parts",
			customID:      "character_create:confirm_proficiency:human:rogue",
			expectedParts: 4,
			shouldHandle:  false,
		},
		{
			name:          "Different action",
			customID:      "character_create:select_proficiencies:human:rogue",
			expectedParts: 4,
			shouldHandle:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts := strings.Split(tt.customID, ":")
			assert.Equal(t, tt.expectedParts, len(parts), "Parts length mismatch")

			// Check if it should be handled by confirm_proficiency case
			if len(parts) >= 2 {
				action := parts[1]
				willHandle := action == "confirm_proficiency" && len(parts) >= 6
				assert.Equal(t, tt.shouldHandle, willHandle,
					"Handler mismatch for action %s with %d parts", action, len(parts))
			}
		})
	}
}

// TestDebugProficiencyFlow helps debug the actual flow
func TestDebugProficiencyFlow(t *testing.T) {
	// This test logs what's happening to help debug
	customID := "character_create:confirm_proficiency:human:rogue:class:0"
	parts := strings.Split(customID, ":")

	t.Logf("Custom ID: %s", customID)
	t.Logf("Parts: %v", parts)
	t.Logf("Parts length: %d", len(parts))
	t.Logf("Action (parts[1]): %s", parts[1])

	if len(parts) >= 6 {
		t.Logf("Race (parts[2]): %s", parts[2])
		t.Logf("Class (parts[3]): %s", parts[3])
		t.Logf("Choice Type (parts[4]): %s", parts[4])
		t.Logf("Choice Index (parts[5]): %s", parts[5])
	}

	// Verify this matches what the handler expects
	assert.Equal(t, "confirm_proficiency", parts[1])
	assert.Equal(t, 6, len(parts))
}
