package encounter

//go:generate mockgen -destination=mock/mock_service.go -package=mockencounter -source=service.go

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/adapters/rpgtoolkit"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/damage"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/equipment"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/combat"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/combat/attack"
	gameSession "github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/session"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e/feats"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"github.com/KirkDiggler/rpg-toolkit/core"
	rpgevents "github.com/KirkDiggler/rpg-toolkit/events"

	"github.com/KirkDiggler/dnd-bot-discord/internal/dice"
	"github.com/KirkDiggler/dnd-bot-discord/internal/effects"
	dnderr "github.com/KirkDiggler/dnd-bot-discord/internal/errors"
	"github.com/KirkDiggler/dnd-bot-discord/internal/repositories/encounters"
	charService "github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	sessService "github.com/KirkDiggler/dnd-bot-discord/internal/services/session"
	"github.com/KirkDiggler/dnd-bot-discord/internal/uuid"
)

// Service defines the encounter service interface
type Service interface {
	// CreateEncounter creates a new encounter in a session
	CreateEncounter(ctx context.Context, input *CreateEncounterInput) (*combat.Encounter, error)

	// GetEncounter retrieves an encounter by ID
	GetEncounter(ctx context.Context, encounterID string) (*combat.Encounter, error)

	// GetActiveEncounter retrieves the active encounter for a session
	GetActiveEncounter(ctx context.Context, sessionID string) (*combat.Encounter, error)

	// AddMonster adds a monster to an encounter
	AddMonster(ctx context.Context, encounterID, userID string, input *AddMonsterInput) (*combat.Combatant, error)

	// AddPlayer adds a player character to an encounter
	AddPlayer(ctx context.Context, encounterID, playerID, characterID string) (*combat.Combatant, error)

	// RemoveCombatant removes a combatant from an encounter
	RemoveCombatant(ctx context.Context, encounterID, combatantID, userID string) error

	// RollInitiative rolls initiative for all combatants
	RollInitiative(ctx context.Context, encounterID, userID string) error

	// StartEncounter begins combat
	StartEncounter(ctx context.Context, encounterID, userID string) error

	// NextTurn advances to the next turn
	NextTurn(ctx context.Context, encounterID, userID string) error

	// PerformAttack executes an attack from one combatant to another
	PerformAttack(ctx context.Context, input *AttackInput) (*AttackResult, error)

	// ApplyDamage applies damage to a combatant
	ApplyDamage(ctx context.Context, encounterID, combatantID, userID string, damage int) error

	// HealCombatant heals a combatant
	HealCombatant(ctx context.Context, encounterID, combatantID, userID string, amount int) error

	// EndEncounter ends the encounter
	EndEncounter(ctx context.Context, encounterID, userID string) error

	// LogCombatAction logs a combat action (like a miss) without damage
	LogCombatAction(ctx context.Context, encounterID, action string) error

	// ProcessMonsterTurn handles a monster's turn automatically
	ProcessMonsterTurn(ctx context.Context, encounterID string, monsterID string) (*AttackResult, error)

	// ProcessAllMonsterTurns processes all consecutive monster turns
	ProcessAllMonsterTurns(ctx context.Context, encounterID string) ([]*AttackResult, error)

	// ExecuteAttackWithTarget handles a complete attack sequence including auto-advancing turns
	ExecuteAttackWithTarget(ctx context.Context, input *ExecuteAttackInput) (*ExecuteAttackResult, error)

	// UpdateMessageID updates the Discord message ID for an encounter
	UpdateMessageID(ctx context.Context, encounterID, messageID, channelID string) error
}

// CreateEncounterInput contains data for creating an encounter
type CreateEncounterInput struct {
	SessionID   string
	ChannelID   string
	Name        string
	Description string
	UserID      string
}

// AddMonsterInput contains data for adding a monster
type AddMonsterInput struct {
	Name            string
	MaxHP           int
	AC              int
	Initiative      int
	InitiativeBonus int
	Speed           int
	CR              float64
	XP              int
	MonsterRef      string // D&D API reference
	Abilities       map[string]int
	Actions         []*combat.MonsterAction
}

// AttackInput contains data for performing an attack
type AttackInput struct {
	EncounterID string
	AttackerID  string
	TargetID    string
	UserID      string // User requesting the attack
	ActionIndex int    // For monsters with multiple attacks (0 = first action)

	// Combat modifiers (optional - used for abilities like sneak attack)
	HasAdvantage    bool // Attacker has advantage on this attack
	HasDisadvantage bool // Attacker has disadvantage on this attack
	AllyAdjacent    bool // An ally is within 5 feet of the target (for sneak attack)
}

// AttackResult contains the results of an attack
type AttackResult struct {
	// Roll information
	AttackRoll  int
	AttackBonus int
	TotalAttack int
	DiceRolls   []int // Individual dice rolls for transparency

	// Hit/Miss information
	TargetAC int
	Hit      bool
	Critical bool

	// Damage information
	Damage      int
	DamageType  string
	DamageRolls []int // Individual damage dice rolls
	DamageBonus int

	// Sneak attack information
	SneakAttackDamage int
	SneakAttackDice   int // Number of d6s rolled
	// Weapon damage dice info (for proper display)
	WeaponDiceCount int
	WeaponDiceSize  int

	// Great Weapon Fighting reroll information
	RerollInfo []attack.DieReroll

	// Combatant information
	AttackerName string
	TargetName   string
	WeaponName   string

	// Results
	TargetNewHP    int
	TargetDefeated bool
	CombatEnded    bool
	PlayersWon     bool

	// Combat log entry
	LogEntry string
}

// ExecuteAttackInput contains data for executing a complete attack sequence
type ExecuteAttackInput struct {
	EncounterID string
	TargetID    string
	UserID      string // User executing the attack
}

// ExecuteAttackResult contains results of the complete attack sequence
type ExecuteAttackResult struct {
	// The initial attack result
	PlayerAttack *AttackResult

	// Any monster attacks that followed
	MonsterAttacks []*AttackResult

	// Updated encounter state
	IsPlayerTurn bool
	CurrentTurn  *combat.Combatant
	CombatEnded  bool
	PlayersWon   bool
}

type service struct {
	repository       encounters.Repository
	sessionService   sessService.Service
	characterService charService.Service
	uuidGenerator    uuid.Generator
	diceRoller       dice.Roller
	eventBus         *rpgevents.Bus
}

// ServiceConfig holds configuration for the service
type ServiceConfig struct {
	Repository       encounters.Repository
	SessionService   sessService.Service
	CharacterService charService.Service
	UUIDGenerator    uuid.Generator
	DiceRoller       dice.Roller
	EventBus         *rpgevents.Bus
}

// NewService creates a new encounter service
func NewService(cfg *ServiceConfig) Service {
	if cfg.Repository == nil {
		panic("repository is required")
	}
	if cfg.SessionService == nil {
		panic("session service is required")
	}
	if cfg.CharacterService == nil {
		panic("character service is required")
	}

	svc := &service{
		repository:       cfg.Repository,
		sessionService:   cfg.SessionService,
		characterService: cfg.CharacterService,
		diceRoller:       cfg.DiceRoller,
		eventBus:         cfg.EventBus,
	}

	if cfg.UUIDGenerator != nil {
		svc.uuidGenerator = cfg.UUIDGenerator
	} else {
		svc.uuidGenerator = uuid.NewGoogleUUIDGenerator()
	}

	// Use random dice roller if none provided
	if svc.diceRoller == nil {
		svc.diceRoller = dice.NewRandomRoller()
	}

	// TODO: Register event handlers when they are migrated to rpg-toolkit
	// - Migrate local spell damage handler to use rpg-toolkit events
	// - Migrate status effect handler to rpg-toolkit
	// - Migrate local proficiency handler to use rpg-toolkit events
	// Note: The proficiency bonus is already calculated in character.Attack()
	// These handlers demonstrate event-driven architecture for future enhancements

	return svc
}

// CreateEncounter creates a new encounter in a session
func (s *service) CreateEncounter(ctx context.Context, input *CreateEncounterInput) (*combat.Encounter, error) {
	if input == nil {
		return nil, dnderr.InvalidArgument("input cannot be nil")
	}

	// Validate input
	if strings.TrimSpace(input.Name) == "" {
		return nil, dnderr.InvalidArgument("encounter name is required")
	}

	// Verify session exists and user is DM
	session, err := s.sessionService.GetSession(ctx, input.SessionID)
	if err != nil {
		return nil, dnderr.Wrap(err, "failed to get session")
	}

	// Check if user is DM or system/bot (for dungeon encounters)
	member, exists := session.Members[input.UserID]
	if !exists {
		// Allow system/bot to create encounters for dungeons
		if !session.IsDungeon() {
			return nil, dnderr.PermissionDenied("only the DM can create encounters")
		}
		// For dungeon sessions, bot/system can create encounters
	} else if member.Role != gameSession.SessionRoleDM {
		return nil, dnderr.PermissionDenied("only the DM can create encounters")
	}

	// Check if there's already an active encounter
	activeEncounter, err := s.repository.GetActiveBySession(ctx, input.SessionID)
	if err != nil {
		// It's OK if no active encounter exists - that's what we want
		if !strings.Contains(err.Error(), "no active encounter") {
			return nil, dnderr.Wrap(err, "failed to get active encounter")
		}
		// No active encounter, we can proceed
	} else if activeEncounter != nil {
		return nil, dnderr.InvalidArgument("session already has an active encounter")
	}

	// Create encounter
	encounterID := s.uuidGenerator.New()
	encounter := combat.NewEncounter(encounterID, input.SessionID, input.ChannelID, input.Name, input.UserID)
	encounter.Description = input.Description

	// Save encounter
	if err := s.repository.Create(ctx, encounter); err != nil {
		return nil, dnderr.Wrap(err, "failed to create encounter")
	}

	// Update session with encounter
	session.Encounters = append(session.Encounters, encounterID)
	session.ActiveEncounterID = &encounterID

	// We should update the session through the session service
	// For now, we'll just return the encounter

	return encounter, nil
}

