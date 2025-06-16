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
	
	// GetRace retrieves race information
	GetRace(ctx context.Context, raceKey string) (*entities.Race, error)
	
	// GetClass retrieves class information
	GetClass(ctx context.Context, classKey string) (*entities.Class, error)
	
	// GetRaces retrieves all available races
	GetRaces(ctx context.Context) ([]*entities.Race, error)
	
	// GetClasses retrieves all available classes
	GetClasses(ctx context.Context) ([]*entities.Class, error)
	
	// GetOrCreateDraftCharacter gets an existing draft or creates a new one
	GetOrCreateDraftCharacter(ctx context.Context, userID, realmID string) (*entities.Character, error)
	
	// UpdateDraftCharacter updates a draft character
	UpdateDraftCharacter(ctx context.Context, characterID string, updates *UpdateDraftInput) (*entities.Character, error)
	
	// FinalizeDraftCharacter marks a draft as active
	FinalizeDraftCharacter(ctx context.Context, characterID string) (*entities.Character, error)
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

// UpdateDraftInput contains fields to update on a draft character
type UpdateDraftInput struct {
	RaceKey            *string
	ClassKey           *string
	AbilityScores      map[string]int    // Legacy: direct ability -> score mapping
	AbilityRolls       []entities.AbilityRoll // New: rolls with IDs
	AbilityAssignments map[string]string      // New: ability -> roll ID mapping
	Proficiencies      []string
	Equipment          []string
	Name               *string
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

// GetRace retrieves race information
func (s *service) GetRace(ctx context.Context, raceKey string) (*entities.Race, error) {
	if strings.TrimSpace(raceKey) == "" {
		return nil, dnderr.InvalidArgument("race key is required")
	}
	
	race, err := s.dndClient.GetRace(raceKey)
	if err != nil {
		return nil, dnderr.Wrapf(err, "failed to get race '%s'", raceKey).
			WithMeta("race_key", raceKey)
	}
	
	return race, nil
}

// GetClass retrieves class information
func (s *service) GetClass(ctx context.Context, classKey string) (*entities.Class, error) {
	if strings.TrimSpace(classKey) == "" {
		return nil, dnderr.InvalidArgument("class key is required")
	}
	
	class, err := s.dndClient.GetClass(classKey)
	if err != nil {
		return nil, dnderr.Wrapf(err, "failed to get class '%s'", classKey).
			WithMeta("class_key", classKey)
	}
	
	return class, nil
}

// GetRaces retrieves all available races
func (s *service) GetRaces(ctx context.Context) ([]*entities.Race, error) {
	races, err := s.dndClient.ListRaces()
	if err != nil {
		return nil, dnderr.Wrap(err, "failed to list races")
	}
	
	return races, nil
}

// GetClasses retrieves all available classes
func (s *service) GetClasses(ctx context.Context) ([]*entities.Class, error) {
	classes, err := s.dndClient.ListClasses()
	if err != nil {
		return nil, dnderr.Wrap(err, "failed to list classes")
	}
	
	return classes, nil
}

// GetOrCreateDraftCharacter gets an existing draft or creates a new one
func (s *service) GetOrCreateDraftCharacter(ctx context.Context, userID, realmID string) (*entities.Character, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, dnderr.InvalidArgument("user ID is required")
	}
	if strings.TrimSpace(realmID) == "" {
		return nil, dnderr.InvalidArgument("realm ID is required")
	}
	
	// Look for existing draft characters
	chars, err := s.repository.GetByOwnerAndRealm(ctx, userID, realmID)
	if err != nil {
		return nil, dnderr.Wrap(err, "failed to get existing characters").
			WithMeta("user_id", userID).
			WithMeta("realm_id", realmID)
	}
	
	// Find a draft character
	for _, char := range chars {
		if char.Status == entities.CharacterStatusDraft {
			return char, nil
		}
	}
	
	// No draft found, create a new one
	character := &entities.Character{
		ID:      generateID(),
		OwnerID: userID,
		RealmID: realmID,
		Name:    "Draft Character",
		Status:  entities.CharacterStatusDraft,
		Level:   1,
		// Initialize empty maps
		Attributes:         make(map[entities.Attribute]*entities.AbilityScore),
		Proficiencies:      make(map[entities.ProficiencyType][]*entities.Proficiency),
		Inventory:          make(map[entities.EquipmentType][]entities.Equipment),
		EquippedSlots:      make(map[entities.Slot]entities.Equipment),
	}
	
	// Save to repository
	if err := s.repository.Create(ctx, character); err != nil {
		return nil, dnderr.Wrap(err, "failed to create draft character").
			WithMeta("character_id", character.ID).
			WithMeta("owner_id", userID)
	}
	
	return character, nil
}

