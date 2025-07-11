package dungeon

import (
	"context"
	"fmt"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/damage"
	combat2 "github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/combat"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/session"
	"log"
	"math/rand"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/handlers/discord/combat"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/encounter"
	"github.com/bwmarrin/discordgo"
)

type EnterRoomHandler struct {
	services *services.Provider
}

func NewEnterRoomHandler(serviceProvider *services.Provider) *EnterRoomHandler {
	return &EnterRoomHandler{
		services: serviceProvider,
	}
}

func (h *EnterRoomHandler) HandleButton(s *discordgo.Session, i *discordgo.InteractionCreate, sessionID, roomType string) error {
	log.Printf("EnterRoom - User %s attempting to enter %s room in session %s", i.Member.User.ID, roomType, sessionID)

	// Defer the response immediately for long operations
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
	if err != nil {
		log.Printf("EnterRoom - Failed to defer response: %v", err)
		// Try to respond normally if defer fails
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Failed to process room entry. Please try again.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// Get session
	sess, err := h.services.SessionService.GetSession(context.Background(), sessionID)
	if err != nil {
		log.Printf("EnterRoom - Session not found: %v", err)
		content := "❌ Session not found!"
		_, editErr := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return editErr
	}

	// Check if user is in the session
	if !sess.IsUserInSession(i.Member.User.ID) {
		log.Printf("EnterRoom - User %s not in session members", i.Member.User.ID)
		content := "❌ You need to join the party first!"
		_, editErr := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return editErr
	}

	// Check if user has a character selected (except for DM/bot)
	member, exists := sess.Members[i.Member.User.ID]
	if exists && member.Role == session.SessionRolePlayer && member.CharacterID == "" {
		log.Printf("EnterRoom - Player %s has no character selected", i.Member.User.ID)
		content := "❌ You need to select a character! Click 'Select Character' first."
		_, editErr := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return editErr
	}

	if exists {
		log.Printf("EnterRoom - User %s has role=%s, characterID=%s", i.Member.User.ID, member.Role, member.CharacterID)
	}

	// Handle based on room type
	switch RoomType(roomType) {
	case RoomTypeCombat:
		return h.handleCombatRoom(s, i, sess)
	case RoomTypePuzzle:
		return h.handlePuzzleRoom(s, i, sess)
	case RoomTypeTrap:
		return h.handleTrapRoom(s, i, sess)
	case RoomTypeTreasure:
		return h.handleTreasureRoom(s, i, sess)
	case RoomTypeRest:
		return h.handleRestRoom(s, i, sess)
	default:
		content := "❌ Unknown room type!"
		_, editErr := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return editErr
	}
}

func (h *EnterRoomHandler) handleCombatRoom(s *discordgo.Session, i *discordgo.InteractionCreate, sess *session.Session) error {
	// Get difficulty and room number from session metadata
	difficulty := sess.GetDifficulty()
	roomNumber := sess.GetRoomNumber()
	fmt.Printf("Combat room - Difficulty: %s, Room Number: %d\n", difficulty, roomNumber)

	// Generate a combat room
	room := h.generateCombatRoom(difficulty, roomNumber)

	// Create encounter
	botID := s.State.User.ID
	log.Printf("Creating encounter - Bot ID: %s, Session ID: %s", botID, sess.ID)

	// Log session members
	log.Printf("Current session members:")
	for userID, member := range sess.Members {
		log.Printf("  - User %s: Role=%s", userID, member.Role)
	}

	encounterInput := &encounter.CreateEncounterInput{
		SessionID:   sess.ID,
		ChannelID:   i.ChannelID,
		Name:        room.Name,
		Description: room.Description,
		UserID:      botID, // Bot manages the encounter
	}

	enc, err := h.services.EncounterService.CreateEncounter(context.Background(), encounterInput)
	if err != nil {
		content := fmt.Sprintf("❌ Failed to create encounter: %v", err)
		_, editErr := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return editErr
	}

	// Add all party members to encounter
	log.Printf("Adding party members to encounter:")
	for userID, member := range sess.Members {
		log.Printf("Processing member - UserID: %s, Role: %s, CharacterID: %s", userID, member.Role, member.CharacterID)
		if member.CharacterID != "" {
			log.Printf("Adding player - UserID: %s, CharacterID: %s", userID, member.CharacterID)
			combatant, addErr := h.services.EncounterService.AddPlayer(context.Background(), enc.ID, userID, member.CharacterID)
			if addErr != nil {
				// Log but continue
				log.Printf("Failed to add player %s: %v", userID, addErr)
			} else if combatant != nil {
				log.Printf("Added player combatant: Name=%s, Type=%s, HP=%d/%d, PlayerID=%s", combatant.Name, combatant.Type, combatant.CurrentHP, combatant.MaxHP, combatant.PlayerID)
			}
		} else {
			log.Printf("Skipping member %s - no character ID", userID)
		}
	}

	// Add monsters from room
	for _, monsterName := range room.Monsters {
		if monster := h.getMonster(monsterName); monster != nil {
			_, err = h.services.EncounterService.AddMonster(context.Background(), enc.ID, botID, monster)
			if err != nil {
				fmt.Printf("Failed to add monster %s: %v\n", monsterName, err)
			}
		}
	}

	// Roll initiative
	err = h.services.EncounterService.RollInitiative(context.Background(), enc.ID, botID)
	if err != nil {
		content := fmt.Sprintf("❌ Failed to roll initiative: %v", err)
		_, editErr := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return editErr
	}

	// Start combat
	err = h.services.EncounterService.StartEncounter(context.Background(), enc.ID, botID)
	if err != nil {
		content := fmt.Sprintf("❌ Failed to start combat: %v", err)
		_, editErr := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return editErr
	}

	// Get updated encounter after all setup is complete
	enc, err = h.services.EncounterService.GetEncounter(context.Background(), enc.ID)
	if err != nil {
		content := fmt.Sprintf("❌ Failed to get encounter: %v", err)
		_, editErr := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return editErr
	}

	// Log the combat log for debugging
	log.Printf("Combat log after setup (%d entries):", len(enc.CombatLog))
	for i, entry := range enc.CombatLog {
		log.Printf("  [%d] %s", i, entry)
	}

	// Process initial monster turns if they go first
	for enc.Status == combat2.EncounterStatusActive {
		current := enc.GetCurrentCombatant()
		if current == nil || current.Type != combat2.CombatantTypeMonster {
			break // Stop when we reach a player's turn
		}

		// Process monster turn
		log.Printf("Processing initial monster turn for %s", current.Name)

		// Find a target (first active player)
		var target *combat2.Combatant
		for _, combatant := range enc.Combatants {
			if combatant.Type == combat2.CombatantTypePlayer && combatant.IsActive {
				target = combatant
				break
			}
		}

		if target != nil {
			// Use PerformAttack for consistency with combat flow
			attackInput := &encounter.AttackInput{
				EncounterID: enc.ID,
				AttackerID:  current.ID,
				TargetID:    target.ID,
				UserID:      botID,
				ActionIndex: 0, // Use first action
			}
			var attackResult *encounter.AttackResult
			attackResult, err = h.services.EncounterService.PerformAttack(
				context.Background(),
				attackInput,
			)
			if err != nil {
				log.Printf("Error performing monster attack: %v", err)
			} else if attackResult != nil {
				log.Printf("Monster attack result: %s", attackResult.LogEntry)
			}
		}

		// Advance turn after monster attack
		err = h.services.EncounterService.NextTurn(context.Background(), enc.ID, botID)
		if err != nil {
			log.Printf("Error advancing turn after monster attack: %v", err)
			break
		}

		// Re-get encounter for next iteration
		enc, err = h.services.EncounterService.GetEncounter(context.Background(), enc.ID)
		if err != nil {
			log.Printf("Error getting encounter after monster turn: %v", err)
			break
		}
	}

	// Use the combat embed builder for consistent formatting
	// Don't pass monsterAttackResults since they're already in the combat log
	embed := combat.BuildCombatStatusEmbed(enc, nil)

	// Add room description to the embed
	if embed.Description != "" {
		embed.Description = fmt.Sprintf("**%s**\n\n%s", room.Description, embed.Description)
	} else {
		embed.Description = fmt.Sprintf("**%s**", room.Description)
	}
	embed.Title = fmt.Sprintf("⚔️ %s", enc.Name)

	// Check if it's a player's turn
	isPlayerTurn := false
	if current := enc.GetCurrentCombatant(); current != nil && current.Type == combat2.CombatantTypePlayer {
		isPlayerTurn = true
	}

	// Combat buttons
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Attack",
					Style:    discordgo.DangerButton,
					CustomID: fmt.Sprintf("combat:attack:%s", enc.ID),
					Emoji:    &discordgo.ComponentEmoji{Name: "⚔️"},
					Disabled: !isPlayerTurn,
				},
				discordgo.Button{
					Label:    "Next Turn",
					Style:    discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("combat:next_turn:%s", enc.ID),
					Emoji:    &discordgo.ComponentEmoji{Name: "➡️"},
				},
				discordgo.Button{
					Label:    "Get My Actions",
					Style:    discordgo.SuccessButton,
					CustomID: fmt.Sprintf("combat:my_actions:%s", enc.ID),
					Emoji:    &discordgo.ComponentEmoji{Name: "🎯"},
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Status",
					Style:    discordgo.SecondaryButton,
					CustomID: fmt.Sprintf("combat:view:%s", enc.ID),
					Emoji:    &discordgo.ComponentEmoji{Name: "📊"},
				},
				discordgo.Button{
					Label:    "History",
					Style:    discordgo.SecondaryButton,
					CustomID: fmt.Sprintf("combat:history:%s", enc.ID),
					Emoji:    &discordgo.ComponentEmoji{Name: "📜"},
				},
			},
		},
	}

	// TODO: Address duplicate combat UI issue - when 2 players are in a dungeon,
	// the first player gets the old style message and needs to enter again to get
	// the newer shared message format. This may be due to multiple handlers responding
	// to the same interaction.

	// Send the combat UI and capture the message ID
	msg, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
	})
	if err != nil {
		return err
	}

	// Store the message ID in the encounter (it's the same message as the lobby)
	if msg != nil && msg.ID != "" {
		enc.MessageID = msg.ID
		enc.ChannelID = msg.ChannelID

		// Update encounter with message ID
		err = h.services.EncounterService.UpdateMessageID(context.Background(), enc.ID, msg.ID, msg.ChannelID)
		if err != nil {
			log.Printf("Failed to store message ID for encounter %s: %v", enc.ID, err)
			// Continue anyway - this is not critical for combat to function
		} else {
			log.Printf("Combat using same message as lobby: %s", msg.ID)
		}
	}

	return nil
}

