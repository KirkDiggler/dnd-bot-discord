package character

import (
	"fmt"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/equipment"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"log"
	"strings"
	"sync"

	"github.com/KirkDiggler/dnd-bot-discord/internal/dice"
	"github.com/KirkDiggler/dnd-bot-discord/internal/effects"
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
	Race               *rulebook.Race
	Class              *rulebook.Class
	Background         *rulebook.Background
	Attributes         map[shared.Attribute]*AbilityScore
	Rolls              []*dice.RollResult
	AbilityRolls       []AbilityRoll     // New field for ability score rolls with IDs
	AbilityAssignments map[string]string // Maps ability name (STR, DEX, etc.) to roll ID
	Proficiencies      map[rulebook.ProficiencyType][]*rulebook.Proficiency
	ProficiencyChoices []*shared.Choice
	Inventory          map[equipment.EquipmentType][]equipment.Equipment
	Features           []*rulebook.CharacterFeature // Character features (class, racial, etc.)

	HitDie           int
	AC               int
	MaxHitPoints     int
	CurrentHitPoints int
	Level            int
	Experience       int
	NextLevel        int

	EquippedSlots map[shared.Slot]equipment.Equipment

	Status shared.CharacterStatus `json:"status"`

	// Resources tracks HP, abilities, spell slots, etc
	Resources *CharacterResources `json:"resources"`

	// Spells tracks known/prepared spells
	Spells *SpellList `json:"spells,omitempty"`

	// EffectManager tracks all active status effects
	EffectManager *effects.Manager `json:"-"`

	// diceRoller is the injected dice roller (defaults to random)
	diceRoller dice.Roller `json:"-"`

	mu sync.Mutex
}

// NewCharacter creates a new character with default dice roller
func NewCharacter() *Character {
	return &Character{
		diceRoller: dice.NewRandomRoller(),
	}
}

// WithDiceRoller sets a custom dice roller (for testing)
func (c *Character) WithDiceRoller(roller dice.Roller) *Character {
	c.diceRoller = roller
	return c
}

// getDiceRoller returns the dice roller, initializing if needed
func (c *Character) getDiceRoller() dice.Roller {
	if c.diceRoller == nil {
		c.diceRoller = dice.NewRandomRoller()
	}
	return c.diceRoller
}

func (c *Character) getEquipment(key string) equipment.Equipment {
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

	equipmentValue := c.getEquipment(key)
	if equipmentValue == nil {
		return false
	}

	if c.EquippedSlots == nil {
		c.EquippedSlots = make(map[shared.Slot]equipment.Equipment)
	}

	c.EquippedSlots[shared.SlotTwoHanded] = nil

	switch equipmentValue.GetSlot() {
	case shared.SlotMainHand:
		if c.EquippedSlots[shared.SlotMainHand] != nil {
			c.EquippedSlots[shared.SlotOffHand] = c.EquippedSlots[shared.SlotMainHand]
		}
	case shared.SlotTwoHanded:
		c.EquippedSlots[shared.SlotMainHand] = nil
		c.EquippedSlots[shared.SlotOffHand] = nil
	}

	c.EquippedSlots[equipmentValue.GetSlot()] = equipmentValue

	return true
}

func (c *Character) calculateAC() {
	// This will be called from the service layer which has access to features package
	// For now, keep the basic calculation
	c.AC = 10

	// First, check for body armor which sets the base AC
	if bodyArmor := c.EquippedSlots[shared.SlotBody]; bodyArmor != nil {
		if bodyArmor.GetEquipmentType() == equipment.EquipmentTypeArmor {
			armor, ok := bodyArmor.(*equipment.Armor)
			if !ok {
				log.Printf("Invalid body armor: %v", bodyArmor)
			}
			if armor.ArmorClass != nil {
				c.AC = armor.ArmorClass.Base
				if armor.ArmorClass.DexBonus {
					// TODO: load max and bonus and limit id applicable
					c.AC += c.Attributes[shared.AttributeDexterity].Bonus
				}
			}
		}
	}

	// Then add bonuses from other armor pieces (like shields)
	for slot, e := range c.EquippedSlots {
		if e == nil || slot == shared.SlotBody {
			continue
		}

		if e.GetEquipmentType() == equipment.EquipmentTypeArmor {
			armor, ok := e.(*equipment.Armor)
			if !ok {
				continue
			}
			if armor.ArmorClass == nil {
				continue
			}
			c.AC += armor.ArmorClass.Base
			if armor.ArmorClass.DexBonus {
				c.AC += c.Attributes[shared.AttributeDexterity].Bonus
			}
		}
	}

	// Apply fighting style bonuses to AC
	c.applyFightingStyleAC()
}

