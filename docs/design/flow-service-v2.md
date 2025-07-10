# Flow Service V2 Design

## Overview

The current flow service has several critical limitations:
1. Hardcoded class logic with switch statements
2. Rigid linear flow with no branching
3. Poor separation between UI and domain logic
4. No way to communicate available actions to the UI
5. Inconsistent step completion checks

This design proposes a flexible, data-driven flow service that can handle complex character creation flows.

## Core Concepts

### 1. Step Registry Pattern

Instead of hardcoding steps, we use a registry pattern where steps are defined as data:

```go
type StepDefinition struct {
    ID          string
    Type        StepType
    Title       string
    Description string
    
    // Dynamic conditions
    When        StepCondition    // When should this step appear?
    Complete    CompletionCheck  // How do we know it's done?
    
    // UI configuration
    Actions     []ActionDef      // What can the user do?
    Layout      LayoutHints      // How should it be displayed?
}
```

### 2. Composable Conditions

Steps appear based on composable conditions:

```go
// Basic conditions
HasClass("wizard")
HasRace("elf")
HasLevel(GreaterThan(2))
HasFeature("divine_domain")

// Composite conditions
And(HasClass("wizard"), HasCompletedStep("ability_scores"))
Or(HasClass("cleric"), HasClass("paladin"))
Not(HasSubclass())
```

### 3. Action-Based UI

Instead of the UI guessing what buttons to show, steps declare their actions:

```go
Actions: []ActionDef{
    {
        ID:      "roll_all",
        Label:   "Roll All Scores",
        Style:   Primary,
        Icon:    "🎲",
        Handler: "ability_scores.roll_all",
    },
    {
        ID:      "roll_individual", 
        Label:   "Roll One at a Time",
        Style:   Secondary,
        Icon:    "🎯",
        Handler: "ability_scores.roll_individual",
    },
    {
        ID:      "standard_array",
        Label:   "Use Standard Array",
        Style:   Secondary,
        Icon:    "📊",
        Handler: "ability_scores.standard",
    },
}
```

### 4. Step Grouping

Steps can be grouped for better organization:

```go
type StepGroup struct {
    ID    string
    Title string
    Steps []StepDefinition
}

// Example groups
"core"          // Race, Class, Abilities
"class_features" // Domain, Fighting Style, etc.
"customization" // Skills, Languages, Equipment
"finalization"  // Name, Backstory
```

## Implementation Plan

### Phase 1: Core Infrastructure

#### 1.1 Create New Interfaces

```go
// internal/domain/character/flow_v2.go
type FlowServiceV2 interface {
    // Get current available steps (may be multiple)
    GetActiveSteps(ctx context.Context, charID string) ([]*Step, error)
    
    // Get complete flow with progress
    GetFlowState(ctx context.Context, charID string) (*FlowState, error)
    
    // Execute an action
    ExecuteAction(ctx context.Context, req ExecuteActionRequest) (*ExecuteActionResponse, error)
}

type Step struct {
    ID          string
    Type        StepType
    Title       string
    Description string
    Actions     []Action
    Progress    *Progress
    Data        map[string]interface{}
}

type FlowState struct {
    CharacterID string
    Groups      []StepGroupState
    ActiveSteps []string
    CanGoBack   bool
}
```

#### 1.2 Step Registry

```go
// internal/services/character/flow_v2/registry.go
type StepRegistry struct {
    steps map[string]*StepDefinition
    mu    sync.RWMutex
}

func (r *StepRegistry) Register(def *StepDefinition) error
func (r *StepRegistry) GetApplicableSteps(ctx context.Context, char *Character) ([]*StepDefinition, error)
```

#### 1.3 Condition System

```go
// internal/services/character/flow_v2/conditions.go
type Condition interface {
    Evaluate(ctx context.Context, char *Character) bool
    Description() string
}

// Implementations
type HasClassCondition struct {
    Classes []string
}

type HasCompletedStepCondition struct {
    StepID string
}

type CompositeCondition struct {
    Op         LogicalOp // AND, OR, NOT
    Conditions []Condition
}
```

### Phase 2: Step Definitions

#### 2.1 Core Steps

```go
// internal/services/character/flow_v2/steps/core.go
func RegisterCoreSteps(registry *StepRegistry) {
    // Race Selection
    registry.Register(&StepDefinition{
        ID:   "race_selection",
        Type: StepTypeRaceSelection,
        Title: "Choose Your Race",
        When: Always(), // No conditions
        Complete: HasRace(),
        Actions: []ActionDef{
            {ID: "select_race", Label: "Choose Race", Style: Primary},
            {ID: "random_race", Label: "Random", Style: Secondary},
        },
    })
    
    // Ability Scores (combined rolling and assignment)
    registry.Register(&StepDefinition{
        ID:   "ability_scores",
        Type: StepTypeAbilityScores,
        Title: "Determine Abilities",
        When: And(HasRace(), HasClass()),
        Complete: HasAssignedAllAbilities(),
        Actions: []ActionDef{
            {ID: "roll_all", Label: "Roll All", Style: Primary},
            {ID: "standard_array", Label: "Standard Array", Style: Secondary},
            {ID: "point_buy", Label: "Point Buy", Style: Secondary},
        },
        Layout: LayoutHints{
            ShowProgress: true,
            Substeps: []string{"roll", "assign"},
        },
    })
}
```

