package encounter

//go:generate mockgen -destination=mock/mock_service.go -package=mockencounter -source=service.go

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/dice"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	dnderr "github.com/KirkDiggler/dnd-bot-discord/internal/errors"
	"github.com/KirkDiggler/dnd-bot-discord/internal/interfaces"
	"github.com/KirkDiggler/dnd-bot-discord/internal/repositories/encounters"
	charService "github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	sessService "github.com/KirkDiggler/dnd-bot-discord/internal/services/session"
	"github.com/KirkDiggler/dnd-bot-discord/internal/uuid"
)

// Service defines the encounter service interface
type Service interface {
	// CreateEncounter creates a new encounter in a session
	CreateEncounter(ctx context.Context, input *CreateEncounterInput) (*entities.Encounter, error)

	// GetEncounter retrieves an encounter by ID
	GetEncounter(ctx context.Context, encounterID string) (*entities.Encounter, error)

	// GetActiveEncounter retrieves the active encounter for a session
	GetActiveEncounter(ctx context.Context, sessionID string) (*entities.Encounter, error)

	// AddMonster adds a monster to an encounter
	AddMonster(ctx context.Context, encounterID, userID string, input *AddMonsterInput) (*entities.Combatant, error)

	// AddPlayer adds a player character to an encounter
	AddPlayer(ctx context.Context, encounterID, playerID, characterID string) (*entities.Combatant, error)

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
	Actions         []*entities.MonsterAction
}

// AttackInput contains data for performing an attack
type AttackInput struct {
	EncounterID string
	AttackerID  string
	TargetID    string
	UserID      string // User requesting the attack
	ActionIndex int    // For monsters with multiple attacks (0 = first action)
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
	CurrentTurn  *entities.Combatant
	CombatEnded  bool
	PlayersWon   bool
}

type service struct {
	repository       encounters.Repository
	sessionService   sessService.Service
	characterService charService.Service
	uuidGenerator    uuid.Generator
	diceRoller       interfaces.DiceRoller
}

// ServiceConfig holds configuration for the service
type ServiceConfig struct {
	Repository       encounters.Repository
	SessionService   sessService.Service
	CharacterService charService.Service
	UUIDGenerator    uuid.Generator
	DiceRoller       interfaces.DiceRoller
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

	return svc
}