// applyFightingStyleAC applies AC bonuses from fighting styles
func (c *Character) applyFightingStyleAC() {
	// Check if the character has fighting style feature with defense
	for _, feature := range c.Features {
		if feature.Key == "fighting_style" && feature.Metadata != nil {
			if style, ok := feature.Metadata["style"].(string); ok && style == "defense" {
				// Defense fighting style gives +1 AC while wearing armor
				// Check if character is wearing any armor
				if c.EquippedSlots != nil && c.EquippedSlots[shared.SlotBody] != nil {
					if c.EquippedSlots[shared.SlotBody].GetEquipmentType() == equipment.EquipmentTypeArmor {
						c.AC += 1
					}
				}
				break
			}
		}
	}
}

func (c *Character) SetHitpoints() {
	if c.Attributes == nil {
		return
	}

	if c.Attributes[shared.AttributeConstitution] == nil {
		return
	}

	if c.HitDie == 0 {
		return
	}

	c.MaxHitPoints = c.HitDie + c.Attributes[shared.AttributeConstitution].Bonus
	c.CurrentHitPoints = c.MaxHitPoints
}

// getCharismaModifier returns the character's Charisma modifier
func (c *Character) getCharismaModifier() int {
	if c.Attributes == nil {
		return 0
	}
	if cha, exists := c.Attributes[shared.AttributeCharisma]; exists && cha != nil {
		return cha.Bonus
	}
	return 0
}

// initializeClassAbilities sets up level 1 class abilities
func (c *Character) initializeClassAbilities() {
	if c.Class == nil || c.Resources == nil {
		return
	}

	// Initialize abilities map if needed
	if c.Resources.Abilities == nil {
		c.Resources.Abilities = make(map[string]*shared.ActiveAbility)
	}

	// Add class-specific abilities based on class key
	switch c.Class.Key {
	case "barbarian":
		c.Resources.Abilities[shared.AbilityKeyRage] = &shared.ActiveAbility{
			Key:           shared.AbilityKeyRage,
			Name:          "Rage",
			Description:   "Enter a battle fury gaining damage bonus and resistance",
			FeatureKey:    "barbarian-rage",
			ActionType:    shared.AbilityTypeBonusAction,
			UsesMax:       2, // 2 uses at level 1
			UsesRemaining: 2,
			RestType:      shared.RestTypeLong,
			Duration:      10, // 10 rounds (1 minute)
		}
	case "fighter":
		c.Resources.Abilities[shared.AbilityKeySecondWind] = &shared.ActiveAbility{
			Key:           shared.AbilityKeySecondWind,
			Name:          "Second Wind",
			Description:   "Regain hit points equal to 1d10 + fighter level",
			FeatureKey:    "fighter-second-wind",
			ActionType:    shared.AbilityTypeBonusAction,
			UsesMax:       1,
			UsesRemaining: 1,
			RestType:      shared.RestTypeShort,
			Duration:      0, // Instant effect
		}
	case "bard":
		c.Resources.Abilities[shared.AbilityKeyBardicInspiration] = &shared.ActiveAbility{
			Key:           shared.AbilityKeyBardicInspiration,
			Name:          "Bardic Inspiration",
			Description:   "Grant an ally a d6 to add to one ability check, attack roll, or saving throw",
			FeatureKey:    "bard-bardic-inspiration",
			ActionType:    shared.AbilityTypeBonusAction,
			UsesMax:       c.getCharismaModifier(), // Uses equal to Charisma modifier
			UsesRemaining: c.getCharismaModifier(),
			RestType:      shared.RestTypeLong,
			Duration:      10, // 10 minutes (100 rounds), but usually consumed on use
		}
		// Initialize bard spells
		c.initializeBardSpells()
	case "paladin":
		c.Resources.Abilities[shared.AbilityKeyLayOnHands] = &shared.ActiveAbility{
			Key:           shared.AbilityKeyLayOnHands,
			Name:          "Lay on Hands",
			Description:   "Heal wounds with a pool of hit points equal to 5 × paladin level",
			FeatureKey:    "paladin-lay-on-hands",
			ActionType:    shared.AbilityTypeAction,
			UsesMax:       5 * c.Level, // 5 HP per level
			UsesRemaining: 5 * c.Level,
			RestType:      shared.RestTypeLong,
			Duration:      0, // Instant effect
		}
		c.Resources.Abilities[shared.AbilityKeyDivineSense] = &shared.ActiveAbility{
			Key:           shared.AbilityKeyDivineSense,
			Name:          "Divine Sense",
			Description:   "Detect celestials, fiends, and undead within 60 feet",
			FeatureKey:    "paladin-divine-sense",
			ActionType:    shared.AbilityTypeAction,
			UsesMax:       1 + c.getCharismaModifier(), // 1 + Charisma modifier
			UsesRemaining: 1 + c.getCharismaModifier(),
			RestType:      shared.RestTypeLong,
			Duration:      0, // Until end of next turn
		}
	case "monk":
		// Monk abilities like Flurry of Blows will come with Ki points later
	case "rogue":
		// Sneak Attack is passive, doesn't use resources
	case "ranger":
		// Rangers get passive features that we'll add as permanent status effects
		// Check if favored enemy is already selected (from character creation)
		var favoredEnemyType string
		for _, feature := range c.Features {
			if feature.Key == "favored_enemy" && feature.Metadata != nil {
				if enemyType, ok := feature.Metadata["enemy_type"].(string); ok {
					favoredEnemyType = enemyType
					break
				}
			}
		}

		// If we have a favored enemy selection, apply the effect
		if favoredEnemyType != "" {
			favoredEnemyEffect := effects.BuildFavoredEnemyEffect(favoredEnemyType)
			if err := c.addStatusEffectInternal(favoredEnemyEffect); err == nil {
				log.Printf("Added Favored Enemy effect for ranger: %s", favoredEnemyType)
			}
		}

		// Natural Explorer is also passive but doesn't have mechanical effects yet
		// TODO: Implement Natural Explorer when terrain tracking is added
	}
}

