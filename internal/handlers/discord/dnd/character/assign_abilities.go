package character

import (
	"context"
	"fmt"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"sort"
	"strings"

	characterService "github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/bwmarrin/discordgo"
)

// AssignAbilitiesHandler handles assigning rolled scores to abilities
type AssignAbilitiesHandler struct {
	characterService characterService.Service
}

// AssignAbilitiesHandlerConfig holds configuration
type AssignAbilitiesHandlerConfig struct {
	CharacterService characterService.Service
}

// NewAssignAbilitiesHandler creates a new handler
func NewAssignAbilitiesHandler(cfg *AssignAbilitiesHandlerConfig) *AssignAbilitiesHandler {
	return &AssignAbilitiesHandler{
		characterService: cfg.CharacterService,
	}
}

// AssignAbilitiesRequest represents the request
type AssignAbilitiesRequest struct {
	Session     *discordgo.Session
	Interaction *discordgo.InteractionCreate
	RaceKey     string
	ClassKey    string
	AutoAssign  bool // Whether to auto-assign abilities
}

// Handle processes ability assignment
func (h *AssignAbilitiesHandler) Handle(req *AssignAbilitiesRequest) error {
	// Update the message first
	err := req.Session.InteractionRespond(req.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content: "Loading...",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to acknowledge interaction: %w", err)
	}

	// Get the draft character
	draftChar, err := h.characterService.GetOrCreateDraftCharacter(
		context.Background(),
		req.Interaction.Member.User.ID,
		req.Interaction.GuildID,
	)
	if err != nil {
		return h.respondWithError(req, "Failed to get draft character.")
	}

	// Make sure we have rolls to work with
	if len(draftChar.AbilityRolls) == 0 {
		return h.respondWithError(req, "No ability rolls found. Please roll abilities first.")
	}

	// Handle auto-assign
	if req.AutoAssign {
		class, getClassErr := h.characterService.GetClass(context.Background(), req.ClassKey)
		if getClassErr == nil {
			assignments := h.autoAssignAbilitiesWithIDs(class.Name, draftChar.AbilityRolls)

			// Save auto-assigned scores immediately
			_, updateErr := h.characterService.UpdateDraftCharacter(
				context.Background(),
				draftChar.ID,
				&characterService.UpdateDraftInput{
					AbilityAssignments: assignments,
				},
			)
			if updateErr != nil {
				return h.respondWithError(req, "Failed to save auto-assignments.")
			}

			// Reload character to get updated state
			draftChar, err = h.characterService.GetCharacter(context.Background(), draftChar.ID)
			if err != nil {
				return h.respondWithError(req, "Failed to reload character.")
			}
		}
	}

	// Handle manual assignment from dropdown
	if req.Interaction.Type == discordgo.InteractionMessageComponent {
		data := req.Interaction.MessageComponentData()
		if strings.HasPrefix(data.CustomID, "character_create:assign_ability:") {
			parts := strings.Split(data.CustomID, ":")
			if len(parts) >= 5 && len(data.Values) > 0 {
				ability := parts[4]
				rollID := data.Values[0]

				// Initialize assignments if nil
				if draftChar.AbilityAssignments == nil {
					draftChar.AbilityAssignments = make(map[string]string)
				}

				// If rollID is "0", remove the assignment
				if rollID == "0" {
					delete(draftChar.AbilityAssignments, ability)
				} else {
					// Remove this roll from any other ability first
					for a, rID := range draftChar.AbilityAssignments {
						if rID == rollID && a != ability {
							delete(draftChar.AbilityAssignments, a)
						}
					}
					// Assign the roll to this ability
					draftChar.AbilityAssignments[ability] = rollID
				}

				// Save the assignment
				_, err = h.characterService.UpdateDraftCharacter(
					context.Background(),
					draftChar.ID,
					&characterService.UpdateDraftInput{
						AbilityAssignments: draftChar.AbilityAssignments,
					},
				)
				if err != nil {
					return h.respondWithError(req, "Failed to save assignment.")
				}

				// Reload character to get updated state
				draftChar, err = h.characterService.GetCharacter(context.Background(), draftChar.ID)
				if err != nil {
					return h.respondWithError(req, "Failed to reload character.")
				}
			}
		}
	}

	// Get race and class for display
	race, err := h.characterService.GetRace(context.Background(), req.RaceKey)
	if err != nil {
		return h.respondWithError(req, "Failed to fetch race details.")
	}

	class, err := h.characterService.GetClass(context.Background(), req.ClassKey)
	if err != nil {
		return h.respondWithError(req, "Failed to fetch class details.")
	}

	// Build the UI
	return h.buildAssignmentUI(req, draftChar, race, class)
}