// GetEncounter retrieves an encounter by ID
func (s *service) GetEncounter(ctx context.Context, encounterID string) (*combat.Encounter, error) {
	if strings.TrimSpace(encounterID) == "" {
		return nil, dnderr.InvalidArgument("encounter ID is required")
	}

	encounter, err := s.repository.Get(ctx, encounterID)
	if err != nil {
		return nil, dnderr.Wrapf(err, "failed to get encounter '%s'", encounterID)
	}

	return encounter, nil
}

// GetActiveEncounter retrieves the active encounter for a session
func (s *service) GetActiveEncounter(ctx context.Context, sessionID string) (*combat.Encounter, error) {
	if strings.TrimSpace(sessionID) == "" {
		return nil, dnderr.InvalidArgument("session ID is required")
	}

	encounter, err := s.repository.GetActiveBySession(ctx, sessionID)
	if err != nil {
		return nil, dnderr.Wrapf(err, "no active encounter in session '%s'", sessionID)
	}

	return encounter, nil
}

// AddMonster adds a monster to an encounter
func (s *service) AddMonster(ctx context.Context, encounterID, userID string, input *AddMonsterInput) (*combat.Combatant, error) {
	if input == nil {
		return nil, dnderr.InvalidArgument("input cannot be nil")
	}

	// Get encounter
	encounter, err := s.repository.Get(ctx, encounterID)
	if err != nil {
		return nil, dnderr.Wrap(err, "failed to get encounter")
	}

	// Verify user is DM or system
	if encounter.CreatedBy != userID {
		// Check if user is DM of the session
		session, err := s.sessionService.GetSession(ctx, encounter.SessionID)
		if err != nil {
			return nil, dnderr.Wrap(err, "failed to get session")
		}

		member, exists := session.Members[userID]
		if !exists {
			// Allow system/bot for dungeon sessions
			if !session.IsDungeon() {
				return nil, dnderr.PermissionDenied("only the DM can add monsters")
			}
		} else if member.Role != gameSession.SessionRoleDM {
			return nil, dnderr.PermissionDenied("only the DM can add monsters")
		}
	}

	// Create combatant
	combatantID := s.uuidGenerator.New()
	combatant := &combat.Combatant{
		ID:              combatantID,
		Name:            input.Name,
		Type:            combat.CombatantTypeMonster,
		Initiative:      input.Initiative,
		InitiativeBonus: input.InitiativeBonus,
		CurrentHP:       input.MaxHP,
		MaxHP:           input.MaxHP,
		AC:              input.AC,
		Speed:           input.Speed,
		IsActive:        true,
		MonsterRef:      input.MonsterRef,
		CR:              input.CR,
		XP:              input.XP,
		Abilities:       input.Abilities,
		Actions:         input.Actions,
	}

	// Add to encounter
	encounter.AddCombatant(combatant)

	// Save changes
	if err := s.repository.Update(ctx, encounter); err != nil {
		return nil, dnderr.Wrap(err, "failed to update encounter")
	}

	return combatant, nil
}

// AddPlayer adds a player character to an encounter
func (s *service) AddPlayer(ctx context.Context, encounterID, playerID, characterID string) (*combat.Combatant, error) {
	// Get encounter
	encounter, err := s.repository.Get(ctx, encounterID)
	if err != nil {
		return nil, dnderr.Wrap(err, "failed to get encounter")
	}

	// Get character details
	char, err := s.characterService.GetByID(characterID)
	if err != nil {
		return nil, dnderr.Wrap(err, "failed to get character")
	}

	// Log character details
	// Character retrieved successfully

	// Ensure resources are initialized (lazy initialization)
	resources := char.GetResources()

	// Check if this is a dungeon session - if so, perform a long rest
	// to reset abilities like rage uses and lay on hands
	if encounter.SessionID != "" {
		session, err := s.sessionService.GetSession(ctx, encounter.SessionID)
		if err == nil && session.Metadata != nil {
			if sessionType, ok := session.Metadata["sessionType"].(string); ok && sessionType == "dungeon" {
				// Dungeon session detected, performing long rest
				resources.LongRest()

				// Save the character to persist the reset abilities
				if err := s.characterService.UpdateEquipment(char); err != nil {
					log.Printf("Failed to save character after long rest: %v", err)
					// Continue anyway - the abilities are reset in memory
				}
			}
		}
	}

	// Reset action economy for the start of combat
	char.StartNewTurn()

	// Save character to persist the reset action economy
	if err := s.characterService.UpdateEquipment(char); err != nil {
		log.Printf("Failed to save character after action economy reset: %v", err)
	}

	// Register feat event handlers if event bus is available
	if s.eventBus != nil {
		// Register handlers for all character feats
		feats.GlobalRegistry.RegisterAllHandlers(char, s.eventBus)
	}

	// Verify character belongs to player
	if char.OwnerID != playerID {
		return nil, dnderr.PermissionDenied("character does not belong to player")
	}

	// Check if player is already in encounter
	for _, combatant := range encounter.Combatants {
		if combatant.PlayerID == playerID {
			return nil, dnderr.InvalidArgument("player is already in the encounter")
		}
	}

	// Create combatant from character
	combatantID := s.uuidGenerator.New()

	// Get dexterity modifier for initiative
	dexBonus := 0
	if dexScore, exists := char.Attributes[shared.AttributeDexterity]; exists {
		dexBonus = dexScore.Bonus
	}

	// Get character class and race info
	className := ""
	if char.Class != nil {
		className = char.Class.Name
	}
	raceName := ""
	if char.Race != nil {
		raceName = char.Race.Name
	}

	combatant := &combat.Combatant{
		ID:              combatantID,
		Name:            char.Name,
		Type:            combat.CombatantTypePlayer,
		InitiativeBonus: dexBonus,
		CurrentHP:       char.CurrentHitPoints,
		MaxHP:           char.MaxHitPoints,
		AC:              char.AC,
		Speed:           30, // Default, should come from race
		IsActive:        true,
		PlayerID:        playerID,
		CharacterID:     characterID,
		Class:           className,
		Race:            raceName,
	}

	// Add to encounter
	encounter.AddCombatant(combatant)

	// Save changes
	if err := s.repository.Update(ctx, encounter); err != nil {
		return nil, dnderr.Wrap(err, "failed to update encounter")
	}

	return combatant, nil
}

// RemoveCombatant removes a combatant from an encounter
func (s *service) RemoveCombatant(ctx context.Context, encounterID, combatantID, userID string) error {
	// Get encounter
	encounter, err := s.repository.Get(ctx, encounterID)
	if err != nil {
		return dnderr.Wrap(err, "failed to get encounter")
	}

	// Check permissions
	if !encounter.CanPlayerAct(userID) {
		return dnderr.PermissionDenied("you cannot modify this encounter")
	}

	// Remove combatant
	encounter.RemoveCombatant(combatantID)

	// Save changes
	if err := s.repository.Update(ctx, encounter); err != nil {
		return dnderr.Wrap(err, "failed to update encounter")
	}

	return nil
}

// RollInitiative rolls initiative for all combatants
func (s *service) RollInitiative(ctx context.Context, encounterID, userID string) error {
	// Get encounter
	encounter, err := s.repository.Get(ctx, encounterID)
	if err != nil {
		return dnderr.Wrap(err, "failed to get encounter")
	}

	// Check permissions
	if encounter.CreatedBy != userID {
		// Allow system/bot for dungeon encounters
		session, err := s.sessionService.GetSession(ctx, encounter.SessionID)
		if err != nil {
			return dnderr.Wrap(err, "failed to get session")
		}
		if !session.IsDungeon() {
			return dnderr.PermissionDenied("only the DM can roll initiative")
		}
	}

	// Check status
	if encounter.Status != combat.EncounterStatusSetup {
		return dnderr.InvalidArgument("encounter is not in setup phase")
	}

	// Clear combat log for new initiative rolls
	encounter.CombatLog = []string{"🎲 **Rolling Initiative**"}

	// Roll initiative for each combatant
	// Sort combatant IDs to ensure deterministic order for testing
	combatantIDs := make([]string, 0, len(encounter.Combatants))
	for id := range encounter.Combatants {
		combatantIDs = append(combatantIDs, id)
	}
	sort.Strings(combatantIDs)

	initiatives := make(map[string]int)
	for _, id := range combatantIDs {
		combatant := encounter.Combatants[id]
		result, err := s.diceRoller.Roll(1, 20, combatant.InitiativeBonus)
		if err != nil {
			return dnderr.Wrap(err, "failed to roll initiative")
		}
		combatant.Initiative = result.Total
		initiatives[id] = combatant.Initiative

		// Log the initiative roll
		logEntry := fmt.Sprintf("**%s** rolls initiative: %v + %d = **%d**",
			combatant.Name,
			result.Rolls[0], // The d20 roll
			combatant.InitiativeBonus,
			combatant.Initiative)
		encounter.CombatLog = append(encounter.CombatLog, logEntry)
	}

	// Sort combatants by initiative (descending)
	encounter.TurnOrder = make([]string, 0, len(encounter.Combatants))
	for id := range encounter.Combatants {
		encounter.TurnOrder = append(encounter.TurnOrder, id)
	}

	// Simple bubble sort for now
	for i := 0; i < len(encounter.TurnOrder)-1; i++ {
		for j := 0; j < len(encounter.TurnOrder)-i-1; j++ {
			if initiatives[encounter.TurnOrder[j]] < initiatives[encounter.TurnOrder[j+1]] {
				encounter.TurnOrder[j], encounter.TurnOrder[j+1] = encounter.TurnOrder[j+1], encounter.TurnOrder[j]
			}
		}
	}

	encounter.Status = combat.EncounterStatusRolling

	// Save changes
	if err := s.repository.Update(ctx, encounter); err != nil {
		return dnderr.Wrap(err, "failed to update encounter")
	}

	return nil
}

