# RPG Toolkit Extraction Guide

## Overview
This guide ensures our character domain refactoring aligns with rpg-toolkit patterns for future extraction.

## Key Principles from RPG Toolkit

### 1. Entity Pattern
All game objects must implement `core.Entity`:
```go
type Entity interface {
    GetID() string
    GetType() string
}
```

**Action Required**: Ensure our Character and CharacterDraft implement these methods.

### 2. Infrastructure vs Implementation
- **Infrastructure** (extractable to toolkit): Generic character concepts, interfaces, workflows
- **Implementation** (stays in bot or examples): D&D 5e specific races, classes, rules

### 3. Event-Driven Communication
- Character changes should emit events through the event bus
- Other systems (combat, UI) react to these events
- No direct coupling between modules

## Recommended Structure for Extraction

### Phase 1: Current Refactoring (In Progress)
```
/internal/domain/character/
  interfaces.go         # ✅ Created - Generic interfaces
  character.go          # TODO: Add GetID(), GetType() methods
  character_draft.go    # TODO: Move to rulebook (D&D specific)
  
/internal/domain/rulebook/dnd5e/
  character_draft.go    # D&D specific creation concept
  creation/             # D&D specific creation flow
  races/                # D&D races
  classes/              # D&D classes
```

### Phase 2: Align with Toolkit Patterns
```go
// character.go - Make toolkit-compatible
type Character struct {
    ID   string
    Name string
    // Use interfaces for rulebook-specific data
    Race  Race      // interface
    Class Class     // interface
}

func (c *Character) GetID() string { return c.ID }
func (c *Character) GetType() string { return "character" }
```

### Phase 3: Future Extraction
When ready to extract to toolkit:

1. **Move to toolkit**:
   - `/internal/domain/character/interfaces.go` → `/rpg-toolkit/character/interfaces.go`
   - `/internal/domain/character/character.go` → `/rpg-toolkit/character/character.go`
   - Generic creation workflow → `/rpg-toolkit/character/creation/`

2. **Keep in bot or move to examples**:
   - All D&D 5e specific implementations
   - CharacterDraft (D&D creation concept)
   - D&D races, classes, features

## Event Integration Points

### Character Events to Emit
```go
// When extracting, these become toolkit events
type CharacterCreated struct {
    CharacterID string
    UserID      string
}

type CharacterLeveledUp struct {
    CharacterID string
    OldLevel    int
    NewLevel    int
}

type CharacterStatChanged struct {
    CharacterID string
    Stat        string
    OldValue    int
    NewValue    int
}
```

### Subscribe to Toolkit Events
- `mechanics.ConditionApplied` - Update character state
- `combat.DamageTaken` - Update HP
- `features.FeatureActivated` - Apply feature effects

## Migration Checklist

### Before Extraction
- [ ] Character implements `core.Entity` interface
- [ ] All D&D specific logic moved to rulebook package
- [ ] Services use only interfaces, no concrete types
- [ ] Event publishing for all state changes
- [ ] Repository interface is generic (no D&D concepts)
- [ ] Tests use mocked interfaces

### During Extraction
- [ ] Copy generic interfaces to toolkit
- [ ] Copy character entity to toolkit
- [ ] Update imports in bot to use toolkit
- [ ] Move D&D implementations to `/examples/dnd5e/`
- [ ] Verify event bus integration works

### After Extraction
- [ ] Bot uses toolkit character package
- [ ] Other games can use the same character infrastructure
- [ ] D&D specific rules isolated in examples
- [ ] Clear documentation for implementing new rulesets

## Example: Future Toolkit Usage

```go
// In rpg-toolkit
package character

type Character struct {
    ID         string
    Name       string
    Level      int
    Attributes map[string]int
    
    // Rulebook-specific data as interfaces
    race       Race
    class      Class
    features   []Feature
}

// In dnd-bot-discord (using toolkit)
import (
    "github.com/yourusername/rpg-toolkit/character"
    "github.com/yourusername/dnd-bot-discord/internal/domain/rulebook/dnd5e"
)

func createDnDCharacter() *character.Character {
    char := character.New("123", "Gandalf")
    char.SetRace(dnd5e.NewRace("elf"))
    char.SetClass(dnd5e.NewClass("wizard"))
    return char
}
```

## Notes
- This is a living document - update as we learn more
- The refactoring we're doing now sets us up for clean extraction later
- Focus on interfaces and separation of concerns
- Event-driven patterns will make extraction much easier