package features

import "github.com/KirkDiggler/dnd-bot-discord/internal/entities"

// ClassFeatures defines the special features for each class
var ClassFeatures = map[string][]entities.CharacterFeature{
	"monk": {
		{
			Key:         "unarmored_defense_monk",
			Name:        "Unarmored Defense",
			Description: "While wearing no armor and not wielding a shield, your AC equals 10 + your Dexterity modifier + your Wisdom modifier.",
			Type:        entities.FeatureTypeClass,
			Level:       1,
			Source:      "Monk",
		},
		{
			Key:         "martial-arts",
			Name:        "Martial Arts",
			Description: "You can use Dexterity instead of Strength for attack and damage rolls of unarmed strikes and monk weapons. Your unarmed strikes use a d4 for damage. This die changes as you gain monk levels: d6 at 5th level, d8 at 11th level, and d10 at 17th level. When you use the Attack action, you can make one unarmed strike as a bonus action.",
			Type:        entities.FeatureTypeClass,
			Level:       1,
			Source:      "Monk",
		},
	},
	"barbarian": {
		{
			Key:         "unarmored_defense_barbarian",
			Name:        "Unarmored Defense",
			Description: "While not wearing armor, your AC equals 10 + your Dexterity modifier + your Constitution modifier. You can use a shield and still gain this benefit.",
			Type:        entities.FeatureTypeClass,
			Level:       1,
			Source:      "Barbarian",
		},
		{
			Key:         "rage",
			Name:        "Rage",
			Description: "In battle, you fight with primal ferocity. As a bonus action, you can enter a rage for 1 minute.",
			Type:        entities.FeatureTypeClass,
			Level:       1,
			Source:      "Barbarian",
		},
	},
	"wizard": {
		{
			Key:         "spellcasting_wizard",
			Name:        "Spellcasting",
			Description: "You have learned to cast spells. See chapter 10 for the general rules of spellcasting and chapter 11 for the wizard spell list.",
			Type:        entities.FeatureTypeClass,
			Level:       1,
			Source:      "Wizard",
		},
		{
			Key:         "arcane_recovery",
			Name:        "Arcane Recovery",
			Description: "Once per day during a short rest, you can recover expended spell slots.",
			Type:        entities.FeatureTypeClass,
			Level:       1,
			Source:      "Wizard",
		},
	},
	"fighter": {
		{
			Key:         "fighting_style",
			Name:        "Fighting Style",
			Description: "You adopt a particular style of fighting as your specialty.",
			Type:        entities.FeatureTypeClass,
			Level:       1,
			Source:      "Fighter",
		},
		{
			Key:         "second_wind",
			Name:        "Second Wind",
			Description: "You have a limited well of stamina that you can draw on to protect yourself from harm.",
			Type:        entities.FeatureTypeClass,
			Level:       1,
			Source:      "Fighter",
		},
	},
	"rogue": {
		{
			Key:         "expertise",
			Name:        "Expertise",
			Description: "Choose two of your skill proficiencies. Your proficiency bonus is doubled for any ability check you make using either of the chosen proficiencies.",
			Type:        entities.FeatureTypeClass,
			Level:       1,
			Source:      "Rogue",
		},
		{
			Key:         "sneak_attack",
			Name:        "Sneak Attack",
			Description: "You know how to strike subtly and exploit a foe's distraction. Once per turn, you can deal an extra 1d6 damage.",
			Type:        entities.FeatureTypeClass,
			Level:       1,
			Source:      "Rogue",
		},
		{
			Key:         "thieves_cant",
			Name:        "Thieves' Cant",
			Description: "You have learned thieves' cant, a secret mix of dialect, jargon, and code.",
			Type:        entities.FeatureTypeClass,
			Level:       1,
			Source:      "Rogue",
		},
	},
	"cleric": {
		{
			Key:         "spellcasting_cleric",
			Name:        "Spellcasting",
			Description: "As a conduit for divine power, you can cast cleric spells.",
			Type:        entities.FeatureTypeClass,
			Level:       1,
			Source:      "Cleric",
		},
		{
			Key:         "divine_domain",
			Name:        "Divine Domain",
			Description: "Choose one domain related to your deity. Your choice grants you domain spells and other features.",
			Type:        entities.FeatureTypeClass,
			Level:       1,
			Source:      "Cleric",
		},
	},
	"ranger": {
		{
			Key:         "favored_enemy",
			Name:        "Favored Enemy",
			Description: "You have significant experience studying, tracking, hunting, and even talking to a certain type of enemy. Choose a type of favored enemy: aberrations, beasts, celestials, constructs, dragons, elementals, fey, fiends, giants, monstrosities, oozes, plants, or undead. Alternatively, you can select two races of humanoid (such as gnolls and orcs) as favored enemies. You have advantage on Wisdom (Survival) checks to track your favored enemies, as well as on Intelligence checks to recall information about them.",
			Type:        entities.FeatureTypeClass,
			Level:       1,
			Source:      "Ranger",
		},
		{
			Key:         "natural_explorer",
			Name:        "Natural Explorer",
			Description: "You are particularly familiar with one type of natural environment and are adept at traveling and surviving in such regions. Choose one type of favored terrain: arctic, coast, desert, forest, grassland, mountain, swamp, or the Underdark. When traveling for an hour or more in your favored terrain, you gain benefits including: difficult terrain doesn't slow your party's travel, your group can't become lost except by magical means, you remain alert to danger while tracking/foraging/navigating, you can move stealthily at a normal pace when alone, you find food and water for up to 6 people daily, and you can track creatures while moving at a fast pace.",
			Type:        entities.FeatureTypeClass,
			Level:       1,
			Source:      "Ranger",
		},
	},
}

