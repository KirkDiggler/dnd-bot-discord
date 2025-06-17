package testutils

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
)

// CreateTestRace creates a test race entity
func CreateTestRace(key, name string) *entities.Race {
	return &entities.Race{
		Key:   key,
		Name:  name,
		Speed: 30,
		AbilityBonuses: []*entities.AbilityBonus{
			{
				Attribute: entities.AttributeStrength,
				Bonus:     2,
			},
		},
		StartingProficiencies: []*entities.ReferenceItem{},
	}
}

// CreateTestClass creates a test class entity
func CreateTestClass(key, name string, hitDie int) *entities.Class {
	return &entities.Class{
		Key:    key,
		Name:   name,
		HitDie: hitDie,
		Proficiencies: []*entities.ReferenceItem{
			{
				Key:  "skill-athletics",
				Name: "Athletics",
			},
		},
		StartingEquipment: []*entities.StartingEquipment{
			{
				Quantity: 1,
				Equipment: &entities.ReferenceItem{
					Key:  "longsword",
					Name: "Longsword",
				},
			},
		},
		ProficiencyChoices:       []*entities.Choice{},
		StartingEquipmentChoices: []*entities.Choice{},
	}
}

// CreateTestCharacter creates a fully formed test character
func CreateTestCharacter(id, ownerID, realmID, name string) *entities.Character {
	char := &entities.Character{
		ID:       id,
		OwnerID:  ownerID,
		RealmID:  realmID,
		Name:     name,
		Race:     CreateTestRace("human", "Human"),
		Class:    CreateTestClass("fighter", "Fighter", 10),
		Level:    1,
		Status:   entities.CharacterStatusActive,
		HitDie:   10,
		Speed:    30,
		Attributes: map[entities.Attribute]*entities.AbilityScore{
			entities.AttributeStrength:     {Score: 16, Bonus: 3},
			entities.AttributeDexterity:    {Score: 14, Bonus: 2},
			entities.AttributeConstitution: {Score: 15, Bonus: 2},
			entities.AttributeIntelligence: {Score: 10, Bonus: 0},
			entities.AttributeWisdom:       {Score: 13, Bonus: 1},
			entities.AttributeCharisma:     {Score: 12, Bonus: 1},
		},
		MaxHitPoints:     12,
		CurrentHitPoints: 12,
		AC:               16,
		Proficiencies:    make(map[entities.ProficiencyType][]*entities.Proficiency),
		Inventory:        make(map[entities.EquipmentType][]entities.Equipment),
		EquippedSlots:    make(map[entities.Slot]entities.Equipment),
		Features:         []*entities.CharacterFeature{},
	}
	
	// Set hit points based on constitution
	char.SetHitpoints()
	
	return char
}

// CreateTestProficiencyChoice creates a test proficiency choice
func CreateTestProficiencyChoice(name string, count int, options []string) *entities.Choice {
	choice := &entities.Choice{
		Name:    name,
		Type:    entities.ChoiceTypeProficiency,
		Key:     "test-prof-choice",
		Count:   count,
		Options: make([]entities.Option, len(options)),
	}
	
	for i, opt := range options {
		choice.Options[i] = &entities.ReferenceOption{
			Reference: &entities.ReferenceItem{
				Key:  opt,
				Name: opt,
			},
		}
	}
	
	return choice
}

// CreateTestEquipmentChoice creates a test equipment choice
func CreateTestEquipmentChoice(name string, count int, options []string) *entities.Choice {
	choice := &entities.Choice{
		Name:    name,
		Type:    entities.ChoiceTypeEquipment,
		Key:     "test-equip-choice",
		Count:   count,
		Options: make([]entities.Option, len(options)),
	}
	
	for i, opt := range options {
		choice.Options[i] = &entities.CountedReferenceOption{
			Count: 1,
			Reference: &entities.ReferenceItem{
				Key:  opt,
				Name: opt,
			},
		}
	}
	
	return choice
}

// CreateTestSession creates a test session
func CreateTestSession(id, dmID, name string) *entities.Session {
	sess := entities.NewSession(id, name, "test-realm", "test-channel", dmID)
	sess.Status = entities.SessionStatusActive
	return sess
}

// CreateTestEncounter creates a test encounter
func CreateTestEncounter(id, sessionID, name string) *entities.Encounter {
	return &entities.Encounter{
		ID:         id,
		SessionID:  sessionID,
		Name:       name,
		Status:     entities.EncounterStatusSetup,
		Round:      1,
		Turn:       0,
		Combatants: make(map[string]*entities.Combatant),
		TurnOrder:  []string{},
	}
}

// CreateTestMonster creates a test monster
func CreateTestMonster(key, name string, cr float32, hp, ac int) *entities.Monster {
	return &entities.Monster{
		Key: key,
		Template: &entities.MonsterTemplate{
			Key:             key,
			Name:            name,
			Type:            "humanoid",
			ArmorClass:      ac,
			HitPoints:       hp,
			HitDice:         "2d8",
			ChallengeRating: cr,
			XP:              50,
			Actions: []*entities.MonsterAction{
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