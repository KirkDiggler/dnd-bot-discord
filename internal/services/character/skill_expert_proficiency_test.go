package character

import (
	"net/http"
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRogueProficienciesFromAPI checks what proficiencies the Rogue gets from the D&D API
func TestRogueProficienciesFromAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping API test in short mode")
	}

	cfg := &dnd5e.Config{
		HttpClient: http.DefaultClient,
	}
	client, err := dnd5e.New(cfg)
	require.NoError(t, err)

	// Get rogue class
	rogue, err := client.GetClass("rogue")
	require.NoError(t, err)
	require.NotNil(t, rogue)

	// Log all proficiencies
	t.Logf("Rogue has %d automatic proficiencies:", len(rogue.Proficiencies))
	for _, prof := range rogue.Proficiencies {
		t.Logf("  - %s (%s)", prof.Name, prof.Key)
	}

	// Log proficiency choices
	t.Logf("\nRogue has %d proficiency choice groups:", len(rogue.ProficiencyChoices))
	for i, choice := range rogue.ProficiencyChoices {
		t.Logf("\nChoice %d: %s (type: %s, count: %d)", i+1, choice.Name, choice.Type, choice.Count)
		t.Logf("  Has %d options", len(choice.Options))
		// Just log first few to avoid spam
		optCount := len(choice.Options)
		if optCount > 5 {
			optCount = 5
		}
		for j := 0; j < optCount; j++ {
			opt := choice.Options[j]
			t.Logf("  - %s (%s)", opt.GetName(), opt.GetKey())
		}
		if len(choice.Options) > 5 {
			t.Logf("  ... and %d more", len(choice.Options)-5)
		}
	}

	// Check expected proficiencies
	expectedProfs := []string{
		"light-armor",
		"simple-weapons",
		"thieves-tools",
		"saving-throw-dex",
		"saving-throw-int",
	}

	foundProfs := make(map[string]bool)
	for _, prof := range rogue.Proficiencies {
		foundProfs[prof.Key] = true
	}

	for _, expected := range expectedProfs {
		assert.True(t, foundProfs[expected], "Rogue should have %s proficiency", expected)
	}

	// Rogue should choose 4 skills
	assert.GreaterOrEqual(t, len(rogue.ProficiencyChoices), 1, "Rogue should have skill choices")
	if len(rogue.ProficiencyChoices) > 0 {
		skillChoice := rogue.ProficiencyChoices[0]
		assert.Equal(t, 4, skillChoice.Count, "Rogue should choose 4 skills")
	}
}

// TestBardProficienciesFromAPI checks Bard proficiencies
func TestBardProficienciesFromAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping API test in short mode")
	}

	cfg := &dnd5e.Config{
		HttpClient: http.DefaultClient,
	}
	client, err := dnd5e.New(cfg)
	require.NoError(t, err)

	bard, err := client.GetClass("bard")
	require.NoError(t, err)
	require.NotNil(t, bard)

	t.Logf("Bard has %d automatic proficiencies:", len(bard.Proficiencies))
	for _, prof := range bard.Proficiencies {
		t.Logf("  - %s (%s)", prof.Name, prof.Key)
	}

	// Log proficiency choices (skills and instruments)
	t.Logf("\nBard has %d proficiency choice groups:", len(bard.ProficiencyChoices))
	for i, choice := range bard.ProficiencyChoices {
		t.Logf("\nChoice %d: %s (type: %s, count: %d)", i+1, choice.Name, choice.Type, choice.Count)
		t.Logf("  Has %d options", len(choice.Options))
	}

	// Check expected proficiencies
	expectedProfs := []string{
		"light-armor",
		"simple-weapons",
		"saving-throw-dex",
		"saving-throw-cha",
	}

	foundProfs := make(map[string]bool)
	for _, prof := range bard.Proficiencies {
		foundProfs[prof.Key] = true
	}

	for _, expected := range expectedProfs {
		assert.True(t, foundProfs[expected], "Bard should have %s proficiency", expected)
	}

	// Bard gets specific weapon proficiencies too
	additionalWeapons := []string{"hand-crossbows", "longswords", "rapiers", "shortswords"}
	for _, weapon := range additionalWeapons {
		if foundProfs[weapon] {
			t.Logf("Bard has %s proficiency", weapon)
		}
	}
}

// TestRangerProficienciesFromAPI checks Ranger proficiencies (already partially tested)
func TestRangerProficienciesFromAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping API test in short mode")
	}

	cfg := &dnd5e.Config{
		HttpClient: http.DefaultClient,
	}
	client, err := dnd5e.New(cfg)
	require.NoError(t, err)

	ranger, err := client.GetClass("ranger")
	require.NoError(t, err)
	require.NotNil(t, ranger)

	t.Logf("Ranger has %d automatic proficiencies:", len(ranger.Proficiencies))
	for _, prof := range ranger.Proficiencies {
		t.Logf("  - %s (%s)", prof.Name, prof.Key)
	}

	// Check expected proficiencies (similar to martial but with DEX save)
	expectedProfs := []string{
		"light-armor",
		"medium-armor",
		"shields",
		"simple-weapons",
		"martial-weapons",
		"saving-throw-str",
		"saving-throw-dex",
	}

	foundProfs := make(map[string]bool)
	for _, prof := range ranger.Proficiencies {
		foundProfs[prof.Key] = true
	}

	for _, expected := range expectedProfs {
		assert.True(t, foundProfs[expected], "Ranger should have %s proficiency", expected)
	}

	// Ranger should choose 3 skills
	assert.GreaterOrEqual(t, len(ranger.ProficiencyChoices), 1, "Ranger should have skill choices")
	if len(ranger.ProficiencyChoices) > 0 {
		skillChoice := ranger.ProficiencyChoices[0]
		assert.Equal(t, 3, skillChoice.Count, "Ranger should choose 3 skills")
	}
}

// TestExpertiseImplementation checks if Rogue's expertise feature exists
func TestExpertiseImplementation(t *testing.T) {
	// This test will help us understand if expertise is handled as:
	// 1. A feature that modifies skill checks
	// 2. A special proficiency type
	// 3. Something else

	t.Log("Expertise is a 1st level Rogue feature that doubles proficiency bonus on chosen skills")
	t.Log("Need to investigate how this is implemented in the system")

	// TODO: Check if there's a feature system that handles expertise
	// TODO: Check if skill checks have a way to apply double proficiency
}
