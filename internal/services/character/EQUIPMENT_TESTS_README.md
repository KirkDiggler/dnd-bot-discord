# Equipment Choice System Tests

This directory contains comprehensive unit tests for the D&D bot's equipment choice system. The tests are designed to catch issues with equipment selection before runtime, particularly focusing on nested choices and complex equipment bundles.

## Test Files

### 1. `equipment_choice_resolver_test.go`
Main test suite for the equipment choice resolver, covering:
- Fighter equipment choices (armor, weapons, packs)
- Monk equipment choices (simple weapons)
- Wizard equipment choices (arcane focus selection)
- Rogue equipment choices (weapon bundles)
- Edge cases (empty choices, nil options, invalid references)

### 2. `equipment_handlers_test.go`
Tests for Discord interaction handlers:
- Equipment choices presentation handler
- Equipment selection handler
- Nested equipment selection handler
- Mock Discord API interactions

### 3. `equipment_integration_test.go`
Integration tests for the full equipment selection flow:
- Complete Fighter equipment selection flow
- Wizard focus selection flow
- Equipment bundle handling
- Nested choice resolution

### 4. `equipment_all_classes_test.go`
Systematic tests for all D&D classes:
- Barbarian (martial weapons, simple weapons)
- Bard (instruments, packs)
- Cleric (holy symbols, armor)
- Druid (shields, simple weapons)
- Fighter (comprehensive choices)
- Monk (simple weapons)
- Paladin (martial weapons, shields)
- Ranger (weapon bundles)
- Rogue (finesse weapons)
- Sorcerer (arcane focus)
- Warlock (pact weapons)
- Wizard (arcane focus)

### 5. `equipment_edge_cases_test.go`
Edge cases and error scenarios:
- Nil and empty inputs
- Invalid references
- Deep nesting
- Special characters
- Count edge cases (zero, negative)
- Large data sets
- Type mismatches

## Key Test Patterns

### 1. Nested Choice Detection
Tests ensure that nested choices (like "choose 2 martial weapons") are properly marked with a "nested-" prefix in their key, which triggers the weapon selection UI.

```go
// Example: Testing nested martial weapon choice
s.Contains(choices[0].Options[1].Key, "nested")
s.Contains(choices[0].Options[1].Description, "Choose 2")
```

### 2. Equipment Bundle Formatting
Tests verify that equipment bundles are properly formatted with counts:

```go
// Example: "20x Arrow" instead of just "Arrow"
s.Contains(choices[0].Options[0].Name, "20x Arrow")
```

### 3. Weapon Descriptions
Tests check that weapon descriptions include damage information:

```go
// Example: Longsword should show damage
s.Contains(option.Description, "1d8 slashing")
```

## Running the Tests

### Run all equipment tests:
```bash
go test ./internal/services/character -run "Equipment"
```

### Run specific test suites:
```bash
# Choice resolver tests
go test ./internal/services/character -run "TestEquipmentChoiceResolverSuite"

# All classes tests
go test ./internal/services/character -run "TestAllClassesEquipmentSuite"

# Edge cases
go test ./internal/services/character -run "TestEquipmentEdgeCasesSuite"
```

### Run with coverage:
```bash
go test ./internal/services/character -cover -run "Equipment"
```

## Common Issues These Tests Catch

1. **Missing Nested Choice Markers**: When a choice like "any martial weapon" isn't properly marked as nested
2. **Incorrect Bundle Formatting**: When item counts aren't displayed properly
3. **Missing Weapon Descriptions**: When weapon stats aren't included in descriptions
4. **Nil Reference Handling**: When equipment references are nil or invalid
5. **Empty Choice Filtering**: When choices with no valid options aren't filtered out

## Adding New Tests

When adding support for new classes or equipment types:

1. Add test cases to `equipment_all_classes_test.go` for the new class
2. Include edge cases in `equipment_edge_cases_test.go`
3. Update integration tests if the flow changes
4. Ensure nested choices are properly tested

## Mock Data Helpers

The test files include helper functions for creating test data:
- `createMartialWeaponOptions()`: All martial weapons
- `createSimpleWeaponOptions()`: All simple weapons
- `createArcaneFocusOptions()`: Arcane focus items
- `createMusicalInstrumentOptions()`: Bard instruments

These helpers ensure consistent test data across all test files.