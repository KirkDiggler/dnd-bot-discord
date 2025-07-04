# rpg-toolkit Extensions Needed for D&D 5e

## Immediate Needs (Phase 1)

Based on our analysis, here's what we need to add to rpg-toolkit to support feats, proficiencies, and conditions for a local D&D 5e rulebook implementation:

### 1. Proficiency System 🎯

**Why we need it**: Characters have proficiencies with weapons, armor, skills, and saving throws that provide bonuses.

```go
// Core proficiency types
type ProficiencyType string
const (
    ProficiencyWeapon      ProficiencyType = "weapon"
    ProficiencyArmor       ProficiencyType = "armor"
    ProficiencySkill       ProficiencyType = "skill"
    ProficiencySave        ProficiencyType = "saving_throw"
    ProficiencyTool        ProficiencyType = "tool"
    ProficiencyInstrument  ProficiencyType = "instrument"
)

// What we can be proficient with
type Proficiency interface {
    Type() ProficiencyType
    Key() string           // "shortsword", "athletics", "constitution"
    Name() string          // "Shortsword", "Athletics", "Constitution"
    Category() string      // "simple-weapons", "martial-weapons", "strength-skills"
}

// Manages proficiencies for entities
type ProficiencyManager interface {
    // Check proficiency
    HasProficiency(entity Entity, profType ProficiencyType, key string) bool
    HasWeaponProficiency(entity Entity, weaponKey string) bool
    HasSkillProficiency(entity Entity, skillKey string) bool
    
    // Get bonuses
    GetProficiencyBonus(entity Entity) int // Based on level: +2, +3, +4, +5, +6
    GetWeaponAttackBonus(entity Entity, weaponKey string) int
    GetSkillBonus(entity Entity, skillKey string) int
    GetSaveBonus(entity Entity, abilityKey string) int
    
    // Manage proficiencies
    AddProficiency(entity Entity, prof Proficiency)
    RemoveProficiency(entity Entity, profKey string)
    GetAllProficiencies(entity Entity) []Proficiency
}
```

**Real Examples**:
- Rogue has proficiency with shortswords → +2 to attack rolls
- Fighter has proficiency with Constitution saves → +2 to Constitution saving throws
- Bard has proficiency with Performance skill → +2 to Performance checks

### 2. Resource Management System ⚡

**Why we need it**: Characters have limited-use abilities (spell slots, rage uses, bardic inspiration, etc.)

```go
type ResourceType string
const (
    ResourceSpellSlots   ResourceType = "spell_slots"
    ResourceAbilityUses  ResourceType = "ability_uses"
    ResourceHitDice      ResourceType = "hit_dice"
    ResourceActionEcon   ResourceType = "action_economy"
)

type RestType string
const (
    RestTypeShort RestType = "short"
    RestTypeLong  RestType = "long"
)

type Resource interface {
    Type() ResourceType
    Key() string        // "spell_slot_1", "rage_uses", "hit_dice_d10"
    Name() string       // "1st Level Spell Slots", "Rage Uses"
    Current() int       // Current available uses
    Maximum() int       // Max uses at full resources
    
    // Resource management
    CanConsume(amount int) bool
    Consume(amount int) error
    Restore(amount int) error
    RestoreToMax()
    
    // Rest mechanics
    RestoresOn() RestType  // Short rest, long rest, or both
    RestoreAmount() int    // How much restores on rest
}

type ResourceManager interface {
    // Get resources
    GetResource(entity Entity, resourceType ResourceType, key string) Resource
    GetSpellSlots(entity Entity, level int) Resource
    GetAbilityUses(entity Entity, abilityKey string) Resource
    
    // Manage resources
    AddResource(entity Entity, resource Resource)
    RemoveResource(entity Entity, resourceKey string)
    
    // Rest mechanics
    ProcessShortRest(entity Entity)
    ProcessLongRest(entity Entity)
    
    // Usage tracking
    ConsumeSpellSlot(entity Entity, level int) error
    ConsumeAbilityUse(entity Entity, abilityKey string) error
    RestoreResource(entity Entity, resourceKey string, amount int) error
}
```

**Real Examples**:
- Wizard has 2 first-level spell slots → can cast 2 first-level spells
- Barbarian has 2 rage uses per long rest → can rage twice before needing long rest
- Fighter has 1 Second Wind use per short rest → regains use after short rest

