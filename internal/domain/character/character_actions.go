package character

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/equipment"
)

// StartNewTurn resets action economy and other per-turn resources
func (c *Character) StartNewTurn() {
	c.mu.Lock()
	defer c.mu.Unlock()

	resources := c.Resources
	if resources == nil {
		c.initializeResourcesInternal()
		resources = c.Resources
	}

	// Reset action economy
	resources.ActionEconomy.Reset()

	// Reset reaction at the start of YOUR turn
	resources.ActionEconomy.ReactionUsed = false

	// Reset sneak attack
	resources.SneakAttackUsedThisTurn = false

	// Update available bonus actions based on character state
	c.updateAvailableBonusActionsInternal()
}

// RecordAction tracks an action taken by the character
func (c *Character) RecordAction(actionType, subtype, weaponKey string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Resources == nil {
		return
	}

	c.Resources.ActionEconomy.RecordAction(actionType, subtype, weaponKey)

	// Update available bonus actions based on new action
	c.updateAvailableBonusActionsInternal()
}

// GetAvailableBonusActions returns the currently available bonus actions
func (c *Character) GetAvailableBonusActions() []BonusActionOption {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Resources == nil {
		return []BonusActionOption{}
	}

	return c.Resources.ActionEconomy.AvailableBonusActions
}

// CanTakeBonusAction checks if the character can take a bonus action
func (c *Character) CanTakeBonusAction() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Resources == nil {
		return false
	}

	return !c.Resources.ActionEconomy.BonusActionUsed &&
		len(c.Resources.ActionEconomy.AvailableBonusActions) > 0
}

// UseBonusAction marks the bonus action as used
func (c *Character) UseBonusAction(bonusActionKey string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Resources == nil || c.Resources.ActionEconomy.BonusActionUsed {
		return false
	}

	// Verify this bonus action is available
	for _, option := range c.Resources.ActionEconomy.AvailableBonusActions {
		if option.Key == bonusActionKey {
			c.Resources.ActionEconomy.BonusActionUsed = true
			c.Resources.ActionEconomy.RecordAction("bonus_action", option.ActionType, "")
			// Update available bonus actions (should be empty now)
			c.updateAvailableBonusActionsInternal()
			return true
		}
	}

	return false
}

// updateAvailableBonusActionsInternal updates available bonus actions based on current state
// Caller must hold the mutex
func (c *Character) updateAvailableBonusActionsInternal() {
	if c.Resources == nil {
		return
	}

	// Clear existing
	c.Resources.ActionEconomy.AvailableBonusActions = []BonusActionOption{}

	// Don't add any if bonus action already used
	if c.Resources.ActionEconomy.BonusActionUsed {
		return
	}

	// Check Martial Arts
	if c.checkMartialArtsBonusActionInternal() {
		c.Resources.ActionEconomy.AvailableBonusActions = append(
			c.Resources.ActionEconomy.AvailableBonusActions,
			BonusActionOption{
				Key:         "martial_arts_strike",
				Name:        "Martial Arts Bonus Strike",
				Description: "Make one unarmed strike as a bonus action",
				Source:      "martial_arts",
				ActionType:  "unarmed_strike",
			},
		)
	}

	// Check Two-Weapon Fighting
	if c.checkTwoWeaponBonusActionInternal() {
		c.Resources.ActionEconomy.AvailableBonusActions = append(
			c.Resources.ActionEconomy.AvailableBonusActions,
			BonusActionOption{
				Key:         "two_weapon_attack",
				Name:        "Off-Hand Attack",
				Description: "Attack with your off-hand weapon",
				Source:      "two_weapon_fighting",
				ActionType:  "weapon_attack",
			},
		)
	}

	// Future: Add more bonus action checks here
	// - Cunning Action (Rogue)
	// - Rage (Barbarian)
	// - Second Wind (Fighter)
	// - Bonus action spells
}

// checkMartialArtsBonusActionInternal checks if Martial Arts bonus action is available
// Caller must hold the mutex
func (c *Character) checkMartialArtsBonusActionInternal() bool {
	// Must have Martial Arts feature
	hasMartialArts := false
	for _, feature := range c.Features {
		if feature != nil && feature.Key == "martial-arts" {
			hasMartialArts = true
			break
		}
	}
	if !hasMartialArts {
		return false
	}

	// Must have taken the Attack action
	if !c.Resources.ActionEconomy.HasTakenAction("attack") {
		return false
	}

	// Check if any attack was with unarmed strike or monk weapon
	attacks := c.Resources.ActionEconomy.GetActionsByType("attack")
	for _, attack := range attacks {
		// Unarmed strike
		if attack.Subtype == "unarmed_strike" {
			return true
		}

		// Monk weapon
		if attack.Subtype == "weapon" && attack.WeaponKey != "" {
			// Check if the weapon used was a monk weapon
			if c.isMonkWeaponByKeyInternal(attack.WeaponKey) {
				return true
			}
		}
	}

	return false
}

// isMonkWeaponByKeyInternal checks if a weapon key is a monk weapon
// Caller must hold the mutex
func (c *Character) isMonkWeaponByKeyInternal(weaponKey string) bool {
	// Check equipped weapons
	for _, equipped := range c.EquippedSlots {
		if weapon, ok := equipped.(*equipment.Weapon); ok && weapon.GetKey() == weaponKey {
			return weapon.IsMonkWeapon()
		}
	}

	// Check inventory
	weapon := c.getEquipment(weaponKey)
	if weapon != nil {
		if w, ok := weapon.(*equipment.Weapon); ok {
			return w.IsMonkWeapon()
		}
	}

	return false
}

// checkTwoWeaponBonusActionInternal checks if two-weapon fighting bonus action is available
// Caller must hold the mutex
func (c *Character) checkTwoWeaponBonusActionInternal() bool {
	// Must have taken the Attack action
	if !c.Resources.ActionEconomy.HasTakenAction("attack") {
		return false
	}

	// Check if any attack was with a light melee weapon
	hasLightWeaponAttack := false
	attacks := c.Resources.ActionEconomy.GetActionsByType("attack")
	for _, attack := range attacks {
		if attack.Subtype == "weapon" && attack.WeaponKey != "" {
			// Get the weapon and check if it's light
			var weapon *equipment.Weapon
			for _, equipped := range c.EquippedSlots {
				if w, ok := equipped.(*equipment.Weapon); ok && w.GetKey() == attack.WeaponKey {
					weapon = w
					break
				}
			}

			if weapon != nil && weapon.HasProperty("light") && weapon.IsMelee() {
				hasLightWeaponAttack = true
				break
			}
		}
	}

	if !hasLightWeaponAttack {
		return false
	}

	// Must have a light weapon in off-hand
	if offHand, ok := c.EquippedSlots[SlotOffHand].(*equipment.Weapon); ok {
		return offHand.HasProperty("light") && offHand.IsMelee()
	}

	return false
}

// HasActionAvailable checks if the character can take their main action
func (c *Character) HasActionAvailable() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Resources == nil {
		return true // Assume available if not initialized
	}

	return !c.Resources.ActionEconomy.ActionUsed
}

// GetActionsTaken returns a summary of actions taken this turn
func (c *Character) GetActionsTaken() []ActionRecord {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Resources == nil {
		return []ActionRecord{}
	}

	return c.Resources.ActionEconomy.ActionsThisTurn
}
