package handlers

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/discord/v2/builders"
	"github.com/KirkDiggler/dnd-bot-discord/internal/discord/v2/core"
	domainCharacter "github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	rulebook "github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
)

// buildEnhancedStepResponse builds a rich Discord response for a creation step
func (h *CharacterCreationHandler) buildEnhancedStepResponse(char *domainCharacter.Character, step *domainCharacter.CreationStep) (*core.Response, error) {
	// For now, we'll build a mock list of all steps
	// TODO: Get this from the flow service when it supports GetAllSteps
	allSteps := h.buildMockAllSteps(char)

	// Build base embed
	embed := builders.NewEmbed().
		Title(fmt.Sprintf("🎭 Character Creation - %s", step.Title)).
		Color(builders.ColorPrimary)

	// Add dynamic progress tracker
	progressTracker := h.buildDynamicProgressTracker(char, step, allSteps)
	if progressTracker != "" {
		embed.AddField("📍 Creation Progress", progressTracker, false)
	}

	// Add character summary
	progressValue := h.buildProgressSummary(char)
	if progressValue != "" {
		embed.AddField("📊 Your Character", progressValue, false)
	}

	// Build components based on step type
	components := builders.NewComponentBuilder(h.customIDBuilder)

	switch step.Type {
	case domainCharacter.StepTypeRaceSelection:
		embed.Description("Choose your character's race. Each race provides unique bonuses and traits that will shape your character's abilities.")

		// For race selection, we'll create a more detailed view
		if len(step.Options) > 0 {
			// Random race button
			components.NewRow()
			components.SecondaryButton("🎲 Random Race", "race_random", char.ID)

			// Add select menu with enhanced descriptions
			options := make([]builders.SelectOption, 0, len(step.Options))
			for _, opt := range step.Options {
				// Get race-specific emoji
				emoji := getRaceEmoji(opt.Key)

				options = append(options, builders.SelectOption{
					Label:       opt.Name,
					Value:       opt.Key,
					Description: opt.Description,
					Emoji:       emoji,
				})
			}

			components.NewRow()
			components.SelectMenuWithTarget(
				"Select a race...",
				"preview_race",
				char.ID,
				options,
			)
		}

	case domainCharacter.StepTypeClassSelection:
		embed.Description("Choose your character's class. Your class determines your abilities, skills, and role in the party.")

		if len(step.Options) > 0 {
			// Add class overview button
			components.NewRow()
			components.PrimaryButton("📖 Class Overview", "class_overview", char.ID)
			components.SecondaryButton("🎲 Random Class", "class_random", char.ID)

			// Create select menu with class info
			options := make([]builders.SelectOption, 0, len(step.Options))
			for _, opt := range step.Options {
				// Add emoji based on class archetype
				emoji := getClassEmoji(opt.Key)

				options = append(options, builders.SelectOption{
					Label:       fmt.Sprintf("%s %s", emoji, opt.Name),
					Value:       opt.Key,
					Description: opt.Description,
				})
			}

			components.NewRow()
			components.SelectMenu(
				"Select a class...",
				fmt.Sprintf("select_%s", char.ID),
				options,
			)
		}

	case domainCharacter.StepTypeAbilityScores:
		embed.Description("Time to determine your character's abilities! You can roll for random scores or use the standard array.")

		// Show current race/class bonuses that will apply
		if char.Race != nil {
			var bonuses []string
			for _, bonus := range char.Race.AbilityBonuses {
				if bonus.Bonus > 0 {
					bonuses = append(bonuses, fmt.Sprintf("%s +%d", getAbilityAbbrev(string(bonus.Attribute)), bonus.Bonus))
				}
			}
			if len(bonuses) > 0 {
				embed.AddField("🎯 Racial Bonuses", strings.Join(bonuses, ", "), true)
			}
		}

		components.PrimaryButton("🎲 Roll Ability Scores", "roll", char.ID)
		components.SecondaryButton("📊 Use Standard Array", "standard", char.ID)
		components.NewRow()
		components.SecondaryButton("❓ About Ability Scores", "ability_info", char.ID)

	case domainCharacter.StepTypeAbilityAssignment:
		embed.Description("Assign your rolled scores to your abilities. Consider your class's primary abilities!")

		// Show class recommendations
		if char.Class != nil {
			primary := getClassPrimaryAbilities(char.Class.Key)
			if primary != "" {
				embed.AddField("💡 Recommended", primary, true)
			}
		}

		components.PrimaryButton("📋 Assign Abilities", "assign", char.ID)
		components.SecondaryButton("🔄 Reroll", "reroll", char.ID)

	case domainCharacter.StepTypeProficiencySelection:
		embed.Description("Choose your character's proficiencies. These determine what skills and tools you're trained in.")
		components.PrimaryButton("🛠️ Choose Proficiencies", "proficiencies", char.ID)

	case domainCharacter.StepTypeEquipmentSelection:
		embed.Description("Select your starting equipment. Choose wisely based on your character's role!")
		components.PrimaryButton("⚔️ Choose Equipment", "equipment", char.ID)

	case domainCharacter.StepTypeCharacterDetails:
		embed.Description("Almost done! Give your character a name and finalize their details.")
		components.PrimaryButton("✏️ Set Character Name", "name", char.ID)

	default:
		// Handle any custom step types
		h.buildDefaultStepComponents(components, step, char)
	}

	// Add help button at the bottom
	components.NewRow()
	components.SecondaryButton("❓ Help", "help", fmt.Sprintf("%s_%s", string(step.Type), char.ID))

	// Add back button if we have state tracking
	if h.canGoBack(char) {
		components.SecondaryButton("⬅️ Back", "back", char.ID)
	}

	response := core.NewResponse("").
		WithEmbeds(embed.Build()).
		WithComponents(components.Build()...).
		AsEphemeral()

	return response, nil
}

