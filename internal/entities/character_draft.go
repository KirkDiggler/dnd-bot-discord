package entities

import (
	"fmt"
	"time"
)

type CreateStep int

const (
	SelectRaceStep          CreateStep = 1 << 0  // 0000 0001
	SelectClassStep         CreateStep = 1 << 1  // 0000 0010
	SelectAbilityScoresStep CreateStep = 1 << 2  // 0000 0100
	SelectProficienciesStep CreateStep = 1 << 3  // 0000 1000
	SelectEquipmentStep     CreateStep = 1 << 4  // 0001 0000
	SelectFeaturesStep      CreateStep = 1 << 5  // 0010 0000
	EnterNameStep           CreateStep = 1 << 6  // 0100 0000
	SelectBackgroundStep    CreateStep = 1 << 7  // 1000 0000
	SelectAlignmentStep     CreateStep = 1 << 8  // 0001 0000 0000
	SelectSkillsStep        CreateStep = 1 << 9  // 0010 0000 0000
)

// StepDependencies defines which steps need to be reset when a step changes
var StepDependencies = map[CreateStep][]CreateStep{
	SelectRaceStep: {
		SelectProficienciesStep, // Race affects available proficiencies
		SelectAbilityScoresStep, // Race might give ability score bonuses
		SelectFeaturesStep,      // Race provides racial features
	},
	SelectClassStep: {
		SelectProficienciesStep, // Class affects available proficiencies
		SelectSkillsStep,        // Class affects skill choices
		SelectEquipmentStep,     // Class affects starting equipment
		SelectFeaturesStep,      // Class provides class features
	},
	SelectBackgroundStep: {
		SelectProficienciesStep, // Background affects proficiencies
		SelectSkillsStep,        // Background affects skills
		SelectEquipmentStep,     // Background affects equipment
		SelectFeaturesStep,      // Background may provide features
	},
}

// StepOrder defines the valid progression of steps
var StepOrder = []CreateStep{
	SelectRaceStep,          // 1. Choose race
	SelectClassStep,         // 2. Choose class
	SelectAbilityScoresStep, // 3. Assign ability scores
	SelectProficienciesStep, // 4. Select proficiencies (from race/class/background)
	SelectEquipmentStep,     // 5. Choose starting equipment
	SelectFeaturesStep,      // 6. Review/select features (if any choices)
	EnterNameStep,           // 7. Enter character name and finalize
	// Future steps:
	// SelectBackgroundStep,    // Background (not implemented yet)
	// SelectAlignmentStep,     // Alignment (not implemented yet)
	// SelectSkillsStep,        // Skills (partially handled by proficiencies)
}

type CharacterDraft struct {
	ID             string     `json:"id"`
	OwnerID        string     `json:"owner_id"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	CurrentStep    CreateStep `json:"current_step"`
	CompletedSteps CreateStep `json:"completed_steps"`
	Character      *Character `json:"character"`
}

func (d *CharacterDraft) IsStepCompleted(step CreateStep) bool {
	return d.CompletedSteps&step != 0
}

func (d *CharacterDraft) CompleteStep(step CreateStep) error {
	// Validate step can be completed based on current progress
	if err := d.canCompleteStep(step); err != nil {
		return err
	}

	d.CompletedSteps |= step
	return nil
}

func (d *CharacterDraft) UncompleteStep(step CreateStep) {
	// Clear the specific step
	d.CompletedSteps &^= step

	// Only clear dependent steps from StepDependencies
	if deps, ok := StepDependencies[step]; ok {
		for _, depStep := range deps {
			d.CompletedSteps &^= depStep
		}
	}
}

func (d *CharacterDraft) AllStepsCompleted() bool {
	// Only check currently implemented steps
	implementedSteps := SelectRaceStep | SelectClassStep | SelectAbilityScoresStep |
		SelectProficienciesStep | SelectEquipmentStep | SelectFeaturesStep | EnterNameStep
	return (d.CompletedSteps & implementedSteps) == implementedSteps
}

func (d *CharacterDraft) canCompleteStep(step CreateStep) error {
	// Find the index of the current step in StepOrder
	currentIdx := -1
	for i, s := range StepOrder {
		if s == step {
			currentIdx = i
			break
		}
	}

	if currentIdx == -1 {
		return fmt.Errorf("invalid step: %d", step)
	}

	// Check if all previous steps are completed
	for i := 0; i < currentIdx; i++ {
		if !d.IsStepCompleted(StepOrder[i]) {
			return fmt.Errorf("previous step %d must be completed first", StepOrder[i])
		}
	}

	return nil
}

func (d *CharacterDraft) NextIncompleteStep() CreateStep {
	for _, step := range StepOrder {
		if !d.IsStepCompleted(step) {
			return step
		}
	}
	return 0
}

// Validate checks if the current step's data is valid based on Character state
func (d *CharacterDraft) Validate() error {
	if d.Character == nil {
		return fmt.Errorf("character not initialized")
	}

	switch d.CurrentStep {
	case SelectRaceStep:
		if d.Character.Race == nil {
			return fmt.Errorf("race must be selected")
		}
	case SelectClassStep:
		if d.Character.Class == nil {
			return fmt.Errorf("class must be selected")
		}
	case EnterNameStep:
		if d.Character.Name == "" {
			return fmt.Errorf("name must be entered")
		}
		// Add validation for other steps...
	}

	return nil
}

func (d *CharacterDraft) ResetStep(step CreateStep) error {
	if !d.IsStepCompleted(step) {
		return fmt.Errorf("step %d is not completed", step)
	}

	// First, reset the specific step
	d.UncompleteStep(step)

	// Reset dependent data in Character based on step
	if d.Character != nil {
		switch step {
		case SelectRaceStep:
			d.Character.Race = nil
			d.Character.resetRacialTraits()
		case SelectClassStep:
			d.Character.Class = nil
			d.Character.resetClassFeatures()
		case SelectBackgroundStep:
			d.Character.resetBackground()
		}
	}

	// Reset dependent steps
	if deps, ok := StepDependencies[step]; ok {
		for _, depStep := range deps {
			d.UncompleteStep(depStep)
		}
	}

	return nil
}
