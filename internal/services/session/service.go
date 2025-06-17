package session

//go:generate mockgen -destination=mock/mock_service.go -package=mocksession -source=service.go

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	dnderr "github.com/KirkDiggler/dnd-bot-discord/internal/errors"
	"github.com/KirkDiggler/dnd-bot-discord/internal/repositories/gamesessions"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/uuid"
)

// Repository is an alias for the game session repository interface
type Repository = gamesessions.Repository

// Service defines the session service interface
type Service interface {
	// CreateSession creates a new game session
	CreateSession(ctx context.Context, input *CreateSessionInput) (*entities.Session, error)

	// GetSession retrieves a session by ID
	GetSession(ctx context.Context, sessionID string) (*entities.Session, error)

	// GetSessionByInviteCode retrieves a session by invite code
	GetSessionByInviteCode(ctx context.Context, code string) (*entities.Session, error)

	// UpdateSession updates session details
	UpdateSession(ctx context.Context, sessionID string, input *UpdateSessionInput) (*entities.Session, error)

	// DeleteSession removes a session
	DeleteSession(ctx context.Context, sessionID string) error

	// InviteToSession sends an invite to a user
	InviteToSession(ctx context.Context, sessionID, inviterID, inviteeID string) error

	// JoinSession adds a user to a session
	JoinSession(ctx context.Context, sessionID, userID string) (*entities.SessionMember, error)

	// JoinSessionByCode adds a user to a session using invite code
	JoinSessionByCode(ctx context.Context, code, userID string) (*entities.SessionMember, error)

	// LeaveSession removes a user from a session
	LeaveSession(ctx context.Context, sessionID, userID string) error

	// SelectCharacter assigns a character to a session member
	SelectCharacter(ctx context.Context, sessionID, userID, characterID string) error

	// StartSession begins a game session
	StartSession(ctx context.Context, sessionID, userID string) error

	// EndSession concludes a game session
	EndSession(ctx context.Context, sessionID, userID string) error

	// PauseSession pauses a game session
	PauseSession(ctx context.Context, sessionID, userID string) error

	// ResumeSession resumes a paused session
	ResumeSession(ctx context.Context, sessionID, userID string) error

	// ListUserSessions lists all sessions for a user
	ListUserSessions(ctx context.Context, userID string) ([]*entities.Session, error)

	// ListRealmSessions lists all sessions for a realm
	ListRealmSessions(ctx context.Context, realmID string) ([]*entities.Session, error)

	// ListActiveUserSessions lists active sessions for a user
	ListActiveUserSessions(ctx context.Context, userID string) ([]*entities.Session, error)

	// ListActiveRealmSessions lists active sessions for a realm
	ListActiveRealmSessions(ctx context.Context, realmID string) ([]*entities.Session, error)
}

// CreateSessionInput contains data for creating a session
type CreateSessionInput struct {
	Name        string
	Description string
	RealmID     string
	ChannelID   string
	CreatorID   string
	Settings    *entities.SessionSettings // Optional, will use defaults if nil
}

// UpdateSessionInput contains fields that can be updated
type UpdateSessionInput struct {
	Name        *string
	Description *string
	Settings    *entities.SessionSettings
}

// service implements the Service interface
type service struct {
	repository       Repository
	characterService character.Service
	uuidGenerator    uuid.Generator
}

// ServiceConfig holds configuration for the service
type ServiceConfig struct {
	Repository       Repository        // Required
	CharacterService character.Service // Required
	UUIDGenerator    uuid.Generator    // Optional, will use default if nil
}

// NewService creates a new session service
func NewService(cfg *ServiceConfig) Service {
	if cfg.Repository == nil {
		panic("repository is required")
	}
	if cfg.CharacterService == nil {
		panic("character service is required")
	}

	svc := &service{
		repository:       cfg.Repository,
		characterService: cfg.CharacterService,
	}

	// Use provided UUID generator or create default
	if cfg.UUIDGenerator != nil {
		svc.uuidGenerator = cfg.UUIDGenerator
	} else {
		svc.uuidGenerator = uuid.NewGoogleUUIDGenerator()
	}

	return svc
}

