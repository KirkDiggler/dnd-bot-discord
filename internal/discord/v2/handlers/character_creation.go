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
	Service         character.Service
	FlowService     domainCharacter.CreationFlowService
	CustomIDBuilder *core.CustomIDBuilder
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

	// Use provided CustomIDBuilder or create default
	customIDBuilder := cfg.CustomIDBuilder
	if customIDBuilder == nil {
		customIDBuilder = core.NewCustomIDBuilder("creation")
	}

	return &CharacterCreationHandler{
		service:         cfg.Service,
		flowService:     cfg.FlowService,
		customIDBuilder: customIDBuilder,
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
	response, err := h.buildEnhancedStepResponse(char, currentStep)
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
	response, err := h.buildEnhancedStepResponse(updatedChar, nextStep)
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
	response, err := h.buildEnhancedStepResponse(updatedChar, prevStep)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	response.AsUpdate()

	return &core.HandlerResult{
		Response: response,
	}, nil
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
