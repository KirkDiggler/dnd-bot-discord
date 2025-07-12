package handlers

import (
	"fmt"

	"github.com/KirkDiggler/dnd-bot-discord/internal/discord/v2/builders"
	"github.com/KirkDiggler/dnd-bot-discord/internal/discord/v2/core"
	domainCharacter "github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/bwmarrin/discordgo"
)

// HandleOpenProficiencySelection shows the proficiency selection UI
func (h *CharacterCreationHandler) HandleOpenProficiencySelection(ctx *core.InteractionContext) (*core.HandlerResult, error) {
	// Parse custom ID to get character ID
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

	// Get current step to verify we're on proficiency selection
	currentStep, err := h.flowService.GetCurrentStep(ctx.Context, char.ID)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	if currentStep.Type != domainCharacter.StepTypeProficiencySelection {
		return nil, core.NewValidationError("Not on proficiency selection step")
	}

	var embed *builders.EmbedBuilder
	var components *builders.ComponentBuilder

	// Check if we have actual choices to make
	if len(currentStep.Options) > 0 && currentStep.MinChoices > 0 {
		// Show actual proficiency selection UI
		embed = builders.NewEmbed().
			Title("🛠️ Choose Proficiencies").
			Description(currentStep.Description).
			Color(builders.ColorPrimary).
			AddField("📚 Current Proficiencies", h.buildProficiencyList(char), false)

		components = builders.NewComponentBuilder(h.customIDBuilder)

		// Build select menu with proficiency options
		var selectOptions []builders.SelectOption
		for _, option := range currentStep.Options {
			selectOptions = append(selectOptions, builders.SelectOption{
				Label:       option.Name,
				Value:       option.Key,
				Description: option.Description,
			})
		}

		if len(selectOptions) > 0 {
			components.NewRow()
			components.SelectMenuWithTarget(
				fmt.Sprintf("Choose %d proficienc%s...", currentStep.MinChoices,
					func() string {
						if currentStep.MinChoices == 1 {
							return "y"
						} else {
							return "ies"
						}
					}()),
				"select_proficiency",
				char.ID,
				selectOptions,
				builders.SelectConfig{
					MinValues: currentStep.MinChoices,
					MaxValues: currentStep.MaxChoices,
				},
			)
		}

		components.NewRow()
		components.SecondaryButton("◀️ Back", "back", char.ID)
	} else {
		// Show auto-applied proficiencies UI
		embed = builders.NewEmbed().
			Title("🛠️ Proficiency Selection").
			Description("Your character's proficiencies are automatically applied based on your race and class.").
			Color(builders.ColorInfo).
			AddField("📚 Current Proficiencies", h.buildProficiencyList(char), false).
			AddField("📝 Note", "No additional proficiency choices needed for this character!", false)

		components = builders.NewComponentBuilder(h.customIDBuilder)

		// Add continue button since no choices are needed
		components.PrimaryButton("✅ Continue to Equipment", "confirm_proficiency_selection", char.ID)
		components.SecondaryButton("◀️ Back", "back", char.ID)
	}

	response := &core.Response{
		Embeds:     []*discordgo.MessageEmbed{embed.Build()},
		Components: components.Build(),
	}
	response.AsUpdate()

	return &core.HandlerResult{
		Response: response,
	}, nil
}

// HandleSelectProficiency handles proficiency selection from select menu
func (h *CharacterCreationHandler) HandleSelectProficiency(ctx *core.InteractionContext) (*core.HandlerResult, error) {
	// Parse custom ID to get character ID
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

	// Get selected values from the interaction
	var selectedValues []string
	if ctx.IsComponent() && ctx.Interaction != nil {
		data := ctx.Interaction.MessageComponentData()
		selectedValues = data.Values
	}

	if len(selectedValues) == 0 {
		return nil, core.NewValidationError("Please select at least one proficiency")
	}

	// Process the proficiency step with actual selections
	result := &domainCharacter.CreationStepResult{
		StepType:   domainCharacter.StepTypeProficiencySelection,
		Selections: selectedValues,
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

	response.AsUpdate()

	return &core.HandlerResult{
		Response: response,
	}, nil
}

// HandleConfirmProficiencySelection processes the proficiency selection (for auto-applied proficiencies)
func (h *CharacterCreationHandler) HandleConfirmProficiencySelection(ctx *core.InteractionContext) (*core.HandlerResult, error) {
	// Parse custom ID to get character ID
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

	// Process the proficiency step with empty selections (auto-applied proficiencies)
	result := &domainCharacter.CreationStepResult{
		StepType:   domainCharacter.StepTypeProficiencySelection,
		Selections: []string{}, // Empty since proficiencies are auto-applied
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

	response.AsUpdate()

	return &core.HandlerResult{
		Response: response,
	}, nil
}

// buildProficiencyList creates a formatted list of character's current proficiencies
func (h *CharacterCreationHandler) buildProficiencyList(char *domainCharacter.Character) string {
	if len(char.Proficiencies) == 0 {
		return "*Proficiencies will be loaded after class selection*"
	}

	var profList []string
	for profType, profs := range char.Proficiencies {
		if len(profs) == 0 {
			continue
		}

		typeStr := string(profType)
		switch profType {
		case "skill":
			typeStr = "📖 Skills"
		case "weapon":
			typeStr = "⚔️ Weapons"
		case "armor":
			typeStr = "🛡️ Armor"
		case "tool":
			typeStr = "🔧 Tools"
		case "language":
			typeStr = "🗣️ Languages"
		}

		var names []string
		for _, prof := range profs {
			names = append(names, prof.Name)
		}

		if len(names) <= 3 {
			profList = append(profList, fmt.Sprintf("%s: %v", typeStr, names))
		} else {
			profList = append(profList, fmt.Sprintf("%s: %v... (+%d more)", typeStr, names[:3], len(names)-3))
		}
	}

	if len(profList) == 0 {
		return "*Loading proficiencies...*"
	}

	result := ""
	for i, prof := range profList {
		if i > 0 {
			result += "\n"
		}
		result += prof
	}
	return result
}
