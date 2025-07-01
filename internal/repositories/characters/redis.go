package characters

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/character"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/equipment"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/shared"

	dnderr "github.com/KirkDiggler/dnd-bot-discord/internal/errors"
	"github.com/KirkDiggler/dnd-bot-discord/internal/uuid"
	"github.com/redis/go-redis/v9"
)

// EquipmentData wraps equipment with type information for JSON marshaling
type EquipmentData struct {
	Type      string          `json:"type"`
	Equipment json.RawMessage `json:"equipment"`
}

// CharacterData represents the serialized form of a character in Redis
type CharacterData struct {
	ID                 string                                               `json:"id"`
	OwnerID            string                                               `json:"owner_id"`
	RealmID            string                                               `json:"realm_id"`
	Name               string                                               `json:"name"`
	Speed              int                                                  `json:"speed"`
	Race               *rulebook.Race                                       `json:"race"`
	Class              *rulebook.Class                                      `json:"class"`
	Background         *rulebook.Background                                 `json:"background"`
	Attributes         map[shared.Attribute]*character.AbilityScore         `json:"attributes"`
	AbilityRolls       []character.AbilityRoll                              `json:"ability_rolls"`
	AbilityAssignments map[string]string                                    `json:"ability_assignments"`
	Proficiencies      map[rulebook.ProficiencyType][]*rulebook.Proficiency `json:"proficiencies"`
	HitDie             int                                                  `json:"hit_die"`
	AC                 int                                                  `json:"ac"`
	MaxHitPoints       int                                                  `json:"max_hit_points"`
	CurrentHitPoints   int                                                  `json:"current_hit_points"`
	Level              int                                                  `json:"level"`
	Experience         int                                                  `json:"experience"`
	Status             shared.CharacterStatus                               `json:"status"`
	Features           []*rulebook.CharacterFeature                         `json:"features"`
	Inventory          map[equipment.EquipmentType][]EquipmentData          `json:"inventory"`
	EquippedSlots      map[shared.Slot]EquipmentData                        `json:"equipped_slots"`
	Resources          *character.CharacterResources                        `json:"resources"`
	CreatedAt          time.Time                                            `json:"created_at"`
	UpdatedAt          time.Time                                            `json:"updated_at"`
}

// redisRepo implements the Repository interface using Redis
type redisRepo struct {
	client        redis.UniversalClient
	uuidGenerator uuid.Generator
	ttl           time.Duration // TTL for draft characters
}

// equipmentToData converts an Equipment interface to EquipmentData for storage
func equipmentToData(eq equipment.Equipment) (EquipmentData, error) {
	// Marshal the concrete type
	data, err := json.Marshal(eq)
	if err != nil {
		return EquipmentData{}, fmt.Errorf("failed to marshal equipment: %w", err)
	}

	// Determine the concrete type
	var typeStr string
	switch eq.(type) {
	case *equipment.Weapon:
		typeStr = "weapon"
	case *equipment.Armor:
		typeStr = "armor"
	case *equipment.BasicEquipment:
		typeStr = "basic"
	default:
		typeStr = "unknown"
	}

	return EquipmentData{
		Type:      typeStr,
		Equipment: data,
	}, nil
}

// dataToEquipment converts EquipmentData back to Equipment interface
func dataToEquipment(data EquipmentData) (equipment.Equipment, error) {
	// Normalize type to handle legacy data
	normalizedType := strings.ToLower(data.Type)

	switch normalizedType {
	case "weapon":
		var weapon equipment.Weapon
		if err := json.Unmarshal(data.Equipment, &weapon); err != nil {
			return nil, fmt.Errorf("failed to unmarshal weapon: %w", err)
		}
		return &weapon, nil
	case "armor":
		var armor equipment.Armor
		if err := json.Unmarshal(data.Equipment, &armor); err != nil {
			return nil, fmt.Errorf("failed to unmarshal armor: %w", err)
		}
		return &armor, nil
	case "basic", "basicequipment", "":
		var basic equipment.BasicEquipment
		if err := json.Unmarshal(data.Equipment, &basic); err != nil {
			return nil, fmt.Errorf("failed to unmarshal basic equipment: %w", err)
		}
		return &basic, nil
	default:
		// Handle unknown types by trying to detect from JSON structure
		var rawData map[string]interface{}
		if err := json.Unmarshal(data.Equipment, &rawData); err == nil {
			// Check for weapon-specific fields
			if _, hasWeaponCategory := rawData["weapon_category"]; hasWeaponCategory {
				var weapon equipment.Weapon
				if err := json.Unmarshal(data.Equipment, &weapon); err != nil {
					return nil, fmt.Errorf("failed to unmarshal weapon: %w", err)
				}
				return &weapon, nil
			}
			// Check for armor-specific fields
			if _, hasArmorCategory := rawData["armor_category"]; hasArmorCategory {
				var armor equipment.Armor
				if err := json.Unmarshal(data.Equipment, &armor); err != nil {
					return nil, fmt.Errorf("failed to unmarshal armor: %w", err)
				}
				return &armor, nil
			}
		}

		// Default to basic equipment
		var basic equipment.BasicEquipment
		if err := json.Unmarshal(data.Equipment, &basic); err != nil {
			return nil, fmt.Errorf("unknown equipment type '%s': %w", data.Type, err)
		}
		return &basic, nil
	}
}

