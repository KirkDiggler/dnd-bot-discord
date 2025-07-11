package features

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
)

// ClassFeatures defines the special features for each class
var ClassFeatures = map[string][]rulebook.CharacterFeature{
	"monk": {
		{
			Key:         "unarmored_defense_monk",
			Name:        "Unarmored Defense",
			Description: "While wearing no armor and not wielding a shield, your AC equals 10 + your Dexterity modifier + your Wisdom modifier.",
			Type:        rulebook.FeatureTypeClass,
			Level:       1,
			Source:      "Monk",
		},
		{
			Key:         "martial-arts",
			Name:        "Martial Arts",
			Description: "You can use Dexterity instead of Strength for attack and damage rolls of unarmed strikes and monk weapons. Monk weapons are shortswords and any simple melee weapons that don't have the two-handed or heavy property. Your unarmed strikes use a d4 for damage. This die changes as you gain monk levels: d6 at 5th level, d8 at 11th level, and d10 at 17th level. When you use the Attack action, you can make one unarmed strike as a bonus action.",
			Type:        rulebook.FeatureTypeClass,
			Level:       1,
			Source:      "Monk",
		},
	},
	"barbarian": {
		{
			Key:         "unarmored_defense_barbarian",
			Name:        "Unarmored Defense",
			Description: "While not wearing armor, your AC equals 10 + your Dexterity modifier + your Constitution modifier. You can use a shield and still gain this benefit.",
			Type:        rulebook.FeatureTypeClass,
			Level:       1,
			Source:      "Barbarian",
		},
		{
			Key:         "rage",
			Name:        "Rage",
			Description: "In battle, you fight with primal ferocity. As a bonus action, you can enter a rage for 1 minute.",
			Type:        rulebook.FeatureTypeClass,
			Level:       1,
			Source:      "Barbarian",
		},
	},
	"wizard": {
		{
			Key:         "spellcasting_wizard",
			Name:        "Spellcasting",
			Description: "You have learned to cast spells. See chapter 10 for the general rules of spellcasting and chapter 11 for the wizard spell list.",
			Type:        rulebook.FeatureTypeClass,
			Level:       1,
			Source:      "Wizard",
		},
		{
			Key:         "arcane_recovery",
			Name:        "Arcane Recovery",
			Description: "Once per day during a short rest, you can recover expended spell slots.",
			Type:        rulebook.FeatureTypeClass,
			Level:       1,
			Source:      "Wizard",
		},
	},
	"fighter": {
		{
			Key:         "fighting_style",
			Name:        "Fighting Style",
			Description: "You adopt a particular style of fighting as your specialty.",
			Type:        rulebook.FeatureTypeClass,
			Level:       1,
			Source:      "Fighter",
		},
		{
			Key:         "second_wind",
			Name:        "Second Wind",
			Description: "You have a limited well of stamina that you can draw on to protect yourself from harm.",
			Type:        rulebook.FeatureTypeClass,
			Level:       1,
			Source:      "Fighter",
		},
	},
	"rogue": {
		{
			Key:         "expertise",
			Name:        "Expertise",
			Description: "Choose two of your skill proficiencies. Your proficiency bonus is doubled for any ability check you make using either of the chosen proficiencies.",
			Type:        rulebook.FeatureTypeClass,
			Level:       1,
			Source:      "Rogue",
		},
		{
			Key:         "sneak_attack",
			Name:        "Sneak Attack",
			Description: "You know how to strike subtly and exploit a foe's distraction. Once per turn, you can deal an extra 1d6 damage.",
			Type:        rulebook.FeatureTypeClass,
			Level:       1,
			Source:      "Rogue",
		},
		{
			Key:         "thieves_cant",
			Name:        "Thieves' Cant",
			Description: "You have learned thieves' cant, a secret mix of dialect, jargon, and code.",
			Type:        rulebook.FeatureTypeClass,
			Level:       1,
			Source:      "Rogue",
		},
	},
	"cleric": {
		{
			Key:         "spellcasting_cleric",
			Name:        "Spellcasting",
			Description: "As a conduit for divine power, you can cast cleric spells.",
			Type:        rulebook.FeatureTypeClass,
			Level:       1,
			Source:      "Cleric",
		},
		{
			Key:         "divine_domain",
			Name:        "Divine Domain",
			Description: "Choose one domain related to your deity. Your choice grants you domain spells and other features.",
			Type:        rulebook.FeatureTypeClass,
			Level:       1,
			Source:      "Cleric",
		},
	},
	"ranger": {
		{
			Key:         "favored_enemy",
			Name:        "Favored Enemy",
			Description: "You have significant experience studying, tracking, hunting, and even talking to a certain type of enemy. Choose a type of favored enemy: aberrations, beasts, celestials, constructs, dragons, elementals, fey, fiends, giants, monstrosities, oozes, plants, or undead. Alternatively, you can select two races of humanoid (such as gnolls and orcs) as favored enemies. You have advantage on Wisdom (Survival) checks to track your favored enemies, as well as on Intelligence checks to recall information about them.",
			Type:        rulebook.FeatureTypeClass,
			Level:       1,
			Source:      "Ranger",
		},
		{
			Key:         "natural_explorer",
			Name:        "Natural Explorer",
			Description: "You are particularly familiar with one type of natural environment and are adept at traveling and surviving in such regions. Choose one type of favored terrain: arctic, coast, desert, forest, grassland, mountain, swamp, or the Underdark. When traveling for an hour or more in your favored terrain, you gain benefits including: difficult terrain doesn't slow your party's travel, your group can't become lost except by magical means, you remain alert to danger while tracking/foraging/navigating, you can move stealthily at a normal pace when alone, you find food and water for up to 6 people daily, and you can track creatures while moving at a fast pace.",
			Type:        rulebook.FeatureTypeClass,
			Level:       1,
			Source:      "Ranger",
		},
	},
	"paladin": {
		{
			Key:         "divine_sense",
			Name:        "Divine Sense",
			Description: "The presence of strong evil registers on your senses like a noxious odor, and powerful good rings like heavenly music in your ears. As an action, you can open your awareness to detect such forces. Until the end of your next turn, you know the location of any celestial, fiend, or undead within 60 feet of you that is not behind total cover.",
			Type:        rulebook.FeatureTypeClass,
			Level:       1,
			Source:      "Paladin",
		},
		{
			Key:         "lay_on_hands",
			Name:        "Lay on Hands",
			Description: "Your blessed touch can heal wounds. You have a pool of healing power that replenishes when you take a long rest. With that pool, you can restore a total number of hit points equal to your paladin level × 5. As an action, you can touch a creature and draw power from the pool to restore a number of hit points to that creature, up to the maximum amount remaining in your pool.",
			Type:        rulebook.FeatureTypeClass,
			Level:       1,
			Source:      "Paladin",
		},
	},
}

