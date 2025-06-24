# GitHub Issues for Combat Testing Infrastructure

## Phase 1: Core Infrastructure

### Issue 1: Create Dice Roller Interface and Implementations
**Title**: Implement Dice Roller Interface with Mock and Real Implementations

**Labels**: `enhancement`, `testing`, `infrastructure`

**Description**:
Create a dice roller interface to abstract randomness in the combat system, enabling deterministic testing.

**Acceptance Criteria**:
- [ ] Define `DiceRoller` interface in `internal/dice/interface.go`
- [ ] Create `RealDiceRoller` implementation using current dice logic
- [ ] Create `MockDiceRoller` for testing with predetermined results
- [ ] Create `FixedDiceRoller` that always returns specific values
- [ ] Update existing dice package to use the interface
- [ ] Add unit tests for all implementations

**Technical Details**:
```go
type DiceRoller interface {
    Roll(count, size, modifier int) (*RollResult, error)
    RollWithAdvantage(count, size, modifier int) (*RollResult, error)
    RollWithDisadvantage(count, size, modifier int) (*RollResult, error)
}
```

---

### Issue 2: Implement Combat Event System
**Title**: Create Event-Driven System for Combat Updates

**Labels**: `enhancement`, `architecture`, `combat`

**Description**:
Implement an event bus system to decouple combat logic from UI updates, making the system more testable and maintainable.

**Acceptance Criteria**:
- [ ] Define event types and interfaces in `internal/combat/events/`
- [ ] Implement in-memory event bus
- [ ] Create event builders for common combat events
- [ ] Add event publishing to existing combat flow
- [ ] Create event handler interface for UI updates
- [ ] Add comprehensive tests for event system

**Technical Details**:
- Events should be immutable
- Event bus should support multiple handlers per event type
- Consider using channels for real-time updates
- Include context for cancellation

---

### Issue 3: Build Combat State Machine
**Title**: Implement State Machine for Combat Flow Control

**Labels**: `enhancement`, `combat`, `architecture`

**Description**:
Create a state machine to manage combat states and transitions, preventing invalid states and making flow predictable.

**Acceptance Criteria**:
- [ ] Define state machine interface and states
- [ ] Implement transition validation logic
- [ ] Create state diagram documentation
- [ ] Integrate with existing combat service
- [ ] Add state persistence to encounter entity
- [ ] Write comprehensive state transition tests

**States**:
- Initiating → Rolling Initiative → In Progress → Round Pending → Complete
- Handle error states and recovery

---

## Phase 2: Service Refactoring

### Issue 4: Extract Combat Service Interface
**Title**: Refactor Combat Service with Testable Interface

**Labels**: `refactoring`, `testing`, `combat`

**Description**:
Extract interface from existing encounter service and refactor for better testability and separation of concerns.

**Acceptance Criteria**:
- [ ] Create `CombatService` interface separate from encounter management
- [ ] Move combat-specific logic to new service
- [ ] Inject dependencies (dice roller, event bus, repositories)
- [ ] Update existing handlers to use new service
- [ ] Maintain backward compatibility
- [ ] Add service-level tests

**Breaking Changes**: None - maintain existing APIs

---

### Issue 5: Implement Turn Processor
**Title**: Create Dedicated Turn Processing System

**Labels**: `enhancement`, `combat`, `testing`

**Description**:
Extract turn processing logic into a dedicated, testable component that handles both player and monster turns.

**Acceptance Criteria**:
- [ ] Define `TurnProcessor` interface
- [ ] Implement player turn processing with validation
- [ ] Implement monster turn processing with AI
- [ ] Add action validation logic
- [ ] Create turn result structures
- [ ] Write comprehensive tests for all turn types

**Key Features**:
- Validate actions before processing
- Support different action types (attack, move, spell, etc.)
- Calculate and apply effects
- Generate appropriate events

---

### Issue 6: Create Combat Repository Interface
**Title**: Implement Repository Pattern for Combat Data

**Labels**: `refactoring`, `persistence`, `testing`

**Description**:
Create a clean repository interface for combat data access, enabling easy mocking and testing.

**Acceptance Criteria**:
- [ ] Define repository interface with atomic operations
- [ ] Implement Redis-based repository
- [ ] Create in-memory repository for testing
- [ ] Add transaction support for complex updates
- [ ] Implement proper error handling
- [ ] Write repository tests

**Atomic Operations**:
- Update turn order
- Update combatant HP
- Append to combat log
- Update round/turn counters

---

## Phase 3: Testing Framework

### Issue 7: Build Combat Test Builders
**Title**: Create Test Builder Pattern for Combat Scenarios

**Labels**: `testing`, `developer-experience`

**Description**:
Implement builder pattern for easily creating test combat scenarios with fluent API.

**Acceptance Criteria**:
- [ ] Create `CombatBuilder` for setting up encounters
- [ ] Create `CombatantBuilder` for players and monsters
- [ ] Add preset scenarios (1v1, party vs monsters, etc.)
- [ ] Include dice roll sequences in builders
- [ ] Document usage with examples
- [ ] Create test helpers package

