package character_test

import (
	"context"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	mockdnd5e "github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e/mock"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
)

// EquipmentIntegrationTestSuite tests the full equipment choice flow
type EquipmentIntegrationTestSuite struct {
	suite.Suite
	ctrl          *gomock.Controller
	mockDNDClient *mockdnd5e.MockClient
	resolver      character.ChoiceResolver
	ctx           context.Context
}

// SetupTest runs before each test
func (s *EquipmentIntegrationTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockDNDClient = mockdnd5e.NewMockClient(s.ctrl)
	s.ctx = context.Background()
	s.resolver = character.NewChoiceResolver(s.mockDNDClient)
}

// TearDownTest runs after each test
func (s *EquipmentIntegrationTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

// Test suite runner
func TestEquipmentIntegrationSuite(t *testing.T) {
	suite.Run(t, new(EquipmentIntegrationTestSuite))
}

// Full Flow Tests

func (s *EquipmentIntegrationTestSuite) TestFullEquipmentSelectionFlow_Fighter() {
	// This test simulates the complete flow of equipment selection for a Fighter

	// Step 1: Get equipment choices
	fighterClass := createFullFighterClass()
	choices, err := s.resolver.ResolveEquipmentChoices(s.ctx, fighterClass)
	s.NoError(err)
	s.Len(choices, 4)

	// Step 2: Verify first choice (armor)
	armorChoice := choices[0]
	s.Equal("fighter-equip-0", armorChoice.ID)
	s.Len(armorChoice.Options, 2)

	// Step 3: Simulate selecting chain mail
	selectedArmor := armorChoice.Options[0]
	s.Equal("chain-mail", selectedArmor.Key)
	s.Equal("Chain Mail", selectedArmor.Name)
	s.Contains(selectedArmor.Description, "16 AC")

	// Step 4: Verify second choice (weapons) has nested options
	weaponChoice := choices[1]
	s.Equal("fighter-equip-1", weaponChoice.ID)
	s.Len(weaponChoice.Options, 2)

	// Option 1 should be nested (martial weapon + shield)
	weaponShieldOption := weaponChoice.Options[0]
	s.Contains(weaponShieldOption.Key, "nested")
	s.Contains(weaponShieldOption.Description, "Choose")

	// Option 2 should also be nested (two martial weapons)
	twoWeaponsOption := weaponChoice.Options[1]
	s.Contains(twoWeaponsOption.Key, "nested")
	s.Contains(twoWeaponsOption.Description, "Choose 2")

	// Step 5: Verify the nested choice would trigger weapon selection UI
	s.NotEmpty(weaponShieldOption.Description)
	s.NotEmpty(twoWeaponsOption.Description)
}

func (s *EquipmentIntegrationTestSuite) TestFullEquipmentSelectionFlow_Wizard() {
	// This test simulates the complete flow for a Wizard with focus selection

	// Step 1: Get equipment choices
	wizardClass := createFullWizardClass()
	choices, err := s.resolver.ResolveEquipmentChoices(s.ctx, wizardClass)
	s.NoError(err)
	s.Len(choices, 3)

	// Step 2: Verify weapon choice
	weaponChoice := choices[0]
	s.Equal("wizard-equip-0", weaponChoice.ID)
	s.Len(weaponChoice.Options, 2)
	s.Equal("quarterstaff", weaponChoice.Options[0].Key)
	s.Equal("dagger", weaponChoice.Options[1].Key)

	// Step 3: Verify focus choice has nested selection
	focusChoice := choices[1]
	s.Equal("wizard-equip-1", focusChoice.ID)
	s.Len(focusChoice.Options, 2)

	// Component pouch is a direct selection
	componentPouch := focusChoice.Options[0]
	s.Equal("component-pouch", componentPouch.Key)

	// Arcane focus should trigger nested selection
	arcaneFocus := focusChoice.Options[1]
	s.Contains(arcaneFocus.Key, "nested")
	s.Equal("arcane focus", arcaneFocus.Name)
	s.Contains(arcaneFocus.Description, "Choose 1")
}

func (s *EquipmentIntegrationTestSuite) TestEquipmentBundleHandling() {
	// Test that equipment bundles are properly formatted

	class := &rulebook.Class{
		Key:  "ranger",
		Name: "Ranger",
		StartingEquipmentChoices: []*shared.Choice{
			{
				Name:  "Ranged weapon choice",
				Count: 1,
				Type:  shared.ChoiceTypeEquipment,
				Options: []shared.Option{
					&shared.MultipleOption{
						Key:  "longbow-bundle",
						Name: "longbow and 20 arrows",
						Items: []shared.Option{
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
					&shared.MultipleOption{
						Key:  "crossbow-bundle",
						Name: "light crossbow and 20 bolts",
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
				},
			},
		},
	}

	choices, err := s.resolver.ResolveEquipmentChoices(s.ctx, class)
	s.NoError(err)
	s.Len(choices, 1)

	// Check bundle formatting
	s.Len(choices[0].Options, 2)
	s.Equal("longbow-bundle", choices[0].Options[0].Key)
	s.Contains(choices[0].Options[0].Name, "Longbow and 20x Arrow")
	s.Equal("crossbow-bundle", choices[0].Options[1].Key)
	s.Contains(choices[0].Options[1].Name, "Light Crossbow and 20x Crossbow Bolt")
}

// Test Data Creation Helpers

func createFullFighterClass() *rulebook.Class {
	return &rulebook.Class{
		Key:  "fighter",
		Name: "Fighter",
		StartingEquipment: []*rulebook.StartingEquipment{
			{
				Equipment: &shared.ReferenceItem{
					Key:  "explorers-pack",
					Name: "Explorer's Pack",
				},
				Quantity: 1,
			},
		},
		StartingEquipmentChoices: []*shared.Choice{
			// Choice 1: Armor
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
			// Choice 2: Weapons
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
								Name:    "martial weapon",
								Count:   1,
								Type:    shared.ChoiceTypeEquipment,
								Options: createMartialWeaponOptions(),
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
						Name:    "two martial weapons",
						Count:   2,
						Type:    shared.ChoiceTypeEquipment,
						Options: createMartialWeaponOptions(),
					},
				},
			},
			// Choice 3: Ranged
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
			// Choice 4: Pack
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
}

func createFullWizardClass() *rulebook.Class {
	return &rulebook.Class{
		Key:  "wizard",
		Name: "Wizard",
		StartingEquipment: []*rulebook.StartingEquipment{
			{
				Equipment: &shared.ReferenceItem{
					Key:  "spellbook",
					Name: "Spellbook",
				},
				Quantity: 1,
			},
		},
		StartingEquipmentChoices: []*shared.Choice{
			// Choice 1: Weapon
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
			// Choice 2: Focus
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
			// Choice 3: Pack
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
}

func createMartialWeaponOptions() []shared.Option {
	return []shared.Option{
		&shared.ReferenceOption{
			Reference: &shared.ReferenceItem{
				Key:  "battleaxe",
				Name: "Battleaxe",
			},
		},
		&shared.ReferenceOption{
			Reference: &shared.ReferenceItem{
				Key:  "flail",
				Name: "Flail",
			},
		},
		&shared.ReferenceOption{
			Reference: &shared.ReferenceItem{
				Key:  "glaive",
				Name: "Glaive",
			},
		},
		&shared.ReferenceOption{
			Reference: &shared.ReferenceItem{
				Key:  "greataxe",
				Name: "Greataxe",
			},
		},
		&shared.ReferenceOption{
			Reference: &shared.ReferenceItem{
				Key:  "greatsword",
				Name: "Greatsword",
			},
		},
		&shared.ReferenceOption{
			Reference: &shared.ReferenceItem{
				Key:  "halberd",
				Name: "Halberd",
			},
		},
		&shared.ReferenceOption{
			Reference: &shared.ReferenceItem{
				Key:  "lance",
				Name: "Lance",
			},
		},
		&shared.ReferenceOption{
			Reference: &shared.ReferenceItem{
				Key:  "longsword",
				Name: "Longsword",
			},
		},
		&shared.ReferenceOption{
			Reference: &shared.ReferenceItem{
				Key:  "maul",
				Name: "Maul",
			},
		},
		&shared.ReferenceOption{
			Reference: &shared.ReferenceItem{
				Key:  "morningstar",
				Name: "Morningstar",
			},
		},
		&shared.ReferenceOption{
			Reference: &shared.ReferenceItem{
				Key:  "pike",
				Name: "Pike",
			},
		},
		&shared.ReferenceOption{
			Reference: &shared.ReferenceItem{
				Key:  "rapier",
				Name: "Rapier",
			},
		},
		&shared.ReferenceOption{
			Reference: &shared.ReferenceItem{
				Key:  "scimitar",
				Name: "Scimitar",
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
				Key:  "trident",
				Name: "Trident",
			},
		},
		&shared.ReferenceOption{
			Reference: &shared.ReferenceItem{
				Key:  "war-pick",
				Name: "War Pick",
			},
		},
		&shared.ReferenceOption{
			Reference: &shared.ReferenceItem{
				Key:  "warhammer",
				Name: "Warhammer",
			},
		},
		&shared.ReferenceOption{
			Reference: &shared.ReferenceItem{
				Key:  "whip",
				Name: "Whip",
			},
		},
	}
}
