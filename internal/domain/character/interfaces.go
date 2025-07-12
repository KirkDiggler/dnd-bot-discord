package character

import (
	"context"
)

// Entity Interfaces - These allow character domain to work with any rulebook implementation

// Race represents a character race from any rulebook
type Race interface {
	GetID() string
	GetName() string
	GetAbilityScoreIncrease() map[string]int
	GetSize() string
	GetSpeed() int
}

// Class represents a character class from any rulebook
type Class interface {
	GetID() string
	GetName() string
	GetHitDice() int
	GetPrimaryAbility() string
	GetSavingThrows() []string
}

// Background represents a character background
type Background interface {
	GetID() string
	GetName() string
	GetSkillProficiencies() []string
}

// Proficiency represents any kind of proficiency (skill, weapon, armor, etc)
type Proficiency interface {
	GetCategory() string // "skill", "weapon", "armor", "tool", etc
	GetName() string
}

// Feature represents a character feature/trait
type Feature interface {
	GetID() string
	GetName() string
	GetDescription() string
	GetSource() string // e.g., "race", "class", "background"
}

// Spell represents a spell reference
type Spell interface {
	GetID() string
	GetName() string
	GetLevel() int
}

// Weapon represents a weapon
type Weapon interface {
	GetID() string
	GetName() string
	GetDamage() string
	GetDamageType() string
	GetProperties() []string
}

// Calculator Interfaces - These handle rulebook-specific calculations

// ProficiencyCalculator calculates proficiency bonus based on level
type ProficiencyCalculator interface {
	CalculateBonus(level int) int
}

// ArmorClassCalculator calculates AC for a character
type ArmorClassCalculator interface {
	Calculate(ctx context.Context, char *Character) (int, error)
}

// CombatCalculator handles attack and damage calculations
type CombatCalculator interface {
	CalculateAttackBonus(ctx context.Context, attacker *Character, weapon Weapon) (int, error)
	CalculateDamage(ctx context.Context, attacker *Character, weapon Weapon) (DamageRoll, error)
}

// DamageRoll represents the result of a damage calculation
type DamageRoll struct {
	DiceCount  int
	DiceSize   int
	Modifier   int
	DamageType string
}

// CharacterValidator validates character data according to rulebook rules
type CharacterValidator interface {
	ValidateAbilityScores(scores map[string]int) error
	ValidateLevel(level int) error
	ValidateCharacter(ctx context.Context, char *Character) error
}

// Creation Service Interfaces - These handle character creation flow

// CharacterCreationService manages the character creation process
type CharacterCreationService interface {
	// StartCreation begins a new character creation session
	StartCreation(ctx context.Context, userID, characterName string) (CreationSession, error)

	// GetSession retrieves an existing creation session
	GetSession(ctx context.Context, sessionID string) (CreationSession, error)

	// GetCurrentStep returns the current step in the creation process
	GetCurrentStep(session CreationSession) Step

	// CompleteStep processes data for the current step and advances to the next
	CompleteStep(ctx context.Context, session CreationSession, stepData interface{}) error

	// FinishCreation completes the creation process and returns the finished character
	FinishCreation(ctx context.Context, session CreationSession) (*Character, error)
}

// CreationSession represents an active character creation session
type CreationSession interface {
	GetID() string
	GetUserID() string
	GetCharacterName() string
	GetCurrentStepID() string
	GetCompletedSteps() []string
	GetStepData(stepID string) interface{}
	IsComplete() bool
}

// Step represents a step in the character creation process
type Step interface {
	GetID() string
	GetName() string
	GetDescription() string
	IsRequired() bool
	GetDependencies() []string // Other step IDs that must be completed first
}

// Repository Interfaces - These define data access patterns

// CharacterRepository defines character persistence operations
type CharacterRepository interface {
	Create(ctx context.Context, character *Character) error
	Get(ctx context.Context, id string) (*Character, error)
	Update(ctx context.Context, character *Character) error
	Delete(ctx context.Context, id string) error
	ListByUserID(ctx context.Context, userID string) ([]*Character, error)
}

// CreationSessionRepository defines creation session persistence
type CreationSessionRepository interface {
	Create(ctx context.Context, session CreationSession) error
	Get(ctx context.Context, id string) (CreationSession, error)
	GetByUserID(ctx context.Context, userID string) (CreationSession, error)
	Update(ctx context.Context, session CreationSession) error
	Delete(ctx context.Context, id string) error
}
