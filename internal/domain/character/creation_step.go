package character

// CreationStepType represents the type of step in character creation
type CreationStepType string

const (
	StepTypeRaceSelection            CreationStepType = "race_selection"
	StepTypeClassSelection           CreationStepType = "class_selection"
	StepTypeAbilityScores            CreationStepType = "ability_scores"
	StepTypeAbilityAssignment        CreationStepType = "ability_assignment"
	StepTypeSkillSelection           CreationStepType = "skill_selection"
	StepTypeLanguageSelection        CreationStepType = "language_selection"
	StepTypeFightingStyleSelection   CreationStepType = "fighting_style_selection"
	StepTypeDivineDomainSelection    CreationStepType = "divine_domain_selection"
	StepTypeFavoredEnemySelection    CreationStepType = "favored_enemy_selection"
	StepTypeNaturalExplorerSelection CreationStepType = "natural_explorer_selection"
	StepTypeProficiencySelection     CreationStepType = "proficiency_selection"
	StepTypeEquipmentSelection       CreationStepType = "equipment_selection"
	StepTypeCharacterDetails         CreationStepType = "character_details"
	StepTypeComplete                 CreationStepType = "complete"
)

// CreationStep represents a single step in character creation
type CreationStep struct {
	Type        CreationStepType `json:"type"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Options     []CreationOption `json:"options,omitempty"`
	MinChoices  int              `json:"min_choices,omitempty"`
	MaxChoices  int              `json:"max_choices,omitempty"`
	Required    bool             `json:"required"`
	Context     map[string]any   `json:"context,omitempty"` // Additional context data
}

// CreationOption represents an option within a creation step
type CreationOption struct {
	Key         string         `json:"key"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// CreationStepResult represents the result of completing a step
type CreationStepResult struct {
	StepType   CreationStepType `json:"step_type"`
	Selections []string         `json:"selections"`
	Metadata   map[string]any   `json:"metadata,omitempty"`
}

// IsComplete returns true if the step has been completed
func (s *CreationStep) IsComplete(result *CreationStepResult) bool {
	if result == nil || result.StepType != s.Type {
		return false
	}

	if s.Required && len(result.Selections) == 0 {
		return false
	}

	if s.MinChoices > 0 && len(result.Selections) < s.MinChoices {
		return false
	}

	if s.MaxChoices > 0 && len(result.Selections) > s.MaxChoices {
		return false
	}

	return true
}