// StartEncounter begins combat
func (s *service) StartEncounter(ctx context.Context, encounterID, userID string) error {
	// Get encounter
	encounter, err := s.repository.Get(ctx, encounterID)
	if err != nil {
		return dnderr.Wrap(err, "failed to get encounter")
	}

	// Check permissions
	if encounter.CreatedBy != userID {
		// Allow system/bot for dungeon encounters
		session, err := s.sessionService.GetSession(ctx, encounter.SessionID)
		if err != nil {
			return dnderr.Wrap(err, "failed to get session")
		}
		if !session.IsDungeon() {
			return dnderr.PermissionDenied("only the DM can start the encounter")
		}
	}

	// Start encounter
	if !encounter.Start() {
		return dnderr.InvalidArgument("encounter cannot be started")
	}

	// Save changes
	if err := s.repository.Update(ctx, encounter); err != nil {
		return dnderr.Wrap(err, "failed to update encounter")
	}

	return nil
}

// NextTurn advances to the next turn
func (s *service) NextTurn(ctx context.Context, encounterID, userID string) error {
	// Get encounter
	encounter, err := s.repository.Get(ctx, encounterID)
	if err != nil {
		return dnderr.Wrap(err, "failed to get encounter")
	}

	// Check if it's the current player's turn or DM
	current := encounter.GetCurrentCombatant()
	if current == nil {
		return dnderr.InvalidArgument("no active combatant")
	}

	// Check permissions based on encounter type
	if session, err := s.sessionService.GetSession(ctx, encounter.SessionID); err == nil {
		if sessionType, ok := session.Metadata["sessionType"].(string); ok && sessionType == "dungeon" {
			// In dungeon encounters:
			// - Players can advance monster turns
			// - Players can advance their own turn
			// - DM (bot) can advance any turn
			if current.Type == combat.CombatantTypeMonster {
				// Any player can advance monster turns in dungeons
			} else if current.PlayerID != userID && encounter.CreatedBy != userID {
				return dnderr.PermissionDenied("not your turn")
			}
		} else {
			// Regular encounter rules
			if current.PlayerID != userID && encounter.CreatedBy != userID {
				return dnderr.PermissionDenied("not your turn")
			}
		}
	} else {
		// Fallback to regular rules if session lookup fails
		if current.PlayerID != userID && encounter.CreatedBy != userID {
			return dnderr.PermissionDenied("not your turn")
		}
	}

	// Track the previous round
	prevRound := encounter.Round

	// Advance turn
	encounter.NextTurn()

	// Emit OnTurnStart event for duration tracking
	if s.eventBus != nil {
		// Get the new current combatant
		newCurrent := encounter.GetCurrentCombatant()
		if newCurrent != nil && newCurrent.Type == combat.CombatantTypePlayer && newCurrent.CharacterID != "" {
			if char, err := s.characterService.GetByID(newCurrent.CharacterID); err == nil {
				// Calculate total turns (rounds * combatants + current turn index)
				totalTurns := (encounter.Round-1)*len(encounter.TurnOrder) + encounter.Turn
				contextData := map[string]interface{}{
					rpgtoolkit.ContextTurnCount:     totalTurns,
					rpgtoolkit.ContextRound:         encounter.Round,
					rpgtoolkit.ContextNumCombatants: len(encounter.TurnOrder),
				}

				if err := rpgtoolkit.EmitEvent(s.eventBus, rpgevents.EventOnTurnStart, char, nil, contextData); err != nil {
					rpgtoolkit.LogEventError("OnTurnStart", err)
				}
			}
		}
	}

	// Check if a new round started
	if encounter.Round > prevRound {
		// Reset per-turn abilities for all player characters
		for _, combatant := range encounter.Combatants {
			if combatant.Type != combat.CombatantTypePlayer || combatant.CharacterID == "" {
				continue
			}

			// Get the character
			char, err := s.characterService.GetByID(combatant.CharacterID)
			if err != nil {
				// Failed to get character for turn reset - continue anyway
				continue
			}

			// Reset per-turn abilities
			char.StartNewTurn()

			// Save character to persist the reset
			if err := s.characterService.UpdateEquipment(char); err != nil {
				log.Printf("Failed to update character %s after turn reset: %v", char.ID, err)
			}
		}
	}

	// Save changes
	if err := s.repository.Update(ctx, encounter); err != nil {
		return dnderr.Wrap(err, "failed to update encounter")
	}

	return nil
}

