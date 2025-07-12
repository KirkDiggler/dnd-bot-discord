package handlers

import (
	"context"
	"errors"
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/discord/v2/core"
	domainCharacter "github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	mockCharacterService "github.com/KirkDiggler/dnd-bot-discord/internal/services/character/mock"
	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

// CharacterCreationResponseTestSuite tests that all character creation handlers
// properly use ephemeral messages and update existing messages
type CharacterCreationResponseTestSuite struct {
	suite.Suite
	ctrl            *gomock.Controller
	handler         *CharacterCreationHandler
	mockService     *mockCharacterService.MockService
	mockFlowService *mockCharacterService.MockCreationFlowService
}

func (s *CharacterCreationResponseTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockService = mockCharacterService.NewMockService(s.ctrl)
	s.mockFlowService = mockCharacterService.NewMockCreationFlowService(s.ctrl)

	customIDBuilder := core.NewCustomIDBuilder("create")

	config := &CharacterCreationHandlerConfig{
		Service:         s.mockService,
		FlowService:     s.mockFlowService,
		CustomIDBuilder: customIDBuilder,
	}

	var err error
	s.handler, err = NewCharacterCreationHandler(config)
	s.Require().NoError(err)
}

func (s *CharacterCreationResponseTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *CharacterCreationResponseTestSuite) createTestCharacter() *domainCharacter.Character {
	return &domainCharacter.Character{
		ID:      "test-char-123",
		OwnerID: "test-user-123",
		RealmID: "test-realm",
		Status:  shared.CharacterStatusDraft,
		Race:    &rulebook.Race{Key: "human", Name: "Human"},
		Class:   &rulebook.Class{Key: "wizard", Name: "Wizard"},
		Attributes: map[shared.Attribute]*domainCharacter.AbilityScore{
			shared.AttributeStrength:     {Score: 10},
			shared.AttributeDexterity:    {Score: 14},
			shared.AttributeConstitution: {Score: 13},
			shared.AttributeIntelligence: {Score: 15},
			shared.AttributeWisdom:       {Score: 12},
			shared.AttributeCharisma:     {Score: 8},
		},
		Features: []*rulebook.CharacterFeature{},
		Spells: &domainCharacter.SpellList{
			Cantrips:    []string{"fire-bolt", "mage-hand", "prestidigitation"},
			KnownSpells: []string{"magic-missile", "shield"},
		},
	}
}

func (s *CharacterCreationResponseTestSuite) createTestContext(customID string) *core.InteractionContext {
	return &core.InteractionContext{
		Context: context.Background(),
		UserID:  "test-user-123",
		GuildID: "test-guild-123",
		Interaction: &discordgo.InteractionCreate{
			Interaction: &discordgo.Interaction{
				Type: discordgo.InteractionMessageComponent,
				Data: discordgo.MessageComponentInteractionData{
					CustomID: customID,
				},
			},
		},
	}
}

// Test that StartCreation uses ephemeral response
func (s *CharacterCreationResponseTestSuite) TestStartCreation_UsesEphemeralResponse() {
	s.Run("initial_creation_is_ephemeral", func() {
		char := s.createTestCharacter()
		char.Race = nil // Starting character has no race yet

		s.mockService.EXPECT().
			GetOrCreateDraftCharacter(gomock.Any(), "test-user-123", "test-guild-123").
			Return(char, nil)

		raceStep := &domainCharacter.CreationStep{
			Type:        domainCharacter.StepTypeRaceSelection,
			Title:       "Choose Race",
			Description: "Select your character's race",
			Options: []domainCharacter.CreationOption{
				{Key: "human", Name: "Human", Description: "Versatile humans"},
				{Key: "elf", Name: "Elf", Description: "Graceful elves"},
			},
		}

		s.mockFlowService.EXPECT().
			GetCurrentStep(gomock.Any(), char.ID).
			Return(raceStep, nil)

		ctx := s.createTestContext("create:character")
		result, err := s.handler.StartCreation(ctx)

		s.Require().NoError(err)
		s.Require().NotNil(result)
		s.Require().NotNil(result.Response)

		// Check that response is ephemeral
		s.Assert().True(result.Response.Ephemeral, "StartCreation should use ephemeral response")
		s.Assert().False(result.Response.Update, "StartCreation should not use update flag")
	})
}

// Test that all step selection handlers use AsUpdate
func (s *CharacterCreationResponseTestSuite) TestHandleStepSelection_UsesUpdate() {
	testCases := []struct {
		name       string
		customID   string
		stepType   domainCharacter.CreationStepType
		selections []string
		setupMocks func()
		isComplete bool
	}{
		{
			name:       "race_selection_uses_update",
			customID:   "create:select:test-char-123",
			stepType:   domainCharacter.StepTypeRaceSelection,
			selections: []string{"human"},
			setupMocks: func() {
				char := s.createTestCharacter()
				char.Race = nil

				s.mockService.EXPECT().
					GetCharacter(gomock.Any(), "test-char-123").
					Return(char, nil)

				currentStep := &domainCharacter.CreationStep{
					Type: domainCharacter.StepTypeRaceSelection,
				}

				s.mockFlowService.EXPECT().
					GetCurrentStep(gomock.Any(), char.ID).
					Return(currentStep, nil)

				nextStep := &domainCharacter.CreationStep{
					Type:  domainCharacter.StepTypeClassSelection,
					Title: "Choose Class",
				}

				s.mockFlowService.EXPECT().
					ProcessStepResult(gomock.Any(), char.ID, gomock.Any()).
					Return(nextStep, nil)

				s.mockFlowService.EXPECT().
					IsCreationComplete(gomock.Any(), char.ID).
					Return(false, nil)

				updatedChar := s.createTestCharacter()
				s.mockService.EXPECT().
					GetCharacter(gomock.Any(), char.ID).
					Return(updatedChar, nil)
			},
			isComplete: false,
		},
		{
			name:       "class_selection_uses_update",
			customID:   "create:select:test-char-123",
			stepType:   domainCharacter.StepTypeClassSelection,
			selections: []string{"wizard"},
			setupMocks: func() {
				char := s.createTestCharacter()

				s.mockService.EXPECT().
					GetCharacter(gomock.Any(), "test-char-123").
					Return(char, nil)

				currentStep := &domainCharacter.CreationStep{
					Type: domainCharacter.StepTypeClassSelection,
				}

				s.mockFlowService.EXPECT().
					GetCurrentStep(gomock.Any(), char.ID).
					Return(currentStep, nil)

				nextStep := &domainCharacter.CreationStep{
					Type:  domainCharacter.StepTypeAbilityAssignment,
					Title: "Assign Abilities",
				}

				s.mockFlowService.EXPECT().
					ProcessStepResult(gomock.Any(), char.ID, gomock.Any()).
					Return(nextStep, nil)

				s.mockFlowService.EXPECT().
					IsCreationComplete(gomock.Any(), char.ID).
					Return(false, nil)

				s.mockService.EXPECT().
					GetCharacter(gomock.Any(), char.ID).
					Return(char, nil)
			},
			isComplete: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.setupMocks()

			ctx := s.createTestContext(tc.customID)
			// Set selected values in the interaction data
			if ctx.Interaction != nil {
				ctx.Interaction.Data = discordgo.MessageComponentInteractionData{
					CustomID: tc.customID,
					Values:   tc.selections,
				}
			}

			result, err := s.handler.HandleStepSelection(ctx)

			s.Require().NoError(err)
			s.Require().NotNil(result)
			s.Require().NotNil(result.Response)

			// Check that response uses update
			s.Assert().True(result.Response.Update, "HandleStepSelection should use update flag")
			s.Assert().False(result.Response.Ephemeral, "HandleStepSelection should not be ephemeral when updating")
		})
	}
}

