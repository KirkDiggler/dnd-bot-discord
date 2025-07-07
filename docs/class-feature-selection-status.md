# Class Feature Selection Status

## Overview
Tracking which classes need feature selections during character creation at level 1.

## Classes That Need Selections at Level 1

### ✅ Implemented
1. **Fighter** - Fighting Style ✅
   - Archery, Defense, Dueling, Great Weapon Fighting, Protection, Two-Weapon Fighting
   
2. **Ranger** - Two selections ✅
   - Favored Enemy (aberrations, beasts, celestials, etc.)
   - Natural Explorer (arctic, coast, desert, etc.)

### ❌ Not Implemented
1. **Cleric** - Divine Domain ❌
   - PHB Domains: Knowledge, Life, Light, Nature, Tempest, Trickery, War
   - Currently shows as TODO in code
   - This is why users see "class features" but get no selection

## Classes With No Level 1 Selections
These classes should work fine as-is:
1. **Barbarian** - No selections (Path at level 3)
2. **Bard** - No selections (College at level 3)  
3. **Monk** - No selections (Tradition at level 3)
4. **Paladin** - No selections (Oath at level 3)
5. **Rogue** - No selections (Archetype at level 3)

## Partially Implemented Classes (Not in Phase 1 yet)
1. **Druid** - Circle selection at level 2
2. **Sorcerer** - Sorcerous Origin at level 1 (needs implementation when class is added)
3. **Warlock** - Otherworldly Patron at level 1 (needs implementation when class is added)
4. **Wizard** - Arcane Tradition at level 2

## Implementation Status

### Current Issues
1. **Cleric Divine Domain** - Not implemented, causing confusion during character creation
2. Handler only checks for `favored_enemy` and `fighting_style` in main handler
3. No `ShowDivineDomainSelection` method exists

### Code Locations
- Feature selection handler: `/internal/handlers/discord/dnd/character/class_features.go`
- Main handler checks: `/internal/handlers/discord/handler.go`
- Feature definitions: `/internal/domain/rulebook/dnd5e/features/features.go`

## Next Steps
1. Implement Divine Domain selection for Cleric
2. Test all implemented classes (Fighter, Ranger, Barbarian, Bard, Monk, Paladin, Rogue, Cleric)
3. Consider if domain spells should be automatically added