// PerformAttack executes an attack from one combatant to another
func (s *service) PerformAttack(ctx context.Context, input *AttackInput) (*AttackResult, error) {
	if input == nil {
		return nil, dnderr.InvalidArgument("input cannot be nil")
	}

	// Get encounter
	encounter, err := s.repository.Get(ctx, input.EncounterID)
	if err != nil {
		return nil, dnderr.Wrap(err, "failed to get encounter")
	}

	// Validate encounter is active
	if encounter.Status != combat.EncounterStatusActive {
		return nil, dnderr.InvalidArgument("encounter is not active")
	}

	// Get attacker
	attacker, exists := encounter.Combatants[input.AttackerID]
	if !exists {
		return nil, dnderr.NotFound("attacker not found")
	}
	if !attacker.IsActive {
		return nil, dnderr.InvalidArgument("attacker is not active")
	}

	// Get target
	target, exists := encounter.Combatants[input.TargetID]
	if !exists {
		return nil, dnderr.NotFound("target not found")
	}
	if !target.IsActive {
		return nil, dnderr.InvalidArgument("target is not active")
	}

	// Check permissions
	current := encounter.GetCurrentCombatant()
	if current == nil || current.ID != input.AttackerID {
		// Special handling for dungeon encounters
		session, err := s.sessionService.GetSession(ctx, encounter.SessionID)
		if err != nil {
			// If we can't get the session, deny the attack for security
			return nil, dnderr.PermissionDenied("unable to verify permissions")
		}

		if sessionType, ok := session.Metadata["sessionType"].(string); ok && sessionType == "dungeon" {
			// In dungeon encounters, bot orchestrates combat
		} else {
			return nil, dnderr.PermissionDenied("not attacker's turn")
		}
	}

	result := &AttackResult{
		AttackerName: attacker.Name,
		TargetName:   target.Name,
		TargetAC:     target.AC,
	}

	// Handle different attacker types
	if attacker.Type == combat.CombatantTypePlayer && attacker.CharacterID != "" {
		// Player attack using character
		char, err := s.characterService.GetByID(attacker.CharacterID)
		if err != nil {
			return nil, dnderr.Wrap(err, "failed to get character")
		}

		// Attack debug logs removed - too verbose during combat

		// Get target character for events
		var targetChar *character.Character
		if target.Type == combat.CombatantTypePlayer && target.CharacterID != "" {
			var getErr error
			targetChar, getErr = s.characterService.GetByID(target.CharacterID)
			if getErr != nil {
				log.Printf("Failed to get target character for events: %v", getErr)
			}
		} else {
			// Create basic character for non-player target
			targetChar = &character.Character{
				ID:   target.ID,
				Name: target.Name,
			}
		}

		// Emit BeforeAttackRoll event
		// TODO: Extract event emission logic into helper function to reduce duplication
		// This pattern is repeated for BeforeAttackRoll, OnAttackRoll, and AfterAttackRoll
		if s.eventBus != nil {
			// Get weapon name for the event
			weaponName := "Unarmed Strike"
			if char.EquippedSlots[shared.SlotMainHand] != nil {
				weaponName = char.EquippedSlots[shared.SlotMainHand].GetName()
			} else if char.EquippedSlots[shared.SlotTwoHanded] != nil {
				weaponName = char.EquippedSlots[shared.SlotTwoHanded].GetName()
			}

			beforeAttackContext := map[string]interface{}{
				"weapon":       weaponName,
				"attack_bonus": 0, // Will be calculated by char.Attack()
			}

			beforeAttackEvent, emitErr := rpgtoolkit.CreateAndEmitEvent(
				s.eventBus,
				rpgevents.EventBeforeAttackRoll,
				char,
				targetChar,
				beforeAttackContext,
			)
			if emitErr != nil {
				rpgtoolkit.LogEventError("BeforeAttackRoll", emitErr)
			}

			// Check if event was cancelled
			if beforeAttackEvent != nil && beforeAttackEvent.IsCancelled() {
				result.Hit = false
				result.Damage = 0
				result.AttackRoll = 0
				result.TotalAttack = 0
				return result, nil
			}
		}

		// Use character's attack method
		attackResults, err := char.Attack()
		if err != nil {
			return nil, dnderr.Wrap(err, "failed to perform character attack")
		}

		if len(attackResults) == 0 {
			return nil, dnderr.InvalidArgument("no attack results")
		}

		// Use first attack result
		attackResult := attackResults[0]

		// Record the attack action for action economy
		// Use the weapon key from the attack result if available
		weaponKey := attackResult.WeaponKey
		if weaponKey == "" {
			// Fallback to equipped weapon if not in result (for backward compatibility)
			if char.EquippedSlots[shared.SlotMainHand] != nil {
				weaponKey = char.EquippedSlots[shared.SlotMainHand].GetKey()
			} else if char.EquippedSlots[shared.SlotTwoHanded] != nil {
				weaponKey = char.EquippedSlots[shared.SlotTwoHanded].GetKey()
			}
		}
		char.RecordAction("attack", "weapon", weaponKey)

		// Save character to persist action economy changes
		if err := s.characterService.UpdateEquipment(char); err != nil {
			log.Printf("Failed to save character after attack action: %v", err)
		}

		// Log action economy state for debugging
		log.Printf("[ACTION ECONOMY] %s attacked with %s - Actions taken: %d, Bonus actions available: %d",
			char.Name, weaponKey, len(char.GetActionsTaken()), len(char.GetAvailableBonusActions()))
		for _, ba := range char.GetAvailableBonusActions() {
			log.Printf("[ACTION ECONOMY] Available bonus action: %s (%s)", ba.Name, ba.Key)
		}
		result.AttackRoll = attackResult.AttackResult.Rolls[0]      // The d20 roll
		result.TotalAttack = attackResult.AttackRoll                // Total including bonuses
		result.AttackBonus = result.TotalAttack - result.AttackRoll // Calculate bonus from total minus d20
		result.DiceRolls = attackResult.AttackResult.Rolls

		// Set weapon damage info
		if attackResult.WeaponDamage != nil {
			result.WeaponDiceCount = attackResult.WeaponDamage.DiceCount
			result.WeaponDiceSize = attackResult.WeaponDamage.DiceSize
		} else {
			// Default to d4 for improvised
			result.WeaponDiceCount = 1
			result.WeaponDiceSize = 4
		}

		// Get weapon name
		if char.EquippedSlots[shared.SlotMainHand] != nil {
			result.WeaponName = char.EquippedSlots[shared.SlotMainHand].GetName()
		} else if char.EquippedSlots[shared.SlotTwoHanded] != nil {
			result.WeaponName = char.EquippedSlots[shared.SlotTwoHanded].GetName()
		} else {
			result.WeaponName = "Unarmed Strike"
		}

		// Emit OnAttackRoll event
		if s.eventBus != nil {
			onAttackContext := map[string]interface{}{
				"weapon":       result.WeaponName,
				"attack_roll":  result.AttackRoll,
				"attack_bonus": result.AttackBonus,
				"total_attack": result.TotalAttack,
				"target_ac":    target.AC,
			}

			onAttackEvent, emitErr := rpgtoolkit.CreateAndEmitEvent(
				s.eventBus,
				rpgevents.EventOnAttackRoll,
				char,
				targetChar,
				onAttackContext,
			)
			if emitErr != nil {
				rpgtoolkit.LogEventError("OnAttackRoll", emitErr)
			}

			// Update attack values from event context if modified
			if onAttackEvent != nil {
				if val, ok := onAttackEvent.Context().Get("total_attack"); ok {
					if modifiedTotal, ok := val.(int); ok {
						result.TotalAttack = modifiedTotal
					}
				}
			}
		}

		// Check hit
		result.Hit = result.TotalAttack >= target.AC
		result.Critical = result.AttackRoll == 20

		// Emit AfterAttackRoll event
		if s.eventBus != nil {
			afterAttackContext := map[string]interface{}{
				"weapon":       result.WeaponName,
				"attack_roll":  result.AttackRoll,
				"attack_bonus": result.AttackBonus,
				"total_attack": result.TotalAttack,
				"target_ac":    target.AC,
				"hit":          result.Hit,
				"critical":     result.Critical,
			}

			afterAttackEvent, emitErr := rpgtoolkit.CreateAndEmitEvent(
				s.eventBus,
				rpgevents.EventAfterAttackRoll,
				char,
				targetChar,
				afterAttackContext,
			)
			if emitErr != nil {
				rpgtoolkit.LogEventError("AfterAttackRoll", emitErr)
			}

			// Update hit/critical from event context if modified
			if afterAttackEvent != nil {
				if val, ok := afterAttackEvent.Context().Get("hit"); ok {
					if modifiedHit, ok := val.(bool); ok {
						result.Hit = modifiedHit
					}
				}
				if val, ok := afterAttackEvent.Context().Get("critical"); ok {
					if modifiedCrit, ok := val.(bool); ok {
						result.Critical = modifiedCrit
					}
				}
			}
		}

		if result.Hit {
			result.Damage = attackResult.DamageRoll
			result.DamageType = string(attackResult.AttackType)
			if attackResult.AllDamageRolls != nil {
				result.DamageRolls = attackResult.AllDamageRolls
				// Calculate dice total from all rolls
				diceTotal := 0
				for _, roll := range attackResult.AllDamageRolls {
					diceTotal += roll
				}
				result.DamageBonus = attackResult.DamageRoll - diceTotal
				// Damage calculation debug logs removed
			}

			// Emit OnDamageRoll event for damage modifiers (like rage and sneak attack)
			if s.eventBus != nil {
				damageContext := map[string]interface{}{
					rpgtoolkit.ContextAttackType: attackResult.AttackType,
					rpgtoolkit.ContextDamage:     result.Damage,
					rpgtoolkit.ContextTargetID:   target.ID,
					rpgtoolkit.ContextIsCritical: result.Critical,
				}

				// Add weapon information for sneak attack
				if attackResult.WeaponKey != "" {
					damageContext[rpgtoolkit.ContextWeaponKey] = attackResult.WeaponKey

					// Check weapon type (ranged vs melee)
					weaponType := "melee"
					if char.EquippedSlots[shared.SlotMainHand] != nil {
						if weapon, ok := char.EquippedSlots[shared.SlotMainHand].(*equipment.Weapon); ok {
							if weapon.WeaponRange == "Ranged" {
								weaponType = "ranged"
							}
							// Check for finesse property
							for _, prop := range weapon.Properties {
								if prop != nil && prop.Key == "finesse" {
									damageContext[rpgtoolkit.ContextWeaponHasFinesse] = true
									break
								}
							}
						}
					}
					damageContext[rpgtoolkit.ContextWeaponType] = weaponType
				}

				// Add combat conditions (these would need to be passed in via input)
				// For now, we'll use defaults - this should be enhanced later
				damageContext[rpgtoolkit.ContextHasAdvantage] = false
				damageContext[rpgtoolkit.ContextHasDisadvantage] = false
				damageContext[rpgtoolkit.ContextAllyAdjacent] = false

				damageEvent, emitErr := rpgtoolkit.CreateAndEmitEvent(
					s.eventBus,
					rpgevents.EventOnDamageRoll,
					char,
					nil, // No target for damage roll
					damageContext,
				)
				if emitErr != nil {
					rpgtoolkit.LogEventError("OnDamageRoll", emitErr)
				}

				// Update damage from event
				if damageEvent != nil {
					if val, ok := damageEvent.Context().Get(rpgtoolkit.ContextDamage); ok {
						if modifiedDamage, ok := val.(int); ok {
							result.Damage = modifiedDamage
						}
					}

					// Check if sneak attack was applied
					if val, ok := damageEvent.Context().Get(rpgtoolkit.ContextSneakAttackDamage); ok {
						if sneakDamage, ok := val.(int); ok && sneakDamage > 0 {
							result.SneakAttackDamage = sneakDamage
							if val2, ok := damageEvent.Context().Get(rpgtoolkit.ContextSneakAttackDice); ok {
								if sneakDice, ok := val2.(string); ok {
									// Extract dice count from format like "3d6"
									_, _ = fmt.Sscanf(sneakDice, "%dd6", &result.SneakAttackDice) //nolint:errcheck
								}
							}
						}
					}
				}
			}

			// Copy reroll information for Great Weapon Fighting display
			if attackResult.RerollInfo != nil {
				result.RerollInfo = attackResult.RerollInfo
			}

			// Check for sneak attack
			if char.Class != nil && char.Class.Key == "rogue" {
				// Get the weapon used
				var weapon *equipment.Weapon
				if char.EquippedSlots[shared.SlotMainHand] != nil {
					if w, ok := char.EquippedSlots[shared.SlotMainHand].(*equipment.Weapon); ok {
						weapon = w
					}
				} else if char.EquippedSlots[shared.SlotTwoHanded] != nil {
					if w, ok := char.EquippedSlots[shared.SlotTwoHanded].(*equipment.Weapon); ok {
						weapon = w
					}
				}

				// Check if sneak attack is eligible
				if weapon != nil && char.CanSneakAttack(weapon, input.HasAdvantage, input.AllyAdjacent, input.HasDisadvantage) {
					// Create combat context for sneak attack
					ctx := &character.CombatContext{
						AttackResult: attackResult,
						IsCritical:   result.Critical,
					}

					// Apply sneak attack damage
					sneakDamage := char.ApplySneakAttack(ctx)
					if sneakDamage > 0 {
						result.SneakAttackDamage = sneakDamage
						result.SneakAttackDice = char.GetSneakAttackDice()
						result.Damage += sneakDamage

						// Save character to persist SneakAttackUsedThisTurn flag
						if err := s.characterService.UpdateEquipment(char); err != nil {
							log.Printf("Failed to update character after sneak attack: %v", err)
						}
					}
				}
			}
		}

	} else if attacker.Type == combat.CombatantTypeMonster && len(attacker.Actions) > 0 && input.ActionIndex >= 0 && input.ActionIndex < len(attacker.Actions) {
		// Monster attack with valid action
		action := attacker.Actions[input.ActionIndex]
		result.WeaponName = action.Name

		// Emit BeforeAttackRoll event
		if s.eventBus != nil {
			// Get target character or create adapter for non-player targets
			var targetChar *character.Character
			if target.Type == combat.CombatantTypePlayer && target.CharacterID != "" {
				var err error
				targetChar, err = s.characterService.GetByID(target.CharacterID)
				if err != nil {
					log.Printf("Failed to get target character: %v", err)
				}
			} else {
				// For non-player targets, we still need a Character for the current event system
				// This will be improved when we fully migrate to rpgtoolkit events
				targetChar = &character.Character{
					ID:   target.ID,
					Name: target.Name,
				}
			}

			// Create event using rpgtoolkit.CreateEntityAdapter for the monster
			actorAdapter := rpgtoolkit.CreateEntityAdapter(attacker)
			var targetCharAdapter core.Entity
			if targetChar != nil {
				targetCharAdapter = rpgtoolkit.WrapCharacter(targetChar)
			}

			beforeAttackContext := map[string]interface{}{
				"weapon":       action.Name,
				"attack_bonus": action.AttackBonus,
			}

			beforeAttackEvent, emitErr := rpgtoolkit.CreateAndEmitEventWithEntities(
				s.eventBus,
				rpgevents.EventBeforeAttackRoll,
				actorAdapter,
				targetCharAdapter,
				beforeAttackContext,
			)
			if emitErr != nil {
				rpgtoolkit.LogEventError("BeforeAttackRoll", emitErr)
			}

			// Check if event was cancelled
			if beforeAttackEvent != nil && beforeAttackEvent.IsCancelled() {
				result.Hit = false
				result.Damage = 0
				result.AttackRoll = 0
				result.TotalAttack = 0
				return result, nil
			}

			// Update attack bonus from event context if modified
			if beforeAttackEvent != nil {
				if val, ok := beforeAttackEvent.Context().Get("attack_bonus"); ok {
					if modifiedBonus, ok := val.(int); ok {
						action.AttackBonus = modifiedBonus
					}
				}
			}
		}

		// Check for disadvantage effects (like Vicious Mockery)
		hasDisadvantage := false
		disadvantageEffectIndex := -1

		if attacker.ActiveEffects != nil {
			for i, effect := range attacker.ActiveEffects {
				if effect.Name == "Vicious Mockery Disadvantage" {
					// Check if this effect applies to attack rolls
					for _, modifier := range effect.Modifiers {
						if modifier.Type == shared.ModifierTypeDisadvantage {
							hasDisadvantage = true
							disadvantageEffectIndex = i
							break
						}
					}
					if hasDisadvantage {
						break
					}
				}
			}
		}

		// Roll attack (with disadvantage if applicable)
		var attackResult struct {
			Total int
			Rolls []int
		}

		if hasDisadvantage {
			// Roll twice for disadvantage
			firstRoll, err := s.diceRoller.Roll(1, 20, action.AttackBonus)
			if err != nil {
				return nil, dnderr.Wrap(err, "failed to roll first attack with disadvantage")
			}

			secondRoll, err := s.diceRoller.Roll(1, 20, action.AttackBonus)
			if err != nil {
				return nil, dnderr.Wrap(err, "failed to roll second attack with disadvantage")
			}

			// Take the lower total
			if firstRoll.Total <= secondRoll.Total {
				attackResult.Total = firstRoll.Total
				attackResult.Rolls = firstRoll.Rolls
			} else {
				attackResult.Total = secondRoll.Total
				attackResult.Rolls = secondRoll.Rolls
			}

			// Log the disadvantage rolls
			encounter.AddCombatLogEntry(fmt.Sprintf("%s attacks with disadvantage (Vicious Mockery) - rolled %d and %d, taking %d",
				attacker.Name, firstRoll.Rolls[0], secondRoll.Rolls[0], attackResult.Rolls[0]))

			// Remove the disadvantage effect after use
			if disadvantageEffectIndex >= 0 {
				attacker.ActiveEffects = append(attacker.ActiveEffects[:disadvantageEffectIndex],
					attacker.ActiveEffects[disadvantageEffectIndex+1:]...)
			}
		} else {
			// Normal attack roll
			normalResult, err := s.diceRoller.Roll(1, 20, action.AttackBonus)
			if err != nil {
				return nil, dnderr.Wrap(err, "failed to roll attack")
			}
			attackResult.Total = normalResult.Total
			attackResult.Rolls = normalResult.Rolls
		}

		result.AttackRoll = attackResult.Rolls[0]
		result.AttackBonus = action.AttackBonus
		result.TotalAttack = attackResult.Total
		result.DiceRolls = attackResult.Rolls

		// Emit OnAttackRoll event
		if s.eventBus != nil {
			// Reuse the actor from the previous event preparation
			actorAdapter := rpgtoolkit.CreateEntityAdapter(attacker)
			var targetCharAdapter core.Entity
			if target.Type == combat.CombatantTypePlayer && target.CharacterID != "" {
				if targetChar, err := s.characterService.GetByID(target.CharacterID); err == nil {
					targetCharAdapter = rpgtoolkit.WrapCharacter(targetChar)
				} else {
					log.Printf("Failed to get target character: %v", err)
				}
			}

			onAttackContext := map[string]interface{}{
				"weapon":       action.Name,
				"attack_roll":  result.AttackRoll,
				"attack_bonus": result.AttackBonus,
				"total_attack": result.TotalAttack,
				"target_ac":    target.AC,
			}

			onAttackEvent, emitErr := rpgtoolkit.CreateAndEmitEventWithEntities(
				s.eventBus,
				rpgevents.EventOnAttackRoll,
				actorAdapter,
				targetCharAdapter,
				onAttackContext,
			)
			if emitErr != nil {
				rpgtoolkit.LogEventError("OnAttackRoll", emitErr)
			}

			// Update attack values from event context if modified
			if onAttackEvent != nil {
				if val, ok := onAttackEvent.Context().Get("total_attack"); ok {
					if modifiedTotal, ok := val.(int); ok {
						result.TotalAttack = modifiedTotal
					}
				}
			}
		}

		// Check hit
		result.Hit = result.TotalAttack >= target.AC
		result.Critical = result.AttackRoll == 20

		// Emit AfterAttackRoll event
		if s.eventBus != nil {
			// Reuse the actor from the previous event preparation
			actorAdapter := rpgtoolkit.CreateEntityAdapter(attacker)
			var targetCharAdapter core.Entity
			if target.Type == combat.CombatantTypePlayer && target.CharacterID != "" {
				if targetChar, err := s.characterService.GetByID(target.CharacterID); err == nil {
					targetCharAdapter = rpgtoolkit.WrapCharacter(targetChar)
				} else {
					log.Printf("Failed to get target character: %v", err)
				}
			}

			afterAttackContext := map[string]interface{}{
				"weapon":       action.Name,
				"attack_roll":  result.AttackRoll,
				"attack_bonus": result.AttackBonus,
				"total_attack": result.TotalAttack,
				"target_ac":    target.AC,
				"hit":          result.Hit,
				"critical":     result.Critical,
			}

			afterAttackEvent, emitErr := rpgtoolkit.CreateAndEmitEventWithEntities(
				s.eventBus,
				rpgevents.EventAfterAttackRoll,
				actorAdapter,
				targetCharAdapter,
				afterAttackContext,
			)
			if emitErr != nil {
				rpgtoolkit.LogEventError("AfterAttackRoll", emitErr)
			}

			// Update hit/critical from event context if modified
			if afterAttackEvent != nil {
				if val, ok := afterAttackEvent.Context().Get("hit"); ok {
					if modifiedHit, ok := val.(bool); ok {
						result.Hit = modifiedHit
					}
				}
				if val, ok := afterAttackEvent.Context().Get("critical"); ok {
					if modifiedCrit, ok := val.(bool); ok {
						result.Critical = modifiedCrit
					}
				}
			}
		}

		if result.Hit {
			// Roll damage for each damage component
			totalDamage := 0
			var allDamageRolls []int
			var totalDamageBonus int

			// Store the primary damage dice info for display
			if len(action.Damage) > 0 {
				primaryDamage := action.Damage[0]
				result.WeaponDiceCount = primaryDamage.DiceCount
				result.WeaponDiceSize = primaryDamage.DiceSize
				log.Printf("Monster action %s damage dice: %dd%d", action.Name, primaryDamage.DiceCount, primaryDamage.DiceSize)
			}

			for _, dmg := range action.Damage {
				damageResult, err := s.diceRoller.Roll(dmg.DiceCount, dmg.DiceSize, dmg.Bonus)
				if err != nil {
					log.Printf("Error rolling damage: %v", err)
					continue
				}

				// Track damage bonuses
				totalDamageBonus += dmg.Bonus

				// Double dice on critical
				if result.Critical {
					critResult, err := s.diceRoller.Roll(dmg.DiceCount, dmg.DiceSize, 0)
					if err == nil {
						damageResult.Total += critResult.Total
						damageResult.Rolls = append(damageResult.Rolls, critResult.Rolls...)
					}
				}

				totalDamage += damageResult.Total
				allDamageRolls = append(allDamageRolls, damageResult.Rolls...)

				// Use first damage type found
				if result.DamageType == "" && dmg.DamageType != "" {
					result.DamageType = string(dmg.DamageType)
				}
			}

			result.Damage = totalDamage
			result.DamageRolls = allDamageRolls
			result.DamageBonus = totalDamageBonus
		}

	} else {
		// Unarmed strike fallback
		result.WeaponName = "Unarmed Strike"
		result.WeaponDiceCount = 1
		result.WeaponDiceSize = 4 // Unarmed strike is always 1d4

		// Roll attack
		attackResult, err := s.diceRoller.Roll(1, 20, 0)
		if err != nil {
			return nil, dnderr.Wrap(err, "failed to roll attack")
		}

		result.AttackRoll = attackResult.Rolls[0]
		result.AttackBonus = 0
		result.TotalAttack = attackResult.Total
		result.DiceRolls = attackResult.Rolls

		// Check hit
		result.Hit = result.TotalAttack >= target.AC
		result.Critical = result.AttackRoll == 20

		if result.Hit {
			// Roll damage
			damageResult, err := s.diceRoller.Roll(1, 4, 0)
			if err != nil {
				return nil, dnderr.Wrap(err, "failed to roll damage")
			}

			if result.Critical {
				critResult, err := s.diceRoller.Roll(1, 4, 0)
				if err == nil {
					damageResult.Total += critResult.Total
					damageResult.Rolls = append(damageResult.Rolls, critResult.Rolls...)
				}
			}

			result.Damage = damageResult.Total
			result.DamageRolls = damageResult.Rolls
			result.DamageType = "bludgeoning"
			result.DamageBonus = 0
		}
	}

	// Apply damage if hit
	if result.Hit && result.Damage > 0 {
		// Use the ApplyDamage method which handles defeat and combat end detection
		// Apply damage with resistance check if target is a player character
		finalDamage := result.Damage
		if target.Type == combat.CombatantTypePlayer && target.CharacterID != "" {
			// Get the character to check for resistances
			if targetChar, err := s.characterService.GetByID(target.CharacterID); err == nil {
				// Apply resistance/vulnerability/immunity
				damageType := damage.TypeSlashing // Default for weapons
				if result.DamageType != "" {
					// Convert damage type string to damage.Type
					switch strings.ToLower(result.DamageType) {
					case "bludgeoning":
						damageType = damage.TypeBludgeoning
					case "piercing":
						damageType = damage.TypePiercing
					case "slashing":
						damageType = damage.TypeSlashing
					case "fire":
						damageType = damage.TypeFire
					case "cold":
						damageType = damage.TypeCold
					case "lightning":
						damageType = damage.TypeLightning
					case "thunder":
						damageType = damage.TypeThunder
					case "acid":
						damageType = damage.TypeAcid
					case "poison":
						damageType = damage.TypePoison
					case "necrotic":
						damageType = damage.TypeNecrotic
					case "radiant":
						damageType = damage.TypeRadiant
					case "psychic":
						damageType = damage.TypePsychic
					case "force":
						damageType = damage.TypeForce
					}
				}

				originalDamage := finalDamage
				finalDamage = targetChar.ApplyDamageResistance(damageType, finalDamage)
				if finalDamage != originalDamage {
					log.Printf("Damage modified by resistance/vulnerability: %d -> %d", originalDamage, finalDamage)
					// Add to combat log
					if finalDamage < originalDamage {
						encounter.CombatLog = append(encounter.CombatLog, fmt.Sprintf("%s's resistance reduces damage from %d to %d", target.Name, originalDamage, finalDamage))
					} else if finalDamage > originalDamage {
						encounter.CombatLog = append(encounter.CombatLog, fmt.Sprintf("%s's vulnerability increases damage from %d to %d", target.Name, originalDamage, finalDamage))
					}
				}
			}
		}

		// Update the result damage to reflect the actual damage dealt
		if finalDamage != result.Damage {
			result.Damage = finalDamage
		}

		// Emit BeforeTakeDamage event for damage resistance (like rage)
		if s.eventBus != nil && finalDamage > 0 {
			// Get target character for event
			var targetCharAdapter core.Entity
			if target.Type == combat.CombatantTypePlayer && target.CharacterID != "" {
				if targetChar, err := s.characterService.GetByID(target.CharacterID); err == nil {
					targetCharAdapter = rpgtoolkit.WrapCharacter(targetChar)
				} else {
					log.Printf("Failed to get target character for BeforeTakeDamage event: %v", err)
				}
			}

			if targetCharAdapter != nil {
				damageType := damage.TypeBludgeoning // Default
				if result.DamageType != "" {
					switch strings.ToLower(result.DamageType) {
					case "slashing":
						damageType = damage.TypeSlashing
					case "piercing":
						damageType = damage.TypePiercing
					case "bludgeoning":
						damageType = damage.TypeBludgeoning
					}
				}

				takeDamageContext := map[string]interface{}{
					"damage":      finalDamage,
					"damage_type": string(damageType),
				}

				takeDamageEvent, emitErr := rpgtoolkit.CreateAndEmitEventWithEntities(
					s.eventBus,
					rpgevents.EventBeforeTakeDamage,
					nil, // No actor for damage taken
					targetCharAdapter,
					takeDamageContext,
				)
				if emitErr != nil {
					rpgtoolkit.LogEventError("BeforeTakeDamage", emitErr)
				}

				// Update damage from event
				if takeDamageEvent != nil {
					if val, ok := takeDamageEvent.Context().Get("damage"); ok {
						if modifiedDamage, ok := val.(int); ok {
							finalDamage = modifiedDamage
							result.Damage = finalDamage
						}
					}
				}
			}
		}

		target.ApplyDamage(finalDamage)
		result.TargetNewHP = target.CurrentHP
		result.TargetDefeated = target.CurrentHP == 0

		// Check if combat should end
		if shouldEnd, playersWon := encounter.CheckCombatEnd(); shouldEnd {
			log.Printf("Combat ending after attack - Players won: %v", playersWon)
			encounter.End()
			result.CombatEnded = true
			result.PlayersWon = playersWon
			if playersWon {
				encounter.AddCombatLogEntry("Victory! All enemies have been defeated!")
			} else {
				encounter.AddCombatLogEntry("Defeat! The party has fallen...")
			}
		}

		// Update encounter
		if err := s.repository.Update(ctx, encounter); err != nil {
			return nil, dnderr.Wrap(err, "failed to update encounter")
		}
	}

	// Generate log entry with dice rolls
	if result.Hit {
		// Format damage dice with expression (e.g., "1d8: [4]+2")
		damageRollStr := ""
		if len(result.DamageRolls) > 0 {
			var diceExpr string
			var diceCount int

			// Always use actual weapon dice - no guessing
			if result.WeaponDiceSize > 0 {
				diceCount = result.WeaponDiceCount
				if result.Critical {
					diceCount *= 2
				}
				diceExpr = fmt.Sprintf("%dd%d", diceCount, result.WeaponDiceSize)
				log.Printf("Using weapon dice: %s", diceExpr)
			} else {
				// This should not happen - log error and use a fallback
				log.Printf("ERROR: No weapon dice info available for attack. Using 1d4 fallback.")
				diceExpr = "1d4"
			}

			// Build damage roll string with Great Weapon Fighting reroll display
			damageRollStr = fmt.Sprintf("%s: [", diceExpr)

			// If we have reroll info, use it to show rerolls with strikethrough
			if len(result.RerollInfo) > 0 {
				// Create a map of positions to reroll info for quick lookup
				rerollMap := make(map[int]attack.DieReroll)
				for _, reroll := range result.RerollInfo {
					rerollMap[reroll.Position] = reroll
				}

				// Build the display string showing rerolls
				for i, finalRoll := range result.DamageRolls {
					if i > 0 {
						damageRollStr += ", "
					}

					// Check if this position had a reroll
					if reroll, hasReroll := rerollMap[i]; hasReroll {
						// Show original roll with strikethrough and new roll
						damageRollStr += fmt.Sprintf("~~%d~~ %d", reroll.OriginalRoll, reroll.NewRoll)
					} else {
						// Show normal roll
						damageRollStr += fmt.Sprintf("%d", finalRoll)
					}
				}
			} else {
				// Standard damage roll display (no rerolls)
				for i, roll := range result.DamageRolls {
					if i > 0 {
						damageRollStr += ", "
					}
					damageRollStr += fmt.Sprintf("%d", roll)
				}
			}

			damageRollStr += "]"
			if result.DamageBonus != 0 {
				damageRollStr += fmt.Sprintf("%+d", result.DamageBonus)

				// Check if there are active damage effects like rage
				if attacker.Type == combat.CombatantTypePlayer && attacker.CharacterID != "" {
					if char, err := s.characterService.GetByID(attacker.CharacterID); err == nil {
						// Check if character has any active damage-modifying effects
						activeEffects := char.GetActiveStatusEffects()
						hasEffects := false
						for _, effect := range activeEffects {
							// Check if this effect modifies damage (like rage)
							for _, mod := range effect.Modifiers {
								if mod.Target == effects.TargetDamage {
									hasEffects = true
									break
								}
							}
							if hasEffects {
								break
							}
						}
						if hasEffects {
							damageRollStr += " (includes effects)"
						}
					}
				}
			}
		}

		// Add sneak attack to damage description if applicable
		sneakAttackStr := ""
		if result.SneakAttackDamage > 0 {
			diceCount := result.SneakAttackDice
			if result.Critical {
				diceCount *= 2
			}
			sneakAttackStr = fmt.Sprintf(" + 🗡️ %dd6 Sneak Attack: %d", diceCount, result.SneakAttackDamage)
		}

		// Add proficiency indicator
		profIndicator := ""
		if attacker.Type == combat.CombatantTypePlayer && attacker.CharacterID != "" && result.WeaponName != "Unarmed Strike" {
			// Check proficiency again for log message
			if char, err := s.characterService.GetByID(attacker.CharacterID); err == nil {
				isProficient := true
				if char.EquippedSlots != nil {
					if item := char.EquippedSlots[shared.SlotMainHand]; item != nil {
						if w, ok := item.(*equipment.Weapon); ok {
							isProficient = char.HasWeaponProficiency(w.Base.Key)
						}
					} else if item := char.EquippedSlots[shared.SlotTwoHanded]; item != nil {
						if w, ok := item.(*equipment.Weapon); ok {
							isProficient = char.HasWeaponProficiency(w.Base.Key)
						}
					}
				}
				if !isProficient {
					profIndicator = " ⚠️ NO PROF"
				}
			}
		}

		if result.Critical {
			result.LogEntry = fmt.Sprintf("⚔️ **%s** → **%s** | 💥 CRIT! 🩸 **%d** ||d20:**%d**%+d=%d vs AC:%d, dmg:%s%s||%s",
				result.AttackerName, result.TargetName,
				result.Damage,
				result.AttackRoll, result.AttackBonus, result.TotalAttack, result.TargetAC,
				damageRollStr, sneakAttackStr, profIndicator)
		} else {
			result.LogEntry = fmt.Sprintf("⚔️ **%s** → **%s** | HIT 🩸 **%d** ||d20:%d%+d=%d vs AC:%d, dmg:%s%s||%s",
				result.AttackerName, result.TargetName,
				result.Damage,
				result.AttackRoll, result.AttackBonus, result.TotalAttack, result.TargetAC,
				damageRollStr, sneakAttackStr, profIndicator)
		}

		if result.TargetDefeated {
			result.LogEntry += " 💀"
		}
	} else {
		// Add proficiency indicator for misses too
		profIndicator := ""
		if attacker.Type == combat.CombatantTypePlayer && attacker.CharacterID != "" {
			if char, err := s.characterService.GetByID(attacker.CharacterID); err == nil {
				var weapon *equipment.Weapon
				isProficient := true // Default to true for unarmed strikes

				if char.EquippedSlots[shared.SlotMainHand] != nil {
					if w, ok := char.EquippedSlots[shared.SlotMainHand].(*equipment.Weapon); ok {
						weapon = w
						isProficient = char.HasWeaponProficiency(w.Base.Key)
					}
				} else if char.EquippedSlots[shared.SlotTwoHanded] != nil {
					if w, ok := char.EquippedSlots[shared.SlotTwoHanded].(*equipment.Weapon); ok {
						weapon = w
						isProficient = char.HasWeaponProficiency(w.Base.Key)
					}
				}

				if !isProficient && weapon != nil {
					profIndicator = " ⚠️ NO PROF"
				}
			}
		}

		result.LogEntry = fmt.Sprintf("⚔️ **%s** → **%s** | ❌ MISS ||d20:%d%+d=%d vs AC:%d||%s",
			result.AttackerName, result.TargetName,
			result.AttackRoll, result.AttackBonus, result.TotalAttack, result.TargetAC, profIndicator)
	}

	// Add to combat log
	encounter.CombatLog = append(encounter.CombatLog, result.LogEntry)
	if err := s.repository.Update(ctx, encounter); err != nil {
		log.Printf("Error updating combat log: %v", err)
	}

	return result, nil
}

