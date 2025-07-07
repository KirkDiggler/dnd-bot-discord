package character

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services"
	"github.com/bwmarrin/discordgo"
)

// FlowHandler handles character creation using the new service-driven approach
type FlowHandler struct {
	services *services.Provider
}

// NewFlowHandler creates a new flow handler
func NewFlowHandler(serviceProvider *services.Provider) *FlowHandler {
	return &FlowHandler{
		services: serviceProvider,
	}
}

// HandleContinue continues character creation from where it left off
func (h *FlowHandler) HandleContinue(s *discordgo.Session, i *discordgo.InteractionCreate, characterID string) error {
	log.Printf("FlowHandler.HandleContinue called for character %s", characterID)
	ctx := context.Background()

	// Get the next step from the service
	step, err := h.services.CreationFlowService.GetNextStep(ctx, characterID)
	if err != nil {
		log.Printf("Error getting next step for character %s: %v", characterID, err)
		return respondWithError(s, i, "Failed to determine next creation step")
	}

	log.Printf("Next step for character %s is %s", characterID, step.Type)

	// Route to appropriate handler based on step type
	return h.routeToStepHandler(s, i, characterID, step)
}

// HandleSelection processes a selection from a creation flow step
func (h *FlowHandler) HandleSelection(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// Parse the custom ID: creation_flow:{characterID}:{stepType}
	parts := strings.Split(i.MessageComponentData().CustomID, ":")
	if len(parts) < 3 {
		return respondWithError(s, i, "Invalid selection")
	}

	characterID := parts[1]
	stepType := character.CreationStepType(parts[2])

	// Get selections from the interaction
	selections := i.MessageComponentData().Values

	ctx := context.Background()

	// Create step result
	result := &character.CreationStepResult{
		StepType:   stepType,
		Selections: selections,
	}

	// Process the result and get next step
	nextStep, err := h.services.CreationFlowService.ProcessStepResult(ctx, characterID, result)
	if err != nil {
		log.Printf("Error processing step result for character %s: %v", characterID, err)
		return respondWithError(s, i, "Failed to process selection")
	}

	// Route to next step handler
	return h.routeToStepHandler(s, i, characterID, nextStep)
}

