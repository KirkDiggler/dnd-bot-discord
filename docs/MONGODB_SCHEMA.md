# MongoDB Schema Design

## Design Philosophy

Using MongoDB's document model to naturally represent D&D entities. Each collection represents a major domain concept, with embedded documents for related data that's always accessed together.

## Collections

### 1. Characters Collection

```javascript
{
  "_id": ObjectId("..."),
  "user_id": "discord_user_id",
  "name": "Thorin Ironforge",
  "level": 5,
  "experience": 6500,
  
  // Embedded for performance (always needed)
  "race": {
    "id": "dwarf",
    "name": "Mountain Dwarf",
    "size": "Medium",
    "speed": 25,
    "darkvision": 60,
    "traits": [
      {
        "name": "Dwarven Resilience",
        "description": "Advantage on saving throws against poison"
      }
    ],
    "ability_bonuses": {
      "str": 2,
      "con": 2
    }
  },
  
  "class": {
    "id": "fighter",
    "name": "Fighter",
    "hit_die": 10,
    "primary_ability": "str",
    "saves": ["str", "con"],
    "skills_available": 2,
    "armor_proficiencies": ["light", "medium", "heavy", "shields"],
    "weapon_proficiencies": ["simple", "martial"]
  },
  
  "background": {
    "id": "soldier",
    "name": "Soldier",
    "skills": ["athletics", "intimidation"],
    "languages": 1,
    "feature": "Military Rank"
  },
  
  "attributes": {
    "str": 18,
    "dex": 13,
    "con": 16,
    "int": 10,
    "wis": 12,
    "cha": 8
  },
  
  "hp": {
    "current": 44,
    "max": 44,
    "temp": 0,
    "hit_dice_current": 5,
    "hit_dice_total": 5
  },
  
  "combat_stats": {
    "ac": 18,
    "initiative_bonus": 1,
    "speed": 25,
    "proficiency_bonus": 3
  },
  
  "skills": {
    "acrobatics": {"proficient": false, "expertise": false},
    "athletics": {"proficient": true, "expertise": false},
    "intimidation": {"proficient": true, "expertise": false}
    // ... all skills
  },
  
  "equipment": [
    {
      "id": "chainmail",
      "name": "Chainmail",
      "type": "armor",
      "equipped": true,
      "armor": {
        "ac": 16,
        "stealth_disadvantage": true,
        "strength_requirement": 13
      }
    },
    {
      "id": "longsword",
      "name": "Longsword",
      "type": "weapon",
      "equipped": true,
      "weapon": {
        "damage": "1d8",
        "damage_type": "slashing",
        "properties": ["versatile"],
        "versatile_damage": "1d10"
      }
    }
  ],
  
  "inventory": [
    {
      "id": "healing_potion",
      "name": "Potion of Healing",
      "quantity": 3,
      "weight": 0.5,
      "value": 50
    }
  ],
  
  "features": [
    {
      "name": "Fighting Style",
      "source": "Fighter 1",
      "description": "Defense: +1 AC while wearing armor"
    },
    {
      "name": "Second Wind",
      "source": "Fighter 1", 
      "description": "Regain 1d10+5 HP",
      "uses_current": 1,
      "uses_max": 1,
      "recharge": "short_rest"
    },
    {
      "name": "Action Surge",
      "source": "Fighter 2",
      "description": "Take an additional action",
      "uses_current": 1,
      "uses_max": 1,
      "recharge": "short_rest"
    }
  ],
  
  "spells": [], // Empty for fighter, populated for casters
  
  "status": "active", // draft, active, archived
  "created_at": ISODate("2024-01-15T10:00:00Z"),
  "updated_at": ISODate("2024-01-15T10:00:00Z")
}
```

### 2. Campaigns Collection

```javascript
{
  "_id": ObjectId("..."),
  "name": "Lost Mines of Phandelver",
  "dm_user_id": "discord_user_id",
  "channel_id": "discord_channel_id",
  "players": [
    {
      "user_id": "discord_user_id",
      "character_id": ObjectId("..."),
      "joined_at": ISODate("...")
    }
  ],
  "status": "active",
  "settings": {
    "allow_multiclass": true,
    "starting_level": 1,
    "ability_score_method": "standard_array",
    "allowed_sources": ["phb", "xgte"]
  },
  "created_at": ISODate("..."),
  "updated_at": ISODate("...")
}
```

### 3. Game Data Collections

#### Spells Collection
```javascript
{
  "_id": "fireball",
  "name": "Fireball",
  "level": 3,
  "school": "evocation",
  "casting_time": "1 action",
  "range": "150 feet",
  "components": ["V", "S", "M"],
  "material": "A tiny ball of bat guano and sulfur",
  "duration": "Instantaneous",
  "description": "A bright streak flashes from your pointing finger...",
  "damage": {
    "dice": "8d6",
    "type": "fire"
  },
  "save": {
    "ability": "dex",
    "effect": "half damage"
  },
  "area": {
    "type": "sphere",
    "size": 20
  },
  "classes": ["sorcerer", "wizard"],
  "source": "phb"
}
```

