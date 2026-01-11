package schema

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// SchemaType represents the type of schema validation
type SchemaType string

const (
	SchemaTypeJSON    SchemaType = "json"
	SchemaTypeProtobuf SchemaType = "protobuf"
)

// SchemaInfo contains metadata about a registered schema
type SchemaInfo struct {
	ID          string            `json:"id"`
	Version     string            `json:"version"`
	Type        SchemaType        `json:"type"`
	Description string            `json:"description"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// SchemaDefinition represents a complete schema with its definition
type SchemaDefinition struct {
	Info      SchemaInfo    `json:"info"`
	Schema    json.RawMessage `json:"schema"`
	Validator Validator     `json:"-"`
}

// Registry provides centralized schema management with versioning support
type Registry struct {
	mu        sync.RWMutex
	schemas   map[string]map[string]*SchemaDefinition // schemaID -> version -> definition
	logger    *zap.Logger
}

// NewRegistry creates a new schema registry instance
func NewRegistry(logger *zap.Logger) *Registry {
	return &Registry{
		schemas: make(map[string]map[string]*SchemaDefinition),
		logger:  logger,
	}
}

// Register adds a new schema to the registry with versioning support
func (r *Registry) Register(schemaID string, version string, schemaType SchemaType, schemaData json.RawMessage, description string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Validate version format
	if err := validateVersion(version); err != nil {
		return fmt.Errorf("invalid version format: %w", err)
	}

	// Create the schema info
	info := SchemaInfo{
		ID:          schemaID,
		Version:     version,
		Type:        schemaType,
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Metadata:    make(map[string]string),
	}

	// Initialize schema map for this ID if needed
	if _, exists := r.schemas[schemaID]; !exists {
		r.schemas[schemaID] = make(map[string]*SchemaDefinition)
	}

	// Check if version already exists
	if _, exists := r.schemas[schemaID][version]; exists {
		return fmt.Errorf("schema %s version %s already exists", schemaID, version)
	}

	// Create validator based on schema type
	var validator Validator
	var err error

	switch schemaType {
	case SchemaTypeJSON:
		validator, err = NewJSONValidator(string(schemaData))
		if err != nil {
			return fmt.Errorf("failed to create JSON validator: %w", err)
		}
	case SchemaTypeProtobuf:
		validator, err = NewProtobufValidator(schemaData)
		if err != nil {
			return fmt.Errorf("failed to create protobuf validator: %w", err)
		}
	default:
		return fmt.Errorf("unsupported schema type: %s", schemaType)
	}

	// Store the schema definition
	definition := &SchemaDefinition{
		Info:      info,
		Schema:    schemaData,
		Validator: validator,
	}

	r.schemas[schemaID][version] = definition

	r.logger.Info("Schema registered",
		zap.String("schema_id", schemaID),
		zap.String("version", version),
		zap.String("type", string(schemaType)))

	return nil
}

// Get retrieves a schema by ID and version
func (r *Registry) Get(schemaID string, version string) (*SchemaDefinition, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	versions, exists := r.schemas[schemaID]
	if !exists {
		return nil, fmt.Errorf("schema not found: %s", schemaID)
	}

	definition, exists := versions[version]
	if !exists {
		return nil, fmt.Errorf("schema version not found: %s:%s", schemaID, version)
	}

	return definition, nil
}

// GetLatest retrieves the latest version of a schema
func (r *Registry) GetLatest(schemaID string) (*SchemaDefinition, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	versions, exists := r.schemas[schemaID]
	if !exists {
		return nil, fmt.Errorf("schema not found: %s", schemaID)
	}

	if len(versions) == 0 {
		return nil, fmt.Errorf("no versions available for schema: %s", schemaID)
	}

	// Find the latest version based on semantic versioning
	var latest *SchemaDefinition
	for _, def := range versions {
		if latest == nil || compareVersions(def.Info.Version, latest.Info.Version) > 0 {
			latest = def
		}
	}

	return latest, nil
}

// GetVersions returns all available versions for a schema
func (r *Registry) GetVersions(schemaID string) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	versions, exists := r.schemas[schemaID]
	if !exists {
		return nil, fmt.Errorf("schema not found: %s", schemaID)
	}

	versionList := make([]string, 0, len(versions))
	for v := range versions {
		versionList = append(versionList, v)
	}

	return versionList, nil
}

// GetValidator returns a validator for a specific schema version
func (r *Registry) GetValidator(schemaID string, version string) (Validator, error) {
	definition, err := r.Get(schemaID, version)
	if err != nil {
		return nil, err
	}

	return definition.Validator, nil
}

// GetLatestValidator returns a validator for the latest schema version
func (r *Registry) GetLatestValidator(schemaID string) (Validator, error) {
	definition, err := r.GetLatest(schemaID)
	if err != nil {
		return nil, err
	}

	return definition.Validator, nil
}

// Migrate migrates data from one schema version to another
func (r *Registry) Migrate(data json.RawMessage, fromVersion, toVersion string) (json.RawMessage, error) {
	// For now, we perform a simple pass-through migration
	// In a real implementation, you would define migration transformations
	return data, nil
}

// ListSchemas returns all registered schema IDs
func (r *Registry) ListSchemas() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	schemaIDs := make([]string, 0, len(r.schemas))
	for id := range r.schemas {
		schemaIDs = append(schemaIDs, id)
	}

	return schemaIDs
}

// GetSchemaInfo returns metadata for a specific schema version
func (r *Registry) GetSchemaInfo(schemaID string, version string) (*SchemaInfo, error) {
	definition, err := r.Get(schemaID, version)
	if err != nil {
		return nil, err
	}

	return &definition.Info, nil
}

// Delete removes a schema version from the registry
func (r *Registry) Delete(schemaID string, version string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	versions, exists := r.schemas[schemaID]
	if !exists {
		return fmt.Errorf("schema not found: %s", schemaID)
	}

	if _, exists := versions[version]; !exists {
		return fmt.Errorf("schema version not found: %s:%s", schemaID, version)
	}

	delete(versions, version)

	// Clean up if no versions left
	if len(versions) == 0 {
		delete(r.schemas, schemaID)
	}

	r.logger.Info("Schema deleted",
		zap.String("schema_id", schemaID),
		zap.String("version", version))

	return nil
}

// Export returns a JSON representation of all registered schemas
func (r *Registry) Export() (map[string]map[string]SchemaInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]map[string]SchemaInfo)
	for schemaID, versions := range r.schemas {
		result[schemaID] = make(map[string]SchemaInfo)
		for version, def := range versions {
			result[schemaID][version] = def.Info
		}
	}

	return result, nil
}
