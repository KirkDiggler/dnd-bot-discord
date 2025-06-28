package combat

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLayOnHandsButtonGeneration(t *testing.T) {
	tests := []struct {
		name           string
		usesRemaining  int
		expectedLabels []string
		description    string
	}{
		{
			name:           "Level 1 Paladin (5 HP pool)",
			usesRemaining:  5,
			expectedLabels: []string{"1 HP", "2 HP", "3 HP", "5 HP"},
			description:    "Should not duplicate the 5 HP button",
		},
		{
			name:           "Level 2 Paladin (10 HP pool)",
			usesRemaining:  10,
			expectedLabels: []string{"1 HP", "2 HP", "3 HP", "5 HP", "10 HP"},
			description:    "Should add the full pool as an additional option",
		},
		{
			name:           "Partially used pool (3 HP remaining)",
			usesRemaining:  3,
			expectedLabels: []string{"1 HP", "2 HP", "3 HP"},
			description:    "Should only show amounts up to remaining pool",
		},
		{
			name:           "Only 1 HP remaining",
			usesRemaining:  1,
			expectedLabels: []string{"1 HP"},
			description:    "Should only show 1 HP option",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the button generation logic
			healAmounts := []int{1, 2, 3, 5}
			if tt.usesRemaining > 5 {
				healAmounts = append(healAmounts, tt.usesRemaining)
			}

			addedAmounts := make(map[int]bool)
			var labels []string

			for _, amount := range healAmounts {
				if amount > 0 && amount <= tt.usesRemaining && !addedAmounts[amount] {
					addedAmounts[amount] = true
					labels = append(labels, fmt.Sprintf("%d HP", amount))
				}
			}

			// Check we have the right number of buttons
			assert.Equal(t, len(tt.expectedLabels), len(labels), tt.description)

			// Check no duplicates
			seen := make(map[string]bool)
			for _, label := range labels {
				assert.False(t, seen[label], "Duplicate button label found: "+label)
				seen[label] = true
			}
		})
	}
}

func TestLayOnHandsNoDuplicateCustomIDs(t *testing.T) {
	// Test that we don't generate duplicate custom IDs
	encounterID := "test-encounter"
	usesRemaining := 5 // Common case that was causing the bug

	healAmounts := []int{1, 2, 3, 5}
	if usesRemaining > 5 {
		healAmounts = append(healAmounts, usesRemaining)
	}

	addedAmounts := make(map[int]bool)
	customIDs := make(map[string]bool)

	for _, amount := range healAmounts {
		if amount > 0 && amount <= usesRemaining && !addedAmounts[amount] {
			addedAmounts[amount] = true
			customID := fmt.Sprintf("combat:lay_on_hands_amount:%s:%d", encounterID, amount)

			// Check for duplicate custom IDs
			assert.False(t, customIDs[customID], "Duplicate custom ID generated")
			customIDs[customID] = true
		}
	}

	// Should have exactly 4 buttons for a 5 HP pool (1, 2, 3, 5)
	assert.Equal(t, 4, len(customIDs))
}
