package character

//go:generate mockgen -destination=mock/mock_service.go -package=mockcharacters -source=service.go

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	charDomain "github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/equipment"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
	features2 "github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e/features"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"

	"github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e"
	dnderr "github.com/KirkDiggler/dnd-bot-discord/internal/errors"
	draftRepo "github.com/KirkDiggler/dnd-bot-discord/internal/repositories/character_draft"
	characterRepo "github.com/KirkDiggler/dnd-bot-discord/internal/repositories/characters"
)

// Repository is an alias for the character repository interface
type Repository = characterRepo.Repository

// Service defines the character service interface
type Service interface {
	// CreateCharacter creates a new character with the given details
	CreateCharacter(ctx context.Context, input *CreateCharacterInput) (*CreateCharacterOutput, error)

	// GetCharacter retrieves a character by ID
	GetCharacter(ctx context.Context, characterID string) (*charDomain.Character, error)

	// ListCharacters lists all characters for a user
	ListCharacters(ctx context.Context, userID string) ([]*charDomain.Character, error)

	// ValidateCharacterCreation validates character creation choices
	ValidateCharacterCreation(ctx context.Context, input *ValidateCharacterInput) error

	// ResolveChoices resolves proficiency/equipment choices for a class/race combo
	ResolveChoices(ctx context.Context, input *ResolveChoicesInput) (*ResolveChoicesOutput, error)

	// GetRace retrieves race information
	GetRace(ctx context.Context, raceKey string) (*rulebook.Race, error)

	// GetClass retrieves class information
	GetClass(ctx context.Context, classKey string) (*rulebook.Class, error)

	// GetRaces retrieves all available races
	GetRaces(ctx context.Context) ([]*rulebook.Race, error)

	// GetClasses retrieves all available classes
	GetClasses(ctx context.Context) ([]*rulebook.Class, error)

	// GetOrCreateDraftCharacter gets an existing draft or creates a new one
	GetOrCreateDraftCharacter(ctx context.Context, userID, realmID string) (*charDomain.Character, error)

	// UpdateDraftCharacter updates a draft character
	UpdateDraftCharacter(ctx context.Context, characterID string, updates *UpdateDraftInput) (*charDomain.Character, error)

	// FinalizeDraftCharacter marks a draft as active
	FinalizeDraftCharacter(ctx context.Context, characterID string) (*charDomain.Character, error)

	// GetEquipmentByCategory retrieves equipment by category (e.g., "martial-weapons")
	GetEquipmentByCategory(ctx context.Context, category string) ([]equipment.Equipment, error)

	// UpdateStatus updates a character's status
	UpdateStatus(characterID string, status shared.CharacterStatus) error

	// UpdateEquipment saves equipment changes for a character
	UpdateEquipment(character *charDomain.Character) error

	// GetPendingFeatureChoices returns feature choices that need to be made for a character
	GetPendingFeatureChoices(ctx context.Context, characterID string) ([]*rulebook.FeatureChoice, error)

	// Delete deletes a character
	Delete(characterID string) error

	// ListByOwner lists all characters for a specific owner
	ListByOwner(ownerID string) ([]*charDomain.Character, error)

	// GetByID retrieves a character by ID
	GetByID(characterID string) (*charDomain.Character, error)

	// FixCharacterAttributes fixes characters that have AbilityAssignments but no Attributes
	FixCharacterAttributes(ctx context.Context, characterID string) (*charDomain.Character, error)

	// FinalizeCharacterWithName sets the name and finalizes a draft character in one operation
	FinalizeCharacterWithName(ctx context.Context, characterID, name, raceKey, classKey string) (*charDomain.Character, error)

	// Character Creation Session methods
	// StartCharacterCreation starts a new character creation session
	StartCharacterCreation(ctx context.Context, userID, guildID string) (*charDomain.CharacterCreationSession, error)

	// GetCharacterCreationSession retrieves an active session
	GetCharacterCreationSession(ctx context.Context, sessionID string) (*charDomain.CharacterCreationSession, error)

	// UpdateCharacterCreationSession updates the session step
	UpdateCharacterCreationSession(ctx context.Context, sessionID, step string) error

	// GetCharacterFromSession gets the character associated with a session
	GetCharacterFromSession(ctx context.Context, sessionID string) (*charDomain.Character, error)

	// StartFreshCharacterCreation gets or creates a draft and clears ability rolls
	StartFreshCharacterCreation(ctx context.Context, userID, realmID string) (*charDomain.Character, error)

	// Spell-related methods
	// GetSpell retrieves spell information
	GetSpell(ctx context.Context, spellKey string) (*rulebook.Spell, error)

	// ListSpellsByClass retrieves all spells available to a class
	ListSpellsByClass(ctx context.Context, classKey string) ([]*rulebook.SpellReference, error)

	// ListSpellsByClassAndLevel retrieves spells available to a class at a specific level
	ListSpellsByClassAndLevel(ctx context.Context, classKey string, level int) ([]*rulebook.SpellReference, error)

	// GetChoiceResolver returns the service's choice resolver for use by other services
	GetChoiceResolver() ChoiceResolver
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
	Character *charDomain.Character
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
	AbilityScores      map[string]int           // Legacy: direct ability -> score mapping
	AbilityRolls       []charDomain.AbilityRoll // New: rolls with IDs
	AbilityAssignments map[string]string        // New: ability -> roll ID mapping
	Proficiencies      []string
	Equipment          []string
	Name               *string
	Spells             *charDomain.SpellList // Spell selection for spellcasters
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
	BundleItems []string // Equipment keys that come with this choice (e.g., shield with weapon+shield)
}

