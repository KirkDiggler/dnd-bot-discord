# Effect Systems Documentation

## Overview

The D&D bot currently maintains a **dual effect system** for managing character effects, buffs, debuffs, and status conditions. This duality exists for backward compatibility while transitioning to a more flexible architecture.

## The Two Systems

### 1. Legacy System: `ActiveEffect` (shared package)

Located in: `/internal/domain/shared/effects.go`

**Purpose**: Original effect system focused on persistence and basic modifiers.

**Structure**:
```go
type ActiveEffect struct {
    ID                    string
    Name                  string
    Description           string
    Source                string       // Spell/Feature that created it
    SourceID              string       // ID of caster/user
    Duration              int          // Amount remaining
    DurationType          DurationType
    Modifiers             []Modifier
    RequiresConcentration bool
}
```

**Key Features**:
- Simple modifier system (damage bonus, resistance, AC bonus)
- Basic duration tracking (rounds, minutes, hours, until rest, permanent)
- Direct persistence to Redis via character Resources
- Used by UI for display and combat calculations

### 2. New System: `StatusEffect` (effects package)

Located in: `/internal/effects/types.go`

**Purpose**: Modern effect system with enhanced flexibility and event integration.

**Structure**:
```go
type StatusEffect struct {
    ID           string
    Source       EffectSource
    SourceID     string
    Name         string
    Description  string
    Duration     Duration
    Modifiers    []Modifier
    Conditions   []Condition     // When the effect applies
    StackingRule StackingRule
    Active       bool
    CreatedAt    time.Time
    ExpiresAt    *time.Time
}
```

**Key Features**:
- Advanced modifier system with conditions
- Flexible stacking rules (replace, stack, take highest/lowest)
- Conditional application (e.g., only vs specific enemy types)
- Integration with event bus for reactive effects
- Built-in effect builders for common effects

## Synchronization Between Systems

The character maintains both systems in sync through careful conversion:

### Adding Effects Flow:
1. New `StatusEffect` is created (e.g., via `effects.BuildRageEffect()`)
2. Added to character's `EffectManager`
3. `syncEffectManagerToResources()` converts to `ActiveEffect` format
4. Both systems now have the effect

### Loading Effects Flow:
1. Character loads from Redis with `ActiveEffect` data
2. `syncResourcestoEffectManager()` rebuilds `StatusEffect` instances
3. Well-known effects (Rage, Favored Enemy) are rebuilt with full modifiers
4. Generic effects are converted with basic properties

## Conversion Mappings

### Duration Types:
```
StatusEffect (New)          →  ActiveEffect (Legacy)
DurationPermanent          →  DurationTypePermanent
DurationRounds             →  DurationTypeRounds
DurationUntilRest          →  DurationTypeUntilRest
DurationInstant            →  DurationTypeRounds
DurationWhileEquipped      →  DurationTypePermanent
DurationConcentration      →  (tracked via RequiresConcentration flag)
```

### Modifier Types:
Currently, modifier conversion is incomplete (marked with TODO). The systems use different approaches:
- **Legacy**: Simple type + value (e.g., damage_bonus: +2)
- **New**: Target + value + conditions (e.g., damage modifier, +2, melee_only)

## Usage Patterns

### Combat Calculations:
```go
// Both systems are checked for effects
character.ApplyDamageResistance() // Checks ActiveEffect.HasResistance()
character.GetDamageBonus()        // Sums ActiveEffect damage bonuses
```

### Ability Execution:
```go
// New abilities use StatusEffect
rageEffect := effects.BuildRageEffect(level)
character.AddStatusEffect(rageEffect) // Automatically syncs to ActiveEffect
```

### UI Display:
```go
// UI primarily uses ActiveEffect for backward compatibility
for _, effect := range character.Resources.ActiveEffects {
    // Display effect name, duration, description
}
```

## Known Issues and Limitations

1. **Incomplete Modifier Conversion**: Modifiers aren't fully converted between systems
2. **Double Resistance Bug** (Fixed in PR #261): Effects were being applied through both systems
3. **Performance Overhead**: Every effect operation requires sync between systems
4. **Complexity**: Developers must understand both systems and their interactions

## Future Direction

The goal is to eventually migrate fully to the `StatusEffect` system:
1. Update UI to read from `EffectManager` directly
2. Implement full modifier conversion
3. Deprecate `ActiveEffect` system
4. Simplify character effect management

## Best Practices

1. **Always use StatusEffect for new features**: The legacy system is for compatibility only
2. **Use effect builders**: Leverage `effects.BuildRageEffect()` and similar helpers
3. **Test both systems**: Ensure effects work in combat calculations and UI display
4. **Document special effects**: Well-known effects need rebuild logic in sync methods

## Examples

### Adding a Simple Effect:
```go
// Create effect using builder
rageEffect := effects.BuildRageEffect(character.Level)

// Add to character (auto-syncs to both systems)
character.AddStatusEffect(rageEffect)
```

### Checking for Effects:
```go
// New system (preferred)
for _, effect := range character.GetActiveStatusEffects() {
    if effect.Name == "Rage" && effect.Active {
        // Character is raging
    }
}

// Legacy system (for UI/combat)
for _, effect := range character.Resources.ActiveEffects {
    if effect.HasResistance("slashing") {
        // Apply resistance
    }
}
```

## Migration Status

- ✅ Core synchronization implemented
- ✅ Well-known effects (Rage, Favored Enemy) rebuild properly
- ⚠️  Modifier conversion incomplete
- ❌ UI still uses legacy system
- ❌ Some combat calculations use legacy system directly

The dual system approach allows gradual migration while maintaining stability, but adds complexity that should be addressed in future refactoring efforts.