package admin

import (
	"context"
	"fmt"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/damage"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/equipment"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"
	"log"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/services"
	characterService "github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
	"github.com/bwmarrin/discordgo"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// InventoryHandler handles admin inventory commands
type InventoryHandler struct {
	characterService characterService.Service
}

// NewInventoryHandler creates a new inventory handler
func NewInventoryHandler(serviceProvider *services.Provider) *InventoryHandler {
	return &InventoryHandler{
		characterService: serviceProvider.CharacterService,
	}
}

// HandleGive handles the admin give command
func (h *InventoryHandler) HandleGive(s *discordgo.Session, i *discordgo.InteractionCreate, characterName, itemKey string, quantity int64) error {
	// Defer the response to give us more time
	if err := h.deferResponse(s, i); err != nil {
		return fmt.Errorf("failed to defer response: %w", err)
	}

	// Find the character by name for the user
	targetChar, err := h.findCharacterByName(i.Member.User.ID, characterName)
	if err != nil {
		return h.respondError(s, i, "Failed to find character", err)
	}
	if targetChar == nil {
		return h.respondError(s, i, fmt.Sprintf("Character '%s' not found", characterName), nil)
	}

	// For testing purposes, create a simple weapon based on the item key
	// In a real implementation, this would fetch from the D&D API
	var equipment equipment.Equipment

	switch itemKey {
	case "longsword", "shortsword", "dagger", "rapier", "greatsword":
		weapon := &equipment.Weapon{
			Base: equipment.BasicEquipment{
				Key:  itemKey,
				Name: cases.Title(language.English).String(itemKey),
			},
			WeaponCategory: "Simple",
			WeaponRange:    "Melee",
		}

		// Set damage based on weapon type
		switch itemKey {
		case "dagger":
			weapon.Damage = &damage.Damage{DiceCount: 1, DiceSize: 4, DamageType: "piercing"}
		case "shortsword":
			weapon.Damage = &damage.Damage{DiceCount: 1, DiceSize: 6, DamageType: "piercing"}
		case "rapier":
			weapon.Damage = &damage.Damage{DiceCount: 1, DiceSize: 8, DamageType: "piercing"}
			weapon.WeaponCategory = "Martial"
		case "longsword":
			weapon.Damage = &damage.Damage{DiceCount: 1, DiceSize: 8, DamageType: "slashing"}
			weapon.WeaponCategory = "Martial"
		case "greatsword":
			weapon.Damage = &damage.Damage{DiceCount: 2, DiceSize: 6, DamageType: "slashing"}
			weapon.TwoHandedDamage = weapon.Damage // Greatsword is always two-handed
			weapon.WeaponCategory = "Martial"
			weapon.Properties = []*shared.ReferenceItem{
				{Key: "two-handed", Name: "Two-Handed"},
				{Key: "heavy", Name: "Heavy"},
			}
		}

		equipment = weapon
	case "shield":
		equipment = &equipment.BasicEquipment{
			Key:  itemKey,
			Name: "Shield",
		}
	default:
		// Generic item
		equipment = &equipment.BasicEquipment{
			Key:  itemKey,
			Name: cases.Title(language.English).String(strings.ReplaceAll(itemKey, "-", " ")),
		}
	}

	// Initialize inventory map if needed
	if targetChar.Inventory == nil {
		targetChar.Inventory = make(map[equipment.EquipmentType][]equipment.Equipment)
	}

	// Determine equipment type
	equipType := equipment.GetEquipmentType()
	if equipType == "BasicEquipment" {
		// Try to determine type from key
		if itemKey == "shield" {
			equipType = equipment.EquipmentTypeArmor
		} else {
			equipType = equipment.EquipmentTypeOther
		}
	}

	// Add the item to the character's inventory
	for j := int64(0); j < quantity; j++ {
		targetChar.Inventory[equipType] = append(targetChar.Inventory[equipType], equipment)
	}

	// Save the character - use UpdateEquipment method
	if updateErr := h.characterService.UpdateEquipment(targetChar); updateErr != nil {
		return h.respondError(s, i, "Failed to save character", updateErr)
	}

	// Count total items
	totalItems := 0
	for _, items := range targetChar.Inventory {
		totalItems += len(items)
	}

	// Send success response
	embed := &discordgo.MessageEmbed{
		Title:       "✅ Item Added",
		Description: fmt.Sprintf("Added %d x **%s** to %s's inventory", quantity, equipment.GetName(), targetChar.Name),
		Color:       0x00ff00,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Current Inventory Count",
				Value:  fmt.Sprintf("%d items", totalItems),
				Inline: true,
			},
		},
	}

	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	})
	return err
}

