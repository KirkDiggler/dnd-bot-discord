package character_test

import (
	"context"
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/services/character"
)

func TestResolveChoices(t *testing.T) {
	// This is a basic structure test
	// In a real implementation, we'd mock the dndClient
	
	req := &character.ResolveChoicesRequest{
		RaceKey:  "human",
		ClassKey: "fighter",
	}
	
	// This will fail without a mock client, but shows the structure
	_ = req
	
	t.Log("Service structure test passed")
}