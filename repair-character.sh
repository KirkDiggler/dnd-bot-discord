#!/bin/bash

CHAR_ID="char_1751254556976130208"
CHAR_KEY="character:$CHAR_ID"

# Get the current broken data
CHAR_DATA=$(redis-cli GET "$CHAR_KEY")

# Parse and fix the JSON structure - remove duplicate keys and ensure proper format
FIXED_DATA=$(echo "$CHAR_DATA" | jq '{
  id: .id,
  owner_id: .owner_id,
  realm_id: .realm_id,
  name: .name,
  speed: .speed,
  race: .race,
  class: .class,
  background: .background,
  attributes: .attributes,
  ability_rolls: .ability_rolls,
  ability_assignments: .ability_assignments,
  proficiencies: .proficiencies,
  hit_die: .hit_die,
  ac: .ac,
  max_hit_points: .max_hit_points,
  current_hit_points: .current_hit_points,
  level: .level,
  experience: .experience,
  status: "active",
  features: .features,
  inventory: (.Inventory // .inventory),
  equipped_slots: .equipped_slots,
  resources: .resources,
  created_at: .created_at,
  updated_at: .updated_at
}')

# Save the fixed data back
echo "$FIXED_DATA" | redis-cli -x SET "$CHAR_KEY"

# Re-add to indexes
redis-cli SADD "character_owner:99945162414772224" "$CHAR_ID"
redis-cli SADD "character_realm:99945541361729536" "$CHAR_ID"
redis-cli SADD "character_owner_realm:99945162414772224:99945541361729536" "$CHAR_ID"

echo "Character repaired!"