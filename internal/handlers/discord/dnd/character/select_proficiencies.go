package character

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
)

// SelectProficienciesHandler handles the actual proficiency selection
type SelectProficienciesHandler struct {
	dndClient dnd5e.Client
}

// SelectProficienciesHandlerConfig holds configuration
type SelectProficienciesHandlerConfig struct {
	DNDClient dnd5e.Client
}

// NewSelectProficienciesHandler creates a new handler
func NewSelectProficienciesHandler(cfg *SelectProficienciesHandlerConfig) *SelectProficienciesHandler {
	return &SelectProficienciesHandler{
		dndClient: cfg.DNDClient,
	}
}

// SelectProficienciesRequest represents the request
type SelectProficienciesRequest struct {
	Session     *discordgo.Session
	Interaction *discordgo.InteractionCreate
	RaceKey     string
	ClassKey    string
	ChoiceIndex int    // Which choice we're showing (class may have multiple)
	ChoiceType  string // "class" or "race"
}

// Handle processes proficiency selection
func (h *SelectProficienciesHandler) Handle(req *SelectProficienciesRequest) error {
	// Update the message
	err := req.Session.InteractionRespond(req.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content: "Loading proficiency options...",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to acknowledge interaction: %w", err)
	}

	// Get race and class details
	race, err := h.dndClient.GetRace(req.RaceKey)
	if err != nil {
		return h.respondWithError(req, "Failed to fetch race details.")
	}

	class, err := h.dndClient.GetClass(req.ClassKey)
	if err != nil {
		return h.respondWithError(req, "Failed to fetch class details.")
	}

	// Determine which choice to show
	var currentChoice *entities.Choice
	var choiceSource string
	totalClassChoices := len(class.ProficiencyChoices)
	
	fmt.Printf("DEBUG: Handling proficiency selection - Type: %s, Index: %d, Total class choices: %d\n", 
		req.ChoiceType, req.ChoiceIndex, totalClassChoices)
	
	// First check class choices
	if req.ChoiceType == "" || req.ChoiceType == "class" {
		if len(class.ProficiencyChoices) > req.ChoiceIndex {
			currentChoice = class.ProficiencyChoices[req.ChoiceIndex]
			choiceSource = fmt.Sprintf("%s Class (Choice %d of %d)", class.Name, req.ChoiceIndex+1, totalClassChoices)
		} else if race.StartingProficiencyOptions != nil && len(race.StartingProficiencyOptions.Options) > 0 {
			// Finished class choices, move to race
			currentChoice = race.StartingProficiencyOptions
			choiceSource = fmt.Sprintf("%s Racial Bonus", race.Name)
			req.ChoiceType = "race"
			req.ChoiceIndex = 0
		}
	}
	
	// If specifically looking for race choice
	if req.ChoiceType == "race" && req.ChoiceIndex == 0 && race.StartingProficiencyOptions != nil {
		currentChoice = race.StartingProficiencyOptions
		choiceSource = fmt.Sprintf("%s Racial Bonus", race.Name)
	}

	if currentChoice == nil {
		fmt.Printf("DEBUG: No current choice found. Race has proficiency options: %v\n", 
			race.StartingProficiencyOptions != nil)
		return h.moveToNextStep(req, race, class, "All proficiency choices complete!")
	}
	
	if len(currentChoice.Options) == 0 {
		fmt.Printf("DEBUG: Current choice has 0 options. Type: %s, Index: %d\n", req.ChoiceType, req.ChoiceIndex)
		return h.moveToNextStep(req, race, class, "All proficiency choices complete!")
	}
	
	fmt.Printf("DEBUG: Current choice has %d options\n", len(currentChoice.Options))

	// Create embed
	embed := &discordgo.MessageEmbed{
		Title:       "Select Proficiencies",
		Description: fmt.Sprintf("**%s**\n\nChoose %d %s:", choiceSource, currentChoice.Count, currentChoice.Name),
		Color:       0x5865F2,
		Fields:      []*discordgo.MessageEmbedField{},
	}
	
	// Add progress if there are multiple choices
	if totalClassChoices > 1 || race.StartingProficiencyOptions != nil {
		progressParts := []string{}
		
		// Show class choice progress
		for i := 0; i < totalClassChoices; i++ {
			if i < req.ChoiceIndex && req.ChoiceType == "class" {
				progressParts = append(progressParts, "✅")
			} else if i == req.ChoiceIndex && req.ChoiceType == "class" {
				progressParts = append(progressParts, "⏳")
			} else {
				progressParts = append(progressParts, "⭕")
			}
		}
		
		// Show race choice progress
		if race.StartingProficiencyOptions != nil {
			if req.ChoiceType == "race" {
				progressParts = append(progressParts, "| 🏃 ⏳")
			} else {
				progressParts = append(progressParts, "| 🏃 ⭕")
			}
		}
		
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Progress",
			Value:  strings.Join(progressParts, " "),
			Inline: false,
		})
	}

	// Show available options
	optionStrings := []string{}
	fmt.Printf("DEBUG: Processing %d options for display\n", len(currentChoice.Options))
	for i, option := range currentChoice.Options {
		if i >= 10 { // Limit display to first 10
			optionStrings = append(optionStrings, fmt.Sprintf("_...and %d more_", len(currentChoice.Options)-10))
			break
		}
		fmt.Printf("DEBUG: Option %d type: %T\n", i, option)
		optionName := h.getOptionName(option)
		if optionName != "" {
			optionStrings = append(optionStrings, fmt.Sprintf("• %s", optionName))
		} else {
			fmt.Printf("DEBUG: Option %d returned empty name\n", i)
		}
	}
	
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "Available Options",
		Value:  strings.Join(optionStrings, "\n"),
		Inline: false,
	})

	// Check if all options are nested choices
	hasNestedChoices := false
	for _, option := range currentChoice.Options {
		if _, ok := option.(*entities.Choice); ok {
			hasNestedChoices = true
			break
		}
	}
	
	// Create components based on option types
	components := []discordgo.MessageComponent{}
	
	if hasNestedChoices {
		// Show nested choices as buttons
		fmt.Println("DEBUG: Detected nested choices, showing as buttons")
		row := discordgo.ActionsRow{Components: []discordgo.MessageComponent{}}
		
		for i, option := range currentChoice.Options {
			if _, ok := option.(*entities.Choice); ok {
				button := discordgo.Button{
					Label:    h.getOptionName(option),
					Style:    discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("character_create:select_tool_type:%s:%s:%d:%d", req.RaceKey, req.ClassKey, req.ChoiceIndex, i),
				}
				row.Components = append(row.Components, button)
			}
		}
		
		if len(row.Components) > 0 {
			components = append(components, row)
		}
		
		embed.Footer = &discordgo.MessageEmbedFooter{
			Text: "Select which type of proficiency you want",
		}
	} else {
		// Regular dropdown for direct options
		selectOptions := []discordgo.SelectMenuOption{}
		for _, option := range currentChoice.Options {
			optionName := h.getOptionName(option)
			optionKey := h.getOptionKey(option)
			if optionName != "" && optionKey != "" {
				selectOptions = append(selectOptions, discordgo.SelectMenuOption{
					Label: optionName,
					Value: optionKey,
				})
			}
		}
		
		// Debug: log the number of options
		fmt.Printf("DEBUG: Found %d options for %s\n", len(selectOptions), choiceSource)
		
		// If no options were parsed, add a debug option
		if len(selectOptions) == 0 {
			fmt.Printf("DEBUG: No options parsed from %d raw options\n", len(currentChoice.Options))
			selectOptions = append(selectOptions, discordgo.SelectMenuOption{
				Label: "No options available",
				Value: "none",
			})
		}
		
		// Limit to 25 options (Discord limit)
		if len(selectOptions) > 25 {
			selectOptions = selectOptions[:25]
			embed.Footer = &discordgo.MessageEmbedFooter{
				Text: fmt.Sprintf("Showing first 25 of %d options", len(currentChoice.Options)),
			}
		}
		
		components = append(components, discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    fmt.Sprintf("character_create:confirm_proficiency:%s:%s:%s:%d", req.RaceKey, req.ClassKey, req.ChoiceType, req.ChoiceIndex),
					Placeholder: fmt.Sprintf("Select %d %s", currentChoice.Count, currentChoice.Name),
					Options:     selectOptions,
					MinValues:   &currentChoice.Count,
					MaxValues:   currentChoice.Count,
				},
			},
		})
	}

	// Update message
	_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
		Content:    &[]string{""}[0],
	})

	return err
}

