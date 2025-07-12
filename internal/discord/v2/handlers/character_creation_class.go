package handlers

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/discord/v2/builders"
	"github.com/KirkDiggler/dnd-bot-discord/internal/discord/v2/core"
	domainCharacter "github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	dnd5eFeatures "github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e/features"
)

// HandleClassOverview shows an overview of all available classes
func (h *CharacterCreationHandler) HandleClassOverview(ctx *core.InteractionContext) (*core.HandlerResult, error) {
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
		return nil, core.NewForbiddenError("You can only view your own characters")
	}

	// Build class overview embed
	embed := builders.NewEmbed().
		Title("⚔️ D&D 5e Classes Overview").
		Color(builders.ColorPrimary).
		Description("Here's an overview of all available classes in D&D 5e:")

	// Add class summaries
	classInfo := map[string]string{
		"barbarian": "Primal warriors who rage in battle",
		"bard":      "Magical performers and jack-of-all-trades",
		"cleric":    "Divine spellcasters who heal and protect",
		"druid":     "Nature priests who shapeshift and cast spells",
		"fighter":   "Masters of martial combat and tactics",
		"monk":      "Martial artists with mystical ki powers",
		"paladin":   "Holy warriors with divine magic",
		"ranger":    "Wilderness experts and hunters",
		"rogue":     "Stealthy skill experts and assassins",
		"sorcerer":  "Innate spellcasters with magical bloodlines",
		"warlock":   "Pact-bound spellcasters with eldritch powers",
		"wizard":    "Scholarly spellcasters who study magic",
	}

	for class, desc := range classInfo {
		emoji := getClassEmoji(class)
		// Capitalize first letter manually instead of using deprecated strings.Title
		className := strings.ToUpper(class[:1]) + class[1:]
		embed.AddField(fmt.Sprintf("%s %s", emoji, className), desc, true)
	}

	// Build components
	components := builders.NewComponentBuilder(h.customIDBuilder)
	components.SecondaryButton("⬅️ Back to Selection", "back", characterID)

	response := core.NewResponse("").
		WithEmbeds(embed.Build()).
		WithComponents(components.Build()...).
		AsUpdate()

	return &core.HandlerResult{
		Response: response,
	}, nil
}

