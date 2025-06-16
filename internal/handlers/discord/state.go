package discord

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// CharacterCreationState holds the state during character creation
type CharacterCreationState struct {
	RaceKey       string         `json:"race_key"`
	ClassKey      string         `json:"class_key"`
	AbilityScores map[string]int `json:"ability_scores,omitempty"`
	Proficiencies []string       `json:"proficiencies,omitempty"`
	Equipment     []string       `json:"equipment,omitempty"`
}

// Encode encodes the state to a base64 string
func (s *CharacterCreationState) Encode() (string, error) {
	data, err := json.Marshal(s)
	if err != nil {
		return "", fmt.Errorf("failed to marshal state: %w", err)
	}
	return base64.URLEncoding.EncodeToString(data), nil
}

// DecodeCharacterCreationState decodes a base64 string to state
func DecodeCharacterCreationState(encoded string) (*CharacterCreationState, error) {
	data, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}
	
	var state CharacterCreationState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state: %w", err)
	}
	
	return &state, nil
}