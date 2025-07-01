package rulebook

// FeatureType represents the type of feature
type FeatureType string

const (
	FeatureTypeRacial  FeatureType = "racial"
	FeatureTypeClass   FeatureType = "class"
	FeatureTypeSubrace FeatureType = "subrace"
	FeatureTypeFeat    FeatureType = "feat"
)

// CharacterFeature represents a character feature (trait, ability, etc)
type CharacterFeature struct {
	Key         string         `json:"key"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Type        FeatureType    `json:"type"`
	Level       int            `json:"level"`              // Level when gained (0 for racial)
	Source      string         `json:"source"`             // Race/Class/Subclass name
	Metadata    map[string]any `json:"metadata,omitempty"` // For storing selections like favored enemy type
}
