package ability

import (
	"context"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
)

// Service defines the ability service interface
type Service interface {
	// UseAbility executes an ability
	UseAbility(ctx context.Context, input *UseAbilityInput) (*UseAbilityResult, error)

	// GetAvailableAbilities returns all abilities a character can currently use
	GetAvailableAbilities(ctx context.Context, characterID string) ([]*AvailableAbility, error)

	// ApplyAbilityEffects applies the effects of an ability
	ApplyAbilityEffects(ctx context.Context, input *ApplyEffectsInput) error
}

// UseAbilityInput contains data for using an ability
type UseAbilityInput struct {
	CharacterID string
	AbilityKey  string
	TargetID    string // For targeted abilities (can be self)
	EncounterID string // Optional, for combat abilities
	Value       int    // For abilities like Lay on Hands that use a resource pool
}

// UseAbilityResult contains the result of using an ability
type UseAbilityResult struct {
	Success       bool
	Message       string
	EffectApplied bool
	UsesRemaining int

	// For healing abilities
	HealingDone int
	TargetNewHP int

	// For buff abilities
	EffectID   string
	EffectName string
	Duration   int
}

// AvailableAbility represents an ability and whether it can be used
type AvailableAbility struct {
	Ability   *shared.ActiveAbility
	Available bool
	Reason    string // Why it's not available (e.g., "No uses remaining", "Wrong action type")
}

// ApplyEffectsInput contains data for applying ability effects
type ApplyEffectsInput struct {
	SourceID    string // Character using the ability
	TargetID    string // Target of the ability
	AbilityKey  string
	EncounterID string
}
