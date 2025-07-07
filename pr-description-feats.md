# Implement D&D 5e Feats System

Part of #74

## Summary
This PR implements a comprehensive D&D 5e feats system and restores the rage ability handler using the rpg-toolkit event bus. The feat system is designed to be modular and event-driven, allowing feats and class abilities to modify character mechanics without tight coupling.

## What's Implemented

### Core Architecture
- **Feat Interface**: Defines the contract for all feats including prerequisites, application, and event handler registration
- **Feat Registry**: Global registry for managing all available feats and checking prerequisites
- **Event Integration**: Feats register handlers with the rpg-toolkit event bus when characters join encounters

### Implemented Feats (6 total)
1. **Alert** (+5 initiative, can't be surprised)
2. **Great Weapon Master** (-5 attack/+10 damage with heavy weapons, bonus action on crit/kill)
3. **Lucky** (3 luck points for rerolls per long rest)
4. **Sharpshooter** (-5 attack/+10 damage with ranged weapons, ignore cover/long range)
5. **Tough** (+2 HP per level)
6. **War Caster** (advantage on concentration saves, cast with hands full, spell opportunity attacks)

### Restored Class Abilities
1. **Rage** (Barbarian) - Fully implemented with event handlers for:
   - +2/+3/+4 damage bonus on melee attacks (level-based)
   - Resistance to bludgeoning, piercing, and slashing damage
   - Proper activation/deactivation with resource tracking
   - Event-driven damage and resistance calculations

## Technical Details
- All feats use the event-driven architecture with rpg-toolkit
- Feats are stored as character features with type "feat"
- Event handlers are registered when characters join encounters
- Feat-specific data (luck points, HP bonus) stored in feature metadata

## Testing Checklist
- [ ] Feats compile without errors
- [ ] Registry properly registers all feats on init
- [ ] Event handlers are registered when character joins encounter
- [ ] Feats can be applied to characters (via code - no UI yet)
- [ ] Rage ability activates and applies damage bonus
- [ ] Rage resistance properly reduces physical damage
- [ ] Rage event handlers register correctly

## Next Steps (Future PRs)
1. Add feat selection to character creation (Variant Human gets feat at level 1)
2. Add feat selection to character advancement (levels 4, 8, 12, etc.)
3. Implement remaining PHB feats (Sentinel, Polearm Master, etc.)
4. Add UI toggles for optional feat abilities (GWM/Sharpshooter power attack)

## Notes
- Great Weapon Master is the feat (not Great Weapon Fighting which is a fighting style)
- All implemented feats are accurate to D&D 5e Player's Handbook rules
- The system is ready for UI integration but currently only works via code