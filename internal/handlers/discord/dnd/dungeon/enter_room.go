package dungeon

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/dice"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities/damage"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/encounter"
	"github.com/bwmarrin/discordgo"
)

// Helper function to convert string to *string
func stringPtr(s string) *string {
	return &s
}

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

	// Get session
	sess, err := h.services.SessionService.GetSession(context.Background(), sessionID)
	if err != nil {
		log.Printf("EnterRoom - Session not found: %v", err)
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Session not found!",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// Check if user is in the session
	if !sess.IsUserInSession(i.Member.User.ID) {
		log.Printf("EnterRoom - User %s not in session members", i.Member.User.ID)
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ You need to join the party first!",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// Check if user has a character selected (except for DM/bot)
	member, exists := sess.Members[i.Member.User.ID]
	if exists && member.Role == entities.SessionRolePlayer && member.CharacterID == "" {
		log.Printf("EnterRoom - Player %s has no character selected", i.Member.User.ID)
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ You need to select a character! Click 'Select Character' first.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
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
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Unknown room type!",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}
}

func (h *EnterRoomHandler) handleCombatRoom(s *discordgo.Session, i *discordgo.InteractionCreate, sess *entities.Session) error {
	// Acknowledge immediately to avoid timeout
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		log.Printf("Error deferring response: %v", err)
		return err
	}

	// Get difficulty and room number from session metadata
	difficulty := "medium"
	roomNumber := 1
	if sess.Metadata != nil {
		if diff, ok := sess.Metadata["difficulty"].(string); ok {
			difficulty = diff
		}
		// Try different type assertions for roomNumber
		switch roomNum := sess.Metadata["roomNumber"].(type) {
		case float64:
			roomNumber = int(roomNum)
		case int:
			roomNumber = roomNum
		}
	}
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
		_, editErr := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: stringPtr(fmt.Sprintf("❌ Failed to create encounter: %v", err)),
		})
		if editErr != nil {
			log.Printf("Error editing response: %v", editErr)
		}
		return err
	}

	// Add all party members to encounter
	log.Printf("Adding party members to encounter:")
	for userID, member := range sess.Members {
		log.Printf("Processing member - UserID: %s, Role: %s, CharacterID: %s", userID, member.Role, member.CharacterID)
		if member.CharacterID != "" {
			log.Printf("Adding player - UserID: %s, CharacterID: %s", userID, member.CharacterID)
			combatant, err := h.services.EncounterService.AddPlayer(context.Background(), enc.ID, userID, member.CharacterID)
			if err != nil {
				// Log but continue
				log.Printf("Failed to add player %s: %v", userID, err)
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
		_, editErr := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: stringPtr(fmt.Sprintf("❌ Failed to roll initiative: %v", err)),
		})
		if editErr != nil {
			log.Printf("Error editing response: %v", editErr)
		}
		return err
	}

	// Start combat
	err = h.services.EncounterService.StartEncounter(context.Background(), enc.ID, botID)
	if err != nil {
		_, editErr := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: stringPtr(fmt.Sprintf("❌ Failed to start combat: %v", err)),
		})
		if editErr != nil {
			log.Printf("Error editing response: %v", editErr)
		}
		return err
	}

	// Get updated encounter
	enc, err = h.services.EncounterService.GetEncounter(context.Background(), enc.ID)
	if err != nil {
		_, editErr := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: stringPtr(fmt.Sprintf("❌ Failed to get encounter: %v", err)),
		})
		if editErr != nil {
			log.Printf("Error editing response: %v", editErr)
		}
		return err
	}

	// Process initial monster turns if they go first
	var monsterActions []string
	for enc.Status == entities.EncounterStatusActive {
		current := enc.GetCurrentCombatant()
		if current == nil || current.Type != entities.CombatantTypeMonster {
			break // Stop when we reach a player's turn
		}

		// Process monster turn
		log.Printf("Processing initial monster turn for %s", current.Name)

		// Find a target (first active player)
		var target *entities.Combatant
		for _, combatant := range enc.Combatants {
			if combatant.Type == entities.CombatantTypePlayer && combatant.IsActive {
				target = combatant
				break
			}
		}

		if target != nil && len(current.Actions) > 0 {
			// Use first available action
			action := current.Actions[0]

			// Roll attack
			attackResult, _ := dice.Roll(1, 20, 0)
			attackRoll := attackResult.Total
			totalAttack := attackRoll + action.AttackBonus

			// Check if hit
			hit := totalAttack >= target.AC
			if hit && len(action.Damage) > 0 {
				totalDamage := 0
				for _, dmg := range action.Damage {
					diceCount := dmg.DiceCount
					if attackRoll == 20 { // Critical hit doubles dice
						diceCount *= 2
					}
					rollResult, _ := dice.Roll(diceCount, dmg.DiceSize, dmg.Bonus)
					totalDamage += rollResult.Total
				}

				// Apply damage
				err = h.services.EncounterService.ApplyDamage(context.Background(), enc.ID, target.ID, botID, totalDamage)
				if err != nil {
					log.Printf("Error applying initial monster damage: %v", err)
				}

				monsterActions = append(monsterActions, fmt.Sprintf("%s attacks %s with %s for %d damage!", current.Name, target.Name, action.Name, totalDamage))
			} else {
				monsterActions = append(monsterActions, fmt.Sprintf("%s misses %s with %s!", current.Name, target.Name, action.Name))
			}
		}

		// Advance turn
		err = h.services.EncounterService.NextTurn(context.Background(), enc.ID, botID)
		if err != nil {
			log.Printf("Error advancing initial monster turn: %v", err)
			break
		}

		// Re-get encounter for next iteration
		enc, err = h.services.EncounterService.GetEncounter(context.Background(), enc.ID)
		if err != nil {
			log.Printf("Error getting encounter after monster turn: %v", err)
			break
		}
	}

	// Build combat display
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("⚔️ Combat: %s", room.Name),
		Description: "The party engages in battle!",
		Color:       0xe74c3c, // Red
		Fields:      []*discordgo.MessageEmbedField{},
	}

	// Show initiative rolls from combat log
	if len(enc.CombatLog) > 0 {
		var initiativeRolls strings.Builder
		// Find all initiative roll entries (they contain "rolls initiative:")
		for _, logEntry := range enc.CombatLog {
			if strings.Contains(logEntry, "rolls initiative:") {
				initiativeRolls.WriteString(logEntry + "\n")
			}
		}
		if initiativeRolls.Len() > 0 {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "🎲 Initiative Rolls",
				Value:  initiativeRolls.String(),
				Inline: false,
			})
		}
	}

	// Show monster actions if any occurred
	if len(monsterActions) > 0 {
		var actionList strings.Builder
		for _, action := range monsterActions {
			actionList.WriteString("⚔️ " + action + "\n")
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "🎲 Monster Actions",
			Value:  actionList.String(),
			Inline: false,
		})
	}

	// Show enemies
	var enemyList strings.Builder
	for _, combatant := range enc.Combatants {
		if combatant.Type == entities.CombatantTypeMonster && combatant.IsActive {
			enemyList.WriteString(fmt.Sprintf("• **%s** (HP: %d/%d, AC: %d)\n", combatant.Name, combatant.CurrentHP, combatant.MaxHP, combatant.AC))
		}
	}
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "🐉 Enemies",
		Value:  enemyList.String(),
		Inline: false,
	})

	// Show turn order
	var turnOrder strings.Builder
	log.Printf("Building turn order display. Total combatants: %d, Current turn: %d", len(enc.Combatants), enc.Turn)
	for i, combatantID := range enc.TurnOrder {
		if combatant, exists := enc.Combatants[combatantID]; exists && combatant.IsActive {
			prefix := "  "
			if i == enc.Turn {
				prefix = "▶️"
			}
			turnOrder.WriteString(fmt.Sprintf("%s %s\n", prefix, combatant.Name))
		}
	}

	// Show whose turn it is
	if current := enc.GetCurrentCombatant(); current != nil {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "🎯 Current Turn",
			Value:  fmt.Sprintf("**%s's turn** (HP: %d/%d)", current.Name, current.CurrentHP, current.MaxHP),
			Inline: false,
		})
	}

	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "📋 Initiative Order",
		Value:  turnOrder.String(),
		Inline: false,
	})

	// Check if it's a player's turn
	isPlayerTurn := false
	if current := enc.GetCurrentCombatant(); current != nil && current.Type == entities.CombatantTypePlayer {
		isPlayerTurn = true
	}

	// Combat buttons
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Attack",
					Style:    discordgo.DangerButton,
					CustomID: fmt.Sprintf("encounter:attack:%s", enc.ID),
					Emoji:    &discordgo.ComponentEmoji{Name: "⚔️"},
					Disabled: !isPlayerTurn,
				},
				discordgo.Button{
					Label:    "Next Turn",
					Style:    discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("encounter:next_turn:%s", enc.ID),
					Emoji:    &discordgo.ComponentEmoji{Name: "➡️"},
				},
				discordgo.Button{
					Label:    "Status",
					Style:    discordgo.SecondaryButton,
					CustomID: fmt.Sprintf("encounter:view:%s", enc.ID),
					Emoji:    &discordgo.ComponentEmoji{Name: "📊"},
				},
			},
		},
	}

	// Edit the deferred response
	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
	})
	return err
}

func (h *EnterRoomHandler) handlePuzzleRoom(s *discordgo.Session, i *discordgo.InteractionCreate, sess *entities.Session) error {
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

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

func (h *EnterRoomHandler) handleTrapRoom(s *discordgo.Session, i *discordgo.InteractionCreate, sess *entities.Session) error {
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

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

func (h *EnterRoomHandler) handleTreasureRoom(s *discordgo.Session, i *discordgo.InteractionCreate, sess *entities.Session) error {
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

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

func (h *EnterRoomHandler) handleRestRoom(s *discordgo.Session, i *discordgo.InteractionCreate, sess *entities.Session) error {
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

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
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
			Actions: []*entities.MonsterAction{
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
			Actions: []*entities.MonsterAction{
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
			Actions: []*entities.MonsterAction{
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
			Actions: []*entities.MonsterAction{
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
