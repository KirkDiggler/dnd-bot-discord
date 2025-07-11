package character

import (
	"context"
	"fmt"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/equipment"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
)

// CreationFlowServiceImpl implements the CreationFlowService interface
type CreationFlowServiceImpl struct {
	characterService Service
	flowBuilder      character.FlowBuilder
}

// NewCreationFlowService creates a new creation flow service
func NewCreationFlowService(characterService Service, flowBuilder character.FlowBuilder) character.CreationFlowService {
	return &CreationFlowServiceImpl{
		characterService: characterService,
		flowBuilder:      flowBuilder,
	}
}

// GetNextStep returns the next step in character creation
func (s *CreationFlowServiceImpl) GetNextStep(ctx context.Context, characterID string) (*character.CreationStep, error) {
	char, err := s.characterService.GetByID(characterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get character: %w", err)
	}

	// Build the complete flow for this character
	flow, err := s.flowBuilder.BuildFlow(ctx, char)
	if err != nil {
		return nil, fmt.Errorf("failed to build flow: %w", err)
	}

	// Find the first incomplete step
	for _, step := range flow.Steps {
		if !s.isStepComplete(char, &step) {
			return &step, nil
		}
	}

	// All steps complete
	return &character.CreationStep{
		Type:        character.StepTypeComplete,
		Title:       "Character Creation Complete",
		Description: "Your character is ready to adventure!",
		Required:    false,
	}, nil
}

// ProcessStepResult processes a completed step and returns the next step
func (s *CreationFlowServiceImpl) ProcessStepResult(ctx context.Context, characterID string, result *character.CreationStepResult) (*character.CreationStep, error) {
	// Apply the step result to the character
	if err := s.applyStepResult(ctx, characterID, result); err != nil {
		return nil, fmt.Errorf("failed to apply step result: %w", err)
	}

	// Return the next step
	return s.GetNextStep(ctx, characterID)
}

// GetCurrentStep returns the current step for a character
func (s *CreationFlowServiceImpl) GetCurrentStep(ctx context.Context, characterID string) (*character.CreationStep, error) {
	return s.GetNextStep(ctx, characterID) // Same as next step for now
}

// IsCreationComplete returns true if character creation is complete
func (s *CreationFlowServiceImpl) IsCreationComplete(ctx context.Context, characterID string) (bool, error) {
	step, err := s.GetNextStep(ctx, characterID)
	if err != nil {
		return false, err
	}
	return step.Type == character.StepTypeComplete, nil
}

// GetProgressSteps returns all steps with completion status
func (s *CreationFlowServiceImpl) GetProgressSteps(ctx context.Context, characterID string) ([]character.ProgressStepInfo, error) {
	char, err := s.characterService.GetByID(characterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get character: %w", err)
	}

	flow, err := s.flowBuilder.BuildFlow(ctx, char)
	if err != nil {
		return nil, fmt.Errorf("failed to build flow: %w", err)
	}

	currentStep, err := s.GetNextStep(ctx, characterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current step: %w", err)
	}

	var progressSteps []character.ProgressStepInfo
	for _, step := range flow.Steps {
		progressSteps = append(progressSteps, character.ProgressStepInfo{
			Step:      step,
			Completed: s.isStepComplete(char, &step),
			Current:   step.Type == currentStep.Type,
		})
	}

	return progressSteps, nil
}

