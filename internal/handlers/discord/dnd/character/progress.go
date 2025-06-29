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
	// Determine which steps are completed based on current step
	completedSteps := getCompletedSteps(currentStep)

	raceCompleted := completedSteps["race"]
	classCompleted := completedSteps["class"]
	abilitiesCompleted := completedSteps["abilities"]

	steps := []ProgressStep{
		{Name: "Step 1: Race", Completed: raceCompleted},
		{Name: "Step 2: Class", Completed: classCompleted},
		{Name: "Step 3: Abilities", Completed: abilitiesCompleted},
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
	// Class features are completed only if we're past that step
	completedSteps := map[string]bool{
		"proficiencies": true,
		"equipment":     true,
		"details":       true,
	}
	return completedSteps[currentStep]
}

// getCompletedSteps returns which steps are completed based on the current step
func getCompletedSteps(currentStep string) map[string]bool {
	// Define the order of steps
	stepOrder := []string{"race", "class", "abilities", "class_features", "proficiencies", "equipment", "details"}

	// Find current step index
	currentIndex := -1
	for i, step := range stepOrder {
		if step == currentStep {
			currentIndex = i
			break
		}
	}

	// Mark all steps before current as completed
	completed := make(map[string]bool)
	for i, step := range stepOrder {
		completed[step] = i < currentIndex
	}

	// Special case: race is always completed once we're past race selection
	if currentStep != "" && currentStep != "race" {
		completed["race"] = true
	}

	// Special case: class is completed once we're at or past class step
	if currentStep != "" && currentStep != "race" {
		completed["class"] = true
	}

	return completed
}
