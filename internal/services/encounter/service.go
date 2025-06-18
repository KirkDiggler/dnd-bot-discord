package encounter

//go:generate mockgen -destination=mock/mock_service.go -package=mockencounter -source=service.go

import (
	"context"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/dice"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	dnderr "github.com/KirkDiggler/dnd-bot-discord/internal/errors"
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

	// ApplyDamage applies damage to a combatant
	ApplyDamage(ctx context.Context, encounterID, combatantID, userID string, damage int) error

	// HealCombatant heals a combatant
	HealCombatant(ctx context.Context, encounterID, combatantID, userID string, amount int) error

	// EndEncounter ends the encounter
	EndEncounter(ctx context.Context, encounterID, userID string) error
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

type service struct {
	repository       encounters.Repository
	sessionService   sessService.Service
	characterService charService.Service
	uuidGenerator    uuid.Generator
}

// ServiceConfig holds configuration for the service
type ServiceConfig struct {
	Repository       encounters.Repository
	SessionService   sessService.Service
	CharacterService charService.Service
	UUIDGenerator    uuid.Generator
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
	}

	if cfg.UUIDGenerator != nil {
		svc.uuidGenerator = cfg.UUIDGenerator
	} else {
		svc.uuidGenerator = uuid.NewGoogleUUIDGenerator()
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

	// Check if user is DM
	member, exists := session.Members[input.UserID]
	if !exists || member.Role != entities.SessionRoleDM {
		return nil, dnderr.PermissionDenied("only the DM can create encounters")
	}

	// Check if there's already an active encounter
	activeEncounter, err := s.repository.GetActiveBySession(ctx, input.SessionID)
	if err != nil {
		return nil, dnderr.Wrap(err, "failed to get active encounter")
	}
	if activeEncounter != nil {
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

	// Verify user is DM
	if encounter.CreatedBy != userID {
		// Check if user is DM of the session
		session, err := s.sessionService.GetSession(ctx, encounter.SessionID)
		if err != nil {
			return nil, dnderr.Wrap(err, "failed to get session")
		}

		member, exists := session.Members[userID]
		if !exists || member.Role != entities.SessionRoleDM {
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
		return dnderr.PermissionDenied("only the DM can roll initiative")
	}

	// Check status
	if encounter.Status != entities.EncounterStatusSetup {
		return dnderr.InvalidArgument("encounter is not in setup phase")
	}

	// Roll initiative for each combatant
	initiatives := make(map[string]int)
	for id, combatant := range encounter.Combatants {
		rollResult, err := dice.RollString("1d20")
		if err != nil {
			return dnderr.Wrap(err, "failed to roll initiative")
		}
		combatant.Initiative = rollResult.Total + combatant.InitiativeBonus
		initiatives[id] = combatant.Initiative
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
		return dnderr.PermissionDenied("only the DM can start the encounter")
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

	// Player can end their own turn, DM can end any turn
	if current.PlayerID != userID && encounter.CreatedBy != userID {
		return dnderr.PermissionDenied("not your turn")
	}

	// Advance turn
	encounter.NextTurn()

	// Save changes
	if err := s.repository.Update(ctx, encounter); err != nil {
		return dnderr.Wrap(err, "failed to update encounter")
	}

	return nil
}

// ApplyDamage applies damage to a combatant
func (s *service) ApplyDamage(ctx context.Context, encounterID, combatantID, userID string, damage int) error {
	// Get encounter
	encounter, err := s.repository.Get(ctx, encounterID)
	if err != nil {
		return dnderr.Wrap(err, "failed to get encounter")
	}

	// Check permissions (DM can always apply damage)
	if !encounter.CanPlayerAct(userID) {
		return dnderr.PermissionDenied("not your turn")
	}

	// Get combatant
	combatant, exists := encounter.Combatants[combatantID]
	if !exists {
		return dnderr.InvalidArgument("combatant not found")
	}

	// Apply damage
	combatant.ApplyDamage(damage)

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
