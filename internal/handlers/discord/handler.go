package discord

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strconv"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/dice"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/handlers/discord/combat"
	"github.com/KirkDiggler/dnd-bot-discord/internal/handlers/discord/dnd/character"
	oldcombat "github.com/KirkDiggler/dnd-bot-discord/internal/handlers/discord/dnd/combat"
	"github.com/KirkDiggler/dnd-bot-discord/internal/handlers/discord/dnd/dungeon"
	encounterHandler "github.com/KirkDiggler/dnd-bot-discord/internal/handlers/discord/dnd/encounter"
	"github.com/KirkDiggler/dnd-bot-discord/internal/handlers/discord/dnd/help"
	sessionHandler "github.com/KirkDiggler/dnd-bot-discord/internal/handlers/discord/dnd/session"
	"github.com/KirkDiggler/dnd-bot-discord/internal/handlers/discord/dnd/testcombat"
	"github.com/KirkDiggler/dnd-bot-discord/internal/handlers/discord/helpers"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services"
	characterService "github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/encounter"
	"github.com/bwmarrin/discordgo"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Handler handles all Discord interactions
type Handler struct {
	ServiceProvider                       *services.Provider
	characterCreateHandler                *character.CreateHandler
	characterRaceSelectHandler            *character.RaceSelectHandler
	characterShowClassesHandler           *character.ShowClassesHandler
	characterClassSelectHandler           *character.ClassSelectHandler
	characterAbilityScoresHandler         *character.AbilityScoresHandler
	characterRollAllHandler               *character.RollAllHandler
	characterRollIndividualHandler        *character.RollIndividualHandler
	characterAssignAbilitiesHandler       *character.AssignAbilitiesHandler
	characterProficiencyChoicesHandler    *character.ProficiencyChoicesHandler
	characterSelectProficienciesHandler   *character.SelectProficienciesHandler
	characterEquipmentChoicesHandler      *character.EquipmentChoicesHandler
	characterSelectEquipmentHandler       *character.SelectEquipmentHandler
	characterSelectNestedEquipmentHandler *character.SelectNestedEquipmentHandler
	characterDetailsHandler               *character.CharacterDetailsHandler
	characterListHandler                  *character.ListHandler
	characterShowHandler                  *character.ShowHandler
	characterWeaponHandler                *character.WeaponHandler
	characterSheetHandler                 *character.SheetHandler
	characterDeleteHandler                *character.DeleteHandler

	// Session handlers
	sessionCreateHandler *sessionHandler.CreateHandler
	sessionListHandler   *sessionHandler.ListHandler
	sessionJoinHandler   *sessionHandler.JoinHandler
	sessionStartHandler  *sessionHandler.StartHandler
	sessionEndHandler    *sessionHandler.EndHandler
	sessionInfoHandler   *sessionHandler.InfoHandler
	sessionLeaveHandler  *sessionHandler.LeaveHandler

	// Encounter handlers
	encounterAddMonsterHandler *encounterHandler.AddMonsterHandler

	// Test combat handler
	testCombatHandler *testcombat.TestCombatHandler

	// Dungeon handlers
	dungeonStartHandler     *dungeon.StartDungeonHandler
	dungeonJoinHandler      *dungeon.JoinPartyHandler
	dungeonEnterRoomHandler *dungeon.EnterRoomHandler

	// Help handler
	helpHandler *help.HelpHandler

	// Combat handlers
	savingThrowHandler *oldcombat.SavingThrowHandler
	skillCheckHandler  *oldcombat.SkillCheckHandler
	combatHandler      *combat.Handler
}

