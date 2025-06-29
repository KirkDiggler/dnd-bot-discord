package character

import (
	"fmt"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/bwmarrin/discordgo"
)

// BuildCreationSuccessResponse builds a simple success message with a button to view the character sheet
func BuildCreationSuccessResponse(char *entities.Character) (*discordgo.MessageEmbed, []discordgo.MessageComponent) {
	// Build a simple success embed
	embed := &discordgo.MessageEmbed{
		Title:       "✅ Character Created Successfully!",
		Description: fmt.Sprintf("**%s** has been created and is ready for adventure!", char.Name),
		Color:       0x00ff00, // Green
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Race & Class",
				Value:  fmt.Sprintf("%s %s (Level %d)", char.Race.Name, char.Class.Name, char.Level),
				Inline: true,
			},
			{
				Name:   "Hit Points",
				Value:  fmt.Sprintf("%d/%d", char.CurrentHitPoints, char.MaxHitPoints),
				Inline: true,
			},
			{
				Name:   "Armor Class",
				Value:  fmt.Sprintf("%d", char.AC),
				Inline: true,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Click the button below to view your complete character sheet",
		},
	}

	// Create a single button to view the character sheet
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "View Character Sheet",
					Style:    discordgo.SuccessButton,
					CustomID: fmt.Sprintf("character:sheet_show:%s", char.ID),
					Emoji:    &discordgo.ComponentEmoji{Name: "📋"},
				},
			},
		},
	}

	return embed, components
}