// buildProgressSummary creates a summary of character creation progress
func (h *CharacterCreationHandler) buildProgressSummary(char *domainCharacter.Character) string {
	var sections []string

	// Name and Level Header
	name := "Unnamed Hero"
	if char.Name != "" && char.Name != "Draft Character" {
		name = char.Name
	}
	level := 1
	if char.Level > 0 {
		level = char.Level
	}
	sections = append(sections, fmt.Sprintf("**%s** • Level %d", name, level))

	// Race Section
	if char.Race != nil {
		var raceDetails []string
		raceDetails = append(raceDetails, fmt.Sprintf("**🎭 Race:** %s", char.Race.Name))

		// Ability bonuses
		var bonuses []string
		for _, bonus := range char.Race.AbilityBonuses {
			if bonus.Bonus > 0 {
				bonuses = append(bonuses, fmt.Sprintf("%s +%d", getAbilityAbbrev(string(bonus.Attribute)), bonus.Bonus))
			}
		}
		if len(bonuses) > 0 {
			raceDetails = append(raceDetails, fmt.Sprintf("• **Ability Bonuses:** %s", strings.Join(bonuses, ", ")))
		}

		// Speed
		if char.Race.Speed > 0 {
			raceDetails = append(raceDetails, fmt.Sprintf("• **Speed:** %d feet", char.Race.Speed))
		}

		// Starting proficiencies
		if len(char.Race.StartingProficiencies) > 0 {
			var profs []string
			for _, prof := range char.Race.StartingProficiencies {
				profs = append(profs, prof.Name)
			}
			if len(profs) > 0 {
				raceDetails = append(raceDetails, fmt.Sprintf("• **Racial Proficiencies:** %s", strings.Join(profs, ", ")))
			}
		}

		sections = append(sections, strings.Join(raceDetails, "\n"))
	} else {
		sections = append(sections, "**🎭 Race:** *Not selected*")
	}

	// Class Section
	if char.Class != nil {
		classDetails := []string{
			fmt.Sprintf("**⚔️ Class:** %s", char.Class.Name),
			fmt.Sprintf("• **Hit Die:** d%d", char.Class.HitDie),
			fmt.Sprintf("• **Primary Ability:** %s", char.Class.GetPrimaryAbility()),
		}

		sections = append(sections, strings.Join(classDetails, "\n"))
	} else {
		sections = append(sections, "**⚔️ Class:** *Not selected*")
	}

	// Ability Scores Section
	if len(char.Attributes) > 0 {
		var abilitySection []string
		abilitySection = append(abilitySection, "**📊 Ability Scores:**")

		// Create a formatted ability score display
		for _, attr := range []shared.Attribute{
			shared.AttributeStrength,
			shared.AttributeDexterity,
			shared.AttributeConstitution,
			shared.AttributeIntelligence,
			shared.AttributeWisdom,
			shared.AttributeCharisma,
		} {
			if score, ok := char.Attributes[attr]; ok {
				modifier := (score.Score - 10) / 2
				modStr := fmt.Sprintf("%+d", modifier)
				if modifier >= 0 {
					modStr = "+" + fmt.Sprintf("%d", modifier)
				}
				abilitySection = append(abilitySection,
					fmt.Sprintf("• **%s:** %d (%s)", getAbilityAbbrev(string(attr)), score.Score, modStr))
			}
		}

		sections = append(sections, strings.Join(abilitySection, "\n"))
	}

	// Proficiencies Section
	if len(char.Proficiencies) > 0 {
		var profSection []string
		profSection = append(profSection, "**🛠️ Proficiencies:**")

		// Group by type
		for profType, profs := range char.Proficiencies {
			if len(profs) == 0 {
				continue
			}
			var profNames []string
			for _, prof := range profs {
				profNames = append(profNames, prof.Name)
			}
			typeLabel := string(profType)
			// Make type label more readable
			switch profType {
			case rulebook.ProficiencyTypeArmor:
				typeLabel = "Armor"
			case rulebook.ProficiencyTypeWeapon:
				typeLabel = "Weapons"
			case rulebook.ProficiencyTypeSkill:
				typeLabel = "Skills"
			case rulebook.ProficiencyTypeTool:
				typeLabel = "Tools"
			case rulebook.ProficiencyTypeSavingThrow:
				typeLabel = "Saving Throws"
			case rulebook.ProficiencyTypeInstrument:
				typeLabel = "Instruments"
			}
			profSection = append(profSection, fmt.Sprintf("• **%s:** %s", typeLabel, strings.Join(profNames, ", ")))
		}

		sections = append(sections, strings.Join(profSection, "\n"))
	}

	// Features Section (if any)
	if len(char.Features) > 0 {
		var featureSection []string
		featureSection = append(featureSection, "**✨ Features:**")
		for _, feature := range char.Features {
			featureSection = append(featureSection, fmt.Sprintf("• %s", feature.Name))
		}
		sections = append(sections, strings.Join(featureSection, "\n"))
	}

	return strings.Join(sections, "\n\n")
}

