package character

//go:generate mockgen -destination=mock/mock_service.go -package=mockcharacters -source=service.go

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities/features"
	dnderr "github.com/KirkDiggler/dnd-bot-discord/internal/errors"
	characterRepo "github.com/KirkDiggler/dnd-bot-discord/internal/repositories/characters"
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

	// GetEquipmentByCategory retrieves equipment by category (e.g., "martial-weapons")
	GetEquipmentByCategory(ctx context.Context, category string) ([]entities.Equipment, error)

	// UpdateStatus updates a character's status
	UpdateStatus(characterID string, status entities.CharacterStatus) error

	// UpdateEquipment saves equipment changes for a character
	UpdateEquipment(character *entities.Character) error

	// Delete deletes a character
	Delete(characterID string) error

	// ListByOwner lists all characters for a specific owner
	ListByOwner(ownerID string) ([]*entities.Character, error)

	// GetByID retrieves a character by ID
	GetByID(characterID string) (*entities.Character, error)

	// FixCharacterAttributes fixes characters that have AbilityAssignments but no Attributes
	FixCharacterAttributes(ctx context.Context, characterID string) (*entities.Character, error)

	// FinalizeCharacterWithName sets the name and finalizes a draft character in one operation
	FinalizeCharacterWithName(ctx context.Context, characterID, name, raceKey, classKey string) (*entities.Character, error)

	// Character Creation Session methods
	// StartCharacterCreation starts a new character creation session
	StartCharacterCreation(ctx context.Context, userID, guildID string) (*entities.CharacterCreationSession, error)

	// GetCharacterCreationSession retrieves an active session
	GetCharacterCreationSession(ctx context.Context, sessionID string) (*entities.CharacterCreationSession, error)

	// UpdateCharacterCreationSession updates the session step
	UpdateCharacterCreationSession(ctx context.Context, sessionID, step string) error

	// GetCharacterFromSession gets the character associated with a session
	GetCharacterFromSession(ctx context.Context, sessionID string) (*entities.Character, error)

	// StartFreshCharacterCreation gets or creates a draft and clears ability rolls
	StartFreshCharacterCreation(ctx context.Context, userID, realmID string) (*entities.Character, error)
}