func (h *EnterRoomHandler) handlePuzzleRoom(s *discordgo.Session, i *discordgo.InteractionCreate, sess *session.Session) error {
	// Placeholder for puzzle room
	embed := &discordgo.MessageEmbed{
		Title:       "🧩 Puzzle Room",
		Description: "This room contains a challenging puzzle!",
		Color:       0x3498db, // Blue
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "🔍 Challenge",
				Value:  "Puzzle mechanics coming soon!",
				Inline: false,
			},
		},
	}

	_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	})
	return err
}

func (h *EnterRoomHandler) handleTrapRoom(s *discordgo.Session, i *discordgo.InteractionCreate, sess *session.Session) error {
	// Placeholder for trap room
	embed := &discordgo.MessageEmbed{
		Title:       "⚠️ Trap Room",
		Description: "Watch your step!",
		Color:       0xf39c12, // Orange
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "💀 Danger",
				Value:  "Trap mechanics coming soon!",
				Inline: false,
			},
		},
	}

	_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	})
	return err
}

func (h *EnterRoomHandler) handleTreasureRoom(s *discordgo.Session, i *discordgo.InteractionCreate, sess *session.Session) error {
	// Placeholder for treasure room
	embed := &discordgo.MessageEmbed{
		Title:       "💰 Treasure Room",
		Description: "Riches await!",
		Color:       0xf1c40f, // Gold
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "✨ Loot",
				Value:  "Treasure mechanics coming soon!",
				Inline: false,
			},
		},
	}

	_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	})
	return err
}