// UpdateDraftCharacter updates a draft character
func (s *service) UpdateDraftCharacter(ctx context.Context, characterID string, updates *UpdateDraftInput) (*entities.Character, error) {
	if strings.TrimSpace(characterID) == "" {
		return nil, dnderr.InvalidArgument("character ID is required")
	}
	if updates == nil {
		return nil, dnderr.InvalidArgument("updates cannot be nil")
	}
	
	// Get the character
	char, err := s.repository.Get(ctx, characterID)
	if err != nil {
		return nil, dnderr.Wrapf(err, "failed to get character '%s'", characterID).
			WithMeta("character_id", characterID)
	}
	
	// Verify it's a draft
	if char.Status != entities.CharacterStatusDraft {
		return nil, dnderr.InvalidArgument("can only update draft characters").
			WithMeta("character_id", characterID).
			WithMeta("status", string(char.Status))
	}
	
	// Update race if provided
	if updates.RaceKey != nil {
		race, err := s.dndClient.GetRace(*updates.RaceKey)
		if err != nil {
			return nil, dnderr.Wrapf(err, "failed to get race '%s'", *updates.RaceKey).
				WithMeta("race_key", *updates.RaceKey)
		}
		char.Race = race
		char.Speed = race.Speed
	}
	
	// Update class if provided
	if updates.ClassKey != nil {
		class, err := s.dndClient.GetClass(*updates.ClassKey)
		if err != nil {
			return nil, dnderr.Wrapf(err, "failed to get class '%s'", *updates.ClassKey).
				WithMeta("class_key", *updates.ClassKey)
		}
		char.Class = class
		char.HitDie = class.HitDie
	}
	
	// Update ability rolls if provided
	if len(updates.AbilityRolls) > 0 {
		char.AbilityRolls = updates.AbilityRolls
	}
	
	// Update ability assignments if provided
	if updates.AbilityAssignments != nil {
		char.AbilityAssignments = updates.AbilityAssignments
		
		// Also update the actual ability scores based on assignments
		if len(char.AbilityRolls) > 0 {
			// Create a map of roll ID to value for easy lookup
			rollValues := make(map[string]int)
			for _, roll := range char.AbilityRolls {
				rollValues[roll.ID] = roll.Value
			}
			
			// Clear existing scores
			char.Attributes = make(map[entities.Attribute]*entities.AbilityScore)
			
			// Set new scores based on assignments
			for ability, rollID := range char.AbilityAssignments {
				if score, ok := rollValues[rollID]; ok {
					attr := stringToAttribute(ability)
					char.AddAttribute(attr, score)
				}
			}
			
			// Apply racial bonuses if race is set
			if char.Race != nil {
				for _, bonus := range char.Race.AbilityBonuses {
					char.AddAbilityBonus(bonus)
				}
			}
			
			// Update hit points
			char.SetHitpoints()
		}
	}
	
	// Legacy: Update ability scores if provided directly
	if len(updates.AbilityScores) > 0 && updates.AbilityAssignments == nil {
		// Clear existing scores
		char.Attributes = make(map[entities.Attribute]*entities.AbilityScore)
		
		// Set new scores
		for ability, score := range updates.AbilityScores {
			attr := stringToAttribute(ability)
			char.AddAttribute(attr, score)
		}
		
		// Apply racial bonuses if race is set
		if char.Race != nil {
			for _, bonus := range char.Race.AbilityBonuses {
				char.AddAbilityBonus(bonus)
			}
		}
		
		// Update hit points
		char.SetHitpoints()
	}
	
	// Update name if provided
	if updates.Name != nil {
		char.Name = *updates.Name
	}
	
	// Update proficiencies if provided
	if len(updates.Proficiencies) > 0 {
		// Clear existing chosen proficiencies (keep starting ones from race/class)
		if char.Proficiencies == nil {
			char.Proficiencies = make(map[entities.ProficiencyType][]*entities.Proficiency)
		}
		
		// Add selected proficiencies
		for _, profKey := range updates.Proficiencies {
			proficiency, err := s.dndClient.GetProficiency(profKey)
			if err == nil && proficiency != nil {
				char.AddProficiency(proficiency)
			}
		}
	}
	
	// Save changes
	if err := s.repository.Update(ctx, char); err != nil {
		return nil, dnderr.Wrap(err, "failed to update character").
			WithMeta("character_id", characterID)
	}
	
	return char, nil
}

// FinalizeDraftCharacter marks a draft as active
func (s *service) FinalizeDraftCharacter(ctx context.Context, characterID string) (*entities.Character, error) {
	if strings.TrimSpace(characterID) == "" {
		return nil, dnderr.InvalidArgument("character ID is required")
	}
	
	// Get the character
	char, err := s.repository.Get(ctx, characterID)
	if err != nil {
		return nil, dnderr.Wrapf(err, "failed to get character '%s'", characterID).
			WithMeta("character_id", characterID)
	}
	
	// Verify it's a draft
	if char.Status != entities.CharacterStatusDraft {
		return nil, dnderr.InvalidArgument("can only finalize draft characters").
			WithMeta("character_id", characterID).
			WithMeta("status", string(char.Status))
	}
	
	// Update status to active
	char.Status = entities.CharacterStatusActive
	
	// Save changes
	if err := s.repository.Update(ctx, char); err != nil {
		return nil, dnderr.Wrap(err, "failed to finalize character").
			WithMeta("character_id", characterID)
	}
	
	return char, nil
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