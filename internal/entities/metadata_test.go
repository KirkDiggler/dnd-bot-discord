package entities

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetadata_GetString(t *testing.T) {
	tests := []struct {
		name        string
		metadata    Metadata
		key         string
		want        string
		wantErr     bool
		errContains string
	}{
		{
			name:     "successful retrieval",
			metadata: Metadata{"sessionType": "dungeon"},
			key:      "sessionType",
			want:     "dungeon",
			wantErr:  false,
		},
		{
			name:        "key not found",
			metadata:    Metadata{"foo": "bar"},
			key:         "sessionType",
			wantErr:     true,
			errContains: "not found",
		},
		{
			name:        "wrong type",
			metadata:    Metadata{"sessionType": 123},
			key:         "sessionType",
			wantErr:     true,
			errContains: "is not a string",
		},
		{
			name:        "nil metadata",
			metadata:    nil,
			key:         "sessionType",
			wantErr:     true,
			errContains: "metadata is nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.metadata.GetString(tt.key)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestMetadata_GetStringOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		metadata     Metadata
		key          string
		defaultValue string
		want         string
	}{
		{
			name:         "returns value when exists",
			metadata:     Metadata{"sessionType": "dungeon"},
			key:          "sessionType",
			defaultValue: "combat",
			want:         "dungeon",
		},
		{
			name:         "returns default when not found",
			metadata:     Metadata{"foo": "bar"},
			key:          "sessionType",
			defaultValue: "combat",
			want:         "combat",
		},
		{
			name:         "returns default when nil",
			metadata:     nil,
			key:          "sessionType",
			defaultValue: "combat",
			want:         "combat",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.metadata.GetStringOrDefault(tt.key, tt.defaultValue)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMetadata_GetInt(t *testing.T) {
	tests := []struct {
		name        string
		metadata    Metadata
		key         string
		want        int
		wantErr     bool
		errContains string
	}{
		{
			name:     "int value",
			metadata: Metadata{"roomNumber": 5},
			key:      "roomNumber",
			want:     5,
			wantErr:  false,
		},
		{
			name:     "float64 value (JSON unmarshaling)",
			metadata: Metadata{"roomNumber": float64(5)},
			key:      "roomNumber",
			want:     5,
			wantErr:  false,
		},
		{
			name:     "int64 value",
			metadata: Metadata{"roomNumber": int64(5)},
			key:      "roomNumber",
			want:     5,
			wantErr:  false,
		},
		{
			name:     "string value parseable",
			metadata: Metadata{"roomNumber": "5"},
			key:      "roomNumber",
			want:     5,
			wantErr:  false,
		},
		{
			name:        "string value not parseable",
			metadata:    Metadata{"roomNumber": "five"},
			key:         "roomNumber",
			wantErr:     true,
			errContains: "invalid syntax",
		},
		{
			name:        "wrong type",
			metadata:    Metadata{"roomNumber": true},
			key:         "roomNumber",
			wantErr:     true,
			errContains: "cannot be converted to int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.metadata.GetInt(tt.key)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestMetadata_Generic(t *testing.T) {
	m := Metadata{
		"sessionType": "dungeon",
		"roomNumber":  5,
		"isActive":    true,
		"difficulty":  "hard",
	}

	// Test Get with string
	sessionType, err := Get[string](m, "sessionType")
	require.NoError(t, err)
	assert.Equal(t, "dungeon", sessionType)

	// Test Get with int
	roomNumber, err := Get[int](m, "roomNumber")
	require.NoError(t, err)
	assert.Equal(t, 5, roomNumber)

	// Test Get with bool
	isActive, err := Get[bool](m, "isActive")
	require.NoError(t, err)
	assert.True(t, isActive)

	// Test GetOrDefault
	difficulty := GetOrDefault(m, "difficulty", "medium")
	assert.Equal(t, "hard", difficulty)

	// Test GetOrDefault with missing key
	missing := GetOrDefault(m, "missing", "default")
	assert.Equal(t, "default", missing)

	// Test type mismatch
	_, err = Get[int](m, "sessionType")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "is not of type int")
}

func TestMetadata_Set(t *testing.T) {
	m := make(Metadata)

	// Set string
	m.Set("sessionType", "dungeon")
	assert.Equal(t, "dungeon", m["sessionType"])

	// Set int
	m.Set("roomNumber", 5)
	assert.Equal(t, 5, m["roomNumber"])

	// Set on nil metadata should not panic
	var nilMetadata Metadata
	assert.NotPanics(t, func() {
		nilMetadata.Set("key", "value")
	})
}

func TestMetadata_Has(t *testing.T) {
	m := Metadata{
		"exists": "value",
	}

	assert.True(t, m.Has("exists"))
	assert.False(t, m.Has("notExists"))

	// Test nil metadata
	var nilMetadata Metadata
	assert.False(t, nilMetadata.Has("anything"))
}