// Test specific handlers for race confirmation
func (s *CharacterCreationResponseTestSuite) TestHandleConfirmRace_UsesUpdate() {
	s.Run("confirm_race_uses_update", func() {
		char := s.createTestCharacter()
		char.Race = nil

		s.mockService.EXPECT().
			GetCharacter(gomock.Any(), "test-char-123").
			Return(char, nil)

		stepResult := &domainCharacter.CreationStepResult{
			StepType:   domainCharacter.StepTypeRaceSelection,
			Selections: []string{"human"},
		}

		nextStep := &domainCharacter.CreationStep{
			Type:  domainCharacter.StepTypeClassSelection,
			Title: "Choose Class",
		}

		s.mockFlowService.EXPECT().
			ProcessStepResult(gomock.Any(), char.ID, stepResult).
			Return(nextStep, nil)

		s.mockFlowService.EXPECT().
			IsCreationComplete(gomock.Any(), char.ID).
			Return(false, nil)

		updatedChar := s.createTestCharacter()
		s.mockService.EXPECT().
			GetCharacter(gomock.Any(), char.ID).
			Return(updatedChar, nil)

		ctx := s.createTestContext("create:confirm_race:test-char-123:human")
		result, err := s.handler.HandleConfirmRace(ctx)

		s.Require().NoError(err)
		s.Require().NotNil(result)
		s.Require().NotNil(result.Response)

		// Check that response uses update
		s.Assert().True(result.Response.Update, "HandleConfirmRace should use update flag")
		s.Assert().False(result.Response.Ephemeral, "HandleConfirmRace should not be ephemeral when updating")
	})
}