// routeToStepHandler routes to the appropriate handler for a step type
func (h *FlowHandler) routeToStepHandler(s *discordgo.Session, i *discordgo.InteractionCreate,
	characterID string, step *character.CreationStep) error {

	switch step.Type {
	case character.StepTypeComplete:
		return h.handleCreationComplete(s, i, characterID)

	case character.StepTypeRaceSelection:
		// Use generic step renderer for race selection
		return h.renderGenericStep(s, i, step, characterID)

	case character.StepTypeClassSelection:
		// Use generic step renderer for class selection
		return h.renderGenericStep(s, i, step, characterID)

	case character.StepTypeAbilityScores:
		// Get character to extract race and class keys
		char, err := h.services.CharacterService.GetByID(characterID)
		if err != nil {
			return respondWithError(s, i, "Failed to get character for ability scores")
		}

		// Use existing ability scores handler
		handler := NewAbilityScoresHandler(&AbilityScoresHandlerConfig{
			CharacterService: h.services.CharacterService,
		})
		req := &AbilityScoresRequest{
			Session:     s,
			Interaction: i,
			RaceKey:     char.Race.Key,
			ClassKey:    char.Class.Key,
		}
		return handler.Handle(req)

	case character.StepTypeAbilityAssignment:
		// Get character to extract race and class keys
		char, err := h.services.CharacterService.GetByID(characterID)
		if err != nil {
			return respondWithError(s, i, "Failed to get character for ability assignment")
		}

		// Use existing ability assignment handler
		handler := NewAssignAbilitiesHandler(&AssignAbilitiesHandlerConfig{
			CharacterService: h.services.CharacterService,
		})
		req := &AssignAbilitiesRequest{
			Session:     s,
			Interaction: i,
			RaceKey:     char.Race.Key,
			ClassKey:    char.Class.Key,
			AutoAssign:  false,
		}
		return handler.Handle(req)

	case character.StepTypeProficiencySelection:
		// Get character to extract race and class keys
		char, err := h.services.CharacterService.GetByID(characterID)
		if err != nil {
			return respondWithError(s, i, "Failed to get character for proficiency selection")
		}

		// Use existing proficiency handler
		handler := NewProficiencyChoicesHandler(&ProficiencyChoicesHandlerConfig{
			CharacterService: h.services.CharacterService,
		})
		req := &ProficiencyChoicesRequest{
			Session:     s,
			Interaction: i,
			RaceKey:     char.Race.Key,
			ClassKey:    char.Class.Key,
		}
		return handler.Handle(req)

	case character.StepTypeEquipmentSelection:
		// Get character to extract race and class keys
		char, err := h.services.CharacterService.GetByID(characterID)
		if err != nil {
			return respondWithError(s, i, "Failed to get character for equipment selection")
		}

		// Use existing equipment handler
		handler := NewEquipmentChoicesHandler(&EquipmentChoicesHandlerConfig{
			CharacterService: h.services.CharacterService,
		})
		req := &EquipmentChoicesRequest{
			Session:     s,
			Interaction: i,
			RaceKey:     char.Race.Key,
			ClassKey:    char.Class.Key,
		}
		return handler.Handle(req)

	case character.StepTypeCharacterDetails:
		// Get character to extract race and class keys
		char, err := h.services.CharacterService.GetByID(characterID)
		if err != nil {
			return respondWithError(s, i, "Failed to get character for details")
		}

		// Use existing details handler
		handler := NewCharacterDetailsHandler(&CharacterDetailsHandlerConfig{
			CharacterService: h.services.CharacterService,
		})
		req := &CharacterDetailsRequest{
			Session:     s,
			Interaction: i,
			RaceKey:     char.Race.Key,
			ClassKey:    char.Class.Key,
		}
		return handler.Handle(req)

	// Class-specific steps handled by the flow handler itself
	case character.StepTypeDivineDomainSelection:
		// Use existing divine domain handler
		handler := NewClassFeaturesHandler(h.services.CharacterService)
		req := &InteractionRequest{
			Session:     s,
			Interaction: i,
			CharacterID: characterID,
		}
		return handler.ShowDivineDomainSelection(req)

	case character.StepTypeFightingStyleSelection:
		// Use existing fighting style handler
		handler := NewClassFeaturesHandler(h.services.CharacterService)
		req := &InteractionRequest{
			Session:     s,
			Interaction: i,
			CharacterID: characterID,
		}
		return handler.ShowFightingStyleSelection(req)

	case character.StepTypeFavoredEnemySelection:
		// Use existing favored enemy handler
		handler := NewClassFeaturesHandler(h.services.CharacterService)
		req := &InteractionRequest{
			Session:     s,
			Interaction: i,
			CharacterID: characterID,
		}
		return handler.ShowFavoredEnemySelection(req)

	case character.StepTypeNaturalExplorerSelection:
		// Use existing natural explorer handler
		handler := NewClassFeaturesHandler(h.services.CharacterService)
		req := &InteractionRequest{
			Session:     s,
			Interaction: i,
			CharacterID: characterID,
		}
		return handler.ShowNaturalExplorerSelection(req)

	// New step types that need generic rendering
	case character.StepTypeSkillSelection,
		character.StepTypeLanguageSelection:
		return h.renderGenericStep(s, i, step, characterID)

	default:
		log.Printf("Unknown step type: %s", step.Type)
		return respondWithError(s, i, "Unknown creation step")
	}
}

