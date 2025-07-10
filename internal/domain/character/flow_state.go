package character

import (
	"time"
)

// FlowState tracks the character creation flow state for draft characters
type FlowState struct {
	CurrentStepID  string         `json:"current_step_id"`
	CompletedSteps []string       `json:"completed_steps"`
	AllSteps       []string       `json:"all_steps"`
	StepData       map[string]any `json:"step_data,omitempty"`
	LastUpdated    time.Time      `json:"last_updated"`
}

// IsStepCompleted checks if a step has been completed
func (fs *FlowState) IsStepCompleted(stepID string) bool {
	if fs == nil {
		return false
	}
	for _, completed := range fs.CompletedSteps {
		if completed == stepID {
			return true
		}
	}
	return false
}

// GetStepIndex returns the index of a step in AllSteps, or -1 if not found
func (fs *FlowState) GetStepIndex(stepID string) int {
	if fs == nil {
		return -1
	}
	for i, step := range fs.AllSteps {
		if step == stepID {
			return i
		}
	}
	return -1
}

// GetCurrentStepIndex returns the index of the current step
func (fs *FlowState) GetCurrentStepIndex() int {
	if fs == nil {
		return -1
	}
	return fs.GetStepIndex(fs.CurrentStepID)
}

// CanNavigateBack returns true if we can go to the previous step
func (fs *FlowState) CanNavigateBack() bool {
	if fs == nil {
		return false
	}
	currentIndex := fs.GetCurrentStepIndex()
	return currentIndex > 0
}

// CanNavigateForward returns true if we can go to the next step
func (fs *FlowState) CanNavigateForward() bool {
	if fs == nil {
		return false
	}
	currentIndex := fs.GetCurrentStepIndex()
	return currentIndex >= 0 && currentIndex < len(fs.AllSteps)-1
}
