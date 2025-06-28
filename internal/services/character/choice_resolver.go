package character

//go:generate mockgen -destination=mock/mock_choice_resolver.go -package=mockcharacters -source=choice_resolver.go

import (
	"context"
	"fmt"
	"strings"

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
			// Only flatten for specific known cases
			if class.Key == "monk" && i == 1 {
				simplified = r.flattenNestedChoice(class.Key, i, choice)
			}
			// For other classes, skip nested choices for now
			// TODO: Properly handle nested choices for all classes
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

	// Handle nil class
	if class == nil {
		return choices, nil
	}

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
			Options:     r.extractEquipmentOptions(choice.Options),
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

// extractEquipmentOptions converts equipment options with descriptions
func (r *choiceResolver) extractEquipmentOptions(options []entities.Option) []ChoiceOption {
	result := []ChoiceOption{}

	for _, opt := range options {
		if opt == nil {
			continue
		}
		switch o := opt.(type) {
		case *entities.ReferenceOption:
			if o.Reference != nil && o.Reference.Key != "" && o.Reference.Name != "" {
				choiceOpt := ChoiceOption{
					Key:  o.Reference.Key,
					Name: o.Reference.Name,
				}

				// Add equipment-specific descriptions
				desc := r.getEquipmentDescription(o.Reference.Key, o.Reference.Name)
				if desc != "" {
					choiceOpt.Description = desc
				}

				result = append(result, choiceOpt)
			}
		case *entities.CountedReferenceOption:
			if o.Reference != nil && o.Reference.Key != "" && o.Reference.Name != "" {
				name := o.Reference.Name
				if o.Count > 1 {
					name = fmt.Sprintf("%dx %s", o.Count, name)
				}

				choiceOpt := ChoiceOption{
					Key:  o.Reference.Key,
					Name: name,
				}

				// Add equipment-specific descriptions
				desc := r.getEquipmentDescription(o.Reference.Key, o.Reference.Name)
				if desc != "" {
					choiceOpt.Description = desc
				}

				result = append(result, choiceOpt)
			}
		case *entities.MultipleOption:
			// Skip if no items
			if len(o.Items) == 0 {
				continue
			}

			// Handle bundles like "weapon and shield" or "two weapons"
			names := []string{}
			descriptions := []string{}
			bundleItems := []string{} // Track non-choice items in the bundle
			bundleKey := o.Key
			if bundleKey == "" {
				bundleKey = fmt.Sprintf("bundle-%d", len(result))
			}
			hasNestedChoice := false
			nestedChoiceDesc := ""

			for _, item := range o.Items {
				if item == nil {
					continue
				}
				switch itemRef := item.(type) {
				case *entities.CountedReferenceOption:
					if itemRef.Reference != nil && itemRef.Reference.Key != "" && itemRef.Reference.Name != "" {
						itemName := itemRef.Reference.Name
						if itemRef.Count > 1 {
							itemName = fmt.Sprintf("%dx %s", itemRef.Count, itemName)
						}
						names = append(names, itemName)

						// Track this as a bundle item (e.g., shield in weapon+shield)
						bundleItems = append(bundleItems, itemRef.Reference.Key)

						// Get description for this item
						if desc := r.getEquipmentDescription(itemRef.Reference.Key, itemRef.Reference.Name); desc != "" {
							descriptions = append(descriptions, fmt.Sprintf("%s (%s)", itemRef.Reference.Name, desc))
						}
					}
				case *entities.ReferenceOption:
					if itemRef.Reference != nil && itemRef.Reference.Key != "" && itemRef.Reference.Name != "" {
						names = append(names, itemRef.Reference.Name)

						// Track this as a bundle item
						bundleItems = append(bundleItems, itemRef.Reference.Key)

						// Get description for this item
						if desc := r.getEquipmentDescription(itemRef.Reference.Key, itemRef.Reference.Name); desc != "" {
							descriptions = append(descriptions, fmt.Sprintf("%s (%s)", itemRef.Reference.Name, desc))
						}
					}
				case *entities.Choice:
					// Handle nested choices like "a martial weapon" or "two martial weapons"
					hasNestedChoice = true
					if itemRef.Count > 1 {
						names = append(names, fmt.Sprintf("%d %s", itemRef.Count, itemRef.Name))
						nestedChoiceDesc = fmt.Sprintf("Choose %d %s", itemRef.Count, itemRef.Name)
					} else {
						names = append(names, itemRef.Name)
						nestedChoiceDesc = fmt.Sprintf("Choose 1 %s", itemRef.Name)
					}
				}
			}

			if len(names) > 0 {
				choiceOpt := ChoiceOption{
					Key:  bundleKey,
					Name: joinWithAnd(names),
				}

				// If this bundle contains nested choices, mark it specially
				if hasNestedChoice {
					choiceOpt.Key = fmt.Sprintf("nested-%d", len(result))
					if nestedChoiceDesc != "" {
						choiceOpt.Description = nestedChoiceDesc
					}
					// Add bundle items to the nested choice
					choiceOpt.BundleItems = bundleItems
				} else if len(descriptions) > 0 {
					choiceOpt.Description = strings.Join(descriptions, ", ")
				}

				result = append(result, choiceOpt)
			}
		case *entities.Choice:
			// Handle nested equipment choices like "any martial weapon"
			if o.Type == entities.ChoiceTypeEquipment {
				// Mark as nested choice so it triggers the weapon selection UI
				result = append(result, ChoiceOption{
					Key:         fmt.Sprintf("nested-%d", len(result)),
					Name:        o.Name,
					Description: fmt.Sprintf("Choose %d from %s", o.Count, o.Name),
				})
			}
		}
	}

	return result
}

// joinWithAnd joins strings with commas and "and" before the last item
func joinWithAnd(items []string) string {
	if len(items) == 0 {
		return ""
	}
	if len(items) == 1 {
		return items[0]
	}
	if len(items) == 2 {
		return items[0] + " and " + items[1]
	}
	result := ""
	for i, item := range items {
		if i == len(items)-1 {
			result += " and " + item
		} else if i > 0 {
			result += ", " + item
		} else {
			result += item
		}
	}
	return result
}

// getEquipmentDescription returns a description for common equipment
func (r *choiceResolver) getEquipmentDescription(key, name string) string {
	// Common weapon descriptions
	weaponDescs := map[string]string{
		"longsword":      "1d8 slashing, versatile (1d10)",
		"shortsword":     "1d6 piercing, finesse, light",
		"battleaxe":      "1d8 slashing, versatile (1d10)",
		"handaxe":        "1d6 slashing, light, thrown (20/60)",
		"warhammer":      "1d8 bludgeoning, versatile (1d10)",
		"mace":           "1d6 bludgeoning",
		"greataxe":       "1d12 slashing, heavy, two-handed",
		"greatsword":     "2d6 slashing, heavy, two-handed",
		"rapier":         "1d8 piercing, finesse",
		"scimitar":       "1d6 slashing, finesse, light",
		"shortbow":       "1d6 piercing, range 80/320",
		"longbow":        "1d8 piercing, range 150/600",
		"light-crossbow": "1d8 piercing, range 80/320",
		"shield":         "+2 AC",
		"dagger":         "1d4 piercing, finesse, light, thrown (20/60)",
		"quarterstaff":   "1d6 bludgeoning, versatile (1d8)",
		"spear":          "1d6 piercing, thrown (20/60), versatile (1d8)",
		"javelin":        "1d6 piercing, thrown (30/120)",
		"club":           "1d4 bludgeoning, light",
	}

	// Common armor descriptions
	armorDescs := map[string]string{
		"leather-armor":   "11 + Dex modifier",
		"scale-mail":      "14 + Dex (max 2)",
		"chain-mail":      "16 AC",
		"chain-shirt":     "13 + Dex (max 2)",
		"padded-armor":    "11 + Dex modifier",
		"studded-leather": "12 + Dex modifier",
		"hide-armor":      "12 + Dex (max 2)",
		"ring-mail":       "14 AC",
		"splint-armor":    "17 AC",
		"plate-armor":     "18 AC",
	}

	// Check weapon descriptions
	if desc, ok := weaponDescs[key]; ok {
		return desc
	}

	// Check armor descriptions
	if desc, ok := armorDescs[key]; ok {
		return desc
	}

	// Check by name if key didn't match
	lowerName := strings.ToLower(name)
	for k, v := range weaponDescs {
		if strings.Contains(lowerName, strings.ReplaceAll(k, "-", " ")) {
			return v
		}
	}
	for k, v := range armorDescs {
		if strings.Contains(lowerName, strings.ReplaceAll(k, "-", " ")) {
			return v
		}
	}

	return ""
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