**Example Usage**:
```go
combat := NewCombatBuilder().
    WithPlayer("Fighter", 20, 16).
    WithMonster("Goblin", 7, 15).
    WithDiceRolls(20, 15, 8, 6). // Predetermined rolls
    Build()
```

---

### Issue 8: Implement Integration Test Suite
**Title**: Create Comprehensive Combat Integration Tests

**Labels**: `testing`, `quality`

**Description**:
Build integration test suite covering all combat scenarios and edge cases.

**Acceptance Criteria**:
- [ ] Test basic combat flow (start to finish)
- [ ] Test round transitions with multiple combatants
- [ ] Test death and removal from combat
- [ ] Test concurrent actions
- [ ] Test error scenarios and recovery
- [ ] Test performance with large encounters
- [ ] Generate test coverage report

**Test Scenarios**:
1. Player defeats monster in 3 rounds
2. Monster defeats player
3. Multi-combatant with mixed initiative
4. Round transitions with pending state
5. Combat with conditions and effects
6. Healing during combat

---

## Phase 4: Monster AI System

### Issue 9: Extract and Enhance Monster AI
**Title**: Create Testable Monster AI System

**Labels**: `enhancement`, `combat`, `ai`

**Description**:
Extract monster decision-making into a separate, configurable AI system.

**Acceptance Criteria**:
- [ ] Define `MonsterAI` interface
- [ ] Implement basic AI (attack nearest)
- [ ] Add tactical AI (target lowest HP)
- [ ] Create AI configuration system
- [ ] Add AI decision logging
- [ ] Write AI behavior tests

**AI Strategies**:
- Basic: Attack nearest enemy
- Aggressive: Target lowest HP
- Defensive: Attack biggest threat
- Support: Heal/buff allies

---

### Issue 10: Implement Monster Action System
**Title**: Refactor Monster Actions for Flexibility

**Labels**: `enhancement`, `combat`

**Description**:
Create a flexible system for monster actions beyond basic attacks.

**Acceptance Criteria**:
- [ ] Define action priority system
- [ ] Support multi-attack monsters
- [ ] Add special ability usage
- [ ] Implement recharge mechanics
- [ ] Add legendary actions support
- [ ] Create action tests

---

## Phase 5: UI and Error Handling

### Issue 11: Implement Discord Event Handlers
**Title**: Create Event-Driven Discord UI Updates

**Labels**: `enhancement`, `ui`, `discord`

**Description**:
Build event handlers that update Discord UI based on combat events.

**Acceptance Criteria**:
- [ ] Create Discord event handler service
- [ ] Map combat events to UI updates
- [ ] Implement message caching for updates
- [ ] Add rate limiting for Discord API
- [ ] Handle stale message errors
- [ ] Test with mock Discord session

---

### Issue 12: Add Combat Recovery System
**Title**: Implement Graceful Recovery from Combat Errors

**Labels**: `enhancement`, `reliability`, `error-handling`

**Description**:
Create system to recover from errors and handle stale encounters gracefully.

**Acceptance Criteria**:
- [ ] Detect stale encounters on interaction
- [ ] Implement encounter recovery mechanism
- [ ] Add combat state validation
- [ ] Create rollback functionality
- [ ] Log errors with context
- [ ] Add recovery tests

**Error Scenarios**:
- Bot restart during combat
- Network interruption
- Invalid state transitions
- Concurrent modifications

---

## Bonus Issues

### Issue 13: Add Combat Metrics and Analytics
**Title**: Implement Combat Performance Metrics

**Labels**: `enhancement`, `monitoring`

**Description**:
Add metrics collection for combat performance and debugging.

**Acceptance Criteria**:
- [ ] Track combat duration
- [ ] Monitor turn processing time
- [ ] Count actions per combat
- [ ] Identify slow operations
- [ ] Create metrics dashboard
- [ ] Add performance tests

---

### Issue 14: Create Combat Debugging Tools
**Title**: Build Developer Tools for Combat Debugging

**Labels**: `developer-experience`, `tooling`

**Description**:
Create tools to help debug combat issues during development.

**Acceptance Criteria**:
- [ ] Combat state inspector command
- [ ] Event log viewer
- [ ] State machine visualizer
- [ ] Turn history replay
- [ ] Combat snapshot/restore
- [ ] Debug mode with verbose logging

---

## Implementation Order

1. **Week 1**: Issues 1-3 (Core Infrastructure)
2. **Week 2**: Issues 4-6 (Service Refactoring)
3. **Week 3**: Issues 7-8 (Testing Framework)
4. **Week 4**: Issues 9-10 (Monster AI)
5. **Week 5**: Issues 11-12 (UI and Error Handling)
6. **Ongoing**: Issues 13-14 (Bonus features)

## Success Metrics

- All combat scenarios have integration tests
- Test coverage > 80% for combat code
- Combat bugs reduced by 90%
- Turn processing time < 100ms
- Zero invalid state transitions
- Monster turns process correctly 100% of the time