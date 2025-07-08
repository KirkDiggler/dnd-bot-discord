package handlers

import (
	"fmt"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/discord/v2/builders"
	"github.com/KirkDiggler/dnd-bot-discord/internal/discord/v2/core"
	domainCharacter "github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	_ "github.com/bwmarrin/discordgo"
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
			// Create buttons for each race that show details when clicked
			components.NewRow()
			components.PrimaryButton("📜 View All Races", "race_list", char.ID)
			components.SecondaryButton("🎲 Random Race", "race_random", char.ID)

			// Add select menu with enhanced descriptions
			options := make([]builders.SelectOption, 0, len(step.Options))
			for _, opt := range step.Options {
				// Parse the bonuses for emoji representation
				description := opt.Description
				if strings.Contains(description, "+") {
					description = "💪 " + description
				}

				options = append(options, builders.SelectOption{
					Label:       opt.Name,
					Value:       opt.Key,
					Description: description,
				})
			}

			components.NewRow()
			components.SelectMenu(
				"Select a race...",
				fmt.Sprintf("select_%s", char.ID),
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
	var parts []string

	if char.Name != "" && char.Name != "Draft Character" {
		parts = append(parts, fmt.Sprintf("**Name:** %s", char.Name))
	}
	if char.Race != nil {
		parts = append(parts, fmt.Sprintf("**Race:** %s", char.Race.Name))
	}
	if char.Class != nil {
		parts = append(parts, fmt.Sprintf("**Class:** %s", char.Class.Name))
	}
	if char.Level > 0 {
		parts = append(parts, fmt.Sprintf("**Level:** %d", char.Level))
	}

	// Show ability scores if assigned
	if len(char.Attributes) > 0 {
		var scores []string
		for _, attr := range []shared.Attribute{
			shared.AttributeStrength,
			shared.AttributeDexterity,
			shared.AttributeConstitution,
			shared.AttributeIntelligence,
			shared.AttributeWisdom,
			shared.AttributeCharisma,
		} {
			if score, ok := char.Attributes[attr]; ok {
				scores = append(scores, fmt.Sprintf("%s %d", getAbilityAbbrev(string(attr)), score.Score))
			}
		}
		if len(scores) > 0 {
			parts = append(parts, fmt.Sprintf("**Abilities:** %s", strings.Join(scores, " | ")))
		}
	}

	return strings.Join(parts, "\n")
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
	customID, err := core.ParseCustomID(ctx.GetCustomID())
	if err != nil {
		return nil, core.NewValidationError("Invalid selection")
	}

	characterID := customID.Target

	// Get character and verify ownership
	char, err := h.service.GetCharacter(ctx.Context, characterID)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	if char.OwnerID != ctx.UserID {
		return nil, core.NewForbiddenError("You can only view details for your own character")
	}

	// Build race overview embed
	embed := builders.NewEmbed().
		Title("🎭 Race Overview").
		Description("Each race in D&D provides unique traits and abilities. Here's what makes each one special:").
		Color(builders.ColorInfo)

	// Add fields for each major race
	raceDetails := map[string]string{
		"Human":      "**Versatile and ambitious**\n+1 to all abilities\nExtra feat at 1st level\nExtra skill proficiency",
		"Elf":        "**Graceful and long-lived**\n+2 DEX\nDarkvision 60ft\nKeen Senses\nFey Ancestry\nTrance",
		"Dwarf":      "**Sturdy and enduring**\n+2 CON\nDarkvision 60ft\nDwarven Resilience\nStonecunning",
		"Halfling":   "**Small but brave**\n+2 DEX\nLucky\nBrave\nHalfling Nimbleness",
		"Dragonborn": "**Draconic heritage**\n+2 STR, +1 CHA\nDraconic Ancestry\nBreath Weapon\nDamage Resistance",
		"Gnome":      "**Clever and curious**\n+2 INT\nDarkvision 60ft\nGnome Cunning",
		"Half-Elf":   "**Best of both worlds**\n+2 CHA, +1 to two others\nDarkvision 60ft\nFey Ancestry\nExtra skills",
		"Half-Orc":   "**Strong and resilient**\n+2 STR, +1 CON\nDarkvision 60ft\nRelentless Endurance\nSavage Attacks",
		"Tiefling":   "**Infernal heritage**\n+2 CHA, +1 INT\nDarkvision 60ft\nHellish Resistance\nInfernal Legacy",
	}

	for race, details := range raceDetails {
		embed.AddField(race, details, true)
	}

	// Add back button
	components := builders.NewComponentBuilder(h.customIDBuilder).
		SecondaryButton("⬅️ Back to Selection", "back", characterID)

	response := core.NewResponse("").
		WithEmbeds(embed.Build()).
		WithComponents(components.Build()...).
		AsUpdate()

	return &core.HandlerResult{
		Response: response,
	}, nil
}

// HandleRandomRace selects a random race
func (h *CharacterCreationHandler) HandleRandomRace(ctx *core.InteractionContext) (*core.HandlerResult, error) {
	// TODO: Implement random race selection
	return nil, core.NewInternalError(fmt.Errorf("random race selection not yet implemented"))
}

// Add these handlers to the router registration
func (h *CharacterCreationHandler) RegisterAdditionalHandlers(router *core.Router) {
	router.ComponentFunc("race_list", h.HandleRaceDetails)
	router.ComponentFunc("race_random", h.HandleRandomRace)
	// Add more handlers as we implement them
}
