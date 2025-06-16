package character

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	dnderr "github.com/KirkDiggler/dnd-bot-discord/internal/errors"
	characterRepo "github.com/KirkDiggler/dnd-bot-discord/internal/repositories/character"
)

// Repository is an alias for the character repository interface
type Repository = characterRepo.Repository

// Service defines the character service interface
type Service interface {
	// CreateCharacter creates a new character with the given details
	CreateCharacter(ctx context.Context, input *CreateCharacterInput) (*CreateCharacterOutput, error)
	
	// GetCharacter retrieves a character by ID
	GetCharacter(ctx context.Context, characterID string) (*entities.Character, error)
	
	// ListCharacters lists all characters for a user
	ListCharacters(ctx context.Context, userID string) ([]*entities.Character, error)
	
	// ValidateCharacterCreation validates character creation choices
	ValidateCharacterCreation(ctx context.Context, input *ValidateCharacterInput) error
	
	// ResolveChoices resolves proficiency/equipment choices for a class/race combo
	ResolveChoices(ctx context.Context, input *ResolveChoicesInput) (*ResolveChoicesOutput, error)
}

// CreateCharacterInput contains all data needed to create a character
type CreateCharacterInput struct {
	UserID       string
	RealmID      string
	Name         string
	RaceKey      string
	ClassKey     string
	AbilityScores map[string]int    // STR, DEX, CON, INT, WIS, CHA -> score
	Proficiencies []string          // Selected proficiency keys
	Equipment     []string          // Selected equipment keys
}

// CreateCharacterOutput contains the created character
type CreateCharacterOutput struct {
	Character *entities.Character
}

// ValidateCharacterInput contains character data to validate
type ValidateCharacterInput struct {
	RaceKey       string
	ClassKey      string
	AbilityScores map[string]int
	Proficiencies []string
}

// ResolveChoicesInput asks for available choices for a race/class
type ResolveChoicesInput struct {
	RaceKey  string
	ClassKey string
}

// ResolveChoicesOutput contains simplified choices for the UI
type ResolveChoicesOutput struct {
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
	repository     Repository
	// Later we'll add:
	// validator  Validator
}

// ServiceConfig holds configuration for the service
type ServiceConfig struct {
	DNDClient      dnd5e.Client
	ChoiceResolver ChoiceResolver // Optional, will create default if nil
	Repository     Repository     // Required
}

// NewService creates a new character service
func NewService(cfg *ServiceConfig) Service {
	if cfg.Repository == nil {
		panic("repository is required")
	}
	
	svc := &service{
		dndClient:  cfg.DNDClient,
		repository: cfg.Repository,
	}
	
	// Use provided choice resolver or create default
	if cfg.ChoiceResolver != nil {
		svc.choiceResolver = cfg.ChoiceResolver
	} else {
		svc.choiceResolver = NewChoiceResolver(cfg.DNDClient)
	}
	
	return svc
}

// CreateCharacter creates a new character
func (s *service) CreateCharacter(ctx context.Context, input *CreateCharacterInput) (*CreateCharacterOutput, error) {
	// Validate input
	if err := ValidateInput(input); err != nil {
		return nil, dnderr.Wrap(err, "invalid character creation input").
			WithMeta("operation", "CreateCharacter")
	}
	
	// Get race and class data
	race, err := s.dndClient.GetRace(input.RaceKey)
	if err != nil {
		return nil, dnderr.Wrapf(err, "failed to get race '%s'", input.RaceKey).
			WithMeta("race_key", input.RaceKey)
	}
	
	class, err := s.dndClient.GetClass(input.ClassKey)
	if err != nil {
		return nil, dnderr.Wrapf(err, "failed to get class '%s'", input.ClassKey).
			WithMeta("class_key", input.ClassKey)
	}
	
	// Create the character entity
	character := &entities.Character{
		ID:      generateID(),
		OwnerID: input.UserID,
		RealmID: input.RealmID,
		Name:    input.Name,
		Race:    race,
		Class:   class,
		Status:  entities.CharacterStatusDraft,
		HitDie:  class.HitDie,
		Speed:   race.Speed,
		Level:   1,
	}
	
	// Set ability scores
	character.Attributes = make(map[entities.Attribute]*entities.AbilityScore)
	for ability, score := range input.AbilityScores {
		attr := stringToAttribute(ability)
		character.AddAttribute(attr, score)
	}
	
	// Apply racial bonuses
	for _, bonus := range race.AbilityBonuses {
		character.AddAbilityBonus(bonus)
	}
	
	// Set hit points based on constitution
	character.SetHitpoints()
	
	// Add starting proficiencies from class
	for _, prof := range class.Proficiencies {
		if prof != nil {
			proficiency, err := s.dndClient.GetProficiency(prof.Key)
			if err == nil && proficiency != nil {
				character.AddProficiency(proficiency)
			}
		}
	}
	
	// Add starting proficiencies from race
	for _, prof := range race.StartingProficiencies {
		if prof != nil {
			proficiency, err := s.dndClient.GetProficiency(prof.Key)
			if err == nil && proficiency != nil {
				character.AddProficiency(proficiency)
			}
		}
	}
	
	// Add selected proficiencies
	for _, profKey := range input.Proficiencies {
		proficiency, err := s.dndClient.GetProficiency(profKey)
		if err == nil && proficiency != nil {
			character.AddProficiency(proficiency)
		}
	}
	
	// Add starting equipment
	for _, se := range class.StartingEquipment {
		if se != nil && se.Equipment != nil {
			equipment, err := s.dndClient.GetEquipment(se.Equipment.Key)
			if err == nil && equipment != nil {
				for i := 0; i < se.Quantity; i++ {
					character.AddInventory(equipment)
				}
			}
		}
	}
	
	// Add selected equipment
	for _, equipKey := range input.Equipment {
		equipment, err := s.dndClient.GetEquipment(equipKey)
		if err == nil && equipment != nil {
			character.AddInventory(equipment)
		}
	}
	
	// Save to repository
	if err := s.repository.Create(ctx, character); err != nil {
		return nil, dnderr.Wrap(err, "failed to save character").
			WithMeta("character_id", character.ID).
			WithMeta("character_name", character.Name).
			WithMeta("owner_id", character.OwnerID)
	}
	
	return &CreateCharacterOutput{
		Character: character,
	}, nil
}

