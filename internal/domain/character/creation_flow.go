package character

import (
	"context"
)

// CreationFlowService defines the interface for managing character creation flow
type CreationFlowService interface {
	// GetNextStep returns the next step in character creation for the given character
	GetNextStep(ctx context.Context, characterID string) (*CreationStep, error)

	// ProcessStepResult processes the result of a completed step and returns the next step
	ProcessStepResult(ctx context.Context, characterID string, result *CreationStepResult) (*CreationStep, error)

	// GetCurrentStep returns the current step for a character in creation
	GetCurrentStep(ctx context.Context, characterID string) (*CreationStep, error)

	// IsCreationComplete returns true if character creation is complete
	IsCreationComplete(ctx context.Context, characterID string) (bool, error)

	// GetProgressSteps returns all steps with their completion status
	GetProgressSteps(ctx context.Context, characterID string) ([]ProgressStepInfo, error)

	// PreviewStepResult creates a preview of what the character would look like with the given selection
	// without actually applying the changes
	PreviewStepResult(ctx context.Context, characterID string, result *CreationStepResult) (*Character, error)
}

// ProgressStepInfo represents a step in the progress display
type ProgressStepInfo struct {
	Step      CreationStep `json:"step"`
	Completed bool         `json:"completed"`
	Current   bool         `json:"current"`
}

// CreationFlow represents the complete flow for character creation
type CreationFlow struct {
	Steps []CreationStep `json:"steps"`
}

// FlowBuilder builds character creation flows based on character state
type FlowBuilder interface {
	// BuildFlow creates a complete flow for the character based on race/class selections
	BuildFlow(ctx context.Context, char *Character) (*CreationFlow, error)
}
