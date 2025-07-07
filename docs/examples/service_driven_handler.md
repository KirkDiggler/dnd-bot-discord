# Service-Driven Character Creation Handler Example

This document shows how handlers will be simplified using the new service-driven architecture.

## Before (Complex Handler Logic)

```go
// Old approach - handler determines flow
func (h *CharacterCreationHandler) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) error {
    char, err := h.characterService.GetByID(characterID)
    if err != nil {
        return err
    }
    
    // Complex logic to determine next step
    if char.Race == nil {
        return h.showRaceSelection(s, i)
    } else if char.Class == nil {
        return h.showClassSelection(s, i)
    } else if char.Class.Key == "cleric" && !hasSelectedDomain(char) {
        return h.showDivineDomainSelection(s, i)
    } else if char.Class.Key == "cleric" && hasSelectedDomain(char, "knowledge") && !hasSelectedKnowledgeSkills(char) {
        return h.showKnowledgeSkillSelection(s, i)
    }
    // ... many more conditions
}
```

## After (Simple Service-Driven)

```go
// New approach - service determines flow
func (h *CharacterCreationHandler) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) error {
    // Service tells us what step to show
    step, err := h.creationFlowService.GetNextStep(ctx, characterID)
    if err != nil {
        return err
    }
    
    // Handler just renders the step
    switch step.Type {
    case character.StepTypeRaceSelection:
        return h.renderStep(s, i, step, h.createRaceComponents)
    case character.StepTypeSkillSelection:
        return h.renderStep(s, i, step, h.createSkillComponents)
    case character.StepTypeLanguageSelection:
        return h.renderStep(s, i, step, h.createLanguageComponents)
    case character.StepTypeDivineDomainSelection:
        return h.renderStep(s, i, step, h.createDomainComponents)
    default:
        return h.renderGenericStep(s, i, step)
    }
}

// Generic step renderer
func (h *CharacterCreationHandler) renderStep(s *discordgo.Session, i *discordgo.InteractionCreate, 
    step *character.CreationStep, componentBuilder func(*character.CreationStep) []discordgo.MessageComponent) error {
    
    embed := &discordgo.MessageEmbed{
        Title:       step.Title,
        Description: step.Description,
        Color:       0x3498db,
    }
    
    components := componentBuilder(step)
    
    return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseUpdateMessage,
        Data: &discordgo.InteractionResponseData{
            Embeds:     []*discordgo.MessageEmbed{embed},
            Components: components,
            Flags:      discordgo.MessageFlagsEphemeral,
        },
    })
}
```

## Processing Step Results

```go
// Old approach - scattered logic
func (h *DomainSelectionHandler) Handle(req *SelectionRequest) error {
    // Find domain feature, update metadata, save character...
    // Complex logic spread across multiple handlers
}

// New approach - service handles the logic
func (h *CharacterCreationHandler) ProcessSelection(ctx context.Context, characterID string, selections []string) error {
    // Create step result
    result := &character.CreationStepResult{
        StepType:   h.getCurrentStepType(ctx, characterID), // or passed from UI
        Selections: selections,
    }
    
    // Service processes the result and returns next step
    nextStep, err := h.creationFlowService.ProcessStepResult(ctx, characterID, result)
    if err != nil {
        return err
    }
    
    // Show next step
    return h.showStep(nextStep)
}
```

## Benefits

1. **Handlers become simple renderers** - No business logic
2. **Service handles all flow logic** - Single source of truth
3. **Easy to add new steps** - Just add to service, handlers automatically work
4. **Testable flow logic** - Can test without Discord dependencies
5. **Consistent behavior** - All creation paths use same service

## Example: Adding Warlock Patron Selection

### Old Way (Multiple Files to Change)
- Update progress.go with new step logic
- Update multiple handlers to check for warlock
- Add new patron selection handler
- Update routing logic

### New Way (Service-Only Change)
```go
// Just add to FlowBuilder.buildClassSpecificSteps()
case "warlock":
    steps = append(steps, b.buildWarlockSteps(ctx, char)...)

func (b *FlowBuilderImpl) buildWarlockSteps(ctx context.Context, char *character.Character) []character.CreationStep {
    // Return patron selection step with options
    patronChoice := rulebook.GetPatronChoice()
    // ... convert to CreationStep
}
```

Handlers automatically work with the new step without any changes!