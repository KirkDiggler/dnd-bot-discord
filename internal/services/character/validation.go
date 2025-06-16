package character

import (
	"fmt"
	"strings"
	
	dnderr "github.com/KirkDiggler/dnd-bot-discord/internal/errors"
)

// Validator interface for input validation
type Validator interface {
	Validate() error
}

// ValidateInput validates any input that implements Validator
func ValidateInput(input Validator) error {
	if input == nil {
		return dnderr.InvalidArgument("input cannot be nil")
	}
	return input.Validate()
}

// Validate checks CreateCharacterInput for validity
func (i *CreateCharacterInput) Validate() error {
	if i == nil {
		return dnderr.InvalidArgument("CreateCharacterInput cannot be nil")
	}

	if strings.TrimSpace(i.UserID) == "" {
		return dnderr.InvalidArgument("user ID is required")
	}

	if strings.TrimSpace(i.RealmID) == "" {
		return dnderr.InvalidArgument("realm ID is required")
	}

	if strings.TrimSpace(i.Name) == "" {
		return dnderr.InvalidArgument("character name is required")
	}

	if len(i.Name) > 50 {
		return dnderr.InvalidArgument("character name cannot exceed 50 characters")
	}

	if strings.TrimSpace(i.RaceKey) == "" {
		return dnderr.InvalidArgument("race is required")
	}

	if strings.TrimSpace(i.ClassKey) == "" {
		return dnderr.InvalidArgument("class is required")
	}

	if i.AbilityScores == nil || len(i.AbilityScores) == 0 {
		return dnderr.InvalidArgument("ability scores are required")
	}

	// Validate ability scores
	requiredAbilities := []string{"STR", "DEX", "CON", "INT", "WIS", "CHA"}
	for _, ability := range requiredAbilities {
		score, ok := i.AbilityScores[ability]
		if !ok {
			return dnderr.InvalidArgument(fmt.Sprintf("missing ability score for %s", ability))
		}
		if score < 3 || score > 18 {
			return dnderr.InvalidArgument(fmt.Sprintf("ability score for %s must be between 3 and 18, got %d", ability, score))
		}
	}

	return nil
}

// Validate checks ResolveChoicesInput for validity
func (i *ResolveChoicesInput) Validate() error {
	if i == nil {
		return dnderr.InvalidArgument("ResolveChoicesInput cannot be nil")
	}

	if strings.TrimSpace(i.RaceKey) == "" {
		return dnderr.InvalidArgument("race key is required")
	}

	if strings.TrimSpace(i.ClassKey) == "" {
		return dnderr.InvalidArgument("class key is required")
	}

	return nil
}

// Validate checks ValidateCharacterInput for validity
func (i *ValidateCharacterInput) Validate() error {
	if i == nil {
		return dnderr.InvalidArgument("ValidateCharacterInput cannot be nil")
	}

	if strings.TrimSpace(i.RaceKey) == "" {
		return dnderr.InvalidArgument("race is required")
	}

	if strings.TrimSpace(i.ClassKey) == "" {
		return dnderr.InvalidArgument("class is required")
	}

	if i.AbilityScores == nil || len(i.AbilityScores) == 0 {
		return dnderr.InvalidArgument("ability scores are required")
	}

	// Validate ability scores
	requiredAbilities := []string{"STR", "DEX", "CON", "INT", "WIS", "CHA"}
	for _, ability := range requiredAbilities {
		score, ok := i.AbilityScores[ability]
		if !ok {
			return dnderr.InvalidArgument(fmt.Sprintf("missing ability score for %s", ability))
		}
		if score < 3 || score > 18 {
			return dnderr.InvalidArgument(fmt.Sprintf("ability score for %s must be between 3 and 18, got %d", ability, score))
		}
	}

	return nil
}