// CreateSession creates a new game session
func (s *service) CreateSession(ctx context.Context, input *CreateSessionInput) (*entities.Session, error) {
	if input == nil {
		return nil, dnderr.InvalidArgument("input cannot be nil")
	}

	// Validate input
	if strings.TrimSpace(input.Name) == "" {
		return nil, dnderr.InvalidArgument("session name is required")
	}
	if strings.TrimSpace(input.RealmID) == "" {
		return nil, dnderr.InvalidArgument("realm ID is required")
	}
	if strings.TrimSpace(input.ChannelID) == "" {
		return nil, dnderr.InvalidArgument("channel ID is required")
	}
	if strings.TrimSpace(input.CreatorID) == "" {
		return nil, dnderr.InvalidArgument("creator ID is required")
	}

	// Generate session ID
	sessionID := s.uuidGenerator.New()

	// Generate invite code
	inviteCode := generateInviteCode()

	// Create session
	session := entities.NewSession(sessionID, input.Name, input.RealmID, input.ChannelID, input.CreatorID)
	session.Description = input.Description
	session.InviteCode = inviteCode

	// Apply custom settings if provided
	if input.Settings != nil {
		session.Settings = input.Settings
	}

	// Add creator as DM
	session.AddMember(input.CreatorID, entities.SessionRoleDM)

	// Save to repository
	if err := s.repository.Create(ctx, session); err != nil {
		return nil, dnderr.Wrap(err, "failed to create session").
			WithMeta("session_id", sessionID).
			WithMeta("session_name", input.Name)
	}

	return session, nil
}

// GetSession retrieves a session by ID
func (s *service) GetSession(ctx context.Context, sessionID string) (*entities.Session, error) {
	if strings.TrimSpace(sessionID) == "" {
		return nil, dnderr.InvalidArgument("session ID is required")
	}

	session, err := s.repository.Get(ctx, sessionID)
	if err != nil {
		return nil, dnderr.Wrapf(err, "failed to get session '%s'", sessionID).
			WithMeta("session_id", sessionID)
	}

	return session, nil
}

// GetSessionByInviteCode retrieves a session by invite code
func (s *service) GetSessionByInviteCode(ctx context.Context, code string) (*entities.Session, error) {
	if strings.TrimSpace(code) == "" {
		return nil, dnderr.InvalidArgument("invite code is required")
	}

	session, err := s.repository.GetByInviteCode(ctx, code)
	if err != nil {
		return nil, dnderr.Wrapf(err, "failed to get session by invite code '%s'", code).
			WithMeta("invite_code", code)
	}

	return session, nil
}

// UpdateSession updates session details
func (s *service) UpdateSession(ctx context.Context, sessionID string, input *UpdateSessionInput) (*entities.Session, error) {
	if strings.TrimSpace(sessionID) == "" {
		return nil, dnderr.InvalidArgument("session ID is required")
	}
	if input == nil {
		return nil, dnderr.InvalidArgument("input cannot be nil")
	}

	// Get existing session
	session, err := s.repository.Get(ctx, sessionID)
	if err != nil {
		return nil, dnderr.Wrapf(err, "failed to get session '%s'", sessionID).
			WithMeta("session_id", sessionID)
	}

	// Apply updates
	if input.Name != nil {
		session.Name = *input.Name
	}
	if input.Description != nil {
		session.Description = *input.Description
	}
	if input.Settings != nil {
		session.Settings = input.Settings
	}

	session.UpdateActivity()

	// Save changes
	if err := s.repository.Update(ctx, session); err != nil {
		return nil, dnderr.Wrap(err, "failed to update session").
			WithMeta("session_id", sessionID)
	}

	return session, nil
}

// DeleteSession removes a session
func (s *service) DeleteSession(ctx context.Context, sessionID string) error {
	if strings.TrimSpace(sessionID) == "" {
		return dnderr.InvalidArgument("session ID is required")
	}

	if err := s.repository.Delete(ctx, sessionID); err != nil {
		return dnderr.Wrapf(err, "failed to delete session '%s'", sessionID).
			WithMeta("session_id", sessionID)
	}

	return nil
}