// CreateEncounter creates a new encounter in a session
func (s *service) CreateEncounter(ctx context.Context, input *CreateEncounterInput) (*entities.Encounter, error) {
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
		if sessionType, ok := session.Metadata["sessionType"].(string); !ok || sessionType != "dungeon" {
			return nil, dnderr.PermissionDenied("only the DM can create encounters")
		}
		// For dungeon sessions, bot/system can create encounters
	} else if member.Role != entities.SessionRoleDM {
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
	encounter := entities.NewEncounter(encounterID, input.SessionID, input.ChannelID, input.Name, input.UserID)
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
func (s *service) GetEncounter(ctx context.Context, encounterID string) (*entities.Encounter, error) {
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
func (s *service) GetActiveEncounter(ctx context.Context, sessionID string) (*entities.Encounter, error) {
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
func (s *service) AddMonster(ctx context.Context, encounterID, userID string, input *AddMonsterInput) (*entities.Combatant, error) {
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
			if sessionType, ok := session.Metadata["sessionType"].(string); !ok || sessionType != "dungeon" {
				return nil, dnderr.PermissionDenied("only the DM can add monsters")
			}
		} else if member.Role != entities.SessionRoleDM {
			return nil, dnderr.PermissionDenied("only the DM can add monsters")
		}
	}

	// Create combatant
	combatantID := s.uuidGenerator.New()
	combatant := &entities.Combatant{
		ID:              combatantID,
		Name:            input.Name,
		Type:            entities.CombatantTypeMonster,
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
func (s *service) AddPlayer(ctx context.Context, encounterID, playerID, characterID string) (*entities.Combatant, error) {
	// Get encounter
	encounter, err := s.repository.Get(ctx, encounterID)
	if err != nil {
		return nil, dnderr.Wrap(err, "failed to get encounter")
	}

	// Get character details
	character, err := s.characterService.GetByID(characterID)
	if err != nil {
		return nil, dnderr.Wrap(err, "failed to get character")
	}

	// Log character details
	log.Printf("AddPlayer - Retrieved character: ID=%s, Name=%s, OwnerID=%s", character.ID, character.Name, character.OwnerID)

	// Verify character belongs to player
	if character.OwnerID != playerID {
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
	if dexScore, exists := character.Attributes[entities.AttributeDexterity]; exists {
		dexBonus = dexScore.Bonus
	}

	combatant := &entities.Combatant{
		ID:              combatantID,
		Name:            character.Name,
		Type:            entities.CombatantTypePlayer,
		InitiativeBonus: dexBonus,
		CurrentHP:       character.CurrentHitPoints,
		MaxHP:           character.MaxHitPoints,
		AC:              character.AC,
		Speed:           30, // Default, should come from race
		IsActive:        true,
		PlayerID:        playerID,
		CharacterID:     characterID,
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
		if sessionType, ok := session.Metadata["sessionType"].(string); !ok || sessionType != "dungeon" {
			return dnderr.PermissionDenied("only the DM can roll initiative")
		}
	}

	// Check status
	if encounter.Status != entities.EncounterStatusSetup {
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
		total, rolls, err := s.diceRoller.Roll(1, 20, combatant.InitiativeBonus)
		if err != nil {
			return dnderr.Wrap(err, "failed to roll initiative")
		}
		combatant.Initiative = total
		initiatives[id] = combatant.Initiative

		// Log the initiative roll
		logEntry := fmt.Sprintf("**%s** rolls initiative: %v + %d = **%d**",
			combatant.Name,
			rolls[0], // The d20 roll
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

	encounter.Status = entities.EncounterStatusRolling

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
		if sessionType, ok := session.Metadata["sessionType"].(string); !ok || sessionType != "dungeon" {
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
			if current.Type == entities.CombatantTypeMonster {
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

	// Advance turn
	encounter.NextTurn()

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
	if encounter.Status != entities.EncounterStatusActive {
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
	if attacker.Type == entities.CombatantTypePlayer && attacker.CharacterID != "" {
		// Player attack using character
		char, err := s.characterService.GetByID(attacker.CharacterID)
		if err != nil {
			return nil, dnderr.Wrap(err, "failed to get character")
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
		result.AttackRoll = attackResult.AttackResult.Rolls[0]      // The d20 roll
		result.TotalAttack = attackResult.AttackRoll                // Total including bonuses
		result.AttackBonus = result.TotalAttack - result.AttackRoll // Calculate bonus from total minus d20
		result.DiceRolls = attackResult.AttackResult.Rolls

		// Get weapon name
		if char.EquippedSlots[entities.SlotMainHand] != nil {
			result.WeaponName = char.EquippedSlots[entities.SlotMainHand].GetName()
		} else if char.EquippedSlots[entities.SlotTwoHanded] != nil {
			result.WeaponName = char.EquippedSlots[entities.SlotTwoHanded].GetName()
		} else {
			result.WeaponName = "Unarmed Strike"
		}

		// Check hit
		result.Hit = result.TotalAttack >= target.AC
		result.Critical = result.AttackRoll == 20

		if result.Hit {
			result.Damage = attackResult.DamageRoll
			result.DamageType = string(attackResult.AttackType)
			if attackResult.DamageResult != nil {
				result.DamageRolls = attackResult.DamageResult.Rolls
				result.DamageBonus = attackResult.DamageRoll - attackResult.DamageResult.Total
			}
		}

	} else if attacker.Type == entities.CombatantTypeMonster && len(attacker.Actions) > 0 && input.ActionIndex >= 0 && input.ActionIndex < len(attacker.Actions) {
		// Monster attack with valid action
		action := attacker.Actions[input.ActionIndex]
		result.WeaponName = action.Name

		// Roll attack
		attackTotal, attackRolls, err := s.diceRoller.Roll(1, 20, action.AttackBonus)
		if err != nil {
			return nil, dnderr.Wrap(err, "failed to roll attack")
		}

		result.AttackRoll = attackRolls[0]
		result.AttackBonus = action.AttackBonus
		result.TotalAttack = attackTotal
		result.DiceRolls = attackRolls

		// Check hit
		result.Hit = result.TotalAttack >= target.AC
		result.Critical = result.AttackRoll == 20

		if result.Hit {
			// Roll damage for each damage component
			totalDamage := 0
			var allDamageRolls []int

			for _, dmg := range action.Damage {
				damageTotal, damageRolls, err := s.diceRoller.Roll(dmg.DiceCount, dmg.DiceSize, dmg.Bonus)
				if err != nil {
					log.Printf("Error rolling damage: %v", err)
					continue
				}

				// Double dice on critical
				if result.Critical {
					critTotal, critRolls, err := s.diceRoller.Roll(dmg.DiceCount, dmg.DiceSize, 0)
					if err == nil {
						damageTotal += critTotal
						damageRolls = append(damageRolls, critRolls...)
					}
				}

				totalDamage += damageTotal
				allDamageRolls = append(allDamageRolls, damageRolls...)

				// Use first damage type found
				if result.DamageType == "" && dmg.DamageType != "" {
					result.DamageType = string(dmg.DamageType)
				}
			}

			result.Damage = totalDamage
			result.DamageRolls = allDamageRolls
		}

	} else {
		// Unarmed strike fallback
		result.WeaponName = "Unarmed Strike"

		// Roll attack
		attackTotal, attackRolls, err := s.diceRoller.Roll(1, 20, 0)
		if err != nil {
			return nil, dnderr.Wrap(err, "failed to roll attack")
		}

		result.AttackRoll = attackRolls[0]
		result.AttackBonus = 0
		result.TotalAttack = attackTotal
		result.DiceRolls = attackRolls

		// Check hit
		result.Hit = result.TotalAttack >= target.AC
		result.Critical = result.AttackRoll == 20

		if result.Hit {
			// Roll damage
			damageTotal, damageRolls, err := s.diceRoller.Roll(1, 4, 0)
			if err != nil {
				return nil, dnderr.Wrap(err, "failed to roll damage")
			}

			if result.Critical {
				critTotal, critRolls, err := s.diceRoller.Roll(1, 4, 0)
				if err == nil {
					damageTotal += critTotal
					damageRolls = append(damageRolls, critRolls...)
				}
			}

			result.Damage = damageTotal
			result.DamageRolls = damageRolls
			result.DamageType = "bludgeoning"
		}
	}

	// Apply damage if hit
	if result.Hit && result.Damage > 0 {
		// Use the ApplyDamage method which handles defeat and combat end detection
		target.ApplyDamage(result.Damage)
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

	// Generate log entry
	if result.Hit {
		if result.Critical {
			result.LogEntry = fmt.Sprintf("⚔️ **CRITICAL HIT!** %s attacks %s with %s: %d + %d = **%d** vs AC %d - **HIT** for %d %s damage!",
				result.AttackerName, result.TargetName, result.WeaponName,
				result.AttackRoll, result.AttackBonus, result.TotalAttack, result.TargetAC,
				result.Damage, result.DamageType)
		} else {
			result.LogEntry = fmt.Sprintf("⚔️ %s attacks %s with %s: %d + %d = **%d** vs AC %d - **HIT** for %d %s damage",
				result.AttackerName, result.TargetName, result.WeaponName,
				result.AttackRoll, result.AttackBonus, result.TotalAttack, result.TargetAC,
				result.Damage, result.DamageType)
		}

		if result.TargetDefeated {
			result.LogEntry += fmt.Sprintf("\n💀 **%s has been defeated!**", result.TargetName)
		}
	} else {
		result.LogEntry = fmt.Sprintf("⚔️ %s attacks %s with %s: %d + %d = **%d** vs AC %d - **MISS**",
			result.AttackerName, result.TargetName, result.WeaponName,
			result.AttackRoll, result.AttackBonus, result.TotalAttack, result.TargetAC)
	}

	// Add to combat log
	encounter.CombatLog = append(encounter.CombatLog, result.LogEntry)
	if err := s.repository.Update(ctx, encounter); err != nil {
		log.Printf("Error updating combat log: %v", err)
	}

	return result, nil
}

// ApplyDamage applies damage to a combatant
func (s *service) ApplyDamage(ctx context.Context, encounterID, combatantID, userID string, damage int) error {
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
	combatant.ApplyDamage(damage)

	// Add to combat log if damage was dealt
	if damage > 0 {
		// Find attacker name (could be current turn or explicit)
		attackerName := "Unknown"
		if current := encounter.GetCurrentCombatant(); current != nil {
			attackerName = current.Name
		}
		encounter.AddCombatLogEntry(fmt.Sprintf("%s hit %s for %d damage", attackerName, combatant.Name, damage))

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
func (s *service) ProcessMonsterTurn(ctx context.Context, encounterID string, monsterID string) (*AttackResult, error) {
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
	var target *entities.Combatant
	for _, combatant := range encounter.Combatants {
		log.Printf("ProcessMonsterTurn - Checking combatant %s: Type=%s, IsActive=%v, HP=%d/%d",
			combatant.Name, combatant.Type, combatant.IsActive, combatant.CurrentHP, combatant.MaxHP)
		if combatant.Type == entities.CombatantTypePlayer && combatant.IsActive {
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
		if current == nil || current.Type != entities.CombatantTypeMonster || !current.IsActive {
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

		if encounter.Status != entities.EncounterStatusActive {
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
	var attacker *entities.Combatant

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

	// Auto-advance turn if it was a player attack
	if attacker.Type == entities.CombatantTypePlayer {
		err = s.NextTurn(ctx, input.EncounterID, input.UserID)
		if err != nil {
			log.Printf("Error auto-advancing turn: %v", err)
		} else {
			// Process any monster turns that follow
			monsterResults, err := s.ProcessAllMonsterTurns(ctx, input.EncounterID)
			if err != nil {
				log.Printf("Error processing monster turns: %v", err)
			} else {
				result.MonsterAttacks = monsterResults
			}
		}
	}

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
		if encounter.Status == entities.EncounterStatusCompleted {
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
