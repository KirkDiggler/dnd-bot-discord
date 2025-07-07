package rulebook

// FeatureChoiceType represents the type of choice
type FeatureChoiceType string

const (
	FeatureChoiceTypeFightingStyle   FeatureChoiceType = "fighting_style"
	FeatureChoiceTypeDivineDomain    FeatureChoiceType = "divine_domain"
	FeatureChoiceTypeFavoredEnemy    FeatureChoiceType = "favored_enemy"
	FeatureChoiceTypeNaturalExplorer FeatureChoiceType = "natural_explorer"
	// Future additions:
	// FeatureChoiceTypeSorcerousOrigin FeatureChoiceType = "sorcerous_origin"
	// FeatureChoiceTypeOtherworldlyPatron FeatureChoiceType = "otherworldly_patron"
)

// FeatureChoice represents a choice that must be made for a feature
type FeatureChoice struct {
	Type        FeatureChoiceType `json:"type"`
	FeatureKey  string            `json:"feature_key"` // Links to CharacterFeature.Key
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Choose      int               `json:"choose"` // How many to choose
	Options     []FeatureOption   `json:"options"`
}

// FeatureOption represents one option in a feature choice
type FeatureOption struct {
	Key         string         `json:"key"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// ClassFeatures represents all features and choices for a class at each level
type ClassFeatures struct {
	ClassName string                      `json:"class_name"`
	Features  map[int][]ClassLevelFeature `json:"features"` // Keyed by level
}

// ClassLevelFeature represents a feature gained at a specific level
type ClassLevelFeature struct {
	Level   int               `json:"level"`
	Feature *CharacterFeature `json:"feature"`
	Choice  *FeatureChoice    `json:"choice,omitempty"` // Optional choice required
}