func (c *Character) AddAttribute(attr shared.Attribute, score int) {
	if c.Attributes == nil {
		c.Attributes = make(map[shared.Attribute]*AbilityScore)
	}

	// Calculate the modifier based on the score
	modifier := (score - 10) / 2

	abilityScore := &AbilityScore{
		Score: score,
		Bonus: modifier,
	}

	c.Attributes[attr] = abilityScore
}
func (c *Character) AddAbilityBonus(ab *shared.AbilityBonus) {
	if c.Attributes == nil {
		c.Attributes = make(map[shared.Attribute]*AbilityScore)
	}

	if _, ok := c.Attributes[ab.Attribute]; !ok {
		c.Attributes[ab.Attribute] = &AbilityScore{Score: 0, Bonus: 0}
	}

	c.Attributes[ab.Attribute] = c.Attributes[ab.Attribute].AddBonus(ab.Bonus)
}

func (c *Character) AddInventory(e equipment.Equipment) {
	if c.Inventory == nil {
		c.Inventory = make(map[equipment.EquipmentType][]equipment.Equipment)
	}

	c.mu.Lock()
	if c.Inventory[e.GetEquipmentType()] == nil {
		c.Inventory[e.GetEquipmentType()] = make([]equipment.Equipment, 0)
	}

	c.Inventory[e.GetEquipmentType()] = append(c.Inventory[e.GetEquipmentType()], e)
	c.mu.Unlock()
}