// getRacialBonus gets the racial bonus for a specific ability
func (h *AssignAbilitiesHandler) getRacialBonus(race *rulebook.Race, ability string) int {
	// Convert ability string to Attribute type
	var attr shared.Attribute
	switch ability {
	case "STR":
		attr = shared.AttributeStrength
	case "DEX":
		attr = shared.AttributeDexterity
	case "CON":
		attr = shared.AttributeConstitution
	case "INT":
		attr = shared.AttributeIntelligence
	case "WIS":
		attr = shared.AttributeWisdom
	case "CHA":
		attr = shared.AttributeCharisma
	}

	for _, bonus := range race.AbilityBonuses {
		if bonus.Attribute == attr {
			return bonus.Bonus
		}
	}
	return 0
}

// getClassRecommendations returns ability score recommendations for a class
func (h *AssignAbilitiesHandler) getClassRecommendations(className string) string {
	recommendations := map[string]string{
		"Fighter":   "High STR or DEX, then CON",
		"Wizard":    "High INT, then CON",
		"Cleric":    "High WIS, then CON",
		"Rogue":     "High DEX, then CON/INT",
		"Ranger":    "High DEX, then WIS",
		"Barbarian": "High STR and CON",
		"Bard":      "High CHA, then DEX",
		"Druid":     "High WIS, then CON",
		"Monk":      "High DEX and WIS",
		"Paladin":   "High STR and CHA",
		"Sorcerer":  "High CHA, then CON",
		"Warlock":   "High CHA, then CON",
	}

	if rec, ok := recommendations[className]; ok {
		return rec
	}

	return "Assign highest to primary ability"
}

// respondWithError updates the message with an error
func (h *AssignAbilitiesHandler) respondWithError(req *AssignAbilitiesRequest, message string) error {
	content := fmt.Sprintf("❌ %s", message)
	_, err := req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
		Content: &content,
		Embeds:  &[]*discordgo.MessageEmbed{},
	})
	return err
}

// autoAssignAbilitiesWithIDs assigns rolls to abilities based on class priorities
func (h *AssignAbilitiesHandler) autoAssignAbilitiesWithIDs(className string, rolls []character.AbilityRoll) map[string]string {
	// Sort rolls by value (highest to lowest)
	sortedRolls := make([]character.AbilityRoll, len(rolls))
	copy(sortedRolls, rolls)
	sort.Slice(sortedRolls, func(i, j int) bool {
		return sortedRolls[i].Value > sortedRolls[j].Value
	})

	// Define priority order for each class
	classPriorities := map[string][]string{
		"Fighter":   {"STR", "CON", "DEX", "WIS", "CHA", "INT"},
		"Wizard":    {"INT", "CON", "DEX", "WIS", "CHA", "STR"},
		"Cleric":    {"WIS", "CON", "STR", "CHA", "DEX", "INT"},
		"Rogue":     {"DEX", "CON", "INT", "WIS", "CHA", "STR"},
		"Ranger":    {"DEX", "WIS", "CON", "STR", "INT", "CHA"},
		"Barbarian": {"STR", "CON", "DEX", "WIS", "CHA", "INT"},
		"Bard":      {"CHA", "DEX", "CON", "WIS", "INT", "STR"},
		"Druid":     {"WIS", "CON", "DEX", "INT", "CHA", "STR"},
		"Monk":      {"DEX", "WIS", "CON", "STR", "INT", "CHA"},
		"Paladin":   {"STR", "CHA", "CON", "WIS", "DEX", "INT"},
		"Sorcerer":  {"CHA", "CON", "DEX", "WIS", "INT", "STR"},
		"Warlock":   {"CHA", "CON", "DEX", "WIS", "INT", "STR"},
	}

	// Get priority order for the class (default to balanced if not found)
	priorities, ok := classPriorities[className]
	if !ok {
		priorities = []string{"STR", "DEX", "CON", "INT", "WIS", "CHA"}
	}

	// Assign roll IDs to abilities based on priority
	assignments := make(map[string]string)
	for i, ability := range priorities {
		if i < len(sortedRolls) {
			assignments[ability] = sortedRolls[i].ID
		}
	}

	return assignments
}

