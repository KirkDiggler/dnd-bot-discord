package feats

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
	rpgevents "github.com/KirkDiggler/rpg-toolkit/events"
)

// Feat represents a D&D 5e feat
type Feat interface {
	// Key returns the unique identifier for this feat
	Key() string

	// Name returns the display name
	Name() string

	// Description returns the feat description
	Description() string

	// Prerequisites returns any requirements for taking this feat
	Prerequisites() []Prerequisite

	// CanTake checks if a character meets the prerequisites
	CanTake(char *character.Character) bool

	// Apply applies the feat's benefits to a character
	Apply(char *character.Character) error

	// RegisterHandlers registers any event handlers for this feat
	RegisterHandlers(bus *rpgevents.Bus, char *character.Character)
}

// Prerequisite represents a requirement for taking a feat
type Prerequisite struct {
	Type        string // "ability_score", "proficiency", "spellcasting", "level"
	Requirement string // Description of the requirement
	Check       func(*character.Character) bool
}

// BaseFeat provides common feat functionality
type BaseFeat struct {
	key           string
	name          string
	description   string
	prerequisites []Prerequisite
}

// Key returns the feat's unique identifier
func (f *BaseFeat) Key() string {
	return f.key
}

// Name returns the feat's display name
func (f *BaseFeat) Name() string {
	return f.name
}

// Description returns the feat's description
func (f *BaseFeat) Description() string {
	return f.description
}

// Prerequisites returns the feat's prerequisites
func (f *BaseFeat) Prerequisites() []Prerequisite {
	return f.prerequisites
}

// CanTake checks if a character can take this feat
func (f *BaseFeat) CanTake(char *character.Character) bool {
	for _, prereq := range f.prerequisites {
		if prereq.Check != nil && !prereq.Check(char) {
			return false
		}
	}
	return true
}

// Apply is a default implementation that adds the feat to the character
func (f *BaseFeat) Apply(char *character.Character) error {
	// Add feat as a character feature
	feature := &rulebook.CharacterFeature{
		Key:         f.key,
		Name:        f.name,
		Description: f.description,
		Type:        rulebook.FeatureTypeFeat,
		Level:       char.Level,
		Source:      "Feat",
	}

	char.Features = append(char.Features, feature)
	return nil
}

// RegisterHandlers is a default no-op implementation
func (f *BaseFeat) RegisterHandlers(bus *rpgevents.Bus, char *character.Character) {
	// Most feats don't need event handlers
}