// Test specific handlers for class confirmation
func (s *CharacterCreationResponseTestSuite) TestHandleConfirmClass_UsesUpdate() {
	s.Run("confirm_class_uses_update", func() {
		char := s.createTestCharacter()

		s.mockService.EXPECT().
			GetCharacter(gomock.Any(), "test-char-123").
			Return(char, nil)

		result := &domainCharacter.CreationStepResult{
			StepType:   domainCharacter.StepTypeClassSelection,
			Selections: []string{"wizard"},
		}

		nextStep := &domainCharacter.CreationStep{
			Type:  domainCharacter.StepTypeAbilityAssignment,
			Title: "Assign Abilities",
		}

		s.mockFlowService.EXPECT().
			ProcessStepResult(gomock.Any(), char.ID, result).
			Return(nextStep, nil)

		s.mockFlowService.EXPECT().
			IsCreationComplete(gomock.Any(), char.ID).
			Return(false, nil)

		s.mockService.EXPECT().
			GetCharacter(gomock.Any(), char.ID).
			Return(char, nil)

		ctx := s.createTestContext("create:confirm_class:test-char-123:wizard")
		handlerResult, err := s.handler.HandleConfirmClass(ctx)

		s.Require().NoError(err)
		s.Require().NotNil(handlerResult)
		s.Require().NotNil(handlerResult.Response)

		// Check that response uses update
		s.Assert().True(handlerResult.Response.Update, "HandleConfirmClass should use update flag")
		s.Assert().False(handlerResult.Response.Ephemeral, "HandleConfirmClass should not be ephemeral when updating")
	})
}

// Test proficiency selection handlers
func (s *CharacterCreationResponseTestSuite) TestHandleOpenProficiencySelection_UsesUpdate() {
	s.Run("open_proficiency_selection_uses_update", func() {
		char := s.createTestCharacter()
		char.Proficiencies = map[rulebook.ProficiencyType][]*rulebook.Proficiency{
			"skill": {
				{Key: "investigation", Name: "Investigation"},
				{Key: "history", Name: "History"},
			},
		}

		s.mockService.EXPECT().
			GetCharacter(gomock.Any(), "test-char-123").
			Return(char, nil)

		profStep := &domainCharacter.CreationStep{
			Type:        domainCharacter.StepTypeProficiencySelection,
			Title:       "Choose Proficiencies",
			Description: "Select proficiencies",
			Options:     []domainCharacter.CreationOption{},
		}

		s.mockFlowService.EXPECT().
			GetCurrentStep(gomock.Any(), char.ID).
			Return(profStep, nil)

		ctx := s.createTestContext("create:proficiencies:test-char-123")
		result, err := s.handler.HandleOpenProficiencySelection(ctx)

		s.Require().NoError(err)
		s.Require().NotNil(result)
		s.Require().NotNil(result.Response)

		// Check that response uses update
		s.Assert().True(result.Response.Update, "HandleOpenProficiencySelection should use update flag")
		// Should NOT be ephemeral as it's part of the flow
		s.Assert().False(result.Response.Ephemeral, "HandleOpenProficiencySelection should not be ephemeral")
	})
}

