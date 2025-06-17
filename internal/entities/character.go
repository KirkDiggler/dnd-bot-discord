package entities

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/KirkDiggler/dnd-bot-discord/internal/dice"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities/attack"
)

type Slot string

const (
	SlotMainHand  Slot = "main-hand"
	SlotOffHand   Slot = "off-hand"
	SlotTwoHanded Slot = "two-handed"
	SlotBody      Slot = "body"
	SlotNone      Slot = "none"
)

type CharacterStatus string

const (
	CharacterStatusDraft    CharacterStatus = "draft"
	CharacterStatusActive   CharacterStatus = "active"
	CharacterStatusArchived CharacterStatus = "archived"
)

// AbilityRoll represents a single ability score roll with a unique ID
type AbilityRoll struct {
	ID    string `json:"id"`
	Value int    `json:"value"`
}

type Character struct {
	ID                 string
	OwnerID            string
	RealmID            string
	Name               string
	Speed              int
	Race               *Race
	Class              *Class
	Background         *Background
	Attributes         map[Attribute]*AbilityScore
	Rolls              []*dice.RollResult
	AbilityRolls       []AbilityRoll // New field for ability score rolls with IDs
	AbilityAssignments map[string]string // Maps ability name (STR, DEX, etc.) to roll ID
	Proficiencies      map[ProficiencyType][]*Proficiency
	ProficiencyChoices []*Choice
	Inventory          map[EquipmentType][]Equipment

	HitDie           int
	AC               int
	MaxHitPoints     int
	CurrentHitPoints int
	Level            int
	Experience       int
	NextLevel        int

	EquippedSlots map[Slot]Equipment

	Status CharacterStatus `json:"status"`

	mu sync.Mutex
}

func (c *Character) Attack() ([]*attack.Result, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.EquippedSlots == nil {
		// Improvised weapon range or melee

		a, err := c.improvisedMelee()
		if err != nil {
			return nil, err
		}

		return []*attack.Result{
			a,
		}, nil

	}

	if c.EquippedSlots[SlotMainHand] != nil {
		if weap, ok := c.EquippedSlots[SlotMainHand].(*Weapon); ok {
			attacks := make([]*attack.Result, 0)
			attak1, err := weap.Attack(c)
			if err != nil {
				return nil, err
			}
			attacks = append(attacks, attak1)

			if offWeap, offOk := c.EquippedSlots[SlotOffHand].(*Weapon); offOk {
				attak2, err := offWeap.Attack(c)
				if err != nil {
					return nil, err
				}
				attacks = append(attacks, attak2)
			}

			return attacks, nil
		}
	}

	if c.EquippedSlots[SlotTwoHanded] != nil {
		if weap, ok := c.EquippedSlots[SlotTwoHanded].(*Weapon); ok {
			a, err := weap.Attack(c)
			if err != nil {
				return nil, err
			}

			return []*attack.Result{
				a,
			}, nil
		}
	}

	a, err := c.improvisedMelee()
	if err != nil {
		return nil, err
	}

	return []*attack.Result{
		a,
	}, nil
}

func (c *Character) improvisedMelee() (*attack.Result, error) {
	bonus := c.Attributes[AttributeStrength].Bonus
	attackRoll, err := dice.Roll(1, 20, 0)
	if err != nil {
		return nil, err
	}
	damageRoll, err := dice.Roll(1, 4, 0)
	if err != nil {
		return nil, err
	}

	return &attack.Result{
		AttackRoll: attackRoll.Total + bonus,
		DamageRoll: damageRoll.Total + bonus,
		AttackType: "bludgening",
	}, nil
}

func (c *Character) getEquipment(key string) Equipment {
	for _, v := range c.Inventory {
		for _, eq := range v {
			if eq.GetKey() == key {
				return eq
			}
		}
	}

	return nil
}

// Equip equips the item if it is found in the inventory, otherwise it is a noop
func (c *Character) Equip(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	defer c.calculateAC()

	equipment := c.getEquipment(key)
	if equipment == nil {
		return false
	}

	if c.EquippedSlots == nil {
		c.EquippedSlots = make(map[Slot]Equipment)
	}

	c.EquippedSlots[SlotTwoHanded] = nil

	switch equipment.GetSlot() {
	case SlotMainHand:
		if c.EquippedSlots[SlotMainHand] != nil {
			c.EquippedSlots[SlotOffHand] = c.EquippedSlots[SlotMainHand]
		}
	case SlotTwoHanded:
		c.EquippedSlots[SlotMainHand] = nil
		c.EquippedSlots[SlotOffHand] = nil
	}

	c.EquippedSlots[equipment.GetSlot()] = equipment

	return true
}

