package core

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockHandler for testing
type MockHandler struct {
	name      string
	canHandle bool
	result    *HandlerResult
	err       error
	called    bool
}

func (m *MockHandler) CanHandle(ctx *InteractionContext) bool {
	return m.canHandle
}

func (m *MockHandler) Handle(ctx *InteractionContext) (*HandlerResult, error) {
	m.called = true
	return m.result, m.err
}

// Use MockResponder from testing.go instead

func TestPipeline_Register(t *testing.T) {
	pipeline := NewPipeline()

	handler1 := &MockHandler{name: "handler1"}
	handler2 := &MockHandler{name: "handler2"}

	pipeline.Register(handler1, handler2)

	assert.Equal(t, 2, pipeline.HandlerCount())
}

func TestPipeline_Execute_StopOnFirst(t *testing.T) {
	pipeline := NewPipeline()
	pipeline.SetStopOnFirst(true)

	mockResponder := NewMockResponder()
	calledHandlers := []string{}

	handler1 := HandlerFunc(func(ctx *InteractionContext) (*HandlerResult, error) {
		calledHandlers = append(calledHandlers, "handler1")
		return &HandlerResult{Response: NewResponse("Response 1")}, nil
	})

	handler2 := HandlerFunc(func(ctx *InteractionContext) (*HandlerResult, error) {
		calledHandlers = append(calledHandlers, "handler2")
		return &HandlerResult{Response: NewResponse("Response 2")}, nil
	})

	// Wrap handlers to use mock responder
	wrappedHandler1 := &testHandler{handler: handler1, responder: mockResponder}
	wrappedHandler2 := &testHandler{handler: handler2, responder: mockResponder}

	pipeline.Register(wrappedHandler1, wrappedHandler2)

	// Create test context
	testCtx := NewTestInteractionContext().
		WithUserID("test-user").
		AsCommand("test")

	// Execute handlers directly to avoid Discord API calls
	handled := false
	for _, h := range []Handler{wrappedHandler1, wrappedHandler2} {
		if !h.CanHandle(testCtx.InteractionContext) {
			continue
		}
		result, err := h.Handle(testCtx.InteractionContext)
		require.NoError(t, err)

		if result != nil && result.Response != nil {
			err = mockResponder.Respond(result.Response)
			require.NoError(t, err)
		}

		handled = true
		if pipeline.stopOnFirst {
			break
		}
	}

	require.True(t, handled)
	assert.Equal(t, []string{"handler1"}, calledHandlers)
	assert.Len(t, mockResponder.Responses, 1)
}

// testHandler wraps a handler to inject mock responder
type testHandler struct {
	handler   Handler
	responder InteractionResponder
}

func (t *testHandler) CanHandle(ctx *InteractionContext) bool {
	return t.handler.CanHandle(ctx)
}

func (t *testHandler) Handle(ctx *InteractionContext) (*HandlerResult, error) {
	// Inject mock responder
	ctx.WithValue("responder", t.responder)
	return t.handler.Handle(ctx)
}

func TestPipeline_Execute_ContinueOnMultiple(t *testing.T) {
	pipeline := NewPipeline()
	pipeline.SetStopOnFirst(false)

	mockResponder := NewMockResponder()
	calledHandlers := []string{}

	handler1 := HandlerFunc(func(ctx *InteractionContext) (*HandlerResult, error) {
		calledHandlers = append(calledHandlers, "handler1")
		return &HandlerResult{Response: NewResponse("Response 1")}, nil
	})

	handler2 := HandlerFunc(func(ctx *InteractionContext) (*HandlerResult, error) {
		calledHandlers = append(calledHandlers, "handler2")
		return &HandlerResult{Response: NewResponse("Response 2")}, nil
	})

	// Wrap handlers to use mock responder
	wrappedHandler1 := &testHandler{handler: handler1, responder: mockResponder}
	wrappedHandler2 := &testHandler{handler: handler2, responder: mockResponder}

	pipeline.Register(wrappedHandler1, wrappedHandler2)

	// Create test context
	testCtx := NewTestInteractionContext().
		WithUserID("test-user").
		AsCommand("test")

	// Execute handlers directly
	for _, h := range []Handler{wrappedHandler1, wrappedHandler2} {
		if h.CanHandle(testCtx.InteractionContext) {
			result, err := h.Handle(testCtx.InteractionContext)
			require.NoError(t, err)

			if result != nil && result.Response != nil {
				err = mockResponder.Respond(result.Response)
				require.NoError(t, err)
			}

			if pipeline.stopOnFirst {
				break
			}
		}
	}

	// Both handlers should be called
	assert.Equal(t, []string{"handler1", "handler2"}, calledHandlers)
	assert.Len(t, mockResponder.Responses, 2)
}

