package core

import (
	"github.com/bwmarrin/discordgo"
)

// Handler defines the interface for all interaction handlers
type Handler interface {
	// CanHandle determines if this handler should process the interaction
	CanHandle(ctx *InteractionContext) bool

	// Handle processes the interaction and returns a result
	Handle(ctx *InteractionContext) (*HandlerResult, error)
}

// HandlerFunc allows functions to implement the Handler interface
type HandlerFunc func(ctx *InteractionContext) (*HandlerResult, error)

// CanHandle for HandlerFunc always returns true
func (f HandlerFunc) CanHandle(ctx *InteractionContext) bool {
	return true
}

// Handle calls the function
func (f HandlerFunc) Handle(ctx *InteractionContext) (*HandlerResult, error) {
	return f(ctx)
}

// HandlerResult contains the response and metadata from a handler
type HandlerResult struct {
	// Response to send to Discord
	Response *Response

	// Whether the response was already deferred
	Deferred bool

	// Whether to stop processing further handlers
	StopPropagation bool

	// Additional context to pass to middleware
	Context map[string]interface{}
}

// Response represents a Discord-agnostic response
type Response struct {
	// Text content of the response
	Content string

	// Discord embeds
	Embeds []*discordgo.MessageEmbed

	// Interactive components (buttons, select menus, etc)
	Components []discordgo.MessageComponent

	// Whether this response should be ephemeral (only visible to the user)
	Ephemeral bool

	// File attachments
	Files []*discordgo.File

	// Whether to update the original message (for deferred responses)
	Update bool

	// Allowed mentions configuration
	AllowedMentions *discordgo.MessageAllowedMentions

	// TTS (text-to-speech) flag
	TTS bool
}

// NewResponse creates a new response with the given content
func NewResponse(content string) *Response {
	return &Response{
		Content: content,
	}
}

// NewEphemeralResponse creates a new ephemeral response
func NewEphemeralResponse(content string) *Response {
	return &Response{
		Content:   content,
		Ephemeral: true,
	}
}

// NewEmbedResponse creates a response with an embed
func NewEmbedResponse(embed *discordgo.MessageEmbed) *Response {
	return &Response{
		Embeds: []*discordgo.MessageEmbed{embed},
	}
}

// WithComponents adds components to the response
func (r *Response) WithComponents(components ...discordgo.MessageComponent) *Response {
	r.Components = components
	return r
}

// WithEmbeds adds embeds to the response
func (r *Response) WithEmbeds(embeds ...*discordgo.MessageEmbed) *Response {
	r.Embeds = embeds
	return r
}

// WithFiles adds file attachments to the response
func (r *Response) WithFiles(files ...*discordgo.File) *Response {
	r.Files = files
	return r
}

// AsEphemeral sets the response to be ephemeral
func (r *Response) AsEphemeral() *Response {
	r.Ephemeral = true
	return r
}

// AsUpdate sets the response to update the original message
func (r *Response) AsUpdate() *Response {
	r.Update = true
	return r
}