// service implements the Service interface
type service struct {
	dndClient       dnd5e.Client
	choiceResolver  ChoiceResolver
	repository      Repository
	draftRepository draftRepo.Repository
	acCalculator    charDomain.ACCalculator
	// Temporary in-memory session store (should be Redis in production)
	sessions map[string]*charDomain.CharacterCreationSession
	// Later we'll add:
	// validator  Validator
}

// ServiceConfig holds configuration for the service
type ServiceConfig struct {
	DNDClient       dnd5e.Client
	ChoiceResolver  ChoiceResolver          // Optional, will create default if nil
	Repository      Repository              // Required
	DraftRepository draftRepo.Repository    // Required
	ACCalculator    charDomain.ACCalculator // Optional, will use default if nil
}

// NewService creates a new character service
func NewService(cfg *ServiceConfig) Service {
	if cfg.Repository == nil {
		panic("repository is required")
	}
	if cfg.DraftRepository == nil {
		panic("draft repository is required")
	}

	svc := &service{
		dndClient:       cfg.DNDClient,
		repository:      cfg.Repository,
		draftRepository: cfg.DraftRepository,
		sessions:        make(map[string]*charDomain.CharacterCreationSession),
	}

	// Use provided choice resolver or create default
	if cfg.ChoiceResolver != nil {
		svc.choiceResolver = cfg.ChoiceResolver
	} else {
		svc.choiceResolver = NewChoiceResolver(cfg.DNDClient)
	}

	// Use provided AC calculator or create default
	if cfg.ACCalculator != nil {
		svc.acCalculator = cfg.ACCalculator
	} else {
		// For now, we'll use a wrapper around the existing function
		// This will be replaced with the proper D&D 5e calculator
		svc.acCalculator = &defaultACCalculator{}
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
	character := &charDomain.Character{
		ID:      generateID(),
		OwnerID: input.UserID,
		RealmID: input.RealmID,
		Name:    input.Name,
		Race:    race,
		Class:   class,
		Status:  shared.CharacterStatusDraft,
		HitDie:  class.HitDie,
		Speed:   race.Speed,
		Level:   1,
	}

	// Set ability scores
	character.Attributes = make(map[shared.Attribute]*charDomain.AbilityScore)
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
			equipmentValue, err := s.dndClient.GetEquipment(se.Equipment.Key)
			if err == nil && equipmentValue != nil {
				for i := 0; i < se.Quantity; i++ {
					character.AddInventory(equipmentValue)
				}
			}
		}
	}

	// Add selected equipment
	for _, equipKey := range input.Equipment {
		equipmentValue, err := s.dndClient.GetEquipment(equipKey)
		if err == nil && equipmentValue != nil {
			character.AddInventory(equipmentValue)
		}
	}

	// Add racial features
	racialFeatures := features2.GetRacialFeatures(race.Key)
	if character.Features == nil {
		character.Features = []*rulebook.CharacterFeature{}
	}
	for _, feat := range racialFeatures {
		featCopy := feat
		character.Features = append(character.Features, &featCopy)
	}

	// Add class features
	classFeatures := features2.GetClassFeatures(class.Key, character.Level)
	for _, feat := range classFeatures {
		featCopy := feat
		character.Features = append(character.Features, &featCopy)
	}

	// Apply passive effects from all features
	if err := features2.DefaultRegistry.ApplyAllPassiveEffects(character); err != nil {
		log.Printf("Error applying passive effects for character %s: %v", character.ID, err)
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
func (s *service) GetCharacter(ctx context.Context, characterID string) (*charDomain.Character, error) {
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
func (s *service) ListCharacters(ctx context.Context, userID string) ([]*charDomain.Character, error) {
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
func (s *service) GetRace(ctx context.Context, raceKey string) (*rulebook.Race, error) {
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
func (s *service) GetClass(ctx context.Context, classKey string) (*rulebook.Class, error) {
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
func (s *service) GetRaces(ctx context.Context) ([]*rulebook.Race, error) {
	races, err := s.dndClient.ListRaces()
	if err != nil {
		return nil, dnderr.Wrap(err, "failed to list races")
	}

	return races, nil
}

// GetClasses retrieves all available classes
func (s *service) GetClasses(ctx context.Context) ([]*rulebook.Class, error) {
	classes, err := s.dndClient.ListClasses()
	if err != nil {
		return nil, dnderr.Wrap(err, "failed to list classes")
	}

	return classes, nil
}

// GetOrCreateDraftCharacter gets an existing draft or creates a new one
func (s *service) GetOrCreateDraftCharacter(ctx context.Context, userID, realmID string) (*charDomain.Character, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, dnderr.InvalidArgument("user ID is required")
	}
	if strings.TrimSpace(realmID) == "" {
		return nil, dnderr.InvalidArgument("realm ID is required")
	}

	// Look for existing draft
	draft, err := s.draftRepository.GetByOwnerAndRealm(ctx, userID, realmID)
	if err == nil && draft != nil {
		// Found existing draft, return the character
		return draft.Character, nil
	}
	// If error is not "not found", return it
	if err != nil && !dnderr.IsNotFound(err) {
		return nil, dnderr.Wrap(err, "failed to get existing draft").
			WithMeta("user_id", userID).
			WithMeta("realm_id", realmID)
	}

	// No draft found, create a new one
	character := &charDomain.Character{
		ID:      generateID(),
		OwnerID: userID,
		RealmID: realmID,
		Name:    "",
		Status:  shared.CharacterStatusDraft,
		Level:   1,
		// Initialize empty maps
		Attributes:    make(map[shared.Attribute]*charDomain.AbilityScore),
		Proficiencies: make(map[rulebook.ProficiencyType][]*rulebook.Proficiency),
		Inventory:     make(map[equipment.EquipmentType][]equipment.Equipment),
		EquippedSlots: make(map[shared.Slot]equipment.Equipment),
	}

	// Create draft wrapper
	newDraft := &charDomain.CharacterDraft{
		ID:        fmt.Sprintf("draft_%d", time.Now().UnixNano()),
		OwnerID:   userID,
		Character: character,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		FlowState: &charDomain.FlowState{
			CurrentStepID:  "race",
			AllSteps:       []string{"race", "class", "abilities", "proficiencies", "equipment", "features", "name"},
			CompletedSteps: []string{},
			LastUpdated:    time.Now(),
		},
	}

	// Save draft
	if err := s.draftRepository.Create(ctx, newDraft); err != nil {
		return nil, dnderr.Wrap(err, "failed to create draft").
			WithMeta("draft_id", newDraft.ID).
			WithMeta("owner_id", userID)
	}

	// Also save character to character repository for backward compatibility
	if err := s.repository.Create(ctx, character); err != nil {
		// Try to clean up the draft
		if deleteErr := s.draftRepository.Delete(ctx, newDraft.ID); deleteErr != nil {
			log.Printf("Failed to clean up draft after character creation failure: %v", deleteErr)
		}
		return nil, dnderr.Wrap(err, "failed to create character").
			WithMeta("character_id", character.ID).
			WithMeta("owner_id", userID)
	}

	return character, nil
}

// UpdateDraftCharacter updates a draft character
func (s *service) UpdateDraftCharacter(ctx context.Context, characterID string, updates *UpdateDraftInput) (*charDomain.Character, error) {
	if strings.TrimSpace(characterID) == "" {
		return nil, dnderr.InvalidArgument("character ID is required")
	}
	if updates == nil {
		return nil, dnderr.InvalidArgument("updates cannot be nil")
	}

	// Try to get draft first
	draft, err := s.draftRepository.GetByCharacterID(ctx, characterID)
	if err != nil && !dnderr.IsNotFound(err) {
		log.Printf("Failed to get draft for character %s: %v", characterID, err)
	}

	// Get the character
	char, err := s.repository.Get(ctx, characterID)
	if err != nil {
		return nil, dnderr.Wrapf(err, "failed to get character '%s'", characterID).
			WithMeta("character_id", characterID)
	}

	// Log character state BEFORE updates

	// Verify it's a draft
	if char.Status != shared.CharacterStatusDraft {
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
		char.Speed = features2.CalculateSpeed(char)

		// Apply racial features
		racialFeatures := features2.GetRacialFeatures(race.Key)
		if char.Features == nil {
			char.Features = []*rulebook.CharacterFeature{}
		}
		// Remove existing racial features
		newFeatures := []*rulebook.CharacterFeature{}
		for _, f := range char.Features {
			if f.Type != rulebook.FeatureTypeRacial {
				newFeatures = append(newFeatures, f)
			}
		}
		// Add new racial features
		for _, feat := range racialFeatures {
			featCopy := feat
			newFeatures = append(newFeatures, &featCopy)
		}
		char.Features = newFeatures

		// Apply passive effects from new racial features
		if err := features2.DefaultRegistry.ApplyAllPassiveEffects(char); err != nil {
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
		classFeatures := features2.GetClassFeatures(class.Key, char.Level)
		if char.Features == nil {
			char.Features = []*rulebook.CharacterFeature{}
		}
		// Create map of existing class features to preserve metadata
		existingClassFeatures := make(map[string]*rulebook.CharacterFeature)
		newFeatures := []*rulebook.CharacterFeature{}

		for _, f := range char.Features {
			if f.Type == rulebook.FeatureTypeClass {
				// Store existing class features by key to preserve metadata
				existingClassFeatures[f.Key] = f
			} else {
				// Keep non-class features as-is
				newFeatures = append(newFeatures, f)
			}
		}

		// Add class features, preserving existing ones with metadata
		for _, feat := range classFeatures {
			if existing, exists := existingClassFeatures[feat.Key]; exists {
				// Preserve existing feature with its metadata
				newFeatures = append(newFeatures, existing)
			} else {
				// Add new feature from template
				featCopy := feat
				newFeatures = append(newFeatures, &featCopy)
			}
		}
		char.Features = newFeatures

		// Recalculate AC with new class features
		char.AC = s.acCalculator.Calculate(char)
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
			char.Attributes = make(map[shared.Attribute]*charDomain.AbilityScore)

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
			char.AC = s.acCalculator.Calculate(char)
		}
	}

	// Legacy: Update ability scores if provided directly
	if len(updates.AbilityScores) > 0 && updates.AbilityAssignments == nil {
		// Clear existing scores
		char.Attributes = make(map[shared.Attribute]*charDomain.AbilityScore)

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
		char.AC = s.acCalculator.Calculate(char)
	}

	// Update name if provided
	if updates.Name != nil {
		char.Name = *updates.Name
	}

	// Update proficiencies if provided
	if len(updates.Proficiencies) > 0 {

		// Clear existing chosen proficiencies (keep starting ones from race/class)
		if char.Proficiencies == nil {
			char.Proficiencies = make(map[rulebook.ProficiencyType][]*rulebook.Proficiency)
		}

		// Separate proficiencies by type
		skillProficiencies := []*rulebook.Proficiency{}

		// Get proficiency objects and filter by type
		for _, profKey := range updates.Proficiencies {
			proficiency, err := s.dndClient.GetProficiency(profKey)
			if err == nil && proficiency != nil {
				// Only handle skill proficiencies as replacements
				// Other types (armor, weapon, etc) should accumulate
				if proficiency.Type == rulebook.ProficiencyTypeSkill {
					skillProficiencies = append(skillProficiencies, proficiency)
				} else {
					// For non-skill proficiencies, add normally (with duplicate check)
					char.AddProficiency(proficiency)
				}
			}
		}

		// Replace all skill proficiencies with the new set
		if len(skillProficiencies) > 0 {
			char.SetProficiencies(rulebook.ProficiencyTypeSkill, skillProficiencies)
		}

	}

	// Update equipment if provided
	if len(updates.Equipment) > 0 {

		// Initialize inventory if needed
		if char.Inventory == nil {
			char.Inventory = make(map[equipment.EquipmentType][]equipment.Equipment)
		}

		// Add selected equipment
		for _, equipKey := range updates.Equipment {
			equipmentValue, err := s.dndClient.GetEquipment(equipKey)
			if err == nil && equipmentValue != nil {
				char.AddInventory(equipmentValue)
			} else if err != nil {
				log.Printf("Failed to get equipment '%s': %v", equipKey, err)
			}
		}

		// Recalculate AC in case equipment affects it
		char.AC = s.acCalculator.Calculate(char)

	}

	// Update spells if provided
	if updates.Spells != nil {
		char.Spells = updates.Spells
	}

	// Save changes
	if err := s.repository.Update(ctx, char); err != nil {
		return nil, dnderr.Wrap(err, "failed to update character").
			WithMeta("character_id", characterID)
	}

	// Update draft flow state if we have a draft
	if draft != nil && draft.FlowState != nil {
		// Track what steps were completed
		if updates.RaceKey != nil && !contains(draft.FlowState.CompletedSteps, "race") {
			draft.FlowState.CompletedSteps = append(draft.FlowState.CompletedSteps, "race")
			draft.FlowState.CurrentStepID = "class"
		}
		if updates.ClassKey != nil && !contains(draft.FlowState.CompletedSteps, "class") {
			draft.FlowState.CompletedSteps = append(draft.FlowState.CompletedSteps, "class")
			draft.FlowState.CurrentStepID = "abilities"
		}
		if len(updates.AbilityAssignments) > 0 && !contains(draft.FlowState.CompletedSteps, "abilities") {
			draft.FlowState.CompletedSteps = append(draft.FlowState.CompletedSteps, "abilities")
			draft.FlowState.CurrentStepID = "proficiencies"
		}
		if len(updates.Proficiencies) > 0 && !contains(draft.FlowState.CompletedSteps, "proficiencies") {
			draft.FlowState.CompletedSteps = append(draft.FlowState.CompletedSteps, "proficiencies")
			draft.FlowState.CurrentStepID = "equipment"
		}
		if len(updates.Equipment) > 0 && !contains(draft.FlowState.CompletedSteps, "equipment") {
			draft.FlowState.CompletedSteps = append(draft.FlowState.CompletedSteps, "equipment")
			draft.FlowState.CurrentStepID = "features"
		}
		// Name is typically the last step
		if updates.Name != nil && *updates.Name != "" && !contains(draft.FlowState.CompletedSteps, "name") {
			draft.FlowState.CompletedSteps = append(draft.FlowState.CompletedSteps, "name")
			draft.FlowState.CurrentStepID = "complete"
		}

		draft.FlowState.LastUpdated = time.Now()
		draft.Character = char
		draft.UpdatedAt = time.Now()

		// Save draft updates
		if err := s.draftRepository.Update(ctx, draft); err != nil {
			// Log but don't fail - character is already saved
			log.Printf("Failed to update draft flow state: %v", err)
		}
	}

	return char, nil
}

// FinalizeDraftCharacter marks a draft as active
func (s *service) FinalizeDraftCharacter(ctx context.Context, characterID string) (*charDomain.Character, error) {
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
	if char.Status != shared.CharacterStatusDraft {
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
		char.Attributes = make(map[shared.Attribute]*charDomain.AbilityScore)

		// Convert assignments to attributes
		for abilityStr, rollID := range char.AbilityAssignments {
			rollValue, rollValueOk := rollValues[rollID]
			if rollValueOk {
				// Parse ability string to Attribute type
				var attr shared.Attribute
				switch abilityStr {
				case "STR":
					attr = shared.AttributeStrength
				case "DEX":
					attr = shared.AttributeDexterity
				case "CON":
					attr = shared.AttributeConstitution
				case "INT":
					attr = shared.AttributeIntelligence
				case "WIS":
					attr = shared.AttributeWisdom
				case "CHA":
					attr = shared.AttributeCharisma
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
				char.Attributes[attr] = &charDomain.AbilityScore{
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
			if con, ok := char.Attributes[shared.AttributeConstitution]; ok && con != nil {
				conMod = con.Bonus
			}
		}
		char.MaxHitPoints = char.Class.HitDie + conMod
		char.CurrentHitPoints = char.MaxHitPoints
		char.HitDie = char.Class.HitDie
	}

	// Ensure all required features are present, preserving existing metadata
	if char.Features == nil {
		char.Features = []*rulebook.CharacterFeature{}
	}

	// Create a map of existing features by key for quick lookup
	existingFeatures := make(map[string]*rulebook.CharacterFeature)
	for _, feat := range char.Features {
		existingFeatures[feat.Key] = feat
	}

	// Add missing racial features (preserve existing ones with metadata)
	if char.Race != nil {
		racialFeatures := features2.GetRacialFeatures(char.Race.Key)
		for _, templateFeat := range racialFeatures {
			if _, exists := existingFeatures[templateFeat.Key]; !exists {
				// Feature doesn't exist, add it from template
				featCopy := templateFeat // Make a copy to avoid reference issues
				char.Features = append(char.Features, &featCopy)
				existingFeatures[templateFeat.Key] = &featCopy
			}
			// If feature already exists, keep it as-is with its metadata
		}
	}

	// Add missing class features (preserve existing ones with metadata)
	if char.Class != nil {
		classFeatures := features2.GetClassFeatures(char.Class.Key, char.Level)
		for _, templateFeat := range classFeatures {
			if _, exists := existingFeatures[templateFeat.Key]; !exists {
				// Feature doesn't exist, add it from template
				featCopy := templateFeat // Make a copy to avoid reference issues
				char.Features = append(char.Features, &featCopy)
				existingFeatures[templateFeat.Key] = &featCopy
			}
			// If feature already exists, keep it as-is with its metadata
		}
	}

	// Apply passive effects from all features
	if applyErr := features2.DefaultRegistry.ApplyAllPassiveEffects(char); applyErr != nil {
		// Log error but don't fail finalization
		log.Printf("Error applying passive effects for character %s (%s %s): %v",
			char.ID, char.Race.Name, char.Class.Name, applyErr)
	}

	// Calculate AC using the features package
	char.AC = s.acCalculator.Calculate(char)

	// Add starting proficiencies from class if not already present
	if s.dndClient != nil && char.Class != nil && char.Class.Proficiencies != nil {
		// Check if we already have class proficiencies (in case they were added earlier)
		hasClassProficiencies := false
		if weaponProfs, exists := char.Proficiencies[rulebook.ProficiencyTypeWeapon]; exists && len(weaponProfs) > 0 {
			hasClassProficiencies = true
		}

		if !hasClassProficiencies {
			for _, prof := range char.Class.Proficiencies {
				if prof != nil {
					proficiency, profErr := s.dndClient.GetProficiency(prof.Key)
					if profErr == nil && proficiency != nil {
						char.AddProficiency(proficiency)
					}
				}
			}
		}
	}

	// Add starting proficiencies from race if not already present
	if s.dndClient != nil && char.Race != nil && char.Race.StartingProficiencies != nil {
		for _, prof := range char.Race.StartingProficiencies {
			if prof != nil {
				proficiency, profErr := s.dndClient.GetProficiency(prof.Key)
				if profErr == nil && proficiency != nil {
					char.AddProficiency(proficiency)
				}
			}
		}
	}

	// Add starting equipment if class is set and DND client is available
	if s.dndClient != nil && char.Class != nil && char.Class.StartingEquipment != nil {
		for _, se := range char.Class.StartingEquipment {
			if se != nil && se.Equipment != nil {
				equipmentValue, eqErr := s.dndClient.GetEquipment(se.Equipment.Key)
				if eqErr != nil {
					// Log the error but don't fail the finalization
					log.Printf("Failed to get starting equipment %s: %v", se.Equipment.Key, eqErr)
					continue
				}
				// If no error, equipment is valid
				for i := 0; i < se.Quantity; i++ {
					char.AddInventory(equipmentValue)
				}
			}
		}
	}

	// Initialize resources before finalizing
	char.InitializeResources()

	// Update status to active
	char.Status = shared.CharacterStatusActive

	// Save changes
	if updateErr := s.repository.Update(ctx, char); updateErr != nil {
		return nil, dnderr.Wrap(updateErr, "failed to finalize character").
			WithMeta("character_id", characterID)
	}

	// Delete draft now that character is finalized
	draft, err := s.draftRepository.GetByCharacterID(ctx, characterID)
	if err == nil && draft != nil {
		// Delete the draft
		if err := s.draftRepository.Delete(ctx, draft.ID); err != nil {
			// Log but don't fail - character is already finalized
			log.Printf("Failed to delete draft %s after finalization: %v", draft.ID, err)
		}
	}

	return char, nil
}

// GetEquipmentByCategory retrieves equipment by category
func (s *service) GetEquipmentByCategory(ctx context.Context, category string) ([]equipment.Equipment, error) {
	if strings.TrimSpace(category) == "" {
		return nil, dnderr.InvalidArgument("category is required")
	}

	equipmentSlice, err := s.dndClient.GetEquipmentByCategory(category)
	if err != nil {
		return nil, dnderr.Wrapf(err, "failed to get equipment for category '%s'", category).
			WithMeta("category", category)
	}

	return equipmentSlice, nil
}

// generateID generates a unique ID for a character
func generateID() string {
	// TODO: Implement proper ID generation (e.g., UUID or snowflake)
	// For now, use timestamp-based ID
	return fmt.Sprintf("char_%d", time.Now().UnixNano())
}

// UpdateStatus updates a character's status
func (s *service) UpdateStatus(characterID string, status shared.CharacterStatus) error {
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
func (s *service) UpdateEquipment(character *charDomain.Character) error {
	if character == nil {
		return dnderr.InvalidArgument("character is required")
	}

	if strings.TrimSpace(character.ID) == "" {
		return dnderr.InvalidArgument("character ID is required")
	}

	ctx := context.Background()

	// Recalculate AC with the features calculator
	character.AC = s.acCalculator.Calculate(character)

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

// GetPendingFeatureChoices returns feature choices that need to be made for a character
func (s *service) GetPendingFeatureChoices(ctx context.Context, characterID string) ([]*rulebook.FeatureChoice, error) {
	if strings.TrimSpace(characterID) == "" {
		return nil, dnderr.InvalidArgument("character ID is required")
	}

	// Get the character
	char, err := s.repository.Get(ctx, characterID)
	if err != nil {
		return nil, dnderr.Wrapf(err, "failed to get character '%s'", characterID).
			WithMeta("character_id", characterID)
	}

	if char.Class == nil {
		return nil, nil // No class, no choices
	}

	// Get all feature choices for this class at current level
	allChoices := rulebook.GetClassFeatureChoices(char.Class.Key, char.Level)

	var pendingChoices []*rulebook.FeatureChoice

	// Check which choices have already been made
	for _, choice := range allChoices {
		// Find the corresponding feature in the character
		var featureFound *rulebook.CharacterFeature
		for _, feature := range char.Features {
			if feature.Key == choice.FeatureKey {
				featureFound = feature
				break
			}
		}

		// Check if choice has been made
		choiceMade := false
		if featureFound != nil && featureFound.Metadata != nil {
			switch choice.Type {
			case rulebook.FeatureChoiceTypeFightingStyle:
				_, choiceMade = featureFound.Metadata["style"]
			case rulebook.FeatureChoiceTypeDivineDomain:
				_, choiceMade = featureFound.Metadata["domain"]
			case rulebook.FeatureChoiceTypeFavoredEnemy:
				_, choiceMade = featureFound.Metadata["enemy_type"]
			case rulebook.FeatureChoiceTypeNaturalExplorer:
				_, choiceMade = featureFound.Metadata["terrain_type"]
			}
		}

		if !choiceMade {
			pendingChoices = append(pendingChoices, choice)
		}
	}

	return pendingChoices, nil
}

// ListByOwner lists all characters for a specific owner
func (s *service) ListByOwner(ownerID string) ([]*charDomain.Character, error) {
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
func (s *service) GetByID(characterID string) (*charDomain.Character, error) {
	if strings.TrimSpace(characterID) == "" {
		return nil, dnderr.InvalidArgument("character ID is required")
	}

	ctx := context.Background()

	char, err := s.repository.Get(ctx, characterID)
	if err != nil {
		return nil, dnderr.Wrapf(err, "failed to get character '%s'", characterID).
			WithMeta("character_id", characterID)
	}

	// Ensure Resources are initialized
	if char.Resources == nil && char.Class != nil {
		log.Printf("Character %s (%s) has no Resources, initializing...", char.ID, char.Name)
		char.InitializeResources()
	}

	return char, nil
}

// stringToAttribute converts a string to an Attribute
func stringToAttribute(s string) shared.Attribute {
	switch strings.ToUpper(s) {
	case "STR":
		return shared.AttributeStrength
	case "DEX":
		return shared.AttributeDexterity
	case "CON":
		return shared.AttributeConstitution
	case "INT":
		return shared.AttributeIntelligence
	case "WIS":
		return shared.AttributeWisdom
	case "CHA":
		return shared.AttributeCharisma
	default:
		return shared.AttributeNone
	}
}

// StartFreshCharacterCreation gets or creates a draft and clears ability rolls
func (s *service) StartFreshCharacterCreation(ctx context.Context, userID, realmID string) (*charDomain.Character, error) {
	draft, err := s.GetOrCreateDraftCharacter(ctx, userID, realmID)
	if err != nil {
		return nil, err
	}

	// Clear ability rolls and assignments if they exist
	if len(draft.AbilityRolls) > 0 || len(draft.AbilityAssignments) > 0 {
		_, err = s.UpdateDraftCharacter(ctx, draft.ID, &UpdateDraftInput{
			AbilityRolls:       []charDomain.AbilityRoll{},
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

// contains is a helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// GetSpell retrieves spell information
func (s *service) GetSpell(ctx context.Context, spellKey string) (*rulebook.Spell, error) {
	if spellKey == "" {
		return nil, dnderr.InvalidArgument("spell key is required")
	}

	spell, err := s.dndClient.GetSpell(spellKey)
	if err != nil {
		return nil, dnderr.Wrapf(err, "failed to get spell '%s'", spellKey).
			WithMeta("spell_key", spellKey)
	}

	return spell, nil
}

// ListSpellsByClass retrieves all spells available to a class
func (s *service) ListSpellsByClass(ctx context.Context, classKey string) ([]*rulebook.SpellReference, error) {
	if classKey == "" {
		return nil, dnderr.InvalidArgument("class key is required")
	}

	spells, err := s.dndClient.ListSpellsByClass(classKey)
	if err != nil {
		return nil, dnderr.Wrapf(err, "failed to list spells for class '%s'", classKey).
			WithMeta("class_key", classKey)
	}

	return spells, nil
}

// ListSpellsByClassAndLevel retrieves spells available to a class at a specific level
func (s *service) ListSpellsByClassAndLevel(ctx context.Context, classKey string, level int) ([]*rulebook.SpellReference, error) {
	if classKey == "" {
		return nil, dnderr.InvalidArgument("class key is required")
	}
	if level < 0 || level > 9 {
		return nil, dnderr.InvalidArgument("spell level must be between 0 and 9")
	}

	spells, err := s.dndClient.ListSpellsByClassAndLevel(classKey, level)
	if err != nil {
		return nil, dnderr.Wrapf(err, "failed to list level %d spells for class '%s'", level, classKey).
			WithMeta("class_key", classKey).
			WithMeta("spell_level", level)
	}

	return spells, nil
}

// GetChoiceResolver returns the service's choice resolver
func (s *service) GetChoiceResolver() ChoiceResolver {
	return s.choiceResolver
}