// HandlerConfig holds configuration for the Discord handler
type HandlerConfig struct {
	ServiceProvider *services.Provider
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
		characterRollAllHandler: character.NewRollAllHandler(&character.RollAllHandlerConfig{
			CharacterService: cfg.ServiceProvider.CharacterService,
		}),
		characterRollIndividualHandler: character.NewRollIndividualHandler(&character.RollIndividualHandlerConfig{
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
		characterWeaponHandler: character.NewWeaponHandler(&character.WeaponHandlerConfig{
			ServiceProvider: cfg.ServiceProvider,
		}),
		characterSheetHandler:  character.NewSheetHandler(cfg.ServiceProvider),
		characterDeleteHandler: character.NewDeleteHandler(cfg.ServiceProvider),

		// Initialize session handlers
		sessionCreateHandler: sessionHandler.NewCreateHandler(cfg.ServiceProvider),
		sessionListHandler:   sessionHandler.NewListHandler(cfg.ServiceProvider),
		sessionJoinHandler:   sessionHandler.NewJoinHandler(cfg.ServiceProvider),
		sessionStartHandler:  sessionHandler.NewStartHandler(cfg.ServiceProvider),
		sessionEndHandler:    sessionHandler.NewEndHandler(cfg.ServiceProvider),
		sessionInfoHandler:   sessionHandler.NewInfoHandler(cfg.ServiceProvider),
		sessionLeaveHandler:  sessionHandler.NewLeaveHandler(cfg.ServiceProvider),

		// Initialize encounter handlers
		encounterAddMonsterHandler: encounterHandler.NewAddMonsterHandler(cfg.ServiceProvider),

		// Initialize test combat handler
		testCombatHandler: testcombat.NewTestCombatHandler(cfg.ServiceProvider),

		// Initialize dungeon handlers
		dungeonStartHandler:     dungeon.NewStartDungeonHandler(cfg.ServiceProvider),
		dungeonJoinHandler:      dungeon.NewJoinPartyHandler(cfg.ServiceProvider),
		dungeonEnterRoomHandler: dungeon.NewEnterRoomHandler(cfg.ServiceProvider),

		// Initialize help handler
		helpHandler: help.NewHelpHandler(),

		// Initialize combat handlers
		savingThrowHandler: oldcombat.NewSavingThrowHandler(&oldcombat.SavingThrowHandlerConfig{
			CharacterService: cfg.ServiceProvider.CharacterService,
			EncounterService: cfg.ServiceProvider.EncounterService,
		}),
		skillCheckHandler: oldcombat.NewSkillCheckHandler(&oldcombat.SkillCheckHandlerConfig{
			CharacterService: cfg.ServiceProvider.CharacterService,
		}),
		combatHandler: combat.NewHandler(cfg.ServiceProvider.EncounterService),
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
							Name:        "delete",
							Description: "Delete one of your characters",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
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
						{
							Name:        "leave",
							Description: "Leave all your active sessions",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
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
				{
					Name:        "testcombat",
					Description: "Start a quick test combat session (bot as DM)",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "monster",
							Description: "Monster to fight (default: goblin)",
							Required:    false,
						},
					},
				},
				{
					Name:        "dungeon",
					Description: "Start a cooperative dungeon delve with friends",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "difficulty",
							Description: "Dungeon difficulty (easy, medium, hard)",
							Required:    false,
							Choices: []*discordgo.ApplicationCommandOptionChoice{
								{Name: "Easy", Value: "easy"},
								{Name: "Medium", Value: "medium"},
								{Name: "Hard", Value: "hard"},
							},
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
		case "testcombat":
			// Get monster from options if provided
			var monster string
			for _, opt := range subcommandGroup.Options {
				if opt.Name == "monster" {
					monster = opt.StringValue()
					break
				}
			}
			req := &testcombat.TestCombatRequest{
				Session:     s,
				Interaction: i,
				MonsterName: monster,
			}
			if err := h.testCombatHandler.Handle(req); err != nil {
				log.Printf("Error handling test combat: %v", err)
			}
		case "dungeon":
			// Get difficulty from options if provided
			difficulty := "medium" // default
			for _, opt := range subcommandGroup.Options {
				if opt.Name == "difficulty" {
					difficulty = opt.StringValue()
					break
				}
			}
			req := &dungeon.StartDungeonRequest{
				Session:     s,
				Interaction: i,
				Difficulty:  difficulty,
			}
			if err := h.dungeonStartHandler.Handle(req); err != nil {
				log.Printf("Error handling dungeon start: %v", err)
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
		case "delete":
			req := &character.DeleteRequest{
				Session:     s,
				Interaction: i,
			}
			if err := h.characterDeleteHandler.Handle(req); err != nil {
				log.Printf("Error handling character delete: %v", err)
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
			req := &sessionHandler.CreateRequest{
				Session:     s,
				Interaction: i,
				Name:        name,
				Description: description,
			}
			if err := h.sessionCreateHandler.Handle(req); err != nil {
				log.Printf("Error handling session create: %v", err)
			}
		case "list":
			req := &sessionHandler.ListRequest{
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
			req := &sessionHandler.JoinRequest{
				Session:     s,
				Interaction: i,
				InviteCode:  code,
			}
			if err := h.sessionJoinHandler.Handle(req); err != nil {
				log.Printf("Error handling session join: %v", err)
			}
		case "info":
			req := &sessionHandler.InfoRequest{
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

			req := &sessionHandler.StartRequest{
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

			req := &sessionHandler.EndRequest{
				Session:     s,
				Interaction: i,
				SessionID:   sessionID,
			}
			if err := h.sessionEndHandler.Handle(req); err != nil {
				log.Printf("Error handling session end: %v", err)
			}
		case "leave":
			req := &sessionHandler.LeaveRequest{
				Session:     s,
				Interaction: i,
			}
			if err := h.sessionLeaveHandler.Handle(req); err != nil {
				log.Printf("Error handling session leave: %v", err)
			}
		}
	} else if subcommandGroup.Name == "encounter" && len(subcommandGroup.Options) > 0 {
		subcommand := subcommandGroup.Options[0]

		if subcommand.Name == "add" {
			// Get monster name from options
			var monsterQuery string
			for _, opt := range subcommand.Options {
				if opt.Name == "monster" {
					monsterQuery = opt.StringValue()
					break
				}
			}

			req := &encounterHandler.AddMonsterRequest{
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
				// Update draft with race and class when we reach ability scores
				draftChar, err := h.ServiceProvider.CharacterService.GetOrCreateDraftCharacter(
					context.Background(),
					i.Member.User.ID,
					i.GuildID,
				)
				if err == nil {
					raceKey := parts[2]
					classKey := parts[3]
					_, err = h.ServiceProvider.CharacterService.UpdateDraftCharacter(
						context.Background(),
						draftChar.ID,
						&characterService.UpdateDraftInput{
							RaceKey:  &raceKey,
							ClassKey: &classKey,
						},
					)
					if err != nil {
						log.Printf("Error updating draft with race/class: %v", err)
					}
				}

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
		case "roll_all":
			if len(parts) >= 4 {
				req := &character.RollAllRequest{
					Session:     s,
					Interaction: i,
					RaceKey:     parts[2],
					ClassKey:    parts[3],
				}
				if err := h.characterRollAllHandler.Handle(req); err != nil {
					log.Printf("Error handling roll all: %v", err)
				}
			}
		case "roll_individual":
			if len(parts) >= 4 {
				rollIndex := 0
				if len(parts) >= 5 {
					if idx, err := strconv.Atoi(parts[4]); err == nil {
						rollIndex = idx
					}
				}
				req := &character.RollIndividualRequest{
					Session:     s,
					Interaction: i,
					RaceKey:     parts[2],
					ClassKey:    parts[3],
					RollIndex:   rollIndex,
				}
				if err := h.characterRollIndividualHandler.Handle(req); err != nil {
					log.Printf("Error handling roll individual: %v", err)
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
				// First check if we already have ability scores saved
				draftChar, err := h.ServiceProvider.CharacterService.GetOrCreateDraftCharacter(
					context.Background(),
					i.Member.User.ID,
					i.GuildID,
				)

				// If we already have ability scores, skip parsing and go straight to proficiencies
				if err == nil && draftChar.Attributes != nil && len(draftChar.Attributes) == 6 {
					log.Printf("Ability scores already saved, moving to proficiencies")
					req := &character.ProficiencyChoicesRequest{
						Session:     s,
						Interaction: i,
						RaceKey:     parts[2],
						ClassKey:    parts[3],
					}
					if err := h.characterProficiencyChoicesHandler.Handle(req); err != nil {
						log.Printf("Error handling confirm abilities: %v", err)
					}
					return
				}

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
								if !strings.Contains(line, ":") && strings.Contains(line, "_Not assigned_") {
									continue
								}

								subParts := strings.Split(line, ":")
								ability := strings.Trim(strings.Trim(subParts[0], "*"), " ")
								scoreStr := strings.TrimSpace(subParts[1])
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

				// Get draft character and update with ability scores
				log.Printf("Parsed ability scores: %+v", abilityScores)
				if len(abilityScores) == 6 {
					draftChar, err := h.ServiceProvider.CharacterService.GetOrCreateDraftCharacter(
						context.Background(),
						i.Member.User.ID,
						i.GuildID,
					)
					if err == nil {
						log.Printf("Updating draft character %s with ability scores and race/class", draftChar.ID)
						raceKey := parts[2]
						classKey := parts[3]
						_, err = h.ServiceProvider.CharacterService.UpdateDraftCharacter(
							context.Background(),
							draftChar.ID,
							&characterService.UpdateDraftInput{
								AbilityScores: abilityScores,
								RaceKey:       &raceKey,
								ClassKey:      &classKey,
							},
						)
						if err != nil {
							log.Printf("Error updating draft with ability scores: %v", err)
						} else {
							log.Printf("Successfully updated draft with ability scores and race/class")
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
					allProfs := slices.Concat(existingProfs, selectedProfs)

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
				choiceIndex, err := strconv.Atoi(parts[5])
				if err != nil {
					log.Printf("Error converting choice index to int: %v", err)
					return
				}

				// Check if there are more proficiency choices to make
				race, err := h.ServiceProvider.CharacterService.GetRace(context.Background(), parts[2])
				if err != nil {
					log.Printf("Error getting race: %v", err)
					return
				}
				class, err := h.ServiceProvider.CharacterService.GetClass(context.Background(), parts[3])
				if err != nil {
					log.Printf("Error getting class: %v", err)
					return
				}

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
				choiceIndex, err := strconv.Atoi(parts[4])
				if err != nil {
					log.Printf("Error converting choice index to int: %v", err)
					return
				}

				// Check if this is a nested choice selection
				selectedValues := i.MessageComponentData().Values
				if len(selectedValues) > 0 && strings.HasPrefix(selectedValues[0], "nested-") {
					// This is a bundle with nested choices - need to expand
					log.Printf("Nested choice selected: %v", selectedValues[0])

					// Get the equipment choices to find the actual selection details
					choices, resolveErr := h.ServiceProvider.CharacterService.ResolveChoices(
						context.Background(),
						&characterService.ResolveChoicesInput{
							RaceKey:  parts[2],
							ClassKey: parts[3],
						},
					)

					selectionCount := 1
					category := "martial-weapons"

					if resolveErr == nil && choiceIndex < len(choices.EquipmentChoices) {
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
					if selectNestedErr := h.characterSelectNestedEquipmentHandler.Handle(req); selectNestedErr != nil {
						log.Printf("Error handling nested equipment selection: %v", selectNestedErr)
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
					_, updateDraftErr := h.ServiceProvider.CharacterService.UpdateDraftCharacter(
						context.Background(),
						draftChar.ID,
						&characterService.UpdateDraftInput{
							Equipment: append(existingEquipment, selectedValues...),
						},
					)
					if updateDraftErr != nil {
						log.Printf("Error updating draft with equipment: %v", updateDraftErr)
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
				if selectErr := h.characterSelectEquipmentHandler.Handle(req); selectErr != nil {
					log.Printf("Error handling next equipment selection: %v", selectErr)
				}
			}
		case "confirm_nested_equipment":
			// Handle nested equipment selection (e.g., selecting specific martial weapons)
			if len(parts) >= 6 {
				choiceIndex, err := strconv.Atoi(parts[4])
				if err != nil {
					log.Printf("Error converting choice index to int: %v", err)
					return
				}
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
						_, editErr := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
							Content: &content,
						})
						if editErr != nil {
							log.Printf("Error sending duplicate weapon error: %v", editErr)
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
					allEquipment := slices.Concat(existingEquipment, selectedWeapons)

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
					err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: content,
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					if err != nil {
						log.Printf("Error confirming deletion: %v", err)
					}
					return
				}

				err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseUpdateMessage,
					Data: &discordgo.InteractionResponseData{
						Content:    "✅ Character successfully deleted.",
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
						Content:    "❌ Action cancelled.",
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
					err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: content,
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					if err != nil {
						log.Printf("Error responding to archive character: %v", err)
					}
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
					err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: content,
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					if err != nil {
						log.Printf("Error responding to restore character: %v", err)
					}
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
			case "edit":
				// Edit character - for now, redirect to character creation flow
				// Get the character first to validate ownership
				char, err := h.ServiceProvider.CharacterService.GetByID(characterID)
				if err != nil {
					content := fmt.Sprintf("❌ Failed to get character: %v", err)
					err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: content,
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					if err != nil {
						log.Printf("Error responding to edit character: %v", err)
					}
					return
				}

				// Verify ownership
				if char.OwnerID != i.Member.User.ID {
					err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: "❌ You can only edit your own characters!",
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					if err != nil {
						log.Printf("Error responding to edit character: %v", err)
					}
					return
				}

				// For now, show a message about what can be edited
				// In the future, this could launch an edit flow
				embed := &discordgo.MessageEmbed{
					Title:       fmt.Sprintf("Edit %s", char.Name),
					Description: "Character editing is currently limited. You can:",
					Color:       0x3498db,
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:   "Available Actions",
							Value:  "• Archive/Restore the character\n• Delete the character\n• Create a new character with updated info",
							Inline: false,
						},
						{
							Name:   "Coming Soon",
							Value:  "• Edit ability scores\n• Change equipment\n• Update proficiencies\n• Level up",
							Inline: false,
						},
					},
					Footer: &discordgo.MessageEmbedFooter{
						Text: "Full character editing will be available in a future update",
					},
				}

				err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Embeds: []*discordgo.MessageEmbed{embed},
						Flags:  discordgo.MessageFlagsEphemeral,
					},
				})
				if err != nil {
					log.Printf("Error showing edit info: %v", err)
				}
			case "continue":
				// Continue creating a draft character
				// Get the character first to validate it's a draft
				char, err := h.ServiceProvider.CharacterService.GetByID(characterID)
				if err != nil {
					content := fmt.Sprintf("❌ Failed to get character: %v", err)
					err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: content,
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					if err != nil {
						log.Printf("Error responding to continue request: %v", err)
					}
					return
				}

				// Verify ownership
				if char.OwnerID != i.Member.User.ID {
					err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: "❌ You can only continue your own draft characters!",
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					if err != nil {
						log.Printf("Error responding to ownership check: %v", err)
					}
					return
				}

				// Verify it's a draft
				if char.Status != entities.CharacterStatusDraft {
					err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: "❌ This character is not a draft!",
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					if err != nil {
						log.Printf("Error responding to draft check: %v", err)
					}
					return
				}

				// Analyze the draft to determine where to resume
				// Follow the new step order: Race -> Class -> Abilities -> Proficiencies -> Equipment -> Features -> Name
				if char.Race == nil {
					// Start from race selection
					req := &character.CreateRequest{
						Session:     s,
						Interaction: i,
					}
					if err := h.characterCreateHandler.Handle(req); err != nil {
						log.Printf("Error resuming character creation at race selection: %v", err)
					}
				} else if char.Class == nil {
					// Continue from class selection
					req := &character.ShowClassesRequest{
						Session:     s,
						Interaction: i,
						RaceKey:     char.Race.Key,
					}
					if err := h.characterShowClassesHandler.Handle(req); err != nil {
						log.Printf("Error resuming character creation at class selection: %v", err)
					}
				} else if len(char.Attributes) == 0 {
					// Continue from ability scores
					req := &character.AbilityScoresRequest{
						Session:     s,
						Interaction: i,
					}
					if err := h.characterAbilityScoresHandler.Handle(req); err != nil {
						log.Printf("Error resuming character creation at ability scores: %v", err)
					}
				} else if len(char.Proficiencies) == 0 {
					// Continue from proficiencies
					req := &character.ProficiencyChoicesRequest{
						Session:     s,
						Interaction: i,
						RaceKey:     char.Race.Key,
						ClassKey:    char.Class.Key,
					}
					if err := h.characterProficiencyChoicesHandler.Handle(req); err != nil {
						log.Printf("Error resuming character creation at proficiencies: %v", err)
					}
				} else if len(char.Inventory) == 0 && len(char.EquippedSlots) == 0 {
					// Continue from equipment
					req := &character.EquipmentChoicesRequest{
						Session:     s,
						Interaction: i,
						RaceKey:     char.Race.Key,
						ClassKey:    char.Class.Key,
					}
					if err := h.characterEquipmentChoicesHandler.Handle(req); err != nil {
						log.Printf("Error resuming character creation at equipment: %v", err)
					}
				} else if char.Name == "" {
					// All gameplay elements done, just need name
					// For now, skip features step and go to name modal
					modal := discordgo.InteractionResponseData{
						CustomID: fmt.Sprintf("character_create:submit_name:%s:%s", char.Race.Key, char.Class.Key),
						Title:    "Name Your Character",
						Components: []discordgo.MessageComponent{
							discordgo.ActionsRow{
								Components: []discordgo.MessageComponent{
									discordgo.TextInput{
										CustomID:    "character_name",
										Label:       "Character Name",
										Style:       discordgo.TextInputShort,
										Placeholder: "Enter your character's name",
										Required:    true,
										MinLength:   1,
										MaxLength:   50,
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
				} else {
					// Character seems complete, show it
					req := &character.ShowRequest{
						Session:     s,
						Interaction: i,
						CharacterID: characterID,
					}
					if err := h.characterShowHandler.Handle(req); err != nil {
						log.Printf("Error showing draft character: %v", err)
					}
				}
			}
		}
	} else if ctx == "character" && action == "delete_select" {
		// Handle character selection for deletion
		if len(parts) >= 3 {
			characterID := parts[2]
			req := &character.DeleteRequest{
				Session:     s,
				Interaction: i,
				CharacterID: characterID,
			}
			if err := h.characterDeleteHandler.Handle(req); err != nil {
				log.Printf("Error handling character delete select: %v", err)
			}
		}
	} else if ctx == "character" && action == "delete_select_menu" {
		// Handle character selection from dropdown menu
		req := &character.DeleteRequest{
			Session:     s,
			Interaction: i,
		}
		if err := h.characterDeleteHandler.HandleSelectMenu(req); err != nil {
			log.Printf("Error handling character delete select menu: %v", err)
		}
	} else if ctx == "character" && action == "delete_confirm" {
		// Handle deletion confirmation
		req := &character.DeleteRequest{
			Session:     s,
			Interaction: i,
		}
		if err := h.characterDeleteHandler.HandleDeleteConfirm(req); err != nil {
			log.Printf("Error handling character delete confirm: %v", err)
		}
	} else if ctx == "character" && action == "delete_cancel" {
		// Handle deletion cancellation
		req := &character.DeleteRequest{
			Session:     s,
			Interaction: i,
		}
		if err := h.characterDeleteHandler.HandleDeleteCancel(req); err != nil {
			log.Printf("Error handling character delete cancel: %v", err)
		}
	} else if ctx == "character" && action == "sheet_refresh" {
		// Refresh character sheet
		if len(parts) >= 3 {
			characterID := parts[2]
			// Use helper function to reduce duplication
			err := helpers.ShowCharacterSheet(s, i, characterID, i.Member.User.ID, h.ServiceProvider, true)
			if err != nil {
				log.Printf("Error refreshing character sheet: %v", err)
				// Provide user feedback on error
				if respondErr := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseUpdateMessage,
					Data: &discordgo.InteractionResponseData{
						Content:    "❌ Failed to refresh character sheet. Please try again.",
						Components: []discordgo.MessageComponent{},
					},
				}); respondErr != nil {
					log.Printf("Error sending error response: %v", respondErr)
				}
			}
		}
	} else if ctx == "character" && action == "sheet_show" {
		// Show character sheet from list
		if len(parts) >= 3 {
			characterID := parts[2]
			// Use helper function to reduce duplication
			err := helpers.ShowCharacterSheet(s, i, characterID, i.Member.User.ID, h.ServiceProvider, false)
			if err != nil {
				log.Printf("Error showing character sheet: %v", err)
				// Provide user feedback on error
				respondErr := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "❌ Failed to display character sheet. Please try again.",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				if respondErr != nil {
					log.Printf("Error sending error response: %v", respondErr)
				}
			}
		}
	} else if ctx == "character" && action == "inventory" {
		// Show inventory management interface
		if len(parts) >= 3 {
			characterID := parts[2]

			// Get the character
			char, err := h.ServiceProvider.CharacterService.GetByID(characterID)
			if err != nil {
				respondWithUpdateError(s, i, fmt.Sprintf("Failed to get character: %v", err))
				return
			}

			// Verify ownership
			if char.OwnerID != i.Member.User.ID {
				respondWithUpdateError(s, i, "You can only manage your own character's inventory!")
				return
			}

			// Build equipment category select menu
			embed := &discordgo.MessageEmbed{
				Title:       fmt.Sprintf("🎒 %s's Equipment", char.Name),
				Description: "Select a category to view and manage your equipment:",
				Color:       0x3498db,
			}

			components := []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.SelectMenu{
							CustomID:    fmt.Sprintf("character:equipment_category:%s", characterID),
							Placeholder: "Choose equipment category...",
							Options: []discordgo.SelectMenuOption{
								{
									Label:       "Weapons",
									Description: "View and equip weapons",
									Value:       "weapons",
									Emoji:       &discordgo.ComponentEmoji{Name: "⚔️"},
								},
								{
									Label:       "Armor",
									Description: "View and equip armor",
									Value:       "armor",
									Emoji:       &discordgo.ComponentEmoji{Name: "🛡️"},
								},
								{
									Label:       "All Items",
									Description: "View all inventory items",
									Value:       "all",
									Emoji:       &discordgo.ComponentEmoji{Name: "📦"},
								},
							},
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label:    "Back to Sheet",
							Style:    discordgo.SecondaryButton,
							CustomID: fmt.Sprintf("character:sheet_refresh:%s", characterID),
							Emoji:    &discordgo.ComponentEmoji{Name: "⬅️"},
						},
					},
				},
			}

			// Update the message
			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseUpdateMessage,
				Data: &discordgo.InteractionResponseData{
					Embeds:     []*discordgo.MessageEmbed{embed},
					Components: components,
				},
			})
			if err != nil {
				log.Printf("Error showing inventory menu: %v", err)
			}
		}
	} else if ctx == "character" && action == "equipment_category" {
		// Handle equipment category selection
		if len(parts) >= 3 && len(i.MessageComponentData().Values) > 0 {
			characterID := parts[2]
			category := i.MessageComponentData().Values[0]

			// Get the character
			char, err := h.ServiceProvider.CharacterService.GetByID(characterID)
			if err != nil {
				respondWithUpdateError(s, i, fmt.Sprintf("Failed to get character: %v", err))
				return
			}

			// Verify ownership
			if char.OwnerID != i.Member.User.ID {
				respondWithUpdateError(s, i, "You can only manage your own character's inventory!")
				return
			}

			// Build equipment list based on category
			var items []entities.Equipment
			var categoryName string

			switch category {
			case "weapons":
				categoryName = "Weapons"
				for _, equipList := range char.Inventory {
					for _, equip := range equipList {
						if weapon, ok := equip.(*entities.Weapon); ok {
							items = append(items, weapon)
						}
					}
				}
			case "armor":
				categoryName = "Armor"
				for _, equipList := range char.Inventory {
					for _, equip := range equipList {
						if armor, ok := equip.(*entities.Armor); ok {
							items = append(items, armor)
						}
					}
				}
			case "all":
				categoryName = "All Equipment"
				for _, equipList := range char.Inventory {
					items = append(items, equipList...)
				}
			}

			// Build the embed
			embed := &discordgo.MessageEmbed{
				Title:       fmt.Sprintf("⚔️ %s - %s", char.Name, categoryName),
				Description: fmt.Sprintf("You have %d item(s) in this category.", len(items)),
				Color:       0x3498db,
				Fields:      []*discordgo.MessageEmbedField{},
			}

			// Add currently equipped items info
			equippedInfo := "**Currently Equipped:**\n"
			hasEquipped := false

			if weapon := char.EquippedSlots[entities.SlotMainHand]; weapon != nil {
				equippedInfo += fmt.Sprintf("Main Hand: %s\n", weapon.GetName())
				hasEquipped = true
			}
			if item := char.EquippedSlots[entities.SlotOffHand]; item != nil {
				equippedInfo += fmt.Sprintf("Off Hand: %s\n", item.GetName())
				hasEquipped = true
			}
			if armor := char.EquippedSlots[entities.SlotBody]; armor != nil {
				equippedInfo += fmt.Sprintf("Armor: %s\n", armor.GetName())
				hasEquipped = true
			}

			if !hasEquipped {
				equippedInfo += "*Nothing equipped*"
			}

			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "🛡️ Current Equipment",
				Value:  equippedInfo,
				Inline: false,
			})

			// Build components - start with category select and back button
			components := []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.SelectMenu{
							CustomID:    fmt.Sprintf("character:equipment_category:%s", characterID),
							Placeholder: "Choose equipment category...",
							Options: []discordgo.SelectMenuOption{
								{
									Label:       "Weapons",
									Description: "View and equip weapons",
									Value:       "weapons",
									Emoji:       &discordgo.ComponentEmoji{Name: "⚔️"},
									Default:     category == "weapons",
								},
								{
									Label:       "Armor",
									Description: "View and equip armor",
									Value:       "armor",
									Emoji:       &discordgo.ComponentEmoji{Name: "🛡️"},
									Default:     category == "armor",
								},
								{
									Label:       "All Items",
									Description: "View all inventory items",
									Value:       "all",
									Emoji:       &discordgo.ComponentEmoji{Name: "📦"},
									Default:     category == "all",
								},
							},
						},
					},
				},
			}

			// Add item selection if there are items
			if len(items) > 0 {
				options := []discordgo.SelectMenuOption{}

				// Track item counts to handle duplicates
				itemCounts := make(map[string]int)
				for _, invItem := range items {
					itemCounts[invItem.GetKey()]++
				}

				// Track current index for each item type
				itemIndices := make(map[string]int)

				for i, invItem := range items {
					if i >= 25 { // Discord limit
						break
					}

					// Build item description
					desc := ""
					switch item := invItem.(type) {
					case *entities.Weapon:
						desc = fmt.Sprintf("Damage: %dd%d", item.Damage.DiceCount, item.Damage.DiceSize)
						if item.Damage.Bonus > 0 {
							desc += fmt.Sprintf("+%d", item.Damage.Bonus)
						}
					case *entities.Armor:
						if item.ArmorClass != nil {
							desc = fmt.Sprintf("AC: %d", item.ArmorClass.Base)
							if item.ArmorClass.MaxBonus > 0 {
								desc += fmt.Sprintf(" (max Dex: %d)", item.ArmorClass.MaxBonus)
							}
						} else {
							desc = fmt.Sprintf("Type: %s", item.ArmorCategory)
						}
					}

					// Check if equipped
					isEquipped := false
					for _, equipped := range char.EquippedSlots {
						if equipped != nil && equipped.GetKey() == invItem.GetKey() {
							isEquipped = true
							break
						}
					}

					label := invItem.GetName()
					if isEquipped {
						label += " (Equipped)"
					}

					// Make value unique for duplicate items
					value := invItem.GetKey()
					if itemCounts[invItem.GetKey()] > 1 {
						itemIndices[invItem.GetKey()]++
						// Use a special delimiter "|||" to avoid conflicts with item keys that contain underscores
						value = fmt.Sprintf("%s|||%d", invItem.GetKey(), itemIndices[invItem.GetKey()])
						// Add index to label if there are duplicates
						if !isEquipped {
							label = fmt.Sprintf("%s #%d", label, itemIndices[invItem.GetKey()])
						}
					}

					options = append(options, discordgo.SelectMenuOption{
						Label:       label,
						Description: desc,
						Value:       value,
					})
				}

				if len(options) > 0 {
					components = append(components, discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.SelectMenu{
								CustomID:    fmt.Sprintf("character:select_item:%s:%s", characterID, category),
								Placeholder: "Select an item to view details...",
								Options:     options,
							},
						},
					})
				}
			} else {
				embed.Description = fmt.Sprintf("You don't have any %s in your inventory.", strings.ToLower(categoryName))
			}

			// Add back button
			components = append(components, discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "Back to Sheet",
						Style:    discordgo.SecondaryButton,
						CustomID: fmt.Sprintf("character:sheet_refresh:%s", characterID),
						Emoji:    &discordgo.ComponentEmoji{Name: "⬅️"},
					},
				},
			})

			// Update the message
			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseUpdateMessage,
				Data: &discordgo.InteractionResponseData{
					Embeds:     []*discordgo.MessageEmbed{embed},
					Components: components,
				},
			})
			if err != nil {
				log.Printf("Error showing equipment category: %v", err)
			}
		}
	} else if ctx == "character" && action == "select_item" {
		// Handle item selection for details and equip/unequip
		if len(parts) >= 4 && len(i.MessageComponentData().Values) > 0 {
			characterID := parts[2]
			category := parts[3]
			itemValue := i.MessageComponentData().Values[0]

			// Extract the base item key (remove index suffix if present)
			// We use a special delimiter pattern "|||" to separate the key from index
			// to avoid conflicts with items that have underscores in their keys
			itemKey := itemValue
			delimiterPattern := "|||"
			if idx := strings.LastIndex(itemValue, delimiterPattern); idx != -1 {
				// Check if the part after delimiter is a number
				if _, err := strconv.Atoi(itemValue[idx+len(delimiterPattern):]); err == nil {
					itemKey = itemValue[:idx]
				}
			}

			// Get the character
			char, err := h.ServiceProvider.CharacterService.GetByID(characterID)
			if err != nil {
				respondWithUpdateError(s, i, fmt.Sprintf("Failed to get character: %v", err))
				return
			}

			// Verify ownership
			if char.OwnerID != i.Member.User.ID {
				respondWithUpdateError(s, i, "You can only manage your own character's inventory!")
				return
			}

			// Find the item in inventory
			var selectedItem entities.Equipment
			currentIndex := make(map[string]int)

			for _, equipList := range char.Inventory {
				for _, equip := range equipList {
					if equip.GetKey() == itemKey {
						currentIndex[itemKey]++
						// If the value has an index, match it
						if strings.Contains(itemValue, delimiterPattern) {
							if itemValue == fmt.Sprintf("%s|||%d", itemKey, currentIndex[itemKey]) {
								selectedItem = equip
								break
							}
						} else {
							// No index, just take the first match
							selectedItem = equip
							break
						}
					}
				}
				if selectedItem != nil {
					break
				}
			}

			if selectedItem == nil {
				respondWithUpdateError(s, i, "Item not found in inventory!")
				return
			}

			// Check if item is equipped
			isEquipped := false
			var equippedSlot entities.Slot
			for slot, equipped := range char.EquippedSlots {
				if equipped != nil && equipped == selectedItem {
					isEquipped = true
					equippedSlot = slot
					break
				}
			}

			// Build item details embed
			embed := &discordgo.MessageEmbed{
				Title:  fmt.Sprintf("📋 %s", selectedItem.GetName()),
				Color:  0x3498db,
				Fields: []*discordgo.MessageEmbedField{},
			}

			// Add item-specific details
			switch item := selectedItem.(type) {
			case *entities.Weapon:
				embed.Fields = append(embed.Fields,
					&discordgo.MessageEmbedField{
						Name: "⚔️ Weapon Details",
						Value: fmt.Sprintf("**Damage:** %dd%d+%d %s\n**Properties:** %s",
							item.Damage.DiceCount,
							item.Damage.DiceSize,
							item.Damage.Bonus,
							item.Damage.DamageType,
							getWeaponPropertiesString(item)),
						Inline: false,
					},
				)

				// Add two-handed damage if applicable
				if item.TwoHandedDamage != nil {
					embed.Fields = append(embed.Fields,
						&discordgo.MessageEmbedField{
							Name: "💪 Two-Handed",
							Value: fmt.Sprintf("**Damage:** %dd%d+%d",
								item.TwoHandedDamage.DiceCount,
								item.TwoHandedDamage.DiceSize,
								item.TwoHandedDamage.Bonus),
							Inline: false,
						},
					)
				}
			case *entities.Armor:
				armorInfo := fmt.Sprintf("**Type:** %s", item.ArmorCategory)
				if item.ArmorClass != nil {
					armorInfo = fmt.Sprintf("**Base AC:** %d\n%s", item.ArmorClass.Base, armorInfo)
					if item.ArmorClass.MaxBonus > 0 {
						armorInfo += fmt.Sprintf("\n**Max Dex Bonus:** %d", item.ArmorClass.MaxBonus)
					}
				}
				if item.StrMin > 0 {
					armorInfo += fmt.Sprintf("\n**Min Strength:** %d", item.StrMin)
				}
				if item.StealthDisadvantage {
					armorInfo += "\n**Stealth:** Disadvantage"
				}

				embed.Fields = append(embed.Fields,
					&discordgo.MessageEmbedField{
						Name:   "🛡️ Armor Details",
						Value:  armorInfo,
						Inline: false,
					},
				)
			}

			// Add equipment status
			statusValue := "Not equipped"
			if isEquipped {
				statusValue = fmt.Sprintf("Equipped in: **%s**", equippedSlot)
			}
			embed.Fields = append(embed.Fields,
				&discordgo.MessageEmbedField{
					Name:   "📍 Status",
					Value:  statusValue,
					Inline: false,
				},
			)

			// Build action buttons
			components := []discordgo.MessageComponent{}

			if isEquipped {
				// Show unequip button
				components = append(components, discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label:    "Unequip",
							Style:    discordgo.DangerButton,
							CustomID: fmt.Sprintf("character:unequip:%s:%s", characterID, itemKey),
							Emoji:    &discordgo.ComponentEmoji{Name: "❌"},
						},
					},
				})
			} else {
				// Show equip buttons based on item type
				buttons := []discordgo.MessageComponent{}

				switch item := selectedItem.(type) {
				case *entities.Weapon:
					// Check if it's two-handed
					isTwoHanded := false
					for _, prop := range item.Properties {
						if prop != nil && strings.EqualFold(prop.Key, "two-handed") {
							isTwoHanded = true
							break
						}
					}

					if isTwoHanded {
						buttons = append(buttons, discordgo.Button{
							Label:    "Equip (Two-Handed)",
							Style:    discordgo.SuccessButton,
							CustomID: fmt.Sprintf("character:equip:%s:%s:two-handed", characterID, itemKey),
							Emoji:    &discordgo.ComponentEmoji{Name: "🗡️"},
						})
					} else {
						buttons = append(buttons,
							discordgo.Button{
								Label:    "Equip Main Hand",
								Style:    discordgo.SuccessButton,
								CustomID: fmt.Sprintf("character:equip:%s:%s:main-hand", characterID, itemKey),
								Emoji:    &discordgo.ComponentEmoji{Name: "✋"},
							},
							discordgo.Button{
								Label:    "Equip Off Hand",
								Style:    discordgo.SuccessButton,
								CustomID: fmt.Sprintf("character:equip:%s:%s:off-hand", characterID, itemKey),
								Emoji:    &discordgo.ComponentEmoji{Name: "🤚"},
							},
						)
					}
				case *entities.Armor:
					buttons = append(buttons, discordgo.Button{
						Label:    "Equip Armor",
						Style:    discordgo.SuccessButton,
						CustomID: fmt.Sprintf("character:equip:%s:%s:body", characterID, itemKey),
						Emoji:    &discordgo.ComponentEmoji{Name: "🛡️"},
					})
				}

				if len(buttons) > 0 {
					components = append(components, discordgo.ActionsRow{
						Components: buttons,
					})
				}
			}

			// Add back button
			components = append(components, discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    fmt.Sprintf("Back to %s", cases.Title(language.English).String(category)),
						Style:    discordgo.SecondaryButton,
						CustomID: fmt.Sprintf("character:equipment_category:%s", characterID),
						Emoji:    &discordgo.ComponentEmoji{Name: "⬅️"},
					},
					discordgo.Button{
						Label:    "Back to Sheet",
						Style:    discordgo.SecondaryButton,
						CustomID: fmt.Sprintf("character:sheet_refresh:%s", characterID),
						Emoji:    &discordgo.ComponentEmoji{Name: "📋"},
					},
				},
			})

			// Update the message
			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseUpdateMessage,
				Data: &discordgo.InteractionResponseData{
					Embeds:     []*discordgo.MessageEmbed{embed},
					Components: components,
				},
			})
			if err != nil {
				log.Printf("Error showing item details: %v", err)
			}
		}
	} else if ctx == "character" && action == "equip" {
		// Handle equipping an item
		if len(parts) >= 5 {
			characterID := parts[2]
			itemKey := parts[3]
			slotType := parts[4]

			// Get the character
			char, err := h.ServiceProvider.CharacterService.GetByID(characterID)
			if err != nil {
				respondWithUpdateError(s, i, fmt.Sprintf("Failed to get character: %v", err))
				return
			}

			// Verify ownership
			if char.OwnerID != i.Member.User.ID {
				respondWithUpdateError(s, i, "You can only manage your own character's inventory!")
				return
			}

			// Find the item in inventory
			var selectedItem entities.Equipment
			for _, equipList := range char.Inventory {
				for _, equip := range equipList {
					if equip.GetKey() == itemKey {
						selectedItem = equip
						break
					}
				}
				if selectedItem != nil {
					break
				}
			}

			if selectedItem == nil {
				respondWithUpdateError(s, i, "Item not found in inventory!")
				return
			}

			// The Character.Equip method handles slot assignment internally based on the item
			// We just need to verify the slot type matches the item type
			switch slotType {
			case "main-hand", "off-hand", "two-handed":
				// Verify it's a weapon
				if _, ok := selectedItem.(*entities.Weapon); !ok {
					respondWithUpdateError(s, i, "This item cannot be equipped as a weapon!")
					return
				}
			case "body":
				// Verify it's armor
				if _, ok := selectedItem.(*entities.Armor); !ok {
					respondWithUpdateError(s, i, "This item cannot be equipped as armor!")
					return
				}
			default:
				respondWithUpdateError(s, i, "Invalid equipment slot!")
				return
			}

			// Equip the item using its key
			success := char.Equip(itemKey)
			if !success {
				respondWithUpdateError(s, i, "Failed to equip item!")
				return
			}

			// Save the character equipment changes
			err = h.ServiceProvider.CharacterService.UpdateEquipment(char)
			if err != nil {
				respondWithUpdateError(s, i, fmt.Sprintf("Failed to save equipment: %v", err))
				return
			}

			// Show success message and refresh inventory
			embed := &discordgo.MessageEmbed{
				Title:       "✅ Item Equipped!",
				Description: fmt.Sprintf("**%s** has been equipped!", selectedItem.GetName()),
				Color:       0x2ecc71,
			}

			// Add updated equipment info
			equippedInfo := "**Currently Equipped:**\n"
			if weapon := char.EquippedSlots[entities.SlotMainHand]; weapon != nil {
				equippedInfo += fmt.Sprintf("Main Hand: %s\n", weapon.GetName())
			}
			if item := char.EquippedSlots[entities.SlotOffHand]; item != nil {
				equippedInfo += fmt.Sprintf("Off Hand: %s\n", item.GetName())
			}
			if weapon := char.EquippedSlots[entities.SlotTwoHanded]; weapon != nil {
				equippedInfo += fmt.Sprintf("Two-Handed: %s\n", weapon.GetName())
			}
			if armor := char.EquippedSlots[entities.SlotBody]; armor != nil {
				equippedInfo += fmt.Sprintf("Armor: %s\n", armor.GetName())
			}

			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "🛡️ Current Equipment",
				Value:  equippedInfo,
				Inline: false,
			})

			// Show buttons to continue
			components := []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label:    "View Inventory",
							Style:    discordgo.PrimaryButton,
							CustomID: fmt.Sprintf("character:inventory:%s", characterID),
							Emoji:    &discordgo.ComponentEmoji{Name: "🎒"},
						},
						discordgo.Button{
							Label:    "Back to Sheet",
							Style:    discordgo.SecondaryButton,
							CustomID: fmt.Sprintf("character:sheet_refresh:%s", characterID),
							Emoji:    &discordgo.ComponentEmoji{Name: "📋"},
						},
					},
				},
			}

			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseUpdateMessage,
				Data: &discordgo.InteractionResponseData{
					Embeds:     []*discordgo.MessageEmbed{embed},
					Components: components,
				},
			})
			if err != nil {
				log.Printf("Error showing equip success: %v", err)
			}
		}
	} else if ctx == "character" && action == "unequip" {
		// Handle unequipping an item
		if len(parts) >= 4 {
			characterID := parts[2]
			itemKey := parts[3]

			// Get the character
			char, err := h.ServiceProvider.CharacterService.GetByID(characterID)
			if err != nil {
				respondWithUpdateError(s, i, fmt.Sprintf("Failed to get character: %v", err))
				return
			}

			// Verify ownership
			if char.OwnerID != i.Member.User.ID {
				respondWithUpdateError(s, i, "You can only manage your own character's inventory!")
				return
			}

			// Find which slot has the item and unequip it
			var foundSlot entities.Slot
			var foundItem entities.Equipment
			for slot, equipped := range char.EquippedSlots {
				if equipped != nil && equipped.GetKey() == itemKey {
					foundSlot = slot
					foundItem = equipped
					break
				}
			}

			if foundItem == nil {
				respondWithUpdateError(s, i, "Item is not equipped!")
				return
			}

			// Unequip the item by setting the slot to nil
			char.EquippedSlots[foundSlot] = nil

			// Save the character equipment changes
			err = h.ServiceProvider.CharacterService.UpdateEquipment(char)
			if err != nil {
				respondWithUpdateError(s, i, fmt.Sprintf("Failed to save equipment: %v", err))
				return
			}

			// Show success message
			embed := &discordgo.MessageEmbed{
				Title:       "✅ Item Unequipped!",
				Description: fmt.Sprintf("**%s** has been unequipped from **%s**", foundItem.GetName(), foundSlot),
				Color:       0x2ecc71,
			}

			// Show buttons to continue
			components := []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label:    "View Inventory",
							Style:    discordgo.PrimaryButton,
							CustomID: fmt.Sprintf("character:inventory:%s", characterID),
							Emoji:    &discordgo.ComponentEmoji{Name: "🎒"},
						},
						discordgo.Button{
							Label:    "Back to Sheet",
							Style:    discordgo.SecondaryButton,
							CustomID: fmt.Sprintf("character:sheet_refresh:%s", characterID),
							Emoji:    &discordgo.ComponentEmoji{Name: "📋"},
						},
					},
				},
			}

			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseUpdateMessage,
				Data: &discordgo.InteractionResponseData{
					Embeds:     []*discordgo.MessageEmbed{embed},
					Components: components,
				},
			})
			if err != nil {
				log.Printf("Error showing unequip success: %v", err)
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
					err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: content,
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					if err != nil {
						log.Printf("Error responding to start session: %v", err)
					}
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
					err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: content,
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					if err != nil {
						log.Printf("Error responding to leave session: %v", err)
					}
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
					err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: content,
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					if err != nil {
						log.Printf("Error responding with error: %v", err)
					}
					return
				}

				// Build character options
				options := make([]discordgo.SelectMenuOption, 0, len(chars))
				for _, char := range chars {
					if char.Status == entities.CharacterStatusActive {
						options = append(options, discordgo.SelectMenuOption{
							Label:       fmt.Sprintf("%s - %s", char.Name, char.GetDisplayInfo()),
							Description: fmt.Sprintf("Level %d | HP: %d/%d | AC: %d", char.Level, char.CurrentHitPoints, char.MaxHitPoints, char.AC),
							Value:       char.ID,
						})
					}
				}

				if len(options) == 0 {
					content := "❌ You don't have any active characters! Create or activate a character first."
					err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: content,
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					if err != nil {
						log.Printf("Error responding with error: %v", err)
					}
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
						err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
							Type: discordgo.InteractionResponseUpdateMessage,
							Data: &discordgo.InteractionResponseData{
								Content:    content,
								Components: []discordgo.MessageComponent{},
							},
						})
						if err != nil {
							log.Printf("Error responding with error: %v", err)
						}
						return
					}

					// Get character details for confirmation
					char, charErr := h.ServiceProvider.CharacterService.GetByID(characterID)
					if charErr != nil {
						content := fmt.Sprintf("❌ Failed to get character: %v", charErr)
						err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
							Type: discordgo.InteractionResponseUpdateMessage,
							Data: &discordgo.InteractionResponseData{
								Content:    content,
								Components: []discordgo.MessageComponent{},
							},
						})
						if err != nil {
							log.Printf("Error responding with error: %v", err)
						}
						return
					}
					content := fmt.Sprintf("✅ Character selected: **%s** the %s", char.Name, char.GetDisplayInfo())
					err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseUpdateMessage,
						Data: &discordgo.InteractionResponseData{
							Content:    content,
							Components: []discordgo.MessageComponent{},
						},
					})
					if err != nil {
						log.Printf("Error confirming character selection: %v", err)
					}

					// Check if there's an active encounter in this session
					// If so, we need to add the player to it and update the shared message
					enc, encErr := h.ServiceProvider.EncounterService.GetActiveEncounter(context.Background(), sessionID)
					if encErr != nil {
						// Log the error but don't fail - there might not be an active encounter
						log.Printf("No active encounter found for session %s: %v", sessionID, encErr)
					} else if enc != nil {
						log.Printf("Found active encounter %s, adding player %s with character %s", enc.ID, i.Member.User.ID, characterID)

						// Check if player is already in the encounter with a different character
						var existingCombatantID string
						for id, combatant := range enc.Combatants {
							if combatant.PlayerID == i.Member.User.ID {
								log.Printf("Player %s is already in encounter as %s (combatant %s)", i.Member.User.ID, combatant.Name, id)
								existingCombatantID = id
								break
							}
						}

						if existingCombatantID != "" {
							// Remove the old combatant first
							log.Printf("Removing old combatant %s to replace with new character", existingCombatantID)
							removeErr := h.ServiceProvider.EncounterService.RemoveCombatant(context.Background(), enc.ID, existingCombatantID, i.Member.User.ID)
							if removeErr != nil {
								log.Printf("Failed to remove old combatant: %v", removeErr)
							}
						}

						// Add the player with the new character
						_, addErr := h.ServiceProvider.EncounterService.AddPlayer(context.Background(), enc.ID, i.Member.User.ID, characterID)
						if addErr != nil {
							log.Printf("Failed to add player to active encounter: %v", addErr)
						} else {
							log.Printf("Successfully added player %s to encounter %s with new character", i.Member.User.ID, enc.ID)
						}

						// Always try to update the shared combat message after character selection
						// This ensures the display is current even if the player was already in the encounter
						log.Printf("Updating shared combat message for encounter %s", enc.ID)

						// Get fresh encounter data
						updatedEnc, getErr := h.ServiceProvider.EncounterService.GetEncounter(context.Background(), enc.ID)
						if getErr != nil {
							log.Printf("Failed to get updated encounter: %v", getErr)
						} else if updatedEnc != nil {
							log.Printf("Got updated encounter. MessageID: %s, ChannelID: %s", updatedEnc.MessageID, updatedEnc.ChannelID)

							if updatedEnc.MessageID != "" && updatedEnc.ChannelID != "" {
								// Build the combat status embed
								embed := combat.BuildCombatStatusEmbed(updatedEnc, nil)
								components := combat.BuildCombatComponents(updatedEnc.ID, &encounter.ExecuteAttackResult{})

								// Update the shared message
								log.Printf("Attempting to update message %s in channel %s", updatedEnc.MessageID, updatedEnc.ChannelID)
								_, err := s.ChannelMessageEditComplex(&discordgo.MessageEdit{
									ID:         updatedEnc.MessageID,
									Channel:    updatedEnc.ChannelID,
									Embeds:     &[]*discordgo.MessageEmbed{embed},
									Components: &components,
								})
								if err != nil {
									log.Printf("Failed to update shared combat message: %v", err)
								} else {
									log.Printf("Successfully updated shared combat message")
								}
							} else {
								log.Printf("Encounter %s has no message ID stored (MessageID: %s, ChannelID: %s)", updatedEnc.ID, updatedEnc.MessageID, updatedEnc.ChannelID)
							}
						}
					}
				}
			case "pause":
				// Pause the session
				err := h.ServiceProvider.SessionService.PauseSession(context.Background(), sessionID, i.Member.User.ID)
				if err != nil {
					content := fmt.Sprintf("❌ Failed to pause session: %v", err)
					err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: content,
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					if err != nil {
						log.Printf("Error responding with error: %v", err)
					}
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
				req := &sessionHandler.EndRequest{
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
					err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: content,
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					if err != nil {
						log.Printf("Error responding with error: %v", err)
					}
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
					err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: content,
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					if err != nil {
						log.Printf("Error responding with error: %v", err)
					}
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
	} else if ctx == "combat" {
		// Use new clean combat handler
		if len(parts) >= 3 {
			encounterID := parts[2]
			if err := h.combatHandler.HandleButton(s, i, action, encounterID); err != nil {
				log.Printf("Error handling combat button: %v", err)
			}
		}
	} else if ctx == "encounter" {
		// Handle encounter management actions
		if len(parts) >= 3 {
			encounterID := parts[2]

			switch action {
			case "roll_initiative":
				// Roll initiative for all combatants
				err := h.ServiceProvider.EncounterService.RollInitiative(context.Background(), encounterID, i.Member.User.ID)
				if err != nil {
					content := fmt.Sprintf("❌ Failed to roll initiative: %v", err)
					err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: content,
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					if err != nil {
						log.Printf("Error responding with error: %v", err)
					}
					return
				}

				// Get encounterResult to show results
				encounterResult, err := h.ServiceProvider.EncounterService.GetEncounter(context.Background(), encounterID)
				if err != nil {
					content := "✅ Initiative rolled! Use View Encounter to see the order."
					err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: content,
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					if err != nil {
						log.Printf("Error responding with error: %v", err)
					}
					return
				}

				// Build initiative order display
				var initiativeList strings.Builder
				for i, combatantID := range encounterResult.TurnOrder {
					if combatant, exists := encounterResult.Combatants[combatantID]; exists {
						initiativeList.WriteString(fmt.Sprintf("%d. **%s** - Initiative: %d\n", i+1, combatant.Name, combatant.Initiative))
					}
				}

				// Get initiative roll details from combat log
				var rollDetails strings.Builder
				if len(encounterResult.CombatLog) > 0 {
					// Skip the first entry which is "Rolling Initiative" header
					for i := 1; i < len(encounterResult.CombatLog) && i <= len(encounterResult.Combatants)+1; i++ {
						rollDetails.WriteString(encounterResult.CombatLog[i] + "\n")
					}
				}

				embed := &discordgo.MessageEmbed{
					Title:       "🎲 Initiative Rolled!",
					Description: "Combat order has been determined:",
					Color:       0x2ecc71, // Green
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:   "🎯 Initiative Rolls",
							Value:  rollDetails.String(),
							Inline: false,
						},
						{
							Name:   "⚔️ Turn Order",
							Value:  initiativeList.String(),
							Inline: false,
						},
					},
				}

				components := []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.Button{
								Label:    "Start Combat",
								Style:    discordgo.SuccessButton,
								CustomID: fmt.Sprintf("encounter:start:%s", encounterID),
								Emoji:    &discordgo.ComponentEmoji{Name: "⚔️"},
							},
							discordgo.Button{
								Label:    "View Encounter",
								Style:    discordgo.SecondaryButton,
								CustomID: fmt.Sprintf("encounter:view:%s", encounterID),
								Emoji:    &discordgo.ComponentEmoji{Name: "👁️"},
							},
						},
					},
				}

				err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Embeds:     []*discordgo.MessageEmbed{embed},
						Components: components,
					},
				})
				if err != nil {
					log.Printf("Error showing initiative results: %v", err)
				}

			case "view":
				// View encounter details
				encounterResult, err := h.ServiceProvider.EncounterService.GetEncounter(context.Background(), encounterID)
				if err != nil {
					content := fmt.Sprintf("❌ Failed to get encounter: %v", err)
					err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: content,
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					if err != nil {
						log.Printf("Error responding with error: %v", err)
					}
					return
				}

				// Build encounter status
				embed := &discordgo.MessageEmbed{
					Title:       fmt.Sprintf("⚔️ %s", encounterResult.Name),
					Description: encounterResult.Description,
					Color:       0x3498db, // Blue
					Fields:      []*discordgo.MessageEmbedField{},
				}

				// Add status field
				statusStr := string(encounterResult.Status)
				if encounterResult.Status == entities.EncounterStatusActive {
					statusStr = fmt.Sprintf("Active - Round %d", encounterResult.Round)
				}
				embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
					Name:   "📊 Status",
					Value:  statusStr,
					Inline: true,
				})

				// Add combatant count
				activeCombatants := 0
				for _, c := range encounterResult.Combatants {
					if c.IsActive {
						activeCombatants++
					}
				}
				embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
					Name:   "👥 Combatants",
					Value:  fmt.Sprintf("%d active / %d total", activeCombatants, len(encounterResult.Combatants)),
					Inline: true,
				})

				// List combatants with HP
				var combatantList strings.Builder
				for _, combatant := range encounterResult.Combatants {
					hpBar := ""
					if combatant.MaxHP > 0 {
						hpPercent := float64(combatant.CurrentHP) / float64(combatant.MaxHP)
						if hpPercent > 0.5 {
							hpBar = "🟢"
						} else if hpPercent > 0.25 {
							hpBar = "🟡"
						} else if combatant.CurrentHP > 0 {
							hpBar = "🔴"
						} else {
							hpBar = "💀"
						}
					}

					combatantList.WriteString(fmt.Sprintf("%s **%s** - HP: %d/%d | AC: %d\n",
						hpBar, combatant.Name, combatant.CurrentHP, combatant.MaxHP, combatant.AC))
				}

				if combatantList.Len() > 0 {
					embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
						Name:   "🗡️ Combatants",
						Value:  combatantList.String(),
						Inline: false,
					})
				}

				// Add appropriate buttons based on status
				var components []discordgo.MessageComponent
				switch encounterResult.Status {
				case entities.EncounterStatusSetup:
					components = []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.Button{
									Label:    "Add Monster",
									Style:    discordgo.PrimaryButton,
									CustomID: fmt.Sprintf("encounter:add_monster:%s", encounterID),
									Emoji:    &discordgo.ComponentEmoji{Name: "➕"},
								},
								discordgo.Button{
									Label:    "Roll Initiative",
									Style:    discordgo.SuccessButton,
									CustomID: fmt.Sprintf("encounter:roll_initiative:%s", encounterID),
									Emoji:    &discordgo.ComponentEmoji{Name: "🎲"},
								},
							},
						},
					}
				case entities.EncounterStatusActive:
					// Show combat controls for active encounters
					// Check if it's the viewing player's turn
					isPlayerTurn := false
					if current := encounterResult.GetCurrentCombatant(); current != nil {
						isPlayerTurn = current.PlayerID == i.Member.User.ID
					}

					components = []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.Button{
									Label:    "Attack",
									Style:    discordgo.DangerButton,
									CustomID: fmt.Sprintf("encounter:attack:%s", encounterID),
									Emoji:    &discordgo.ComponentEmoji{Name: "⚔️"},
									Disabled: !isPlayerTurn,
								},
								discordgo.Button{
									Label:    "Next Turn",
									Style:    discordgo.PrimaryButton,
									CustomID: fmt.Sprintf("encounter:next_turn:%s", encounterID),
									Emoji:    &discordgo.ComponentEmoji{Name: "➡️"},
								},
								discordgo.Button{
									Label:    "Status",
									Style:    discordgo.SecondaryButton,
									CustomID: fmt.Sprintf("encounter:view:%s", encounterID),
									Emoji:    &discordgo.ComponentEmoji{Name: "📊"},
								},
								discordgo.Button{
									Label:    "History",
									Style:    discordgo.SecondaryButton,
									CustomID: fmt.Sprintf("encounter:history:%s", encounterID),
									Emoji:    &discordgo.ComponentEmoji{Name: "📜"},
								},
							},
						},
					}
				}

				// Make the response ephemeral if this is an active encounter view
				flags := discordgo.MessageFlags(0)
				if encounterResult.Status == entities.EncounterStatusActive {
					flags = discordgo.MessageFlagsEphemeral
				}

				err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Embeds:     []*discordgo.MessageEmbed{embed},
						Components: components,
						Flags:      flags,
					},
				})
				if err != nil {
					log.Printf("Error showing encounter view: %v", err)
				}

			case "add_monster":
				// Prompt for monster name
				content := "Use `/dnd encounter add <monster>` to add a monster to this encounter."
				err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: content,
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				if err != nil {
					log.Printf("Error prompting for add monster: %v", err)
				}
			case "start":
				// Start combat
				err := h.ServiceProvider.EncounterService.StartEncounter(context.Background(), encounterID, i.Member.User.ID)
				if err != nil {
					content := fmt.Sprintf("❌ Failed to start combat: %v", err)
					if responseErr := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: content,
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					}); responseErr != nil {
						log.Printf("Failed to respond with error message: %v", responseErr)
					}
					return
				}

				// Get encounterResult to show combat status
				encounterResult, err := h.ServiceProvider.EncounterService.GetEncounter(context.Background(), encounterID)
				if err != nil {
					content := "✅ Combat started!"
					if responseErr := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: content,
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					}); responseErr != nil {
						log.Printf("Failed to respond with error message: %v", responseErr)
					}
					return
				}

				// Build combat tracker display
				embed := &discordgo.MessageEmbed{
					Title:       "⚔️ Combat Started!",
					Description: fmt.Sprintf("**%s** - Round %d", encounterResult.Name, encounterResult.Round),
					Color:       0xe74c3c, // Red
					Fields:      []*discordgo.MessageEmbedField{},
				}

				// Show current turn
				if current := encounterResult.GetCurrentCombatant(); current != nil {
					embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
						Name:   "🎯 Current Turn",
						Value:  fmt.Sprintf("**%s** (HP: %d/%d | AC: %d)", current.Name, current.CurrentHP, current.MaxHP, current.AC),
						Inline: false,
					})
				}

				// Show turn order
				var turnOrder strings.Builder
				for i, combatantID := range encounterResult.TurnOrder {
					if combatant, exists := encounterResult.Combatants[combatantID]; exists && combatant.IsActive {
						prefix := "  "
						if i == encounterResult.Turn {
							prefix = "▶️"
						}
						turnOrder.WriteString(fmt.Sprintf("%s %s (Initiative: %d)\n", prefix, combatant.Name, combatant.Initiative))
					}
				}

				embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
					Name:   "📋 Turn Order",
					Value:  turnOrder.String(),
					Inline: false,
				})

				// Add combat action buttons
				components := []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.Button{
								Label:    "Attack",
								Style:    discordgo.DangerButton,
								CustomID: fmt.Sprintf("encounter:attack:%s", encounterID),
								Emoji:    &discordgo.ComponentEmoji{Name: "⚔️"},
							},
							discordgo.Button{
								Label:    "Apply Damage",
								Style:    discordgo.DangerButton,
								CustomID: fmt.Sprintf("encounter:damage:%s", encounterID),
								Emoji:    &discordgo.ComponentEmoji{Name: "💥"},
							},
							discordgo.Button{
								Label:    "Heal",
								Style:    discordgo.SuccessButton,
								CustomID: fmt.Sprintf("encounter:heal:%s", encounterID),
								Emoji:    &discordgo.ComponentEmoji{Name: "💚"},
							},
							discordgo.Button{
								Label:    "Next Turn",
								Style:    discordgo.PrimaryButton,
								CustomID: fmt.Sprintf("encounter:next_turn:%s", encounterID),
								Emoji:    &discordgo.ComponentEmoji{Name: "➡️"},
							},
						},
					},
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.Button{
								Label:    "View Full",
								Style:    discordgo.SecondaryButton,
								CustomID: fmt.Sprintf("encounter:view_full:%s", encounterID),
								Emoji:    &discordgo.ComponentEmoji{Name: "📊"},
							},
							discordgo.Button{
								Label:    "End Combat",
								Style:    discordgo.SecondaryButton,
								CustomID: fmt.Sprintf("encounter:end:%s", encounterID),
								Emoji:    &discordgo.ComponentEmoji{Name: "🏁"},
							},
						},
					},
				}

				err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Embeds:     []*discordgo.MessageEmbed{embed},
						Components: components,
					},
				})
				if err != nil {
					log.Printf("Error showing combat started: %v", err)
				}
			case "next_turn":
				// Advance to next turn
				err := h.ServiceProvider.EncounterService.NextTurn(context.Background(), encounterID, i.Member.User.ID)
				if err != nil {
					content := fmt.Sprintf("❌ Failed to advance turn: %v", err)
					if responseErr := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: content,
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					}); responseErr != nil {
						log.Printf("Failed to respond with error message: %v", responseErr)
					}
					return
				}

				// Get updated encounter
				encounterResult, err := h.ServiceProvider.EncounterService.GetEncounter(context.Background(), encounterID)
				if err != nil {
					content := "✅ Turn advanced!"
					if responseErr := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: content,
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					}); responseErr != nil {
						log.Printf("Failed to respond with error message: %v", responseErr)
					}
					return
				}

				// Build turn update display
				embed := &discordgo.MessageEmbed{
					Title:       "➡️ Next Turn!",
					Description: fmt.Sprintf("**%s** - Round %d", encounterResult.Name, encounterResult.Round),
					Color:       0x3498db, // Blue
					Fields:      []*discordgo.MessageEmbedField{},
				}

				// Process any monster turns
				monsterActed := false
				if current := encounterResult.GetCurrentCombatant(); current != nil && current.Type == entities.CombatantTypeMonster && current.CanAct() {
					monsterActed = true
					log.Printf("Processing monster turn for %s (HP: %d/%d)", current.Name, current.CurrentHP, current.MaxHP)

					// Find a target (first active player)
					var target *entities.Combatant
					for _, combatant := range encounterResult.Combatants {
						if combatant.Type == entities.CombatantTypePlayer && combatant.IsActive {
							target = combatant
							break
						}
					}

					if target != nil && len(current.Actions) > 0 {
						// Use first available action
						action := current.Actions[0]

						// Roll attack
						attackResult, rollErr := dice.Roll(1, 20, 0)
						if rollErr != nil {
							log.Printf("Failed to roll attack: %v", rollErr)

							//TODO: Handle error
						}
						attackRoll := attackResult.Total
						totalAttack := attackRoll + action.AttackBonus

						embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
							Name:   "🐉 Monster Attack!",
							Value:  fmt.Sprintf("%s uses **%s** against %s", current.Name, action.Name, target.Name),
							Inline: false,
						})

						// Check if hit
						hit := totalAttack >= target.AC
						hitText := "❌ **MISS!**"
						if attackRoll == 20 {
							hitText = "🎆 **CRITICAL HIT!**"
							hit = true
						} else if attackRoll == 1 {
							hitText = "⚠️ **CRITICAL MISS!**"
							hit = false
						} else if hit {
							hitText = "✅ **HIT!**"
						}

						embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
							Name:   "Attack Roll",
							Value:  fmt.Sprintf("🎲 %d + %d = **%d** vs AC %d\n%s", attackRoll, action.AttackBonus, totalAttack, target.AC, hitText),
							Inline: false,
						})

						// Apply damage if hit
						if hit && len(action.Damage) > 0 {
							totalDamage := 0
							var damageDetails strings.Builder

							for _, dmg := range action.Damage {
								diceCount := dmg.DiceCount
								if attackRoll == 20 { // Critical hit doubles dice
									diceCount *= 2
								}
								rollResult, rollErr := dice.Roll(diceCount, dmg.DiceSize, dmg.Bonus)
								if rollErr != nil {
									log.Printf("Failed to roll damage: %v", rollErr)
									//TODO: Handle error
								}
								dmgTotal := rollResult.Total
								totalDamage += dmgTotal
								damageDetails.WriteString(fmt.Sprintf("🎲 %dd%d+%d = **%d** %s\n", diceCount, dmg.DiceSize, dmg.Bonus, dmgTotal, dmg.DamageType))
							}

							embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
								Name:   "Damage Roll",
								Value:  damageDetails.String(),
								Inline: false,
							})

							// Apply damage
							log.Printf("Monster %s dealt %d damage to %s", current.Name, totalDamage, target.Name)

							err = h.ServiceProvider.EncounterService.ApplyDamage(context.Background(), encounterID, target.ID, i.Member.User.ID, totalDamage)
							if err != nil {
								log.Printf("Error applying monster damage: %v", err)
							} else {
								target.CurrentHP -= totalDamage
								if target.CurrentHP < 0 {
									target.CurrentHP = 0
								}

								embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
									Name:   "🩸 Target Status",
									Value:  fmt.Sprintf("%s now has **%d/%d HP**", target.Name, target.CurrentHP, target.MaxHP),
									Inline: false,
								})

								if target.CurrentHP == 0 {
									embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
										Name:   "💀 Player Down!",
										Value:  fmt.Sprintf("%s has been knocked unconscious!", target.Name),
										Inline: false,
									})
								}
							}
						} else {
							// Log miss
							err = h.ServiceProvider.EncounterService.LogCombatAction(context.Background(), encounterID,
								fmt.Sprintf("%s missed %s", current.Name, target.Name))
							if err != nil {
								log.Printf("Error logging miss: %v", err)
							}
						}
					} else if target == nil {
						embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
							Name:   "🎯 No Targets",
							Value:  "The monster has no valid targets!",
							Inline: false,
						})
					}
				}

				// If a monster acted, auto-advance and update display
				if monsterActed {
					log.Printf("Auto-advancing turn after monster attack")
					err = h.ServiceProvider.EncounterService.NextTurn(context.Background(), encounterID, i.Member.User.ID)
					if err != nil {
						log.Printf("Error auto-advancing turn: %v", err)
					} else {
						// Re-get encounter to show updated turn
						encounterResult, err = h.ServiceProvider.EncounterService.GetEncounter(context.Background(), encounterID)
						if err != nil {
							log.Printf("Error getting updated encounter: %v", err)
						}
					}
				}

				// Check if combat ended
				if encounterResult.Status == entities.EncounterStatusCompleted {
					// Show victory/defeat message
					shouldEnd, playersWon := encounterResult.CheckCombatEnd()
					if shouldEnd && playersWon {
						embed.Title = "🎉 Victory!"
						embed.Color = 0x2ecc71 // Green
						embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
							Name:   "Combat Complete",
							Value:  "All enemies have been defeated! The party is victorious!",
							Inline: false,
						})
					} else if shouldEnd && !playersWon {
						embed.Title = "💀 Defeat..."
						embed.Color = 0xe74c3c // Red
						embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
							Name:   "Combat Complete",
							Value:  "The party has fallen in battle...",
							Inline: false,
						})
					}
				} else {
					// Show current turn after any updates
					if current := encounterResult.GetCurrentCombatant(); current != nil {
						embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
							Name:   "🎯 Now Up",
							Value:  fmt.Sprintf("**%s** (HP: %d/%d | AC: %d)", current.Name, current.CurrentHP, current.MaxHP, current.AC),
							Inline: false,
						})
					}
				}

				// Show upcoming turns
				var upcoming strings.Builder
				for i := 0; i < 3 && i < len(encounterResult.TurnOrder); i++ {
					idx := (encounterResult.Turn + i) % len(encounterResult.TurnOrder)
					if combatant, exists := encounterResult.Combatants[encounterResult.TurnOrder[idx]]; exists && combatant.IsActive {
						if i == 0 {
							upcoming.WriteString(fmt.Sprintf("▶️ **%s** (current)\n", combatant.Name))
						} else {
							upcoming.WriteString(fmt.Sprintf("%d. %s\n", i, combatant.Name))
						}
					}
				}

				embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
					Name:   "📋 Turn Order",
					Value:  upcoming.String(),
					Inline: true,
				})

				// Combat action buttons
				var components []discordgo.MessageComponent
				if encounterResult.Status == entities.EncounterStatusCompleted {
					// Combat ended - show different buttons
					components = []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.Button{
									Label:    "View History",
									Style:    discordgo.SecondaryButton,
									CustomID: fmt.Sprintf("encounter:history:%s", encounterID),
									Emoji:    &discordgo.ComponentEmoji{Name: "📜"},
								},
								discordgo.Button{
									Label:    "Continue Dungeon",
									Style:    discordgo.SuccessButton,
									CustomID: fmt.Sprintf("dungeon:next_room:%s", encounterResult.SessionID),
									Emoji:    &discordgo.ComponentEmoji{Name: "🚪"},
								},
							},
						},
					}
				} else if encounterResult.IsRoundComplete() {
					// Round is complete - show continue button
					embed.Title = "🔄 Round Complete!"
					embed.Fields = append([]*discordgo.MessageEmbedField{
						{
							Name:   "📊 Round Summary",
							Value:  fmt.Sprintf("Round %d has ended. All combatants have acted.", encounterResult.Round),
							Inline: false,
						},
					}, embed.Fields...)

					components = []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.Button{
									Label:    "Continue to Next Round",
									Style:    discordgo.SuccessButton,
									CustomID: fmt.Sprintf("encounter:next_turn:%s", encounterID),
									Emoji:    &discordgo.ComponentEmoji{Name: "▶️"},
								},
								discordgo.Button{
									Label:    "View Status",
									Style:    discordgo.SecondaryButton,
									CustomID: fmt.Sprintf("encounter:view_full:%s", encounterID),
									Emoji:    &discordgo.ComponentEmoji{Name: "📊"},
								},
							},
						},
					}
				} else {
					// Combat ongoing - show normal buttons
					// Check if it's the player's turn
					isPlayerTurn := false
					if current := encounterResult.GetCurrentCombatant(); current != nil {
						isPlayerTurn = current.PlayerID == i.Member.User.ID
					}

					components = []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.Button{
									Label:    "Attack",
									Style:    discordgo.DangerButton,
									CustomID: fmt.Sprintf("encounter:attack:%s", encounterID),
									Emoji:    &discordgo.ComponentEmoji{Name: "⚔️"},
									Disabled: !isPlayerTurn,
								},
								discordgo.Button{
									Label:    "Apply Damage",
									Style:    discordgo.DangerButton,
									CustomID: fmt.Sprintf("encounter:damage:%s", encounterID),
									Emoji:    &discordgo.ComponentEmoji{Name: "💥"},
								},
								discordgo.Button{
									Label:    "Heal",
									Style:    discordgo.SuccessButton,
									CustomID: fmt.Sprintf("encounter:heal:%s", encounterID),
									Emoji:    &discordgo.ComponentEmoji{Name: "💚"},
								},
								discordgo.Button{
									Label:    "Next Turn",
									Style:    discordgo.PrimaryButton,
									CustomID: fmt.Sprintf("encounter:next_turn:%s", encounterID),
									Emoji:    &discordgo.ComponentEmoji{Name: "➡️"},
								},
								discordgo.Button{
									Label:    "View Full",
									Style:    discordgo.SecondaryButton,
									CustomID: fmt.Sprintf("encounter:view_full:%s", encounterID),
									Emoji:    &discordgo.ComponentEmoji{Name: "📊"},
								},
							},
						},
					}
				}

				err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Embeds:     []*discordgo.MessageEmbed{embed},
						Components: components,
					},
				})
				if err != nil {
					log.Printf("Error showing next turn: %v", err)
				}
			case "damage":
				// Show damage modal
				modal := discordgo.InteractionResponseData{
					CustomID: fmt.Sprintf("encounter:apply_damage:%s", encounterID),
					Title:    "Apply Damage",
					Components: []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:    "damage_amount",
									Label:       "Damage Amount",
									Style:       discordgo.TextInputShort,
									Placeholder: "Enter damage (e.g., 12)",
									Required:    true,
									MaxLength:   3,
								},
							},
						},
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:    "target_name",
									Label:       "Target Name",
									Style:       discordgo.TextInputShort,
									Placeholder: "Enter target name",
									Required:    true,
									MaxLength:   50,
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
					log.Printf("Error showing damage modal: %v", err)
				}
			case "heal":
				// Show heal modal
				modal := discordgo.InteractionResponseData{
					CustomID: fmt.Sprintf("encounter:apply_heal:%s", encounterID),
					Title:    "Heal Target",
					Components: []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:    "heal_amount",
									Label:       "Healing Amount",
									Style:       discordgo.TextInputShort,
									Placeholder: "Enter healing (e.g., 8)",
									Required:    true,
									MaxLength:   3,
								},
							},
						},
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:    "target_name",
									Label:       "Target Name",
									Style:       discordgo.TextInputShort,
									Placeholder: "Enter target name",
									Required:    true,
									MaxLength:   50,
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
					log.Printf("Error showing heal modal: %v", err)
				}
			case "end":
				// End the encounter
				err := h.ServiceProvider.EncounterService.EndEncounter(context.Background(), encounterID, i.Member.User.ID)
				if err != nil {
					content := fmt.Sprintf("❌ Failed to end encounter: %v", err)
					if responseErr := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: content,
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					}); responseErr != nil {
						log.Printf("Failed to respond with error message: %v", responseErr)
					}
					return
				}

				// Show end summary
				embed := &discordgo.MessageEmbed{
					Title:       "🏁 Combat Ended!",
					Description: "The encounter has concluded.",
					Color:       0x2ecc71, // Green
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:   "📊 Summary",
							Value:  "Combat statistics will be available in a future update!",
							Inline: false,
						},
					},
				}

				err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Embeds: []*discordgo.MessageEmbed{embed},
					},
				})
				if err != nil {
					log.Printf("Error showing end combat: %v", err)
				}
			case "history":
				// View combat history
				encounterResult, err := h.ServiceProvider.EncounterService.GetEncounter(context.Background(), encounterID)
				if err != nil {
					content := fmt.Sprintf("❌ Failed to get encounter: %v", err)
					err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: content,
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					if err != nil {
						log.Printf("Error responding with error: %v", err)
					}
					return
				}

				// Build history embed
				embed := &discordgo.MessageEmbed{
					Title:       fmt.Sprintf("📜 Combat History - %s", encounterResult.Name),
					Description: fmt.Sprintf("Round %d", encounterResult.Round),
					Color:       0x9b59b6, // Purple
					Fields:      []*discordgo.MessageEmbedField{},
				}

				// Show recent combat log
				if len(encounterResult.CombatLog) > 0 {
					var logText strings.Builder
					// Show last 10 entries
					start := 0
					if len(encounterResult.CombatLog) > 10 {
						start = len(encounterResult.CombatLog) - 10
					}

					for i := start; i < len(encounterResult.CombatLog); i++ {
						logText.WriteString(encounterResult.CombatLog[i] + "\n")
					}

					embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
						Name:   "Recent Actions",
						Value:  logText.String(),
						Inline: false,
					})
				} else {
					embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
						Name:   "Recent Actions",
						Value:  "*No combat actions yet*",
						Inline: false,
					})
				}

				// Add back button
				components := []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.Button{
								Label:    "Back",
								Style:    discordgo.SecondaryButton,
								CustomID: fmt.Sprintf("encounter:view:%s", encounterID),
								Emoji:    &discordgo.ComponentEmoji{Name: "🔙"},
							},
						},
					},
				}

				err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Embeds:     []*discordgo.MessageEmbed{embed},
						Components: components,
						Flags:      discordgo.MessageFlagsEphemeral,
					},
				})
				if err != nil {
					log.Printf("Error showing combat history: %v", err)
				}
			case "attack":
				log.Printf("Attack button pressed for encounter: %s by user: %s", encounterID, i.Member.User.ID)
				// Simple attack handler for testing
				// Get encounter
				encounterResult, err := h.ServiceProvider.EncounterService.GetEncounter(context.Background(), encounterID)
				if err != nil {
					// For dungeon encounters, try to find the active encounter for the channel
					log.Printf("Failed to get encounter %s, looking for active encounter in channel", encounterID)

					// Try to get session from channel metadata or find active encounter
					// This is a workaround for stale encounter IDs in old Discord messages
					content := "❌ This encounter has expired. Please start a new room!"
					if responseErr := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: content,
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					}); responseErr != nil {
						log.Printf("Failed to respond with error message: %v", responseErr)
					}
					return
				}

				// Find the attacker - the player who clicked attack
				var current *entities.Combatant
				attackerID := i.Member.User.ID

				// Find the player's combatant
				log.Printf("Looking for attacker with PlayerID=%s among %d combatants", attackerID, len(encounterResult.Combatants))
				for id, combatant := range encounterResult.Combatants {
					log.Printf("Checking combatant %s: Name=%s, Type=%s, PlayerID=%s", id, combatant.Name, combatant.Type, combatant.PlayerID)
					if combatant.PlayerID == attackerID {
						current = combatant
						log.Printf("Found player %s attacking as %s", attackerID, combatant.Name)
						break
					}
				}

				// If player has no combatant, check if they're the DM
				if current == nil && encounterResult.CreatedBy == attackerID {
					// DM can control current turn's combatant
					current = encounterResult.GetCurrentCombatant()
					if current != nil {
						log.Printf("DM controlling %s for attack", current.Name)
					}
				}

				if current == nil {
					content := "❌ You don't have a character in this encounter!"
					if responseErr := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: content,
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					}); responseErr != nil {
						log.Printf("Failed to respond with error message: %v", responseErr)
					}
					return
				}

				// Show target selection
				// Build target list from encounter combatants
				var targetButtons []discordgo.MessageComponent
				targetCount := 0

				for id, combatant := range encounterResult.Combatants {
					// Don't show self as target, inactive combatants, or defeated enemies
					if combatant.ID == current.ID || !combatant.IsActive || combatant.CurrentHP <= 0 {
						log.Printf("Skipping combatant %s (ID: %s): self=%v, active=%v, HP=%d/%d",
							combatant.Name, id, combatant.ID == current.ID, combatant.IsActive, combatant.CurrentHP, combatant.MaxHP)
						continue
					}

					log.Printf("Adding target button for %s (ID: %s)", combatant.Name, id)
					// Create button for this target
					emoji := "🧑"
					if combatant.Type == entities.CombatantTypeMonster {
						emoji = "👹"
					}

					targetButtons = append(targetButtons, discordgo.Button{
						Label:    fmt.Sprintf("%s (HP: %d/%d)", combatant.Name, combatant.CurrentHP, combatant.MaxHP),
						Style:    discordgo.PrimaryButton,
						CustomID: fmt.Sprintf("encounter:select_target:%s:%s", encounterID, id),
						Emoji:    &discordgo.ComponentEmoji{Name: emoji},
					})
					targetCount++

					// Discord limits 5 buttons per row
					if targetCount >= 5 {
						break
					}
				}

				if len(targetButtons) == 0 {
					content := "❌ No valid targets available!"
					if responseErr := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: content,
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					}); responseErr != nil {
						log.Printf("Failed to respond with error message: %v", responseErr)
					}
					return
				}

				// Create embed for target selection
				embed := &discordgo.MessageEmbed{
					Title:       fmt.Sprintf("⚔️ %s's Attack", current.Name),
					Description: "Select your target:",
					Color:       0xe74c3c,
				}

				components := []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: targetButtons,
					},
				}

				err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Embeds:     []*discordgo.MessageEmbed{embed},
						Components: components,
						Flags:      discordgo.MessageFlagsEphemeral,
					},
				})
				if err != nil {
					log.Printf("Error showing target selection: %v", err)
				}
			case "select_target":
				log.Printf("=== ENTERING select_target handler ===")
				// Handle target selection for attack
				if len(parts) < 4 {
					log.Printf("Invalid select_target interaction: %v", parts)
					err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: "❌ Invalid target selection",
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					})
					if err != nil {
						log.Printf("Failed to respond with error message: %v", err)
					}
					return
				}
				targetID := parts[3]
				log.Printf("Target selected: %s for encounter: %s", targetID, encounterID)

				// Defer the response immediately since attack processing can take time
				err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseDeferredMessageUpdate,
				})
				if err != nil {
					log.Printf("Failed to defer response: %v", err)
				}

				// Use the new service method to handle the complete attack sequence
				attackResult, err := h.ServiceProvider.EncounterService.ExecuteAttackWithTarget(
					context.Background(),
					&encounter.ExecuteAttackInput{
						EncounterID: encounterID,
						TargetID:    targetID,
						UserID:      i.Member.User.ID,
					},
				)
				if err != nil {
					log.Printf("Error executing attack: %v", err)
					content := fmt.Sprintf("❌ Failed to execute attack: %v", err)
					_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
						Content: &content,
						Embeds:  &[]*discordgo.MessageEmbed{},
					})
					if err != nil {
						log.Printf("Failed to edit interaction response: %v", err)
					}
					return
				}

				// Build result embed
				playerAttack := attackResult.PlayerAttack
				embed := &discordgo.MessageEmbed{
					Title:       fmt.Sprintf("⚔️ %s attacks %s!", playerAttack.AttackerName, playerAttack.TargetName),
					Description: fmt.Sprintf("**Attack:** %s", playerAttack.WeaponName),
					Color:       0xe74c3c,
					Fields:      []*discordgo.MessageEmbedField{},
				}

				// Display attack roll and result
				var hitText string
				if playerAttack.Critical {
					hitText = "🎆 **CRITICAL HIT!**"
				} else if playerAttack.AttackRoll == 1 {
					hitText = "⚠️ **CRITICAL MISS!**"
				} else if playerAttack.Hit {
					hitText = "✅ **HIT!**"
				} else {
					hitText = "❌ **MISS!**"
				}

				// Attack roll details
				attackDetails := fmt.Sprintf("Roll: %v + %d = **%d**\nvs AC %d\n%s",
					playerAttack.DiceRolls,
					playerAttack.AttackBonus,
					playerAttack.TotalAttack,
					playerAttack.TargetAC,
					hitText)

				embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
					Name:   "🎲 Attack Roll",
					Value:  attackDetails,
					Inline: true,
				})

				// Damage details if hit
				if playerAttack.Hit && playerAttack.Damage > 0 {
					damageDetails := fmt.Sprintf("Roll: %v",
						playerAttack.DamageRolls)
					if playerAttack.DamageBonus != 0 {
						damageDetails += fmt.Sprintf(" + %d", playerAttack.DamageBonus)
					}
					damageDetails += fmt.Sprintf(" = **%d** %s", playerAttack.Damage, playerAttack.DamageType)

					embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
						Name:   "💥 Damage",
						Value:  damageDetails,
						Inline: true,
					}, &discordgo.MessageEmbedField{
						Name:   "🩸 Target Status",
						Value:  fmt.Sprintf("%s now has **%d HP**", playerAttack.TargetName, playerAttack.TargetNewHP),
						Inline: false,
					})

					if playerAttack.TargetDefeated {
						embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
							Name:   "💀 Defeated!",
							Value:  fmt.Sprintf("%s has been defeated!", playerAttack.TargetName),
							Inline: false,
						})
					}
				}

				// Display any monster attacks that followed
				for _, monsterAttack := range attackResult.MonsterAttacks {
					var monsterValue string
					if monsterAttack.Hit {
						monsterValue = fmt.Sprintf("%s attacks %s with %s!\n🎲 Attack: %d vs AC %d - **HIT!**\n💥 Damage: **%d**",
							monsterAttack.AttackerName, monsterAttack.TargetName, monsterAttack.WeaponName,
							monsterAttack.TotalAttack, monsterAttack.TargetAC, monsterAttack.Damage)
					} else {
						monsterValue = fmt.Sprintf("%s attacks %s with %s!\n🎲 Attack: %d vs AC %d - **MISS!**",
							monsterAttack.AttackerName, monsterAttack.TargetName, monsterAttack.WeaponName,
							monsterAttack.TotalAttack, monsterAttack.TargetAC)
					}
					embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
						Name:   fmt.Sprintf("🐉 %s's Turn", monsterAttack.AttackerName),
						Value:  monsterValue,
						Inline: false,
					})
				}

				// Check if combat ended
				if attackResult.CombatEnded && attackResult.PlayersWon {
					embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
						Name:   "🎉 Victory!",
						Value:  "All enemies have been defeated!",
						Inline: false,
					})
				}

				// Add combat control buttons
				components := []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.Button{
								Label:    "Attack Again",
								Style:    discordgo.DangerButton,
								CustomID: fmt.Sprintf("encounter:attack:%s", encounterID),
								Emoji:    &discordgo.ComponentEmoji{Name: "⚔️"},
								Disabled: !attackResult.IsPlayerTurn || attackResult.CombatEnded,
							},
							discordgo.Button{
								Label:    "Next Turn",
								Style:    discordgo.PrimaryButton,
								CustomID: fmt.Sprintf("encounter:next_turn:%s", encounterID),
								Emoji:    &discordgo.ComponentEmoji{Name: "➡️"},
								Disabled: attackResult.CombatEnded,
							},
							discordgo.Button{
								Label:    "View Status",
								Style:    discordgo.SecondaryButton,
								CustomID: fmt.Sprintf("encounter:view:%s", encounterID),
								Emoji:    &discordgo.ComponentEmoji{Name: "📊"},
							},
							discordgo.Button{
								Label:    "History",
								Style:    discordgo.SecondaryButton,
								CustomID: fmt.Sprintf("encounter:history:%s", encounterID),
								Emoji:    &discordgo.ComponentEmoji{Name: "📜"},
							},
						},
					},
				}

				_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
					Embeds:     &[]*discordgo.MessageEmbed{embed},
					Components: &components,
				})
				if err != nil {
					log.Printf("Error showing attack result: %v", err)
				}
			default:
				log.Printf("Unknown encounter action: %s", action)
			}
		}
	} else if ctx == "ea" {
		// Handle execute attack with shortened IDs
		// ea:{encIDShort}:{targetIDShort}:{attackType}
		if len(parts) < 4 {
			log.Printf("Invalid ea interaction: %v", parts)
			return
		}

		encIDShort := parts[1]
		targetIDShort := parts[2]
		attackType := parts[3]

		log.Printf("Execute attack - enc: %s, target: %s, type: %s", encIDShort, targetIDShort, attackType)

		// We need to find the full encounter ID and target ID
		// First, let's try to get active encounter for the session
		// This requires us to know the session ID from context

		// For now, we'll search through all encounters (this is not ideal for production)
		// In production, you'd want to store session context or use a more efficient lookup

		// Alternative approach: Store the full IDs in the message metadata or use a cache
		// For now, let's respond with an error and ask user to try again
		content := "⚠️ Attack action requires full context. Please use the Attack button to select your target again."
		if responseErr := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content:    content,
				Components: []discordgo.MessageComponent{},
			},
		}); responseErr != nil {
			log.Printf("Failed to respond with error message: %v", responseErr)
		}
	} else if ctx == "dungeon" {
		// Handle dungeon actions
		if len(parts) >= 3 {
			sessionID := parts[2]

			switch action {
			case "join":
				if err := h.dungeonJoinHandler.HandleButton(s, i, sessionID); err != nil {
					log.Printf("Error handling dungeon join: %v", err)
				}
			case "select_character":
				// Handle character selection for dungeon
				if len(i.MessageComponentData().Values) > 0 {
					characterID := i.MessageComponentData().Values[0]

					// Get the session to check if user is already in it
					sess, err := h.ServiceProvider.SessionService.GetSession(context.Background(), sessionID)
					if err != nil {
						respondWithUpdateError(s, i, fmt.Sprintf("Failed to get session: %v", err))
						return
					}

					// If not in session, join it first
					if !sess.IsUserInSession(i.Member.User.ID) {
						log.Printf("User %s not in session, joining...", i.Member.User.ID)
						_, joinErr := h.ServiceProvider.SessionService.JoinSession(context.Background(), sessionID, i.Member.User.ID)
						if joinErr != nil {
							respondWithUpdateError(s, i, fmt.Sprintf("Failed to join party: %v", joinErr))
							return
						}
					}

					// Select the character
					err = h.ServiceProvider.SessionService.SelectCharacter(context.Background(), sessionID, i.Member.User.ID, characterID)
					if err != nil {
						respondWithUpdateError(s, i, fmt.Sprintf("Failed to select character: %v", err))
						return
					}

					// Get character details for confirmation
					char, charErr := h.ServiceProvider.CharacterService.GetByID(characterID)
					if charErr != nil {
						respondWithUpdateError(s, i, fmt.Sprintf("Failed to get character: %v", charErr))
						return
					}

					// Success response
					embed := &discordgo.MessageEmbed{
						Title:       "🎉 Joined the Party!",
						Description: fmt.Sprintf("**%s** has joined the dungeon delve!", char.Name),
						Color:       0x2ecc71, // Green
						Fields: []*discordgo.MessageEmbedField{
							{
								Name:   "Character",
								Value:  fmt.Sprintf("%s (Level %d)", char.GetDisplayInfo(), char.Level),
								Inline: true,
							},
							{
								Name:   "HP",
								Value:  fmt.Sprintf("%d/%d", char.CurrentHitPoints, char.MaxHitPoints),
								Inline: true,
							},
							{
								Name:   "AC",
								Value:  fmt.Sprintf("%d", char.AC),
								Inline: true,
							},
						},
					}

					if responseErr := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseUpdateMessage,
						Data: &discordgo.InteractionResponseData{
							Embeds:     []*discordgo.MessageEmbed{embed},
							Components: []discordgo.MessageComponent{},
						},
					}); responseErr != nil {
						log.Printf("Failed to respond with success message: %v", responseErr)
					}

					// Update the shared dungeon lobby message using stored message ID
					freshSess, err := h.ServiceProvider.SessionService.GetSession(context.Background(), sessionID)
					if err != nil {
						log.Printf("Failed to get session: %v", err)
						return
					}

					if freshSess != nil && freshSess.Metadata != nil {
						// Get stored message ID from session metadata
						if messageID, ok := freshSess.Metadata["lobbyMessageID"].(string); ok {
							if channelID, ok := freshSess.Metadata["lobbyChannelID"].(string); ok {
								log.Printf("Updating dungeon lobby message %s with new party member", messageID)

								// Use the helper function to update the lobby message
								if updateErr := dungeon.UpdateDungeonLobbyMessage(s, h.ServiceProvider.SessionService, h.ServiceProvider.CharacterService, sessionID, messageID, channelID); updateErr != nil {
									log.Printf("Failed to update dungeon lobby message: %v", updateErr)
								} else {
									log.Printf("Successfully updated dungeon lobby with new party member")
								}
							}
						} else {
							log.Printf("No lobby message ID found in session metadata")
						}
					}
				}
			case "enter":
				if len(parts) >= 4 {
					roomType := parts[3]
					if err := h.dungeonEnterRoomHandler.HandleButton(s, i, sessionID, roomType); err != nil {
						log.Printf("Error handling dungeon enter: %v", err)
					}
				}
			case "status":
				// Get session to show party status
				sess, err := h.ServiceProvider.SessionService.GetSession(context.Background(), sessionID)
				if err != nil {
					content := "❌ Session not found!"
					if responseErr := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: content,
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					}); responseErr != nil {
						log.Printf("Failed to respond with error message: %v", responseErr)
					}
					return
				}

				// Build party status
				embed := &discordgo.MessageEmbed{
					Title:       "🎭 Party Status",
					Description: "Current adventurers:",
					Color:       0x3498db,
					Fields:      []*discordgo.MessageEmbedField{},
				}

				for _, member := range sess.Members {
					if member.CharacterID != "" {
						// Get character details
						char, getCharErr := h.ServiceProvider.CharacterService.GetByID(member.CharacterID)
						if getCharErr == nil {
							embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
								Name:   char.Name,
								Value:  fmt.Sprintf("%s | HP: %d/%d | AC: %d", char.GetDisplayInfo(), char.CurrentHitPoints, char.MaxHitPoints, char.AC),
								Inline: false,
							})
						}
					}
				}

				if len(embed.Fields) == 0 {
					embed.Description = "No party members yet!"
				}

				err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Embeds: []*discordgo.MessageEmbed{embed},
						Flags:  discordgo.MessageFlagsEphemeral,
					},
				})
				if err != nil {
					log.Printf("Error showing party status: %v", err)
				}
			}
		}
	} else if ctx == "saving_throw" {
		// Handle saving throw rolls
		if len(parts) >= 4 {
			characterID := parts[1]
			attribute := entities.Attribute(parts[2])
			dc, err := strconv.Atoi(parts[3])
			if err == nil {
				if err := h.savingThrowHandler.HandleSavingThrowRoll(s, i, characterID, attribute, dc); err != nil {
					log.Printf("Error handling saving throw: %v", err)
				}
			}
		}
	} else if ctx == "skill_check" {
		// Handle skill check rolls
		if len(parts) >= 4 {
			characterID := parts[1]
			skillKey := parts[2]
			dc, err := strconv.Atoi(parts[3])
			if err == nil {
				if err := h.skillCheckHandler.HandleSkillCheckRoll(s, i, characterID, skillKey, dc); err != nil {
					log.Printf("Error handling skill check: %v", err)
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
				getDraftErr := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "❌ Failed to get character draft. Please try again.",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				if getDraftErr != nil {
					log.Printf("Error responding with error: %v", getDraftErr)
				}
				return
			}

			// Update the draft with the final name and ensure race/class are set
			raceKey := parts[2]
			classKey := parts[3]

			_, err = h.ServiceProvider.CharacterService.UpdateDraftCharacter(
				context.Background(),
				draftChar.ID,
				&characterService.UpdateDraftInput{
					Name:     &characterName,
					RaceKey:  &raceKey,
					ClassKey: &classKey,
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
				getCharErr := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "❌ Failed to finalize character. Please try again.",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				if getCharErr != nil {
					log.Printf("Error responding with error: %v", getCharErr)
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

			// Finalize the draft character (marking it as active)
			finalChar, err := h.ServiceProvider.CharacterService.FinalizeDraftCharacter(context.Background(), draftChar.ID)
			if err != nil {
				log.Printf("Error finalizing character: %v", err)
				// Show error to user
				finalCharErr := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: fmt.Sprintf("❌ Failed to finalize character: %v", err),
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				if finalCharErr != nil {
					log.Printf("Error responding with error: %v", finalCharErr)
				}
				return
			}

			// Show success with character details
			description := fmt.Sprintf("**Name:** %s", finalChar.Name)
			if finalChar.Race != nil {
				description += fmt.Sprintf("\n**Race:** %s", finalChar.Race.Name)
			}
			if finalChar.Class != nil {
				description += fmt.Sprintf("\n**Class:** %s", finalChar.Class.Name)
			}

			embed := &discordgo.MessageEmbed{
				Title:       "Character Created!",
				Description: description,
				Color:       0x00ff00,
				Fields:      []*discordgo.MessageEmbedField{},
			}

			// Only add ability scores if we have them
			if len(abilityScores) > 0 {
				embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
					Name: "💪 Base Abilities",
					Value: fmt.Sprintf("STR: %d, DEX: %d, CON: %d\nINT: %d, WIS: %d, CHA: %d",
						abilityScores["STR"], abilityScores["DEX"], abilityScores["CON"],
						abilityScores["INT"], abilityScores["WIS"], abilityScores["CHA"],
					),
					Inline: true,
				})
			}

			// Add other fields
			embed.Fields = append(embed.Fields,
				&discordgo.MessageEmbedField{
					Name:   "❤️ Hit Points",
					Value:  fmt.Sprintf("%d", finalChar.MaxHitPoints),
					Inline: true,
				},
				&discordgo.MessageEmbedField{
					Name:   "🛡️ Hit Die",
					Value:  fmt.Sprintf("d%d", finalChar.HitDie),
					Inline: true,
				},
				&discordgo.MessageEmbedField{
					Name:   "✅ Character Complete",
					Value:  "Your character has been created and saved successfully!",
					Inline: false,
				},
			)

			embed.Footer = &discordgo.MessageEmbedFooter{
				Text: "Ready for adventure!",
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
	} else if ctx == "encounter" {
		if len(parts) >= 3 {
			encounterID := parts[2]

			switch action {
			case "apply_damage":
				// Extract values from modal
				damageAmount := 0
				targetName := ""

				for _, comp := range data.Components {
					if row, ok := comp.(*discordgo.ActionsRow); ok {
						for _, rowComp := range row.Components {
							if input, ok := rowComp.(*discordgo.TextInput); ok {
								switch input.CustomID {
								case "damage_amount":
									var err error
									damageAmount, err = strconv.Atoi(input.Value)
									if err != nil {
										log.Printf("Error parsing damage amount: %v", err)
										return
									}
								case "target_name":
									targetName = input.Value
								}
							}
						}
					}
				}

				// Get encounter to find target
				encounterResult, err := h.ServiceProvider.EncounterService.GetEncounter(context.Background(), encounterID)
				if err != nil {
					content := fmt.Sprintf("❌ Failed to get encounter: %v", err)
					if responseErr := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: content,
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					}); responseErr != nil {
						log.Printf("Failed to respond with error message: %v", responseErr)
					}
					return
				}

				// Find target combatant
				var targetID string
				for id, combatant := range encounterResult.Combatants {
					if strings.EqualFold(combatant.Name, targetName) {
						targetID = id
						break
					}
				}

				if targetID == "" {
					content := fmt.Sprintf("❌ Target '%s' not found in encounter!", targetName)
					if responseErr := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: content,
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					}); responseErr != nil {
						log.Printf("Failed to respond with error message: %v", responseErr)
					}
					return
				}

				// Apply damage
				err = h.ServiceProvider.EncounterService.ApplyDamage(context.Background(), encounterID, targetID, i.Member.User.ID, damageAmount)
				if err != nil {
					content := fmt.Sprintf("❌ Failed to apply damage: %v", err)
					if responseErr := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: content,
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					}); responseErr != nil {
						log.Printf("Failed to respond with error message: %v", responseErr)
					}
					return
				}

				// Get updated combatant
				encounterResult, err = h.ServiceProvider.EncounterService.GetEncounter(context.Background(), encounterID)
				if err != nil {
					content := fmt.Sprintf("❌ Failed to get encounter: %v", err)
					if responseErr := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: content,
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					}); responseErr != nil {
						log.Printf("Failed to respond with error message: %v", responseErr)
					}
					return
				}
				target := encounterResult.Combatants[targetID]

				// Show result
				embed := &discordgo.MessageEmbed{
					Title:       "💥 Damage Applied!",
					Description: fmt.Sprintf("**%s** takes %d damage!", target.Name, damageAmount),
					Color:       0xe74c3c, // Red
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:   "❤️ Hit Points",
							Value:  fmt.Sprintf("%d / %d", target.CurrentHP, target.MaxHP),
							Inline: true,
						},
					},
				}

				if target.CurrentHP <= 0 {
					embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
						Name:   "💀 Status",
						Value:  "**Unconscious!**",
						Inline: true,
					})
				}

				if responseErr := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Embeds: []*discordgo.MessageEmbed{embed},
					},
				}); responseErr != nil {
					log.Printf("Failed to respond with error message: %v", responseErr)
				}
				if err != nil {
					log.Printf("Error responding to damage: %v", err)
				}

			case "apply_heal":
				// Extract values from modal
				healAmount := 0
				targetName := ""

				for _, comp := range data.Components {
					if row, ok := comp.(*discordgo.ActionsRow); ok {
						for _, rowComp := range row.Components {
							if input, ok := rowComp.(*discordgo.TextInput); ok {
								switch input.CustomID {
								case "heal_amount":
									var err error
									healAmount, err = strconv.Atoi(input.Value)
									if err != nil {
										log.Printf("Error parsing heal amount: %v", err)
										return
									}
								case "target_name":
									targetName = input.Value
								}
							}
						}
					}
				}

				// Get encounter to find target
				encounterResult, err := h.ServiceProvider.EncounterService.GetEncounter(context.Background(), encounterID)
				if err != nil {
					content := fmt.Sprintf("❌ Failed to get encounter: %v", err)
					if responseErr := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: content,
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					}); responseErr != nil {
						log.Printf("Failed to respond with error message: %v", responseErr)
					}
					return
				}

				// Find target combatant
				var targetID string
				for id, combatant := range encounterResult.Combatants {
					if strings.EqualFold(combatant.Name, targetName) {
						targetID = id
						break
					}
				}

				if targetID == "" {
					content := fmt.Sprintf("❌ Target '%s' not found in encounter!", targetName)
					if responseErr := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: content,
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					}); responseErr != nil {
						log.Printf("Failed to respond with error message: %v", responseErr)
					}
					return
				}

				// Apply healing
				err = h.ServiceProvider.EncounterService.HealCombatant(context.Background(), encounterID, targetID, i.Member.User.ID, healAmount)
				if err != nil {
					content := fmt.Sprintf("❌ Failed to apply healing: %v", err)
					if responseErr := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: content,
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					}); responseErr != nil {
						log.Printf("Failed to respond with error message: %v", responseErr)
					}
					return
				}

				// Get updated combatant
				encounterResult, err = h.ServiceProvider.EncounterService.GetEncounter(context.Background(), encounterID)
				if err != nil {
					content := fmt.Sprintf("❌ Failed to get encounter: %v", err)
					if responseErr := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: content,
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					}); responseErr != nil {
						log.Printf("Failed to respond with error message: %v", responseErr)
					}
					return
				}
				target := encounterResult.Combatants[targetID]

				// Show result
				embed := &discordgo.MessageEmbed{
					Title:       "💚 Healing Applied!",
					Description: fmt.Sprintf("**%s** is healed for %d points!", target.Name, healAmount),
					Color:       0x2ecc71, // Green
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:   "❤️ Hit Points",
							Value:  fmt.Sprintf("%d / %d", target.CurrentHP, target.MaxHP),
							Inline: true,
						},
					},
				}

				err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Embeds: []*discordgo.MessageEmbed{embed},
					},
				})
				if err != nil {
					log.Printf("Error responding to heal: %v", err)
				}
			case "execute_attack":
				// Extract values from modal
				targetName := ""
				attackRoll := 0

				for _, comp := range data.Components {
					if row, ok := comp.(*discordgo.ActionsRow); ok {
						for _, rowComp := range row.Components {
							if input, ok := rowComp.(*discordgo.TextInput); ok {
								switch input.CustomID {
								case "target_name":
									targetName = input.Value
								case "attack_roll":
									var attackErr error
									attackRoll, attackErr = strconv.Atoi(input.Value)
									if attackErr != nil {
										log.Printf("Error parsing attack roll: %v", attackErr)
										return
									}
								}
							}
						}
					}
				}

				// Validate attack roll
				if attackRoll < 1 || attackRoll > 20 {
					content := "❌ Invalid attack roll! Must be between 1-20."
					if responseErr := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: content,
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					}); responseErr != nil {
						log.Printf("Failed to respond with error message: %v", responseErr)
					}
					return
				}

				// Get encounter
				encounterResult, err := h.ServiceProvider.EncounterService.GetEncounter(context.Background(), encounterID)
				if err != nil {
					content := fmt.Sprintf("❌ Failed to get encounter: %v", err)
					if responseErr := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: content,
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					}); responseErr != nil {
						log.Printf("Failed to respond with error message: %v", responseErr)
					}
					return
				}

				// Get attacker
				attacker := encounterResult.GetCurrentCombatant()
				if attacker == nil {
					content := "❌ No active attacker!"
					if responseErr := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: content,
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					}); responseErr != nil {
						log.Printf("Failed to respond with error message: %v", responseErr)
					}
					return
				}

				// Find target
				var targetID string
				var target *entities.Combatant
				for id, combatant := range encounterResult.Combatants {
					if strings.EqualFold(combatant.Name, targetName) {
						targetID = id
						target = combatant
						break
					}
				}

				if target == nil {
					content := fmt.Sprintf("❌ Target '%s' not found!", targetName)
					if responseErr := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: content,
							Flags:   discordgo.MessageFlagsEphemeral,
						},
					}); responseErr != nil {
						log.Printf("Failed to respond with error message: %v", responseErr)
					}
					return
				}

				// Calculate attack bonus (simplified for testing)
				attackBonus := 5 // Default
				if attacker.Type == entities.CombatantTypeMonster && len(attacker.Actions) > 0 {
					attackBonus = attacker.Actions[0].AttackBonus
				}

				totalAttack := attackRoll + attackBonus
				hit := totalAttack >= target.AC

				// Build result
				embed := &discordgo.MessageEmbed{
					Title: "⚔️ Attack Result",
					Color: 0x3498db, // Blue
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:   "🎲 Attack Roll",
							Value:  fmt.Sprintf("d20: %d + %d = **%d**", attackRoll, attackBonus, totalAttack),
							Inline: true,
						},
						{
							Name:   "🎯 Target AC",
							Value:  fmt.Sprintf("%d", target.AC),
							Inline: true,
						},
					},
				}

				if hit {
					// Roll damage (simplified - using 1d8+3 for all attacks)
					damageRoll := 1 + (attackRoll % 8) // Simulate 1d8
					damageMod := 3
					totalDamage := damageRoll + damageMod

					embed.Description = fmt.Sprintf("**%s** hits **%s**!", attacker.Name, target.Name)
					embed.Color = 0x2ecc71 // Green
					embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
						Name:   "💥 Damage",
						Value:  fmt.Sprintf("1d8+3: %d + %d = **%d**", damageRoll, damageMod, totalDamage),
						Inline: false,
					})

					// Apply damage
					encounterErr := h.ServiceProvider.EncounterService.ApplyDamage(context.Background(), encounterID, targetID, i.Member.User.ID, totalDamage)
					if encounterErr != nil {
						log.Printf("Error applying damage: %v", encounterErr)
						return
					}

					// Get updated target
					encounterResult, encounterErr := h.ServiceProvider.EncounterService.GetEncounter(context.Background(), encounterID)
					if encounterErr != nil {
						log.Printf("Error getting encounter: %v", encounterErr)
						return
					}
					updatedTarget := encounterResult.Combatants[targetID]

					embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
						Name:   "❤️ Target HP",
						Value:  fmt.Sprintf("%d / %d", updatedTarget.CurrentHP, updatedTarget.MaxHP),
						Inline: true,
					})

					if updatedTarget.CurrentHP <= 0 {
						embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
							Name:   "💀 Status",
							Value:  "**Target defeated!**",
							Inline: true,
						})
					}
				} else {
					embed.Description = fmt.Sprintf("**%s** misses **%s**!", attacker.Name, target.Name)
					embed.Color = 0xe74c3c // Red
				}

				// Add next turn button
				components := []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.Button{
								Label:    "Next Turn",
								Style:    discordgo.PrimaryButton,
								CustomID: fmt.Sprintf("encounter:next_turn:%s", encounterID),
								Emoji:    &discordgo.ComponentEmoji{Name: "➡️"},
							},
							discordgo.Button{
								Label:    "View Status",
								Style:    discordgo.SecondaryButton,
								CustomID: fmt.Sprintf("encounter:view:%s", encounterID),
								Emoji:    &discordgo.ComponentEmoji{Name: "📊"},
							},
						},
					},
				}

				log.Printf("Sending attack response - Embed fields: %d, Components: %d", len(embed.Fields), len(components))
				err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseUpdateMessage,
					Data: &discordgo.InteractionResponseData{
						Embeds:     []*discordgo.MessageEmbed{embed},
						Components: components,
					},
				})
				if err != nil {
					log.Printf("Error responding to attack: %v", err)
				} else {
					log.Printf("Attack response sent successfully")
				}
			}
		}
	}
}

// respondWithUpdateError is a helper function to respond with an error message using UpdateMessage
func respondWithUpdateError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content:    fmt.Sprintf("❌ %s", message),
			Components: []discordgo.MessageComponent{},
		},
	}); err != nil {
		log.Printf("Failed to respond with error message: %v", err)
	}
}

// getWeaponPropertiesString converts weapon properties to a comma-separated string
func getWeaponPropertiesString(weapon *entities.Weapon) string {
	if len(weapon.Properties) == 0 {
		return "None"
	}

	var props []string
	for _, prop := range weapon.Properties {
		if prop != nil {
			props = append(props, prop.Name)
		}
	}

	if len(props) == 0 {
		return "None"
	}

	return strings.Join(props, ", ")
}
