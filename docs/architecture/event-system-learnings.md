# Event System Learnings

## Overview
This document captures key learnings from implementing the spell damage event system, providing insights for future event-driven architecture development.

## What Worked Well

### 1. Event Bus Pattern
- The centralized EventBus made it easy to add new event types without modifying existing code
- Subscribe/Emit pattern allowed clean separation between spell casting and damage application
- Priority system in handlers enables proper ordering of event processing

### 2. Context System
- The key-value context system on GameEvent proved flexible for passing varied data
- Type-safe getter methods (GetIntContext, GetStringContext) prevent runtime errors
- Constants for context keys improve maintainability and prevent typos

### 3. Handler Interface
- Simple interface (HandleEvent, Priority) made it easy to implement new handlers
- SpellDamageHandler cleanly separated spell damage logic from encounter service

## Challenges and Solutions

### 1. Cross-Service Dependencies
**Challenge**: SpellDamageHandler needed access to the encounter service to apply damage
**Solution**: Passed service as dependency to handler constructor
**Learning**: Event handlers often need service dependencies; consider a handler factory pattern

### 2. Missing Context Data
**Challenge**: Events needed encounter_id and user_id but these weren't always available
**Solution**: Added context fields when emitting events from spell implementations
**Learning**: Define required context fields for each event type upfront

### 3. Target Identification
**Challenge**: Monsters don't have Character objects, but events expected them
**Solution**: Created minimal Character objects for monster targets
**Learning**: Need a more generic "Combatant" interface for events

## Event System Needs Discovered

### 1. Event Type Definitions
Each event type should clearly define:
- Required context fields
- Optional context fields
- Expected actor/target types
- Whether the event can be cancelled

Example:
```go
// OnSpellDamage Event
// Required Context:
//   - ContextDamage (int): Amount of damage
//   - ContextDamageType (string): Type of damage
//   - ContextEncounterID (string): Current encounter
// Optional Context:
//   - ContextSpellName (string): Name of spell
//   - ContextSpellLevel (int): Level spell was cast at
// Actor: The caster (Character)
// Target: The target (Character or minimal stub)
```

### 2. Event Flow Documentation
Need clear documentation of event chains:
```
OnSpellCast -> OnSpellDamage -> BeforeTakeDamage -> AfterTakeDamage
                     |
                     v
              OnSavingThrow (future)
```

### 3. Handler Registration Patterns
Current: Manual registration in service constructors
Better: Automatic registration based on handler interfaces
```go
// Future pattern
type AutoRegisterHandler interface {
    Handler
    EventTypes() []EventType
}
```

### 4. Testing Utilities
Need test utilities for event-driven code:
- Mock event bus that captures emitted events
- Test helpers to assert event emissions
- Fixtures for common event scenarios

## Recommendations for Future Development

### 1. Standardize Event Context
Create standard context fields that all combat events should include:
- `ContextEncounterID`: Always needed for combat events
- `ContextUserID`: Who initiated the action
- `ContextRound`: Current combat round
- `ContextTurnCount`: Total turns elapsed

### 2. Generic Combatant Interface
Instead of requiring Character objects, define a minimal interface:
```go
type Combatant interface {
    GetID() string
    GetName() string
    GetType() CombatantType // Player or Monster
}
```

### 3. Event Builder Pattern
Make event creation more ergonomic:
```go
event := events.NewCombatEvent(events.OnSpellDamage).
    InEncounter(encounterID).
    ByUser(userID).
    FromActor(caster).
    ToTarget(target).
    WithDamage(damage, damageType).
    Build()
```

### 4. Event Documentation Generator
Consider tooling to generate event documentation from code:
- Extract required/optional context from handler implementations
- Generate event flow diagrams
- Validate event emissions match documented contracts

## Next Steps

1. **Condition System**: The event system is well-suited for status effects
   - OnConditionApplied, OnConditionExpired events
   - Conditions can listen for relevant events (e.g., Concentration listens for damage)

2. **Saving Throws**: Natural fit for events
   - OnSavingThrowRequired event
   - Handlers can modify DC or advantage/disadvantage
   - Results trigger follow-up events

3. **Action Economy**: Events can track action usage
   - OnActionUsed, OnBonusActionUsed events
   - Abilities can listen and enable bonus actions

## Conclusion

The event system shows great promise for implementing D&D mechanics in a modular way. The key insight is that D&D rules are largely about things reacting to other things - perfect for event-driven architecture. With some refinements to handle cross-cutting concerns (encounter context, generic targets), this pattern can elegantly handle complex rule interactions.