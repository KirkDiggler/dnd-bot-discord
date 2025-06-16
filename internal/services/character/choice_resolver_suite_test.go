package character_test

import (
	"context"
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	mockdnd5e "github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

// ChoiceResolverTestSuite defines the test suite for choice resolver
type ChoiceResolverTestSuite struct {
	suite.Suite
	ctrl          *gomock.Controller
	mockDNDClient *mockdnd5e.MockClient
	resolver      character.ChoiceResolver
	ctx           context.Context
}

// SetupTest runs before each test
func (s *ChoiceResolverTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockDNDClient = mockdnd5e.NewMockClient(s.ctrl)
	s.ctx = context.Background()
	s.resolver = character.NewChoiceResolver(s.mockDNDClient)
}

// TearDownTest runs after each test
func (s *ChoiceResolverTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

// Test suite runner
func TestChoiceResolverSuite(t *testing.T) {
	suite.Run(t, new(ChoiceResolverTestSuite))
}

// ResolveProficiencyChoices Tests

func (s *ChoiceResolverTestSuite) TestResolveProficiencyChoices_FighterWithSkillChoices() {
	// Setup
	race := &entities.Race{
		Key:  "human",
		Name: "Human",
	}
	
	class := &entities.Class{
		Key:  "fighter",
		Name: "Fighter",
		ProficiencyChoices: []*entities.Choice{
			{
				Name:  "Choose 2 skills",
				Count: 2,
				Type:  entities.ChoiceTypeProficiency,
				Options: []entities.Option{
					&entities.ReferenceOption{
						Reference: &entities.ReferenceItem{
							Key:  "skill-athletics",
							Name: "Athletics",
						},
					},
					&entities.ReferenceOption{
						Reference: &entities.ReferenceItem{
							Key:  "skill-intimidation",
							Name: "Intimidation",
						},
					},
				},
			},
		},
	}
	
	// Execute
	choices, err := s.resolver.ResolveProficiencyChoices(s.ctx, race, class)
	
	// Assert
	s.NoError(err)
	s.Len(choices, 1)
	s.Equal("fighter-prof-0", choices[0].ID)
	s.Equal("Choose 2 skills", choices[0].Name)
	s.Equal(2, choices[0].Choose)
	s.Len(choices[0].Options, 2)
	s.Equal("skill-athletics", choices[0].Options[0].Key)
	s.Equal("Athletics", choices[0].Options[0].Name)
}

func (s *ChoiceResolverTestSuite) TestResolveProficiencyChoices_MonkWithNestedToolChoice() {
	// Setup
	race := &entities.Race{
		Key:  "human",
		Name: "Human",
	}
	
	class := &entities.Class{
		Key:  "monk",
		Name: "Monk",
		ProficiencyChoices: []*entities.Choice{
			{
				Name:  "Choose 2 skills",
				Count: 2,
				Type:  entities.ChoiceTypeProficiency,
				Options: []entities.Option{
					&entities.ReferenceOption{
						Reference: &entities.ReferenceItem{
							Key:  "skill-acrobatics",
							Name: "Acrobatics",
						},
					},
				},
			},
			{
				Name:  "Choose 1 tool or instrument",
				Count: 1,
				Type:  entities.ChoiceTypeProficiency,
				Options: []entities.Option{
					// Nested choice detected
					&entities.Choice{
						Name:  "Artisan's Tools",
						Count: 1,
					},
				},
			},
		},
	}
	
	// Execute
	choices, err := s.resolver.ResolveProficiencyChoices(s.ctx, race, class)
	
	// Assert
	s.NoError(err)
	s.Len(choices, 2)
	
	// First choice should be normal
	s.Equal("monk-prof-0", choices[0].ID)
	s.Equal("Choose 2 skills", choices[0].Name)
	
	// Second choice should be flattened
	s.Equal("monk-prof-1", choices[1].ID)
	s.Equal("Tools or Instrument", choices[1].Name)
	s.Equal("Choose 1 artisan's tool or musical instrument", choices[1].Description)
	s.GreaterOrEqual(len(choices[1].Options), 3) // Should have multiple tool options
}

func (s *ChoiceResolverTestSuite) TestResolveProficiencyChoices_RaceWithProficiencyChoices() {
	// Setup
	race := &entities.Race{
		Key:  "half-elf",
		Name: "Half-Elf",
		StartingProficiencyOptions: &entities.Choice{
			Name:  "Choose 2 skills",
			Count: 2,
			Type:  entities.ChoiceTypeProficiency,
			Options: []entities.Option{
				&entities.ReferenceOption{
					Reference: &entities.ReferenceItem{
						Key:  "skill-perception",
						Name: "Perception",
					},
				},
				&entities.ReferenceOption{
					Reference: &entities.ReferenceItem{
						Key:  "skill-insight",
						Name: "Insight",
					},
				},
			},
		},
	}
	
	class := &entities.Class{
		Key:  "fighter",
		Name: "Fighter",
		// No proficiency choices
	}
	
	// Execute
	choices, err := s.resolver.ResolveProficiencyChoices(s.ctx, race, class)
	
	// Assert
	s.NoError(err)
	s.Len(choices, 1)
	s.Equal("half-elf-prof", choices[0].ID)
	s.Equal("Choose 2 skills", choices[0].Name)
	s.Contains(choices[0].Description, "racial proficiency")
	s.Len(choices[0].Options, 2)
}

func (s *ChoiceResolverTestSuite) TestResolveProficiencyChoices_EmptyChoices() {
	// Setup
	race := &entities.Race{
		Key:  "human",
		Name: "Human",
	}
	
	class := &entities.Class{
		Key:  "fighter",
		Name: "Fighter",
		// No proficiency choices
	}
	
	// Execute
	choices, err := s.resolver.ResolveProficiencyChoices(s.ctx, race, class)
	
	// Assert
	s.NoError(err)
	s.Empty(choices)
}

// ResolveEquipmentChoices Tests

func (s *ChoiceResolverTestSuite) TestResolveEquipmentChoices_FighterEquipment() {
	// Setup
	class := &entities.Class{
		Key:  "fighter",
		Name: "Fighter",
		StartingEquipmentChoices: []*entities.Choice{
			{
				Name:  "Choose armor",
				Count: 1,
				Options: []entities.Option{
					&entities.ReferenceOption{
						Reference: &entities.ReferenceItem{
							Key:  "chain-mail",
							Name: "Chain Mail",
						},
					},
					&entities.ReferenceOption{
						Reference: &entities.ReferenceItem{
							Key:  "leather-armor",
							Name: "Leather Armor",
						},
					},
				},
			},
			{
				Name:  "Choose weapon",
				Count: 1,
				Options: []entities.Option{
					&entities.CountedReferenceOption{
						Count: 2,
						Reference: &entities.ReferenceItem{
							Key:  "shortsword",
							Name: "Shortsword",
						},
					},
				},
			},
		},
	}
	
	// Execute
	choices, err := s.resolver.ResolveEquipmentChoices(s.ctx, class)
	
	// Assert
	s.NoError(err)
	s.Len(choices, 2)
	
	// First choice - armor
	s.Equal("fighter-equip-0", choices[0].ID)
	s.Equal("Choose armor", choices[0].Name)
	s.Equal("equipment", choices[0].Type)
	s.Len(choices[0].Options, 2)
	s.Equal("chain-mail", choices[0].Options[0].Key)
	s.Equal("Chain Mail", choices[0].Options[0].Name)
	
	// Second choice - weapon with count
	s.Equal("fighter-equip-1", choices[1].ID)
	s.Equal("Choose weapon", choices[1].Name)
	s.Len(choices[1].Options, 1)
	s.Equal("shortsword", choices[1].Options[0].Key)
	s.Equal("2x Shortsword", choices[1].Options[0].Name) // Should show count
}

func (s *ChoiceResolverTestSuite) TestResolveEquipmentChoices_NoChoices() {
	// Setup
	class := &entities.Class{
		Key:  "monk",
		Name: "Monk",
		// No equipment choices
	}
	
	// Execute
	choices, err := s.resolver.ResolveEquipmentChoices(s.ctx, class)
	
	// Assert
	s.NoError(err)
	s.Empty(choices)
}

func (s *ChoiceResolverTestSuite) TestResolveEquipmentChoices_NilOptions() {
	// Setup
	class := &entities.Class{
		Key:  "wizard",
		Name: "Wizard",
		StartingEquipmentChoices: []*entities.Choice{
			{
				Name:    "Choose focus",
				Count:   1,
				Options: nil, // Nil options
			},
			{
				Name:    "Choose weapon",
				Count:   1,
				Options: []entities.Option{}, // Empty options
			},
		},
	}
	
	// Execute
	choices, err := s.resolver.ResolveEquipmentChoices(s.ctx, class)
	
	// Assert
	s.NoError(err)
	s.Empty(choices) // Should skip choices with nil or empty options
}

// ValidateProficiencySelections Tests

func (s *ChoiceResolverTestSuite) TestValidateProficiencySelections_NotImplemented() {
	// Setup
	race := &entities.Race{Key: "human", Name: "Human"}
	class := &entities.Class{Key: "fighter", Name: "Fighter"}
	selections := []string{"skill-athletics", "skill-intimidation"}
	
	// Execute
	err := s.resolver.ValidateProficiencySelections(s.ctx, race, class, selections)
	
	// Assert
	s.NoError(err) // Currently returns nil (not implemented)
}