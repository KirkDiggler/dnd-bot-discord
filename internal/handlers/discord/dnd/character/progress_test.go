package character

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildProgressValue(t *testing.T) {
	tests := []struct {
		name             string
		classKey         string
		currentStep      string
		expectedSteps    []string
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			name:        "fighter class after class selection",
			classKey:    "fighter",
			currentStep: "class",
			expectedSteps: []string{
				"✅ Step 1: Race",
				"✅ Step 2: Class",
				"⏳ Step 3: Abilities",
				"⏳ Step 4: Class Features",
				"⏳ Step 5: Proficiencies",
				"⏳ Step 6: Equipment",
				"⏳ Step 7: Details",
			},
			shouldContain: []string{"Class Features"},
		},
		{
			name:        "wizard class after class selection",
			classKey:    "wizard",
			currentStep: "class",
			expectedSteps: []string{
				"✅ Step 1: Race",
				"✅ Step 2: Class",
				"⏳ Step 3: Abilities",
				"⏳ Step 4: Proficiencies",
				"⏳ Step 5: Equipment",
				"⏳ Step 6: Details",
			},
			shouldNotContain: []string{"Class Features"},
		},
		{
			name:          "ranger class after class selection",
			classKey:      "ranger",
			currentStep:   "class",
			shouldContain: []string{"Class Features"},
		},
		{
			name:        "fighter on abilities step",
			classKey:    "fighter",
			currentStep: "abilities",
			shouldContain: []string{
				"⏳ Step 3: Abilities",
				"⏳ Step 4: Class Features",
			},
		},
		{
			name:        "fighter on class features step",
			classKey:    "fighter",
			currentStep: "class_features",
			shouldContain: []string{
				"✅ Step 3: Abilities",
				"⏳ Step 4: Class Features",
				"⏳ Step 5: Proficiencies",
			},
		},
		{
			name:        "fighter on proficiencies step",
			classKey:    "fighter",
			currentStep: "proficiencies",
			shouldContain: []string{
				"✅ Step 3: Abilities",
				"✅ Step 4: Class Features",
				"⏳ Step 5: Proficiencies",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildProgressValue(tt.classKey, tt.currentStep)

			// Check that expected strings are present
			for _, expected := range tt.shouldContain {
				assert.Contains(t, result, expected, "Progress should contain: %s", expected)
			}

			// Check that unexpected strings are not present
			for _, unexpected := range tt.shouldNotContain {
				assert.NotContains(t, result, unexpected, "Progress should not contain: %s", unexpected)
			}

			// If full expected steps are provided, check line by line
			if len(tt.expectedSteps) > 0 {
				lines := strings.Split(result, "\n")
				assert.Equal(t, len(tt.expectedSteps), len(lines), "Number of progress steps should match")

				for i, expected := range tt.expectedSteps {
					if i < len(lines) {
						assert.Equal(t, expected, lines[i], "Step %d should match", i+1)
					}
				}
			}
		})
	}
}

func TestNeedsClassFeatures(t *testing.T) {
	tests := []struct {
		classKey string
		expected bool
	}{
		{"fighter", true},
		{"ranger", true},
		{"cleric", true},
		{"warlock", true},
		{"sorcerer", true},
		{"wizard", false},
		{"rogue", false},
		{"barbarian", false},
		{"monk", false},
		{"paladin", false},
	}

	for _, tt := range tests {
		t.Run(tt.classKey, func(t *testing.T) {
			result := needsClassFeatures(tt.classKey)
			assert.Equal(t, tt.expected, result, "Class %s should have needsClassFeatures=%v", tt.classKey, tt.expected)
		})
	}
}
