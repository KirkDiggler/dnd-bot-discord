package character

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
)

// FlowBuilderImpl implements the FlowBuilder interface
type FlowBuilderImpl struct {
	dndClient dnd5e.Client
}

// NewFlowBuilder creates a new flow builder
func NewFlowBuilder(dndClient dnd5e.Client) character.FlowBuilder {
	return &FlowBuilderImpl{
		dndClient: dndClient,
	}
}

// BuildFlow creates a complete character creation flow based on character state
func (b *FlowBuilderImpl) BuildFlow(ctx context.Context, char *character.Character) (*character.CreationFlow, error) {
	var steps []character.CreationStep

	// 1. Race Selection (with options if not yet selected)
	if char.Race == nil {
		step := character.CreationStep{
			Type:        character.StepTypeRaceSelection,
			Title:       "Choose Your Race",
			Description: "Select your character's race, which determines starting abilities and traits.",
			Required:    true,
			MinChoices:  1,
			MaxChoices:  1,
		}

		// Fetch race options if client is available
		if b.dndClient != nil {
			raceRefs, err := b.dndClient.ListRaces()
			if err != nil {
				return nil, fmt.Errorf("failed to fetch races: %w", err)
			}

			// Fetch full race details concurrently
			type raceResult struct {
				race *rulebook.Race
				err  error
			}

			results := make(chan raceResult, len(raceRefs))
			for _, raceRef := range raceRefs {
				go func(key string) {
					race, err := b.dndClient.GetRace(key)
					results <- raceResult{race: race, err: err}
				}(raceRef.Key)
			}

			// Collect results
			races := make([]*rulebook.Race, 0, len(raceRefs))
			for i := 0; i < len(raceRefs); i++ {
				result := <-results
				if result.err != nil {
					// Log error but continue with other races
					fmt.Printf("Warning: Failed to fetch race details: %v\n", result.err)
					continue
				}
				if result.race != nil {
					races = append(races, result.race)
				}
			}
			close(results)

			var raceOptions []character.CreationOption
			for _, race := range races {
				// Build comprehensive race details
				var details []string

				// Ability bonuses
				var bonuses []string
				for _, bonus := range race.AbilityBonuses {
					if bonus.Bonus > 0 {
						bonuses = append(bonuses, fmt.Sprintf("%s +%d", bonus.Attribute, bonus.Bonus))
					}
				}
				if len(bonuses) > 0 {
					details = append(details, strings.Join(bonuses, ", "))
				}

				// Speed
				if race.Speed > 0 {
					details = append(details, fmt.Sprintf("%dft", race.Speed))
				}

				// Add key proficiencies if any
				if len(race.StartingProficiencies) > 0 {
					// Just show count to keep it concise
					details = append(details, fmt.Sprintf("%d prof", len(race.StartingProficiencies)))
				}

				// Build description (Discord limits to 100 chars)
				description := "No special traits"
				if len(details) > 0 {
					description = strings.Join(details, " • ")
					if len(description) > 100 {
						description = description[:97] + "..."
					}
				}

				// Store full race data in metadata for later use
				metadata := make(map[string]any)
				metadata["race"] = race
				metadata["bonuses"] = bonuses

				raceOptions = append(raceOptions, character.CreationOption{
					Key:         race.Key,
					Name:        race.Name,
					Description: description,
					Metadata:    metadata,
				})
			}

			// Sort race options alphabetically by name
			sort.Slice(raceOptions, func(i, j int) bool {
				return raceOptions[i].Name < raceOptions[j].Name
			})

			step.Options = raceOptions
		}

		steps = append(steps, step)
	}

	// 2. Class Selection (with options if not yet selected)
	if char.Class == nil {
		step := character.CreationStep{
			Type:        character.StepTypeClassSelection,
			Title:       "Choose Your Class",
			Description: "Select your character's class, which determines abilities, proficiencies, and features.",
			Required:    true,
			MinChoices:  1,
			MaxChoices:  1,
		}

		// Fetch class options if client is available
		if b.dndClient != nil {
			classes, err := b.dndClient.ListClasses()
			if err != nil {
				return nil, fmt.Errorf("failed to fetch classes: %w", err)
			}

			var classOptions []character.CreationOption
			for _, class := range classes {
				// Build a more useful description
				desc := fmt.Sprintf("Hit Die: d%d", class.HitDie)

				// Add primary abilities from class definition
				primaryAbility := class.GetPrimaryAbility()
				if primaryAbility != "" {
					// Convert full names to abbreviations for compact display
					abbrev := primaryAbility
					abbrev = strings.ReplaceAll(abbrev, "Strength", "STR")
					abbrev = strings.ReplaceAll(abbrev, "Dexterity", "DEX")
					abbrev = strings.ReplaceAll(abbrev, "Constitution", "CON")
					abbrev = strings.ReplaceAll(abbrev, "Intelligence", "INT")
					abbrev = strings.ReplaceAll(abbrev, "Wisdom", "WIS")
					abbrev = strings.ReplaceAll(abbrev, "Charisma", "CHA")
					abbrev = strings.ReplaceAll(abbrev, " and ", " & ")
					desc = abbrev + " primary • " + desc
				}

				// Add class-specific features preview
				features := b.getClassFeaturesPreview(class.Key)
				if features != "" {
					desc += " • " + features
				}

				classOptions = append(classOptions, character.CreationOption{
					Key:         class.Key,
					Name:        class.Name,
					Description: desc,
				})
			}
			step.Options = classOptions
		}

		steps = append(steps, step)
	}

	// Only add subsequent steps if race and class are selected
	if char.Race != nil && char.Class != nil {
		// 3. Ability Scores
		steps = append(steps, character.CreationStep{
			Type:        character.StepTypeAbilityScores,
			Title:       "Roll Ability Scores",
			Description: "Generate your character's six ability scores.",
			Required:    true,
		})

		// 4. Ability Assignment (only if scores exist but not assigned)
		if len(char.Attributes) == 0 {
			steps = append(steps, character.CreationStep{
				Type:        character.StepTypeAbilityAssignment,
				Title:       "Assign Ability Scores",
				Description: "Assign your rolled scores to the six abilities.",
				Required:    true,
			})
		}

		// 5. Class-specific features (dynamic based on class)
		if char.Class != nil {
			classSteps, err := b.buildClassSpecificSteps(ctx, char)
			if err != nil {
				return nil, fmt.Errorf("failed to build class steps: %w", err)
			}
			steps = append(steps, classSteps...)
		}

		// Final steps - only add if character has race, class, and abilities
		if len(char.Attributes) > 0 {
			steps = append(steps,
				// 6. Proficiency Selection
				character.CreationStep{
					Type:        character.StepTypeProficiencySelection,
					Title:       "Choose Proficiencies",
					Description: "Select your character's skill and tool proficiencies.",
					Required:    true,
				},
				// 7. Equipment Selection
				character.CreationStep{
					Type:        character.StepTypeEquipmentSelection,
					Title:       "Choose Equipment",
					Description: "Select your starting equipment and gear.",
					Required:    true,
				},
				// 8. Character Details
				character.CreationStep{
					Type:        character.StepTypeCharacterDetails,
					Title:       "Character Details",
					Description: "Choose your character's name and other details.",
					Required:    true,
				},
			)
		}
	}

	return &character.CreationFlow{
		Steps: steps,
	}, nil
}