### 3. Enhanced Condition System 🎭

**Why we need it**: D&D 5e has standardized conditions that apply specific mechanical effects.

```go
// Enhanced condition with mechanical effects
type EnhancedCondition interface {
    Condition
    
    // Mechanical effects
    GetModifiers() []Modifier
    GetRestrictedActions() []ActionType
    GetAutomaticFailures() []string    // "strength_saves", "dexterity_saves"
    GetAutomaticSuccesses() []string
    GetMovementRestrictions() MovementRestriction
    
    // Interaction rules
    ConflictsWith() []string           // Conditions that remove this one
    PreventsConditions() []string      // Conditions this prevents
    UpgradesTo() map[string]string     // "restrained" -> "paralyzed"
}

// Standard D&D 5e conditions with proper mechanics
var StandardConditions = map[string]EnhancedCondition{
    "blinded": NewCondition("blinded").
        WithModifiers(DisadvantageOn("attack_rolls")).
        WithModifiers(AdvantageAgainst("attack_rolls")),
        
    "frightened": NewCondition("frightened").
        WithModifiers(DisadvantageOn("ability_checks", "attack_rolls")).
        WithMovementRestriction(CannotMoveTowardsSource()),
        
    "paralyzed": NewCondition("paralyzed").
        WithRestrictedActions("move", "speak", "actions", "reactions").
        WithAutomaticFailures("strength_saves", "dexterity_saves").
        WithModifiers(AdvantageAgainst("attack_rolls")).
        WithModifiers(AutoCriticalHitsWithin(5)),
        
    "poisoned": NewCondition("poisoned").
        WithModifiers(DisadvantageOn("attack_rolls", "ability_checks")),
        
    "prone": NewCondition("prone").
        WithModifiers(DisadvantageOn("attack_rolls")).
        WithModifiers(AdvantageAgainst("melee_attacks")).
        WithModifiers(DisadvantageAgainst("ranged_attacks")),
}
```

**Real Examples**:
- Poisoned creature has disadvantage on attack rolls and ability checks
- Prone creature has disadvantage on attack rolls, advantage against melee attacks
- Paralyzed creature auto-fails Strength and Dexterity saves, takes automatic crits

### 4. Equipment & Weapon System ⚔️

**Why we need it**: Equipment provides bonuses, properties, and determines proficiency requirements.

```go
type EquipmentSlot string
const (
    SlotMainHand   EquipmentSlot = "main_hand"
    SlotOffHand    EquipmentSlot = "off_hand"
    SlotArmor      EquipmentSlot = "armor"
    SlotShield     EquipmentSlot = "shield"
    SlotRing       EquipmentSlot = "ring"
    SlotNecklace   EquipmentSlot = "necklace"
)

type WeaponProperty string
const (
    PropertyFinesse    WeaponProperty = "finesse"     // Can use DEX for attack/damage
    PropertyVersatile  WeaponProperty = "versatile"   // 1d8 one-handed, 1d10 two-handed
    PropertyTwoHanded  WeaponProperty = "two_handed"  // Requires two hands
    PropertyLight      WeaponProperty = "light"       // Two-weapon fighting
    PropertyHeavy      WeaponProperty = "heavy"       // Small creatures have disadvantage
    PropertyReach      WeaponProperty = "reach"       // 10 ft reach instead of 5 ft
    PropertyThrown     WeaponProperty = "thrown"      // Can be thrown
)

type Weapon interface {
    Equipment
    
    // Attack properties
    AttackBonus(wielder Entity) int
    DamageBonus(wielder Entity) int
    DamageDice() string              // "1d8", "1d6"
    DamageType() string              // "slashing", "piercing", "bludgeoning"
    
    // Weapon properties
    HasProperty(prop WeaponProperty) bool
    IsFinesse() bool
    IsVersatile() bool
    GetVersatileDamage() string      // "1d10" for versatile weapons
    
    // Proficiency requirements
    ProficiencyCategory() string     // "simple-weapons", "martial-weapons"
    RequiresProficiency(wielder Entity) bool
    
    // Attack mechanics
    CanUseAbilityScore(ability string) bool  // Finesse weapons can use DEX
    GetAttackRange() int                     // 5 ft melee, varies for ranged
}

type EquipmentManager interface {
    // Equipment management
    EquipItem(entity Entity, item Equipment, slot EquipmentSlot) error
    UnequipItem(entity Entity, slot EquipmentSlot) error
    GetEquippedItem(entity Entity, slot EquipmentSlot) Equipment
    
    // Weapon-specific
    GetMainHandWeapon(entity Entity) Weapon
    GetOffHandWeapon(entity Entity) Weapon
    CanDualWield(entity Entity) bool
    
    // AC calculation
    GetArmorClass(entity Entity) int
    GetArmorBonus(entity Entity) int
    GetShieldBonus(entity Entity) int
}
```