// buildAssignmentUI builds the assignment interface
func (h *AssignAbilitiesHandler) buildAssignmentUI(req *AssignAbilitiesRequest, char *character.Character, race *rulebook.Race, class *rulebook.Class) error {
	// Create a map of roll ID to value for easy lookup
	rollValues := make(map[string]int)
	for _, roll := range char.AbilityRolls {
		rollValues[roll.ID] = roll.Value
	}

	// Create embed
	embed := &discordgo.MessageEmbed{
		Title:       "Create New Character - Assign Abilities",
		Description: fmt.Sprintf("**Race:** %s\n**Class:** %s\n\nAssign your rolled scores to each ability.\n\n💡 **Tip:** Numbers in [brackets] are your d20 roll modifiers", race.Name, class.Name),
		Color:       0x5865F2,
		Fields:      []*discordgo.MessageEmbedField{},
	}

	// Show rolled values
	rollStrings := []string{}
	usedRolls := make(map[string]bool)
	if char.AbilityAssignments != nil {
		for _, rollID := range char.AbilityAssignments {
			usedRolls[rollID] = true
		}
	}

	for _, roll := range char.AbilityRolls {
		status := ""
		if usedRolls[roll.ID] {
			status = " ✓"
		}
		rollStrings = append(rollStrings, fmt.Sprintf("**%d**%s", roll.Value, status))
	}

	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "🎲 Your Rolls",
		Value:  strings.Join(rollStrings, " • "),
		Inline: false,
	})

	// Show current assignments
	abilities := []string{"STR", "DEX", "CON", "INT", "WIS", "CHA"}
	physicalStrings := []string{}
	mentalStrings := []string{}

	for i, ability := range abilities {
		var line string
		if char.AbilityAssignments != nil {
			if rollID, assigned := char.AbilityAssignments[ability]; assigned {
				if score, exists := rollValues[rollID]; exists {
					racialBonus := h.getRacialBonus(race, ability)
					total := score + racialBonus
					modifier := (total - 10) / 2
					modStr := fmt.Sprintf("%+d", modifier)

					line = fmt.Sprintf("**%s:** %d", ability, score)
					if racialBonus > 0 {
						line += fmt.Sprintf(" (+%d) = %d [%s]", racialBonus, total, modStr)
					} else {
						line += fmt.Sprintf(" = %d [%s]", total, modStr)
					}
				} else {
					line = fmt.Sprintf("**%s:** _Error_", ability)
				}
			} else {
				line = fmt.Sprintf("**%s:** _Not assigned_", ability)
				racialBonus := h.getRacialBonus(race, ability)
				if racialBonus > 0 {
					line += fmt.Sprintf(" (+%d racial)", racialBonus)
				}
			}
		} else {
			line = fmt.Sprintf("**%s:** _Not assigned_", ability)
			racialBonus := h.getRacialBonus(race, ability)
			if racialBonus > 0 {
				line += fmt.Sprintf(" (+%d racial)", racialBonus)
			}
		}

		// Split into physical (STR, DEX, CON) and mental (INT, WIS, CHA)
		if i < 3 {
			physicalStrings = append(physicalStrings, line)
		} else {
			mentalStrings = append(mentalStrings, line)
		}
	}

	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "💪 Physical",
		Value:  strings.Join(physicalStrings, "\n"),
		Inline: true,
	}, &discordgo.MessageEmbedField{
		Name:   "🧠 Mental",
		Value:  strings.Join(mentalStrings, "\n"),
		Inline: true,
	})

	// Class recommendations
	recommendations := h.getClassRecommendations(class.Name)
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("💡 %s Tips", class.Name),
		Value:  recommendations,
		Inline: true,
	}, &discordgo.MessageEmbedField{
		Name:   "Progress",
		Value:  "✅ Step 1: Race\n✅ Step 2: Class\n⏳ Step 3: Abilities\n⏳ Step 4: Details",
		Inline: false,
	})

	// Create components
	components := h.buildComponents(req, char, abilities)

	embed.Footer = &discordgo.MessageEmbedFooter{
		Text: "Click ability buttons to assign scores, or use Auto-assign for optimal placement.",
	}

	// Update message
	_, err := req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
		Content:    &[]string{""}[0],
	})

	return err
}

