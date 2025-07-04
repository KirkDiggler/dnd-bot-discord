# RPG Toolkit Event Bus Adapter Usage Examples

This document shows how to use the RPG Toolkit event bus adapter to integrate rpg-toolkit's event system with the Discord bot.

## Basic Setup

```go
package main

import (
    "github.com/KirkDiggler/dnd-bot-discord/internal/adapters/rpgtoolkit"
    "github.com/KirkDiggler/dnd-bot-discord/internal/domain/events"
)

func main() {
    // Create the adapter
    eventBus := rpgtoolkit.NewEventBusAdapter()
    
    // Use it like the regular Discord bot event bus
    // but get the benefits of rpg-toolkit's features
}
```

## Example 1: Subscribing to Attack Events

```go
// Subscribe to attack roll events
eventBus.Subscribe(events.BeforeAttackRoll, &AttackModifierHandler{})

// Handler that adds bless bonus to attacks
type AttackModifierHandler struct{}

func (h *AttackModifierHandler) Handle(event interface{}) {
    attackEvent, ok := event.(*events.AttackEvent)
    if !ok {
        return
    }
    
    // Check if attacker has bless
    if attackEvent.Attacker.HasStatusEffect("blessed") {
        // In rpg-toolkit, this would add a modifier to the event context
        // For now, we'd need to handle it in the Discord bot way
    }
}
```

## Example 2: Publishing Spell Events

```go
// When a spell is cast
spellEvent := &events.SpellEvent{
    Caster:     caster,
    Targets:    targets,
    SpellName:  "vicious-mockery",
    SpellLevel: 0, // Cantrip
    SaveDC:     13,
    SaveType:   "wisdom",
}

// Publish through the adapter
err := eventBus.Publish(events.OnSpellCast, spellEvent)
if err != nil {
    log.Printf("Failed to publish spell event: %v", err)
}
```

## Example 3: Using RPG Toolkit Features Directly

Since the adapter exposes the underlying rpg-toolkit bus, you can also use rpg-toolkit features directly:

```go
// Get the rpg-toolkit bus
rpgBus := eventBus.GetRPGBus()

// Create a disadvantage condition using rpg-toolkit
import (
    rpgevents "github.com/KirkDiggler/rpg-toolkit/events"
    "github.com/KirkDiggler/rpg-toolkit/mechanics/conditions"
)

// Subscribe directly to rpg-toolkit events
rpgBus.SubscribeFunc(rpgevents.EventBeforeAttackRoll, 100, func(ctx context.Context, e rpgevents.Event) error {
    // Check if attacker has disadvantage from vicious mockery
    if e.Source() != nil {
        // This is where we'd check for the disadvantage condition
        // and apply it to the attack roll
    }
    return nil
})
```

## Example 4: Vicious Mockery with Event System

```go
// In vicious_mockery.go
func (v *ViciousMockeryHandler) Execute(ctx context.Context, input *SpellInput) (*SpellResult, error) {
    // ... save logic ...
    
    if saveFailed {
        // Publish damage event
        damageEvent := &events.DamageEvent{
            Source:     input.Caster,
            Target:     target,
            Damage:     damage,
            DamageType: "psychic",
        }
        
        err := v.eventBus.Publish(events.OnSpellDamage, damageEvent)
        if err != nil {
            return nil, err
        }
        
        // Apply disadvantage effect
        // This would be better handled through rpg-toolkit conditions
        // but for now we use the existing system
        disadvantageEffect := effects.NewBuilder().
            WithName("Vicious Mockery").
            WithDuration(effects.DurationRounds, 1).
            WithCondition(effects.ConditionDisadvantageAttack, "").
            Build()
            
        target.AddStatusEffect(disadvantageEffect)
    }
    
    return result, nil
}
```

## Migration Path

1. **Phase 1**: Use the adapter as a drop-in replacement for the current event bus
2. **Phase 2**: Start using rpg-toolkit modifiers in event handlers
3. **Phase 3**: Migrate conditions/effects to rpg-toolkit's system
4. **Phase 4**: Remove Discord bot's event system entirely

## Benefits of Using the Adapter

1. **Cancellable Events**: Can implement counterspell by cancelling spell events
2. **Event Priorities**: Handlers execute in a defined order
3. **Better Modifiers**: Use rpg-toolkit's typed modifier system
4. **Future-Proof**: Easy migration path to full rpg-toolkit integration

## Testing

```go
func TestViciousMockeryWithAdapter(t *testing.T) {
    // Create adapter
    eventBus := rpgtoolkit.NewEventBusAdapter()
    
    // Subscribe to spell damage events
    var damageCaptured int
    eventBus.Subscribe(events.OnSpellDamage, &TestHandler{
        onHandle: func(e interface{}) {
            if dmgEvent, ok := e.(*events.DamageEvent); ok {
                damageCaptured = dmgEvent.Damage
            }
        },
    })
    
    // Cast vicious mockery
    handler := &ViciousMockeryHandler{eventBus: eventBus}
    result, err := handler.Execute(ctx, input)
    
    // Verify damage event was published
    assert.NoError(t, err)
    assert.Equal(t, expectedDamage, damageCaptured)
}
```