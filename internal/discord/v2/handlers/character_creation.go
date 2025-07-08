package handlers

import (
	"fmt"

	"github.com/KirkDiggler/dnd-bot-discord/internal/discord/v2/builders"
	"github.com/KirkDiggler/dnd-bot-discord/internal/discord/v2/core"
	domainCharacter "github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
)

// CharacterCreationHandler handles the character creation flow
type CharacterCreationHandler struct {
	service         character.Service
	flowService     domainCharacter.CreationFlowService
	customIDBuilder *core.CustomIDBuilder
}

// CharacterCreationHandlerConfig holds the configuration
type CharacterCreationHandlerConfig struct {
	Service     character.Service
	FlowService domainCharacter.CreationFlowService
}

// NewCharacterCreationHandler creates a new character creation handler
func NewCharacterCreationHandler(cfg *CharacterCreationHandlerConfig) (*CharacterCreationHandler, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}
	if cfg.Service == nil {
		return nil, fmt.Errorf("service is required")
	}
	if cfg.FlowService == nil {
		return nil, fmt.Errorf("flowService is required")
	}

	return &CharacterCreationHandler{
		service:         cfg.Service,
		flowService:     cfg.FlowService,
		customIDBuilder: core.NewCustomIDBuilder("creation"),
	}, nil
}

// StartCreation handles the initial character creation command
func (h *CharacterCreationHandler) StartCreation(ctx *core.InteractionContext) (*core.HandlerResult, error) {
	// Get or create draft character
	char, err := h.service.GetOrCreateDraftCharacter(ctx.Context, ctx.UserID, ctx.GuildID)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Get current step
	currentStep, err := h.flowService.GetCurrentStep(ctx.Context, char.ID)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Build response for current step
	response, err := h.buildStepResponse(char, currentStep)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	return &core.HandlerResult{
		Response: response,
	}, nil
}

// HandleStepSelection handles when a user makes a selection in the creation flow
func (h *CharacterCreationHandler) HandleStepSelection(ctx *core.InteractionContext) (*core.HandlerResult, error) {
	// Parse custom ID
	customID, err := core.ParseCustomID(ctx.GetCustomID())
	if err != nil {
		return nil, core.NewValidationError("Invalid selection")
	}

	characterID := customID.Target

	// Get character
	char, err := h.service.GetCharacter(ctx.Context, characterID)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Verify ownership
	if char.OwnerID != ctx.UserID {
		return nil, core.NewForbiddenError("You can only edit your own characters")
	}

	// Handle navigation
	if customID.Action == "back" {
		return h.handleBack(ctx, char)
	}

	// Get selected values
	var selectedValues []string

	// First check if we have test values (for unit testing)
	if testValues, ok := ctx.GetParam("selected_values").([]string); ok {
		selectedValues = testValues
	} else if ctx.IsComponent() && ctx.Interaction != nil {
		// Otherwise get from Discord interaction
		data := ctx.Interaction.MessageComponentData()
		selectedValues = data.Values
	}

	if len(selectedValues) == 0 && customID.Action == "select" {
		return nil, core.NewValidationError("Please make a selection")
	}

	// Get current step to determine the type
	currentStep, err := h.flowService.GetCurrentStep(ctx.Context, char.ID)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Create step result with context from the step
	result := &domainCharacter.CreationStepResult{
		StepType:   currentStep.Type,
		Selections: selectedValues,
	}

	// Add context metadata if available
	if currentStep.Context != nil {
		result.Metadata = make(map[string]any)
		if source, ok := currentStep.Context["source"].(string); ok {
			result.Metadata["source"] = source
		}
	}

	// Process the step result and get the next step
	nextStep, err := h.flowService.ProcessStepResult(ctx.Context, char.ID, result)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Check if we're done
	isComplete, err := h.flowService.IsCreationComplete(ctx.Context, char.ID)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	if isComplete {
		// Get the updated character
		updatedChar, updateErr := h.service.GetCharacter(ctx.Context, char.ID)
		if updateErr != nil {
			return nil, core.NewInternalError(updateErr)
		}
		return h.completeCreation(ctx, updatedChar)
	}

	// Get updated character for display
	updatedChar, err := h.service.GetCharacter(ctx.Context, char.ID)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Build response for next step
	response, err := h.buildStepResponse(updatedChar, nextStep)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	response.AsUpdate() // Update the original message

	return &core.HandlerResult{
		Response: response,
	}, nil
}

// handleBack handles the back navigation
func (h *CharacterCreationHandler) handleBack(ctx *core.InteractionContext, char *domainCharacter.Character) (*core.HandlerResult, error) {
	// For now, just return to the current step
	// TODO: Implement proper back navigation when needed
	prevStep, err := h.flowService.GetCurrentStep(ctx.Context, char.ID)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Get updated character
	updatedChar, err := h.service.GetCharacter(ctx.Context, char.ID)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Build response for previous step
	response, err := h.buildStepResponse(updatedChar, prevStep)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	response.AsUpdate()

	return &core.HandlerResult{
		Response: response,
	}, nil
}

