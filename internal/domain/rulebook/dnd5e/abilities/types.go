package abilities

import (
	"context"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
)

// Input contains the input for executing an ability
type Input struct {
	CharacterID string
	AbilityKey  string
	TargetID    string // Optional: for targeted abilities
	Value       int    // Optional: for abilities that need a value (e.g., lay on hands healing amount)
	EncounterID string // Optional: for combat context
}

// Result contains the result of executing an ability
type Result struct {
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

// Handler is the interface that D&D 5e abilities implement
type Handler interface {
	Key() string
	Execute(ctx context.Context, char *character.Character, ability *shared.ActiveAbility, input *Input) (*Result, error)
}
