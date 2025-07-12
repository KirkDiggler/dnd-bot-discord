package handlers

import (
	"log"

	"github.com/KirkDiggler/dnd-bot-discord/internal/discord/v2/core"
	charService "github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/bwmarrin/discordgo"
)

// HandleOpenNameModal opens a modal for character name entry
func (h *CharacterCreationHandler) HandleOpenNameModal(ctx *core.InteractionContext) (*core.HandlerResult, error) {
	// Parse custom ID to get character ID
	customID, err := core.ParseCustomID(ctx.GetCustomID())
	if err != nil {
		return nil, core.NewValidationError("Invalid selection")
	}

	characterID := customID.Target

	// Get character to verify ownership
	char, err := h.service.GetCharacter(ctx.Context, characterID)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Verify ownership
	if char.OwnerID != ctx.UserID {
		return nil, core.NewForbiddenError("You can only edit your own characters")
	}

	// Create modal custom ID
	modalCustomID := h.customIDBuilder.Modal("submit_name", characterID)

	// Show modal directly through Discord API
	err = ctx.Session.InteractionRespond(ctx.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: modalCustomID,
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
		},
	})

	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Modal responses don't need a response object
	return &core.HandlerResult{}, nil
}

// HandleSubmitName processes the submitted character name from the modal
func (h *CharacterCreationHandler) HandleSubmitName(ctx *core.InteractionContext) (*core.HandlerResult, error) {
	log.Printf("[HandleSubmitName] Starting - IsModal: %v, InteractionType: %v", ctx.IsModal(), ctx.Interaction.Type)

	// Defer the response immediately to prevent timeout
	responder := core.NewDiscordResponder(ctx.Session, ctx.Interaction)
	if err := responder.Defer(true); err != nil {
		log.Printf("[HandleSubmitName] Failed to defer response: %v", err)
		return nil, core.NewInternalError(err)
	}

	// Parse custom ID to get character ID
	customID, err := core.ParseCustomID(ctx.GetCustomID())
	if err != nil {
		log.Printf("[HandleSubmitName] Failed to parse custom ID: %v", err)
		return nil, core.NewValidationError("Invalid submission")
	}

	characterID := customID.Target
	log.Printf("[HandleSubmitName] Character ID: %s", characterID)

	// Get character
	char, err := h.service.GetCharacter(ctx.Context, characterID)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Verify ownership
	if char.OwnerID != ctx.UserID {
		return nil, core.NewForbiddenError("You can only edit your own characters")
	}

	// Get the submitted name from modal data
	var characterName string
	if ctx.IsModal() && ctx.Interaction != nil {
		data := ctx.Interaction.ModalSubmitData()
		log.Printf("[HandleSubmitName] Modal data custom ID: %s, components: %d", data.CustomID, len(data.Components))

		// Modal components are ActionsRows containing TextInputs
		for i, component := range data.Components {
			log.Printf("[HandleSubmitName] Component %d type: %T", i, component)
			if row, ok := component.(*discordgo.ActionsRow); ok {
				log.Printf("[HandleSubmitName] ActionsRow has %d components", len(row.Components))
				for j, rowComponent := range row.Components {
					log.Printf("[HandleSubmitName] Row component %d type: %T", j, rowComponent)
					if input, ok := rowComponent.(*discordgo.TextInput); ok {
						log.Printf("[HandleSubmitName] TextInput custom ID: %s, value: %s", input.CustomID, input.Value)
						if input.CustomID == "character_name" {
							characterName = input.Value
							break
						}
					}
				}
			}
		}
	}

	if characterName == "" {
		log.Printf("[HandleSubmitName] Character name is empty after parsing modal data")
		return nil, core.NewValidationError("Character name cannot be empty")
	}

	log.Printf("[HandleSubmitName] Received character name: %s for character ID: %s", characterName, characterID)

	// Update character name using UpdateDraftCharacter
	updateInput := &charService.UpdateDraftInput{
		Name: &characterName,
	}
	updatedChar, err := h.service.UpdateDraftCharacter(ctx.Context, char.ID, updateInput)
	if err != nil {
		return nil, core.NewInternalError(err)
	}

	// Complete creation and get the response
	result, err := h.completeCreation(ctx, updatedChar)
	if err != nil {
		// Edit the deferred response with an error message
		errorResponse := core.NewEphemeralResponse("Failed to finalize character: " + err.Error())
		editErr := responder.Edit(errorResponse)
		if editErr != nil {
			log.Printf("[HandleSubmitName] Failed to edit response with error: %v", editErr)
		}
		return nil, err
	}

	// Edit the deferred response with the completion result
	if result != nil && result.Response != nil {
		if err := responder.Edit(result.Response); err != nil {
			log.Printf("[HandleSubmitName] Failed to edit response: %v", err)
			return nil, core.NewInternalError(err)
		}
	}

	// Return empty result since we've already handled the response
	return &core.HandlerResult{}, nil
}
