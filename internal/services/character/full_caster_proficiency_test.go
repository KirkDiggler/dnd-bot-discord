package character

import (
	"net/http"
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWizardProficienciesFromAPI checks what proficiencies the Wizard gets from the D&D API
func TestWizardProficienciesFromAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping API test in short mode")
	}

	cfg := &dnd5e.Config{
		HttpClient: http.DefaultClient,
	}
	client, err := dnd5e.New(cfg)
	require.NoError(t, err)

	// Get wizard class
	wizard, err := client.GetClass("wizard")
	require.NoError(t, err)
	require.NotNil(t, wizard)

	// Log all proficiencies
	t.Logf("Wizard has %d automatic proficiencies:", len(wizard.Proficiencies))
	for _, prof := range wizard.Proficiencies {
		t.Logf("  - %s (%s)", prof.Name, prof.Key)
	}

	// Log proficiency choices
	t.Logf("\nWizard has %d proficiency choice groups:", len(wizard.ProficiencyChoices))
	for i, choice := range wizard.ProficiencyChoices {
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
		"saving-throw-int",
		"saving-throw-wis",
	}

	foundProfs := make(map[string]bool)
	for _, prof := range wizard.Proficiencies {
		foundProfs[prof.Key] = true
	}

	for _, expected := range expectedProfs {
		assert.True(t, foundProfs[expected], "Wizard should have %s proficiency", expected)
	}

	// Wizard should choose 2 skills
	assert.GreaterOrEqual(t, len(wizard.ProficiencyChoices), 1, "Wizard should have skill choices")
	if len(wizard.ProficiencyChoices) > 0 {
		skillChoice := wizard.ProficiencyChoices[0]
		assert.Equal(t, 2, skillChoice.Count, "Wizard should choose 2 skills")
	}
}

// TestSorcererProficienciesFromAPI checks Sorcerer proficiencies
func TestSorcererProficienciesFromAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping API test in short mode")
	}

	cfg := &dnd5e.Config{
		HttpClient: http.DefaultClient,
	}
	client, err := dnd5e.New(cfg)
	require.NoError(t, err)

	sorcerer, err := client.GetClass("sorcerer")
	require.NoError(t, err)
	require.NotNil(t, sorcerer)

	t.Logf("Sorcerer has %d automatic proficiencies:", len(sorcerer.Proficiencies))
	for _, prof := range sorcerer.Proficiencies {
		t.Logf("  - %s (%s)", prof.Name, prof.Key)
	}

	// Check expected proficiencies
	expectedProfs := []string{
		"saving-throw-con",
		"saving-throw-cha",
	}

	foundProfs := make(map[string]bool)
	for _, prof := range sorcerer.Proficiencies {
		foundProfs[prof.Key] = true
	}

	for _, expected := range expectedProfs {
		assert.True(t, foundProfs[expected], "Sorcerer should have %s proficiency", expected)
	}
}

// TestWarlockProficienciesFromAPI checks Warlock proficiencies
func TestWarlockProficienciesFromAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping API test in short mode")
	}

	cfg := &dnd5e.Config{
		HttpClient: http.DefaultClient,
	}
	client, err := dnd5e.New(cfg)
	require.NoError(t, err)

	warlock, err := client.GetClass("warlock")
	require.NoError(t, err)
	require.NotNil(t, warlock)

	t.Logf("Warlock has %d automatic proficiencies:", len(warlock.Proficiencies))
	for _, prof := range warlock.Proficiencies {
		t.Logf("  - %s (%s)", prof.Name, prof.Key)
	}

	// Check expected proficiencies
	expectedProfs := []string{
		"light-armor",
		"simple-weapons",
		"saving-throw-wis",
		"saving-throw-cha",
	}

	foundProfs := make(map[string]bool)
	for _, prof := range warlock.Proficiencies {
		foundProfs[prof.Key] = true
	}

	for _, expected := range expectedProfs {
		assert.True(t, foundProfs[expected], "Warlock should have %s proficiency", expected)
	}
}

// TestClericProficienciesFromAPI checks Cleric proficiencies
func TestClericProficienciesFromAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping API test in short mode")
	}

	cfg := &dnd5e.Config{
		HttpClient: http.DefaultClient,
	}
	client, err := dnd5e.New(cfg)
	require.NoError(t, err)

	cleric, err := client.GetClass("cleric")
	require.NoError(t, err)
	require.NotNil(t, cleric)

	t.Logf("Cleric has %d automatic proficiencies:", len(cleric.Proficiencies))
	for _, prof := range cleric.Proficiencies {
		t.Logf("  - %s (%s)", prof.Name, prof.Key)
	}

	// Check expected proficiencies
	expectedProfs := []string{
		"light-armor",
		"medium-armor",
		"shields",
		"simple-weapons",
		"saving-throw-wis",
		"saving-throw-cha",
	}

	foundProfs := make(map[string]bool)
	for _, prof := range cleric.Proficiencies {
		foundProfs[prof.Key] = true
	}

	for _, expected := range expectedProfs {
		assert.True(t, foundProfs[expected], "Cleric should have %s proficiency", expected)
	}
}

// TestDruidProficienciesFromAPI checks Druid proficiencies
func TestDruidProficienciesFromAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping API test in short mode")
	}

	cfg := &dnd5e.Config{
		HttpClient: http.DefaultClient,
	}
	client, err := dnd5e.New(cfg)
	require.NoError(t, err)

	druid, err := client.GetClass("druid")
	require.NoError(t, err)
	require.NotNil(t, druid)

	t.Logf("Druid has %d automatic proficiencies:", len(druid.Proficiencies))
	for _, prof := range druid.Proficiencies {
		t.Logf("  - %s (%s)", prof.Name, prof.Key)
	}

	// Log proficiency choices for tool proficiencies
	t.Logf("\nDruid has %d proficiency choice groups:", len(druid.ProficiencyChoices))
	for i, choice := range druid.ProficiencyChoices {
		t.Logf("\nChoice %d: %s (type: %s, count: %d)", i+1, choice.Name, choice.Type, choice.Count)
	}

	// Check expected proficiencies
	expectedProfs := []string{
		"light-armor",
		"medium-armor",
		"shields",
		"saving-throw-int",
		"saving-throw-wis",
	}

	foundProfs := make(map[string]bool)
	for _, prof := range druid.Proficiencies {
		foundProfs[prof.Key] = true
	}

	for _, expected := range expectedProfs {
		assert.True(t, foundProfs[expected], "Druid should have %s proficiency", expected)
	}

	// Check for specific weapon proficiencies
	expectedWeapons := []string{
		"clubs", "daggers", "darts", "javelins", "maces",
		"quarterstaffs", "scimitars", "sickles", "slings", "spears",
	}

	t.Log("\nChecking for specific Druid weapon proficiencies:")
	for _, weapon := range expectedWeapons {
		if foundProfs[weapon] {
			t.Logf("  ✓ Found %s", weapon)
		} else {
			t.Logf("  ✗ Missing %s", weapon)
		}
	}

	// Note: Druids don't get simple-weapons as a category, they get specific weapons
	assert.False(t, foundProfs["simple-weapons"], "Druid should NOT have simple-weapons category")
	assert.False(t, foundProfs["martial-weapons"], "Druid should NOT have martial-weapons category")
}
