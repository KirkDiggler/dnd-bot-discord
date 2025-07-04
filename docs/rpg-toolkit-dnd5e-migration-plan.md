# rpg-toolkit D&D 5e Migration Plan

## Goal: Complete Pure rpg-toolkit Implementation

Replace the Discord bot's event system and D&D 5e logic with a comprehensive rpg-toolkit-based implementation that can power the full D&D 5e rulebook locally.

## Current State Analysis

### ✅ What Works in rpg-toolkit
- **Event System**: Comprehensive event bus with proper entity/context support
- **Basic Modifiers**: Attack bonus, damage bonus, advantage/disadvantage
- **Entity System**: Clean Entity interface for characters/monsters
- **Condition Foundation**: Basic condition interface exists

### ❌ What's Missing in rpg-toolkit
- **Proficiency System**: No proficiency bonus calculations or type checking
- **Resource Management**: No spell slots, ability uses, hit dice tracking
- **Equipment System**: No weapons, armor, or equipment properties
- **Feature System**: No class/race features implementation
- **Duration Tracking**: Limited compared to Discord bot's robust system
- **Action Economy**: No action/bonus action/reaction tracking

## Migration Strategy: 3 Phases

### Phase 1: Foundation Systems (Current Sprint)
Build the core D&D 5e systems that everything else depends on.

#### 1.1 Proficiency System
```go
// Add to rpg-toolkit
type ProficiencyType string
const (
    ProficiencyWeapon     ProficiencyType = "weapon"
    ProficiencyArmor      ProficiencyType = "armor"
    ProficiencySkill      ProficiencyType = "skill"
    ProficiencySave       ProficiencyType = "saving_throw"
    ProficiencyTool       ProficiencyType = "tool"
)

type Proficiency interface {
    Type() ProficiencyType
    Key() string
    Name() string
    Category() string // "simple-weapons", "martial-weapons", etc.
}

type ProficiencyManager interface {
    HasProficiency(entity Entity, profType ProficiencyType, key string) bool
    GetProficiencyBonus(entity Entity) int
    AddProficiency(entity Entity, prof Proficiency)
    RemoveProficiency(entity Entity, prof Proficiency)
}
```

#### 1.2 Resource Management System
```go
type ResourceType string
const (
    ResourceSpellSlots ResourceType = "spell_slots"
    ResourceHitDice    ResourceType = "hit_dice"
    ResourceAbility    ResourceType = "ability_use"
    ResourceAction     ResourceType = "action_economy"
)

type Resource interface {
    Type() ResourceType
    Key() string
    Current() int
    Maximum() int
    Consume(amount int) error
    Restore(amount int) error
    RestoreToMax()
}

type ResourceManager interface {
    GetResource(entity Entity, resourceType ResourceType, key string) Resource
    AddResource(entity Entity, resource Resource)
    ProcessShortRest(entity Entity)
    ProcessLongRest(entity Entity)
}
```

#### 1.3 Equipment System
```go
type EquipmentType string
const (
    EquipmentWeapon EquipmentType = "weapon"
    EquipmentArmor  EquipmentType = "armor"
    EquipmentItem   EquipmentType = "item"
)

type Equipment interface {
    Type() EquipmentType
    Key() string
    Name() string
    Properties() []string
    IsEquipped() bool
}

type Weapon interface {
    Equipment
    AttackBonus(wielder Entity) int
    DamageBonus(wielder Entity) int
    DamageDice() string
    IsFinesse() bool
    IsVersatile() bool
    IsTwoHanded() bool
}

type EquipmentManager interface {
    GetEquippedWeapons(entity Entity) []Weapon
    GetEquippedArmor(entity Entity) []Equipment
    EquipItem(entity Entity, item Equipment) error
    UnequipItem(entity Entity, item Equipment) error
}
```

### Phase 2: Feature System (Next Sprint)
Implement the D&D 5e class/race features that modify gameplay.

#### 2.1 Feature System
```go
type FeatureType string
const (
    FeatureRacial    FeatureType = "racial"
    FeatureClass     FeatureType = "class"
    FeatureSubclass  FeatureType = "subclass"
    FeatureFeat      FeatureType = "feat"
)

type Feature interface {
    Type() FeatureType
    Key() string
    Name() string
    Description() string
    Level() int
    Source() string
    
    // Feature behavior
    IsPassive() bool
    GetModifiers() []Modifier
    GetEventListeners() []EventListener
}

type FeatureManager interface {
    GetFeatures(entity Entity, featureType FeatureType) []Feature
    AddFeature(entity Entity, feature Feature)
    RemoveFeature(entity Entity, featureKey string)
    ProcessLevelUp(entity Entity, newLevel int)
}
```

