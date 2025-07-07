package rulebook

// FightingStyle represents a combat style that can be chosen by certain classes
type FightingStyle struct {
	Key         string   `json:"key"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Classes     []string `json:"classes"` // Classes that can choose this style
}

// GetFightingStyles returns all available fighting styles
func GetFightingStyles() []FightingStyle {
	return []FightingStyle{
		{
			Key:         "archery",
			Name:        "Archery",
			Description: "You gain a +2 bonus to attack rolls you make with ranged weapons.",
			Classes:     []string{"fighter", "ranger"},
		},
		{
			Key:         "defense",
			Name:        "Defense",
			Description: "While you are wearing armor, you gain a +1 bonus to AC.",
			Classes:     []string{"fighter", "ranger", "paladin"},
		},
		{
			Key:         "dueling",
			Name:        "Dueling",
			Description: "+2 damage when wielding a melee weapon in one hand with no other weapons.",
			Classes:     []string{"fighter", "ranger", "paladin"},
		},
		{
			Key:         "great_weapon_fighting",
			Name:        "Great Weapon Fighting",
			Description: "Reroll 1-2 on damage dice with two-handed or versatile weapons.",
			Classes:     []string{"fighter", "paladin"},
		},
		{
			Key:         "protection",
			Name:        "Protection",
			Description: "Use reaction with shield to impose disadvantage on an attack near you.",
			Classes:     []string{"fighter", "paladin"},
		},
		{
			Key:         "two_weapon_fighting",
			Name:        "Two-Weapon Fighting",
			Description: "Add ability modifier to off-hand weapon damage.",
			Classes:     []string{"fighter", "ranger"},
		},
	}
}

// GetFightingStylesForClass returns fighting styles available to a specific class
func GetFightingStylesForClass(className string) []FightingStyle {
	allStyles := GetFightingStyles()
	var classStyles []FightingStyle

	for _, style := range allStyles {
		for _, class := range style.Classes {
			if class == className {
				classStyles = append(classStyles, style)
				break
			}
		}
	}

	return classStyles
}

// GetFightingStyleChoice returns a FeatureChoice for fighting style selection
func GetFightingStyleChoice(className string) *FeatureChoice {
	styles := GetFightingStylesForClass(className)
	if len(styles) == 0 {
		return nil
	}

	options := make([]FeatureOption, len(styles))
	for i, style := range styles {
		options[i] = FeatureOption{
			Key:         style.Key,
			Name:        style.Name,
			Description: style.Description,
		}
	}

	return &FeatureChoice{
		Type:        FeatureChoiceTypeFightingStyle,
		FeatureKey:  "fighting_style",
		Name:        "Fighting Style",
		Description: "You adopt a particular style of fighting as your specialty.",
		Choose:      1,
		Options:     options,
	}
}