// Test proficiency confirmation
func (s *CharacterCreationResponseTestSuite) TestHandleConfirmProficiencySelection_UsesUpdate() {
	s.Run("confirm_proficiency_uses_update", func() {
		char := s.createTestCharacter()

		s.mockService.EXPECT().
			GetCharacter(gomock.Any(), "test-char-123").
			Return(char, nil)

		stepResult := &domainCharacter.CreationStepResult{
			StepType:   domainCharacter.StepTypeProficiencySelection,
			Selections: []string{},
		}

		nextStep := &domainCharacter.CreationStep{
			Type:  domainCharacter.StepTypeEquipmentSelection,
			Title: "Choose Equipment",
		}

		s.mockFlowService.EXPECT().
			ProcessStepResult(gomock.Any(), char.ID, stepResult).
			Return(nextStep, nil)

		s.mockFlowService.EXPECT().
			IsCreationComplete(gomock.Any(), char.ID).
			Return(false, nil)

		s.mockService.EXPECT().
			GetCharacter(gomock.Any(), char.ID).
			Return(char, nil)

		ctx := s.createTestContext("create:confirm_proficiency_selection:test-char-123")
		result, err := s.handler.HandleConfirmProficiencySelection(ctx)

		s.Require().NoError(err)
		s.Require().NotNil(result)
		s.Require().NotNil(result.Response)

		// Check that response uses update
		s.Assert().True(result.Response.Update, "HandleConfirmProficiencySelection should use update flag")
		s.Assert().False(result.Response.Ephemeral, "HandleConfirmProficiencySelection should not be ephemeral when updating")
	})
}

// Test equipment selection handlers
func (s *CharacterCreationResponseTestSuite) TestHandleOpenEquipmentSelection_UsesUpdate() {
	s.Run("open_equipment_uses_update", func() {
		char := s.createTestCharacter()

		s.mockService.EXPECT().
			GetCharacter(gomock.Any(), "test-char-123").
			Return(char, nil)

		equipStep := &domainCharacter.CreationStep{
			Type:        domainCharacter.StepTypeEquipmentSelection,
			Title:       "Choose Equipment",
			Description: "Select equipment",
			Options:     []domainCharacter.CreationOption{},
		}

		s.mockFlowService.EXPECT().
			GetCurrentStep(gomock.Any(), char.ID).
			Return(equipStep, nil)

		ctx := s.createTestContext("create:equipment:test-char-123")
		result, err := s.handler.HandleOpenEquipmentSelection(ctx)

		s.Require().NoError(err)
		s.Require().NotNil(result)
		s.Require().NotNil(result.Response)

		// Check that response uses update
		s.Assert().True(result.Response.Update, "HandleOpenEquipmentSelection should use update flag")
		// Should NOT be ephemeral as it's part of the flow
		s.Assert().False(result.Response.Ephemeral, "HandleOpenEquipmentSelection should not be ephemeral")
	})
}

// Test spell selection handlers
func (s *CharacterCreationResponseTestSuite) TestHandleConfirmSpellSelection_UsesUpdate() {
	s.Run("confirm_spell_selection_uses_update", func() {
		char := s.createTestCharacter()
		char.Features = []*rulebook.CharacterFeature{
			{Key: "cantrips_selection_confirmed"},
		}

		s.mockService.EXPECT().
			GetCharacter(gomock.Any(), "test-char-123").
			Return(char, nil)

		currentStep := &domainCharacter.CreationStep{
			Type:        domainCharacter.StepTypeSpellbookSelection,
			Title:       "Choose Spells",
			Description: "Select your spells",
			Options:     []domainCharacter.CreationOption{},
		}

		s.mockFlowService.EXPECT().
			GetCurrentStep(gomock.Any(), char.ID).
			Return(currentStep, nil)

		stepResult := &domainCharacter.CreationStepResult{
			StepType:   domainCharacter.StepTypeSpellbookSelection,
			Selections: char.Spells.KnownSpells,
		}

		nextStep := &domainCharacter.CreationStep{
			Type:        domainCharacter.StepTypeProficiencySelection,
			Title:       "Choose Proficiencies",
			Description: "Select proficiencies",
			Options:     []domainCharacter.CreationOption{},
		}

		s.mockFlowService.EXPECT().
			ProcessStepResult(gomock.Any(), char.ID, stepResult).
			Return(nextStep, nil)

		updatedChar := s.createTestCharacter()
		updatedChar.Features = append(updatedChar.Features, &rulebook.CharacterFeature{
			Key: "spells_selection_confirmed",
		})

		s.mockService.EXPECT().
			GetCharacter(gomock.Any(), char.ID).
			Return(updatedChar, nil)

		ctx := s.createTestContext("create:confirm_spell_selection:test-char-123")
		result, err := s.handler.HandleConfirmSpellSelection(ctx)

		s.Require().NoError(err)
		s.Require().NotNil(result)
		s.Require().NotNil(result.Response)

		// Check that response uses update
		s.Assert().True(result.Response.Update, "HandleConfirmSpellSelection should use update flag")
		s.Assert().False(result.Response.Ephemeral, "HandleConfirmSpellSelection should not be ephemeral when updating")
	})
}

