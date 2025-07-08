package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCustomID_Encode(t *testing.T) {
	tests := []struct {
		name     string
		customID *CustomID
		expected string
		wantErr  bool
	}{
		{
			name: "simple domain and action",
			customID: &CustomID{
				Domain: "character",
				Action: "list",
			},
			expected: "character:list",
		},
		{
			name: "with target",
			customID: &CustomID{
				Domain: "character",
				Action: "show",
				Target: "char_123",
			},
			expected: "character:show:char_123",
		},
		{
			name: "with args",
			customID: &CustomID{
				Domain: "combat",
				Action: "attack",
				Target: "char_123",
				Args:   []string{"goblin_1", "weapon_sword"},
			},
			expected: "combat:attack:char_123:goblin_1:weapon_sword",
		},
		{
			name: "with data",
			customID: &CustomID{
				Domain: "character",
				Action: "equip",
				Target: "char_123",
				Data: map[string]interface{}{
					"slot": "mainhand",
					"item": "sword",
				},
			},
			expected: "character:equip:char_123:data:",
		},
		{
			name: "exceeds max length",
			customID: &CustomID{
				Domain: "verylongdomainname",
				Action: "verylongactionname",
				Target: "verylongtargetidentifier",
				Args: []string{
					"verylongargument1",
					"verylongargument2",
					"verylongargument3",
					"verylongargument4",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.customID.Encode()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			// For data test, just check prefix since JSON encoding may vary
			if len(tt.customID.Data) > 0 {
				assert.Contains(t, result, tt.expected)
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestParseCustomID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *CustomID
		wantErr  bool
	}{
		{
			name:  "simple domain and action",
			input: "character:list",
			expected: &CustomID{
				Domain: "character",
				Action: "list",
				Args:   []string{},
				Data:   map[string]interface{}{},
			},
		},
		{
			name:  "with target",
			input: "character:show:char_123",
			expected: &CustomID{
				Domain: "character",
				Action: "show",
				Target: "char_123",
				Args:   []string{},
				Data:   map[string]interface{}{},
			},
		},
		{
			name:  "with args",
			input: "combat:attack:char_123:goblin_1:weapon_sword",
			expected: &CustomID{
				Domain: "combat",
				Action: "attack",
				Target: "char_123",
				Args:   []string{"goblin_1", "weapon_sword"},
				Data:   map[string]interface{}{},
			},
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "missing action",
			input:   "character",
			wantErr: true,
		},
		{
			name:  "with colons in args",
			input: "modal:submit:form_123:field:value",
			expected: &CustomID{
				Domain: "modal",
				Action: "submit",
				Target: "form_123",
				Args:   []string{"field", "value"},
				Data:   map[string]interface{}{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseCustomID(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCustomID_RoundTrip(t *testing.T) {
	// Test that encode/decode is symmetric
	original := &CustomID{
		Domain: "character",
		Action: "equip",
		Target: "char_123",
		Args:   []string{"slot_mainhand"},
		Data: map[string]interface{}{
			"item":     "longsword",
			"quantity": float64(1), // JSON numbers decode as float64
		},
	}

	encoded, err := original.Encode()
	require.NoError(t, err)

	decoded, err := ParseCustomID(encoded)
	require.NoError(t, err)

	assert.Equal(t, original.Domain, decoded.Domain)
	assert.Equal(t, original.Action, decoded.Action)
	assert.Equal(t, original.Target, decoded.Target)
	assert.Equal(t, original.Args, decoded.Args)
	assert.Equal(t, original.Data, decoded.Data)
}

func TestCustomIDBuilder(t *testing.T) {
	builder := NewCustomIDBuilder("character")

	t.Run("button", func(t *testing.T) {
		result := builder.Button("show", "char_123")
		assert.Equal(t, "character:show:char_123", result)
	})

	t.Run("button with args", func(t *testing.T) {
		result := builder.Button("equip", "char_123", "sword", "mainhand")
		assert.Equal(t, "character:equip:char_123:sword:mainhand", result)
	})

	t.Run("select with data", func(t *testing.T) {
		result := builder.Select("filter", map[string]interface{}{
			"class": "fighter",
			"level": 5,
		})
		assert.Contains(t, result, "character:filter:")
		assert.Contains(t, result, "data:")

		// Verify it can be parsed
		parsed, err := ParseCustomID(result)
		require.NoError(t, err)
		assert.Equal(t, "fighter", parsed.Data["class"])
		assert.Equal(t, float64(5), parsed.Data["level"])
	})

	t.Run("modal", func(t *testing.T) {
		result := builder.Modal("create", "template_1")
		assert.Equal(t, "character:create:template_1", result)
	})
}

func TestCustomIDMatcher(t *testing.T) {
	matcher := NewCustomIDMatcher("character", "show")

	t.Run("exact match", func(t *testing.T) {
		assert.True(t, matcher.Matches("character:show:char_123"))
	})

	t.Run("different action", func(t *testing.T) {
		assert.False(t, matcher.Matches("character:list"))
	})

	t.Run("different domain", func(t *testing.T) {
		assert.False(t, matcher.Matches("combat:show:char_123"))
	})

	t.Run("wildcard action", func(t *testing.T) {
		wildcardMatcher := NewCustomIDMatcher("character", "*")
		assert.True(t, wildcardMatcher.Matches("character:show:char_123"))
		assert.True(t, wildcardMatcher.Matches("character:list"))
		assert.False(t, wildcardMatcher.Matches("combat:attack"))
	})

	t.Run("extract", func(t *testing.T) {
		customID, ok := matcher.Extract("character:show:char_123:extra")
		require.True(t, ok)
		assert.Equal(t, "character", customID.Domain)
		assert.Equal(t, "show", customID.Action)
		assert.Equal(t, "char_123", customID.Target)
		assert.Equal(t, []string{"extra"}, customID.Args)
	})

	t.Run("extract no match", func(t *testing.T) {
		customID, ok := matcher.Extract("combat:attack")
		assert.False(t, ok)
		assert.Nil(t, customID)
	})
}

func TestCustomID_FluentAPI(t *testing.T) {
	customID := NewCustomID("character", "equip").
		WithTarget("char_123").
		WithArgs("sword", "mainhand").
		WithData("enchanted", true).
		WithData("damage", "1d8")

	encoded := customID.MustEncode()

	decoded, err := ParseCustomID(encoded)
	require.NoError(t, err)

	assert.Equal(t, "character", decoded.Domain)
	assert.Equal(t, "equip", decoded.Action)
	assert.Equal(t, "char_123", decoded.Target)
	assert.Equal(t, []string{"sword", "mainhand"}, decoded.Args)
	assert.Equal(t, true, decoded.Data["enchanted"])
	assert.Equal(t, "1d8", decoded.Data["damage"])
}
