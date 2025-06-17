package entities

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChoice_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		validate func(t *testing.T, c *Choice)
		wantErr  bool
	}{
		{
			name: "simple choice with reference options",
			json: `{
				"name": "Skill Proficiencies",
				"type": "proficiency",
				"key": "skill-prof",
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
							"key": "skill-athletics",
							"name": "Athletics"
						}
					}
				]
			}`,
			validate: func(t *testing.T, c *Choice) {
				assert.Equal(t, "Skill Proficiencies", c.Name)
				assert.Equal(t, ChoiceTypeProficiency, c.Type)
				assert.Equal(t, 2, c.Count)
				assert.Len(t, c.Options, 2)

				// Check first option
				opt1, ok := c.Options[0].(*ReferenceOption)
				require.True(t, ok, "First option should be ReferenceOption")
				assert.Equal(t, "skill-acrobatics", opt1.GetKey())
				assert.Equal(t, "Acrobatics", opt1.GetName())
			},
		},
		{
			name: "choice with counted reference options",
			json: `{
				"name": "Starting Equipment",
				"type": "equipment",
				"key": "starting-equip",
				"count": 1,
				"options": [
					{
						"count": 1,
						"reference": {
							"key": "longsword",
							"name": "Longsword"
						}
					},
					{
						"count": 2,
						"reference": {
							"key": "shortsword",
							"name": "Shortsword"
						}
					}
				]
			}`,
			validate: func(t *testing.T, c *Choice) {
				assert.Equal(t, "Starting Equipment", c.Name)
				assert.Equal(t, ChoiceTypeEquipment, c.Type)
				assert.Len(t, c.Options, 2)

				// Check counted reference
				opt1, ok := c.Options[0].(*CountedReferenceOption)
				require.True(t, ok, "First option should be CountedReferenceOption")
				assert.Equal(t, 1, opt1.Count)
				assert.Equal(t, "longsword", opt1.GetKey())
			},
		},
		{
			name: "choice with empty options array",
			json: `{
				"name": "Empty Choice",
				"type": "proficiency",
				"key": "empty",
				"count": 0,
				"options": []
			}`,
			validate: func(t *testing.T, c *Choice) {
				assert.Equal(t, "Empty Choice", c.Name)
				assert.Len(t, c.Options, 0)
			},
		},
		{
			name: "choice with options as object (should be ignored)",
			json: `{
				"name": "Object Options",
				"type": "proficiency",
				"key": "object-opts",
				"count": 1,
				"options": {
					"some": "object"
				}
			}`,
			validate: func(t *testing.T, c *Choice) {
				assert.Equal(t, "Object Options", c.Name)
				assert.Len(t, c.Options, 0) // Should be empty as we ignore object format
			},
		},
		{
			name: "nested choice options",
			json: `{
				"name": "Complex Choice",
				"type": "proficiency",
				"key": "complex",
				"count": 1,
				"options": [
					{
						"name": "Martial Weapons",
						"type": "proficiency",
						"key": "martial-weapons",
						"count": 2,
						"options": []
					}
				]
			}`,
			validate: func(t *testing.T, c *Choice) {
				assert.Equal(t, "Complex Choice", c.Name)
				assert.Len(t, c.Options, 1)

				// Check nested choice
				opt1, ok := c.Options[0].(*Choice)
				require.True(t, ok, "First option should be Choice")
				assert.Equal(t, "Martial Weapons", opt1.Name)
				assert.Equal(t, 2, opt1.Count)
			},
		},
		{
			name: "multiple option type",
			json: `{
				"name": "Multiple Choice",
				"type": "equipment",
				"key": "multiple",
				"count": 1,
				"options": [
					{
						"key": "explorers-pack",
						"name": "Explorer's Pack",
						"items": [
							{
								"reference": {
									"key": "bedroll",
									"name": "Bedroll"
								}
							},
							{
								"reference": {
									"key": "mess-kit",
									"name": "Mess Kit"
								}
							}
						]
					}
				]
			}`,
			validate: func(t *testing.T, c *Choice) {
				assert.Equal(t, "Multiple Choice", c.Name)
				assert.Len(t, c.Options, 1)

				// Check multiple option
				opt1, ok := c.Options[0].(*MultipleOption)
				require.True(t, ok, "First option should be MultipleOption")
				assert.Equal(t, "explorers-pack", opt1.GetKey())
				assert.Len(t, opt1.Items, 2)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var c Choice
			err := json.Unmarshal([]byte(tt.json), &c)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			tt.validate(t, &c)
		})
	}
}

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

func TestMultipleOption_UnmarshalJSON(t *testing.T) {
	multiJSON := `{
		"key": "pack",
		"name": "Equipment Pack",
		"items": [
			{
				"reference": {
					"key": "rope",
					"name": "Rope, hempen (50 feet)"
				}
			},
			{
				"count": 10,
				"reference": {
					"key": "piton",
					"name": "Piton"
				}
			}
		]
	}`

	var multi MultipleOption
	err := json.Unmarshal([]byte(multiJSON), &multi)
	require.NoError(t, err)

	assert.Equal(t, "pack", multi.Key)
	assert.Equal(t, "Equipment Pack", multi.Name)
	assert.Len(t, multi.Items, 2)

	// Check first item (reference)
	item1, ok := multi.Items[0].(*ReferenceOption)
	require.True(t, ok, "First item should be ReferenceOption")
	assert.Equal(t, "rope", item1.GetKey())

	// Check second item (counted reference)
	item2, ok := multi.Items[1].(*CountedReferenceOption)
	require.True(t, ok, "Second item should be CountedReferenceOption")
	assert.Equal(t, "piton", item2.GetKey())
	assert.Equal(t, 10, item2.Count)
}
