# ADR 004: Character Creation Step Handler Registry

Date: 2025-01-11

## Status

Proposed

## Context

During implementation of paginated spell selection, we discovered that the connection between character creation steps (defined by the flow service) and their Discord UI handlers is implicit and fragile. When adding new step types like `StepTypeCantripsSelection`, we must remember to:

1. Define the step type constant
2. Create the step in the flow builder
3. Add UI hints with action IDs
4. Register handlers for those action IDs
5. Hope everything connects properly

This implicit mapping has led to:
- Lost functionality (class preview disappeared)
- Runtime failures when handlers are missing
- Confusion about where rendering logic belongs
- Difficulty validating complete coverage

Additionally, this Discord bot serves as a prototype for the universal rpg-toolkit. The patterns we establish here will inform how different RPG systems integrate with various UI platforms.

## Decision

We will implement an explicit step handler registry that validates complete coverage at startup:

```go
// In internal/domain/rulebook/dnd5e/creation_steps.go
package dnd5e

var CreationStepTypes = []character.CreationStepType{
    character.StepTypeRaceSelection,
    character.StepTypeClassSelection,
    character.StepTypeAbilityScoreGeneration,
    character.StepTypeCantripsSelection,
    character.StepTypeSpellbookSelection,
    // ... all other D&D 5e steps
}

// In internal/discord/v2/handlers/character_creation_enhanced.go
type CharacterCreationHandler struct {
    stepHandlers map[character.CreationStepType]StepHandlerFunc
}

func NewCharacterCreationHandler(cfg *Config) (*CharacterCreationHandler, error) {
    h := &CharacterCreationHandler{
        stepHandlers: make(map[character.CreationStepType]StepHandlerFunc),
    }
    
    // Register all handlers explicitly
    h.stepHandlers[character.StepTypeRaceSelection] = h.handleRaceSelection
    h.stepHandlers[character.StepTypeClassSelection] = h.handleClassSelection
    h.stepHandlers[character.StepTypeCantripsSelection] = h.handleSpellSelection
    h.stepHandlers[character.StepTypeSpellbookSelection] = h.handleSpellSelection
    
    // Validate complete coverage
    for _, stepType := range dnd5e.CreationStepTypes {
        if _, ok := h.stepHandlers[stepType]; !ok {
            return nil, fmt.Errorf("no handler registered for step type: %s", stepType)
        }
    }
    
    return h, nil
}
```

## Consequences

### Positive

- **Fail Fast**: Missing handlers are caught at startup, not runtime
- **Explicit Contract**: Clear mapping between steps and handlers
- **Single Source of Truth**: D&D 5e defines its steps in one place
- **Clean Separation**: Rulebooks define WHAT, Discord defines HOW
- **Type Safety**: No string-based action IDs or interface{} casting
- **Toolkit Ready**: Pattern works for any rulebook/UI combination

### Negative

- **More Boilerplate**: Must explicitly register each handler
- **Compile-Time Coupling**: Discord handler must know all D&D 5e step types
- **No Dynamic Registration**: Can't add step types without recompiling

### Neutral

- **Different Pattern**: Moves away from action-based routing to type-based
- **Explicit Dependencies**: Makes rulebook → UI dependency visible

## Alternatives Considered

1. **Keep Current UI Hints System**: Continue with action IDs in UI hints
   - Rejected: Too fragile, no validation

2. **Protocol-Based Rendering**: Steps declare generic protocols
   - Rejected: Too abstract for current needs

3. **Generated Code**: Generate handlers from step definitions
   - Rejected: Adds build complexity

4. **Runtime Registration**: Handlers register themselves
   - Rejected: Can't validate complete coverage

## Implementation Notes

1. Start by extracting all step types to `dnd5e.CreationStepTypes`
2. Implement registry in character creation handler
3. Convert existing switch statements to use registry
4. Remove UI hints action routing in favor of direct step type handling
5. Add tests to ensure all step types have handlers

## Future Considerations

This pattern establishes clear boundaries that will benefit the rpg-toolkit:
- Other rulesets can define their own step types
- Other UI platforms can implement their own handler registries
- The contract between rulebook and UI becomes explicit and testable

When we extract patterns to the toolkit, this registry approach could become a standard interface that any UI adapter must implement.