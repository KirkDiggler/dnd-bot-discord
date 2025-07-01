package dungeon

//go:generate mockgen -destination=mock/mock_service.go -package=mockdungeon -source=service.go

import (
	"context"
	"fmt"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/exploration"
	"math/rand"
	"time"

	dnderr "github.com/KirkDiggler/dnd-bot-discord/internal/errors"
	"github.com/KirkDiggler/dnd-bot-discord/internal/repositories/dungeons"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/encounter"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/loot"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/monster"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/session"
	"github.com/KirkDiggler/dnd-bot-discord/internal/uuid"
)

// Repository is an alias for the dungeon repository interface
type Repository = dungeons.Repository

// Service defines the dungeon service interface
type Service interface {
	// CreateDungeon creates a new dungeon instance
	CreateDungeon(ctx context.Context, input *CreateDungeonInput) (*exploration.Dungeon, error)

	// GetDungeon retrieves a dungeon by ID
	GetDungeon(ctx context.Context, dungeonID string) (*exploration.Dungeon, error)

	// JoinDungeon adds a player to the dungeon party
	JoinDungeon(ctx context.Context, dungeonID, userID, characterID string) error

	// EnterRoom handles entering the current room
	EnterRoom(ctx context.Context, dungeonID string) (*exploration.DungeonRoom, error)

	// CompleteRoom marks the current room as completed
	CompleteRoom(ctx context.Context, dungeonID string) error

	// ProceedToNextRoom generates and moves to the next room
	ProceedToNextRoom(ctx context.Context, dungeonID string) (*exploration.DungeonRoom, error)

	// GetAvailableActions returns actions available in current state
	GetAvailableActions(ctx context.Context, dungeonID string) ([]DungeonAction, error)

	// AbandonDungeon ends the dungeon run
	AbandonDungeon(ctx context.Context, dungeonID string) error
}

// CreateDungeonInput contains data for creating a dungeon
type CreateDungeonInput struct {
	SessionID  string
	Difficulty string
	CreatorID  string
}

// DungeonAction represents an action players can take
type DungeonAction struct {
	ID          string
	Label       string
	Description string
	Available   bool
	RequiresAll bool // all party members must agree
}

// service implements the Service interface
type service struct {
	repository       Repository
	sessionService   session.Service
	encounterService encounter.Service
	monsterService   monster.Service
	lootService      loot.Service
	uuidGenerator    uuid.Generator
	random           *rand.Rand
}

// ServiceConfig holds configuration for the service
type ServiceConfig struct {
	Repository       Repository        // Required
	SessionService   session.Service   // Required
	EncounterService encounter.Service // Required
	MonsterService   monster.Service   // Optional (will use hardcoded if nil)
	LootService      loot.Service      // Optional (will use hardcoded if nil)
	UUIDGenerator    uuid.Generator    // Optional
}

