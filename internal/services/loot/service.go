package loot

//go:generate mockgen -destination=mock/mock_service.go -package=mockloot -source=service.go

import (
	"context"
	"fmt"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/equipment"
	"log"
	"math/rand"
	"time"

	"github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e"
	dnderr "github.com/KirkDiggler/dnd-bot-discord/internal/errors"
)

// Service defines the loot service interface
type Service interface {
	// GenerateTreasure generates treasure based on difficulty and room number
	GenerateTreasure(ctx context.Context, difficulty string, roomNumber int) ([]string, error)

	// GenerateLootTable creates a loot table for a given CR
	GenerateLootTable(ctx context.Context, challengeRating float64) (*LootTable, error)
}

// LootTable represents a collection of possible loot
type LootTable struct {
	Gold      GoldRange
	Items     []LootItem
	MagicItem *equipment.Equipment // Optional magic item
}

// GoldRange represents a range of gold pieces
type GoldRange struct {
	Min int
	Max int
}

// LootItem represents a single item that can be looted
type LootItem struct {
	Name   string
	Chance float64 // 0.0 to 1.0
}

type service struct {
	dndClient dnd5e.Client
	random    *rand.Rand
}

// ServiceConfig holds configuration for the service
type ServiceConfig struct {
	DNDClient dnd5e.Client // Optional - will use hardcoded loot if nil
}

// NewService creates a new loot service
func NewService(cfg *ServiceConfig) Service {
	svc := &service{
		random: rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	if cfg != nil && cfg.DNDClient != nil {
		svc.dndClient = cfg.DNDClient
	}

	return svc
}

// GenerateTreasure generates treasure based on difficulty and room number
func (s *service) GenerateTreasure(ctx context.Context, difficulty string, roomNumber int) ([]string, error) {
	if difficulty == "" {
		return nil, dnderr.InvalidArgument("difficulty is required")
	}

	var treasure []string

	// Base gold amount
	goldMin, goldMax := s.getGoldRange(difficulty, roomNumber)
	goldAmount := goldMin + s.random.Intn(goldMax-goldMin+1)
	treasure = append(treasure, fmt.Sprintf("%d gold pieces", goldAmount))

	// Healing potions
	potionChance := 0.3
	if difficulty == "hard" {
		potionChance = 0.5
	}
	if s.random.Float64() < potionChance {
		potionCount := 1 + s.random.Intn(2)
		if potionCount == 1 {
			treasure = append(treasure, "healing potion")
		} else {
			treasure = append(treasure, fmt.Sprintf("%d healing potions", potionCount))
		}
	}

	// Equipment from API if available
	if s.dndClient != nil && roomNumber%5 == 0 {
		// Try to get equipment from API
		equipment, err := s.getRandomEquipment(ctx)
		if err == nil && equipment != nil {
			treasure = append(treasure, equipment.GetName())
		}
	}

	// Special items for higher room numbers
	if roomNumber >= 5 {
		specialItems := s.getSpecialItems(difficulty, roomNumber)
		treasure = append(treasure, specialItems...)
	}

	return treasure, nil
}

// GenerateLootTable creates a loot table for a given CR
func (s *service) GenerateLootTable(ctx context.Context, challengeRating float64) (*LootTable, error) {
	table := &LootTable{
		Gold:  s.getGoldRangeForCR(challengeRating),
		Items: []LootItem{},
	}

	// Add common items
	if challengeRating < 1 {
		table.Items = append(table.Items,
			LootItem{Name: "healing potion", Chance: 0.25},
			LootItem{Name: "torch (5)", Chance: 0.15},
			LootItem{Name: "rations (3 days)", Chance: 0.20},
		)
	} else if challengeRating < 5 {
		table.Items = append(table.Items,
			LootItem{Name: "healing potion", Chance: 0.35},
			LootItem{Name: "greater healing potion", Chance: 0.10},
			LootItem{Name: "antitoxin", Chance: 0.15},
			LootItem{Name: "alchemist's fire", Chance: 0.10},
		)
	} else {
		table.Items = append(table.Items,
			LootItem{Name: "greater healing potion", Chance: 0.30},
			LootItem{Name: "superior healing potion", Chance: 0.15},
			LootItem{Name: "potion of speed", Chance: 0.05},
			LootItem{Name: "potion of giant strength", Chance: 0.05},
		)
	}

	// Try to add magic item from API
	if s.dndClient != nil && challengeRating >= 3 {
		magicItem, err := s.getRandomEquipment(ctx)
		if err != nil {
			log.Println("Failed to get random equipment:", err)
		} else if magicItem != nil {
			table.MagicItem = &magicItem
		}
	}

	return table, nil
}

// getGoldRange returns min/max gold for difficulty and room
func (s *service) getGoldRange(difficulty string, roomNumber int) (minValue, maxValue int) {
	base := 10
	switch difficulty {
	case "easy":
		base = 10
	case "medium":
		base = 25
	case "hard":
		base = 50
	}

	// Scale with room number
	multiplier := 1 + (roomNumber / 3)
	minValue = base * multiplier
	maxValue = base * multiplier * 3

	return minValue, maxValue
}

// getGoldRangeForCR returns gold range based on challenge rating
func (s *service) getGoldRangeForCR(cr float64) GoldRange {
	if cr < 1 {
		return GoldRange{Min: 5, Max: 30}
	} else if cr < 5 {
		return GoldRange{Min: 30, Max: 150}
	} else if cr < 10 {
		return GoldRange{Min: 150, Max: 500}
	}
	return GoldRange{Min: 500, Max: 2000}
}

// getRandomEquipment tries to get a random equipment item from the API
func (s *service) getRandomEquipment(ctx context.Context) (equipment.Equipment, error) {
	if s.dndClient == nil {
		return nil, dnderr.NotFound("DND client not available")
	}

	// Try to get equipment list
	equipment, err := s.dndClient.ListEquipment()
	if err != nil {
		return nil, err
	}

	if len(equipment) == 0 {
		return nil, dnderr.NotFound("no equipment available")
	}

	// Return random equipment
	return equipment[s.random.Intn(len(equipment))], nil
}

// getSpecialItems returns special items based on room progression
func (s *service) getSpecialItems(difficulty string, roomNumber int) []string {
	var items []string

	// Every 5 rooms, add a special item
	if roomNumber%5 == 0 {
		specialItems := []string{
			"mysterious scroll",
			"ancient tome",
			"gemstone",
			"silver necklace",
			"enchanted ring",
			"magical component",
		}
		items = append(items, specialItems[s.random.Intn(len(specialItems))])
	}

	// Boss rooms (every 10 rooms) get extra special loot
	if roomNumber%10 == 0 {
		bossItems := []string{
			"legendary artifact fragment",
			"rare spell scroll",
			"masterwork weapon",
			"suit of fine armor",
			"bag of holding",
		}
		items = append(items, bossItems[s.random.Intn(len(bossItems))])
	}

	return items
}
