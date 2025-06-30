package testcombat

import (
	"context"
	"fmt"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/damage"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/combat"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/services"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/encounter"
	sessService "github.com/KirkDiggler/dnd-bot-discord/internal/services/session"
	"github.com/bwmarrin/discordgo"
)

type TestCombatRequest struct {
	Session     *discordgo.Session
	Interaction *discordgo.InteractionCreate
	MonsterName string
}

type TestCombatHandler struct {
	services *services.Provider
}

func NewTestCombatHandler(servicesProvider *services.Provider) *TestCombatHandler {
	return &TestCombatHandler{
		services: servicesProvider,
	}
}

func (h *TestCombatHandler) Handle(req *TestCombatRequest) error {
	// Defer acknowledge the interaction
	err := req.Session.InteractionRespond(req.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		return fmt.Errorf("failed to acknowledge interaction: %w", err)
	}

	// Create a test session with bot as DM
	botID := req.Session.State.User.ID
	sessionInput := &sessService.CreateSessionInput{
		Name:        "Test Combat Arena",
		Description: "Quick combat test session",
		CreatorID:   botID,
		RealmID:     req.Interaction.GuildID,
		ChannelID:   req.Interaction.ChannelID,
	}

	session, err := h.services.SessionService.CreateSession(context.Background(), sessionInput)
	if err != nil {
		content := fmt.Sprintf("❌ Failed to create test session: %v", err)
		_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return err
	}

	// Get user's active character
	chars, err := h.services.CharacterService.ListByOwner(req.Interaction.Member.User.ID)
	if err != nil || len(chars) == 0 {
		content := "❌ You need an active character first! Use `/dnd character create`"
		_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return err
	}

	// Find first active character
	var playerChar *character.Character
	for _, char := range chars {
		if char.Status == character.CharacterStatusActive {
			playerChar = char
			break
		}
	}

	if playerChar == nil {
		content := "❌ No active character found! Activate a character first."
		_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return err
	}

	// Join the player to session automatically
	_, err = h.services.SessionService.JoinSession(context.Background(), session.ID, req.Interaction.Member.User.ID)
	if err != nil {
		content := fmt.Sprintf("❌ Failed to join session: %v", err)
		_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return err
	}

	// Select character for the player
	err = h.services.SessionService.SelectCharacter(context.Background(), session.ID, req.Interaction.Member.User.ID, playerChar.ID)
	if err != nil {
		content := fmt.Sprintf("❌ Failed to select character: %v", err)
		_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return err
	}

	// Start the session
	err = h.services.SessionService.StartSession(context.Background(), session.ID, botID)
	if err != nil {
		content := fmt.Sprintf("❌ Failed to start session: %v", err)
		_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return err
	}

	// Create encounter
	encounterInput := &encounter.CreateEncounterInput{
		SessionID:   session.ID,
		ChannelID:   req.Interaction.ChannelID,
		Name:        "Test Combat",
		Description: "Testing combat mechanics",
		UserID:      botID, // Bot is DM
	}

	enc, err := h.services.EncounterService.CreateEncounter(context.Background(), encounterInput)
	if err != nil {
		content := fmt.Sprintf("❌ Failed to create encounter: %v", err)
		_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return err
	}

	// Add player character to encounter
	_, err = h.services.EncounterService.AddPlayer(context.Background(), enc.ID, req.Interaction.Member.User.ID, playerChar.ID)
	if err != nil {
		content := fmt.Sprintf("❌ Failed to add player to encounter: %v", err)
		_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return err
	}

	// Add requested monster
	monsterName := strings.TrimSpace(req.MonsterName)
	if monsterName == "" {
		monsterName = "goblin" // Default
	}

	// Get monster from our predefined list
	monster := h.getMonster(monsterName)
	if monster == nil {
		content := fmt.Sprintf("❌ Unknown monster '%s'. Try: goblin, orc, skeleton, dire wolf, or zombie", monsterName)
		_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return err
	}

	_, err = h.services.EncounterService.AddMonster(context.Background(), enc.ID, botID, monster)
	if err != nil {
		content := fmt.Sprintf("❌ Failed to add monster: %v", err)
		_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return err
	}

	// Roll initiative
	err = h.services.EncounterService.RollInitiative(context.Background(), enc.ID, botID)
	if err != nil {
		content := fmt.Sprintf("❌ Failed to roll initiative: %v", err)
		_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return err
	}

	// Start combat
	err = h.services.EncounterService.StartEncounter(context.Background(), enc.ID, botID)
	if err != nil {
		content := fmt.Sprintf("❌ Failed to start combat: %v", err)
		_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return err
	}

	// Get final encounter state
	enc, err = h.services.EncounterService.GetEncounter(context.Background(), enc.ID)
	if err != nil {
		content := fmt.Sprintf("❌ Failed to get encounter: %v", err)
		_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return err
	}

	// Build response
	embed := &discordgo.MessageEmbed{
		Title:       "⚔️ Test Combat Started!",
		Description: fmt.Sprintf("**%s** vs **%s**", playerChar.Name, monster.Name),
		Color:       0xe74c3c, // Red
		Fields:      []*discordgo.MessageEmbedField{},
	}

	// Show turn order
	var turnOrder strings.Builder
	for i, combatantID := range enc.TurnOrder {
		if combatant, exists := enc.Combatants[combatantID]; exists {
			prefix := "  "
			if i == 0 {
				prefix = "▶️"
			}
			turnOrder.WriteString(fmt.Sprintf("%s %s (Initiative: %d)\n", prefix, combatant.Name, combatant.Initiative))
		}
	}

	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "📋 Turn Order",
		Value:  turnOrder.String(),
		Inline: false,
	})

	// Show current turn
	if current := enc.GetCurrentCombatant(); current != nil {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "🎯 Current Turn",
			Value:  fmt.Sprintf("**%s** (HP: %d/%d | AC: %d)", current.Name, current.CurrentHP, current.MaxHP, current.AC),
			Inline: false,
		})
	}

	// Add combat action buttons
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Attack",
					Style:    discordgo.DangerButton,
					CustomID: fmt.Sprintf("encounter:attack:%s", enc.ID),
					Emoji:    &discordgo.ComponentEmoji{Name: "⚔️"},
				},
				discordgo.Button{
					Label:    "Next Turn",
					Style:    discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("encounter:next_turn:%s", enc.ID),
					Emoji:    &discordgo.ComponentEmoji{Name: "➡️"},
				},
				discordgo.Button{
					Label:    "View Status",
					Style:    discordgo.SecondaryButton,
					CustomID: fmt.Sprintf("encounter:view:%s", enc.ID),
					Emoji:    &discordgo.ComponentEmoji{Name: "📊"},
				},
			},
		},
	}

	embed.Footer = &discordgo.MessageEmbedFooter{
		Text: "Bot is acting as DM. Use Attack when it's your turn!",
	}

	_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
	})
	return err
}

// getMonster returns a predefined monster by name
func (h *TestCombatHandler) getMonster(name string) *encounter.AddMonsterInput {
	name = strings.ToLower(name)

	monsters := map[string]*encounter.AddMonsterInput{
		"goblin": {
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
			},
		},
		"orc": {
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
			},
		},
	}

	return monsters[name]
}
