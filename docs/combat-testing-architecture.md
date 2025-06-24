# Combat Testing Infrastructure Architecture

## Overview
This document outlines the architecture for building a testable combat system that addresses the current issues with combat flow, monster turns, and UI updates.

## Core Design Principles

1. **Dependency Injection**: All dependencies should be injected, making components easily testable
2. **Interface-Based Design**: Use interfaces for all major components to enable mocking
3. **Deterministic Testing**: Abstract randomness through a dice roller interface
4. **Event-Driven Updates**: Use an event system for UI updates to decouple combat logic from Discord UI
5. **State Machine**: Implement combat as a state machine with clear transitions

## Key Components

### 1. Dice Roller Interface
```go
type DiceRoller interface {
    Roll(count, size, modifier int) (*RollResult, error)
    RollWithAdvantage(count, size, modifier int) (*RollResult, error)
    RollWithDisadvantage(count, size, modifier int) (*RollResult, error)
}
```

### 2. Combat State Machine
```go
type CombatState string

const (
    StateInitiating      CombatState = "initiating"
    StateRollingInitiative CombatState = "rolling_initiative"
    StateInProgress      CombatState = "in_progress"
    StateRoundPending    CombatState = "round_pending"
    StateComplete        CombatState = "complete"
)

type CombatStateMachine interface {
    CurrentState() CombatState
    CanTransition(to CombatState) bool
    Transition(to CombatState) error
    ValidActions() []string
}
```

### 3. Combat Event System
```go
type EventType string

const (
    EventCombatStarted   EventType = "combat_started"
    EventTurnStarted     EventType = "turn_started"
    EventTurnEnded       EventType = "turn_ended"
    EventRoundStarted    EventType = "round_started"
    EventRoundEnded      EventType = "round_ended"
    EventDamageDealt     EventType = "damage_dealt"
    EventCombatantDowned EventType = "combatant_downed"
    EventCombatEnded     EventType = "combat_ended"
)

type CombatEvent interface {
    Type() EventType
    EncounterID() string
    Timestamp() time.Time
    Data() map[string]interface{}
}

type EventBus interface {
    Subscribe(eventType EventType, handler EventHandler) string
    Unsubscribe(subscriptionID string)
    Publish(event CombatEvent) error
}
```

### 4. Combat Service Refactor
```go
type CombatService interface {
    // Core combat operations
    InitiateCombat(ctx context.Context, input InitiateCombatInput) (*Combat, error)
    ProcessTurn(ctx context.Context, combatID string, action CombatAction) (*TurnResult, error)
    AdvanceTurn(ctx context.Context, combatID string) error
    ContinueRound(ctx context.Context, combatID string) error
    
    // State queries
    GetCombatState(ctx context.Context, combatID string) (*CombatState, error)
    GetCurrentTurn(ctx context.Context, combatID string) (*TurnInfo, error)
    
    // Testing helpers
    SetDiceRoller(roller DiceRoller)
    SetEventBus(bus EventBus)
}
```

### 5. Turn Processor
```go
type TurnProcessor interface {
    ProcessPlayerTurn(ctx context.Context, combat *Combat, action PlayerAction) (*TurnResult, error)
    ProcessMonsterTurn(ctx context.Context, combat *Combat, monster *Combatant) (*TurnResult, error)
    ValidateAction(combat *Combat, actor *Combatant, action CombatAction) error
}
```

### 6. Combat Repository Interface
```go
type CombatRepository interface {
    Create(ctx context.Context, combat *Combat) error
    Get(ctx context.Context, id string) (*Combat, error)
    Update(ctx context.Context, combat *Combat) error
    Delete(ctx context.Context, id string) error
    
    // Atomic operations
    UpdateTurnOrder(ctx context.Context, id string, turnOrder []string) error
    UpdateCombatantHP(ctx context.Context, combatID, combatantID string, newHP int) error
    AppendToLog(ctx context.Context, combatID string, entry LogEntry) error
}
```

## Testing Strategy

### 1. Unit Tests
- Mock dice roller for deterministic results
- Mock event bus to verify event publishing
- Test state transitions
- Test turn processing logic
- Test damage calculations

### 2. Integration Tests
- Test full combat flows with predetermined dice rolls
- Test round transitions
- Test monster AI decisions
- Test concurrent turn processing

### 3. Test Scenarios
- Basic combat flow (player vs single monster)
- Multi-combatant scenarios
- Death and removal from turn order
- Round transitions
- Combat completion conditions
- Edge cases (0 HP, negative damage, etc.)

## Implementation Phases

### Phase 1: Core Infrastructure
- Dice roller interface and implementations
- Event system
- State machine

### Phase 2: Service Refactoring
- Extract interfaces from existing services
- Implement dependency injection
- Add event publishing

### Phase 3: Testing Framework
- Mock implementations
- Test builders for combat scenarios
- Integration test suite

### Phase 4: Monster AI
- Extract monster decision logic
- Make AI testable and configurable
- Add different AI strategies

### Phase 5: UI Updates
- Implement event handlers for Discord updates
- Decouple UI from combat logic
- Add proper error handling for stale encounters

## Benefits
1. **Predictable Testing**: Can test exact combat scenarios
2. **Better Debugging**: Event log shows exact flow
3. **Extensibility**: Easy to add new features
4. **Reliability**: State machine prevents invalid states
5. **Performance**: Can optimize without breaking tests