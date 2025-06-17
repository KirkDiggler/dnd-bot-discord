package discord

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/handlers/discord/dnd/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/handlers/discord/dnd/encounter"
	"github.com/KirkDiggler/dnd-bot-discord/internal/handlers/discord/dnd/help"
	"github.com/KirkDiggler/dnd-bot-discord/internal/handlers/discord/dnd/session"
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
	characterEquipmentChoicesHandler   *character.EquipmentChoicesHandler
	characterSelectEquipmentHandler    *character.SelectEquipmentHandler
	characterSelectNestedEquipmentHandler *character.SelectNestedEquipmentHandler
	characterDetailsHandler            *character.CharacterDetailsHandler
	characterListHandler               *character.ListHandler
	characterShowHandler               *character.ShowHandler
	
	// Session handlers
	sessionCreateHandler               *session.CreateHandler
	sessionListHandler                 *session.ListHandler
	sessionJoinHandler                 *session.JoinHandler
	sessionStartHandler                *session.StartHandler
	sessionEndHandler                  *session.EndHandler
	sessionInfoHandler                 *session.InfoHandler
	
	// Encounter handlers
	encounterAddMonsterHandler         *encounter.AddMonsterHandler
	
	// Help handler
	helpHandler                        *help.HelpHandler
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
		characterEquipmentChoicesHandler: character.NewEquipmentChoicesHandler(&character.EquipmentChoicesHandlerConfig{
			CharacterService: cfg.ServiceProvider.CharacterService,
		}),
		characterSelectEquipmentHandler: character.NewSelectEquipmentHandler(&character.SelectEquipmentHandlerConfig{
			CharacterService: cfg.ServiceProvider.CharacterService,
		}),
		characterSelectNestedEquipmentHandler: character.NewSelectNestedEquipmentHandler(&character.SelectNestedEquipmentHandlerConfig{
			CharacterService: cfg.ServiceProvider.CharacterService,
		}),
		characterDetailsHandler: character.NewCharacterDetailsHandler(&character.CharacterDetailsHandlerConfig{
			CharacterService: cfg.ServiceProvider.CharacterService,
		}),
		characterListHandler: character.NewListHandler(cfg.ServiceProvider),
		characterShowHandler: character.NewShowHandler(cfg.ServiceProvider),
		
		// Initialize session handlers
		sessionCreateHandler: session.NewCreateHandler(cfg.ServiceProvider),
		sessionListHandler:   session.NewListHandler(cfg.ServiceProvider),
		sessionJoinHandler:   session.NewJoinHandler(cfg.ServiceProvider),
		sessionStartHandler:  session.NewStartHandler(cfg.ServiceProvider),
		sessionEndHandler:    session.NewEndHandler(cfg.ServiceProvider),
		sessionInfoHandler:   session.NewInfoHandler(cfg.ServiceProvider),
		
		// Initialize encounter handlers
		encounterAddMonsterHandler: encounter.NewAddMonsterHandler(cfg.ServiceProvider),
		
		// Initialize help handler
		helpHandler: help.NewHelpHandler(),
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
						{
							Name:        "list",
							Description: "List all your characters",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
						},
						{
							Name:        "show",
							Description: "Show details of a character",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionString,
									Name:        "id",
									Description: "Character ID to show",
									Required:    true,
								},
							},
						},
					},
				},
				{
					Name:        "session",
					Description: "Session management commands",
					Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "create",
							Description: "Create a new game session",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionString,
									Name:        "name",
									Description: "Session name",
									Required:    true,
								},
								{
									Type:        discordgo.ApplicationCommandOptionString,
									Name:        "description",
									Description: "Session description (optional)",
									Required:    false,
								},
							},
						},
						{
							Name:        "list",
							Description: "List all your sessions",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
						},
						{
							Name:        "join",
							Description: "Join a session with invite code",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionString,
									Name:        "code",
									Description: "Invite code",
									Required:    true,
								},
							},
						},
						{
							Name:        "info",
							Description: "Show info about your current session",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
						},
						{
							Name:        "start",
							Description: "Start a session (DM only)",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionString,
									Name:        "id",
									Description: "Session ID (optional if you have only one session)",
									Required:    false,
								},
							},
						},
						{
							Name:        "end",
							Description: "End a session (DM only)",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionString,
									Name:        "id",
									Description: "Session ID (optional if you have only one session)",
									Required:    false,
								},
							},
						},
					},
				},
				{
					Name:        "encounter",
					Description: "Encounter and combat management",
					Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "add",
							Description: "Add a monster to the encounter",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionString,
									Name:        "monster",
									Description: "Monster name to search for",
									Required:    true,
								},
							},
						},
					},
				},
				{
					Name:        "help",
					Description: "Get help on using the bot",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "topic",
							Description: "Specific help topic (character, session, encounter, combat)",
							Required:    false,
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
	
	// Handle direct subcommands (like help)
	if subcommandGroup.Type == discordgo.ApplicationCommandOptionSubCommand {
		switch subcommandGroup.Name {
		case "help":
			// Get topic from options if provided
			var topic string
			for _, opt := range subcommandGroup.Options {
				if opt.Name == "topic" {
					topic = opt.StringValue()
					break
				}
			}
			req := &help.HelpRequest{
				Session:     s,
				Interaction: i,
				Topic:       topic,
			}
			if err := h.helpHandler.Handle(req); err != nil {
				log.Printf("Error handling help command: %v", err)
			}
		}
		return
	}
	
	// Handle subcommand groups
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
		case "list":
			req := &character.ListRequest{
				Session:     s,
				Interaction: i,
			}
			if err := h.characterListHandler.Handle(req); err != nil {
				log.Printf("Error handling character list: %v", err)
			}
		case "show":
			// Get character ID from options
			var characterID string
			for _, opt := range subcommand.Options {
				if opt.Name == "id" {
					characterID = opt.StringValue()
					break
				}
			}
			req := &character.ShowRequest{
				Session:     s,
				Interaction: i,
				CharacterID: characterID,
			}
			if err := h.characterShowHandler.Handle(req); err != nil {
				log.Printf("Error handling character show: %v", err)
			}
		}
	} else if subcommandGroup.Name == "session" && len(subcommandGroup.Options) > 0 {
		subcommand := subcommandGroup.Options[0]
		
		switch subcommand.Name {
		case "create":
			// Get name and description from options
			var name, description string
			for _, opt := range subcommand.Options {
				switch opt.Name {
				case "name":
					name = opt.StringValue()
				case "description":
					description = opt.StringValue()
				}
			}
			req := &session.CreateRequest{
				Session:     s,
				Interaction: i,
				Name:        name,
				Description: description,
			}
			if err := h.sessionCreateHandler.Handle(req); err != nil {
				log.Printf("Error handling session create: %v", err)
			}
		case "list":
			req := &session.ListRequest{
				Session:     s,
				Interaction: i,
			}
			if err := h.sessionListHandler.Handle(req); err != nil {
				log.Printf("Error handling session list: %v", err)
			}
		case "join":
			// Get invite code from options
			var code string
			for _, opt := range subcommand.Options {
				if opt.Name == "code" {
					code = opt.StringValue()
					break
				}
			}
			req := &session.JoinRequest{
				Session:     s,
				Interaction: i,
				InviteCode:  code,
			}
			if err := h.sessionJoinHandler.Handle(req); err != nil {
				log.Printf("Error handling session join: %v", err)
			}
		case "info":
			req := &session.InfoRequest{
				Session:     s,
				Interaction: i,
			}
			if err := h.sessionInfoHandler.Handle(req); err != nil {
				log.Printf("Error handling session info: %v", err)
			}
		case "start":
			// Get session ID from options (optional)
			var sessionID string
			for _, opt := range subcommand.Options {
				if opt.Name == "id" {
					sessionID = opt.StringValue()
					break
				}
			}
			
			// If no session ID provided, try to find the user's session
			if sessionID == "" {
				sessions, err := h.ServiceProvider.SessionService.ListUserSessions(context.Background(), i.Member.User.ID)
				if err == nil && len(sessions) == 1 {
					sessionID = sessions[0].ID
				} else if len(sessions) > 1 {
					// User has multiple sessions, they need to specify
					err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: "❌ You have multiple sessions. Please specify the session ID.",
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					if err != nil {
						log.Printf("Error responding to start command: %v", err)
					}
					return
				}
			}
			
			req := &session.StartRequest{
				Session:     s,
				Interaction: i,
				SessionID:   sessionID,
			}
			if err := h.sessionStartHandler.Handle(req); err != nil {
				log.Printf("Error handling session start: %v", err)
			}
		case "end":
			// Get session ID from options (optional)
			var sessionID string
			for _, opt := range subcommand.Options {
				if opt.Name == "id" {
					sessionID = opt.StringValue()
					break
				}
			}
			
			// If no session ID provided, try to find the user's session
			if sessionID == "" {
				sessions, err := h.ServiceProvider.SessionService.ListUserSessions(context.Background(), i.Member.User.ID)
				if err == nil && len(sessions) == 1 {
					sessionID = sessions[0].ID
				} else if len(sessions) > 1 {
					// User has multiple sessions, they need to specify
					err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: "❌ You have multiple sessions. Please specify the session ID.",
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					if err != nil {
						log.Printf("Error responding to end command: %v", err)
					}
					return
				}
			}
			
			req := &session.EndRequest{
				Session:     s,
				Interaction: i,
				SessionID:   sessionID,
			}
			if err := h.sessionEndHandler.Handle(req); err != nil {
				log.Printf("Error handling session end: %v", err)
			}
		}
	} else if subcommandGroup.Name == "encounter" && len(subcommandGroup.Options) > 0 {
		subcommand := subcommandGroup.Options[0]
		
		switch subcommand.Name {
		case "add":
			// Get monster name from options
			var monsterQuery string
			for _, opt := range subcommand.Options {
				if opt.Name == "monster" {
					monsterQuery = opt.StringValue()
					break
				}
			}
			
			req := &encounter.AddMonsterRequest{
				Session:     s,
				Interaction: i,
				Query:       monsterQuery,
			}
			if err := h.encounterAddMonsterHandler.Handle(req); err != nil {
				log.Printf("Error handling add monster: %v", err)
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
			if len(parts) >= 4 {
				req := &character.AssignAbilitiesRequest{
					Session:     s,
					Interaction: i,
					RaceKey:     parts[2],
					ClassKey:    parts[3],
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
		case "auto_assign":
			// Auto-assign ability scores based on class
			if len(parts) >= 4 {
				req := &character.AssignAbilitiesRequest{
					Session:     s,
					Interaction: i,
					RaceKey:     parts[2],
					ClassKey:    parts[3],
					AutoAssign:  true,
				}
				if err := h.characterAssignAbilitiesHandler.Handle(req); err != nil {
					log.Printf("Error handling auto assign: %v", err)
				}
			}
		case "confirm_abilities":
			// Save ability scores to draft character before moving to proficiency choices
			if len(parts) >= 4 {
				// Parse ability scores from the current message embed
				abilityScores := make(map[string]int)
				if i.Message != nil && len(i.Message.Embeds) > 0 {
					embed := i.Message.Embeds[0]
					for _, field := range embed.Fields {
						if field.Name == "💪 Physical" || field.Name == "🧠 Mental" {
							// Parse ability scores from the field
							lines := strings.Split(field.Value, "\n")
							for _, line := range lines {
								// Parse lines like "**STR:** 17 (+2) = 19 [+4]" or "**STR:** 17 [+3]"
								if strings.Contains(line, ":") && !strings.Contains(line, "_Not assigned_") {
									parts := strings.Split(line, ":")
									ability := strings.Trim(strings.Trim(parts[0], "*"), " ")
									scoreStr := strings.TrimSpace(parts[1])
									// Extract just the base score
									if idx := strings.Index(scoreStr, " "); idx > 0 {
										scoreStr = scoreStr[:idx]
									}
									if score, err := strconv.Atoi(scoreStr); err == nil {
										abilityScores[ability] = score
									}
								}
							}
						}
					}
				}
				
				// Get draft character and update with ability scores
				log.Printf("Parsed ability scores: %+v", abilityScores)
				if len(abilityScores) == 6 {
					draftChar, err := h.ServiceProvider.CharacterService.GetOrCreateDraftCharacter(
						context.Background(),
						i.Member.User.ID,
						i.GuildID,
					)
					if err == nil {
						log.Printf("Updating draft character %s with ability scores", draftChar.ID)
						_, err = h.ServiceProvider.CharacterService.UpdateDraftCharacter(
							context.Background(),
							draftChar.ID,
							&characterService.UpdateDraftInput{
								AbilityScores: abilityScores,
							},
						)
						if err != nil {
							log.Printf("Error updating draft with ability scores: %v", err)
						} else {
							log.Printf("Successfully updated draft with ability scores")
						}
					} else {
						log.Printf("Error getting draft character: %v", err)
					}
				} else {
					log.Printf("Warning: Only parsed %d ability scores, expected 6", len(abilityScores))
				}
				
				// Move to proficiency choices
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
			// Handle proficiency selection and save to draft
			if len(parts) >= 6 {
				// Get the draft character
				draftChar, err := h.ServiceProvider.CharacterService.GetOrCreateDraftCharacter(
					context.Background(),
					i.Member.User.ID,
					i.GuildID,
				)
				if err != nil {
					log.Printf("Error getting draft character: %v", err)
				} else {
					// Get current proficiencies and add new ones
					selectedProfs := i.MessageComponentData().Values
					log.Printf("Selected proficiencies: %v", selectedProfs)
					
					// Get existing proficiencies if any
					existingProfs := []string{}
					if draftChar.Proficiencies != nil {
						for _, profList := range draftChar.Proficiencies {
							for _, prof := range profList {
								existingProfs = append(existingProfs, prof.Key)
							}
						}
					}
					
					// Combine existing and new proficiencies
					allProfs := append(existingProfs, selectedProfs...)
					
					// Update the draft with selected proficiencies
					_, err = h.ServiceProvider.CharacterService.UpdateDraftCharacter(
						context.Background(),
						draftChar.ID,
						&characterService.UpdateDraftInput{
							Proficiencies: allProfs,
						},
					)
					if err != nil {
						log.Printf("Error updating draft with proficiencies: %v", err)
					} else {
						log.Printf("Successfully updated draft with proficiencies")
					}
				}
				
				// Parse choice type and index
				choiceType := parts[4]
				choiceIndex, _ := strconv.Atoi(parts[5])
				
				// Check if there are more proficiency choices to make
				race, _ := h.ServiceProvider.CharacterService.GetRace(context.Background(), parts[2])
				class, _ := h.ServiceProvider.CharacterService.GetClass(context.Background(), parts[3])
				
				// Determine if we need to show more choices
				hasMoreChoices := false
				nextChoiceType := choiceType
				nextChoiceIndex := choiceIndex + 1
				
				if choiceType == "class" && class != nil {
					if nextChoiceIndex < len(class.ProficiencyChoices) {
						hasMoreChoices = true
					} else if race != nil && race.StartingProficiencyOptions != nil {
						// Move to race choices
						hasMoreChoices = true
						nextChoiceType = "race"
						nextChoiceIndex = 0
					}
				}
				
				if hasMoreChoices {
					// Show next proficiency choice
					req := &character.SelectProficienciesRequest{
						Session:     s,
						Interaction: i,
						RaceKey:     parts[2],
						ClassKey:    parts[3],
						ChoiceIndex: nextChoiceIndex,
						ChoiceType:  nextChoiceType,
					}
					if err := h.characterSelectProficienciesHandler.Handle(req); err != nil {
						log.Printf("Error handling next proficiency selection: %v", err)
					}
				} else {
					// All proficiencies selected, move to equipment choices
					req := &character.EquipmentChoicesRequest{
						Session:     s,
						Interaction: i,
						RaceKey:     parts[2],
						ClassKey:    parts[3],
					}
					if err := h.characterEquipmentChoicesHandler.Handle(req); err != nil {
						log.Printf("Error handling equipment choices: %v", err)
					}
				}
			}
		case "select_tool_type":
			// Handle selection of tool/instrument category (nested choices)
			// For now, since monk choices are already flattened by the choice resolver,
			// we can proceed directly to equipment choices
			if len(parts) >= 4 {
				req := &character.EquipmentChoicesRequest{
					Session:     s,
					Interaction: i,
					RaceKey:     parts[2],
					ClassKey:    parts[3],
				}
				if err := h.characterEquipmentChoicesHandler.Handle(req); err != nil {
					log.Printf("Error handling equipment choices: %v", err)
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
		case "select_equipment":
			// Show equipment selection interface
			if len(parts) >= 4 {
				req := &character.SelectEquipmentRequest{
					Session:     s,
					Interaction: i,
					RaceKey:     parts[2],
					ClassKey:    parts[3],
					ChoiceIndex: 0,
				}
				if err := h.characterSelectEquipmentHandler.Handle(req); err != nil {
					log.Printf("Error handling select equipment: %v", err)
				}
			}
		case "confirm_equipment":
			// Handle equipment selection confirmation
			if len(parts) >= 5 {
				choiceIndex, _ := strconv.Atoi(parts[4])
				
				// Check if this is a nested choice selection
				selectedValues := i.MessageComponentData().Values
				if len(selectedValues) > 0 && strings.HasPrefix(selectedValues[0], "nested-") {
					// This is a bundle with nested choices - need to expand
					log.Printf("Nested choice selected: %v", selectedValues[0])
					
					// Get the equipment choices to find the actual selection details
					choices, err := h.ServiceProvider.CharacterService.ResolveChoices(
						context.Background(),
						&characterService.ResolveChoicesInput{
							RaceKey:  parts[2],
							ClassKey: parts[3],
						},
					)
					
					selectionCount := 1
					category := "martial-weapons"
					
					if err == nil && choiceIndex < len(choices.EquipmentChoices) {
						// Find the selected option in the choice
						choice := choices.EquipmentChoices[choiceIndex]
						for _, opt := range choice.Options {
							if opt.Key == selectedValues[0] {
								// Parse the description to get the count
								if strings.Contains(opt.Description, "Choose 2") || strings.Contains(opt.Name, "2") || strings.Contains(opt.Name, "two") {
									selectionCount = 2
								}
								// Could also parse category from description if needed
								break
							}
						}
					}
					
					// Show martial weapon selection UI
					req := &character.SelectNestedEquipmentRequest{
						Session:        s,
						Interaction:    i,
						RaceKey:        parts[2],
						ClassKey:       parts[3],
						ChoiceIndex:    choiceIndex,
						BundleKey:      selectedValues[0],
						SelectionCount: selectionCount,
						Category:       category,
					}
					if err := h.characterSelectNestedEquipmentHandler.Handle(req); err != nil {
						log.Printf("Error handling nested equipment selection: %v", err)
					}
					return
				}
				
				// Get the draft character and update with selected equipment
				draftChar, err := h.ServiceProvider.CharacterService.GetOrCreateDraftCharacter(
					context.Background(),
					i.Member.User.ID,
					i.GuildID,
				)
				if err == nil {
					log.Printf("Selected equipment: %v", selectedValues)
					
					// Get existing equipment if any
					existingEquipment := []string{}
					// Note: We would need to track equipment selections in the draft character
					// For now, just use the new selections
					
					// Update the draft with selected equipment
					_, err = h.ServiceProvider.CharacterService.UpdateDraftCharacter(
						context.Background(),
						draftChar.ID,
						&characterService.UpdateDraftInput{
							Equipment: append(existingEquipment, selectedValues...),
						},
					)
					if err != nil {
						log.Printf("Error updating draft with equipment: %v", err)
					} else {
						log.Printf("Successfully updated draft with equipment")
					}
				} else {
					log.Printf("Error getting draft character: %v", err)
				}
				
				// Move to next equipment choice or character details
				req := &character.SelectEquipmentRequest{
					Session:     s,
					Interaction: i,
					RaceKey:     parts[2],
					ClassKey:    parts[3],
					ChoiceIndex: choiceIndex + 1,
				}
				if err := h.characterSelectEquipmentHandler.Handle(req); err != nil {
					log.Printf("Error handling next equipment selection: %v", err)
				}
			}
		case "confirm_nested_equipment":
			// Handle nested equipment selection (e.g., selecting specific martial weapons)
			if len(parts) >= 6 {
				choiceIndex, _ := strconv.Atoi(parts[4])
				bundleKey := parts[5]
				
				// Get selected weapons
				selectedWeapons := i.MessageComponentData().Values
				log.Printf("Selected weapons: %v for bundle: %s", selectedWeapons, bundleKey)
				
				// Check for duplicate selections
				weaponMap := make(map[string]bool)
				for _, weapon := range selectedWeapons {
					if weaponMap[weapon] {
						// Duplicate found - show error
						content := "❌ You cannot select the same weapon twice. Please choose different weapons."
						_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
							Content: &content,
						})
						if err != nil {
							log.Printf("Error sending duplicate weapon error: %v", err)
						}
						return
					}
					weaponMap[weapon] = true
				}
				
				// Get the draft character
				draftChar, err := h.ServiceProvider.CharacterService.GetOrCreateDraftCharacter(
					context.Background(),
					i.Member.User.ID,
					i.GuildID,
				)
				if err == nil {
					// Get existing equipment
					existingEquipment := []string{}
					// TODO: Track equipment properly in draft
					
					// Add selected weapons
					allEquipment := append(existingEquipment, selectedWeapons...)
					
					// If this was "weapon + shield" choice, add shield
					if strings.Contains(strings.ToLower(bundleKey), "shield") && len(selectedWeapons) == 1 {
						allEquipment = append(allEquipment, "shield")
					}
					
					// Update draft with equipment
					_, err = h.ServiceProvider.CharacterService.UpdateDraftCharacter(
						context.Background(),
						draftChar.ID,
						&characterService.UpdateDraftInput{
							Equipment: allEquipment,
						},
					)
					if err != nil {
						log.Printf("Error updating draft with nested equipment: %v", err)
					}
				}
				
				// Continue to next equipment choice or character details
				// Since the interaction is already acknowledged, we need to check if there are more choices
				choices, err := h.ServiceProvider.CharacterService.ResolveChoices(
					context.Background(),
					&characterService.ResolveChoicesInput{
						RaceKey:  parts[2],
						ClassKey: parts[3],
					},
				)
				if err == nil && choiceIndex+1 < len(choices.EquipmentChoices) {
					// More equipment choices available
					req := &character.SelectEquipmentRequest{
						Session:     s,
						Interaction: i,
						RaceKey:     parts[2],
						ClassKey:    parts[3],
						ChoiceIndex: choiceIndex + 1,
					}
					if err := h.characterSelectEquipmentHandler.Handle(req); err != nil {
						log.Printf("Error handling next equipment selection: %v", err)
					}
				} else {
					// No more equipment choices, go to character details
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
	} else if ctx == "character_manage" {
		// Handle character management actions (edit, archive, delete, etc.)
		if len(parts) >= 3 {
			characterID := parts[2]
			
			switch action {
			case "delete":
				// Confirm deletion with a modal or direct action
				err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "⚠️ Are you sure you want to delete this character? This action cannot be undone!",
						Flags:   discordgo.MessageFlagsEphemeral,
						Components: []discordgo.MessageComponent{
							discordgo.ActionsRow{
								Components: []discordgo.MessageComponent{
									discordgo.Button{
										Label:    "Confirm Delete",
										Style:    discordgo.DangerButton,
										CustomID: fmt.Sprintf("character_manage:confirm_delete:%s", characterID),
										Emoji: &discordgo.ComponentEmoji{
											Name: "🗑️",
										},
									},
									discordgo.Button{
										Label:    "Cancel",
										Style:    discordgo.SecondaryButton,
										CustomID: "character_manage:cancel",
										Emoji: &discordgo.ComponentEmoji{
											Name: "❌",
										},
									},
								},
							},
						},
					},
				})
				if err != nil {
					log.Printf("Error showing delete confirmation: %v", err)
				}
			case "confirm_delete":
				// Actually delete the character
				err := h.ServiceProvider.CharacterService.Delete(characterID)
				if err != nil {
					content := fmt.Sprintf("❌ Failed to delete character: %v", err)
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: content,
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					return
				}
				
				err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseUpdateMessage,
					Data: &discordgo.InteractionResponseData{
						Content: "✅ Character successfully deleted.",
						Components: []discordgo.MessageComponent{},
					},
				})
				if err != nil {
					log.Printf("Error confirming deletion: %v", err)
				}
			case "cancel":
				// Cancel the action
				err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseUpdateMessage,
					Data: &discordgo.InteractionResponseData{
						Content: "❌ Action cancelled.",
						Components: []discordgo.MessageComponent{},
					},
				})
				if err != nil {
					log.Printf("Error cancelling action: %v", err)
				}
			case "archive":
				// Archive the character
				err := h.ServiceProvider.CharacterService.UpdateStatus(characterID, entities.CharacterStatusArchived)
				if err != nil {
					content := fmt.Sprintf("❌ Failed to archive character: %v", err)
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: content,
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					return
				}
				
				err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "✅ Character archived successfully! Use `/dnd character list` to see all your characters.",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				if err != nil {
					log.Printf("Error archiving character: %v", err)
				}
			case "restore":
				// Restore archived character to active
				err := h.ServiceProvider.CharacterService.UpdateStatus(characterID, entities.CharacterStatusActive)
				if err != nil {
					content := fmt.Sprintf("❌ Failed to restore character: %v", err)
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: content,
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					return
				}
				
				err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "✅ Character restored to active status!",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				if err != nil {
					log.Printf("Error restoring character: %v", err)
				}
			}
		}
	} else if ctx == "session_manage" {
		// Handle session management actions
		if len(parts) >= 3 {
			sessionID := parts[2]
			
			switch action {
			case "start":
				// Start the session
				err := h.ServiceProvider.SessionService.StartSession(context.Background(), sessionID, i.Member.User.ID)
				if err != nil {
					content := fmt.Sprintf("❌ Failed to start session: %v", err)
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: content,
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					return
				}
				
				err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "🎲 Session started! Let the adventure begin!",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				if err != nil {
					log.Printf("Error starting session: %v", err)
				}
			case "leave":
				// Leave the session
				err := h.ServiceProvider.SessionService.LeaveSession(context.Background(), sessionID, i.Member.User.ID)
				if err != nil {
					content := fmt.Sprintf("❌ Failed to leave session: %v", err)
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: content,
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					return
				}
				
				err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "👋 You've left the session.",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				if err != nil {
					log.Printf("Error leaving session: %v", err)
				}
			case "select_character":
				// Show character selection menu
				// Get user's characters
				chars, err := h.ServiceProvider.CharacterService.ListByOwner(i.Member.User.ID)
				if err != nil || len(chars) == 0 {
					content := "❌ You need to create a character first! Use `/dnd character create`"
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: content,
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					return
				}
				
				// Build character options
				options := make([]discordgo.SelectMenuOption, 0, len(chars))
				for _, char := range chars {
					if char.Status == entities.CharacterStatusActive {
						options = append(options, discordgo.SelectMenuOption{
							Label:       fmt.Sprintf("%s - %s %s", char.Name, char.Race.Name, char.Class.Name),
							Description: fmt.Sprintf("Level %d | HP: %d/%d | AC: %d", char.Level, char.CurrentHitPoints, char.MaxHitPoints, char.AC),
							Value:       char.ID,
						})
					}
				}
				
				if len(options) == 0 {
					content := "❌ You don't have any active characters! Create or activate a character first."
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: content,
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					return
				}
				
				// Show character selection menu
				err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "🎭 Select your character for this session:",
						Flags:   discordgo.MessageFlagsEphemeral,
						Components: []discordgo.MessageComponent{
							discordgo.ActionsRow{
								Components: []discordgo.MessageComponent{
									discordgo.SelectMenu{
										CustomID:    fmt.Sprintf("session_manage:confirm_character:%s", sessionID),
										Placeholder: "Choose your character...",
										Options:     options,
									},
								},
							},
						},
					},
				})
				if err != nil {
					log.Printf("Error showing character selection: %v", err)
				}
			case "confirm_character":
				// Set the selected character
				if len(i.MessageComponentData().Values) > 0 {
					characterID := i.MessageComponentData().Values[0]
					err := h.ServiceProvider.SessionService.SelectCharacter(context.Background(), sessionID, i.Member.User.ID, characterID)
					if err != nil {
						content := fmt.Sprintf("❌ Failed to select character: %v", err)
						s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
							Type: discordgo.InteractionResponseUpdateMessage,
							Data: &discordgo.InteractionResponseData{
								Content: content,
								Components: []discordgo.MessageComponent{},
							},
						})
						return
					}
					
					// Get character details for confirmation
					char, _ := h.ServiceProvider.CharacterService.GetByID(characterID)
					content := fmt.Sprintf("✅ Character selected: **%s** the %s %s", char.Name, char.Race.Name, char.Class.Name)
					err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseUpdateMessage,
						Data: &discordgo.InteractionResponseData{
							Content: content,
							Components: []discordgo.MessageComponent{},
						},
					})
					if err != nil {
						log.Printf("Error confirming character selection: %v", err)
					}
				}
			case "pause":
				// Pause the session
				err := h.ServiceProvider.SessionService.PauseSession(context.Background(), sessionID, i.Member.User.ID)
				if err != nil {
					content := fmt.Sprintf("❌ Failed to pause session: %v", err)
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: content,
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					return
				}
				
				err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "⏸️ Session paused. Use `/dnd session info` to see options.",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				if err != nil {
					log.Printf("Error pausing session: %v", err)
				}
			case "end":
				// Use the end handler
				req := &session.EndRequest{
					Session:     s,
					Interaction: i,
					SessionID:   sessionID,
				}
				if err := h.sessionEndHandler.Handle(req); err != nil {
					log.Printf("Error handling session end button: %v", err)
				}
			case "resume":
				// Resume the session
				err := h.ServiceProvider.SessionService.ResumeSession(context.Background(), sessionID, i.Member.User.ID)
				if err != nil {
					content := fmt.Sprintf("❌ Failed to resume session: %v", err)
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: content,
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					return
				}
				
				err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "▶️ Session resumed! The adventure continues!",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				if err != nil {
					log.Printf("Error resuming session: %v", err)
				}
			case "invite":
				// Show invite interface
				// Get session details
				session, err := h.ServiceProvider.SessionService.GetSession(context.Background(), sessionID)
				if err != nil {
					content := fmt.Sprintf("❌ Failed to get session: %v", err)
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: content,
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					return
				}
				
				// Show invite code and instructions
				content := fmt.Sprintf("📨 **Invite Players to %s**\n\n🔑 Invite Code: `%s`\n\nPlayers can join using:\n```/dnd session join %s```", 
					session.Name, session.InviteCode, session.InviteCode)
				err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: content,
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				if err != nil {
					log.Printf("Error showing invite: %v", err)
				}
			case "settings":
				// TODO: Implement session settings modal
				content := "🔧 Session settings coming soon!"
				err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: content,
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				if err != nil {
					log.Printf("Error showing settings placeholder: %v", err)
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
			
			// Get the draft character
			draftChar, err := h.ServiceProvider.CharacterService.GetOrCreateDraftCharacter(
				context.Background(),
				i.Member.User.ID,
				i.GuildID,
			)
			if err != nil {
				log.Printf("Error getting draft character: %v", err)
				// Show error to user
				err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "❌ Failed to get character draft. Please try again.",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				if err != nil {
					log.Printf("Error responding with error: %v", err)
				}
				return
			}
			
			// Update the draft with the final name
			_, err = h.ServiceProvider.CharacterService.UpdateDraftCharacter(
				context.Background(),
				draftChar.ID,
				&characterService.UpdateDraftInput{
					Name: &characterName,
				},
			)
			if err != nil {
				log.Printf("Error updating draft with name: %v", err)
			}
			
			// Get the updated character to ensure we have all the data
			char, err := h.ServiceProvider.CharacterService.GetCharacter(context.Background(), draftChar.ID)
			if err != nil {
				log.Printf("Error getting updated character: %v", err)
				// Show error to user
				err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "❌ Failed to finalize character. Please try again.",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				if err != nil {
					log.Printf("Error responding with error: %v", err)
				}
				return
			}
			
			// Build ability scores from the character
			abilityScores := make(map[string]int)
			log.Printf("Character attributes: %+v", char.Attributes)
			if char.Attributes != nil {
				for attr, score := range char.Attributes {
					var key string
					switch attr {
					case entities.AttributeStrength:
						key = "STR"
					case entities.AttributeDexterity:
						key = "DEX"
					case entities.AttributeConstitution:
						key = "CON"
					case entities.AttributeIntelligence:
						key = "INT"
					case entities.AttributeWisdom:
						key = "WIS"
					case entities.AttributeCharisma:
						key = "CHA"
					}
					if key != "" && score != nil {
						abilityScores[key] = score.Score
						log.Printf("Attribute %s (%s): %d", key, attr, score.Score)
					}
				}
			}
			log.Printf("Final ability scores map: %+v", abilityScores)
			
			// Collect proficiencies
			proficiencies := []string{}
			if char.Proficiencies != nil {
				for _, profList := range char.Proficiencies {
					for _, prof := range profList {
						if prof != nil {
							proficiencies = append(proficiencies, prof.Key)
						}
					}
				}
			}
			
			// Finalize the draft character (marking it as active)
			finalChar, err := h.ServiceProvider.CharacterService.FinalizeDraftCharacter(context.Background(), draftChar.ID)
			if err != nil {
				log.Printf("Error finalizing character: %v", err)
				// Show error to user
				err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: fmt.Sprintf("❌ Failed to finalize character: %v", err),
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
				Description: fmt.Sprintf("**Name:** %s\n**Race:** %s\n**Class:** %s", finalChar.Name, finalChar.Race.Name, finalChar.Class.Name),
				Color:       0x00ff00,
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "💪 Base Abilities",
						Value:  fmt.Sprintf("STR: %d, DEX: %d, CON: %d\nINT: %d, WIS: %d, CHA: %d",
							abilityScores["STR"], abilityScores["DEX"], abilityScores["CON"],
							abilityScores["INT"], abilityScores["WIS"], abilityScores["CHA"],
						),
						Inline: true,
					},
					{
						Name:   "❤️ Hit Points",
						Value:  fmt.Sprintf("%d", finalChar.MaxHitPoints),
						Inline: true,
					},
					{
						Name:   "🛡️ Hit Die",
						Value:  fmt.Sprintf("d%d", finalChar.HitDie),
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