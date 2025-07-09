package builders

import (
	"fmt"

	"github.com/KirkDiggler/dnd-bot-discord/internal/discord/v2/core"
	"github.com/bwmarrin/discordgo"
)

// ComponentBuilder builds Discord message components
type ComponentBuilder struct {
	rows            []discordgo.MessageComponent
	currentRow      []discordgo.MessageComponent
	customIDBuilder *core.CustomIDBuilder
}

// NewComponentBuilder creates a new component builder
func NewComponentBuilder(customIDBuilder *core.CustomIDBuilder) *ComponentBuilder {
	return &ComponentBuilder{
		rows:            make([]discordgo.MessageComponent, 0),
		currentRow:      make([]discordgo.MessageComponent, 0, 5), // Max 5 per row
		customIDBuilder: customIDBuilder,
	}
}

// Button adds a button to the current row
func (b *ComponentBuilder) Button(label string, style discordgo.ButtonStyle, action string, args ...string) *ComponentBuilder {
	customID := ""
	if b.customIDBuilder != nil {
		customID = b.customIDBuilder.Button(action, args[0], args[1:]...)
	} else {
		// Fallback for no builder
		id := core.NewCustomID("default", action)
		if len(args) > 0 {
			id.WithTarget(args[0]).WithArgs(args[1:]...)
		}
		customID = id.MustEncode()
	}

	button := discordgo.Button{
		Label:    label,
		Style:    style,
		CustomID: customID,
	}

	b.addComponent(button)
	return b
}

// LinkButton adds a URL button
func (b *ComponentBuilder) LinkButton(label, url string) *ComponentBuilder {
	button := discordgo.Button{
		Label: label,
		Style: discordgo.LinkButton,
		URL:   url,
	}

	b.addComponent(button)
	return b
}

// EmojiButton adds a button with emoji
func (b *ComponentBuilder) EmojiButton(label, emoji string, style discordgo.ButtonStyle, action string, args ...string) *ComponentBuilder {
	customID := ""
	if b.customIDBuilder != nil {
		customID = b.customIDBuilder.Button(action, args[0], args[1:]...)
	} else {
		id := core.NewCustomID("default", action)
		if len(args) > 0 {
			id.WithTarget(args[0]).WithArgs(args[1:]...)
		}
		customID = id.MustEncode()
	}

	button := discordgo.Button{
		Label:    label,
		Style:    style,
		CustomID: customID,
		Emoji: &discordgo.ComponentEmoji{
			Name: emoji,
		},
	}

	b.addComponent(button)
	return b
}

// DisabledButton adds a disabled button
func (b *ComponentBuilder) DisabledButton(label string, style discordgo.ButtonStyle) *ComponentBuilder {
	button := discordgo.Button{
		Label:    label,
		Style:    style,
		CustomID: "disabled",
		Disabled: true,
	}

	b.addComponent(button)
	return b
}

// SelectMenu adds a select menu
func (b *ComponentBuilder) SelectMenu(placeholder, action string, options []SelectOption, config ...SelectConfig) *ComponentBuilder {
	customID := ""
	if b.customIDBuilder != nil {
		customID = b.customIDBuilder.Select(action, nil)
	} else {
		customID = core.NewCustomID("default", action).MustEncode()
	}

	return b.selectMenuWithCustomID(placeholder, customID, options, config...)
}

// SelectMenuWithTarget adds a select menu with a target ID
func (b *ComponentBuilder) SelectMenuWithTarget(placeholder, action, target string, options []SelectOption, config ...SelectConfig) *ComponentBuilder {
	customID := ""
	if b.customIDBuilder != nil {
		customID = b.customIDBuilder.Build(action).WithTarget(target).MustEncode()
	} else {
		customID = core.NewCustomID("default", action).WithTarget(target).MustEncode()
	}

	return b.selectMenuWithCustomID(placeholder, customID, options, config...)
}

// selectMenuWithCustomID is the internal method that builds the select menu
func (b *ComponentBuilder) selectMenuWithCustomID(placeholder, customID string, options []SelectOption, config ...SelectConfig) *ComponentBuilder {
	// Convert options
	discordOptions := make([]discordgo.SelectMenuOption, len(options))
	for i, opt := range options {
		discordOptions[i] = discordgo.SelectMenuOption{
			Label:       opt.Label,
			Value:       opt.Value,
			Description: opt.Description,
			Default:     opt.Default,
		}
		if opt.Emoji != "" {
			discordOptions[i].Emoji = &discordgo.ComponentEmoji{
				Name: opt.Emoji,
			}
		}
	}

	selectMenu := discordgo.SelectMenu{
		CustomID:    customID,
		Placeholder: placeholder,
		Options:     discordOptions,
	}

	// Apply config if provided
	if len(config) > 0 {
		cfg := config[0]
		if cfg.MinValues > 0 {
			minVal := cfg.MinValues
			selectMenu.MinValues = &minVal
		}
		if cfg.MaxValues > 0 {
			selectMenu.MaxValues = cfg.MaxValues
		}
		selectMenu.Disabled = cfg.Disabled
	}

	b.addComponent(selectMenu)
	return b
}