// Helper functions
func getClassEmoji(classKey string) string {
	emojis := map[string]string{
		"barbarian": "🪓",
		"bard":      "🎵",
		"cleric":    "⛪",
		"druid":     "🌿",
		"fighter":   "⚔️",
		"monk":      "👊",
		"paladin":   "🛡️",
		"ranger":    "🏹",
		"rogue":     "🗡️",
		"sorcerer":  "✨",
		"warlock":   "👹",
		"wizard":    "🧙",
	}

	if emoji, ok := emojis[classKey]; ok {
		return emoji
	}
	return "📚"
}

func getAbilityAbbrev(ability string) string {
	abbrevs := map[string]string{
		"Strength":                           "STR",
		"Dexterity":                          "DEX",
		"Constitution":                       "CON",
		"Intelligence":                       "INT",
		"Wisdom":                             "WIS",
		"Charisma":                           "CHA",
		string(shared.AttributeStrength):     "STR",
		string(shared.AttributeDexterity):    "DEX",
		string(shared.AttributeConstitution): "CON",
		string(shared.AttributeIntelligence): "INT",
		string(shared.AttributeWisdom):       "WIS",
		string(shared.AttributeCharisma):     "CHA",
	}

	if abbrev, ok := abbrevs[ability]; ok {
		return abbrev
	}
	return ability
}

