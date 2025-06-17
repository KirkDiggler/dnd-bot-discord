package dnd5e

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
	GetClassLevels(classKey string) ([]*entities.ClassLevel, error)
	GetFeature(featureKey string) (*entities.Feature, error)
}
