package combat

import (
	"fmt"
	"log"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/bwmarrin/discordgo"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Common skill mappings to attributes
var SkillToAttribute = map[string]entities.Attribute{
	"skill-acrobatics":      entities.AttributeDexterity,
	"skill-animal-handling": entities.AttributeWisdom,
	"skill-arcana":          entities.AttributeIntelligence,
	"skill-athletics":       entities.AttributeStrength,
	"skill-deception":       entities.AttributeCharisma,
	"skill-history":         entities.AttributeIntelligence,
	"skill-insight":         entities.AttributeWisdom,
	"skill-intimidation":    entities.AttributeCharisma,
	"skill-investigation":   entities.AttributeIntelligence,
	"skill-medicine":        entities.AttributeWisdom,
	"skill-nature":          entities.AttributeIntelligence,
	"skill-perception":      entities.AttributeWisdom,
	"skill-performance":     entities.AttributeCharisma,
	"skill-persuasion":      entities.AttributeCharisma,
	"skill-religion":        entities.AttributeIntelligence,
	"skill-sleight-of-hand": entities.AttributeDexterity,
	"skill-stealth":         entities.AttributeDexterity,
	"skill-survival":        entities.AttributeWisdom,
}

// SkillCheckHandler handles skill check interactions
type SkillCheckHandler struct {
	characterService character.Service
}

// SkillCheckHandlerConfig holds configuration for the skill check handler
type SkillCheckHandlerConfig struct {
	CharacterService character.Service
}

// NewSkillCheckHandler creates a new skill check handler
func NewSkillCheckHandler(cfg *SkillCheckHandlerConfig) *SkillCheckHandler {
	return &SkillCheckHandler{
		characterService: cfg.CharacterService,
	}
}

// ShowSkillCheckPrompt displays a prompt for a player to make a skill check
func (h *SkillCheckHandler) ShowSkillCheckPrompt(s *discordgo.Session, i *discordgo.InteractionCreate,
	char *entities.Character, skillKey string, dc int, reason string) error {

	// Get the attribute for this skill
	attribute, ok := SkillToAttribute[skillKey]
	if !ok {
		return fmt.Errorf("unknown skill: %s", skillKey)
	}

	// Calculate the bonus
	bonus := char.GetSkillBonus(skillKey, attribute)
	profIndicator := ""
	if char.HasSkillProficiency(skillKey) {
		profIndicator = " (PROF)"
	}

	// Format skill name
	skillName := cases.Title(language.English).String(strings.ReplaceAll(strings.TrimPrefix(skillKey, "skill-"), "-", " "))

	embed := &discordgo.MessageEmbed{
		Title:       "🎯 Skill Check Required!",
		Description: fmt.Sprintf("%s must make a **%s** check!\n\n%s", char.Name, skillName, reason),
		Color:       0x3498db, // Blue
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "DC",
				Value:  fmt.Sprintf("%d", dc),
				Inline: true,
			},
			{
				Name:   "Your Bonus",
				Value:  fmt.Sprintf("%+d%s (%s)", bonus, profIndicator, attribute.Short()),
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
					Label:    fmt.Sprintf("Roll %s", skillName),
					Style:    discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("skill_check:%s:%s:%d", char.ID, skillKey, dc),
					Emoji: &discordgo.ComponentEmoji{
						Name: "🎯",
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

// HandleSkillCheckRoll processes a skill check roll
func (h *SkillCheckHandler) HandleSkillCheckRoll(s *discordgo.Session, i *discordgo.InteractionCreate,
	characterID string, skillKey string, dc int) error {

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

	// Get the attribute for this skill
	attribute, ok := SkillToAttribute[skillKey]
	if !ok {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Unknown skill!",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// Roll the skill check
	roll, total, err := char.RollSkillCheck(skillKey, attribute)
	if err != nil {
		log.Printf("Failed to roll skill check: %v", err)
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Failed to roll skill check!",
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

	bonus := char.GetSkillBonus(skillKey, attribute)
	profIndicator := ""
	if char.HasSkillProficiency(skillKey) {
		profIndicator = " (PROF)"
	}

	// Format skill name
	skillName := cases.Title(language.English).String(strings.ReplaceAll(strings.TrimPrefix(skillKey, "skill-"), "-", " "))

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s Check", skillName),
		Description: fmt.Sprintf("%s rolls a %s check!", char.Name, skillName),
		Color:       resultColor,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Roll",
				Value:  fmt.Sprintf("🎲 %d", roll.Rolls[0]),
				Inline: true,
			},
			{
				Name:   "Bonus",
				Value:  fmt.Sprintf("%+d%s (%s)", bonus, profIndicator, attribute.Short()),
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