// RedisRepoConfig holds configuration for the Redis repository
type RedisRepoConfig struct {
	Client        redis.UniversalClient
	UUIDGenerator uuid.Generator
	DraftTTL      time.Duration // How long to keep draft characters (default: 24 hours)
}

// NewRedisRepository creates a new Redis-backed character repository
func NewRedisRepository(cfg *RedisRepoConfig) Repository {
	if cfg == nil {
		panic("RedisRepoConfig cannot be nil")
	}
	if cfg.Client == nil {
		panic("Redis client cannot be nil")
	}
	if cfg.UUIDGenerator == nil {
		cfg.UUIDGenerator = uuid.NewGoogleUUIDGenerator()
	}

	ttl := cfg.DraftTTL
	if ttl == 0 {
		ttl = 24 * time.Hour // Default to 24 hours for drafts
	}

	return &redisRepo{
		client:        cfg.Client,
		uuidGenerator: cfg.UUIDGenerator,
		ttl:           ttl,
	}
}

// key generates the Redis key for a character
func (r *redisRepo) key(id string) string {
	return fmt.Sprintf("character:%s", id)
}

// ownerCharactersKey generates the Redis key for an owner's character list
func (r *redisRepo) ownerCharactersKey(ownerID string) string {
	return fmt.Sprintf("owner:%s:characters", ownerID)
}

// realmCharactersKey generates the Redis key for a realm's character list
func (r *redisRepo) realmCharactersKey(realmID string) string {
	return fmt.Sprintf("realm:%s:characters", realmID)
}

// ownerRealmCharactersKey generates the Redis key for an owner's characters in a specific realm
func (r *redisRepo) ownerRealmCharactersKey(ownerID, realmID string) string {
	return fmt.Sprintf("owner:%s:realm:%s:characters", ownerID, realmID)
}

// Create stores a new character
func (r *redisRepo) Create(ctx context.Context, char *character.Character) error {
	if char == nil {
		return dnderr.InvalidArgument("character cannot be nil")
	}
	if char.ID == "" {
		return dnderr.InvalidArgument("character ID is required")
	}
	if char.OwnerID == "" {
		return dnderr.InvalidArgument("character owner ID is required")
	}
	if char.RealmID == "" {
		return dnderr.InvalidArgument("character realm ID is required")
	}

	// Check if character already exists
	exists, err := r.client.Exists(ctx, r.key(char.ID)).Result()
	if err != nil {
		return fmt.Errorf("failed to check character existence: %w", err)
	}
	if exists > 0 {
		return dnderr.AlreadyExistsf("character with ID '%s' already exists", char.ID).
			WithMeta("character_id", char.ID)
	}

	// Convert to data struct
	data, err := r.toCharacterData(char)
	if err != nil {
		return fmt.Errorf("failed to convert character data: %w", err)
	}
	data.CreatedAt = time.Now().UTC()
	data.UpdatedAt = data.CreatedAt

	// Serialize character
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal character: %w", err)
	}

	// Store in Redis using pipeline for atomicity
	pipe := r.client.Pipeline()

	// Store character data
	pipe.Set(ctx, r.key(char.ID), jsonData, 0) // No expiration for finalized characters

	// Add to various index sets
	pipe.SAdd(ctx, r.ownerCharactersKey(char.OwnerID), char.ID)
	pipe.SAdd(ctx, r.realmCharactersKey(char.RealmID), char.ID)
	pipe.SAdd(ctx, r.ownerRealmCharactersKey(char.OwnerID, char.RealmID), char.ID)

	// Execute pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create character: %w", err)
	}

	return nil
}

