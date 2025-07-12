package handlers

import (
	"fmt"

	"github.com/KirkDiggler/dnd-bot-discord/internal/discord/v2/builders"
	"github.com/KirkDiggler/dnd-bot-discord/internal/discord/v2/core"
	domainCharacter "github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/bwmarrin/discordgo"
)

// HandleOpenEquipmentSelection shows the equipment selection UI
func (h *CharacterCreationHandler) HandleOpenEquipmentSelection(ctx *core.InteractionContext) (*core.HandlerResult, error) {
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

	// Get current step to verify we're on equipment selection
	currentStep, err := h.flowService.GetCurrentStep(ctx.Context, char.ID)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	if currentStep.Type != domainCharacter.StepTypeEquipmentSelection {
		return nil, core.NewValidationError("Not on equipment selection step")
	}

	var embed *builders.EmbedBuilder
	var components *builders.ComponentBuilder

	// Check if we have actual choices to make
	if len(currentStep.Options) > 0 && currentStep.MinChoices > 0 {
		// Show actual equipment selection UI
		embed = builders.NewEmbed().
			Title("⚔️ Choose Equipment").
			Description(currentStep.Description).
			Color(builders.ColorPrimary).
			AddField("🎒 Current Equipment", h.buildEquipmentList(char), false)

		components = builders.NewComponentBuilder(h.customIDBuilder)

		// Equipment choices can be complex with bundles, so we'll show them one at a time
		// Discord has a 25-option limit, so we need to handle this carefully
		var selectOptions []builders.SelectOption
		for i, option := range currentStep.Options {
			if i >= 24 { // Leave room for navigation
				break
			}
			selectOptions = append(selectOptions, builders.SelectOption{
				Label:       option.Name,
				Value:       option.Key,
				Description: option.Description,
			})
		}

		if len(selectOptions) > 0 {
			components.NewRow()
			components.SelectMenuWithTarget(
				fmt.Sprintf("Choose %d equipment option%s...", currentStep.MinChoices,
					func() string {
						if currentStep.MinChoices == 1 {
							return ""
						} else {
							return "s"
						}
					}()),
				"select_equipment",
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
		// Show auto-applied equipment UI
		embed = builders.NewEmbed().
			Title("⚔️ Equipment Selection").
			Description("Your character's starting equipment is automatically provided based on your class.").
			Color(builders.ColorInfo).
			AddField("🎒 Current Equipment", h.buildEquipmentList(char), false).
			AddField("📝 Note", "No additional equipment choices needed for this character!", false)

		components = builders.NewComponentBuilder(h.customIDBuilder)

		// Add continue button since no choices are needed
		components.PrimaryButton("✅ Continue to Character Details", "confirm_equipment_selection", char.ID)
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

// HandleSelectEquipment handles equipment selection from select menu
func (h *CharacterCreationHandler) HandleSelectEquipment(ctx *core.InteractionContext) (*core.HandlerResult, error) {
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
		return nil, core.NewValidationError("Please select at least one equipment option")
	}

	// Process the equipment step with actual selections
	result := &domainCharacter.CreationStepResult{
		StepType:   domainCharacter.StepTypeEquipmentSelection,
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

// HandleConfirmEquipmentSelection processes the equipment selection (for auto-applied equipment)
func (h *CharacterCreationHandler) HandleConfirmEquipmentSelection(ctx *core.InteractionContext) (*core.HandlerResult, error) {
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

	// Process the equipment step with empty selections (auto-applied equipment)
	result := &domainCharacter.CreationStepResult{
		StepType:   domainCharacter.StepTypeEquipmentSelection,
		Selections: []string{}, // Empty since equipment is auto-applied
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

// buildEquipmentList creates a formatted list of character's current equipment
func (h *CharacterCreationHandler) buildEquipmentList(char *domainCharacter.Character) string {
	if len(char.Inventory) == 0 && len(char.EquippedSlots) == 0 {
		return "*Starting equipment will be provided based on your class*"
	}

	var equipList []string

	// Show equipped items first
	if len(char.EquippedSlots) > 0 {
		equipList = append(equipList, "**⚡ Equipped:**")
		for slot, item := range char.EquippedSlots {
			equipList = append(equipList, fmt.Sprintf("• %s: %s", slot, item.GetName()))
		}
	}

	// Show inventory items
	if len(char.Inventory) > 0 {
		if len(equipList) > 0 {
			equipList = append(equipList, "") // Add blank line
		}
		equipList = append(equipList, "**🎒 Inventory:**")

		itemCount := 0
		for _, items := range char.Inventory {
			for _, item := range items {
				if itemCount >= 5 { // Limit display to first 5 items
					// Convert inventory to count total items
					totalItems := 0
					for _, items := range char.Inventory {
						totalItems += len(items)
					}
					equipList = append(equipList, fmt.Sprintf("• ... and %d more items", totalItems-5))
					break
				}
				equipList = append(equipList, fmt.Sprintf("• %s", item.GetName()))
				itemCount++
			}
			if itemCount >= 5 {
				break
			}
		}
	}

	if len(equipList) == 0 {
		return "*Loading equipment...*"
	}

	result := ""
	for i, item := range equipList {
		if i > 0 && item != "" {
			result += "\n"
		}
		result += item
	}
	return result
}
