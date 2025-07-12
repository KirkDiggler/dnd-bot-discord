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
	r.router.ComponentFunc("assign", r.creationHandler.HandleStepSelection)
	r.router.ComponentFunc("proficiencies", r.creationHandler.HandleOpenProficiencySelection)
	r.router.ComponentFunc("equipment", r.creationHandler.HandleOpenEquipmentSelection)
	r.router.ComponentFunc("name", r.creationHandler.HandleOpenNameModal)
	r.router.ModalFunc("submit_name", r.creationHandler.HandleSubmitName)
	r.router.ComponentFunc("option", r.creationHandler.HandleStepSelection)

	// Enhanced UI handlers
	r.router.ComponentFunc("race_list", r.creationHandler.HandleRaceDetails)
	r.router.ComponentFunc("race_random", r.creationHandler.HandleRandomRace)
	r.router.ComponentFunc("preview_race", r.creationHandler.HandleRacePreview)
	r.router.ComponentFunc("confirm_race", r.creationHandler.HandleConfirmRace)

	// Class selection handlers
	r.router.ComponentFunc("class_overview", r.creationHandler.HandleClassOverview)
	r.router.ComponentFunc("class_random", r.creationHandler.HandleRandomClass)
	r.router.ComponentFunc("preview_class", r.creationHandler.HandleClassPreview)
	r.router.ComponentFunc("confirm_class", r.creationHandler.HandleConfirmClass)

	// Ability scores handlers
	r.router.ComponentFunc("roll", r.creationHandler.HandleRollAbilityScores)
	r.router.ComponentFunc("standard", r.creationHandler.HandleStandardArray)
	r.router.ComponentFunc("ability_info", r.creationHandler.HandleAbilityInfo)
	r.router.ComponentFunc("confirm_rolled_scores", r.creationHandler.HandleConfirmRolledScores)
	r.router.ComponentFunc("assign_standard_array", r.creationHandler.HandleAssignStandardArray)
	r.router.ComponentFunc("start_ability_assignment", r.creationHandler.HandleStartAbilityAssignment)
	r.router.ComponentFunc("start_fresh", r.creationHandler.HandleStartFresh)
	r.router.ComponentFunc("back_to_abilities", r.creationHandler.HandleBackToAbilities)

	// Manual assignment handlers
	r.router.ComponentFunc("auto_assign_and_continue", r.creationHandler.HandleAutoAssignAndContinue)
	r.router.ComponentFunc("start_manual_assignment", r.creationHandler.HandleStartManualAssignment)
	r.router.ComponentFunc("assign_to_ability", r.creationHandler.HandleAssignToAbility)
	r.router.ComponentFunc("apply_ability_assignment", r.creationHandler.HandleApplyAbilityAssignment)
	r.router.ComponentFunc("apply_direct_assignment", r.creationHandler.HandleApplyDirectAssignment)
	r.router.ComponentFunc("confirm_ability_assignment", r.creationHandler.HandleConfirmAbilityAssignment)

	// One-at-a-time rolling handlers
	r.router.ComponentFunc("roll_single_ability", r.creationHandler.HandleRollSingleAbility)
	r.router.ComponentFunc("continue_rolling", r.creationHandler.HandleContinueRolling)
	r.router.ComponentFunc("start_interactive_assignment", r.creationHandler.HandleStartInteractiveAssignment)

	// Wizard-specific handlers (registered for UI hints actions)
	r.router.ComponentFunc("select_cantrips", r.creationHandler.HandleStepSelection)
	r.router.ComponentFunc("suggested_cantrips", r.creationHandler.HandleStepSelection)
	r.router.ComponentFunc("select_spells", r.creationHandler.HandleStepSelection)
	r.router.ComponentFunc("quick_build", r.creationHandler.HandleStepSelection)
	r.router.ComponentFunc("select_tradition", r.creationHandler.HandleStepSelection)

	// Spell selection handlers (paginated)
	r.router.ComponentFunc("open_spell_selection", r.creationHandler.HandleOpenSpellSelection)
	r.router.ComponentFunc("spell_page", r.creationHandler.HandleSpellPageChange)
	r.router.ComponentFunc("select_spell", r.creationHandler.HandleSpellToggle)
	r.router.ComponentFunc("confirm_spell_selection", r.creationHandler.HandleConfirmSpellSelection)
	r.router.ComponentFunc("cancel_spell_selection", r.creationHandler.HandleCancelSpellSelection)
	r.router.ComponentFunc("spell_details", r.creationHandler.HandleSpellDetails)
	r.router.ComponentFunc("continue", r.creationHandler.HandleStepSelection)

	// Proficiency and Equipment selection handlers
	r.router.ComponentFunc("select_proficiency", r.creationHandler.HandleSelectProficiency)
	r.router.ComponentFunc("confirm_proficiency_selection", r.creationHandler.HandleConfirmProficiencySelection)
	r.router.ComponentFunc("select_equipment", r.creationHandler.HandleSelectEquipment)
	r.router.ComponentFunc("confirm_equipment_selection", r.creationHandler.HandleConfirmEquipmentSelection)

	// Future: /dnd create encounter, /dnd create item, etc.
	// r.router.ActionFunc("encounter", r.handleCreateEncounter)
	// r.router.ActionFunc("item", r.handleCreateItem)
}
