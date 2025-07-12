package handlers

import (
	"fmt"

	"github.com/KirkDiggler/dnd-bot-discord/internal/discord/v2/builders"
	"github.com/KirkDiggler/dnd-bot-discord/internal/discord/v2/core"
	domainCharacter "github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	rulebook "github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
)

// StepHandlerFunc handles rendering for a specific step type
type StepHandlerFunc func(char *domainCharacter.Character, step *domainCharacter.CreationStep) (*core.Response, error)

// CharacterCreationHandler handles the character creation flow
type CharacterCreationHandler struct {
	service         character.Service
	flowService     domainCharacter.CreationFlowService
	customIDBuilder *core.CustomIDBuilder
	stepHandlers    map[domainCharacter.CreationStepType]StepHandlerFunc
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

	h := &CharacterCreationHandler{
		service:         cfg.Service,
		flowService:     cfg.FlowService,
		customIDBuilder: customIDBuilder,
		stepHandlers:    make(map[domainCharacter.CreationStepType]StepHandlerFunc),
	}

	// Register all step handlers explicitly
	h.registerStepHandlers()

	// Validate complete coverage for D&D 5e steps
	for _, stepTypeStr := range rulebook.CharacterCreationSteps {
		stepType := domainCharacter.CreationStepType(stepTypeStr)
		if _, ok := h.stepHandlers[stepType]; !ok {
			return nil, fmt.Errorf("no handler registered for step type: %s", stepType)
		}
	}

	return h, nil
}

// registerStepHandlers registers all step type handlers
func (h *CharacterCreationHandler) registerStepHandlers() {
	// Core steps
	h.stepHandlers[domainCharacter.StepTypeRaceSelection] = h.handleRaceSelection
	h.stepHandlers[domainCharacter.StepTypeClassSelection] = h.handleClassSelection
	h.stepHandlers[domainCharacter.StepTypeAbilityScores] = h.handleAbilityScores
	h.stepHandlers[domainCharacter.StepTypeAbilityAssignment] = h.handleAbilityAssignment
	h.stepHandlers[domainCharacter.StepTypeProficiencySelection] = h.handleProficiencySelection
	h.stepHandlers[domainCharacter.StepTypeEquipmentSelection] = h.handleEquipmentSelection
	h.stepHandlers[domainCharacter.StepTypeCharacterDetails] = h.handleCharacterDetails

	// Class-specific steps
	h.stepHandlers[domainCharacter.StepTypeSkillSelection] = h.handleGenericSelection
	h.stepHandlers[domainCharacter.StepTypeLanguageSelection] = h.handleGenericSelection
	h.stepHandlers[domainCharacter.StepTypeFightingStyleSelection] = h.handleGenericSelection
	h.stepHandlers[domainCharacter.StepTypeDivineDomainSelection] = h.handleGenericSelection
	h.stepHandlers[domainCharacter.StepTypeFavoredEnemySelection] = h.handleGenericSelection
	h.stepHandlers[domainCharacter.StepTypeNaturalExplorerSelection] = h.handleGenericSelection

	// Spellcaster steps
	h.stepHandlers[domainCharacter.StepTypeCantripsSelection] = h.handleSpellSelection
	h.stepHandlers[domainCharacter.StepTypeSpellSelection] = h.handleSpellSelection
	h.stepHandlers[domainCharacter.StepTypeSpellbookSelection] = h.handleSpellSelection
	h.stepHandlers[domainCharacter.StepTypeSpellsKnownSelection] = h.handleSpellSelection

	// Class specialization steps
	h.stepHandlers[domainCharacter.StepTypeExpertiseSelection] = h.handleGenericSelection
	h.stepHandlers[domainCharacter.StepTypeSubclassSelection] = h.handleGenericSelection

	// Class-specific steps
	h.stepHandlers[domainCharacter.StepTypeFightingStyleSelection] = h.handleFightingStyleSelection
	h.stepHandlers[domainCharacter.StepTypeDivineDomainSelection] = h.handleDivineDomainSelection
	h.stepHandlers[domainCharacter.StepTypeFavoredEnemySelection] = h.handleFavoredEnemySelection
	h.stepHandlers[domainCharacter.StepTypeNaturalExplorerSelection] = h.handleNaturalExplorerSelection

	// Skill/language steps
	h.stepHandlers[domainCharacter.StepTypeSkillSelection] = h.handleSkillSelection
	h.stepHandlers[domainCharacter.StepTypeLanguageSelection] = h.handleLanguageSelection

	// Spellcaster steps
	h.stepHandlers[domainCharacter.StepTypeCantripsSelection] = h.handleSpellSelection
	h.stepHandlers[domainCharacter.StepTypeSpellSelection] = h.handleSpellSelection

	// Final step
	h.stepHandlers[domainCharacter.StepTypeComplete] = h.handleComplete
}

// Step handler implementations - these define HOW each step is rendered in Discord