// HandleRandomClass selects a random class for the character
func (h *CharacterCreationHandler) HandleRandomClass(ctx *core.InteractionContext) (*core.HandlerResult, error) {
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

	// Get current step to fetch class options
	currentStep, err := h.flowService.GetCurrentStep(ctx.Context, char.ID)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	if len(currentStep.Options) == 0 {
		return nil, core.NewInternalError(fmt.Errorf("no class options available"))
	}

	// Select a random class
	randomIndex := rand.Intn(len(currentStep.Options))
	selectedClass := currentStep.Options[randomIndex].Key

	// Use the flow service to process the selection
	result := &domainCharacter.CreationStepResult{
		StepType:   domainCharacter.StepTypeClassSelection,
		Selections: []string{selectedClass},
	}

	nextStep, err := h.flowService.ProcessStepResult(ctx.Context, char.ID, result)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Get updated character
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

// HandleClassPreview shows a preview of the character with the selected class
func (h *CharacterCreationHandler) HandleClassPreview(ctx *core.InteractionContext) (*core.HandlerResult, error) {
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

	// Get selected class value
	var selectedClass string
	if ctx.IsComponent() && ctx.Interaction != nil {
		data := ctx.Interaction.MessageComponentData()
		if len(data.Values) > 0 {
			selectedClass = data.Values[0]
		}
	}

	if selectedClass == "" {
		return nil, core.NewValidationError("Please select a class")
	}

	// Get current step to fetch class options
	currentStep, err := h.flowService.GetCurrentStep(ctx.Context, char.ID)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Find the selected class in options for display info
	var selectedClassOption *domainCharacter.CreationOption
	for _, opt := range currentStep.Options {
		if opt.Key == selectedClass {
			selectedClassOption = &opt
			break
		}
	}

	if selectedClassOption == nil {
		return nil, core.NewValidationError("Invalid class selection")
	}

	// Use the flow service to preview the selection
	previewResult := &domainCharacter.CreationStepResult{
		StepType:   domainCharacter.StepTypeClassSelection,
		Selections: []string{selectedClass},
	}

	previewChar, err := h.flowService.PreviewStepResult(ctx.Context, char.ID, previewResult)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Build enhanced response showing the preview
	embed := builders.NewEmbed().
		Title(fmt.Sprintf("%s Class Preview - %s", getClassEmoji(selectedClass), selectedClassOption.Name)).
		Color(builders.ColorPrimary).
		Description("Here's how your character looks with this class. You can select a different class or click 'Confirm Class Selection' to proceed.")

	// Add class details
	if previewChar.Class != nil {
		var features []string

		// Hit dice
		if previewChar.Class.HitDie > 0 {
			features = append(features,
				fmt.Sprintf("**Hit Dice:** 1d%d", previewChar.Class.HitDie),
				fmt.Sprintf("**Hit Points at 1st Level:** %d + CON modifier", previewChar.Class.HitDie))
		}

		// Primary ability
		primary := getClassPrimaryAbilities(selectedClass)
		if primary != "" {
			features = append(features, "**Primary Abilities:** "+primary)
		}

		// TODO: Add saving throw proficiencies when available in the Class struct

		if len(features) > 0 {
			embed.AddField("Class Basics", strings.Join(features, "\n"), false)
		}

		// Get level 1 class features
		classFeatures := dnd5eFeatures.GetClassFeatures(selectedClass, 1)
		if len(classFeatures) > 0 {
			var featureList []string
			for _, feat := range classFeatures {
				featureList = append(featureList, fmt.Sprintf("**%s** - %s", feat.Name, feat.Description))
			}
			embed.AddField("Level 1 Features", strings.Join(featureList, "\n\n"), false)
		}

		// Show starting proficiencies
		if len(previewChar.Class.Proficiencies) > 0 {
			var profs []string
			profCount := 0
			for _, prof := range previewChar.Class.Proficiencies {
				if profCount >= 5 {
					profs = append(profs, fmt.Sprintf("*...and %d more*", len(previewChar.Class.Proficiencies)-5))
					break
				}
				profs = append(profs, "• "+prof.Name)
				profCount++
			}
			embed.AddField("Starting Proficiencies", strings.Join(profs, "\n"), false)
		}
	}

	// Add character summary with preview
	progressValue := h.buildProgressSummary(previewChar)
	if progressValue != "" {
		embed.AddField("📊 Character Preview", progressValue, false)
	}

	// Build components
	components := builders.NewComponentBuilder(h.customIDBuilder)

	// Confirm button - pass the selected class as an argument
	components.SuccessButton("✅ Confirm Class Selection", "confirm_class", characterID, selectedClass)

	// Add select menu to choose a different class
	if len(currentStep.Options) > 0 {
		options := make([]builders.SelectOption, 0, len(currentStep.Options))
		for _, opt := range currentStep.Options {
			emoji := getClassEmoji(opt.Key)
			selected := opt.Key == selectedClass

			options = append(options, builders.SelectOption{
				Label:       fmt.Sprintf("%s %s", emoji, opt.Name),
				Value:       opt.Key,
				Description: opt.Description,
				Default:     selected,
			})
		}

		components.NewRow()
		components.SelectMenuWithTarget(
			"Or select a different class...",
			"preview_class",
			characterID,
			options,
		)
	}

	// Back button
	components.NewRow()
	components.SecondaryButton("⬅️ Back to Selection", "back", characterID)

	response := core.NewResponse("").
		WithEmbeds(embed.Build()).
		WithComponents(components.Build()...).
		AsUpdate()

	return &core.HandlerResult{
		Response: response,
	}, nil
}

// HandleConfirmClass confirms the class selection and proceeds to the next step
func (h *CharacterCreationHandler) HandleConfirmClass(ctx *core.InteractionContext) (*core.HandlerResult, error) {
	// Parse custom ID
	customID, err := core.ParseCustomID(ctx.GetCustomID())
	if err != nil {
		return nil, core.NewValidationError("Invalid selection")
	}

	characterID := customID.Target

	// Get the selected class from the args (passed from the confirm button)
	if len(customID.Args) == 0 {
		return nil, core.NewValidationError("No class specified")
	}
	selectedClass := customID.Args[0]

	// Get character
	char, err := h.service.GetCharacter(ctx.Context, characterID)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Verify ownership
	if char.OwnerID != ctx.UserID {
		return nil, core.NewForbiddenError("You can only edit your own characters")
	}

	// Confirm the class selection by processing it
	result := &domainCharacter.CreationStepResult{
		StepType:   domainCharacter.StepTypeClassSelection,
		Selections: []string{selectedClass},
	}

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