// PreviewStepResult creates a preview of what the character would look like with the given selection
func (s *CreationFlowServiceImpl) PreviewStepResult(ctx context.Context, characterID string, result *character.CreationStepResult) (*character.Character, error) {
	char, err := s.characterService.GetByID(characterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get character: %w", err)
	}

	// Create a copy of the character for preview
	previewChar := &character.Character{
		ID:               char.ID,
		Name:             char.Name,
		OwnerID:          char.OwnerID,
		RealmID:          char.RealmID,
		Level:            char.Level,
		Status:           char.Status,
		Race:             char.Race,
		Class:            char.Class,
		Attributes:       make(map[shared.Attribute]*character.AbilityScore),
		Proficiencies:    make(map[rulebook.ProficiencyType][]*rulebook.Proficiency),
		Features:         make([]*rulebook.CharacterFeature, 0),
		Inventory:        make(map[equipment.EquipmentType][]equipment.Equipment),
		EquippedSlots:    make(map[shared.Slot]equipment.Equipment),
		CurrentHitPoints: char.CurrentHitPoints,
		MaxHitPoints:     char.MaxHitPoints,
		AC:               char.AC,
		Speed:            char.Speed,
	}

	// Copy existing data
	for k, v := range char.Attributes {
		previewChar.Attributes[k] = v
	}
	for k, v := range char.Proficiencies {
		previewChar.Proficiencies[k] = v
	}
	previewChar.Features = append(previewChar.Features, char.Features...)

	// Apply the preview change based on step type
	switch result.StepType {
	case character.StepTypeRaceSelection:
		if len(result.Selections) > 0 {
			// Get the race from the character service
			race, err := s.characterService.GetRace(ctx, result.Selections[0])
			if err == nil && race != nil {
				previewChar.Race = race
			}
		}
	case character.StepTypeClassSelection:
		if len(result.Selections) > 0 {
			// Get the class from the character service
			class, err := s.characterService.GetClass(ctx, result.Selections[0])
			if err == nil && class != nil {
				previewChar.Class = class
			}
		}
		// Add other preview cases as needed
	}

	return previewChar, nil
}

// isStepComplete checks if a step is complete for the given character
func (s *CreationFlowServiceImpl) isStepComplete(char *character.Character, step *character.CreationStep) bool {
	switch step.Type {
	case character.StepTypeRaceSelection:
		return char.Race != nil
	case character.StepTypeClassSelection:
		return char.Class != nil
	case character.StepTypeAbilityScores:
		// Check if abilities have been rolled (not if they're assigned to attributes)
		return len(char.AbilityRolls) >= 6
	case character.StepTypeAbilityAssignment:
		// Check if all ability scores are assigned (not default values)
		return s.hasAssignedAbilities(char)
	case character.StepTypeSkillSelection:
		// Check if character has completed domain-specific skill selection
		return s.hasCompletedDomainSkills(char)
	case character.StepTypeLanguageSelection:
		// Check if character has completed domain-specific language selection
		return s.hasCompletedDomainLanguages(char)
	case character.StepTypeFightingStyleSelection:
		return s.hasSelectedFightingStyle(char)
	case character.StepTypeDivineDomainSelection:
		return s.hasSelectedDivineDomain(char)
	case character.StepTypeFavoredEnemySelection:
		return s.hasSelectedFavoredEnemy(char)
	case character.StepTypeNaturalExplorerSelection:
		return s.hasSelectedNaturalExplorer(char)
	case character.StepTypeProficiencySelection:
		// Check if user has made proficiency choices beyond automatic ones
		// This is a simplified check - ideally we'd track which are from choices
		return s.hasUserSelectedProficiencies(char)
	case character.StepTypeEquipmentSelection:
		// Check if equipment has been selected (not just starting equipment)
		return s.hasUserSelectedEquipment(char)
	case character.StepTypeCharacterDetails:
		// Character needs a name AND to be finalized
		return char.Name != "" && char.Status != shared.CharacterStatusDraft

	// Spellcaster steps
	case character.StepTypeCantripsSelection:
		return s.hasSelectedCantrips(char)
	case character.StepTypeSpellSelection, character.StepTypeSpellbookSelection:
		return s.hasSelectedSpells(char)

	// Subclass steps
	case character.StepTypeSubclassSelection, character.StepTypePatronSelection, character.StepTypeSorcerousOriginSelection:
		return s.hasSelectedSubclass(char)

	default:
		return false
	}
}

// Helper methods for checking step completion
func (s *CreationFlowServiceImpl) hasAssignedAbilities(char *character.Character) bool {
	// Check if all 6 ability scores have been assigned
	if len(char.AbilityAssignments) != 6 {
		return false
	}

	// Verify each ability has an assignment
	requiredAbilities := []string{"STR", "DEX", "CON", "INT", "WIS", "CHA"}
	for _, ability := range requiredAbilities {
		if _, exists := char.AbilityAssignments[ability]; !exists {
			return false
		}
	}

	return true
}

