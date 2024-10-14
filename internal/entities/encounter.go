package entities

import "encoding/json"

type Encounter struct {
	ID        string
	MessageID string
	Players   []string
}

func (e *Encounter) MarshallJSON() ([]byte, error) {
	return json.Marshal(e)
}
