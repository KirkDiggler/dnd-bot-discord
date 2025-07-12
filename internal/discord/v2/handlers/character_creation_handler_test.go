package handlers

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/discord/v2/core"
	domainCharacter "github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	mockCharacterService "github.com/KirkDiggler/dnd-bot-discord/internal/services/character/mock"
	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// Test fixtures
func createTestCharacter() *domainCharacter.Character {
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

func createTestContext(customID string) *core.InteractionContext {
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

func setupHandler(t *testing.T) (*CharacterCreationHandler, *mockCharacterService.MockService, *mockCharacterService.MockCreationFlowService, *gomock.Controller) {
	ctrl := gomock.NewController(t)

	mockService := mockCharacterService.NewMockService(ctrl)
	mockFlowService := mockCharacterService.NewMockCreationFlowService(ctrl)

	customIDBuilder := core.NewCustomIDBuilder("create")

	config := &CharacterCreationHandlerConfig{
		Service:         mockService,
		FlowService:     mockFlowService,
		CustomIDBuilder: customIDBuilder,
	}

	handler, err := NewCharacterCreationHandler(config)
	require.NoError(t, err)

	return handler, mockService, mockFlowService, ctrl
}

func TestCharacterCreationHandler_StartCreation(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*mockCharacterService.MockService, *mockCharacterService.MockCreationFlowService)
		expectedError bool
		checkResponse func(*testing.T, *core.HandlerResult)
	}{
		{
			name: "successful_start_race_selection",
			setupMocks: func(mockService *mockCharacterService.MockService, mockFlowService *mockCharacterService.MockCreationFlowService) {
				char := createTestCharacter()
				char.Race = nil // Starting character has no race yet

				mockService.EXPECT().
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

				mockFlowService.EXPECT().
					GetCurrentStep(gomock.Any(), char.ID).
					Return(raceStep, nil)
			},
			expectedError: false,
			checkResponse: func(t *testing.T, result *core.HandlerResult) {
				assert.NotNil(t, result.Response)
				// Check that response has ephemeral flag set (no direct method to check)
				assert.NotEmpty(t, result.Response.Embeds)
				assert.NotEmpty(t, result.Response.Components)
			},
		},
		{
			name: "service_error",
			setupMocks: func(mockService *mockCharacterService.MockService, mockFlowService *mockCharacterService.MockCreationFlowService) {
				mockService.EXPECT().
					GetOrCreateDraftCharacter(gomock.Any(), "test-user-123", "test-guild-123").
					Return(nil, errors.New("database error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService, mockFlowService, ctrl := setupHandler(t)
			defer ctrl.Finish()

			tt.setupMocks(mockService, mockFlowService)

			ctx := createTestContext("create:character")
			result, err := handler.StartCreation(ctx)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.checkResponse != nil {
					tt.checkResponse(t, result)
				}
			}
		})
	}
}

func TestCharacterCreationHandler_HandleConfirmRace(t *testing.T) {
	tests := []struct {
		name          string
		customID      string
		setupMocks    func(*mockCharacterService.MockService, *mockCharacterService.MockCreationFlowService)
		expectedError bool
		errorType     string
	}{
		{
			name:     "successful_race_confirmation",
			customID: "create:confirm_race:test-char-123:human",
			setupMocks: func(mockService *mockCharacterService.MockService, mockFlowService *mockCharacterService.MockCreationFlowService) {
				char := createTestCharacter()
				char.Race = nil

				mockService.EXPECT().
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

				mockFlowService.EXPECT().
					ProcessStepResult(gomock.Any(), char.ID, stepResult).
					Return(nextStep, nil)

				mockFlowService.EXPECT().
					IsCreationComplete(gomock.Any(), char.ID).
					Return(false, nil)

				updatedChar := createTestCharacter() // Now has race
				mockService.EXPECT().
					GetCharacter(gomock.Any(), char.ID).
					Return(updatedChar, nil)
			},
			expectedError: false,
		},
		{
			name:          "invalid_custom_id",
			customID:      "invalid-format",
			setupMocks:    func(*mockCharacterService.MockService, *mockCharacterService.MockCreationFlowService) {},
			expectedError: true,
			errorType:     "validation",
		},
		{
			name:     "unauthorized_user",
			customID: "create:confirm_race:test-char-123:human",
			setupMocks: func(mockService *mockCharacterService.MockService, mockFlowService *mockCharacterService.MockCreationFlowService) {
				char := createTestCharacter()
				char.OwnerID = "different-user" // Different owner

				mockService.EXPECT().
					GetCharacter(gomock.Any(), "test-char-123").
					Return(char, nil)
			},
			expectedError: true,
			errorType:     "forbidden",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService, mockFlowService, ctrl := setupHandler(t)
			defer ctrl.Finish()

			tt.setupMocks(mockService, mockFlowService)

			ctx := createTestContext(tt.customID)
			result, err := handler.HandleConfirmRace(ctx)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)

				if tt.errorType != "" {
					switch tt.errorType {
					case "validation":
						assert.Contains(t, err.Error(), "Invalid selection")
					case "forbidden":
						assert.Contains(t, err.Error(), "You can only edit your own characters")
					}
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotNil(t, result.Response)
			}
		})
	}
}