func getClassPrimaryAbilities(classKey string) string {
	primaryAbilities := map[string]string{
		"barbarian": "STR & CON - Strength for attacks, Constitution for survivability",
		"bard":      "CHA - Charisma powers your spells and social skills",
		"cleric":    "WIS - Wisdom determines your spellcasting power",
		"druid":     "WIS - Wisdom for spells, Constitution for wild shape",
		"fighter":   "STR or DEX - Depending on melee or ranged focus",
		"monk":      "DEX & WIS - Dexterity for attacks, Wisdom for Ki",
		"paladin":   "STR & CHA - Strength for combat, Charisma for spells",
		"ranger":    "DEX & WIS - Dexterity for attacks, Wisdom for spells",
		"rogue":     "DEX - Dexterity for stealth and sneak attacks",
		"sorcerer":  "CHA - Charisma fuels your innate magic",
		"warlock":   "CHA - Charisma channels your patron's power",
		"wizard":    "INT - Intelligence determines spell power and knowledge",
	}

	if primary, ok := primaryAbilities[classKey]; ok {
		return primary
	}
	return ""
}

func (h *CharacterCreationHandler) buildDefaultStepComponents(components *builders.ComponentBuilder, step *domainCharacter.CreationStep, char *domainCharacter.Character) {
	if len(step.Options) > 0 {
		// Check if this is a multi-select
		if step.MaxChoices > 1 {
			options := make([]builders.SelectOption, 0, len(step.Options))
			for _, opt := range step.Options {
				options = append(options, builders.SelectOption{
					Label:       opt.Name,
					Value:       opt.Key,
					Description: opt.Description,
				})
			}

			placeholder := fmt.Sprintf("Select %d options...", step.MinChoices)
			if step.MinChoices != step.MaxChoices {
				placeholder = fmt.Sprintf("Select %d-%d options...", step.MinChoices, step.MaxChoices)
			}

			components.SelectMenuWithOptions(
				placeholder,
				fmt.Sprintf("select_%s", char.ID),
				options,
				step.MinChoices,
				step.MaxChoices,
			)
		} else {
			// Single select with buttons
			for i, opt := range step.Options {
				if i > 0 && i%5 == 0 {
					components.NewRow()
				}
				components.PrimaryButton(opt.Name, fmt.Sprintf("option_%s", opt.Key), char.ID)
			}
		}
	}
}

func (h *CharacterCreationHandler) canGoBack(char *domainCharacter.Character) bool {
	// TODO: Implement when we have state tracking
	return false
}

