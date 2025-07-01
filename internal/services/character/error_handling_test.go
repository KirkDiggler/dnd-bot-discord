package character_test

import (
	"context"
	"errors"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
	"testing"

	mockdnd5e "github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e/mock"
	dnderr "github.com/KirkDiggler/dnd-bot-discord/internal/errors"
	mockrepo "github.com/KirkDiggler/dnd-bot-discord/internal/repositories/characters/mock"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

// ErrorHandlingTestSuite tests error handling and propagation
type ErrorHandlingTestSuite struct {
	suite.Suite
	ctrl           *gomock.Controller
	mockDNDClient  *mockdnd5e.MockClient
	mockRepository *mockrepo.MockRepository
	service        character.Service
	ctx            context.Context
}

func (s *ErrorHandlingTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockDNDClient = mockdnd5e.NewMockClient(s.ctrl)
	s.mockRepository = mockrepo.NewMockRepository(s.ctrl)
	s.ctx = context.Background()

	s.service = character.NewService(&character.ServiceConfig{
		DNDClient:  s.mockDNDClient,
		Repository: s.mockRepository,
	})
}

func (s *ErrorHandlingTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func TestErrorHandlingSuite(t *testing.T) {
	suite.Run(t, new(ErrorHandlingTestSuite))
}

func (s *ErrorHandlingTestSuite) TestCreateCharacter_ValidationError() {
	// Test nil input
	output, err := s.service.CreateCharacter(s.ctx, nil)

	s.Error(err)
	s.Nil(output)
	s.True(dnderr.IsInvalidArgument(err))
	s.Contains(err.Error(), "invalid character creation input")

	// Check metadata
	meta := dnderr.GetMeta(err)
	s.Equal("CreateCharacter", meta["operation"])
}

func (s *ErrorHandlingTestSuite) TestCreateCharacter_RaceNotFoundWithContext() {
	// Setup
	input := &character.CreateCharacterInput{
		UserID:   "user123",
		RealmID:  "guild456",
		Name:     "Thorin",
		RaceKey:  "invalid-race",
		ClassKey: "fighter",
		AbilityScores: map[string]int{
			"STR": 16, "DEX": 12, "CON": 15,
			"INT": 10, "WIS": 13, "CHA": 8,
		},
	}

	// Mock race not found
	s.mockDNDClient.EXPECT().GetRace("invalid-race").
		Return(nil, errors.New("404: race not found"))

	// Execute
	output, err := s.service.CreateCharacter(s.ctx, input)

	// Assert
	s.Error(err)
	s.Nil(output)
	s.Contains(err.Error(), "failed to get race 'invalid-race'")
	s.Contains(err.Error(), "404: race not found")

	// Check metadata
	meta := dnderr.GetMeta(err)
	s.Equal("invalid-race", meta["race_key"])
}

func (s *ErrorHandlingTestSuite) TestCreateCharacter_RepositoryErrorWithContext() {
	// Setup
	input := &character.CreateCharacterInput{
		UserID:   "user123",
		RealmID:  "guild456",
		Name:     "Thorin",
		RaceKey:  "dwarf",
		ClassKey: "fighter",
		AbilityScores: map[string]int{
			"STR": 16, "DEX": 12, "CON": 15,
			"INT": 10, "WIS": 13, "CHA": 8,
		},
	}

	// Setup mocks
	s.mockDNDClient.EXPECT().GetRace("dwarf").Return(&rulebook.Race{
		Key: "dwarf", Name: "Dwarf", Speed: 25,
	}, nil)
	s.mockDNDClient.EXPECT().GetClass("fighter").Return(&rulebook.Class{
		Key: "fighter", Name: "Fighter", HitDie: 10,
	}, nil)

	// Repository returns already exists error
	s.mockRepository.EXPECT().Create(s.ctx, gomock.Any()).
		Return(dnderr.AlreadyExists("character already exists").
			WithMeta("existing_id", "char_123"))

	// Execute
	output, err := s.service.CreateCharacter(s.ctx, input)

	// Assert
	s.Error(err)
	s.Nil(output)
	s.Contains(err.Error(), "failed to save character")
	s.Contains(err.Error(), "character already exists")

	// Check error propagation - should preserve original error code
	s.True(dnderr.IsAlreadyExists(err))

	// Check metadata accumulation
	meta := dnderr.GetMeta(err)
	s.Equal("Thorin", meta["character_name"])
	s.Equal("user123", meta["owner_id"])
	s.NotEmpty(meta["character_id"])
	s.Equal("char_123", meta["existing_id"]) // From original error
}

func (s *ErrorHandlingTestSuite) TestGetCharacter_EmptyID() {
	// Execute
	char, err := s.service.GetCharacter(s.ctx, "")

	// Assert
	s.Error(err)
	s.Nil(char)
	s.True(dnderr.IsInvalidArgument(err))
	s.Equal("character ID is required", err.Error())
}

func (s *ErrorHandlingTestSuite) TestGetCharacter_NotFoundWithContext() {
	// Mock repository returns not found
	s.mockRepository.EXPECT().Get(s.ctx, "char_999").
		Return(nil, dnderr.NotFound("character not found").
			WithMeta("searched_at", "2024-01-01"))

	// Execute
	char, err := s.service.GetCharacter(s.ctx, "char_999")

	// Assert
	s.Error(err)
	s.Nil(char)
	s.True(dnderr.IsNotFound(err))
	s.Contains(err.Error(), "failed to get character 'char_999'")

	// Check metadata
	meta := dnderr.GetMeta(err)
	s.Equal("char_999", meta["character_id"])
	s.Equal("2024-01-01", meta["searched_at"]) // From original error
}

func (s *ErrorHandlingTestSuite) TestResolveChoices_InvalidInput() {
	// Test nil input
	output, err := s.service.ResolveChoices(s.ctx, nil)

	s.Error(err)
	s.Nil(output)
	s.True(dnderr.IsInvalidArgument(err))
	s.Contains(err.Error(), "invalid resolve choices input")

	// Test empty race key
	output, err = s.service.ResolveChoices(s.ctx, &character.ResolveChoicesInput{
		RaceKey:  "",
		ClassKey: "fighter",
	})

	s.Error(err)
	s.Nil(output)
	s.True(dnderr.IsInvalidArgument(err))
	s.Contains(err.Error(), "race key is required")
}

func (s *ErrorHandlingTestSuite) TestErrorChaining() {
	// Create a chain of errors
	baseErr := errors.New("connection timeout")

	// Wrap at repository level
	repoErr := dnderr.WrapWithCode(baseErr, dnderr.CodeUnavailable, "database unavailable").
		WithMeta("retry_after", 30).
		WithMeta("connection_pool", "exhausted")

	// Wrap at service level
	serviceErr := dnderr.Wrap(repoErr, "failed to save character").
		WithMeta("character_name", "Thorin").
		WithMeta("operation", "CreateCharacter")

	// Assert error chain
	s.Error(serviceErr)
	s.Contains(serviceErr.Error(), "failed to save character")
	s.Contains(serviceErr.Error(), "database unavailable")
	s.Contains(serviceErr.Error(), "connection timeout")

	// Check error code is preserved
	s.True(dnderr.Is(serviceErr, dnderr.CodeUnavailable))

	// Check all metadata is preserved
	meta := dnderr.GetMeta(serviceErr)
	s.Equal(30, meta["retry_after"])
	s.Equal("exhausted", meta["connection_pool"])
	s.Equal("Thorin", meta["character_name"])
	s.Equal("CreateCharacter", meta["operation"])

	// Check we can still access the base error
	s.True(errors.Is(serviceErr, baseErr))
}
