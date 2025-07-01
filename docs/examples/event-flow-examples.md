# Event Flow Examples

This document shows step-by-step examples of how events flow through the modifier system for common game scenarios.

## Example 1: Barbarian Rage Attack

**Scenario**: Level 5 Barbarian with active Rage attacks a goblin with a greataxe

### Setup
```go
// Active modifiers on the barbarian
modifiers := []Modifier{
    &RageModifier{characterID: "barb-123", damageBonus: 2},
    &GreatWeaponFightingModifier{characterID: "barb-123"},
    &ProficiencyModifier{characterID: "barb-123", bonus: 3},
}

weapon := &Weapon{
    Name: "Greataxe", 
    Damage: "1d12", 
    DamageType: "slashing",
    Properties: []string{"heavy", "two-handed"},
}
```

### Event Flow

#### Step 1: Attack Roll
```go
// Character.Attack() emits initial event
attackEvent := NewGameEvent(OnAttackRoll).
    WithActor(barbarian).
    WithTarget(goblin).
    WithContext("weapon", weapon).
    WithContext("base_attack", 8) // STR modifier

eventBus.Emit(attackEvent)
```

#### Step 2: Modifiers Process Attack
```go
// ProficiencyModifier (Priority: 50)
func (p *ProficiencyModifier) HandleEvent(event) {
    if weapon.Category == "martial" && character.HasProficiency("martial-weapons") {
        event.SetContext("proficiency_bonus", 3)
    }
}

// No other modifiers affect attack rolls
// Final attack roll: 1d20 + 4 (STR) + 3 (prof) = 1d20 + 7
```

#### Step 3: Attack Hits
```go
// Roll: 15 + 7 = 22 vs AC 15 → Hit!
hitEvent := NewGameEvent(OnAttackHit).
    WithActor(barbarian).
    WithTarget(goblin).
    WithContext("weapon", weapon).
    WithContext("attack_roll", 22)

eventBus.Emit(hitEvent)
```

#### Step 4: Damage Roll
```go
damageEvent := NewGameEvent(OnDamageRoll).
    WithActor(barbarian).
    WithTarget(goblin).
    WithContext("weapon", weapon).
    WithContext("base_damage", "1d12").
    WithContext("ability_modifier", 4) // STR mod

eventBus.Emit(damageEvent)
```

#### Step 5: Modifiers Process Damage
```go
// GreatWeaponFightingModifier (Priority: 100)
func (g *GreatWeaponFightingModifier) HandleEvent(event) {
    if weapon.HasProperty("two-handed") {
        // Reroll 1s and 2s - rolled 2, reroll to 8
        event.SetContext("gwf_reroll", "2→8")
    }
}

// RageModifier (Priority: 150) 
func (r *RageModifier) HandleEvent(event) {
    if weapon.IsMelee() && r.tracker.IsActive() {
        event.SetContext("rage_damage", 2)
    }
}

// Final damage: 8 (rerolled) + 4 (STR) + 2 (rage) = 14 slashing
```

#### Step 6: Apply Damage
```go
takeDamageEvent := NewGameEvent(OnTakeDamage).
    WithActor(goblin).
    WithContext("damage", 14).
    WithContext("damage_type", "slashing").
    WithContext("source", barbarian)

eventBus.Emit(takeDamageEvent)

// No resistance modifiers on goblin
// Goblin takes 14 damage: 7 HP → -7 HP (dead)
```

### Result Log
```
Grunk attacks Goblin with Greataxe (+7 to hit): 22 vs AC 15 - HIT!
Damage: 1d12+4 = 8+4=12 +2 rage = 14 slashing damage (includes effects)
Goblin takes 14 damage and dies!
```

## Example 2: Rogue Sneak Attack with Magic Weapon

**Scenario**: Level 3 Rogue with +1 Rapier attacks surprised orc (has advantage)