// NewService creates a new dungeon service
func NewService(cfg *ServiceConfig) Service {
	if cfg.Repository == nil {
		panic("repository is required")
	}
	if cfg.SessionService == nil {
		panic("session service is required")
	}
	if cfg.EncounterService == nil {
		panic("encounter service is required")
	}

	svc := &service{
		repository:       cfg.Repository,
		sessionService:   cfg.SessionService,
		encounterService: cfg.EncounterService,
		monsterService:   cfg.MonsterService,
		lootService:      cfg.LootService,
		random:           rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	// Use provided UUID generator or create default
	if cfg.UUIDGenerator != nil {
		svc.uuidGenerator = cfg.UUIDGenerator
	} else {
		svc.uuidGenerator = uuid.NewGoogleUUIDGenerator()
	}

	return svc
}

// CreateDungeon creates a new dungeon instance
func (s *service) CreateDungeon(ctx context.Context, input *CreateDungeonInput) (*exploration.Dungeon, error) {
	if input == nil {
		return nil, dnderr.InvalidArgument("input cannot be nil")
	}

	// Validate input
	if input.SessionID == "" {
		return nil, dnderr.InvalidArgument("session ID is required")
	}
	if input.Difficulty == "" {
		input.Difficulty = "medium"
	}

	// Validate difficulty
	switch input.Difficulty {
	case "easy", "medium", "hard":
		// valid
	default:
		return nil, dnderr.InvalidArgument("difficulty must be easy, medium, or hard")
	}

	// Generate dungeon ID
	dungeonID := s.uuidGenerator.New()

	// Generate first room
	room := s.generateRoom(input.Difficulty, 1)

	// Create dungeon
	dungeon := &exploration.Dungeon{
		ID:           dungeonID,
		SessionID:    input.SessionID,
		State:        exploration.DungeonStateAwaitingParty,
		CurrentRoom:  room,
		RoomNumber:   1,
		Difficulty:   input.Difficulty,
		Party:        []exploration.PartyMember{},
		RoomsCleared: 0,
		CreatedAt:    time.Now(),
	}

	// Save to repository
	if err := s.repository.Create(ctx, dungeon); err != nil {
		return nil, dnderr.Wrap(err, "failed to create dungeon").
			WithMeta("dungeon_id", dungeonID)
	}

	// Also update session metadata for backward compatibility
	sess, err := s.sessionService.GetSession(ctx, input.SessionID)
	if err != nil {
		return nil, dnderr.Wrap(err, "failed to get session")
	}

	if sess.Metadata == nil {
		sess.Metadata = make(map[string]interface{})
	}
	sess.Metadata["dungeonID"] = dungeonID
	sess.Metadata["currentRoom"] = room
	sess.Metadata["roomNumber"] = 1
	sess.Metadata["difficulty"] = input.Difficulty

	if err := s.sessionService.SaveSession(ctx, sess); err != nil {
		// Log but don't fail
		fmt.Printf("Warning: Failed to update session metadata: %v\n", err)
	}

	return dungeon, nil
}

// GetDungeon retrieves a dungeon by ID
func (s *service) GetDungeon(ctx context.Context, dungeonID string) (*exploration.Dungeon, error) {
	if dungeonID == "" {
		return nil, dnderr.InvalidArgument("dungeon ID is required")
	}

	dungeon, err := s.repository.Get(ctx, dungeonID)
	if err != nil {
		return nil, dnderr.Wrapf(err, "failed to get dungeon '%s'", dungeonID).
			WithMeta("dungeon_id", dungeonID)
	}

	return dungeon, nil
}

// JoinDungeon adds a player to the dungeon party
func (s *service) JoinDungeon(ctx context.Context, dungeonID, userID, characterID string) error {
	if dungeonID == "" {
		return dnderr.InvalidArgument("dungeon ID is required")
	}
	if userID == "" {
		return dnderr.InvalidArgument("user ID is required")
	}
	if characterID == "" {
		return dnderr.InvalidArgument("character ID is required")
	}

	dungeon, err := s.repository.Get(ctx, dungeonID)
	if err != nil {
		return dnderr.Wrapf(err, "failed to get dungeon '%s'", dungeonID)
	}

	// Check state
	if dungeon.State != exploration.DungeonStateAwaitingParty && dungeon.State != exploration.DungeonStateRoomReady {
		return dnderr.InvalidArgument("cannot join dungeon in current state").
			WithMeta("state", string(dungeon.State))
	}

	// Check if already in party
	for _, member := range dungeon.Party {
		if member.UserID == userID {
			return dnderr.InvalidArgument("user already in party")
		}
	}

	// Add to party
	dungeon.Party = append(dungeon.Party, exploration.PartyMember{
		UserID:      userID,
		CharacterID: characterID,
		Status:      "active",
	})

	// Update state if this is the first member
	if len(dungeon.Party) == 1 {
		dungeon.State = exploration.DungeonStateRoomReady
	}

	// Save
	return s.repository.Update(ctx, dungeon)
}

// generateRoom creates a room based on difficulty and room number
func (s *service) generateRoom(difficulty string, roomNumber int) *exploration.DungeonRoom {
	// First room is always combat to start the adventure
	if roomNumber == 1 {
		return s.generateCombatRoom(difficulty, roomNumber)
	}

	// Room type probabilities
	roomTypes := []exploration.RoomType{
		exploration.RoomTypeCombat,
		exploration.RoomTypeCombat,
		exploration.RoomTypeCombat, // Higher chance for combat
		exploration.RoomTypePuzzle,
		exploration.RoomTypeTrap,
		exploration.RoomTypeRest,
	}

	// Special rooms every 5 rooms
	if roomNumber%5 == 0 {
		roomTypes = append(roomTypes, exploration.RoomTypeTreasure, exploration.RoomTypeTreasure)
	}

	roomType := roomTypes[s.random.Intn(len(roomTypes))]

	switch roomType {
	case exploration.RoomTypeCombat:
		return s.generateCombatRoom(difficulty, roomNumber)
	case exploration.RoomTypePuzzle:
		return s.generatePuzzleRoom(difficulty, roomNumber)
	case exploration.RoomTypeTrap:
		return s.generateTrapRoom(difficulty, roomNumber)
	case exploration.RoomTypeTreasure:
		return s.generateTreasureRoom(difficulty, roomNumber)
	case exploration.RoomTypeRest:
		return s.generateRestRoom(roomNumber)
	default:
		return s.generateCombatRoom(difficulty, roomNumber)
	}
}

// generateCombatRoom creates a combat encounter room
func (s *service) generateCombatRoom(difficulty string, roomNumber int) *exploration.DungeonRoom {
	rooms := []struct {
		name        string
		description string
	}{
		{"Guard Chamber", "Stone walls echo with the sounds of movement. Weapons glint in the torchlight."},
		{"Ancient Crypt", "Dusty sarcophagi line the walls. Something stirs in the darkness."},
		{"Goblin Warren", "The stench is overwhelming. Crude weapons and bones litter the floor."},
		{"Spider's Den", "Thick webs cover every surface. Multiple eyes gleam from the shadows."},
	}

	selected := rooms[s.random.Intn(len(rooms))]

	// Determine number of monsters based on difficulty and room number
	var baseCount int
	switch difficulty {
	case "easy":
		baseCount = 1
	case "medium":
		baseCount = 2
	case "hard":
		baseCount = 3
	default:
		baseCount = 2
	}

	// Scale with room number
	extraMonsters := roomNumber / 3
	totalCount := baseCount + extraMonsters

	// Use monster service if available, otherwise fallback to hardcoded
	var monsters []string
	if s.monsterService != nil {
		// Try to get dynamic monsters from the API
		ctx := context.Background()
		monsterTemplates, err := s.monsterService.GetRandomMonsters(ctx, difficulty, totalCount)
		if err == nil && len(monsterTemplates) > 0 {
			for _, template := range monsterTemplates {
				monsters = append(monsters, template.Key)
			}
		}
	}

	// Fallback to hardcoded monsters if monster service failed or unavailable
	if len(monsters) == 0 {
		// Hardcoded fallback
		switch difficulty {
		case "easy":
			monsters = []string{"goblin", "skeleton"}
		case "medium":
			monsters = []string{"orc", "goblin", "goblin"}
		case "hard":
			monsters = []string{"orc", "dire-wolf", "skeleton", "skeleton"}
		default:
			monsters = []string{"goblin"}
		}

		// Ensure we have the right number of monsters
		for len(monsters) < totalCount {
			monsters = append(monsters, monsters[s.random.Intn(len(monsters))])
		}
		if len(monsters) > totalCount {
			monsters = monsters[:totalCount]
		}
	}

	return &exploration.DungeonRoom{
		Type:        exploration.RoomTypeCombat,
		Name:        selected.name,
		Description: selected.description,
		Completed:   false,
		Monsters:    monsters,
		Challenge:   fmt.Sprintf("Defeat all %d enemies!", len(monsters)),
	}
}

// generatePuzzleRoom creates a puzzle room
func (s *service) generatePuzzleRoom(difficulty string, roomNumber int) *exploration.DungeonRoom {
	return &exploration.DungeonRoom{
		Type:        exploration.RoomTypePuzzle,
		Name:        "The Riddler's Chamber",
		Description: "Ancient runes glow on the walls. A voice echoes: 'Answer wisely or face the consequences.'",
		Completed:   false,
		Challenge:   "Solve the ancient riddle to proceed.",
	}
}

// generateTrapRoom creates a trap room
func (s *service) generateTrapRoom(difficulty string, roomNumber int) *exploration.DungeonRoom {
	return &exploration.DungeonRoom{
		Type:        exploration.RoomTypeTrap,
		Name:        "Hall of Dangers",
		Description: "The floor tiles look suspicious. Strange holes dot the walls.",
		Completed:   false,
		Challenge:   "Navigate the trapped hallway safely.",
	}
}

// generateTreasureRoom creates a treasure room
func (s *service) generateTreasureRoom(difficulty string, roomNumber int) *exploration.DungeonRoom {
	// Generate treasure using loot service if available
	var treasure []string
	if s.lootService != nil {
		ctx := context.Background()
		generatedTreasure, err := s.lootService.GenerateTreasure(ctx, difficulty, roomNumber)
		if err == nil && len(generatedTreasure) > 0 {
			treasure = generatedTreasure
		}
	}

	// Fallback to hardcoded treasure if loot service failed or unavailable
	if len(treasure) == 0 {
		treasure = []string{"gold", "healing potion", "mysterious artifact"}
	}

	return &exploration.DungeonRoom{
		Type:        exploration.RoomTypeTreasure,
		Name:        "Treasury Vault",
		Description: "Chests and artifacts fill the room. Gold coins glitter in piles.",
		Completed:   false,
		Treasure:    treasure,
		Challenge:   "Claim your rewards!",
	}
}

// generateRestRoom creates a rest room
func (s *service) generateRestRoom(roomNumber int) *exploration.DungeonRoom {
	return &exploration.DungeonRoom{
		Type:        exploration.RoomTypeRest,
		Name:        "Safe Haven",
		Description: "A peaceful chamber with fresh water and comfortable bedrolls.",
		Completed:   false,
		Challenge:   "Rest and recover before continuing.",
	}
}

// EnterRoom handles entering the current room
func (s *service) EnterRoom(ctx context.Context, dungeonID string) (*exploration.DungeonRoom, error) {
	dungeon, err := s.GetDungeon(ctx, dungeonID)
	if err != nil {
		return nil, err
	}

	// Check if can enter
	if !dungeon.CanEnterRoom() {
		return nil, dnderr.InvalidArgument("cannot enter room in current state").
			WithMeta("state", string(dungeon.State))
	}

	// Update state
	dungeon.State = exploration.DungeonStateInProgress

	// Handle room-specific logic
	switch dungeon.CurrentRoom.Type {
	case exploration.RoomTypeCombat:
		// Create encounter will be handled by the handler
		// Just update state here
	case exploration.RoomTypePuzzle:
		// Puzzle logic would go here
	case exploration.RoomTypeTrap:
		// Trap logic would go here
	case exploration.RoomTypeTreasure:
		// Mark as completed immediately for treasure rooms
		dungeon.CurrentRoom.Completed = true
		dungeon.State = exploration.DungeonStateRoomCleared
	case exploration.RoomTypeRest:
		// Rest logic would go here
	}

	if err := s.repository.Update(ctx, dungeon); err != nil {
		return nil, err
	}

	return dungeon.CurrentRoom, nil
}

// CompleteRoom marks the current room as completed
func (s *service) CompleteRoom(ctx context.Context, dungeonID string) error {
	dungeon, err := s.GetDungeon(ctx, dungeonID)
	if err != nil {
		return err
	}

	if dungeon.State != exploration.DungeonStateInProgress {
		return dnderr.InvalidArgument("room is not in progress")
	}

	dungeon.CurrentRoom.Completed = true
	dungeon.State = exploration.DungeonStateRoomCleared
	dungeon.RoomsCleared++

	return s.repository.Update(ctx, dungeon)
}

// ProceedToNextRoom generates and moves to the next room
func (s *service) ProceedToNextRoom(ctx context.Context, dungeonID string) (*exploration.DungeonRoom, error) {
	dungeon, err := s.GetDungeon(ctx, dungeonID)
	if err != nil {
		return nil, err
	}

	if !dungeon.CanProceed() {
		return nil, dnderr.InvalidArgument("cannot proceed to next room").
			WithMeta("state", string(dungeon.State))
	}

	// Generate next room
	dungeon.RoomNumber++
	dungeon.CurrentRoom = s.generateRoom(dungeon.Difficulty, dungeon.RoomNumber)
	dungeon.State = exploration.DungeonStateRoomReady

	// Check for completion (e.g., after 10 rooms)
	if dungeon.RoomNumber > 10 {
		dungeon.State = exploration.DungeonStateComplete
		now := time.Now()
		dungeon.CompletedAt = &now
	}

	if err := s.repository.Update(ctx, dungeon); err != nil {
		return nil, err
	}

	return dungeon.CurrentRoom, nil
}

// GetAvailableActions returns actions available in current state
func (s *service) GetAvailableActions(ctx context.Context, dungeonID string) ([]DungeonAction, error) {
	dungeon, err := s.GetDungeon(ctx, dungeonID)
	if err != nil {
		return nil, err
	}

	var actions []DungeonAction

	switch dungeon.State {
	case exploration.DungeonStateAwaitingParty:
		actions = append(actions, DungeonAction{
			ID:          "join",
			Label:       "Join Party",
			Description: "Join the dungeon expedition",
			Available:   true,
		})

	case exploration.DungeonStateRoomReady:
		actions = append(actions, DungeonAction{
			ID:          "enter",
			Label:       "Enter Room",
			Description: "Enter the current room",
			Available:   true,
		})

	case exploration.DungeonStateRoomCleared:
		actions = append(actions, DungeonAction{
			ID:          "proceed",
			Label:       "Next Room",
			Description: "Proceed to the next room",
			Available:   true,
		}, DungeonAction{
			ID:          "rest",
			Label:       "Take a Rest",
			Description: "Rest and recover before continuing",
			Available:   true,
		})

	case exploration.DungeonStateComplete:
		actions = append(actions, DungeonAction{
			ID:          "claim",
			Label:       "Claim Rewards",
			Description: "Claim your dungeon completion rewards",
			Available:   true,
		})
	}

	// Always available actions
	if dungeon.IsActive() {
		actions = append(actions, DungeonAction{
			ID:          "status",
			Label:       "Party Status",
			Description: "View party status and statistics",
			Available:   true,
		}, DungeonAction{
			ID:          "abandon",
			Label:       "Abandon Dungeon",
			Description: "Give up and leave the dungeon",
			Available:   true,
			RequiresAll: true,
		})
	}

	return actions, nil
}

// AbandonDungeon ends the dungeon run
func (s *service) AbandonDungeon(ctx context.Context, dungeonID string) error {
	dungeon, err := s.GetDungeon(ctx, dungeonID)
	if err != nil {
		return err
	}

	if !dungeon.IsActive() {
		return dnderr.InvalidArgument("dungeon is not active")
	}

	dungeon.State = exploration.DungeonStateFailed
	now := time.Now()
	dungeon.CompletedAt = &now

	return s.repository.Update(ctx, dungeon)
}
