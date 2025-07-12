# ADR-003: Character-Rulebook Domain Separation

## Status
Proposed

## Context
The character domain currently has tight coupling to D&D 5e specific rulebook implementations. This creates several problems:

1. **Portability**: Cannot extract character domain to rpg-toolkit without bringing D&D 5e dependencies
2. **Testing**: Tests require concrete D&D types, making them brittle and hard to maintain
3. **Extensibility**: Adding new rulebook systems would require modifying character domain
4. **Separation of Concerns**: Business logic is scattered across domains instead of centralized

Current coupling points include:
- Direct imports of rulebook types in character domain
- Hardcoded D&D 5e calculations (proficiency bonus, ability scores)
- Character creation logic mixed with domain model
- Services containing rulebook-specific business logic

## Decision
We will separate the character and rulebook domains using interface-based abstraction:

1. **Define Interfaces**: Character domain will define interfaces for all rulebook concepts
2. **Move Creation Logic**: Character creation becomes a rulebook responsibility
3. **Extract Calculations**: All game mechanics move to rulebook-specific calculators
4. **Thin Services**: Services become orchestrators only, no business logic
5. **Dependency Injection**: Rulebook implementations injected at runtime

### Interface Categories

#### Entity Interfaces
```go
type Race interface {
    GetID() string
    GetName() string
    GetAbilityScoreIncrease() map[string]int
    GetSize() string
    GetSpeed() int
}

type Class interface {
    GetID() string
    GetName() string
    GetHitDice() int
    GetPrimaryAbility() string
}
```

#### Calculator Interfaces
```go
type ProficiencyCalculator interface {
    CalculateBonus(level int) int
}

type ArmorClassCalculator interface {
    Calculate(ctx context.Context, char *Character) (int, error)
}
```

#### Service Interfaces
```go
type CharacterCreationService interface {
    StartCreation(ctx context.Context, userID, name string) (CreationSession, error)
    CompleteStep(ctx context.Context, session CreationSession, data interface{}) error
    FinishCreation(ctx context.Context, session CreationSession) (*Character, error)
}
```

## Consequences

### Positive
- **Clean Architecture**: Clear boundaries between domains
- **Testability**: Can test with mocked interfaces, reducing test maintenance
- **Portability**: Character domain can move to rpg-toolkit without modifications
- **Extensibility**: New rulebook systems just implement interfaces
- **Single Responsibility**: Each domain has one clear purpose

### Negative
- **More Interfaces**: Additional abstraction layer to maintain
- **Wiring Complexity**: Dependency injection setup more complex
- **Initial Refactoring**: Significant upfront work to migrate
- **Indirection**: One more layer between handlers and implementations

### Migration Strategy
1. **Phase 1**: Create interfaces without breaking existing code
2. **Phase 2**: Move creation logic to rulebook incrementally
3. **Phase 3**: Refactor services to use interfaces
4. **Phase 4**: Remove direct imports from character domain
5. **Phase 5**: Update handlers and tests
6. **Phase 6**: Documentation and cleanup

## Examples

### Before (Tight Coupling)
```go
// character/character.go
import "internal/domain/rulebook/dnd5e"

type Character struct {
    Race  *dnd5e.Race
    Class *dnd5e.Class
}

func (c *Character) GetProficiencyBonus() int {
    return 2 + ((c.Level - 1) / 4) // D&D 5e specific
}
```

### After (Interface-Based)
```go
// character/character.go - NO rulebook imports!
type Character struct {
    Race  Race  // interface
    Class Class // interface
}

// character/service.go
type Service struct {
    proficiencyCalc ProficiencyCalculator // injected
}

func (s *Service) GetProficiencyBonus(char *Character) int {
    return s.proficiencyCalc.CalculateBonus(char.Level)
}
```

### Testing Before
```go
// Had to create concrete D&D types
race := &dnd5e.Race{ID: "elf", Name: "Elf"}
class := &dnd5e.Class{ID: "wizard", Name: "Wizard"}
```

### Testing After
```go
// Simple mocks
race := &MockRace{}
race.On("GetID").Return("elf")
race.On("GetName").Return("Elf")
```

## Related
- Follows: ADR-001 (CharacterDraft with FlowState)
- Follows: ADR-002 (Step Handler Registry)
- Enables: Future extraction to rpg-toolkit
- Fixes: Issue #247 (Decouple character from rulebook)