### Setup
```go
modifiers := []Modifier{
    &SneakAttackModifier{characterID: "rogue-456", level: 3}, // 2d6
    &MagicWeaponModifier{weaponID: "rapier-789", bonus: 1},
    &ProficiencyModifier{characterID: "rogue-456", bonus: 2},
}

weapon := &Weapon{
    ID: "rapier-789",
    Name: "+1 Rapier",
    Damage: "1d8", 
    DamageType: "piercing",
    Properties: []string{"finesse"},
}
```

### Event Flow

#### Step 1: Attack with Advantage
```go
attackEvent := NewGameEvent(OnAttackRoll).
    WithActor(rogue).
    WithTarget(orc).
    WithContext("weapon", weapon).
    WithContext("base_attack", 3). // DEX modifier
    WithContext("has_advantage", true)

eventBus.Emit(attackEvent)
```

#### Step 2: Modifiers Process Attack
```go
// ProficiencyModifier
event.SetContext("proficiency_bonus", 2)

// MagicWeaponModifier  
event.SetContext("magic_attack_bonus", 1)

// SneakAttackModifier
func (s *SneakAttackModifier) HandleEvent(event) {
    weapon := event.GetContext("weapon")
    hasAdvantage := event.GetBoolContext("has_advantage")
    
    if weapon.HasProperty("finesse") && hasAdvantage && s.canUse() {
        event.SetContext("sneak_attack_available", true)
    }
}

// Attack: 2d20kh1 + 3 (DEX) + 2 (prof) + 1 (magic) = 2d20kh1 + 6
// Rolled: [18, 12] keep highest = 18 + 6 = 24 vs AC 13 → HIT!
```

#### Step 3: Damage with Sneak Attack
```go
damageEvent := NewGameEvent(OnDamageRoll).
    WithActor(rogue).
    WithTarget(orc).
    WithContext("weapon", weapon).
    WithContext("sneak_attack_available", true)

eventBus.Emit(damageEvent)
```

#### Step 4: All Damage Modifiers
```go
// Base weapon damage
baseDamage := rollDice("1d8") // = 6

// MagicWeaponModifier
event.SetContext("magic_damage_bonus", 1)

// SneakAttackModifier  
func (s *SneakAttackModifier) HandleEvent(event) {
    if event.GetBoolContext("sneak_attack_available") {
        sneakDamage := rollDice("2d6") // = 7 (3+4)
        event.SetContext("sneak_attack_damage", 7)
        s.markUsed() // Can't sneak attack again this turn
    }
}

// Final: 6 (base) + 3 (DEX) + 1 (magic) + 7 (sneak) = 17 piercing
```

### Result Log
```
Slyvia attacks Orc with +1 Rapier (advantage): 24 vs AC 13 - HIT!
Damage: 1d8+3 = 6+3=9 +1 magic +7 sneak attack = 17 piercing damage 🗡️
Orc takes 17 damage: 15 HP → -2 HP (dead)
```

## Example 3: Complex Spell Combat

**Scenario**: Blessed Fighter with Action Surge casts Haste then attacks twice

### Setup
```go
modifiers := []Modifier{
    &BlessModifier{targetID: "fighter-999"},
    &ActionSurgeModifier{characterID: "fighter-999"},
    &HasteModifier{targetID: "fighter-999"},
    &SpellSlotModifier{characterID: "fighter-999"},
}
```

### Event Flow

#### Step 1: Cast Haste Spell
```go
castEvent := NewGameEvent(OnCastSpell).
    WithActor(fighter).
    WithContext("spell", "haste").
    WithContext("spell_level", 3).
    WithContext("slot_level", 3).
    WithContext("target", fighter)

eventBus.Emit(castEvent)
```

```go
// SpellSlotModifier consumes 3rd level slot
// HasteModifier activates on fighter
type HasteModifier struct {
    targetID string
    tracker  *DurationTracker // 1 minute
}

func (h *HasteModifier) HandleEvent(event) {
    if event.Actor.ID != h.targetID {
        return
    }
    
    switch event.Type {
    case OnAttackRoll:
        event.SetContext("haste_attack_bonus", 2) // +2 AC, +2 to DEX saves
    case OnMove:
        speed := event.GetIntContext("speed")
        event.SetContext("speed", speed*2) // Double speed
    case OnTurnStart:
        event.SetContext("extra_action", true) // One additional action
    }
}
```

