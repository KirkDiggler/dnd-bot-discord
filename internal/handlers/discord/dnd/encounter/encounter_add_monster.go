package encounter

import (
	"context"
	"fmt"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/damage"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/combat"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/session"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/services"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/encounter"
	"github.com/bwmarrin/discordgo"
)

type AddMonsterRequest struct {
	Session     *discordgo.Session
	Interaction *discordgo.InteractionCreate
	Query       string
}

type AddMonsterHandler struct {
	services *services.Provider
}

func NewAddMonsterHandler(serviceProvider *services.Provider) *AddMonsterHandler {
	return &AddMonsterHandler{
		services: serviceProvider,
	}
}

func (h *AddMonsterHandler) Handle(req *AddMonsterRequest) error {
	// Defer acknowledge the interaction
	err := req.Session.InteractionRespond(req.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		return fmt.Errorf("failed to acknowledge interaction: %w", err)
	}

	// Get active session for the user
	sessions, err := h.services.SessionService.ListActiveUserSessions(context.Background(), req.Interaction.Member.User.ID)
	if err != nil || len(sessions) == 0 {
		content := "❌ You need to be in an active session to add monsters!"
		_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return err
	}

	// For now, use the first active session
	session := sessions[0]

	// Check if user is DM
	member, exists := session.Members[req.Interaction.Member.User.ID]
	if !exists || member.Role != session.SessionRoleDM {
		content := "❌ Only the DM can add monsters to encounters!"
		_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return err
	}

	// Get active encounter or create one if needed
	activeEncounter, err := h.services.EncounterService.GetActiveEncounter(context.Background(), session.ID)
	if err != nil {
		// No active encounter, create one
		encounterInput := &encounter.CreateEncounterInput{
			SessionID:   session.ID,
			ChannelID:   req.Interaction.ChannelID,
			Name:        "New Encounter",
			Description: "Combat encounter",
			UserID:      req.Interaction.Member.User.ID,
		}

		activeEncounter, err = h.services.EncounterService.CreateEncounter(context.Background(), encounterInput)
		if err != nil {
			content := fmt.Sprintf("❌ Failed to create encounter: %v", err)
			_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
				Content: &content,
			})
			return err
		}
	}

	// Search for monsters using D&D API
	searchQuery := strings.ToLower(strings.TrimSpace(req.Query))
	if searchQuery == "" {
		content := "❌ Please provide a monster name to search for!"
		_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return err
	}

	// For now, we'll use some common monsters as examples
	// In a real implementation, this would search the D&D API
	monsterOptions := h.getCommonMonsters(searchQuery)

	if len(monsterOptions) == 0 {
		content := fmt.Sprintf("❌ No monsters found matching '%s'", req.Query)
		_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return err
	}

	// If only one match, add it directly
	if len(monsterOptions) == 1 {
		monster := monsterOptions[0]
		_, err = h.services.EncounterService.AddMonster(context.Background(), activeEncounter.ID, req.Interaction.Member.User.ID, monster)
		if err != nil {
			content := fmt.Sprintf("❌ Failed to add monster: %v", err)
			_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
				Content: &content,
			})
			return err
		}

		// Success message
		embed := &discordgo.MessageEmbed{
			Title:       "🐉 Monster Added!",
			Description: fmt.Sprintf("**%s** has been added to the encounter!", monster.Name),
			Color:       0x2ecc71, // Green
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "📊 Stats",
					Value:  fmt.Sprintf("**HP:** %d | **AC:** %d | **CR:** %.1f", monster.MaxHP, monster.AC, monster.CR),
					Inline: true,
				},
			},
		}

		// Add action buttons
		components := []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "Add Another",
						Style:    discordgo.PrimaryButton,
						CustomID: fmt.Sprintf("encounter:add_monster:%s", activeEncounter.ID),
						Emoji:    &discordgo.ComponentEmoji{Name: "➕"},
					},
					discordgo.Button{
						Label:    "Roll Initiative",
						Style:    discordgo.SuccessButton,
						CustomID: fmt.Sprintf("encounter:roll_initiative:%s", activeEncounter.ID),
						Emoji:    &discordgo.ComponentEmoji{Name: "🎲"},
					},
					discordgo.Button{
						Label:    "View Encounter",
						Style:    discordgo.SecondaryButton,
						CustomID: fmt.Sprintf("encounter:view:%s", activeEncounter.ID),
						Emoji:    &discordgo.ComponentEmoji{Name: "👁️"},
					},
				},
			},
		}

		_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
			Embeds:     &[]*discordgo.MessageEmbed{embed},
			Components: &components,
		})
		return err
	}

	// Multiple matches, show selection menu
	options := make([]discordgo.SelectMenuOption, 0, len(monsterOptions))
	for i, monster := range monsterOptions {
		if i >= 25 { // Discord limit
			break
		}
		options = append(options, discordgo.SelectMenuOption{
			Label:       monster.Name,
			Description: fmt.Sprintf("CR %.1f | HP %d | AC %d", monster.CR, monster.MaxHP, monster.AC),
			Value:       fmt.Sprintf("%d", i),
		})
	}

	content := fmt.Sprintf("🔍 Found %d monsters matching '%s'. Select one:", len(monsterOptions), req.Query)
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    fmt.Sprintf("encounter:select_monster:%s", activeEncounter.ID),
					Placeholder: "Choose a monster...",
					Options:     options,
				},
			},
		},
	}

	// Store monster options in session state for later retrieval
	// For now, we'll encode them in the interaction response

	_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
		Content:    &content,
		Components: &components,
	})
	return err
}