func (s *CreationFlowServiceImpl) hasCompletedDomainSkills(char *character.Character) bool {
	// Check if Knowledge domain cleric has selected bonus skills
	if char.Class == nil || char.Class.Key != "cleric" {
		return true // Not applicable
	}

	// Check if divine domain is Knowledge
	for _, feature := range char.Features {
		if feature.Key == "divine_domain" && feature.Metadata != nil {
			if domain, ok := feature.Metadata["domain"].(string); ok && domain == "knowledge" {
				// Check if bonus skills have been selected
				// Check both []string and []interface{} since JSON unmarshaling can produce either
				if skills, ok := feature.Metadata["bonus_skills"].([]string); ok && len(skills) >= 2 {
					return true
				}
				if skills, ok := feature.Metadata["bonus_skills"].([]any); ok && len(skills) >= 2 {
					return true
				}
				return false
			}
		}
	}
	return true // No Knowledge domain, skill selection not needed
}

func (s *CreationFlowServiceImpl) hasCompletedDomainLanguages(char *character.Character) bool {
	// Similar to skills but for languages
	if char.Class == nil || char.Class.Key != "cleric" {
		return true
	}

	for _, feature := range char.Features {
		if feature.Key == "divine_domain" && feature.Metadata != nil {
			if domain, ok := feature.Metadata["domain"].(string); ok && domain == "knowledge" {
				// Check both []string and []interface{} since JSON unmarshaling can produce either
				if languages, ok := feature.Metadata["bonus_languages"].([]string); ok && len(languages) >= 2 {
					return true
				}
				if languages, ok := feature.Metadata["bonus_languages"].([]any); ok && len(languages) >= 2 {
					return true
				}
				return false
			}
		}
	}
	return true
}

func (s *CreationFlowServiceImpl) hasSelectedFightingStyle(char *character.Character) bool {
	for _, feature := range char.Features {
		if feature.Key == "fighting_style" && feature.Metadata != nil {
			if _, ok := feature.Metadata["style"].(string); ok {
				return true
			}
		}
	}
	return false
}

func (s *CreationFlowServiceImpl) hasSelectedDivineDomain(char *character.Character) bool {
	for _, feature := range char.Features {
		if feature.Key == "divine_domain" && feature.Metadata != nil {
			if _, ok := feature.Metadata["domain"].(string); ok {
				return true
			}
		}
	}
	return false
}

func (s *CreationFlowServiceImpl) hasSelectedFavoredEnemy(char *character.Character) bool {
	for _, feature := range char.Features {
		if feature.Key == "favored_enemy" && feature.Metadata != nil {
			if _, ok := feature.Metadata["enemy_type"].(string); ok {
				return true
			}
		}
	}
	return false
}

func (s *CreationFlowServiceImpl) hasSelectedNaturalExplorer(char *character.Character) bool {
	for _, feature := range char.Features {
		if feature.Key == "natural_explorer" && feature.Metadata != nil {
			if _, ok := feature.Metadata["terrain_type"].(string); ok {
				return true
			}
		}
	}
	return false
}

func (s *CreationFlowServiceImpl) hasUserSelectedProficiencies(char *character.Character) bool {
	// For now, return false to ensure the step is shown
	// In the future, track which proficiencies are from user choices vs automatic
	return false
}

func (s *CreationFlowServiceImpl) hasUserSelectedEquipment(char *character.Character) bool {
	// For now, return false to ensure the step is shown
	// In the future, track which equipment is from user choices vs starting equipment
	return false
}

func (s *CreationFlowServiceImpl) hasSelectedCantrips(char *character.Character) bool {
	// Check if character has selected required cantrips
	if char.Spells == nil {
		return false
	}

	// Get required number of cantrips for the class/level
	requiredCantrips := s.getRequiredCantrips(char)
	if requiredCantrips == 0 {
		return true // No cantrips required
	}

	// Check if enough cantrips have been selected
	return len(char.Spells.Cantrips) >= requiredCantrips
}

