package character_test

import (
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/stretchr/testify/suite"
)

// ValidationTestSuite tests input validation
type ValidationTestSuite struct {
	suite.Suite
}

func TestValidationSuite(t *testing.T) {
	suite.Run(t, new(ValidationTestSuite))
}

// CreateCharacterInput Validation Tests

func (s *ValidationTestSuite) TestCreateCharacterInput_Validate_NilInput() {
	var input *character.CreateCharacterInput
	err := character.ValidateInput(input)
	
	s.Error(err)
	s.Contains(err.Error(), "input cannot be nil")
}

func (s *ValidationTestSuite) TestCreateCharacterInput_Validate_NilReceiver() {
	var input *character.CreateCharacterInput
	err := input.Validate()
	
	s.Error(err)
	s.Contains(err.Error(), "CreateCharacterInput cannot be nil")
}

func (s *ValidationTestSuite) TestCreateCharacterInput_Validate_MissingUserID() {
	input := &character.CreateCharacterInput{
		RealmID:  "realm123",
		Name:     "Thorin",
		RaceKey:  "dwarf",
		ClassKey: "fighter",
		AbilityScores: map[string]int{
			"STR": 15, "DEX": 14, "CON": 13,
			"INT": 12, "WIS": 10, "CHA": 8,
		},
	}
	
	err := input.Validate()
	s.Error(err)
	s.Contains(err.Error(), "user ID is required")
}

func (s *ValidationTestSuite) TestCreateCharacterInput_Validate_EmptyName() {
	input := &character.CreateCharacterInput{
		UserID:   "user123",
		RealmID:  "realm123",
		Name:     "   ", // Only whitespace
		RaceKey:  "dwarf",
		ClassKey: "fighter",
		AbilityScores: map[string]int{
			"STR": 15, "DEX": 14, "CON": 13,
			"INT": 12, "WIS": 10, "CHA": 8,
		},
	}
	
	err := input.Validate()
	s.Error(err)
	s.Contains(err.Error(), "character name is required")
}

func (s *ValidationTestSuite) TestCreateCharacterInput_Validate_NameTooLong() {
	input := &character.CreateCharacterInput{
		UserID:   "user123",
		RealmID:  "realm123",
		Name:     "ThisIsAVeryLongCharacterNameThatExceedsTheFiftyCharacterLimit",
		RaceKey:  "dwarf",
		ClassKey: "fighter",
		AbilityScores: map[string]int{
			"STR": 15, "DEX": 14, "CON": 13,
			"INT": 12, "WIS": 10, "CHA": 8,
		},
	}
	
	err := input.Validate()
	s.Error(err)
	s.Contains(err.Error(), "character name cannot exceed 50 characters")
}

func (s *ValidationTestSuite) TestCreateCharacterInput_Validate_MissingAbilityScore() {
	input := &character.CreateCharacterInput{
		UserID:   "user123",
		RealmID:  "realm123",
		Name:     "Thorin",
		RaceKey:  "dwarf",
		ClassKey: "fighter",
		AbilityScores: map[string]int{
			"STR": 15, "DEX": 14, "CON": 13,
			"INT": 12, "WIS": 10,
			// Missing CHA
		},
	}
	
	err := input.Validate()
	s.Error(err)
	s.Contains(err.Error(), "missing ability score for CHA")
}

func (s *ValidationTestSuite) TestCreateCharacterInput_Validate_ScoreTooLow() {
	input := &character.CreateCharacterInput{
		UserID:   "user123",
		RealmID:  "realm123",
		Name:     "Thorin",
		RaceKey:  "dwarf",
		ClassKey: "fighter",
		AbilityScores: map[string]int{
			"STR": 2,  // Too low!
			"DEX": 14, "CON": 13,
			"INT": 12, "WIS": 10, "CHA": 8,
		},
	}
	
	err := input.Validate()
	s.Error(err)
	s.Contains(err.Error(), "ability score for STR must be between 3 and 18, got 2")
}