// InviteToSession sends an invite to a user (permission check needed)
func (s *service) InviteToSession(ctx context.Context, sessionID, inviterID, inviteeID string) error {
	if strings.TrimSpace(sessionID) == "" {
		return dnderr.InvalidArgument("session ID is required")
	}
	if strings.TrimSpace(inviterID) == "" {
		return dnderr.InvalidArgument("inviter ID is required")
	}
	if strings.TrimSpace(inviteeID) == "" {
		return dnderr.InvalidArgument("invitee ID is required")
	}

	// Get session
	session, err := s.repository.Get(ctx, sessionID)
	if err != nil {
		return dnderr.Wrapf(err, "failed to get session '%s'", sessionID).
			WithMeta("session_id", sessionID)
	}

	// Check if inviter is DM
	inviter, exists := session.Members[inviterID]
	if !exists || inviter.Role != entities.SessionRoleDM {
		return dnderr.PermissionDenied("only the DM can invite players").
			WithMeta("inviter_id", inviterID).
			WithMeta("session_id", sessionID)
	}

	// Check if invitee is already in session
	if session.IsUserInSession(inviteeID) {
		return dnderr.InvalidArgument("user is already in the session").
			WithMeta("invitee_id", inviteeID).
			WithMeta("session_id", sessionID)
	}

	// In a real implementation, this would send a notification to the invitee
	// For now, we just return success and they can join with the code

	return nil
}

// JoinSession adds a user to a session
func (s *service) JoinSession(ctx context.Context, sessionID, userID string) (*entities.SessionMember, error) {
	if strings.TrimSpace(sessionID) == "" {
		return nil, dnderr.InvalidArgument("session ID is required")
	}
	if strings.TrimSpace(userID) == "" {
		return nil, dnderr.InvalidArgument("user ID is required")
	}

	// Get session
	session, err := s.repository.Get(ctx, sessionID)
	if err != nil {
		return nil, dnderr.Wrapf(err, "failed to get session '%s'", sessionID).
			WithMeta("session_id", sessionID)
	}

	// Check if user is already in session
	if session.IsUserInSession(userID) {
		return nil, dnderr.InvalidArgument("user is already in the session").
			WithMeta("user_id", userID).
			WithMeta("session_id", sessionID)
	}

	// Check if session can be joined
	if !session.CanJoin() {
		return nil, dnderr.InvalidArgument("cannot join this session").
			WithMeta("session_id", sessionID).
			WithMeta("status", string(session.Status))
	}

	// Add as player by default
	member := session.AddMember(userID, entities.SessionRolePlayer)

	// Save changes
	if err := s.repository.Update(ctx, session); err != nil {
		return nil, dnderr.Wrap(err, "failed to update session").
			WithMeta("session_id", sessionID)
	}

	return member, nil
}

// JoinSessionByCode adds a user to a session using invite code
func (s *service) JoinSessionByCode(ctx context.Context, code, userID string) (*entities.SessionMember, error) {
	if strings.TrimSpace(code) == "" {
		return nil, dnderr.InvalidArgument("invite code is required")
	}
	if strings.TrimSpace(userID) == "" {
		return nil, dnderr.InvalidArgument("user ID is required")
	}

	// Get session by invite code
	session, err := s.repository.GetByInviteCode(ctx, code)
	if err != nil {
		return nil, dnderr.Wrapf(err, "failed to get session by invite code '%s'", code).
			WithMeta("invite_code", code)
	}

	// Use JoinSession to handle the actual joining
	return s.JoinSession(ctx, session.ID, userID)
}

// LeaveSession removes a user from a session
func (s *service) LeaveSession(ctx context.Context, sessionID, userID string) error {
	if strings.TrimSpace(sessionID) == "" {
		return dnderr.InvalidArgument("session ID is required")
	}
	if strings.TrimSpace(userID) == "" {
		return dnderr.InvalidArgument("user ID is required")
	}

	// Get session
	session, err := s.repository.Get(ctx, sessionID)
	if err != nil {
		return dnderr.Wrapf(err, "failed to get session '%s'", sessionID).
			WithMeta("session_id", sessionID)
	}

	// Check if user is in session
	member, exists := session.Members[userID]
	if !exists {
		return dnderr.InvalidArgument("user is not in the session").
			WithMeta("user_id", userID).
			WithMeta("session_id", sessionID)
	}

	// Don't allow DM to leave unless ending session (skip this for planning sessions)
	if member.Role == entities.SessionRoleDM && session.Status != entities.SessionStatusEnded && session.Status != entities.SessionStatusPlanning {
		return dnderr.InvalidArgument("DM cannot leave an active session").
			WithMeta("user_id", userID).
			WithMeta("session_id", sessionID)
	}

	// Remove member
	session.RemoveMember(userID)

	// Save changes
	if err := s.repository.Update(ctx, session); err != nil {
		return dnderr.Wrap(err, "failed to update session").
			WithMeta("session_id", sessionID)
	}

	return nil
}

