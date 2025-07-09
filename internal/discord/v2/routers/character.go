package routers

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/discord/v2/builders"
	"github.com/KirkDiggler/dnd-bot-discord/internal/discord/v2/core"
	"github.com/KirkDiggler/dnd-bot-discord/internal/discord/v2/handlers"
	"github.com/KirkDiggler/dnd-bot-discord/internal/discord/v2/middleware"
	domainCharacter "github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
)

// CharacterRouter handles all character-related interactions
type CharacterRouter struct {
	router          *core.Router
	service         character.Service
	flowService     domainCharacter.CreationFlowService
	creationHandler *handlers.CharacterCreationHandler
	idBuilder       *core.CustomIDBuilder
}

type CharacterRouterConfig struct {
	Pipeline *core.Pipeline
	Provider *services.Provider
}

func (cr *CharacterRouterConfig) Validate() error {
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

// NewCharacterRouter creates a new character router
func NewCharacterRouter(cfg *CharacterRouterConfig) (*CharacterRouter, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	router := core.NewRouter("character", cfg.Pipeline)

	cr := &CharacterRouter{
		router:      router,
		service:     cfg.Provider.CharacterService,
		flowService: cfg.Provider.CreationFlowService,
		idBuilder:   router.GetCustomIDBuilder(),
	}

	// Create character creation handler
	creationConfig := &handlers.CharacterCreationHandlerConfig{
		Service:         cfg.Provider.CharacterService,
		FlowService:     cfg.Provider.CreationFlowService,
		CustomIDBuilder: router.GetCustomIDBuilder(),
	}
	creationHandler, err := handlers.NewCharacterCreationHandler(creationConfig)
	if err != nil {
		// Log error but continue - creation won't work but other features will
		log.Printf("Failed to create character creation handler: %v", err)
		return nil, err
	}
	cr.creationHandler = creationHandler

	// Apply router-specific middleware
	router.Use(
		middleware.UserRateLimitMiddleware(10, 60), // 10 requests per minute
	)

	// Register routes
	cr.registerRoutes()

	// Register with pipeline
	router.Register()

	return cr, nil
}

// registerRoutes sets up all character routes
func (r *CharacterRouter) registerRoutes() {
	// Slash command actions (domain is "character")
	r.router.ActionFunc("list", r.handleList)     // /dnd character list
	r.router.ActionFunc("show", r.handleShow)     // /dnd character show
	r.router.ActionFunc("create", r.handleCreate) // /dnd character create

	// Component interactions
	r.router.ComponentFunc("quickshow", r.handleQuickShow)
	r.router.ComponentFunc("delete", r.handleDelete)
	r.router.ComponentFunc("delete_confirm", r.handleDeleteConfirm)
	r.router.ComponentFunc("delete_cancel", r.handleDeleteCancel)
	r.router.ComponentFunc("page", r.handlePageChange)

	// Note: Character creation components are handled by CharacterCreationRouter
	// This router only handles the initial slash command
}

// handleList handles the character list command
func (r *CharacterRouter) handleList(ctx *core.InteractionContext) (*core.HandlerResult, error) {
	// Get user's characters
	characters, err := r.service.ListCharacters(ctx.Context, ctx.UserID)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Filter out empty drafts
	filtered := make([]*domainCharacter.Character, 0)
	for _, char := range characters {
		if char.Status == "active" || (char.Name != "" && char.Race != nil && char.Class != nil) {
			filtered = append(filtered, char)
		}
	}

	// Build embed
	embed := builders.NewListEmbed("Your Characters", 10).
		SetTotalItems(len(filtered))

	if len(filtered) == 0 {
		embed.Description("You don't have any characters yet. Use `/dnd character create` to create one!")
	} else {
		for i, char := range filtered {
			if i >= 10 {
				break // First page only
			}

			status := "🟢 Active"
			if char.Status == "draft" {
				status = "📝 Draft"
			}

			value := fmt.Sprintf("Level %d %s %s • %s",
				char.Level,
				char.Race.Name,
				char.Class.Name,
				status,
			)
			embed.AddItem(char.Name, value)
		}
	}

	// Build components
	components := builders.NewComponentBuilder(r.idBuilder)

	// Add character buttons
	for i, char := range filtered {
		if i >= 5 {
			break // Discord limit
		}
		components.PrimaryButton(char.Name, "quickshow", char.ID)
	}

	// Add pagination if needed
	if len(filtered) > 10 {
		components.NewRow()
		totalPages := (len(filtered) + 9) / 10
		components.PaginationButtons(1, totalPages, "page")
	}

	// Create response
	response := core.NewResponse("").
		WithEmbeds(embed.Build()).
		WithComponents(components.Build()...).
		AsEphemeral()

	return &core.HandlerResult{
		Response: response,
	}, nil
}

// handleShow handles showing a specific character
func (r *CharacterRouter) handleShow(ctx *core.InteractionContext) (*core.HandlerResult, error) {
	// Get character ID from options
	charID := ctx.GetStringParam("character")
	if charID == "" {
		return nil, core.NewValidationError("Please select a character to show")
	}

	// Get character
	char, err := r.service.GetCharacter(ctx.Context, charID)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Check ownership
	if char.OwnerID != ctx.UserID {
		return nil, core.NewForbiddenError("You can only view your own characters")
	}

	// Build character embed
	embedBuilder := r.buildCharacterEmbed(char)

	// Build components
	components := builders.NewComponentBuilder(r.idBuilder).
		ActionButtons(builders.ActionButtonsConfig{
			ShowEdit:   char.Status == "draft",
			ShowDelete: true,
			ShowInfo:   true,
			TargetID:   char.ID,
		})

	// Create response
	response := core.NewResponse("").
		WithEmbeds(embedBuilder.Build()).
		WithComponents(components.Build()...).
		AsEphemeral()

	return &core.HandlerResult{
		Response: response,
	}, nil
}

// handleQuickShow handles the quick show button
func (r *CharacterRouter) handleQuickShow(ctx *core.InteractionContext) (*core.HandlerResult, error) {
	// Get character ID from custom ID
	customID, err := core.ParseCustomID(ctx.GetCustomID())
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	charID := customID.Target

	// Get character
	char, err := r.service.GetCharacter(ctx.Context, charID)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Check ownership
	if char.OwnerID != ctx.UserID {
		return nil, core.NewForbiddenError("You can only view your own characters")
	}

	// Build character embed
	embedBuilder := r.buildCharacterEmbed(char)

	// Create response (update original message)
	response := core.NewResponse("").
		WithEmbeds(embedBuilder.Build()).
		AsUpdate()

	return &core.HandlerResult{
		Response: response,
	}, nil
}

// handleCreate starts character creation
func (r *CharacterRouter) handleCreate(ctx *core.InteractionContext) (*core.HandlerResult, error) {
	// Use the character creation handler if available
	if r.creationHandler != nil {
		return r.creationHandler.StartCreation(ctx)
	}

	// Fallback if handler not available
	return nil, core.NewInternalError(fmt.Errorf("character creation handler not available"))
}

// handleDelete shows delete confirmation
func (r *CharacterRouter) handleDelete(ctx *core.InteractionContext) (*core.HandlerResult, error) {
	customID, err := core.ParseCustomID(ctx.GetCustomID())
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	charID := customID.Target

	// Get character to show name
	char, err := r.service.GetCharacter(ctx.Context, charID)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	if char.OwnerID != ctx.UserID {
		return nil, core.NewForbiddenError("You can only delete your own characters")
	}

	// Confirmation embed
	embed := builders.WarningEmbed(
		"Delete Character?",
		fmt.Sprintf("Are you sure you want to delete **%s**? This action cannot be undone.", char.Name),
	).Build()

	// Confirmation buttons
	components := builders.NewComponentBuilder(r.idBuilder).
		ConfirmationButtons("delete_confirm", "delete_cancel", charID)

	response := core.NewResponse("").
		WithEmbeds(embed).
		WithComponents(components.Build()...).
		AsUpdate()

	return &core.HandlerResult{
		Response: response,
	}, nil
}

// handleDeleteConfirm processes character deletion
func (r *CharacterRouter) handleDeleteConfirm(ctx *core.InteractionContext) (*core.HandlerResult, error) {
	customID, err := core.ParseCustomID(ctx.GetCustomID())
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	charID := customID.Target

	// Delete character
	err = r.service.Delete(charID)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Success message
	embed := builders.SuccessEmbed(
		"Character Deleted",
		"The character has been successfully deleted.",
	).Build()

	response := core.NewResponse("").
		WithEmbeds(embed).
		AsUpdate()

	return &core.HandlerResult{
		Response: response,
	}, nil
}

// handleDeleteCancel cancels deletion
func (r *CharacterRouter) handleDeleteCancel(ctx *core.InteractionContext) (*core.HandlerResult, error) {
	embed := builders.InfoEmbed(
		"Deletion Cancelled",
		"The character was not deleted.",
	).Build()

	response := core.NewResponse("").
		WithEmbeds(embed).
		AsUpdate()

	return &core.HandlerResult{
		Response: response,
	}, nil
}

// handlePageChange handles pagination
func (r *CharacterRouter) handlePageChange(ctx *core.InteractionContext) (*core.HandlerResult, error) {
	customID, err := core.ParseCustomID(ctx.GetCustomID())
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Extract page number from args
	if len(customID.Args) == 0 {
		return nil, core.NewValidationError("Invalid page number")
	}

	page, err := strconv.Atoi(customID.Args[0])
	if err != nil {
		return nil, core.NewValidationError("Invalid page number")
	}

	// TODO: Implement pagination support in handleList
	// For now, just show the list (which defaults to page 1)
	_ = page

	return r.handleList(ctx)
}

// buildCharacterEmbed creates a character display embed
func (r *CharacterRouter) buildCharacterEmbed(char *domainCharacter.Character) *builders.EmbedBuilder {
	embed := builders.NewCharacterEmbed().
		SetCharacter(char.Name, char.Race.Name, char.Class.Name, char.Level).
		AddStats(
			char.Attributes[shared.AttributeStrength].Score,
			char.Attributes[shared.AttributeDexterity].Score,
			char.Attributes[shared.AttributeConstitution].Score,
			char.Attributes[shared.AttributeIntelligence].Score,
			char.Attributes[shared.AttributeWisdom].Score,
			char.Attributes[shared.AttributeCharisma].Score,
		).
		AddCombatInfo(char.CurrentHitPoints, char.MaxHitPoints, char.AC)

	// Add proficiencies grouped by type
	if len(char.Proficiencies) > 0 {
		profsByType := make(map[string][]string)
		for profType, profList := range char.Proficiencies {
			typeName := string(profType)
			for _, prof := range profList {
				profsByType[typeName] = append(profsByType[typeName], prof.Name)
			}
		}

		// Add each proficiency type as a field
		for profType, profs := range profsByType {
			embed.Field(profType+" Proficiencies", strings.Join(profs, ", "), false)
		}
	}

	// Add equipped items
	if len(char.EquippedSlots) > 0 {
		for slot, item := range char.EquippedSlots {
			embed.Field(string(slot), item.GetName(), true)
		}
	}

	return embed.EmbedBuilder
}
