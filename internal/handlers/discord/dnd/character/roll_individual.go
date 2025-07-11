package character

import (
	"context"
	"fmt"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"math/rand"
	"sort"
	"strings"
	"time"

	characterService "github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/bwmarrin/discordgo"
)

// RollIndividualHandler handles rolling ability scores one at a time
type RollIndividualHandler struct {
	characterService characterService.Service
}

// RollIndividualHandlerConfig holds configuration
type RollIndividualHandlerConfig struct {
	CharacterService characterService.Service
}

// NewRollIndividualHandler creates a new handler
func NewRollIndividualHandler(cfg *RollIndividualHandlerConfig) *RollIndividualHandler {
	return &RollIndividualHandler{
		characterService: cfg.CharacterService,
	}
}

// RollIndividualRequest represents the request
type RollIndividualRequest struct {
	Session     *discordgo.Session
	Interaction *discordgo.InteractionCreate
	RaceKey     string
	ClassKey    string
	RollIndex   int // Which roll we're on (0-5)
}

// Handle processes rolling individual ability scores
func (h *RollIndividualHandler) Handle(req *RollIndividualRequest) error {
	// Update the message
	err := req.Session.InteractionRespond(req.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content: "Rolling ability score...",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to acknowledge interaction: %w", err)
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

	// Get or create draft character
	draftChar, err := h.characterService.GetOrCreateDraftCharacter(
		context.Background(),
		req.Interaction.Member.User.ID,
		req.Interaction.GuildID,
	)
	if err != nil {
		return h.respondWithError(req, "Failed to get character draft.")
	}

	// Get existing rolls or initialize empty
	existingRolls := draftChar.AbilityRolls
	if existingRolls == nil {
		existingRolls = []character.AbilityRoll{}
	}

	// If this is a new roll (not viewing existing), roll the dice
	var currentRoll character.AbilityRoll
	var dice []int
	if req.RollIndex == len(existingRolls) {
		// Roll 4d6
		dice = make([]int, 4)
		for j := 0; j < 4; j++ {
			dice[j] = rand.Intn(6) + 1
		}

		// Sort to identify lowest
		sortedDice := make([]int, 4)
		copy(sortedDice, dice)
		sort.Ints(sortedDice)

		// Calculate total (drop lowest)
		total := 0
		for j := 1; j < 4; j++ {
			total += sortedDice[j]
		}

		currentRoll = character.AbilityRoll{
			ID:    fmt.Sprintf("roll_%d_%d", time.Now().UnixNano(), req.RollIndex),
			Value: total,
		}

		// Add to existing rolls
		existingRolls = append(existingRolls, currentRoll)

		// Save updated rolls
		_, err = h.characterService.UpdateDraftCharacter(
			context.Background(),
			draftChar.ID,
			&characterService.UpdateDraftInput{
				AbilityRolls: existingRolls,
			},
		)
		if err != nil {
			return h.respondWithError(req, "Failed to save ability roll.")
		}
	} else if req.RollIndex < len(existingRolls) {
		// Viewing an existing roll
		currentRoll = existingRolls[req.RollIndex]
	}

	// Create embed
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Rolling Ability Score %d of 6", req.RollIndex+1),
		Description: fmt.Sprintf("**Race:** %s\n**Class:** %s", race.Name, class.Name),
		Color:       0x5865F2,
		Fields:      []*discordgo.MessageEmbedField{},
	}

	// Show dice details for new roll
	if dice != nil {
		diceStr := []string{}
		for _, d := range dice {
			diceStr = append(diceStr, fmt.Sprintf("%d", d))
		}

		sortedDice := make([]int, 4)
		copy(sortedDice, dice)
		sort.Ints(sortedDice)

		// Add flavor text based on roll quality
		rollFlavor := ""
		if currentRoll.Value >= 16 {
			rollFlavor = "\n*Exceptional!*"
		} else if currentRoll.Value >= 14 {
			rollFlavor = "\n*A strong roll!*"
		} else if currentRoll.Value <= 8 {
			rollFlavor = "\n*The dice are cruel...*"
		} else if currentRoll.Value <= 10 {
			rollFlavor = "\n*Below average, but workable.*"
		}

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "🎲 Dice Rolled",
			Value:  fmt.Sprintf("**Rolled:** %s\n**Dropped:** %d (lowest)\n**Total:** %d%s", strings.Join(diceStr, ", "), sortedDice[0], currentRoll.Value, rollFlavor),
			Inline: false,
		})
	}

	// Show all rolls so far
	if len(existingRolls) > 0 {
		rollStrings := []string{}
		for i, roll := range existingRolls {
			rollStrings = append(rollStrings, fmt.Sprintf("**Roll %d:** %d", i+1, roll.Value))
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "📊 Rolls So Far",
			Value:  strings.Join(rollStrings, "\n"),
			Inline: true,
		})
	}

	// Show progress
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "Progress",
		Value:  fmt.Sprintf("🎲 Rolling: %d/6 completed", len(existingRolls)),
		Inline: true,
	})

	// Show racial bonuses
	if len(race.AbilityBonuses) > 0 {
		bonuses := []string{}
		for _, bonus := range race.AbilityBonuses {
			if bonus.Bonus > 0 {
				bonuses = append(bonuses, fmt.Sprintf("%s +%d", bonus.Attribute, bonus.Bonus))
			}
		}
		if len(bonuses) > 0 {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "🏃 Racial Bonuses",
				Value:  strings.Join(bonuses, "\n") + "\n*(Applied after assignment)*",
				Inline: false,
			})
		}
	}

	// Components
	var components []discordgo.MessageComponent

	if len(existingRolls) < 6 {
		// Still rolling
		components = []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "Roll Next",
						Style:    discordgo.PrimaryButton,
						CustomID: fmt.Sprintf("character_create:roll_individual:%s:%s:%d", req.RaceKey, req.ClassKey, len(existingRolls)),
						Emoji: &discordgo.ComponentEmoji{
							Name: "🎲",
						},
					},
					discordgo.Button{
						Label:    "Use These & Assign",
						Style:    discordgo.SecondaryButton,
						CustomID: fmt.Sprintf("character_create:start_assign:%s:%s", req.RaceKey, req.ClassKey),
						Emoji: &discordgo.ComponentEmoji{
							Name: "📊",
						},
						Disabled: len(existingRolls) == 0,
					},
				},
			},
		}
		embed.Footer = &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Roll %d more times or use current rolls", 6-len(existingRolls)),
		}
	} else {
		// All rolls complete - calculate total and add appropriate flavor text
		totalScore := 0
		for _, roll := range existingRolls {
			totalScore += roll.Value
		}

		flavorText := "The dice have spoken! Your fate is sealed."
		if totalScore >= 78 { // Average of 13+ per stat
			flavorText = "The gods smile upon you! An exceptional set of rolls."
		} else if totalScore <= 60 { // Average of 10- per stat
			flavorText = "The dice show no mercy... But legends are born from adversity!"
		}

		embed.Title = "All Ability Scores Rolled!"
		embed.Footer = &discordgo.MessageEmbedFooter{
			Text: flavorText,
		}
		components = []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "Assign to Abilities",
						Style:    discordgo.PrimaryButton,
						CustomID: fmt.Sprintf("character_create:start_assign:%s:%s", req.RaceKey, req.ClassKey),
						Emoji: &discordgo.ComponentEmoji{
							Name: "📊",
						},
					},
				},
			},
		}
	}

	// Update message
	_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
		Content:    &[]string{""}[0],
	})

	return err
}

// respondWithError updates the message with an error
func (h *RollIndividualHandler) respondWithError(req *RollIndividualRequest, message string) error {
	content := fmt.Sprintf("❌ %s", message)
	_, err := req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
		Content: &content,
		Embeds:  &[]*discordgo.MessageEmbed{},
	})
	return err
}