// SelectCharacter assigns a character to a session member
func (s *service) SelectCharacter(ctx context.Context, sessionID, userID, characterID string) error {
	if strings.TrimSpace(sessionID) == "" {
		return dnderr.InvalidArgument("session ID is required")
	}
	if strings.TrimSpace(userID) == "" {
		return dnderr.InvalidArgument("user ID is required")
	}
	if strings.TrimSpace(characterID) == "" {
		return dnderr.InvalidArgument("character ID is required")
	}

	// Verify character exists and belongs to user
	character, err := s.characterService.GetByID(characterID)
	if err != nil {
		return dnderr.Wrapf(err, "failed to get character '%s'", characterID).
			WithMeta("character_id", characterID)
	}

	if character.OwnerID != userID {
		return dnderr.PermissionDenied("character does not belong to user").
			WithMeta("user_id", userID).
			WithMeta("character_id", characterID)
	}

	// Get session
	session, err := s.repository.Get(ctx, sessionID)
	if err != nil {
		return dnderr.Wrapf(err, "failed to get session '%s'", sessionID).
			WithMeta("session_id", sessionID)
	}

	// Set character
	if !session.SetCharacter(userID, characterID) {
		return dnderr.InvalidArgument("user is not in the session").
			WithMeta("user_id", userID).
			WithMeta("session_id", sessionID)
	}

	// Save changes
	if err := s.repository.Update(ctx, session); err != nil {
		return dnderr.Wrap(err, "failed to update session").
			WithMeta("session_id", sessionID)
	}

	return nil
}

// StartSession begins a game session
func (s *service) StartSession(ctx context.Context, sessionID, userID string) error {
	if strings.TrimSpace(sessionID) == "" {
		return dnderr.InvalidArgument("session ID is required")
	}
	if strings.TrimSpace(userID) == "" {
		return dnderr.InvalidArgument("user ID is required")
	}

	// Get session
	session, err := s.repository.Get(ctx, sessionID)
	if err != nil {
		return dnderr.Wrapf(err, "failed to get session '%s'", sessionID).
			WithMeta("session_id", sessionID)
	}

	// Check if user is DM
	member, exists := session.Members[userID]
	if !exists || member.Role != entities.SessionRoleDM {
		return dnderr.PermissionDenied("only the DM can start the session").
			WithMeta("user_id", userID).
			WithMeta("session_id", sessionID)
	}

	// Start session
	if !session.Start() {
		return dnderr.InvalidArgument("cannot start session in current state").
			WithMeta("session_id", sessionID).
			WithMeta("status", string(session.Status))
	}

	// Save changes
	if err := s.repository.Update(ctx, session); err != nil {
		return dnderr.Wrap(err, "failed to update session").
			WithMeta("session_id", sessionID)
	}

	return nil
}

// EndSession concludes a game session
func (s *service) EndSession(ctx context.Context, sessionID, userID string) error {
	if strings.TrimSpace(sessionID) == "" {
		return dnderr.InvalidArgument("session ID is required")
	}
	if strings.TrimSpace(userID) == "" {
		return dnderr.InvalidArgument("user ID is required")
	}

	// Get session
	session, err := s.repository.Get(ctx, sessionID)
	if err != nil {
		return dnderr.Wrapf(err, "failed to get session '%s'", sessionID).
			WithMeta("session_id", sessionID)
	}

	// Check if user is DM
	member, exists := session.Members[userID]
	if !exists || member.Role != entities.SessionRoleDM {
		return dnderr.PermissionDenied("only the DM can end the session").
			WithMeta("user_id", userID).
			WithMeta("session_id", sessionID)
	}

	// End session
	if !session.End() {
		return dnderr.InvalidArgument("session is already ended").
			WithMeta("session_id", sessionID).
			WithMeta("status", string(session.Status))
	}

	// Save changes
	if err := s.repository.Update(ctx, session); err != nil {
		return dnderr.Wrap(err, "failed to update session").
			WithMeta("session_id", sessionID)
	}

	return nil
}