// buildComponents creates the UI components for assignment
func (h *AssignAbilitiesHandler) buildComponents(req *AssignAbilitiesRequest, char *character.Character, abilities []string) []discordgo.MessageComponent {
	components := []discordgo.MessageComponent{}

	// Check if we're showing a dropdown for a specific ability
	showDropdownFor := ""
	if req.Interaction.Type == discordgo.InteractionMessageComponent {
		data := req.Interaction.MessageComponentData()
		if strings.HasPrefix(data.CustomID, "character_create:show_assign:") {
			parts := strings.Split(data.CustomID, ":")
			if len(parts) >= 5 {
				showDropdownFor = parts[4]
			}
		}
	}

	// If showing dropdown for a specific ability
	if showDropdownFor != "" {
		var scoreOptions []discordgo.SelectMenuOption

		// Add unassign option
		currentRollID := ""
		if char.AbilityAssignments != nil {
			currentRollID = char.AbilityAssignments[showDropdownFor]
		}

		scoreOptions = append(scoreOptions, discordgo.SelectMenuOption{
			Label:   "Not assigned",
			Value:   "0",
			Default: currentRollID == "",
		})

		// Track which rolls are used elsewhere
		usedBy := make(map[string]string) // rollID -> ability
		if char.AbilityAssignments != nil {
			for ability, rollID := range char.AbilityAssignments {
				if ability != showDropdownFor {
					usedBy[rollID] = ability
				}
			}
		}

		// Add each roll as an option
		for _, roll := range char.AbilityRolls {
			option := discordgo.SelectMenuOption{
				Label:   fmt.Sprintf("%d", roll.Value),
				Value:   roll.ID,
				Default: currentRollID == roll.ID,
			}

			// Show if this roll is used elsewhere
			if usedAbility, exists := usedBy[roll.ID]; exists {
				option.Description = fmt.Sprintf("Currently assigned to %s", usedAbility)
			}

			scoreOptions = append(scoreOptions, option)
		}

		components = append(components, discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    fmt.Sprintf("character_create:assign_ability:%s:%s:%s", req.RaceKey, req.ClassKey, showDropdownFor),
					Placeholder: fmt.Sprintf("Select score for %s", showDropdownFor),
					Options:     scoreOptions,
				},
			},
		})
	}

	// Always show ability buttons (3 per row)
	row1 := discordgo.ActionsRow{Components: []discordgo.MessageComponent{}}
	row2 := discordgo.ActionsRow{Components: []discordgo.MessageComponent{}}

	rollValues := make(map[string]int)
	for _, roll := range char.AbilityRolls {
		rollValues[roll.ID] = roll.Value
	}

	for i, ability := range abilities {
		label := ability
		style := discordgo.SecondaryButton

		if char.AbilityAssignments != nil {
			if rollID, assigned := char.AbilityAssignments[ability]; assigned {
				if score, exists := rollValues[rollID]; exists {
					label = fmt.Sprintf("%s: %d", ability, score)
					style = discordgo.SuccessButton
				}
			}
		}

		button := discordgo.Button{
			Label:    label,
			Style:    style,
			CustomID: fmt.Sprintf("character_create:show_assign:%s:%s:%s", req.RaceKey, req.ClassKey, ability),
		}

		if i < 3 {
			row1.Components = append(row1.Components, button)
		} else {
			row2.Components = append(row2.Components, button)
		}
	}

	components = append(components, row1, row2)

	// Add action buttons
	allAssigned := len(char.AbilityAssignments) == 6
	components = append(components, discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				Label:    "Confirm Abilities",
				Style:    discordgo.SuccessButton,
				CustomID: fmt.Sprintf("character_create:confirm_abilities:%s:%s", req.RaceKey, req.ClassKey),
				Disabled: !allAssigned,
				Emoji: &discordgo.ComponentEmoji{
					Name: "✅",
				},
			},
			discordgo.Button{
				Label:    "Auto-assign",
				Style:    discordgo.PrimaryButton,
				CustomID: fmt.Sprintf("character_create:auto_assign:%s:%s", req.RaceKey, req.ClassKey),
				Emoji: &discordgo.ComponentEmoji{
					Name: "⚡",
				},
			},
			discordgo.Button{
				Label:    "Reroll",
				Style:    discordgo.SecondaryButton,
				CustomID: fmt.Sprintf("character_create:ability_scores:%s:%s", req.RaceKey, req.ClassKey),
				Emoji: &discordgo.ComponentEmoji{
					Name: "🎲",
				},
			},
		},
	})

	return components
}
