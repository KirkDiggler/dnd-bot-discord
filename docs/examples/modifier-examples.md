# Modifier Implementation Examples

This document shows concrete examples of how different types of modifiers would be implemented using the event-driven system.

## Event Types Reference

```go
// Core character events that modifiers can listen to
const (
    // Combat Events
    OnAttackRoll     = "attack_roll"      // Before attack roll is made
    OnAttackHit      = "attack_hit"       // After successful hit
    OnAttackMiss     = "attack_miss"      // After attack misses
    OnDamageRoll     = "damage_roll"      // Before damage is calculated
    OnTakeDamage     = "take_damage"      // Before damage is applied to target
    OnDealDamage     = "deal_damage"      // After damage is dealt
    OnHeal           = "heal"             // Before healing is applied
    
    // Movement Events
    OnMove           = "move"             // Character moves
    OnMoveStart      = "move_start"       // Before movement begins
    OnMoveEnd        = "move_end"         // After movement completes
    
    // Action Events
    OnUseAbility     = "use_ability"      // Ability is activated
    OnCastSpell      = "cast_spell"       // Spell is cast
    OnTurnStart      = "turn_start"       // Turn begins
    OnTurnEnd        = "turn_end"         // Turn ends
    
    // State Events
    OnEquipItem      = "equip_item"       // Item equipped
    OnRestStart      = "rest_start"       // Rest begins
    OnRestComplete   = "rest_complete"    // Rest finishes
    OnLevelUp        = "level_up"         // Character gains level
)
```

## 1. Duration-Based Modifiers (Spell Effects)

### Bless Spell
**Effect**: +1d4 to attack rolls and saving throws for 1 minute

```go
type BlessModifier struct {
    targetID     string
    tracker      *DurationTracker
    diceRoller   dice.Roller
}

func NewBlessModifier(targetID string) *BlessModifier {
    return &BlessModifier{
        targetID: targetID,
        tracker: &DurationTracker{
            RemainingMinutes: 1,
            Type: DurationMinutes,
        },
        diceRoller: dice.NewRandomRoller(),
    }
}

func (b *BlessModifier) HandleEvent(event *GameEvent) error {
    // Only affect our target
    if event.Actor.ID != b.targetID {
        return nil
    }
    
    // Check if effect is still active
    if !b.tracker.IsActive() {
        return ErrModifierExpired
    }
    
    switch event.Type {
    case OnAttackRoll:
        bonus, _ := b.diceRoller.Roll(1, 4, 0)
        event.SetContext("bless_bonus", bonus.Total)
        
    case OnSavingThrow: 
        bonus, _ := b.diceRoller.Roll(1, 4, 0)
        event.SetContext("bless_bonus", bonus.Total)
        
    case OnTurnEnd:
        b.tracker.AdvanceTime(6) // 6 seconds per turn
        if !b.tracker.IsActive() {
            return ErrModifierExpired // Signals removal
        }
    }
    
    return nil
}
```

### Rage (Conditional Cancellation)
**Effect**: +2 damage, resistance to physical, ends if no damage dealt/taken

```go
type RageModifier struct {
    characterID      string
    damageBonus      int
    tracker          *DurationTracker
    damageTaken      bool
    damageDealt      bool
}

func (r *RageModifier) HandleEvent(event *GameEvent) error {
    if event.Actor.ID != r.characterID {
        return nil
    }
    
    switch event.Type {
    case OnDamageRoll:
        // Add rage damage to melee attacks
        weapon, _ := event.GetContext("weapon")
        if weapon.IsMelee() {
            event.SetContext("rage_damage", r.damageBonus)
            r.damageDealt = true
        }
        
    case OnTakeDamage:
        // Apply resistance to physical damage
        damageType, _ := event.GetStringContext("damage_type")
        if isPhysicalDamage(damageType) {
            currentDamage, _ := event.GetIntContext("damage")
            event.SetContext("damage", currentDamage/2)
        }
        r.damageTaken = true
        
    case OnTurnEnd:
        // Check if rage should end
        if !r.damageTaken && !r.damageDealt {
            return ErrModifierExpired // End rage
        }
        
        // Reset for next turn
        r.damageTaken = false
        r.damageDealt = false
        
        // Check duration
        r.tracker.AdvanceRounds(1)
        if !r.tracker.IsActive() {
            return ErrModifierExpired
        }
    }
    
    return nil
}
```