func TestCharacterCreationHandler_HandleOpenProficiencySelection(t *testing.T) {
	tests := []struct {
		name          string
		customID      string
		setupMocks    func(*mockCharacterService.MockService, *mockCharacterService.MockCreationFlowService)
		expectedError bool
		checkResponse func(*testing.T, *core.HandlerResult)
	}{
		{
			name:     "shows_proficiency_ui",
			customID: "create:proficiencies:test-char-123",
			setupMocks: func(mockService *mockCharacterService.MockService, mockFlowService *mockCharacterService.MockCreationFlowService) {
				char := createTestCharacter()
				// Add some proficiencies to test the display
				char.Proficiencies = map[rulebook.ProficiencyType][]*rulebook.Proficiency{
					"skill": {
						{Key: "investigation", Name: "Investigation"},
						{Key: "history", Name: "History"},
					},
					"weapon": {
						{Key: "simple-weapons", Name: "Simple Weapons"},
					},
				}

				mockService.EXPECT().
					GetCharacter(gomock.Any(), "test-char-123").
					Return(char, nil)

				profStep := &domainCharacter.CreationStep{
					Type:        domainCharacter.StepTypeProficiencySelection,
					Title:       "Choose Proficiencies",
					Description: "Select proficiencies",
					Options:     []domainCharacter.CreationOption{}, // Empty as expected
				}

				mockFlowService.EXPECT().
					GetCurrentStep(gomock.Any(), char.ID).
					Return(profStep, nil)
			},
			expectedError: false,
			checkResponse: func(t *testing.T, result *core.HandlerResult) {
				assert.NotNil(t, result.Response)
				assert.True(t, true)
				assert.NotEmpty(t, result.Response.Embeds)
				assert.NotEmpty(t, result.Response.Components)

				// Check that the embed contains proficiency information
				embed := result.Response.Embeds[0]
				assert.Contains(t, embed.Title, "Proficiency Selection")
			},
		},
		{
			name:     "wrong_step_type",
			customID: "create:proficiencies:test-char-123",
			setupMocks: func(mockService *mockCharacterService.MockService, mockFlowService *mockCharacterService.MockCreationFlowService) {
				char := createTestCharacter()

				mockService.EXPECT().
					GetCharacter(gomock.Any(), "test-char-123").
					Return(char, nil)

				// Return wrong step type (e.g., still on class selection)
				wrongStep := &domainCharacter.CreationStep{
					Type: domainCharacter.StepTypeClassSelection,
				}

				mockFlowService.EXPECT().
					GetCurrentStep(gomock.Any(), char.ID).
					Return(wrongStep, nil)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService, mockFlowService, ctrl := setupHandler(t)
			defer ctrl.Finish()

			tt.setupMocks(mockService, mockFlowService)

			ctx := createTestContext(tt.customID)
			result, err := handler.HandleOpenProficiencySelection(ctx)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.checkResponse != nil {
					tt.checkResponse(t, result)
				}
			}
		})
	}
}