func (c *Character) calculateAC() {
	c.AC = 10
	
	// First, check for body armor which sets the base AC
	if bodyArmor := c.EquippedSlots[SlotBody]; bodyArmor != nil {
		if bodyArmor.GetEquipmentType() == "Armor" {
			armor := bodyArmor.(*Armor)
			if armor.ArmorClass != nil {
				c.AC = armor.ArmorClass.Base
				if armor.ArmorClass.DexBonus {
					// TODO: load max and bonus and limit id applicable
					c.AC += c.Attributes[AttributeDexterity].Bonus
				}
			}
		}
	}
	
	// Then add bonuses from other armor pieces (like shields)
	for slot, e := range c.EquippedSlots {
		if e == nil || slot == SlotBody {
			continue
		}

		if e.GetEquipmentType() == "Armor" {
			armor := e.(*Armor)
			if armor.ArmorClass == nil {
				continue
			}
			c.AC += armor.ArmorClass.Base
			if armor.ArmorClass.DexBonus {
				c.AC += c.Attributes[AttributeDexterity].Bonus
			}
		}
	}
}

func (c *Character) SetHitpoints() {
	if c.Attributes == nil {
		return
	}

	if c.Attributes[AttributeConstitution] == nil {
		return
	}

	if c.HitDie == 0 {
		return
	}

	c.MaxHitPoints = c.HitDie + c.Attributes[AttributeConstitution].Bonus
	c.CurrentHitPoints = c.MaxHitPoints
}

func (c *Character) AddAttribute(attr Attribute, score int) {
	if c.Attributes == nil {
		c.Attributes = make(map[Attribute]*AbilityScore)
	}

	bonus := 0
	if _, ok := c.Attributes[attr]; ok {
		bonus = c.Attributes[attr].Bonus
	}
	abilityScore := &AbilityScore{
		Score: score,
		Bonus: bonus,
	}
	switch {
	case score == 1:
		abilityScore.Bonus += -5
	case score < 4 && score > 1:
		abilityScore.Bonus += -4
	case score < 6 && score > 3:
		abilityScore.Bonus += -3
	case score < 8 && score > 5:
		abilityScore.Bonus += -2
	case score < 10 && score >= 8:
		abilityScore.Bonus += -1
	case score < 12 && score > 9:
		abilityScore.Bonus += 0
	case score < 14 && score > 11:
		abilityScore.Bonus += 1
	case score < 16 && score > 13:
		abilityScore.Bonus += 2
	case score < 18 && score > 15:
		abilityScore.Bonus += 3
	case score < 20 && score > 17:
		abilityScore.Bonus += 4
	case score == 20:
		abilityScore.Bonus += 5
	}

	c.Attributes[attr] = abilityScore
}
func (c *Character) AddAbilityBonus(ab *AbilityBonus) {
	if c.Attributes == nil {
		c.Attributes = make(map[Attribute]*AbilityScore)
	}

	if _, ok := c.Attributes[ab.Attribute]; !ok {
		c.Attributes[ab.Attribute] = &AbilityScore{}
	}

	c.Attributes[ab.Attribute] = c.Attributes[ab.Attribute].AddBonus(ab.Bonus)
}

func (c *Character) AddInventory(e Equipment) {
	if c.Inventory == nil {
		c.Inventory = make(map[EquipmentType][]Equipment)
	}

	c.mu.Lock()
	if c.Inventory[e.GetEquipmentType()] == nil {
		c.Inventory[e.GetEquipmentType()] = make([]Equipment, 0)
	}

	c.Inventory[e.GetEquipmentType()] = append(c.Inventory[e.GetEquipmentType()], e)
	c.mu.Unlock()
}

func (c *Character) AddProficiency(p *Proficiency) {
	if c.Proficiencies == nil {
		c.Proficiencies = make(map[ProficiencyType][]*Proficiency)
	}
	c.mu.Lock()
	if c.Proficiencies[p.Type] == nil {
		c.Proficiencies[p.Type] = make([]*Proficiency, 0)
	}

	c.Proficiencies[p.Type] = append(c.Proficiencies[p.Type], p)
	c.mu.Unlock()
}

func (c *Character) AddAbilityScoreBonus(attr Attribute, bonus int) {
	if c.Attributes == nil {
		c.Attributes = make(map[Attribute]*AbilityScore)
	}

	c.Attributes[attr] = c.Attributes[attr].AddBonus(bonus)
}

func (c *Character) NameString() string {
	if c.Race == nil || c.Class == nil {
		return "Character not fully created"
	}

	return fmt.Sprintf("%s the %s %s", c.Name, c.Race.Name, c.Class.Name)
}

