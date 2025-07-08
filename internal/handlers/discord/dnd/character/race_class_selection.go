package character

import (
	"context"
	"fmt"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services"
	characterService "github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/bwmarrin/discordgo"
)

// RaceClassSelectionHandler handles combined race and class selection
type RaceClassSelectionHandler struct {
	services *services.Provider
}

// NewRaceClassSelectionHandler creates a new handler
func NewRaceClassSelectionHandler(serviceProvider *services.Provider) *RaceClassSelectionHandler {
	return &RaceClassSelectionHandler{
		services: serviceProvider,
	}
}

// Handle shows the race/class selection interface
func (h *RaceClassSelectionHandler) Handle(s *discordgo.Session, i *discordgo.InteractionCreate, characterID string) error {
	ctx := context.Background()

	// Get the character
	char, err := h.services.CharacterService.GetByID(characterID)
	if err != nil {
		return respondWithError(s, i, "Failed to get character")
	}

	// Get the flow steps to find options
	progressSteps, err := h.services.CreationFlowService.GetProgressSteps(ctx, characterID)
	if err != nil {
		return respondWithError(s, i, "Failed to get character creation steps")
	}

	// Find race and class steps
	var raceStep, classStep *character.CreationStep
	for _, stepInfo := range progressSteps {
		switch stepInfo.Step.Type {
		case character.StepTypeRaceSelection:
			raceStep = &stepInfo.Step
		case character.StepTypeClassSelection:
			classStep = &stepInfo.Step
		}
	}

	// Build embed
	embed := &discordgo.MessageEmbed{
		Title:       "Create Your Character",
		Description: "Select your race and class. You can change your selections before continuing.",
		Color:       0x5865F2,
		Fields:      []*discordgo.MessageEmbedField{},
	}

	// Add creation steps preview
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name: "Character Creation Steps",
		Value: "1️⃣ **Race & Class** (current)\n" +
			"2️⃣ **Ability Scores** - Roll your stats\n" +
			"3️⃣ **Class Features** - Choose fighting style, domain, etc.\n" +
			"4️⃣ **Proficiencies** - Pick skills and tools\n" +
			"5️⃣ **Equipment** - Select starting gear\n" +
			"6️⃣ **Character Details** - Name and finalize",
		Inline: false,
	})

	// Add instructions
	if char.Race == nil {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "🏁 Getting Started",
			Value:  "First, select your character's race from the dropdown below. Each race provides different ability bonuses and traits.",
			Inline: false,
		})
	} else if char.Class == nil {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "⚔️ Choose Your Path",
			Value:  "Now select your class. This determines your abilities, combat style, and role in the party.",
			Inline: false,
		})
	}

	// Show selected race details
	if char.Race != nil {
		race, err := h.services.CharacterService.GetRace(ctx, char.Race.Key)
		if err == nil {
			details := []string{
				fmt.Sprintf("**Speed:** %d ft", race.Speed),
			}

			// Ability bonuses
			if len(race.AbilityBonuses) > 0 {
				var bonuses []string
				for _, bonus := range race.AbilityBonuses {
					if bonus.Bonus > 0 {
						bonuses = append(bonuses, fmt.Sprintf("%s +%d", bonus.Attribute, bonus.Bonus))
					}
				}
				details = append(details, fmt.Sprintf("**Ability Bonuses:** %s", strings.Join(bonuses, ", ")))
			}

			// Size would go here if available in the API

			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   fmt.Sprintf("Selected Race: %s", race.Name),
				Value:  strings.Join(details, "\n"),
				Inline: false,
			})
		}
	}

	// Show selected class details
	if char.Class != nil {
		class, err := h.services.CharacterService.GetClass(ctx, char.Class.Key)
		if err == nil {
			details := []string{
				fmt.Sprintf("**Hit Die:** d%d", class.HitDie),
				fmt.Sprintf("**Hit Points at 1st Level:** %d + CON modifier", class.HitDie),
			}

			// Primary ability
			primaryAbility := class.GetPrimaryAbility()
			if primaryAbility != "" {
				details = append(details, fmt.Sprintf("**Primary Ability:** %s", primaryAbility))
			}

			// Class features
			features := h.getClassFeaturesList(class.Key)
			if features != "" {
				details = append(details, fmt.Sprintf("**Starting Features:** %s", features))
			}

			// Proficiencies
			if len(class.Proficiencies) > 0 {
				var profs []string
				for _, prof := range class.Proficiencies {
					profs = append(profs, prof.Name)
				}
				if len(profs) > 0 {
					details = append(details, fmt.Sprintf("**Proficiencies:** %s", strings.Join(profs[:minInt(3, len(profs))], ", ")+"..."))
				}
			}

			// Proficiency choices
			if len(class.ProficiencyChoices) > 0 {
				for _, choice := range class.ProficiencyChoices {
					if choice.Count > 0 {
						details = append(details, fmt.Sprintf("**Choose %d %s**", choice.Count, choice.Name))
					}
				}
			}

			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   fmt.Sprintf("Selected Class: %s", class.Name),
				Value:  strings.Join(details, "\n"),
				Inline: false,
			})
		}
	}

	// Build components
	var components []discordgo.MessageComponent

	// Race selection dropdown
	if raceStep != nil && len(raceStep.Options) > 0 {
		var raceOptions []discordgo.SelectMenuOption
		for _, opt := range raceStep.Options {
			selected := char.Race != nil && char.Race.Key == opt.Key
			desc := opt.Description
			if len(desc) > 100 {
				desc = desc[:97] + "..."
			}
			raceOptions = append(raceOptions, discordgo.SelectMenuOption{
				Label:       opt.Name,
				Value:       opt.Key,
				Description: desc,
				Default:     selected,
			})
		}

		components = append(components, discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    fmt.Sprintf("creation_flow:%s:race_select", characterID),
					Placeholder: "Select your race...",
					Options:     raceOptions,
				},
			},
		})
	}

	// Class selection dropdown (only if race is selected)
	if char.Race != nil && classStep != nil && len(classStep.Options) > 0 {
		var classOptions []discordgo.SelectMenuOption
		for _, opt := range classStep.Options {
			selected := char.Class != nil && char.Class.Key == opt.Key
			desc := opt.Description
			if len(desc) > 100 {
				desc = desc[:97] + "..."
			}
			classOptions = append(classOptions, discordgo.SelectMenuOption{
				Label:       opt.Name,
				Value:       opt.Key,
				Description: desc,
				Default:     selected,
			})
		}

		components = append(components, discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    fmt.Sprintf("creation_flow:%s:class_select", characterID),
					Placeholder: "Select your class...",
					Options:     classOptions,
				},
			},
		})
	}

	// Continue button (only if both race and class are selected)
	if char.Race != nil && char.Class != nil {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Ready to Continue?",
			Value:  "You can still change your race or class selections using the dropdowns above.",
			Inline: false,
		})

		components = append(components, discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Continue to Ability Scores",
					Style:    discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("creation_flow:%s:continue", characterID),
					Emoji:    &discordgo.ComponentEmoji{Name: "➡️"},
				},
			},
		})
	}

	// Send response
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
			Flags:      discordgo.MessageFlagsEphemeral,
		},
	})
}

