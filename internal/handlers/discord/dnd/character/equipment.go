package character

import (
	"fmt"
	"log"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services"
	"github.com/bwmarrin/discordgo"
)

type EquipmentHandler struct {
	ServiceProvider *services.Provider
}

type EquipmentHandlerConfig struct {
	ServiceProvider *services.Provider
}

func NewEquipmentHandler(cfg *EquipmentHandlerConfig) *EquipmentHandler {
	return &EquipmentHandler{
		ServiceProvider: cfg.ServiceProvider,
	}
}

// HandleEquip handles the /dnd character equip command for weapons and shields
func (h *EquipmentHandler) HandleEquip(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// Get character ID from command options
	options := i.ApplicationCommandData().Options[0].Options[0].Options // dnd -> character -> equip
	var characterID string
	var itemKey string

	for _, opt := range options {
		switch opt.Name {
		case "character_id":
			characterID = opt.StringValue()
		case "item":
			itemKey = opt.StringValue()
		}
	}

	if characterID == "" || itemKey == "" {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Both character ID and weapon/shield are required!",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// Get character
	char, err := h.ServiceProvider.CharacterService.GetByID(characterID)
	if err != nil {
		log.Printf("Failed to get character %s: %v", characterID, err)
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Character not found!",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// Check ownership
	if char.OwnerID != i.Member.User.ID {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ You can only equip items on your own characters!",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// Find item in inventory (weapon or shield)
	var foundItem entities.Equipment
	for _, equipmentList := range char.Inventory {
		for _, equipment := range equipmentList {
			if equipment.GetKey() == itemKey {
				foundItem = equipment
				break
			}
		}
		if foundItem != nil {
			break
		}
	}

	if foundItem == nil {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("❌ Item '%s' not found in your inventory!", itemKey),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// Equip the item
	success := char.Equip(itemKey)
	if !success {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Failed to equip item!",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// Save the equipment changes
	if err := h.ServiceProvider.CharacterService.UpdateEquipment(char); err != nil {
		log.Printf("Failed to save equipment changes for character %s: %v", characterID, err)
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "⚠️ Weapon equipped but failed to save changes. Changes may be lost on restart.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// Success response with equipped item info
	embed := &discordgo.MessageEmbed{
		Title:       "⚔️ Item Equipped!",
		Description: fmt.Sprintf("**%s** equipped **%s**", char.Name, foundItem.GetName()),
		Color:       0x2ecc71, // Green
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Item",
				Value:  foundItem.GetName(),
				Inline: true,
			},
			{
				Name:   "Slot",
				Value:  string(foundItem.GetSlot()),
				Inline: true,
			},
		},
	}

	// Add weapon stats if it's a weapon
	if weapon, ok := foundItem.(*entities.Weapon); ok {
		if weapon.Damage != nil {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "Damage",
				Value:  fmt.Sprintf("%dd%d+%d %s", weapon.Damage.DiceCount, weapon.Damage.DiceSize, weapon.Damage.Bonus, weapon.Damage.DamageType),
				Inline: true,
			})
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Type",
			Value:  fmt.Sprintf("%s %s", weapon.WeaponRange, weapon.WeaponCategory),
			Inline: true,
		})
	} else if armor, ok := foundItem.(*entities.Armor); ok && armor.ArmorCategory == entities.ArmorCategoryShield {
		// Add shield stats
		if armor.ArmorClass != nil {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "AC Bonus",
				Value:  fmt.Sprintf("+%d", armor.ArmorClass.Base),
				Inline: true,
			})
		}
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
}

