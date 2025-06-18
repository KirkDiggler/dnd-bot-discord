package dungeon

import (
	"context"
	"fmt"
	"log"
	"math/rand"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/session"
	"github.com/bwmarrin/discordgo"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type StartDungeonRequest struct {
	Session     *discordgo.Session
	Interaction *discordgo.InteractionCreate
	Difficulty  string // easy, medium, hard
}

type StartDungeonHandler struct {
	services *services.Provider
}

func NewStartDungeonHandler(serviceProvider *services.Provider) *StartDungeonHandler {
	return &StartDungeonHandler{
		services: serviceProvider,
	}
}

// RoomType represents different types of dungeon rooms
type RoomType string

const (
	RoomTypeCombat   RoomType = "combat"
	RoomTypePuzzle   RoomType = "puzzle"
	RoomTypeTreasure RoomType = "treasure"
	RoomTypeTrap     RoomType = "trap"
	RoomTypeRest     RoomType = "rest"
)

// Room represents a dungeon room
type Room struct {
	Type        RoomType
	Name        string
	Description string
	Completed   bool
	Monsters    []string
	Treasure    []string
	Challenge   string
}

func (h *StartDungeonHandler) Handle(req *StartDungeonRequest) error {
	// Defer acknowledge the interaction
	err := req.Session.InteractionRespond(req.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		return fmt.Errorf("failed to acknowledge interaction: %w", err)
	}

	// Check if user is already in an active session
	activeSessions, err := h.services.SessionService.ListActiveUserSessions(context.Background(), req.Interaction.Member.User.ID)
	if err == nil && len(activeSessions) > 0 {
		// Leave any existing sessions first
		for _, activeSession := range activeSessions {
			err = h.services.SessionService.LeaveSession(context.Background(), activeSession.ID, req.Interaction.Member.User.ID)
			if err != nil {
				log.Printf("Error leaving session %s: %v", activeSession.ID, err)
			}
		}
	}

	// Create a cooperative session (no DM required)
	sessionInput := &session.CreateSessionInput{
		Name:        "Dungeon Delve",
		Description: "Cooperative dungeon exploration",
		CreatorID:   req.Interaction.Member.User.ID, // User creates it
		RealmID:     req.Interaction.GuildID,
		ChannelID:   req.Interaction.ChannelID,
	}

	sess, err := h.services.SessionService.CreateSession(context.Background(), sessionInput)
	if err != nil {
		content := fmt.Sprintf("❌ Failed to create dungeon session: %v", err)
		_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return err
	}

	// Add the bot as DM for dungeon mode
	botID := req.Session.State.User.ID
	log.Printf("Adding bot %s as DM to session %s", botID, sess.ID)
	sess.AddMember(botID, entities.SessionRoleDM)
	sess.DMID = botID // Set bot as the DM

	// Creator is automatically added as DM, but for dungeon mode we want them as a player
	// Update their role to player
	if member, exists := sess.Members[req.Interaction.Member.User.ID]; exists {
		member.Role = entities.SessionRolePlayer
	}

	// Log session members
	log.Printf("Session members after modification:")
	for userID, member := range sess.Members {
		log.Printf("  - User %s: Role=%s", userID, member.Role)
	}

	// Get user's active character and select it
	var characterName string
	chars, err := h.services.CharacterService.ListByOwner(req.Interaction.Member.User.ID)
	if err == nil && len(chars) > 0 {
		// Find first active character
		for _, char := range chars {
			if char.Status == entities.CharacterStatusActive {
				// Select this character for the session
				err = h.services.SessionService.SelectCharacter(context.Background(), sess.ID, req.Interaction.Member.User.ID, char.ID)
				if err != nil {
					log.Printf("Warning: Failed to auto-select character: %v", err)
				} else {
					characterName = char.Name
				}
				break
			}
		}
	}

	// Generate first room
	room := h.generateRoom(req.Difficulty, 1)

	// Build the dungeon entrance message
	embed := &discordgo.MessageEmbed{
		Title:       "🏰 Dungeon Delve Started!",
		Description: "A cooperative adventure awaits! Work together to explore the dungeon.",
		Color:       0x9b59b6, // Purple
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "📍 Current Room",
				Value:  fmt.Sprintf("**%s**\n%s", room.Name, room.Description),
				Inline: false,
			},
			{
				Name:   "🎯 Objective",
				Value:  h.getRoomObjective(room),
				Inline: false,
			},
			{
				Name:   "👥 Party",
				Value:  h.formatPartyMember(req.Interaction.Member.User.ID, characterName),
				Inline: true,
			},
			{
				Name:   "🏆 Difficulty",
				Value:  cases.Title(language.English).String(req.Difficulty),
				Inline: true,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Other players can join with the buttons below!",
		},
	}

	// Add action buttons based on room type
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Join Party",
					Style:    discordgo.SuccessButton,
					CustomID: fmt.Sprintf("dungeon:join:%s", sess.ID),
					Emoji:    &discordgo.ComponentEmoji{Name: "🤝"},
				},
				discordgo.Button{
					Label:    "Enter Room",
					Style:    discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("dungeon:enter:%s:%s", sess.ID, room.Type),
					Emoji:    &discordgo.ComponentEmoji{Name: "🚪"},
				},
				discordgo.Button{
					Label:    "Party Status",
					Style:    discordgo.SecondaryButton,
					CustomID: fmt.Sprintf("dungeon:status:%s", sess.ID),
					Emoji:    &discordgo.ComponentEmoji{Name: "📊"},
				},
			},
		},
	}

	// Store room data in session metadata
	sess.Metadata = map[string]interface{}{
		"currentRoom": room,
		"roomNumber":  1,
		"difficulty":  req.Difficulty,
	}

	log.Printf("Dungeon started with difficulty: %s, bot ID: %s as DM", req.Difficulty, botID)

	// Save the session with the bot as DM and metadata
	err = h.services.SessionService.SaveSession(context.Background(), sess)
	if err != nil {
		log.Printf("Warning: Failed to save session updates: %v", err)
	}

	_, err = req.Session.InteractionResponseEdit(req.Interaction.Interaction, &discordgo.WebhookEdit{
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
	})
	return err
}

