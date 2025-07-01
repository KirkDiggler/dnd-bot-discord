package character_test

import (
	"context"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"testing"

	mockdnd5e "github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e/mock"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

// EquipmentChoiceResolverTestSuite tests equipment choice resolution
type EquipmentChoiceResolverTestSuite struct {
	suite.Suite
	ctrl          *gomock.Controller
	mockDNDClient *mockdnd5e.MockClient
	resolver      character.ChoiceResolver
	ctx           context.Context
}

// SetupTest runs before each test
func (s *EquipmentChoiceResolverTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockDNDClient = mockdnd5e.NewMockClient(s.ctrl)
	s.ctx = context.Background()
	s.resolver = character.NewChoiceResolver(s.mockDNDClient)
}

// TearDownTest runs after each test
func (s *EquipmentChoiceResolverTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

// Test suite runner
func TestEquipmentChoiceResolverSuite(t *testing.T) {
	suite.Run(t, new(EquipmentChoiceResolverTestSuite))
}

// Fighter Equipment Tests

func (s *EquipmentChoiceResolverTestSuite) TestFighterEquipmentChoices_Complete() {
	// Setup - Fighter with all typical equipment choices
	class := &rulebook.Class{
		Key:  "fighter",
		Name: "Fighter",
		StartingEquipmentChoices: []*shared.Choice{
			// Choice 1: (a) chain mail or (b) leather armor, longbow, and 20 arrows
			{
				Name:  "(a) chain mail or (b) leather armor, longbow, and 20 arrows",
				Count: 1,
				Type:  shared.ChoiceTypeEquipment,
				Options: []shared.Option{
					&shared.ReferenceOption{
						Reference: &shared.ReferenceItem{
							Key:  "chain-mail",
							Name: "Chain Mail",
						},
					},
					&shared.MultipleOption{
						Key:  "armor-bow-bundle",
						Name: "leather armor, longbow, and 20 arrows",
						Items: []shared.Option{
							&shared.ReferenceOption{
								Reference: &shared.ReferenceItem{
									Key:  "leather-armor",
									Name: "Leather Armor",
								},
							},
							&shared.ReferenceOption{
								Reference: &shared.ReferenceItem{
									Key:  "longbow",
									Name: "Longbow",
								},
							},
							&shared.CountedReferenceOption{
								Count: 20,
								Reference: &shared.ReferenceItem{
									Key:  "arrow",
									Name: "Arrow",
								},
							},
						},
					},
				},
			},
			// Choice 2: (a) a martial weapon and a shield or (b) two martial weapons
			{
				Name:  "(a) a martial weapon and a shield or (b) two martial weapons",
				Count: 1,
				Type:  shared.ChoiceTypeEquipment,
				Options: []shared.Option{
					&shared.MultipleOption{
						Key:  "weapon-shield",
						Name: "a martial weapon and a shield",
						Items: []shared.Option{
							&shared.Choice{
								Name:  "martial weapon",
								Count: 1,
								Type:  shared.ChoiceTypeEquipment,
								Options: []shared.Option{
									&shared.ReferenceOption{
										Reference: &shared.ReferenceItem{
											Key:  "longsword",
											Name: "Longsword",
										},
									},
									&shared.ReferenceOption{
										Reference: &shared.ReferenceItem{
											Key:  "battleaxe",
											Name: "Battleaxe",
										},
									},
								},
							},
							&shared.ReferenceOption{
								Reference: &shared.ReferenceItem{
									Key:  "shield",
									Name: "Shield",
								},
							},
						},
					},
					&shared.Choice{
						Name:  "two martial weapons",
						Count: 2,
						Type:  shared.ChoiceTypeEquipment,
						Options: []shared.Option{
							&shared.ReferenceOption{
								Reference: &shared.ReferenceItem{
									Key:  "longsword",
									Name: "Longsword",
								},
							},
							&shared.ReferenceOption{
								Reference: &shared.ReferenceItem{
									Key:  "shortsword",
									Name: "Shortsword",
								},
							},
							&shared.ReferenceOption{
								Reference: &shared.ReferenceItem{
									Key:  "rapier",
									Name: "Rapier",
								},
							},
						},
					},
				},
			},
			// Choice 3: (a) a light crossbow and 20 bolts or (b) two handaxes
			{
				Name:  "(a) a light crossbow and 20 bolts or (b) two handaxes",
				Count: 1,
				Type:  shared.ChoiceTypeEquipment,
				Options: []shared.Option{
					&shared.MultipleOption{
						Key:  "crossbow-bundle",
						Name: "a light crossbow and 20 bolts",
						Items: []shared.Option{
							&shared.ReferenceOption{
								Reference: &shared.ReferenceItem{
									Key:  "light-crossbow",
									Name: "Light Crossbow",
								},
							},
							&shared.CountedReferenceOption{
								Count: 20,
								Reference: &shared.ReferenceItem{
									Key:  "crossbow-bolt",
									Name: "Crossbow Bolt",
								},
							},
						},
					},
					&shared.CountedReferenceOption{
						Count: 2,
						Reference: &shared.ReferenceItem{
							Key:  "handaxe",
							Name: "Handaxe",
						},
					},
				},
			},
			// Choice 4: (a) a dungeoneer's pack or (b) an explorer's pack
			{
				Name:  "(a) a dungeoneer's pack or (b) an explorer's pack",
				Count: 1,
				Type:  shared.ChoiceTypeEquipment,
				Options: []shared.Option{
					&shared.ReferenceOption{
						Reference: &shared.ReferenceItem{
							Key:  "dungeoneers-pack",
							Name: "Dungeoneer's Pack",
						},
					},
					&shared.ReferenceOption{
						Reference: &shared.ReferenceItem{
							Key:  "explorers-pack",
							Name: "Explorer's Pack",
						},
					},
				},
			},
		},
	}

	// Execute
	choices, err := s.resolver.ResolveEquipmentChoices(s.ctx, class)

	// Assert
	s.NoError(err)
	s.Len(choices, 4)

	// Choice 1 - Armor choice
	s.Equal("fighter-equip-0", choices[0].ID)
	s.Contains(choices[0].Name, "chain mail")
	s.Equal("equipment", choices[0].Type)
	s.Equal(1, choices[0].Choose)
	s.Len(choices[0].Options, 2)
	s.Equal("chain-mail", choices[0].Options[0].Key)
	s.Equal("Chain Mail", choices[0].Options[0].Name)
	s.Contains(choices[0].Options[0].Description, "16 AC") // Should have armor description
	s.Equal("armor-bow-bundle", choices[0].Options[1].Key)
	s.Contains(choices[0].Options[1].Name, "Leather Armor")
	s.Contains(choices[0].Options[1].Name, "Longbow")
	s.Contains(choices[0].Options[1].Name, "20x Arrow")

	// Choice 2 - Weapon choice with nested selection
	s.Equal("fighter-equip-1", choices[1].ID)
	s.Len(choices[1].Options, 2)
	// First option should be marked as nested (triggers weapon selection)
	s.Contains(choices[1].Options[0].Key, "nested")
	s.Contains(choices[1].Options[0].Name, "martial weapon")
	s.Contains(choices[1].Options[0].Description, "Choose")
	// Second option should also be nested (choose 2 weapons)
	s.Contains(choices[1].Options[1].Key, "nested")
	s.Contains(choices[1].Options[1].Name, "two martial weapons")

	// Choice 3 - Ranged weapon choice
	s.Equal("fighter-equip-2", choices[2].ID)
	s.Len(choices[2].Options, 2)
	s.Contains(choices[2].Options[0].Name, "Light Crossbow")
	s.Contains(choices[2].Options[0].Name, "20x Crossbow Bolt")
	s.Equal("handaxe", choices[2].Options[1].Key)
	s.Equal("2x Handaxe", choices[2].Options[1].Name)

	// Choice 4 - Pack choice
	s.Equal("fighter-equip-3", choices[3].ID)
	s.Len(choices[3].Options, 2)
	s.Equal("dungeoneers-pack", choices[3].Options[0].Key)
	s.Equal("explorers-pack", choices[3].Options[1].Key)
}

// Monk Equipment Tests

func (s *EquipmentChoiceResolverTestSuite) TestMonkEquipmentChoices_SimpleWeapons() {
	// Setup - Monk with simple weapon choices
	class := &rulebook.Class{
		Key:  "monk",
		Name: "Monk",
		StartingEquipmentChoices: []*shared.Choice{
			// Choice 1: (a) a shortsword or (b) any simple weapon
			{
				Name:  "(a) a shortsword or (b) any simple weapon",
				Count: 1,
				Type:  shared.ChoiceTypeEquipment,
				Options: []shared.Option{
					&shared.ReferenceOption{
						Reference: &shared.ReferenceItem{
							Key:  "shortsword",
							Name: "Shortsword",
						},
					},
					&shared.Choice{
						Name:  "any simple weapon",
						Count: 1,
						Type:  shared.ChoiceTypeEquipment,
						Options: []shared.Option{
							&shared.ReferenceOption{
								Reference: &shared.ReferenceItem{
									Key:  "club",
									Name: "Club",
								},
							},
							&shared.ReferenceOption{
								Reference: &shared.ReferenceItem{
									Key:  "quarterstaff",
									Name: "Quarterstaff",
								},
							},
							&shared.ReferenceOption{
								Reference: &shared.ReferenceItem{
									Key:  "dagger",
									Name: "Dagger",
								},
							},
						},
					},
				},
			},
			// Choice 2: (a) a dungeoneer's pack or (b) an explorer's pack
			{
				Name:  "(a) a dungeoneer's pack or (b) an explorer's pack",
				Count: 1,
				Type:  shared.ChoiceTypeEquipment,
				Options: []shared.Option{
					&shared.ReferenceOption{
						Reference: &shared.ReferenceItem{
							Key:  "dungeoneers-pack",
							Name: "Dungeoneer's Pack",
						},
					},
					&shared.ReferenceOption{
						Reference: &shared.ReferenceItem{
							Key:  "explorers-pack",
							Name: "Explorer's Pack",
						},
					},
				},
			},
		},
	}

	// Execute
	choices, err := s.resolver.ResolveEquipmentChoices(s.ctx, class)

	// Assert
	s.NoError(err)
	s.Len(choices, 2)

	// Choice 1 - Weapon choice
	s.Equal("monk-equip-0", choices[0].ID)
	s.Len(choices[0].Options, 2)
	s.Equal("shortsword", choices[0].Options[0].Key)
	s.Equal("Shortsword", choices[0].Options[0].Name)
	s.Contains(choices[0].Options[0].Description, "1d6 piercing") // Weapon description
	// Second option should be marked as nested choice
	s.Contains(choices[0].Options[1].Key, "nested")
	s.Equal("any simple weapon", choices[0].Options[1].Name)
	s.Contains(choices[0].Options[1].Description, "Choose")

	// Choice 2 - Pack choice
	s.Equal("monk-equip-1", choices[1].ID)
	s.Len(choices[1].Options, 2)
}

// Wizard Equipment Tests

func (s *EquipmentChoiceResolverTestSuite) TestWizardEquipmentChoices_WithFocus() {
	// Setup - Wizard with component pouch/focus choice
	class := &rulebook.Class{
		Key:  "wizard",
		Name: "Wizard",
		StartingEquipmentChoices: []*shared.Choice{
			// Choice 1: (a) a quarterstaff or (b) a dagger
			{
				Name:  "(a) a quarterstaff or (b) a dagger",
				Count: 1,
				Type:  shared.ChoiceTypeEquipment,
				Options: []shared.Option{
					&shared.ReferenceOption{
						Reference: &shared.ReferenceItem{
							Key:  "quarterstaff",
							Name: "Quarterstaff",
						},
					},
					&shared.ReferenceOption{
						Reference: &shared.ReferenceItem{
							Key:  "dagger",
							Name: "Dagger",
						},
					},
				},
			},
			// Choice 2: (a) a component pouch or (b) an arcane focus
			{
				Name:  "(a) a component pouch or (b) an arcane focus",
				Count: 1,
				Type:  shared.ChoiceTypeEquipment,
				Options: []shared.Option{
					&shared.ReferenceOption{
						Reference: &shared.ReferenceItem{
							Key:  "component-pouch",
							Name: "Component Pouch",
						},
					},
					&shared.Choice{
						Name:  "arcane focus",
						Count: 1,
						Type:  shared.ChoiceTypeEquipment,
						Options: []shared.Option{
							&shared.ReferenceOption{
								Reference: &shared.ReferenceItem{
									Key:  "crystal",
									Name: "Crystal",
								},
							},
							&shared.ReferenceOption{
								Reference: &shared.ReferenceItem{
									Key:  "orb",
									Name: "Orb",
								},
							},
							&shared.ReferenceOption{
								Reference: &shared.ReferenceItem{
									Key:  "rod",
									Name: "Rod",
								},
							},
							&shared.ReferenceOption{
								Reference: &shared.ReferenceItem{
									Key:  "staff",
									Name: "Staff",
								},
							},
							&shared.ReferenceOption{
								Reference: &shared.ReferenceItem{
									Key:  "wand",
									Name: "Wand",
								},
							},
						},
					},
				},
			},
			// Choice 3: (a) a scholar's pack or (b) an explorer's pack
			{
				Name:  "(a) a scholar's pack or (b) an explorer's pack",
				Count: 1,
				Type:  shared.ChoiceTypeEquipment,
				Options: []shared.Option{
					&shared.ReferenceOption{
						Reference: &shared.ReferenceItem{
							Key:  "scholars-pack",
							Name: "Scholar's Pack",
						},
					},
					&shared.ReferenceOption{
						Reference: &shared.ReferenceItem{
							Key:  "explorers-pack",
							Name: "Explorer's Pack",
						},
					},
				},
			},
		},
	}

	// Execute
	choices, err := s.resolver.ResolveEquipmentChoices(s.ctx, class)

	// Assert
	s.NoError(err)
	s.Len(choices, 3)

	// Choice 2 - Focus choice with nested selection
	s.Equal("wizard-equip-1", choices[1].ID)
	s.Len(choices[1].Options, 2)
	s.Equal("component-pouch", choices[1].Options[0].Key)
	// Second option should trigger focus selection
	s.Contains(choices[1].Options[1].Key, "nested")
	s.Equal("arcane focus", choices[1].Options[1].Name)
	s.Contains(choices[1].Options[1].Description, "Choose 1")
}

// Rogue Equipment Tests

func (s *EquipmentChoiceResolverTestSuite) TestRogueEquipmentChoices_WithBundles() {
	// Setup - Rogue with weapon bundles
	class := &rulebook.Class{
		Key:  "rogue",
		Name: "Rogue",
		StartingEquipmentChoices: []*shared.Choice{
			// Choice 1: (a) a rapier or (b) a shortsword
			{
				Name:  "(a) a rapier or (b) a shortsword",
				Count: 1,
				Type:  shared.ChoiceTypeEquipment,
				Options: []shared.Option{
					&shared.ReferenceOption{
						Reference: &shared.ReferenceItem{
							Key:  "rapier",
							Name: "Rapier",
						},
					},
					&shared.ReferenceOption{
						Reference: &shared.ReferenceItem{
							Key:  "shortsword",
							Name: "Shortsword",
						},
					},
				},
			},
			// Choice 2: (a) a shortbow and 20 arrows or (b) a shortsword
			{
				Name:  "(a) a shortbow and 20 arrows or (b) a shortsword",
				Count: 1,
				Type:  shared.ChoiceTypeEquipment,
				Options: []shared.Option{
					&shared.MultipleOption{
						Key:  "bow-bundle",
						Name: "a shortbow and 20 arrows",
						Items: []shared.Option{
							&shared.ReferenceOption{
								Reference: &shared.ReferenceItem{
									Key:  "shortbow",
									Name: "Shortbow",
								},
							},
							&shared.CountedReferenceOption{
								Count: 20,
								Reference: &shared.ReferenceItem{
									Key:  "arrow",
									Name: "Arrow",
								},
							},
						},
					},
					&shared.ReferenceOption{
						Reference: &shared.ReferenceItem{
							Key:  "shortsword",
							Name: "Shortsword",
						},
					},
				},
			},
		},
	}

	// Execute
	choices, err := s.resolver.ResolveEquipmentChoices(s.ctx, class)

	// Assert
	s.NoError(err)
	s.Len(choices, 2)

	// Choice 2 - Bundle choice
	s.Equal("rogue-equip-1", choices[1].ID)
	s.Len(choices[1].Options, 2)
	s.Equal("bow-bundle", choices[1].Options[0].Key)
	s.Contains(choices[1].Options[0].Name, "Shortbow and 20x Arrow")
	s.Equal("shortsword", choices[1].Options[1].Key)
}

// Edge Cases and Error Handling

func (s *EquipmentChoiceResolverTestSuite) TestResolveEquipmentChoices_EmptyChoices() {
	// Setup - Class with no equipment choices
	class := &rulebook.Class{
		Key:                      "custom",
		Name:                     "Custom Class",
		StartingEquipmentChoices: []*shared.Choice{},
	}

	// Execute
	choices, err := s.resolver.ResolveEquipmentChoices(s.ctx, class)

	// Assert
	s.NoError(err)
	s.Empty(choices)
}

func (s *EquipmentChoiceResolverTestSuite) TestResolveEquipmentChoices_NilChoices() {
	// Setup - Class with nil equipment choices
	class := &rulebook.Class{
		Key:                      "custom",
		Name:                     "Custom Class",
		StartingEquipmentChoices: nil,
	}

	// Execute
	choices, err := s.resolver.ResolveEquipmentChoices(s.ctx, class)

	// Assert
	s.NoError(err)
	s.Empty(choices)
}

func (s *EquipmentChoiceResolverTestSuite) TestResolveEquipmentChoices_ChoicesWithNilOptions() {
	// Setup - Choices with nil or empty options
	class := &rulebook.Class{
		Key:  "custom",
		Name: "Custom Class",
		StartingEquipmentChoices: []*shared.Choice{
			{
				Name:    "Empty choice",
				Count:   1,
				Type:    shared.ChoiceTypeEquipment,
				Options: nil, // Nil options
			},
			{
				Name:    "No options",
				Count:   1,
				Type:    shared.ChoiceTypeEquipment,
				Options: []shared.Option{}, // Empty options
			},
			{
				Name:  "Valid choice",
				Count: 1,
				Type:  shared.ChoiceTypeEquipment,
				Options: []shared.Option{
					&shared.ReferenceOption{
						Reference: &shared.ReferenceItem{
							Key:  "dagger",
							Name: "Dagger",
						},
					},
				},
			},
		},
	}

	// Execute
	choices, err := s.resolver.ResolveEquipmentChoices(s.ctx, class)

	// Assert
	s.NoError(err)
	s.Len(choices, 1) // Only the valid choice should be included
	s.Equal("custom-equip-2", choices[0].ID)
	s.Equal("Valid choice", choices[0].Name)
}

func (s *EquipmentChoiceResolverTestSuite) TestResolveEquipmentChoices_InvalidOptionTypes() {
	// Setup - Choice with invalid reference
	class := &rulebook.Class{
		Key:  "custom",
		Name: "Custom Class",
		StartingEquipmentChoices: []*shared.Choice{
			{
				Name:  "Choice with nil reference",
				Count: 1,
				Type:  shared.ChoiceTypeEquipment,
				Options: []shared.Option{
					&shared.ReferenceOption{
						Reference: nil, // Nil reference
					},
					&shared.ReferenceOption{
						Reference: &shared.ReferenceItem{
							Key:  "valid-item",
							Name: "Valid Item",
						},
					},
				},
			},
		},
	}

	// Execute
	choices, err := s.resolver.ResolveEquipmentChoices(s.ctx, class)

	// Assert
	s.NoError(err)
	s.Len(choices, 1)
	s.Len(choices[0].Options, 1) // Only valid option should be included
	s.Equal("valid-item", choices[0].Options[0].Key)
}

// Complex Nested Choice Tests

func (s *EquipmentChoiceResolverTestSuite) TestResolveEquipmentChoices_DeepNestedChoices() {
	// Setup - Complex nested martial weapon choice
	class := &rulebook.Class{
		Key:  "fighter",
		Name: "Fighter",
		StartingEquipmentChoices: []*shared.Choice{
			{
				Name:  "Weapon choice with deep nesting",
				Count: 1,
				Type:  shared.ChoiceTypeEquipment,
				Options: []shared.Option{
					&shared.MultipleOption{
						Key:  "complex-bundle",
						Name: "weapon bundle",
						Items: []shared.Option{
							&shared.Choice{
								Name:  "a martial melee weapon",
								Count: 1,
								Type:  shared.ChoiceTypeEquipment,
								Options: []shared.Option{
									&shared.ReferenceOption{
										Reference: &shared.ReferenceItem{
											Key:  "longsword",
											Name: "Longsword",
										},
									},
									&shared.ReferenceOption{
										Reference: &shared.ReferenceItem{
											Key:  "battleaxe",
											Name: "Battleaxe",
										},
									},
								},
							},
							&shared.Choice{
								Name:  "a simple ranged weapon",
								Count: 1,
								Type:  shared.ChoiceTypeEquipment,
								Options: []shared.Option{
									&shared.ReferenceOption{
										Reference: &shared.ReferenceItem{
											Key:  "shortbow",
											Name: "Shortbow",
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

	// Execute
	choices, err := s.resolver.ResolveEquipmentChoices(s.ctx, class)

	// Assert
	s.NoError(err)
	s.Len(choices, 1)
	// The bundle with nested choices should be marked appropriately
	s.Contains(choices[0].Options[0].Key, "nested")
	s.Contains(choices[0].Options[0].Description, "Choose")
}

// Cleric Equipment Tests

func (s *EquipmentChoiceResolverTestSuite) TestClericEquipmentChoices_WithHolySymbol() {
	// Setup - Cleric with holy symbol choice
	class := &rulebook.Class{
		Key:  "cleric",
		Name: "Cleric",
		StartingEquipmentChoices: []*shared.Choice{
			// Choice 1: (a) a mace or (b) a warhammer (if proficient)
			{
				Name:  "(a) a mace or (b) a warhammer (if proficient)",
				Count: 1,
				Type:  shared.ChoiceTypeEquipment,
				Options: []shared.Option{
					&shared.ReferenceOption{
						Reference: &shared.ReferenceItem{
							Key:  "mace",
							Name: "Mace",
						},
					},
					&shared.ReferenceOption{
						Reference: &shared.ReferenceItem{
							Key:  "warhammer",
							Name: "Warhammer",
						},
					},
				},
			},
			// Choice 2: (a) scale mail, (b) leather armor, or (c) chain mail (if proficient)
			{
				Name:  "(a) scale mail, (b) leather armor, or (c) chain mail (if proficient)",
				Count: 1,
				Type:  shared.ChoiceTypeEquipment,
				Options: []shared.Option{
					&shared.ReferenceOption{
						Reference: &shared.ReferenceItem{
							Key:  "scale-mail",
							Name: "Scale Mail",
						},
					},
					&shared.ReferenceOption{
						Reference: &shared.ReferenceItem{
							Key:  "leather-armor",
							Name: "Leather Armor",
						},
					},
					&shared.ReferenceOption{
						Reference: &shared.ReferenceItem{
							Key:  "chain-mail",
							Name: "Chain Mail",
						},
					},
				},
			},
			// Choice 3: holy symbol choice
			{
				Name:  "a holy symbol",
				Count: 1,
				Type:  shared.ChoiceTypeEquipment,
				Options: []shared.Option{
					&shared.Choice{
						Name:  "holy symbol",
						Count: 1,
						Type:  shared.ChoiceTypeEquipment,
						Options: []shared.Option{
							&shared.ReferenceOption{
								Reference: &shared.ReferenceItem{
									Key:  "amulet",
									Name: "Amulet",
								},
							},
							&shared.ReferenceOption{
								Reference: &shared.ReferenceItem{
									Key:  "emblem",
									Name: "Emblem",
								},
							},
							&shared.ReferenceOption{
								Reference: &shared.ReferenceItem{
									Key:  "reliquary",
									Name: "Reliquary",
								},
							},
						},
					},
				},
			},
		},
	}

	// Execute
	choices, err := s.resolver.ResolveEquipmentChoices(s.ctx, class)

	// Assert
	s.NoError(err)
	s.Len(choices, 3)

	// Choice 1 - Weapon
	s.Equal("cleric-equip-0", choices[0].ID)
	s.Len(choices[0].Options, 2)
	s.Equal("mace", choices[0].Options[0].Key)
	s.Contains(choices[0].Options[0].Description, "1d6 bludgeoning")

	// Choice 2 - Armor
	s.Equal("cleric-equip-1", choices[1].ID)
	s.Len(choices[1].Options, 3)
	s.Contains(choices[1].Options[0].Description, "14 + Dex")
	s.Contains(choices[1].Options[1].Description, "11 + Dex")
	s.Contains(choices[1].Options[2].Description, "16 AC")

	// Choice 3 - Holy symbol (nested)
	s.Equal("cleric-equip-2", choices[2].ID)
	s.Len(choices[2].Options, 1)
	s.Contains(choices[2].Options[0].Key, "nested")
	s.Equal("holy symbol", choices[2].Options[0].Name)
}

// Ranger Equipment Tests

func (s *EquipmentChoiceResolverTestSuite) TestRangerEquipmentChoices_WithMultipleOptions() {
	// Setup - Ranger with varied equipment choices
	class := &rulebook.Class{
		Key:  "ranger",
		Name: "Ranger",
		StartingEquipmentChoices: []*shared.Choice{
			// Choice 1: (a) scale mail or (b) leather armor
			{
				Name:  "(a) scale mail or (b) leather armor",
				Count: 1,
				Type:  shared.ChoiceTypeEquipment,
				Options: []shared.Option{
					&shared.ReferenceOption{
						Reference: &shared.ReferenceItem{
							Key:  "scale-mail",
							Name: "Scale Mail",
						},
					},
					&shared.ReferenceOption{
						Reference: &shared.ReferenceItem{
							Key:  "leather-armor",
							Name: "Leather Armor",
						},
					},
				},
			},
			// Choice 2: (a) two shortswords or (b) two simple melee weapons
			{
				Name:  "(a) two shortswords or (b) two simple melee weapons",
				Count: 1,
				Type:  shared.ChoiceTypeEquipment,
				Options: []shared.Option{
					&shared.CountedReferenceOption{
						Count: 2,
						Reference: &shared.ReferenceItem{
							Key:  "shortsword",
							Name: "Shortsword",
						},
					},
					&shared.Choice{
						Name:  "two simple melee weapons",
						Count: 2,
						Type:  shared.ChoiceTypeEquipment,
						Options: []shared.Option{
							&shared.ReferenceOption{
								Reference: &shared.ReferenceItem{
									Key:  "handaxe",
									Name: "Handaxe",
								},
							},
							&shared.ReferenceOption{
								Reference: &shared.ReferenceItem{
									Key:  "club",
									Name: "Club",
								},
							},
							&shared.ReferenceOption{
								Reference: &shared.ReferenceItem{
									Key:  "dagger",
									Name: "Dagger",
								},
							},
						},
					},
				},
			},
		},
	}

	// Execute
	choices, err := s.resolver.ResolveEquipmentChoices(s.ctx, class)

	// Assert
	s.NoError(err)
	s.Len(choices, 2)

	// Choice 2 - Weapon choice
	s.Equal("ranger-equip-1", choices[1].ID)
	s.Len(choices[1].Options, 2)
	s.Equal("shortsword", choices[1].Options[0].Key)
	s.Equal("2x Shortsword", choices[1].Options[0].Name)
	// Second option should be nested (choose 2 simple weapons)
	s.Contains(choices[1].Options[1].Key, "nested")
	s.Contains(choices[1].Options[1].Description, "Choose 2")
}
