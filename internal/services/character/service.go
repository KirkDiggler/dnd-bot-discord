package character

import (
	"context"
	"fmt"

	"github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
)

// Service defines the character service interface
type Service interface {
	// CreateCharacter creates a new character with the given details
	CreateCharacter(ctx context.Context, req *CreateCharacterRequest) (*CreateCharacterResponse, error)
	
	// GetCharacter retrieves a character by ID
	GetCharacter(ctx context.Context, characterID string) (*entities.Character, error)
	
	// ListCharacters lists all characters for a user
	ListCharacters(ctx context.Context, userID string) ([]*entities.Character, error)
	
	// ValidateCharacterCreation validates character creation choices
	ValidateCharacterCreation(ctx context.Context, req *ValidateCharacterRequest) error
	
	// ResolveChoices resolves proficiency/equipment choices for a class/race combo
	ResolveChoices(ctx context.Context, req *ResolveChoicesRequest) (*ResolveChoicesResponse, error)
}

// CreateCharacterRequest contains all data needed to create a character
type CreateCharacterRequest struct {
	UserID       string
	GuildID      string
	Name         string
	RaceKey      string
	ClassKey     string
	AbilityScores map[string]int    // STR, DEX, CON, INT, WIS, CHA -> score
	Proficiencies []string          // Selected proficiency keys
	Equipment     []string          // Selected equipment keys
}

// CreateCharacterResponse contains the created character
type CreateCharacterResponse struct {
	Character *entities.Character
}

// ValidateCharacterRequest contains character data to validate
type ValidateCharacterRequest struct {
	RaceKey       string
	ClassKey      string
	AbilityScores map[string]int
	Proficiencies []string
}

// ResolveChoicesRequest asks for available choices for a race/class
type ResolveChoicesRequest struct {
	RaceKey  string
	ClassKey string
}

// ResolveChoicesResponse contains simplified choices for the UI
type ResolveChoicesResponse struct {
	ProficiencyChoices []SimplifiedChoice
	EquipmentChoices   []SimplifiedChoice
}

// SimplifiedChoice represents a choice in a UI-friendly format
type SimplifiedChoice struct {
	ID          string
	Name        string
	Description string
	Type        string // "skill", "tool", "language", "equipment"
	Choose      int    // How many to choose
	Options     []ChoiceOption
}

// ChoiceOption represents a single option within a choice
type ChoiceOption struct {
	Key         string
	Name        string
	Description string
}

// service implements the Service interface
type service struct {
	dndClient      dnd5e.Client
	choiceResolver ChoiceResolver
	// Later we'll add:
	// repository Repository
	// validator  Validator
}

// ServiceConfig holds configuration for the service
type ServiceConfig struct {
	DNDClient dnd5e.Client
}

// NewService creates a new character service
func NewService(cfg *ServiceConfig) Service {
	return &service{
		dndClient:      cfg.DNDClient,
		choiceResolver: NewChoiceResolver(cfg.DNDClient),
	}
}

// CreateCharacter creates a new character
func (s *service) CreateCharacter(ctx context.Context, req *CreateCharacterRequest) (*CreateCharacterResponse, error) {
	// TODO: Implement character creation logic
	// For now, return a mock response
	
	// Validate the request
	if err := s.ValidateCharacterCreation(ctx, &ValidateCharacterRequest{
		RaceKey:       req.RaceKey,
		ClassKey:      req.ClassKey,
		AbilityScores: req.AbilityScores,
		Proficiencies: req.Proficiencies,
	}); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}
	
	// Create the character entity
	character := &entities.Character{
		ID:      generateID(), // We'll implement this
		OwnerID: req.UserID,
		Name:    req.Name,
		// We'll populate more fields as we implement
	}
	
	// TODO: Save to repository
	
	return &CreateCharacterResponse{
		Character: character,
	}, nil
}

// GetCharacter retrieves a character by ID
func (s *service) GetCharacter(ctx context.Context, characterID string) (*entities.Character, error) {
	// TODO: Implement character retrieval
	return nil, fmt.Errorf("not implemented")
}

// ListCharacters lists all characters for a user
func (s *service) ListCharacters(ctx context.Context, userID string) ([]*entities.Character, error) {
	// TODO: Implement character listing
	return nil, fmt.Errorf("not implemented")
}

// ValidateCharacterCreation validates character creation choices
func (s *service) ValidateCharacterCreation(ctx context.Context, req *ValidateCharacterRequest) error {
	// TODO: Implement validation logic
	// - Check ability scores are valid (3-18 before racial bonuses)
	// - Check proficiencies match available choices
	// - Check all required choices are made
	
	// For now, just basic validation
	if req.RaceKey == "" {
		return fmt.Errorf("race is required")
	}
	
	if req.ClassKey == "" {
		return fmt.Errorf("class is required")
	}
	
	// Validate ability scores
	requiredAbilities := []string{"STR", "DEX", "CON", "INT", "WIS", "CHA"}
	for _, ability := range requiredAbilities {
		score, ok := req.AbilityScores[ability]
		if !ok {
			return fmt.Errorf("missing ability score for %s", ability)
		}
		if score < 3 || score > 18 {
			return fmt.Errorf("ability score for %s must be between 3 and 18", ability)
		}
	}
	
	return nil
}

// ResolveChoices resolves proficiency/equipment choices for a class/race combo
func (s *service) ResolveChoices(ctx context.Context, req *ResolveChoicesRequest) (*ResolveChoicesResponse, error) {
	// Fetch race and class data
	race, err := s.dndClient.GetRace(req.RaceKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get race: %w", err)
	}
	
	class, err := s.dndClient.GetClass(req.ClassKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get class: %w", err)
	}
	
	// Resolve proficiency choices
	proficiencyChoices, err := s.choiceResolver.ResolveProficiencyChoices(ctx, race, class)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve proficiency choices: %w", err)
	}
	
	// Resolve equipment choices
	equipmentChoices, err := s.choiceResolver.ResolveEquipmentChoices(ctx, class)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve equipment choices: %w", err)
	}
	
	return &ResolveChoicesResponse{
		ProficiencyChoices: proficiencyChoices,
		EquipmentChoices:   equipmentChoices,
	}, nil
}

// generateID generates a unique ID for a character
func generateID() string {
	// TODO: Implement proper ID generation
	// For now, return a placeholder
	return "char_123456"
}