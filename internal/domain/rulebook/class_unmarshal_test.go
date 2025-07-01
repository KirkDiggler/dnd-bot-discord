package rulebook

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClass_UnmarshalJSON_WithProficiencyChoices(t *testing.T) {
	// Test that a full class with proficiency choices unmarshals correctly
	classJSON := `{
		"key": "fighter",
		"name": "Fighter",
		"hit_die": 10,
		"proficiency_choices": [
			{
				"name": "Skill Proficiencies",
				"type": "proficiency",
				"key": "fighter-skills",
				"count": 2,
				"options": [
					{
						"reference": {
							"key": "skill-acrobatics",
							"name": "Acrobatics"
						}
					},
					{
						"reference": {
							"key": "skill-animal-handling",
							"name": "Animal Handling"
						}
					}
				]
			}
		],
		"proficiencies": [],
		"starting_equipment": [],
		"starting_equipment_choices": []
	}`

	var class Class
	err := json.Unmarshal([]byte(classJSON), &class)
	require.NoError(t, err)

	assert.Equal(t, "fighter", class.Key)
	assert.Equal(t, "Fighter", class.Name)
	assert.Equal(t, 10, class.HitDie)
	assert.Len(t, class.ProficiencyChoices, 1)

	// Verify the proficiency choice
	choice := class.ProficiencyChoices[0]
	assert.Equal(t, "Skill Proficiencies", choice.Name)
	assert.Equal(t, 2, choice.Count)
	assert.Len(t, choice.Options, 2)
}
