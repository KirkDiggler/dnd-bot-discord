package ability

import (
	"context"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
)

// Executor defines the interface for ability-specific execution logic
type Executor interface {
	// Key returns the ability key this executor handles
	Key() string

	// Execute processes the ability activation
	Execute(ctx context.Context, char *character.Character, ability *shared.ActiveAbility, input *UseAbilityInput) (*UseAbilityResult, error)
}

// ExecutorRegistry manages ability executors
type ExecutorRegistry struct {
	executors map[string]Executor
}

// NewExecutorRegistry creates a new executor registry
func NewExecutorRegistry() *ExecutorRegistry {
	return &ExecutorRegistry{
		executors: make(map[string]Executor),
	}
}

// Register adds a new executor to the registry
func (r *ExecutorRegistry) Register(executor Executor) {
	r.executors[executor.Key()] = executor
}

// Get returns the executor for a specific ability key
func (r *ExecutorRegistry) Get(abilityKey string) (Executor, bool) {
	executor, exists := r.executors[abilityKey]
	return executor, exists
}

// Has checks if an executor exists for the ability key
func (r *ExecutorRegistry) Has(abilityKey string) bool {
	_, exists := r.executors[abilityKey]
	return exists
}