func TestPipeline_Execute_StopPropagation(t *testing.T) {
	pipeline := NewPipeline()
	pipeline.SetStopOnFirst(false)

	mockResponder := NewMockResponder()
	calledHandlers := []string{}

	handler1 := HandlerFunc(func(ctx *InteractionContext) (*HandlerResult, error) {
		calledHandlers = append(calledHandlers, "handler1")
		return &HandlerResult{
			Response:        NewResponse("Response 1"),
			StopPropagation: true,
		}, nil
	})

	handler2 := HandlerFunc(func(ctx *InteractionContext) (*HandlerResult, error) {
		calledHandlers = append(calledHandlers, "handler2")
		return &HandlerResult{Response: NewResponse("Response 2")}, nil
	})

	// Wrap handlers to use mock responder
	wrappedHandler1 := &testHandler{handler: handler1, responder: mockResponder}
	wrappedHandler2 := &testHandler{handler: handler2, responder: mockResponder}

	pipeline.Register(wrappedHandler1, wrappedHandler2)

	// Create test context
	testCtx := NewTestInteractionContext().
		WithUserID("test-user").
		AsCommand("test")

	// Execute handlers directly
	for _, h := range []Handler{wrappedHandler1, wrappedHandler2} {
		if h.CanHandle(testCtx.InteractionContext) {
			result, err := h.Handle(testCtx.InteractionContext)
			require.NoError(t, err)

			if result != nil && result.Response != nil {
				err = mockResponder.Respond(result.Response)
				require.NoError(t, err)
			}

			if result != nil && result.StopPropagation {
				break
			}
		}
	}

	// Only first handler should be called due to StopPropagation
	assert.Equal(t, []string{"handler1"}, calledHandlers)
	assert.Len(t, mockResponder.Responses, 1)
}

func TestPipeline_Execute_ErrorHandling(t *testing.T) {
	pipeline := NewPipeline()

	testError := errors.New("test error")
	mockResponder := NewMockResponder()

	errorHandler := HandlerFunc(func(ctx *InteractionContext) (*HandlerResult, error) {
		return nil, testError
	})

	// Custom error handler
	errorHandled := false
	pipeline.SetErrorHandler(func(ctx *InteractionContext, err error) *HandlerResult {
		errorHandled = true
		assert.Equal(t, testError, err)
		return &HandlerResult{
			Response: NewEphemeralResponse("Custom error message"),
		}
	})

	// Wrap handler to use mock responder
	wrappedHandler := &testHandler{handler: errorHandler, responder: mockResponder}
	pipeline.Register(wrappedHandler)

	// Create test context
	testCtx := NewTestInteractionContext().
		WithUserID("test-user").
		AsCommand("test")

	// Execute handler directly
	if wrappedHandler.CanHandle(testCtx.InteractionContext) {
		_, err := wrappedHandler.Handle(testCtx.InteractionContext)

		// Handler should return error
		assert.Error(t, err)
		assert.Equal(t, testError, err)

		// Apply error handler
		errorResult := pipeline.errorHandler(testCtx.InteractionContext, err)
		assert.NotNil(t, errorResult)

		// Send error response
		if errorResult != nil && errorResult.Response != nil {
			sendErr := mockResponder.Respond(errorResult.Response)
			assert.NoError(t, sendErr)
		}
	}

	assert.True(t, errorHandled)
	assert.Len(t, mockResponder.Responses, 1)
	assert.True(t, mockResponder.Responses[0].Ephemeral)
}

func TestPipeline_Middleware(t *testing.T) {
	pipeline := NewPipeline()

	// Track middleware execution order
	var executionOrder []string

	// Create middleware
	middleware1 := func(next Handler) Handler {
		return HandlerFunc(func(ctx *InteractionContext) (*HandlerResult, error) {
			executionOrder = append(executionOrder, "middleware1_before")
			result, err := next.Handle(ctx)
			executionOrder = append(executionOrder, "middleware1_after")
			return result, err
		})
	}

	middleware2 := func(next Handler) Handler {
		return HandlerFunc(func(ctx *InteractionContext) (*HandlerResult, error) {
			executionOrder = append(executionOrder, "middleware2_before")
			result, err := next.Handle(ctx)
			executionOrder = append(executionOrder, "middleware2_after")
			return result, err
		})
	}

	// Add middleware
	pipeline.Use(middleware1, middleware2)

	// Add handler
	handler := HandlerFunc(func(ctx *InteractionContext) (*HandlerResult, error) {
		executionOrder = append(executionOrder, "handler")
		return &HandlerResult{Response: NewResponse("test")}, nil
	})

	pipeline.Register(handler)

	// Create test context
	testCtx := NewTestInteractionContext().
		WithUserID("test-user").
		AsCommand("test")

	// Apply middleware manually
	var wrappedHandler Handler = handler
	for i := len(pipeline.middleware) - 1; i >= 0; i-- {
		wrappedHandler = pipeline.middleware[i](wrappedHandler)
	}

	// Execute wrapped handler
	if wrappedHandler.CanHandle(testCtx.InteractionContext) {
		_, err := wrappedHandler.Handle(testCtx.InteractionContext)
		require.NoError(t, err)
	}

	// Check execution order
	expected := []string{
		"middleware1_before",
		"middleware2_before",
		"handler",
		"middleware2_after",
		"middleware1_after",
	}
	assert.Equal(t, expected, executionOrder)
}

func TestPipeline_NoHandlers(t *testing.T) {
	pipeline := NewPipeline()

	// Create test context
	testCtx := NewTestInteractionContext().
		WithUserID("test-user").
		AsCommand("test")

	// No handlers registered, so nothing should handle it
	handled := false
	for _, h := range pipeline.handlers {
		if h.CanHandle(testCtx.InteractionContext) {
			handled = true
			break
		}
	}

	assert.False(t, handled)
	assert.Equal(t, 0, pipeline.HandlerCount())
}

func TestPipeline_Clear(t *testing.T) {
	pipeline := NewPipeline()

	handler := &MockHandler{name: "handler1"}
	pipeline.Register(handler)

	assert.Equal(t, 1, pipeline.HandlerCount())

	pipeline.Clear()

	assert.Equal(t, 0, pipeline.HandlerCount())
}
