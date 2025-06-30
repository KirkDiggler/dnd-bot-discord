package character

import (
	"fmt"
	character2 "github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"log"

	"github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/bwmarrin/discordgo"
)

// ClassFeaturesHandler handles class-specific feature selections during character creation
type ClassFeaturesHandler struct {
	characterService character.Service
}

// NewClassFeaturesHandler creates a new class features handler
func NewClassFeaturesHandler(characterService character.Service) *ClassFeaturesHandler {
	return &ClassFeaturesHandler{
		characterService: characterService,
	}
}

// ClassFeaturesRequest contains data for class feature selection
type ClassFeaturesRequest struct {
	Session     *discordgo.Session
	Interaction *discordgo.InteractionCreate
	CharacterID string
	FeatureType string // e.g., "favored_enemy", "divine_domain"
	Selection   string // The selected option
}

// Handle processes class feature selections
func (h *ClassFeaturesHandler) Handle(req *ClassFeaturesRequest) error {
	log.Printf("DEBUG: ClassFeaturesHandler.Handle called with FeatureType=%s, Selection=%s, CharacterID=%s", req.FeatureType, req.Selection, req.CharacterID)

	// Get the character
	char, err := h.characterService.GetByID(req.CharacterID)
	if err != nil {
		return fmt.Errorf("failed to get character: %w", err)
	}
	log.Printf("DEBUG: Got character %s with %d features", char.Name, len(char.Features))

	// Store the selection based on feature type
	switch req.FeatureType {
	case "favored_enemy":
		err = h.handleFavoredEnemy(req, char)
	case "natural_explorer":
		err = h.handleNaturalExplorer(req, char)
	case "fighting_style":
		err = h.handleFightingStyle(req, char)
	// Add other class features here (divine_domain, patron, etc.)
	default:
		err = fmt.Errorf("unknown feature type: %s", req.FeatureType)
	}

	if err != nil {
		return err
	}

	// Save the character (using UpdateEquipment which saves the whole character)
	log.Printf("DEBUG: Saving character %s with %d features", char.Name, len(char.Features))
	for _, feature := range char.Features {
		if feature.Key == "fighting_style" {
			log.Printf("DEBUG: Fighting style feature metadata: %v", feature.Metadata)
		}
	}
	if err := h.characterService.UpdateEquipment(char); err != nil {
		log.Printf("DEBUG: Error saving character: %v", err)
		return fmt.Errorf("failed to update character: %w", err)
	}
	log.Printf("DEBUG: Character saved successfully")

	return nil
}

// handleFavoredEnemy stores the ranger's favored enemy selection
func (h *ClassFeaturesHandler) handleFavoredEnemy(req *ClassFeaturesRequest, char *character2.Character) error {
	// Find the favored enemy feature and update its metadata
	for _, feature := range char.Features {
		if feature.Key == "favored_enemy" {
			if feature.Metadata == nil {
				feature.Metadata = make(map[string]any)
			}
			feature.Metadata["enemy_type"] = req.Selection

			// Log for debugging
			log.Printf("Set favored enemy for %s to %s", char.Name, req.Selection)
			break
		}
	}

	return nil
}

// handleNaturalExplorer stores the ranger's natural explorer terrain selection
func (h *ClassFeaturesHandler) handleNaturalExplorer(req *ClassFeaturesRequest, char *character2.Character) error {
	// Find the natural explorer feature and update its metadata
	for _, feature := range char.Features {
		if feature.Key == "natural_explorer" {
			if feature.Metadata == nil {
				feature.Metadata = make(map[string]any)
			}
			feature.Metadata["terrain_type"] = req.Selection

			// Log for debugging
			log.Printf("Set natural explorer terrain for %s to %s", char.Name, req.Selection)
			break
		}
	}

	return nil
}

// handleFightingStyle stores the fighter's fighting style selection
func (h *ClassFeaturesHandler) handleFightingStyle(req *ClassFeaturesRequest, char *character2.Character) error {
	log.Printf("DEBUG: handleFightingStyle called with selection: %s", req.Selection)

	// Find the fighting style feature and update its metadata
	found := false
	for _, feature := range char.Features {
		log.Printf("DEBUG: Checking feature %s (key=%s)", feature.Name, feature.Key)
		if feature.Key == "fighting_style" {
			found = true
			log.Printf("DEBUG: Found fighting_style feature, current metadata: %v", feature.Metadata)
			if feature.Metadata == nil {
				log.Printf("DEBUG: Creating new metadata map")
				feature.Metadata = make(map[string]any)
			}
			log.Printf("DEBUG: Setting style to: %s", req.Selection)
			feature.Metadata["style"] = req.Selection
			log.Printf("DEBUG: Metadata after setting: %v", feature.Metadata)

			// Log for debugging
			log.Printf("Set fighting style for %s to %s", char.Name, req.Selection)
			break
		}
	}

	if !found {
		log.Printf("DEBUG: ERROR - fighting_style feature not found!")
	}

	return nil
}

