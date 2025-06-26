# Combat End Detection Test Coverage

## Overview
This document describes the test coverage for the combat end detection feature implemented in PR #110 (Issue #109).

## Test Files Created

### 1. `/internal/entities/encounter_combat_end_test.go`
Unit tests for the core `CheckCombatEnd()` method and related functionality:
- `TestCheckCombatEnd_PlayersWin` - All monsters defeated
- `TestCheckCombatEnd_PlayersLose` - All players defeated  
- `TestCheckCombatEnd_CombatContinues` - Both sides have active combatants
- `TestCheckCombatEnd_MonsterOnlyBattle` - Edge case with no players
- `TestCheckCombatEnd_AllDefeated` - Edge case with everyone defeated
- `TestApplyDamage_Defeat` - Combatant defeat mechanics
- `TestApplyDamage_TempHP` - Temporary HP absorption

### 2. `/internal/services/encounter/combat_end_integration_test.go`
Integration tests demonstrating combat end detection through the service layer:
- `TestCombatEndIntegration_MonstersDefeatPlayer` - Monster attack defeats last player
- `TestCombatEndIntegration_PlayerDefeatsLastMonster` - Demonstrates monster-on-monster combat

## Testing Limitations

### Character Attack Dice Mocking
The main limitation in testing is that `Character.Attack()` uses the global `dice.Roll()` function rather than an injected dice roller. This means:
- We cannot mock player attack rolls in tests
- Player victory scenarios are difficult to test reliably
- Tests focusing on player attacks may fail randomly

### Workaround
Tests focus on monster attacks (which do use the injected dice roller) to demonstrate the combat end detection logic works correctly.

## Test Coverage Summary

✅ **Core Logic** - `CheckCombatEnd()` method fully tested
✅ **Monster Victories** - Monster defeats last player triggers defeat
✅ **Combat Continuation** - Combat continues when enemies remain
✅ **Edge Cases** - Monster-only battles, all defeated scenarios
✅ **Integration** - Service layer properly detects and handles combat end
⚠️  **Player Victories** - Limited by dice mocking architecture

## Future Improvements

To achieve full test coverage, consider:
1. Refactoring `Character.Attack()` to use injected dice roller
2. Adding an interface for attack calculations
3. Creating end-to-end tests with Discord bot integration

## Running the Tests

```bash
# Run all combat end tests
go test ./internal/entities -run TestCheckCombatEnd -v
go test ./internal/entities -run TestApplyDamage -v
go test ./internal/services/encounter -run TestCombatEndIntegration -v

# Run all encounter tests
go test ./internal/services/encounter -v
```