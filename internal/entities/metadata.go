package entities

import (
	"fmt"
	"strconv"
)

// Metadata provides type-safe access to map[string]interface{} data
type Metadata map[string]interface{}

// GetString retrieves a string value from metadata
func (m Metadata) GetString(key string) (string, error) {
	if m == nil {
		return "", fmt.Errorf("metadata is nil")
	}

	value, exists := m[key]
	if !exists {
		return "", fmt.Errorf("key %q not found", key)
	}

	str, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("key %q is not a string (got %T)", key, value)
	}

	return str, nil
}

// GetStringOrDefault retrieves a string value or returns the default
func (m Metadata) GetStringOrDefault(key, defaultValue string) string {
	str, err := m.GetString(key)
	if err != nil {
		return defaultValue
	}
	return str
}

// GetInt retrieves an int value from metadata
func (m Metadata) GetInt(key string) (int, error) {
	if m == nil {
		return 0, fmt.Errorf("metadata is nil")
	}

	value, exists := m[key]
	if !exists {
		return 0, fmt.Errorf("key %q not found", key)
	}

	// Handle various numeric types
	switch v := value.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	case string:
		// Try to parse string as int
		return strconv.Atoi(v)
	default:
		return 0, fmt.Errorf("key %q cannot be converted to int (got %T)", key, value)
	}
}

// GetIntOrDefault retrieves an int value or returns the default
func (m Metadata) GetIntOrDefault(key string, defaultValue int) int {
	val, err := m.GetInt(key)
	if err != nil {
		return defaultValue
	}
	return val
}

// GetBool retrieves a bool value from metadata
func (m Metadata) GetBool(key string) (bool, error) {
	if m == nil {
		return false, fmt.Errorf("metadata is nil")
	}

	value, exists := m[key]
	if !exists {
		return false, fmt.Errorf("key %q not found", key)
	}

	b, ok := value.(bool)
	if !ok {
		return false, fmt.Errorf("key %q is not a bool (got %T)", key, value)
	}

	return b, nil
}

// GetBoolOrDefault retrieves a bool value or returns the default
func (m Metadata) GetBoolOrDefault(key string, defaultValue bool) bool {
	val, err := m.GetBool(key)
	if err != nil {
		return defaultValue
	}
	return val
}

// Set sets a value in the metadata
func (m Metadata) Set(key string, value interface{}) {
	if m == nil {
		return
	}
	m[key] = value
}

// Has checks if a key exists in metadata
func (m Metadata) Has(key string) bool {
	if m == nil {
		return false
	}
	_, exists := m[key]
	return exists
}

// Generic Get method using type parameters (requires Go 1.18+)
func Get[T any](m Metadata, key string) (T, error) {
	var zero T

	if m == nil {
		return zero, fmt.Errorf("metadata is nil")
	}

	value, exists := m[key]
	if !exists {
		return zero, fmt.Errorf("key %q not found", key)
	}

	typed, ok := value.(T)
	if !ok {
		return zero, fmt.Errorf("key %q is not of type %T (got %T)", key, zero, value)
	}

	return typed, nil
}

// GetOrDefault is a generic method that returns a default value on error
func GetOrDefault[T any](m Metadata, key string, defaultValue T) T {
	val, err := Get[T](m, key)
	if err != nil {
		return defaultValue
	}
	return val
}
