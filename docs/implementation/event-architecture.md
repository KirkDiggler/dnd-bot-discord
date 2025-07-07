# Event-Driven Architecture Documentation

## Overview

The D&D bot uses the **rpg-toolkit** event bus for implementing cross-cutting game mechanics. This event-driven architecture allows abilities, feats, and features to react to game events without tight coupling between systems.

## Event Bus Integration

### Core Components

**Event Bus**: `github.com/KirkDiggler/rpg-toolkit/events`
- Central message broker for game events
- Priority-based subscription system
- Synchronous event processing
- Event cancellation support

**Adapter Layer**: `/internal/adapters/rpgtoolkit/`
- Wraps D&D entities as rpg-toolkit entities
- Provides helper functions for event emission
- Maps between internal and toolkit event types

## Event Types

### Combat Events

```go
// Attack sequence
rpgevents.EventBeforeAttackRoll    // Can modify attack roll
rpgevents.EventOnAttackRoll        // Attack roll made
rpgevents.EventAfterAttackRoll     // Attack result determined

// Damage sequence
rpgevents.EventOnDamageRoll        // Damage being rolled
rpgevents.EventBeforeTakeDamage    // Can modify incoming damage
rpgevents.EventOnTakeDamage        // Damage applied

// Turn management
rpgevents.EventOnTurnStart         // Character's turn begins
rpgevents.EventOnTurnEnd           // Character's turn ends
```

### Spell Events

```go
rpgevents.EventOnSpellCast         // Spell being cast
rpgevents.EventOnSpellDamage       // Spell damage rolled
```

### Status Events

```go
rpgevents.EventOnConditionApplied  // Status effect applied
rpgevents.EventOnConditionRemoved  // Status effect removed
```

### Custom Events

```go
"dndbot.rage_activated"            // Rage ability activated
"dndbot.rage_deactivated"          // Rage ended
"dndbot.sneak_attack_applied"      // Sneak attack damage added
```

## Event Context

Events carry contextual data accessed via `event.Context()`:

### Common Context Keys

```go
// Core combat data
ContextDamage           = "damage"          // int: damage amount
ContextDamageType       = "damage_type"     // string: damage type
ContextWeapon           = "weapon"          // string: weapon key
ContextAttackBonus      = "attack_bonus"    // int: total attack bonus
ContextIsCritical       = "is_critical"     // bool: critical hit

// Advantage/Disadvantage
ContextHasAdvantage     = "has_advantage"    // bool
ContextHasDisadvantage  = "has_disadvantage" // bool

// Turn tracking
ContextRound            = "round"            // int: current round
ContextTurnCount        = "turn_count"       // int: total turns
```

## Implementation Patterns

### 1. Event Emission

```go
// Basic event emission
EmitEvent(bus, rpgevents.EventOnAttackRoll, attacker, target, map[string]interface{}{
    "weapon": weaponKey,
    "attack_bonus": attackBonus,
    "total_attack": totalRoll,
})

// Event with cancellation check
event, err := CreateAndEmitEvent(bus, rpgevents.EventBeforeAttackRoll, attacker, target, context)
if event.IsCancelled() {
    return // Attack prevented
}
```

### 2. Event Subscription

```go
// Subscribe to events (in ability/feat registration)
bus.SubscribeFunc(rpgevents.EventOnDamageRoll, 50, func(ctx context.Context, event rpgevents.Event) error {
    // Extract actor
    actor, ok := rpgtoolkit.ExtractCharacter(event.Source())
    if !ok || actor.ID != characterID {
        return nil
    }
    
    // Modify damage
    damage, _ := rpgtoolkit.GetIntContext(event, ContextDamage)
    event.Context().Set(ContextDamage, damage + bonusDamage)
    
    return nil
})
```

### 3. Priority System

Handlers execute in priority order (lower number = higher priority):

```go
Priority 10:  Critical hit determination
Priority 30:  Resistance application
Priority 50:  Damage bonuses (rage, etc.)
Priority 70:  UI updates
Priority 100: Logging
```

## Entity Wrapping

D&D characters are wrapped to implement rpg-toolkit's Entity interface:

```go
// Wrap character for event system
entity := rpgtoolkit.WrapCharacter(character)

// Extract character from event
if char, ok := rpgtoolkit.ExtractCharacter(event.Source()); ok {
    // Process character-specific logic
}
```

## Event Flow Examples

### Attack with Rage Active

```
1. Player clicks "Attack" → Select target
2. BeforeAttackRoll event
   - Great Weapon Master can modify (-5 attack)
   - Other feats can add advantage
3. Attack roll made
4. OnAttackRoll event
   - Log the roll
5. AfterAttackRoll event
   - Determine hit/miss
6. OnDamageRoll event (if hit)
   - Rage adds +2 damage (priority 50)
   - Great Weapon Master adds +10 (priority 50)
   - Sneak Attack adds dice (priority 60)
7. BeforeTakeDamage event
   - Target's resistance applied (priority 30)
   - Defensive abilities trigger
8. OnTakeDamage event
   - Damage applied to target
   - Combat log updated
```

### Spell Casting Flow

```
1. Cast Magic Missile
2. OnSpellCast event
   - Counterspell opportunity (future)
   - War Caster feat bonuses
3. For each missile:
   - OnSpellDamage event
   - Damage rolled and applied
4. Status effects applied if applicable
```

## Best Practices

### Do:
- Use appropriate event types from rpg-toolkit
- Set meaningful priorities for handlers
- Check actor/target IDs to filter relevant events
- Use helper functions for context extraction
- Clean up subscriptions when appropriate

### Don't:
- Modify event data that other handlers depend on
- Use very low priorities unless necessary
- Emit events from within event handlers (avoid loops)
- Assume event order beyond priority guarantees
- Store state in event handlers (use character/effect systems)

## Integration Points

### Abilities
- Rage subscribes to damage events
- Second Wind doesn't use events (instant effect)
- Divine Sense could emit detection events

### Feats
- Great Weapon Master modifies attacks
- Alert modifies initiative
- Lucky could intercept roll events

### Features
- Sneak Attack checks conditions on damage roll
- Favored Enemy could modify tracking/damage vs types
- Unarmored Defense is passive (no events)

## Future Enhancements

1. **Reaction System**: Use events for opportunity attacks, counterspell
2. **Aura Effects**: Subscribe to movement/proximity events
3. **Conditional Triggers**: "When ally drops to 0 HP" type effects
4. **Battle Master Maneuvers**: Modify attacks with special effects
5. **Environmental Effects**: Area damage, difficult terrain

## Debugging

### Event Logging
```go
// Enable verbose event logging
log.Printf("[EVENT] %s: %s → %s (context: %+v)", 
    event.Type(), 
    event.Source().Name(), 
    event.Target().Name(),
    event.Context())
```

### Common Issues
1. **Handler not firing**: Check priority and actor ID filtering
2. **Modifications not applying**: Ensure correct context key usage
3. **Infinite loops**: Avoid emitting events from handlers
4. **Race conditions**: Event bus is synchronous, but beware of state changes

The event-driven architecture provides powerful extensibility while maintaining clean separation between game systems. New mechanics can be added by subscribing to appropriate events without modifying core combat logic.