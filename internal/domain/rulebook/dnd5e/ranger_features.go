package rulebook

// GetFavoredEnemyChoice returns a FeatureChoice for ranger's favored enemy selection
func GetFavoredEnemyChoice() *FeatureChoice {
	return &FeatureChoice{
		Type:        FeatureChoiceTypeFavoredEnemy,
		FeatureKey:  "favored_enemy",
		Name:        "Favored Enemy",
		Description: "You have significant experience studying, tracking, hunting, and even talking to a certain type of enemy. You gain advantage on Wisdom (Survival) checks to track your favored enemies, as well as on Intelligence checks to recall information about them.",
		Choose:      1,
		Options: []FeatureOption{
			{Key: "aberrations", Name: "Aberrations", Description: "Beholders, mind flayers, and other alien creatures"},
			{Key: "beasts", Name: "Beasts", Description: "Bears, wolves, dire animals, and other natural creatures"},
			{Key: "celestials", Name: "Celestials", Description: "Angels, pegasi, unicorns, and other heavenly beings"},
			{Key: "constructs", Name: "Constructs", Description: "Golems, animated objects, and other artificial creatures"},
			{Key: "dragons", Name: "Dragons", Description: "True dragons and dragonkin"},
			{Key: "elementals", Name: "Elementals", Description: "Creatures from the elemental planes"},
			{Key: "fey", Name: "Fey", Description: "Sprites, dryads, pixies, and other fey creatures"},
			{Key: "fiends", Name: "Fiends", Description: "Devils, demons, yugoloths, and other evil outsiders"},
			{Key: "giants", Name: "Giants", Description: "Hill giants, storm giants, ogres, and other large humanoids"},
			{Key: "monstrosities", Name: "Monstrosities", Description: "Griffons, hydras, owlbears, and other unnatural creatures"},
			{Key: "oozes", Name: "Oozes", Description: "Black puddings, gelatinous cubes, and other amorphous creatures"},
			{Key: "plants", Name: "Plants", Description: "Shambling mounds, treants, and other plant creatures"},
			{Key: "undead", Name: "Undead", Description: "Zombies, skeletons, vampires, and other undead"},
			{Key: "humanoids", Name: "Two Humanoid Races", Description: "Choose two races of humanoid (such as gnolls and orcs)"},
		},
	}
}

// GetNaturalExplorerChoice returns a FeatureChoice for ranger's natural explorer selection
func GetNaturalExplorerChoice() *FeatureChoice {
	return &FeatureChoice{
		Type:        FeatureChoiceTypeNaturalExplorer,
		FeatureKey:  "natural_explorer",
		Name:        "Natural Explorer",
		Description: "You are particularly familiar with one type of natural environment and are adept at traveling and surviving in such regions. While traveling in your favored terrain, you gain various benefits.",
		Choose:      1,
		Options: []FeatureOption{
			{Key: "arctic", Name: "Arctic", Description: "Frozen tundra, glaciers, and icy wastes"},
			{Key: "coast", Name: "Coast", Description: "Beaches, cliffs, shores, and islands"},
			{Key: "desert", Name: "Desert", Description: "Sandy and rocky badlands, hot and cold deserts"},
			{Key: "forest", Name: "Forest", Description: "Woodlands, jungles, and rainforests"},
			{Key: "grassland", Name: "Grassland", Description: "Plains, savannas, meadows, and prairies"},
			{Key: "mountain", Name: "Mountain", Description: "Hills, peaks, and alpine regions"},
			{Key: "swamp", Name: "Swamp", Description: "Marshes, bogs, fens, and moors"},
			{Key: "underdark", Name: "Underdark", Description: "Caves, caverns, and underground passages"},
		},
	}
}
