# Flow Service V2 Implementation Plan

## Current State Analysis

### Problems to Solve
1. **Hardcoded Class Logic**: Switch statements for each class
2. **Missing Classes**: No wizard, barbarian, sorcerer, warlock steps
3. **Wrong Completion Checks**: Checking wrong fields (Attributes vs AbilityRolls)
4. **No UI Hints**: UI has to guess what buttons to show
5. **Linear Only**: Can't handle parallel steps or branching
6. **Poor Separation**: Business logic mixed with UI concerns

### What's Working
- Basic flow progression
- Race and class selection
- Equipment and proficiency selection (mostly)
- Character finalization

## Implementation Phases

### Phase 1: Minimal Viable Change (1-2 days)
**Goal**: Fix immediate issues without major refactor

#### Tasks:
1. **Add Missing Step Types**
   ```go
   // internal/domain/character/creation_step.go
   const (
       // Add these:
       StepTypeSpellSelection    CreationStepType = "spell_selection"
       StepTypeCantripsSelection CreationStepType = "cantrips_selection" 
       StepTypeSubclassSelection CreationStepType = "subclass_selection"
   )
   ```

2. **Fix Wizard Flow**
   ```go
   // internal/services/character/flow_builder_wizard.go
   func (b *FlowBuilderImpl) buildWizardSteps(ctx context.Context, char *character.Character) []character.CreationStep {
       return []character.CreationStep{
           {
               Type:        character.StepTypeCantripsSelection,
               Title:       "Choose Your Cantrips",
               Description: "Select 3 cantrips from the wizard spell list",
               Required:    true,
               MinChoices:  3,
               MaxChoices:  3,
           },
           {
               Type:        character.StepTypeSpellSelection,
               Title:       "Select Your Spells",
               Description: "Choose 6 1st-level spells for your spellbook",
               Required:    true,
               MinChoices:  6,
               MaxChoices:  6,
           },
       }
   }
   ```

3. **Add Step Metadata**
   ```go
   // Add to CreationStep
   type CreationStep struct {
       // ... existing fields
       UIHints *StepUIHints `json:"ui_hints,omitempty"`
   }
   
   type StepUIHints struct {
       Actions []StepAction `json:"actions"`
       Layout  string       `json:"layout"` // "default", "grid", "list"
       Color   int          `json:"color"`
   }
   
   type StepAction struct {
       ID    string `json:"id"`
       Label string `json:"label"`
       Style string `json:"style"` // "primary", "secondary", "danger"
       Icon  string `json:"icon"`
   }
   ```

### Phase 2: New Architecture Foundation (3-5 days)
**Goal**: Build V2 alongside V1 without breaking anything

#### Directory Structure:
```
internal/services/character/
├── flow_v2/
│   ├── service.go          # Main V2 service
│   ├── registry.go         # Step registry
│   ├── conditions.go       # Condition system
│   ├── completion.go       # Completion checkers
│   ├── steps/
│   │   ├── core.go        # Race, class, abilities
│   │   ├── wizard.go      # Wizard-specific
│   │   ├── cleric.go      # Cleric-specific
│   │   └── ...
│   └── tests/
├── flow_builder.go         # Current V1 (keep for now)
└── creation_flow_service.go # Current V1 (keep for now)
```

#### Core Components:

1. **Registry Implementation**
   ```go
   // internal/services/character/flow_v2/registry.go
   type Registry struct {
       steps map[string]*StepDefinition
       mu    sync.RWMutex
   }
   
   func NewRegistry() *Registry {
       r := &Registry{
           steps: make(map[string]*StepDefinition),
       }
       // Auto-register all steps
       registerAllSteps(r)
       return r
   }
   ```

2. **Condition System**
   ```go
   // internal/services/character/flow_v2/conditions.go
   type Condition interface {
       IsMet(ctx context.Context, char *character.Character) bool
   }
   
   // Common conditions
   func Always() Condition
   func HasClass(classes ...string) Condition
   func HasRace(races ...string) Condition
   func HasLevel(check func(int) bool) Condition
   func And(conditions ...Condition) Condition
   func Or(conditions ...Condition) Condition
   ```

