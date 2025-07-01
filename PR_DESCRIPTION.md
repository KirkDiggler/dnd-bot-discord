# Fix: Circular Import Dependencies in Domain/Game Architecture

Fixes #230

## Summary

This PR resolves circular import dependencies that emerged after reorganizing the codebase into a domain/game architecture. The refactoring follows clean architecture principles where the `shared` package has zero dependencies on other domain packages.

## Problem

After moving files to the new domain/game structure, we encountered multiple circular import cycles:
- `rulebook` → `character` → `rulebook` (via CharacterResources)
- `character` → `equipment` → `character` (via weapon.Attack())
- `shared` → `rulebook` → `shared` (via resources importing rulebook.Class)

## Solution

### 1. **Shared Package Refactoring**
- Moved `CharacterResources` from `shared` to `character` package (it's character-specific state)
- Created `shared/ability_bonus.go` for the `AbilityBonus` type (used by both character and rulebook)
- Created `shared/slot.go` for equipment slot constants
- Ensured `shared` package has zero imports from other domain packages

### 2. **Type Relocations**
- `AbilityBonus`: `character` → `shared` (breaks rulebook→character dependency)
- `Slot` and slot constants: `character` → `shared` (used by equipment system)
- `CharacterResources`: `shared` → `character` (belongs with character state)

### 3. **Equipment System Simplification**
- Removed `weapon.Attack()` method that created circular dependency
- Weapons now only provide data (GetDamage(), IsRanged(), etc.)
- Attack logic remains in the character package where it belongs

### 4. **Import Updates**
All imports have been updated across the codebase to use the new package locations. The IDE's refactoring tools were used to ensure all references were properly updated.

## Testing

- ✅ All unit tests passing
- ✅ Bot successfully starts and runs
- ✅ Dungeon system tested end-to-end (user confirmed: "oh my god, we are on the other side now. bot can run and i did a dungeon")
- ✅ No more circular import errors during build

## Architecture Benefits

This refactoring positions us well for the planned event-driven architecture (#202):
- Clear domain boundaries with no circular dependencies
- Shared types truly shared without coupling
- Each domain package can evolve independently
- Ready for event-driven modifier system implementation

## Files Changed

### New Files
- `/internal/domain/shared/ability_bonus.go` - Moved from character package
- `/internal/domain/shared/slot.go` - Equipment slot constants

### Modified Files
- `/internal/domain/character/character_resources.go` - Now contains CharacterResources
- `/internal/domain/shared/resources.go` - Removed CharacterResources, now only simple types
- `/internal/domain/rulebook/race.go` - Updated to use shared.AbilityBonus
- `/internal/domain/equipment/weapon.go` - Removed Attack() method
- Plus all files with import updates

## Lessons Learned

As documented in CLAUDE.md:
> "oh boy you are still at this... you are not good at moving contents of files and changing their imports. When you do large refactors like this, you should just ask me (the user) to use their IDE"

This PR was completed using IDE refactoring tools for reliable import updates across the entire codebase.

## Next Steps

With circular dependencies resolved, we can now proceed with:
1. Implementing the event-driven modifier system
2. Adding new character abilities without coupling
3. Expanding the combat system with clean domain boundaries

---

🎉 This was a challenging refactor, but the codebase is now much cleaner and ready for future enhancements!