// ApplyDamage applies damage to a combatant
func (s *service) ApplyDamage(ctx context.Context, encounterID, combatantID, userID string, damageAmount int) error {
	// Get encounter
	encounter, err := s.repository.Get(ctx, encounterID)
	if err != nil {
		return dnderr.Wrap(err, "failed to get encounter")
	}

	// Check permissions
	// For dungeon encounters, allow damage application regardless of turn
	// since the bot orchestrates combat on behalf of players
	if session, err := s.sessionService.GetSession(ctx, encounter.SessionID); err == nil {
		if sessionType, ok := session.Metadata["sessionType"].(string); ok && sessionType == "dungeon" {
			// Allow damage in dungeon encounters
		} else if !encounter.CanPlayerAct(userID) {
			// For regular encounters, check turn order
			return dnderr.PermissionDenied("not your turn")
		}
	} else if !encounter.CanPlayerAct(userID) {
		// Fallback to turn check if session lookup fails
		return dnderr.PermissionDenied("not your turn")
	}

	// Get combatant
	combatant, exists := encounter.Combatants[combatantID]
	if !exists {
		return dnderr.InvalidArgument("combatant not found")
	}

	// Apply damage
	combatant.ApplyDamage(damageAmount)

	// Add to combat log if damage was dealt
	if damageAmount > 0 {
		// Find attacker name (could be current turn or explicit)
		attackerName := "Unknown"
		if current := encounter.GetCurrentCombatant(); current != nil {
			attackerName = current.Name
		}
		encounter.AddCombatLogEntry(fmt.Sprintf("%s hit %s for %d damage", attackerName, combatant.Name, damageAmount))

		if combatant.CurrentHP == 0 {
			encounter.AddCombatLogEntry(fmt.Sprintf("%s was defeated!", combatant.Name))
		}
	}

	// Check if combat should end
	if shouldEnd, playersWon := encounter.CheckCombatEnd(); shouldEnd {
		log.Printf("Combat ending - Players won: %v", playersWon)
		encounter.End()
		if playersWon {
			encounter.AddCombatLogEntry("Victory! All enemies have been defeated!")
		} else {
			encounter.AddCombatLogEntry("Defeat! The party has fallen...")
		}
	}

	// Save changes
	if err := s.repository.Update(ctx, encounter); err != nil {
		return dnderr.Wrap(err, "failed to update encounter")
	}

	return nil
}