func (c *Character) StatsString() string {
	msg := strings.Builder{}
	msg.WriteString(fmt.Sprintf("  -  Speed: %d\n", c.Speed))
	msg.WriteString(fmt.Sprintf("  -  Hit Die: %d\n", c.HitDie))
	msg.WriteString(fmt.Sprintf("  -  AC: %d\n", c.AC))
	msg.WriteString(fmt.Sprintf("  -  Max Hit Points: %d\n", c.MaxHitPoints))
	msg.WriteString(fmt.Sprintf("  -  Current Hit Points: %d\n", c.CurrentHitPoints))
	msg.WriteString(fmt.Sprintf("  -  Level: %d\n", c.Level))
	msg.WriteString(fmt.Sprintf("  -  Experience: %d\n", c.Experience))

	return msg.String()
}

func (c *Character) String() string {
	msg := strings.Builder{}
	if c.Race == nil || c.Class == nil {
		return "Character not fully created"
	}

	msg.WriteString(fmt.Sprintf("%s the %s %s\n", c.Name, c.Race.Name, c.Class.Name))

	msg.WriteString("**Rolls**:\n")
	for _, roll := range c.Rolls {
		msg.WriteString(fmt.Sprintf("%s, ", roll))
	}
	msg.WriteString("\n")
	msg.WriteString("\n**Stats**:\n")
	msg.WriteString(fmt.Sprintf("  -  Speed: %d\n", c.Speed))
	msg.WriteString(fmt.Sprintf("  -  Hit Die: %d\n", c.HitDie))
	msg.WriteString(fmt.Sprintf("  -  AC: %d\n", c.AC))
	msg.WriteString(fmt.Sprintf("  -  Max Hit Points: %d\n", c.MaxHitPoints))
	msg.WriteString(fmt.Sprintf("  -  Current Hit Points: %d\n", c.CurrentHitPoints))
	msg.WriteString(fmt.Sprintf("  -  Level: %d\n", c.Level))
	msg.WriteString(fmt.Sprintf("  -  Experience: %d\n", c.Experience))

	msg.WriteString("\n**Attributes**:\n")
	for _, attr := range Attributes {
		if c.Attributes[attr] == nil {
			continue
		}
		msg.WriteString(fmt.Sprintf("  -  %s: %s\n", attr, c.Attributes[attr]))
	}

	msg.WriteString("\n**Proficiencies**:\n")
	for _, key := range ProficiencyTypes {
		if c.Proficiencies[key] == nil {
			continue
		}

		msg.WriteString(fmt.Sprintf("  -  **%s**:\n", key))
		for _, prof := range c.Proficiencies[key] {
			msg.WriteString(fmt.Sprintf("    -  %s\n", prof.Name))
		}
	}

	msg.WriteString("\n**Inventory**:\n")
	for key := range c.Inventory {
		if c.Inventory[key] == nil {
			continue
		}

		msg.WriteString(fmt.Sprintf("  -  **%s**:\n", key))
		for _, item := range c.Inventory[key] {
			if c.IsEquipped(item) {
				msg.WriteString(fmt.Sprintf("    -  %s (Equipped)\n", item.GetName()))
				continue
			}

			msg.WriteString(fmt.Sprintf("    -  %s \n", item.GetName()))
		}

	}
	return msg.String()
}

func (c *Character) IsEquipped(e Equipment) bool {
	for _, item := range c.EquippedSlots {
		if item == nil {
			continue
		}
		log.Printf("item: %s, e: %s", item.GetKey(), e.GetKey())

		if item.GetKey() == e.GetKey() {
			return true
		}
	}

	return false
}

func (c *Character) resetRacialTraits() {
	// Clear all downstream data
	c.Proficiencies = make(map[ProficiencyType][]*Proficiency)
	c.resetAbilityScores()
}

func (c *Character) resetClassFeatures() {
	// Clear all downstream data
	c.Proficiencies = make(map[ProficiencyType][]*Proficiency)
	c.Inventory = make(map[EquipmentType][]Equipment)
	c.EquippedSlots = make(map[Slot]Equipment)
}

func (c *Character) resetBackground() {
	c.Background = nil
	// Clear all downstream data
	c.Proficiencies = make(map[ProficiencyType][]*Proficiency)
	c.Inventory = make(map[EquipmentType][]Equipment)
}

func (c *Character) resetAbilityScores() {
	c.Attributes = make(map[Attribute]*AbilityScore)
}

func (c *Character) resetProficienciesBySource(step CreateStep) {}
func (c *Character) resetSkillsBySource(step CreateStep)        {}
func (c *Character) resetEquipmentBySource(step CreateStep)     {}
