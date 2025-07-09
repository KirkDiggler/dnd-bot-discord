package routers

import (
	"errors"

	"github.com/KirkDiggler/dnd-bot-discord/internal/discord/v2/core"
	"github.com/KirkDiggler/dnd-bot-discord/internal/discord/v2/handlers"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services"
)

// CreateRouter handles /dnd create commands (character, encounter, item, etc.)
type CreateRouter struct {
	router          *core.Router
	creationHandler *handlers.CharacterCreationHandler
}

type CreateRouterConfig struct {
	Pipeline *core.Pipeline
	Provider *services.Provider
}

func (cr *CreateRouterConfig) Validate() error {
	if cr.Pipeline == nil {
		return errors.New("pipeline is required")
	}
	if cr.Provider == nil {
		return errors.New("provider is required")
	}

	if cr.Provider.CharacterService == nil {
		return errors.New("provider.CharacterService is required")
	}
	if cr.Provider.CreationFlowService == nil {
		return errors.New("provider.CreationFlowService is required")
	}

	return nil
}

// NewCreateRouter creates a router for /dnd create commands
func NewCreateRouter(cfg *CreateRouterConfig) (*CreateRouter, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// Use "create" domain for /dnd create commands
	router := core.NewRouter("create", cfg.Pipeline)

	// Create the handler with the creation domain
	creationConfig := &handlers.CharacterCreationHandlerConfig{
		Service:         cfg.Provider.CharacterService,
		FlowService:     cfg.Provider.CreationFlowService,
		CustomIDBuilder: router.GetCustomIDBuilder(),
	}

	creationHandler, err := handlers.NewCharacterCreationHandler(creationConfig)
	if err != nil {
		// This is a critical error - character creation won't work
		return nil, err
	}

	cr := &CreateRouter{
		router:          router,
		creationHandler: creationHandler,
	}

	// Register all creation flow routes
	cr.registerRoutes()

	// Register with pipeline
	router.Register()

	return cr, nil
}

// registerRoutes sets up all create command routes
func (r *CreateRouter) registerRoutes() {
	// Command: /dnd create character
	r.router.ActionFunc("character", r.creationHandler.StartCreation)

	// Core flow components
	r.router.ComponentFunc("select", r.creationHandler.HandleStepSelection)
	r.router.ComponentFunc("back", r.creationHandler.HandleStepSelection)
	r.router.ComponentFunc("roll", r.creationHandler.HandleStepSelection)
	r.router.ComponentFunc("assign", r.creationHandler.HandleStepSelection)
	r.router.ComponentFunc("proficiencies", r.creationHandler.HandleStepSelection)
	r.router.ComponentFunc("equipment", r.creationHandler.HandleStepSelection)
	r.router.ComponentFunc("name", r.creationHandler.HandleStepSelection)
	r.router.ComponentFunc("option", r.creationHandler.HandleStepSelection)

	// Enhanced UI handlers
	r.router.ComponentFunc("race_list", r.creationHandler.HandleRaceDetails)
	r.router.ComponentFunc("race_random", r.creationHandler.HandleRandomRace)
	r.router.ComponentFunc("preview_race", r.creationHandler.HandleRacePreview)
	r.router.ComponentFunc("confirm_race", r.creationHandler.HandleConfirmRace)

	// Future: /dnd create encounter, /dnd create item, etc.
	// r.router.ActionFunc("encounter", r.handleCreateEncounter)
	// r.router.ActionFunc("item", r.handleCreateItem)
}
