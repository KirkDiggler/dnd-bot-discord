package testutils

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/equipment"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/combat"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/session"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
)

// CreateTestRace creates a test race entity
func CreateTestRace(key, name string) *rulebook.Race {
	return &rulebook.Race{
		Key:   key,
		Name:  name,
		Speed: 30,
		AbilityBonuses: []*shared.AbilityBonus{
			{
				Attribute: shared.AttributeStrength,
				Bonus:     2,
			},
		},
		StartingProficiencies: []*shared.ReferenceItem{},
	}
}

// CreateTestClass creates a test class entity
func CreateTestClass(key, name string, hitDie int) *rulebook.Class {
	return &rulebook.Class{
		Key:    key,
		Name:   name,
		HitDie: hitDie,
		Proficiencies: []*shared.ReferenceItem{
			{
				Key:  "skill-athletics",
				Name: "Athletics",
			},
		},
		StartingEquipment: []*rulebook.StartingEquipment{
			{
				Quantity: 1,
				Equipment: &shared.ReferenceItem{
					Key:  "longsword",
					Name: "Longsword",
				},
			},
		},
		ProficiencyChoices:       []*shared.Choice{},
		StartingEquipmentChoices: []*shared.Choice{},
	}
}

// CreateTestCharacter creates a fully formed test character
func CreateTestCharacter(id, ownerID, realmID, name string) *character.Character {
	char := &character.Character{
		ID:      id,
		OwnerID: ownerID,
		RealmID: realmID,
		Name:    name,
		Race:    CreateTestRace("human", "Human"),
		Class:   CreateTestClass("fighter", "Fighter", 10),
		Level:   1,
		Status:  shared.CharacterStatusActive,
		HitDie:  10,
		Speed:   30,
		Attributes: map[shared.Attribute]*character.AbilityScore{
			shared.AttributeStrength:     {Score: 16, Bonus: 3},
			shared.AttributeDexterity:    {Score: 14, Bonus: 2},
			shared.AttributeConstitution: {Score: 15, Bonus: 2},
			shared.AttributeIntelligence: {Score: 10, Bonus: 0},
			shared.AttributeWisdom:       {Score: 13, Bonus: 1},
			shared.AttributeCharisma:     {Score: 12, Bonus: 1},
		},
		MaxHitPoints:     12,
		CurrentHitPoints: 12,
		AC:               16,
		Proficiencies:    make(map[rulebook.ProficiencyType][]*rulebook.Proficiency),
		Inventory:        make(map[equipment.EquipmentType][]equipment.Equipment),
		EquippedSlots:    make(map[shared.Slot]equipment.Equipment),
		Features:         []*rulebook.CharacterFeature{},
	}

	// Set hit points based on constitution
	char.SetHitpoints()

	return char
}

// CreateTestProficiencyChoice creates a test proficiency choice
func CreateTestProficiencyChoice(name string, count int, options []string) *shared.Choice {
	choice := &shared.Choice{
		Name:    name,
		Type:    shared.ChoiceTypeProficiency,
		Key:     "test-prof-choice",
		Count:   count,
		Options: make([]shared.Option, len(options)),
	}

	for i, opt := range options {
		choice.Options[i] = &shared.ReferenceOption{
			Reference: &shared.ReferenceItem{
				Key:  opt,
				Name: opt,
			},
		}
	}

	return choice
}

// CreateTestEquipmentChoice creates a test equipment choice
func CreateTestEquipmentChoice(name string, count int, options []string) *shared.Choice {
	choice := &shared.Choice{
		Name:    name,
		Type:    shared.ChoiceTypeEquipment,
		Key:     "test-equip-choice",
		Count:   count,
		Options: make([]shared.Option, len(options)),
	}

	for i, opt := range options {
		choice.Options[i] = &shared.CountedReferenceOption{
			Count: 1,
			Reference: &shared.ReferenceItem{
				Key:  opt,
				Name: opt,
			},
		}
	}

	return choice
}

// CreateTestSession creates a test session
func CreateTestSession(id, dmID, name string) *session.Session {
	sess := session.NewSession(id, name, "test-realm", "test-channel", dmID)
	sess.Status = session.SessionStatusActive
	return sess
}

// CreateTestEncounter creates a test encounter
func CreateTestEncounter(id, sessionID, name string) *combat.Encounter {
	return &combat.Encounter{
		ID:         id,
		SessionID:  sessionID,
		Name:       name,
		Status:     combat.EncounterStatusSetup,
		Round:      1,
		Turn:       0,
		Combatants: make(map[string]*combat.Combatant),
		TurnOrder:  []string{},
	}
}

// CreateTestMonster creates a test monster
func CreateTestMonster(key, name string, cr float32, hp, ac int) *combat.Monster {
	return &combat.Monster{
		Key: key,
		Template: &combat.MonsterTemplate{
			Key:             key,
			Name:            name,
			Type:            "humanoid",
			ArmorClass:      ac,
			HitPoints:       hp,
			HitDice:         "2d8",
			ChallengeRating: cr,
			XP:              50,
			Actions: []*combat.MonsterAction{
				{
					Name:        "Scimitar",
					AttackBonus: 4,
					Description: "Melee Weapon Attack: +4 to hit, reach 5 ft., one target. Hit: 5 (1d6 + 2) slashing damage.",
				},
			},
		},
		CurrentHP: hp,
	}
}
