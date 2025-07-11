# ADR-004: Character Creation Flow State Management

## Status
Accepted

## Context
We need to track the state of character creation flow, including:
- Current step in the creation process
- Completed steps
- Dynamic, class-specific steps (e.g., wizard spell selection, fighter combat styles)
- Step dependencies and validation
- Ability to navigate back/forward through steps

Initially, we added a `FlowState` field directly to the `Character` model. However, this mixes workflow concerns with the domain entity.

Furthermore, we realized that character creation flow is inherently a **rulebook-specific concept**. Different game systems have different creation processes, and even within a system, different campaigns might want custom flows.

## Decision
We will treat `CharacterDraft` as a rulebook-specific concept with configurable flow steps.

### Architecture:
```go
// In rulebook/dnd5e package
type CharacterDraft struct {
    ID          string
    OwnerID     string
    Character   *Character  // Generic character entity
    FlowState   *FlowState  // Rulebook-specific workflow
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type FlowState struct {
    CurrentStepID  string
    CompletedSteps []string
    AllSteps       []string  // Configured dynamically
    StepData       map[string]interface{}
}

// Configurable flow definition
type FlowConfig struct {
    Steps []StepConfig `json:"steps"`
}
```

## Consequences

### Positive
- **Clean separation of concerns**: Character remains a pure domain entity
- **Rulebook encapsulation**: Creation rules live where they belong
- **Maximum flexibility**: Flows can be configured without code changes
- **Multi-system support**: Each rulebook defines its own creation process
- **House rule friendly**: DMs can customize flows per campaign
- **Toolkit ready**: Clean pattern for any game system

### Negative
- **Additional complexity**: Configuration layer adds complexity
- **Migration effort**: Moving from current architecture requires work
- **Validation complexity**: Need to validate custom flow configurations

## Implementation Plan
1. Remove FlowState from Character model
2. Move CharacterDraft to rulebook/dnd5e package
3. Create FlowConfig structure for configurable steps
4. Implement default D&D 5e flow configuration
5. Add flow customization per realm/campaign
6. Update repositories to handle rulebook-specific drafts
7. Update services to use rulebook's creation flow
8. Implement CharacterDraft → Character finalization

## Notes
This architecture enables:
- Different game systems to have completely different creation flows
- House rules and campaign variants within a system
- Runtime configuration without recompilation
- Clean separation between generic characters and rulebook-specific creation

The key insight: character creation is not a generic concept - it's inherently tied to the rules of the game being played.