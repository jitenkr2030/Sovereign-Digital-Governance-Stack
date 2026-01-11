package schema

import (
	"encoding/json"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestRegistry_Register(t *testing.T) {
	logger := zap.NewNop()
	registry := NewRegistry(logger)

	// Test registering a JSON schema
	schemaJSON := `{
		"type": "object",
		"properties": {
			"id": {"type": "string"},
			"value": {"type": "number"}
		},
		"required": ["id"]
	}`

	err := registry.Register("test-schema", "1.0.0", SchemaTypeJSON, []byte(schemaJSON), "Test schema")
	if err != nil {
		t.Fatalf("Failed to register schema: %v", err)
	}

	// Verify the schema was registered
	def, err := registry.Get("test-schema", "1.0.0")
	if err != nil {
		t.Fatalf("Failed to get schema: %v", err)
	}

	if def.Info.ID != "test-schema" {
		t.Errorf("Expected schema ID 'test-schema', got '%s'", def.Info.ID)
	}

	if def.Info.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", def.Info.Version)
	}
}

func TestRegistry_GetLatest(t *testing.T) {
	logger := zap.NewNop()
	registry := NewRegistry(logger)

	// Register multiple versions
	schemaJSON := `{"type": "object", "properties": {"id": {"type": "string"}}}`

	registry.Register("test-schema", "1.0.0", SchemaTypeJSON, []byte(schemaJSON), "Version 1")
	registry.Register("test-schema", "1.1.0", SchemaTypeJSON, []byte(schemaJSON), "Version 2")
	registry.Register("test-schema", "2.0.0", SchemaTypeJSON, []byte(schemaJSON), "Version 3")

	// Get latest version
	latest, err := registry.GetLatest("test-schema")
	if err != nil {
		t.Fatalf("Failed to get latest schema: %v", err)
	}

	if latest.Info.Version != "2.0.0" {
		t.Errorf("Expected latest version '2.0.0', got '%s'", latest.Info.Version)
	}
}

func TestRegistry_ListSchemas(t *testing.T) {
	logger := zap.NewNop()
	registry := NewRegistry(logger)

	schemaJSON := `{"type": "object"}`

	registry.Register("schema-1", "1.0.0", SchemaTypeJSON, []byte(schemaJSON), "")
	registry.Register("schema-2", "1.0.0", SchemaTypeJSON, []byte(schemaJSON), "")

	schemas := registry.ListSchemas()
	if len(schemas) != 2 {
		t.Errorf("Expected 2 schemas, got %d", len(schemas))
	}
}

func TestRegistry_Delete(t *testing.T) {
	logger := zap.NewNop()
	registry := NewRegistry(logger)

	schemaJSON := `{"type": "object"}`

	registry.Register("test-schema", "1.0.0", SchemaTypeJSON, []byte(schemaJSON), "")

	// Delete the schema
	err := registry.Delete("test-schema", "1.0.0")
	if err != nil {
		t.Fatalf("Failed to delete schema: %v", err)
	}

	// Verify it's gone
	_, err = registry.Get("test-schema", "1.0.0")
	if err == nil {
		t.Error("Expected error when getting deleted schema")
	}
}

func TestRegistry_Export(t *testing.T) {
	logger := zap.NewNop()
	registry := NewRegistry(logger)

	schemaJSON := `{"type": "object"}`

	registry.Register("test-schema", "1.0.0", SchemaTypeJSON, []byte(schemaJSON), "Test description")

	export, err := registry.Export()
	if err != nil {
		t.Fatalf("Failed to export registry: %v", err)
	}

	if _, exists := export["test-schema"]; !exists {
		t.Error("Expected test-schema in export")
	}

	if _, exists := export["test-schema"]["1.0.0"]; !exists {
		t.Error("Expected version 1.0.0 in export")
	}
}

func TestValidateVersion(t *testing.T) {
	tests := []struct {
		version string
		valid   bool
	}{
		{"1.0.0", true},
		{"1.0", false},
		{"v1.0.0", false},
		{"1.0.0.0", false},
		{"1", false},
	}

	for _, tt := range tests {
		err := validateVersion(tt.version)
		if tt.valid && err != nil {
			t.Errorf("Expected version %s to be valid, got error: %v", tt.version, err)
		}
		if !tt.valid && err == nil {
			t.Errorf("Expected version %s to be invalid", tt.version)
		}
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		v1       string
		v2       string
		expected int
	}{
		{"1.0.0", "1.0.0", 0},
		{"2.0.0", "1.0.0", 1},
		{"1.0.0", "2.0.0", -1},
		{"1.1.0", "1.0.0", 1},
		{"1.0.1", "1.0.0", 1},
	}

	for _, tt := range tests {
		result := compareVersions(tt.v1, tt.v2)
		if result != tt.expected {
			t.Errorf("compareVersions(%s, %s) = %d, expected %d", tt.v1, tt.v2, result, tt.expected)
		}
	}
}