// generateRoom creates a random room based on difficulty and room number
func (h *StartDungeonHandler) generateRoom(difficulty string, roomNumber int) *Room {
	// First room is always combat to start the adventure
	if roomNumber == 1 {
		return h.generateCombatRoom(difficulty, roomNumber)
	}

	// Room type probabilities
	roomTypes := []RoomType{
		RoomTypeCombat,
		RoomTypeCombat,
		RoomTypeCombat, // Higher chance for combat
		RoomTypePuzzle,
		RoomTypeTrap,
		RoomTypeRest,
	}

	// Special rooms every 5 rooms
	if roomNumber%5 == 0 {
		roomTypes = append(roomTypes, RoomTypeTreasure, RoomTypeTreasure)
	}

	roomType := roomTypes[rand.Intn(len(roomTypes))]

	switch roomType {
	case RoomTypeCombat:
		return h.generateCombatRoom(difficulty, roomNumber)
	case RoomTypePuzzle:
		return h.generatePuzzleRoom(difficulty, roomNumber)
	case RoomTypeTrap:
		return h.generateTrapRoom(difficulty, roomNumber)
	case RoomTypeTreasure:
		return h.generateTreasureRoom(difficulty, roomNumber)
	case RoomTypeRest:
		return h.generateRestRoom(roomNumber)
	default:
		return h.generateCombatRoom(difficulty, roomNumber)
	}
}