func (s *ValidationTestSuite) TestCreateCharacterInput_Validate_ScoreTooHigh() {
	input := &character.CreateCharacterInput{
		UserID:   "user123",
		RealmID:  "realm123",
		Name:     "Thorin",
		RaceKey:  "dwarf",
		ClassKey: "fighter",
		AbilityScores: map[string]int{
			"STR": 19, // Too high!
			"DEX": 14, "CON": 13,
			"INT": 12, "WIS": 10, "CHA": 8,
		},
	}
	
	err := input.Validate()
	s.Error(err)
	s.Contains(err.Error(), "ability score for STR must be between 3 and 18, got 19")
}

func (s *ValidationTestSuite) TestCreateCharacterInput_Validate_Success() {
	input := &character.CreateCharacterInput{
		UserID:   "user123",
		RealmID:  "realm123",
		Name:     "Thorin Oakenshield",
		RaceKey:  "dwarf",
		ClassKey: "fighter",
		AbilityScores: map[string]int{
			"STR": 16, "DEX": 12, "CON": 15,
			"INT": 10, "WIS": 13, "CHA": 8,
		},
		Proficiencies: []string{"skill-athletics", "skill-intimidation"},
		Equipment:     []string{"chain-mail", "longsword"},
	}
	
	err := input.Validate()
	s.NoError(err)
}

// ResolveChoicesInput Validation Tests

func (s *ValidationTestSuite) TestResolveChoicesInput_Validate_NilReceiver() {
	var input *character.ResolveChoicesInput
	err := input.Validate()
	
	s.Error(err)
	s.Contains(err.Error(), "ResolveChoicesInput cannot be nil")
}

func (s *ValidationTestSuite) TestResolveChoicesInput_Validate_EmptyRaceKey() {
	input := &character.ResolveChoicesInput{
		RaceKey:  "",
		ClassKey: "fighter",
	}
	
	err := input.Validate()
	s.Error(err)
	s.Contains(err.Error(), "race key is required")
}

func (s *ValidationTestSuite) TestResolveChoicesInput_Validate_EmptyClassKey() {
	input := &character.ResolveChoicesInput{
		RaceKey:  "human",
		ClassKey: "",
	}
	
	err := input.Validate()
	s.Error(err)
	s.Contains(err.Error(), "class key is required")
}

func (s *ValidationTestSuite) TestResolveChoicesInput_Validate_Success() {
	input := &character.ResolveChoicesInput{
		RaceKey:  "human",
		ClassKey: "fighter",
	}
	
	err := input.Validate()
	s.NoError(err)
}

// ValidateCharacterInput Validation Tests

func (s *ValidationTestSuite) TestValidateCharacterInput_Validate_NilAbilityScores() {
	input := &character.ValidateCharacterInput{
		RaceKey:       "human",
		ClassKey:      "fighter",
		AbilityScores: nil,
	}
	
	err := input.Validate()
	s.Error(err)
	s.Contains(err.Error(), "ability scores are required")
}

func (s *ValidationTestSuite) TestValidateCharacterInput_Validate_EmptyAbilityScores() {
	input := &character.ValidateCharacterInput{
		RaceKey:       "human",
		ClassKey:      "fighter",
		AbilityScores: map[string]int{},
	}
	
	err := input.Validate()
	s.Error(err)
	s.Contains(err.Error(), "ability scores are required")
}

func (s *ValidationTestSuite) TestValidateCharacterInput_Validate_Success() {
	input := &character.ValidateCharacterInput{
		RaceKey:  "human",
		ClassKey: "fighter",
		AbilityScores: map[string]int{
			"STR": 15, "DEX": 14, "CON": 13,
			"INT": 12, "WIS": 10, "CHA": 8,
		},
		Proficiencies: []string{"skill-athletics"},
	}
	
	err := input.Validate()
	s.NoError(err)
}