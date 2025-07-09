package builders

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

// EmbedBuilder provides a fluent API for building Discord embeds
type EmbedBuilder struct {
	embed *discordgo.MessageEmbed
}

// NewEmbed creates a new embed builder
func NewEmbed() *EmbedBuilder {
	return &EmbedBuilder{
		embed: &discordgo.MessageEmbed{
			Type:   discordgo.EmbedTypeRich,
			Fields: make([]*discordgo.MessageEmbedField, 0),
		},
	}
}

// Title sets the embed title
func (b *EmbedBuilder) Title(title string) *EmbedBuilder {
	b.embed.Title = title
	return b
}

// Description sets the embed description
func (b *EmbedBuilder) Description(description string) *EmbedBuilder {
	b.embed.Description = description
	return b
}

// URL sets the embed URL
func (b *EmbedBuilder) URL(url string) *EmbedBuilder {
	b.embed.URL = url
	return b
}

// Color sets the embed color
func (b *EmbedBuilder) Color(color int) *EmbedBuilder {
	b.embed.Color = color
	return b
}

// Timestamp sets the embed timestamp
func (b *EmbedBuilder) Timestamp(timestamp time.Time) *EmbedBuilder {
	b.embed.Timestamp = timestamp.Format(time.RFC3339)
	return b
}

// Footer sets the embed footer
func (b *EmbedBuilder) Footer(text string, iconURL ...string) *EmbedBuilder {
	b.embed.Footer = &discordgo.MessageEmbedFooter{
		Text: text,
	}
	if len(iconURL) > 0 {
		b.embed.Footer.IconURL = iconURL[0]
	}
	return b
}

// Image sets the embed image
func (b *EmbedBuilder) Image(url string) *EmbedBuilder {
	b.embed.Image = &discordgo.MessageEmbedImage{
		URL: url,
	}
	return b
}

// Thumbnail sets the embed thumbnail
func (b *EmbedBuilder) Thumbnail(url string) *EmbedBuilder {
	b.embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
		URL: url,
	}
	return b
}

// Author sets the embed author
func (b *EmbedBuilder) Author(name, url, iconURL string) *EmbedBuilder {
	b.embed.Author = &discordgo.MessageEmbedAuthor{
		Name:    name,
		URL:     url,
		IconURL: iconURL,
	}
	return b
}

// Field adds a field to the embed
func (b *EmbedBuilder) Field(name, value string, inline bool) *EmbedBuilder {
	b.embed.Fields = append(b.embed.Fields, &discordgo.MessageEmbedField{
		Name:   name,
		Value:  value,
		Inline: inline,
	})
	return b
}

// AddField is an alias for Field
func (b *EmbedBuilder) AddField(name, value string, inline bool) *EmbedBuilder {
	return b.Field(name, value, inline)
}

// AddBlankField adds a blank field (useful for spacing)
func (b *EmbedBuilder) AddBlankField(inline bool) *EmbedBuilder {
	return b.Field("\u200b", "\u200b", inline)
}

// Build returns the constructed embed
func (b *EmbedBuilder) Build() *discordgo.MessageEmbed {
	return b.embed
}

// Common embed colors
const (
	ColorSuccess = 0x00ff00 // Green
	ColorError   = 0xff0000 // Red
	ColorWarning = 0xffaa00 // Orange
	ColorInfo    = 0x0099ff // Blue
	ColorPrimary = 0x7289da // Discord Blurple
)

// SuccessEmbed creates a pre-styled success embed
func SuccessEmbed(title, description string) *EmbedBuilder {
	return NewEmbed().
		Title("✅ " + title).
		Description(description).
		Color(ColorSuccess).
		Timestamp(time.Now())
}

// ErrorEmbed creates a pre-styled error embed
func ErrorEmbed(title, description string) *EmbedBuilder {
	return NewEmbed().
		Title("❌ " + title).
		Description(description).
		Color(ColorError).
		Timestamp(time.Now())
}

// WarningEmbed creates a pre-styled warning embed
func WarningEmbed(title, description string) *EmbedBuilder {
	return NewEmbed().
		Title("⚠️ " + title).
		Description(description).
		Color(ColorWarning).
		Timestamp(time.Now())
}

// InfoEmbed creates a pre-styled info embed
func InfoEmbed(title, description string) *EmbedBuilder {
	return NewEmbed().
		Title("ℹ️ " + title).
		Description(description).
		Color(ColorInfo).
		Timestamp(time.Now())
}

// CharacterEmbed creates a character display embed
type CharacterEmbedBuilder struct {
	*EmbedBuilder
}

// NewCharacterEmbed creates a new character embed builder
func NewCharacterEmbed() *CharacterEmbedBuilder {
	return &CharacterEmbedBuilder{
		EmbedBuilder: NewEmbed().Color(ColorPrimary),
	}
}

// SetCharacter sets basic character info
func (b *CharacterEmbedBuilder) SetCharacter(name, race, class string, level int) *CharacterEmbedBuilder {
	b.Title(name)
	b.Description(fmt.Sprintf("Level %d %s %s", level, race, class))
	return b
}

// AddStats adds ability scores
func (b *CharacterEmbedBuilder) AddStats(str, dex, con, intValue, wis, cha int) *CharacterEmbedBuilder {
	stats := fmt.Sprintf("STR: %d | DEX: %d | CON: %d\nINT: %d | WIS: %d | CHA: %d",
		str, dex, con, intValue, wis, cha)
	b.Field("Ability Scores", stats, false)
	return b
}

// AddCombatInfo adds HP, AC, etc
func (b *CharacterEmbedBuilder) AddCombatInfo(hp, maxHP, ac int) *CharacterEmbedBuilder {
	b.Field("HP", fmt.Sprintf("%d/%d", hp, maxHP), true)
	b.Field("AC", fmt.Sprintf("%d", ac), true)
	b.AddBlankField(true) // For alignment
	return b
}

// ListEmbed creates a paginated list embed
type ListEmbedBuilder struct {
	*EmbedBuilder
	itemsPerPage int
	currentPage  int
	totalItems   int
}

// NewListEmbed creates a new list embed builder
func NewListEmbed(title string, itemsPerPage int) *ListEmbedBuilder {
	return &ListEmbedBuilder{
		EmbedBuilder: NewEmbed().Title(title).Color(ColorPrimary),
		itemsPerPage: itemsPerPage,
		currentPage:  1,
	}
}

// SetPage sets the current page
func (b *ListEmbedBuilder) SetPage(page int) *ListEmbedBuilder {
	b.currentPage = page
	return b
}

// SetTotalItems sets the total item count
func (b *ListEmbedBuilder) SetTotalItems(total int) *ListEmbedBuilder {
	b.totalItems = total
	b.updateFooter()
	return b
}

// AddItem adds an item to the list
func (b *ListEmbedBuilder) AddItem(name, value string) *ListEmbedBuilder {
	b.Field(name, value, false)
	return b
}

// updateFooter updates the pagination footer
func (b *ListEmbedBuilder) updateFooter() {
	totalPages := (b.totalItems + b.itemsPerPage - 1) / b.itemsPerPage
	b.Footer(fmt.Sprintf("Page %d of %d • Total: %d", b.currentPage, totalPages, b.totalItems))
}
