package rulebook

// GetClassFeatureChoices returns all feature choices required for a class at a given level
func GetClassFeatureChoices(className string, level int) []*FeatureChoice {
	var choices []*FeatureChoice

	// Only return choices for features gained up to the specified level
	switch className {
	case "cleric":
		if level >= 1 {
			choices = append(choices, GetDivineDomainChoice())
		}
	case "fighter":
		if level >= 1 {
			choices = append(choices, GetFightingStyleChoice(className))
		}
	case "ranger":
		if level >= 1 {
			choices = append(choices, GetFavoredEnemyChoice(), GetNaturalExplorerChoice())
		}
	case "paladin":
		if level >= 2 {
			choices = append(choices, GetFightingStyleChoice(className))
		}
	// Other classes don't have choices at early levels
	case "barbarian", "bard", "monk", "rogue":
		// No choices at level 1
	// Not yet implemented classes
	case "druid", "sorcerer", "warlock", "wizard":
		// TODO: Add when these classes are fully implemented
	}

	return choices
}

// GetPendingFeatureChoices returns feature choices that haven't been made yet
func GetPendingFeatureChoices(char *CharacterFeature, className string, level int) []*FeatureChoice {
	allChoices := GetClassFeatureChoices(className, level)
	var pendingChoices []*FeatureChoice

	// Check which choices have already been made
	for _, choice := range allChoices {
		// Check if this choice has been made by looking at feature metadata
		choiceMade := false
		if char != nil && char.Key == choice.FeatureKey {
			if char.Metadata != nil {
				// Different features store their selections differently
				switch choice.Type {
				case FeatureChoiceTypeFightingStyle:
					_, choiceMade = char.Metadata["style"]
				case FeatureChoiceTypeDivineDomain:
					_, choiceMade = char.Metadata["domain"]
				case FeatureChoiceTypeFavoredEnemy:
					_, choiceMade = char.Metadata["enemy_type"]
				case FeatureChoiceTypeNaturalExplorer:
					_, choiceMade = char.Metadata["terrain_type"]
				}
			}
		}

		if !choiceMade {
			pendingChoices = append(pendingChoices, choice)
		}
	}

	return pendingChoices
}

// GetFeatureChoiceByType returns a specific feature choice for a class
func GetFeatureChoiceByType(className string, choiceType FeatureChoiceType) *FeatureChoice {
	choices := GetClassFeatureChoices(className, 20) // Get all choices up to level 20

	for _, choice := range choices {
		if choice.Type == choiceType {
			return choice
		}
	}

	return nil
}