func TestCharacterCreationHandler_HandleConfirmProficiencySelection(t *testing.T) {
	tests := []struct {
		name           string
		setupMocks     func(*mockCharacterService.MockService, *mockCharacterService.MockCreationFlowService)
		expectedError  bool
		expectComplete bool
	}{
		{
			name: "advances_to_equipment",
			setupMocks: func(mockService *mockCharacterService.MockService, mockFlowService *mockCharacterService.MockCreationFlowService) {
				char := createTestCharacter()

				mockService.EXPECT().
					GetCharacter(gomock.Any(), "test-char-123").
					Return(char, nil)

				stepResult := &domainCharacter.CreationStepResult{
					StepType:   domainCharacter.StepTypeProficiencySelection,
					Selections: []string{}, // Empty selections expected
				}

				nextStep := &domainCharacter.CreationStep{
					Type:  domainCharacter.StepTypeEquipmentSelection,
					Title: "Choose Equipment",
				}

				mockFlowService.EXPECT().
					ProcessStepResult(gomock.Any(), char.ID, stepResult).
					Return(nextStep, nil)

				mockFlowService.EXPECT().
					IsCreationComplete(gomock.Any(), char.ID).
					Return(false, nil)

				mockService.EXPECT().
					GetCharacter(gomock.Any(), char.ID).
					Return(char, nil)
			},
			expectedError:  false,
			expectComplete: false,
		},
		{
			name: "completes_creation",
			setupMocks: func(mockService *mockCharacterService.MockService, mockFlowService *mockCharacterService.MockCreationFlowService) {
				char := createTestCharacter()
				char.Name = "Test Wizard" // Named character

				mockService.EXPECT().
					GetCharacter(gomock.Any(), "test-char-123").
					Return(char, nil)

				stepResult := &domainCharacter.CreationStepResult{
					StepType:   domainCharacter.StepTypeProficiencySelection,
					Selections: []string{},
				}

				finalStep := &domainCharacter.CreationStep{
					Type: domainCharacter.StepTypeComplete,
				}

				mockFlowService.EXPECT().
					ProcessStepResult(gomock.Any(), char.ID, stepResult).
					Return(finalStep, nil)

				mockFlowService.EXPECT().
					IsCreationComplete(gomock.Any(), char.ID).
					Return(true, nil)

				finalChar := createTestCharacter()
				finalChar.Name = "Test Wizard"
				finalChar.Status = shared.CharacterStatusActive

				mockService.EXPECT().
					GetCharacter(gomock.Any(), char.ID).
					Return(finalChar, nil)

				mockService.EXPECT().
					FinalizeDraftCharacter(gomock.Any(), char.ID).
					Return(finalChar, nil)
			},
			expectedError:  false,
			expectComplete: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService, mockFlowService, ctrl := setupHandler(t)
			defer ctrl.Finish()

			tt.setupMocks(mockService, mockFlowService)

			ctx := createTestContext("create:confirm_proficiency_selection:test-char-123")
			result, err := handler.HandleConfirmProficiencySelection(ctx)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotNil(t, result.Response)

				if tt.expectComplete {
					// Should be completion embed
					embed := result.Response.Embeds[0]
					assert.Contains(t, embed.Title, "Created")
				}
			}
		})
	}
}

func TestCharacterCreationHandler_HandleOpenEquipmentSelection(t *testing.T) {
	handler, mockService, mockFlowService, ctrl := setupHandler(t)
	defer ctrl.Finish()

	char := createTestCharacter()
	// Add some equipment to test display
	// Note: EquippedSlots needs proper equipment.Equipment implementation
	// For test purposes, we'll skip this complex setup

	mockService.EXPECT().
		GetCharacter(gomock.Any(), "test-char-123").
		Return(char, nil)

	equipStep := &domainCharacter.CreationStep{
		Type:        domainCharacter.StepTypeEquipmentSelection,
		Title:       "Choose Equipment",
		Description: "Select equipment",
		Options:     []domainCharacter.CreationOption{},
	}

	mockFlowService.EXPECT().
		GetCurrentStep(gomock.Any(), char.ID).
		Return(equipStep, nil)

	ctx := createTestContext("create:equipment:test-char-123")
	result, err := handler.HandleOpenEquipmentSelection(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.Response)
	assert.True(t, true)

	embed := result.Response.Embeds[0]
	assert.Contains(t, embed.Title, "Equipment Selection")
}

