# Test Migration Example: From Raw Tests to Suite Pattern

This document shows a real example of migrating from raw tests to the suite pattern, demonstrating the benefits.

## Before: discord_flow_test.go (Raw Test)

```go
package character_test

import (
    "context"
    "testing"
    // ... 7 import statements
)

func TestDiscordCharacterCreationFlow(t *testing.T) {
    // Every test needs this boilerplate
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockClient := mockdnd5e.NewMockClient(ctrl)
    mockRepo := mockcharrepo.NewMockRepository(ctrl)

    svc := character.NewService(&character.ServiceConfig{
        DNDClient:  mockClient,
        Repository: mockRepo,
        // OOPS! Forgot DraftRepository - test fails!
    })

    ctx := context.Background()
    userID := "discord_user"
    realmID := "discord_realm"
    characterID := "char_123"

    // ... test logic ...
}
```

### Problems with Raw Tests:
1. **Boilerplate in every test** - 10+ lines of setup code
2. **Easy to forget dependencies** - Missing DraftRepository
3. **No code reuse** - Every test duplicates the setup
4. **Maintenance nightmare** - Change constructor = update every test file

## After: flow_suite_test.go (Suite Pattern)

```go
package character_test

// CharacterFlowSuite tests multi-step character workflows
type CharacterFlowSuite struct {
    suite.Suite
    ctrl           *gomock.Controller
    ctx            context.Context
    mockDNDClient  *mockdnd5e.MockClient
    mockRepository *mockcharrepo.MockRepository
    mockDraftRepo  inmemoryDraft.Repository
    mockResolver   *mockcharacters.MockChoiceResolver
    service        character.Service

    // Test data - available to all tests
    testUserID      string
    testRealmID     string
    testCharacterID string
    elfRace         *rulebook.Race
    monkClass       *rulebook.Class
}

// SetupTest runs before each test method
func (s *CharacterFlowSuite) SetupTest() {
    // ALL setup in ONE place
    s.ctrl = gomock.NewController(s.T())
    s.ctx = context.Background()
    
    // Create all mocks
    s.mockDNDClient = mockdnd5e.NewMockClient(s.ctrl)
    s.mockRepository = mockcharrepo.NewMockRepository(s.ctrl)
    s.mockDraftRepo = inmemoryDraft.NewInMemoryRepository()
    s.mockResolver = mockcharacters.NewMockChoiceResolver(s.ctrl)
    
    // Create service with all dependencies
    s.service = character.NewService(&character.ServiceConfig{
        DNDClient:       s.mockDNDClient,
        Repository:      s.mockRepository,
        DraftRepository: s.mockDraftRepo,  // Can't forget - it's right here!
        ChoiceResolver:  s.mockResolver,
    })
    
    s.setupTestData()
}

// Now tests are clean and focused
func (s *CharacterFlowSuite) TestDiscordCharacterCreationFlow() {
    // No boilerplate! Jump straight to test logic
    charState := &character2.Character{
        ID:      s.testCharacterID,  // Use suite data
        OwnerID: s.testUserID,
        // ... setup specific to this test ...
    }

    s.mockRepository.EXPECT().GetByOwner(s.ctx, s.testUserID).Return([]*character2.Character{charState}, nil)

    chars, err := s.service.ListByOwner(s.testUserID)
    s.Require().NoError(err)
    // ... assertions ...
}

// Adding new tests is trivial
func (s *CharacterFlowSuite) TestCharacterWithNoRace() {
    // Another test - no setup needed!
    charWithoutRace := &character2.Character{
        ID:      "test-char",
        OwnerID: s.testUserID,  // Reuse test data
        Name:    "No Race Character",
        // ... test specific setup ...
    }
    
    s.mockRepository.EXPECT().Get(s.ctx, "test-char").Return(charWithoutRace, nil)
    
    char, err := s.service.GetCharacter(s.ctx, "test-char")
    s.Require().NoError(err)
    s.Nil(char.Race)
}
```

## The Impact of This Change

### Before (Raw Tests):
- **54+ files to update** when DraftRepository was added
- **10+ lines of boilerplate** per test
- **Import cycles** with test helpers
- **4+ hours** to fix all tests

### After (Suite Pattern):
- **5-6 SetupTest methods to update** when dependencies change
- **Zero boilerplate** in test methods
- **No import cycles** - suite handles everything
- **17 minutes** to update all suites

## Real Example: Adding DraftRepository

### With Raw Tests (What Actually Happened):
```bash
# 54+ files failed
FAIL: draft repository is required
FAIL: draft repository is required
FAIL: draft repository is required
... (54 times)

# Had to update EVERY test file:
- internal/services/character/discord_flow_test.go
- internal/services/character/finalize_character_test.go
- internal/services/character/ability_assignment_test.go
... (51 more files)
```

### With Suite Pattern (What Could Have Been):
```bash
# Only update the suite SetupTest methods:
- CharacterServiceTestSuite.SetupTest()      # Core tests
- CharacterFlowSuite.SetupTest()             # Flow tests
- CharacterFinalizationSuite.SetupTest()     # Finalization tests
- CharacterProficiencySuite.SetupTest()      # Proficiency tests
- EncounterServiceTestSuite.SetupTest()      # Encounter tests

# Done in 17 minutes!
```

## Key Takeaways

1. **Test code is production code** - It needs proper architecture
2. **DRY applies to tests too** - Don't repeat setup code
3. **Suites scale better** - Adding tests is trivial
4. **Maintenance matters** - One change shouldn't break 54 files
5. **Clear > Clever** - `TestCharacterWithNoRace()` beats table entry `{"no race", args{...}, false}`

## Migration Strategy

When migrating tests to suites:

1. **Group related tests** - Create suites by functionality
2. **Start with setup** - Move all mock creation to SetupTest
3. **Extract test data** - Put common data on the suite struct
4. **One test at a time** - Migrate incrementally
5. **Delete the old** - Don't leave both versions

The suite pattern isn't just about organization - it's about making tests maintainable and resilient to change.