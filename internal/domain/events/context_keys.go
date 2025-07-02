package events

// Context keys for event data
// These constants ensure consistent access to event context across the system
const (
	// Combat context keys
	ContextAttackType       = "attack_type"        // string: "melee" or "ranged"
	ContextDamage           = "damage"             // int: current damage value
	ContextDamageType       = "damage_type"        // string: damage type (bludgeoning, piercing, etc.)
	ContextTargetID         = "target_id"          // string: ID of the target
	ContextIsCritical       = "is_critical"        // bool: whether the attack was a critical hit
	ContextWeaponKey        = "weapon_key"         // string: key of the weapon used
	ContextWeaponType       = "weapon_type"        // string: "melee" or "ranged"
	ContextWeaponHasFinesse = "weapon_has_finesse" // bool: whether the weapon has finesse property

	// Combat conditions
	ContextHasAdvantage    = "has_advantage"    // bool: attacker has advantage
	ContextHasDisadvantage = "has_disadvantage" // bool: attacker has disadvantage
	ContextAllyAdjacent    = "ally_adjacent"    // bool: ally is adjacent to target

	// Turn tracking
	ContextTurnCount     = "turn_count"     // int: total turns elapsed
	ContextRound         = "round"          // int: current round number
	ContextNumCombatants = "num_combatants" // int: number of combatants

	// Spell context keys
	ContextSpellLevel      = "spell_level"      // int: level at which spell was cast
	ContextSpellSchool     = "spell_school"     // string: school of magic
	ContextSpellComponents = "spell_components" // []string: required components (V, S, M)
	ContextSpellSaveType   = "spell_save_type"  // string: ability for save (STR, DEX, etc.)
	ContextSpellSaveDC     = "spell_save_dc"    // int: difficulty class for saves
	ContextConcentration   = "concentration"    // bool: whether spell requires concentration

	// Ability check context
	ContextAbilityType     = "ability_type"     // string: STR, DEX, CON, INT, WIS, CHA
	ContextSkillType       = "skill_type"       // string: specific skill being used
	ContextDifficultyClass = "difficulty_class" // int: DC for the check
	ContextCheckResult     = "check_result"     // int: result of the check

	// Status effect context
	ContextStatusType     = "status"          // string: type of status effect
	ContextEffectDuration = "effect_duration" // int: duration in rounds
	ContextEffectSource   = "effect_source"   // string: source of the effect

	// Special ability context
	ContextSneakAttackDamage = "sneak_attack_damage" // int: sneak attack damage dealt
	ContextSneakAttackDice   = "sneak_attack_dice"   // string: dice notation (e.g., "3d6")
	ContextDamageBonusSource = "damage_bonus_source" // string: source of damage bonus
	ContextResistanceApplied = "resistance_applied"  // string: type of resistance applied
)