## 2. Resource-Based Modifiers (Consumed on Use)

### Ki Points - Flurry of Blows
**Effect**: Spend 1 ki for bonus action unarmed strikes

```go
type FlurryOfBlowsModifier struct {
    characterID string
    kiTracker   *PoolTracker
}

func (f *FlurryOfBlowsModifier) HandleEvent(event *GameEvent) error {
    if event.Actor.ID != f.characterID {
        return nil
    }
    
    switch event.Type {
    case OnAttackRoll:
        // After Attack action with unarmed strike, enable flurry
        actionType, _ := event.GetStringContext("action_type")
        weapon, _ := event.GetContext("weapon")
        
        if actionType == "action" && weapon.IsUnarmed() {
            if f.kiTracker.CanSpend(1) {
                // Enable bonus action attacks
                event.SetContext("flurry_available", true)
            }
        }
        
    case OnUseAbility:
        abilityKey, _ := event.GetStringContext("ability_key")
        if abilityKey == "flurry_of_blows" {
            if f.kiTracker.Spend(1) {
                // Grant two bonus action unarmed strikes
                event.SetContext("bonus_attacks", 2)
            } else {
                return ErrInsufficientResources
            }
        }
        
    case OnRestComplete:
        restType, _ := event.GetStringContext("rest_type")
        if restType == "short" || restType == "long" {
            f.kiTracker.RestoreAll()
        }
    }
    
    return nil
}
```

### Spell Slots
**Effect**: Consume spell slots to cast spells

```go
type SpellSlotModifier struct {
    characterID string
    slots       map[int]*PoolTracker // level -> tracker
}

func (s *SpellSlotModifier) HandleEvent(event *GameEvent) error {
    if event.Actor.ID != s.characterID {
        return nil
    }
    
    switch event.Type {
    case OnCastSpell:
        spellLevel, _ := event.GetIntContext("spell_level")
        slotUsed, _ := event.GetIntContext("slot_level")
        
        // Must use slot of equal or higher level
        if slotUsed < spellLevel {
            return ErrInvalidSlotLevel
        }
        
        tracker := s.slots[slotUsed]
        if !tracker.CanSpend(1) {
            return ErrNoSpellSlots
        }
        
        tracker.Spend(1)
        event.SetContext("slot_consumed", slotUsed)
        
    case OnRestComplete:
        restType, _ := event.GetStringContext("rest_type")
        if restType == "long" {
            for _, tracker := range s.slots {
                tracker.RestoreAll()
            }
        }
    }
    
    return nil
}
```

## 3. Movement-Triggered Modifiers

### Bleed Effect
**Effect**: Take damage and move slower when moving

