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

type RaceSelectHandlerTestSuite struct {
	suite.Suite
	ctrl         *gomock.Controller
	handler      *RaceSelectHandler
	mockClient   *mock.MockClient
	mockSession  *discordgo.Session
	interaction  *discordgo.InteractionCreate
}

func TestRaceSelectHandlerSuite(t *testing.T) {
	suite.Run(t, new(RaceSelectHandlerTestSuite))
}

func (s *RaceSelectHandlerTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockClient = mock.NewMockClient(s.ctrl)
	s.handler = NewRaceSelectHandler(&RaceSelectHandlerConfig{
		DNDClient: s.mockClient,
	})

	s.mockSession = &discordgo.Session{}
	
	s.interaction = &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			Type: discordgo.InteractionMessageComponent,
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

func (s *RaceSelectHandlerTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *RaceSelectHandlerTestSuite) TestHandle_Success() {
	// Setup test data
	selectedRace := &entities.Race{
		Key:   "dwarf",
		Name:  "Dwarf",
		Speed: 25,
		AbilityBonuses: []*entities.AbilityBonus{
			{Attribute: entities.AttributeConstitution, Bonus: 2},
			{Attribute: entities.AttributeWisdom, Bonus: 1},
		},
		StartingProficiencies: []*entities.ReferenceItem{
			{Name: "Battleaxe"},
			{Name: "Handaxe"},
		},
	}

	classes := []*entities.Class{
		{
			Key:    "fighter",
			Name:   "Fighter",
			HitDie: 10,
		},
		{
			Key:    "wizard",
			Name:   "Wizard",
			HitDie: 6,
		},
	}

	// Mock expectations
	s.mockClient.EXPECT().
		GetRace("dwarf").
		Return(selectedRace, nil)

	s.mockClient.EXPECT().
		ListClasses(gomock.Any()).
		Return(classes, nil)

	// Track responses
	var deferredResponse *discordgo.InteractionResponse
	var editedResponse *discordgo.WebhookEdit

	s.mockSession.InteractionRespond = func(interaction *discordgo.Interaction, resp *discordgo.InteractionResponse) error {
		deferredResponse = resp
		return nil
	}

	s.mockSession.InteractionResponseEdit = func(interaction *discordgo.Interaction, edit *discordgo.WebhookEdit) (*discordgo.Message, error) {
		editedResponse = edit
		return &discordgo.Message{}, nil
	}

	// Execute
	req := &RaceSelectRequest{
		Interaction: s.interaction,
		RaceKey:     "dwarf",
	}
	err := s.handler.Handle(req)

	// Verify
	s.NoError(err)
	s.NotNil(deferredResponse)
	s.NotNil(editedResponse)

	// Check embed
	s.Require().Len(*editedResponse.Embeds, 1)
	embed := (*editedResponse.Embeds)[0]
	s.Equal("Selected Race: Dwarf", embed.Title)
	s.Equal("Step 2: Choose your class", embed.Description)

	// Check fields
	fieldMap := make(map[string]string)
	for _, field := range embed.Fields {
		fieldMap[field.Name] = field.Value
	}

	s.Equal("25 feet", fieldMap["Speed"])
	s.Equal("Con +2, Wis +1", fieldMap["Ability Bonuses"])
	s.Equal("Battleaxe, Handaxe", fieldMap["Racial Proficiencies"])

	// Check class selection menu
	s.Require().Len(*editedResponse.Components, 1)
	row := (*editedResponse.Components)[0].(discordgo.ActionsRow)
	selectMenu := row.Components[0].(discordgo.SelectMenu)

	s.Equal("character_create:class_select:dwarf", selectMenu.CustomID)
	s.Equal("Choose your class", selectMenu.Placeholder)
	s.Len(selectMenu.Options, 2)

	s.Equal("Fighter", selectMenu.Options[0].Label)
	s.Equal("fighter", selectMenu.Options[0].Value)
	s.Equal("Hit Die: d10", selectMenu.Options[0].Description)
}

func (s *RaceSelectHandlerTestSuite) TestHandle_RaceNotFound() {
	// Mock race not found
	s.mockClient.EXPECT().
		GetRace("invalid").
		Return(nil, errors.New("race not found"))

	// Track responses
	var editedResponse *discordgo.WebhookEdit

	s.mockSession.InteractionRespond = func(interaction *discordgo.Interaction, resp *discordgo.InteractionResponse) error {
		return nil
	}

	s.mockSession.InteractionResponseEdit = func(interaction *discordgo.Interaction, edit *discordgo.WebhookEdit) (*discordgo.Message, error) {
		editedResponse = edit
		return &discordgo.Message{}, nil
	}

	// Execute
	req := &RaceSelectRequest{
		Interaction: s.interaction,
		RaceKey:     "invalid",
	}
	err := s.handler.Handle(req)

	// Verify
	s.NoError(err)
	s.Require().NotNil(editedResponse)
	s.Require().NotNil(editedResponse.Content)
	s.Equal("Failed to fetch race details. Please try again.", *editedResponse.Content)
}