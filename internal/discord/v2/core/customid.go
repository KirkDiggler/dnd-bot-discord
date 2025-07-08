package core

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

const (
	// CustomIDSeparator is the character used to separate parts
	CustomIDSeparator = ":"

	// MaxCustomIDLength is Discord's limit for custom IDs
	MaxCustomIDLength = 100
)

// CustomID represents a parsed custom ID with type-safe access
type CustomID struct {
	// Domain is the top-level category (e.g., "character", "combat")
	Domain string

	// Action is the specific action (e.g., "list", "show", "attack")
	Action string

	// Target is the primary target of the action (e.g., character ID)
	Target string

	// Args are additional arguments
	Args []string

	// Data is optional JSON-encoded data for complex payloads
	Data map[string]interface{}
}

// NewCustomID creates a new CustomID
func NewCustomID(domain, action string) *CustomID {
	return &CustomID{
		Domain: domain,
		Action: action,
		Args:   make([]string, 0),
		Data:   make(map[string]interface{}),
	}
}

// WithTarget sets the target
func (c *CustomID) WithTarget(target string) *CustomID {
	c.Target = target
	return c
}

// WithArgs adds arguments
func (c *CustomID) WithArgs(args ...string) *CustomID {
	c.Args = append(c.Args, args...)
	return c
}

// WithData adds data to be JSON encoded
func (c *CustomID) WithData(key string, value interface{}) *CustomID {
	c.Data[key] = value
	return c
}

// Encode converts the CustomID to a string
func (c *CustomID) Encode() (string, error) {
	parts := []string{c.Domain, c.Action}

	if c.Target != "" {
		parts = append(parts, c.Target)
	}

	// Add args
	parts = append(parts, c.Args...)

	// If we have data, encode it at the end
	if len(c.Data) > 0 {
		jsonData, err := json.Marshal(c.Data)
		if err != nil {
			return "", fmt.Errorf("failed to marshal data: %w", err)
		}

		// Use base64 to ensure no separator conflicts
		encoded := base64.RawURLEncoding.EncodeToString(jsonData)
		parts = append(parts, "data:"+encoded)
	}

	result := strings.Join(parts, CustomIDSeparator)

	// Check length limit
	if len(result) > MaxCustomIDLength {
		return "", fmt.Errorf("custom ID exceeds maximum length of %d characters", MaxCustomIDLength)
	}

	return result, nil
}

// MustEncode is like Encode but panics on error
func (c *CustomID) MustEncode() string {
	result, err := c.Encode()
	if err != nil {
		panic(err)
	}
	return result
}

// ParseCustomID parses a custom ID string
func ParseCustomID(customID string) (*CustomID, error) {
	if customID == "" {
		return nil, fmt.Errorf("empty custom ID")
	}

	parts := strings.Split(customID, CustomIDSeparator)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid custom ID format: expected at least domain:action")
	}

	result := &CustomID{
		Domain: parts[0],
		Action: parts[1],
		Args:   make([]string, 0),
		Data:   make(map[string]interface{}),
	}

	// Process remaining parts
	if len(parts) > 2 {
		// Check if we have data (will be split as "data" and the encoded part)
		hasData := false
		dataIndex := -1

		for i := 2; i < len(parts)-1; i++ {
			if parts[i] == "data" {
				// Next part should be the base64 encoded data
				encoded := parts[i+1]
				decoded, err := base64.RawURLEncoding.DecodeString(encoded)
				if err == nil {
					// Successfully decoded, this is our data
					if err := json.Unmarshal(decoded, &result.Data); err == nil {
						hasData = true
						dataIndex = i
						break
					}
				}
			}
		}

		// Now process target and args
		result.Target = parts[2]

		// Collect args, skipping data parts
		for i := 3; i < len(parts); i++ {
			if hasData && (i == dataIndex || i == dataIndex+1) {
				// Skip the "data" marker and encoded data
				continue
			}
			result.Args = append(result.Args, parts[i])
		}
	}

	return result, nil
}

// CustomIDBuilder provides a fluent interface for building custom IDs
type CustomIDBuilder struct {
	domain string
}

// NewCustomIDBuilder creates a new builder for a domain
func NewCustomIDBuilder(domain string) *CustomIDBuilder {
	return &CustomIDBuilder{domain: domain}
}

// Build creates a CustomID for an action
func (b *CustomIDBuilder) Build(action string) *CustomID {
	return NewCustomID(b.domain, action)
}

// Button creates a button custom ID
func (b *CustomIDBuilder) Button(action, target string, args ...string) string {
	return NewCustomID(b.domain, action).
		WithTarget(target).
		WithArgs(args...).
		MustEncode()
}

// Select creates a select menu custom ID
func (b *CustomIDBuilder) Select(action string, data map[string]interface{}) string {
	id := NewCustomID(b.domain, action)
	for k, v := range data {
		id.WithData(k, v)
	}
	return id.MustEncode()
}

// Modal creates a modal custom ID
func (b *CustomIDBuilder) Modal(action, target string) string {
	return NewCustomID(b.domain, action).
		WithTarget(target).
		MustEncode()
}

// CustomIDMatcher helps match custom IDs in handlers
type CustomIDMatcher struct {
	domain string
	action string
}

// NewCustomIDMatcher creates a matcher for a specific domain/action
func NewCustomIDMatcher(domain, action string) *CustomIDMatcher {
	return &CustomIDMatcher{
		domain: domain,
		action: action,
	}
}

// Matches checks if a custom ID matches this pattern
func (m *CustomIDMatcher) Matches(customID string) bool {
	parsed, err := ParseCustomID(customID)
	if err != nil {
		return false
	}

	return parsed.Domain == m.domain && (m.action == "*" || parsed.Action == m.action)
}

// Extract parses and returns the custom ID if it matches
func (m *CustomIDMatcher) Extract(customID string) (*CustomID, bool) {
	if !m.Matches(customID) {
		return nil, false
	}

	parsed, err := ParseCustomID(customID)
	if err != nil {
		return nil, false
	}

	return parsed, true
}