// getCommonMonsters returns some common D&D monsters for demonstration
// In a real implementation, this would query the D&D API
func (h *AddMonsterHandler) getCommonMonsters(query string) []*encounter.AddMonsterInput {
	monsters := []*encounter.AddMonsterInput{
		{
			Name:            "Goblin",
			MaxHP:           7,
			AC:              15,
			InitiativeBonus: 2,
			Speed:           30,
			CR:              0.25,
			XP:              50,
			MonsterRef:      "goblin",
			Abilities: map[string]int{
				"strength":     8,
				"dexterity":    14,
				"constitution": 10,
				"intelligence": 10,
				"wisdom":       8,
				"charisma":     8,
			},
			Actions: []*combat.MonsterAction{
				{
					Name:        "Scimitar",
					AttackBonus: 4,
					Description: "Melee Weapon Attack: +4 to hit, reach 5 ft., one target.",
					Damage: []*damage.Damage{
						{DiceCount: 1, DiceSize: 6, Bonus: 2, DamageType: damage.TypeSlashing},
					},
				},
				{
					Name:        "Shortbow",
					AttackBonus: 4,
					Description: "Ranged Weapon Attack: +4 to hit, range 80/320 ft., one target.",
					Damage: []*damage.Damage{
						{DiceCount: 1, DiceSize: 6, Bonus: 2, DamageType: damage.TypePiercing},
					},
				},
			},
		},
		{
			Name:            "Orc",
			MaxHP:           15,
			AC:              13,
			InitiativeBonus: 1,
			Speed:           30,
			CR:              0.5,
			XP:              100,
			MonsterRef:      "orc",
			Abilities: map[string]int{
				"strength":     16,
				"dexterity":    12,
				"constitution": 16,
				"intelligence": 7,
				"wisdom":       11,
				"charisma":     10,
			},
			Actions: []*combat.MonsterAction{
				{
					Name:        "Greataxe",
					AttackBonus: 5,
					Description: "Melee Weapon Attack: +5 to hit, reach 5 ft., one target.",
					Damage: []*damage.Damage{
						{DiceCount: 1, DiceSize: 12, Bonus: 3, DamageType: damage.TypeSlashing},
					},
				},
				{
					Name:        "Javelin",
					AttackBonus: 5,
					Description: "Melee or Ranged Weapon Attack: +5 to hit, reach 5 ft. or range 30/120 ft., one target.",
					Damage: []*damage.Damage{
						{DiceCount: 1, DiceSize: 6, Bonus: 3, DamageType: damage.TypePiercing},
					},
				},
			},
		},
		{
			Name:            "Skeleton",
			MaxHP:           13,
			AC:              13,
			InitiativeBonus: 2,
			Speed:           30,
			CR:              0.25,
			XP:              50,
			MonsterRef:      "skeleton",
			Abilities: map[string]int{
				"strength":     10,
				"dexterity":    14,
				"constitution": 15,
				"intelligence": 6,
				"wisdom":       8,
				"charisma":     5,
			},
			Actions: []*combat.MonsterAction{
				{
					Name:        "Shortsword",
					AttackBonus: 4,
					Description: "Melee Weapon Attack: +4 to hit, reach 5 ft., one target.",
					Damage: []*damage.Damage{
						{DiceCount: 1, DiceSize: 6, Bonus: 2, DamageType: damage.TypePiercing},
					},
				},
				{
					Name:        "Shortbow",
					AttackBonus: 4,
					Description: "Ranged Weapon Attack: +4 to hit, range 80/320 ft., one target.",
					Damage: []*damage.Damage{
						{DiceCount: 1, DiceSize: 6, Bonus: 2, DamageType: damage.TypePiercing},
					},
				},
			},
		},
		{
			Name:            "Dire Wolf",
			MaxHP:           37,
			AC:              14,
			InitiativeBonus: 2,
			Speed:           50,
			CR:              1,
			XP:              200,
			MonsterRef:      "dire-wolf",
			Abilities: map[string]int{
				"strength":     17,
				"dexterity":    15,
				"constitution": 15,
				"intelligence": 3,
				"wisdom":       12,
				"charisma":     7,
			},
			Actions: []*combat.MonsterAction{
				{
					Name:        "Bite",
					AttackBonus: 5,
					Description: "Melee Weapon Attack: +5 to hit, reach 5 ft., one target. If the target is a creature, it must succeed on a DC 13 Strength saving throw or be knocked prone.",
					Damage: []*damage.Damage{
						{DiceCount: 2, DiceSize: 6, Bonus: 3, DamageType: damage.TypePiercing},
					},
				},
			},
		},
		{
			Name:            "Zombie",
			MaxHP:           22,
			AC:              8,
			InitiativeBonus: -2,
			Speed:           20,
			CR:              0.25,
			XP:              50,
			MonsterRef:      "zombie",
			Abilities: map[string]int{
				"strength":     13,
				"dexterity":    6,
				"constitution": 16,
				"intelligence": 3,
				"wisdom":       6,
				"charisma":     5,
			},
			Actions: []*combat.MonsterAction{
				{
					Name:        "Slam",
					AttackBonus: 3,
					Description: "Melee Weapon Attack: +3 to hit, reach 5 ft., one target.",
					Damage: []*damage.Damage{
						{DiceCount: 1, DiceSize: 6, Bonus: 1, DamageType: damage.TypeBludgeoning},
					},
				},
			},
		},
	}

	// Filter by query
	var filtered []*encounter.AddMonsterInput
	for _, monster := range monsters {
		if strings.Contains(strings.ToLower(monster.Name), query) {
			filtered = append(filtered, monster)
		}
	}

	return filtered
}