// HealCombatant heals a combatant
func (s *service) HealCombatant(ctx context.Context, encounterID, combatantID, userID string, amount int) error {
	// Get encounter
	encounter, err := s.repository.Get(ctx, encounterID)
	if err != nil {
		return dnderr.Wrap(err, "failed to get encounter")
	}

	// Check permissions
	if !encounter.CanPlayerAct(userID) {
		return dnderr.PermissionDenied("not your turn")
	}

	// Get combatant
	combatant, exists := encounter.Combatants[combatantID]
	if !exists {
		return dnderr.InvalidArgument("combatant not found")
	}

	// Apply healing
	combatant.Heal(amount)

	// Save changes
	if err := s.repository.Update(ctx, encounter); err != nil {
		return dnderr.Wrap(err, "failed to update encounter")
	}

	return nil
}

// EndEncounter ends the encounter
func (s *service) EndEncounter(ctx context.Context, encounterID, userID string) error {
	// Get encounter
	encounter, err := s.repository.Get(ctx, encounterID)
	if err != nil {
		return dnderr.Wrap(err, "failed to get encounter")
	}

	// Check permissions
	if encounter.CreatedBy != userID {
		return dnderr.PermissionDenied("only the DM can end the encounter")
	}

	// End encounter
	encounter.End()

	// Save changes
	if err := s.repository.Update(ctx, encounter); err != nil {
		return dnderr.Wrap(err, "failed to update encounter")
	}

	return nil
}

