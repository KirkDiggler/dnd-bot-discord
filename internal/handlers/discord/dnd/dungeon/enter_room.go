package dungeon

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities/damage"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/encounter"
	"github.com/bwmarrin/discordgo"
)

type EnterRoomHandler struct {
	services *services.Provider
}

func NewEnterRoomHandler(services *services.Provider) *EnterRoomHandler {
	return &EnterRoomHandler{
		services: services,
	}
}

func (h *EnterRoomHandler) HandleButton(s *discordgo.Session, i *discordgo.InteractionCreate, sessionID string, roomType string) error {
	// Get session
	sess, err := h.services.SessionService.GetSession(context.Background(), sessionID)
	if err != nil {
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
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ You need to join the party first!",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
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
	// Get difficulty and room number from session metadata
	difficulty := "medium"
	roomNumber := 1
	if sess.Metadata != nil {
		if diff, ok := sess.Metadata["difficulty"].(string); ok {
			difficulty = diff
		}
		// Try different type assertions for roomNumber
		if roomNum, ok := sess.Metadata["roomNumber"].(float64); ok {
			roomNumber = int(roomNum)
		} else if roomNum, ok := sess.Metadata["roomNumber"].(int); ok {
			roomNumber = roomNum
		}
	}
	fmt.Printf("Combat room - Difficulty: %s, Room Number: %d\n", difficulty, roomNumber)
	
	// Generate a combat room
	room := h.generateCombatRoom(difficulty, roomNumber)
	
	// Create encounter
	botID := s.State.User.ID
	encounterInput := &encounter.CreateEncounterInput{
		SessionID:   sess.ID,
		ChannelID:   i.ChannelID,
		Name:        room.Name,
		Description: room.Description,
		UserID:      botID, // Bot manages the encounter
	}
	
	enc, err := h.services.EncounterService.CreateEncounter(context.Background(), encounterInput)
	if err != nil {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("❌ Failed to create encounter: %v", err),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}
	
	// Add all party members to encounter
	for userID, member := range sess.Members {
		if member.CharacterID != "" {
			_, err = h.services.EncounterService.AddPlayer(context.Background(), enc.ID, userID, member.CharacterID)
			if err != nil {
				// Log but continue
				fmt.Printf("Failed to add player %s: %v\n", userID, err)
			}
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
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("❌ Failed to roll initiative: %v", err),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}
	
	// Start combat
	err = h.services.EncounterService.StartEncounter(context.Background(), enc.ID, botID)
	if err != nil {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("❌ Failed to start combat: %v", err),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}
	
	// Get updated encounter
	enc, _ = h.services.EncounterService.GetEncounter(context.Background(), enc.ID)
	
	// Build combat display
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("⚔️ Combat: %s", room.Name),
		Description: "The party engages in battle!",
		Color:       0xe74c3c, // Red
		Fields:      []*discordgo.MessageEmbedField{},
	}
	
	// Show enemies
	var enemyList strings.Builder
	for _, combatant := range enc.Combatants {
		if combatant.Type == entities.CombatantTypeMonster {
			enemyList.WriteString(fmt.Sprintf("• **%s** (HP: %d, AC: %d)\n", combatant.Name, combatant.MaxHP, combatant.AC))
		}
	}
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "🐉 Enemies",
		Value:  enemyList.String(),
		Inline: false,
	})
	
	// Show turn order
	var turnOrder strings.Builder
	for i, combatantID := range enc.TurnOrder {
		if combatant, exists := enc.Combatants[combatantID]; exists {
			prefix := "  "
			if i == 0 {
				prefix = "▶️"
			}
			turnOrder.WriteString(fmt.Sprintf("%s %s\n", prefix, combatant.Name))
		}
	}
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "📋 Initiative Order",
		Value:  turnOrder.String(),
		Inline: false,
	})
	
	// Combat buttons
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
					Label:    "Status",
					Style:    discordgo.SecondaryButton,
					CustomID: fmt.Sprintf("encounter:view:%s", enc.ID),
					Emoji:    &discordgo.ComponentEmoji{Name: "📊"},
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