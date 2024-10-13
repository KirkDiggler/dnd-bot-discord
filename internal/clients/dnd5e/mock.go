package dnd5e

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/stretchr/testify/mock"
)

type Mock struct {
	mock.Mock
}

func (m *Mock) ListClasses() ([]*entities.Class, error) {
	args := m.Called()

	if args.Error(1) != nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]*entities.Class), nil
}

func (m *Mock) ListRaces() ([]*entities.Race, error) {
	args := m.Called()

	if args.Error(1) != nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]*entities.Race), nil
}

func (m *Mock) GetRace(key string) (*entities.Race, error) {
	args := m.Called(key)

	if args.Error(1) != nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*entities.Race), nil
}

func (m *Mock) GetClass(key string) (*entities.Class, error) {
	args := m.Called(key)

	if args.Error(1) != nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*entities.Class), nil
}

func (m *Mock) GetProficiency(key string) (*entities.Proficiency, error) {
	args := m.Called(key)

	if args.Error(1) != nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*entities.Proficiency), nil
}

func (m *Mock) GetEquipment(key string) (entities.Equipment, error) {
	args := m.Called(key)

	if args.Error(1) != nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(entities.Equipment), nil
}

func (m *Mock) GetMonster(key string) (*entities.MonsterTemplate, error) {
	args := m.Called(key)

	if args.Error(1) != nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*entities.MonsterTemplate), nil
}