#### Monsters Collection
```javascript
{
  "_id": "goblin",
  "name": "Goblin",
  "size": "small",
  "type": "humanoid",
  "subtype": "goblinoid",
  "alignment": "neutral evil",
  "ac": 15,
  "armor_desc": "leather armor, shield",
  "hp": 7,
  "hit_dice": "2d6",
  "speed": {
    "walk": 30
  },
  "attributes": {
    "str": 8,
    "dex": 14,
    "con": 10,
    "int": 10,
    "wis": 8,
    "cha": 8
  },
  "skills": {
    "stealth": 6
  },
  "senses": {
    "darkvision": 60,
    "passive_perception": 9
  },
  "languages": ["Common", "Goblin"],
  "challenge_rating": 0.25,
  "xp": 50,
  "special_abilities": [
    {
      "name": "Nimble Escape",
      "desc": "The goblin can take the Disengage or Hide action as a bonus action on each of its turns."
    }
  ],
  "actions": [
    {
      "name": "Scimitar",
      "desc": "Melee Weapon Attack: +4 to hit, reach 5 ft., one target. Hit: 5 (1d6 + 2) slashing damage."
    },
    {
      "name": "Shortbow",
      "desc": "Ranged Weapon Attack: +4 to hit, range 80/320 ft., one target. Hit: 5 (1d6 + 2) piercing damage."
    }
  ]
}
```

### 4. User Preferences Collection

```javascript
{
  "_id": "discord_user_id",
  "display_name": "PlayerName",
  "active_character_id": ObjectId("..."),
  "preferences": {
    "dice_notation": "standard", // or "verbose"
    "auto_roll_damage": false,
    "whisper_rolls": false
  },
  "statistics": {
    "total_rolls": 1247,
    "nat_20s": 67,
    "nat_1s": 58,
    "combats_participated": 23
  },
  "created_at": ISODate("..."),
  "last_active": ISODate("...")
}
```

## Redis Schema (Session Data)

### Combat Session
```
KEY: combat:{session_id}
TYPE: JSON
TTL: 4 hours
VALUE: {
  "id": "uuid",
  "channel_id": "discord_channel_id",
  "message_id": "discord_message_id",
  "state": "active",
  "participants": {
    "participant_id": {
      "character_id": "mongodb_object_id",
      "user_id": "discord_user_id",
      "name": "Thorin",
      "initiative": 15,
      "current_hp": 44,
      "max_hp": 44,
      "ac": 18,
      "position": {"x": 5, "y": 10},
      "conditions": [],
      "death_saves": {"successes": 0, "failures": 0}
    }
  },
  "initiative_order": ["participant_id1", "participant_id2"],
  "current_turn": 0,
  "round": 1,
  "map": {
    "width": 20,
    "height": 20,
    "grid_size": 5
  },
  "started_at": "2024-01-15T10:00:00Z"
}
```

### Character Sheet Cache
```
KEY: character:{character_id}
TYPE: JSON
TTL: 5 minutes
VALUE: {complete character document from MongoDB}
```

### Active Encounters by Channel
```
KEY: encounter:channel:{channel_id}
TYPE: String
TTL: 4 hours
VALUE: {session_id}
```

## Indexes

### Characters Collection
```javascript
db.characters.createIndex({ "user_id": 1 })
db.characters.createIndex({ "name": 1 })
db.characters.createIndex({ "user_id": 1, "status": 1 })
```

### Campaigns Collection
```javascript
db.campaigns.createIndex({ "dm_user_id": 1 })
db.campaigns.createIndex({ "players.user_id": 1 })
db.campaigns.createIndex({ "channel_id": 1 })
```

### Spells Collection
```javascript
db.spells.createIndex({ "level": 1 })
db.spells.createIndex({ "classes": 1 })
db.spells.createIndex({ "name": "text" })
```

### Monsters Collection
```javascript
db.monsters.createIndex({ "challenge_rating": 1 })
db.monsters.createIndex({ "type": 1 })
db.monsters.createIndex({ "name": "text" })
```

## Migration Strategy

When schema changes are needed:

1. **Backward Compatible Changes**: Add new fields with defaults
2. **Breaking Changes**: Use versioning in document (`schema_version` field)
3. **Data Migrations**: Go scripts in `/migrations` folder

Example migration script:
```go
// migrations/001_add_death_saves.go
func AddDeathSaves(ctx context.Context, db *mongo.Database) error {
    collection := db.Collection("characters")
    
    filter := bson.M{
        "death_saves": bson.M{"$exists": false},
    }
    
    update := bson.M{
        "$set": bson.M{
            "death_saves": bson.M{
                "successes": 0,
                "failures": 0,
            },
        },
    }
    
    _, err := collection.UpdateMany(ctx, filter, update)
    return err
}
```