// buildDynamicProgressTracker creates a visual progress tracker that shows completed and remaining steps
func (h *CharacterCreationHandler) buildDynamicProgressTracker(char *domainCharacter.Character, currentStep *domainCharacter.CreationStep, allSteps []domainCharacter.CreationStep) string {
	var tracker []string

	// Define the base steps that everyone goes through
	baseSteps := []struct {
		stepType domainCharacter.CreationStepType
		name     string
		emoji    string
	}{
		{domainCharacter.StepTypeRaceSelection, "Race", "🎭"},
		{domainCharacter.StepTypeClassSelection, "Class", "⚔️"},
		{domainCharacter.StepTypeAbilityScores, "Abilities", "🎲"},
		{domainCharacter.StepTypeAbilityAssignment, "Assign", "📊"},
	}

	// Track completed steps
	completedSteps := make(map[domainCharacter.CreationStepType]bool)
	if char.Race != nil {
		completedSteps[domainCharacter.StepTypeRaceSelection] = true
	}
	if char.Class != nil {
		completedSteps[domainCharacter.StepTypeClassSelection] = true
	}
	if len(char.Attributes) > 0 {
		completedSteps[domainCharacter.StepTypeAbilityScores] = true
		completedSteps[domainCharacter.StepTypeAbilityAssignment] = true
	}

	// Build the base progress
	for _, step := range baseSteps {
		if completedSteps[step.stepType] {
			tracker = append(tracker, fmt.Sprintf("✅ %s %s", step.emoji, step.name))
		} else if currentStep.Type == step.stepType {
			tracker = append(tracker, fmt.Sprintf("▶️ **%s %s**", step.emoji, step.name))
		} else {
			tracker = append(tracker, fmt.Sprintf("⬜ %s %s", step.emoji, step.name))
		}
	}

	// Add class-specific steps dynamically
	if char.Class != nil {
		// Check for class-specific steps in allSteps
		classSpecificSteps := []struct {
			stepType domainCharacter.CreationStepType
			name     string
			emoji    string
			classes  []string
		}{
			{domainCharacter.StepTypeDivineDomainSelection, "Divine Domain", "⛪", []string{"cleric"}},
			{domainCharacter.StepTypeFightingStyleSelection, "Fighting Style", "🛡️", []string{"fighter", "ranger", "paladin"}},
			{domainCharacter.StepTypeFavoredEnemySelection, "Favored Enemy", "🎯", []string{"ranger"}},
			{domainCharacter.StepTypeNaturalExplorerSelection, "Natural Explorer", "🏔️", []string{"ranger"}},
		}

		for _, specStep := range classSpecificSteps {
			// Check if this step applies to the character's class
			if contains(specStep.classes, char.Class.Key) {
				// Check if this step exists in allSteps
				for _, step := range allSteps {
					if step.Type == specStep.stepType {
						if h.isStepComplete(char, specStep.stepType) {
							tracker = append(tracker, fmt.Sprintf("✅ %s %s", specStep.emoji, specStep.name))
						} else if currentStep.Type == specStep.stepType {
							tracker = append(tracker, fmt.Sprintf("▶️ **%s %s**", specStep.emoji, specStep.name))
						} else {
							tracker = append(tracker, fmt.Sprintf("⬜ %s %s", specStep.emoji, specStep.name))
						}
						break
					}
				}
			}
		}
	}

	// Add remaining universal steps
	universalSteps := []struct {
		stepType domainCharacter.CreationStepType
		name     string
		emoji    string
	}{
		{domainCharacter.StepTypeProficiencySelection, "Proficiencies", "🛠️"},
		{domainCharacter.StepTypeEquipmentSelection, "Equipment", "🎒"},
		{domainCharacter.StepTypeCharacterDetails, "Details", "📝"},
	}

	for _, step := range universalSteps {
		if h.isStepComplete(char, step.stepType) {
			tracker = append(tracker, fmt.Sprintf("✅ %s %s", step.emoji, step.name))
		} else if currentStep.Type == step.stepType {
			tracker = append(tracker, fmt.Sprintf("▶️ **%s %s**", step.emoji, step.name))
		} else {
			tracker = append(tracker, fmt.Sprintf("⬜ %s %s", step.emoji, step.name))
		}
	}

	// Add step counter
	totalSteps := len(tracker)
	completedCount := 0
	for _, t := range tracker {
		if strings.HasPrefix(t, "✅") {
			completedCount++
		}
	}

	// Build the final tracker with counter
	result := fmt.Sprintf("**Step %d of %d**\n\n", completedCount+1, totalSteps)
	result += strings.Join(tracker, "\n")

	// Add estimated time remaining
	stepsRemaining := totalSteps - completedCount
	if stepsRemaining > 0 {
		minutesRemaining := stepsRemaining * 2 // Estimate 2 minutes per step
		result += fmt.Sprintf("\n\n⏱️ *Estimated time: ~%d minutes*", minutesRemaining)
	}

	return result
}

// isStepComplete checks if a specific step type has been completed
func (h *CharacterCreationHandler) isStepComplete(char *domainCharacter.Character, stepType domainCharacter.CreationStepType) bool {
	// TODO: This should check the actual completion state from the flow service
	// For now, we'll use some basic checks
	switch stepType {
	case domainCharacter.StepTypeRaceSelection:
		return char.Race != nil
	case domainCharacter.StepTypeClassSelection:
		return char.Class != nil
	case domainCharacter.StepTypeAbilityScores, domainCharacter.StepTypeAbilityAssignment:
		return len(char.Attributes) == 6
	case domainCharacter.StepTypeProficiencySelection:
		return len(char.Proficiencies) > 0
	case domainCharacter.StepTypeEquipmentSelection:
		return len(char.EquippedSlots) > 0
	case domainCharacter.StepTypeCharacterDetails:
		return char.Name != "" && char.Name != "Draft Character"
	default:
		// For class-specific steps, we'd need to check the character's state
		// This would be better handled by the flow service
		return false
	}
}

