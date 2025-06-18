package help

import (
	"github.com/bwmarrin/discordgo"
)

type HelpRequest struct {
	Session     *discordgo.Session
	Interaction *discordgo.InteractionCreate
	Topic       string // Optional specific help topic
}

type HelpHandler struct{}

func NewHelpHandler() *HelpHandler {
	return &HelpHandler{}
}

func (h *HelpHandler) Handle(req *HelpRequest) error {
	// Create help embed based on topic
	var embed *discordgo.MessageEmbed

	if req.Topic == "" {
		embed = h.getGeneralHelp()
	} else {
		switch req.Topic {
		case "character":
			embed = h.getCharacterHelp()
		case "session":
			embed = h.getSessionHelp()
		case "encounter":
			embed = h.getEncounterHelp()
		case "combat":
			embed = h.getCombatHelp()
		default:
			embed = h.getGeneralHelp()
		}
	}

	// Respond with the help embed
	err := req.Session.InteractionRespond(req.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral, // Only visible to the user
		},
	})

	return err
}

func (h *HelpHandler) getGeneralHelp() *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       "🎲 D&D Bot Help",
		Description: "Welcome to the D&D 5e Discord Bot! Here's how to get started:",
		Color:       0x3498db, // Blue
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "📚 Getting Started",
				Value:  "1. Create a character: `/dnd character create`\n2. Create or join a session: `/dnd session create` or `/dnd session join <code>`\n3. Start playing!",
				Inline: false,
			},
			{
				Name:   "🎭 Character Commands",
				Value:  "`/dnd character create` - Create a new character\n`/dnd character list` - View your characters\n`/dnd character show <id>` - Show character details",
				Inline: false,
			},
			{
				Name:   "🎮 Session Commands",
				Value:  "`/dnd session create` - Start a new game session\n`/dnd session join <code>` - Join with invite code\n`/dnd session list` - View your sessions\n`/dnd session info` - Current session details",
				Inline: false,
			},
			{
				Name:   "⚔️ Encounter Commands",
				Value:  "`/dnd encounter add <monster>` - Add monster to encounter (DM only)",
				Inline: false,
			},
			{
				Name:   "❓ More Help",
				Value:  "Use `/dnd help <topic>` for detailed help on:\n• `character` - Character creation & management\n• `session` - Game session management\n• `encounter` - Running encounters\n• `combat` - Combat mechanics",
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Need more help? Contact your server admin!",
		},
	}
}

func (h *HelpHandler) getCharacterHelp() *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       "🎭 Character Management Help",
		Description: "Everything you need to know about creating and managing characters.",
		Color:       0x2ecc71, // Green
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Creating a Character",
				Value:  "Use `/dnd character create` to start the character creation wizard. You'll be guided through:\n• Choosing a race\n• Selecting a class\n• Rolling ability scores\n• Picking proficiencies\n• Selecting starting equipment",
				Inline: false,
			},
			{
				Name:   "Managing Characters",
				Value:  "`/dnd character list` - See all your characters\n`/dnd character show <id>` - View full character sheet\n`/dnd character delete <id>` - Delete a character (⚠️ permanent!)",
				Inline: false,
			},
			{
				Name:   "Character Status",
				Value:  "• **Active** - Available for play\n• **Retired** - No longer actively played\n• **Deceased** - Met an unfortunate end\n• **Draft** - Still being created",
				Inline: false,
			},
			{
				Name:   "💡 Tips",
				Value:  "• You can have multiple characters\n• Characters are saved automatically\n• Join a session to use your character in play",
				Inline: false,
			},
		},
	}
}

func (h *HelpHandler) getSessionHelp() *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       "🎮 Session Management Help",
		Description: "Learn how to create and manage D&D game sessions.",
		Color:       0x9b59b6, // Purple
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Starting a Session",
				Value:  "**As a DM:**\n1. `/dnd session create` - Create a new session\n2. Share the invite code with players\n3. `/dnd session start` - Begin when ready\n\n**As a Player:**\n1. `/dnd session join <code>` - Join with invite code\n2. Select your character when prompted",
				Inline: false,
			},
			{
				Name:   "Session Commands",
				Value:  "**DM Commands:**\n`/dnd session start` - Begin the session\n`/dnd session end` - Conclude the session\n`/dnd session info` - View session details\n\n**Player Commands:**\n`/dnd session list` - View your sessions\n`/dnd session info` - Current session info\n`Leave Session` button - Exit a session",
				Inline: false,
			},
			{
				Name:   "Session States",
				Value:  "• **Planning** - Setting up, players joining\n• **Active** - Game in progress\n• **Paused** - Temporarily stopped\n• **Ended** - Session complete",
				Inline: false,
			},
		},
	}
}

func (h *HelpHandler) getEncounterHelp() *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       "⚔️ Encounter Management Help",
		Description: "Running combat encounters in your D&D session.",
		Color:       0xe74c3c, // Red
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Starting an Encounter",
				Value:  "**DM Only:**\n1. `/dnd encounter add <monster>` - Add monsters\n2. Players are automatically added from session\n3. Click `Roll Initiative` when ready\n4. Combat begins in initiative order",
				Inline: false,
			},
			{
				Name:   "Adding Monsters",
				Value:  "• Search by name: `/dnd encounter add goblin`\n• Select from menu if multiple matches\n• Add multiple monsters of same type\n• Common monsters available: Goblin, Orc, Skeleton, Dire Wolf, Zombie",
				Inline: false,
			},
			{
				Name:   "During Combat",
				Value:  "• Track initiative order automatically\n• Monitor HP for all combatants\n• Apply damage or healing\n• Advance turns in order",
				Inline: false,
			},
		},
	}
}

func (h *HelpHandler) getCombatHelp() *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       "⚔️ Combat Mechanics Help",
		Description: "Understanding D&D 5e combat in the bot.",
		Color:       0xf39c12, // Orange
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Initiative & Turns",
				Value:  "• Initiative: 1d20 + DEX modifier\n• Higher initiative acts first\n• Each round, everyone gets one turn\n• DM controls monster turns",
				Inline: false,
			},
			{
				Name:   "Hit Points & Damage",
				Value:  "• **Current HP** - Health remaining\n• **Max HP** - Full health\n• **Temp HP** - Bonus HP (doesn't stack)\n• **0 HP** - Unconscious (death saves)",
				Inline: false,
			},
			{
				Name:   "Actions in Combat",
				Value:  "On your turn you can:\n• **Attack** - Roll to hit, then damage\n• **Cast a Spell** - If you know spells\n• **Move** - Up to your speed\n• **Other** - Dodge, help, hide, etc.",
				Inline: false,
			},
			{
				Name:   "Dice Notation",
				Value:  "• `1d20` - Roll 1 twenty-sided die\n• `2d6+3` - Roll 2 six-sided dice, add 3\n• `4d4` - Roll 4 four-sided dice\n• Attack rolls: d20 + modifiers\n• Damage rolls: weapon dice + STR/DEX",
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Based on D&D 5th Edition rules",
		},
	}
}
