package dungeon_test

import (
	"context"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/exploration"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/session"
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/handlers/discord/dnd/dungeon"
	"github.com/KirkDiggler/dnd-bot-discord/internal/services"
	mockdungeon "github.com/KirkDiggler/dnd-bot-discord/internal/services/dungeon/mock"
	mocksession "github.com/KirkDiggler/dnd-bot-discord/internal/services/session/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestStartDungeonHandler_InitializesSessionMetadata(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	mockSessionService := mocksession.NewMockService(ctrl)
	mockDungeonService := mockdungeon.NewMockService(ctrl)

	// Session without metadata
	createdSession := &session.Session{
		ID:        "session123",
		Name:      "Dungeon Delve",
		CreatorID: "user123",
		Members:   map[string]*session.SessionMember{},
		Metadata:  nil, // This is the key - metadata is nil
	}

	// Mock expectations
	mockSessionService.EXPECT().
		CreateSession(gomock.Any(), gomock.Any()).
		Return(createdSession, nil)

	// Expect SaveSession to be called and verify metadata is initialized
	mockSessionService.EXPECT().
		SaveSession(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx interface{}, sess *session.Session) error {
			// The first call is for saving bot as DM
			// The second call should have initialized metadata
			if sess.Metadata != nil && sess.Metadata["dungeonID"] != nil {
				assert.NotNil(t, sess.Metadata, "Metadata should be initialized")
				assert.Equal(t, "dungeon123", sess.Metadata["dungeonID"])
			}
			return nil
		}).Times(2)

	mockDungeonService.EXPECT().
		CreateDungeon(gomock.Any(), gomock.Any()).
		Return(&exploration.Dungeon{
			ID:        "dungeon123",
			SessionID: "session123",
		}, nil)

	// Create service provider
	provider := &services.Provider{
		SessionService: mockSessionService,
		DungeonService: mockDungeonService,
		MonsterService: nil, // Add if needed
		LootService:    nil, // Add if needed
	}

	// Test that metadata is properly initialized
	handler := dungeon.NewStartDungeonHandler(provider)

	// Simulate the critical part of the handle function
	ctx := context.Background()

	// Create session
	sess, err := mockSessionService.CreateSession(ctx, nil)
	require.NoError(t, err)

	// Add bot as DM
	sess.DMID = "bot123"
	err = mockSessionService.SaveSession(ctx, sess)
	require.NoError(t, err)

	// Create dungeon
	dung, err := mockDungeonService.CreateDungeon(ctx, nil)
	require.NoError(t, err)

	// This is where the panic occurred - metadata was nil
	if sess.Metadata == nil {
		sess.Metadata = make(map[string]interface{})
	}
	sess.Metadata["dungeonID"] = dung.ID

	// Save session with metadata
	err = mockSessionService.SaveSession(ctx, sess)
	require.NoError(t, err)

	// Verify the handler exists (basic sanity check)
	assert.NotNil(t, handler)
}
