package handlers

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/discord/v2/builders"
	"github.com/KirkDiggler/dnd-bot-discord/internal/discord/v2/core"
	domainCharacter "github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	rulebook "github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	characterService "github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/bwmarrin/discordgo"
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

	// Build UI based on step type
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
			components.SelectMenuWithTarget(
				"Select a class...",
				"preview_class",
				char.ID,
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
		components.DangerButton("🗑️ Start Fresh", "start_fresh", char.ID)

	case domainCharacter.StepTypeAbilityAssignment:
		embed.Description("Your ability scores have been rolled! Now assign them to your abilities.")

		// Show rolled scores if available
		if len(char.AbilityRolls) > 0 {
			var scoreDisplay []string
			for _, roll := range char.AbilityRolls {
				scoreDisplay = append(scoreDisplay, fmt.Sprintf("**%d**", roll.Value))
			}
			embed.AddField("🎯 Available Scores", strings.Join(scoreDisplay, " • "), false)
		}

		// Show class recommendations
		if char.Class != nil {
			primary := getClassPrimaryAbilities(char.Class.Key)
			if primary != "" {
				embed.AddField("💡 Class Recommendations", primary, false)
			}
		}

		components.SuccessButton("🤖 Auto-Assign & Continue", "auto_assign_and_continue", char.ID)
		components.PrimaryButton("📝 Manual Assignment", "start_manual_assignment", char.ID)

	case domainCharacter.StepTypeProficiencySelection:
		embed.Description("Choose your character's proficiencies. These determine what skills and tools you're trained in.")
		components.PrimaryButton("🛠️ Choose Proficiencies", "proficiencies", char.ID)

	case domainCharacter.StepTypeEquipmentSelection:
		embed.Description("Select your starting equipment. Choose wisely based on your character's role!")
		components.PrimaryButton("⚔️ Choose Equipment", "equipment", char.ID)

	case domainCharacter.StepTypeCharacterDetails:
		embed.Description("Almost done! Give your character a name and finalize their details.")
		components.PrimaryButton("✏️ Set Character Name", "name", char.ID)

	case domainCharacter.StepTypeSkillSelection:
		embed.Description(step.Description)
		h.buildSelectMenuFromOptions(components, step, char.ID, "skill")

	case domainCharacter.StepTypeLanguageSelection:
		embed.Description(step.Description)
		h.buildSelectMenuFromOptions(components, step, char.ID, "language")

	case domainCharacter.StepTypeFightingStyleSelection:
		embed.Description(step.Description)
		h.buildSelectMenuFromOptions(components, step, char.ID, "fighting_style")

	case domainCharacter.StepTypeDivineDomainSelection:
		embed.Description(step.Description)
		h.buildSelectMenuFromOptions(components, step, char.ID, "divine_domain")

	case domainCharacter.StepTypeFavoredEnemySelection:
		embed.Description(step.Description)
		h.buildSelectMenuFromOptions(components, step, char.ID, "favored_enemy")

	case domainCharacter.StepTypeNaturalExplorerSelection:
		embed.Description(step.Description)
		h.buildSelectMenuFromOptions(components, step, char.ID, "natural_explorer")

	case domainCharacter.StepTypeExpertiseSelection:
		embed.Description(step.Description)
		h.buildSelectMenuFromOptions(components, step, char.ID, "expertise")

	case domainCharacter.StepTypeSubclassSelection:
		embed.Description(step.Description)
		h.buildSelectMenuFromOptions(components, step, char.ID, "subclass")

	case domainCharacter.StepTypeCantripsSelection, domainCharacter.StepTypeSpellSelection, domainCharacter.StepTypeSpellbookSelection, domainCharacter.StepTypeSpellsKnownSelection:
		// These require pagination due to large number of options
		embed.Description(step.Description)
		components.PrimaryButton("📜 Browse Spell List", "open_spell_selection", char.ID)

	default:
		// Handle any custom step types
		embed.Description(fmt.Sprintf("Step type %s is not yet implemented", step.Type))
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
		WithComponents(components.Build()...)

	// Don't set ephemeral/update here - let the caller decide
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
		classEmoji := getClassEmoji(char.Class.Key)
		classDetails := []string{
			fmt.Sprintf("**%s Class:** %s", classEmoji, char.Class.Name),
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

			components.SelectMenuWithTarget(
				placeholder,
				"select",
				char.ID,
				options,
				builders.SelectConfig{
					MinValues: step.MinChoices,
					MaxValues: step.MaxChoices,
				},
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

// buildDynamicProgressTracker creates an inline visual progress tracker
func (h *CharacterCreationHandler) buildDynamicProgressTracker(char *domainCharacter.Character, currentStep *domainCharacter.CreationStep, allSteps []domainCharacter.CreationStep) string {
	// Define all possible steps in order
	allPossibleSteps := []struct {
		stepType domainCharacter.CreationStepType
		name     string
		emoji    string
		classes  []string // empty means universal step
	}{
		{domainCharacter.StepTypeRaceSelection, "Race", "👤", nil},
		{domainCharacter.StepTypeClassSelection, "Class", "⚔️", nil},
		{domainCharacter.StepTypeAbilityScores, "Abilities", "🎲", nil},
		{domainCharacter.StepTypeAbilityAssignment, "Assign", "📊", nil},
		{domainCharacter.StepTypeDivineDomainSelection, "Domain", "⛪", []string{"cleric"}},
		{domainCharacter.StepTypeFightingStyleSelection, "Fighting Style", "🛡️", []string{"fighter", "ranger", "paladin"}},
		{domainCharacter.StepTypeFavoredEnemySelection, "Favored Enemy", "🎯", []string{"ranger"}},
		{domainCharacter.StepTypeNaturalExplorerSelection, "Natural Explorer", "🏔️", []string{"ranger"}},
		{domainCharacter.StepTypeSkillSelection, "Skills", "🎯", []string{"wizard", "rogue", "bard"}},
		{domainCharacter.StepTypeLanguageSelection, "Languages", "📚", []string{"wizard", "cleric"}},
		{domainCharacter.StepTypeCantripsSelection, "Cantrips", "✨", []string{"wizard", "cleric", "sorcerer", "warlock", "bard", "druid"}},
		{domainCharacter.StepTypeSpellbookSelection, "Spells", "📖", []string{"wizard"}},
		{domainCharacter.StepTypeSpellsKnownSelection, "Spells", "🌟", []string{"sorcerer", "bard", "warlock"}},
		{domainCharacter.StepTypeProficiencySelection, "Proficiencies", "🛠️", nil},
		{domainCharacter.StepTypeEquipmentSelection, "Equipment", "🎒", nil},
		{domainCharacter.StepTypeCharacterDetails, "Name & Finalize", "📝", nil},
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

	// Build list of steps that apply to this character
	var applicableSteps []struct {
		stepType domainCharacter.CreationStepType
		name     string
		emoji    string
		status   string // "completed", "current", "pending"
	}

	for _, step := range allPossibleSteps {
		// Check if this step applies to the character
		if step.classes != nil && char.Class != nil && !contains(step.classes, char.Class.Key) {
			continue // Skip class-specific steps that don't apply
		}

		// For class-specific steps, also check if it exists in allSteps
		if step.classes != nil {
			found := false
			for _, actualStep := range allSteps {
				if actualStep.Type == step.stepType {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Determine step status
		var status string
		if h.isStepComplete(char, step.stepType) {
			status = "completed"
		} else if currentStep.Type == step.stepType {
			status = "current"
		} else {
			status = "pending"
		}

		applicableSteps = append(applicableSteps, struct {
			stepType domainCharacter.CreationStepType
			name     string
			emoji    string
			status   string
		}{step.stepType, step.name, step.emoji, status})
	}

	// Build inline progress tracker with cleaner visualization
	var parts []string
	completedCount := 0

	for i, step := range applicableSteps {
		var stepDisplay string

		switch step.status {
		case "completed":
			completedCount++
			// Show actual selection instead of generic step name
			switch step.stepType {
			case domainCharacter.StepTypeRaceSelection:
				if char.Race != nil {
					raceEmoji := getRaceEmoji(char.Race.Key)
					stepDisplay = fmt.Sprintf("%s %s", raceEmoji, char.Race.Name)
				}
			case domainCharacter.StepTypeClassSelection:
				if char.Class != nil {
					classEmoji := getClassEmoji(char.Class.Key)
					stepDisplay = fmt.Sprintf("%s %s", classEmoji, char.Class.Name)
				}
			case domainCharacter.StepTypeAbilityScores:
				stepDisplay = "🎲 Rolled"
			case domainCharacter.StepTypeAbilityAssignment:
				stepDisplay = "📊 Assigned"
			case domainCharacter.StepTypeCantripsSelection:
				if char.Spells != nil && len(char.Spells.Cantrips) > 0 {
					stepDisplay = fmt.Sprintf("✨ %d Cantrips", len(char.Spells.Cantrips))
				} else {
					stepDisplay = "✨ Cantrips"
				}
			case domainCharacter.StepTypeSpellbookSelection, domainCharacter.StepTypeSpellsKnownSelection:
				if char.Spells != nil && len(char.Spells.KnownSpells) > 0 {
					stepDisplay = fmt.Sprintf("📖 %d Spells", len(char.Spells.KnownSpells))
				} else {
					stepDisplay = step.emoji + " " + step.name
				}
			default:
				stepDisplay = step.emoji + " " + step.name
			}

		case "current":
			// Highlight current step prominently
			stepDisplay = fmt.Sprintf("**→ %s %s**", step.emoji, step.name)

		case "pending":
			// Show pending steps more subtly
			stepDisplay = step.name
		}

		parts = append(parts, stepDisplay)

		// Add separator between steps
		if i < len(applicableSteps)-1 {
			if step.status == "completed" && applicableSteps[i+1].status == "completed" {
				parts = append(parts, "•") // Dot between completed steps
			} else if step.status == "completed" && applicableSteps[i+1].status == "current" {
				parts = append(parts, " ") // Space before current
			} else if step.status == "current" {
				parts = append(parts, " ") // Space after current
			} else {
				parts = append(parts, "·") // Light separator for pending steps
			}
		}
	}

	// Build the result with cleaner formatting
	totalSteps := len(applicableSteps)
	currentStepNum := completedCount + 1

	// Create the progress line
	result := strings.Join(parts, " ")

	// Add step counter below for clarity
	result += fmt.Sprintf("\n*Step %d of %d*", currentStepNum, totalSteps)

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
	case domainCharacter.StepTypeCantripsSelection:
		// Check for confirmation marker
		for _, feature := range char.Features {
			if feature.Key == "cantrips_selection_confirmed" {
				return true
			}
		}
		return false
	case domainCharacter.StepTypeSpellSelection, domainCharacter.StepTypeSpellbookSelection:
		// Check for confirmation marker
		for _, feature := range char.Features {
			if feature.Key == "spells_selection_confirmed" {
				return true
			}
		}
		return false
	case domainCharacter.StepTypeProficiencySelection:
		// Check for confirmation marker
		for _, feature := range char.Features {
			if feature.Key == "proficiency_selection_complete" {
				return true
			}
		}
		return false
	case domainCharacter.StepTypeEquipmentSelection:
		// Check for confirmation marker
		for _, feature := range char.Features {
			if feature.Key == "equipment_selection_complete" {
				return true
			}
		}
		return false
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
			steps = append(steps,
				domainCharacter.CreationStep{Type: domainCharacter.StepTypeDivineDomainSelection},
				domainCharacter.CreationStep{Type: domainCharacter.StepTypeLanguageSelection},
			)
		case "wizard":
			steps = append(steps,
				domainCharacter.CreationStep{Type: domainCharacter.StepTypeCantripsSelection},
				domainCharacter.CreationStep{Type: domainCharacter.StepTypeSpellbookSelection},
			)
		case "fighter", "paladin":
			steps = append(steps, domainCharacter.CreationStep{Type: domainCharacter.StepTypeFightingStyleSelection})
		case "ranger":
			steps = append(steps,
				domainCharacter.CreationStep{Type: domainCharacter.StepTypeFightingStyleSelection},
				domainCharacter.CreationStep{Type: domainCharacter.StepTypeFavoredEnemySelection},
				domainCharacter.CreationStep{Type: domainCharacter.StepTypeNaturalExplorerSelection},
			)
		case "rogue", "bard":
			steps = append(steps, domainCharacter.CreationStep{Type: domainCharacter.StepTypeSkillSelection})
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

	response.AsUpdate()

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
		"gnome":      "🧚", // Changed from wizard emoji to fairy/gnome emoji
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

// HandleRollAbilityScores handles rolling all ability scores at once
func (h *CharacterCreationHandler) HandleRollAbilityScores(ctx *core.InteractionContext) (*core.HandlerResult, error) {
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

	// Roll all ability scores at once (4d6 drop lowest for each)
	abilities := []string{"Strength", "Dexterity", "Constitution", "Intelligence", "Wisdom", "Charisma"}
	rolledScores := make([]int, 6)
	rollDetails := make([]string, 6)

	for i := range abilities {
		// Roll 4d6 and drop the lowest
		rolls := make([]int, 4)
		for j := 0; j < 4; j++ {
			rolls[j] = 1 + (rand.Intn(6)) // 1-6
		}

		// Sort and drop lowest
		sort.Ints(rolls)
		total := rolls[1] + rolls[2] + rolls[3] // Sum the 3 highest
		rolledScores[i] = total
		rollDetails[i] = fmt.Sprintf("**%d** [%d,%d,%d,~~%d~~]", total, rolls[3], rolls[2], rolls[1], rolls[0])
	}

	// Create AbilityRolls from the rolled scores
	abilityRolls := make([]domainCharacter.AbilityRoll, 6)
	for i, score := range rolledScores {
		abilityRolls[i] = domainCharacter.AbilityRoll{
			ID:    fmt.Sprintf("roll_%d", i+1),
			Value: score,
		}
	}

	// Store the rolled scores in the character
	updateInput := &characterService.UpdateDraftInput{
		AbilityRolls: abilityRolls,
	}

	_, err = h.service.UpdateDraftCharacter(ctx.Context, char.ID, updateInput)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Build response showing all rolled scores
	embed := builders.NewEmbed().
		Title("🎲 Ability Scores Rolled!").
		Color(builders.ColorSuccess).
		Description("Here are your rolled ability scores. You can auto-assign them based on your class or assign them manually.")

	// Add rolled scores in a clean format
	embed.AddField("🎯 Your Rolls", strings.Join(rollDetails, "  •  "), false)

	// Calculate total and average
	total := 0
	for _, score := range rolledScores {
		total += score
	}
	average := float64(total) / 6.0
	embed.AddField("📊 Stats", fmt.Sprintf("**Total**: %d • **Average**: %.1f", total, average), true)

	// Show what racial bonuses will apply
	if char.Race != nil {
		var bonuses []string
		for _, bonus := range char.Race.AbilityBonuses {
			if bonus.Bonus > 0 {
				bonuses = append(bonuses, fmt.Sprintf("%s +%d", getAbilityAbbrev(string(bonus.Attribute)), bonus.Bonus))
			}
		}
		if len(bonuses) > 0 {
			embed.AddField("🎭 Racial Bonuses", strings.Join(bonuses, ", "), true)
		}
	}

	// Build components
	components := builders.NewComponentBuilder(h.customIDBuilder)
	components.SuccessButton("🤖 Auto-Assign & Continue", "auto_assign_and_continue", char.ID)
	components.PrimaryButton("📝 Manual Assignment", "start_manual_assignment", char.ID)
	components.NewRow()
	components.SecondaryButton("📊 Use Standard Array Instead", "standard", char.ID)
	components.DangerButton("🗑️ Start Fresh", "start_fresh", char.ID)

	response := core.NewResponse("").
		WithEmbeds(embed.Build()).
		WithComponents(components.Build()...).
		AsUpdate()

	return &core.HandlerResult{
		Response: response,
	}, nil
}

// HandleAutoAssignAndContinue auto-assigns abilities and continues to next step
func (h *CharacterCreationHandler) HandleAutoAssignAndContinue(ctx *core.InteractionContext) (*core.HandlerResult, error) {
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

	// Create smart auto-assignment
	assignments := h.createSmartAutoAssignment(char)

	// Update character with assignments
	updateInput := &characterService.UpdateDraftInput{
		AbilityAssignments: assignments,
	}

	_, err = h.service.UpdateDraftCharacter(ctx.Context, char.ID, updateInput)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Get next step from flow service
	nextStep, err := h.flowService.GetNextStep(ctx.Context, char.ID)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Re-fetch character to get updated state
	char, err = h.service.GetCharacter(ctx.Context, characterID)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Build response for next step
	response, err := h.buildEnhancedStepResponse(char, nextStep)
	if err != nil {
		return nil, err
	}

	response.AsUpdate()

	return &core.HandlerResult{
		Response: response,
	}, nil
}

// HandleStartManualAssignment shows the manual assignment UI with buttons per ability
func (h *CharacterCreationHandler) HandleStartManualAssignment(ctx *core.InteractionContext) (*core.HandlerResult, error) {
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

	// Start with smart auto-assignment as the base
	assignments := h.createSmartAutoAssignment(char)

	// Store initial assignments
	updateInput := &characterService.UpdateDraftInput{
		AbilityAssignments: assignments,
	}

	_, err = h.service.UpdateDraftCharacter(ctx.Context, char.ID, updateInput)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Show the manual assignment UI
	return h.buildManualAssignmentUI(ctx, char, assignments)
}

// buildManualAssignmentUI creates the manual assignment interface with buttons per ability
func (h *CharacterCreationHandler) buildManualAssignmentUI(ctx *core.InteractionContext, char *domainCharacter.Character, assignments map[string]string) (*core.HandlerResult, error) {
	// Build embed
	embed := builders.NewEmbed().
		Title("📊 Assign Ability Scores").
		Color(builders.ColorPrimary).
		Description("Click on any ability below to change its assigned score. Auto-assignment is applied based on your class - tweak as needed!")

	// Show available scores (clean format)
	var scoreDisplay []string
	for i, roll := range char.AbilityRolls {
		scoreDisplay = append(scoreDisplay, fmt.Sprintf("**%d**", roll.Value))
		if i < len(char.AbilityRolls)-1 {
			scoreDisplay = append(scoreDisplay, "•")
		}
	}
	embed.AddField("🎯 Available Scores", strings.Join(scoreDisplay, " "), false)

	// Show current assignments
	var assignmentDisplay []string
	abilities := []struct{ name, abbrev, emoji string }{
		{"Strength", "STR", "💪"},
		{"Dexterity", "DEX", "🏃"},
		{"Constitution", "CON", "❤️"},
		{"Intelligence", "INT", "🧠"},
		{"Wisdom", "WIS", "👁️"},
		{"Charisma", "CHA", "💬"},
	}

	for _, ability := range abilities {
		if rollID, exists := assignments[ability.abbrev]; exists {
			// Find the score for this roll ID
			for _, roll := range char.AbilityRolls {
				if roll.ID == rollID {
					// Show final score with racial bonuses
					finalScore := roll.Value
					if char.Race != nil {
						for _, bonus := range char.Race.AbilityBonuses {
							if getAbilityAbbrev(string(bonus.Attribute)) == ability.abbrev {
								finalScore += bonus.Bonus
								break
							}
						}
					}

					if finalScore != roll.Value {
						assignmentDisplay = append(assignmentDisplay, fmt.Sprintf("%s **%s**: %d (%d+%d)",
							ability.emoji, ability.abbrev, finalScore, roll.Value, finalScore-roll.Value))
					} else {
						assignmentDisplay = append(assignmentDisplay, fmt.Sprintf("%s **%s**: %d",
							ability.emoji, ability.abbrev, roll.Value))
					}
					break
				}
			}
		}
	}
	embed.AddField("🎯 Current Assignment", strings.Join(assignmentDisplay, "\n"), false)

	// Show class recommendations
	if char.Class != nil {
		primary := getClassPrimaryAbilities(char.Class.Key)
		if primary != "" {
			embed.AddField("💡 Class Recommendations", primary, false)
		}
	}

	// Build assignment interface with buttons for each ability
	components := builders.NewComponentBuilder(h.customIDBuilder)

	// Create buttons for each ability in two rows
	for i, ability := range abilities {
		if i == 3 {
			components.NewRow()
		}

		// Find current assigned score for display
		currentScore := "?"
		if rollID, exists := assignments[ability.abbrev]; exists {
			for _, roll := range char.AbilityRolls {
				if roll.ID == rollID {
					currentScore = fmt.Sprintf("%d", roll.Value)
					break
				}
			}
		}

		components.SecondaryButton(fmt.Sprintf("%s %s: %s", ability.emoji, ability.abbrev, currentScore),
			"assign_to_ability", char.ID, ability.abbrev)
	}

	// Action buttons
	components.NewRow()
	components.SuccessButton("✅ Confirm Assignment", "confirm_ability_assignment", char.ID)
	components.SecondaryButton("🤖 Re-Auto-Assign", "auto_assign_abilities", char.ID)
	components.SecondaryButton("⬅️ Back to Rolling", "roll", char.ID)

	response := core.NewResponse("").
		WithEmbeds(embed.Build()).
		WithComponents(components.Build()...).
		AsUpdate()

	return &core.HandlerResult{
		Response: response,
	}, nil
}

// HandleStandardArray handles using the standard array for ability scores
func (h *CharacterCreationHandler) HandleStandardArray(ctx *core.InteractionContext) (*core.HandlerResult, error) {
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

	// Standard array values
	standardArray := []int{15, 14, 13, 12, 10, 8}

	// Build response showing standard array
	embed := builders.NewEmbed().
		Title("📊 Standard Array Selected").
		Color(builders.ColorPrimary).
		Description("You've chosen to use the standard array. Now assign these scores to your abilities based on your class and playstyle.")

	// Show the standard array
	var arrayDisplay []string
	for _, score := range standardArray {
		arrayDisplay = append(arrayDisplay, fmt.Sprintf("**%d**", score))
	}
	embed.AddField("🎯 Available Scores", strings.Join(arrayDisplay, " • "), false)

	// Show class recommendations
	if char.Class != nil {
		primary := getClassPrimaryAbilities(char.Class.Key)
		if primary != "" {
			embed.AddField("💡 Class Recommendations", primary, false)
		}
	}

	// Show what racial bonuses will apply
	if char.Race != nil {
		var bonuses []string
		for _, bonus := range char.Race.AbilityBonuses {
			if bonus.Bonus > 0 {
				bonuses = append(bonuses, fmt.Sprintf("%s +%d", getAbilityAbbrev(string(bonus.Attribute)), bonus.Bonus))
			}
		}
		if len(bonuses) > 0 {
			embed.AddField("🎭 Racial Bonuses", strings.Join(bonuses, ", "), true)
		}
	}

	// Build components for assignment
	components := builders.NewComponentBuilder(h.customIDBuilder)
	components.PrimaryButton("📝 Assign Scores", "assign_standard_array", char.ID)
	components.SecondaryButton("🎲 Roll Instead", "roll", char.ID)
	components.NewRow()
	components.SecondaryButton("❓ About Standard Array", "ability_info", char.ID)

	response := core.NewResponse("").
		WithEmbeds(embed.Build()).
		WithComponents(components.Build()...).
		AsUpdate()

	return &core.HandlerResult{
		Response: response,
	}, nil
}

// HandleAbilityInfo shows information about ability scores
func (h *CharacterCreationHandler) HandleAbilityInfo(ctx *core.InteractionContext) (*core.HandlerResult, error) {
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

	// Build informational embed
	embed := builders.NewEmbed().
		Title("❓ About Ability Scores").
		Color(builders.ColorInfo).
		Description("Ability scores determine your character's capabilities. Here's what you need to know:")

	// Explain the two methods
	embed.AddField("🎲 Rolling Method",
		"• Roll 4d6 for each ability\n• Drop the lowest die\n• Sum the remaining 3 dice\n• More random, can get very high or low scores\n• Average total: ~72-73", false)

	embed.AddField("📊 Standard Array",
		"• Use predetermined scores: 15, 14, 13, 12, 10, 8\n• Assign them to abilities as you choose\n• More balanced and predictable\n• Total: 72", false)

	// Explain ability score effects
	embed.AddField("💪 What Abilities Do",
		"• **Strength**: Melee attacks, Athletics, carrying capacity\n• **Dexterity**: Ranged attacks, AC, Stealth, Initiative\n• **Constitution**: Hit points, concentration saves\n• **Intelligence**: Spells (Wizard), Investigation, History\n• **Wisdom**: Spells (Cleric/Druid), Perception, Insight\n• **Charisma**: Spells (Bard/Sorcerer), social skills", false)

	// Explain modifiers
	embed.AddField("🔢 Ability Modifiers",
		"• Score 8-9 = -1 modifier\n• Score 10-11 = +0 modifier\n• Score 12-13 = +1 modifier\n• Score 14-15 = +2 modifier\n• Score 16-17 = +3 modifier\n• Score 18-19 = +4 modifier", false)

	// Build components
	components := builders.NewComponentBuilder(h.customIDBuilder)
	components.PrimaryButton("🎲 Roll Ability Scores", "roll", char.ID)
	components.SecondaryButton("📊 Use Standard Array", "standard", char.ID)
	components.NewRow()
	components.SecondaryButton("⬅️ Back", "back_to_abilities", char.ID)

	response := core.NewResponse("").
		WithEmbeds(embed.Build()).
		WithComponents(components.Build()...).
		AsUpdate()

	return &core.HandlerResult{
		Response: response,
	}, nil
}

// HandleConfirmRolledScores handles confirming the rolled ability scores
func (h *CharacterCreationHandler) HandleConfirmRolledScores(ctx *core.InteractionContext) (*core.HandlerResult, error) {
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

	// For now, we'll re-roll the scores since we don't have them stored
	// TODO: In the future, we should store them in the interaction or character
	rolledScores := make([]int, 6)
	for i := range rolledScores {
		// Roll 4d6 and drop the lowest
		rolls := make([]int, 4)
		for j := 0; j < 4; j++ {
			rolls[j] = 1 + (rand.Intn(6)) // 1-6
		}

		// Sort and drop lowest
		sort.Ints(rolls)
		rolledScores[i] = rolls[1] + rolls[2] + rolls[3] // Sum the 3 highest
	}

	// Create AbilityRolls from the rolled scores
	abilityRolls := make([]domainCharacter.AbilityRoll, 6)
	for i, score := range rolledScores {
		abilityRolls[i] = domainCharacter.AbilityRoll{
			ID:    fmt.Sprintf("roll_%d", i+1),
			Value: score,
		}
	}

	// Store the rolled scores in the character
	updateInput := &characterService.UpdateDraftInput{
		AbilityRolls: abilityRolls,
	}

	_, err = h.service.UpdateDraftCharacter(ctx.Context, char.ID, updateInput)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Show assignment UI
	return h.buildAbilityAssignmentUI(ctx, char, abilityRolls)
}

// HandleAssignStandardArray handles assigning the standard array scores
func (h *CharacterCreationHandler) HandleAssignStandardArray(ctx *core.InteractionContext) (*core.HandlerResult, error) {
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

	// Standard array values
	standardArray := []int{15, 14, 13, 12, 10, 8}

	// Create AbilityRolls from the standard array
	abilityRolls := make([]domainCharacter.AbilityRoll, 6)
	for i, score := range standardArray {
		abilityRolls[i] = domainCharacter.AbilityRoll{
			ID:    fmt.Sprintf("standard_%d", i+1),
			Value: score,
		}
	}

	// Store the standard array in the character
	updateInput := &characterService.UpdateDraftInput{
		AbilityRolls: abilityRolls,
	}

	_, err = h.service.UpdateDraftCharacter(ctx.Context, char.ID, updateInput)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Show assignment UI
	return h.buildAbilityAssignmentUI(ctx, char, abilityRolls)
}

// HandleBackToAbilities handles going back to ability selection
func (h *CharacterCreationHandler) HandleBackToAbilities(ctx *core.InteractionContext) (*core.HandlerResult, error) {
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

	// Get current step
	currentStep, err := h.flowService.GetCurrentStep(ctx.Context, char.ID)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Build the response for the ability scores step
	response, err := h.buildEnhancedStepResponse(char, currentStep)
	if err != nil {
		return nil, err
	}

	response.AsUpdate()

	return &core.HandlerResult{
		Response: response,
	}, nil
}

// buildAbilityAssignmentUI builds the UI for assigning ability scores
func (h *CharacterCreationHandler) buildAbilityAssignmentUI(ctx *core.InteractionContext, char *domainCharacter.Character, abilityRolls []domainCharacter.AbilityRoll) (*core.HandlerResult, error) {
	// Build embed
	embed := builders.NewEmbed().
		Title("📊 Assign Ability Scores").
		Color(builders.ColorPrimary).
		Description("Assign your rolled/standard scores to your abilities. Choose wisely based on your class and playstyle!")

	// Show available scores
	var scoreDisplay []string
	for _, roll := range abilityRolls {
		scoreDisplay = append(scoreDisplay, fmt.Sprintf("**%d** (ID: %s)", roll.Value, roll.ID))
	}
	embed.AddField("🎯 Available Scores", strings.Join(scoreDisplay, " • "), false)

	// Show class recommendations
	if char.Class != nil {
		primary := getClassPrimaryAbilities(char.Class.Key)
		if primary != "" {
			embed.AddField("💡 Class Recommendations", primary, false)
		}
	}

	// Show racial bonuses
	if char.Race != nil {
		var bonuses []string
		for _, bonus := range char.Race.AbilityBonuses {
			if bonus.Bonus > 0 {
				bonuses = append(bonuses, fmt.Sprintf("%s +%d", getAbilityAbbrev(string(bonus.Attribute)), bonus.Bonus))
			}
		}
		if len(bonuses) > 0 {
			embed.AddField("🎭 Racial Bonuses", strings.Join(bonuses, ", "), true)
		}
	}

	// Build assignment components
	components := builders.NewComponentBuilder(h.customIDBuilder)

	// For now, we'll use buttons to start the assignment process
	// In a full implementation, we'd need multiple select menus or a more complex UI
	components.PrimaryButton("📝 Start Assignment", "start_ability_assignment", char.ID)
	components.SecondaryButton("← Back to Abilities", "back_to_abilities", char.ID)

	embed.AddField("⚠️ Next Steps", "Click 'Start Assignment' to begin assigning your scores to abilities.", false)

	response := core.NewResponse("").
		WithEmbeds(embed.Build()).
		WithComponents(components.Build()...).
		AsUpdate()

	return &core.HandlerResult{
		Response: response,
	}, nil
}

// HandleStartAbilityAssignment handles starting the ability assignment process
func (h *CharacterCreationHandler) HandleStartAbilityAssignment(ctx *core.InteractionContext) (*core.HandlerResult, error) {
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

	// For now, we'll do a simple automatic assignment based on class
	// In a full implementation, this would be an interactive UI
	if len(char.AbilityRolls) == 0 {
		return nil, core.NewValidationError("No ability rolls found. Please generate scores first.")
	}

	// Sort rolls by value (highest first)
	sortedRolls := make([]domainCharacter.AbilityRoll, len(char.AbilityRolls))
	copy(sortedRolls, char.AbilityRolls)
	sort.Slice(sortedRolls, func(i, j int) bool {
		return sortedRolls[i].Value > sortedRolls[j].Value
	})

	// Create a basic assignment (this is simplified - in practice we'd want user interaction)
	assignments := make(map[string]string)

	// Basic assignment logic based on class
	if char.Class != nil {
		switch char.Class.Key {
		case "wizard":
			assignments["INT"] = sortedRolls[0].ID // Highest to INT
			assignments["DEX"] = sortedRolls[1].ID // Second highest to DEX
			assignments["CON"] = sortedRolls[2].ID // Third highest to CON
			assignments["WIS"] = sortedRolls[3].ID
			assignments["CHA"] = sortedRolls[4].ID
			assignments["STR"] = sortedRolls[5].ID // Lowest to STR
		case "fighter":
			assignments["STR"] = sortedRolls[0].ID // Highest to STR
			assignments["CON"] = sortedRolls[1].ID // Second highest to CON
			assignments["DEX"] = sortedRolls[2].ID // Third highest to DEX
			assignments["WIS"] = sortedRolls[3].ID
			assignments["CHA"] = sortedRolls[4].ID
			assignments["INT"] = sortedRolls[5].ID // Lowest to INT
		case "rogue":
			assignments["DEX"] = sortedRolls[0].ID // Highest to DEX
			assignments["CON"] = sortedRolls[1].ID // Second highest to CON
			assignments["INT"] = sortedRolls[2].ID // Third highest to INT
			assignments["WIS"] = sortedRolls[3].ID
			assignments["CHA"] = sortedRolls[4].ID
			assignments["STR"] = sortedRolls[5].ID // Lowest to STR
		default:
			// Default assignment
			assignments["STR"] = sortedRolls[0].ID
			assignments["DEX"] = sortedRolls[1].ID
			assignments["CON"] = sortedRolls[2].ID
			assignments["INT"] = sortedRolls[3].ID
			assignments["WIS"] = sortedRolls[4].ID
			assignments["CHA"] = sortedRolls[5].ID
		}
	} else {
		// Default assignment without class
		assignments["STR"] = sortedRolls[0].ID
		assignments["DEX"] = sortedRolls[1].ID
		assignments["CON"] = sortedRolls[2].ID
		assignments["INT"] = sortedRolls[3].ID
		assignments["WIS"] = sortedRolls[4].ID
		assignments["CHA"] = sortedRolls[5].ID
	}

	// Update character with assignments
	updateInput := &characterService.UpdateDraftInput{
		AbilityAssignments: assignments,
	}

	_, err = h.service.UpdateDraftCharacter(ctx.Context, char.ID, updateInput)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Get next step from flow service
	nextStep, err := h.flowService.GetNextStep(ctx.Context, char.ID)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Re-fetch character to get updated state
	char, err = h.service.GetCharacter(ctx.Context, characterID)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Build response for next step
	response, err := h.buildEnhancedStepResponse(char, nextStep)
	if err != nil {
		return nil, err
	}

	response.AsUpdate()

	return &core.HandlerResult{
		Response: response,
	}, nil
}

// HandleStartFresh handles starting fresh character creation
func (h *CharacterCreationHandler) HandleStartFresh(ctx *core.InteractionContext) (*core.HandlerResult, error) {
	// Parse custom ID to get character ID
	customID, err := core.ParseCustomID(ctx.GetCustomID())
	if err != nil {
		return nil, core.NewValidationError("Invalid selection")
	}

	characterID := customID.Target

	// Get character to verify ownership
	char, err := h.service.GetCharacter(ctx.Context, characterID)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Verify ownership
	if char.OwnerID != ctx.UserID {
		return nil, core.NewForbiddenError("You can only edit your own characters")
	}

	// Clear ability rolls and assignments from the CURRENT character (not create new one)
	_, err = h.service.UpdateDraftCharacter(ctx.Context, char.ID, &characterService.UpdateDraftInput{
		AbilityRolls:       []domainCharacter.AbilityRoll{},
		AbilityAssignments: map[string]string{},
	})
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Get the updated character
	freshChar, err := h.service.GetCharacter(ctx.Context, char.ID)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Get the current step (should be ability scores since we cleared them)
	currentStep, err := h.flowService.GetCurrentStep(ctx.Context, freshChar.ID)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Build response for the current step
	response, err := h.buildEnhancedStepResponse(freshChar, currentStep)
	if err != nil {
		return nil, err
	}

	response.AsUpdate()

	return &core.HandlerResult{
		Response: response,
	}, nil
}

// HandleRollSingleAbility handles rolling a single ability score
func (h *CharacterCreationHandler) HandleRollSingleAbility(ctx *core.InteractionContext) (*core.HandlerResult, error) {
	// Parse custom ID to get character ID and ability index
	customID, err := core.ParseCustomID(ctx.GetCustomID())
	if err != nil {
		return nil, core.NewValidationError("Invalid selection")
	}

	characterID := customID.Target
	if len(customID.Args) == 0 {
		return nil, core.NewValidationError("No ability index provided")
	}

	abilityIndex := 0
	if len(customID.Args) > 0 {
		// Parse ability index from args
		switch customID.Args[0] {
		case "0":
			abilityIndex = 0
		case "1":
			abilityIndex = 1
		case "2":
			abilityIndex = 2
		case "3":
			abilityIndex = 3
		case "4":
			abilityIndex = 4
		case "5":
			abilityIndex = 5
		}
	}

	// Get character
	char, err := h.service.GetCharacter(ctx.Context, characterID)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Verify ownership
	if char.OwnerID != ctx.UserID {
		return nil, core.NewForbiddenError("You can only edit your own characters")
	}

	abilities := []string{"Strength", "Dexterity", "Constitution", "Intelligence", "Wisdom", "Charisma"}

	// Roll 4d6 drop lowest for this ability
	rolls := make([]int, 4)
	for j := 0; j < 4; j++ {
		rolls[j] = 1 + (rand.Intn(6)) // 1-6
	}

	// Sort and drop lowest
	sort.Ints(rolls)
	total := rolls[1] + rolls[2] + rolls[3] // Sum the 3 highest

	// Ensure we have enough rolls array space
	for len(char.AbilityRolls) <= abilityIndex {
		char.AbilityRolls = append(char.AbilityRolls, domainCharacter.AbilityRoll{})
	}

	// Store this roll
	char.AbilityRolls[abilityIndex] = domainCharacter.AbilityRoll{
		ID:    fmt.Sprintf("roll_%d", abilityIndex+1),
		Value: total,
	}

	// Update character with the new roll
	updateInput := &characterService.UpdateDraftInput{
		AbilityRolls: char.AbilityRolls,
	}

	_, err = h.service.UpdateDraftCharacter(ctx.Context, char.ID, updateInput)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Show the roll result
	embed := builders.NewEmbed().
		Title(fmt.Sprintf("🎲 %s Rolled!", abilities[abilityIndex])).
		Color(builders.ColorSuccess).
		Description(fmt.Sprintf("You rolled **%d** for %s!", total, abilities[abilityIndex]))

	// Show the detailed roll
	embed.AddField("🎯 Roll Details",
		fmt.Sprintf("Rolled: [%d, %d, %d, ~~%d~~]\nKeeping highest 3: **%d**",
			rolls[3], rolls[2], rolls[1], rolls[0], total), false)

	// Build components
	components := builders.NewComponentBuilder(h.customIDBuilder)

	// Continue to next ability or assignment
	if abilityIndex < 5 {
		components.PrimaryButton("➡️ Continue to Next Ability", "continue_rolling", char.ID, fmt.Sprintf("%d", abilityIndex+1))
	} else {
		components.SuccessButton("📝 Assign Abilities", "start_interactive_assignment", char.ID)
	}

	// Option to start fresh if you don't like your rolls
	components.NewRow()
	components.DangerButton("🗑️ Start Fresh", "start_fresh", char.ID)

	response := core.NewResponse("").
		WithEmbeds(embed.Build()).
		WithComponents(components.Build()...).
		AsUpdate()

	return &core.HandlerResult{
		Response: response,
	}, nil
}

// HandleContinueRolling continues to the next ability in the rolling sequence
func (h *CharacterCreationHandler) HandleContinueRolling(ctx *core.InteractionContext) (*core.HandlerResult, error) {
	// Parse custom ID to get character ID and next ability index
	customID, err := core.ParseCustomID(ctx.GetCustomID())
	if err != nil {
		return nil, core.NewValidationError("Invalid selection")
	}

	characterID := customID.Target
	nextIndex := 0
	if len(customID.Args) > 0 {
		// Parse next ability index from args
		switch customID.Args[0] {
		case "1":
			nextIndex = 1
		case "2":
			nextIndex = 2
		case "3":
			nextIndex = 3
		case "4":
			nextIndex = 4
		case "5":
			nextIndex = 5
		}
	}

	// Get character
	char, err := h.service.GetCharacter(ctx.Context, characterID)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Verify ownership
	if char.OwnerID != ctx.UserID {
		return nil, core.NewForbiddenError("You can only edit your own characters")
	}

	// Continue to the next ability
	return h.buildRollingUI(ctx, char, nextIndex)
}

// HandleStartInteractiveAssignment starts the interactive assignment UI
func (h *CharacterCreationHandler) HandleStartInteractiveAssignment(ctx *core.InteractionContext) (*core.HandlerResult, error) {
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

	// Show the interactive assignment UI
	return h.buildInteractiveAssignmentUI(ctx, char)
}

// buildInteractiveAssignmentUI creates an interactive assignment interface
func (h *CharacterCreationHandler) buildInteractiveAssignmentUI(ctx *core.InteractionContext, char *domainCharacter.Character) (*core.HandlerResult, error) {
	// Build embed
	embed := builders.NewEmbed().
		Title("📊 Assign Ability Scores").
		Color(builders.ColorPrimary).
		Description("Great! Now assign your rolled scores to your abilities. You can swap assignments around until you are happy with them.")

	// Show available scores (without IDs!)
	var scoreDisplay []string
	for i, roll := range char.AbilityRolls {
		scoreDisplay = append(scoreDisplay, fmt.Sprintf("**%d**", roll.Value))
		if i < len(char.AbilityRolls)-1 {
			scoreDisplay = append(scoreDisplay, "•")
		}
	}
	embed.AddField("🎯 Your Rolled Scores", strings.Join(scoreDisplay, " "), false)

	// Show current assignments (start with auto-assignment)
	assignments := h.createSmartAutoAssignment(char)

	var assignmentDisplay []string
	abilities := []string{"STR", "DEX", "CON", "INT", "WIS", "CHA"}
	abilityEmojis := []string{"💪", "🏃", "❤️", "🧠", "👁️", "💬"}

	for i, ability := range abilities {
		if rollID, exists := assignments[ability]; exists {
			// Find the score for this roll ID
			for _, roll := range char.AbilityRolls {
				if roll.ID == rollID {
					assignmentDisplay = append(assignmentDisplay, fmt.Sprintf("%s **%s**: %d", abilityEmojis[i], ability, roll.Value))
					break
				}
			}
		}
	}
	embed.AddField("🎯 Current Assignment", strings.Join(assignmentDisplay, "\n"), false)

	// Show what racial bonuses will apply
	if char.Race != nil {
		var bonuses []string
		for _, bonus := range char.Race.AbilityBonuses {
			if bonus.Bonus > 0 {
				bonuses = append(bonuses, fmt.Sprintf("%s +%d", getAbilityAbbrev(string(bonus.Attribute)), bonus.Bonus))
			}
		}
		if len(bonuses) > 0 {
			embed.AddField("🎭 Racial Bonuses (will be applied)", strings.Join(bonuses, ", "), true)
		}
	}

	// Show class recommendations
	if char.Class != nil {
		primary := getClassPrimaryAbilities(char.Class.Key)
		if primary != "" {
			embed.AddField("💡 Class Recommendations", primary, false)
		}
	}

	// Build assignment interface components
	components := builders.NewComponentBuilder(h.customIDBuilder)

	// For now, use auto-assignment with option to proceed
	components.SuccessButton("✅ Confirm Assignment", "confirm_ability_assignment", char.ID)
	components.SecondaryButton("🔄 Auto-Assign", "auto_assign_abilities", char.ID)
	components.SecondaryButton("⬅️ Back to Rolling", "roll", char.ID)

	response := core.NewResponse("").
		WithEmbeds(embed.Build()).
		WithComponents(components.Build()...).
		AsUpdate()

	return &core.HandlerResult{
		Response: response,
	}, nil
}

// createSmartAutoAssignment creates an intelligent auto-assignment based on class and race
func (h *CharacterCreationHandler) createSmartAutoAssignment(char *domainCharacter.Character) map[string]string {
	if len(char.AbilityRolls) != 6 {
		return make(map[string]string)
	}

	// Sort rolls by value (highest first)
	sortedRolls := make([]domainCharacter.AbilityRoll, len(char.AbilityRolls))
	copy(sortedRolls, char.AbilityRolls)
	sort.Slice(sortedRolls, func(i, j int) bool {
		return sortedRolls[i].Value > sortedRolls[j].Value
	})

	assignments := make(map[string]string)

	// Smart assignment based on class
	if char.Class != nil {
		switch char.Class.Key {
		case "wizard":
			assignments["INT"] = sortedRolls[0].ID // Highest to INT
			assignments["DEX"] = sortedRolls[1].ID // Second highest to DEX
			assignments["CON"] = sortedRolls[2].ID // Third highest to CON
			assignments["WIS"] = sortedRolls[3].ID
			assignments["CHA"] = sortedRolls[4].ID
			assignments["STR"] = sortedRolls[5].ID // Lowest to STR
		case "fighter":
			assignments["STR"] = sortedRolls[0].ID // Highest to STR
			assignments["CON"] = sortedRolls[1].ID // Second highest to CON
			assignments["DEX"] = sortedRolls[2].ID // Third highest to DEX
			assignments["WIS"] = sortedRolls[3].ID
			assignments["CHA"] = sortedRolls[4].ID
			assignments["INT"] = sortedRolls[5].ID // Lowest to INT
		case "rogue":
			assignments["DEX"] = sortedRolls[0].ID // Highest to DEX
			assignments["CON"] = sortedRolls[1].ID // Second highest to CON
			assignments["INT"] = sortedRolls[2].ID // Third highest to INT (for skills)
			assignments["WIS"] = sortedRolls[3].ID
			assignments["CHA"] = sortedRolls[4].ID
			assignments["STR"] = sortedRolls[5].ID // Lowest to STR
		case "cleric":
			assignments["WIS"] = sortedRolls[0].ID // Highest to WIS
			assignments["CON"] = sortedRolls[1].ID // Second highest to CON
			assignments["STR"] = sortedRolls[2].ID // Third highest to STR
			assignments["DEX"] = sortedRolls[3].ID
			assignments["CHA"] = sortedRolls[4].ID
			assignments["INT"] = sortedRolls[5].ID // Lowest to INT
		default:
			// Default assignment
			assignments["STR"] = sortedRolls[0].ID
			assignments["DEX"] = sortedRolls[1].ID
			assignments["CON"] = sortedRolls[2].ID
			assignments["INT"] = sortedRolls[3].ID
			assignments["WIS"] = sortedRolls[4].ID
			assignments["CHA"] = sortedRolls[5].ID
		}
	} else {
		// Default assignment without class
		assignments["STR"] = sortedRolls[0].ID
		assignments["DEX"] = sortedRolls[1].ID
		assignments["CON"] = sortedRolls[2].ID
		assignments["INT"] = sortedRolls[3].ID
		assignments["WIS"] = sortedRolls[4].ID
		assignments["CHA"] = sortedRolls[5].ID
	}

	return assignments
}

// HandleConfirmAbilityAssignment confirms the ability assignment and proceeds
func (h *CharacterCreationHandler) HandleConfirmAbilityAssignment(ctx *core.InteractionContext) (*core.HandlerResult, error) {
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

	// Create smart auto-assignment
	assignments := h.createSmartAutoAssignment(char)

	// Update character with assignments
	updateInput := &characterService.UpdateDraftInput{
		AbilityAssignments: assignments,
	}

	_, err = h.service.UpdateDraftCharacter(ctx.Context, char.ID, updateInput)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Get next step from flow service
	nextStep, err := h.flowService.GetNextStep(ctx.Context, char.ID)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Re-fetch character to get updated state
	char, err = h.service.GetCharacter(ctx.Context, characterID)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Build response for next step
	response, err := h.buildEnhancedStepResponse(char, nextStep)
	if err != nil {
		return nil, err
	}

	response.AsUpdate()

	return &core.HandlerResult{
		Response: response,
	}, nil
}

// HandleAssignToAbility shows dropdown for assigning a score to specific ability
func (h *CharacterCreationHandler) HandleAssignToAbility(ctx *core.InteractionContext) (*core.HandlerResult, error) {
	// Parse custom ID to get character ID and ability
	customID, err := core.ParseCustomID(ctx.GetCustomID())
	if err != nil {
		return nil, core.NewValidationError("Invalid selection")
	}

	characterID := customID.Target
	if len(customID.Args) == 0 {
		return nil, core.NewValidationError("No ability specified")
	}
	targetAbility := customID.Args[0]

	// Get character
	char, err := h.service.GetCharacter(ctx.Context, characterID)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Verify ownership
	if char.OwnerID != ctx.UserID {
		return nil, core.NewForbiddenError("You can only edit your own characters")
	}

	// Get current assignments
	var currentAssignments map[string]string
	if len(char.AbilityAssignments) > 0 {
		currentAssignments = char.AbilityAssignments
	} else {
		// If no assignments, create smart auto-assignment
		currentAssignments = h.createSmartAutoAssignment(char)
	}

	abilityNames := map[string]string{
		"STR": "Strength",
		"DEX": "Dexterity",
		"CON": "Constitution",
		"INT": "Intelligence",
		"WIS": "Wisdom",
		"CHA": "Charisma",
	}

	// Build embed for score selection
	embed := builders.NewEmbed().
		Title(fmt.Sprintf("📊 Assign Score to %s", abilityNames[targetAbility])).
		Color(builders.ColorPrimary).
		Description(fmt.Sprintf("Choose which score to assign to **%s**:", abilityNames[targetAbility]))

	// Show current assignments for context
	var assignmentDisplay []string
	abilities := []string{"STR", "DEX", "CON", "INT", "WIS", "CHA"}
	abilityEmojis := map[string]string{
		"STR": "💪", "DEX": "🏃", "CON": "❤️",
		"INT": "🧠", "WIS": "👁️", "CHA": "💬",
	}

	for _, ability := range abilities {
		if rollID, exists := currentAssignments[ability]; exists {
			for _, roll := range char.AbilityRolls {
				if roll.ID == rollID {
					status := ""
					if ability == targetAbility {
						status = " **(current)**"
					}
					assignmentDisplay = append(assignmentDisplay,
						fmt.Sprintf("%s **%s**: %d%s", abilityEmojis[ability], ability, roll.Value, status))
					break
				}
			}
		}
	}
	embed.AddField("🎯 Current Assignments", strings.Join(assignmentDisplay, "\n"), false)

	// Create button options for each available score
	// This allows us to pass both the roll ID and target ability
	components := builders.NewComponentBuilder(h.customIDBuilder)
	components.NewRow()
	count := 0
	for _, roll := range char.AbilityRolls {
		if count > 0 && count%5 == 0 {
			components.NewRow()
		}

		// Find what this roll is currently assigned to
		assignedTo := ""
		for ability, assignedRollID := range currentAssignments {
			if assignedRollID == roll.ID {
				assignedTo = ability
				break
			}
		}

		label := fmt.Sprintf("%d", roll.Value)
		if assignedTo != "" {
			label = fmt.Sprintf("%d (%s)", roll.Value, assignedTo)
		}

		// Pass both roll ID and target ability in args
		components.SecondaryButton(label, "apply_direct_assignment", char.ID, roll.ID, targetAbility)
		count++
	}

	// Add a note about what we're assigning to
	components.NewRow()
	components.DisabledButton(fmt.Sprintf("→ %s", abilityNames[targetAbility]), discordgo.SecondaryButton)

	// Back button
	components.NewRow()
	components.SecondaryButton("⬅️ Back to Assignment", "start_manual_assignment", char.ID)

	response := core.NewResponse("").
		WithEmbeds(embed.Build()).
		WithComponents(components.Build()...).
		AsUpdate()

	return &core.HandlerResult{
		Response: response,
	}, nil
}

// HandleApplyDirectAssignment handles direct assignment with both roll ID and target ability
func (h *CharacterCreationHandler) HandleApplyDirectAssignment(ctx *core.InteractionContext) (*core.HandlerResult, error) {
	// Parse custom ID to get character ID, roll ID, and target ability
	customID, err := core.ParseCustomID(ctx.GetCustomID())
	if err != nil {
		return nil, core.NewValidationError("Invalid selection")
	}

	characterID := customID.Target
	if len(customID.Args) < 2 {
		return nil, core.NewValidationError("Missing roll ID or target ability")
	}
	selectedRollID := customID.Args[0]
	targetAbility := customID.Args[1]

	// Get character
	char, err := h.service.GetCharacter(ctx.Context, characterID)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Verify ownership
	if char.OwnerID != ctx.UserID {
		return nil, core.NewForbiddenError("You can only edit your own characters")
	}

	// Get current assignments
	currentAssignments := make(map[string]string)
	if len(char.AbilityAssignments) > 0 {
		for k, v := range char.AbilityAssignments {
			currentAssignments[k] = v
		}
	} else {
		currentAssignments = h.createSmartAutoAssignment(char)
	}

	// Find what ability currently has the selected roll assigned (for swapping)
	var previousAbility string
	for ability, rollID := range currentAssignments {
		if rollID == selectedRollID {
			previousAbility = ability
			break
		}
	}

	// If target ability already has a roll assigned, we need to swap
	var rollToSwap string
	if currentRollID, exists := currentAssignments[targetAbility]; exists {
		rollToSwap = currentRollID
	}

	// Perform the assignment (with swapping if needed)
	currentAssignments[targetAbility] = selectedRollID

	// If we need to swap, assign the displaced roll to the previous ability
	if rollToSwap != "" && previousAbility != "" && previousAbility != targetAbility {
		currentAssignments[previousAbility] = rollToSwap
	}

	// Store the updated assignments
	updateInput := &characterService.UpdateDraftInput{
		AbilityAssignments: currentAssignments,
	}

	_, err = h.service.UpdateDraftCharacter(ctx.Context, char.ID, updateInput)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Re-fetch character to get updated state
	char, err = h.service.GetCharacter(ctx.Context, characterID)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Return to manual assignment UI with updated assignments
	return h.buildManualAssignmentUI(ctx, char, currentAssignments)
}

// HandleApplyAbilityAssignment applies the selected score assignment (with swapping)
func (h *CharacterCreationHandler) HandleApplyAbilityAssignment(ctx *core.InteractionContext) (*core.HandlerResult, error) {
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

	// Get selected roll ID from the dropdown
	var selectedRollID string
	if ctx.IsComponent() && ctx.Interaction != nil {
		data := ctx.Interaction.MessageComponentData()
		if len(data.Values) > 0 {
			selectedRollID = data.Values[0]
		}
	}

	if selectedRollID == "" {
		return nil, core.NewValidationError("No score selected")
	}

	// Get current assignments
	currentAssignments := make(map[string]string)
	if len(char.AbilityAssignments) > 0 {
		for k, v := range char.AbilityAssignments {
			currentAssignments[k] = v
		}
	} else {
		currentAssignments = h.createSmartAutoAssignment(char)
	}

	// LIMITATION: Due to Discord interaction limitations, we can't easily track
	// which specific ability the user was assigning to. For now, we'll just
	// return to the manual assignment UI and let them try again.
	//
	// In a production system, you'd want to:
	// 1. Store the target ability in Redis with a temporary key
	// 2. Use a more sophisticated custom ID encoding
	// 3. Or redesign the UI to be a single-step selection
	//
	// For this demo, we'll show them a message and return to the assignment UI

	// Store the updated assignments
	updateInput := &characterService.UpdateDraftInput{
		AbilityAssignments: currentAssignments,
	}

	_, err = h.service.UpdateDraftCharacter(ctx.Context, char.ID, updateInput)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Re-fetch character to get updated state
	char, err = h.service.GetCharacter(ctx.Context, characterID)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Return to manual assignment UI with updated assignments
	return h.buildManualAssignmentUI(ctx, char, currentAssignments)
}

// buildRollingUI builds the UI for rolling a specific ability (one-at-a-time rolling)
func (h *CharacterCreationHandler) buildRollingUI(ctx *core.InteractionContext, char *domainCharacter.Character, abilityIndex int) (*core.HandlerResult, error) {
	abilities := []string{"Strength", "Dexterity", "Constitution", "Intelligence", "Wisdom", "Charisma"}

	if abilityIndex >= len(abilities) {
		return nil, core.NewValidationError("Invalid ability index")
	}

	// Build embed for the ability to roll
	embed := builders.NewEmbed().
		Title(fmt.Sprintf("🎲 Roll %s", abilities[abilityIndex])).
		Color(builders.ColorPrimary).
		Description(fmt.Sprintf("Ready to roll for **%s**! Click the button below to roll 4d6 and drop the lowest.", abilities[abilityIndex]))

	// Show progress
	embed.AddField("📊 Progress", fmt.Sprintf("Rolling ability %d of 6", abilityIndex+1), true)

	// Show previously rolled abilities if any
	if len(char.AbilityRolls) > 0 {
		var previousRolls []string
		for i, roll := range char.AbilityRolls {
			if i < abilityIndex {
				previousRolls = append(previousRolls, fmt.Sprintf("**%s**: %d", abilities[i], roll.Value))
			}
		}
		if len(previousRolls) > 0 {
			embed.AddField("✅ Previous Rolls", strings.Join(previousRolls, " • "), false)
		}
	}

	// Build components
	components := builders.NewComponentBuilder(h.customIDBuilder)
	components.PrimaryButton("🎲 Roll!", "roll_single_ability", char.ID, fmt.Sprintf("%d", abilityIndex))
	components.NewRow()
	components.SecondaryButton("📊 Use Standard Array Instead", "standard", char.ID)
	components.DangerButton("🗑️ Start Fresh", "start_fresh", char.ID)

	response := core.NewResponse("").
		WithEmbeds(embed.Build()).
		WithComponents(components.Build()...).
		AsUpdate()

	return &core.HandlerResult{
		Response: response,
	}, nil
}

// HandleOpenSpellSelection opens the paginated spell selection interface
func (h *CharacterCreationHandler) HandleOpenSpellSelection(ctx *core.InteractionContext) (*core.HandlerResult, error) {
	// Parse custom ID
	customID, err := core.ParseCustomID(ctx.GetCustomID())
	if err != nil {
		return nil, core.NewValidationError("Invalid selection")
	}

	characterID := customID.Target

	// Get character to determine class
	char, err := h.service.GetCharacter(ctx.Context, characterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get character: %w", err)
	}

	// Validate character has required fields - never assume, always verify
	if char == nil {
		return nil, core.NewValidationError("Character not found")
	}
	if char.Class == nil {
		return nil, core.NewValidationError("No class selected")
	}
	if char.Class.Key == "" {
		return nil, core.NewValidationError("Invalid class data")
	}

	// Ensure Spells field is initialized for spellcasting classes
	// This handles characters created before SpellList initialization was added
	if isSpellcastingClass(char.Class.Key) {
		if char.Spells == nil {
			char.Spells = &domainCharacter.SpellList{
				Cantrips:       []string{},
				KnownSpells:    []string{},
				PreparedSpells: []string{},
			}
		}
		// Ensure nested fields are initialized - never assume they exist
		if char.Spells.Cantrips == nil {
			char.Spells.Cantrips = []string{}
		}
		if char.Spells.KnownSpells == nil {
			char.Spells.KnownSpells = []string{}
		}
		if char.Spells.PreparedSpells == nil {
			char.Spells.PreparedSpells = []string{}
		}
	}

	// Determine spell level based on the current step
	spellLevel := 1

	// Check if we're selecting cantrips by getting the current step
	ctx2 := context.Background()
	if currentStep, stepErr := h.flowService.GetCurrentStep(ctx2, characterID); stepErr == nil {
		if currentStep.Type == domainCharacter.StepTypeCantripsSelection {
			spellLevel = 0 // Cantrips are level 0
		}
	}

	// Get spells for the class and level
	spellRefs, err := h.service.ListSpellsByClassAndLevel(ctx.Context, char.Class.Key, spellLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to get spells: %w", err)
	}

	// Build paginated spell list
	const spellsPerPage = 10
	page := 0

	response := h.buildSpellSelectionPage(ctx.Context, char, spellRefs, spellLevel, page, spellsPerPage)
	return &core.HandlerResult{
		Response: response,
	}, nil
}

// buildSpellSelectionPage builds a page of the spell selection interface
func (h *CharacterCreationHandler) buildSpellSelectionPage(ctx context.Context, char *domainCharacter.Character, spells []*rulebook.SpellReference, spellLevel, page, perPage int) *core.Response {
	// Validate inputs - never assume, always verify to prevent panics
	if char == nil {
		return core.NewResponse("Error: Character data is missing").AsEphemeral()
	}
	if char.Class == nil {
		return core.NewResponse("Error: Character class is not set").AsEphemeral()
	}
	if char.Class.Name == "" {
		return core.NewResponse("Error: Character class name is missing").AsEphemeral()
	}

	// Calculate pagination
	totalSpells := len(spells)
	totalPages := (totalSpells + perPage - 1) / perPage
	startIdx := page * perPage
	endIdx := startIdx + perPage
	if endIdx > totalSpells {
		endIdx = totalSpells
	}

	// Build embed
	title := fmt.Sprintf("📜 Select Level %d Spells - Page %d/%d", spellLevel, page+1, totalPages)
	maxSpells := h.getMaxSpells(char)
	description := fmt.Sprintf("Choose spells for your %s. You can select up to %d spells.", char.Class.Name, maxSpells)

	if spellLevel == 0 {
		title = fmt.Sprintf("✨ Select Cantrips - Page %d/%d", page+1, totalPages)
		maxCantrips := h.getMaxCantrips(char)
		description = fmt.Sprintf("Choose cantrips for your %s. You can select up to %d cantrips. Cantrips can be cast at will without using spell slots.", char.Class.Name, maxCantrips)
	}

	// Add total spell count to description
	description += fmt.Sprintf("\n\n📚 **Total Available: %d spells across %d pages**", totalSpells, totalPages)

	embed := builders.NewEmbed().
		Title(title).
		Description(description).
		Color(0x9146FF) // Purple for magic

	// Add spell list for current page with descriptions
	var spellList []string
	pageSpells := spells[startIdx:endIdx]
	for i, spellRef := range pageSpells {
		num := startIdx + i + 1

		// Fetch full spell details to get description
		spell, err := h.service.GetSpell(ctx, spellRef.Key)
		if err != nil || spell == nil {
			// Fallback to just name if API call fails
			spellList = append(spellList, fmt.Sprintf("**%d.** %s", num, spellRef.Name))
			continue
		}

		// Truncate description for selection page (keep it concise)
		desc := spell.Description
		if len(desc) > 100 {
			desc = desc[:97] + "..."
		}

		spellEntry := fmt.Sprintf("**%d.** %s\n*%s*", num, spell.Name, desc)
		spellList = append(spellList, spellEntry)
	}

	if len(spellList) > 0 {
		currentPageTitle := fmt.Sprintf("📖 This Page (%d-%d)", startIdx+1, endIdx)
		embed.AddField(currentPageTitle, strings.Join(spellList, "\n"), false)
	}

	// Add preview of what's on other pages
	if totalPages > 1 && page < totalPages-1 {
		// Show a preview of next page
		nextPageStart := (page + 1) * perPage
		nextPageEnd := nextPageStart + 3 // Show first 3 of next page
		if nextPageEnd > totalSpells {
			nextPageEnd = totalSpells
		}

		var nextPagePreview []string
		for i := nextPageStart; i < nextPageEnd && i < len(spells); i++ {
			nextPagePreview = append(nextPagePreview, spells[i].Name)
		}

		if len(nextPagePreview) > 0 {
			remaining := 0
			nextPageActualEnd := (page + 2) * perPage
			if nextPageActualEnd > totalSpells {
				nextPageActualEnd = totalSpells
			}
			remaining = nextPageActualEnd - nextPageStart - len(nextPagePreview)

			previewText := strings.Join(nextPagePreview, ", ")
			if remaining > 0 {
				previewText += fmt.Sprintf(" ... +%d more", remaining)
			}
			embed.AddField("📑 Next Page Preview", previewText, false)
		}
	}

	// Add selected spells tracker
	selectedCount := 0
	maxAllowed := 0
	selectedDisplay := "*None selected yet*"

	// Check if character already has cantrips/spells selected
	// Validate Spells field exists before accessing - prevent panic
	if char.Spells != nil {
		if spellLevel == 0 && char.Spells.Cantrips != nil && len(char.Spells.Cantrips) > 0 {
			selectedCount = len(char.Spells.Cantrips)
			maxAllowed = h.getMaxCantrips(char)
			var selected []string
			for _, cantripKey := range char.Spells.Cantrips {
				// Find the cantrip name
				for _, spell := range spells {
					if spell.Key == cantripKey {
						selected = append(selected, spell.Name)
						break
					}
				}
			}
			if len(selected) > 0 {
				selectedDisplay = strings.Join(selected, ", ")
			}
		} else if spellLevel > 0 && char.Spells.KnownSpells != nil && len(char.Spells.KnownSpells) > 0 {
			selectedCount = len(char.Spells.KnownSpells)
			maxAllowed = h.getMaxSpells(char)
			var selected []string
			for _, spellKey := range char.Spells.KnownSpells {
				// Find the spell name
				for _, spell := range spells {
					if spell.Key == spellKey {
						selected = append(selected, spell.Name)
						break
					}
				}
			}
			if len(selected) > 0 {
				selectedDisplay = strings.Join(selected, ", ")
			}
		} else {
			// No spells selected yet, get max allowed
			if spellLevel == 0 {
				maxAllowed = h.getMaxCantrips(char)
			} else {
				maxAllowed = h.getMaxSpells(char)
			}
		}
	}

	// Add field with count
	fieldTitle := fmt.Sprintf("Selected Spells (%d/%d)", selectedCount, maxAllowed)
	embed.AddField(fieldTitle, selectedDisplay, false)

	// Build components
	components := builders.NewComponentBuilder(h.customIDBuilder)

	// Spell selection menu
	if len(pageSpells) > 0 {
		components.NewRow()

		var options []builders.SelectOption
		for i, spell := range pageSpells {
			num := startIdx + i + 1
			option := builders.SelectOption{
				Label:       spell.Name,
				Value:       spell.Key,
				Description: fmt.Sprintf("#%d", num),
			}

			// Check if this spell is already selected - validate before accessing
			if char.Spells != nil {
				if spellLevel == 0 && char.Spells.Cantrips != nil {
					for _, selected := range char.Spells.Cantrips {
						if selected == spell.Key {
							option.Default = true
							break
						}
					}
				} else if spellLevel > 0 && char.Spells.KnownSpells != nil {
					for _, selected := range char.Spells.KnownSpells {
						if selected == spell.Key {
							option.Default = true
							break
						}
					}
				}
			}

			options = append(options, option)
		}

		// Calculate max selections based on spell type and class
		var maxSelections int

		if spellLevel == 0 {
			// Cantrips - limit to class cantrips known
			maxSelections = h.getMaxCantrips(char)
		} else {
			// Spells - limit to spells known/prepared
			maxSelections = h.getMaxSpells(char)
		}

		// Don't exceed the number of spells on this page
		if maxSelections > len(pageSpells) {
			maxSelections = len(pageSpells)
		}

		// Create select menu with proper custom ID including page number
		placeholder := "Select spells..."
		if spellLevel == 0 {
			placeholder = "Select cantrips..."
		}

		// Use the new method that supports args for page number
		components.SelectMenuWithTargetAndArgs(
			placeholder,
			"select_spell",
			char.ID,
			[]string{fmt.Sprintf("%d", page)},
			options,
			builders.SelectConfig{
				MinValues: 0,
				MaxValues: maxSelections,
			},
		)
	}

	// Navigation buttons
	components.NewRow()

	// Previous button
	if page > 0 {
		components.PrimaryButton("⬅️ Previous", "spell_page", char.ID, fmt.Sprintf("%d", page-1))
	} else {
		components.DisabledButton("⬅️ Previous", discordgo.SecondaryButton)
	}

	// Page indicator
	components.DisabledButton(fmt.Sprintf("Page %d/%d", page+1, totalPages), discordgo.SecondaryButton)

	// Next button
	if page < totalPages-1 {
		components.PrimaryButton("➡️ Next", "spell_page", char.ID, fmt.Sprintf("%d", page+1))
	} else {
		components.DisabledButton("➡️ Next", discordgo.SecondaryButton)
	}

	// Action buttons
	components.NewRow()
	components.SecondaryButton("🔍 View Details", "spell_details", char.ID)
	components.SuccessButton("✅ Confirm Selection", "confirm_spell_selection", char.ID)

	// Back to character creation
	components.NewRow()
	components.DangerButton("❌ Cancel", "cancel_spell_selection", char.ID)

	return core.NewResponse("").
		WithEmbeds(embed.Build()).
		WithComponents(components.Build()...).
		AsUpdate()
}

// HandleSpellPageChange handles pagination
func (h *CharacterCreationHandler) HandleSpellPageChange(ctx *core.InteractionContext) (*core.HandlerResult, error) {
	// Parse custom ID
	customID, err := core.ParseCustomID(ctx.GetCustomID())
	if err != nil {
		return nil, core.NewValidationError("Invalid selection")
	}

	characterID := customID.Target
	pageStr := ""
	if len(customID.Args) > 0 {
		pageStr = customID.Args[0]
	}

	page := 0
	if pageStr != "" {
		if _, scanErr := fmt.Sscanf(pageStr, "%d", &page); scanErr != nil {
			// Default to page 0 on parse error
			page = 0
		}
	}

	// Get character
	char, err := h.service.GetCharacter(ctx.Context, characterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get character: %w", err)
	}

	// Determine spell level based on current step
	spellLevel := 1
	ctx2 := context.Background()
	if currentStep, stepErr := h.flowService.GetCurrentStep(ctx2, characterID); stepErr == nil {
		if currentStep.Type == domainCharacter.StepTypeCantripsSelection {
			spellLevel = 0 // Cantrips are level 0
		}
	}
	spellRefs, err := h.service.ListSpellsByClassAndLevel(ctx.Context, char.Class.Key, spellLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to get spells: %w", err)
	}

	response := h.buildSpellSelectionPage(ctx.Context, char, spellRefs, spellLevel, page, 10)
	return &core.HandlerResult{
		Response: response,
	}, nil
}

// HandleSpellToggle handles spell selection/deselection
func (h *CharacterCreationHandler) HandleSpellToggle(ctx *core.InteractionContext) (*core.HandlerResult, error) {
	// Parse custom ID to get character ID
	customID, err := core.ParseCustomID(ctx.GetCustomID())
	if err != nil {
		return nil, core.NewValidationError("Invalid selection")
	}

	characterID := customID.Target

	// Get character
	char, err := h.service.GetCharacter(ctx.Context, characterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get character: %w", err)
	}

	// Validate inputs - never assume, always verify to prevent panics
	if char == nil {
		return &core.HandlerResult{
			Response: core.NewResponse("Error: Character data is missing").AsEphemeral(),
		}, nil
	}

	// Ensure character has spell list initialized
	if char.Spells == nil {
		char.Spells = &domainCharacter.SpellList{
			Cantrips:       []string{},
			KnownSpells:    []string{},
			PreparedSpells: []string{},
		}
	}

	// Get selected spell keys from the select menu
	interaction := ctx.Interaction
	if interaction.MessageComponentData().ComponentType != discordgo.SelectMenuComponent {
		return nil, core.NewValidationError("Invalid component type")
	}

	selectedSpells := interaction.MessageComponentData().Values

	// Determine if we're selecting cantrips or spells based on current step
	step, err := h.flowService.GetCurrentStep(ctx.Context, char.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current step: %w", err)
	}

	// Parse page number from custom ID if available
	page := 0
	if len(customID.Args) > 0 {
		if p, parseErr := strconv.Atoi(customID.Args[0]); parseErr == nil {
			page = p
		}
	}

	// Get all spells for the class to find spell names and current page spells
	spellLevel := 1 // Default to level 1 spells
	if step.Type == domainCharacter.StepTypeCantripsSelection {
		spellLevel = 0
	}

	allSpells, err := h.service.ListSpellsByClassAndLevel(ctx.Context, char.Class.Key, spellLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to get spell list: %w", err)
	}

	// Build a map of spells on the current page
	const spellsPerPage = 10
	startIdx := page * spellsPerPage
	endIdx := startIdx + spellsPerPage
	if endIdx > len(allSpells) {
		endIdx = len(allSpells)
	}

	currentPageSpellKeys := make(map[string]bool)
	for i := startIdx; i < endIdx; i++ {
		currentPageSpellKeys[allSpells[i].Key] = true
	}

	// Update character spell selection based on step type
	switch step.Type {
	case domainCharacter.StepTypeCantripsSelection:
		// Remove any deselected spells from the current page
		newCantrips := []string{}
		for _, existing := range char.Spells.Cantrips {
			// Keep spells from other pages
			if !currentPageSpellKeys[existing] {
				newCantrips = append(newCantrips, existing)
			}
		}
		// Add the newly selected spells from this page
		newCantrips = append(newCantrips, selectedSpells...)
		char.Spells.Cantrips = newCantrips

	case domainCharacter.StepTypeSpellbookSelection, domainCharacter.StepTypeSpellSelection:
		// Remove any deselected spells from the current page
		newSpells := []string{}
		for _, existing := range char.Spells.KnownSpells {
			// Keep spells from other pages
			if !currentPageSpellKeys[existing] {
				newSpells = append(newSpells, existing)
			}
		}
		// Add the newly selected spells from this page
		newSpells = append(newSpells, selectedSpells...)
		char.Spells.KnownSpells = newSpells

	default:
		return nil, core.NewValidationError("Invalid step for spell selection")
	}

	// Save the character with updated spell selection using UpdateDraftCharacter
	_, err = h.service.UpdateDraftCharacter(ctx.Context, char.ID, &characterService.UpdateDraftInput{
		Spells: char.Spells,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to save character: %w", err)
	}

	// Return to the same page of spell selection with updated selection
	// Extract page from the custom ID if present
	currentPage := 0
	if len(customID.Args) > 0 {
		if p, parseErr := strconv.Atoi(customID.Args[0]); parseErr == nil {
			currentPage = p
		}
	}

	// Get spell refs to rebuild the page
	spellRefs, err := h.service.ListSpellsByClassAndLevel(ctx.Context, char.Class.Key, spellLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to get spells: %w", err)
	}

	// Build and return the updated page
	response := h.buildSpellSelectionPage(ctx.Context, char, spellRefs, spellLevel, currentPage, 10)
	return &core.HandlerResult{
		Response: response,
	}, nil
}

// HandleSpellDetails shows detailed information about selected spells
func (h *CharacterCreationHandler) HandleSpellDetails(ctx *core.InteractionContext) (*core.HandlerResult, error) {
	// Parse custom ID to get character ID
	customID, err := core.ParseCustomID(ctx.GetCustomID())
	if err != nil {
		return nil, core.NewValidationError("Invalid selection")
	}

	characterID := customID.Target

	// Get character
	char, err := h.service.GetCharacter(ctx.Context, characterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get character: %w", err)
	}

	// Validate inputs - never assume, always verify to prevent panics
	if char == nil {
		return &core.HandlerResult{
			Response: core.NewResponse("Error: Character data is missing").AsEphemeral(),
		}, nil
	}

	// Build detailed spell information
	embed := builders.NewEmbed().
		Title("🔍 Selected Spell Details").
		Color(0x9146FF) // Purple for magic

	if char.Spells == nil || (len(char.Spells.Cantrips) == 0 && len(char.Spells.KnownSpells) == 0) {
		embed.Description("No spells selected yet.")
	} else {
		// Fetch spell details from API
		cantripDetails := make(map[string]*rulebook.Spell)
		spellDetails := make(map[string]*rulebook.Spell)

		// Fetch cantrip details
		for _, cantripKey := range char.Spells.Cantrips {
			spell, err := h.service.GetSpell(ctx.Context, cantripKey)
			if err == nil && spell != nil {
				cantripDetails[cantripKey] = spell
			}
		}

		// Fetch spell details
		for _, spellKey := range char.Spells.KnownSpells {
			spell, err := h.service.GetSpell(ctx.Context, spellKey)
			if err == nil && spell != nil {
				spellDetails[spellKey] = spell
			}
		}

		// Show cantrips with details
		if len(char.Spells.Cantrips) > 0 {
			embed.AddField(fmt.Sprintf("✨ Cantrips (%d)", len(char.Spells.Cantrips)), "", false)

			for _, cantripKey := range char.Spells.Cantrips {
				if spell, ok := cantripDetails[cantripKey]; ok {
					// Format spell info
					spellInfo := fmt.Sprintf("**School:** %s\n", spell.School)
					spellInfo += fmt.Sprintf("**Casting Time:** %s\n", spell.CastingTime)
					spellInfo += fmt.Sprintf("**Range:** %s\n", spell.Range)

					spellInfo += fmt.Sprintf("**Description:** %s", spell.Description)

					embed.AddField(fmt.Sprintf("📜 %s", spell.Name), spellInfo, false)
				} else {
					// Fallback if API call failed
					embed.AddField(fmt.Sprintf("📜 %s", cantripKey), "*Details unavailable*", false)
				}
			}
		}

		// Show known spells with details
		if len(char.Spells.KnownSpells) > 0 {
			embed.AddField(fmt.Sprintf("📖 Spells (%d)", len(char.Spells.KnownSpells)), "", false)

			for _, spellKey := range char.Spells.KnownSpells {
				if spell, ok := spellDetails[spellKey]; ok {
					// Format spell info
					spellInfo := fmt.Sprintf("**Level:** %d | **School:** %s\n", spell.Level, spell.School)
					spellInfo += fmt.Sprintf("**Casting Time:** %s | **Duration:** %s\n", spell.CastingTime, spell.Duration)
					spellInfo += fmt.Sprintf("**Range:** %s | **Components:** %s\n", spell.Range, spell.Components)

					spellInfo += fmt.Sprintf("**Effect:** %s", spell.Description)

					embed.AddField(fmt.Sprintf("🌟 %s", spell.Name), spellInfo, false)
				} else {
					// Fallback if API call failed
					embed.AddField(fmt.Sprintf("🌟 %s", spellKey), "*Details unavailable*", false)
				}
			}
		}

		// Add note about spell slots if applicable
		if len(char.Spells.KnownSpells) > 0 {
			embed.Footer("💡 Spell slots determine how many spells you can cast per day")
		}
	}

	// Add back button
	components := builders.NewComponentBuilder(h.customIDBuilder)
	components.SecondaryButton("⬅️ Back to Selection", "open_spell_selection", char.ID)

	response := core.NewResponse("").
		AsUpdate().
		WithEmbeds(embed.Build()).
		WithComponents(components.Build()...)

	return &core.HandlerResult{
		Response: response,
	}, nil
}

// HandleConfirmSpellSelection confirms and saves spell selection
func (h *CharacterCreationHandler) HandleConfirmSpellSelection(ctx *core.InteractionContext) (*core.HandlerResult, error) {
	// Parse custom ID
	customID, err := core.ParseCustomID(ctx.GetCustomID())
	if err != nil {
		return nil, core.NewValidationError("Invalid selection")
	}

	characterID := customID.Target

	// Get character to check current step
	char, err := h.service.GetCharacter(ctx.Context, characterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get character: %w", err)
	}

	// Verify ownership
	if char.OwnerID != ctx.UserID {
		return nil, core.NewForbiddenError("You can only edit your own characters")
	}

	// Get current step to determine spell type
	currentStep, err := h.flowService.GetCurrentStep(ctx.Context, char.ID)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Create step result based on the current step type
	result := &domainCharacter.CreationStepResult{
		StepType: currentStep.Type,
	}

	// Add the spell selections as metadata
	if char.Spells != nil {
		switch currentStep.Type {
		case domainCharacter.StepTypeCantripsSelection:
			result.Selections = char.Spells.Cantrips
		case domainCharacter.StepTypeSpellbookSelection, domainCharacter.StepTypeSpellSelection:
			result.Selections = char.Spells.KnownSpells
		}
	}

	// Process the step result to advance to next step
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

	response.AsUpdate()

	return &core.HandlerResult{
		Response: response,
	}, nil
}

// getMaxCantrips returns the maximum number of cantrips for the character's class
func (h *CharacterCreationHandler) getMaxCantrips(char *domainCharacter.Character) int {
	if char.Class == nil {
		return 0
	}

	// Get cantrips known for this class at level 1
	switch char.Class.Key {
	case "wizard":
		return 3
	case "cleric":
		return 3
	case "druid":
		return 2
	case "bard":
		return 2
	case "sorcerer":
		return 4
	case "warlock":
		return 2
	default:
		return 0
	}
}

// getMaxSpells returns the maximum number of spells for the character's class
func (h *CharacterCreationHandler) getMaxSpells(char *domainCharacter.Character) int {
	if char.Class == nil {
		return 0
	}

	// Get spells known for this class at level 1
	switch char.Class.Key {
	case "wizard":
		return 6 // Wizards start with 6 spells in spellbook
	case "sorcerer":
		return 2 // Sorcerers know 2 spells at level 1
	case "bard":
		return 4 // Bards know 4 spells at level 1
	case "ranger":
		return 0 // Rangers don't get spells until level 2
	case "warlock":
		return 2 // Warlocks know 2 spells at level 1
	case "cleric", "druid", "paladin":
		// These classes prepare spells - use prepared spell calculation
		return h.calculatePreparedSpells(char)
	default:
		return 0
	}
}

// calculatePreparedSpells calculates how many spells a character can prepare
func (h *CharacterCreationHandler) calculatePreparedSpells(char *domainCharacter.Character) int {
	if char.Class == nil {
		return 0
	}

	// Base calculation for most classes: casting ability modifier + level
	// For now, use a simplified version
	level := 1
	if char.Level > 0 {
		level = char.Level
	}

	// Get the primary ability modifier
	modifier := 0
	switch char.Class.Key {
	case "wizard", "artificer":
		if intScore, ok := char.Attributes[shared.AttributeIntelligence]; ok {
			modifier = (intScore.Score - 10) / 2
		}
	case "cleric", "druid", "ranger":
		if wisScore, ok := char.Attributes[shared.AttributeWisdom]; ok {
			modifier = (wisScore.Score - 10) / 2
		}
	}

	// Minimum 1 spell
	prepared := modifier + level
	if prepared < 1 {
		prepared = 1
	}

	return prepared
}

// buildSelectMenuFromOptions creates a select menu for simple single-choice steps
func (h *CharacterCreationHandler) buildSelectMenuFromOptions(components *builders.ComponentBuilder, step *domainCharacter.CreationStep, characterID, targetType string) {
	if len(step.Options) == 0 {
		components.DangerButton("No options available", "error", characterID)
		return
	}

	options := make([]builders.SelectOption, 0, len(step.Options))
	for _, opt := range step.Options {
		options = append(options, builders.SelectOption{
			Label:       opt.Name,
			Value:       opt.Key,
			Description: opt.Description,
		})
	}

	components.SelectMenuWithTarget(
		fmt.Sprintf("Choose %s...", step.Title),
		targetType,
		characterID,
		options,
	)
}

// isSpellcastingClass returns true if the given class key represents a spellcasting class
func isSpellcastingClass(classKey string) bool {
	spellcastingClasses := map[string]bool{
		"wizard":    true,
		"cleric":    true,
		"sorcerer":  true,
		"warlock":   true,
		"bard":      true,
		"druid":     true,
		"paladin":   true,
		"ranger":    true,
		"artificer": true,
	}
	return spellcastingClasses[classKey]
}
