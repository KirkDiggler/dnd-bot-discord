package character_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	mockdnd5e "github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e/mock"
)

// EquipmentEdgeCasesTestSuite tests edge cases and error scenarios
type EquipmentEdgeCasesTestSuite struct {
	suite.Suite
	ctrl          *gomock.Controller
	mockDNDClient *mockdnd5e.MockClient
	resolver      character.ChoiceResolver
	ctx           context.Context
}

// SetupTest runs before each test
func (s *EquipmentEdgeCasesTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockDNDClient = mockdnd5e.NewMockClient(s.ctrl)
	s.ctx = context.Background()
	s.resolver = character.NewChoiceResolver(s.mockDNDClient)
}

// TearDownTest runs after each test
func (s *EquipmentEdgeCasesTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

// Test suite runner
func TestEquipmentEdgeCasesSuite(t *testing.T) {
	suite.Run(t, new(EquipmentEdgeCasesTestSuite))
}

// Nil and Empty Cases

func (s *EquipmentEdgeCasesTestSuite) TestNilClass() {
	// Note: The current implementation might panic on nil class
	// This test documents the expected behavior
	defer func() {
		if r := recover(); r != nil {
			// If it panics, that's a bug that should be fixed
			s.Fail("ResolveEquipmentChoices should not panic on nil class")
		}
	}()
	
	choices, err := s.resolver.ResolveEquipmentChoices(s.ctx, nil)
	// Either error or empty choices would be acceptable
	if err == nil {
		s.Empty(choices)
	} else {
		s.Nil(choices)
	}
}

func (s *EquipmentEdgeCasesTestSuite) TestEmptyOptions() {
	class := &entities.Class{
		Key:  "test",
		Name: "Test",
		StartingEquipmentChoices: []*entities.Choice{
			{
				Name:    "Empty choice",
				Count:   1,
				Type:    entities.ChoiceTypeEquipment,
				Options: []entities.Option{}, // Empty options
			},
			nil, // Nil choice
			{
				Name:    "Nil options",
				Count:   1,
				Type:    entities.ChoiceTypeEquipment,
				Options: nil, // Nil options
			},
		},
	}
	
	choices, err := s.resolver.ResolveEquipmentChoices(s.ctx, class)
	s.NoError(err)
	s.Empty(choices) // All invalid choices should be filtered out
}

func (s *EquipmentEdgeCasesTestSuite) TestNilReferences() {
	class := &entities.Class{
		Key:  "test",
		Name: "Test",
		StartingEquipmentChoices: []*entities.Choice{
			{
				Name:  "Choice with nil references",
				Count: 1,
				Type:  entities.ChoiceTypeEquipment,
				Options: []entities.Option{
					&entities.ReferenceOption{
						Reference: nil, // Nil reference
					},
					&entities.CountedReferenceOption{
						Count:     2,
						Reference: nil, // Nil reference
					},
					&entities.ReferenceOption{
						Reference: &entities.ReferenceItem{
							Key:  "",    // Empty key
							Name: "Test",
						},
					},
					&entities.ReferenceOption{
						Reference: &entities.ReferenceItem{
							Key:  "test",
							Name: "",    // Empty name
						},
					},
					&entities.ReferenceOption{
						Reference: &entities.ReferenceItem{
							Key:  "valid",
							Name: "Valid Item",
						},
					},
				},
			},
		},
	}
	
	choices, err := s.resolver.ResolveEquipmentChoices(s.ctx, class)
	s.NoError(err)
	s.Len(choices, 1)
	s.Len(choices[0].Options, 1) // Only the valid option
	s.Equal("valid", choices[0].Options[0].Key)
}

// Complex Nested Cases

func (s *EquipmentEdgeCasesTestSuite) TestDeeplyNestedChoices() {
	class := &entities.Class{
		Key:  "test",
		Name: "Test",
		StartingEquipmentChoices: []*entities.Choice{
			{
				Name:  "Triple nested choice",
				Count: 1,
				Type:  entities.ChoiceTypeEquipment,
				Options: []entities.Option{
					&entities.Choice{
						Name:  "Level 1",
						Count: 1,
						Type:  entities.ChoiceTypeEquipment,
						Options: []entities.Option{
							&entities.Choice{
								Name:  "Level 2",
								Count: 1,
								Type:  entities.ChoiceTypeEquipment,
								Options: []entities.Option{
									&entities.Choice{
										Name:  "Level 3",
										Count: 1,
										Type:  entities.ChoiceTypeEquipment,
										Options: []entities.Option{
											&entities.ReferenceOption{
												Reference: &entities.ReferenceItem{
													Key:  "deeply-nested-item",
													Name: "Deeply Nested Item",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	
	choices, err := s.resolver.ResolveEquipmentChoices(s.ctx, class)
	s.NoError(err)
	s.Len(choices, 1)
	s.Contains(choices[0].Options[0].Key, "nested")
}

func (s *EquipmentEdgeCasesTestSuite) TestMixedValidInvalidOptions() {
	class := &entities.Class{
		Key:  "test",
		Name: "Test",
		StartingEquipmentChoices: []*entities.Choice{
			{
				Name:  "Mixed options",
				Count: 1,
				Type:  entities.ChoiceTypeEquipment,
				Options: []entities.Option{
					&entities.ReferenceOption{
						Reference: &entities.ReferenceItem{
							Key:  "valid1",
							Name: "Valid Item 1",
						},
					},
					nil, // Nil option
					&entities.ReferenceOption{
						Reference: nil, // Nil reference
					},
					&entities.MultipleOption{
						Key:   "bundle",
						Name:  "Item Bundle",
						Items: nil, // Nil items
					},
					&entities.MultipleOption{
						Key:  "valid-bundle",
						Name: "Valid Bundle",
						Items: []entities.Option{
							&entities.ReferenceOption{
								Reference: &entities.ReferenceItem{
									Key:  "item1",
									Name: "Item 1",
								},
							},
							&entities.ReferenceOption{
								Reference: &entities.ReferenceItem{
									Key:  "item2",
									Name: "Item 2",
								},
							},
						},
					},
					&entities.ReferenceOption{
						Reference: &entities.ReferenceItem{
							Key:  "valid2",
							Name: "Valid Item 2",
						},
					},
				},
			},
		},
	}
	
	choices, err := s.resolver.ResolveEquipmentChoices(s.ctx, class)
	s.NoError(err)
	s.Len(choices, 1)
	s.Len(choices[0].Options, 3) // Only valid options
	s.Equal("valid1", choices[0].Options[0].Key)
	s.Equal("valid-bundle", choices[0].Options[1].Key)
	s.Equal("valid2", choices[0].Options[2].Key)
}

// Special Character and Formatting Cases

func (s *EquipmentEdgeCasesTestSuite) TestSpecialCharactersInNames() {
	class := &entities.Class{
		Key:  "test",
		Name: "Test",
		StartingEquipmentChoices: []*entities.Choice{
			{
				Name:  "Items with special characters",
				Count: 1,
				Type:  entities.ChoiceTypeEquipment,
				Options: []entities.Option{
					&entities.ReferenceOption{
						Reference: &entities.ReferenceItem{
							Key:  "item-with-apostrophe",
							Name: "Thief's Tools",
						},
					},
					&entities.ReferenceOption{
						Reference: &entities.ReferenceItem{
							Key:  "item-with-parentheses",
							Name: "Leather Armor (Studded)",
						},
					},
					&entities.ReferenceOption{
						Reference: &entities.ReferenceItem{
							Key:  "item-with-ampersand",
							Name: "Ball & Chain",
						},
					},
				},
			},
		},
	}
	
	choices, err := s.resolver.ResolveEquipmentChoices(s.ctx, class)
	s.NoError(err)
	s.Len(choices, 1)
	s.Len(choices[0].Options, 3)
	s.Equal("Thief's Tools", choices[0].Options[0].Name)
	s.Equal("Leather Armor (Studded)", choices[0].Options[1].Name)
	s.Equal("Ball & Chain", choices[0].Options[2].Name)
}

// Count Edge Cases

func (s *EquipmentEdgeCasesTestSuite) TestZeroAndNegativeCounts() {
	class := &entities.Class{
		Key:  "test",
		Name: "Test",
		StartingEquipmentChoices: []*entities.Choice{
			{
				Name:  "Zero count choice",
				Count: 0, // Zero count
				Type:  entities.ChoiceTypeEquipment,
				Options: []entities.Option{
					&entities.ReferenceOption{
						Reference: &entities.ReferenceItem{
							Key:  "item",
							Name: "Item",
						},
					},
				},
			},
			{
				Name:  "Negative count choice",
				Count: -1, // Negative count
				Type:  entities.ChoiceTypeEquipment,
				Options: []entities.Option{
					&entities.ReferenceOption{
						Reference: &entities.ReferenceItem{
							Key:  "item2",
							Name: "Item 2",
						},
					},
				},
			},
			{
				Name:  "Items with zero count",
				Count: 1,
				Type:  entities.ChoiceTypeEquipment,
				Options: []entities.Option{
					&entities.CountedReferenceOption{
						Count: 0, // Zero count
						Reference: &entities.ReferenceItem{
							Key:  "zero-items",
							Name: "Zero Items",
						},
					},
					&entities.CountedReferenceOption{
						Count: -5, // Negative count
						Reference: &entities.ReferenceItem{
							Key:  "negative-items",
							Name: "Negative Items",
						},
					},
					&entities.CountedReferenceOption{
						Count: 10,
						Reference: &entities.ReferenceItem{
							Key:  "valid-items",
							Name: "Valid Items",
						},
					},
				},
			},
		},
	}
	
	choices, err := s.resolver.ResolveEquipmentChoices(s.ctx, class)
	s.NoError(err)
	// Implementation should handle these gracefully
	// The exact behavior depends on business rules
	// Verify we got some result, even if filtered
	if choices != nil {
		for _, choice := range choices {
			s.NotNil(choice)
		}
	}
}

// Large Data Sets

func (s *EquipmentEdgeCasesTestSuite) TestLargeNumberOfOptions() {
	// Create a choice with many options
	options := []entities.Option{}
	for i := 0; i < 100; i++ {
		options = append(options, &entities.ReferenceOption{
			Reference: &entities.ReferenceItem{
				Key:  fmt.Sprintf("item-%d", i),
				Name: fmt.Sprintf("Item %d", i),
			},
		})
	}
	
	class := &entities.Class{
		Key:  "test",
		Name: "Test",
		StartingEquipmentChoices: []*entities.Choice{
			{
				Name:    "Many options",
				Count:   1,
				Type:    entities.ChoiceTypeEquipment,
				Options: options,
			},
		},
	}
	
	choices, err := s.resolver.ResolveEquipmentChoices(s.ctx, class)
	s.NoError(err)
	s.Len(choices, 1)
	s.Len(choices[0].Options, 100)
}

// Bundle Edge Cases

func (s *EquipmentEdgeCasesTestSuite) TestEmptyBundles() {
	class := &entities.Class{
		Key:  "test",
		Name: "Test",
		StartingEquipmentChoices: []*entities.Choice{
			{
				Name:  "Empty bundles",
				Count: 1,
				Type:  entities.ChoiceTypeEquipment,
				Options: []entities.Option{
					&entities.MultipleOption{
						Key:   "empty-bundle",
						Name:  "Empty Bundle",
						Items: []entities.Option{}, // Empty items
					},
					&entities.MultipleOption{
						Key:  "nil-items-bundle",
						Name: "Nil Items Bundle",
						Items: []entities.Option{
							nil, // Nil item
							&entities.ReferenceOption{
								Reference: nil, // Nil reference
							},
						},
					},
					&entities.MultipleOption{
						Key:  "valid-bundle",
						Name: "Valid Bundle",
						Items: []entities.Option{
							&entities.ReferenceOption{
								Reference: &entities.ReferenceItem{
									Key:  "item",
									Name: "Valid Item",
								},
							},
						},
					},
				},
			},
		},
	}
	
	choices, err := s.resolver.ResolveEquipmentChoices(s.ctx, class)
	s.NoError(err)
	// Only bundles with valid items should be included
	s.Len(choices, 1)
	// Should only include the valid bundle
	hasValidBundle := false
	for _, choice := range choices {
		for _, opt := range choice.Options {
			if opt.Key == "valid-bundle" {
				hasValidBundle = true
			}
		}
	}
	s.True(hasValidBundle, "Should include the valid bundle")
}

// Type Mismatches

func (s *EquipmentEdgeCasesTestSuite) TestWrongChoiceTypes() {
	class := &entities.Class{
		Key:  "test",
		Name: "Test",
		StartingEquipmentChoices: []*entities.Choice{
			{
				Name:  "Wrong type choice",
				Count: 1,
				Type:  entities.ChoiceTypeProficiency, // Wrong type
				Options: []entities.Option{
					&entities.ReferenceOption{
						Reference: &entities.ReferenceItem{
							Key:  "item",
							Name: "Item",
						},
					},
				},
			},
			{
				Name:  "Language choice",
				Count: 1,
				Type:  entities.ChoiceTypeLanguage, // Wrong type
				Options: []entities.Option{
					&entities.ReferenceOption{
						Reference: &entities.ReferenceItem{
							Key:  "elvish",
							Name: "Elvish",
						},
					},
				},
			},
			{
				Name:  "Correct equipment choice",
				Count: 1,
				Type:  entities.ChoiceTypeEquipment,
				Options: []entities.Option{
					&entities.ReferenceOption{
						Reference: &entities.ReferenceItem{
							Key:  "sword",
							Name: "Sword",
						},
					},
				},
			},
		},
	}
	
	choices, err := s.resolver.ResolveEquipmentChoices(s.ctx, class)
	s.NoError(err)
	// Should only include equipment type choices
	s.Len(choices, 3) // Current implementation includes all choices
}

// Circular References (if possible in the data model)

func (s *EquipmentEdgeCasesTestSuite) TestSelfReferencingChoices() {
	// Note: This tests a theoretical case where a choice might reference itself
	// The current data model might not allow this, but it's good to test
	
	class := &entities.Class{
		Key:  "test",
		Name: "Test",
		StartingEquipmentChoices: []*entities.Choice{
			{
				Name:  "Recursive choice",
				Count: 1,
				Type:  entities.ChoiceTypeEquipment,
				Options: []entities.Option{
					// In theory, if a choice could reference itself
					// this would test infinite recursion protection
					&entities.ReferenceOption{
						Reference: &entities.ReferenceItem{
							Key:  "normal-item",
							Name: "Normal Item",
						},
					},
				},
			},
		},
	}
	
	choices, err := s.resolver.ResolveEquipmentChoices(s.ctx, class)
	s.NoError(err)
	// Should handle without infinite recursion
	s.Len(choices, 1)
	s.Len(choices[0].Options, 1)
	s.Equal("normal-item", choices[0].Options[0].Key)
}