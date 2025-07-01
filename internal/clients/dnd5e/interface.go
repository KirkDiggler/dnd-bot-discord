package dnd5e

//go:generate mockgen -destination=mock/mock_client.go -package=mockdnd5e . Client

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/equipment"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/combat"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
)

type Client interface {
	ListClasses() ([]*rulebook.Class, error)
	ListRaces() ([]*rulebook.Race, error)
	GetRace(key string) (*rulebook.Race, error)
	GetClass(key string) (*rulebook.Class, error)
	GetProficiency(key string) (*rulebook.Proficiency, error)
	GetMonster(key string) (*combat.MonsterTemplate, error)
	GetEquipment(key string) (equipment.Equipment, error)
	GetEquipmentByCategory(category string) ([]equipment.Equipment, error)

	// New methods needed by services
	GetClassFeatures(classKey string, level int) ([]*rulebook.CharacterFeature, error)
	ListMonstersByCR(minCR, maxCR float32) ([]*combat.MonsterTemplate, error)
	ListEquipment() ([]equipment.Equipment, error)
}
