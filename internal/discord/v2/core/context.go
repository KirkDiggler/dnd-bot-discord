package core

import (
	"context"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// InteractionContext wraps a Discord interaction with useful helpers and context
type InteractionContext struct {
	// Core Discord objects
	Session     *discordgo.Session
	Interaction *discordgo.InteractionCreate

	// Extracted common fields for convenience
	UserID    string
	GuildID   string
	ChannelID string
	Member    *discordgo.Member

	// Context for cancellation and values
	Context context.Context

	// Parsed interaction data
	params map[string]interface{}
}

// NewInteractionContext creates a new InteractionContext from a Discord interaction
func NewInteractionContext(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) *InteractionContext {
	ic := &InteractionContext{
		Session:     s,
		Interaction: i,
		Context:     ctx,
		params:      make(map[string]interface{}),
	}

	// Extract common fields
	if i.Member != nil {
		ic.Member = i.Member
		ic.UserID = i.Member.User.ID
	} else if i.User != nil {
		ic.UserID = i.User.ID
	}

	if i.GuildID != "" {
		ic.GuildID = i.GuildID
	}

	if i.ChannelID != "" {
		ic.ChannelID = i.ChannelID
	}

	// Parse parameters based on interaction type
	ic.parseParams()

	return ic
}

// parseParams extracts parameters from different interaction types
func (ic *InteractionContext) parseParams() {
	switch ic.Interaction.Type {
	case discordgo.InteractionApplicationCommand:
		ic.parseCommandParams()
	case discordgo.InteractionMessageComponent:
		ic.parseComponentParams()
	case discordgo.InteractionModalSubmit:
		ic.parseModalParams()
	}
}

// parseCommandParams extracts options from slash commands
func (ic *InteractionContext) parseCommandParams() {
	if ic.Interaction.ApplicationCommandData().Options == nil {
		return
	}

	// Recursively parse options
	ic.parseOptions(ic.Interaction.ApplicationCommandData().Options)
}

// parseOptions recursively extracts command options
func (ic *InteractionContext) parseOptions(options []*discordgo.ApplicationCommandInteractionDataOption) {
	for _, opt := range options {
		// If this option has sub-options, it's a subcommand
		if len(opt.Options) > 0 {
			ic.params["subcommand"] = opt.Name
			ic.parseOptions(opt.Options)
		} else {
			// Store the option value
			ic.params[opt.Name] = opt.Value
		}
	}
}

// parseComponentParams extracts custom ID parts
func (ic *InteractionContext) parseComponentParams() {
	customID := ic.Interaction.MessageComponentData().CustomID
	parts := strings.Split(customID, ":")

	if len(parts) >= 1 {
		ic.params["component_action"] = parts[0]
	}
	if len(parts) >= 2 {
		ic.params["component_target"] = parts[1]
	}

	// Store remaining parts as indexed params
	for i := 2; i < len(parts); i++ {
		ic.params[fmt.Sprintf("component_arg_%d", i-2)] = parts[i]
	}

	// Store the full custom ID for reference
	ic.params["custom_id"] = customID
}

// parseModalParams extracts modal submit data
func (ic *InteractionContext) parseModalParams() {
	data := ic.Interaction.ModalSubmitData()
	ic.params["modal_id"] = data.CustomID

	// Extract text input components
	for _, comp := range data.Components {
		if textInput, ok := comp.(*discordgo.ActionsRow); ok {
			for _, innerComp := range textInput.Components {
				if input, ok := innerComp.(*discordgo.TextInput); ok {
					ic.params[input.CustomID] = input.Value
				}
			}
		}
	}
}

// GetParam retrieves a parameter by name
func (ic *InteractionContext) GetParam(name string) interface{} {
	return ic.params[name]
}

// GetStringParam retrieves a string parameter or returns empty string
func (ic *InteractionContext) GetStringParam(name string) string {
	if val, ok := ic.params[name]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return ""
}

// GetIntParam retrieves an int parameter or returns 0
func (ic *InteractionContext) GetIntParam(name string) int {
	if val, ok := ic.params[name]; ok {
		switch v := val.(type) {
		case float64:
			return int(v)
		case int:
			return v
		case int64:
			return int(v)
		}
	}
	return 0
}

// GetBoolParam retrieves a bool parameter or returns false
func (ic *InteractionContext) GetBoolParam(name string) bool {
	if val, ok := ic.params[name]; ok {
		if boolVal, ok := val.(bool); ok {
			return boolVal
		}
	}
	return false
}

// IsCommand checks if this is a slash command interaction
func (ic *InteractionContext) IsCommand() bool {
	return ic.Interaction.Type == discordgo.InteractionApplicationCommand
}

// IsComponent checks if this is a message component interaction
func (ic *InteractionContext) IsComponent() bool {
	return ic.Interaction.Type == discordgo.InteractionMessageComponent
}

// IsModal checks if this is a modal submit interaction
func (ic *InteractionContext) IsModal() bool {
	return ic.Interaction.Type == discordgo.InteractionModalSubmit
}

// GetCustomID returns the custom ID for component interactions
func (ic *InteractionContext) GetCustomID() string {
	if ic.IsComponent() {
		return ic.Interaction.MessageComponentData().CustomID
	}
	if ic.IsModal() {
		return ic.Interaction.ModalSubmitData().CustomID
	}
	return ""
}

// GetCommandName returns the command name for slash commands
func (ic *InteractionContext) GetCommandName() string {
	if ic.IsCommand() {
		return ic.Interaction.ApplicationCommandData().Name
	}
	return ""
}

// GetSubcommand returns the subcommand name if present
func (ic *InteractionContext) GetSubcommand() string {
	return ic.GetStringParam("subcommand")
}

// WithValue adds a value to the context
func (ic *InteractionContext) WithValue(key, val interface{}) {
	ic.Context = context.WithValue(ic.Context, key, val)
}

// Value retrieves a value from the context
func (ic *InteractionContext) Value(key interface{}) interface{} {
	return ic.Context.Value(key)
}