func (s *CreationFlowServiceImpl) getRequiredCantrips(char *character.Character) int {
	if char.Class == nil {
		return 0
	}

	// Get cantrips known for this class at level 1
	// This should ideally come from the class data
	switch char.Class.Key {
	case "wizard":
		return 3
	case "cleric":
		return 3
	case "druid":
		return 2
	case "bard":
		return 2
	case "sorcerer":
		return 4
	case "warlock":
		return 2
	default:
		return 0
	}
}

func (s *CreationFlowServiceImpl) hasSelectedSpells(char *character.Character) bool {
	// Check if character has selected required spells
	if char.Spells == nil {
		return false
	}

	// Get required number of spells for the class/level
	requiredSpells := s.getRequiredSpells(char)
	if requiredSpells == 0 {
		return true // No spells required
	}

	// Check if enough spells have been selected
	return len(char.Spells.KnownSpells) >= requiredSpells
}

func (s *CreationFlowServiceImpl) getRequiredSpells(char *character.Character) int {
	if char.Class == nil {
		return 0
	}

	// Get spells known for this class at level 1
	switch char.Class.Key {
	case "wizard":
		return 6 // Wizards start with 6 spells in spellbook
	case "sorcerer":
		return 2 // Sorcerers know 2 spells at level 1
	case "bard":
		return 4 // Bards know 4 spells at level 1
	case "ranger":
		return 0 // Rangers don't get spells until level 2
	case "warlock":
		return 2 // Warlocks know 2 spells at level 1
	case "cleric", "druid", "paladin":
		return 0 // These classes prepare spells, not learn them
	default:
		return 0
	}
}

func (s *CreationFlowServiceImpl) hasSelectedSubclass(char *character.Character) bool {
	// TODO: Implement when we have subclass tracking
	// For now, return false to show the step
	return false
}

// applyStepResult applies the result of a step to the character
func (s *CreationFlowServiceImpl) applyStepResult(ctx context.Context, characterID string, result *character.CreationStepResult) error {
	char, err := s.characterService.GetByID(characterID)
	if err != nil {
		return fmt.Errorf("failed to get character: %w", err)
	}

	switch result.StepType {
	case character.StepTypeRaceSelection:
		return s.applyRaceSelection(ctx, char, result)
	case character.StepTypeClassSelection:
		return s.applyClassSelection(ctx, char, result)
	case character.StepTypeSkillSelection:
		return s.applySkillSelection(char, result)
	case character.StepTypeLanguageSelection:
		return s.applyLanguageSelection(char, result)
	case character.StepTypeCantripsSelection:
		return s.applyCantripSelection(char, result)
	case character.StepTypeSpellSelection, character.StepTypeSpellbookSelection, character.StepTypeSpellsKnownSelection:
		return s.applySpellSelection(char, result)
	// Add other step result handlers as needed
	default:
		// For steps handled by existing handlers, we don't need to do anything here
		return nil
	}
}

func (s *CreationFlowServiceImpl) applySkillSelection(char *character.Character, result *character.CreationStepResult) error {
	// Find the divine domain feature and add bonus skills
	for _, feature := range char.Features {
		if feature.Key != "divine_domain" {
			continue
		}

		if feature.Metadata == nil {
			feature.Metadata = make(map[string]any)
		}
		feature.Metadata["bonus_skills"] = result.Selections

		// Also add to character proficiencies
		for _, skillKey := range result.Selections {
			// Convert skill key to proficiency (this would need proper mapping)
			prof := &rulebook.Proficiency{
				Key:  skillKey,
				Name: skillKey, // This should be proper name mapping
				Type: rulebook.ProficiencyTypeSkill,
			}
			char.AddProficiency(prof)
		}

		// Log for debugging
		fmt.Printf("Applied skill selections to character %s: %v\n", char.ID, result.Selections)
		break
	}

	// Save the character
	return s.characterService.UpdateEquipment(char)
}

func (s *CreationFlowServiceImpl) applyLanguageSelection(char *character.Character, result *character.CreationStepResult) error {
	// Find the divine domain feature and add bonus languages
	for _, feature := range char.Features {
		if feature.Key != "divine_domain" {
			continue
		}

		if feature.Metadata == nil {
			feature.Metadata = make(map[string]any)
		}
		feature.Metadata["bonus_languages"] = result.Selections

		// Also add to character languages (would need proper language system)
		// For now just store in metadata
		break
	}

	// Save the character
	return s.characterService.UpdateEquipment(char)
}