#### Step 2: Use Action Surge
```go
abilityEvent := NewGameEvent(OnUseAbility).
    WithActor(fighter).
    WithContext("ability_key", "action_surge")

// ActionSurgeModifier grants additional action this turn
event.SetContext("extra_actions", 2) // Normal + Action Surge + Haste
```

#### Step 3: First Attack
```go
attackEvent1 := NewGameEvent(OnAttackRoll).
    WithActor(fighter).
    WithTarget(enemy).
    WithContext("action_type", "action")

// BlessModifier: +1d4 = 3
// HasteModifier: +2 
// Attack: 1d20 + 5 (base) + 3 (bless) + 2 (haste) = 1d20 + 10
```

#### Step 4: Second Attack (Action Surge)
```go
attackEvent2 := NewGameEvent(OnAttackRoll).
    WithActor(fighter).
    WithTarget(enemy).
    WithContext("action_type", "extra_action")

// Same bonuses apply
// Attack: 1d20 + 10
```

#### Step 5: Third Attack (Haste Action)
```go
attackEvent3 := NewGameEvent(OnAttackRoll).
    WithActor(fighter).
    WithTarget(enemy).
    WithContext("action_type", "haste_action")

// Haste action can only make one weapon attack
// Attack: 1d20 + 10
```

### Result Log
```
Gareth casts Haste on himself (3rd level slot consumed)
Gareth uses Action Surge (1 use remaining)
Gareth attacks with Longsword (+10): 18 - HIT! 9 damage
Gareth attacks with Longsword (+10): 15 - HIT! 7 damage  
Gareth attacks with Longsword (Haste, +10): 23 - HIT! 11 damage
Total: 27 damage dealt this turn!
```

## Example 4: Environmental and Movement Effects

**Scenario**: Character with bleeding wound moves through difficult terrain while poisoned

### Setup
```go
modifiers := []Modifier{
    &BleedModifier{targetID: "char-123", damagePerMove: 2, speedReduction: 10},
    &PoisonedModifier{targetID: "char-123"}, // Disadvantage on attacks/abilities
    &DifficultTerrainModifier{area: swampArea},
}
```

### Event Flow

#### Step 1: Attempt to Move
```go
moveEvent := NewGameEvent(OnMoveStart).
    WithActor(character).
    WithContext("current_position", pos1).
    WithContext("target_position", pos2).
    WithContext("distance", 30).
    WithContext("speed", 30) // Normal speed

eventBus.Emit(moveEvent)
```

#### Step 2: Modifiers Affect Movement
```go
// BleedModifier reduces speed
func (b *BleedModifier) HandleEvent(event) {
    if event.Type == OnMoveStart {
        speed := event.GetIntContext("speed")
        newSpeed := max(0, speed - 10) // Reduce by 10
        event.SetContext("speed", newSpeed)
        event.SetContext("bleed_speed_reduction", true)
    }
}

// DifficultTerrainModifier doubles cost  
func (d *DifficultTerrainModifier) HandleEvent(event) {
    if d.area.Contains(event.GetContext("target_position")) {
        cost := event.GetIntContext("movement_cost") // 1 foot per foot
        event.SetContext("movement_cost", cost * 2)   // 2 feet per foot
        event.SetContext("difficult_terrain", true)
    }
}

// Final: 20 speed, 30 feet to move, costs 60 feet of movement
// Result: Can only move 20 feet this turn
```

#### Step 3: Execute Partial Movement
```go
actualMoveEvent := NewGameEvent(OnMove).
    WithActor(character).
    WithContext("distance", 20). // Only moved 20 feet
    WithContext("stopped_by", "insufficient_movement")

eventBus.Emit(actualMoveEvent)
```

