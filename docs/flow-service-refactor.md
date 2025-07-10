# Flow Service Refactor Plan

## Current Pain Points

### 1. **Hardcoded Class Logic**
The flow builder has hardcoded switch statements for each class:
```go
switch char.Class.Key {
case "cleric":
case "fighter": 
case "ranger":
// Missing: wizard, rogue, barbarian, etc.
}
```

### 2. **Inconsistent Step Completion Checks**
- `StepTypeAbilityScores` checks `len(char.Attributes) > 0` but should check `AbilityRolls`
- `hasAssignedAbilities` has a weak check (just looks for any score > 8)
- Some steps always return `false` (e.g., `hasUserSelectedProficiencies`)

### 3. **Limited Step Types**
Missing critical step types:
- Spell selection (wizards, clerics, etc.)
- Cantrip selection
- Skill expertise (rogues)
- Sorcerous origin
- Warlock pact
- Subclass features

### 4. **No UI Hints/Options**
The flow service doesn't communicate available actions to the UI:
- Can't indicate "roll individually" vs "roll all at once"
- No way to specify which buttons should appear
- No contextual help or recommendations

### 5. **Rigid Linear Flow**
- Can't skip optional steps
- Can't go back and change earlier choices
- No branching paths based on choices
- No way to handle prerequisites

### 6. **Poor Separation of Concerns**
- Flow builder knows too much about D&D rules
- Step completion logic mixed with flow logic
- UI concerns bleeding into domain logic

## Proposed Architecture

### 1. **Step Registry Pattern**
```go
// Step definition is data, not code
type StepDefinition struct {
    ID          string
    Type        StepType
    Title       string
    Description string
    
    // When does this step apply?
    Conditions  []StepCondition
    
    // What makes it complete?
    Completion  CompletionChecker
    
    // What can the user do?
    Actions     []StepAction
    
    // Any special UI hints
    UIHints     UIHints
}

// Register steps at startup
func init() {
    registry.Register(StepDefinition{
        ID:   "wizard.spell_selection",
        Type: StepTypeSpellSelection,
        Title: "Choose Your Spells",
        Conditions: []StepCondition{
            HasClass("wizard"),
            HasCompletedStep(StepTypeAbilityScores),
        },
        Actions: []StepAction{
            {ID: "select_spells", Label: "Choose Spells", Primary: true},
            {ID: "random_spells", Label: "Random Selection"},
        },
    })
}
```

### 2. **Flexible Conditions System**
```go
type StepCondition interface {
    IsMet(ctx context.Context, char *Character) bool
}

// Composable conditions
type AndCondition struct {
    Conditions []StepCondition
}

type OrCondition struct {
    Conditions []StepCondition  
}

type HasClassCondition struct {
    ClassKeys []string
}

type HasFeatureCondition struct {
    FeatureKey string
}
```

### 3. **Smart Completion Checking**
```go
type CompletionChecker interface {
    IsComplete(ctx context.Context, char *Character) bool
    GetProgress(ctx context.Context, char *Character) Progress
}

type Progress struct {
    Current int
    Total   int
    Message string
}

// Example: Ability score completion
type AbilityScoreCompletion struct{}

func (a AbilityScoreCompletion) IsComplete(ctx context.Context, char *Character) bool {
    // Check if abilities are rolled AND assigned
    if len(char.AbilityRolls) < 6 {
        return false
    }
    
    // Check if all rolls are assigned
    assignedCount := 0
    for _, assignment := range char.AbilityAssignments {
        if assignment != "" {
            assignedCount++
        }
    }
    
    return assignedCount == 6
}
```

### 4. **Action System for UI**
```go
type StepAction struct {
    ID          string
    Label       string
    Style       ActionStyle // Primary, Secondary, Danger
    Icon        string
    
    // When is this action available?
    Condition   ActionCondition
    
    // What handler processes it?
    Handler     string
}

// The UI can query available actions
actions := flowService.GetAvailableActions(ctx, characterID, stepID)
```

### 5. **Flow Service Interface**
```go
type FlowService interface {
    // Get the current step(s) - could be multiple in parallel
    GetActiveSteps(ctx context.Context, characterID string) ([]Step, error)
    
    // Get all steps with completion status
    GetFlowProgress(ctx context.Context, characterID string) (*FlowProgress, error)
    
    // Get available actions for a step
    GetStepActions(ctx context.Context, characterID string, stepID string) ([]StepAction, error)
    
    // Process an action
    ProcessAction(ctx context.Context, characterID string, actionID string, data map[string]any) (*ActionResult, error)
    
    // Navigate flow
    CanGoBack(ctx context.Context, characterID string) bool
    GoBack(ctx context.Context, characterID string) (*Step, error)
}
```

### 6. **Benefits**

1. **Extensible**: Add new classes/features without touching core code
2. **Testable**: Each component is isolated and mockable
3. **Flexible**: Support complex flows with branches and prerequisites
4. **UI-Friendly**: Provides all info needed for rich UIs
5. **Maintainable**: Clear separation of concerns
6. **Data-Driven**: Most changes are configuration, not code

## Migration Strategy

### Phase 1: Add Missing Features (Quick Fixes)
1. Add wizard spell selection to current flow
2. Fix ability score completion checks
3. Add missing step types enum values

### Phase 2: Create New System Alongside Old
1. Build step registry system
2. Create condition and completion interfaces
3. Implement for one class (wizard) as proof of concept

### Phase 3: Gradual Migration
1. Migrate one class at a time to new system
2. Keep old system as fallback
3. Add feature flag to toggle between systems

### Phase 4: Full Cutover
1. Remove old flow builder
2. Delete legacy code
3. Optimize and refine new system

## Example: Wizard Flow in New System

```go
// Wizard spell selection step
registry.Register(StepDefinition{
    ID:          "wizard.spells",
    Type:        StepTypeSpellSelection,
    Title:       "Choose Your Spells",
    Description: "As a wizard, you begin with 6 spells in your spellbook.",
    
    Conditions: []StepCondition{
        HasClass("wizard"),
        HasAssignedAbilities(), // Need INT score
    },
    
    Completion: SpellSelectionCompletion{
        RequiredCount: 6,
        SpellList:     "wizard",
        Level:         1,
    },
    
    Actions: []StepAction{
        {
            ID:    "choose_spells",
            Label: "Select Spells",
            Style: Primary,
            Icon:  "📜",
        },
        {
            ID:    "suggested_spells",  
            Label: "Use Suggested",
            Style: Secondary,
            Icon:  "🎯",
        },
    },
    
    UIHints: UIHints{
        ShowProgress: true,
        Grouping:     "class-features",
        Color:        0x6B46C1, // Purple for arcane
    },
})
```

## Next Steps

1. **Review and Refine**: Get feedback on this architecture
2. **Prototype**: Build a minimal version to validate the design
3. **Test**: Ensure it handles edge cases better than current system
4. **Document**: Create developer guide for adding new steps
5. **Implement**: Start with Phase 1 quick fixes while building Phase 2