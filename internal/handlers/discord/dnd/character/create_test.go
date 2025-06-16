package character

import (
	"context"
	"errors"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e/mock"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
)

type CreateHandlerTestSuite struct {
	suite.Suite
	ctrl         *gomock.Controller
	handler      *CreateHandler
	mockClient   *mock.MockClient
	mockSession  *discordgo.Session
	interaction  *discordgo.InteractionCreate
}

func TestCreateHandlerSuite(t *testing.T) {
	suite.Run(t, new(CreateHandlerTestSuite))
}

func (s *CreateHandlerTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockClient = mock.NewMockClient(s.ctrl)
	s.handler = NewCreateHandler(&CreateHandlerConfig{
		DNDClient: s.mockClient,
	})

	// Create a minimal mock session
	s.mockSession = &discordgo.Session{}
	
	s.interaction = &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			Type: discordgo.InteractionApplicationCommand,
			Member: &discordgo.Member{
				User: &discordgo.User{
					ID:       "user123",
					Username: "testuser",
				},
			},
		},
		Session: s.mockSession,
	}
}

func (s *CreateHandlerTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *CreateHandlerTestSuite) TestHandle_Success() {
	// Setup test data
	races := []*entities.Race{
		{
			Key:   "dwarf",
			Name:  "Dwarf",
			Speed: 25,
		},
		{
			Key:   "elf",
			Name:  "Elf", 
			Speed: 30,
		},
		{
			Key:   "human",
			Name:  "Human",
			Speed: 30,
		},
	}

	// Mock expectations
	s.mockClient.EXPECT().
		ListRaces(gomock.Any()).
		Return(races, nil)

	// We need to track the interaction responses
	var deferredResponse *discordgo.InteractionResponse
	var editedResponse *discordgo.WebhookEdit

	// Mock the Session methods
	s.mockSession.InteractionRespond = func(interaction *discordgo.Interaction, resp *discordgo.InteractionResponse) error {
		deferredResponse = resp
		return nil
	}

	s.mockSession.InteractionResponseEdit = func(interaction *discordgo.Interaction, edit *discordgo.WebhookEdit) (*discordgo.Message, error) {
		editedResponse = edit
		return &discordgo.Message{}, nil
	}

	// Execute
	req := &CreateRequest{
		Interaction: s.interaction,
	}
	err := s.handler.Handle(req)

	// Verify
	s.NoError(err)

	// Check deferred response
	s.Require().NotNil(deferredResponse)
	s.Equal(discordgo.InteractionResponseDeferredChannelMessageWithSource, deferredResponse.Type)
	s.True(deferredResponse.Data.Flags&discordgo.MessageFlagsEphemeral != 0)

	// Check edited response
	s.Require().NotNil(editedResponse)
	s.Require().NotNil(editedResponse.Embeds)
	s.Require().Len(*editedResponse.Embeds, 1)
	
	embed := (*editedResponse.Embeds)[0]
	s.Equal("Create New Character - Step 1: Choose Race", embed.Title)
	s.Contains(embed.Description, "Select your character's race")

	// Check components
	s.Require().NotNil(editedResponse.Components)
	s.Require().Len(*editedResponse.Components, 1)
	
	row := (*editedResponse.Components)[0].(discordgo.ActionsRow)
	s.Require().Len(row.Components, 1)
	
	selectMenu := row.Components[0].(discordgo.SelectMenu)
	s.Equal("character_create:race_select", selectMenu.CustomID)
	s.Equal("Choose your race", selectMenu.Placeholder)
	s.Len(selectMenu.Options, 3)

	// Verify options
	s.Equal("Dwarf", selectMenu.Options[0].Label)
	s.Equal("dwarf", selectMenu.Options[0].Value)
	s.Equal("Speed: 25 ft", selectMenu.Options[0].Description)

	s.Equal("Elf", selectMenu.Options[1].Label)
	s.Equal("elf", selectMenu.Options[1].Value)
	s.Equal("Speed: 30 ft", selectMenu.Options[1].Description)
}

func (s *CreateHandlerTestSuite) TestHandle_APIError() {
	// Mock API error
	s.mockClient.EXPECT().
		ListRaces(gomock.Any()).
		Return(nil, errors.New("API error"))

	// Mock session methods
	var editedResponse *discordgo.WebhookEdit

	s.mockSession.InteractionRespond = func(interaction *discordgo.Interaction, resp *discordgo.InteractionResponse) error {
		return nil
	}

	s.mockSession.InteractionResponseEdit = func(interaction *discordgo.Interaction, edit *discordgo.WebhookEdit) (*discordgo.Message, error) {
		editedResponse = edit
		return &discordgo.Message{}, nil
	}

	// Execute
	req := &CreateRequest{
		Interaction: s.interaction,
	}
	err := s.handler.Handle(req)

	// Verify
	s.NoError(err) // The handler returns nil but sends error to user
	s.Require().NotNil(editedResponse)
	s.Require().NotNil(editedResponse.Content)
	s.Equal("Failed to fetch races. Please try again.", *editedResponse.Content)
}