// buildClassSpecificSteps creates steps specific to the character's class
func (b *FlowBuilderImpl) buildClassSpecificSteps(ctx context.Context, char *character.Character) ([]character.CreationStep, error) {
	var steps []character.CreationStep

	switch char.Class.Key {
	case "cleric":
		steps = append(steps, b.buildClericSteps(ctx, char)...)
	case "fighter":
		steps = append(steps, b.buildFighterSteps(ctx, char)...)
	case "ranger":
		steps = append(steps, b.buildRangerSteps(ctx, char)...)
	case "wizard":
		steps = append(steps, b.buildWizardSteps(ctx, char)...)
		// Add other classes as needed
	}

	return steps, nil
}

// buildClericSteps creates cleric-specific steps
func (b *FlowBuilderImpl) buildClericSteps(ctx context.Context, char *character.Character) []character.CreationStep {
	var steps []character.CreationStep

	// Divine Domain Selection
	domainChoice := rulebook.GetDivineDomainChoice()
	var domainOptions []character.CreationOption
	for _, option := range domainChoice.Options {
		domainOptions = append(domainOptions, character.CreationOption{
			Key:         option.Key,
			Name:        option.Name,
			Description: option.Description,
			Metadata:    option.Metadata,
		})
	}

	steps = append(steps, character.CreationStep{
		Type:        character.StepTypeDivineDomainSelection,
		Title:       "Choose Your Divine Domain",
		Description: "Choose one domain related to your deity. Your choice grants you domain spells and other features.",
		Options:     domainOptions,
		MinChoices:  1,
		MaxChoices:  1,
		Required:    true,
		Context: map[string]any{
			"color":       0xf1c40f, // Gold
			"placeholder": "Make your selection...",
		},
	})

	// Knowledge Domain specific steps
	if b.hasSelectedDomain(char, "knowledge") {
		// Skill Selection
		skillOptions := []character.CreationOption{
			{Key: "arcana", Name: "Arcana", Description: "Your knowledge of magic and magical theory"},
			{Key: "history", Name: "History", Description: "Your knowledge of historical events and lore"},
			{Key: "nature", Name: "Nature", Description: "Your knowledge of the natural world"},
			{Key: "religion", Name: "Religion", Description: "Your knowledge of deities and religious practices"},
		}

		steps = append(steps, character.CreationStep{
			Type:        character.StepTypeSkillSelection,
			Title:       "Choose Knowledge Domain Skills",
			Description: "As a Knowledge domain cleric, choose 2 additional skill proficiencies.",
			Options:     skillOptions,
			MinChoices:  2,
			MaxChoices:  2,
			Required:    true,
			Context: map[string]any{
				"source":      "knowledge_domain",
				"color":       0x9b59b6, // Purple
				"placeholder": "Select 2 skills...",
			},
		})

		// Language Selection (simplified - would need full language list)
		languageOptions := []character.CreationOption{
			{Key: "draconic", Name: "Draconic", Description: "The language of dragons and dragonborn"},
			{Key: "elvish", Name: "Elvish", Description: "The language of elves"},
			{Key: "dwarvish", Name: "Dwarvish", Description: "The language of dwarves"},
			{Key: "celestial", Name: "Celestial", Description: "The language of celestials"},
			{Key: "abyssal", Name: "Abyssal", Description: "The language of demons"},
			{Key: "infernal", Name: "Infernal", Description: "The language of devils"},
		}

		steps = append(steps, character.CreationStep{
			Type:        character.StepTypeLanguageSelection,
			Title:       "Choose Knowledge Domain Languages",
			Description: "As a Knowledge domain cleric, choose 2 additional languages.",
			Options:     languageOptions,
			MinChoices:  2,
			MaxChoices:  2,
			Required:    true,
			Context: map[string]any{
				"source":      "knowledge_domain",
				"color":       0xe67e22, // Orange
				"placeholder": "Select 2 languages...",
			},
		})
	}

	return steps
}

