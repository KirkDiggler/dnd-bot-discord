package character_test

import (
	"context"
	"errors"
	"testing"

	mockdnd5e "github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e/mock"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	mockrepo "github.com/KirkDiggler/dnd-bot-discord/internal/repositories/characters/mock"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	mockcharacters "github.com/KirkDiggler/dnd-bot-discord/internal/services/character/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

// CharacterServiceTestSuite defines the test suite for character service
type CharacterServiceTestSuite struct {
	suite.Suite
	ctrl           *gomock.Controller
	mockDNDClient  *mockdnd5e.MockClient
	mockResolver   *mockcharacters.MockChoiceResolver
	mockRepository *mockrepo.MockRepository
	service        character.Service
	ctx            context.Context
}

// SetupTest runs before each test
func (s *CharacterServiceTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockDNDClient = mockdnd5e.NewMockClient(s.ctrl)
	s.mockResolver = mockcharacters.NewMockChoiceResolver(s.ctrl)
	s.mockRepository = mockrepo.NewMockRepository(s.ctrl)
	s.ctx = context.Background()

	s.service = character.NewService(&character.ServiceConfig{
		DNDClient:      s.mockDNDClient,
		ChoiceResolver: s.mockResolver,
		Repository:     s.mockRepository,
	})
}

// TearDownTest runs after each test
func (s *CharacterServiceTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

// Test suite runner
func TestCharacterServiceSuite(t *testing.T) {
	suite.Run(t, new(CharacterServiceTestSuite))
}

// ResolveChoices Tests

func (s *CharacterServiceTestSuite) TestResolveChoices_Success() {
	// Setup
	input := &character.ResolveChoicesInput{
		RaceKey:  "human",
		ClassKey: "fighter",
	}

	humanRace := &entities.Race{
		Key:  "human",
		Name: "Human",
	}

	fighterClass := &entities.Class{
		Key:  "fighter",
		Name: "Fighter",
	}

	profChoices := []character.SimplifiedChoice{
		{
			ID:     "fighter-prof-0",
			Name:   "Fighter Skills",
			Choose: 2,
		},
	}

	equipChoices := []character.SimplifiedChoice{
		{
			ID:     "fighter-equip-0",
			Name:   "Starting Weapon",
			Choose: 1,
		},
	}

	// Expectations
	s.mockDNDClient.EXPECT().GetRace("human").Return(humanRace, nil)
	s.mockDNDClient.EXPECT().GetClass("fighter").Return(fighterClass, nil)
	s.mockResolver.EXPECT().ResolveProficiencyChoices(s.ctx, humanRace, fighterClass).Return(profChoices, nil)
	s.mockResolver.EXPECT().ResolveEquipmentChoices(s.ctx, fighterClass).Return(equipChoices, nil)

	// Execute
	output, err := s.service.ResolveChoices(s.ctx, input)

	// Assert
	s.NoError(err)
	s.NotNil(output)
	s.Len(output.ProficiencyChoices, 1)
	s.Len(output.EquipmentChoices, 1)
	s.Equal("fighter-prof-0", output.ProficiencyChoices[0].ID)
	s.Equal("fighter-equip-0", output.EquipmentChoices[0].ID)
}

func (s *CharacterServiceTestSuite) TestResolveChoices_RaceNotFound() {
	// Setup
	input := &character.ResolveChoicesInput{
		RaceKey:  "invalid-race",
		ClassKey: "fighter",
	}

	// Expectations
	s.mockDNDClient.EXPECT().GetRace("invalid-race").Return(nil, errors.New("race not found"))

	// Execute
	output, err := s.service.ResolveChoices(s.ctx, input)

	// Assert
	s.Error(err)
	s.Nil(output)
	s.Contains(err.Error(), "failed to get race")
}

func (s *CharacterServiceTestSuite) TestResolveChoices_ClassNotFound() {
	// Setup
	input := &character.ResolveChoicesInput{
		RaceKey:  "human",
		ClassKey: "invalid-class",
	}

	// Expectations
	s.mockDNDClient.EXPECT().GetRace("human").Return(&entities.Race{Key: "human", Name: "Human"}, nil)
	s.mockDNDClient.EXPECT().GetClass("invalid-class").Return(nil, errors.New("class not found"))

	// Execute
	output, err := s.service.ResolveChoices(s.ctx, input)

	// Assert
	s.Error(err)
	s.Nil(output)
	s.Contains(err.Error(), "failed to get class")
}

// ValidateCharacterCreation Tests

func (s *CharacterServiceTestSuite) TestValidateCharacterCreation_ValidInput() {
	// Setup
	input := &character.ValidateCharacterInput{
		RaceKey:  "human",
		ClassKey: "fighter",
		AbilityScores: map[string]int{
			"STR": 15,
			"DEX": 14,
			"CON": 13,
			"INT": 12,
			"WIS": 10,
			"CHA": 8,
		},
	}

	// Execute
	err := s.service.ValidateCharacterCreation(s.ctx, input)

	// Assert
	s.NoError(err)
}

func (s *CharacterServiceTestSuite) TestValidateCharacterCreation_MissingRace() {
	// Setup
	input := &character.ValidateCharacterInput{
		ClassKey: "fighter",
		AbilityScores: map[string]int{
			"STR": 15,
			"DEX": 14,
			"CON": 13,
			"INT": 12,
			"WIS": 10,
			"CHA": 8,
		},
	}

	// Execute
	err := s.service.ValidateCharacterCreation(s.ctx, input)

	// Assert
	s.Error(err)
	s.Contains(err.Error(), "race is required")
}

func (s *CharacterServiceTestSuite) TestValidateCharacterCreation_ScoreTooLow() {
	// Setup
	input := &character.ValidateCharacterInput{
		RaceKey:  "human",
		ClassKey: "fighter",
		AbilityScores: map[string]int{
			"STR": 2, // Too low!
			"DEX": 14,
			"CON": 13,
			"INT": 12,
			"WIS": 10,
			"CHA": 8,
		},
	}

	// Execute
	err := s.service.ValidateCharacterCreation(s.ctx, input)

	// Assert
	s.Error(err)
	s.Contains(err.Error(), "ability score for STR must be between 3 and 18")
}

func (s *CharacterServiceTestSuite) TestValidateCharacterCreation_MissingAbility() {
	// Setup
	input := &character.ValidateCharacterInput{
		RaceKey:  "human",
		ClassKey: "fighter",
		AbilityScores: map[string]int{
			"STR": 15,
			"DEX": 14,
			"CON": 13,
			"INT": 12,
			"WIS": 10,
			// Missing CHA!
		},
	}

	// Execute
	err := s.service.ValidateCharacterCreation(s.ctx, input)

	// Assert
	s.Error(err)
	s.Contains(err.Error(), "missing ability score for CHA")
}

// CreateCharacter Tests

func (s *CharacterServiceTestSuite) TestCreateCharacter_Success() {
	// Setup
	input := &character.CreateCharacterInput{
		UserID:   "user123",
		RealmID:  "guild456",
		Name:     "Thorin",
		RaceKey:  "dwarf",
		ClassKey: "fighter",
		AbilityScores: map[string]int{
			"STR": 16,
			"DEX": 12,
			"CON": 15,
			"INT": 10,
			"WIS": 13,
			"CHA": 8,
		},
		Proficiencies: []string{"skill-athletics", "skill-intimidation"},
		Equipment:     []string{"chain-mail", "longsword", "shield"},
	}

	dwarfRace := &entities.Race{
		Key:   "dwarf",
		Name:  "Dwarf",
		Speed: 25,
		AbilityBonuses: []*entities.AbilityBonus{
			{Attribute: entities.AttributeConstitution, Bonus: 2},
		},
	}

	fighterClass := &entities.Class{
		Key:    "fighter",
		Name:   "Fighter",
		HitDie: 10,
		StartingEquipment: []*entities.StartingEquipment{
			{Quantity: 1, Equipment: &entities.ReferenceItem{Key: "chain-shirt"}},
		},
	}

	// Expectations
	s.mockDNDClient.EXPECT().GetRace("dwarf").Return(dwarfRace, nil)
	s.mockDNDClient.EXPECT().GetClass("fighter").Return(fighterClass, nil)

	// Proficiency expectations
	s.mockDNDClient.EXPECT().GetProficiency("skill-athletics").Return(&entities.Proficiency{
		Key:  "skill-athletics",
		Name: "Athletics",
		Type: entities.ProficiencyTypeSkill,
	}, nil)
	s.mockDNDClient.EXPECT().GetProficiency("skill-intimidation").Return(&entities.Proficiency{
		Key:  "skill-intimidation",
		Name: "Intimidation",
		Type: entities.ProficiencyTypeSkill,
	}, nil)

	// Equipment expectations
	s.mockDNDClient.EXPECT().GetEquipment("chain-shirt").Return(&entities.BasicEquipment{
		Key:  "chain-shirt",
		Name: "Chain Shirt",
	}, nil)
	s.mockDNDClient.EXPECT().GetEquipment("chain-mail").Return(&entities.Armor{
		Base: entities.BasicEquipment{Key: "chain-mail", Name: "Chain Mail"},
	}, nil)
	s.mockDNDClient.EXPECT().GetEquipment("longsword").Return(&entities.Weapon{
		Base: entities.BasicEquipment{Key: "longsword", Name: "Longsword"},
	}, nil)
	s.mockDNDClient.EXPECT().GetEquipment("shield").Return(&entities.BasicEquipment{
		Key:  "shield",
		Name: "Shield",
	}, nil)

	// Repository expectation
	s.mockRepository.EXPECT().Create(s.ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, char *entities.Character) error {
		// Validate the character being saved
		s.Equal("Thorin", char.Name)
		s.Equal("user123", char.OwnerID)
		s.Equal("guild456", char.RealmID)
		s.Equal(25, char.Speed)
		s.Equal(10, char.HitDie)
		s.Equal(1, char.Level)
		s.NotEmpty(char.ID)
		return nil
	})

	// Execute
	output, err := s.service.CreateCharacter(s.ctx, input)

	// Assert
	s.NoError(err)
	s.NotNil(output)
	s.NotNil(output.Character)
	s.Equal("Thorin", output.Character.Name)
	s.Equal("user123", output.Character.OwnerID)
	s.Equal("guild456", output.Character.RealmID)
	s.NotEmpty(output.Character.ID)
}

func (s *CharacterServiceTestSuite) TestCreateCharacter_ValidationFailure() {
	// Setup
	input := &character.CreateCharacterInput{
		UserID:   "user123",
		RealmID:  "guild456",
		Name:     "BadChar",
		RaceKey:  "human",
		ClassKey: "fighter",
		AbilityScores: map[string]int{
			"STR": 15,
			"DEX": 14,
			// Missing other abilities
		},
	}

	// Execute
	output, err := s.service.CreateCharacter(s.ctx, input)

	// Assert
	s.Error(err)
	s.Nil(output)
	s.Contains(err.Error(), "invalid character creation input")
}

func (s *CharacterServiceTestSuite) TestCreateCharacter_RepositoryError() {
	// Setup
	input := &character.CreateCharacterInput{
		UserID:   "user123",
		RealmID:  "guild456",
		Name:     "Thorin",
		RaceKey:  "dwarf",
		ClassKey: "fighter",
		AbilityScores: map[string]int{
			"STR": 16,
			"DEX": 12,
			"CON": 15,
			"INT": 10,
			"WIS": 13,
			"CHA": 8,
		},
	}

	// Expectations
	s.mockDNDClient.EXPECT().GetRace("dwarf").Return(&entities.Race{
		Key:   "dwarf",
		Name:  "Dwarf",
		Speed: 25,
	}, nil)
	s.mockDNDClient.EXPECT().GetClass("fighter").Return(&entities.Class{
		Key:    "fighter",
		Name:   "Fighter",
		HitDie: 10,
	}, nil)
	s.mockRepository.EXPECT().Create(s.ctx, gomock.Any()).Return(errors.New("database error"))

	// Execute
	output, err := s.service.CreateCharacter(s.ctx, input)

	// Assert
	s.Error(err)
	s.Nil(output)
	s.Contains(err.Error(), "failed to save character")
}

// GetCharacter Tests

func (s *CharacterServiceTestSuite) TestGetCharacter_Success() {
	// Setup
	expectedChar := &entities.Character{
		ID:      "char_123",
		Name:    "Thorin",
		OwnerID: "user_123",
	}

	// Expectations
	s.mockRepository.EXPECT().Get(s.ctx, "char_123").Return(expectedChar, nil)

	// Execute
	char, err := s.service.GetCharacter(s.ctx, "char_123")

	// Assert
	s.NoError(err)
	s.NotNil(char)
	s.Equal("Thorin", char.Name)
}

func (s *CharacterServiceTestSuite) TestGetCharacter_NotFound() {
	// Expectations
	s.mockRepository.EXPECT().Get(s.ctx, "nonexistent").Return(nil, errors.New("not found"))

	// Execute
	char, err := s.service.GetCharacter(s.ctx, "nonexistent")

	// Assert
	s.Error(err)
	s.Nil(char)
}

// ListCharacters Tests

func (s *CharacterServiceTestSuite) TestListCharacters_Success() {
	// Setup
	expectedChars := []*entities.Character{
		{ID: "char_1", Name: "Thorin", OwnerID: "user_123"},
		{ID: "char_2", Name: "Gandalf", OwnerID: "user_123"},
	}

	// Expectations
	s.mockRepository.EXPECT().GetByOwner(s.ctx, "user_123").Return(expectedChars, nil)

	// Execute
	chars, err := s.service.ListCharacters(s.ctx, "user_123")

	// Assert
	s.NoError(err)
	s.Len(chars, 2)
	s.Equal("Thorin", chars[0].Name)
	s.Equal("Gandalf", chars[1].Name)
}

func (s *CharacterServiceTestSuite) TestListCharacters_Empty() {
	// Expectations
	s.mockRepository.EXPECT().GetByOwner(s.ctx, "user_123").Return([]*entities.Character{}, nil)

	// Execute
	chars, err := s.service.ListCharacters(s.ctx, "user_123")

	// Assert
	s.NoError(err)
	s.Empty(chars)
}
