package character_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	mockdnd5e "github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e/mock"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
)

// AllClassesEquipmentTestSuite tests equipment choices for all D&D classes
type AllClassesEquipmentTestSuite struct {
	suite.Suite
	ctrl          *gomock.Controller
	mockDNDClient *mockdnd5e.MockClient
	resolver      character.ChoiceResolver
	ctx           context.Context
}

// SetupTest runs before each test
func (s *AllClassesEquipmentTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.mockDNDClient = mockdnd5e.NewMockClient(s.ctrl)
	s.ctx = context.Background()
	s.resolver = character.NewChoiceResolver(s.mockDNDClient)
}

// TearDownTest runs after each test
func (s *AllClassesEquipmentTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

// Test suite runner
func TestAllClassesEquipmentSuite(t *testing.T) {
	suite.Run(t, new(AllClassesEquipmentTestSuite))
}

// Barbarian Tests

func (s *AllClassesEquipmentTestSuite) TestBarbarianEquipmentChoices() {
	class := &entities.Class{
		Key:  "barbarian",
		Name: "Barbarian",
		StartingEquipmentChoices: []*entities.Choice{
			{
				Name:  "(a) a greataxe or (b) any martial melee weapon",
				Count: 1,
				Type:  entities.ChoiceTypeEquipment,
				Options: []entities.Option{
					&entities.ReferenceOption{
						Reference: &entities.ReferenceItem{
							Key:  "greataxe",
							Name: "Greataxe",
						},
					},
					&entities.Choice{
						Name:    "any martial melee weapon",
						Count:   1,
						Type:    entities.ChoiceTypeEquipment,
						Options: createMartialMeleeWeaponOptions(),
					},
				},
			},
			{
				Name:  "(a) two handaxes or (b) any simple weapon",
				Count: 1,
				Type:  entities.ChoiceTypeEquipment,
				Options: []entities.Option{
					&entities.CountedReferenceOption{
						Count: 2,
						Reference: &entities.ReferenceItem{
							Key:  "handaxe",
							Name: "Handaxe",
						},
					},
					&entities.Choice{
						Name:    "any simple weapon",
						Count:   1,
						Type:    entities.ChoiceTypeEquipment,
						Options: createSimpleWeaponOptions(),
					},
				},
			},
		},
	}

	choices, err := s.resolver.ResolveEquipmentChoices(s.ctx, class)
	s.NoError(err)
	s.Len(choices, 2)

	// First choice - weapon
	s.Equal("barbarian-equip-0", choices[0].ID)
	s.Equal("greataxe", choices[0].Options[0].Key)
	s.Contains(choices[0].Options[0].Description, "1d12 slashing")
	s.Contains(choices[0].Options[1].Key, "nested") // Martial weapon choice

	// Second choice - secondary weapon
	s.Equal("barbarian-equip-1", choices[1].ID)
	s.Equal("2x Handaxe", choices[1].Options[0].Name)
	s.Contains(choices[1].Options[1].Key, "nested") // Simple weapon choice
}

// Bard Tests

func (s *AllClassesEquipmentTestSuite) TestBardEquipmentChoices() {
	class := &entities.Class{
		Key:  "bard",
		Name: "Bard",
		StartingEquipmentChoices: []*entities.Choice{
			{
				Name:  "(a) a rapier, (b) a longsword, or (c) any simple weapon",
				Count: 1,
				Type:  entities.ChoiceTypeEquipment,
				Options: []entities.Option{
					&entities.ReferenceOption{
						Reference: &entities.ReferenceItem{
							Key:  "rapier",
							Name: "Rapier",
						},
					},
					&entities.ReferenceOption{
						Reference: &entities.ReferenceItem{
							Key:  "longsword",
							Name: "Longsword",
						},
					},
					&entities.Choice{
						Name:    "any simple weapon",
						Count:   1,
						Type:    entities.ChoiceTypeEquipment,
						Options: createSimpleWeaponOptions(),
					},
				},
			},
			{
				Name:  "(a) a diplomat's pack or (b) an entertainer's pack",
				Count: 1,
				Type:  entities.ChoiceTypeEquipment,
				Options: []entities.Option{
					&entities.ReferenceOption{
						Reference: &entities.ReferenceItem{
							Key:  "diplomats-pack",
							Name: "Diplomat's Pack",
						},
					},
					&entities.ReferenceOption{
						Reference: &entities.ReferenceItem{
							Key:  "entertainers-pack",
							Name: "Entertainer's Pack",
						},
					},
				},
			},
			{
				Name:  "(a) a lute or (b) any other musical instrument",
				Count: 1,
				Type:  entities.ChoiceTypeEquipment,
				Options: []entities.Option{
					&entities.ReferenceOption{
						Reference: &entities.ReferenceItem{
							Key:  "lute",
							Name: "Lute",
						},
					},
					&entities.Choice{
						Name:    "any other musical instrument",
						Count:   1,
						Type:    entities.ChoiceTypeEquipment,
						Options: createMusicalInstrumentOptions(),
					},
				},
			},
		},
	}

	choices, err := s.resolver.ResolveEquipmentChoices(s.ctx, class)
	s.NoError(err)
	s.Len(choices, 3)

	// Weapon choice
	s.Len(choices[0].Options, 3)
	s.Equal("rapier", choices[0].Options[0].Key)
	s.Equal("longsword", choices[0].Options[1].Key)
	s.Contains(choices[0].Options[2].Key, "nested")

	// Pack choice
	s.Len(choices[1].Options, 2)

	// Instrument choice
	s.Equal("lute", choices[2].Options[0].Key)
	s.Contains(choices[2].Options[1].Key, "nested")
}

// Druid Tests

func (s *AllClassesEquipmentTestSuite) TestDruidEquipmentChoices() {
	class := &entities.Class{
		Key:  "druid",
		Name: "Druid",
		StartingEquipmentChoices: []*entities.Choice{
			{
				Name:  "(a) a wooden shield or (b) any simple weapon",
				Count: 1,
				Type:  entities.ChoiceTypeEquipment,
				Options: []entities.Option{
					&entities.ReferenceOption{
						Reference: &entities.ReferenceItem{
							Key:  "shield",
							Name: "Shield",
						},
					},
					&entities.Choice{
						Name:    "any simple weapon",
						Count:   1,
						Type:    entities.ChoiceTypeEquipment,
						Options: createSimpleWeaponOptions(),
					},
				},
			},
			{
				Name:  "(a) a scimitar or (b) any simple melee weapon",
				Count: 1,
				Type:  entities.ChoiceTypeEquipment,
				Options: []entities.Option{
					&entities.ReferenceOption{
						Reference: &entities.ReferenceItem{
							Key:  "scimitar",
							Name: "Scimitar",
						},
					},
					&entities.Choice{
						Name:    "any simple melee weapon",
						Count:   1,
						Type:    entities.ChoiceTypeEquipment,
						Options: createSimpleMeleeWeaponOptions(),
					},
				},
			},
		},
	}

	choices, err := s.resolver.ResolveEquipmentChoices(s.ctx, class)
	s.NoError(err)
	s.Len(choices, 2)

	// Shield/weapon choice
	s.Equal("shield", choices[0].Options[0].Key)
	s.Contains(choices[0].Options[0].Description, "+2 AC")
	s.Contains(choices[0].Options[1].Key, "nested")

	// Weapon choice
	s.Equal("scimitar", choices[1].Options[0].Key)
	s.Contains(choices[1].Options[0].Description, "1d6 slashing")
	s.Contains(choices[1].Options[1].Key, "nested")
}

// Paladin Tests

func (s *AllClassesEquipmentTestSuite) TestPaladinEquipmentChoices() {
	class := &entities.Class{
		Key:  "paladin",
		Name: "Paladin",
		StartingEquipmentChoices: []*entities.Choice{
			{
				Name:  "(a) a martial weapon and a shield or (b) two martial weapons",
				Count: 1,
				Type:  entities.ChoiceTypeEquipment,
				Options: []entities.Option{
					&entities.MultipleOption{
						Key:  "weapon-shield",
						Name: "a martial weapon and a shield",
						Items: []entities.Option{
							&entities.Choice{
								Name:    "martial weapon",
								Count:   1,
								Type:    entities.ChoiceTypeEquipment,
								Options: createMartialWeaponOptions(),
							},
							&entities.ReferenceOption{
								Reference: &entities.ReferenceItem{
									Key:  "shield",
									Name: "Shield",
								},
							},
						},
					},
					&entities.Choice{
						Name:    "two martial weapons",
						Count:   2,
						Type:    entities.ChoiceTypeEquipment,
						Options: createMartialWeaponOptions(),
					},
				},
			},
			{
				Name:  "(a) five javelins or (b) any simple melee weapon",
				Count: 1,
				Type:  entities.ChoiceTypeEquipment,
				Options: []entities.Option{
					&entities.CountedReferenceOption{
						Count: 5,
						Reference: &entities.ReferenceItem{
							Key:  "javelin",
							Name: "Javelin",
						},
					},
					&entities.Choice{
						Name:    "any simple melee weapon",
						Count:   1,
						Type:    entities.ChoiceTypeEquipment,
						Options: createSimpleMeleeWeaponOptions(),
					},
				},
			},
			{
				Name:  "(a) a priest's pack or (b) an explorer's pack",
				Count: 1,
				Type:  entities.ChoiceTypeEquipment,
				Options: []entities.Option{
					&entities.ReferenceOption{
						Reference: &entities.ReferenceItem{
							Key:  "priests-pack",
							Name: "Priest's Pack",
						},
					},
					&entities.ReferenceOption{
						Reference: &entities.ReferenceItem{
							Key:  "explorers-pack",
							Name: "Explorer's Pack",
						},
					},
				},
			},
		},
	}

	choices, err := s.resolver.ResolveEquipmentChoices(s.ctx, class)
	s.NoError(err)
	s.Len(choices, 3)

	// Weapon choice - both options are nested
	s.Contains(choices[0].Options[0].Key, "nested")
	s.Contains(choices[0].Options[1].Key, "nested")

	// Javelin/weapon choice
	s.Equal("5x Javelin", choices[1].Options[0].Name)
	s.Contains(choices[1].Options[1].Key, "nested")
}

// Sorcerer Tests

func (s *AllClassesEquipmentTestSuite) TestSorcererEquipmentChoices() {
	class := &entities.Class{
		Key:  "sorcerer",
		Name: "Sorcerer",
		StartingEquipmentChoices: []*entities.Choice{
			{
				Name:  "(a) a light crossbow and 20 bolts or (b) any simple weapon",
				Count: 1,
				Type:  entities.ChoiceTypeEquipment,
				Options: []entities.Option{
					&entities.MultipleOption{
						Key:  "crossbow-bundle",
						Name: "a light crossbow and 20 bolts",
						Items: []entities.Option{
							&entities.ReferenceOption{
								Reference: &entities.ReferenceItem{
									Key:  "light-crossbow",
									Name: "Light Crossbow",
								},
							},
							&entities.CountedReferenceOption{
								Count: 20,
								Reference: &entities.ReferenceItem{
									Key:  "crossbow-bolt",
									Name: "Crossbow Bolt",
								},
							},
						},
					},
					&entities.Choice{
						Name:    "any simple weapon",
						Count:   1,
						Type:    entities.ChoiceTypeEquipment,
						Options: createSimpleWeaponOptions(),
					},
				},
			},
			{
				Name:  "(a) a component pouch or (b) an arcane focus",
				Count: 1,
				Type:  entities.ChoiceTypeEquipment,
				Options: []entities.Option{
					&entities.ReferenceOption{
						Reference: &entities.ReferenceItem{
							Key:  "component-pouch",
							Name: "Component Pouch",
						},
					},
					&entities.Choice{
						Name:    "arcane focus",
						Count:   1,
						Type:    entities.ChoiceTypeEquipment,
						Options: createArcaneFocusOptions(),
					},
				},
			},
		},
	}

	choices, err := s.resolver.ResolveEquipmentChoices(s.ctx, class)
	s.NoError(err)
	s.Len(choices, 2)

	// Weapon choice
	s.Contains(choices[0].Options[0].Name, "Light Crossbow and 20x Crossbow Bolt")
	s.Contains(choices[0].Options[1].Key, "nested")

	// Focus choice
	s.Equal("component-pouch", choices[1].Options[0].Key)
	s.Contains(choices[1].Options[1].Key, "nested")
}

// Warlock Tests

func (s *AllClassesEquipmentTestSuite) TestWarlockEquipmentChoices() {
	class := &entities.Class{
		Key:  "warlock",
		Name: "Warlock",
		StartingEquipmentChoices: []*entities.Choice{
			{
				Name:  "(a) a light crossbow and 20 bolts or (b) any simple weapon",
				Count: 1,
				Type:  entities.ChoiceTypeEquipment,
				Options: []entities.Option{
					&entities.MultipleOption{
						Key:  "crossbow-bundle",
						Name: "a light crossbow and 20 bolts",
						Items: []entities.Option{
							&entities.ReferenceOption{
								Reference: &entities.ReferenceItem{
									Key:  "light-crossbow",
									Name: "Light Crossbow",
								},
							},
							&entities.CountedReferenceOption{
								Count: 20,
								Reference: &entities.ReferenceItem{
									Key:  "crossbow-bolt",
									Name: "Crossbow Bolt",
								},
							},
						},
					},
					&entities.Choice{
						Name:    "any simple weapon",
						Count:   1,
						Type:    entities.ChoiceTypeEquipment,
						Options: createSimpleWeaponOptions(),
					},
				},
			},
			{
				Name:  "(a) a component pouch or (b) an arcane focus",
				Count: 1,
				Type:  entities.ChoiceTypeEquipment,
				Options: []entities.Option{
					&entities.ReferenceOption{
						Reference: &entities.ReferenceItem{
							Key:  "component-pouch",
							Name: "Component Pouch",
						},
					},
					&entities.Choice{
						Name:    "arcane focus",
						Count:   1,
						Type:    entities.ChoiceTypeEquipment,
						Options: createArcaneFocusOptions(),
					},
				},
			},
			{
				Name:  "(a) a scholar's pack or (b) a dungeoneer's pack",
				Count: 1,
				Type:  entities.ChoiceTypeEquipment,
				Options: []entities.Option{
					&entities.ReferenceOption{
						Reference: &entities.ReferenceItem{
							Key:  "scholars-pack",
							Name: "Scholar's Pack",
						},
					},
					&entities.ReferenceOption{
						Reference: &entities.ReferenceItem{
							Key:  "dungeoneers-pack",
							Name: "Dungeoneer's Pack",
						},
					},
				},
			},
		},
	}

	choices, err := s.resolver.ResolveEquipmentChoices(s.ctx, class)
	s.NoError(err)
	s.Len(choices, 3)
}

// Helper functions for creating weapon/item options

func createMartialMeleeWeaponOptions() []entities.Option {
	weapons := []string{
		"battleaxe", "flail", "glaive", "greataxe", "greatsword",
		"halberd", "lance", "longsword", "maul", "morningstar",
		"pike", "rapier", "scimitar", "shortsword", "trident",
		"war-pick", "warhammer", "whip",
	}

	options := []entities.Option{}
	for _, weapon := range weapons {
		options = append(options, &entities.ReferenceOption{
			Reference: &entities.ReferenceItem{
				Key:  weapon,
				Name: capitalize(weapon),
			},
		})
	}
	return options
}

func createSimpleWeaponOptions() []entities.Option {
	weapons := []string{
		"club", "dagger", "greatclub", "handaxe", "javelin",
		"light-hammer", "mace", "quarterstaff", "sickle", "spear",
		"light-crossbow", "dart", "shortbow", "sling",
	}

	options := []entities.Option{}
	for _, weapon := range weapons {
		options = append(options, &entities.ReferenceOption{
			Reference: &entities.ReferenceItem{
				Key:  weapon,
				Name: capitalize(weapon),
			},
		})
	}
	return options
}

func createSimpleMeleeWeaponOptions() []entities.Option {
	weapons := []string{
		"club", "dagger", "greatclub", "handaxe", "javelin",
		"light-hammer", "mace", "quarterstaff", "sickle", "spear",
	}

	options := []entities.Option{}
	for _, weapon := range weapons {
		options = append(options, &entities.ReferenceOption{
			Reference: &entities.ReferenceItem{
				Key:  weapon,
				Name: capitalize(weapon),
			},
		})
	}
	return options
}

func createMusicalInstrumentOptions() []entities.Option {
	instruments := []string{
		"bagpipes", "drum", "dulcimer", "flute", "horn",
		"lyre", "pan-flute", "shawm", "viol",
	}

	options := []entities.Option{}
	for _, instrument := range instruments {
		options = append(options, &entities.ReferenceOption{
			Reference: &entities.ReferenceItem{
				Key:  instrument,
				Name: capitalize(instrument),
			},
		})
	}
	return options
}

func createArcaneFocusOptions() []entities.Option {
	focuses := []string{
		"crystal", "orb", "rod", "staff", "wand",
	}

	options := []entities.Option{}
	for _, focus := range focuses {
		options = append(options, &entities.ReferenceOption{
			Reference: &entities.ReferenceItem{
				Key:  focus,
				Name: capitalize(focus),
			},
		})
	}
	return options
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	// Simple capitalization - in production you'd handle multi-word items better
	return string(s[0]-32) + s[1:]
}
