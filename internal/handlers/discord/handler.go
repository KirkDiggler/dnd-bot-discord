package discord

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/KirkDiggler/dnd-bot-discord/internal/handlers/discord/dnd/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services"
	characterService "github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
)

// Handler handles all Discord interactions
type Handler struct {
	ServiceProvider                    *services.Provider
	characterCreateHandler             *character.CreateHandler
	characterRaceSelectHandler         *character.RaceSelectHandler
	characterShowClassesHandler        *character.ShowClassesHandler
	characterClassSelectHandler        *character.ClassSelectHandler
	characterAbilityScoresHandler      *character.AbilityScoresHandler
	characterAssignAbilitiesHandler    *character.AssignAbilitiesHandler
	characterProficiencyChoicesHandler *character.ProficiencyChoicesHandler
	characterSelectProficienciesHandler *character.SelectProficienciesHandler
	characterDetailsHandler            *character.CharacterDetailsHandler
}

// HandlerConfig holds configuration for the Discord handler
type HandlerConfig struct {
	ServiceProvider  *services.Provider
}

// NewHandler creates a new Discord handler
func NewHandler(cfg *HandlerConfig) *Handler {
	return &Handler{
		ServiceProvider: cfg.ServiceProvider,
		characterCreateHandler: character.NewCreateHandler(&character.CreateHandlerConfig{
			CharacterService: cfg.ServiceProvider.CharacterService,
		}),
		characterRaceSelectHandler: character.NewRaceSelectHandler(&character.RaceSelectHandlerConfig{
			CharacterService: cfg.ServiceProvider.CharacterService,
		}),
		characterShowClassesHandler: character.NewShowClassesHandler(&character.ShowClassesHandlerConfig{
			CharacterService: cfg.ServiceProvider.CharacterService,
		}),
		characterClassSelectHandler: character.NewClassSelectHandler(&character.ClassSelectHandlerConfig{
			CharacterService: cfg.ServiceProvider.CharacterService,
		}),
		characterAbilityScoresHandler: character.NewAbilityScoresHandler(&character.AbilityScoresHandlerConfig{
			CharacterService: cfg.ServiceProvider.CharacterService,
		}),
		characterAssignAbilitiesHandler: character.NewAssignAbilitiesHandler(&character.AssignAbilitiesHandlerConfig{
			CharacterService: cfg.ServiceProvider.CharacterService,
		}),
		characterProficiencyChoicesHandler: character.NewProficiencyChoicesHandler(&character.ProficiencyChoicesHandlerConfig{
			CharacterService: cfg.ServiceProvider.CharacterService,
		}),
		characterSelectProficienciesHandler: character.NewSelectProficienciesHandler(&character.SelectProficienciesHandlerConfig{
			CharacterService: cfg.ServiceProvider.CharacterService,
		}),
		characterDetailsHandler: character.NewCharacterDetailsHandler(&character.CharacterDetailsHandlerConfig{
			CharacterService: cfg.ServiceProvider.CharacterService,
		}),
	}
}

// RegisterCommands registers all slash commands with Discord
func (h *Handler) RegisterCommands(s *discordgo.Session, guildID string) error {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "dnd",
			Description: "D&D 5e bot commands",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "character",
					Description: "Character management commands",
					Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "create",
							Description: "Create a new character",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
						},
					},
				},
			},
		},
	}

	for _, cmd := range commands {
		_, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, cmd)
		if err != nil {
			return fmt.Errorf("failed to create command %s: %w", cmd.Name, err)
		}
		log.Printf("Registered command: %s", cmd.Name)
	}

	return nil
}

// HandleInteraction handles all Discord interactions
func (h *Handler) HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		h.handleCommand(s, i)
	case discordgo.InteractionMessageComponent:
		h.handleComponent(s, i)
	case discordgo.InteractionModalSubmit:
		h.handleModalSubmit(s, i)
	}
}

// handleCommand handles slash command interactions
func (h *Handler) handleCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()
	
	// Check for /dnd command
	if data.Name != "dnd" {
		return
	}

	// Check for subcommand group
	if len(data.Options) == 0 {
		return
	}

	subcommandGroup := data.Options[0]
	if subcommandGroup.Name == "character" && len(subcommandGroup.Options) > 0 {
		subcommand := subcommandGroup.Options[0]
		
		switch subcommand.Name {
		case "create":
			req := &character.CreateRequest{
				Session:     s,
				Interaction: i,
			}
			if err := h.characterCreateHandler.Handle(req); err != nil {
				log.Printf("Error handling character create: %v", err)
			}
		}
	}
}

