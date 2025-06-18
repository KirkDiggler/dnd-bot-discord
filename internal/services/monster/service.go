package monster

//go:generate mockgen -destination=mock/mock_service.go -package=mockmonster -source=service.go

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

	"github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/entities/damage"
	dnderr "github.com/KirkDiggler/dnd-bot-discord/internal/errors"
)

// Service defines the monster service interface

type Service interface {
	// GetMonster fetches a specific monster by key
	GetMonster(ctx context.Context, key string) (*entities.MonsterTemplate, error)

	// GetMonstersByCR returns monsters within a CR range
	GetMonstersByCR(ctx context.Context, minCR, maxCR float32) ([]*entities.MonsterTemplate, error)

	// GetRandomMonsters returns random monsters for a given difficulty
	GetRandomMonsters(ctx context.Context, difficulty string, count int) ([]*entities.MonsterTemplate, error)

	// GetMonsterForEncounter converts a monster template to encounter format
	GetMonsterForEncounter(template *entities.MonsterTemplate) *MonsterEncounterData
}

// MonsterEncounterData represents monster data formatted for the encounter service
type MonsterEncounterData struct {
	Name            string
	MaxHP           int
	AC              int
	Speed           int
	InitiativeBonus int
	CR              float64
	XP              int
	Abilities       map[string]int
	Actions         []*entities.MonsterAction
}

type service struct {
	dndClient dnd5e.Client
	// Cache of monster templates by key
	monsterCache map[string]*entities.MonsterTemplate
}

// ServiceConfig holds configuration for the service
type ServiceConfig struct {
	DNDClient dnd5e.Client // Required
}

// NewService creates a new monster service
func NewService(cfg *ServiceConfig) Service {
	if cfg.DNDClient == nil {
		panic("DND client is required")
	}

	return &service{
		dndClient:    cfg.DNDClient,
		monsterCache: make(map[string]*entities.MonsterTemplate),
	}
}

// GetMonster fetches a specific monster by key
func (s *service) GetMonster(ctx context.Context, key string) (*entities.MonsterTemplate, error) {
	if key == "" {
		return nil, dnderr.InvalidArgument("monster key is req")
	}

	// Check cache first
	if cached, ok := s.monsterCache[key]; ok {
		return cached, nil
	}

	// Fetch from API
	monster, err := s.dndClient.GetMonster(key)
	if err != nil {
		return nil, dnderr.Wrapf(err, "failed to get monster '%s'", key)
	}

	// Cache it
	s.monsterCache[key] = monster

	return monster, nil
}

// GetMonstersByCR returns monsters within a CR range
func (s *service) GetMonstersByCR(ctx context.Context, minCR, maxCR float32) ([]*entities.MonsterTemplate, error) {
	// Try to use API if available
	if s.dndClient != nil {
		var allMonsters []*entities.MonsterTemplate

		// The API requires exact CR values, so we need to query for each CR in range
		// Common CR values: 0, 0.125, 0.25, 0.5, 1, 2, 3, 4, 5, etc.
		crValues := []float64{0, 0.125, 0.25, 0.5, 1, 2, 3, 4, 5, 6, 7, 8}

		for _, cr := range crValues {
			if float32(cr) >= minCR && float32(cr) <= maxCR {
				// For now, get all monsters within the CR range
				// TODO: Update when API supports better CR filtering
				monsters, err := s.dndClient.ListMonstersByCR(minCR, maxCR)
				if err == nil {
					allMonsters = append(allMonsters, monsters...)
					break // Don't duplicate results
				}
			}
		}

		if len(allMonsters) > 0 {
			return allMonsters, nil
		}
	}

	// Fallback to hardcoded list
	var monsters []*entities.MonsterTemplate

	// CR 0-0.5
	if minCR <= 0.5 {
		monsters = append(monsters, s.getHardcodedMonsters("goblin", "skeleton", "kobold", "rat", "spider")...)
	}

	// CR 0.5-1
	if minCR <= 1 && maxCR >= 0.5 {
		monsters = append(monsters, s.getHardcodedMonsters("orc", "wolf", "zombie", "hobgoblin")...)
	}

	// CR 1-2
	if minCR <= 2 && maxCR >= 1 {
		monsters = append(monsters, s.getHardcodedMonsters("dire-wolf", "ghoul", "ogre", "bugbear")...)
	}

	// CR 2+
	if maxCR >= 2 {
		monsters = append(monsters, s.getHardcodedMonsters("owlbear", "troll", "wight")...)
	}

	return monsters, nil
}

// GetRandomMonsters returns random monsters for a given difficulty
func (s *service) GetRandomMonsters(ctx context.Context, difficulty string, count int) ([]*entities.MonsterTemplate, error) {
	var minCR, maxCR float32

	switch strings.ToLower(difficulty) {
	case "easy":
		minCR, maxCR = 0, 0.5
	case "medium":
		minCR, maxCR = 0.25, 1
	case "hard":
		minCR, maxCR = 0.5, 2
	case "deadly":
		minCR, maxCR = 1, 3
	default:
		return nil, dnderr.InvalidArgument("difficulty must be easy, medium, hard, or deadly")
	}

	// Get monsters in CR range
	availableMonsters, err := s.GetMonstersByCR(ctx, minCR, maxCR)
	if err != nil {
		return nil, err
	}

	if len(availableMonsters) == 0 {
		return nil, dnderr.NotFound("no monsters found for difficulty")
	}

	// Select random monsters
	result := make([]*entities.MonsterTemplate, 0, count)
	for i := 0; i < count; i++ {
		idx := rand.Intn(len(availableMonsters))
		result = append(result, availableMonsters[idx])
	}

	return result, nil
}