// LogCombatAction logs a combat action without applying damage
func (s *service) LogCombatAction(ctx context.Context, encounterID, action string) error {
	// Get encounter
	encounter, err := s.repository.Get(ctx, encounterID)
	if err != nil {
		return dnderr.Wrap(err, "failed to get encounter")
	}

	// Add to combat log
	encounter.AddCombatLogEntry(action)

	// Save changes
	if err := s.repository.Update(ctx, encounter); err != nil {
		return dnderr.Wrap(err, "failed to update encounter")
	}

	return nil
}

// ProcessMonsterTurn handles a monster's turn automatically
func (s *service) ProcessMonsterTurn(ctx context.Context, encounterID, monsterID string) (*AttackResult, error) {
	// Get encounter
	encounter, err := s.repository.Get(ctx, encounterID)
	if err != nil {
		return nil, dnderr.Wrap(err, "failed to get encounter")
	}

	// Get the monster
	monster, exists := encounter.Combatants[monsterID]
	if !exists || !monster.IsActive {
		return nil, dnderr.InvalidArgument("monster not found or inactive")
	}

	// Find a target (first active player)
	var target *combat.Combatant
	for _, combatant := range encounter.Combatants {
		if combatant.Type == combat.CombatantTypePlayer && combatant.IsActive {
			target = combatant
			break
		}
	}

	if target == nil {
		log.Printf("ProcessMonsterTurn - No valid player targets found for monster %s", monster.Name)
		return nil, dnderr.NotFound("no valid target found")
	}

	// Use PerformAttack with the first action
	actionIndex := 0
	if len(monster.Actions) == 0 {
		actionIndex = -1 // Will trigger unarmed strike
	}

	return s.PerformAttack(ctx, &AttackInput{
		EncounterID: encounterID,
		AttackerID:  monsterID,
		TargetID:    target.ID,
		UserID:      encounter.CreatedBy, // DM/bot
		ActionIndex: actionIndex,
	})
}