// HandleTake handles the admin take command
func (h *InventoryHandler) HandleTake(s *discordgo.Session, i *discordgo.InteractionCreate, characterName, itemKey string, quantity int64) error {
	// Defer the response to give us more time
	if err := h.deferResponse(s, i); err != nil {
		return fmt.Errorf("failed to defer response: %w", err)
	}

	// Find the character by name for the user
	targetChar, err := h.findCharacterByName(i.Member.User.ID, characterName)
	if err != nil {
		return h.respondError(s, i, "Failed to find character", err)
	}
	if targetChar == nil {
		return h.respondError(s, i, fmt.Sprintf("Character '%s' not found", characterName), nil)
	}

	// Count how many of this item the character has across all equipment types
	itemCount := 0
	itemName := ""
	for _, items := range targetChar.Inventory {
		for _, item := range items {
			if item.GetKey() == itemKey {
				itemCount++
				if itemName == "" {
					itemName = item.GetName()
				}
			}
		}
	}

	if itemCount == 0 {
		return h.respondError(s, i, fmt.Sprintf("Character doesn't have any '%s'", itemKey), nil)
	}

	// Determine how many to remove
	removeCount := itemCount // Remove all by default
	if quantity > 0 && int(quantity) < itemCount {
		removeCount = int(quantity)
	}

	// Remove the items
	removed := 0
	for equipType, items := range targetChar.Inventory {
		newItems := []equipment.Equipment{}
		for _, item := range items {
			if item.GetKey() == itemKey && removed < removeCount {
				removed++
				continue
			}
			newItems = append(newItems, item)
		}
		if len(newItems) > 0 {
			targetChar.Inventory[equipType] = newItems
		} else {
			// Remove empty equipment type entries
			delete(targetChar.Inventory, equipType)
		}
	}

	// Also check if the item is equipped and unequip if necessary
	if targetChar.EquippedSlots != nil {
		for slot, equipped := range targetChar.EquippedSlots {
			if equipped != nil && equipped.GetKey() == itemKey {
				delete(targetChar.EquippedSlots, slot)
			}
		}
	}

	// Save the character
	if updateErr := h.characterService.UpdateEquipment(targetChar); updateErr != nil {
		return h.respondError(s, i, "Failed to save character", updateErr)
	}

	// Count total items remaining
	totalItems := 0
	for _, items := range targetChar.Inventory {
		totalItems += len(items)
	}

	// Send success response
	embed := &discordgo.MessageEmbed{
		Title:       "✅ Item Removed",
		Description: fmt.Sprintf("Removed %d x **%s** from %s's inventory", removed, itemName, targetChar.Name),
		Color:       0x00ff00,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Remaining in Inventory",
				Value:  fmt.Sprintf("%d x %s", itemCount-removed, itemName),
				Inline: true,
			},
			{
				Name:   "Total Inventory Count",
				Value:  fmt.Sprintf("%d items", totalItems),
				Inline: true,
			},
		},
	}

	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	})
	return err
}

func (h *InventoryHandler) respondError(s *discordgo.Session, i *discordgo.InteractionCreate, message string, err error) error {
	if err != nil {
		log.Printf("Admin inventory error: %s: %v", message, err)
	}

	embed := &discordgo.MessageEmbed{
		Title:       "❌ Error",
		Description: message,
		Color:       0xff0000,
	}

	_, editErr := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	})
	if editErr != nil {
		log.Printf("Failed to edit interaction response: %v", editErr)
	}
	return nil
}

// deferResponse defers the interaction response
func (h *InventoryHandler) deferResponse(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
}

// findCharacterByName finds a character by name for a given user
func (h *InventoryHandler) findCharacterByName(userID, characterName string) (*character.Character, error) {
	characters, err := h.characterService.ListCharacters(context.Background(), userID)
	if err != nil {
		return nil, err
	}

	for _, char := range characters {
		if char.Name == characterName {
			return char, nil
		}
	}

	return nil, nil
}
