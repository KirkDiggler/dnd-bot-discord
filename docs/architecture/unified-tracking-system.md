# Unified Tracking System Architecture

## Problem Statement

Currently, we have multiple tracking systems that overlap in functionality:
- **Ability Service**: Tracks cooldowns, uses remaining, action economy
- **Status Effects**: Tracks temporary conditions with durations  
- **Resources**: Tracks consumable pools (HP, spell slots, ki points)
- **Event Listeners**: Tracks active modifiers (like rage damage bonus)

This leads to:
- Duplicated tracking logic
- Tight coupling between systems
- Difficulty adding new abilities/effects
- Hardcoded ability logic in services

## Proposed Solution: Tracker + Modifier System

### Core Concept
Split each game effect into two parts:
1. **Tracker**: Manages lifecycle (duration, uses, pools)
2. **Modifier**: Applies game effect (damage bonus, resistance, etc.)

### Tracking Strategies

Based on **lifecycle** and **scope**:

#### Duration-Based Tracking
- **Use Case**: Temporary effects with time limits
- **Examples**: Rage (10 rounds), Bless (1 minute), Concentration
- **Cleanup**: Auto-expire when duration ends

#### Pool-Based Tracking  
- **Use Case**: Consumable resources
- **Examples**: Ki points, Spell slots, Lay on Hands pool
- **Cleanup**: Recharge on rest

#### Usage-Based Tracking
- **Use Case**: Limited uses per time period
- **Examples**: Ability uses, Channel Divinity, Action Surge
- **Cleanup**: Recharge on rest/turn

#### Event-Based Tracking
- **Use Case**: Active modifiers tied to conditions
- **Examples**: Rage damage bonus, Sneak attack eligibility
- **Cleanup**: When triggering condition ends

### Modifier System

Modifiers hook into game calculations and events:

#### Combat Modifiers
- **Attack Roll Modifiers**: Bless (+1d4), Advantage/Disadvantage
- **Damage Modifiers**: Rage (+2), Sneak Attack (+Xd6)
- **Defense Modifiers**: Resistance (half damage), AC bonuses

#### Resource Modifiers
- **Action Economy**: Action Surge (extra action), Martial Arts (bonus unarmed strike)
- **Spell Casting**: Spell slots enable casting, Metamagic modifies spells
- **Movement**: Monk speed bonus, difficult terrain penalties

### Composition Examples

```go
// Rage Effect Composition
rage := &TrackedEffect{
    Name: "Rage",
    Trackers: []Tracker{
        &DurationTracker{RemainingRounds: 10},
        &UsageTracker{Uses: 2, RechargeOn: LongRest},
    },
    Modifiers: []Modifier{
        &DamageModifier{Types: []string{"melee"}, Bonus: 2},
        &ResistanceModifier{DamageTypes: []string{"bludgeoning", "piercing", "slashing"}},
    },
}

// Bardic Inspiration Composition
inspiration := &TrackedEffect{
    Name: "Bardic Inspiration",
    Trackers: []Tracker{
        &UsageTracker{Uses: 3, RechargeOn: ShortRest},
        &TargetTracker{TargetID: "character-123"},
    },
    Modifiers: []Modifier{
        &AbilityCheckModifier{Bonus: "1d6", Uses: 1},
    },
}

// Monk Ki Points Composition
ki := &TrackedEffect{
    Name: "Ki",
    Trackers: []Tracker{
        &PoolTracker{Current: 5, Max: 5, RechargeOn: ShortRest},
    },
    Modifiers: []Modifier{
        &ActionModifier{Type: "bonus_action", Enable: "flurry_of_blows", Cost: 1},
        &ActionModifier{Type: "bonus_action", Enable: "patient_defense", Cost: 1},
        &ActionModifier{Type: "bonus_action", Enable: "step_of_wind", Cost: 1},
    },
}
```

## Benefits

### For Developers
- **Composable**: Mix and match trackers/modifiers for new effects
- **Extensible**: Add new tracker types or modifiers without changing existing code
- **Testable**: Each component can be tested in isolation
- **Reusable**: Same tracker/modifier can be used across different abilities

### For Rulesets
- **Ruleset Agnostic**: Services only manage tracking, not game rules
- **Easy Implementation**: New abilities are just data composition
- **Flexible**: Can model any game system's mechanics
- **Maintainable**: Game rules live in data, not code

### For Features
- **Consistent UI**: All tracked effects show in standardized format
- **Automatic Cleanup**: Trackers handle their own expiration logic
- **Event Integration**: Modifiers automatically hook into relevant events
- **Persistence**: Unified save/load for all tracked effects

## Implementation Strategy

### Phase 1: Interface Design
1. Define `Tracker` and `Modifier` interfaces
2. Create basic implementations for each tracking strategy
3. Design `TrackedEffect` composition system

### Phase 2: Service Refactor
1. Remove hardcoded ability logic from ability service
2. Implement generic tracking management
3. Move D&D 5e specific logic to rulebook package

### Phase 3: Migration
1. Convert existing abilities to use new system
2. Update event system to work with modifiers
3. Add UI support for tracked effects

### Phase 4: Enhancement
1. Add more modifier types as needed
2. Implement complex interactions (stacking, conflicts)
3. Add debugging/introspection tools

## Related Issues
- **Issue #237**: Refactor ability service to be ruleset-agnostic
- **Future**: Implement remaining D&D 5e abilities using this system
- **Future**: Add support for other rulesets (Pathfinder, etc.)

## Success Metrics
- Ability service has zero hardcoded ability knowledge
- New abilities can be added without code changes
- All tracking systems use unified interface
- Performance comparable to current implementation