# Event-Driven Modifier System

## Core Concept

**Things just do things and emit events. Everything else listens and modifies.**

### The Flow
1. **Core entities** (Character, Encounter, Monster) perform actions
2. **Emit standardized events** when actions happen  
3. **Modifiers listen** to relevant events and apply changes
4. **No direct coupling** between entities and game rules

## Event Categories

### Character Events
```go
// Combat Actions
OnAttackRoll     // Character attempts attack
OnDamageRoll     // Character deals damage  
OnTakeDamage     // Character receives damage
OnHeal           // Character gets healed

// Movement & Actions
OnMove           // Character changes position
OnUseAbility     // Character activates ability
OnCastSpell      // Character casts spell
OnTurnStart      // Character's turn begins
OnTurnEnd        // Character's turn ends

// State Changes  
OnEquipItem      // Character equips gear
OnLevelUp        // Character gains level
OnRestStart      // Character begins rest
OnRestComplete   // Character finishes rest
```

### Encounter Events
```go
OnEncounterStart // Combat begins
OnEncounterEnd   // Combat ends
OnRoundStart     // New round begins
OnInitiativeRoll // Rolling for turn order
```

### Monster Events
```go
OnMonsterAttack  // Monster attacks
OnMonsterDeath   // Monster dies
OnMonsterSpawn   // Monster appears
```

## Modifier Integration

### Attack Modifiers
Listen to `OnAttackRoll` and modify:
```go
// Rage: +2 to melee damage rolls
rageModifier.Listen(OnDamageRoll) -> if melee && hasRage: +2

// Bless: +1d4 to attack rolls  
blessModifier.Listen(OnAttackRoll) -> if blessed: +1d4

// Sneak Attack: +Xd6 if conditions met
sneakModifier.Listen(OnDamageRoll) -> if advantage || allyAdjacent: +Xd6
```

### Defense Modifiers
Listen to `OnTakeDamage` and modify:
```go
// Rage: Resistance to physical damage
rageResistance.Listen(OnTakeDamage) -> if physical: damage / 2

// Shield spell: +5 AC against one attack
shieldSpell.Listen(OnAttackRoll) -> if targeted && hasReaction: +5 AC

// Uncanny Dodge: Half damage from one attack per turn
uncannyDodge.Listen(OnTakeDamage) -> if dex save && first this turn: damage / 2
```

### Resource Modifiers
Listen to various events and enable/modify:
```go
// Martial Arts: Bonus unarmed strike after Attack action
martialArts.Listen(OnAttackRoll) -> if unarmed && action: enable bonus action

// Action Surge: Extra action this turn
actionSurge.Listen(OnTurnStart) -> if activated: grant extra action

// Ki Points: Enable special abilities
ki.Listen(OnUseAbility) -> if flurry/dodge/dash: consume 1 ki
```

## Implementation Pattern

### Entities Stay Pure
```go
// Character.Attack() just does the mechanical action
func (c *Character) Attack(target *Character) *AttackResult {
    // Roll attack, calculate base damage
    result := &AttackResult{...}
    
    // Emit event - let modifiers handle the rest
    event := NewGameEvent(OnAttackRoll).
        WithActor(c).
        WithTarget(target).
        WithContext("weapon", weapon).
        WithContext("base_damage", baseDamage)
    
    eventBus.Emit(event)
    
    // Event listeners have modified the result
    return result
}
```

### Modifiers Listen and Modify
```go
// Rage modifier listens for damage events
type RageModifier struct {
    characterID string
    damageBonus int
}

func (r *RageModifier) HandleEvent(event *GameEvent) error {
    if event.Type != OnDamageRoll {
        return nil
    }
    
    if event.Actor.ID != r.characterID {
        return nil // Not our character
    }
    
    weapon, _ := event.GetContext("weapon")
    if weapon.IsMelee() {
        // Modify the damage in the event
        event.SetContext("damage_bonus", r.damageBonus)
    }
    
    return nil
}
```

### Rulesets Just Register Listeners
```go
// D&D 5e ruleset registers all its modifiers
func (d *DND5eRuleset) RegisterModifiers(bus *EventBus) {
    // Register all class features
    bus.Subscribe(OnDamageRoll, &RageModifier{})
    bus.Subscribe(OnAttackRoll, &SneakAttackModifier{})
    bus.Subscribe(OnTakeDamage, &UncannyDodgeModifier{})
    
    // Register all spells
    bus.Subscribe(OnAttackRoll, &BlessModifier{})
    bus.Subscribe(OnTakeDamage, &ShieldModifier{})
    
    // Register all conditions
    bus.Subscribe(OnAttackRoll, &PoisonedModifier{})
    bus.Subscribe(OnMove, &GrappledModifier{})
}
```

## Benefits

### For Core System
- **Zero game knowledge**: Character.Attack() doesn't know about rage, sneak attack, etc.
- **Consistent interface**: All modifiers use same event pattern
- **Easy testing**: Mock the event bus to test in isolation

### For Rulesets  
- **Pure composition**: Just register the right listeners
- **No code changes**: Add new abilities by adding new listeners
- **Multiple rulesets**: Different rulesets can coexist

### For Features
- **Emergent behavior**: Complex interactions arise from simple rules
- **Easy debugging**: Can log all events and modifier applications
- **Flexible UI**: Can show "why" damage was X by showing modifier chain

## Examples

### Adding New Ability
```go
// Adding Fighter's "Great Weapon Fighting" - reroll 1s and 2s on damage
type GreatWeaponFighting struct {
    characterID string
}

func (g *GreatWeaponFighting) HandleEvent(event *GameEvent) error {
    if event.Type != OnDamageRoll || event.Actor.ID != g.characterID {
        return nil
    }
    
    weapon, _ := event.GetContext("weapon")
    if weapon.IsTwoHanded() {
        // Reroll any 1s or 2s in damage dice
        // Modify the damage roll in the event
    }
    
    return nil
}

// To add this ability: just register the listener!
eventBus.Subscribe(OnDamageRoll, &GreatWeaponFighting{characterID: "fighter-123"})
```

### Complex Interactions Work Automatically
```go
// Barbarian with rage + great weapon fighting + magic weapon
// Each modifier listens and adds its piece:
// 1. Magic weapon: +1 to attack and damage
// 2. Great weapon fighting: reroll 1s and 2s  
// 3. Rage: +2 damage to melee attacks
// Final damage = base + magic + rage (with rerolled dice)
```

## Architecture Diagram

```
[Character.Attack()] 
        ↓ emits
[OnAttackRoll Event]
        ↓ flows to
[Event Bus] 
        ↓ notifies
[All Registered Modifiers]
        ↓ each applies
[BlessModifier: +1d4]
[AdvantageModifier: roll twice]  
[MagicWeaponModifier: +1]
        ↓ result
[Modified Attack Result]
```

## Integration with Tracking System

Modifiers can have their own trackers:
```go
type SneakAttackModifier struct {
    tracker *UsageTracker // Once per turn
    damage  string        // "3d6" 
}

func (s *SneakAttackModifier) HandleEvent(event *GameEvent) error {
    if !s.tracker.CanUse() {
        return nil // Already used this turn
    }
    
    if s.meetsConditions(event) {
        event.SetContext("sneak_attack_damage", s.damage)
        s.tracker.Use() // Mark as used
    }
    
    return nil
}
```

This gives us the best of both worlds:
- **Tracking** manages when/how often things can be used
- **Events** manage what happens when they are used