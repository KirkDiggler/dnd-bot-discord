package character

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
)

// AssignAbilitiesHandler handles assigning rolled scores to abilities
type AssignAbilitiesHandler struct {
	dndClient dnd5e.Client
}

// AssignAbilitiesHandlerConfig holds configuration
type AssignAbilitiesHandlerConfig struct {
	DNDClient dnd5e.Client
}

// NewAssignAbilitiesHandler creates a new handler
func NewAssignAbilitiesHandler(cfg *AssignAbilitiesHandlerConfig) *AssignAbilitiesHandler {
	return &AssignAbilitiesHandler{
		dndClient: cfg.DNDClient,
	}
}

// AssignAbilitiesRequest represents the request
type AssignAbilitiesRequest struct {
	Session     *discordgo.Session
	Interaction *discordgo.InteractionCreate
	RaceKey     string
	ClassKey    string
	RolledScores []int // The 6 rolled values
	Assignments  map[string]int // Current assignments (ability -> score)
}

// Handle processes ability assignment
func (h *AssignAbilitiesHandler) Handle(req *AssignAbilitiesRequest) error {
	// If we don't have rolled scores, we need to parse them from the message
	if len(req.RolledScores) == 0 && req.Interaction.Type == discordgo.InteractionMessageComponent {
		// Get the current message to parse state
		message := req.Interaction.Message
		if message != nil && len(message.Embeds) > 0 {
			embed := message.Embeds[0]
			// Parse rolled scores from the embed
			for _, field := range embed.Fields {
				if field.Name == "🎲 Your Rolls" {
					// Parse the rolls from the field value
					parts := strings.Split(field.Value, " • ")
					for _, part := range parts {
						// Extract number from format like "**17** ✓"
						cleaned := strings.TrimSpace(part)
						cleaned = strings.TrimPrefix(cleaned, "**")
						if idx := strings.Index(cleaned, "**"); idx > 0 {
							cleaned = cleaned[:idx]
						}
						if score, err := strconv.Atoi(cleaned); err == nil {
							req.RolledScores = append(req.RolledScores, score)
						}
					}
				}
			}
			
			// Parse current assignments from embed
			if req.Assignments == nil {
				req.Assignments = make(map[string]int)
			}
			for _, field := range embed.Fields {
				if field.Name == "📊 Ability Scores" {
					lines := strings.Split(field.Value, "\n")
					for _, line := range lines {
						// Parse lines like "**STR:** 17 (+2) = 19 [+4]"
						if strings.Contains(line, ":") && !strings.Contains(line, "_Not assigned_") {
							parts := strings.Split(line, ":")
							ability := strings.Trim(strings.Trim(parts[0], "*"), " ")
							scoreStr := strings.TrimSpace(parts[1])
							// Extract just the base score
							if idx := strings.Index(scoreStr, " "); idx > 0 {
								scoreStr = scoreStr[:idx]
							}
							if score, err := strconv.Atoi(scoreStr); err == nil {
								req.Assignments[ability] = score
							}
						}
					}
				}
			}
		}
	}
	
	// Parse any new assignment from the interaction
	if req.Interaction.Type == discordgo.InteractionMessageComponent {
		data := req.Interaction.MessageComponentData()
		if strings.HasPrefix(data.CustomID, "character_create:assign_ability:") {
			parts := strings.Split(data.CustomID, ":")
			if len(parts) >= 5 && len(data.Values) > 0 {
				ability := parts[4]
				valueStr := data.Values[0]
				
				// Parse score from "score:index" format
				score := 0
				if valueStr != "0" {
					valueParts := strings.Split(valueStr, ":")
					if len(valueParts) > 0 {
						score, _ = strconv.Atoi(valueParts[0])
					}
				}
				
				// Initialize assignments if needed
				if req.Assignments == nil {
					req.Assignments = make(map[string]int)
				}
				
				// If score is 0, remove the assignment
				if score == 0 {
					delete(req.Assignments, ability)
				} else {
					// Remove this score from any other ability
					for a := range req.Assignments {
						if req.Assignments[a] == score && a != ability {
							delete(req.Assignments, a)
						}
					}
					// Assign the score
					req.Assignments[ability] = score
				}
			}
		}
	}

	// Update the message
	err := req.Session.InteractionRespond(req.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content: "Updating ability assignments...",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to acknowledge interaction: %w", err)
	}

	// Get race and class for context
	race, err := h.dndClient.GetRace(req.RaceKey)
	if err != nil {
		return h.respondWithError(req, "Failed to fetch race details.")
	}

	class, err := h.dndClient.GetClass(req.ClassKey)
	if err != nil {
		return h.respondWithError(req, "Failed to fetch class details.")
	}

	// Create embed
	embed := &discordgo.MessageEmbed{
		Title:       "Create New Character - Assign Abilities",
		Description: fmt.Sprintf("**Race:** %s\n**Class:** %s\n\nAssign your rolled scores to each ability.", race.Name, class.Name),
		Color:       0x5865F2,
		Fields:      []*discordgo.MessageEmbedField{},
	}

	// Show rolled values
	rollStrings := []string{}
	for _, roll := range req.RolledScores {
		assigned := h.isScoreAssigned(roll, req.Assignments)
		status := ""
		if assigned {
			status = " ✓"
		}
		rollStrings = append(rollStrings, fmt.Sprintf("**%d**%s", roll, status))
	}
	
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "🎲 Your Rolls",
		Value:  strings.Join(rollStrings, " • "),
		Inline: false,
	})

	// Show current assignments in two columns
	abilities := []string{"STR", "DEX", "CON", "INT", "WIS", "CHA"}
	physicalStrings := []string{}
	mentalStrings := []string{}
	
	for i, ability := range abilities {
		score, assigned := req.Assignments[ability]
		racialBonus := h.getRacialBonus(race, ability)
		
		var line string
		if assigned {
			total := score + racialBonus
			modifier := (total - 10) / 2
			modStr := fmt.Sprintf("%+d", modifier)
			
			line = fmt.Sprintf("**%s:** %d", ability, score)
			if racialBonus > 0 {
				line += fmt.Sprintf(" (+%d) = %d [%s]", racialBonus, total, modStr)
			} else {
				line += fmt.Sprintf(" [%s]", modStr)
			}
		} else {
			line = fmt.Sprintf("**%s:** _Not assigned_", ability)
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
	})
	
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
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
	})

	// Progress
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "Progress",
		Value:  "✅ Step 1: Race\n✅ Step 2: Class\n⏳ Step 3: Abilities\n⏳ Step 4: Details",
		Inline: false,
	})

	// Create ability buttons
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
		// Create score options
		var scoreOptions []discordgo.SelectMenuOption
		
		// Add unassign option
		scoreOptions = append(scoreOptions, discordgo.SelectMenuOption{
			Label:   "Not assigned",
			Value:   "0",
			Default: req.Assignments[showDropdownFor] == 0,
		})
		
		// Add each rolled score with index to ensure uniqueness
		for i, score := range req.RolledScores {
			assignedTo := ""
			for a, s := range req.Assignments {
				if s == score && a != showDropdownFor {
					assignedTo = a
					break
				}
			}
			
			option := discordgo.SelectMenuOption{
				Label:   fmt.Sprintf("%d", score),
				Value:   fmt.Sprintf("%d:%d", score, i), // Include index to ensure uniqueness
				Default: req.Assignments[showDropdownFor] == score,
			}
			
			if assignedTo != "" {
				option.Description = fmt.Sprintf("Currently assigned to %s", assignedTo)
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
	} else {
		// Show ability buttons (3 per row)
		row1 := discordgo.ActionsRow{Components: []discordgo.MessageComponent{}}
		row2 := discordgo.ActionsRow{Components: []discordgo.MessageComponent{}}
		
		for i, ability := range abilities {
			score, assigned := req.Assignments[ability]
			
			label := ability
			if assigned {
				label = fmt.Sprintf("%s: %d", ability, score)
			}
			
			style := discordgo.SecondaryButton
			if assigned {
				style = discordgo.SuccessButton
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
	}

	// Add action buttons
	allAssigned := len(req.Assignments) == 6
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
				Label:    "Reroll",
				Style:    discordgo.SecondaryButton,
				CustomID: fmt.Sprintf("character_create:ability_scores:%s:%s", req.RaceKey, req.ClassKey),
				Emoji: &discordgo.ComponentEmoji{
					Name: "🎲",
				},
			},
		},
	})

	embed.Footer = &discordgo.MessageEmbedFooter{
		Text: "Assign each rolled score to an ability. Racial bonuses will be added automatically.",
	}

	// Update message
	_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
		Content:    &[]string{""}[0],
	})

	return err
}

