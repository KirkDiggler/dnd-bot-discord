package core

import (
	"context"
	"github.com/bwmarrin/discordgo"
)

// TestInteractionContext creates an InteractionContext for testing
type TestInteractionContext struct {
	*InteractionContext
	TestParams map[string]interface{}
}

// NewTestInteractionContext creates a test interaction context
func NewTestInteractionContext() *TestInteractionContext {
	ctx := &InteractionContext{
		Context: context.Background(),
		UserID:  "test-user-123",
		GuildID: "test-guild-123",
		params:  make(map[string]interface{}),
	}

	return &TestInteractionContext{
		InteractionContext: ctx,
		TestParams:         make(map[string]interface{}),
	}
}

// WithParam adds a parameter for testing
func (t *TestInteractionContext) WithParam(key string, value interface{}) *TestInteractionContext {
	t.params[key] = value
	return t
}

// WithUserID sets the user ID
func (t *TestInteractionContext) WithUserID(userID string) *TestInteractionContext {
	t.UserID = userID
	return t
}

// WithGuildID sets the guild ID
func (t *TestInteractionContext) WithGuildID(guildID string) *TestInteractionContext {
	t.GuildID = guildID
	return t
}

// WithRoles sets the roles for the member
func (t *TestInteractionContext) WithRoles(roles []string) *TestInteractionContext {
	if t.Interaction == nil {
		t.Interaction = &discordgo.InteractionCreate{
			Interaction: &discordgo.Interaction{},
		}
	}
	if t.Interaction.Member == nil {
		t.Interaction.Member = &discordgo.Member{
			User: &discordgo.User{
				ID: t.UserID,
			},
		}
	}
	t.Interaction.Member.Roles = roles

	// Update the InteractionContext's Member field as well
	t.Member = t.Interaction.Member
	return t
}

// WithValue adds a value to the context
func (t *TestInteractionContext) WithValue(key string, value interface{}) *TestInteractionContext {
	t.params[key] = value
	return t
}

// AsCommand simulates a command interaction
func (t *TestInteractionContext) AsCommand(name string, subcommand ...string) *TestInteractionContext {
	t.Interaction = &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			Type: discordgo.InteractionApplicationCommand,
			Data: discordgo.ApplicationCommandInteractionData{
				Name: name,
			},
		},
	}

	if len(subcommand) > 0 {
		t.params["subcommand"] = subcommand[0]
	}

	return t
}

// AsComponent simulates a component interaction
func (t *TestInteractionContext) AsComponent(customID string) *TestInteractionContext {
	t.Interaction = &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			Type: discordgo.InteractionMessageComponent,
			Data: discordgo.MessageComponentInteractionData{
				CustomID: customID,
			},
		},
	}
	return t
}

// MockResponder is a test implementation of InteractionResponder
type MockResponder struct {
	DeferCalls   []bool // Track ephemeral flags
	Responses    []*Response
	Edits        []*Response
	FollowUps    []*Response
	DeferError   error
	RespondError error
	EditError    error
	Deferred     bool
	Responded    bool
}

// NewMockResponder creates a new mock responder
func NewMockResponder() *MockResponder {
	return &MockResponder{
		DeferCalls: make([]bool, 0),
		Responses:  make([]*Response, 0),
		Edits:      make([]*Response, 0),
		FollowUps:  make([]*Response, 0),
	}
}

func (m *MockResponder) Defer(ephemeral bool) error {
	m.DeferCalls = append(m.DeferCalls, ephemeral)
	m.Deferred = true
	return m.DeferError
}

func (m *MockResponder) Respond(response *Response) error {
	m.Responses = append(m.Responses, response)
	m.Responded = true
	return m.RespondError
}

func (m *MockResponder) Edit(response *Response) error {
	m.Edits = append(m.Edits, response)
	return m.EditError
}

func (m *MockResponder) FollowUp(response *Response) (*discordgo.Message, error) {
	m.FollowUps = append(m.FollowUps, response)
	return &discordgo.Message{ID: "test-message-123"}, nil
}

func (m *MockResponder) DeleteFollowUp(messageID string) error {
	return nil
}

func (m *MockResponder) DeleteOriginal() error {
	return nil
}

func (m *MockResponder) HasResponded() bool {
	return m.Responded
}

func (m *MockResponder) IsDeferred() bool {
	return m.Deferred
}

// LastResponse returns the last response sent
func (m *MockResponder) LastResponse() *Response {
	if len(m.Responses) > 0 {
		return m.Responses[len(m.Responses)-1]
	}
	if len(m.Edits) > 0 {
		return m.Edits[len(m.Edits)-1]
	}
	return nil
}