// handleComponent handles message component interactions (buttons, select menus)
func (h *Handler) handleComponent(s *discordgo.Session, i *discordgo.InteractionCreate) {
	customID := i.MessageComponentData().CustomID
	
	// Parse custom ID format: "context:action:data"
	// e.g., "character_create:race_select"
	
	// Split the custom ID to parse it
	parts := strings.Split(customID, ":")
	if len(parts) < 2 {
		return
	}
	
	ctx := parts[0]
	action := parts[1]
	
	if ctx == "character_create" {
		switch action {
		case "race_select":
			// Get selected value
			if len(i.MessageComponentData().Values) > 0 {
				req := &character.RaceSelectRequest{
					Session:     s,
					Interaction: i,
					RaceKey:     i.MessageComponentData().Values[0],
				}
				if err := h.characterRaceSelectHandler.Handle(req); err != nil {
					log.Printf("Error handling race selection: %v", err)
				}
			}
		case "show_classes":
			if len(parts) >= 3 {
				req := &character.ShowClassesRequest{
					Session:     s,
					Interaction: i,
					RaceKey:     parts[2],
				}
				if err := h.characterShowClassesHandler.Handle(req); err != nil {
					log.Printf("Error handling show classes: %v", err)
				}
			}
		case "class_select":
			if len(parts) >= 3 && len(i.MessageComponentData().Values) > 0 {
				req := &character.ClassSelectRequest{
					Session:     s,
					Interaction: i,
					RaceKey:     parts[2],
					ClassKey:    i.MessageComponentData().Values[0],
				}
				if err := h.characterClassSelectHandler.Handle(req); err != nil {
					log.Printf("Error handling class selection: %v", err)
				}
			}
		case "ability_scores":
			if len(parts) >= 4 {
				req := &character.AbilityScoresRequest{
					Session:     s,
					Interaction: i,
					RaceKey:     parts[2],
					ClassKey:    parts[3],
				}
				if err := h.characterAbilityScoresHandler.Handle(req); err != nil {
					log.Printf("Error handling ability scores: %v", err)
				}
			}
		case "start_assign":
			if len(parts) >= 5 {
				// Parse rolled scores from custom ID
				rollsStr := parts[4]
				rollsParts := strings.Split(rollsStr, ",")
				rolls := make([]int, 0, len(rollsParts))
				for _, r := range rollsParts {
					if score, err := strconv.Atoi(r); err == nil {
						rolls = append(rolls, score)
					}
				}
				
				req := &character.AssignAbilitiesRequest{
					Session:      s,
					Interaction:  i,
					RaceKey:      parts[2],
					ClassKey:     parts[3],
					RolledScores: rolls,
					Assignments:  make(map[string]int),
				}
				if err := h.characterAssignAbilitiesHandler.Handle(req); err != nil {
					log.Printf("Error handling start assign: %v", err)
				}
			}
		case "assign_ability":
			// This handles individual ability assignments
			// The assignments will be parsed in the handler itself
			if len(parts) >= 4 {
				req := &character.AssignAbilitiesRequest{
					Session:     s,
					Interaction: i,
					RaceKey:     parts[2],
					ClassKey:    parts[3],
					// RolledScores and Assignments will be reconstructed from the current message
				}
				if err := h.characterAssignAbilitiesHandler.Handle(req); err != nil {
					log.Printf("Error handling assign ability: %v", err)
				}
			}
		case "show_assign":
			// This shows the dropdown for a specific ability
			if len(parts) >= 4 {
				req := &character.AssignAbilitiesRequest{
					Session:     s,
					Interaction: i,
					RaceKey:     parts[2],
					ClassKey:    parts[3],
					// RolledScores and Assignments will be reconstructed from the current message
				}
				if err := h.characterAssignAbilitiesHandler.Handle(req); err != nil {
					log.Printf("Error handling show assign: %v", err)
				}
			}
		case "confirm_abilities":
			// Move to proficiency choices
			if len(parts) >= 4 {
				req := &character.ProficiencyChoicesRequest{
					Session:     s,
					Interaction: i,
					RaceKey:     parts[2],
					ClassKey:    parts[3],
				}
				if err := h.characterProficiencyChoicesHandler.Handle(req); err != nil {
					log.Printf("Error handling confirm abilities: %v", err)
				}
			}
		case "select_proficiencies":
			// Show proficiency selection interface
			if len(parts) >= 4 {
				req := &character.SelectProficienciesRequest{
					Session:     s,
					Interaction: i,
					RaceKey:     parts[2],
					ClassKey:    parts[3],
					ChoiceIndex: 0,
					ChoiceType:  "class",
				}
				if err := h.characterSelectProficienciesHandler.Handle(req); err != nil {
					log.Printf("Error handling select proficiencies: %v", err)
				}
			}
		case "confirm_proficiency":
			// For now, just move to character details
			if len(parts) >= 4 {
				// Log selections for future use
				if len(i.MessageComponentData().Values) > 0 {
					log.Printf("Selected proficiencies: %v", i.MessageComponentData().Values)
				}
				
				// Move directly to character details
				req := &character.CharacterDetailsRequest{
					Session:     s,
					Interaction: i,
					RaceKey:     parts[2],
					ClassKey:    parts[3],
				}
				if err := h.characterDetailsHandler.Handle(req); err != nil {
					log.Printf("Error handling character details: %v", err)
				}
			}
		case "select_tool_type":
			// Handle selection of tool/instrument category
			if len(parts) >= 6 {
				// For now, just skip to character details
				req := &character.CharacterDetailsRequest{
					Session:     s,
					Interaction: i,
					RaceKey:     parts[2],
					ClassKey:    parts[3],
				}
				if err := h.characterDetailsHandler.Handle(req); err != nil {
					log.Printf("Error handling select tool type: %v", err)
				}
			}
		case "character_details":
			// Show character details screen
			if len(parts) >= 4 {
				req := &character.CharacterDetailsRequest{
					Session:     s,
					Interaction: i,
					RaceKey:     parts[2],
					ClassKey:    parts[3],
				}
				if err := h.characterDetailsHandler.Handle(req); err != nil {
					log.Printf("Error handling character details: %v", err)
				}
			}
		case "name_modal":
			// Show modal for character name input
			if len(parts) >= 4 {
				modal := discordgo.InteractionResponseData{
					CustomID: fmt.Sprintf("character_create:submit_name:%s:%s", parts[2], parts[3]),
					Title:    "Character Name",
					Components: []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:    "character_name",
									Label:       "What is your character's name?",
									Style:       discordgo.TextInputShort,
									Placeholder: "Enter character name",
									Required:    true,
									MinLength:   1,
									MaxLength:   32,
								},
							},
						},
					},
				}
				
				err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseModal,
					Data: &modal,
				})
				if err != nil {
					log.Printf("Error showing name modal: %v", err)
				}
			}
		}
	}
}