// generateCombatRoom creates a combat encounter room
func (h *StartDungeonHandler) generateCombatRoom(difficulty string, roomNumber int) *Room {
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

// generatePuzzleRoom creates a puzzle room
func (h *StartDungeonHandler) generatePuzzleRoom(difficulty string, roomNumber int) *Room {
	puzzles := []struct {
		name        string
		description string
		challenge   string
	}{
		{
			"Hall of Riddles",
			"Ancient runes glow on the walls. A stone door bars your way forward.",
			"Solve the riddle: 'I have cities, but no houses. Mountains, but no trees. Water, but no fish. What am I?'",
		},
		{
			"Pressure Plate Puzzle",
			"The floor is covered in stone tiles, each marked with a different symbol.",
			"Step on the correct sequence of tiles to open the door. Look for the pattern!",
		},
		{
			"Mirror Chamber",
			"Mirrors line every wall, reflecting infinite versions of yourselves.",
			"Find the one mirror that shows the truth to reveal the exit.",
		},
	}

	selected := puzzles[rand.Intn(len(puzzles))]

	return &Room{
		Type:        RoomTypePuzzle,
		Name:        selected.name,
		Description: selected.description,
		Challenge:   selected.challenge,
	}
}

// generateTrapRoom creates a trap room
func (h *StartDungeonHandler) generateTrapRoom(difficulty string, roomNumber int) *Room {
	traps := []struct {
		name        string
		description string
		challenge   string
	}{
		{
			"Spike Pit Corridor",
			"A narrow hallway stretches before you. The floor looks suspiciously worn.",
			"Navigate carefully! Make DEX saves to avoid the hidden pits.",
		},
		{
			"Dart Gallery",
			"Small holes pepper the walls. The air smells of ancient poison.",
			"Time your movements to avoid the poison darts!",
		},
		{
			"Crushing Walls",
			"The walls begin to close in as you enter. Ancient gears groan to life.",
			"Find the lever to stop the walls before it's too late!",
		},
	}

	selected := traps[rand.Intn(len(traps))]

	return &Room{
		Type:        RoomTypeTrap,
		Name:        selected.name,
		Description: selected.description,
		Challenge:   selected.challenge,
	}
}

// generateTreasureRoom creates a treasure room
func (h *StartDungeonHandler) generateTreasureRoom(difficulty string, roomNumber int) *Room {
	return &Room{
		Type:        RoomTypeTreasure,
		Name:        "Treasury Vault",
		Description: "Golden light spills from overflowing chests. Ancient artifacts line the shelves.",
		Challenge:   "Claim your rewards! But choose wisely...",
		Treasure:    []string{"Gold coins", "Healing potions", "Magic weapon", "Ancient tome"},
	}
}

// generateRestRoom creates a safe rest area
func (h *StartDungeonHandler) generateRestRoom(roomNumber int) *Room {
	return &Room{
		Type:        RoomTypeRest,
		Name:        "Safe Haven",
		Description: "A peaceful chamber with a fountain of clear water. You can rest here safely.",
		Challenge:   "Take a short rest to recover HP and prepare for the challenges ahead.",
	}
}

// getRoomObjective returns the objective text for a room
func (h *StartDungeonHandler) getRoomObjective(room *Room) string {
	switch room.Type {
	case RoomTypeCombat:
		return fmt.Sprintf("⚔️ Combat: %s", room.Challenge)
	case RoomTypePuzzle:
		return fmt.Sprintf("🧩 Puzzle: %s", room.Challenge)
	case RoomTypeTrap:
		return fmt.Sprintf("⚠️ Trap: %s", room.Challenge)
	case RoomTypeTreasure:
		return fmt.Sprintf("💰 Treasure: %s", room.Challenge)
	case RoomTypeRest:
		return fmt.Sprintf("💤 Rest: %s", room.Challenge)
	default:
		return room.Challenge
	}
}

// formatPartyMember formats a party member display
func (h *StartDungeonHandler) formatPartyMember(userID, characterName string) string {
	if characterName != "" {
		return fmt.Sprintf("<@%s> - %s", userID, characterName)
	}
	return fmt.Sprintf("<@%s> (no character selected)", userID)
}