// Get retrieves a character by ID
func (r *redisRepo) Get(ctx context.Context, id string) (*character.Character, error) {
	if id == "" {
		return nil, dnderr.InvalidArgument("character ID is required")
	}

	// Get character data from Redis
	jsonData, err := r.client.Get(ctx, r.key(id)).Result()
	if err == redis.Nil {
		return nil, dnderr.NotFoundf("character with ID '%s' not found", id).
			WithMeta("character_id", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get character: %w", err)
	}

	// Deserialize character data
	var data CharacterData
	if unmarshalErr := json.Unmarshal([]byte(jsonData), &data); unmarshalErr != nil {
		return nil, fmt.Errorf("failed to unmarshal character: %w", unmarshalErr)
	}

	// Debug: Log features being loaded
	log.Printf("DEBUG REDIS: Loading character %s with %d features", data.Name, len(data.Features))
	for i, feature := range data.Features {
		log.Printf("DEBUG REDIS: Loaded Feature %d: key=%s, metadata=%v", i, feature.Key, feature.Metadata)
	}

	// Convert to entity
	char, err := r.fromCharacterData(&data)
	if err != nil {
		return nil, fmt.Errorf("failed to convert character from data: %w", err)
	}
	return char, nil
}

// GetByOwner retrieves all characters for a specific owner
func (r *redisRepo) GetByOwner(ctx context.Context, ownerID string) ([]*character.Character, error) {
	if ownerID == "" {
		return nil, dnderr.InvalidArgument("owner ID is required")
	}

	// Get character IDs for owner
	ids, err := r.client.SMembers(ctx, r.ownerCharactersKey(ownerID)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to list character IDs: %w", err)
	}

	// Get each character
	characters := make([]*character.Character, 0, len(ids))
	for _, id := range ids {
		char, err := r.Get(ctx, id)
		if err != nil {
			// Skip characters that can't be loaded
			continue
		}
		characters = append(characters, char)
	}

	return characters, nil
}

// GetByOwnerAndRealm retrieves all characters for a specific owner in a realm
func (r *redisRepo) GetByOwnerAndRealm(ctx context.Context, ownerID, realmID string) ([]*character.Character, error) {
	if ownerID == "" {
		return nil, dnderr.InvalidArgument("owner ID is required")
	}
	if realmID == "" {
		return nil, dnderr.InvalidArgument("realm ID is required")
	}

	// Get character IDs for owner in realm
	ids, err := r.client.SMembers(ctx, r.ownerRealmCharactersKey(ownerID, realmID)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to list character IDs: %w", err)
	}

	// Get each character
	characters := make([]*character.Character, 0, len(ids))
	for _, id := range ids {
		char, err := r.Get(ctx, id)
		if err != nil {
			// Skip characters that can't be loaded
			continue
		}
		characters = append(characters, char)
	}

	return characters, nil
}

// Update updates an existing character
func (r *redisRepo) Update(ctx context.Context, char *character.Character) error {
	if char == nil {
		return dnderr.InvalidArgument("character cannot be nil")
	}
	if char.ID == "" {
		return dnderr.InvalidArgument("character ID is required")
	}

	// Get existing character to verify it exists and preserve created timestamp
	existingData, err := r.client.Get(ctx, r.key(char.ID)).Result()
	if err == redis.Nil {
		return dnderr.NotFoundf("character with ID '%s' not found", char.ID).
			WithMeta("character_id", char.ID)
	}
	if err != nil {
		return fmt.Errorf("failed to get existing character: %w", err)
	}

	// Parse existing data to preserve created timestamp
	var existing CharacterData
	if marshErr := json.Unmarshal([]byte(existingData), &existing); marshErr != nil {
		return fmt.Errorf("failed to unmarshal existing character: %w", marshErr)
	}

	// Convert to data struct
	data, err := r.toCharacterData(char)
	if err != nil {
		return fmt.Errorf("failed to convert character data: %w", err)
	}
	data.CreatedAt = existing.CreatedAt // Preserve creation time
	data.UpdatedAt = time.Now().UTC()

	// Debug: Log features being saved
	log.Printf("DEBUG REDIS: Saving character %s with %d features", char.Name, len(data.Features))
	for i, feature := range data.Features {
		log.Printf("DEBUG REDIS: Feature %d: key=%s, metadata=%v", i, feature.Key, feature.Metadata)
	}

	// Serialize updated character
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal character: %w", err)
	}

	// Update in Redis
	err = r.client.Set(ctx, r.key(char.ID), jsonData, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to update character: %w", err)
	}

	// If owner or realm changed, update indexes
	if existing.OwnerID != char.OwnerID || existing.RealmID != char.RealmID {
		pipe := r.client.Pipeline()

		// Remove from old indexes
		pipe.SRem(ctx, r.ownerCharactersKey(existing.OwnerID), char.ID)
		pipe.SRem(ctx, r.realmCharactersKey(existing.RealmID), char.ID)
		pipe.SRem(ctx, r.ownerRealmCharactersKey(existing.OwnerID, existing.RealmID), char.ID)

		// Add to new indexes
		pipe.SAdd(ctx, r.ownerCharactersKey(char.OwnerID), char.ID)
		pipe.SAdd(ctx, r.realmCharactersKey(char.RealmID), char.ID)
		pipe.SAdd(ctx, r.ownerRealmCharactersKey(char.OwnerID, char.RealmID), char.ID)

		_, err = pipe.Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to update character indexes: %w", err)
		}
	}

	return nil
}

