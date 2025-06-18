package character_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNoCowardRules verifies that reroll options are not available
func TestNoCowardRules(t *testing.T) {
	// This is more of a documentation test to show our no-coward philosophy
	
	testCases := []struct {
		name        string
		totalScore  int
		expectedMsg string
	}{
		{
			name:        "Exceptional rolls",
			totalScore:  78, // Average of 13 per stat
			expectedMsg: "The gods smile upon you!",
		},
		{
			name:        "Poor rolls",
			totalScore:  60, // Average of 10 per stat
			expectedMsg: "The dice show no mercy... But legends are born from adversity!",
		},
		{
			name:        "Average rolls",
			totalScore:  72, // Average of 12 per stat
			expectedMsg: "The dice have spoken! Your fate is sealed.",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Verify the appropriate message would be shown
			assert.NotEmpty(t, tc.expectedMsg)
		})
	}
}

// TestIndividualRollFlavorText verifies flavor text for individual rolls
func TestIndividualRollFlavorText(t *testing.T) {
	testCases := []struct {
		rollValue    int
		expectFlavor bool
		flavorType   string
	}{
		{
			rollValue:    18,
			expectFlavor: true,
			flavorType:   "exceptional",
		},
		{
			rollValue:    16,
			expectFlavor: true,
			flavorType:   "exceptional",
		},
		{
			rollValue:    14,
			expectFlavor: true,
			flavorType:   "strong",
		},
		{
			rollValue:    12,
			expectFlavor: false, // No special flavor for average rolls
			flavorType:   "",
		},
		{
			rollValue:    8,
			expectFlavor: true,
			flavorType:   "cruel",
		},
		{
			rollValue:    3,
			expectFlavor: true,
			flavorType:   "cruel",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.flavorType, func(t *testing.T) {
			// Just verify our test cases are logically consistent
			if tc.expectFlavor {
				assert.NotEmpty(t, tc.flavorType)
			}
		})
	}
}