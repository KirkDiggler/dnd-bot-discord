# Rulebook System Implementation Guide

## Overview

The D&D bot implements game mechanics through three distinct but interconnected systems: **Abilities**, **Features**, and **Feats**. Each serves a specific purpose in modeling D&D 5e rules while maintaining clean architecture boundaries.

## System Categories

### 1. Abilities - Active Character Actions

**Definition**: Actions that characters can actively use, typically consuming resources (action, bonus action, uses per rest).

**Location**: `/internal/domain/rulebook/dnd5e/abilities/`

**Core Interface**:
```go
type Handler interface {
    Key() string
    Execute(ctx context.Context, 
            char *character.Character, 
            ability *shared.ActiveAbility, 
            input *Input) (*Result, error)
}
```

**Characteristics**:
- Require explicit activation by the player
- Track usage (e.g., 3 uses per long rest)
- May have duration (10 rounds for Rage)
- Can apply temporary effects
- Consume action economy resources

**Examples**:
- **Rage** (Barbarian): Bonus action, limited uses, 10-round duration
- **Second Wind** (Fighter): Bonus action, 1 use per short rest, instant healing
- **Divine Smite** (Paladin): No action, uses spell slots, instant damage
- **Action Surge** (Fighter): Free action, grants additional action

### 2. Features - Passive Character Traits

**Definition**: Passive abilities that are always active or trigger automatically under specific conditions.

**Location**: `/internal/domain/rulebook/dnd5e/features/`

**Core Interface**:
```go
type FeatureHandler interface {
    GetKey() string
    ApplyPassiveEffects(char *character.Character) error
    ModifySkillCheck(char *character.Character, skillKey string, baseResult int) (int, bool)
    GetPassiveDisplayInfo(char *character.Character) (string, bool)
}
```

**Characteristics**:
- Always active or conditionally automatic
- Don't consume resources
- May grant proficiencies or bonuses
- Can modify other mechanics

**Examples**:
- **Darkvision**: Always active, grants vision in darkness
- **Keen Senses** (Elf): Grants Perception proficiency
- **Lucky** (Halfling): Reroll natural 1s
- **Fey Ancestry** (Elf): Advantage against charm

### 3. Feats - Optional Character Enhancements

**Definition**: Optional rules that provide special abilities, often mixing passive and active components.

**Location**: `/internal/domain/rulebook/dnd5e/feats/`

**Core Interface**:
```go
type Feat interface {
    Key() string
    Name() string
    Description() string
    Prerequisites() []Prerequisite
    CanTake(char *character.Character) bool
    Apply(char *character.Character) error
    RegisterHandlers(bus *rpgevents.Bus, char *character.Character)
}
```

**Characteristics**:
- Optional character choices
- May have prerequisites
- Often provide both passive and active benefits
- Can hook into event system

**Examples**:
- **Alert**: +5 initiative, can't be surprised
- **Great Weapon Master**: -5 attack/+10 damage option
- **Lucky**: 3 luck points per long rest
- **Sharpshooter**: Ignore cover, long range, power attack

## Implementation Patterns

### 1. Registry Pattern

All handlers are registered centrally for dynamic lookup:

```go
// Abilities Registry
func RegisterAll(registry interface{}, cfg *RegistryConfig) {
    registry.RegisterHandler(NewRageHandler(cfg.EventBus))
    registry.RegisterHandler(NewSecondWindHandler(cfg.DiceRoller))
    // ... more abilities
}
```

### 2. Handler Pattern

Each ability/feature/feat implements a specific handler interface:

```go
// Example: Rage Handler
type RageHandler struct {
    eventBus         *rpgevents.Bus
    characterService charService.Service
}

func (h *RageHandler) Execute(ctx context.Context, ...) (*Result, error) {
    // Toggle rage on/off
    // Apply/remove effects
    // Update character state
}
```

### 3. Event-Driven Integration

Complex mechanics use the event system:

