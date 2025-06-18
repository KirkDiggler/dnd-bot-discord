package characters

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
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
	Race               *entities.Race                                       `json:"race"`
	Class              *entities.Class                                      `json:"class"`
	Background         *entities.Background                                 `json:"background"`
	Attributes         map[entities.Attribute]*entities.AbilityScore        `json:"attributes"`
	AbilityRolls       []entities.AbilityRoll                               `json:"ability_rolls"`
	AbilityAssignments map[string]string                                    `json:"ability_assignments"`
	Proficiencies      map[entities.ProficiencyType][]*entities.Proficiency `json:"proficiencies"`
	HitDie             int                                                  `json:"hit_die"`
	AC                 int                                                  `json:"ac"`
	MaxHitPoints       int                                                  `json:"max_hit_points"`
	CurrentHitPoints   int                                                  `json:"current_hit_points"`
	Level              int                                                  `json:"level"`
	Experience         int                                                  `json:"experience"`
	Status             entities.CharacterStatus                             `json:"status"`
	Features           []*entities.CharacterFeature                         `json:"features"`
	Inventory          map[entities.EquipmentType][]EquipmentData           `json:"inventory"`
	EquippedSlots      map[entities.Slot]EquipmentData                      `json:"equipped_slots"`
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
func equipmentToData(eq entities.Equipment) (EquipmentData, error) {
	// Marshal the concrete type
	data, err := json.Marshal(eq)
	if err != nil {
		return EquipmentData{}, fmt.Errorf("failed to marshal equipment: %w", err)
	}

	// Determine the concrete type
	var typeStr string
	switch eq.(type) {
	case *entities.Weapon:
		typeStr = "weapon"
	case *entities.Armor:
		typeStr = "armor"
	case *entities.BasicEquipment:
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
func dataToEquipment(data EquipmentData) (entities.Equipment, error) {
	switch data.Type {
	case "weapon":
		var weapon entities.Weapon
		if err := json.Unmarshal(data.Equipment, &weapon); err != nil {
			return nil, fmt.Errorf("failed to unmarshal weapon: %w", err)
		}
		return &weapon, nil
	case "armor":
		var armor entities.Armor
		if err := json.Unmarshal(data.Equipment, &armor); err != nil {
			return nil, fmt.Errorf("failed to unmarshal armor: %w", err)
		}
		return &armor, nil
	case "basic":
		var basic entities.BasicEquipment
		if err := json.Unmarshal(data.Equipment, &basic); err != nil {
			return nil, fmt.Errorf("failed to unmarshal basic equipment: %w", err)
		}
		return &basic, nil
	default:
		return nil, fmt.Errorf("unknown equipment type: %s", data.Type)
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
func (r *redisRepo) Create(ctx context.Context, character *entities.Character) error {
	if character == nil {
		return dnderr.InvalidArgument("character cannot be nil")
	}
	if character.ID == "" {
		return dnderr.InvalidArgument("character ID is required")
	}
	if character.OwnerID == "" {
		return dnderr.InvalidArgument("character owner ID is required")
	}
	if character.RealmID == "" {
		return dnderr.InvalidArgument("character realm ID is required")
	}

	// Check if character already exists
	exists, err := r.client.Exists(ctx, r.key(character.ID)).Result()
	if err != nil {
		return fmt.Errorf("failed to check character existence: %w", err)
	}
	if exists > 0 {
		return dnderr.AlreadyExistsf("character with ID '%s' already exists", character.ID).
			WithMeta("character_id", character.ID)
	}

	// Convert to data struct
	data, err := r.toCharacterData(character)
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
	pipe.Set(ctx, r.key(character.ID), jsonData, 0) // No expiration for finalized characters

	// Add to various index sets
	pipe.SAdd(ctx, r.ownerCharactersKey(character.OwnerID), character.ID)
	pipe.SAdd(ctx, r.realmCharactersKey(character.RealmID), character.ID)
	pipe.SAdd(ctx, r.ownerRealmCharactersKey(character.OwnerID, character.RealmID), character.ID)

	// Execute pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create character: %w", err)
	}

	return nil
}

// Get retrieves a character by ID
func (r *redisRepo) Get(ctx context.Context, id string) (*entities.Character, error) {
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

	// Convert to entity
	char, err := r.fromCharacterData(&data)
	if err != nil {
		return nil, fmt.Errorf("failed to convert character from data: %w", err)
	}
	return char, nil
}

// GetByOwner retrieves all characters for a specific owner
func (r *redisRepo) GetByOwner(ctx context.Context, ownerID string) ([]*entities.Character, error) {
	if ownerID == "" {
		return nil, dnderr.InvalidArgument("owner ID is required")
	}

	// Get character IDs for owner
	ids, err := r.client.SMembers(ctx, r.ownerCharactersKey(ownerID)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to list character IDs: %w", err)
	}

	// Get each character
	characters := make([]*entities.Character, 0, len(ids))
	for _, id := range ids {
		character, err := r.Get(ctx, id)
		if err != nil {
			// Skip characters that can't be loaded
			continue
		}
		characters = append(characters, character)
	}

	return characters, nil
}

// GetByOwnerAndRealm retrieves all characters for a specific owner in a realm
func (r *redisRepo) GetByOwnerAndRealm(ctx context.Context, ownerID, realmID string) ([]*entities.Character, error) {
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
	characters := make([]*entities.Character, 0, len(ids))
	for _, id := range ids {
		character, err := r.Get(ctx, id)
		if err != nil {
			// Skip characters that can't be loaded
			continue
		}
		characters = append(characters, character)
	}

	return characters, nil
}

// Update updates an existing character
func (r *redisRepo) Update(ctx context.Context, character *entities.Character) error {
	if character == nil {
		return dnderr.InvalidArgument("character cannot be nil")
	}
	if character.ID == "" {
		return dnderr.InvalidArgument("character ID is required")
	}

	// Get existing character to verify it exists and preserve created timestamp
	existingData, err := r.client.Get(ctx, r.key(character.ID)).Result()
	if err == redis.Nil {
		return dnderr.NotFoundf("character with ID '%s' not found", character.ID).
			WithMeta("character_id", character.ID)
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
	data, err := r.toCharacterData(character)
	if err != nil {
		return fmt.Errorf("failed to convert character data: %w", err)
	}
	data.CreatedAt = existing.CreatedAt // Preserve creation time
	data.UpdatedAt = time.Now().UTC()

	// Serialize updated character
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal character: %w", err)
	}

	// Update in Redis
	err = r.client.Set(ctx, r.key(character.ID), jsonData, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to update character: %w", err)
	}

	// If owner or realm changed, update indexes
	if existing.OwnerID != character.OwnerID || existing.RealmID != character.RealmID {
		pipe := r.client.Pipeline()

		// Remove from old indexes
		pipe.SRem(ctx, r.ownerCharactersKey(existing.OwnerID), character.ID)
		pipe.SRem(ctx, r.realmCharactersKey(existing.RealmID), character.ID)
		pipe.SRem(ctx, r.ownerRealmCharactersKey(existing.OwnerID, existing.RealmID), character.ID)

		// Add to new indexes
		pipe.SAdd(ctx, r.ownerCharactersKey(character.OwnerID), character.ID)
		pipe.SAdd(ctx, r.realmCharactersKey(character.RealmID), character.ID)
		pipe.SAdd(ctx, r.ownerRealmCharactersKey(character.OwnerID, character.RealmID), character.ID)

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
	character, err := r.Get(ctx, id)
	if err != nil {
		return err
	}

	// Remove using pipeline
	pipe := r.client.Pipeline()

	// Remove character data
	pipe.Del(ctx, r.key(id))

	// Remove from index sets
	pipe.SRem(ctx, r.ownerCharactersKey(character.OwnerID), id)
	pipe.SRem(ctx, r.realmCharactersKey(character.RealmID), id)
	pipe.SRem(ctx, r.ownerRealmCharactersKey(character.OwnerID, character.RealmID), id)

	// Execute pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete character: %w", err)
	}

	return nil
}

// toCharacterData converts an entity to the data struct for storage
func (r *redisRepo) toCharacterData(char *entities.Character) (*CharacterData, error) {
	// Convert inventory
	inventory := make(map[entities.EquipmentType][]EquipmentData)
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
	equippedSlots := make(map[entities.Slot]EquipmentData)
	for slot, item := range char.EquippedSlots {
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
	}, nil
}

// fromCharacterData converts a data struct to an entity
func (r *redisRepo) fromCharacterData(data *CharacterData) (*entities.Character, error) {
	// Convert inventory back
	inventory := make(map[entities.EquipmentType][]entities.Equipment)
	for eqType, items := range data.Inventory {
		var eqItems []entities.Equipment
		for _, item := range items {
			eq, err := dataToEquipment(item)
			if err != nil {
				return nil, fmt.Errorf("failed to convert inventory data: %w", err)
			}
			eqItems = append(eqItems, eq)
		}
		inventory[eqType] = eqItems
	}

	// Convert equipped slots back
	equippedSlots := make(map[entities.Slot]entities.Equipment)
	for slot, item := range data.EquippedSlots {
		eq, err := dataToEquipment(item)
		if err != nil {
			return nil, fmt.Errorf("failed to convert equipped data: %w", err)
		}
		equippedSlots[slot] = eq
	}

	return &entities.Character{
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
	}, nil
}