// contains checks if a string is in a slice
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// buildMockAllSteps creates a mock list of all steps for the character
// TODO: Replace this with actual flow service method when available
func (h *CharacterCreationHandler) buildMockAllSteps(char *domainCharacter.Character) []domainCharacter.CreationStep {
	steps := []domainCharacter.CreationStep{
		{Type: domainCharacter.StepTypeRaceSelection},
		{Type: domainCharacter.StepTypeClassSelection},
		{Type: domainCharacter.StepTypeAbilityScores},
		{Type: domainCharacter.StepTypeAbilityAssignment},
	}

	// Add class-specific steps if class is selected
	if char.Class != nil {
		switch char.Class.Key {
		case "cleric":
			steps = append(steps, domainCharacter.CreationStep{Type: domainCharacter.StepTypeDivineDomainSelection})
		case "fighter", "paladin":
			steps = append(steps, domainCharacter.CreationStep{Type: domainCharacter.StepTypeFightingStyleSelection})
		case "ranger":
			steps = append(steps,
				domainCharacter.CreationStep{Type: domainCharacter.StepTypeFightingStyleSelection},
				domainCharacter.CreationStep{Type: domainCharacter.StepTypeFavoredEnemySelection},
				domainCharacter.CreationStep{Type: domainCharacter.StepTypeNaturalExplorerSelection},
			)
		}
	}

	// Add universal end steps
	steps = append(steps,
		domainCharacter.CreationStep{Type: domainCharacter.StepTypeProficiencySelection},
		domainCharacter.CreationStep{Type: domainCharacter.StepTypeEquipmentSelection},
		domainCharacter.CreationStep{Type: domainCharacter.StepTypeCharacterDetails},
	)

	return steps
}

// HandleRaceDetails shows detailed information about races
func (h *CharacterCreationHandler) HandleRaceDetails(ctx *core.InteractionContext) (*core.HandlerResult, error) {
	// This isn't needed since all races are in the dropdown
	return nil, core.NewNotImplementedError("race details view")
}

// HandleRandomRace selects a random race
func (h *CharacterCreationHandler) HandleRandomRace(ctx *core.InteractionContext) (*core.HandlerResult, error) {
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

	// Get current step to get race options
	currentStep, err := h.flowService.GetCurrentStep(ctx.Context, char.ID)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	if len(currentStep.Options) == 0 {
		return nil, core.NewInternalError(fmt.Errorf("no race options available"))
	}

	// Select a random race
	randomIndex := rand.Intn(len(currentStep.Options))
	selectedRace := currentStep.Options[randomIndex]

	// Process the selection
	result := &domainCharacter.CreationStepResult{
		StepType:   domainCharacter.StepTypeRaceSelection,
		Selections: []string{selectedRace.Key},
	}

	nextStep, err := h.flowService.ProcessStepResult(ctx.Context, char.ID, result)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Re-fetch character to get updated state
	char, err = h.service.GetCharacter(ctx.Context, characterID)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Build the response for the next step
	response, err := h.buildEnhancedStepResponse(char, nextStep)
	if err != nil {
		return nil, err
	}

	response.Ephemeral = true

	return &core.HandlerResult{
		Response: response,
	}, nil
}

