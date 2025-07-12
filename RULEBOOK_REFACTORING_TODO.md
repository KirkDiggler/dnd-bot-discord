# Rulebook Domain Refactoring TODO

This document identifies all the coupling points between character and rulebook domains that need to be refactored.

## 1. Direct Imports from Rulebook in Character Domain

### Files with direct rulebook imports:
- `/internal/domain/character/character_proficiencies.go` - imports `dnd5e` and uses `rulebook.Proficiency`
- `/internal/domain/character/spells.go` - imports `dnd5e` 
- `/internal/domain/character/character_resource.go` - imports `dnd5e`
- `/internal/domain/character/character_combat.go` - imports `dnd5e` and uses `rulebook.ProficiencyTypeWeapon`
- `/internal/domain/character/character_test.go` - imports `dnd5e` (test file)
- `/internal/domain/character/character_draft_test.go` - imports `dnd5e` (test file)
- `/internal/domain/character/character_draft_flow_test.go` - imports `dnd5e` (test file)
- `/internal/domain/character/action_economy_test.go` - imports `dnd5e` (test file)
- `/internal/domain/character/character_proficiency_fix_test.go` - imports `dnd5e` (test file)

## 2. D&D 5e Specific Business Logic in Character Domain

### Proficiency Bonus Calculation
**Location**: `/internal/domain/character/character_proficiencies.go`
```go
func (c *Character) GetProficiencyBonus() int {
    // D&D 5e specific formula: 2 + ((level - 1) / 4)
    return 2 + ((c.Level - 1) / 4)
}
```
**Action**: Move to `/internal/domain/rulebook/dnd5e/calculators/proficiency_calculator.go`

### Combat Calculations
**Location**: `/internal/domain/character/character_combat.go`
- Proficiency bonus calculation (lines 95, 214): `proficiencyBonus = 2 + ((c.Level - 1) / 4)`
- Weapon proficiency category mapping (lines 338-342)
- Fighting style specific logic
- Finesse weapon rules
- Monk weapon rules
- Martial Arts damage dice progression (lines 398-408)

**Action**: Create `/internal/domain/rulebook/dnd5e/calculators/combat_calculator.go`

### Sneak Attack Rules
**Location**: `/internal/domain/character/character_sneak_attack.go`
- Rogue-specific class check
- Sneak attack dice calculation: `(c.Level + 1) / 2`
- Weapon eligibility rules (finesse/ranged)

**Action**: Move to `/internal/domain/rulebook/dnd5e/abilities/sneak_attack.go` (already exists, needs integration)

### Spell Slot Calculation
**Location**: `/internal/domain/character/character_resource.go`
- References to `rulebook.GetSpellSlotTable()`
- Spell slot progression logic

**Action**: Already exists in rulebook, remove direct references

## 3. Business Logic in Services That Should Move

### AC Calculator
**Location**: `/internal/services/character/default_ac_calculator.go`
- Currently wraps `features2.CalculateAC(char)`
- Should use the rulebook's AC calculator interface

**Action**: Already exists in `/internal/domain/rulebook/dnd5e/calculators/ac_calculator.go`, update service to use it

### Ability Score Validation
**Location**: `/internal/services/character/validation.go`
- D&D 5e specific ability score range (3-18)
- Required abilities list (STR, DEX, CON, INT, WIS, CHA)

**Action**: Create `/internal/domain/rulebook/dnd5e/validators/ability_score_validator.go`

## 4. Hardcoded D&D 5e Constants

### Weapon Categories
**Location**: `/internal/domain/character/character_combat.go`
```go
categoryMap := map[string]string{
    "simple":  "simple-weapons",
    "martial": "martial-weapons",
}
```
**Action**: Move to rulebook constants

### Ability Names
**Location**: Multiple files
- Hardcoded references to "STR", "DEX", "CON", "INT", "WIS", "CHA"
- Uses `shared.AttributeStrength`, `shared.AttributeDexterity`, etc.

**Action**: Create rulebook-specific attribute definitions

## 5. Existing Rulebook Infrastructure

### Already exists in `/internal/domain/rulebook/dnd5e/`:
- `calculators/ac_calculator.go` - AC calculation
- `abilities/` - Various ability implementations (rage, sneak attack, etc.)
- `features/` - Feature implementations
- `proficiency.go` - Proficiency types and definitions

### Needs to be created:
- `calculators/proficiency_calculator.go` - Proficiency bonus calculation
- `calculators/combat_calculator.go` - Attack/damage calculations
- `validators/ability_score_validator.go` - Ability score validation
- `validators/character_validator.go` - Character creation validation
- `constants.go` - D&D 5e specific constants

## 6. Interface Design

### Proposed Calculator Interfaces:
```go
// In rulebook package
type ProficiencyCalculator interface {
    CalculateBonus(level int) int
}

type CombatCalculator interface {
    CalculateAttackBonus(character *Character, weapon *Weapon) int
    CalculateDamageBonus(character *Character, weapon *Weapon) int
    GetWeaponAbilityModifier(character *Character, weapon *Weapon) int
}

type CharacterValidator interface {
    ValidateAbilityScores(scores map[string]int) error
    ValidateCharacterCreation(input ValidationInput) error
}
```

## 7. Migration Strategy

1. **Phase 1**: Create new calculator/validator implementations in rulebook
2. **Phase 2**: Update character domain to use interfaces instead of direct logic
3. **Phase 3**: Update services to inject rulebook-specific implementations
4. **Phase 4**: Remove all direct rulebook imports from character domain
5. **Phase 5**: Update tests to use mocked interfaces

## 8. Test Considerations

- Test files in character domain should not import rulebook packages
- Create mock implementations of calculator/validator interfaces
- Move D&D 5e specific test cases to rulebook package tests