// RacialFeatures defines features for each race
var RacialFeatures = map[string][]rulebook.CharacterFeature{
	"dragonborn": {
		{
			Key:         "draconic_ancestry",
			Name:        "Draconic Ancestry",
			Description: "You have draconic ancestry. Choose one type of dragon from the Draconic Ancestry table.",
			Type:        rulebook.FeatureTypeRacial,
			Level:       0,
			Source:      "Dragonborn",
		},
		{
			Key:         "breath_weapon",
			Name:        "Breath Weapon",
			Description: "You can use your action to exhale destructive energy based on your draconic ancestry.",
			Type:        rulebook.FeatureTypeRacial,
			Level:       0,
			Source:      "Dragonborn",
		},
		{
			Key:         "damage_resistance",
			Name:        "Damage Resistance",
			Description: "You have resistance to the damage type associated with your draconic ancestry.",
			Type:        rulebook.FeatureTypeRacial,
			Level:       0,
			Source:      "Dragonborn",
		},
	},
	"elf": {
		{
			Key:         "darkvision",
			Name:        "Darkvision",
			Description: "You can see in dim light within 60 feet as if it were bright light.",
			Type:        rulebook.FeatureTypeRacial,
			Level:       0,
			Source:      "Elf",
		},
		{
			Key:         "keen_senses",
			Name:        "Keen Senses",
			Description: "You have proficiency in the Perception skill.",
			Type:        rulebook.FeatureTypeRacial,
			Level:       0,
			Source:      "Elf",
		},
		{
			Key:         "fey_ancestry",
			Name:        "Fey Ancestry",
			Description: "You have advantage on saving throws against being charmed, and magic can't put you to sleep.",
			Type:        rulebook.FeatureTypeRacial,
			Level:       0,
			Source:      "Elf",
		},
		{
			Key:         "trance",
			Name:        "Trance",
			Description: "Elves don't need to sleep. Instead, they meditate deeply for 4 hours a day.",
			Type:        rulebook.FeatureTypeRacial,
			Level:       0,
			Source:      "Elf",
		},
	},
	"dwarf": {
		{
			Key:         "darkvision",
			Name:        "Darkvision",
			Description: "You can see in dim light within 60 feet as if it were bright light.",
			Type:        rulebook.FeatureTypeRacial,
			Level:       0,
			Source:      "Dwarf",
		},
		{
			Key:         "dwarven_resilience",
			Name:        "Dwarven Resilience",
			Description: "You have advantage on saving throws against poison, and resistance against poison damage.",
			Type:        rulebook.FeatureTypeRacial,
			Level:       0,
			Source:      "Dwarf",
		},
		{
			Key:         "stonecunning",
			Name:        "Stonecunning",
			Description: "Whenever you make an Intelligence (History) check related to stonework, you are considered proficient and add double your proficiency bonus.",
			Type:        rulebook.FeatureTypeRacial,
			Level:       0,
			Source:      "Dwarf",
		},
	},
	"halfling": {
		{
			Key:         "lucky",
			Name:        "Lucky",
			Description: "When you roll a 1 on the d20 for an attack roll, ability check, or saving throw, you can reroll the die and must use the new roll.",
			Type:        rulebook.FeatureTypeRacial,
			Level:       0,
			Source:      "Halfling",
		},
		{
			Key:         "brave",
			Name:        "Brave",
			Description: "You have advantage on saving throws against being frightened.",
			Type:        rulebook.FeatureTypeRacial,
			Level:       0,
			Source:      "Halfling",
		},
		{
			Key:         "halfling_nimbleness",
			Name:        "Halfling Nimbleness",
			Description: "You can move through the space of any creature that is of a size larger than yours.",
			Type:        rulebook.FeatureTypeRacial,
			Level:       0,
			Source:      "Halfling",
		},
	},
	"human": {
		{
			Key:         "human_versatility",
			Name:        "Versatility",
			Description: "Humans gain +1 to all ability scores.",
			Type:        rulebook.FeatureTypeRacial,
			Level:       0,
			Source:      "Human",
		},
	},
	"tiefling": {
		{
			Key:         "darkvision",
			Name:        "Darkvision",
			Description: "You can see in dim light within 60 feet as if it were bright light.",
			Type:        rulebook.FeatureTypeRacial,
			Level:       0,
			Source:      "Tiefling",
		},
		{
			Key:         "hellish_resistance",
			Name:        "Hellish Resistance",
			Description: "You have resistance to fire damage.",
			Type:        rulebook.FeatureTypeRacial,
			Level:       0,
			Source:      "Tiefling",
		},
		{
			Key:         "infernal_legacy",
			Name:        "Infernal Legacy",
			Description: "You know the thaumaturgy cantrip. At 3rd level, you can cast hellish rebuke once per day. At 5th level, you can cast darkness once per day.",
			Type:        rulebook.FeatureTypeRacial,
			Level:       0,
			Source:      "Tiefling",
		},
	},
}

// GetClassFeatures returns features for a class at a given level
func GetClassFeatures(classKey string, level int) []rulebook.CharacterFeature {
	features := []rulebook.CharacterFeature{}
	if classFeats, exists := ClassFeatures[classKey]; exists {
		for _, feat := range classFeats {
			if feat.Level <= level {
				features = append(features, feat)
			}
		}
	}
	return features
}

// GetRacialFeatures returns all features for a race
func GetRacialFeatures(raceKey string) []rulebook.CharacterFeature {
	if features, exists := RacialFeatures[raceKey]; exists {
		return features
	}
	return []rulebook.CharacterFeature{}
}

// HasFeature checks if a list of features contains a specific feature key
func HasFeature(features []rulebook.CharacterFeature, key string) bool {
	for _, f := range features {
		if f.Key == key {
			return true
		}
	}
	return false
}