// InteractionRequest contains basic interaction data
type InteractionRequest struct {
	Session     *discordgo.Session
	Interaction *discordgo.InteractionCreate
	CharacterID string
}

// ShowFavoredEnemySelection displays the favored enemy selection UI
func (h *ClassFeaturesHandler) ShowFavoredEnemySelection(req *InteractionRequest) error {
	// Get the character
	char, err := h.characterService.GetByID(req.CharacterID)
	if err != nil {
		return fmt.Errorf("failed to get character: %w", err)
	}

	// Enemy type options
	enemyOptions := []discordgo.SelectMenuOption{
		{Label: "Aberrations", Value: "aberrations", Description: "Beholders, mind flayers, etc."},
		{Label: "Beasts", Value: "beasts", Description: "Bears, wolves, dire animals"},
		{Label: "Celestials", Value: "celestials", Description: "Angels, pegasi, unicorns"},
		{Label: "Constructs", Value: "constructs", Description: "Golems, animated objects"},
		{Label: "Dragons", Value: "dragons", Description: "True dragons and dragonkin"},
		{Label: "Elementals", Value: "elementals", Description: "Creatures from elemental planes"},
		{Label: "Fey", Value: "fey", Description: "Sprites, dryads, pixies"},
		{Label: "Fiends", Value: "fiends", Description: "Devils, demons, yugoloths"},
		{Label: "Giants", Value: "giants", Description: "Hill giants, storm giants, ogres"},
		{Label: "Monstrosities", Value: "monstrosities", Description: "Griffons, hydras, owlbears"},
		{Label: "Oozes", Value: "oozes", Description: "Black puddings, gelatinous cubes"},
		{Label: "Plants", Value: "plants", Description: "Shambling mounds, treants"},
		{Label: "Undead", Value: "undead", Description: "Zombies, skeletons, vampires"},
		{Label: "Two Humanoid Races", Value: "humanoids", Description: "Orcs & goblins, elves & dwarves, etc."},
	}

	// Create embed
	embed := &discordgo.MessageEmbed{
		Title:       "Choose Your Favored Enemy",
		Description: fmt.Sprintf("**%s**, as a Ranger you have studied a specific type of enemy extensively.\n\nYou gain advantage on Wisdom (Survival) checks to track them and Intelligence checks to recall information about them.", char.Name),
		Color:       0x228B22, // Forest Green
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Your Selection",
				Value:  "Choose the type of creature you have dedicated yourself to hunting.",
				Inline: false,
			},
		},
	}

	// Add progress field
	classKey := ""
	if char.Class != nil {
		classKey = char.Class.Key
	}
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "Progress",
		Value:  BuildProgressValue(classKey, "class_features"),
		Inline: false,
	})

	// Create components
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    fmt.Sprintf("character_create:class_features:%s:favored_enemy", char.ID),
					Placeholder: "Select your favored enemy...",
					Options:     enemyOptions,
				},
			},
		},
	}

	// Update the interaction response
	return req.Session.InteractionRespond(req.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
			Flags:      discordgo.MessageFlagsEphemeral,
		},
	})
}

