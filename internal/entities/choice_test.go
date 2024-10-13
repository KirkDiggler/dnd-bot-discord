package entities

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type suiteChoice struct {
	suite.Suite

	choice *Choice
}

func (s *suiteChoice) SetupTest() {
	s.choice = &Choice{
		Name:  "(a) a martial weapon and a shield or (b) two martial weapons",
		Count: 1,
		Type:  ChoiceTypeEquipment,
		Options: []Option{
			&MultipleOption{
				Key:  "option-1",
				Name: "a martial weapon and a shield",
				Items: []Option{
					&Choice{
						Name:  "a martial weapon",
						Key:   "option-1-1",
						Count: 1,
						Type:  ChoiceTypeEquipment,
						Options: []Option{
							&ReferenceOption{
								Reference: &ReferenceItem{
									Key:  "battleaxe",
									Name: "Battleaxe",
								},
							},
							&ReferenceOption{
								Reference: &ReferenceItem{
									Key:  "flail",
									Name: "Flail",
								},
							},
						},
					},
					&CountedReferenceOption{
						Count: 1,
						Reference: &ReferenceItem{
							Key:  "shield",
							Name: "Shield",
						},
					},
				},
			},
			&Choice{
				Name:  "two martial weapons",
				Key:   "option-2",
				Count: 2,
				Type:  ChoiceTypeEquipment,
				Options: []Option{
					&ReferenceOption{
						Reference: &ReferenceItem{
							Key:  "handaxe",
							Name: "Handaxe",
						},
					},
					&ReferenceOption{
						Reference: &ReferenceItem{
							Key:  "longsword",
							Name: "Longsword",
						},
					},
					&ReferenceOption{
						Reference: &ReferenceItem{
							Key:  "rapier",
							Name: "Rapier",
						},
					},
				},
			},
		},
	}
}

func (s *suiteChoice) TestSelectsMultiple() {
	multi := s.choice.Select("option-1")
	s.Equal("a martial weapon and a shield", multi.Option.GetName())

	choice := multi.Option.Select("option-1-1")
	s.Equal("a martial weapon", choice.Option.GetName())
	s.Equal(ChoiceStatusActive, s.choice.GetStatus())

	s.Equal(ChoiceStatusActive, multi.Option.GetStatus())

	axe := choice.Option.Select("battleaxe")
	s.Equal("Battleaxe", axe.Option.GetName())
	s.Equal(ChoiceStatusActive, multi.Option.GetStatus())
	s.Equal(ChoiceStatusActive, s.choice.GetStatus())

	shield := multi.Option.Select("shield")
	s.Equal("Shield", shield.Option.GetName())

	s.Equal(ChoiceStatusSelected, multi.Option.GetStatus())
}

func TestChoice(t *testing.T) {
	suite.Run(t, new(suiteChoice))
}
