package routers

import (
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

// NewCharacterRouter creates a new character router
func NewCharacterRouter(pipeline *core.Pipeline, provider *services.Provider) *CharacterRouter {
	router := core.NewRouter("character", pipeline)

	cr := &CharacterRouter{
		router:      router,
		service:     provider.CharacterService,
		flowService: provider.CreationFlowService,
		idBuilder:   router.GetCustomIDBuilder(),
	}

	// Create character creation handler
	creationConfig := &handlers.CharacterCreationHandlerConfig{
		Service:     provider.CharacterService,
		FlowService: provider.CreationFlowService,
	}
	creationHandler, err := handlers.NewCharacterCreationHandler(creationConfig)
	if err != nil {
		// Log error but continue - creation won't work but other features will
		log.Printf("Failed to create character creation handler: %v", err)
	} else {
		cr.creationHandler = creationHandler
	}

	// Apply router-specific middleware
	router.Use(
		middleware.UserRateLimitMiddleware(10, 60), // 10 requests per minute
	)

	// Register routes
	cr.registerRoutes()

	// Register with pipeline
	router.Register()

	return cr
}

// registerRoutes sets up all character routes
func (r *CharacterRouter) registerRoutes() {
	// Slash commands
	r.router.SubcommandFunc("dnd", "character list", r.handleList)
	r.router.SubcommandFunc("dnd", "character show", r.handleShow)
	r.router.SubcommandFunc("dnd", "character create", r.handleCreate)

	// Component interactions
	r.router.ComponentFunc("quickshow", r.handleQuickShow)
	r.router.ComponentFunc("delete", r.handleDelete)
	r.router.ComponentFunc("delete_confirm", r.handleDeleteConfirm)
	r.router.ComponentFunc("delete_cancel", r.handleDeleteCancel)
	r.router.ComponentFunc("page", r.handlePageChange)

	// Character creation flow components (if handler is available)
	if r.creationHandler != nil {
		// Selection components - using wildcards to match any creation action
		r.router.ComponentFunc("creation:select", r.creationHandler.HandleStepSelection)
		r.router.ComponentFunc("creation:back", r.creationHandler.HandleStepSelection)
		r.router.ComponentFunc("creation:roll", r.creationHandler.HandleStepSelection)
		r.router.ComponentFunc("creation:assign", r.creationHandler.HandleStepSelection)
		r.router.ComponentFunc("creation:proficiencies", r.creationHandler.HandleStepSelection)
		r.router.ComponentFunc("creation:equipment", r.creationHandler.HandleStepSelection)
		r.router.ComponentFunc("creation:name", r.creationHandler.HandleStepSelection)
		r.router.ComponentFunc("creation:option", r.creationHandler.HandleStepSelection)

		// Register additional enhanced handlers
		r.creationHandler.RegisterAdditionalHandlers(r.router)
	}
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