// SelectMenuWithOptions is a convenience method for multi-select menus
func (b *ComponentBuilder) SelectMenuWithOptions(placeholder, action string, options []SelectOption, minValues, maxValues int) *ComponentBuilder {
	return b.SelectMenu(placeholder, action, options, SelectConfig{
		MinValues: minValues,
		MaxValues: maxValues,
	})
}

// NewRow starts a new action row
func (b *ComponentBuilder) NewRow() *ComponentBuilder {
	if len(b.currentRow) > 0 {
		b.rows = append(b.rows, discordgo.ActionsRow{
			Components: b.currentRow,
		})
		b.currentRow = make([]discordgo.MessageComponent, 0, 5)
	}
	return b
}

// Build returns the built components
func (b *ComponentBuilder) Build() []discordgo.MessageComponent {
	// Add any remaining components in current row
	if len(b.currentRow) > 0 {
		b.rows = append(b.rows, discordgo.ActionsRow{
			Components: b.currentRow,
		})
	}

	return b.rows
}

// addComponent adds a component to the current row
func (b *ComponentBuilder) addComponent(component discordgo.MessageComponent) {
	// Check if we need a new row
	if len(b.currentRow) >= 5 {
		b.NewRow()
	}

	b.currentRow = append(b.currentRow, component)
}

// SelectOption represents an option in a select menu
type SelectOption struct {
	Label       string
	Value       string
	Description string
	Emoji       string
	Default     bool
}

// SelectConfig configures a select menu
type SelectConfig struct {
	MinValues int
	MaxValues int
	Disabled  bool
}

// Common button styles helpers
func (b *ComponentBuilder) PrimaryButton(label, action string, args ...string) *ComponentBuilder {
	return b.Button(label, discordgo.PrimaryButton, action, args...)
}

func (b *ComponentBuilder) SecondaryButton(label, action string, args ...string) *ComponentBuilder {
	return b.Button(label, discordgo.SecondaryButton, action, args...)
}

func (b *ComponentBuilder) SuccessButton(label, action string, args ...string) *ComponentBuilder {
	return b.Button(label, discordgo.SuccessButton, action, args...)
}

func (b *ComponentBuilder) DangerButton(label, action string, args ...string) *ComponentBuilder {
	return b.Button(label, discordgo.DangerButton, action, args...)
}

// Pagination helpers
func (b *ComponentBuilder) PaginationButtons(currentPage, totalPages int, baseAction string) *ComponentBuilder {
	// Previous button
	if currentPage > 1 {
		b.EmojiButton("Previous", "⬅️", discordgo.SecondaryButton, baseAction+"_prev", fmt.Sprintf("%d", currentPage-1))
	} else {
		b.DisabledButton("Previous", discordgo.SecondaryButton)
	}

	// Page indicator
	b.DisabledButton(fmt.Sprintf("%d/%d", currentPage, totalPages), discordgo.SecondaryButton)

	// Next button
	if currentPage < totalPages {
		b.EmojiButton("Next", "➡️", discordgo.SecondaryButton, baseAction+"_next", fmt.Sprintf("%d", currentPage+1))
	} else {
		b.DisabledButton("Next", discordgo.SecondaryButton)
	}

	return b
}

// ConfirmationButtons adds Yes/No confirmation buttons
func (b *ComponentBuilder) ConfirmationButtons(confirmAction, cancelAction, targetID string) *ComponentBuilder {
	b.SuccessButton("Yes", confirmAction, targetID)
	b.DangerButton("No", cancelAction, targetID)
	return b
}

// ActionButtons creates a standard set of action buttons
type ActionButtonsConfig struct {
	ShowEdit   bool
	ShowDelete bool
	ShowInfo   bool
	TargetID   string
}

func (b *ComponentBuilder) ActionButtons(config ActionButtonsConfig) *ComponentBuilder {
	if config.ShowInfo {
		b.PrimaryButton("Info", "info", config.TargetID)
	}
	if config.ShowEdit {
		b.SecondaryButton("Edit", "edit", config.TargetID)
	}
	if config.ShowDelete {
		b.DangerButton("Delete", "delete", config.TargetID)
	}
	return b
}