// Test that completion also uses update
func (s *CharacterCreationResponseTestSuite) TestCompleteCreation_UsesUpdate() {
	s.T().Skip("Skipping test - mock setup needs refactoring")
	s.Run("completion_uses_update", func() {
		char := s.createTestCharacter()
		char.Name = "Test Wizard"

		s.mockService.EXPECT().
			GetCharacter(gomock.Any(), "test-char-123").
			Return(char, nil)

		stepResult := &domainCharacter.CreationStepResult{
			StepType:   domainCharacter.StepTypeCharacterDetails,
			Selections: []string{},
		}

		finalStep := &domainCharacter.CreationStep{
			Type: domainCharacter.StepTypeComplete,
		}

		s.mockFlowService.EXPECT().
			ProcessStepResult(gomock.Any(), char.ID, stepResult).
			Return(finalStep, nil)

		s.mockFlowService.EXPECT().
			IsCreationComplete(gomock.Any(), char.ID).
			Return(true, nil)

		s.mockService.EXPECT().
			GetCharacter(gomock.Any(), char.ID).
			Return(char, nil)

		finalChar := s.createTestCharacter()
		finalChar.Name = "Test Wizard"
		finalChar.Status = shared.CharacterStatusActive

		s.mockService.EXPECT().
			FinalizeDraftCharacter(gomock.Any(), char.ID).
			Return(finalChar, nil)

		ctx := s.createTestContext("create:name:test-char-123")
		// Set character name in the interaction data
		if ctx.Interaction != nil {
			ctx.Interaction.Data = discordgo.MessageComponentInteractionData{
				CustomID: "create:name:test-char-123",
				Values:   []string{"Test Wizard"},
			}
		}

		result, err := s.handler.HandleStepSelection(ctx)

		s.Require().NoError(err)
		s.Require().NotNil(result)
		s.Require().NotNil(result.Response)

		// Check that response uses update
		s.Assert().True(result.Response.Update, "Completion should use update flag")
		s.Assert().False(result.Response.Ephemeral, "Completion should not be ephemeral")
	})
}

// Test that navigation (back button) uses update
func (s *CharacterCreationResponseTestSuite) TestHandleBack_UsesUpdate() {
	s.T().Skip("Skipping test - mock setup needs refactoring")
	s.Run("back_navigation_uses_update", func() {
		char := s.createTestCharacter()

		s.mockService.EXPECT().
			GetCharacter(gomock.Any(), "test-char-123").
			Return(char, nil)

		currentStep := &domainCharacter.CreationStep{
			Type:  domainCharacter.StepTypeClassSelection,
			Title: "Choose Class",
		}

		s.mockFlowService.EXPECT().
			GetCurrentStep(gomock.Any(), char.ID).
			Return(currentStep, nil)

		ctx := s.createTestContext("create:back:test-char-123")
		result, err := s.handler.HandleStepSelection(ctx)

		s.Require().NoError(err)
		s.Require().NotNil(result)
		s.Require().NotNil(result.Response)

		// Check that response uses update
		s.Assert().True(result.Response.Update, "Back navigation should use update flag")
		s.Assert().False(result.Response.Ephemeral, "Back navigation should not be ephemeral when updating")
	})
}

