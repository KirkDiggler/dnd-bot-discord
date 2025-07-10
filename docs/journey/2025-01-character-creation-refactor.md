# Character Creation Flow Refactor Journey

## January 2025

### The Problem
We were trying to implement a clean character creation flow that could:
- Track current position in multi-step creation
- Support dynamic, class-specific steps
- Allow navigation (back/forward)
- Eventually become part of a reusable toolkit

### Initial Attempts
1. **v2 Flow Service**: Created complex handler registry with UI handlers embedded in service layer
   - Too coupled - service shouldn't know about UI
   - Overly complex with handler registration

2. **Simple Flow Service**: Tried to simplify but ended up with 3 different flow services
   - Too many versions, confusing architecture
   - Still had coupling issues

3. **FlowState on Character**: Added flow state directly to Character model
   - Mixed concerns - workflow state polluting domain entity
   - Character shouldn't know about creation workflow

### The Realization
User challenged the assumption: "Why not use the existing CharacterDraft?"

This led to the key insight:
- CharacterDraft already exists as a wrapper around Character
- It's the perfect place for workflow state
- Keeps Character pure as a domain entity
- Draft → Active is a clear lifecycle transition

### The Solution
```go
// CharacterDraft wraps a Character during creation
type CharacterDraft struct {
    ID          string
    OwnerID     string
    Character   *Character  // Pure domain entity
    FlowState   *FlowState  // Workflow state
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// FlowState is now simple and focused
type FlowState struct {
    CurrentStepID  string
    CompletedSteps []string
    AllSteps       []string
    StepData       map[string]interface{}
}
```

### Key Learnings
1. **Challenge architectural assumptions** - The best solution was hiding in plain sight
2. **Separation of concerns matters** - Workflow != Domain entity
3. **Simpler is better** - FlowState doesn't need complex features, just track position
4. **Think toolkit-first** - This pattern will work well for any game system

### Next Steps
1. Remove FlowState from Character model
2. Enhance CharacterDraft with FlowState
3. Update services to use CharacterDraft during creation
4. Keep Discord UI handlers separate and clean