// ProcessAllMonsterTurns processes all consecutive monster turns
func (s *service) ProcessAllMonsterTurns(ctx context.Context, encounterID string) ([]*AttackResult, error) {
	var results []*AttackResult

	for {
		// Get current encounter state
		encounter, err := s.repository.Get(ctx, encounterID)
		if err != nil {
			return results, dnderr.Wrap(err, "failed to get encounter")
		}

		// Check if current turn is a monster
		current := encounter.GetCurrentCombatant()
		if current == nil || current.Type != combat.CombatantTypeMonster || !current.IsActive {
			break // Not a monster's turn or no current combatant
		}

		// Process this monster's turn
		result, err := s.ProcessMonsterTurn(ctx, encounterID, current.ID)
		if err != nil {
			log.Printf("Error processing monster turn: %v", err)
			// Continue anyway, the monster might just not have a valid target
		} else if result != nil {
			results = append(results, result)
		}

		// Advance to next turn
		err = s.NextTurn(ctx, encounterID, encounter.CreatedBy)
		if err != nil {
			return results, dnderr.Wrap(err, "failed to advance turn")
		}

		// Check if combat ended
		encounter, err = s.repository.Get(ctx, encounterID)
		if err != nil {
			return results, dnderr.Wrap(err, "failed to get encounter after turn")
		}

		if encounter.Status != combat.EncounterStatusActive {
			break // Combat ended
		}
	}

	return results, nil
}

// ExecuteAttackWithTarget handles a complete attack sequence including auto-advancing turns
func (s *service) ExecuteAttackWithTarget(ctx context.Context, input *ExecuteAttackInput) (*ExecuteAttackResult, error) {
	result := &ExecuteAttackResult{
		MonsterAttacks: []*AttackResult{},
	}

	// Get encounter
	encounter, err := s.repository.Get(ctx, input.EncounterID)
	if err != nil {
		return nil, dnderr.Wrap(err, "failed to get encounter")
	}

	// Find the attacker - could be the player who clicked or current turn (for DM)
	var attacker *combat.Combatant

	// First, try to find the player's combatant
	for _, combatant := range encounter.Combatants {
		if combatant.PlayerID == input.UserID && combatant.IsActive {
			attacker = combatant
			break
		}
	}

	// If not found, use current turn (for DM controlling monsters)
	if attacker == nil {
		attacker = encounter.GetCurrentCombatant()
		if attacker == nil || !attacker.IsActive {
			return nil, dnderr.InvalidArgument("no active attacker found")
		}
	}

	// Execute the attack
	attackInput := &AttackInput{
		EncounterID: input.EncounterID,
		AttackerID:  attacker.ID,
		TargetID:    input.TargetID,
		UserID:      input.UserID,
		ActionIndex: 0, // Default to first action
	}

	result.PlayerAttack, err = s.PerformAttack(ctx, attackInput)
	if err != nil {
		return nil, dnderr.Wrap(err, "failed to perform attack")
	}

	// Check if combat ended from the attack
	if result.PlayerAttack.CombatEnded {
		result.CombatEnded = true
		result.PlayersWon = result.PlayerAttack.PlayersWon
		return result, nil // Don't process turns if combat ended
	}

	// Don't auto-advance turn for player attacks - they may have bonus actions
	// Only auto-advance for monster attacks
	if attacker.Type == combat.CombatantTypeMonster {
		err = s.NextTurn(ctx, input.EncounterID, input.UserID)
		if err != nil {
			log.Printf("Error auto-advancing turn after monster attack: %v", err)
		} else {
			// Process any consecutive monster turns
			monsterResults, monstErr := s.ProcessAllMonsterTurns(ctx, input.EncounterID)
			if monstErr != nil {
				log.Printf("Error processing monster turns: %v", monstErr)
			} else {
				result.MonsterAttacks = monsterResults
			}
		}
	}
	// For player attacks, let them manually end their turn after considering bonus actions

	// Get updated encounter state
	encounter, err = s.repository.Get(ctx, input.EncounterID)
	if err != nil {
		log.Printf("Error getting updated encounter: %v", err)
	} else {
		// Check current turn
		if current := encounter.GetCurrentCombatant(); current != nil {
			result.CurrentTurn = current
			result.IsPlayerTurn = current.PlayerID == input.UserID
		}

		// Check if combat ended
		if encounter.Status == combat.EncounterStatusCompleted {
			result.CombatEnded = true
			_, result.PlayersWon = encounter.CheckCombatEnd()
		}
	}

	return result, nil
}

// UpdateMessageID updates the Discord message ID for an encounter
func (s *service) UpdateMessageID(ctx context.Context, encounterID, messageID, channelID string) error {
	// Validate input
	if encounterID == "" {
		return dnderr.InvalidArgument("encounter ID is required")
	}
	if messageID == "" {
		return dnderr.InvalidArgument("message ID is required")
	}
	if channelID == "" {
		return dnderr.InvalidArgument("channel ID is required")
	}

	// Get the encounter
	encounter, err := s.repository.Get(ctx, encounterID)
	if err != nil {
		return dnderr.Wrap(err, "failed to get encounter")
	}
	if encounter == nil {
		return dnderr.NotFound("encounter not found")
	}

	// Update the message ID and channel ID
	encounter.MessageID = messageID
	encounter.ChannelID = channelID

	// Save the updated encounter
	err = s.repository.Update(ctx, encounter)
	if err != nil {
		return dnderr.Wrap(err, "failed to update encounter")
	}

	log.Printf("Updated encounter %s with message ID %s in channel %s", encounterID, messageID, channelID)
	return nil
}

// StatusEffectHandler handles status effect events for encounters
type StatusEffectHandler struct {
	service Service
}

// NewStatusEffectHandler creates a new status effect handler
func NewStatusEffectHandler(service Service) *StatusEffectHandler {
	return &StatusEffectHandler{
		service: service,
	}
}

// HandleEvent processes status applied events
func (h *StatusEffectHandler) HandleEvent(event rpgevents.Event) error {
	// Only handle status applied events
	if event.Type() != rpgevents.EventOnStatusApplied {
		return nil
	}

	// Get target ID
	var targetID string
	if val, ok := event.Context().Get(rpgtoolkit.ContextTargetID); ok {
		if id, ok := val.(string); ok {
			targetID = id
		}
	}
	if targetID == "" {
		log.Printf("StatusEffectHandler: No target ID for status effect")
		return nil
	}

	// Get effect details
	var effectName, effectType string
	var duration int
	if val, ok := event.Context().Get("effect_name"); ok {
		if name, ok := val.(string); ok {
			effectName = name
		}
	}
	if val, ok := event.Context().Get("effect_type"); ok {
		if eType, ok := val.(string); ok {
			effectType = eType
		}
	}
	if val, ok := event.Context().Get("effect_duration"); ok {
		if d, ok := val.(int); ok {
			duration = d
		}
	}

	// Check if this is vicious mockery disadvantage
	if effectType == "disadvantage_next_attack" && effectName == "Vicious Mockery Disadvantage" {
		// Apply to both players and monsters
		log.Printf("StatusEffectHandler: Applying vicious mockery disadvantage to target %s", targetID)

		// First check if target is a player character
		// If event.Target is not nil, it's a player character
		if event.Target() != nil {
			if targetEntity := event.Target(); targetEntity != nil {
				if char, ok := rpgtoolkit.ExtractCharacter(targetEntity); ok && char != nil && char.OwnerID != "" {
					// For players, the effect is handled through character.Resources.ActiveEffects
					// This is already done in the vicious_mockery.go ApplyViciousMockeryDisadvantage function
					log.Printf("StatusEffectHandler: Target is a player, effect already applied to character")
					return nil
				}
			}
		}

		// For monsters, we need to find them in an active encounter
		// Get encounter ID from context if available
		var encounterID string
		if val, ok := event.Context().Get(rpgtoolkit.ContextEncounterID); ok {
			if encID, ok := val.(string); ok {
				encounterID = encID
			}
		}
		if encounterID == "" {
			// Try to find an active encounter containing this target
			// This would require additional context or a repository method to search
			log.Printf("StatusEffectHandler: No encounter ID, cannot apply effect to monster")
			return nil
		}

		// Get the encounter
		ctx := context.Background()
		encounter, err := h.service.GetEncounter(ctx, encounterID)
		if err != nil {
			log.Printf("StatusEffectHandler: Failed to get encounter: %v", err)
			return nil
		}

		// Find the target combatant
		combatant, exists := encounter.Combatants[targetID]
		if !exists {
			log.Printf("StatusEffectHandler: Target %s not found in encounter", targetID)
			return nil
		}

		// Create the effect
		effect := &shared.ActiveEffect{
			ID:           uuid.NewGoogleUUIDGenerator().New(),
			Name:         effectName,
			Description:  "Disadvantage on next attack roll",
			Source:       "Vicious Mockery",
			Duration:     duration,
			DurationType: shared.DurationTypeRounds,
			Modifiers: []shared.Modifier{
				{
					Type: shared.ModifierTypeDisadvantage,
				},
			},
		}

		// Initialize ActiveEffects if nil
		if combatant.ActiveEffects == nil {
			combatant.ActiveEffects = []*shared.ActiveEffect{}
		}

		// Add the effect to the combatant
		combatant.ActiveEffects = append(combatant.ActiveEffects, effect)

		// Update the encounter - we can't access repository directly from Service interface
		// So we'll rely on the encounter being saved elsewhere in the flow
		log.Printf("StatusEffectHandler: Effect applied, encounter will be saved by calling method")

		log.Printf("StatusEffectHandler: Successfully applied vicious mockery disadvantage to %s (%s)",
			combatant.Name, combatant.Type)
	}

	return nil
}

// Priority returns the handler priority
func (h *StatusEffectHandler) Priority() int {
	return 100
}
