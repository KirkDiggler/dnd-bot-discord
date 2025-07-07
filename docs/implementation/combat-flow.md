# Combat Flow Documentation

## Overview

This document details the complete flow of combat calculations in the D&D bot, from attack initiation through damage application. The system implements D&D 5e rules with an event-driven architecture for extensibility.

## Attack Resolution Flow

### Phase 1: Attack Initiation

```
User clicks "Attack" → Select Target → PerformAttack()
```

**Validations**:
- Is it the attacker's turn?
- Is the attacker conscious?
- Is the target valid and alive?
- Does attacker have actions remaining?

### Phase 2: Attack Roll

```
1. BeforeAttackRoll Event
   ├─ Can be cancelled (e.g., Sanctuary)
   ├─ Can add advantage/disadvantage
   └─ Can modify attack bonus

2. Roll Attack
   ├─ Roll 1d20 (or 2d20 for advantage/disadvantage)
   ├─ Add ability modifier (STR/DEX)
   ├─ Add proficiency bonus (if proficient)
   ├─ Add fighting style bonus (e.g., Archery +2)
   └─ Apply event modifications

3. OnAttackRoll Event
   └─ Log the roll, allow inspection

4. Determine Result
   ├─ Natural 20 = Critical Hit
   ├─ Natural 1 = Critical Miss
   └─ Total ≥ Target AC = Hit

5. AfterAttackRoll Event
   └─ Final chance to modify hit/miss/crit
```

### Phase 3: Damage Calculation (if hit)

```
1. Base Damage
   ├─ Roll weapon dice (e.g., 1d8 for longsword)
   ├─ Critical hits: roll dice twice
   └─ Great Weapon Fighting: reroll 1s and 2s once

2. Add Modifiers
   ├─ Ability modifier (STR/DEX)
   ├─ Fighting style (e.g., Dueling +2)
   └─ Enhancement bonuses (magic weapons)

3. OnDamageRoll Event
   ├─ Rage damage bonus (+2/+3/+4)
   ├─ Feat bonuses (GWM +10)
   └─ Other effect bonuses

4. Special Damage
   ├─ Sneak Attack (if eligible)
   ├─ Divine Smite
   └─ Other triggered abilities
```

### Phase 4: Damage Application

```
1. Target Defenses
   ├─ Check immunities (damage → 0)
   ├─ Check resistances (damage → half)
   └─ Check vulnerabilities (damage → double)

2. BeforeTakeDamage Event
   ├─ Final damage modifications
   ├─ Defensive abilities trigger
   └─ Damage can still be prevented

3. Apply Damage
   ├─ Reduce target HP
   ├─ Check for unconsciousness (0 HP)
   └─ Update combat state

4. OnTakeDamage Event
   └─ Post-damage effects trigger
```

## Damage Calculation Details

### Attack Bonus Calculation

```go
Attack Bonus = Base + Proficiency + Ability Modifier + Bonuses

Where:
- Base = 0
- Proficiency = 2 + ((level-1) / 4) [if proficient]
- Ability Modifier = STR or DEX modifier
- Bonuses = Fighting style + magic + effects
```

### Damage Roll Calculation

```go
Total Damage = Weapon Dice + Ability Modifier + Bonuses

Where:
- Weapon Dice = base dice (doubled on crit)
- Ability Modifier = STR or DEX
- Bonuses = Fighting style + Rage + Feats + Effects
```

### Critical Hit Rules

- **Doubled**: All damage dice (weapon, sneak attack, smite)
- **Not Doubled**: Ability modifiers, flat bonuses
- **Example**: Longsword crit = 2d8 + STR (not 2×STR)

## Special Cases

### Finesse Weapons
- Can use DEX instead of STR for attack and damage
- Player chooses higher modifier
- Monk weapons always use DEX if higher

### Two-Weapon Fighting
- Main hand uses action
- Off-hand uses bonus action
- Off-hand doesn't add ability modifier to damage (unless TWF style)

### Sneak Attack
- Once per turn (not per round)
- Requires finesse or ranged weapon
- Needs advantage OR ally adjacent
- Damage: (level + 1) / 2 d6

### Great Weapon Fighting
```go
for each damage die:
    if roll == 1 or 2:
        reroll once
        use new value
```

## Resistance and Vulnerability

### Application Order
1. Calculate total damage
2. Apply immunities (damage = 0)
3. Apply resistances (damage / 2, round down)
4. Apply vulnerabilities (damage × 2)

### Stacking Rules
- Multiple resistances don't stack
- Resistance + Vulnerability = normal damage
- Immunities override everything

## Event System Integration

### Event Priorities
```
10:  System critical (hit determination)
30:  Defensive modifications (resistance)
50:  Offensive modifications (damage bonus)
70:  UI updates
100: Logging
```

### Key Events

**BeforeAttackRoll**: Modify attack before rolling
- Great Weapon Master -5 penalty
- Advantage/disadvantage application
- Attack cancellation (Sanctuary)

**OnDamageRoll**: Modify damage after rolling
- Rage damage bonus
- Feat damage bonuses
- Sneak attack addition

**BeforeTakeDamage**: Final damage modifications
- Resistance application
- Damage reduction abilities
- Defensive reactions

## Action Economy

### Action Types
- **Action**: Standard attacks, most abilities
- **Bonus Action**: Off-hand attacks, some abilities
- **Reaction**: Opportunity attacks, defensive abilities
- **Free**: Drawing weapons, speaking

### Tracking
```go
type ActionEconomy struct {
    ActionsUsed       int
    BonusActionUsed   bool
    ReactionUsed      bool
    MovementUsed      int
}
```

## Combat Log Integration

Each phase generates combat log entries:
```
→ Grunk attacks Goblin with Greataxe
  Attack: 1d20:14 + 5 = 19 vs AC 15 [HIT]
  Damage: 1d12:7 + 3 + 2 (rage) = 12 slashing
  Goblin takes 12 damage (HP: 7 → 0)
  Goblin is defeated!
```

## Performance Considerations

1. **Event Bus**: Synchronous processing ensures order
2. **Effect Caching**: Active effects cached on character
3. **Dice Rolling**: Uses crypto/rand for fairness
4. **Database Updates**: Batched after combat round

## Future Enhancements

1. **Saving Throws**: Spell DCs and save modifiers
2. **Conditions**: Comprehensive condition system
3. **Reactions**: Opportunity attacks, Shield spell
4. **Area Effects**: Fireball, breath weapons
5. **Cover System**: Half/three-quarters/full cover

This combat flow provides a robust foundation for D&D 5e combat while maintaining extensibility through the event system.