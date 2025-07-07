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
			Description: "When you are wielding a melee weapon in one hand and no other weapons, you gain a +2 bonus to damage rolls with that weapon.",
			Classes:     []string{"fighter", "ranger", "paladin"},
		},
		{
			Key:         "great_weapon_fighting",
			Name:        "Great Weapon Fighting",
			Description: "When you roll a 1 or 2 on a damage die for an attack you make with a melee weapon that you are wielding with two hands, you can reroll the die and must use the new roll. The weapon must have the two-handed or versatile property.",
			Classes:     []string{"fighter", "paladin"},
		},
		{
			Key:         "protection",
			Name:        "Protection",
			Description: "When a creature you can see attacks a target other than you that is within 5 feet of you, you can use your reaction to impose disadvantage on the attack roll. You must be wielding a shield.",
			Classes:     []string{"fighter", "paladin"},
		},
		{
			Key:         "two_weapon_fighting",
			Name:        "Two-Weapon Fighting",
			Description: "When you engage in two-weapon fighting, you can add your ability modifier to the damage of the second attack.",
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
