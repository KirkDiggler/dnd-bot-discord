package entities

import (
	"time"
)

type CreateStep int

const (
	SelectRaceStep          CreateStep = 1 << iota
	SelectClassStep
	EnterNameStep
	SelectBackgroundStep
	SelectAlignmentStep
	SelectAbilityScoresStep
	SelectSkillsStep
	SelectEquipmentStep
	SelectProficienciesStep
)

type CharacterDraft struct {
	ID             string    `json:"id"`
	OwnerID        string    `json:"owner_id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	CurrentStep    CreateStep `json:"current_step"`
	CompletedSteps CreateStep `json:"completed_steps"`
	Character      *Character `json:"character"`
}

func (d *CharacterDraft) IsStepCompleted(step CreateStep) bool {
	return d.CompletedSteps&step != 0
}

func (d *CharacterDraft) CompleteStep(step CreateStep) {
	d.CompletedSteps |= step
}

func (d *CharacterDraft) UncompleteStep(step CreateStep) {
	d.CompletedSteps &^= step
}

func (d *CharacterDraft) AllStepsCompleted() bool {
	allSteps := SelectRaceStep | SelectClassStep | EnterNameStep | SelectBackgroundStep |
				SelectAlignmentStep | SelectAbilityScoresStep | SelectSkillsStep |
				SelectEquipmentStep | SelectProficienciesStep
	return d.CompletedSteps == allSteps
}