```go
type BleedModifier struct {
    targetID       string
    damagePerMove  int
    speedReduction int
    tracker        *DurationTracker
}

func (b *BleedModifier) HandleEvent(event *GameEvent) error {
    if event.Actor.ID != b.targetID {
        return nil
    }
    
    if !b.tracker.IsActive() {
        return ErrModifierExpired
    }
    
    switch event.Type {
    case OnMoveStart:
        // Reduce movement speed
        currentSpeed, _ := event.GetIntContext("speed")
        newSpeed := max(0, currentSpeed-b.speedReduction)
        event.SetContext("speed", newSpeed)
        event.SetContext("bleed_speed_reduction", b.speedReduction)
        
    case OnMove:
        // Take damage for moving
        distance, _ := event.GetIntContext("distance")
        if distance > 0 {
            damageEvent := NewGameEvent(OnTakeDamage).
                WithActor(event.Actor).
                WithContext("damage", b.damagePerMove).
                WithContext("damage_type", "necrotic").
                WithContext("source", "bleed")
            
            // Emit damage event
            return EmitEvent(damageEvent)
        }
        
    case OnTurnEnd:
        b.tracker.AdvanceRounds(1)
        if !b.tracker.IsActive() {
            return ErrModifierExpired
        }
        
    case OnHeal:
        // Healing can remove bleed
        healType, _ := event.GetStringContext("heal_type")
        if healType == "magical" {
            return ErrModifierExpired // Remove bleed
        }
    }
    
    return nil
}
```

### Difficult Terrain
**Effect**: Double movement cost in certain areas

```go
type DifficultTerrainModifier struct {
    affectedArea Area
}

func (d *DifficultTerrainModifier) HandleEvent(event *GameEvent) error {
    switch event.Type {
    case OnMoveStart:
        currentPos, _ := event.GetContext("current_position")
        targetPos, _ := event.GetContext("target_position")
        
        // Check if movement crosses difficult terrain
        if d.affectedArea.Intersects(currentPos, targetPos) {
            // Double movement cost
            cost, _ := event.GetIntContext("movement_cost")
            event.SetContext("movement_cost", cost*2)
            event.SetContext("difficult_terrain", true)
        }
    }
    
    return nil
}
```

## 4. Complex Conditional Modifiers

### Sneak Attack
**Effect**: Add damage if advantage or ally adjacent, once per turn

```go
type SneakAttackModifier struct {
    characterID   string
    damageLevel   int // scales with rogue level
    usageTracker  *TurnUsageTracker
}

func (s *SneakAttackModifier) HandleEvent(event *GameEvent) error {
    if event.Actor.ID != s.characterID {
        return nil
    }
    
    switch event.Type {
    case OnAttackRoll:
        // Check if we can sneak attack
        if !s.usageTracker.CanUse() {
            return nil // Already used this turn
        }
        
        weapon, _ := event.GetContext("weapon")
        hasAdvantage, _ := event.GetBoolContext("has_advantage")
        hasDisadvantage, _ := event.GetBoolContext("has_disadvantage")
        allyAdjacent, _ := event.GetBoolContext("ally_adjacent")
        
        canSneakAttack := (weapon.IsFinesse() || weapon.IsRanged()) &&
                         (hasAdvantage && !hasDisadvantage) || allyAdjacent
        
        if canSneakAttack {
            event.SetContext("sneak_attack_available", true)
        }
        
    case OnAttackHit:
        available, _ := event.GetBoolContext("sneak_attack_available")
        if available {
            // Calculate sneak attack damage
            dice := (s.damageLevel + 1) / 2 // 1d6 per 2 levels
            damage, _ := dice.Roll(dice, 6, 0)
            
            event.SetContext("sneak_attack_damage", damage.Total)
            s.usageTracker.Use()
        }
        
    case OnTurnStart:
        s.usageTracker.Reset() // Can sneak attack again
    }
    
    return nil
}
```

### Magic Weapon
**Effect**: +1 to attack and damage, counts as magical

```go
type MagicWeaponModifier struct {
    weaponID string
    bonus    int
    tracker  *DurationTracker
}

func (m *MagicWeaponModifier) HandleEvent(event *GameEvent) error {
    weapon, _ := event.GetContext("weapon")
    if weapon.ID != m.weaponID {
        return nil // Not affecting this weapon
    }
    
    if !m.tracker.IsActive() {
        return ErrModifierExpired
    }
    
    switch event.Type {
    case OnAttackRoll:
        // Add bonus to attack
        event.SetContext("magic_attack_bonus", m.bonus)
        
    case OnDamageRoll:
        // Add bonus to damage  
        event.SetContext("magic_damage_bonus", m.bonus)
        
    case OnTakeDamage:
        // Mark damage as magical (for resistance purposes)
        event.SetContext("damage_magical", true)
        
    case OnTurnEnd:
        m.tracker.AdvanceRounds(1)
        if !m.tracker.IsActive() {
            return ErrModifierExpired
        }
    }
    
    return nil
}
```

