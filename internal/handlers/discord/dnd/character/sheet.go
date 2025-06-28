package character

import (
	"fmt"
	"log"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/handlers/discord/utils"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services"
	"github.com/bwmarrin/discordgo"
)

// SheetHandler handles the character sheet display command
type SheetHandler struct {
	services *services.Provider
}

// NewSheetHandler creates a new sheet handler
func NewSheetHandler(serviceProvider *services.Provider) *SheetHandler {
	return &SheetHandler{
		services: serviceProvider,
	}
}

// Handle processes the character sheet command
func (h *SheetHandler) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// Get character ID from options safely
	characterID := utils.GetStringOption(i, "character_id")
	if characterID == "" {
		return respondWithError(s, i, "Character ID is required")
	}

	// Get the character from service (all business logic is in the service)
	character, err := h.services.CharacterService.GetByID(characterID)
	if err != nil {
		log.Printf("Error getting character %s: %v", characterID, err)
		return respondWithError(s, i, "Character not found")
	}

	// Verify ownership
	if character.OwnerID != i.Member.User.ID {
		return respondWithError(s, i, "You can only view your own characters!")
	}

	// Build the character sheet embed (pure presentation logic)
	embed := BuildCharacterSheetEmbed(character)

	// Build interactive components
	components := BuildCharacterSheetComponents(characterID)

	// Send ephemeral response
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
			Flags:      discordgo.MessageFlagsEphemeral, // Only visible to the user
		},
	})
}

// BuildCharacterSheetEmbed creates the main character sheet embed
func BuildCharacterSheetEmbed(char *entities.Character) *discordgo.MessageEmbed {
	// Build title
	title := fmt.Sprintf("%s - Level %d %s", char.Name, char.Level, char.Class.Name)
	if char.Race != nil {
		title = fmt.Sprintf("%s %s", char.Race.Name, title)
	}

	// Build HP/AC line
	hpAcLine := fmt.Sprintf("**HP:** %d/%d | **AC:** %d",
		char.CurrentHitPoints, char.MaxHitPoints, char.AC)

	// Calculate initiative bonus
	initiativeBonus := 0
	if dex, exists := char.Attributes[entities.AttributeDexterity]; exists {
		initiativeBonus = dex.Bonus
	}
	hpAcLine += fmt.Sprintf(" | **Initiative:** %+d", initiativeBonus)

	// Build ability scores
	abilityLines := []string{
		"**Physical:**",
		buildAbilityLine("STR", char.Attributes[entities.AttributeStrength]),
		buildAbilityLine("DEX", char.Attributes[entities.AttributeDexterity]),
		buildAbilityLine("CON", char.Attributes[entities.AttributeConstitution]),
		"",
		"**Mental:**",
		buildAbilityLine("INT", char.Attributes[entities.AttributeIntelligence]),
		buildAbilityLine("WIS", char.Attributes[entities.AttributeWisdom]),
		buildAbilityLine("CHA", char.Attributes[entities.AttributeCharisma]),
	}

	// Build equipment display
	equipmentLines := buildEquipmentDisplay(char)

	// Build proficiencies summary
	proficiencyLines := buildProficiencySummary(char)

	// Build features summary
	featureLines := buildFeatureSummary(char)

	embed := &discordgo.MessageEmbed{
		Title:       title,
		Description: hpAcLine,
		Color:       0x3498db, // Blue
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "📊 Ability Scores",
				Value:  strings.Join(abilityLines, "\n"),
				Inline: true,
			},
			{
				Name:   "⚔️ Equipment",
				Value:  strings.Join(equipmentLines, "\n"),
				Inline: true,
			},
			{
				Name:   "📚 Proficiencies",
				Value:  strings.Join(proficiencyLines, "\n"),
				Inline: false,
			},
			{
				Name:   "✨ Features",
				Value:  strings.Join(featureLines, "\n"),
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Character ID: %s", char.ID),
		},
	}

	return embed
}

// buildAbilityLine formats a single ability score line
func buildAbilityLine(name string, score *entities.AbilityScore) string {
	if score == nil {
		return fmt.Sprintf("**%s:** 10 (+0)", name)
	}
	return fmt.Sprintf("**%s:** %d (%+d)", name, score.Score, score.Bonus)
}

