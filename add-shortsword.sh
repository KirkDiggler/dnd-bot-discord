#!/bin/bash

# Check if character ID is provided
if [ -z "$1" ]; then
    echo "Usage: ./add-shortsword.sh CHARACTER_ID"
    echo "To find your character ID, run: redis-cli KEYS 'character:*'"
    exit 1
fi

CHAR_ID=$1
CHAR_KEY="character:$CHAR_ID"

# Get current character data
CHAR_DATA=$(redis-cli GET "$CHAR_KEY")

if [ -z "$CHAR_DATA" ]; then
    echo "Character not found with ID: $CHAR_ID"
    exit 1
fi

# Create a shortsword JSON object
SHORTSWORD='{
  "Type": "weapon",
  "Data": {
    "base": {
      "key": "shortsword",
      "name": "Shortsword",
      "equipment_category": "weapon",
      "weight": 2,
      "cost": {
        "quantity": 10,
        "unit": "gp"
      }
    },
    "damage": {
      "damage_dice": "1d6",
      "damage_type": "piercing",
      "dice_count": 1,
      "dice_size": 6
    },
    "range": 5,
    "weapon_category": "Martial",
    "weapon_range": "Melee",
    "category_range": "martial-melee",
    "properties": [
      {
        "key": "finesse",
        "name": "Finesse",
        "url": "/api/weapon-properties/finesse"
      },
      {
        "key": "light",
        "name": "Light",
        "url": "/api/weapon-properties/light"
      },
      {
        "key": "monk",
        "name": "Monk",
        "url": "/api/weapon-properties/monk"
      }
    ],
    "two_handed_damage": null
  }
}'

# Use jq to add the shortsword to inventory
# This assumes you have jq installed. If not, you'll need to install it: sudo apt-get install jq
if ! command -v jq &> /dev/null; then
    echo "jq is required but not installed. Install it with: sudo apt-get install jq"
    exit 1
fi

# Add shortsword to weapon inventory
UPDATED_DATA=$(echo "$CHAR_DATA" | jq --argjson sword "$SHORTSWORD" '.Inventory.weapon += [$sword]')

# Save back to Redis
echo "$UPDATED_DATA" | redis-cli -x SET "$CHAR_KEY"

echo "Shortsword added to character inventory!"
echo "You can now equip it with: /dnd character equip"