## 5. What Can Be Modified

### Attack Events
```go
// OnAttackRoll - Can modify:
event.SetContext("attack_bonus", 5)        // Add to attack roll
event.SetContext("advantage", true)        // Grant advantage  
event.SetContext("disadvantage", true)     // Impose disadvantage
event.SetContext("critical_range", 19)     // Expand crit range
event.SetContext("auto_hit", true)         // Force hit
event.SetContext("auto_miss", true)        // Force miss

// OnDamageRoll - Can modify:
event.SetContext("damage_bonus", 4)        // Add flat damage
event.SetContext("damage_dice", "2d6")     // Add dice damage
event.SetContext("damage_type", "fire")    // Change damage type
event.SetContext("damage_magical", true)   // Mark as magical
event.SetContext("critical_multiplier", 3) // Change crit multiplier
```

### Defense Events
```go
// OnTakeDamage - Can modify:
event.SetContext("damage", newAmount)      // Change damage amount
event.SetContext("damage_type", "psychic") // Change damage type
event.SetContext("resistance", true)       // Apply resistance
event.SetContext("immunity", true)         // Apply immunity
event.SetContext("vulnerability", true)    // Apply vulnerability
event.SetContext("damage_negated", true)   // Negate all damage
```

### Movement Events
```go
// OnMove - Can modify:
event.SetContext("speed", 40)              // Change movement speed
event.SetContext("movement_cost", 10)      // Change cost per foot
event.SetContext("movement_blocked", true) // Prevent movement
event.SetContext("teleport", true)         // Change to teleportation
event.SetContext("difficult_terrain", true) // Mark as difficult
```

### Resource Events
```go
// OnUseAbility - Can modify:
event.SetContext("resource_cost", 2)       // Change resource cost
event.SetContext("ability_blocked", true)  // Prevent usage
event.SetContext("free_use", true)         // Make free this time
event.SetContext("enhanced_effect", true)  // Boost the effect
```

## 6. Modification Rules

### Direct vs Indirect Changes
```go
// WRONG: Direct modification bypasses event system
character.CurrentHP -= damage

// RIGHT: Emit event, let modifiers handle it
damageEvent := NewGameEvent(OnTakeDamage).
    WithActor(character).
    WithContext("damage", damage).
    WithContext("damage_type", "fire")
eventBus.Emit(damageEvent)

// Then apply the final modified damage
finalDamage, _ := damageEvent.GetIntContext("damage")
character.CurrentHP -= finalDamage
```

### Modification Priority
```go
// Modifiers run in priority order (lower numbers first)
type PriorityModifier interface {
    Priority() int // 0-999, lower runs first
}

// Priority guidelines:
// 0-99:   Core mechanics (weapon properties, class features)
// 100-199: Spells and magical effects  
// 200-299: Conditions and temporary effects
// 300-399: Equipment bonuses
// 400-499: Environmental effects
// 500+:    Situational modifiers
```

### Modification Stacking
```go
// Multiple modifiers can affect the same value
event.SetContext("damage_bonus", 2)  // Rage: +2
event.SetContext("damage_bonus", 1)  // Magic weapon: +1
// Result: damage + 2 + 1 = damage + 3

// Or use arrays for complex stacking
bonuses := event.GetIntArray("damage_bonuses")
bonuses = append(bonuses, 2) // Rage
bonuses = append(bonuses, 1) // Magic weapon
event.SetContext("damage_bonuses", bonuses)
```

This system provides incredible flexibility while maintaining clean separation between core mechanics and game rules!