package rulebook

// DivineDomain represents a cleric's divine domain
type DivineDomain struct {
	Key          string           `json:"key"`
	Name         string           `json:"name"`
	Description  string           `json:"description"`
	DomainSpells map[int][]string `json:"domain_spells"` // Spell keys by spell level
	Features     []DomainFeature  `json:"features"`      // Additional features granted
}

// DomainFeature represents a feature granted by a divine domain
type DomainFeature struct {
	Level       int    `json:"level"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// GetDivineDomains returns all PHB divine domains
func GetDivineDomains() []DivineDomain {
	return []DivineDomain{
		{
			Key:         "knowledge",
			Name:        "Knowledge Domain",
			Description: "The gods of knowledge value learning and understanding above all. You gain proficiency in two skills and languages.",
			DomainSpells: map[int][]string{
				1: {"command", "identify"},
				3: {"augury", "suggestion"},
				5: {"nondetection", "speak-with-dead"},
				7: {"arcane-eye", "confusion"},
				9: {"legend-lore", "scrying"},
			},
		},
		{
			Key:         "life",
			Name:        "Life Domain",
			Description: "The Life domain focuses on the vibrant positive energy that sustains all life. You gain heavy armor proficiency and improved healing.",
			DomainSpells: map[int][]string{
				1: {"bless", "cure-wounds"},
				3: {"lesser-restoration", "spiritual-weapon"},
				5: {"beacon-of-hope", "revivify"},
				7: {"death-ward", "guardian-of-faith"},
				9: {"mass-cure-wounds", "raise-dead"},
			},
		},
		{
			Key:         "light",
			Name:        "Light Domain",
			Description: "Gods of light promote the ideals of rebirth, renewal, truth, vigilance, and beauty. You gain the light cantrip and defensive abilities.",
			DomainSpells: map[int][]string{
				1: {"burning-hands", "faerie-fire"},
				3: {"flaming-sphere", "scorching-ray"},
				5: {"daylight", "fireball"},
				7: {"guardian-of-faith", "wall-of-fire"},
				9: {"flame-strike", "scrying"},
			},
		},
		{
			Key:         "nature",
			Name:        "Nature Domain",
			Description: "Gods of nature are as varied as the natural world itself. You gain a druid cantrip and proficiency with heavy armor.",
			DomainSpells: map[int][]string{
				1: {"animal-friendship", "speak-with-animals"},
				3: {"barkskin", "spike-growth"},
				5: {"plant-growth", "wind-wall"},
				7: {"dominate-beast", "grasping-vine"},
				9: {"insect-plague", "tree-stride"},
			},
		},
		{
			Key:         "tempest",
			Name:        "Tempest Domain",
			Description: "Gods of the tempest govern storms, sea, and sky. You gain proficiency with martial weapons and heavy armor.",
			DomainSpells: map[int][]string{
				1: {"fog-cloud", "thunderwave"},
				3: {"gust-of-wind", "shatter"},
				5: {"call-lightning", "sleet-storm"},
				7: {"control-water", "ice-storm"},
				9: {"destructive-wave", "insect-plague"},
			},
		},
		{
			Key:         "trickery",
			Name:        "Trickery Domain",
			Description: "Gods of trickery are mischief-makers and instigators. You can use your Channel Divinity to create an illusory duplicate.",
			DomainSpells: map[int][]string{
				1: {"charm-person", "disguise-self"},
				3: {"mirror-image", "pass-without-trace"},
				5: {"blink", "dispel-magic"},
				7: {"dimension-door", "polymorph"},
				9: {"dominate-person", "modify-memory"},
			},
		},
		{
			Key:         "war",
			Name:        "War Domain",
			Description: "Gods of war watch over warriors and reward acts of violence. You gain proficiency with martial weapons and heavy armor.",
			DomainSpells: map[int][]string{
				1: {"divine-favor", "shield-of-faith"},
				3: {"magic-weapon", "spiritual-weapon"},
				5: {"crusaders-mantle", "spirit-guardians"},
				7: {"freedom-of-movement", "stoneskin"},
				9: {"flame-strike", "hold-monster"},
			},
		},
	}
}

// GetDivineDomainChoice returns a FeatureChoice for divine domain selection
func GetDivineDomainChoice() *FeatureChoice {
	domains := GetDivineDomains()
	options := make([]FeatureOption, len(domains))

	for i, domain := range domains {
		options[i] = FeatureOption{
			Key:         domain.Key,
			Name:        domain.Name,
			Description: domain.Description,
			Metadata: map[string]any{
				"domain_spells": domain.DomainSpells,
			},
		}
	}

	return &FeatureChoice{
		Type:        FeatureChoiceTypeDivineDomain,
		FeatureKey:  "divine_domain",
		Name:        "Divine Domain",
		Description: "Choose one domain related to your deity. Your choice grants you domain spells and other features.",
		Choose:      1,
		Options:     options,
	}
}
