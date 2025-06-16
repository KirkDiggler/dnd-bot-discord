package character

import (
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e"
)

// AbilityScoresHandler handles the ability score generation
type AbilityScoresHandler struct {
	dndClient dnd5e.Client
}

// AbilityScoresHandlerConfig holds configuration
type AbilityScoresHandlerConfig struct {
	DNDClient dnd5e.Client
}

// NewAbilityScoresHandler creates a new handler
func NewAbilityScoresHandler(cfg *AbilityScoresHandlerConfig) *AbilityScoresHandler {
	return &AbilityScoresHandler{
		dndClient: cfg.DNDClient,
	}
}

// AbilityScoresRequest represents the request
type AbilityScoresRequest struct {
	Session     *discordgo.Session
	Interaction *discordgo.InteractionCreate
	RaceKey     string
	ClassKey    string
}

// Handle processes ability score generation
func (h *AbilityScoresHandler) Handle(req *AbilityScoresRequest) error {
	// Update the message
	err := req.Session.InteractionRespond(req.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content: "Rolling ability scores...",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to acknowledge interaction: %w", err)
	}

	// Get race and class for context
	race, err := h.dndClient.GetRace(req.RaceKey)
	if err != nil {
		return h.respondWithError(req, "Failed to fetch race details.")
	}

	class, err := h.dndClient.GetClass(req.ClassKey)
	if err != nil {
		return h.respondWithError(req, "Failed to fetch class details.")
	}

	// Roll ability scores using 4d6 drop lowest
	rolls := h.rollAbilityScores()

	// Create embed
	embed := &discordgo.MessageEmbed{
		Title:       "Create New Character - Ability Scores",
		Description: fmt.Sprintf("**Race:** %s\n**Class:** %s\n\nYour ability scores have been rolled! Assign them to your attributes.", race.Name, class.Name),
		Color:       0x5865F2,
		Fields:      []*discordgo.MessageEmbedField{},
	}

	// Show rolled values
	rollStrings := []string{}
	for i, roll := range rolls {
		rollStrings = append(rollStrings, fmt.Sprintf("**Roll %d:** %d", i+1, roll))
	}
	
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "🎲 Rolled Values",
		Value:  strings.Join(rollStrings, "\n"),
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
				Inline: true,
			})
		}
	}

	// Class recommendations
	recommendations := h.getClassRecommendations(class.Name)
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("💡 %s Recommendations", class.Name),
		Value:  recommendations,
		Inline: false,
	})

	// Progress
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "Progress",
		Value:  "✅ Step 1: Race\n✅ Step 2: Class\n⏳ Step 3: Abilities\n⏳ Step 4: Details",
		Inline: false,
	})

	embed.Footer = &discordgo.MessageEmbedFooter{
		Text: "Click 'Assign to Abilities' to assign each roll to a specific ability",
	}

	// Store rolls in custom ID for the assign button
	rollsStr := []string{}
	for _, roll := range rolls {
		rollsStr = append(rollsStr, strconv.Itoa(roll))
	}
	rollsData := strings.Join(rollsStr, ",")
	
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Assign to Abilities",
					Style:    discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("character_create:start_assign:%s:%s:%s", req.RaceKey, req.ClassKey, rollsData),
					Emoji: &discordgo.ComponentEmoji{
						Name: "📊",
					},
				},
				discordgo.Button{
					Label:    "Reroll",
					Style:    discordgo.SecondaryButton,
					CustomID: fmt.Sprintf("character_create:ability_scores:%s:%s", req.RaceKey, req.ClassKey),
					Emoji: &discordgo.ComponentEmoji{
						Name: "🎲",
					},
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

// rollAbilityScores rolls 6 ability scores using 4d6 drop lowest
func (h *AbilityScoresHandler) rollAbilityScores() []int {
	scores := make([]int, 6)
	
	for i := 0; i < 6; i++ {
		// Roll 4d6
		rolls := make([]int, 4)
		for j := 0; j < 4; j++ {
			rolls[j] = rand.Intn(6) + 1
		}
		
		// Sort and drop lowest
		sort.Ints(rolls)
		total := 0
		for j := 1; j < 4; j++ { // Skip index 0 (lowest)
			total += rolls[j]
		}
		
		scores[i] = total
	}
	
	// Sort scores highest to lowest for display
	sort.Sort(sort.Reverse(sort.IntSlice(scores)))
	
	return scores
}

// getClassRecommendations returns ability score recommendations for a class
func (h *AbilityScoresHandler) getClassRecommendations(className string) string {
	recommendations := map[string]string{
		"Fighter":   "Prioritize **Strength** or **Dexterity** for attacks, then **Constitution** for survivability.",
		"Wizard":    "Prioritize **Intelligence** for spellcasting, then **Constitution** for survivability.",
		"Cleric":    "Prioritize **Wisdom** for spellcasting, then **Constitution** for survivability.",
		"Rogue":     "Prioritize **Dexterity** for attacks and AC, then **Constitution** or **Intelligence**.",
		"Ranger":    "Prioritize **Dexterity** for attacks, then **Wisdom** for spells and perception.",
		"Barbarian": "Prioritize **Strength** for attacks and **Constitution** for HP and AC.",
		"Bard":      "Prioritize **Charisma** for spellcasting, then **Dexterity** for AC.",
		"Druid":     "Prioritize **Wisdom** for spellcasting, then **Constitution** for survivability.",
		"Monk":      "Prioritize **Dexterity** and **Wisdom** equally for AC and features.",
		"Paladin":   "Prioritize **Strength** for attacks and **Charisma** for spells and auras.",
		"Sorcerer":  "Prioritize **Charisma** for spellcasting, then **Constitution** for survivability.",
		"Warlock":   "Prioritize **Charisma** for spellcasting, then **Constitution** for survivability.",
	}
	
	if rec, ok := recommendations[className]; ok {
		return rec
	}
	
	return "Assign your highest scores to your primary abilities."
}

// respondWithError updates the message with an error
func (h *AbilityScoresHandler) respondWithError(req *AbilityScoresRequest, message string) error {
	content := fmt.Sprintf("❌ %s", message)
	_, err := req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
		Content: &content,
		Embeds:  &[]*discordgo.MessageEmbed{},
	})
	return err
}