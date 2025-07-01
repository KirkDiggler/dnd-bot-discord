package ability

import (
	"context"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
)

// Handler defines the interface for ability handlers
// This is what rulebook-specific abilities must implement
type Handler interface {
	// Key returns the unique identifier for this ability (e.g., "rage", "second-wind")
	Key() string

	// Execute performs the ability's effect
	Execute(ctx context.Context, char *character.Character, ability *shared.ActiveAbility, input *HandlerInput) (*HandlerResult, error)
}

// HandlerInput contains the input for executing an ability
type HandlerInput struct {
	CharacterID string
	AbilityKey  string
	TargetID    string // Optional: for targeted abilities
	Value       int    // Optional: for abilities that need a value (e.g., lay on hands healing amount)
	EncounterID string // Optional: for combat context
}

// HandlerResult contains the result of executing an ability
type HandlerResult struct {
	Success       bool
	Message       string
	UsesRemaining int

	// Optional fields based on ability type
	EffectApplied bool
	EffectID      string
	EffectName    string
	Duration      int
	HealingDone   int
	DamageBonus   int
	TargetNewHP   int
}
