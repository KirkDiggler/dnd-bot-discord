package dnd5e

//go:generate mockgen -destination=mock/mock_client.go -package=mockdnd5e . Client

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
)

type Client interface {
	ListClasses() ([]*entities.Class, error)
	ListRaces() ([]*entities.Race, error)
	GetRace(key string) (*entities.Race, error)
	GetClass(key string) (*entities.Class, error)
	GetProficiency(key string) (*entities.Proficiency, error)
	GetMonster(key string) (*entities.MonsterTemplate, error)
	GetEquipment(key string) (entities.Equipment, error)
	GetEquipmentByCategory(category string) ([]entities.Equipment, error)

	// New methods needed by services
	ListClassFeatures(classKey string, level int) ([]*entities.CharacterFeature, error)
	ListMonstersByCR(minCR, maxCR float32) ([]*entities.MonsterTemplate, error)
	ListEquipment() ([]entities.Equipment, error)
}
