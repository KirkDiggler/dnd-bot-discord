package combat

import (
	"fmt"
	"log"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/encounter"
	"github.com/bwmarrin/discordgo"
)

// SavingThrowHandler handles saving throw interactions
type SavingThrowHandler struct {
	characterService character.Service
	encounterService encounter.Service
}

// SavingThrowHandlerConfig holds configuration for the saving throw handler
type SavingThrowHandlerConfig struct {
	CharacterService character.Service
	EncounterService encounter.Service
}

// NewSavingThrowHandler creates a new saving throw handler
func NewSavingThrowHandler(cfg *SavingThrowHandlerConfig) *SavingThrowHandler {
	return &SavingThrowHandler{
		characterService: cfg.CharacterService,
		encounterService: cfg.EncounterService,
	}
}

// ShowSavingThrowPrompt displays a prompt for a player to make a saving throw
func (h *SavingThrowHandler) ShowSavingThrowPrompt(s *discordgo.Session, i *discordgo.InteractionCreate,
	character *entities.Character, attribute entities.Attribute, dc int, reason string) error {

	// Calculate the bonus
	bonus := character.GetSavingThrowBonus(attribute)
	profIndicator := ""
	if character.HasSavingThrowProficiency(attribute) {
		profIndicator = " (PROF)"
	}

	embed := &discordgo.MessageEmbed{
		Title:       "🎲 Saving Throw Required!",
		Description: fmt.Sprintf("%s must make a **%s saving throw**!\n\n%s", character.Name, strings.ToUpper(string(attribute)), reason),
		Color:       0xe74c3c, // Red
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "DC",
				Value:  fmt.Sprintf("%d", dc),
				Inline: true,
			},
			{
				Name:   "Your Bonus",
				Value:  fmt.Sprintf("%+d%s", bonus, profIndicator),
				Inline: true,
			},
			{
				Name:   "Need to Roll",
				Value:  fmt.Sprintf("%d or higher", dc-bonus),
				Inline: true,
			},
		},
	}

	// Create roll button
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    fmt.Sprintf("Roll %s Save", strings.ToUpper(string(attribute))),
					Style:    discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("saving_throw:%s:%s:%d", character.ID, attribute, dc),
					Emoji: &discordgo.ComponentEmoji{
						Name: "🎲",
					},
				},
			},
		},
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
		},
	})
}

// HandleSavingThrowRoll processes a saving throw roll
func (h *SavingThrowHandler) HandleSavingThrowRoll(s *discordgo.Session, i *discordgo.InteractionCreate,
	characterID string, attribute entities.Attribute, dc int) error {

	// Get the character
	char, err := h.characterService.GetByID(characterID)
	if err != nil {
		log.Printf("Failed to get character %s: %v", characterID, err)
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Failed to find your character!",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// Roll the saving throw
	roll, total, err := char.RollSavingThrow(attribute)
	if err != nil {
		log.Printf("Failed to roll saving throw: %v", err)
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Failed to roll saving throw!",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// Determine success/failure
	success := total >= dc
	resultText := "❌ **FAILED!**"
	resultColor := 0xe74c3c // Red
	if success {
		resultText = "✅ **SUCCESS!**"
		resultColor = 0x2ecc71 // Green
	}

	bonus := char.GetSavingThrowBonus(attribute)
	profIndicator := ""
	if char.HasSavingThrowProficiency(attribute) {
		profIndicator = " (PROF)"
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s Saving Throw", strings.ToUpper(string(attribute))),
		Description: fmt.Sprintf("%s rolls a saving throw!", char.Name),
		Color:       resultColor,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Roll",
				Value:  fmt.Sprintf("🎲 %d", roll.Rolls[0]),
				Inline: true,
			},
			{
				Name:   "Bonus",
				Value:  fmt.Sprintf("%+d%s", bonus, profIndicator),
				Inline: true,
			},
			{
				Name:   "Total",
				Value:  fmt.Sprintf("**%d**", total),
				Inline: true,
			},
			{
				Name:   "DC",
				Value:  fmt.Sprintf("%d", dc),
				Inline: true,
			},
			{
				Name:   "Result",
				Value:  resultText,
				Inline: true,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Proficiency Bonus: +%d", char.GetProficiencyBonus()),
		},
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: []discordgo.MessageComponent{}, // Remove buttons
		},
	})
}