// HandleRacePreview shows a preview of the character with the selected race
func (h *CharacterCreationHandler) HandleRacePreview(ctx *core.InteractionContext) (*core.HandlerResult, error) {
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

	// Get selected race value
	var selectedRace string
	if ctx.IsComponent() && ctx.Interaction != nil {
		data := ctx.Interaction.MessageComponentData()
		if len(data.Values) > 0 {
			selectedRace = data.Values[0]
		}
	}

	if selectedRace == "" {
		return nil, core.NewValidationError("Please select a race")
	}

	// Get current step to fetch race options
	currentStep, err := h.flowService.GetCurrentStep(ctx.Context, char.ID)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Find the selected race in options for display info
	var selectedRaceOption *domainCharacter.CreationOption
	for _, opt := range currentStep.Options {
		if opt.Key == selectedRace {
			selectedRaceOption = &opt
			break
		}
	}

	if selectedRaceOption == nil {
		return nil, core.NewValidationError("Invalid race selection")
	}

	// Use the flow service to preview the selection
	previewResult := &domainCharacter.CreationStepResult{
		StepType:   domainCharacter.StepTypeRaceSelection,
		Selections: []string{selectedRace},
	}

	previewChar, err := h.flowService.PreviewStepResult(ctx.Context, char.ID, previewResult)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Build enhanced response showing the preview
	embed := builders.NewEmbed().
		Title("🎭 Race Preview - " + selectedRaceOption.Name).
		Color(builders.ColorPrimary).
		Description("Here's how your character looks with this race. You can select a different race or click 'Confirm Race Selection' to proceed.")

	// Add race details
	if previewChar.Race != nil {
		var traits []string

		// Ability bonuses
		if bonuses, ok := selectedRaceOption.Metadata["bonuses"].([]string); ok && len(bonuses) > 0 {
			traits = append(traits, "**Ability Bonuses:** "+strings.Join(bonuses, ", "))
		}

		// Speed
		if previewChar.Race.Speed > 0 {
			traits = append(traits, fmt.Sprintf("**Speed:** %d feet", previewChar.Race.Speed))
		}

		// Additional traits could be added here based on race key
		// For example, we could have a mapping of race to size

		if len(traits) > 0 {
			embed.AddField("Race Traits", strings.Join(traits, "\n"), false)
		}
	}

	// Add character summary with preview
	progressValue := h.buildProgressSummary(previewChar)
	if progressValue != "" {
		embed.AddField("📊 Character Preview", progressValue, false)
	}

	// Build components
	components := builders.NewComponentBuilder(h.customIDBuilder)

	// Race selection dropdown with current selection
	if len(currentStep.Options) > 0 {
		options := make([]builders.SelectOption, 0, len(currentStep.Options))
		for _, opt := range currentStep.Options {
			option := builders.SelectOption{
				Label:       opt.Name,
				Value:       opt.Key,
				Description: opt.Description,
			}
			// Mark the currently selected option
			if opt.Key == selectedRace {
				option.Default = true
			}
			options = append(options, option)
		}

		components.SelectMenu(
			"Preview a different race...",
			h.customIDBuilder.Build("preview_race").WithTarget(char.ID).MustEncode(),
			options,
		)
	}

	// Action buttons
	components.NewRow()
	components.SuccessButton("✅ Confirm Race Selection", "confirm_race", char.ID, selectedRace)
	components.SecondaryButton("🎲 Random Race", "race_random", char.ID)

	response := core.NewResponse("").
		WithEmbeds(embed.Build()).
		WithComponents(components.Build()...).
		AsUpdate()

	return &core.HandlerResult{
		Response: response,
	}, nil
}

// HandleConfirmRace processes the confirmed race selection
func (h *CharacterCreationHandler) HandleConfirmRace(ctx *core.InteractionContext) (*core.HandlerResult, error) {
	// Parse custom ID
	customID, err := core.ParseCustomID(ctx.GetCustomID())
	if err != nil {
		return nil, core.NewValidationError("Invalid selection")
	}

	characterID := customID.Target

	// Get the selected race from the args (passed from the confirm button)
	if len(customID.Args) == 0 {
		return nil, core.NewValidationError("No race specified")
	}
	selectedRace := customID.Args[0]

	// Get character
	char, err := h.service.GetCharacter(ctx.Context, characterID)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Verify ownership
	if char.OwnerID != ctx.UserID {
		return nil, core.NewForbiddenError("You can only edit your own characters")
	}

	// Create step result and process it
	result := &domainCharacter.CreationStepResult{
		StepType:   domainCharacter.StepTypeRaceSelection,
		Selections: []string{selectedRace},
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

// getRaceEmoji returns an appropriate emoji for a race
func getRaceEmoji(raceKey string) string {
	emojiMap := map[string]string{
		"dragonborn": "🐲",
		"dwarf":      "⛏️",
		"elf":        "🧝",
		"gnome":      "🧙",
		"half-elf":   "🌗",
		"halfling":   "🍄",
		"half-orc":   "💪",
		"human":      "👤",
		"tiefling":   "😈",
	}

	if emoji, ok := emojiMap[raceKey]; ok {
		return emoji
	}
	return "🎭" // Default race emoji
}
