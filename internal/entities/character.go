package entities

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/KirkDiggler/dnd-bot-discord/internal/dice"
	"github.com/KirkDiggler/dnd-bot-discord/internal/effects"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities/attack"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities/damage"
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
	AbilityRolls       []AbilityRoll     // New field for ability score rolls with IDs
	AbilityAssignments map[string]string // Maps ability name (STR, DEX, etc.) to roll ID
	Proficiencies      map[ProficiencyType][]*Proficiency
	ProficiencyChoices []*Choice
	Inventory          map[EquipmentType][]Equipment
	Features           []*CharacterFeature // Character features (class, racial, etc.)

	HitDie           int
	AC               int
	MaxHitPoints     int
	CurrentHitPoints int
	Level            int
	Experience       int
	NextLevel        int

	EquippedSlots map[Slot]Equipment

	Status CharacterStatus `json:"status"`

	// Resources tracks HP, abilities, spell slots, etc
	Resources *CharacterResources `json:"resources"`

	// EffectManager tracks all active status effects
	EffectManager *effects.Manager `json:"-"`

	mu sync.Mutex
}

func (c *Character) Attack() ([]*attack.Result, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.EquippedSlots == nil {
		// Improvised weapon range or melee
		// No equipped slots, using improvised melee
		a, err := c.improvisedMelee()
		if err != nil {
			return nil, err
		}

		return []*attack.Result{
			a,
		}, nil

	}

	// Check main hand slot
	if c.EquippedSlots[SlotMainHand] != nil {
		if weap, ok := c.EquippedSlots[SlotMainHand].(*Weapon); ok {
			attacks := make([]*attack.Result, 0)

			// Check proficiency while we have the mutex
			directProf := c.hasWeaponProficiencyInternal(weap.GetKey())
			categoryProf := c.hasWeaponCategoryProficiency(weap.WeaponCategory)
			isProficient := directProf || categoryProf
			// Proficiency check completed

			// Calculate ability bonus based on weapon type
			var abilityBonus int
			if c.Attributes != nil {
				switch weap.WeaponRange {
				case "Ranged":
					if c.Attributes[AttributeDexterity] != nil {
						abilityBonus = c.Attributes[AttributeDexterity].Bonus
					}
				case "Melee":
					if c.Attributes[AttributeStrength] != nil {
						abilityBonus = c.Attributes[AttributeStrength].Bonus
					}
				}
			}

			// Calculate proficiency bonus if proficient
			proficiencyBonus := 0
			if isProficient {
				proficiencyBonus = 2 + ((c.Level - 1) / 4)
			}

			attackBonus := abilityBonus + proficiencyBonus
			damageBonus := abilityBonus // Base damage bonus from ability modifier

			// Apply fighting style bonuses
			attackBonus, damageBonus = c.applyFightingStyleBonuses(weap, attackBonus, damageBonus)

			// Apply damage bonuses from active effects (e.g., rage)
			// Use the weapon's actual range type
			attackType := strings.ToLower(weap.WeaponRange)
			var err error
			damageBonus, err = c.applyActiveEffectDamageBonus(damageBonus, attackType)
			if err != nil {
				log.Printf("ERROR: Failed to apply active effect damage bonus: %v", err)
				// Continue with base damage bonus
			}

			log.Printf("Final attack bonus: +%d (ability: %d, proficiency: %d)", attackBonus, abilityBonus, proficiencyBonus)
			log.Printf("Final damage bonus: +%d", damageBonus)

			// Roll the attack
			var attak1 *attack.Result
			if weap.IsTwoHanded() && weap.TwoHandedDamage != nil {
				attak1, err = attack.RollAttack(attackBonus, damageBonus, weap.TwoHandedDamage)
			} else {
				attak1, err = attack.RollAttack(attackBonus, damageBonus, weap.Damage)
			}

			if err != nil {
				log.Printf("Weapon attack error: %v", err)
				return nil, err
			}
			log.Printf("Weapon attack successful")
			attacks = append(attacks, attak1)

			if c.EquippedSlots[SlotOffHand] != nil {
				if offWeap, offOk := c.EquippedSlots[SlotOffHand].(*Weapon); offOk {
					// Same process for off-hand weapon
					offHandProficient := c.hasWeaponProficiencyInternal(offWeap.GetKey()) ||
						c.hasWeaponCategoryProficiency(offWeap.WeaponCategory)

					var offHandAbilityBonus int
					switch offWeap.WeaponRange {
					case "Ranged":
						offHandAbilityBonus = c.Attributes[AttributeDexterity].Bonus
					case "Melee":
						offHandAbilityBonus = c.Attributes[AttributeStrength].Bonus
					}

					offHandProficiencyBonus := 0
					if offHandProficient {
						offHandProficiencyBonus = 2 + ((c.Level - 1) / 4)
					}

					offHandAttackBonus := offHandAbilityBonus + offHandProficiencyBonus
					offHandDamageBonus := offHandAbilityBonus

					// Apply fighting style bonuses to off-hand
					offHandAttackBonus, offHandDamageBonus = c.applyFightingStyleBonusesWithHand(offWeap, offHandAttackBonus, offHandDamageBonus, SlotOffHand)

					// Apply damage bonuses from active effects (e.g., rage) to off-hand
					offHandDamageBonus, err = c.applyActiveEffectDamageBonus(offHandDamageBonus, "melee")
					if err != nil {
						log.Printf("ERROR: Failed to apply active effect damage bonus to off-hand: %v", err)
						// Continue with current bonus
					}

					attak2, err := attack.RollAttack(offHandAttackBonus, offHandDamageBonus, offWeap.Damage)
					if err != nil {
						return nil, err
					}
					attacks = append(attacks, attak2)
				}
			}

			log.Printf("Returning %d attack results", len(attacks))
			return attacks, nil
		} else {
			log.Printf("Main hand equipment is not a weapon: %T", c.EquippedSlots[SlotMainHand])
		}
	} else {
		log.Printf("No main hand equipment")
	}

	if c.EquippedSlots[SlotTwoHanded] != nil {
		log.Printf("Checking two-handed slot...")
		log.Printf("Two-handed slot type: %T", c.EquippedSlots[SlotTwoHanded])
		if weap, ok := c.EquippedSlots[SlotTwoHanded].(*Weapon); ok {
			log.Printf("Two-handed weapon found: %s", weap.GetName())
			// Check proficiency while we have the mutex
			directProf := c.hasWeaponProficiencyInternal(weap.GetKey())
			categoryProf := c.hasWeaponCategoryProficiency(weap.WeaponCategory)
			isProficient := directProf || categoryProf
			// Proficiency check completed

			// Calculate ability bonus based on weapon type
			var abilityBonus int
			if c.Attributes != nil {
				switch weap.WeaponRange {
				case "Ranged":
					if c.Attributes[AttributeDexterity] != nil {
						abilityBonus = c.Attributes[AttributeDexterity].Bonus
					}
				case "Melee":
					if c.Attributes[AttributeStrength] != nil {
						abilityBonus = c.Attributes[AttributeStrength].Bonus
					}
				}
			}

			// Calculate proficiency bonus if proficient
			proficiencyBonus := 0
			if isProficient {
				proficiencyBonus = 2 + ((c.Level - 1) / 4)
			}

			attackBonus := abilityBonus + proficiencyBonus
			damageBonus := abilityBonus

			// Apply fighting style bonuses
			attackBonus, damageBonus = c.applyFightingStyleBonuses(weap, attackBonus, damageBonus)

			// Apply damage bonuses from active effects (e.g., rage)
			// Use the weapon's actual range type
			attackType := strings.ToLower(weap.WeaponRange)
			var err error
			damageBonus, err = c.applyActiveEffectDamageBonus(damageBonus, attackType)
			if err != nil {
				log.Printf("ERROR: Failed to apply active effect damage bonus: %v", err)
				// Continue with base damage bonus
			}

			log.Printf("Final attack bonus: +%d (ability: %d, proficiency: %d)", attackBonus, abilityBonus, proficiencyBonus)
			log.Printf("Final damage bonus: +%d", damageBonus)

			// Two-handed weapons often have special damage
			var dmg *damage.Damage
			if weap.TwoHandedDamage != nil {
				dmg = weap.TwoHandedDamage
			} else {
				dmg = weap.Damage
			}

			a, err := attack.RollAttack(attackBonus, damageBonus, dmg)
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

// HasWeaponProficiency checks if the character is proficient with a weapon (thread-safe)
func (c *Character) HasWeaponProficiency(weaponKey string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	// First check direct weapon proficiency
	if c.hasWeaponProficiencyInternal(weaponKey) {
		return true
	}

	// Check if we have the weapon to get its category
	weapon := c.getEquipment(weaponKey)
	if weapon != nil {
		if w, ok := weapon.(*Weapon); ok && w.WeaponCategory != "" {
			return c.hasWeaponCategoryProficiency(w.WeaponCategory)
		}
	}

	// Also check equipped weapons
	for _, equipped := range c.EquippedSlots {
		if equipped != nil && equipped.GetKey() == weaponKey {
			if w, ok := equipped.(*Weapon); ok && w.WeaponCategory != "" {
				return c.hasWeaponCategoryProficiency(w.WeaponCategory)
			}
		}
	}

	return false
}

// hasWeaponProficiencyInternal checks proficiency without locking (must be called with mutex held)
func (c *Character) hasWeaponProficiencyInternal(weaponKey string) bool {
	if c.Proficiencies == nil {
		return false
	}

	weaponProficiencies, exists := c.Proficiencies[ProficiencyTypeWeapon]
	if !exists {
		return false
	}

	for _, prof := range weaponProficiencies {
		if prof.Key == weaponKey {
			return true
		}
	}

	return false
}

// hasWeaponCategoryProficiency checks if the character has proficiency with a weapon category
// This handles proficiencies like "simple-weapons" or "martial-weapons"
func (c *Character) hasWeaponCategoryProficiency(weaponCategory string) bool {
	if c.Proficiencies == nil || weaponCategory == "" {
		return false
	}

	weaponProficiencies, exists := c.Proficiencies[ProficiencyTypeWeapon]
	if !exists {
		return false
	}

	// Map weapon categories to proficiency keys (case-insensitive)
	categoryMap := map[string]string{
		"simple":  "simple-weapons",
		"martial": "martial-weapons",
	}

	// Convert to lowercase for case-insensitive comparison
	lowerCategory := strings.ToLower(weaponCategory)
	profKey, exists := categoryMap[lowerCategory]
	if !exists {
		return false
	}

	for _, prof := range weaponProficiencies {
		if prof.Key == profKey {
			return true
		}
	}

	return false
}

func (c *Character) improvisedMelee() (*attack.Result, error) {
	bonus := 0
	if c.Attributes != nil && c.Attributes[AttributeStrength] != nil {
		bonus = c.Attributes[AttributeStrength].Bonus
	}

	// Apply damage bonuses from active effects (e.g., rage) to improvised attacks
	damageBonus, err := c.applyActiveEffectDamageBonus(bonus, "melee")
	if err != nil {
		log.Printf("ERROR: Failed to apply active effect damage bonus to improvised: %v", err)
		damageBonus = bonus // Fall back to base
	}

	attackRoll, err := dice.Roll(1, 20, 0)
	if err != nil {
		return nil, err
	}
	damageRoll, err := dice.Roll(1, 4, 0)
	if err != nil {
		return nil, err
	}

	return &attack.Result{
		AttackRoll:   attackRoll.Total + bonus,
		DamageRoll:   damageRoll.Total + damageBonus,
		AttackType:   damage.TypeBludgeoning,
		AttackResult: attackRoll,
		DamageResult: damageRoll,
		WeaponDamage: &damage.Damage{
			DiceCount:  1,
			DiceSize:   4,
			Bonus:      0,
			DamageType: damage.TypeBludgeoning,
		},
	}, nil
}

// applyActiveEffectDamageBonus applies damage bonuses from active effects like rage
// Returns the modified damage and any error encountered
func (c *Character) applyActiveEffectDamageBonus(baseDamage int, damageType string) (int, error) {
	// Get damage modifiers from the new status effect system
	conditions := map[string]string{
		"attack_type": damageType,
	}

	modifiers := c.getDamageModifiersInternal(conditions)
	effectBonus := 0

	for _, mod := range modifiers {
		// Parse modifier value (e.g., "+2", "+3", "+4")
		if mod.Value != "" && mod.Value[0] == '+' {
			var parsedBonus int
			if n, err := fmt.Sscanf(mod.Value, "+%d", &parsedBonus); err == nil && n == 1 {
				effectBonus += parsedBonus
			}
		}
	}

	baseDamage += effectBonus

	// Also check the old system for backward compatibility
	if c.Resources != nil {
		oldEffectBonus := c.Resources.GetTotalDamageBonus(damageType)
		baseDamage += oldEffectBonus
	}

	return baseDamage, nil
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
	// This will be called from the service layer which has access to features package
	// For now, keep the basic calculation
	c.AC = 10

	// First, check for body armor which sets the base AC
	if bodyArmor := c.EquippedSlots[SlotBody]; bodyArmor != nil {
		if bodyArmor.GetEquipmentType() == EquipmentTypeArmor {
			armor, ok := bodyArmor.(*Armor)
			if !ok {
				log.Printf("Invalid body armor: %v", bodyArmor)
			}
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

		if e.GetEquipmentType() == EquipmentTypeArmor {
			armor, ok := e.(*Armor)
			if !ok {
				continue
			}
			if armor.ArmorClass == nil {
				continue
			}
			c.AC += armor.ArmorClass.Base
			if armor.ArmorClass.DexBonus {
				c.AC += c.Attributes[AttributeDexterity].Bonus
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
				if c.EquippedSlots != nil && c.EquippedSlots[SlotBody] != nil {
					if c.EquippedSlots[SlotBody].GetEquipmentType() == EquipmentTypeArmor {
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

	if c.Attributes[AttributeConstitution] == nil {
		return
	}

	if c.HitDie == 0 {
		return
	}

	c.MaxHitPoints = c.HitDie + c.Attributes[AttributeConstitution].Bonus
	c.CurrentHitPoints = c.MaxHitPoints
}

// GetResources returns the character's resources, initializing if needed
func (c *Character) GetResources() *CharacterResources {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Resources == nil {
		c.initializeResourcesInternal()
	}
	return c.Resources
}

// InitializeResources sets up the character's resources based on class and level
func (c *Character) InitializeResources() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.initializeResourcesInternal()
}

// initializeResourcesInternal is the internal resource initialization (caller must hold lock)
func (c *Character) initializeResourcesInternal() {
	if c.Resources == nil {
		c.Resources = &CharacterResources{}
	}

	// Initialize basic resources
	if c.Class != nil {
		c.Resources.Initialize(c.Class, c.Level)
	}

	// Set HP based on character's max HP (includes CON bonus)
	c.Resources.HP = HPResource{
		Current: c.CurrentHitPoints,
		Max:     c.MaxHitPoints,
	}

	// Initialize class-specific abilities at level 1
	c.initializeClassAbilities()
}

// getCharismaModifier returns the character's Charisma modifier
func (c *Character) getCharismaModifier() int {
	if c.Attributes == nil {
		return 0
	}
	if cha, exists := c.Attributes[AttributeCharisma]; exists && cha != nil {
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
		c.Resources.Abilities = make(map[string]*ActiveAbility)
	}

	// Add class-specific abilities based on class key
	switch c.Class.Key {
	case "barbarian":
		c.Resources.Abilities["rage"] = &ActiveAbility{
			Key:           "rage",
			Name:          "Rage",
			Description:   "Enter a battle fury gaining damage bonus and resistance",
			FeatureKey:    "barbarian-rage",
			ActionType:    AbilityTypeBonusAction,
			UsesMax:       2, // 2 uses at level 1
			UsesRemaining: 2,
			RestType:      RestTypeLong,
			Duration:      10, // 10 rounds (1 minute)
		}
	case "fighter":
		c.Resources.Abilities["second-wind"] = &ActiveAbility{
			Key:           "second-wind",
			Name:          "Second Wind",
			Description:   "Regain hit points equal to 1d10 + fighter level",
			FeatureKey:    "fighter-second-wind",
			ActionType:    AbilityTypeBonusAction,
			UsesMax:       1,
			UsesRemaining: 1,
			RestType:      RestTypeShort,
			Duration:      0, // Instant effect
		}
	case "bard":
		c.Resources.Abilities["bardic-inspiration"] = &ActiveAbility{
			Key:           "bardic-inspiration",
			Name:          "Bardic Inspiration",
			Description:   "Grant an ally a d6 to add to one ability check, attack roll, or saving throw",
			FeatureKey:    "bard-bardic-inspiration",
			ActionType:    AbilityTypeBonusAction,
			UsesMax:       c.getCharismaModifier(), // Uses equal to Charisma modifier
			UsesRemaining: c.getCharismaModifier(),
			RestType:      RestTypeLong,
			Duration:      10, // 10 minutes (100 rounds), but usually consumed on use
		}
	case "paladin":
		c.Resources.Abilities["lay-on-hands"] = &ActiveAbility{
			Key:           "lay-on-hands",
			Name:          "Lay on Hands",
			Description:   "Heal wounds with a pool of hit points equal to 5 × paladin level",
			FeatureKey:    "paladin-lay-on-hands",
			ActionType:    AbilityTypeAction,
			UsesMax:       5 * c.Level, // 5 HP per level
			UsesRemaining: 5 * c.Level,
			RestType:      RestTypeLong,
			Duration:      0, // Instant effect
		}
		c.Resources.Abilities["divine-sense"] = &ActiveAbility{
			Key:           "divine-sense",
			Name:          "Divine Sense",
			Description:   "Detect celestials, fiends, and undead within 60 feet",
			FeatureKey:    "paladin-divine-sense",
			ActionType:    AbilityTypeAction,
			UsesMax:       1 + c.getCharismaModifier(), // 1 + Charisma modifier
			UsesRemaining: 1 + c.getCharismaModifier(),
			RestType:      RestTypeLong,
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

func (c *Character) AddAttribute(attr Attribute, score int) {
	if c.Attributes == nil {
		c.Attributes = make(map[Attribute]*AbilityScore)
	}

	// Calculate the modifier based on the score
	modifier := (score - 10) / 2

	abilityScore := &AbilityScore{
		Score: score,
		Bonus: modifier,
	}

	c.Attributes[attr] = abilityScore
}
func (c *Character) AddAbilityBonus(ab *AbilityBonus) {
	if c.Attributes == nil {
		c.Attributes = make(map[Attribute]*AbilityScore)
	}

	if _, ok := c.Attributes[ab.Attribute]; !ok {
		c.Attributes[ab.Attribute] = &AbilityScore{Score: 0, Bonus: 0}
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
	defer c.mu.Unlock()

	if c.Proficiencies[p.Type] == nil {
		c.Proficiencies[p.Type] = make([]*Proficiency, 0)
	}

	// Check for duplicates
	for _, existing := range c.Proficiencies[p.Type] {
		if existing.Key == p.Key {
			return // Already have this proficiency
		}
	}

	c.Proficiencies[p.Type] = append(c.Proficiencies[p.Type], p)
}

// SetProficiencies replaces all proficiencies of a given type
// This is used when selecting proficiencies during character creation
func (c *Character) SetProficiencies(profType ProficiencyType, proficiencies []*Proficiency) {
	if c.Proficiencies == nil {
		c.Proficiencies = make(map[ProficiencyType][]*Proficiency)
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	// Replace all proficiencies of this type
	c.Proficiencies[profType] = proficiencies
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
	clone.Attributes = make(map[Attribute]*AbilityScore)
	for k, v := range c.Attributes {
		if v != nil {
			scoreCopy := *v
			clone.Attributes[k] = &scoreCopy
		}
	}

	// Deep copy Inventory map
	clone.Inventory = make(map[EquipmentType][]Equipment)
	for k, v := range c.Inventory {
		if v != nil {
			clone.Inventory[k] = append([]Equipment(nil), v...)
		}
	}

	// Deep copy Proficiencies map
	clone.Proficiencies = make(map[ProficiencyType][]*Proficiency)
	for k, v := range c.Proficiencies {
		if v != nil {
			profCopy := make([]*Proficiency, len(v))
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
		clone.Features = make([]*CharacterFeature, len(c.Features))
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
	clone.EquippedSlots = make(map[Slot]Equipment)
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
			clone.Resources.SpellSlots = make(map[int]SpellSlotInfo)
			for level, slot := range c.Resources.SpellSlots {
				clone.Resources.SpellSlots[level] = slot
			}
		}

		// Deep copy abilities
		if c.Resources.Abilities != nil {
			clone.Resources.Abilities = make(map[string]*ActiveAbility)
			for key, ability := range c.Resources.Abilities {
				if ability != nil {
					abilityCopy := *ability
					clone.Resources.Abilities[key] = &abilityCopy
				}
			}
		}

		// Deep copy active effects
		if c.Resources.ActiveEffects != nil {
			clone.Resources.ActiveEffects = make([]*ActiveEffect, len(c.Resources.ActiveEffects))
			for i, effect := range c.Resources.ActiveEffects {
				if effect != nil {
					effectCopy := *effect
					// Deep copy modifiers
					if effect.Modifiers != nil {
						effectCopy.Modifiers = make([]Modifier, len(effect.Modifiers))
						copy(effectCopy.Modifiers, effect.Modifiers)
					}
					clone.Resources.ActiveEffects[i] = &effectCopy
				}
			}
		}
	}

	return clone
}

// GetProficiencyBonus returns the character's proficiency bonus based on level
func (c *Character) GetProficiencyBonus() int {
	if c.Level == 0 {
		return 2 // Default for level 1
	}
	return 2 + ((c.Level - 1) / 4)
}

// HasSavingThrowProficiency checks if the character is proficient in a saving throw
func (c *Character) HasSavingThrowProficiency(attribute Attribute) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Proficiencies == nil || c.Proficiencies[ProficiencyTypeSavingThrow] == nil {
		return false
	}

	// Convert attribute to the expected saving throw key format
	savingThrowKey := fmt.Sprintf("saving-throw-%s", strings.ToLower(string(attribute)))

	for _, prof := range c.Proficiencies[ProficiencyTypeSavingThrow] {
		if strings.EqualFold(prof.Key, savingThrowKey) {
			return true
		}
	}

	return false
}

// GetSavingThrowBonus calculates the total saving throw bonus for an attribute
func (c *Character) GetSavingThrowBonus(attribute Attribute) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Get ability modifier
	modifier := 0
	if abilityScore, exists := c.Attributes[attribute]; exists && abilityScore != nil {
		modifier = abilityScore.Bonus
	}

	// Check proficiency without locking again
	isProficient := false
	if c.Proficiencies != nil && c.Proficiencies[ProficiencyTypeSavingThrow] != nil {
		savingThrowKey := fmt.Sprintf("saving-throw-%s", strings.ToLower(string(attribute)))
		for _, prof := range c.Proficiencies[ProficiencyTypeSavingThrow] {
			if strings.EqualFold(prof.Key, savingThrowKey) {
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

// RollSavingThrow rolls a saving throw for the given attribute
func (c *Character) RollSavingThrow(attribute Attribute) (*dice.RollResult, int, error) {
	bonus := c.GetSavingThrowBonus(attribute)

	// Roll 1d20
	result, err := dice.Roll(1, 20, bonus)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to roll saving throw: %w", err)
	}

	return result, result.Total, nil
}

// HasSkillProficiency checks if the character is proficient in a skill
func (c *Character) HasSkillProficiency(skillKey string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Proficiencies == nil || c.Proficiencies[ProficiencyTypeSkill] == nil {
		return false
	}

	for _, prof := range c.Proficiencies[ProficiencyTypeSkill] {
		if strings.EqualFold(prof.Key, skillKey) {
			return true
		}
	}

	return false
}

// GetSkillBonus calculates the total skill bonus
func (c *Character) GetSkillBonus(skillKey string, attribute Attribute) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Get ability modifier
	modifier := 0
	if abilityScore, exists := c.Attributes[attribute]; exists && abilityScore != nil {
		modifier = abilityScore.Bonus
	}

	// Check proficiency without locking again
	isProficient := false
	if c.Proficiencies != nil && c.Proficiencies[ProficiencyTypeSkill] != nil {
		for _, prof := range c.Proficiencies[ProficiencyTypeSkill] {
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
func (c *Character) RollSkillCheck(skillKey string, attribute Attribute) (*dice.RollResult, int, error) {
	bonus := c.GetSkillBonus(skillKey, attribute)

	// Roll 1d20
	result, err := dice.Roll(1, 20, bonus)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to roll skill check: %w", err)
	}

	return result, result.Total, nil
}

// GetEffectManager returns the character's effect manager, initializing it if needed
func (c *Character) GetEffectManager() *effects.Manager {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.getEffectManagerInternal()
}

// getEffectManagerInternal returns the effect manager without locking (caller must hold lock)
func (c *Character) getEffectManagerInternal() *effects.Manager {
	if c.EffectManager == nil {
		c.EffectManager = effects.NewManager()
		// Restore effects from persisted data
		c.syncResourcestoEffectManager()
	}
	return c.EffectManager
}

// syncResourcestoEffectManager restores EffectManager from persisted ActiveEffects
func (c *Character) syncResourcestoEffectManager() {
	if c.Resources == nil || c.Resources.ActiveEffects == nil {
		return
	}

	// Convert old ActiveEffect format back to new StatusEffect format
	for _, oldEffect := range c.Resources.ActiveEffects {
		if oldEffect == nil {
			continue
		}

		// Create new status effect
		newEffect := &effects.StatusEffect{
			ID:          oldEffect.ID,
			Name:        oldEffect.Name,
			Description: oldEffect.Description,
			Source:      effects.EffectSource(oldEffect.Source),
			SourceID:    oldEffect.SourceID,
			Duration: effects.Duration{
				Rounds: oldEffect.Duration,
			},
			Modifiers:  []effects.Modifier{}, // TODO: Convert modifiers if needed
			Conditions: []effects.Condition{},
			Active:     true,
		}

		// Set duration type
		switch oldEffect.DurationType {
		case DurationTypePermanent:
			newEffect.Duration.Type = effects.DurationPermanent
		case DurationTypeRounds:
			newEffect.Duration.Type = effects.DurationRounds
		case DurationTypeUntilRest:
			newEffect.Duration.Type = effects.DurationUntilRest
		default:
			newEffect.Duration.Type = effects.DurationPermanent
		}

		// For well-known effects, rebuild them properly
		if oldEffect.Name == "Rage" && oldEffect.Source == string(effects.SourceAbility) {
			// Rebuild rage effect with proper modifiers
			rageEffect := effects.BuildRageEffect(c.Level)
			rageEffect.ID = oldEffect.ID // Keep the same ID
			if err := c.EffectManager.AddEffect(rageEffect); err != nil {
				log.Printf("Failed to add rage effect: %v", err)
			}
		} else if oldEffect.Name == "Favored Enemy" && oldEffect.Source == string(effects.SourceFeature) {
			// Rebuild favored enemy effect with the selected enemy type
			enemyType := "humanoids" // default fallback
			for _, feature := range c.Features {
				if feature.Key == "favored_enemy" && feature.Metadata != nil {
					if et, ok := feature.Metadata["enemy_type"].(string); ok {
						enemyType = et
						break
					}
				}
			}
			favoredEnemyEffect := effects.BuildFavoredEnemyEffect(enemyType)
			favoredEnemyEffect.ID = oldEffect.ID
			if err := c.EffectManager.AddEffect(favoredEnemyEffect); err != nil {
				log.Printf("Failed to add favored enemy effect: %v", err)
			}
		} else {
			// Add generic effect
			if err := c.EffectManager.AddEffect(newEffect); err != nil {
				log.Printf("Failed to add effect %s: %v", newEffect.Name, err)
			}
		}
	}
}

// AddStatusEffect adds a new status effect to the character
func (c *Character) AddStatusEffect(effect *effects.StatusEffect) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.addStatusEffectInternal(effect)
}

// addStatusEffectInternal adds a status effect without locking (caller must hold lock)
func (c *Character) addStatusEffectInternal(effect *effects.StatusEffect) error {
	err := c.getEffectManagerInternal().AddEffect(effect)
	if err != nil {
		return err
	}

	// Also sync to the persisted ActiveEffects for backward compatibility
	c.syncEffectManagerToResources()
	return nil
}

// syncEffectManagerToResources syncs EffectManager to the persisted ActiveEffects
func (c *Character) syncEffectManagerToResources() {
	if c.EffectManager == nil || c.Resources == nil {
		return
	}

	// Get active effects from manager
	activeEffects := c.EffectManager.GetActiveEffects()

	// Clear old active effects
	c.Resources.ActiveEffects = []*ActiveEffect{}

	// Convert new status effects to old ActiveEffect format for persistence
	for _, effect := range activeEffects {
		if effect == nil {
			continue
		}

		// Convert to old format (simplified for now)
		oldEffect := &ActiveEffect{
			ID:           effect.ID,
			Name:         effect.Name,
			Description:  effect.Description,
			Source:       string(effect.Source),
			SourceID:     effect.SourceID,
			Duration:     effect.Duration.Rounds,
			DurationType: DurationTypeRounds, // Simplified
			Modifiers:    []Modifier{},       // TODO: Convert modifiers if needed
		}

		// Set duration type based on new system
		switch effect.Duration.Type {
		case effects.DurationPermanent:
			oldEffect.DurationType = DurationTypePermanent
		case effects.DurationRounds:
			oldEffect.DurationType = DurationTypeRounds
		case effects.DurationInstant:
			oldEffect.DurationType = DurationTypeRounds // Instant effects map to rounds
		case effects.DurationWhileEquipped:
			oldEffect.DurationType = DurationTypePermanent
		case effects.DurationUntilRest:
			oldEffect.DurationType = DurationTypeUntilRest
		}

		c.Resources.ActiveEffects = append(c.Resources.ActiveEffects, oldEffect)
	}
}

// RemoveStatusEffect removes a status effect by ID
func (c *Character) RemoveStatusEffect(effectID string) {
	c.GetEffectManager().RemoveEffect(effectID)
}

// GetActiveStatusEffects returns all active status effects
func (c *Character) GetActiveStatusEffects() []*effects.StatusEffect {
	return c.GetEffectManager().GetActiveEffects()
}

// GetAttackModifiers returns all modifiers that apply to attack rolls
func (c *Character) GetAttackModifiers(conditions map[string]string) []effects.Modifier {
	return c.GetEffectManager().GetModifiers(effects.TargetAttackRoll, conditions)
}

// GetDamageModifiers returns all modifiers that apply to damage rolls
func (c *Character) GetDamageModifiers(conditions map[string]string) []effects.Modifier {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.getDamageModifiersInternal(conditions)
}

// getDamageModifiersInternal returns damage modifiers without locking (caller must hold lock)
func (c *Character) getDamageModifiersInternal(conditions map[string]string) []effects.Modifier {
	return c.getEffectManagerInternal().GetModifiers(effects.TargetDamage, conditions)
}

// ApplyDamageResistance applies resistance/vulnerability/immunity from status effects
func (c *Character) ApplyDamageResistance(damageType damage.Type, amount int) int {
	conditions := map[string]string{
		"damage_type": string(damageType),
	}

	// Check for immunities
	immunities := c.GetEffectManager().GetModifiers(effects.TargetImmunity, conditions)
	for _, mod := range immunities {
		if mod.DamageType == string(damageType) {
			return 0 // Immune to this damage type
		}
	}

	// Check for resistances
	resistances := c.GetEffectManager().GetModifiers(effects.TargetResistance, conditions)
	hasResistance := false
	for _, mod := range resistances {
		if mod.DamageType == string(damageType) {
			hasResistance = true
			break
		}
	}

	// Check for vulnerabilities
	vulnerabilities := c.GetEffectManager().GetModifiers(effects.TargetVulnerability, conditions)
	hasVulnerability := false
	for _, mod := range vulnerabilities {
		if mod.DamageType == string(damageType) {
			hasVulnerability = true
			break
		}
	}

	// Apply modifiers
	if hasVulnerability {
		amount *= 2
	}
	if hasResistance {
		amount /= 2
	}

	return amount
}

// applyFightingStyleBonuses applies bonuses from fighter fighting styles
func (c *Character) applyFightingStyleBonuses(weapon *Weapon, attackBonus, damageBonus int) (finalAttackBonus, finalDamageBonus int) {
	return c.applyFightingStyleBonusesWithHand(weapon, attackBonus, damageBonus, SlotMainHand)
}

// applyFightingStyleBonusesWithHand applies bonuses from fighter fighting styles for a specific hand
func (c *Character) applyFightingStyleBonusesWithHand(weapon *Weapon, attackBonus, damageBonus int, hand Slot) (finalAttackBonus, finalDamageBonus int) {
	// Check if the character has fighting style feature
	var fightingStyle string
	log.Printf("DEBUG: Checking for fighting style among %d features", len(c.Features))
	for _, feature := range c.Features {
		if feature.Key == "fighting_style" {
			log.Printf("DEBUG: Found fighting_style feature, metadata=%v", feature.Metadata)
			if feature.Metadata != nil {
				if style, ok := feature.Metadata["style"].(string); ok {
					fightingStyle = style
					log.Printf("DEBUG: Found fighting style: %s", style)
					break
				} else {
					log.Printf("DEBUG: Fighting style metadata exists but no 'style' key or wrong type")
				}
			} else {
				log.Printf("DEBUG: Fighting style feature found but metadata is nil")
			}
		}
	}

	if fightingStyle == "" {
		log.Printf("DEBUG: No fighting style found")
		return attackBonus, damageBonus
	}
	log.Printf("DEBUG: Applying fighting style: %s", fightingStyle)

	// Apply fighting style bonuses
	switch fightingStyle {
	case "archery":
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
			offHand := c.EquippedSlots[SlotOffHand]
			log.Printf("DEBUG: Off-hand equipment: %v", offHand)

			offHandHasWeapon := false
			if offHand != nil {
				_, isWeapon := offHand.(*Weapon)
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
		if hand == SlotOffHand && weapon.IsMelee() {
			// Two-weapon fighting allows adding ability modifier to off-hand damage
			// The base off-hand damage calculation doesn't include ability modifier
			// So we need to add it here
			if c.Attributes != nil && c.Attributes[AttributeStrength] != nil {
				damageBonus = c.Attributes[AttributeStrength].Bonus
			}
		}
	}

	return attackBonus, damageBonus
}