// getOptionName extracts the display name from an option
func (h *SelectProficienciesHandler) getOptionName(option entities.Option) string {
	if option == nil {
		fmt.Println("DEBUG: getOptionName received nil option")
		return ""
	}
	
	switch opt := option.(type) {
	case *entities.ReferenceOption:
		if opt.Reference != nil {
			return opt.Reference.Name
		}
		fmt.Println("DEBUG: ReferenceOption has nil Reference")
	case *entities.CountedReferenceOption:
		if opt.Reference != nil {
			return fmt.Sprintf("%s (×%d)", opt.Reference.Name, opt.Count)
		}
		fmt.Println("DEBUG: CountedReferenceOption has nil Reference")
	case *entities.MultipleOption:
		// For multiple options, show combined name
		names := []string{}
		for _, item := range opt.Items {
			if name := h.getOptionName(item); name != "" {
				names = append(names, name)
			}
		}
		return strings.Join(names, " + ")
	case *entities.Choice:
		// Handle nested choices - show what type of choice it is
		if opt.Name != "" {
			return opt.Name
		}
		return fmt.Sprintf("Choose %d items", opt.Count)
	default:
		fmt.Printf("DEBUG: Unknown option type: %T\n", option)
	}
	return ""
}

// getOptionKey extracts a unique key from an option
func (h *SelectProficienciesHandler) getOptionKey(option entities.Option) string {
	switch opt := option.(type) {
	case *entities.ReferenceOption:
		if opt.Reference != nil {
			return opt.Reference.Key
		}
	case *entities.CountedReferenceOption:
		if opt.Reference != nil {
			return opt.Reference.Key
		}
	case *entities.MultipleOption:
		// For multiple options, combine keys
		keys := []string{}
		for _, item := range opt.Items {
			if key := h.getOptionKey(item); key != "" {
				keys = append(keys, key)
			}
		}
		return strings.Join(keys, "+")
	case *entities.Choice:
		// For nested choices, use the choice key or type
		if opt.Key != "" {
			return opt.Key
		}
		return fmt.Sprintf("choice_%s", opt.Type)
	}
	return ""
}

// moveToNextStep transitions to the next part of character creation
func (h *SelectProficienciesHandler) moveToNextStep(req *SelectProficienciesRequest, race *entities.Race, class *entities.Class, message string) error {
	embed := &discordgo.MessageEmbed{
		Title:       "Proficiencies Complete",
		Description: fmt.Sprintf("**Race:** %s\n**Class:** %s\n\n%s", race.Name, class.Name, message),
		Color:       0x5865F2,
	}

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Continue to Character Details",
					Style:    discordgo.SuccessButton,
					CustomID: fmt.Sprintf("character_create:character_details:%s:%s", req.RaceKey, req.ClassKey),
					Emoji: &discordgo.ComponentEmoji{
						Name: "➡️",
					},
				},
			},
		},
	}

	_, err := req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
		Content:    &[]string{""}[0],
	})

	return err
}

// respondWithError updates the message with an error
func (h *SelectProficienciesHandler) respondWithError(req *SelectProficienciesRequest, message string) error {
	content := fmt.Sprintf("❌ %s", message)
	_, err := req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
		Content: &content,
		Embeds:  &[]*discordgo.MessageEmbed{},
	})
	return err
}