func (h *CharacterCreationHandler) handleRaceSelection(char *domainCharacter.Character, step *domainCharacter.CreationStep) (*core.Response, error) {
	// Use the existing buildEnhancedStepResponse for now
	// TODO: Implement clean race selection UI without UIHints
	return h.buildEnhancedStepResponse(char, step)
}

func (h *CharacterCreationHandler) handleClassSelection(char *domainCharacter.Character, step *domainCharacter.CreationStep) (*core.Response, error) {
	// TODO: Implement class selection UI
	return h.buildEnhancedStepResponse(char, step)
}

func (h *CharacterCreationHandler) handleAbilityScores(char *domainCharacter.Character, step *domainCharacter.CreationStep) (*core.Response, error) {
	// TODO: Implement ability score UI
	return nil, fmt.Errorf("ability scores UI not implemented")
}

func (h *CharacterCreationHandler) handleAbilityAssignment(char *domainCharacter.Character, step *domainCharacter.CreationStep) (*core.Response, error) {
	// TODO: Implement ability assignment UI
	return h.buildEnhancedStepResponse(char, step)
}

func (h *CharacterCreationHandler) handleProficiencySelection(char *domainCharacter.Character, step *domainCharacter.CreationStep) (*core.Response, error) {
	// TODO: Implement proficiency selection UI
	return h.buildEnhancedStepResponse(char, step)
}

func (h *CharacterCreationHandler) handleEquipmentSelection(char *domainCharacter.Character, step *domainCharacter.CreationStep) (*core.Response, error) {
	// TODO: Implement equipment selection UI
	return h.buildEnhancedStepResponse(char, step)
}

func (h *CharacterCreationHandler) handleCharacterDetails(char *domainCharacter.Character, step *domainCharacter.CreationStep) (*core.Response, error) {
	// TODO: Implement character details UI
	return h.buildEnhancedStepResponse(char, step)
}

func (h *CharacterCreationHandler) handleFightingStyleSelection(char *domainCharacter.Character, step *domainCharacter.CreationStep) (*core.Response, error) {
	// TODO: Implement fighting style selection UI
	return nil, fmt.Errorf("fighting style selection UI not implemented")
}

func (h *CharacterCreationHandler) handleDivineDomainSelection(char *domainCharacter.Character, step *domainCharacter.CreationStep) (*core.Response, error) {
	// TODO: Implement divine domain selection UI
	return nil, fmt.Errorf("divine domain selection UI not implemented")
}

func (h *CharacterCreationHandler) handleFavoredEnemySelection(char *domainCharacter.Character, step *domainCharacter.CreationStep) (*core.Response, error) {
	// TODO: Implement favored enemy selection UI
	return nil, fmt.Errorf("favored enemy selection UI not implemented")
}

func (h *CharacterCreationHandler) handleNaturalExplorerSelection(char *domainCharacter.Character, step *domainCharacter.CreationStep) (*core.Response, error) {
	// TODO: Implement natural explorer selection UI
	return nil, fmt.Errorf("natural explorer selection UI not implemented")
}

func (h *CharacterCreationHandler) handleSkillSelection(char *domainCharacter.Character, step *domainCharacter.CreationStep) (*core.Response, error) {
	// TODO: Implement skill selection UI
	return nil, fmt.Errorf("skill selection UI not implemented")
}

func (h *CharacterCreationHandler) handleLanguageSelection(char *domainCharacter.Character, step *domainCharacter.CreationStep) (*core.Response, error) {
	// TODO: Implement language selection UI
	return nil, fmt.Errorf("language selection UI not implemented")
}

func (h *CharacterCreationHandler) handleGenericSelection(char *domainCharacter.Character, step *domainCharacter.CreationStep) (*core.Response, error) {
	// Generic handler for simple selection steps
	// The enhanced handler will provide the actual UI
	return h.buildEnhancedStepResponse(char, step)
}

func (h *CharacterCreationHandler) handleSpellSelection(char *domainCharacter.Character, step *domainCharacter.CreationStep) (*core.Response, error) {
	// TODO: Implement spell/cantrip selection UI
	// This handles both cantrips and regular spells
	return h.buildEnhancedStepResponse(char, step)
}

func (h *CharacterCreationHandler) handleComplete(char *domainCharacter.Character, step *domainCharacter.CreationStep) (*core.Response, error) {
	// TODO: Implement completion UI
	return h.buildEnhancedStepResponse(char, step)
}

// buildStepResponse routes to the appropriate handler based on step type
func (h *CharacterCreationHandler) buildStepResponse(char *domainCharacter.Character, step *domainCharacter.CreationStep) (*core.Response, error) {
	handler, ok := h.stepHandlers[step.Type]
	if !ok {
		return nil, fmt.Errorf("no handler registered for step type: %s", step.Type)
	}

	return handler(char, step)
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

	// Build response for current step using the handler registry
	response, err := h.buildStepResponse(char, currentStep)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Make the initial response ephemeral
	response.AsEphemeral()

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