// HandleUnequip handles the /dnd character unequip command
func (h *EquipmentHandler) HandleUnequip(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// Get character ID and slot from command options
	options := i.ApplicationCommandData().Options[0].Options[0].Options // dnd -> character -> unequip
	var characterID string
	var slotName string

	for _, opt := range options {
		switch opt.Name {
		case "character_id":
			characterID = opt.StringValue()
		case "slot":
			slotName = opt.StringValue()
		}
	}

	if characterID == "" || slotName == "" {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Both character ID and slot are required!",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// Get character
	char, err := h.ServiceProvider.CharacterService.GetByID(characterID)
	if err != nil {
		log.Printf("Failed to get character %s: %v", characterID, err)
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Character not found!",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// Check ownership
	if char.OwnerID != i.Member.User.ID {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ You can only unequip items from your own characters!",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// Convert slot name to Slot type
	slot := entities.Slot(slotName)

	// Check if anything is equipped in that slot
	if char.EquippedSlots == nil || char.EquippedSlots[slot] == nil {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("❌ No item equipped in %s slot!", slotName),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	equippedItem := char.EquippedSlots[slot]
	itemName := equippedItem.GetName()

	// Unequip the item
	char.EquippedSlots[slot] = nil

	// Save the equipment changes
	if err := h.ServiceProvider.CharacterService.UpdateEquipment(char); err != nil {
		log.Printf("Failed to save equipment changes for character %s: %v", characterID, err)
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "⚠️ Item unequipped but failed to save changes. Changes may be lost on restart.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// Success response
	embed := &discordgo.MessageEmbed{
		Title:       "🛡️ Item Unequipped!",
		Description: fmt.Sprintf("**%s** unequipped **%s** from %s slot", char.Name, itemName, slotName),
		Color:       0xe74c3c, // Red
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
}

// HandleInventory shows the character's weapon inventory with equip buttons
func (h *EquipmentHandler) HandleInventory(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// Get character ID from command options
	options := i.ApplicationCommandData().Options[0].Options[0].Options // dnd -> character -> inventory
	var characterID string

	for _, opt := range options {
		if opt.Name == "character_id" {
			characterID = opt.StringValue()
		}
	}

	if characterID == "" {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Character ID is required!",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// Get character
	char, err := h.ServiceProvider.CharacterService.GetByID(characterID)
	if err != nil {
		log.Printf("Failed to get character %s: %v", characterID, err)
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Character not found!",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// Check ownership
	if char.OwnerID != i.Member.User.ID {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ You can only view your own character's inventory!",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// Build inventory display
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("🎒 %s's Inventory", char.Name),
		Description: "Weapons and equipped items",
		Color:       0x3498db, // Blue
		Fields:      []*discordgo.MessageEmbedField{},
	}

	// Show currently equipped items
	var equippedItems []string
	if char.EquippedSlots != nil {
		for slot, equipment := range char.EquippedSlots {
			if equipment != nil {
				equippedItems = append(equippedItems, fmt.Sprintf("**%s**: %s", slot, equipment.GetName()))
			}
		}
	}

	if len(equippedItems) > 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "⚔️ Currently Equipped",
			Value:  strings.Join(equippedItems, "\n"),
			Inline: false,
		})
	} else {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "⚔️ Currently Equipped",
			Value:  "*No items equipped*",
			Inline: false,
		})
	}

	// Show available weapons
	var weaponList []string
	if weapons, exists := char.Inventory[entities.EquipmentTypeWeapon]; exists {
		for _, weapon := range weapons {
			weaponList = append(weaponList, fmt.Sprintf("• %s (%s)", weapon.GetName(), weapon.GetKey()))
		}
	}

	if len(weaponList) > 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "🗡️ Available Weapons",
			Value:  strings.Join(weaponList, "\n"),
			Inline: false,
		})
	} else {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "🗡️ Available Weapons",
			Value:  "*No weapons in inventory*",
			Inline: false,
		})
	}

	// Show available shields
	var shieldList []string
	if armor, exists := char.Inventory[entities.EquipmentTypeArmor]; exists {
		for _, item := range armor {
			if armorItem, ok := item.(*entities.Armor); ok && armorItem.ArmorCategory == entities.ArmorCategoryShield {
				shieldList = append(shieldList, fmt.Sprintf("• %s (%s)", item.GetName(), item.GetKey()))
			}
		}
	}

	if len(shieldList) > 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "🛡️ Available Shields",
			Value:  strings.Join(shieldList, "\n"),
			Inline: false,
		})
	} else {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "🛡️ Available Shields",
			Value:  "*No shields in inventory*",
			Inline: false,
		})
	}

	embed.Footer = &discordgo.MessageEmbedFooter{
		Text: "Use /dnd character equip to equip items, /dnd character unequip to remove them",
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
}