// CreateCharacterInput contains all data needed to create a character
type CreateCharacterInput struct {
	UserID        string
	RealmID       string
	Name          string
	RaceKey       string
	ClassKey      string
	AbilityScores map[string]int // STR, DEX, CON, INT, WIS, CHA -> score
	Proficiencies []string       // Selected proficiency keys
	Equipment     []string       // Selected equipment keys
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
	AbilityScores      map[string]int         // Legacy: direct ability -> score mapping
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
	// Temporary in-memory session store (should be Redis in production)
	sessions map[string]*entities.CharacterCreationSession
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
		sessions:   make(map[string]*entities.CharacterCreationSession),
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

	// Skip the in-memory active draft check - let's rely on Redis
	// The in-memory map doesn't work well with Redis persistence

	// Look for existing draft characters
	chars, err := s.repository.GetByOwnerAndRealm(ctx, userID, realmID)
	if err != nil {
		return nil, dnderr.Wrap(err, "failed to get existing characters").
			WithMeta("user_id", userID).
			WithMeta("realm_id", realmID)
	}

	// Find all draft characters
	var drafts []*entities.Character
	for _, char := range chars {
		if char.Status == entities.CharacterStatusDraft {
			drafts = append(drafts, char)
		}
	}

	// If we have any drafts, use the most recent one
	if len(drafts) > 0 {
		// Sort drafts by ID (which includes timestamp) to find most recent
		sort.Slice(drafts, func(i, j int) bool {
			return drafts[i].ID > drafts[j].ID // Descending order
		})

		// Delete any extra drafts
		if len(drafts) > 1 {
			for i := 1; i < len(drafts); i++ {
				if err := s.repository.Delete(ctx, drafts[i].ID); err != nil {
					log.Printf("Failed to delete old draft character %s: %v", drafts[i].ID, err)
				}
			}
		}

		return drafts[0], nil
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
		Attributes:    make(map[entities.Attribute]*entities.AbilityScore),
		Proficiencies: make(map[entities.ProficiencyType][]*entities.Proficiency),
		Inventory:     make(map[entities.EquipmentType][]entities.Equipment),
		EquippedSlots: make(map[entities.Slot]entities.Equipment),
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

	// Log character state BEFORE updates

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
		char.Speed = features.CalculateSpeed(char)

		// Apply racial features
		racialFeatures := features.GetRacialFeatures(race.Key)
		if char.Features == nil {
			char.Features = []*entities.CharacterFeature{}
		}
		// Remove existing racial features
		newFeatures := []*entities.CharacterFeature{}
		for _, f := range char.Features {
			if f.Type != entities.FeatureTypeRacial {
				newFeatures = append(newFeatures, f)
			}
		}
		// Add new racial features
		for _, feat := range racialFeatures {
			newFeatures = append(newFeatures, &feat)
		}
		char.Features = newFeatures

		// Apply passive effects from new racial features
		if err := features.DefaultRegistry.ApplyAllPassiveEffects(char); err != nil {
			log.Printf("Error applying passive effects for character %s (race: %s): %v",
				characterID, race.Name, err)
		}
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

		// Apply class features
		classFeatures := features.GetClassFeatures(class.Key, char.Level)
		if char.Features == nil {
			char.Features = []*entities.CharacterFeature{}
		}
		// Remove existing class features
		newFeatures := []*entities.CharacterFeature{}
		for _, f := range char.Features {
			if f.Type != entities.FeatureTypeClass {
				newFeatures = append(newFeatures, f)
			}
		}
		// Add new class features
		for _, feat := range classFeatures {
			newFeatures = append(newFeatures, &feat)
		}
		char.Features = newFeatures

		// Recalculate AC with new class features
		char.AC = features.CalculateAC(char)
	}

	// Update ability rolls if provided (including clearing with empty slice)
	if updates.AbilityRolls != nil {
		char.AbilityRolls = updates.AbilityRolls
	}

	// Update ability assignments if provided
	if updates.AbilityAssignments != nil {
		char.AbilityAssignments = updates.AbilityAssignments

		// Also update the actual ability scores based on assignments
		if len(char.AbilityRolls) > 0 && len(updates.AbilityAssignments) > 0 {
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

			// Recalculate AC with new ability scores
			char.AC = features.CalculateAC(char)
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

		// Recalculate AC with new ability scores
		char.AC = features.CalculateAC(char)
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

		// Separate proficiencies by type
		skillProficiencies := []*entities.Proficiency{}

		// Get proficiency objects and filter by type
		for _, profKey := range updates.Proficiencies {
			proficiency, err := s.dndClient.GetProficiency(profKey)
			if err == nil && proficiency != nil {
				// Only handle skill proficiencies as replacements
				// Other types (armor, weapon, etc) should accumulate
				if proficiency.Type == entities.ProficiencyTypeSkill {
					skillProficiencies = append(skillProficiencies, proficiency)
				} else {
					// For non-skill proficiencies, add normally (with duplicate check)
					char.AddProficiency(proficiency)
				}
			}
		}

		// Replace all skill proficiencies with the new set
		if len(skillProficiencies) > 0 {
			char.SetProficiencies(entities.ProficiencyTypeSkill, skillProficiencies)
		}

	}

	// Update equipment if provided
	if len(updates.Equipment) > 0 {

		// Initialize inventory if needed
		if char.Inventory == nil {
			char.Inventory = make(map[entities.EquipmentType][]entities.Equipment)
		}

		// Add selected equipment
		for _, equipKey := range updates.Equipment {
			equipment, err := s.dndClient.GetEquipment(equipKey)
			if err == nil && equipment != nil {
				char.AddInventory(equipment)
			}
		}

		// Recalculate AC in case equipment affects it
		char.AC = features.CalculateAC(char)

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

	// Convert AbilityAssignments to Attributes if needed
	if len(char.Attributes) == 0 && len(char.AbilityAssignments) > 0 && len(char.AbilityRolls) > 0 {

		// Create roll ID to value map
		rollValues := make(map[string]int)
		for _, roll := range char.AbilityRolls {
			rollValues[roll.ID] = roll.Value
		}

		// Initialize attributes map
		char.Attributes = make(map[entities.Attribute]*entities.AbilityScore)

		// Convert assignments to attributes
		for abilityStr, rollID := range char.AbilityAssignments {
			rollValue, rollValueOk := rollValues[rollID]
			if rollValueOk {
				// Parse ability string to Attribute type
				var attr entities.Attribute
				switch abilityStr {
				case "STR":
					attr = entities.AttributeStrength
				case "DEX":
					attr = entities.AttributeDexterity
				case "CON":
					attr = entities.AttributeConstitution
				case "INT":
					attr = entities.AttributeIntelligence
				case "WIS":
					attr = entities.AttributeWisdom
				case "CHA":
					attr = entities.AttributeCharisma
				default:
					continue
				}

				// Create base ability score
				score := rollValue

				// Apply racial bonuses
				if char.Race != nil {
					for _, bonus := range char.Race.AbilityBonuses {
						if bonus.Attribute == attr {
							score += bonus.Bonus
						}
					}
				}

				// Calculate modifier
				modifier := (score - 10) / 2

				// Create ability score
				char.Attributes[attr] = &entities.AbilityScore{
					Score: score,
					Bonus: modifier,
				}

			}
		}

	}

	// Calculate hit points if not set
	if char.MaxHitPoints == 0 && char.Class != nil {
		// Base HP = HitDie + Constitution modifier
		conMod := 0
		if char.Attributes != nil {
			if con, ok := char.Attributes[entities.AttributeConstitution]; ok && con != nil {
				conMod = con.Bonus
			}
		}
		char.MaxHitPoints = char.Class.HitDie + conMod
		char.CurrentHitPoints = char.MaxHitPoints
		char.HitDie = char.Class.HitDie
	}

	// Apply features if not already loaded
	if len(char.Features) == 0 {
		char.Features = []*entities.CharacterFeature{}

		// Apply racial features
		if char.Race != nil {
			racialFeatures := features.GetRacialFeatures(char.Race.Key)
			for _, feat := range racialFeatures {
				featCopy := feat // Make a copy to avoid reference issues
				char.Features = append(char.Features, &featCopy)
			}
		}

		// Apply class features
		if char.Class != nil {
			classFeatures := features.GetClassFeatures(char.Class.Key, char.Level)
			for _, feat := range classFeatures {
				featCopy := feat // Make a copy to avoid reference issues
				char.Features = append(char.Features, &featCopy)
			}
		}
	}

	// Apply passive effects from all features
	if err := features.DefaultRegistry.ApplyAllPassiveEffects(char); err != nil {
		// Log error but don't fail finalization
		log.Printf("Error applying passive effects for character %s (%s %s): %v",
			char.ID, char.Race.Name, char.Class.Name, err)
	}

	// Calculate AC using the features package
	char.AC = features.CalculateAC(char)

	// Add starting equipment if class is set and DND client is available
	if s.dndClient != nil && char.Class != nil && char.Class.StartingEquipment != nil {
		for _, se := range char.Class.StartingEquipment {
			if se != nil && se.Equipment != nil {
				equipment, err := s.dndClient.GetEquipment(se.Equipment.Key)
				if err != nil {
					// Log the error but don't fail the finalization
					log.Printf("Failed to get starting equipment %s: %v", se.Equipment.Key, err)
					continue
				}
				// If no error, equipment is valid
				for i := 0; i < se.Quantity; i++ {
					char.AddInventory(equipment)
				}
			}
		}
	}

	// Initialize resources before finalizing
	char.InitializeResources()

	// Update status to active
	char.Status = entities.CharacterStatusActive

	// Save changes
	if err := s.repository.Update(ctx, char); err != nil {
		return nil, dnderr.Wrap(err, "failed to finalize character").
			WithMeta("character_id", characterID)
	}

	return char, nil
}

// GetEquipmentByCategory retrieves equipment by category
func (s *service) GetEquipmentByCategory(ctx context.Context, category string) ([]entities.Equipment, error) {
	if strings.TrimSpace(category) == "" {
		return nil, dnderr.InvalidArgument("category is required")
	}

	equipment, err := s.dndClient.GetEquipmentByCategory(category)
	if err != nil {
		return nil, dnderr.Wrapf(err, "failed to get equipment for category '%s'", category).
			WithMeta("category", category)
	}

	return equipment, nil
}

// generateID generates a unique ID for a character
func generateID() string {
	// TODO: Implement proper ID generation (e.g., UUID or snowflake)
	// For now, use timestamp-based ID
	return fmt.Sprintf("char_%d", time.Now().UnixNano())
}

// UpdateStatus updates a character's status
func (s *service) UpdateStatus(characterID string, status entities.CharacterStatus) error {
	if strings.TrimSpace(characterID) == "" {
		return dnderr.InvalidArgument("character ID is required")
	}

	ctx := context.Background()

	// Get the character
	char, err := s.repository.Get(ctx, characterID)
	if err != nil {
		return dnderr.Wrapf(err, "failed to get character '%s'", characterID).
			WithMeta("character_id", characterID)
	}

	// Update status
	char.Status = status

	// Save changes
	if err := s.repository.Update(ctx, char); err != nil {
		return dnderr.Wrap(err, "failed to update character status").
			WithMeta("character_id", characterID).
			WithMeta("status", string(status))
	}

	return nil
}

// UpdateEquipment saves equipment changes for a character
func (s *service) UpdateEquipment(character *entities.Character) error {
	if character == nil {
		return dnderr.InvalidArgument("character is required")
	}

	if strings.TrimSpace(character.ID) == "" {
		return dnderr.InvalidArgument("character ID is required")
	}

	ctx := context.Background()

	// Recalculate AC with the features calculator
	character.AC = features.CalculateAC(character)

	// Save the character with updated equipment
	if err := s.repository.Update(ctx, character); err != nil {
		return dnderr.Wrap(err, "failed to update character equipment").
			WithMeta("character_id", character.ID)
	}

	return nil
}

// Delete deletes a character
func (s *service) Delete(characterID string) error {
	if strings.TrimSpace(characterID) == "" {
		return dnderr.InvalidArgument("character ID is required")
	}

	ctx := context.Background()

	if err := s.repository.Delete(ctx, characterID); err != nil {
		return dnderr.Wrapf(err, "failed to delete character '%s'", characterID).
			WithMeta("character_id", characterID)
	}

	return nil
}

// ListByOwner lists all characters for a specific owner
func (s *service) ListByOwner(ownerID string) ([]*entities.Character, error) {
	if strings.TrimSpace(ownerID) == "" {
		return nil, dnderr.InvalidArgument("owner ID is required")
	}

	ctx := context.Background()

	chars, err := s.repository.GetByOwner(ctx, ownerID)
	if err != nil {
		return nil, dnderr.Wrapf(err, "failed to list characters for owner '%s'", ownerID).
			WithMeta("owner_id", ownerID)
	}

	return chars, nil
}

// GetByID retrieves a character by ID
func (s *service) GetByID(characterID string) (*entities.Character, error) {
	if strings.TrimSpace(characterID) == "" {
		return nil, dnderr.InvalidArgument("character ID is required")
	}

	ctx := context.Background()

	char, err := s.repository.Get(ctx, characterID)
	if err != nil {
		return nil, dnderr.Wrapf(err, "failed to get character '%s'", characterID).
			WithMeta("character_id", characterID)
	}

	return char, nil
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

// StartFreshCharacterCreation gets or creates a draft and clears ability rolls
func (s *service) StartFreshCharacterCreation(ctx context.Context, userID, realmID string) (*entities.Character, error) {
	draft, err := s.GetOrCreateDraftCharacter(ctx, userID, realmID)
	if err != nil {
		return nil, err
	}

	// Clear ability rolls and assignments if they exist
	if len(draft.AbilityRolls) > 0 || len(draft.AbilityAssignments) > 0 {
		_, err = s.UpdateDraftCharacter(ctx, draft.ID, &UpdateDraftInput{
			AbilityRolls:       []entities.AbilityRoll{},
			AbilityAssignments: map[string]string{},
		})
		if err != nil {
			return nil, dnderr.Wrap(err, "failed to clear ability rolls").
				WithMeta("character_id", draft.ID)
		}

		// Refetch the updated draft
		draft, err = s.GetCharacter(ctx, draft.ID)
		if err != nil {
			return nil, err
		}
	}

	return draft, nil
}