#### 2.2 Class-Specific Steps

```go
// internal/services/character/flow_v2/steps/wizard.go
func RegisterWizardSteps(registry *StepRegistry) {
    // Cantrip Selection
    registry.Register(&StepDefinition{
        ID:   "wizard_cantrips",
        Type: StepTypeCantrips,
        Title: "Choose Cantrips",
        When: And(
            HasClass("wizard"),
            HasAssignedAllAbilities(),
        ),
        Complete: HasSelectedCantrips(3),
        Actions: []ActionDef{
            {ID: "select_cantrips", Label: "Choose Cantrips", Style: Primary},
            {ID: "suggest", Label: "Suggested", Style: Secondary},
        },
    })
    
    // Spell Selection
    registry.Register(&StepDefinition{
        ID:   "wizard_spells",
        Type: StepTypeSpells,
        Title: "Select Spells",
        When: And(
            HasClass("wizard"),
            HasCompletedStep("wizard_cantrips"),
        ),
        Complete: HasSelectedSpells(6),
        Actions: []ActionDef{
            {ID: "select_spells", Label: "Choose Spells", Style: Primary},
            {ID: "quick_build", Label: "Quick Build", Style: Secondary},
        },
        Layout: LayoutHints{
            GroupBy: "school",
            Filter:  true,
        },
    })
}
```

### Phase 3: UI Integration

#### 3.1 Enhanced Discord Handler

```go
// internal/discord/v2/handlers/character_creation_v2.go
type CharacterCreationV2Handler struct {
    flowService FlowServiceV2
    uiBuilder   UIBuilder
}

func (h *CharacterCreationV2Handler) ShowCurrentStep(ctx *core.InteractionContext) (*core.HandlerResult, error) {
    // Get active steps
    steps, err := h.flowService.GetActiveSteps(ctx.Context, characterID)
    
    // Build UI from step definitions
    response := h.uiBuilder.BuildStepUI(steps[0])
    
    return &core.HandlerResult{Response: response}, nil
}
```

#### 3.2 Dynamic Component Builder

```go
// internal/discord/v2/builders/step_ui.go
type StepUIBuilder struct {
    customIDBuilder *core.CustomIDBuilder
}

func (b *StepUIBuilder) BuildStepUI(step *Step) *core.Response {
    embed := builders.NewEmbed().
        Title(step.Title).
        Description(step.Description)
    
    components := builders.NewComponentBuilder(b.customIDBuilder)
    
    // Add progress if applicable
    if step.Progress != nil {
        embed.AddField("Progress", 
            fmt.Sprintf("%d/%d", step.Progress.Current, step.Progress.Required),
            true)
    }
    
    // Build action buttons from step definition
    for i, action := range step.Actions {
        if i > 0 && i%5 == 0 {
            components.NewRow()
        }
        
        switch action.Style {
        case ActionStylePrimary:
            components.PrimaryButton(action.Label, action.ID, step.ID)
        case ActionStyleSecondary:
            components.SecondaryButton(action.Label, action.ID, step.ID)
        // etc...
        }
    }
    
    return core.NewResponse("").
        WithEmbeds(embed.Build()).
        WithComponents(components.Build()...)
}
```

### Phase 4: Migration Strategy

#### 4.1 Feature Flag

```go
// Start with feature flag
if features.IsEnabled("flow_v2") {
    return flowServiceV2.GetActiveSteps(ctx, charID)
} else {
    return oldFlowService.GetNextStep(ctx, charID)
}
```

#### 4.2 Gradual Migration

1. Implement V2 alongside V1
2. Migrate one class at a time
3. A/B test with subset of users
4. Monitor metrics and feedback
5. Full cutover when stable

## Benefits

### 1. Extensibility
- Add new classes/races without touching core code
- Support homebrew content easily
- Enable/disable features with conditions

### 2. Flexibility
- Non-linear flows (e.g., choose skills and languages in parallel)
- Conditional steps based on choices
- Different UI layouts per step

### 3. Testability
- Pure functions for conditions
- Mock registry for testing
- Isolated step definitions

### 4. Maintainability
- Clear separation of concerns
- Self-documenting step definitions
- Centralized step logic

## Example: Complete Wizard Flow

```go
// Registration
registry := NewStepRegistry()
RegisterCoreSteps(registry)
RegisterWizardSteps(registry)

// Runtime
char := &Character{
    Class: &Class{Key: "wizard"},
    Race: &Race{Key: "elf"},
    Level: 1,
}

// Get applicable steps
steps := registry.GetApplicableSteps(ctx, char)
// Returns: [wizard_cantrips, skill_selection, equipment_selection]

// User selects cantrips
flowService.ExecuteAction(ctx, ExecuteActionRequest{
    CharacterID: char.ID,
    StepID:      "wizard_cantrips",
    ActionID:    "select_cantrips",
    Data: map[string]interface{}{
        "selections": []string{"mage_hand", "prestidigitation", "fire_bolt"},
    },
})

// Get next steps
steps = registry.GetApplicableSteps(ctx, char)
// Returns: [wizard_spells, skill_selection, equipment_selection]
```

## Next Steps

1. Create proof of concept with single class (Wizard)
2. Benchmark performance vs current system
3. Gather feedback on API design
4. Build migration tooling
5. Document for other developers