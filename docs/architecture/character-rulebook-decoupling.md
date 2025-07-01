# Character-Rulebook Decoupling Plan

## Current State Analysis

### Direct Rulebook Imports in Character Package
1. **Type Dependencies**:
   - `*rulebook.Race`
   - `*rulebook.Class`
   - `*rulebook.Background`
   - `*rulebook.Proficiency`
   - `*rulebook.CharacterFeature`
   - `rulebook.ProficiencyType`

2. **Logic Dependencies**:
   - AC calculation has D&D 5e-specific rules
   - Fighting style bonuses are hardcoded
   - Feature effects are coupled to character methods

### Current Dependency Flow
```
rulebook ← character (WRONG!)
         ↓
       shared
```

### Desired Dependency Flow
```
rulebook → character → shared
```

## Proposed Architecture

### Phase 1: Extract Interfaces for Core Types

Following the same pattern as abilities, create interfaces in the character package that rulebook types will implement:

```go
// internal/domain/character/interfaces.go
package character

type RaceInterface interface {
    GetKey() string
    GetName() string
    GetBaseSpeed() int
    GetSize() string
    // Note: No game-specific methods like GetAbilityBonuses()
}

type ClassInterface interface {
    GetKey() string
    GetName() string
    GetLevel() int
    // Note: No game-specific methods like GetHitDie()
}

type BackgroundInterface interface {
    GetKey() string
    GetName() string
    GetDescription() string
}

type ProficiencyInterface interface {
    GetKey() string
    GetName() string
    GetType() string // Generic type, not rulebook-specific
}

type FeatureInterface interface {
    GetKey() string
    GetName() string
    GetDescription() string
    IsActive() bool
}
```

### Consistent Pattern Application

This follows the same successful pattern we used for abilities:
1. **Service defines interface** for what it needs
2. **Rulebook implements** game-specific version
3. **Adapter/Registry** pattern for registration
4. **No circular dependencies**

### Phase 2: Create Calculator Pattern

Move all ruleset-specific calculations to the rulebook:

```go
// internal/domain/rulebook/dnd5e/calculators/ac_calculator.go
package calculators

type ACCalculator struct{}

func (a *ACCalculator) Calculate(char *character.Character) int {
    // All D&D 5e-specific AC logic here
    // Including unarmored defense, fighting styles, etc.
}
```

### Phase 3: Use Service Layer for Rule Application

The service layer orchestrates between character and rulebook:

```go
// internal/services/character/service.go
func (s *service) UpdateCharacterAC(char *character.Character) {
    calculator := s.rulebook.GetACCalculator()
    char.AC = calculator.Calculate(char)
}
```

### Phase 4: Event-Driven Feature System

Features emit events that the rulebook listens to:

```go
// Character has features but doesn't know what they do
char.Features = []FeatureInterface{
    &GenericFeature{Key: "rage", Active: true},
}

// Rulebook listens for feature activation
eventBus.Subscribe("feature.activated", rageHandler)
```

## Implementation Steps

### Step 1: Create Core Interfaces
- [ ] Define RaceInterface, ClassInterface, etc. in character package
- [ ] Update Character struct to use interfaces
- [ ] Create adapters in rulebook to implement interfaces

### Step 2: Extract AC Calculation
- [ ] Create ACCalculator in rulebook
- [ ] Move all AC logic from character.calculateAC()
- [ ] Update service to use calculator

### Step 3: Extract Fighting Style Logic
- [ ] Move applyFightingStyleBonuses to rulebook
- [ ] Create FightingStyleCalculator
- [ ] Update attack methods to use calculator

### Step 4: Generalize Proficiencies
- [ ] Replace rulebook.ProficiencyType with string constants
- [ ] Create generic proficiency structure
- [ ] Move proficiency logic to rulebook

### Step 5: Feature System Refactor
- [ ] Create generic feature interface
- [ ] Move feature effects to event handlers
- [ ] Update character to store features without knowing effects

### Step 6: Update Tests
- [ ] Create mock implementations of interfaces
- [ ] Update existing tests
- [ ] Add integration tests for calculators

## Benefits

1. **Multi-Ruleset Support**: Easy to add Pathfinder, etc.
2. **Clean Architecture**: Clear separation of concerns
3. **Better Testing**: Can test character logic without rulebook
4. **Flexibility**: Rules can change without touching character
5. **Maintainability**: Changes to rules don't cascade

## Example: AC Calculation After Refactor

```go
// Before (in character package)
func (c *Character) calculateAC() {
    c.AC = 10
    // D&D 5e specific logic...
}

// After (in character package)
type Character struct {
    AC int // Just stores the value
    // No calculation logic
}

// In rulebook package
func (calc *ACCalculator) Calculate(char CharacterInterface) int {
    ac := 10
    
    // All D&D 5e logic here
    if armor := char.GetEquipment(SlotBody); armor != nil {
        // Apply armor AC
    }
    
    // Check for unarmored defense
    if char.HasFeature("unarmored_defense") {
        // Apply monk/barbarian logic
    }
    
    return ac
}
```

## Migration Strategy

1. **Start with new features**: Implement new features using the decoupled pattern
2. **Gradual refactor**: Move one calculation at a time
3. **Maintain compatibility**: Use adapters during transition
4. **Test extensively**: Ensure behavior doesn't change

## Open Questions

1. Should we use events for all feature effects or just some?
2. How do we handle feature interactions (e.g., multiclass)?
3. Should calculators be stateless or maintain state?
4. How do we handle UI display of rulebook-specific data?

## Additional Pattern Applications

Following the successful ability handler pattern, we can apply the same approach to:

### 1. Equipment Effects
```go
// Character has equipment interface
type EquipmentInterface interface {
    GetKey() string
    GetName() string
    GetSlot() string
}

// Rulebook handles equipment effects
type EquipmentHandler interface {
    OnEquip(char *Character, item EquipmentInterface)
    OnUnequip(char *Character, item EquipmentInterface)
}
```

### 2. Condition Effects
```go
// Character tracks conditions as data
type ConditionInterface interface {
    GetKey() string
    GetName() string
    GetDuration() int
}

// Rulebook handles condition effects
type ConditionHandler interface {
    OnApply(char *Character, condition ConditionInterface)
    OnRemove(char *Character, condition ConditionInterface)
    OnTurnStart(char *Character, condition ConditionInterface)
}
```

### 3. Spell Effects
```go
// Character knows spells as data
type SpellInterface interface {
    GetKey() string
    GetName() string
    GetLevel() int
}

// Rulebook handles spell mechanics
type SpellHandler interface {
    CanCast(char *Character, spell SpellInterface) bool
    Cast(char *Character, spell SpellInterface, target interface{})
}
```

### 4. Level Progression
```go
// Character has level, rulebook handles progression
type LevelProgressionHandler interface {
    OnLevelUp(char *Character, newLevel int)
    GetExperienceRequired(currentLevel int) int
    GetFeaturesAtLevel(class ClassInterface, level int) []FeatureInterface
}
```

## Pattern Benefits

By consistently applying this pattern:
1. **Predictable architecture** - Developers know where to look
2. **Easy ruleset swapping** - Just register different handlers
3. **Clean testing** - Mock interfaces easily
4. **No circular dependencies** - Clear hierarchy
5. **Extensible** - Add new rulesets without touching core

## Next Steps

1. Review and refine this plan
2. Create proof-of-concept for AC calculator (#248)
3. If successful, apply to race/class interfaces
4. Create detailed tickets for each phase
5. Document the pattern for future developers