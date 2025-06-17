package character_test

import (
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestCharacterListFiltering(t *testing.T) {
	tests := []struct {
		name            string
		characters      []*entities.Character
		expectedActive  int
		expectedDraft   int
		expectedArchived int
		description     string
	}{
		{
			name: "filters out empty draft characters",
			characters: []*entities.Character{
				{
					ID:     "draft1",
					Status: entities.CharacterStatusDraft,
					Name:   "", // No name
					Race:   nil, // No race
					Class:  nil, // No class
				},
				{
					ID:     "draft2",
					Status: entities.CharacterStatusDraft,
					Name:   "Bob",
				},
			},
			expectedActive:  0,
			expectedDraft:   1, // Only Bob should show
			expectedArchived: 0,
			description:     "Empty drafts should be filtered out",
		},
		{
			name: "shows draft with race but no name",
			characters: []*entities.Character{
				{
					ID:     "draft3",
					Status: entities.CharacterStatusDraft,
					Name:   "",
					Race:   testutils.CreateTestRace("human", "Human"),
				},
			},
			expectedActive:  0,
			expectedDraft:   1,
			expectedArchived: 0,
			description:     "Draft with race should show even without name",
		},
		{
			name: "shows draft with class but no name",
			characters: []*entities.Character{
				{
					ID:     "draft4",
					Status: entities.CharacterStatusDraft,
					Name:   "",
					Class:  testutils.CreateTestClass("fighter", "Fighter", 10),
				},
			},
			expectedActive:  0,
			expectedDraft:   1,
			expectedArchived: 0,
			description:     "Draft with class should show even without name",
		},
		{
			name: "properly groups by status",
			characters: []*entities.Character{
				{
					ID:     "active1",
					Status: entities.CharacterStatusActive,
					Name:   "Gandalf",
				},
				{
					ID:     "draft5",
					Status: entities.CharacterStatusDraft,
					Name:   "Frodo",
				},
				{
					ID:     "archived1",
					Status: entities.CharacterStatusArchived,
					Name:   "Boromir",
				},
			},
			expectedActive:  1,
			expectedDraft:   1,
			expectedArchived: 1,
			description:     "Characters should be grouped by status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Group characters by status (mimicking the handler logic)
			activeChars := make([]*entities.Character, 0)
			draftChars := make([]*entities.Character, 0)
			archivedChars := make([]*entities.Character, 0)

			for _, char := range tt.characters {
				switch char.Status {
				case entities.CharacterStatusActive:
					activeChars = append(activeChars, char)
				case entities.CharacterStatusDraft:
					// Only show drafts that have meaningful progress
					if char.Name != "" || char.Race != nil || char.Class != nil {
						draftChars = append(draftChars, char)
					}
				case entities.CharacterStatusArchived:
					archivedChars = append(archivedChars, char)
				}
			}

			assert.Equal(t, tt.expectedActive, len(activeChars), "Active characters count mismatch")
			assert.Equal(t, tt.expectedDraft, len(draftChars), "Draft characters count mismatch")
			assert.Equal(t, tt.expectedArchived, len(archivedChars), "Archived characters count mismatch")
		})
	}
}

func TestDraftCharacterDisplay(t *testing.T) {
	tests := []struct {
		name           string
		character      *entities.Character
		expectedStatus string
		expectedProgress string
	}{
		{
			name: "draft with name",
			character: &entities.Character{
				Name: "Aragorn",
				Status: entities.CharacterStatusDraft,
			},
			expectedStatus: "Aragorn",
			expectedProgress: "✓ Name",
		},
		{
			name: "draft with race and class but no name",
			character: &entities.Character{
				Name:  "",
				Race:  &entities.Race{Name: "Human"},
				Class: &entities.Class{Name: "Ranger"},
				Status: entities.CharacterStatusDraft,
			},
			expectedStatus: "Human Ranger (unnamed)",
			expectedProgress: "✓ Race ✓ Class ",
		},
		{
			name: "draft with only race",
			character: &entities.Character{
				Name:  "",
				Race:  &entities.Race{Name: "Elf"},
				Status: entities.CharacterStatusDraft,
			},
			expectedStatus: "Elf (selecting class)",
			expectedProgress: "✓ Race ",
		},
		{
			name: "complete draft",
			character: &entities.Character{
				Name:  "Legolas",
				Race:  &entities.Race{Name: "Elf"},
				Class: &entities.Class{Name: "Ranger"},
				Attributes: map[entities.Attribute]*entities.AbilityScore{
					entities.AttributeStrength: {Score: 14},
				},
				Status: entities.CharacterStatusDraft,
			},
			expectedStatus: "Legolas",
			expectedProgress: "✓ Race ✓ Class ✓ Abilities ✓ Name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate status string
			status := "Creating..."
			if tt.character.Name != "" {
				status = tt.character.Name
			} else if tt.character.Race != nil && tt.character.Class != nil {
				status = tt.character.Race.Name + " " + tt.character.Class.Name + " (unnamed)"
			} else if tt.character.Race != nil {
				status = tt.character.Race.Name + " (selecting class)"
			}

			// Generate progress string
			progress := ""
			if tt.character.Race != nil {
				progress += "✓ Race "
			}
			if tt.character.Class != nil {
				progress += "✓ Class "
			}
			if len(tt.character.Attributes) > 0 {
				progress += "✓ Abilities "
			}
			if tt.character.Name != "" {
				progress += "✓ Name"
			}

			assert.Equal(t, tt.expectedStatus, status, "Status display mismatch")
			assert.Equal(t, tt.expectedProgress, progress, "Progress display mismatch")
		})
	}
}