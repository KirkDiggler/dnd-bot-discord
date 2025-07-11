# Character Creation Architecture Journey - January 11, 2025

## Context

While implementing spell selection with pagination, we discovered a deeper architectural issue: the connection between the flow service (which defines character creation steps) and the Discord handlers (which render those steps) is implicit and fragile.

### The Problem

1. **Lost Class Preview** - While adding spell selection, we lost the class preview functionality
2. **Implicit Handler Mapping** - UI hints contain action IDs that "hope" a handler exists
3. **No Validation** - Can't verify all steps have handlers until runtime failures
4. **Mixed Concerns** - Flow service starting to contain rendering logic

## The Realization

> "we are adding rendering logic in our flow service. its job is just to say the steps. the discord rendering just needs to decide what select race render looks like"

The flow service should be pure domain logic: "these are the D&D 5e character creation steps." The Discord layer should own ALL rendering decisions.

## Brainstorming Session

### Option 1: Step Type Registry Pattern (Pragmatic Winner ✅)
```go
// Each rulebook declares its step types
var DnD5eStepTypes = []StepType{
    StepTypeRaceSelection,
    StepTypeClassSelection,
    StepTypeAbilityScoreGeneration,
    StepTypeCantripsSelection,
    StepTypeSpellbookSelection,
}

// Discord handler validates coverage at startup
type CharacterCreationHandler struct {
    stepHandlers map[StepType]StepHandlerFunc
}

func NewCharacterCreationHandler(cfg *Config) (*CharacterCreationHandler, error) {
    h := &CharacterCreationHandler{
        stepHandlers: make(map[StepType]StepHandlerFunc),
    }
    
    // Register all handlers explicitly
    h.stepHandlers[StepTypeRaceSelection] = h.handleRaceSelection
    h.stepHandlers[StepTypeClassSelection] = h.handleClassSelection
    h.stepHandlers[StepTypeCantripsSelection] = h.handleSpellSelection
    
    // Validate we have handlers for all step types
    for _, stepType := range DnD5eStepTypes {
        if _, ok := h.stepHandlers[stepType]; !ok {
            return nil, fmt.Errorf("no handler registered for step type: %s", stepType)
        }
    }
    
    return h, nil
}
```

### Option 2: Protocol-Based Rendering (Too Meta ❌)
```go
// Steps declare what protocol they use
type CreationStep struct {
    Type     StepType
    Protocol RenderProtocol  // "select_one", "select_many", "roll_dice"
}
```
Rejected: Too abstract, adds unnecessary indirection

### Option 3: Capability-Based Discovery (Overengineered ❌)
```go
type StepCapabilities struct {
    HasOptions      bool
    RequiresDice    bool
    AllowsPreview   bool
}
```
Rejected: Trying to be too generic for our needs

### Option 4: Event-Driven Registration (Overkill ❌)
```go
eventBus.Emit("rulebook.registered", RulebookRegistered{
    Rulebook: "dnd5e",
    StepTypes: []StepType{...},
})
```
Rejected: We're not building a multi-rulebook system (yet)

### Option 5: Generated Code (The Nuclear Option ☢️)
```go
//go:generate go run gen_step_handlers.go
```
Rejected: Adds build complexity for marginal benefit

### Option 6: Step Contracts with Type Assertions (Please No 🤮)
```go
if contract, ok := step.(RaceSelectionContract); ok {
    if contract.SupportsPreview() {
        // Add preview UI
    }
}
```
Rejected: Runtime type checking = bad, prefer generics

## Key Insights

1. **Separation of Concerns**: Rulebooks define WHAT, Discord defines HOW
2. **Explicit > Implicit**: Better to have explicit handler registration than magic
3. **Validate Early**: Fail at startup if handlers are missing, not at runtime
4. **Keep It Simple**: We're building a D&D 5e bot, not a universal RPG framework

## The Path Forward

Implement Option 1 with these benefits:
- **Compile-time safety** - Can't forget to add a handler
- **Clear separation** - Flow service defines steps, Discord defines rendering
- **No magic** - Explicit registration, easy to debug
- **Type safe** - No interface{} or type assertions

## Future Dreams (label:dreams)

> "we may have a need for a generic layer to hand off to other areas but the discord is like the game we are not trying to make a single bot that you can choose different rulebooks .... -.- i mean sounds cool when you type it out loud"

Maybe someday: Call of Cthulhu, Gamma World, or other systems. But not today.

## Next Steps

1. Implement step handler registry pattern
2. Fix the missing class preview
3. Ensure all step types have registered handlers
4. Add startup validation to catch missing handlers early

## Lesson Learned

When the architecture feels wrong, stop and think. Sometimes a late-night brainstorming session can prevent days of technical debt.