**Real Examples**:
- Shortsword (finesse, light) → can use DEX for attacks, enables two-weapon fighting
- Longsword (versatile) → 1d8 damage one-handed, 1d10 damage two-handed
- Greataxe (heavy, two-handed) → requires two hands, small creatures have disadvantage

### 5. Feature System 🌟

**Why we need it**: Class and race features provide the core mechanical differences between characters.

```go
type FeatureType string
const (
    FeatureRacial    FeatureType = "racial"
    FeatureClass     FeatureType = "class" 
    FeatureSubclass  FeatureType = "subclass"
    FeatureFeat      FeatureType = "feat"
    FeatureItem      FeatureType = "item"
)

type FeatureTiming string
const (
    TimingPassive     FeatureTiming = "passive"      // Always active
    TimingTriggered   FeatureTiming = "triggered"    // Reacts to events
    TimingActivated   FeatureTiming = "activated"    // Player chooses to use
)

type Feature interface {
    // Basic info
    Key() string
    Name() string
    Description() string
    Type() FeatureType
    Level() int
    Source() string
    
    // Mechanical effects
    IsPassive() bool
    GetTiming() FeatureTiming
    GetModifiers() []Modifier
    GetProficiencies() []Proficiency
    GetResources() []Resource
    
    // Event integration
    GetEventListeners() []EventListener
    CanTrigger(event Event) bool
    TriggerFeature(entity Entity, event Event) error
    
    // Prerequisites
    HasPrerequisites() bool
    MeetsPrerequisites(entity Entity) bool
    GetPrerequisites() []string
}

// Examples of how features work
var ExampleFeatures = map[string]Feature{
    "rage": NewFeature("rage", FeatureClass).
        WithResources(NewResource("rage_uses", 2, RestTypeLong)).
        WithEventListeners(RageListener{}).
        WithTiming(TimingActivated),
        
    "sneak_attack": NewFeature("sneak_attack", FeatureClass).
        WithEventListeners(SneakAttackListener{}).
        WithTiming(TimingTriggered),
        
    "darkvision": NewFeature("darkvision", FeatureRacial).
        WithModifiers(VisionModifier("darkvision", 60)).
        WithTiming(TimingPassive),
}
```

**Real Examples**:
- Rage (Barbarian class feature) → +2 damage, resistance to physical damage, limited uses
- Sneak Attack (Rogue class feature) → Extra damage when conditions met
- Darkvision (Half-orc racial feature) → See in darkness up to 60 feet

## Integration Strategy

### Step 1: Add to rpg-toolkit Core
These systems should be added as core rpg-toolkit packages that any game system can use:
- `proficiency/` - Generic proficiency system
- `resources/` - Resource management and rest mechanics  
- `equipment/` - Equipment slots and bonuses
- `features/` - Feature system with event integration

### Step 2: Create D&D 5e Rulebook Package
Create a D&D 5e specific package that uses the core systems:
- `dnd5e/proficiencies.go` - All D&D 5e proficiencies
- `dnd5e/conditions.go` - All D&D 5e conditions with proper mechanics
- `dnd5e/equipment.go` - D&D 5e weapons and armor
- `dnd5e/features.go` - All class/race features
- `dnd5e/spells.go` - Spell implementations

### Step 3: Replace Discord Bot Systems
Migrate the Discord bot to use the rpg-toolkit D&D 5e implementation:
- Remove EventBusAdapter entirely
- Convert all listeners to use rpg-toolkit events
- Use rpg-toolkit proficiency/resource/equipment managers
- Replace character features with rpg-toolkit feature system

This creates a comprehensive, reusable D&D 5e rules engine in rpg-toolkit that can power not just our Discord bot, but any D&D 5e application.