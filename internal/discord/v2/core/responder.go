package core

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// InteractionResponder provides an abstraction over Discord's interaction response API
type InteractionResponder interface {
	// Defer sends a deferred response, optionally ephemeral
	Defer(ephemeral bool) error

	// Respond sends an immediate response
	Respond(response *Response) error

	// Edit updates a previous response (after defer or respond)
	Edit(response *Response) error

	// FollowUp sends an additional message after the initial response
	FollowUp(response *Response) (*discordgo.Message, error)

	// DeleteFollowUp deletes a follow-up message
	DeleteFollowUp(messageID string) error

	// DeleteOriginal deletes the original response
	DeleteOriginal() error
}

// DiscordResponder implements InteractionResponder using Discord's API
type DiscordResponder struct {
	session     *discordgo.Session
	interaction *discordgo.InteractionCreate
	responded   bool
	deferred    bool
}

// NewDiscordResponder creates a new Discord responder
func NewDiscordResponder(s *discordgo.Session, i *discordgo.InteractionCreate) *DiscordResponder {
	return &DiscordResponder{
		session:     s,
		interaction: i,
	}
}

// Defer sends a deferred response
func (r *DiscordResponder) Defer(ephemeral bool) error {
	if r.responded || r.deferred {
		return fmt.Errorf("interaction already responded to")
	}

	flags := discordgo.MessageFlags(0)
	if ephemeral {
		flags = discordgo.MessageFlagsEphemeral
	}

	err := r.session.InteractionRespond(r.interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: flags,
		},
	})

	if err == nil {
		r.deferred = true
		r.responded = true
	}

	return err
}

// Respond sends an immediate response
func (r *DiscordResponder) Respond(response *Response) error {
	if r.responded {
		// If we've already responded, edit instead
		return r.Edit(response)
	}

	data := r.buildResponseData(response)

	err := r.session.InteractionRespond(r.interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: data,
	})

	if err == nil {
		r.responded = true
	}

	return err
}

// Edit updates a previous response
func (r *DiscordResponder) Edit(response *Response) error {
	if !r.responded {
		return fmt.Errorf("cannot edit before responding")
	}

	webhook := &discordgo.WebhookEdit{
		Content:         &response.Content,
		Embeds:          &response.Embeds,
		Components:      &response.Components,
		Files:           response.Files,
		AllowedMentions: response.AllowedMentions,
	}

	_, err := r.session.InteractionResponseEdit(r.interaction.Interaction, webhook)
	return err
}

// FollowUp sends an additional message after the initial response
func (r *DiscordResponder) FollowUp(response *Response) (*discordgo.Message, error) {
	if !r.responded {
		return nil, fmt.Errorf("cannot follow up before responding")
	}

	data := r.buildFollowUpData(response)
	return r.session.FollowupMessageCreate(r.interaction.Interaction, true, data)
}

// DeleteFollowUp deletes a follow-up message
func (r *DiscordResponder) DeleteFollowUp(messageID string) error {
	return r.session.FollowupMessageDelete(r.interaction.Interaction, messageID)
}

// DeleteOriginal deletes the original response
func (r *DiscordResponder) DeleteOriginal() error {
	return r.session.InteractionResponseDelete(r.interaction.Interaction)
}

// buildResponseData converts our Response to Discord's InteractionResponseData
func (r *DiscordResponder) buildResponseData(response *Response) *discordgo.InteractionResponseData {
	data := &discordgo.InteractionResponseData{
		Content:         response.Content,
		Embeds:          response.Embeds,
		Components:      response.Components,
		Files:           response.Files,
		AllowedMentions: response.AllowedMentions,
		TTS:             response.TTS,
	}

	if response.Ephemeral {
		data.Flags = discordgo.MessageFlagsEphemeral
	}

	return data
}

// buildFollowUpData converts our Response to Discord's WebhookParams
func (r *DiscordResponder) buildFollowUpData(response *Response) *discordgo.WebhookParams {
	params := &discordgo.WebhookParams{
		Content:         response.Content,
		Embeds:          response.Embeds,
		Components:      response.Components,
		Files:           response.Files,
		AllowedMentions: response.AllowedMentions,
		TTS:             response.TTS,
	}

	if response.Ephemeral {
		params.Flags = discordgo.MessageFlagsEphemeral
	}

	return params
}

// HasResponded returns whether this responder has already sent a response
func (r *DiscordResponder) HasResponded() bool {
	return r.responded
}

// IsDeferred returns whether this responder has sent a deferred response
func (r *DiscordResponder) IsDeferred() bool {
	return r.deferred
}