// Delete removes a character
func (r *redisRepo) Delete(ctx context.Context, id string) error {
	if id == "" {
		return dnderr.InvalidArgument("character ID is required")
	}

	// Get character to find owner/realm for cleanup
	char, err := r.Get(ctx, id)
	if err != nil {
		return err
	}

	// Remove using pipeline
	pipe := r.client.Pipeline()

	// Remove character data
	pipe.Del(ctx, r.key(id))

	// Remove from index sets
	pipe.SRem(ctx, r.ownerCharactersKey(char.OwnerID), id)
	pipe.SRem(ctx, r.realmCharactersKey(char.RealmID), id)
	pipe.SRem(ctx, r.ownerRealmCharactersKey(char.OwnerID, char.RealmID), id)

	// Execute pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete character: %w", err)
	}

	return nil
}

// toCharacterData converts an entity to the data struct for storage
func (r *redisRepo) toCharacterData(char *character.Character) (*CharacterData, error) {
	// Convert inventory
	inventory := make(map[equipment.EquipmentType][]EquipmentData)
	for eqType, items := range char.Inventory {
		var dataItems []EquipmentData
		for _, item := range items {
			data, err := equipmentToData(item)
			if err != nil {
				return nil, fmt.Errorf("failed to convert inventory item: %w", err)
			}
			dataItems = append(dataItems, data)
		}
		inventory[eqType] = dataItems
	}

	// Convert equipped slots
	equippedSlots := make(map[shared.Slot]EquipmentData)
	for slot, item := range char.EquippedSlots {
		// Skip nil items (empty slots)
		if item == nil {
			continue
		}
		data, err := equipmentToData(item)
		if err != nil {
			return nil, fmt.Errorf("failed to convert equipped item: %w", err)
		}
		equippedSlots[slot] = data
	}

	return &CharacterData{
		ID:                 char.ID,
		OwnerID:            char.OwnerID,
		RealmID:            char.RealmID,
		Name:               char.Name,
		Speed:              char.Speed,
		Race:               char.Race,
		Class:              char.Class,
		Background:         char.Background,
		Attributes:         char.Attributes,
		AbilityRolls:       char.AbilityRolls,
		AbilityAssignments: char.AbilityAssignments,
		Proficiencies:      char.Proficiencies,
		HitDie:             char.HitDie,
		AC:                 char.AC,
		MaxHitPoints:       char.MaxHitPoints,
		CurrentHitPoints:   char.CurrentHitPoints,
		Level:              char.Level,
		Experience:         char.Experience,
		Status:             char.Status,
		Features:           char.Features,
		Inventory:          inventory,
		EquippedSlots:      equippedSlots,
		Resources:          char.Resources,
	}, nil
}

// fromCharacterData converts a data struct to an entity
func (r *redisRepo) fromCharacterData(data *CharacterData) (*character.Character, error) {
	// Convert inventory back
	inventory := make(map[equipment.EquipmentType][]equipment.Equipment)
	for eqType, items := range data.Inventory {
		var eqItems []equipment.Equipment
		for _, item := range items {
			eq, err := DataToEquipmentWithMigration(item)
			if err != nil {
				return nil, fmt.Errorf("failed to convert inventory data: %w", err)
			}
			eqItems = append(eqItems, eq)
		}
		inventory[eqType] = eqItems
	}

	// Convert equipped slots back
	equippedSlots := make(map[shared.Slot]equipment.Equipment)
	for slot, item := range data.EquippedSlots {
		eq, err := DataToEquipmentWithMigration(item)
		if err != nil {
			return nil, fmt.Errorf("failed to convert equipped data: %w", err)
		}
		equippedSlots[slot] = eq
	}

	return &character.Character{
		ID:                 data.ID,
		OwnerID:            data.OwnerID,
		RealmID:            data.RealmID,
		Name:               data.Name,
		Speed:              data.Speed,
		Race:               data.Race,
		Class:              data.Class,
		Background:         data.Background,
		Attributes:         data.Attributes,
		AbilityRolls:       data.AbilityRolls,
		AbilityAssignments: data.AbilityAssignments,
		Proficiencies:      data.Proficiencies,
		HitDie:             data.HitDie,
		AC:                 data.AC,
		MaxHitPoints:       data.MaxHitPoints,
		CurrentHitPoints:   data.CurrentHitPoints,
		Level:              data.Level,
		Experience:         data.Experience,
		Status:             data.Status,
		Features:           data.Features,
		Inventory:          inventory,
		EquippedSlots:      equippedSlots,
		Resources:          data.Resources,
	}, nil
}
