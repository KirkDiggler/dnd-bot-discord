package entities

import (
	"time"
)

// CharacterCreationSession tracks a character creation in progress
type CharacterCreationSession struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	GuildID     string    `json:"guild_id"`
	CharacterID string    `json:"character_id"`
	CurrentStep string    `json:"current_step"` // race, class, abilities, etc
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// CharacterCreationStep represents the current step in character creation
const (
	StepRaceSelection        = "race_selection"
	StepClassSelection       = "class_selection"
	StepAbilityScores        = "ability_scores"
	StepAbilityAssignment    = "ability_assignment"
	StepProficiencySelection = "proficiency_selection"
	StepEquipmentSelection   = "equipment_selection"
	StepCharacterDetails     = "character_details"
	StepComplete             = "complete"
)

// IsExpired checks if the session has expired
func (s *CharacterCreationSession) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// UpdateStep updates the current step and timestamp
func (s *CharacterCreationSession) UpdateStep(step string) {
	s.CurrentStep = step
	s.UpdatedAt = time.Now()
}