// handleModalSubmit handles modal form submissions
func (h *Handler) handleModalSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ModalSubmitData()
	
	// Parse custom ID
	parts := strings.Split(data.CustomID, ":")
	if len(parts) < 2 {
		return
	}
	
	ctx := parts[0]
	action := parts[1]
	
	if ctx == "character_create" && action == "submit_name" {
		if len(parts) >= 4 {
			// Get the character name from the modal
			characterName := ""
			for _, comp := range data.Components {
				if row, ok := comp.(*discordgo.ActionsRow); ok {
					for _, rowComp := range row.Components {
						if input, ok := rowComp.(*discordgo.TextInput); ok && input.CustomID == "character_name" {
							characterName = input.Value
							break
						}
					}
				}
			}
			
			// Get race and class info
			raceKey := parts[2]
			classKey := parts[3]
			
			// Parse ability scores from the message that triggered the modal
			// The character details screen should have the summary
			abilityScores := make(map[string]int)
			proficiencies := []string{}
			
			// For now, parse from the interaction message if available
			if i.Message != nil && len(i.Message.Embeds) > 0 {
				embed := i.Message.Embeds[0]
				for _, field := range embed.Fields {
					// Look for the summary field that might contain ability scores
					if field.Name == "📊 Summary" && strings.Contains(field.Value, "Abilities: Assigned") {
						// Abilities were assigned, but we need to parse them from somewhere
						// For now, use defaults but mark that we need to implement proper state tracking
						abilityScores = map[string]int{
							"STR": 15,
							"DEX": 14,
							"CON": 13,
							"INT": 12,
							"WIS": 10,
							"CHA": 8,
						}
					}
				}
			}
			
			// If we couldn't parse, use defaults
			if len(abilityScores) == 0 {
				abilityScores = map[string]int{
					"STR": 15,
					"DEX": 14,
					"CON": 13,
					"INT": 12,
					"WIS": 10,
					"CHA": 8,
				}
			}
			
			// Create the character using the service
			input := &characterService.CreateCharacterInput{
				UserID:        i.Member.User.ID,
				RealmID:       i.GuildID, // Discord guild becomes realm
				Name:          characterName,
				RaceKey:       raceKey,
				ClassKey:      classKey,
				AbilityScores: abilityScores,
				Proficiencies: proficiencies,
				Equipment:     []string{}, // Standard starting equipment
			}
			
			output, err := h.ServiceProvider.CharacterService.CreateCharacter(context.Background(), input)
			if err != nil {
				log.Printf("Error creating character: %v", err)
				// Show error to user
				err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: fmt.Sprintf("❌ Failed to create character: %v", err),
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				if err != nil {
					log.Printf("Error responding with error: %v", err)
				}
				return
			}
			
			// Show success with character details
			embed := &discordgo.MessageEmbed{
				Title:       "Character Created!",
				Description: fmt.Sprintf("**Name:** %s\n**Race:** %s\n**Class:** %s", output.Character.Name, output.Character.Race.Name, output.Character.Class.Name),
				Color:       0x00ff00,
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "🆔 Character ID",
						Value:  output.Character.ID,
						Inline: true,
					},
					{
						Name:   "❤️ Hit Points",
						Value:  fmt.Sprintf("%d", output.Character.MaxHitPoints),
						Inline: true,
					},
					{
						Name:   "🛡️ Hit Die",
						Value:  fmt.Sprintf("d%d", output.Character.HitDie),
						Inline: true,
					},
					{
						Name:   "✅ Character Complete",
						Value:  "Your character has been created and saved successfully!",
						Inline: false,
					},
				},
				Footer: &discordgo.MessageEmbedFooter{
					Text: "Ready for adventure!",
				},
			}
			
			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{embed},
				},
			})
			
			if err != nil {
				log.Printf("Error responding to modal submit: %v", err)
			}
		}
	}
}