// ShowNaturalExplorerSelection displays the natural explorer terrain selection UI
func (h *ClassFeaturesHandler) ShowNaturalExplorerSelection(req *InteractionRequest) error {
	// Get the character
	char, err := h.characterService.GetByID(req.CharacterID)
	if err != nil {
		return fmt.Errorf("failed to get character: %w", err)
	}

	// Terrain options
	terrainOptions := []discordgo.SelectMenuOption{
		{Label: "Arctic", Value: "arctic", Description: "Frozen tundra and icy wastes"},
		{Label: "Coast", Value: "coast", Description: "Beaches, cliffs, and shores"},
		{Label: "Desert", Value: "desert", Description: "Sandy and rocky badlands"},
		{Label: "Forest", Value: "forest", Description: "Woodlands and jungles"},
		{Label: "Grassland", Value: "grassland", Description: "Plains, savannas, and meadows"},
		{Label: "Mountain", Value: "mountain", Description: "Hills and peaks"},
		{Label: "Swamp", Value: "swamp", Description: "Marshes, bogs, and fens"},
		{Label: "Underdark", Value: "underdark", Description: "Caves and underground"},
	}

	// Create embed
	embed := &discordgo.MessageEmbed{
		Title:       "Choose Your Favored Terrain",
		Description: fmt.Sprintf("**%s**, you are particularly familiar with one type of natural environment.\n\nWhile in your favored terrain:\n• Your group can't become lost except by magical means\n• You remain alert while tracking, foraging, or navigating\n• You can move stealthily at normal pace when alone\n• You find food and water for up to 6 people daily", char.Name),
		Color:       0x228B22, // Forest Green
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Your Selection",
				Value:  "Choose the terrain where you feel most at home.",
				Inline: false,
			},
		},
	}

	// Add progress field
	classKey := ""
	if char.Class != nil {
		classKey = char.Class.Key
	}
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "Progress",
		Value:  BuildProgressValue(classKey, "class_features"),
		Inline: false,
	})

	// Create components
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    fmt.Sprintf("character_create:class_features:%s:natural_explorer", char.ID),
					Placeholder: "Select your favored terrain...",
					Options:     terrainOptions,
				},
			},
		},
	}

	// Update the interaction response
	return req.Session.InteractionRespond(req.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
			Flags:      discordgo.MessageFlagsEphemeral,
		},
	})
}

// ShouldShowClassFeatures checks if a character needs to select class features
func (h *ClassFeaturesHandler) ShouldShowClassFeatures(char *character2.Character) (needsSelection bool, featureType string) {
	if char.Class == nil {
		return false, ""
	}

	// Check each class for required selections
	switch char.Class.Key {
	case "ranger":
		// Check if ranger has selected favored enemy
		for _, feature := range char.Features {
			if feature.Key == "favored_enemy" {
				if feature.Metadata == nil || feature.Metadata["enemy_type"] == nil {
					return true, "favored_enemy"
				}
			}
			if feature.Key == "natural_explorer" {
				if feature.Metadata == nil || feature.Metadata["terrain_type"] == nil {
					return true, "natural_explorer"
				}
			}
		}
	// Add other classes with level 1 choices here
	case "cleric":
		// TODO: Check for divine domain selection
	case "fighter":
		// Check if fighter has selected fighting style
		for _, feature := range char.Features {
			if feature.Key == "fighting_style" {
				if feature.Metadata == nil || feature.Metadata["style"] == nil {
					return true, "fighting_style"
				}
			}
		}
	case "warlock":
		// TODO: Check for patron selection
	case "sorcerer":
		// TODO: Check for sorcerous origin selection
	}

	return false, ""
}

// ShowFightingStyleSelection displays the fighting style selection UI
func (h *ClassFeaturesHandler) ShowFightingStyleSelection(req *InteractionRequest) error {
	// Get the character
	char, err := h.characterService.GetByID(req.CharacterID)
	if err != nil {
		return fmt.Errorf("failed to get character: %w", err)
	}

	// Fighting style options
	styleOptions := []discordgo.SelectMenuOption{
		{Label: "Archery", Value: "archery", Description: "+2 to attack rolls with ranged weapons"},
		{Label: "Defense", Value: "defense", Description: "+1 to AC while wearing armor"},
		{Label: "Dueling", Value: "dueling", Description: "+2 damage with one-handed weapons"},
		{Label: "Great Weapon Fighting", Value: "great_weapon", Description: "Reroll 1s and 2s on damage with two-handed weapons"},
		{Label: "Protection", Value: "protection", Description: "Use shield to impose disadvantage on attacks"},
		{Label: "Two-Weapon Fighting", Value: "two_weapon", Description: "Add modifier to off-hand damage"},
	}

	// Create embed
	embed := &discordgo.MessageEmbed{
		Title:       "Choose Your Fighting Style",
		Description: fmt.Sprintf("**%s**, as a Fighter you adopt a particular style of fighting as your specialty.\n\nChoose the fighting style that matches your combat approach.", char.Name),
		Color:       0x722F37, // Dark red for fighter
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Your Selection",
				Value:  "Choose the fighting style that defines your combat technique.",
				Inline: false,
			},
		},
	}

	// Add progress field
	classKey := ""
	if char.Class != nil {
		classKey = char.Class.Key
	}
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "Progress",
		Value:  BuildProgressValue(classKey, "class_features"),
		Inline: false,
	})

	// Create components
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    fmt.Sprintf("character_create:class_features:%s:fighting_style", char.ID),
					Placeholder: "Select your fighting style...",
					Options:     styleOptions,
				},
			},
		},
	}

	// Update the interaction response
	return req.Session.InteractionRespond(req.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
			Flags:      discordgo.MessageFlagsEphemeral,
		},
	})
}