// buildFighterSteps creates fighter-specific steps
func (b *FlowBuilderImpl) buildFighterSteps(ctx context.Context, char *character.Character) []character.CreationStep {
	var steps []character.CreationStep

	// Fighting Style Selection
	fightingStyleChoice := rulebook.GetFightingStyleChoice("fighter")
	var styleOptions []character.CreationOption
	for _, option := range fightingStyleChoice.Options {
		styleOptions = append(styleOptions, character.CreationOption{
			Key:         option.Key,
			Name:        option.Name,
			Description: option.Description,
			Metadata:    option.Metadata,
		})
	}

	steps = append(steps, character.CreationStep{
		Type:        character.StepTypeFightingStyleSelection,
		Title:       "Choose Your Fighting Style",
		Description: "Choose a fighting style that defines your combat technique.",
		Options:     styleOptions,
		MinChoices:  1,
		MaxChoices:  1,
		Required:    true,
		Context: map[string]any{
			"color":       0xe74c3c, // Red
			"placeholder": "Make your selection...",
		},
	})

	return steps
}

// buildRangerSteps creates ranger-specific steps
func (b *FlowBuilderImpl) buildRangerSteps(ctx context.Context, char *character.Character) []character.CreationStep {
	var steps []character.CreationStep

	// Favored Enemy Selection
	favoredEnemyChoice := rulebook.GetFavoredEnemyChoice()
	var enemyOptions []character.CreationOption
	for _, option := range favoredEnemyChoice.Options {
		enemyOptions = append(enemyOptions, character.CreationOption{
			Key:         option.Key,
			Name:        option.Name,
			Description: option.Description,
			Metadata:    option.Metadata,
		})
	}

	steps = append(steps, character.CreationStep{
		Type:        character.StepTypeFavoredEnemySelection,
		Title:       "Choose Your Favored Enemy",
		Description: "Choose the type of creature you have dedicated yourself to hunting.",
		Options:     enemyOptions,
		MinChoices:  1,
		MaxChoices:  1,
		Required:    true,
	})

	// Natural Explorer Selection
	naturalExplorerChoice := rulebook.GetNaturalExplorerChoice()
	var terrainOptions []character.CreationOption
	for _, option := range naturalExplorerChoice.Options {
		terrainOptions = append(terrainOptions, character.CreationOption{
			Key:         option.Key,
			Name:        option.Name,
			Description: option.Description,
			Metadata:    option.Metadata,
		})
	}

	steps = append(steps, character.CreationStep{
		Type:        character.StepTypeNaturalExplorerSelection,
		Title:       "Choose Your Favored Terrain",
		Description: "Choose the terrain where you feel most at home.",
		Options:     terrainOptions,
		MinChoices:  1,
		MaxChoices:  1,
		Required:    true,
	})

	return steps
}