// Test error cases to ensure they don't create new messages
func (s *CharacterCreationResponseTestSuite) TestErrorHandling_NoNewMessages() {
	testCases := []struct {
		name       string
		setupMocks func()
		handler    func(*core.InteractionContext) (*core.HandlerResult, error)
		customID   string
	}{
		{
			name: "service_error_returns_error_not_response",
			setupMocks: func() {
				s.mockService.EXPECT().
					GetCharacter(gomock.Any(), "test-char-123").
					Return(nil, errors.New("database error"))
			},
			handler: func(ctx *core.InteractionContext) (*core.HandlerResult, error) {
				return s.handler.HandleConfirmRace(ctx)
			},
			customID: "create:confirm_race:test-char-123:human",
		},
		{
			name: "flow_service_error_returns_error_not_response",
			setupMocks: func() {
				char := s.createTestCharacter()
				s.mockService.EXPECT().
					GetCharacter(gomock.Any(), "test-char-123").
					Return(char, nil)

				s.mockFlowService.EXPECT().
					ProcessStepResult(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("flow error"))
			},
			handler: func(ctx *core.InteractionContext) (*core.HandlerResult, error) {
				return s.handler.HandleConfirmRace(ctx)
			},
			customID: "create:confirm_race:test-char-123:human",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.setupMocks()

			ctx := s.createTestContext(tc.customID)
			result, err := tc.handler(ctx)

			// Should return error, not a result
			s.Assert().Error(err)
			s.Assert().Nil(result, "Error cases should not return a HandlerResult")
		})
	}
}

// Test that preview handlers use update
func (s *CharacterCreationResponseTestSuite) TestPreviewHandlers_UseUpdate() {
	testCases := []struct {
		name       string
		handler    string
		customID   string
		setupMocks func()
	}{
		{
			name:     "race_preview_uses_update",
			handler:  "HandleRacePreview",
			customID: "create:preview_race:test-char-123",
			setupMocks: func() {
				char := s.createTestCharacter()
				char.Race = nil

				s.mockService.EXPECT().
					GetCharacter(gomock.Any(), "test-char-123").
					Return(char, nil)

				currentStep := &domainCharacter.CreationStep{
					Type: domainCharacter.StepTypeRaceSelection,
					Options: []domainCharacter.CreationOption{
						{Key: "human", Name: "Human", Description: "Versatile"},
					},
				}

				s.mockFlowService.EXPECT().
					GetCurrentStep(gomock.Any(), char.ID).
					Return(currentStep, nil)

				previewResult := &domainCharacter.CreationStepResult{
					StepType:   domainCharacter.StepTypeRaceSelection,
					Selections: []string{"human"},
				}

				previewChar := s.createTestCharacter()
				s.mockFlowService.EXPECT().
					PreviewStepResult(gomock.Any(), char.ID, previewResult).
					Return(previewChar, nil)
			},
		},
		{
			name:     "class_preview_uses_update",
			handler:  "HandleClassPreview",
			customID: "create:preview_class:test-char-123",
			setupMocks: func() {
				char := s.createTestCharacter()

				s.mockService.EXPECT().
					GetCharacter(gomock.Any(), "test-char-123").
					Return(char, nil)

				currentStep := &domainCharacter.CreationStep{
					Type: domainCharacter.StepTypeClassSelection,
					Options: []domainCharacter.CreationOption{
						{Key: "wizard", Name: "Wizard", Description: "Scholar"},
					},
				}

				s.mockFlowService.EXPECT().
					GetCurrentStep(gomock.Any(), char.ID).
					Return(currentStep, nil)

				previewResult := &domainCharacter.CreationStepResult{
					StepType:   domainCharacter.StepTypeClassSelection,
					Selections: []string{"wizard"},
				}

				previewChar := s.createTestCharacter()
				s.mockFlowService.EXPECT().
					PreviewStepResult(gomock.Any(), char.ID, previewResult).
					Return(previewChar, nil)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.setupMocks()

			ctx := s.createTestContext(tc.customID)
			// Add selected values for preview
			ctx.Interaction.Data = discordgo.MessageComponentInteractionData{
				CustomID: tc.customID,
				Values: func() []string {
					if tc.handler == "HandleRacePreview" {
						return []string{"human"}
					}
					return []string{"wizard"}
				}(),
			}

			var result *core.HandlerResult
			var err error

			switch tc.handler {
			case "HandleRacePreview":
				result, err = s.handler.HandleRacePreview(ctx)
			case "HandleClassPreview":
				result, err = s.handler.HandleClassPreview(ctx)
			}

			s.Require().NoError(err)
			s.Require().NotNil(result)
			s.Require().NotNil(result.Response)

			// Check that response uses update
			s.Assert().True(result.Response.Update, "%s should use update flag", tc.handler)
			s.Assert().False(result.Response.Ephemeral, "%s should not be ephemeral when updating", tc.handler)
		})
	}
}