// isScoreAssigned checks if a score is assigned to any ability
func (h *AssignAbilitiesHandler) isScoreAssigned(score int, assignments map[string]int) bool {
	for _, s := range assignments {
		if s == score {
			return true
		}
	}
	return false
}

// getRacialBonus gets the racial bonus for a specific ability
func (h *AssignAbilitiesHandler) getRacialBonus(race *entities.Race, ability string) int {
	// Convert ability string to Attribute type
	var attr entities.Attribute
	switch ability {
	case "STR":
		attr = entities.AttributeStrength
	case "DEX":
		attr = entities.AttributeDexterity
	case "CON":
		attr = entities.AttributeConstitution
	case "INT":
		attr = entities.AttributeIntelligence
	case "WIS":
		attr = entities.AttributeWisdom
	case "CHA":
		attr = entities.AttributeCharisma
	}
	
	for _, bonus := range race.AbilityBonuses {
		if bonus.Attribute == attr {
			return bonus.Bonus
		}
	}
	return 0
}

// getAbilityDescription returns a description for an ability
func (h *AssignAbilitiesHandler) getAbilityDescription(ability string) string {
	descriptions := map[string]string{
		"STR": "Physical power",
		"DEX": "Agility and reflexes",
		"CON": "Endurance and health",
		"INT": "Reasoning and memory",
		"WIS": "Perception and insight",
		"CHA": "Force of personality",
	}
	
	if desc, ok := descriptions[ability]; ok {
		return desc
	}
	
	return ""
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