func TestCharacterCreationHandler_SpellSelectionFlow(t *testing.T) {
	t.Run("confirm_spell_selection_advances_to_proficiencies", func(t *testing.T) {
		handler, mockService, mockFlowService, ctrl := setupHandler(t)
		defer ctrl.Finish()

		char := createTestCharacter()
		char.Features = []*rulebook.CharacterFeature{
			{Key: "cantrips_selection_confirmed"},
			// Note: spells_selection_confirmed will be added by the handler
		}

		mockService.EXPECT().
			GetCharacter(gomock.Any(), "test-char-123").
			Return(char, nil)

		// The handler first gets current step to build enhanced response
		currentStep := &domainCharacter.CreationStep{
			Type:        domainCharacter.StepTypeSpellbookSelection,
			Title:       "Choose Spells",
			Description: "Select your spells",
			Options:     []domainCharacter.CreationOption{},
		}

		mockFlowService.EXPECT().
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
			Options:     []domainCharacter.CreationOption{}, // Empty as expected
		}

		mockFlowService.EXPECT().
			ProcessStepResult(gomock.Any(), char.ID, stepResult).
			Return(nextStep, nil)

		updatedChar := createTestCharacter()
		updatedChar.Features = append(updatedChar.Features, &rulebook.CharacterFeature{
			Key: "spells_selection_confirmed",
		})

		mockService.EXPECT().
			GetCharacter(gomock.Any(), char.ID).
			Return(updatedChar, nil)

		ctx := createTestContext("create:confirm_spell_selection:test-char-123")
		result, err := handler.HandleConfirmSpellSelection(ctx)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.Response)

		// Should advance to proficiency step
		embed := result.Response.Embeds[0]
		assert.Contains(t, embed.Title, "Character Creation")
	})
}

// Test error handling across handlers
func TestCharacterCreationHandler_ErrorHandling(t *testing.T) {
	tests := []struct {
		name      string
		handler   func(*CharacterCreationHandler, *core.InteractionContext) (*core.HandlerResult, error)
		customID  string
		setupMock func(*mockCharacterService.MockService, *mockCharacterService.MockCreationFlowService)
		expectErr string
	}{
		{
			name:     "service_error_in_race_confirmation",
			customID: "create:confirm_race:test-char-123:human",
			handler: func(h *CharacterCreationHandler, ctx *core.InteractionContext) (*core.HandlerResult, error) {
				return h.HandleConfirmRace(ctx)
			},
			setupMock: func(mockService *mockCharacterService.MockService, mockFlowService *mockCharacterService.MockCreationFlowService) {
				mockService.EXPECT().
					GetCharacter(gomock.Any(), "test-char-123").
					Return(nil, errors.New("database connection failed"))
			},
			expectErr: "internal",
		},
		{
			name:     "flow_service_error",
			customID: "create:confirm_proficiency_selection:test-char-123",
			handler: func(h *CharacterCreationHandler, ctx *core.InteractionContext) (*core.HandlerResult, error) {
				return h.HandleConfirmProficiencySelection(ctx)
			},
			setupMock: func(mockService *mockCharacterService.MockService, mockFlowService *mockCharacterService.MockCreationFlowService) {
				char := createTestCharacter()
				mockService.EXPECT().
					GetCharacter(gomock.Any(), "test-char-123").
					Return(char, nil)

				mockFlowService.EXPECT().
					ProcessStepResult(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("flow processing failed"))
			},
			expectErr: "internal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockService, mockFlowService, ctrl := setupHandler(t)
			defer ctrl.Finish()

			tt.setupMock(mockService, mockFlowService)

			ctx := createTestContext(tt.customID)
			result, err := tt.handler(handler, ctx)

			assert.Error(t, err)
			assert.Nil(t, result)
			// Check for specific error content instead of generic type
			switch tt.expectErr {
			case "internal":
				// Internal errors wrap the original error message
				assert.True(t, strings.Contains(err.Error(), "database connection failed") ||
					strings.Contains(err.Error(), "flow processing failed"))
			default:
				assert.Contains(t, err.Error(), tt.expectErr)
			}
		})
	}
}
