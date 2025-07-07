package character

import (
	"net/http"
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFighterProficienciesFromAPI checks what proficiencies the Fighter gets from the D&D API
func TestFighterProficienciesFromAPI(t *testing.T) {
	// Skip in CI, this is for local testing with real API
	if testing.Short() {
		t.Skip("Skipping API test in short mode")
	}

	// Create real API client
	cfg := &dnd5e.Config{
		HttpClient: http.DefaultClient,
	}
	client, err := dnd5e.New(cfg)
	require.NoError(t, err)

	// Get fighter class
	fighter, err := client.GetClass("fighter")
	require.NoError(t, err)
	require.NotNil(t, fighter)

	// Log all proficiencies
	t.Logf("Fighter has %d automatic proficiencies:", len(fighter.Proficiencies))
	for _, prof := range fighter.Proficiencies {
		t.Logf("  - %s (%s)", prof.Name, prof.Key)
	}

	// Log proficiency choices
	t.Logf("\nFighter has %d proficiency choice groups:", len(fighter.ProficiencyChoices))
	for i, choice := range fighter.ProficiencyChoices {
		t.Logf("\nChoice %d: %s (type: %s, count: %d)", i+1, choice.Name, choice.Type, choice.Count)
		t.Logf("  Has %d options", len(choice.Options))
		for _, opt := range choice.Options {
			t.Logf("  - %s (%s)", opt.GetName(), opt.GetKey())
		}
	}

	// Check expected proficiencies
	expectedProfs := map[string]bool{
		"light-armor":      false,
		"medium-armor":     false,
		"heavy-armor":      false,
		"shields":          false,
		"simple-weapons":   false,
		"martial-weapons":  false,
		"saving-throw-str": false,
		"saving-throw-con": false,
	}

	// Mark found proficiencies
	for _, prof := range fighter.Proficiencies {
		if _, exists := expectedProfs[prof.Key]; exists {
			expectedProfs[prof.Key] = true
		}
	}

	// Report missing proficiencies
	t.Log("\nProficiency Check Results:")
	for key, found := range expectedProfs {
		if found {
			t.Logf("✓ %s", key)
		} else {
			t.Logf("✗ %s (MISSING)", key)
		}
	}
}

// TestBarbarianProficienciesFromAPI checks Barbarian proficiencies
func TestBarbarianProficienciesFromAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping API test in short mode")
	}

	cfg := &dnd5e.Config{
		HttpClient: http.DefaultClient,
	}
	client, err := dnd5e.New(cfg)
	require.NoError(t, err)

	barbarian, err := client.GetClass("barbarian")
	require.NoError(t, err)
	require.NotNil(t, barbarian)

	t.Logf("Barbarian has %d automatic proficiencies:", len(barbarian.Proficiencies))
	for _, prof := range barbarian.Proficiencies {
		t.Logf("  - %s (%s)", prof.Name, prof.Key)
	}

	// Barbarian should have same armor/weapon proficiencies as Fighter
	expectedProfs := []string{
		"light-armor",
		"medium-armor",
		"shields",
		"simple-weapons",
		"martial-weapons",
		"saving-throw-str",
		"saving-throw-con",
	}

	foundProfs := make(map[string]bool)
	for _, prof := range barbarian.Proficiencies {
		foundProfs[prof.Key] = true
	}

	for _, expected := range expectedProfs {
		assert.True(t, foundProfs[expected], "Barbarian should have %s proficiency", expected)
	}
}

// TestPaladinProficienciesFromAPI checks Paladin proficiencies
func TestPaladinProficienciesFromAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping API test in short mode")
	}

	cfg := &dnd5e.Config{
		HttpClient: http.DefaultClient,
	}
	client, err := dnd5e.New(cfg)
	require.NoError(t, err)

	paladin, err := client.GetClass("paladin")
	require.NoError(t, err)
	require.NotNil(t, paladin)

	t.Logf("Paladin has %d automatic proficiencies:", len(paladin.Proficiencies))
	for _, prof := range paladin.Proficiencies {
		t.Logf("  - %s (%s)", prof.Name, prof.Key)
	}

	// Paladin has different saving throws
	expectedSaves := []string{
		"saving-throw-wis",
		"saving-throw-cha",
	}

	foundProfs := make(map[string]bool)
	for _, prof := range paladin.Proficiencies {
		foundProfs[prof.Key] = true
	}

	for _, expected := range expectedSaves {
		assert.True(t, foundProfs[expected], "Paladin should have %s proficiency", expected)
	}
}
