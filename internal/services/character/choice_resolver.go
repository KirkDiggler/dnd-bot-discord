package character

//go:generate mockgen -destination=mock/mock_choice_resolver.go -package=mockcharacter -source=choice_resolver.go

import (
	"context"
	"fmt"

	"github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
)

// ChoiceResolver handles the complex D&D 5e choice resolution
type ChoiceResolver interface {
	// ResolveProficiencyChoices returns simplified proficiency choices for a race/class
	ResolveProficiencyChoices(ctx context.Context, race *entities.Race, class *entities.Class) ([]SimplifiedChoice, error)
	
	// ResolveEquipmentChoices returns simplified equipment choices for a class
	ResolveEquipmentChoices(ctx context.Context, class *entities.Class) ([]SimplifiedChoice, error)
	
	// ValidateProficiencySelections validates that the selected proficiencies are valid
	ValidateProficiencySelections(ctx context.Context, race *entities.Race, class *entities.Class, selections []string) error
}

// choiceResolver implements ChoiceResolver
type choiceResolver struct {
	dndClient dnd5e.Client
}

// NewChoiceResolver creates a new choice resolver
func NewChoiceResolver(dndClient dnd5e.Client) ChoiceResolver {
	return &choiceResolver{
		dndClient: dndClient,
	}
}

// ResolveProficiencyChoices returns simplified proficiency choices
func (r *choiceResolver) ResolveProficiencyChoices(ctx context.Context, race *entities.Race, class *entities.Class) ([]SimplifiedChoice, error) {
	choices := []SimplifiedChoice{}
	
	// Process class proficiency choices
	for i, choice := range class.ProficiencyChoices {
		if choice == nil || len(choice.Options) == 0 {
			continue
		}
		
		simplified := SimplifiedChoice{
			ID:          fmt.Sprintf("%s-prof-%d", class.Key, i),
			Name:        choice.Name,
			Description: fmt.Sprintf("Choose %d", choice.Count),
			Type:        string(choice.Type),
			Choose:      choice.Count,
			Options:     r.extractOptions(choice.Options),
		}
		
		// Handle nested choices (like Monk's tools)
		if r.hasNestedChoices(choice.Options) {
			// For now, flatten to just the common ones
			simplified = r.flattenNestedChoice(class.Key, i, choice)
		}
		
		choices = append(choices, simplified)
	}
	
	// Process racial proficiency choices
	if race.StartingProficiencyOptions != nil && len(race.StartingProficiencyOptions.Options) > 0 {
		simplified := SimplifiedChoice{
			ID:          fmt.Sprintf("%s-prof", race.Key),
			Name:        race.StartingProficiencyOptions.Name,
			Description: fmt.Sprintf("Choose %d racial proficiency", race.StartingProficiencyOptions.Count),
			Type:        string(race.StartingProficiencyOptions.Type),
			Choose:      race.StartingProficiencyOptions.Count,
			Options:     r.extractOptions(race.StartingProficiencyOptions.Options),
		}
		choices = append(choices, simplified)
	}
	
	return choices, nil
}

// ResolveEquipmentChoices returns simplified equipment choices
func (r *choiceResolver) ResolveEquipmentChoices(ctx context.Context, class *entities.Class) ([]SimplifiedChoice, error) {
	choices := []SimplifiedChoice{}
	
	// Process starting equipment choices
	for i, choice := range class.StartingEquipmentChoices {
		if choice == nil || len(choice.Options) == 0 {
			continue
		}
		
		simplified := SimplifiedChoice{
			ID:          fmt.Sprintf("%s-equip-%d", class.Key, i),
			Name:        choice.Name,
			Description: fmt.Sprintf("Choose %d", choice.Count),
			Type:        "equipment",
			Choose:      choice.Count,
			Options:     r.extractOptions(choice.Options),
		}
		
		choices = append(choices, simplified)
	}
	
	return choices, nil
}

// ValidateProficiencySelections validates proficiency selections
func (r *choiceResolver) ValidateProficiencySelections(ctx context.Context, race *entities.Race, class *entities.Class, selections []string) error {
	// TODO: Implement validation
	// Check that each selection is valid for the available choices
	return nil
}

// extractOptions converts entity options to simple choice options
func (r *choiceResolver) extractOptions(options []entities.Option) []ChoiceOption {
	result := []ChoiceOption{}
	
	for _, opt := range options {
		switch o := opt.(type) {
		case *entities.ReferenceOption:
			if o.Reference != nil {
				result = append(result, ChoiceOption{
					Key:  o.Reference.Key,
					Name: o.Reference.Name,
				})
			}
		case *entities.CountedReferenceOption:
			if o.Reference != nil {
				result = append(result, ChoiceOption{
					Key:  o.Reference.Key,
					Name: fmt.Sprintf("%s (x%d)", o.Reference.Name, o.Count),
				})
			}
		// Skip nested choices for now
		}
	}
	
	return result
}

// hasNestedChoices checks if options contain nested choices
func (r *choiceResolver) hasNestedChoices(options []entities.Option) bool {
	for _, opt := range options {
		if _, ok := opt.(*entities.Choice); ok {
			return true
		}
	}
	return false
}

// flattenNestedChoice handles special cases like Monk's tool choices
func (r *choiceResolver) flattenNestedChoice(classKey string, index int, choice *entities.Choice) SimplifiedChoice {
	// Special handling for known nested choices
	if classKey == "monk" && index == 1 {
		// Monk's second choice: artisan tools or musical instrument
		return SimplifiedChoice{
			ID:          fmt.Sprintf("%s-prof-%d", classKey, index),
			Name:        "Tools or Instrument",
			Description: "Choose 1 artisan's tool or musical instrument",
			Type:        "tool",
			Choose:      1,
			Options: []ChoiceOption{
				// Common tools that would be in the nested choices
				{Key: "alchemists-supplies", Name: "Alchemist's Supplies"},
				{Key: "brewers-supplies", Name: "Brewer's Supplies"},
				{Key: "calligraphers-supplies", Name: "Calligrapher's Supplies"},
				{Key: "carpenters-tools", Name: "Carpenter's Tools"},
				{Key: "cooks-utensils", Name: "Cook's Utensils"},
				{Key: "flute", Name: "Flute"},
				{Key: "lute", Name: "Lute"},
				{Key: "horn", Name: "Horn"},
			},
		}
	}
	
	// Default: return the original with empty options
	return SimplifiedChoice{
		ID:          fmt.Sprintf("%s-prof-%d", classKey, index),
		Name:        choice.Name,
		Description: fmt.Sprintf("Choose %d", choice.Count),
		Type:        string(choice.Type),
		Choose:      choice.Count,
		Options:     []ChoiceOption{},
	}
}