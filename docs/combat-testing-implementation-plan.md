# Combat Testing Infrastructure Implementation Plan

## Executive Summary

This plan addresses the critical issues in the combat system:
- Combat rounds not advancing properly
- Monster turns not being processed correctly
- UI not updating after round continuation
- Dead monsters still affecting turn order
- General combat flow confusion

By implementing a testable, event-driven architecture with proper separation of concerns, we'll create a reliable and maintainable combat system.

## Current Problems Analysis

### 1. Round Advancement Issues
**Root Cause**: State management is scattered across multiple components without clear ownership.
**Solution**: Implement a state machine with validated transitions.

### 2. Monster Turn Processing
**Root Cause**: Monster turn logic is embedded in UI handlers, making it hard to test and debug.
**Solution**: Extract turn processing into a dedicated service with AI strategies.

### 3. UI Update Failures
**Root Cause**: Tight coupling between combat logic and Discord UI updates.
**Solution**: Event-driven architecture with decoupled event handlers.

### 4. Dead Monster Turn Order
**Root Cause**: Turn order isn't properly updated when combatants are removed.
**Solution**: Atomic repository operations with proper state validation.

### 5. Combat Flow Confusion
**Root Cause**: No clear state machine or flow documentation.
**Solution**: Implement explicit state machine with transition rules.

## Architecture Benefits

### 1. Testability
- **Mock Dice Roller**: Test exact combat scenarios
- **Event Verification**: Ensure proper event flow
- **State Validation**: Prevent invalid states

### 2. Maintainability
- **Clear Interfaces**: Easy to understand and modify
- **Decoupled Components**: Changes don't cascade
- **Event Logging**: Built-in debugging

### 3. Reliability
- **State Machine**: Prevents invalid transitions
- **Error Recovery**: Graceful handling of failures
- **Atomic Operations**: Consistent state updates

### 4. Performance
- **Efficient Updates**: Only update what changed
- **Concurrent Safety**: Proper locking strategies
- **Optimized Queries**: Repository pattern enables caching

## Quick Start Guide

### For Developers

1. **Start with Core Infrastructure** (Week 1)
   - Implement dice roller interface
   - Build event system
   - Create state machine

2. **Refactor Services** (Week 2)
   - Extract interfaces
   - Add dependency injection
   - Integrate events

3. **Build Tests** (Week 3)
   - Create test builders
   - Write integration tests
   - Verify all scenarios

### For Testing

```go
// Example: Testing a combat scenario
func TestPlayerDefeatsGoblin(t *testing.T) {
    // Setup
    dice := NewMockDiceRoller()
    dice.SetRolls([]int{20, 15, 8, 6}) // Predetermined rolls
    
    combat := NewCombatBuilder().
        WithPlayer("Fighter", 20, 16).
        WithMonster("Goblin", 7, 15).
        WithDiceRoller(dice).
        Build()
    
    service := NewCombatService(
        WithDiceRoller(dice),
        WithEventBus(NewTestEventBus()),
    )
    
    // Act
    result, err := service.ProcessTurn(ctx, combat.ID, AttackAction{
        Target: "goblin-1",
    })
    
    // Assert
    assert.NoError(t, err)
    assert.True(t, result.Hit)
    assert.Equal(t, 8, result.Damage)
}
```

## Migration Strategy

### Phase 1: Parallel Implementation
- Build new system alongside existing
- No breaking changes
- Gradual migration

### Phase 2: Feature Flag Rollout
- Use feature flags for new combat system
- A/B test with select users
- Monitor metrics

### Phase 3: Full Migration
- Migrate all combat to new system
- Deprecate old code
- Clean up legacy

## Success Criteria

### Technical Metrics
- **Test Coverage**: >80% for combat code
- **Bug Reduction**: 90% fewer combat bugs
- **Performance**: <100ms turn processing
- **Reliability**: Zero invalid states

### User Experience
- **Smooth Combat**: No UI glitches
- **Fast Responses**: Instant feedback
- **Clear Flow**: Obvious what's happening
- **Error Recovery**: Graceful handling

## Risk Mitigation

### Risk 1: Breaking Existing Functionality
**Mitigation**: Comprehensive integration tests before migration

### Risk 2: Performance Degradation
**Mitigation**: Benchmark tests and optimization phase

### Risk 3: Complex Migration
**Mitigation**: Gradual rollout with feature flags

## Timeline

### Month 1: Foundation
- Week 1: Core infrastructure
- Week 2: Service refactoring
- Week 3: Testing framework
- Week 4: Integration and testing

### Month 2: Enhancement
- Week 1: Monster AI system
- Week 2: UI improvements
- Week 3: Error handling
- Week 4: Performance optimization

### Month 3: Polish
- Week 1: Monitoring and metrics
- Week 2: Developer tools
- Week 3: Documentation
- Week 4: Full rollout

## Conclusion

This combat testing infrastructure will transform the combat system from a source of bugs to a reliable, enjoyable feature. The investment in proper architecture and testing will pay dividends in reduced maintenance and improved user experience.

The modular approach allows incremental implementation while maintaining system stability. Each phase builds on the previous, creating a solid foundation for future enhancements.

## Next Steps

1. Review and approve the architecture
2. Create GitHub issues from the templates
3. Assign team members to Phase 1 issues
4. Set up weekly progress reviews
5. Begin implementation with Issue #1 (Dice Roller Interface)