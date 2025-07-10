package character_test

import (
	"context"
	"testing"

	mockdnd5e "github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e/mock"
	character2 "github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/equipment"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	inmemoryDraft "github.com/KirkDiggler/dnd-bot-discord/internal/repositories/character_draft"
	mockcharrepo "github.com/KirkDiggler/dnd-bot-discord/internal/repositories/characters/mock"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	mockcharacters "github.com/KirkDiggler/dnd-bot-discord/internal/services/character/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

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

	// Test data
	testUserID      string
	testRealmID     string
	testCharacterID string
	elfRace         *rulebook.Race
	monkClass       *rulebook.Class
}

// SetupTest runs before each test method
func (s *CharacterFlowSuite) SetupTest() {
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
		DraftRepository: s.mockDraftRepo,
		ChoiceResolver:  s.mockResolver,
	})

	// Initialize test data
	s.setupTestData()
}

// TearDownTest runs after each test
func (s *CharacterFlowSuite) TearDownTest() {
	s.ctrl.Finish()
}

// setupTestData initializes common test data
func (s *CharacterFlowSuite) setupTestData() {
	s.testUserID = "discord_user"
	s.testRealmID = "discord_realm"
	s.testCharacterID = "char_123"

	// Common race data
	s.elfRace = &rulebook.Race{
		Key:  "elf",
		Name: "Elf",
		AbilityBonuses: []*shared.AbilityBonus{
			{Attribute: shared.AttributeDexterity, Bonus: 2},
		},
	}

	// Common class data
	s.monkClass = &rulebook.Class{
		Key:    "monk",
		Name:   "Monk",
		HitDie: 8,
	}
}

// TestDiscordCharacterCreationFlow simulates the exact Discord flow to identify where attributes might be lost
// This was migrated from discord_flow_test.go
func (s *CharacterFlowSuite) TestDiscordCharacterCreationFlow() {
	// Setup character state after ability assignment
	charState := &character2.Character{
		ID:      s.testCharacterID,
		OwnerID: s.testUserID,
		RealmID: s.testRealmID,
		Name:    "NotGonnaWorkHere",
		Status:  shared.CharacterStatusActive, // Already finalized
		Level:   1,
		Race:    s.elfRace,
		Class:   s.monkClass,
		// This is what might be happening - character was finalized but attributes weren't saved
		AbilityRolls: []character2.AbilityRoll{
			{ID: "roll_1", Value: 15},
			{ID: "roll_2", Value: 14},
			{ID: "roll_3", Value: 13},
			{ID: "roll_4", Value: 12},
			{ID: "roll_5", Value: 11},
			{ID: "roll_6", Value: 10},
		},
		AbilityAssignments: map[string]string{
			"STR": "roll_3",
			"DEX": "roll_2",
			"CON": "roll_4",
			"INT": "roll_1",
			"WIS": "roll_5",
			"CHA": "roll_6",
		},
		Attributes:       map[shared.Attribute]*character2.AbilityScore{}, // Empty!
		Proficiencies:    make(map[rulebook.ProficiencyType][]*rulebook.Proficiency),
		Inventory:        make(map[equipment.EquipmentType][]equipment.Equipment),
		EquippedSlots:    make(map[shared.Slot]equipment.Equipment),
		MaxHitPoints:     9,
		CurrentHitPoints: 9,
		AC:               13,
	}

	// Test 1: ListByOwner returns character with empty attributes
	s.mockRepository.EXPECT().GetByOwner(s.ctx, s.testUserID).Return([]*character2.Character{charState}, nil)

	chars, err := s.service.ListByOwner(s.testUserID)
	s.Require().NoError(err)
	s.Require().Len(chars, 1)

	char := chars[0]
	s.T().Logf("Character from ListByOwner - Name: %s, Status: %s, Attributes: %d, AbilityAssignments: %d",
		char.Name, char.Status, len(char.Attributes), len(char.AbilityAssignments))

	// This reproduces the bug
	s.Empty(char.Attributes, "Character has empty attributes")
	s.NotEmpty(char.AbilityAssignments, "Character has ability assignments")
	s.False(char.IsComplete(), "Character shows as incomplete due to missing attributes")

	// Test 2: What happens if we try to "fix" the character
	// The service doesn't have a method to fix already finalized characters
	// This might be the root cause - characters get finalized before attributes are properly converted
}

// TestCharacterWithNoRace demonstrates what happens when we forget to set race
func (s *CharacterFlowSuite) TestCharacterWithNoRace() {
	// Create a character without race
	charWithoutRace := &character2.Character{
		ID:      "test-char",
		OwnerID: s.testUserID,
		RealmID: s.testRealmID,
		Name:    "No Race Character",
		Status:  shared.CharacterStatusDraft,
		Level:   1,
		Class:   s.monkClass,
		// No race set!
	}

	// Mock the repository to return this character
	s.mockRepository.EXPECT().Get(s.ctx, "test-char").Return(charWithoutRace, nil)

	// Try to get the character
	char, err := s.service.GetCharacter(s.ctx, "test-char")
	s.Require().NoError(err)
	s.Nil(char.Race, "Character should have no race")
	s.False(char.IsComplete(), "Character without race should not be complete")
}

// TestMultipleCharacterCreation tests creating multiple characters in sequence
func (s *CharacterFlowSuite) TestMultipleCharacterCreation() {
	// This demonstrates how the suite pattern makes it easy to test multiple scenarios

	// Create first character
	s.mockRepository.EXPECT().Create(s.ctx, gomock.Any()).Return(nil)
	char1, err := s.service.GetOrCreateDraftCharacter(s.ctx, "user1", "realm1")
	s.Require().NoError(err)
	s.NotNil(char1)

	// Create second character for different user
	s.mockRepository.EXPECT().Create(s.ctx, gomock.Any()).Return(nil)
	char2, err := s.service.GetOrCreateDraftCharacter(s.ctx, "user2", "realm1")
	s.Require().NoError(err)
	s.NotNil(char2)

	// Verify they have different IDs
	s.NotEqual(char1.ID, char2.ID, "Characters should have unique IDs")
}

// Run the test suite
func TestCharacterFlowSuite(t *testing.T) {
	suite.Run(t, new(CharacterFlowSuite))
}

// MIGRATION NOTES:
//
// This suite demonstrates the benefits of the suite pattern:
// 1. All mock setup is in one place (SetupTest)
// 2. Common test data is defined once (setupTestData)
// 3. Each test method is focused on the scenario, not setup
// 4. When we added DraftRepository, we only had to update SetupTest
//
// Before: Every test file had to create mocks and service
// After: Just add a test method to the appropriate suite
//
// Key Benefits Demonstrated:
// - No import cycles (suite is in _test package)
// - DRY principle applied to test code
// - Clear test names instead of table entries
// - Easy to add new tests without copy/paste
// - Single point of update when dependencies change
