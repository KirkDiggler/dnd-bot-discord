# Character Actions Implementation Guide

## Overview

This guide documents how character actions are implemented in the D&D Discord bot, from user interaction through game mechanics execution. It serves as the central reference for understanding the complete action flow.

## Architecture Overview

```
Discord UI → Handlers → Services → Domain Logic → Effects/Events → State Updates
```

### Key Components

1. **Discord Handlers** (`/internal/handlers/discord/`)
   - Process Discord interactions (buttons, commands)
   - Manage UI state and ephemeral messages
   - Route actions to appropriate services

2. **Service Layer** (`/internal/services/`)
   - Business logic orchestration
   - Transaction boundaries
   - Cross-domain coordination

3. **Domain Layer** (`/internal/domain/`)
   - Core game rules and entities
   - Character abilities and combat mechanics
   - Effect management

4. **Event System** (`rpg-toolkit`)
   - Cross-cutting concerns
   - Reactive mechanics
   - Loose coupling between systems

## Action Types

### 1. Combat Actions

**Attack Action**
- UI: "Attack" button → target selection
- Flow: `handleAttack()` → `encounterService.ExecuteAttack()`
- Consumes: Action
- Events: BeforeAttackRoll, OnDamageRoll, BeforeTakeDamage

**Bonus Action Attack**
- UI: Automatic after main attack (TWF, Martial Arts)
- Flow: `handleBonusAttack()` → same as attack
- Consumes: Bonus Action
- Special: No ability modifier to damage (unless TWF style)

### 2. Ability Actions

**Active Abilities**
- UI: "Abilities" button → ability selection
- Flow: `handleUseAbility()` → `abilityService.UseAbility()`
- Consumes: Varies (Action, Bonus Action, None)
- Examples: Rage, Second Wind, Divine Smite

**Targeted Abilities**
- Additional target selection step
- Examples: Bardic Inspiration, Healing Word
- Flow includes target validation

### 3. Movement Actions
- Currently not implemented
- Planned: Grid-based movement system

### 4. Other Actions
- **Dodge**: Disadvantage on attacks against you
- **Dash**: Double movement (future)
- **Help**: Grant advantage to ally (future)

## Action Economy

### Resource Tracking
```go
type ActionEconomy struct {
    ActionsRemaining   int  // Usually 1, 2 with Action Surge
    BonusActionUsed    bool
    ReactionUsed       bool
    MovementRemaining  int  // Future
}
```

### Turn Flow
1. Start of turn → Reset action economy
2. Player actions → Consume resources
3. End of turn → Trigger end-of-turn effects
4. Next combatant → Repeat

## UI Integration

### Message Types

**Shared Combat Message**
- Visible to all players
- Shows combat state, turn order
- Updated after each action

**Ephemeral Action Controller**
- Private to each player
- Shows available actions
- "Get My Actions" button access

### Button States
- **Enabled**: Action available, resources exist
- **Disabled**: Not your turn, no resources, or invalid
- **Hidden**: Action not applicable

## Implementation Patterns

### Adding a New Action Type

1. **Define the UI**
```go
// In combat handler
func (h *Handler) createActionButton(actionType string) *discordgo.Button {
    return &discordgo.Button{
        Label:    "Action Name",
        Style:    discordgo.PrimaryButton,
        CustomID: fmt.Sprintf("combat_action_%s", actionType),
        Disabled: !h.canPerformAction(actionType),
    }
}
```

2. **Create Handler Method**
```go
func (h *Handler) handleNewAction(i *discordgo.InteractionCreate) error {
    // Validate turn and resources
    // Execute action logic
    // Update combat state
    // Refresh UI
}
```

3. **Implement Service Logic**
```go
func (s *Service) ExecuteNewAction(input ActionInput) (*ActionResult, error) {
    // Core game logic
    // Event emission
    // State updates
}
```

4. **Register with Event System** (if needed)
```go
bus.SubscribeFunc(rpgevents.EventOnNewAction, priority, handler)
```

## State Management

### Character State
- Persisted to database after actions
- Includes HP, resources, effects
- Synchronized across all systems

### Combat State
- Maintained in encounter service
- Tracks turn order, round count
- Updates after each action

### UI State
- Discord message IDs tracked
- Ephemeral responses per user
- Bulk updates for efficiency

## Error Handling

### Validation Layers
1. **UI Level**: Button disabled states
2. **Handler Level**: Turn and permission checks
3. **Service Level**: Business rule validation
4. **Domain Level**: Game rule enforcement

### Recovery
- Actions are atomic (all or nothing)
- Resources restored on failure
- Clear error messages to users

## Performance Considerations

1. **Batch Operations**: Update multiple UI elements together
2. **Event Priorities**: Critical events process first
3. **Caching**: Active effects cached on character
4. **Database**: Optimize queries, batch updates

## Testing Strategy

### Unit Tests
- Test each handler method
- Mock Discord interactions
- Verify service calls

### Integration Tests
- Full action flows
- Event system integration
- State persistence

### Manual Testing
- Discord bot interaction
- Multi-player scenarios
- Edge cases

## Common Issues and Solutions

### "Not Your Turn"
- Ensure turn validation in handler
- Check encounter state consistency
- Verify player-character mapping

### Missing Actions
- Check action economy state
- Verify ability prerequisites
- Ensure proper registration

### UI Not Updating
- Check ephemeral vs shared messages
- Verify message ID tracking
- Ensure proper error handling

## Related Documentation

- [Combat Flow](./combat-flow.md) - Detailed combat calculations
- [Event Architecture](./event-architecture.md) - Event system details
- [Effect Systems](./effect-systems.md) - Effect management
- [Rulebook System](./rulebook-system.md) - Abilities and features

## Future Enhancements

1. **Reaction System**: Opportunity attacks, Shield spell
2. **Ritual Casting**: Long-duration actions
3. **Mounted Combat**: Special movement rules
4. **Environmental Actions**: Use objects, terrain
5. **Social Actions**: Persuasion, deception in combat

This documentation provides a comprehensive guide to implementing and understanding character actions in the D&D bot. Follow these patterns for consistency and maintainability.