// RacialFeatures defines features for each race
var RacialFeatures = map[string][]entities.CharacterFeature{
	"dragonborn": {
		{
			Key:         "draconic_ancestry",
			Name:        "Draconic Ancestry",
			Description: "You have draconic ancestry. Choose one type of dragon from the Draconic Ancestry table.",
			Type:        entities.FeatureTypeRacial,
			Level:       0,
			Source:      "Dragonborn",
		},
		{
			Key:         "breath_weapon",
			Name:        "Breath Weapon",
			Description: "You can use your action to exhale destructive energy based on your draconic ancestry.",
			Type:        entities.FeatureTypeRacial,
			Level:       0,
			Source:      "Dragonborn",
		},
		{
			Key:         "damage_resistance",
			Name:        "Damage Resistance",
			Description: "You have resistance to the damage type associated with your draconic ancestry.",
			Type:        entities.FeatureTypeRacial,
			Level:       0,
			Source:      "Dragonborn",
		},
	},
	"elf": {
		{
			Key:         "darkvision",
			Name:        "Darkvision",
			Description: "You can see in dim light within 60 feet as if it were bright light.",
			Type:        entities.FeatureTypeRacial,
			Level:       0,
			Source:      "Elf",
		},
		{
			Key:         "keen_senses",
			Name:        "Keen Senses",
			Description: "You have proficiency in the Perception skill.",
			Type:        entities.FeatureTypeRacial,
			Level:       0,
			Source:      "Elf",
		},
		{
			Key:         "fey_ancestry",
			Name:        "Fey Ancestry",
			Description: "You have advantage on saving throws against being charmed, and magic can't put you to sleep.",
			Type:        entities.FeatureTypeRacial,
			Level:       0,
			Source:      "Elf",
		},
		{
			Key:         "trance",
			Name:        "Trance",
			Description: "Elves don't need to sleep. Instead, they meditate deeply for 4 hours a day.",
			Type:        entities.FeatureTypeRacial,
			Level:       0,
			Source:      "Elf",
		},
	},
	"dwarf": {
		{
			Key:         "darkvision",
			Name:        "Darkvision",
			Description: "You can see in dim light within 60 feet as if it were bright light.",
			Type:        entities.FeatureTypeRacial,
			Level:       0,
			Source:      "Dwarf",
		},
		{
			Key:         "dwarven_resilience",
			Name:        "Dwarven Resilience",
			Description: "You have advantage on saving throws against poison, and resistance against poison damage.",
			Type:        entities.FeatureTypeRacial,
			Level:       0,
			Source:      "Dwarf",
		},
		{
			Key:         "stonecunning",
			Name:        "Stonecunning",
			Description: "Whenever you make an Intelligence (History) check related to stonework, you are considered proficient and add double your proficiency bonus.",
			Type:        entities.FeatureTypeRacial,
			Level:       0,
			Source:      "Dwarf",
		},
	},
	"halfling": {
		{
			Key:         "lucky",
			Name:        "Lucky",
			Description: "When you roll a 1 on the d20 for an attack roll, ability check, or saving throw, you can reroll the die and must use the new roll.",
			Type:        entities.FeatureTypeRacial,
			Level:       0,
			Source:      "Halfling",
		},
		{
			Key:         "brave",
			Name:        "Brave",
			Description: "You have advantage on saving throws against being frightened.",
			Type:        entities.FeatureTypeRacial,
			Level:       0,
			Source:      "Halfling",
		},
		{
			Key:         "halfling_nimbleness",
			Name:        "Halfling Nimbleness",
			Description: "You can move through the space of any creature that is of a size larger than yours.",
			Type:        entities.FeatureTypeRacial,
			Level:       0,
			Source:      "Halfling",
		},
	},
	"human": {
		{
			Key:         "human_versatility",
			Name:        "Versatility",
			Description: "Humans gain +1 to all ability scores.",
			Type:        entities.FeatureTypeRacial,
			Level:       0,
			Source:      "Human",
		},
	},
	"tiefling": {
		{
			Key:         "darkvision",
			Name:        "Darkvision",
			Description: "You can see in dim light within 60 feet as if it were bright light.",
			Type:        entities.FeatureTypeRacial,
			Level:       0,
			Source:      "Tiefling",
		},
		{
			Key:         "hellish_resistance",
			Name:        "Hellish Resistance",
			Description: "You have resistance to fire damage.",
			Type:        entities.FeatureTypeRacial,
			Level:       0,
			Source:      "Tiefling",
		},
		{
			Key:         "infernal_legacy",
			Name:        "Infernal Legacy",
			Description: "You know the thaumaturgy cantrip. At 3rd level, you can cast hellish rebuke once per day. At 5th level, you can cast darkness once per day.",
			Type:        entities.FeatureTypeRacial,
			Level:       0,
			Source:      "Tiefling",
		},
	},
}

// GetClassFeatures returns features for a class at a given level
func GetClassFeatures(classKey string, level int) []entities.CharacterFeature {
	features := []entities.CharacterFeature{}
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
func GetRacialFeatures(raceKey string) []entities.CharacterFeature {
	if features, exists := RacialFeatures[raceKey]; exists {
		return features
	}
	return []entities.CharacterFeature{}
}

// HasFeature checks if a list of features contains a specific feature key
func HasFeature(features []entities.CharacterFeature, key string) bool {
	for _, f := range features {
		if f.Key == key {
			return true
		}
	}
	return false
}
