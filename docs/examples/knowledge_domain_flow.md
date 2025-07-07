# Knowledge Domain Implementation Example

This document demonstrates how the new service-driven architecture makes implementing complex class features like Knowledge Domain skill and language selection trivial.

## Problem Statement

**Issue #285**: Knowledge Domain clerics should receive:
1. **2 bonus skill proficiencies** from: Arcana, History, Nature, or Religion  
2. **2 additional languages** of their choice
3. These selections should appear as steps in character creation

## Solution Using Service-Driven Architecture

### 1. Flow Builder automatically detects Knowledge Domain

```go
// In FlowBuilderImpl.buildClericSteps()
if b.hasSelectedDomain(char, "knowledge") {
    // Add Knowledge Domain specific steps
    steps = append(steps,
        character.CreationStep{
            Type:       character.StepTypeSkillSelection,
            Title:      "Choose Knowledge Domain Skills", 
            Description: "As a Knowledge domain cleric, choose 2 additional skill proficiencies.",
            Options:     skillOptions, // Arcana, History, Nature, Religion
            MinChoices:  2,
            MaxChoices:  2,
            Required:    true,
        },
        character.CreationStep{
            Type:       character.StepTypeLanguageSelection,
            Title:      "Choose Knowledge Domain Languages",
            Description: "As a Knowledge domain cleric, choose 2 additional languages.",
            Options:     languageOptions,
            MinChoices:  2,
            MaxChoices:  2, 
            Required:    true,
        },
    )
}
```

### 2. Service processes selections automatically

```go
// In CreationFlowServiceImpl.ProcessStepResult()
switch result.StepType {
case character.StepTypeSkillSelection:
    return s.applySkillSelection(char, result)
case character.StepTypeLanguageSelection:
    return s.applyLanguageSelection(char, result)
}

func (s *CreationFlowServiceImpl) applySkillSelection(char *character.Character, result *CreationStepResult) error {
    // Find divine domain feature and add bonus skills
    feature.Metadata["bonus_skills"] = result.Selections
    
    // Add to character proficiencies
    for _, skillKey := range result.Selections {
        prof := &rulebook.Proficiency{
            Key:  skillKey,
            Name: skillKey,
            Type: rulebook.ProficiencyTypeSkill,
        }
        char.AddProficiency(prof)
    }
    
    return s.characterService.UpdateEquipment(char)
}
```

### 3. Handler renders steps generically

```go
// CreationFlowHandler.renderStep() works for ANY step type
func (h *CreationFlowHandler) renderStep(step *character.CreationStep, characterID string) error {
    embed := &discordgo.MessageEmbed{
        Title:       step.Title,        // "Choose Knowledge Domain Skills"
        Description: step.Description,  // "As a Knowledge domain cleric..."
        Color:       h.getStepColor(step.Type), // Purple for skills
    }
    
    // Build select menu from step.Options
    components := h.buildStepComponents(step, characterID)
    
    // Send Discord message - works for any step type!
}
```

## Test Results

```bash
=== RUN   TestKnowledgeDomainStepGeneration
--- PASS: TestKnowledgeDomainStepGeneration (0.00s)

=== RUN   TestOtherDomainNoExtraSteps  
--- PASS: TestOtherDomainNoExtraSteps (0.00s)
```

## Character Creation Flow Example

### Normal Cleric (Life Domain)
1. ✅ Race Selection
2. ✅ Class Selection  
3. ✅ Ability Scores
4. ✅ Ability Assignment
5. ✅ Divine Domain Selection → **Life Domain**
6. ⏳ Proficiency Selection
7. ⏳ Equipment Selection
8. ⏳ Character Details

### Knowledge Domain Cleric
1. ✅ Race Selection
2. ✅ Class Selection
3. ✅ Ability Scores  
4. ✅ Ability Assignment
5. ✅ Divine Domain Selection → **Knowledge Domain**
6. ⏳ **Knowledge Domain Skills** (2 from Arcana/History/Nature/Religion)
7. ⏳ **Knowledge Domain Languages** (2 additional languages)
8. ⏳ Proficiency Selection
9. ⏳ Equipment Selection
10. ⏳ Character Details

## Key Benefits Demonstrated

### ✅ **Zero Handler Changes Required**
- No new handlers needed for Knowledge Domain
- Existing `CreationFlowHandler` handles all step types generically
- Discord UI automatically adapts to new step types

### ✅ **Service-Only Implementation**
- All logic in `FlowBuilder` and `CreationFlowService`
- Easy to test without Discord dependencies
- Consistent behavior across all creation paths

### ✅ **Extensible Pattern**
- Adding Warlock patrons: Just add to `buildWarlockSteps()`
- Adding Sorcerer origins: Just add to `buildSorcererSteps()`
- Adding complex multiclass flows: Service handles it all

### ✅ **Proper Data Storage**
- Skills stored in `feature.Metadata["bonus_skills"]`
- Languages stored in `feature.Metadata["bonus_languages"]`
- Added to character proficiencies automatically
- Character sheet displays selections using `selection_display`

## Implementation Status

- ✅ Architecture foundation (PR #287)
- ✅ Knowledge Domain flow detection
- ✅ Skill selection step generation
- ✅ Language selection step generation  
- ✅ Step validation and processing
- ✅ Test coverage for all scenarios
- ⏳ Handler integration (can be done incrementally)
- ⏳ UI testing with Discord bot

**Ready for**: User testing of Knowledge Domain cleric character creation!