func (h *EnterRoomHandler) handleRestRoom(s *discordgo.Session, i *discordgo.InteractionCreate, sess *session.Session) error {
	// Placeholder for rest room
	embed := &discordgo.MessageEmbed{
		Title:       "💤 Rest Area",
		Description: "A safe place to recover.",
		Color:       0x2ecc71, // Green
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "🏕️ Rest",
				Value:  "Rest mechanics coming soon!",
				Inline: false,
			},
		},
	}

	_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	})
	return err
}

// getMonster returns a predefined monster by name
func (h *EnterRoomHandler) getMonster(name string) *encounter.AddMonsterInput {
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
			Actions: []*combat2.MonsterAction{
				{
					Name:        "Scimitar",
					AttackBonus: 4,
					Description: "Melee Weapon Attack: +4 to hit",
					Damage: []*damage.Damage{
						{DiceCount: 1, DiceSize: 6, Bonus: 2, DamageType: damage.TypeSlashing},
					},
				},
			},
		},
		"skeleton": {
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
			Actions: []*combat2.MonsterAction{
				{
					Name:        "Shortsword",
					AttackBonus: 4,
					Description: "Melee Weapon Attack: +4 to hit",
					Damage: []*damage.Damage{
						{DiceCount: 1, DiceSize: 6, Bonus: 2, DamageType: damage.TypePiercing},
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
			Actions: []*combat2.MonsterAction{
				{
					Name:        "Greataxe",
					AttackBonus: 5,
					Description: "Melee Weapon Attack: +5 to hit",
					Damage: []*damage.Damage{
						{DiceCount: 1, DiceSize: 12, Bonus: 3, DamageType: damage.TypeSlashing},
					},
				},
			},
		},
		"dire wolf": {
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
			Actions: []*combat2.MonsterAction{
				{
					Name:        "Bite",
					AttackBonus: 5,
					Description: "Melee Weapon Attack: +5 to hit",
					Damage: []*damage.Damage{
						{DiceCount: 2, DiceSize: 6, Bonus: 3, DamageType: damage.TypePiercing},
					},
				},
			},
		},
	}

	return monsters[name]
}

// generateCombatRoom creates a combat encounter room
func (h *EnterRoomHandler) generateCombatRoom(difficulty string, roomNumber int) *Room {
	rooms := []struct {
		name        string
		description string
	}{
		{"Guard Chamber", "Stone walls echo with the sounds of movement. Weapons glint in the torchlight."},
		{"Ancient Crypt", "Dusty sarcophagi line the walls. Something stirs in the darkness."},
		{"Goblin Warren", "The stench is overwhelming. Crude weapons and bones litter the floor."},
		{"Spider's Den", "Thick webs cover every surface. Multiple eyes gleam from the shadows."},
	}

	selected := rooms[rand.Intn(len(rooms))]

	// Determine monsters based on difficulty
	var monsters []string
	switch difficulty {
	case "easy":
		monsters = []string{"goblin", "skeleton"}
	case "medium":
		monsters = []string{"orc", "goblin", "goblin"}
	case "hard":
		monsters = []string{"orc", "dire wolf", "skeleton", "skeleton"}
	default:
		monsters = []string{"goblin"}
	}

	// Scale with room number
	extraMonsters := roomNumber / 3
	for i := 0; i < extraMonsters; i++ {
		monsters = append(monsters, monsters[rand.Intn(len(monsters))])
	}

	return &Room{
		Type:        RoomTypeCombat,
		Name:        selected.name,
		Description: selected.description,
		Monsters:    monsters,
		Challenge:   fmt.Sprintf("Defeat all %d enemies!", len(monsters)),
	}
}
