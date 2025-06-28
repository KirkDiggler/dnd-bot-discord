package character

import (
	"fmt"
	"strings"
)

// ProgressStep represents a step in character creation
type ProgressStep struct {
	Name      string
	Completed bool
}

// BuildProgressValue creates a progress field value for character creation
// It dynamically adjusts based on the character's class
func BuildProgressValue(classKey, currentStep string) string {
	steps := []ProgressStep{
		{Name: "Step 1: Race", Completed: true},
		{Name: "Step 2: Class", Completed: true},
		{Name: "Step 3: Abilities", Completed: true},
	}

	// Add class features step for classes that need it
	nextStep := 4
	if needsClassFeatures(classKey) {
		featureStepName := fmt.Sprintf("Step %d: Class Features", nextStep)
		steps = append(steps, ProgressStep{
			Name:      featureStepName,
			Completed: isClassFeaturesStepCompleted(currentStep),
		})
		nextStep++
	}

	// Add remaining steps
	steps = append(steps,
		ProgressStep{
			Name:      fmt.Sprintf("Step %d: Proficiencies", nextStep),
			Completed: currentStep == "equipment" || currentStep == "details",
		},
		ProgressStep{
			Name:      fmt.Sprintf("Step %d: Equipment", nextStep+1),
			Completed: currentStep == "details",
		},
		ProgressStep{
			Name:      fmt.Sprintf("Step %d: Details", nextStep+2),
			Completed: false,
		},
	)

	// Build the progress string
	var progressLines []string
	for _, step := range steps {
		icon := "⏳"
		if step.Completed {
			icon = "✅"
		}
		progressLines = append(progressLines, fmt.Sprintf("%s %s", icon, step.Name))
	}

	return strings.Join(progressLines, "\n")
}

// needsClassFeatures returns true if the class requires feature selection at level 1
func needsClassFeatures(classKey string) bool {
	switch classKey {
	case "ranger", "cleric", "warlock", "fighter", "sorcerer":
		return true
	default:
		return false
	}
}

// isClassFeaturesStepCompleted returns true if the class features step is completed
func isClassFeaturesStepCompleted(currentStep string) bool {
	// Class features are completed if we're past that step
	// (i.e., not currently on "class_features" or "abilities")
	return currentStep != "class_features" && currentStep != "abilities"
}