3. **Service Implementation**
   ```go
   // internal/services/character/flow_v2/service.go
   type ServiceV2 struct {
       registry    *Registry
       charService character.Service
   }
   
   func (s *ServiceV2) GetActiveSteps(ctx context.Context, characterID string) ([]*Step, error) {
       char, err := s.charService.GetCharacter(ctx, characterID)
       if err != nil {
           return nil, err
       }
       
       // Get all registered steps
       allSteps := s.registry.GetAll()
       
       // Filter to applicable steps
       var activeSteps []*Step
       for _, def := range allSteps {
           if def.When.IsMet(ctx, char) && !def.Complete.IsMet(ctx, char) {
               activeSteps = append(activeSteps, s.buildStep(def, char))
           }
       }
       
       return activeSteps, nil
   }
   ```

### Phase 3: UI Integration (2-3 days)
**Goal**: Update Discord handlers to use V2 when available

1. **Update Handler to Check V2**
   ```go
   // internal/discord/v2/handlers/character_creation_enhanced.go
   func (h *CharacterCreationHandler) buildEnhancedStepResponse(char *domainCharacter.Character, step *domainCharacter.CreationStep) (*core.Response, error) {
       // Check if step has UI hints
       if step.UIHints != nil {
           return h.buildV2StepResponse(char, step)
       }
       
       // Fall back to current implementation
       // ... existing code
   }
   ```

2. **Dynamic Action Building**
   ```go
   func (h *CharacterCreationHandler) buildV2StepResponse(char *domainCharacter.Character, step *domainCharacter.CreationStep) (*core.Response, error) {
       embed := builders.NewEmbed().
           Title(step.Title).
           Description(step.Description)
       
       components := builders.NewComponentBuilder(h.customIDBuilder)
       
       // Build actions from UI hints
       for i, action := range step.UIHints.Actions {
           if i > 0 && i%5 == 0 {
               components.NewRow()
           }
           
           switch action.Style {
           case "primary":
               components.PrimaryButton(action.Label, action.ID, char.ID)
           case "secondary":
               components.SecondaryButton(action.Label, action.ID, char.ID)
           case "danger":
               components.DangerButton(action.Label, action.ID, char.ID)
           }
       }
       
       return core.NewResponse("").
           WithEmbeds(embed.Build()).
           WithComponents(components.Build()...), nil
   }
   ```

### Phase 4: Migration (1-2 days per class)
**Goal**: Migrate each class to V2

1. **Start with Wizard** (simplest spell caster)
2. **Then Cleric** (has existing domain selection)
3. **Fighter** (simple, good test case)
4. **Continue with others**

### Phase 5: Cleanup (2-3 days)
**Goal**: Remove V1 code and optimize

1. Remove old flow builder
2. Remove old completion checks
3. Optimize registry lookups
4. Add caching where appropriate

## Success Metrics

### Functionality
- [ ] All classes have appropriate steps
- [ ] Wizard spell selection works
- [ ] Steps appear/disappear based on conditions
- [ ] UI shows appropriate actions per step

### Code Quality
- [ ] No hardcoded switch statements for classes
- [ ] Clear separation between domain and UI
- [ ] Easy to add new classes/races
- [ ] Comprehensive test coverage

### Performance
- [ ] Step calculation < 50ms
- [ ] No memory leaks from registry
- [ ] Efficient condition evaluation

## Risk Mitigation

### Feature Flag
```go
// internal/features/flags.go
var FlowV2Enabled = os.Getenv("FLOW_V2_ENABLED") == "true"

// Usage
if features.FlowV2Enabled {
    return flowV2.GetActiveSteps(ctx, charID)
}
return flowV1.GetNextStep(ctx, charID)
```

### Rollback Plan
1. Keep V1 code until V2 is proven
2. Log metrics comparing V1 vs V2
3. A/B test with small percentage
4. Have quick disable switch

## Development Order

1. **Week 1**: Phase 1 (Quick fixes) + Start Phase 2
2. **Week 2**: Complete Phase 2 + Phase 3
3. **Week 3**: Phase 4 (Migrate 3-4 classes)
4. **Week 4**: Complete Phase 4 + Phase 5

Total estimate: ~4 weeks with testing and iterations