// Test overview handlers use update
func (s *CharacterCreationResponseTestSuite) TestOverviewHandlers_UseUpdate() {
	testCases := []struct {
		name     string
		handler  func(*core.InteractionContext) (*core.HandlerResult, error)
		customID string
	}{
		{
			name:     "class_overview_uses_update",
			handler:  s.handler.HandleClassOverview,
			customID: "create:class_overview:test-char-123",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			char := s.createTestCharacter()

			s.mockService.EXPECT().
				GetCharacter(gomock.Any(), "test-char-123").
				Return(char, nil)

			ctx := s.createTestContext(tc.customID)
			result, err := tc.handler(ctx)

			s.Require().NoError(err)
			s.Require().NotNil(result)
			s.Require().NotNil(result.Response)

			// Check that response uses update
			s.Assert().True(result.Response.Update, "%s should use update flag", tc.name)
			s.Assert().False(result.Response.Ephemeral, "%s should not be ephemeral when updating", tc.name)
		})
	}
}

// Test random selection handlers use update
func (s *CharacterCreationResponseTestSuite) TestRandomSelectionHandlers_UseUpdate() {
	testCases := []struct {
		name       string
		handler    func(*core.InteractionContext) (*core.HandlerResult, error)
		customID   string
		stepType   domainCharacter.CreationStepType
		setupMocks func()
	}{
		{
			name:     "random_race_uses_update",
			handler:  s.handler.HandleRandomRace,
			customID: "create:random_race:test-char-123",
			stepType: domainCharacter.StepTypeRaceSelection,
			setupMocks: func() {
				char := s.createTestCharacter()
				char.Race = nil

				s.mockService.EXPECT().
					GetCharacter(gomock.Any(), "test-char-123").
					Return(char, nil)

				currentStep := &domainCharacter.CreationStep{
					Type: domainCharacter.StepTypeRaceSelection,
					Options: []domainCharacter.CreationOption{
						{Key: "human", Name: "Human"},
						{Key: "elf", Name: "Elf"},
					},
				}

				s.mockFlowService.EXPECT().
					GetCurrentStep(gomock.Any(), char.ID).
					Return(currentStep, nil)

				nextStep := &domainCharacter.CreationStep{
					Type:  domainCharacter.StepTypeClassSelection,
					Title: "Choose Class",
				}

				s.mockFlowService.EXPECT().
					ProcessStepResult(gomock.Any(), char.ID, gomock.Any()).
					Return(nextStep, nil)

				updatedChar := s.createTestCharacter()
				s.mockService.EXPECT().
					GetCharacter(gomock.Any(), char.ID).
					Return(updatedChar, nil)
			},
		},
		{
			name:     "random_class_uses_update",
			handler:  s.handler.HandleRandomClass,
			customID: "create:random_class:test-char-123",
			stepType: domainCharacter.StepTypeClassSelection,
			setupMocks: func() {
				char := s.createTestCharacter()

				s.mockService.EXPECT().
					GetCharacter(gomock.Any(), "test-char-123").
					Return(char, nil)

				currentStep := &domainCharacter.CreationStep{
					Type: domainCharacter.StepTypeClassSelection,
					Options: []domainCharacter.CreationOption{
						{Key: "wizard", Name: "Wizard"},
						{Key: "fighter", Name: "Fighter"},
					},
				}

				s.mockFlowService.EXPECT().
					GetCurrentStep(gomock.Any(), char.ID).
					Return(currentStep, nil)

				nextStep := &domainCharacter.CreationStep{
					Type:  domainCharacter.StepTypeAbilityAssignment,
					Title: "Assign Abilities",
				}

				s.mockFlowService.EXPECT().
					ProcessStepResult(gomock.Any(), char.ID, gomock.Any()).
					Return(nextStep, nil)

				updatedChar := s.createTestCharacter()
				s.mockService.EXPECT().
					GetCharacter(gomock.Any(), char.ID).
					Return(updatedChar, nil)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.setupMocks()

			ctx := s.createTestContext(tc.customID)
			result, err := tc.handler(ctx)

			s.Require().NoError(err)
			s.Require().NotNil(result)
			s.Require().NotNil(result.Response)

			// Check that response uses update
			s.Assert().True(result.Response.Update, "%s should use update flag", tc.name)
			s.Assert().False(result.Response.Ephemeral, "%s should not be ephemeral when updating", tc.name)
		})
	}
}

func TestCharacterCreationResponseTestSuite(t *testing.T) {
	suite.Run(t, new(CharacterCreationResponseTestSuite))
}