```go
// Great Weapon Master subscribes to attack events
func (f *GreatWeaponMaster) RegisterHandlers(bus *rpgevents.Bus, char *character.Character) {
    bus.SubscribeFunc(rpgevents.EventBeforeAttackRoll, func(ctx context.Context, event rpgevents.Event) error {
        // Modify attack roll if power attack is chosen
    })
}
```

### 4. Effect Application

Effects are applied through the dual effect system:

```go
// Rage applies both damage bonus and resistance
rageEffect := effects.BuildRageEffect(char.Level)
char.AddStatusEffect(rageEffect) // Syncs to both effect systems
```

## Data Flow

### Ability Execution Flow:
```
Discord Button Click
    ↓
Combat Handler (handleUseAbility)
    ↓
Ability Service (UseAbility)
    ↓
Specific Handler (e.g., RageHandler.Execute)
    ↓
Character State Update
    ↓
Effect Application
    ↓
Event Publication
    ↓
UI Update
```

### Feature Application Flow:
```
Character Creation/Load
    ↓
Feature Registry Lookup
    ↓
FeatureHandler.ApplyPassiveEffects
    ↓
Character Modifications (proficiencies, bonuses)
    ↓
Permanent Effect Creation
```

## Key Design Decisions

### 1. Separation of Concerns
- **Abilities**: Active actions with resource management
- **Features**: Passive modifications and proficiencies
- **Feats**: Optional enhancements with mixed behavior

### 2. Event Bus Integration
- Allows loose coupling between systems
- Enables complex interactions (feats modifying attacks)
- Supports future extensibility

### 3. Dual Effect System
- Maintains backward compatibility
- Allows gradual migration to new architecture
- Ensures persistence works correctly

### 4. Registry Pattern
- Dynamic handler registration
- Easy to add new abilities/features
- Supports multiple rulesets (future)

## Common Patterns

### Adding a New Ability:
1. Create handler implementing `abilities.Handler`
2. Register in `abilities.RegisterAll`
3. Add to character initialization if class-specific
4. Create UI integration in combat handler

### Adding a New Feature:
1. Create handler implementing `features.FeatureHandler`
2. Register in feature registry
3. Add to race/class definitions
4. Apply during character creation

### Adding a New Feat:
1. Implement `feats.Feat` interface
2. Define prerequisites
3. Register event handlers if needed
4. Add to feat selection UI

## Consistency Guidelines

### Do:
- Use handler interfaces consistently
- Register all handlers centrally
- Apply effects through the effect system
- Use events for cross-cutting concerns
- Document special mechanics

### Don't:
- Hardcode ability logic in services
- Skip the registry pattern
- Apply effects directly without the effect system
- Create tight coupling between systems
- Mix UI concerns with domain logic

## Future Improvements

1. **Unified Handler Interface**: Consider unifying ability/feature/feat handlers
2. **Better Effect Conversion**: Complete the modifier conversion between effect systems
3. **Resource Management**: Centralized resource tracking (spell slots, ki points, etc.)
4. **Condition System**: Proper condition tracking (stunned, paralyzed, etc.)
5. **Action Economy**: More sophisticated action tracking

## Examples

### Simple Ability (Second Wind):
```go
func (h *SecondWindHandler) Execute(...) (*Result, error) {
    // Roll healing
    healing := h.diceRoller.RollDice(fmt.Sprintf("1d10+%d", char.Level))
    
    // Apply healing
    char.Heal(healing.Total)
    
    // Mark as used
    ability.UsesRemaining--
    
    return &Result{Success: true, Message: fmt.Sprintf("Healed for %d", healing.Total)}, nil
}
```

### Event-Driven Feat (Great Weapon Master):
```go
func (f *GreatWeaponMaster) RegisterHandlers(bus *rpgevents.Bus, char *character.Character) {
    bus.SubscribeFunc(rpgevents.EventBeforeAttackRoll, func(ctx context.Context, event rpgevents.Event) error {
        if powerAttack {
            event.Context().Set("attack_modifier", -5)
            event.Context().Set("damage_bonus", 10)
        }
        return nil
    })
}
```

This architecture provides a solid foundation for implementing D&D 5e mechanics while maintaining clean code boundaries and extensibility for future enhancements.