#### 2.2 Migrate Core Features
- **Rage**: Event-driven damage bonus and resistance
- **Sneak Attack**: Conditional damage with usage tracking
- **Bardic Inspiration**: Resource-based ability with action economy
- **Spell Casting**: Spell slot management and casting mechanics

### Phase 3: Advanced Mechanics (Final Sprint)
Complete the D&D 5e implementation with advanced systems.

#### 3.1 Condition System Enhancement
```go
type ConditionEffect interface {
    Condition
    GetModifiers() []Modifier
    GetRestrictedActions() []ActionType
    GetAutomaticFailures() []string
    GetAutomaticSuccesses() []string
}

// Standard D&D 5e conditions
var (
    ConditionBlinded      = NewStandardCondition("blinded", /* modifiers */)
    ConditionCharmed      = NewStandardCondition("charmed", /* modifiers */)
    ConditionFrightened   = NewStandardCondition("frightened", /* modifiers */)
    ConditionParalyzed    = NewStandardCondition("paralyzed", /* modifiers */)
    ConditionPoisoned     = NewStandardCondition("poisoned", /* modifiers */)
    ConditionProne        = NewStandardCondition("prone", /* modifiers */)
    ConditionRestrained   = NewStandardCondition("restrained", /* modifiers */)
    ConditionStunned      = NewStandardCondition("stunned", /* modifiers */)
    ConditionUnconscious  = NewStandardCondition("unconscious", /* modifiers */)
)
```

#### 3.2 Combat State Tracking
```go
type CombatState interface {
    GetInitiativeOrder() []Entity
    GetCurrentTurn() Entity
    GetRoundNumber() int
    AdvanceTurn()
    AddCombatant(entity Entity, initiative int)
    RemoveCombatant(entity Entity)
}

type ActionEconomy interface {
    HasAction(entity Entity) bool
    HasBonusAction(entity Entity) bool
    HasReaction(entity Entity) bool
    ConsumeAction(entity Entity, actionType ActionType) error
    ResetTurn(entity Entity)
}
```

## Implementation Approach

### Step 1: Remove EventBusAdapter
- Migrate all Discord bot listeners to use rpg-toolkit events directly
- Remove the `EventBusAdapter` and conversion logic
- Update all event emissions to use rpg-toolkit event types

### Step 2: Build Foundation Systems
- Implement ProficiencyManager in rpg-toolkit
- Add ResourceManager for spell slots and abilities
- Create basic EquipmentManager

### Step 3: Migrate Features
- Convert Rage to pure rpg-toolkit implementation
- Convert SneakAttack to pure rpg-toolkit implementation
- Convert ViciousMockery to pure rpg-toolkit implementation

### Step 4: Test & Validate
- Ensure all existing functionality still works
- Add comprehensive integration tests
- Validate performance (should be better without adapter overhead)

## Benefits of Pure rpg-toolkit Approach

1. **No Translation Bugs**: Eliminates context-loss bugs like we just fixed
2. **Better Performance**: No event conversion overhead
3. **Cleaner Architecture**: Single event system instead of dual system
4. **rpg-toolkit Improvements**: Our D&D 5e implementation becomes reusable
5. **Local Rulebook**: Full D&D 5e rules engine in rpg-toolkit

## Files to Create

### rpg-toolkit Extensions
1. `proficiency/manager.go` - Proficiency system
2. `resources/manager.go` - Resource management (spell slots, abilities)
3. `equipment/manager.go` - Equipment and weapon system
4. `features/manager.go` - Class/race feature system
5. `conditions/dnd5e.go` - Standard D&D 5e conditions
6. `combat/state.go` - Combat state and action economy

### Migration Files
1. `internal/adapters/rpgtoolkit/dnd5e/` - D&D 5e specific rpg-toolkit implementations
2. Updated listeners in `internal/domain/rulebook/dnd5e/` to use pure rpg-toolkit
3. Remove `internal/adapters/rpgtoolkit/event_bus.go` entirely

This plan transforms the Discord bot from using rpg-toolkit as a library to making rpg-toolkit the authoritative D&D 5e rules engine.