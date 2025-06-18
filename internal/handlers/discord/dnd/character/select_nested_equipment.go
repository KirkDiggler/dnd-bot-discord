package character

import (
	"context"
	"fmt"

	characterService "github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/bwmarrin/discordgo"
)

// SelectNestedEquipmentHandler handles nested equipment selection (e.g., choosing martial weapons)
type SelectNestedEquipmentHandler struct {
	characterService characterService.Service
}

// SelectNestedEquipmentHandlerConfig holds configuration
type SelectNestedEquipmentHandlerConfig struct {
	CharacterService characterService.Service
}

// NewSelectNestedEquipmentHandler creates a new handler
func NewSelectNestedEquipmentHandler(cfg *SelectNestedEquipmentHandlerConfig) *SelectNestedEquipmentHandler {
	return &SelectNestedEquipmentHandler{
		characterService: cfg.CharacterService,
	}
}

// SelectNestedEquipmentRequest represents the request
type SelectNestedEquipmentRequest struct {
	Session        *discordgo.Session
	Interaction    *discordgo.InteractionCreate
	RaceKey        string
	ClassKey       string
	ChoiceIndex    int
	BundleKey      string // e.g., "nested-0"
	SelectionCount int    // How many to select (1 or 2)
	Category       string // e.g., "martial-weapons"
}

// Handle processes nested equipment selection
func (h *SelectNestedEquipmentHandler) Handle(req *SelectNestedEquipmentRequest) error {
	// Update the message first
	err := req.Session.InteractionRespond(req.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content: "Loading weapon options...",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to acknowledge interaction: %w", err)
	}

	// Get equipment from category
	equipment, err := h.characterService.GetEquipmentByCategory(context.Background(), req.Category)
	if err != nil {
		return h.respondWithError(req, fmt.Sprintf("Failed to fetch %s.", req.Category))
	}

	// Get race and class for context
	race, err := h.characterService.GetRace(context.Background(), req.RaceKey)
	if err != nil {
		return h.respondWithError(req, "Failed to fetch race details.")
	}

	class, err := h.characterService.GetClass(context.Background(), req.ClassKey)
	if err != nil {
		return h.respondWithError(req, "Failed to fetch class details.")
	}

	// Create embed
	var title string
	if req.SelectionCount > 1 {
		title = fmt.Sprintf("Select %d Weapons", req.SelectionCount)
	} else {
		title = "Select Weapon"
	}

	embed := &discordgo.MessageEmbed{
		Title:       title,
		Description: fmt.Sprintf("**Race:** %s\n**Class:** %s\n\nChoose from the available weapons.", race.Name, class.Name),
		Color:       0x5865F2,
		Fields:      []*discordgo.MessageEmbedField{},
	}

	// Create select menu options
	var selectOptions []discordgo.SelectMenuOption
	for _, equip := range equipment {
		if equip == nil {
			continue
		}

		option := discordgo.SelectMenuOption{
			Label: equip.GetName(),
			Value: equip.GetKey(),
		}

		// Add description based on equipment type
		desc := h.getWeaponDescription(equip)
		if desc != "" {
			option.Description = desc
		}

		selectOptions = append(selectOptions, option)
	}

	// Limit to 25 options (Discord limitation)
	if len(selectOptions) > 25 {
		selectOptions = selectOptions[:25]
		embed.Footer = &discordgo.MessageEmbedFooter{
			Text: "Showing first 25 weapons. Common choices include: Longsword, Battleaxe, Warhammer, Rapier",
		}
	}

	// Progress
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "Progress",
		Value:  "✅ Step 1: Race\n✅ Step 2: Class\n✅ Step 3: Abilities\n✅ Step 4: Proficiencies\n⏳ Step 5: Equipment\n⏳ Step 6: Details",
		Inline: false,
	})

	// Create components
	minValues := 1
	maxValues := req.SelectionCount
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    fmt.Sprintf("character_create:confirm_nested_equipment:%s:%s:%d:%s", req.RaceKey, req.ClassKey, req.ChoiceIndex, req.BundleKey),
					Placeholder: fmt.Sprintf("Select %d weapon(s)", req.SelectionCount),
					MinValues:   &minValues,
					MaxValues:   maxValues,
					Options:     selectOptions,
				},
			},
		},
	}

	// Update message
	_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
		Content:    &[]string{""}[0],
	})

	return err
}

// getWeaponDescription returns a formatted description for a weapon
func (h *SelectNestedEquipmentHandler) getWeaponDescription(equip interface{}) string {
	// Check if it's a weapon type with damage info
	// This would need to be implemented based on your equipment entity structure
	// For now, return a simple placeholder
	switch e := equip.(type) {
	case interface{ GetDamage() string }:
		return e.GetDamage()
	default:
		return ""
	}
}

// respondWithError updates the message with an error
func (h *SelectNestedEquipmentHandler) respondWithError(req *SelectNestedEquipmentRequest, message string) error {
	content := fmt.Sprintf("❌ %s", message)
	_, err := req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
		Content: &content,
		Embeds:  &[]*discordgo.MessageEmbed{},
	})
	return err
}