// hasSelectedDomain checks if the character has selected a specific divine domain
func (b *FlowBuilderImpl) hasSelectedDomain(char *character.Character, domain string) bool {
	for _, feature := range char.Features {
		if feature.Key == "divine_domain" && feature.Metadata != nil {
			if selectedDomain, ok := feature.Metadata["domain"].(string); ok {
				return selectedDomain == domain
			}
		}
	}
	return false
}

// getClassFeaturesPreview returns a short preview of class features and choices
func (b *FlowBuilderImpl) getClassFeaturesPreview(classKey string) string {
	switch classKey {
	case "barbarian":
		return "Rage, Unarmored Defense"
	case "bard":
		return "Bardic Inspiration, Spellcasting, 3 skills"
	case "cleric":
		return "Choose Domain, Spellcasting, 2 skills"
	case "druid":
		return "Druidic, Spellcasting, 2 skills"
	case "fighter":
		return "Choose Fighting Style, Second Wind, 2 skills"
	case "monk":
		return "Martial Arts, Unarmored Defense, 2 skills"
	case "paladin":
		return "Divine Sense, Lay on Hands, 2 skills"
	case "ranger":
		return "Choose Favored Enemy & Terrain, 3 skills"
	case "rogue":
		return "Sneak Attack, Expertise, 4 skills"
	case "sorcerer":
		return "Sorcerous Origin, Spellcasting, 2 skills"
	case "warlock":
		return "Otherworldly Patron, Pact Magic, 2 skills"
	case "wizard":
		return "Arcane Recovery, Spellcasting, 2 skills"
	default:
		return ""
	}
}