// buildEquipmentDisplay builds the equipment section
func buildEquipmentDisplay(char *entities.Character) []string {
	lines := []string{}

	// Main hand
	if weapon := char.EquippedSlots[entities.SlotMainHand]; weapon != nil {
		lines = append(lines, fmt.Sprintf("**Main Hand:** %s", weapon.GetName()))
	} else {
		lines = append(lines, "**Main Hand:** Empty")
	}

	// Off hand
	if item := char.EquippedSlots[entities.SlotOffHand]; item != nil {
		lines = append(lines, fmt.Sprintf("**Off Hand:** %s", item.GetName()))
	} else {
		lines = append(lines, "**Off Hand:** Empty")
	}

	// Two-handed (only show if no main/off hand)
	if char.EquippedSlots[entities.SlotMainHand] == nil && char.EquippedSlots[entities.SlotOffHand] == nil {
		if weapon := char.EquippedSlots[entities.SlotTwoHanded]; weapon != nil {
			lines = append(lines, fmt.Sprintf("**Two-Handed:** %s", weapon.GetName()))
		}
	}

	// Armor
	if armor := char.EquippedSlots[entities.SlotBody]; armor != nil {
		lines = append(lines, fmt.Sprintf("**Armor:** %s", armor.GetName()))
	} else {
		lines = append(lines, "**Armor:** Empty")
	}

	return lines
}

// buildProficiencySummary builds a summary of proficiencies
func buildProficiencySummary(char *entities.Character) []string {
	lines := []string{}

	// Weapon proficiencies
	if weapons, exists := char.Proficiencies[entities.ProficiencyTypeWeapon]; exists && len(weapons) > 0 {
		weaponNames := []string{}
		for _, prof := range weapons {
			weaponNames = append(weaponNames, prof.Name)
		}
		lines = append(lines, fmt.Sprintf("**Weapons:** %s", strings.Join(weaponNames, ", ")))
	}

	// Armor proficiencies
	if armors, exists := char.Proficiencies[entities.ProficiencyTypeArmor]; exists && len(armors) > 0 {
		armorNames := []string{}
		for _, prof := range armors {
			armorNames = append(armorNames, prof.Name)
		}
		lines = append(lines, fmt.Sprintf("**Armor:** %s", strings.Join(armorNames, ", ")))
	}

	// Skill proficiencies
	if skills, exists := char.Proficiencies[entities.ProficiencyTypeSkill]; exists && len(skills) > 0 {
		skillNames := []string{}
		for _, prof := range skills {
			skillNames = append(skillNames, prof.Name)
		}
		lines = append(lines, fmt.Sprintf("**Skills:** %s", strings.Join(skillNames, ", ")))
	}

	if len(lines) == 0 {
		lines = append(lines, "*No proficiencies*")
	}

	return lines
}

// BuildCharacterSheetComponents builds the interactive buttons
func BuildCharacterSheetComponents(characterID string) []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "View Inventory",
					Style:    discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("character:inventory:%s", characterID),
					Emoji:    &discordgo.ComponentEmoji{Name: "🎒"},
				},
				discordgo.Button{
					Label:    "View Details",
					Style:    discordgo.SecondaryButton,
					CustomID: fmt.Sprintf("character:details:%s", characterID),
					Emoji:    &discordgo.ComponentEmoji{Name: "📋"},
				},
				discordgo.Button{
					Label:    "Refresh",
					Style:    discordgo.SecondaryButton,
					CustomID: fmt.Sprintf("character:sheet_refresh:%s", characterID),
					Emoji:    &discordgo.ComponentEmoji{Name: "🔄"},
				},
			},
		},
	}
}

// buildFeatureSummary builds a summary of character features
func buildFeatureSummary(char *entities.Character) []string {
	lines := []string{}

	if len(char.Features) == 0 {
		lines = append(lines, "*No features*")
		return lines
	}

	// Group features by type
	classFeatures := []*entities.CharacterFeature{}
	racialFeatures := []*entities.CharacterFeature{}
	otherFeatures := []*entities.CharacterFeature{}

	for _, feature := range char.Features {
		if feature == nil {
			continue
		}
		switch feature.Type {
		case entities.FeatureTypeClass:
			classFeatures = append(classFeatures, feature)
		case entities.FeatureTypeRacial:
			racialFeatures = append(racialFeatures, feature)
		default:
			otherFeatures = append(otherFeatures, feature)
		}
	}

	// Add class features
	if len(classFeatures) > 0 {
		lines = append(lines, "**Class Features:**")
		for _, feat := range classFeatures {
			lines = append(lines, fmt.Sprintf("• %s", feat.Name))
		}
	}

	// Add racial features
	if len(racialFeatures) > 0 {
		if len(lines) > 0 {
			lines = append(lines, "") // Add spacing
		}
		lines = append(lines, "**Racial Features:**")
		for _, feat := range racialFeatures {
			lines = append(lines, fmt.Sprintf("• %s", feat.Name))
		}
	}

	// Add other features
	if len(otherFeatures) > 0 {
		if len(lines) > 0 {
			lines = append(lines, "") // Add spacing
		}
		lines = append(lines, "**Other Features:**")
		for _, feat := range otherFeatures {
			lines = append(lines, fmt.Sprintf("• %s", feat.Name))
		}
	}

	return lines
}

// respondWithError sends an error message as ephemeral
func respondWithError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("❌ %s", message),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}