func (c *Character) AddAbilityScoreBonus(attr shared.Attribute, bonus int) {
	if c.Attributes == nil {
		c.Attributes = make(map[shared.Attribute]*AbilityScore)
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

// IsComplete checks if the character has all required fields
func (c *Character) IsComplete() bool {
	return c != nil && c.Race != nil && c.Class != nil && c.Name != "" && len(c.Attributes) > 0
}

// GetDisplayInfo returns a safe string representation of race and class
func (c *Character) GetDisplayInfo() string {
	if c == nil {
		return "Unknown Character"
	}

	if c.Race != nil && c.Class != nil {
		return fmt.Sprintf("%s %s", c.Race.Name, c.Class.Name)
	} else if c.Race != nil {
		return c.Race.Name
	} else if c.Class != nil {
		return c.Class.Name
	}
	return "Incomplete Character"
}

func (c *Character) String() string {
	msg := strings.Builder{}
	if !c.IsComplete() {
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

	// Add features section
	if len(c.Features) > 0 {
		msg.WriteString("\n**Features**:\n")
		for _, feat := range c.Features {
			msg.WriteString(fmt.Sprintf("  - **%s**: %s\n", feat.Name, feat.Description))
		}
	}

	msg.WriteString("\n**Attributes**:\n")
	for _, attr := range shared.Attributes {
		if c.Attributes[attr] == nil {
			continue
		}
		msg.WriteString(fmt.Sprintf("  -  %s: %s\n", attr, c.Attributes[attr]))
	}

	msg.WriteString("\n**Proficiencies**:\n")
	for _, key := range rulebook.ProficiencyTypes {
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

func (c *Character) IsEquipped(e equipment.Equipment) bool {
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
	c.Proficiencies = make(map[rulebook.ProficiencyType][]*rulebook.Proficiency)
	c.resetAbilityScores()
}

func (c *Character) resetClassFeatures() {
	// Clear all downstream data
	c.Proficiencies = make(map[rulebook.ProficiencyType][]*rulebook.Proficiency)
	c.Inventory = make(map[equipment.EquipmentType][]equipment.Equipment)
	c.EquippedSlots = make(map[shared.Slot]equipment.Equipment)
}

func (c *Character) resetBackground() {
	c.Background = nil
	// Clear all downstream data
	c.Proficiencies = make(map[rulebook.ProficiencyType][]*rulebook.Proficiency)
	c.Inventory = make(map[equipment.EquipmentType][]equipment.Equipment)
}

func (c *Character) resetAbilityScores() {
	c.Attributes = make(map[shared.Attribute]*AbilityScore)
}

// Clone creates a deep copy of the character without copying the mutex
func (c *Character) Clone() *Character {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Create a new character with all fields except mutex
	clone := &Character{
		ID:               c.ID,
		OwnerID:          c.OwnerID,
		RealmID:          c.RealmID,
		Name:             c.Name,
		Level:            c.Level,
		Experience:       c.Experience,
		CurrentHitPoints: c.CurrentHitPoints,
		// TemporaryHitPoints: c.TemporaryHitPoints,
		MaxHitPoints: c.MaxHitPoints,
		AC:           c.AC,
		HitDie:       c.HitDie,
		Speed:        c.Speed,
		// Initiative:        c.Initiative,
		// PassivePerception: c.PassivePerception,
		// ProficiencyBonus:  c.ProficiencyBonus,
		Status:     c.Status,
		Background: c.Background,
		// Alignment:         c.Alignment,
		// Age:               c.Age,
		// Height:            c.Height,
		// Weight:            c.Weight,
		// Eyes:              c.Eyes,
		// Skin:              c.Skin,
		// Hair:              c.Hair,
		// Backstory:         c.Backstory,
		// Portrait:          c.Portrait,
		// Note: mu sync.Mutex is not copied - new instance gets its own
	}

	// Deep copy Race
	if c.Race != nil {
		raceCopy := *c.Race
		clone.Race = &raceCopy
	}

	// Deep copy Class
	if c.Class != nil {
		classCopy := *c.Class
		clone.Class = &classCopy
	}

	// Deep copy Attributes map
	clone.Attributes = make(map[shared.Attribute]*AbilityScore)
	for k, v := range c.Attributes {
		if v != nil {
			scoreCopy := *v
			clone.Attributes[k] = &scoreCopy
		}
	}

	// Deep copy Inventory map
	clone.Inventory = make(map[equipment.EquipmentType][]equipment.Equipment)
	for k, v := range c.Inventory {
		if v != nil {
			clone.Inventory[k] = append([]equipment.Equipment(nil), v...)
		}
	}

	// Deep copy Proficiencies map
	clone.Proficiencies = make(map[rulebook.ProficiencyType][]*rulebook.Proficiency)
	for k, v := range c.Proficiencies {
		if v != nil {
			profCopy := make([]*rulebook.Proficiency, len(v))
			for i, prof := range v {
				if prof != nil {
					p := *prof
					profCopy[i] = &p
				}
			}
			clone.Proficiencies[k] = profCopy
		}
	}

	// Deep copy Features slice
	if c.Features != nil {
		clone.Features = make([]*rulebook.CharacterFeature, len(c.Features))
		for i, feat := range c.Features {
			if feat != nil {
				f := *feat
				clone.Features[i] = &f
			}
		}
	}

	// Deep copy Languages slice
	// if c.Languages != nil {
	// 	clone.Languages = append([]Language(nil), c.Languages...)
	// }

	// Deep copy Skills map
	// clone.Skills = make(map[SkillType]*Skill)
	// for k, v := range c.Skills {
	// 	if v != nil {
	// 		skillCopy := *v
	// 		clone.Skills[k] = &skillCopy
	// 	}
	// }

	// Deep copy EquippedSlots map
	clone.EquippedSlots = make(map[shared.Slot]equipment.Equipment)
	for k, v := range c.EquippedSlots {
		clone.EquippedSlots[k] = v
	}

	// Deep copy Notes slice
	// if c.Notes != nil {
	// 	clone.Notes = append([]string(nil), c.Notes...)
	// }

	// Deep copy Resources
	if c.Resources != nil {
		clone.Resources = &CharacterResources{
			HP:      c.Resources.HP,      // HPResource is a value type
			HitDice: c.Resources.HitDice, // Also a value type
		}

		// Deep copy spell slots
		if c.Resources.SpellSlots != nil {
			clone.Resources.SpellSlots = make(map[int]shared.SpellSlotInfo)
			for level, slot := range c.Resources.SpellSlots {
				clone.Resources.SpellSlots[level] = slot
			}
		}

		// Deep copy abilities
		if c.Resources.Abilities != nil {
			clone.Resources.Abilities = make(map[string]*shared.ActiveAbility)
			for key, ability := range c.Resources.Abilities {
				if ability != nil {
					abilityCopy := *ability
					clone.Resources.Abilities[key] = &abilityCopy
				}
			}
		}

		// Deep copy active effects
		if c.Resources.ActiveEffects != nil {
			clone.Resources.ActiveEffects = make([]*shared.ActiveEffect, len(c.Resources.ActiveEffects))
			for i, effect := range c.Resources.ActiveEffects {
				if effect != nil {
					effectCopy := *effect
					// Deep copy modifiers
					if effect.Modifiers != nil {
						effectCopy.Modifiers = make([]shared.Modifier, len(effect.Modifiers))
						copy(effectCopy.Modifiers, effect.Modifiers)
					}
					clone.Resources.ActiveEffects[i] = &effectCopy
				}
			}
		}
	}

	return clone
}

// GetSkillBonus calculates the total skill bonus
func (c *Character) GetSkillBonus(skillKey string, attribute shared.Attribute) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Get ability modifier
	modifier := 0
	if abilityScore, exists := c.Attributes[attribute]; exists && abilityScore != nil {
		modifier = abilityScore.Bonus
	}

	// Check proficiency without locking again
	isProficient := false
	if c.Proficiencies != nil && c.Proficiencies[rulebook.ProficiencyTypeSkill] != nil {
		for _, prof := range c.Proficiencies[rulebook.ProficiencyTypeSkill] {
			if strings.EqualFold(prof.Key, skillKey) {
				isProficient = true
				break
			}
		}
	}

	// Add proficiency bonus if proficient
	if isProficient {
		modifier += c.GetProficiencyBonus()
	}

	return modifier
}

// RollSkillCheck rolls a skill check
func (c *Character) RollSkillCheck(skillKey string, attribute shared.Attribute) (*dice.RollResult, int, error) {
	bonus := c.GetSkillBonus(skillKey, attribute)

	// Roll 1d20
	result, err := c.getDiceRoller().Roll(1, 20, bonus)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to roll skill check: %w", err)
	}

	return result, result.Total, nil
}

// applyFightingStyleBonuses applies bonuses from fighter fighting styles
func (c *Character) applyFightingStyleBonuses(weapon *equipment.Weapon, attackBonus, damageBonus int) (finalAttackBonus, finalDamageBonus int) {
	return c.applyFightingStyleBonusesWithHand(weapon, attackBonus, damageBonus, shared.SlotMainHand)
}

// getFightingStyle returns the character's fighting style if they have one
func (c *Character) getFightingStyle() string {
	for _, feature := range c.Features {
		if feature.Key == "fighting_style" && feature.Metadata != nil {
			if style, ok := feature.Metadata["style"].(string); ok {
				return style
			}
		}
	}
	return ""
}

// applyFightingStyleBonusesWithHand applies bonuses from fighter fighting styles for a specific hand
func (c *Character) applyFightingStyleBonusesWithHand(weapon *equipment.Weapon, attackBonus, damageBonus int, hand shared.Slot) (finalAttackBonus, finalDamageBonus int) {
	// Check if the character has fighting style feature
	fightingStyle := c.getFightingStyle()
	log.Printf("DEBUG: Checking for fighting style among %d features", len(c.Features))

	if fightingStyle == "" {
		log.Printf("DEBUG: No fighting style found")
		return attackBonus, damageBonus
	}
	log.Printf("DEBUG: Applying fighting style: %s", fightingStyle)

	// Apply fighting style bonuses
	switch fightingStyle {
	case "archery": // TODO: these keys should be a typed constant
		// +2 to attack rolls with ranged weapons
		if weapon.IsRanged() {
			attackBonus += 2
		}
	case "defense":
		// +1 to AC while wearing armor (handled elsewhere in AC calculation)
		// No attack/damage bonus
	case "dueling":
		// +2 damage with one-handed weapons when no other weapon equipped (shields are OK)
		log.Printf("DEBUG: Checking dueling for weapon %s, IsMelee=%v, IsTwoHanded=%v", weapon.GetName(), weapon.IsMelee(), weapon.IsTwoHanded())
		if weapon.IsMelee() && !weapon.IsTwoHanded() {
			// Check if off-hand has a weapon (shields are allowed)
			offHand := c.EquippedSlots[shared.SlotOffHand]
			log.Printf("DEBUG: Off-hand equipment: %v", offHand)

			offHandHasWeapon := false
			if offHand != nil {
				_, isWeapon := offHand.(*equipment.Weapon)
				offHandHasWeapon = isWeapon
				log.Printf("DEBUG: Off-hand has weapon: %v (type: %T)", offHandHasWeapon, offHand)
			}

			if !offHandHasWeapon {
				log.Printf("DEBUG: Applying dueling bonus +2")
				damageBonus += 2
			} else {
				log.Printf("DEBUG: Not applying dueling bonus - weapon in off-hand")
			}
		} else {
			log.Printf("DEBUG: Not applying dueling - weapon not melee one-handed")
		}
	case "great_weapon":
		// Reroll 1s and 2s on damage with two-handed weapons
		// This is handled in damage rolling, not as a flat bonus
	case "protection":
		// Reaction ability to impose disadvantage (handled elsewhere)
		// No attack/damage bonus
	case "two_weapon":
		// Add ability modifier to off-hand damage
		if hand == shared.SlotOffHand && weapon.IsMelee() {
			// Two-weapon fighting allows adding ability modifier to off-hand damage
			// The base off-hand damage calculation doesn't include ability modifier
			// So we need to add it here
			if c.Attributes != nil && c.Attributes[shared.AttributeStrength] != nil {
				damageBonus = c.Attributes[shared.AttributeStrength].Bonus
			}
		}
	}

	return attackBonus, damageBonus
}

// initializeBardSpells adds default bard cantrips
func (c *Character) initializeBardSpells() {
	// For now, just add vicious mockery as a default cantrip
	// In a full implementation, this would be part of character creation choices
	c.AddCantrip(shared.SpellKeyViciousMockery)

	// We could also add it as an ability that references the spell
	// This maintains backward compatibility while we transition
	c.Resources.Abilities[shared.AbilityKeyViciousMockery] = &shared.ActiveAbility{
		Key:           shared.AbilityKeyViciousMockery,
		Name:          "Vicious Mockery",
		Description:   "Unleash a string of insults laced with subtle enchantments",
		FeatureKey:    "bard-vicious-mockery",
		ActionType:    shared.AbilityTypeAction,
		UsesMax:       -1, // Unlimited (cantrip)
		UsesRemaining: -1,
		RestType:      shared.RestTypeNone,
		Targeting: &shared.AbilityTargeting{
			TargetType: shared.TargetTypeSingleEnemy,
			RangeType:  shared.RangeTypeRanged,
			Range:      60,
			Components: []shared.ComponentType{shared.ComponentVerbal},
			SaveType:   shared.AttributeWisdom,
		},
	}
}
