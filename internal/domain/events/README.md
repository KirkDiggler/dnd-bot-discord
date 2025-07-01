# Events Package

## Purpose
The Events package provides a game-wide event bus that enables loose coupling between different parts of the system. It allows ruleset-specific code to react to game events without the core services knowing about specific implementations.

## Core Components

### EventBus
Central message broker that manages event subscriptions and dispatching.

```go
eventBus := events.NewEventBus()
eventBus.Subscribe(OnDamageRoll, myListener)
eventBus.Emit(event)
```

### GameEvent
The core event type that flows through the system.

```go
event := NewGameEvent(OnAttackRoll).
    WithActor(attacker).
    WithTarget(defender).
    WithContext("weapon_id", weaponID)
```

### EventListener
Interface that all event handlers must implement.

```go
type EventListener interface {
    HandleEvent(event *GameEvent) error
    Priority() int  // Lower numbers run first
}
```

## Event Types
- **Combat Events**: Attack rolls, damage rolls, hits, misses
- **Turn Events**: Turn start/end, round start/end  
- **Ability Events**: Ability used, spell cast
- **Status Events**: Effect applied/removed, condition changes
- **Rest Events**: Short rest, long rest

## Event Flow Example
```
1. Player attacks with sword
2. EncounterService emits OnAttackRoll event
3. Listeners process in priority order:
   - MagicWeaponListener adds +1 to attack
   - BlessListener adds 1d4 to attack
   - DisadvantageListener applies disadvantage
4. Attack resolves
5. EncounterService emits OnDamageRoll event
6. Damage listeners apply:
   - RageListener adds +2 damage
   - ResistanceListener halves damage
```

## Context Keys
Events carry context data as key-value pairs. Common keys should be constants:

```go
const (
    ContextAttackType = "attack_type"
    ContextDamage = "damage" 
    ContextTargetID = "target_id"
    ContextAbilityKey = "ability_key"
    // etc...
)
```

## Design Principles
1. **Loose Coupling**: Publishers don't know about subscribers
2. **Priority Ordering**: Critical modifiers run first
3. **Immutable Events**: Events shouldn't be modified after creation
4. **Context Over Structure**: Use context map for flexibility

## Best Practices
- Keep event handlers focused and fast
- Use meaningful event types (not generic "something_happened")
- Document what context keys each event type expects
- Handle errors gracefully - one listener shouldn't break others
- Use priorities to ensure correct order of operations

## Future Enhancements
- Event history/replay for debugging
- Async event processing for non-critical events
- Event filtering/routing based on context
- Performance metrics for event processing