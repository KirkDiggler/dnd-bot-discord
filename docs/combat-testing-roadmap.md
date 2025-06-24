# Combat Testing Roadmap

## Overview
We've created a systematic approach to fixing combat issues through proper testing infrastructure rather than debugging in production.

## Created GitHub Issues

### Phase 1: Core Infrastructure (Foundation)
- **#85**: [Create Dice Roller Interface](https://github.com/KirkDiggler/dnd-bot-discord/issues/85) - Enable deterministic testing
- **#86**: [Implement Combat Event System](https://github.com/KirkDiggler/dnd-bot-discord/issues/86) - Decouple combat from UI
- **#87**: [Combat State Machine](https://github.com/KirkDiggler/dnd-bot-discord/issues/87) - Prevent invalid states

### Phase 2: Service Refactoring (Clean Architecture)
- **#88**: [Extract Combat Service](https://github.com/KirkDiggler/dnd-bot-discord/issues/88) - Separate combat logic
- **#89**: [Turn Processor](https://github.com/KirkDiggler/dnd-bot-discord/issues/89) - Fix turn order bugs

### Phase 3: Testing Framework (Quality Assurance)
- **#90**: [Integration Test Suite](https://github.com/KirkDiggler/dnd-bot-discord/issues/90) - Test combat scenarios
- **#91**: [Test Builder Pattern](https://github.com/KirkDiggler/dnd-bot-discord/issues/91) - Easy test creation

## Why This Approach?

### Current Problems We're Solving:
1. **"Round not pending" errors** → State machine prevents invalid transitions
2. **Dead monsters attacking** → Turn processor skips inactive combatants  
3. **UI not updating** → Event system ensures reliable updates
4. **Can't reproduce bugs** → Mock dice enables exact scenario testing
5. **Combat gets stuck** → Clear state transitions and recovery

### Benefits:
- **Deterministic Testing**: Mock dice roller lets us test exact scenarios
- **Clear Separation**: Combat logic separate from Discord UI
- **Event-Driven**: UI updates through events, not scattered calls
- **State Safety**: State machine prevents invalid combat states
- **Comprehensive Tests**: Can test edge cases and complex scenarios

## Implementation Order

1. **Start with #85 (Dice Interface)** - Quick win, enables all other testing
2. **Then #87 (State Machine)** - Fixes core state issues
3. **Then #89 (Turn Processor)** - Fixes turn order bugs
4. **Then #90 (Integration Tests)** - Validates everything works

## Example Test Scenario

With the new infrastructure, we can write tests like:

```go
func TestDeadMonstersSkipped(t *testing.T) {
    combat := NewCombatBuilder(t).
        WithPlayer("Fighter", 20, 16).
        WithMonster("Goblin 1", 7, 15).
        WithMonster("Goblin 2", 7, 15).
        WithInitiativeRolls(10, 15, 12). // Goblin 1, Goblin 2, Fighter
        Build()
    
    // Goblin 1 attacks and misses
    combat.ExpectTurn("Goblin 1").
        WithAttackRoll(5). // Miss
        AssertMiss()
    
    // Goblin 2 attacks and hits
    combat.ExpectTurn("Goblin 2").
        WithAttackRoll(18).
        WithDamage(5).
        AssertHit("Fighter", 5)
    
    // Fighter kills Goblin 1
    combat.ExpectTurn("Fighter").
        PlayerAttack("Goblin 1").
        WithRoll(16).
        WithDamage(8).
        AssertKill("Goblin 1")
    
    // Round 2 - Goblin 1 should be skipped
    combat.ContinueRound().
        ExpectTurn("Goblin 2"). // Skips dead Goblin 1
        AssertRound(2)
}
```

## Next Steps

1. Review the GitHub issues and adjust if needed
2. Start with issue #85 (Dice Roller Interface) - it's small and enables everything else
3. Each issue can be a separate PR
4. We can run the integration tests in CI to catch regressions

This approach gives us confidence that combat works correctly before deploying!