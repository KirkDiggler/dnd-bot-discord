package utils

import "github.com/bwmarrin/discordgo"

// GetCommandOption safely retrieves a command option by name from interaction data
func GetCommandOption(i *discordgo.InteractionCreate, name string) *discordgo.ApplicationCommandInteractionDataOption {
	if i.ApplicationCommandData().Options == nil {
		return nil
	}

	// Start with the root options
	options := i.ApplicationCommandData().Options

	// Navigate through subcommand groups and subcommands
	for len(options) > 0 {
		// If we find the option we're looking for, return it
		for _, opt := range options {
			if opt.Name == name {
				return opt
			}
		}

		// If the first option has sub-options (it's a subcommand group or subcommand), drill down
		if len(options[0].Options) > 0 {
			options = options[0].Options
		} else {
			// No more sub-options to explore
			break
		}
	}

	return nil
}

// GetStringOption safely retrieves a string option value by name
func GetStringOption(i *discordgo.InteractionCreate, name string) string {
	opt := GetCommandOption(i, name)
	if opt == nil {
		return ""
	}
	return opt.StringValue()
}

// GetIntOption safely retrieves an integer option value by name
func GetIntOption(i *discordgo.InteractionCreate, name string) int64 {
	opt := GetCommandOption(i, name)
	if opt == nil {
		return 0
	}
	return opt.IntValue()
}

// GetBoolOption safely retrieves a boolean option value by name
func GetBoolOption(i *discordgo.InteractionCreate, name string) bool {
	opt := GetCommandOption(i, name)
	if opt == nil {
		return false
	}
	return opt.BoolValue()
}