// renderGenericStep renders a step using the generic flow UI
func (h *FlowHandler) renderGenericStep(s *discordgo.Session, i *discordgo.InteractionCreate,
	step *character.CreationStep, characterID string) error {

	// Build embed
	embed := &discordgo.MessageEmbed{
		Title:       step.Title,
		Description: step.Description,
		Color:       h.getStepColor(step.Type),
	}

	// Get character to show current selections
	ctx := context.Background()
	char, err := h.services.CharacterService.GetByID(characterID)
	if err == nil {
		// Show current selections
		var currentInfo []string
		if char.Race != nil {
			currentInfo = append(currentInfo, fmt.Sprintf("**Race:** %s", char.Race.Name))
		}
		if char.Class != nil {
			currentInfo = append(currentInfo, fmt.Sprintf("**Class:** %s", char.Class.Name))
		}
		if len(currentInfo) > 0 {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "Current Character",
				Value:  strings.Join(currentInfo, "\n"),
				Inline: false,
			})
		}
	}

	// Add progress field
	if progressSteps, err := h.services.CreationFlowService.GetProgressSteps(ctx, characterID); err == nil {
		var lines []string
		for idx, stepInfo := range progressSteps {
			icon := "⏳"
			if stepInfo.Completed {
				icon = "✅"
			} else if stepInfo.Current {
				icon = "🔄"
			}
			lines = append(lines, fmt.Sprintf("%s Step %d: %s", icon, idx+1, stepInfo.Step.Title))
		}

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Progress",
			Value:  strings.Join(lines, "\n"),
			Inline: false,
		})
	}

	// Build components
	if len(step.Options) == 0 {
		return respondWithError(s, i, "No options available for this step")
	}

	// Convert options to select menu
	var selectOptions []discordgo.SelectMenuOption
	for _, option := range step.Options {
		desc := option.Description
		if len(desc) > 100 {
			desc = desc[:97] + "..."
		}
		selectOptions = append(selectOptions, discordgo.SelectMenuOption{
			Label:       option.Name,
			Value:       option.Key,
			Description: desc,
		})
	}

	minValues := 1
	if step.MinChoices > 0 {
		minValues = step.MinChoices
	}

	maxValues := 1
	if step.MaxChoices > 0 {
		maxValues = step.MaxChoices
	}

	customID := fmt.Sprintf("creation_flow:%s:%s", characterID, step.Type)

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    customID,
					Placeholder: h.getStepPlaceholder(step),
					Options:     selectOptions,
					MinValues:   &minValues,
					MaxValues:   maxValues,
				},
			},
		},
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

// handleCreationComplete handles when character creation is finished
func (h *FlowHandler) handleCreationComplete(s *discordgo.Session, i *discordgo.InteractionCreate, characterID string) error {
	ctx := context.Background()

	// Finalize the character
	if _, err := h.services.CharacterService.FinalizeDraftCharacter(ctx, characterID); err != nil {
		log.Printf("Error finalizing character %s: %v", characterID, err)
		return respondWithError(s, i, "Failed to finalize character")
	}

	// Show success message
	embed := &discordgo.MessageEmbed{
		Title:       "Character Creation Complete! 🎉",
		Description: "Your character is ready for adventure!",
		Color:       0x00ff00, // Green
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "What's Next?",
				Value:  "• View your character sheet\n• Join a session\n• Start adventuring!",
				Inline: false,
			},
		},
	}

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "View Character",
					Style:    discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("character:sheet_refresh:%s", characterID),
					Emoji:    &discordgo.ComponentEmoji{Name: "📄"},
				},
			},
		},
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
			Flags:      discordgo.MessageFlagsEphemeral,
		},
	})
}

// Helper methods
func (h *FlowHandler) getStepColor(stepType character.CreationStepType) int {
	switch stepType {
	case character.StepTypeSkillSelection:
		return 0x9b59b6 // Purple
	case character.StepTypeLanguageSelection:
		return 0xe67e22 // Orange
	case character.StepTypeDivineDomainSelection:
		return 0xf1c40f // Gold
	case character.StepTypeFightingStyleSelection:
		return 0xe74c3c // Red
	default:
		return 0x3498db // Blue
	}
}

func (h *FlowHandler) getStepPlaceholder(step *character.CreationStep) string {
	switch step.Type {
	case character.StepTypeSkillSelection:
		if step.MaxChoices > 1 {
			return fmt.Sprintf("Select %d skills...", step.MaxChoices)
		}
		return "Select a skill..."
	case character.StepTypeLanguageSelection:
		if step.MaxChoices > 1 {
			return fmt.Sprintf("Select %d languages...", step.MaxChoices)
		}
		return "Select a language..."
	case character.StepTypeRaceSelection:
		return "Select your race..."
	case character.StepTypeClassSelection:
		return "Select your class..."
	default:
		return "Make your selection..."
	}
}