// HandleRaceSelect handles race selection
func (h *RaceClassSelectionHandler) HandleRaceSelect(s *discordgo.Session, i *discordgo.InteractionCreate, characterID, raceKey string) error {
	ctx := context.Background()

	// Update the character with selected race
	updateInput := &characterService.UpdateDraftInput{
		RaceKey: &raceKey,
	}

	_, err := h.services.CharacterService.UpdateDraftCharacter(ctx, characterID, updateInput)
	if err != nil {
		return respondWithError(s, i, "Failed to update character race")
	}

	// Re-render the selection UI
	return h.Handle(s, i, characterID)
}

// HandleClassSelect handles class selection
func (h *RaceClassSelectionHandler) HandleClassSelect(s *discordgo.Session, i *discordgo.InteractionCreate, characterID, classKey string) error {
	ctx := context.Background()

	// Update the character with selected class
	updateInput := &characterService.UpdateDraftInput{
		ClassKey: &classKey,
	}

	_, err := h.services.CharacterService.UpdateDraftCharacter(ctx, characterID, updateInput)
	if err != nil {
		return respondWithError(s, i, "Failed to update character class")
	}

	// Re-render the selection UI
	return h.Handle(s, i, characterID)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// getClassFeaturesList returns a formatted list of class features
func (h *RaceClassSelectionHandler) getClassFeaturesList(classKey string) string {
	switch classKey {
	case "barbarian":
		return "Rage (2/day), Unarmored Defense"
	case "bard":
		return "Bardic Inspiration (d6), Spellcasting"
	case "cleric":
		return "Divine Domain, Spellcasting, Channel Divinity"
	case "druid":
		return "Druidic language, Spellcasting"
	case "fighter":
		return "Fighting Style, Second Wind, Action Surge"
	case "monk":
		return "Martial Arts (1d4), Unarmored Defense"
	case "paladin":
		return "Divine Sense, Lay on Hands"
	case "ranger":
		return "Favored Enemy, Natural Explorer"
	case "rogue":
		return "Sneak Attack (1d6), Thieves' Cant, Expertise"
	case "sorcerer":
		return "Sorcerous Origin, Spellcasting, Font of Magic"
	case "warlock":
		return "Otherworldly Patron, Pact Magic"
	case "wizard":
		return "Spellcasting, Arcane Recovery, Spell Book"
	default:
		return ""
	}
}