// PauseSession pauses a game session
func (s *service) PauseSession(ctx context.Context, sessionID, userID string) error {
	if strings.TrimSpace(sessionID) == "" {
		return dnderr.InvalidArgument("session ID is required")
	}
	if strings.TrimSpace(userID) == "" {
		return dnderr.InvalidArgument("user ID is required")
	}

	// Get session
	session, err := s.repository.Get(ctx, sessionID)
	if err != nil {
		return dnderr.Wrapf(err, "failed to get session '%s'", sessionID).
			WithMeta("session_id", sessionID)
	}

	// Check if user is DM
	member, exists := session.Members[userID]
	if !exists || member.Role != entities.SessionRoleDM {
		return dnderr.PermissionDenied("only the DM can pause the session").
			WithMeta("user_id", userID).
			WithMeta("session_id", sessionID)
	}

	// Pause session
	if !session.Pause() {
		return dnderr.InvalidArgument("cannot pause session in current state").
			WithMeta("session_id", sessionID).
			WithMeta("status", string(session.Status))
	}

	// Save changes
	if err := s.repository.Update(ctx, session); err != nil {
		return dnderr.Wrap(err, "failed to update session").
			WithMeta("session_id", sessionID)
	}

	return nil
}

// ResumeSession resumes a paused session
func (s *service) ResumeSession(ctx context.Context, sessionID, userID string) error {
	if strings.TrimSpace(sessionID) == "" {
		return dnderr.InvalidArgument("session ID is required")
	}
	if strings.TrimSpace(userID) == "" {
		return dnderr.InvalidArgument("user ID is required")
	}

	// Get session
	session, err := s.repository.Get(ctx, sessionID)
	if err != nil {
		return dnderr.Wrapf(err, "failed to get session '%s'", sessionID).
			WithMeta("session_id", sessionID)
	}

	// Check if user is DM
	member, exists := session.Members[userID]
	if !exists || member.Role != entities.SessionRoleDM {
		return dnderr.PermissionDenied("only the DM can resume the session").
			WithMeta("user_id", userID).
			WithMeta("session_id", sessionID)
	}

	// Resume session
	if !session.Resume() {
		return dnderr.InvalidArgument("cannot resume session in current state").
			WithMeta("session_id", sessionID).
			WithMeta("status", string(session.Status))
	}

	// Save changes
	if err := s.repository.Update(ctx, session); err != nil {
		return dnderr.Wrap(err, "failed to update session").
			WithMeta("session_id", sessionID)
	}

	return nil
}

// ListUserSessions lists all sessions for a user
func (s *service) ListUserSessions(ctx context.Context, userID string) ([]*entities.Session, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, dnderr.InvalidArgument("user ID is required")
	}

	sessions, err := s.repository.GetByUser(ctx, userID)
	if err != nil {
		return nil, dnderr.Wrapf(err, "failed to list sessions for user '%s'", userID).
			WithMeta("user_id", userID)
	}

	return sessions, nil
}

// ListRealmSessions lists all sessions for a realm
func (s *service) ListRealmSessions(ctx context.Context, realmID string) ([]*entities.Session, error) {
	if strings.TrimSpace(realmID) == "" {
		return nil, dnderr.InvalidArgument("realm ID is required")
	}

	sessions, err := s.repository.GetByRealm(ctx, realmID)
	if err != nil {
		return nil, dnderr.Wrapf(err, "failed to list sessions for realm '%s'", realmID).
			WithMeta("realm_id", realmID)
	}

	return sessions, nil
}

// ListActiveUserSessions lists active sessions for a user
func (s *service) ListActiveUserSessions(ctx context.Context, userID string) ([]*entities.Session, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, dnderr.InvalidArgument("user ID is required")
	}

	sessions, err := s.repository.GetActiveByUser(ctx, userID)
	if err != nil {
		return nil, dnderr.Wrapf(err, "failed to list active sessions for user '%s'", userID).
			WithMeta("user_id", userID)
	}

	return sessions, nil
}

// ListActiveRealmSessions lists active sessions for a realm
func (s *service) ListActiveRealmSessions(ctx context.Context, realmID string) ([]*entities.Session, error) {
	if strings.TrimSpace(realmID) == "" {
		return nil, dnderr.InvalidArgument("realm ID is required")
	}

	sessions, err := s.repository.GetActiveByRealm(ctx, realmID)
	if err != nil {
		return nil, dnderr.Wrapf(err, "failed to list active sessions for realm '%s'", realmID).
			WithMeta("realm_id", realmID)
	}

	return sessions, nil
}

// generateInviteCode generates a unique invite code
func generateInviteCode() string {
	// Generate 4 bytes (8 hex chars)
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp-based code
		return fmt.Sprintf("INV%d", time.Now().Unix())
	}
	return strings.ToUpper(hex.EncodeToString(b))
}
