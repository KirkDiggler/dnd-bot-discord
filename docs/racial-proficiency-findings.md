# Racial Proficiency Testing Results

## Summary
Racial proficiencies are correctly loaded from the D&D 5e API and applied during character finalization. However, duplicate proficiency prevention is not yet implemented at the choice selection level.

## Key Findings

### Racial Proficiencies Working
1. **Half-Orc**: Gets Intimidation skill automatically
2. **Elf**: Gets Perception skill automatically  
3. **Dwarf**: Gets battleaxe, handaxe, light hammer, warhammer proficiencies + tool choice
4. **Half-Elf**: Gets to choose 2 skills from any list
5. **Human**: No proficiencies (but should get +1 to all abilities)

### Duplicate Proficiency Handling
- **Current Behavior**: 
  - Duplicates appear in choice lists (e.g., Half-Orc Barbarian can still choose Intimidation)
  - Finalization correctly deduplicates (only one instance in final character)
- **Desired Behavior**: 
  - Duplicates should be filtered from choice lists
  - Prevents confusion and wasted choices

### Racial Ability Score Improvements
- **Issue Found**: Racial ability bonuses are showing as the scores themselves, not added to base
- **Known Issue**: This matches issue #155 (Human racial ability bonuses not being applied)
- **Expected**: Base score + racial bonus (e.g., 10 + 2 = 12)
- **Actual**: Just the bonus value (e.g., 2)

## Implementation Status

### What's Working ✅
1. Racial proficiencies load from API
2. Proficiencies are applied during finalization
3. Deduplication happens at finalization (no duplicates in final character)
4. Racial choice options (Half-Elf skills, Dwarf tools) are presented

### What Needs Implementation ❌
1. Filter duplicate proficiencies from choice lists
2. Fix racial ability score bonuses (issue #155)
3. Validate proficiency selections to prevent invalid choices

## Test Results
```
✅ TestRacialProficienciesFromAPI - All races load correct proficiencies
✅ TestCharacterFinalizationDeduplication - Duplicates removed in final character
✅ TestRaceClassProficiencyOverlap - Shows duplicates still in choices (expected)
❌ TestRacialAbilityScoreImprovements - Ability scores not calculated correctly
```

## Recommendations
1. Implement duplicate filtering in `ResolveProficiencyChoices`
2. Add validation in `ValidateProficiencySelections` 
3. Fix ability score calculation (separate issue #155)
4. Consider showing already-granted proficiencies in UI but disabled/grayed out

## Code Location
- Choice resolution: `internal/services/character/choice_resolver.go`
- Proficiency application: `internal/services/character/service.go`
- Character finalization: `FinalizeDraftCharacter` method