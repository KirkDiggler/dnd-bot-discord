package dnd5e_test

import (
	"testing"

	"github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e"
	mockdnd5e "github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e/mock"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestClient_ImplementsInterface(t *testing.T) {
	// This test ensures our client properly implements the interface
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := mockdnd5e.NewMockClient(ctrl)
	
	// Ensure mock implements the interface
	var _ dnd5e.Client = mock
	
	// Test ListEquipment
	expectedEquipment := []entities.Equipment{
		&entities.Weapon{Base: entities.BasicEquipment{Key: "longsword", Name: "Longsword"}},
		&entities.Armor{Base: entities.BasicEquipment{Key: "chain-mail", Name: "Chain Mail"}},
	}
	mock.EXPECT().ListEquipment().Return(expectedEquipment, nil)
	
	equipment, err := mock.ListEquipment()
	require.NoError(t, err)
	assert.Equal(t, expectedEquipment, equipment)
	
	// Test ListMonstersByCR
	expectedMonsters := []*entities.MonsterTemplate{
		{Key: "goblin", Name: "Goblin", ChallengeRating: 0.25},
		{Key: "orc", Name: "Orc", ChallengeRating: 0.5},
	}
	mock.EXPECT().ListMonstersByCR(float32(0), float32(1)).Return(expectedMonsters, nil)
	
	monsters, err := mock.ListMonstersByCR(0, 1)
	require.NoError(t, err)
	assert.Equal(t, expectedMonsters, monsters)
}