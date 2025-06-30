package character

import (
	"fmt"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/damage"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/equipment"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/combat/attack"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"log"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/effects"
)

// calculateWeaponAbilityBonus determines the ability bonus for a weapon attack
// Takes into account finesse weapons and monk weapons that can use DEX
func (c *Character) calculateWeaponAbilityBonus(weap *equipment.Weapon, hasMartialArts bool) int {
	if c.Attributes == nil {
		return 0
	}

	switch weap.WeaponRange {
	case "Ranged":
		if c.Attributes[shared.AttributeDexterity] != nil {
			return c.Attributes[shared.AttributeDexterity].Bonus
		}
	case "Melee":
		// Finesse weapons and monk weapons can use DEX instead of STR
		if weap.IsFinesse() || (hasMartialArts && weap.IsMonkWeapon()) {
			strBonus := 0
			dexBonus := 0
			if c.Attributes[shared.AttributeStrength] != nil {
				strBonus = c.Attributes[shared.AttributeStrength].Bonus
			}
			if c.Attributes[shared.AttributeDexterity] != nil {
				dexBonus = c.Attributes[shared.AttributeDexterity].Bonus
			}
			// Use the higher of STR or DEX
			if dexBonus > strBonus {
				return dexBonus
			}
			return strBonus
		} else if c.Attributes[shared.AttributeStrength] != nil {
			return c.Attributes[shared.AttributeStrength].Bonus
		}
	}
	return 0
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
		if weap, ok := c.EquippedSlots[SlotMainHand].(*equipment.Weapon); ok {
			attacks := make([]*attack.Result, 0)

			// Check proficiency while we have the mutex
			directProf := c.hasWeaponProficiencyInternal(weap.GetKey())
			categoryProf := c.hasWeaponCategoryProficiency(weap.WeaponCategory)
			isProficient := directProf || categoryProf
			// Proficiency check completed

			// Check if character has Martial Arts
			hasMartialArts := false
			for _, feature := range c.Features {
				if feature != nil && feature.Key == "martial-arts" {
					hasMartialArts = true
					break
				}
			}

			// Calculate ability bonus based on weapon type
			abilityBonus := c.calculateWeaponAbilityBonus(weap, hasMartialArts)

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

			// Roll the attack with fighting style consideration
			var attak1 *attack.Result
			fightingStyle := c.getFightingStyle()
			// Great Weapon Fighting only applies to two-handed melee weapons
			useGreatWeaponFighting := fightingStyle == "great_weapon" && weap.IsTwoHanded() && weap.IsMelee()

			if weap.IsTwoHanded() && weap.TwoHandedDamage != nil {
				if useGreatWeaponFighting {
					attak1, err = attack.RollAttackWithFightingStyle(c.getDiceRoller(), attackBonus, damageBonus, weap.TwoHandedDamage, "great_weapon")
				} else {
					attak1, err = attack.RollAttack(c.getDiceRoller(), attackBonus, damageBonus, weap.TwoHandedDamage)
				}
			} else {
				attak1, err = attack.RollAttack(c.getDiceRoller(), attackBonus, damageBonus, weap.Damage)
			}

			if err != nil {
				log.Printf("Weapon attack error: %v", err)
				return nil, err
			}
			// Set the weapon key for action economy tracking
			attak1.WeaponKey = weap.GetKey()
			log.Printf("Weapon attack successful")
			attacks = append(attacks, attak1)

			if c.EquippedSlots[SlotOffHand] != nil {
				if offWeap, offOk := c.EquippedSlots[SlotOffHand].(*equipment.Weapon); offOk {
					// Same process for off-hand weapon
					offHandProficient := c.hasWeaponProficiencyInternal(offWeap.GetKey()) ||
						c.hasWeaponCategoryProficiency(offWeap.WeaponCategory)

					// Calculate off-hand ability bonus
					offHandAbilityBonus := c.calculateWeaponAbilityBonus(offWeap, hasMartialArts)

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

					attak2, err := attack.RollAttack(c.getDiceRoller(), offHandAttackBonus, offHandDamageBonus, offWeap.Damage)
					if err != nil {
						return nil, err
					}
					// Set the weapon key for off-hand attack
					attak2.WeaponKey = offWeap.GetKey()
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
		if weap, ok := c.EquippedSlots[SlotTwoHanded].(*equipment.Weapon); ok {
			log.Printf("Two-handed weapon found: %s", weap.GetName())
			// Check proficiency while we have the mutex
			directProf := c.hasWeaponProficiencyInternal(weap.GetKey())
			categoryProf := c.hasWeaponCategoryProficiency(weap.WeaponCategory)
			isProficient := directProf || categoryProf
			// Proficiency check completed

			// Check if character has Martial Arts
			hasMartialArts := false
			for _, feature := range c.Features {
				if feature != nil && feature.Key == "martial-arts" {
					hasMartialArts = true
					break
				}
			}

			// Calculate ability bonus based on weapon type
			abilityBonus := c.calculateWeaponAbilityBonus(weap, hasMartialArts)

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

			// Apply Great Weapon Fighting for two-handed melee weapons
			fightingStyle := c.getFightingStyle()
			var a *attack.Result
			if fightingStyle == "great_weapon" && weap.IsMelee() {
				a, err = attack.RollAttackWithFightingStyle(c.getDiceRoller(), attackBonus, damageBonus, dmg, "great_weapon")
			} else {
				a, err = attack.RollAttack(c.getDiceRoller(), attackBonus, damageBonus, dmg)
			}
			if err != nil {
				return nil, err
			}

			// Set the weapon key for two-handed weapon
			a.WeaponKey = weap.GetKey()

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
		if w, ok := weapon.(*equipment.Weapon); ok && w.WeaponCategory != "" {
			return c.hasWeaponCategoryProficiency(w.WeaponCategory)
		}
	}

	// Also check equipped weapons
	for _, equipped := range c.EquippedSlots {
		if equipped != nil && equipped.GetKey() == weaponKey {
			if w, ok := equipped.(*equipment.Weapon); ok && w.WeaponCategory != "" {
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

	weaponProficiencies, exists := c.Proficiencies[rulebook.ProficiencyTypeWeapon]
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

	weaponProficiencies, exists := c.Proficiencies[rulebook.ProficiencyTypeWeapon]
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
	// Check if character has Martial Arts feature (monks)
	hasMartialArts := false
	for _, feature := range c.Features {
		if feature != nil && feature.Key == "martial-arts" {
			hasMartialArts = true
			break
		}
	}

	// Determine ability bonus - monks can use DEX instead of STR
	bonus := 0
	if hasMartialArts && c.Attributes != nil {
		// Use the higher of STR or DEX for monks
		strBonus := 0
		dexBonus := 0
		if c.Attributes[shared.AttributeStrength] != nil {
			strBonus = c.Attributes[shared.AttributeStrength].Bonus
		}
		if c.Attributes[shared.AttributeDexterity] != nil {
			dexBonus = c.Attributes[shared.AttributeDexterity].Bonus
		}
		bonus = strBonus
		if dexBonus > strBonus {
			bonus = dexBonus
		}
	} else if c.Attributes != nil && c.Attributes[shared.AttributeStrength] != nil {
		// Non-monks use STR
		bonus = c.Attributes[shared.AttributeStrength].Bonus
	}

	// Apply damage bonuses from active effects (e.g., rage) to improvised attacks
	damageBonus, err := c.applyActiveEffectDamageBonus(bonus, "melee")
	if err != nil {
		log.Printf("ERROR: Failed to apply active effect damage bonus to improvised: %v", err)
		damageBonus = bonus // Fall back to base
	}

	// Determine damage dice size based on monk level
	diceSize := 4 // Default for non-monks and level 1-4 monks
	if hasMartialArts {
		switch {
		case c.Level >= 17:
			diceSize = 10
		case c.Level >= 11:
			diceSize = 8
		case c.Level >= 5:
			diceSize = 6
		}
	}

	attackResult, err := c.getDiceRoller().Roll(1, 20, 0)
	if err != nil {
		return nil, err
	}
	damageResult, err := c.getDiceRoller().Roll(1, diceSize, 0)
	if err != nil {
		return nil, err
	}

	return &attack.Result{
		AttackRoll:   attackResult.Total + bonus,
		DamageRoll:   damageResult.Total + damageBonus,
		AttackType:   damage.TypeBludgeoning,
		AttackResult: attackResult,
		DamageResult: damageResult,
		WeaponDamage: &damage.Damage{
			DiceCount:  1,
			DiceSize:   diceSize,
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
