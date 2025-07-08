package v2_test

import (
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/discord/v2/builders"
	"github.com/KirkDiggler/dnd-bot-discord/internal/discord/v2/core"
	"github.com/KirkDiggler/dnd-bot-discord/internal/discord/v2/middleware"
	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockDiscordSession implements a mock Discord session for testing
type MockDiscordSession struct {
	InteractionResponses []discordgo.InteractionResponse
	WebhookEdits         []discordgo.WebhookEdit
	RespondError         error
	EditError            error
}

func (m *MockDiscordSession) InteractionRespond(interaction *discordgo.Interaction, resp *discordgo.InteractionResponse) error {
	m.InteractionResponses = append(m.InteractionResponses, *resp)
	return m.RespondError
}

func (m *MockDiscordSession) InteractionResponseEdit(interaction *discordgo.Interaction, edit *discordgo.WebhookEdit) (*discordgo.Message, error) {
	m.WebhookEdits = append(m.WebhookEdits, *edit)
	return nil, m.EditError
}

func (m *MockDiscordSession) FollowupMessageCreate(interaction *discordgo.Interaction, wait bool, data *discordgo.WebhookParams) (*discordgo.Message, error) {
	return nil, nil
}

func (m *MockDiscordSession) FollowupMessageDelete(interaction *discordgo.Interaction, messageID string) error {
	return nil
}

func (m *MockDiscordSession) InteractionResponseDelete(interaction *discordgo.Interaction) error {
	return nil
}

func (m *MockDiscordSession) Guild(guildID string) (*discordgo.Guild, error) {
	return &discordgo.Guild{
		ID:      guildID,
		OwnerID: "owner123",
	}, nil
}

func (m *MockDiscordSession) UserChannelPermissions(userID, channelID string) (int64, error) {
	// Return admin permissions for testing
	return discordgo.PermissionAdministrator, nil
}

func TestPipelineIntegration(t *testing.T) {
	// Create a test pipeline
	pipeline := core.NewPipeline()

	// Track handler execution
	handlerCalled := false
	receivedResponse := (*core.Response)(nil)

	// Create a test handler
	testHandler := core.HandlerFunc(func(ctx *core.InteractionContext) (*core.HandlerResult, error) {
		handlerCalled = true

		// Verify context parsing
		assert.Equal(t, "test", ctx.GetCommandName())
		assert.Equal(t, "user123", ctx.UserID)

		// Return a response
		response := core.NewResponse("Test successful!").
			AsEphemeral()

		receivedResponse = response

		return &core.HandlerResult{
			Response: response,
		}, nil
	})

	pipeline.Register(testHandler)

	// Create test context directly
	_ = core.NewTestInteractionContext().
		WithUserID("user123").
		AsCommand("test")

	// Verify handler was registered
	require.Equal(t, 1, pipeline.HandlerCount())

	// Execute the handler directly (since pipeline.Execute requires Discord session)
	testCtx := core.NewTestInteractionContext().
		WithUserID("user123").
		AsCommand("test")

	// Create a mock responder
	mockResponder := core.NewMockResponder()
	testCtx.WithValue("responder", mockResponder)

	// The handler should be called by the pipeline
	result, err := testHandler.Handle(testCtx.InteractionContext)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify handler was called and response was created
	assert.True(t, handlerCalled)
	assert.NotNil(t, receivedResponse)
	assert.Equal(t, "Test successful!", receivedResponse.Content)
	assert.True(t, receivedResponse.Ephemeral)
}

func TestRouterIntegration(t *testing.T) {
	// Create pipeline
	pipeline := core.NewPipeline()

	// Create router
	router := core.NewRouter("test", pipeline)

	// Track handler calls
	listCalled := false
	showCalled := false
	buttonCalled := false
	var lastResponse *core.Response

	// Register routes for the "test" domain
	router.HandleFunc("cmd:test:list", func(ctx *core.InteractionContext) (*core.HandlerResult, error) {
		listCalled = true
		lastResponse = core.NewResponse("List response")
		return &core.HandlerResult{
			Response: lastResponse,
		}, nil
	})

	router.HandleFunc("cmd:test:show", func(ctx *core.InteractionContext) (*core.HandlerResult, error) {
		showCalled = true
		id := ctx.GetStringParam("id")
		lastResponse = core.NewResponse("Showing: " + id)
		return &core.HandlerResult{
			Response: lastResponse,
		}, nil
	})

	router.ComponentFunc("button", func(ctx *core.InteractionContext) (*core.HandlerResult, error) {
		buttonCalled = true
		customID, err := core.ParseCustomID(ctx.GetCustomID())
		if err != nil {
			return nil, core.NewInternalError(err)
		}

		lastResponse = core.NewResponse("Button pressed: " + customID.Target)
		return &core.HandlerResult{
			Response: lastResponse,
		}, nil
	})

	// Register router
	router.Register()

	// Verify router was registered
	require.Equal(t, 1, pipeline.HandlerCount())

	// Get the built handler from the router
	handler := router.Build()

	// Test list command
	t.Run("ListCommand", func(t *testing.T) {
		listCalled = false
		lastResponse = nil

		testCtx := core.NewTestInteractionContext().
			WithUserID("user123").
			AsCommand("test").
			WithParam("subcommand", "list")

		if handler.CanHandle(testCtx.InteractionContext) {
			result, err := handler.Handle(testCtx.InteractionContext)
			require.NoError(t, err)
			require.NotNil(t, result)
		}

		assert.True(t, listCalled)
		assert.NotNil(t, lastResponse)
		assert.Equal(t, "List response", lastResponse.Content)
	})

	// Test show command with parameter
	t.Run("ShowCommand", func(t *testing.T) {
		showCalled = false
		lastResponse = nil

		testCtx := core.NewTestInteractionContext().
			WithUserID("user123").
			AsCommand("test").
			WithParam("subcommand", "show").
			WithParam("id", "item123")

		if handler.CanHandle(testCtx.InteractionContext) {
			result, err := handler.Handle(testCtx.InteractionContext)
			require.NoError(t, err)
			require.NotNil(t, result)
		}

		assert.True(t, showCalled)
		assert.NotNil(t, lastResponse)
		assert.Equal(t, "Showing: item123", lastResponse.Content)
	})

	// Test button component
	t.Run("ButtonComponent", func(t *testing.T) {
		buttonCalled = false
		lastResponse = nil

		testCtx := core.NewTestInteractionContext().
			WithUserID("user123").
			AsComponent("test:button:target123")

		if handler.CanHandle(testCtx.InteractionContext) {
			result, err := handler.Handle(testCtx.InteractionContext)
			require.NoError(t, err)
			require.NotNil(t, result)
		}

		assert.True(t, buttonCalled)
		assert.NotNil(t, lastResponse)
		assert.Equal(t, "Button pressed: target123", lastResponse.Content)
	})
}

func TestMiddlewareIntegration(t *testing.T) {
	t.Run("ErrorMiddleware", func(t *testing.T) {
		pipeline := core.NewPipeline()

		// Create error middleware config
		errorConfig := &middleware.ErrorConfig{}
		pipeline.Use(middleware.ErrorMiddleware(errorConfig))

		var errorResponse *core.Response

		// Handler that returns an error
		pipeline.Register(core.HandlerFunc(func(ctx *core.InteractionContext) (*core.HandlerResult, error) {
			return nil, core.NewValidationError("Invalid input")
		}))

		// Verify handler was registered
		require.Equal(t, 1, pipeline.HandlerCount())

		// Create mock responder to capture response
		mockResponder := core.NewMockResponder()

		testCtx := core.NewTestInteractionContext().
			WithUserID("user123").
			AsCommand("test").
			WithValue("responder", mockResponder)

		// Since middleware is applied during registration, test the error handling directly
		// Create a test handler that should have middleware applied
		testHandler := core.HandlerFunc(func(ctx *core.InteractionContext) (*core.HandlerResult, error) {
			return nil, core.NewValidationError("Invalid input")
		})

		// Apply middleware manually for testing
		errorMiddleware := middleware.ErrorMiddleware(errorConfig)
		wrappedHandler := errorMiddleware(testHandler)

		result, err := wrappedHandler.Handle(testCtx.InteractionContext)

		// Error should be handled by middleware
		assert.NoError(t, err)
		assert.NotNil(t, result)

		if result != nil && result.Response != nil {
			errorResponse = result.Response
		}

		// Check error response
		assert.NotNil(t, errorResponse)
		assert.Contains(t, errorResponse.Content, "Invalid input")
		assert.True(t, errorResponse.Ephemeral)
	})

	t.Run("AuthMiddleware", func(t *testing.T) {
		pipeline := core.NewPipeline()

		// Use role required middleware
		pipeline.Use(middleware.RoleRequiredMiddleware("admin_role"))

		// Handler that should only run if authorized
		handlerCalled := false
		var lastResponse *core.Response

		pipeline.Register(core.HandlerFunc(func(ctx *core.InteractionContext) (*core.HandlerResult, error) {
			handlerCalled = true
			lastResponse = core.NewResponse("Authorized!")
			return &core.HandlerResult{
				Response: lastResponse,
			}, nil
		}))

		// Verify handler was registered
		require.Equal(t, 1, pipeline.HandlerCount())

		// Apply middleware manually for testing
		authMiddleware := middleware.RoleRequiredMiddleware("admin_role")
		testHandler := core.HandlerFunc(func(ctx *core.InteractionContext) (*core.HandlerResult, error) {
			handlerCalled = true
			lastResponse = core.NewResponse("Authorized!")
			return &core.HandlerResult{
				Response: lastResponse,
			}, nil
		})
		wrappedHandler := authMiddleware(testHandler)

		// Test without required role
		handlerCalled = false
		testCtx := core.NewTestInteractionContext().
			WithUserID("user123").
			AsCommand("test").
			WithRoles([]string{"user_role"})

		result, err := wrappedHandler.Handle(testCtx.InteractionContext)
		require.NoError(t, err)
		assert.False(t, handlerCalled)
		assert.NotNil(t, result)
		assert.NotNil(t, result.Response)
		assert.Contains(t, result.Response.Content, "required role")

		// Test with required role
		handlerCalled = false
		testCtx = core.NewTestInteractionContext().
			WithUserID("user123").
			AsCommand("test").
			WithRoles([]string{"admin_role"})

		_, err = wrappedHandler.Handle(testCtx.InteractionContext)
		require.NoError(t, err)
		assert.True(t, handlerCalled)
		assert.NotNil(t, lastResponse)
		assert.Equal(t, "Authorized!", lastResponse.Content)
	})
}

func TestBuildersIntegration(t *testing.T) {
	t.Run("EmbedBuilder", func(t *testing.T) {
		embed := builders.SuccessEmbed("Test Passed", "Everything is working!").
			AddField("Status", "✅ Operational", true).
			AddField("Version", "v2.0", true).
			Footer("Test Suite").
			Build()

		assert.Equal(t, "✅ Test Passed", embed.Title)
		assert.Equal(t, "Everything is working!", embed.Description)
		assert.Equal(t, builders.ColorSuccess, embed.Color)
		assert.Len(t, embed.Fields, 2)
	})

	t.Run("ComponentBuilder", func(t *testing.T) {
		customIDBuilder := core.NewCustomIDBuilder("test")
		builder := builders.NewComponentBuilder(customIDBuilder)

		components := builder.
			PrimaryButton("Click Me", "action", "target123").
			SecondaryButton("Cancel", "cancel", "target123").
			NewRow().
			SelectMenu("Choose an option", "select", []builders.SelectOption{
				{Label: "Option 1", Value: "opt1"},
				{Label: "Option 2", Value: "opt2"},
			}).
			Build()

		assert.Len(t, components, 2) // Two rows

		// Check first row
		row1, ok := components[0].(discordgo.ActionsRow)
		assert.True(t, ok)
		assert.Len(t, row1.Components, 2) // Two buttons

		// Check button custom IDs
		btn1, ok := row1.Components[0].(discordgo.Button)
		assert.True(t, ok)
		assert.Equal(t, "test:action:target123", btn1.CustomID)
	})
}