#### Step 4: Bleed Damage from Movement
```go
// BleedModifier triggers damage
func (b *BleedModifier) HandleEvent(event) {
    if event.Type == OnMove {
        distance := event.GetIntContext("distance")
        if distance > 0 {
            damageEvent := NewGameEvent(OnTakeDamage).
                WithActor(event.Actor).
                WithContext("damage", 2).
                WithContext("damage_type", "necrotic").
                WithContext("source", "bleeding")
            
            EmitEvent(damageEvent)
        }
    }
}

// Character takes 2 necrotic damage from moving while bleeding
```

### Result Log
```
Wounded Explorer moves through swamp (bleeding, difficult terrain)
Movement: 30 feet attempted, 20 feet completed (reduced speed: 30→20, difficult terrain)
Bleeding damage: 2 necrotic damage from movement
HP: 15 → 13
```

## Example 5: Multi-Character Combat Round

**Scenario**: Full combat round with multiple characters and overlapping effects

### Setup
- **Barbarian**: Raging, attacking with greataxe
- **Cleric**: Blessed, casting Healing Word on barbarian  
- **Rogue**: Hidden, attempting sneak attack
- **Wizard**: Concentrating on Web spell, enemies have disadvantage

### Turn 1: Barbarian's Turn
```go
// Barbarian attacks orc
attackEvent := NewGameEvent(OnAttackRoll).
    WithActor(barbarian).
    WithTarget(orc).
    WithContext("weapon", greataxe)

// Modifiers: Rage (no effect on attack), Proficiency (+3), Bless (+1d4)
// Result: 1d20 + 4 (STR) + 3 (prof) + 2 (bless) = 19 - HIT!

damageEvent := NewGameEvent(OnDamageRoll).
    WithActor(barbarian).
    WithTarget(orc).
    WithContext("weapon", greataxe)

// Modifiers: Great Weapon Fighting (reroll), Rage (+2)
// Result: 1d12+4+2 = 13 slashing damage
```

### Turn 2: Orc's Turn (Attacks Barbarian)
```go
attackEvent := NewGameEvent(OnAttackRoll).
    WithActor(orc).
    WithTarget(barbarian).
    WithContext("weapon", scimitar).
    WithContext("has_disadvantage", true) // From Web spell

// Web spell gives disadvantage: 2d20kl1 + 5 = 8 - MISS!
```

### Turn 3: Cleric's Turn
```go
// Cast Healing Word as bonus action
spellEvent := NewGameEvent(OnCastSpell).
    WithActor(cleric).
    WithContext("spell", "healing_word").
    WithContext("target", barbarian).
    WithContext("slot_level", 1)

// Healing: 1d4 + 4 (WIS) = 7 HP restored to barbarian
```

### Turn 4: Rogue's Turn
```go
// Rogue attacks from hiding
attackEvent := NewGameEvent(OnAttackRoll).
    WithActor(rogue).
    WithTarget(orc).
    WithContext("weapon", shortsword).
    WithContext("has_advantage", true). // Hidden
    WithContext("ally_adjacent", true)  // Barbarian next to orc

// Sneak attack conditions met: advantage AND ally adjacent
// Attack: 2d20kh1 + 5 (DEX) + 2 (prof) = 19 - HIT!
// Damage: 1d6+3 + 2d6 (sneak) = 11 piercing damage
```

### Round Summary
```
=== ROUND 1 SUMMARY ===
Grunk (Barbarian): Attacks Orc for 13 damage (rage active)
Orc: Attacks Grunk with disadvantage - MISS (Web spell)  
Sister Mary (Cleric): Heals Grunk for 7 HP with Healing Word
Slyvia (Rogue): Sneak attacks Orc for 11 damage (2d6 sneak attack)

Orc HP: 15 → 2 → -9 (DEAD)
Grunk HP: 8 → 15 (healed)
```

This shows how multiple modifiers interact seamlessly - each one only caring about its specific triggers while complex emergent behavior arises naturally!