// GetCharacter retrieves a character by ID
func (s *service) GetCharacter(ctx context.Context, characterID string) (*entities.Character, error) {
	if strings.TrimSpace(characterID) == "" {
		return nil, dnderr.InvalidArgument("character ID is required")
	}
	
	char, err := s.repository.Get(ctx, characterID)
	if err != nil {
		return nil, dnderr.Wrapf(err, "failed to get character '%s'", characterID).
			WithMeta("character_id", characterID)
	}
	
	return char, nil
}

// ListCharacters lists all characters for a user
func (s *service) ListCharacters(ctx context.Context, userID string) ([]*entities.Character, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, dnderr.InvalidArgument("user ID is required")
	}
	
	chars, err := s.repository.GetByOwner(ctx, userID)
	if err != nil {
		return nil, dnderr.Wrapf(err, "failed to list characters for user '%s'", userID).
			WithMeta("user_id", userID)
	}
	
	return chars, nil
}

// ValidateCharacterCreation validates character creation choices
func (s *service) ValidateCharacterCreation(ctx context.Context, input *ValidateCharacterInput) error {
	// Use the Validate method
	if err := ValidateInput(input); err != nil {
		return dnderr.Wrap(err, "validation failed").
			WithMeta("operation", "ValidateCharacterCreation")
	}
	
	// TODO: Additional validation
	// - Check proficiencies match available choices from race/class
	// - Check all required choices are made
	
	return nil
}

// ResolveChoices resolves proficiency/equipment choices for a class/race combo
func (s *service) ResolveChoices(ctx context.Context, input *ResolveChoicesInput) (*ResolveChoicesOutput, error) {
	// Validate input
	if err := ValidateInput(input); err != nil {
		return nil, dnderr.Wrap(err, "invalid resolve choices input").
			WithMeta("operation", "ResolveChoices")
	}
	
	// Fetch race and class data
	race, err := s.dndClient.GetRace(input.RaceKey)
	if err != nil {
		return nil, dnderr.Wrapf(err, "failed to get race '%s'", input.RaceKey).
			WithMeta("race_key", input.RaceKey)
	}
	
	class, err := s.dndClient.GetClass(input.ClassKey)
	if err != nil {
		return nil, dnderr.Wrapf(err, "failed to get class '%s'", input.ClassKey).
			WithMeta("class_key", input.ClassKey)
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
	
	return &ResolveChoicesOutput{
		ProficiencyChoices: proficiencyChoices,
		EquipmentChoices:   equipmentChoices,
	}, nil
}

// generateID generates a unique ID for a character
func generateID() string {
	// TODO: Implement proper ID generation (e.g., UUID or snowflake)
	// For now, use timestamp-based ID
	return fmt.Sprintf("char_%d", time.Now().UnixNano())
}

// stringToAttribute converts a string to an Attribute
func stringToAttribute(s string) entities.Attribute {
	switch strings.ToUpper(s) {
	case "STR":
		return entities.AttributeStrength
	case "DEX":
		return entities.AttributeDexterity
	case "CON":
		return entities.AttributeConstitution
	case "INT":
		return entities.AttributeIntelligence
	case "WIS":
		return entities.AttributeWisdom
	case "CHA":
		return entities.AttributeCharisma
	default:
		return entities.AttributeNone
	}
}