package character

import (
	"fmt"
	"strings"
)

// Validator interface for input validation
type Validator interface {
	Validate() error
}

// ValidateInput validates any input that implements Validator
func ValidateInput(input Validator) error {
	if input == nil {
		return fmt.Errorf("input cannot be nil")
	}
	return input.Validate()
}

// Validate checks CreateCharacterInput for validity
func (i *CreateCharacterInput) Validate() error {
	if i == nil {
		return fmt.Errorf("CreateCharacterInput cannot be nil")
	}

	if strings.TrimSpace(i.UserID) == "" {
		return fmt.Errorf("user ID is required")
	}

	if strings.TrimSpace(i.RealmID) == "" {
		return fmt.Errorf("realm ID is required")
	}

	if strings.TrimSpace(i.Name) == "" {
		return fmt.Errorf("character name is required")
	}

	if len(i.Name) > 50 {
		return fmt.Errorf("character name cannot exceed 50 characters")
	}

	if strings.TrimSpace(i.RaceKey) == "" {
		return fmt.Errorf("race is required")
	}

	if strings.TrimSpace(i.ClassKey) == "" {
		return fmt.Errorf("class is required")
	}

	if i.AbilityScores == nil || len(i.AbilityScores) == 0 {
		return fmt.Errorf("ability scores are required")
	}

	// Validate ability scores
	requiredAbilities := []string{"STR", "DEX", "CON", "INT", "WIS", "CHA"}
	for _, ability := range requiredAbilities {
		score, ok := i.AbilityScores[ability]
		if !ok {
			return fmt.Errorf("missing ability score for %s", ability)
		}
		if score < 3 || score > 18 {
			return fmt.Errorf("ability score for %s must be between 3 and 18, got %d", ability, score)
		}
	}

	return nil
}

// Validate checks ResolveChoicesInput for validity
func (i *ResolveChoicesInput) Validate() error {
	if i == nil {
		return fmt.Errorf("ResolveChoicesInput cannot be nil")
	}

	if strings.TrimSpace(i.RaceKey) == "" {
		return fmt.Errorf("race key is required")
	}

	if strings.TrimSpace(i.ClassKey) == "" {
		return fmt.Errorf("class key is required")
	}

	return nil
}

// Validate checks ValidateCharacterInput for validity
func (i *ValidateCharacterInput) Validate() error {
	if i == nil {
		return fmt.Errorf("ValidateCharacterInput cannot be nil")
	}

	if strings.TrimSpace(i.RaceKey) == "" {
		return fmt.Errorf("race is required")
	}

	if strings.TrimSpace(i.ClassKey) == "" {
		return fmt.Errorf("class is required")
	}

	if i.AbilityScores == nil || len(i.AbilityScores) == 0 {
		return fmt.Errorf("ability scores are required")
	}

	// Validate ability scores
	requiredAbilities := []string{"STR", "DEX", "CON", "INT", "WIS", "CHA"}
	for _, ability := range requiredAbilities {
		score, ok := i.AbilityScores[ability]
		if !ok {
			return fmt.Errorf("missing ability score for %s", ability)
		}
		if score < 3 || score > 18 {
			return fmt.Errorf("ability score for %s must be between 3 and 18, got %d", ability, score)
		}
	}

	return nil
}