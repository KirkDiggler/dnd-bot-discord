package character

import (
	"fmt"
	"log"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"

	"github.com/KirkDiggler/dnd-bot-discord/internal/effects"
	"github.com/KirkDiggler/dnd-bot-discord/internal/handlers/discord/utils"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services"
	"github.com/bwmarrin/discordgo"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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
	char, err := h.services.CharacterService.GetByID(characterID)
	if err != nil {
		log.Printf("Error getting character %s: %v", characterID, err)
		return respondWithError(s, i, "Character not found")
	}

	// Verify ownership
	if char.OwnerID != i.Member.User.ID {
		return respondWithError(s, i, "You can only view your own characters!")
	}

	// Build the character sheet embed (pure presentation logic)
	embed := BuildCharacterSheetEmbed(char)

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
func BuildCharacterSheetEmbed(char *character.Character) *discordgo.MessageEmbed {
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
	if dex, exists := char.Attributes[shared.AttributeDexterity]; exists {
		initiativeBonus = dex.Bonus
	}
	hpAcLine += fmt.Sprintf(" | **Initiative:** %+d", initiativeBonus)

	// Build ability scores
	abilityLines := []string{
		"**Physical:**",
		buildAbilityLine("STR", char.Attributes[shared.AttributeStrength]),
		buildAbilityLine("DEX", char.Attributes[shared.AttributeDexterity]),
		buildAbilityLine("CON", char.Attributes[shared.AttributeConstitution]),
		"",
		"**Mental:**",
		buildAbilityLine("INT", char.Attributes[shared.AttributeIntelligence]),
		buildAbilityLine("WIS", char.Attributes[shared.AttributeWisdom]),
		buildAbilityLine("CHA", char.Attributes[shared.AttributeCharisma]),
	}

	// Build equipment display
	equipmentLines := buildEquipmentDisplay(char)

	// Build proficiencies summary
	proficiencyLines := buildProficiencySummary(char)

	// Build features summary
	featureLines := buildFeatureSummary(char)

	// Build active effects summary
	effectLines := buildActiveEffectsDisplay(char)

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
			{
				Name:   "🔮 Active Effects",
				Value:  strings.Join(effectLines, "\n"),
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
func buildAbilityLine(name string, score *character.AbilityScore) string {
	if score == nil {
		return fmt.Sprintf("**%s:** 10 (+0)", name)
	}
	return fmt.Sprintf("**%s:** %d (%+d)", name, score.Score, score.Bonus)
}

// buildEquipmentDisplay builds the equipment section
func buildEquipmentDisplay(char *character.Character) []string {
	lines := []string{}

	// Main hand
	if weapon := char.EquippedSlots[shared.SlotMainHand]; weapon != nil {
		lines = append(lines, fmt.Sprintf("**Main Hand:** %s", weapon.GetName()))
	} else {
		lines = append(lines, "**Main Hand:** Empty")
	}

	// Off hand
	if item := char.EquippedSlots[shared.SlotOffHand]; item != nil {
		lines = append(lines, fmt.Sprintf("**Off Hand:** %s", item.GetName()))
	} else {
		lines = append(lines, "**Off Hand:** Empty")
	}

	// Two-handed (only show if no main/off hand)
	if char.EquippedSlots[shared.SlotMainHand] == nil && char.EquippedSlots[shared.SlotOffHand] == nil {
		if weapon := char.EquippedSlots[shared.SlotTwoHanded]; weapon != nil {
			lines = append(lines, fmt.Sprintf("**Two-Handed:** %s", weapon.GetName()))
		}
	}

	// Armor
	if armor := char.EquippedSlots[shared.SlotBody]; armor != nil {
		lines = append(lines, fmt.Sprintf("**Armor:** %s", armor.GetName()))
	} else {
		lines = append(lines, "**Armor:** Empty")
	}

	return lines
}

// buildProficiencySummary builds a summary of proficiencies
func buildProficiencySummary(char *character.Character) []string {
	lines := []string{}

	// Weapon proficiencies
	if weapons, exists := char.Proficiencies[rulebook.ProficiencyTypeWeapon]; exists && len(weapons) > 0 {
		weaponNames := []string{}
		for _, prof := range weapons {
			weaponNames = append(weaponNames, prof.Name)
		}
		lines = append(lines, fmt.Sprintf("**Weapons:** %s", strings.Join(weaponNames, ", ")))
	}

	// Armor proficiencies
	if armors, exists := char.Proficiencies[rulebook.ProficiencyTypeArmor]; exists && len(armors) > 0 {
		armorNames := []string{}
		for _, prof := range armors {
			armorNames = append(armorNames, prof.Name)
		}
		lines = append(lines, fmt.Sprintf("**Armor:** %s", strings.Join(armorNames, ", ")))
	}

	// Skill proficiencies
	if skills, exists := char.Proficiencies[rulebook.ProficiencyTypeSkill]; exists && len(skills) > 0 {
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
					Label:    "Manage Equipment",
					Style:    discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("character:inventory:%s", characterID),
					Emoji:    &discordgo.ComponentEmoji{Name: "🎒"},
				},
				discordgo.Button{
					Label:    "Edit Character",
					Style:    discordgo.SecondaryButton,
					CustomID: fmt.Sprintf("character:edit_menu:%s", characterID),
					Emoji:    &discordgo.ComponentEmoji{Name: "✏️"},
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
func buildFeatureSummary(char *character.Character) []string {
	lines := []string{}

	if len(char.Features) == 0 {
		lines = append(lines, "*No features*")
		return lines
	}

	// Group features by type
	classFeatures := []*rulebook.CharacterFeature{}
	racialFeatures := []*rulebook.CharacterFeature{}
	otherFeatures := []*rulebook.CharacterFeature{}

	for _, feature := range char.Features {
		if feature == nil {
			continue
		}
		switch feature.Type {
		case rulebook.FeatureTypeClass:
			classFeatures = append(classFeatures, feature)
		case rulebook.FeatureTypeRacial:
			racialFeatures = append(racialFeatures, feature)
		default:
			otherFeatures = append(otherFeatures, feature)
		}
	}

	// Add class features
	if len(classFeatures) > 0 {
		lines = append(lines, "**Class Features:**")
		caser := cases.Title(language.English)
		for _, feat := range classFeatures {
			featName := feat.Name
			// Add specific selections from metadata
			if feat.Key == "favored_enemy" && feat.Metadata != nil {
				if enemyType, ok := feat.Metadata["enemy_type"].(string); ok {
					// Capitalize the enemy type for display
					enemyDisplay := caser.String(enemyType)
					if enemyType == "humanoids" {
						enemyDisplay = "Two Humanoid Races"
					}
					featName = fmt.Sprintf("%s (%s)", feat.Name, enemyDisplay)
				}
			} else if feat.Key == "natural_explorer" && feat.Metadata != nil {
				if terrainType, ok := feat.Metadata["terrain_type"].(string); ok {
					// Capitalize the terrain type
					terrainDisplay := caser.String(terrainType)
					featName = fmt.Sprintf("%s (%s)", feat.Name, terrainDisplay)
				}
			} else if feat.Key == "fighting_style" && feat.Metadata != nil {
				if style, ok := feat.Metadata["style"].(string); ok {
					// Format fighting style for display
					styleDisplay := ""
					switch style {
					case "archery":
						styleDisplay = "Archery"
					case "defense":
						styleDisplay = "Defense"
					case "dueling":
						styleDisplay = "Dueling"
					case "great_weapon":
						styleDisplay = "Great Weapon Fighting"
					case "protection":
						styleDisplay = "Protection"
					case "two_weapon":
						styleDisplay = "Two-Weapon Fighting"
					default:
						styleDisplay = caser.String(style)
					}
					featName = fmt.Sprintf("%s (%s)", feat.Name, styleDisplay)
				}
			}
			lines = append(lines, fmt.Sprintf("• %s", featName))
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

// buildActiveEffectsDisplay builds the active effects section
func buildActiveEffectsDisplay(char *character.Character) []string {
	lines := []string{}

	// Get active status effects
	activeEffects := char.GetActiveStatusEffects()

	if len(activeEffects) == 0 {
		lines = append(lines, "*No active effects*")
		return lines
	}

	// Group effects by source
	abilityEffects := []string{}
	spellEffects := []string{}
	featureEffects := []string{}
	itemEffects := []string{}
	conditionEffects := []string{}

	for _, effect := range activeEffects {
		if effect == nil {
			continue
		}

		// Format duration
		durationStr := ""
		switch effect.Duration.Type {
		case effects.DurationPermanent:
			durationStr = "permanent"
		case effects.DurationRounds:
			durationStr = fmt.Sprintf("%d rounds", effect.Duration.Rounds)
		case effects.DurationInstant:
			durationStr = "instant"
		case effects.DurationWhileEquipped:
			durationStr = "while equipped"
		case effects.DurationUntilRest:
			durationStr = "until rest"
		}

		if effect.Duration.Concentration {
			if durationStr != "" {
				durationStr += ", concentration"
			} else {
				durationStr = "concentration"
			}
		}

		effectDisplay := fmt.Sprintf("• **%s**", effect.Name)
		if durationStr != "" && durationStr != "permanent" {
			effectDisplay += fmt.Sprintf(" (%s)", durationStr)
		}

		// Group by source type
		switch effect.Source {
		case effects.SourceAbility:
			abilityEffects = append(abilityEffects, effectDisplay)
		case effects.SourceSpell:
			spellEffects = append(spellEffects, effectDisplay)
		case effects.SourceFeature:
			featureEffects = append(featureEffects, effectDisplay)
		case effects.SourceItem:
			itemEffects = append(itemEffects, effectDisplay)
		case effects.SourceCondition:
			conditionEffects = append(conditionEffects, effectDisplay)
		default:
			// Default to features for unknown sources
			featureEffects = append(featureEffects, effectDisplay)
		}
	}

	// Add grouped effects
	if len(abilityEffects) > 0 {
		lines = append(lines, "**Abilities:**")
		lines = append(lines, abilityEffects...)
	}

	if len(spellEffects) > 0 {
		if len(lines) > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, "**Spells:**")
		lines = append(lines, spellEffects...)
	}

	if len(featureEffects) > 0 {
		if len(lines) > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, "**Features:**")
		lines = append(lines, featureEffects...)
	}

	if len(itemEffects) > 0 {
		if len(lines) > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, "**Items:**")
		lines = append(lines, itemEffects...)
	}

	if len(conditionEffects) > 0 {
		if len(lines) > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, "**Conditions:**")
		lines = append(lines, conditionEffects...)
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