// GetMonsterForEncounter converts a monster template to encounter format
func (s *service) GetMonsterForEncounter(template *entities.MonsterTemplate) *MonsterEncounterData {
	if template == nil {
		return nil
	}

	// For now, return hardcoded data based on monster key
	// In a full implementation, we'd parse the monster template data
	return s.getHardcodedEncounterData(template.Key)
}

// getHardcodedMonsters returns hardcoded monster templates
func (s *service) getHardcodedMonsters(keys ...string) []*entities.MonsterTemplate {
	var monsters []*entities.MonsterTemplate
	for _, key := range keys {
		monsters = append(monsters, &entities.MonsterTemplate{
			Key:  key,
			Name: strings.Title(strings.ReplaceAll(key, "-", " ")),
		})
	}
	return monsters
}

// getHardcodedEncounterData returns hardcoded encounter data for common monsters
func (s *service) getHardcodedEncounterData(key string) *MonsterEncounterData {
	monsters := map[string]*MonsterEncounterData{
		"goblin": {
			Name:            "Goblin",
			MaxHP:           7,
			AC:              15,
			Speed:           30,
			InitiativeBonus: 2,
			CR:              0.25,
			XP:              50,
			Abilities: map[string]int{
				"STR": 8, "DEX": 14, "CON": 10,
				"INT": 10, "WIS": 8, "CHA": 8,
			},
			Actions: []*entities.MonsterAction{
				{
					Name:        "Scimitar",
					AttackBonus: 4,
					Description: "Melee Weapon Attack: +4 to hit, reach 5 ft., one target.",
					Damage: []*damage.Damage{{
						DiceCount: 1, DiceSize: 6, Bonus: 2,
						DamageType: damage.TypeSlashing,
					}},
				},
			},
		},
		"skeleton": {
			Name:            "Skeleton",
			MaxHP:           13,
			AC:              13,
			Speed:           30,
			InitiativeBonus: 2,
			CR:              0.25,
			XP:              50,
			Abilities: map[string]int{
				"STR": 10, "DEX": 14, "CON": 15,
				"INT": 6, "WIS": 8, "CHA": 5,
			},
			Actions: []*entities.MonsterAction{
				{
					Name:        "Shortsword",
					AttackBonus: 4,
					Description: "Melee Weapon Attack: +4 to hit, reach 5 ft., one target.",
					Damage: []*damage.Damage{{
						DiceCount: 1, DiceSize: 6, Bonus: 2,
						DamageType: damage.TypePiercing,
					}},
				},
			},
		},
		"orc": {
			Name:            "Orc",
			MaxHP:           15,
			AC:              13,
			Speed:           30,
			InitiativeBonus: 1,
			CR:              0.5,
			XP:              100,
			Abilities: map[string]int{
				"STR": 16, "DEX": 12, "CON": 16,
				"INT": 7, "WIS": 11, "CHA": 10,
			},
			Actions: []*entities.MonsterAction{
				{
					Name:        "Greataxe",
					AttackBonus: 5,
					Description: "Melee Weapon Attack: +5 to hit, reach 5 ft., one target.",
					Damage: []*damage.Damage{{
						DiceCount: 1, DiceSize: 12, Bonus: 3,
						DamageType: damage.TypeSlashing,
					}},
				},
			},
		},
		"dire-wolf": {
			Name:            "Dire Wolf",
			MaxHP:           37,
			AC:              14,
			Speed:           50,
			InitiativeBonus: 2,
			CR:              1,
			XP:              200,
			Abilities: map[string]int{
				"STR": 17, "DEX": 15, "CON": 15,
				"INT": 3, "WIS": 12, "CHA": 7,
			},
			Actions: []*entities.MonsterAction{
				{
					Name:        "Bite",
					AttackBonus: 5,
					Description: "Melee Weapon Attack: +5 to hit, reach 5 ft., one target.",
					Damage: []*damage.Damage{{
						DiceCount: 2, DiceSize: 6, Bonus: 3,
						DamageType: damage.TypePiercing,
					}},
				},
			},
		},
		"owlbear": {
			Name:            "Owlbear",
			MaxHP:           59,
			AC:              13,
			Speed:           40,
			InitiativeBonus: 1,
			CR:              3,
			XP:              700,
			Abilities: map[string]int{
				"STR": 20, "DEX": 12, "CON": 17,
				"INT": 3, "WIS": 12, "CHA": 7,
			},
			Actions: []*entities.MonsterAction{
				{
					Name:        "Beak",
					AttackBonus: 7,
					Description: "Melee Weapon Attack: +7 to hit, reach 5 ft., one creature.",
					Damage: []*damage.Damage{{
						DiceCount: 1, DiceSize: 10, Bonus: 5,
						DamageType: damage.TypePiercing,
					}},
				},
			},
		},
	}

	if data, ok := monsters[key]; ok {
		return data
	}

	// Default monster if not found
	return &MonsterEncounterData{
		Name:            fmt.Sprintf("Unknown Monster (%s)", key),
		MaxHP:           10,
		AC:              12,
		Speed:           30,
		InitiativeBonus: 0,
		CR:              0.25,
		XP:              50,
		Abilities: map[string]int{
			"STR": 10, "DEX": 10, "CON": 10,
			"INT": 10, "WIS": 10, "CHA": 10,
		},
	}
}