// buildStepResponse builds the Discord response for a creation step
func (h *CharacterCreationHandler) buildStepResponse(char *domainCharacter.Character, step *domainCharacter.CreationStep) (*core.Response, error) {
	// Build embed
	embed := builders.NewEmbed().
		Title(fmt.Sprintf("Character Creation - %s", step.Title)).
		Description(step.Description).
		Color(builders.ColorPrimary)

	// Add current progress
	if char.Name != "" {
		embed.AddField("Name", char.Name, true)
	}
	if char.Race != nil {
		embed.AddField("Race", char.Race.Name, true)
	}
	if char.Class != nil {
		embed.AddField("Class", char.Class.Name, true)
	}

	// Build components based on step type
	components := builders.NewComponentBuilder(h.customIDBuilder)

	switch step.Type {
	case domainCharacter.StepTypeRaceSelection, domainCharacter.StepTypeClassSelection,
		domainCharacter.StepTypeDivineDomainSelection, domainCharacter.StepTypeFightingStyleSelection,
		domainCharacter.StepTypeFavoredEnemySelection, domainCharacter.StepTypeNaturalExplorerSelection:
		// Single-choice selection menus
		if len(step.Options) > 0 {
			options := make([]builders.SelectOption, 0, len(step.Options))
			for _, opt := range step.Options {
				options = append(options, builders.SelectOption{
					Label:       opt.Name,
					Value:       opt.Key,
					Description: opt.Description,
				})
			}

			placeholder := "Choose an option..."
			if step.Context != nil {
				if ph, ok := step.Context["placeholder"].(string); ok {
					placeholder = ph
				}
			}

			components.SelectMenu(
				placeholder,
				fmt.Sprintf("select_%s", char.ID),
				options,
			)
		}

	case domainCharacter.StepTypeSkillSelection, domainCharacter.StepTypeLanguageSelection:
		// Multi-choice selection menus
		if len(step.Options) > 0 {
			options := make([]builders.SelectOption, 0, len(step.Options))
			for _, opt := range step.Options {
				options = append(options, builders.SelectOption{
					Label:       opt.Name,
					Value:       opt.Key,
					Description: opt.Description,
				})
			}

			placeholder := fmt.Sprintf("Select %d options...", step.MinChoices)
			if step.Context != nil {
				if ph, ok := step.Context["placeholder"].(string); ok {
					placeholder = ph
				}
			}

			components.SelectMenuWithOptions(
				placeholder,
				fmt.Sprintf("select_%s", char.ID),
				options,
				step.MinChoices,
				step.MaxChoices,
			)
		}

	case domainCharacter.StepTypeAbilityScores:
		// For ability scores, add a roll button
		components.PrimaryButton("Roll Ability Scores", "roll", char.ID)

	case domainCharacter.StepTypeAbilityAssignment:
		// For ability assignment, add an assign button
		components.PrimaryButton("Assign Abilities", "assign", char.ID)

	case domainCharacter.StepTypeProficiencySelection:
		// For proficiency selection, add a select button
		components.PrimaryButton("Choose Proficiencies", "proficiencies", char.ID)

	case domainCharacter.StepTypeEquipmentSelection:
		// For equipment selection, add a select button
		components.PrimaryButton("Choose Equipment", "equipment", char.ID)

	case domainCharacter.StepTypeCharacterDetails:
		// For final details, add a name button
		components.PrimaryButton("Set Character Name", "name", char.ID)

	default:
		// For other steps, build based on options
		if len(step.Options) > 0 {
			for i, opt := range step.Options {
				if i > 0 && i%5 == 0 {
					components.NewRow() // Discord limit of 5 buttons per row
				}
				components.PrimaryButton(opt.Name, fmt.Sprintf("option_%s", opt.Key), char.ID)
			}
		}
	}

	// Add navigation buttons
	// TODO: Add back button when CreationState is available
	// components.NewRow()
	// components.SecondaryButton("⬅️ Back", "back", char.ID)

	response := core.NewResponse("").
		WithEmbeds(embed.Build()).
		WithComponents(components.Build()...).
		AsEphemeral()

	return response, nil
}

// completeCreation finalizes the character creation
func (h *CharacterCreationHandler) completeCreation(ctx *core.InteractionContext, char *domainCharacter.Character) (*core.HandlerResult, error) {
	// Finalize the character
	finalChar, err := h.service.FinalizeDraftCharacter(ctx.Context, char.ID)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Build success embed
	embed := builders.SuccessEmbed(
		"Character Created!",
		fmt.Sprintf("Your character **%s** has been created successfully!", finalChar.Name),
	).
		AddField("Race", finalChar.Race.Name, true).
		AddField("Class", finalChar.Class.Name, true).
		AddField("Level", fmt.Sprintf("%d", finalChar.Level), true).
		Build()

	response := core.NewResponse("").
		WithEmbeds(embed).
		AsUpdate()

	return &core.HandlerResult{
		Response: response,
	}, nil
}