func (s *CreationFlowServiceImpl) applyRaceSelection(ctx context.Context, char *character.Character, result *character.CreationStepResult) error {
	if len(result.Selections) == 0 {
		return fmt.Errorf("no race selected")
	}

	raceKey := result.Selections[0]

	// Update the character with the selected race
	updateInput := &UpdateDraftInput{
		RaceKey: &raceKey,
	}

	_, err := s.characterService.UpdateDraftCharacter(ctx, char.ID, updateInput)
	if err != nil {
		return fmt.Errorf("failed to update character race: %w", err)
	}

	return nil
}

func (s *CreationFlowServiceImpl) applyClassSelection(ctx context.Context, char *character.Character, result *character.CreationStepResult) error {
	if len(result.Selections) == 0 {
		return fmt.Errorf("no class selected")
	}

	classKey := result.Selections[0]

	// Update the character with the selected class
	updateInput := &UpdateDraftInput{
		ClassKey: &classKey,
	}

	_, err := s.characterService.UpdateDraftCharacter(ctx, char.ID, updateInput)
	if err != nil {
		return fmt.Errorf("failed to update character class: %w", err)
	}

	// Initialize Spells field for spellcasting classes
	if isSpellcastingClass(classKey) {
		// Get the updated character
		updatedChar, err := s.characterService.GetByID(char.ID)
		if err != nil {
			return fmt.Errorf("failed to get updated character: %w", err)
		}

		// Initialize Spells field if not already present
		if updatedChar.Spells == nil {
			updatedChar.Spells = &character.SpellList{
				Cantrips:       []string{},
				KnownSpells:    []string{},
				PreparedSpells: []string{},
			}

			// Save the updated character with initialized Spells
			err = s.characterService.UpdateEquipment(updatedChar)
			if err != nil {
				return fmt.Errorf("failed to initialize spells for character: %w", err)
			}
		}
	}

	return nil
}

// isSpellcastingClass returns true if the given class key represents a spellcasting class
func isSpellcastingClass(classKey string) bool {
	spellcastingClasses := map[string]bool{
		"wizard":    true,
		"cleric":    true,
		"sorcerer":  true,
		"warlock":   true,
		"bard":      true,
		"druid":     true,
		"paladin":   true,
		"ranger":    true,
		"artificer": true,
	}
	return spellcastingClasses[classKey]
}

func (s *CreationFlowServiceImpl) applyCantripSelection(char *character.Character, result *character.CreationStepResult) error {
	// Get a fresh copy of the character to ensure we have the latest state
	freshChar, err := s.characterService.GetByID(char.ID)
	if err != nil {
		return fmt.Errorf("failed to get character: %w", err)
	}

	// Initialize spells if needed
	if freshChar.Spells == nil {
		freshChar.Spells = &character.SpellList{
			Cantrips:       []string{},
			KnownSpells:    []string{},
			PreparedSpells: []string{},
		}
	}

	// Set the selected cantrips
	freshChar.Spells.Cantrips = result.Selections

	// Save the character
	return s.characterService.UpdateEquipment(freshChar)
}

func (s *CreationFlowServiceImpl) applySpellSelection(char *character.Character, result *character.CreationStepResult) error {
	// Get a fresh copy of the character to ensure we have the latest state
	freshChar, err := s.characterService.GetByID(char.ID)
	if err != nil {
		return fmt.Errorf("failed to get character: %w", err)
	}

	// Initialize spells if needed
	if freshChar.Spells == nil {
		freshChar.Spells = &character.SpellList{
			Cantrips:       []string{},
			KnownSpells:    []string{},
			PreparedSpells: []string{},
		}
	}

	// Set the selected spells based on step type
	switch result.StepType {
	case character.StepTypeSpellSelection, character.StepTypeSpellbookSelection, character.StepTypeSpellsKnownSelection:
		freshChar.Spells.KnownSpells = result.Selections
	}

	// Save the character
	return s.characterService.UpdateEquipment(freshChar)
}
