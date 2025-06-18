// uuid simple generator that allows mocking
package uuid

import (
	"github.com/google/uuid"
)

// Generator is an interface for generating UUIDs
type Generator interface {
	New() string
}

// GoogleUUIDGenerator implements the Generator interface using Google's UUID package
type GoogleUUIDGenerator struct{}

// New generates a new UUID string
func (g *GoogleUUIDGenerator) New() string {
	return uuid.New().String()
}

// NewGoogleUUIDGenerator creates a new GoogleUUIDGenerator
func NewGoogleUUIDGenerator() *GoogleUUIDGenerator {
	return &GoogleUUIDGenerator{}
}
