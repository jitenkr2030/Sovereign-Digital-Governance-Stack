package adapters

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"neam-platform/integration-gateway/config"
	"neam-platform/integration-gateway/models"
	"neam-platform/integration-gateway/transformers"
)

// StateAdapter handles integration with state government systems
type StateAdapter struct {
	*BaseAdapter
	cfg *config.Config
}

// NewStateAdapter creates a new StateAdapter
func NewStateAdapter(cfg *config.Config) *StateAdapter {
	transformer := transformers.NewDataTransformer()
	return &StateAdapter{
		BaseAdapter: NewBaseAdapter(cfg, "state", "state-adapter", transformer),
		cfg:         cfg,
	}
}

// Connect establishes connection to state systems
func (a *StateAdapter) Connect(ctx context.Context) error {
	return a.HealthCheck(ctx)
}

// Disconnect closes the connection
func (a *StateAdapter) Disconnect(ctx context.Context) error {
	return nil
}

// Fetch retrieves data from state systems
func (a *StateAdapter) Fetch(ctx context.Context, request *models.AdapterRequest) (*models.AdapterResponse, error) {
	url := a.getEndpoint(request.Endpoint)

	// Add state code header if configured
	if a.config.Adapters.State.StateCode != "" {
		ctx = context.WithValue(ctx, "state_code", a.config.Adapters.State.StateCode)
	}

	resp, err := a.doRequest(ctx, request.Method, url, request.Payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	response, err := a.parseResponse(resp)
	if err != nil {
		return nil, err
	}

	// Transform data for state-specific formats
	if request.Transform {
		transformedData, err := a.transformer.TransformStateData(response.Body, request.DataType)
		if err != nil {
			return nil, fmt.Errorf("state data transformation failed: %w", err)
		}
		response.Body = transformedData
	}

	return response, nil
}

// Push sends data to state systems
func (a *StateAdapter) Push(ctx context.Context, message *models.IntegrationMessage) error {
	endpoint := a.getEndpointForMessage(message)

	request := &models.AdapterRequest{
		Method:  http.MethodPost,
		Endpoint: endpoint,
		Payload: message.Payload,
		DataType: message.DataType,
	}

	resp, err := a.Fetch(ctx, request)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("state push failed with status: %d", resp.StatusCode)
	}

	return nil
}

// HealthCheck performs health check on state connection
func (a *StateAdapter) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	healthURL := BuildRequestURL(a.config.Adapters.State.Endpoint, "/health", nil)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
	if err != nil {
		return err
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("state health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("state system unhealthy, status: %d", resp.StatusCode)
	}

	return nil
}

// getEndpointForMessage returns the appropriate endpoint for state messages
func (a *StateAdapter) getEndpointForMessage(message *models.IntegrationMessage) string {
	switch message.DataType {
	case "state_citizen":
		return "/api/v1/state/citizens"
	case "state_benefit":
		return "/api/v1/state/benefits"
	case "state_tax":
		return "/api/v1/state/tax"
	case "state_infrastructure":
		return "/api/v1/state/infrastructure"
	case "state_health":
		return "/api/v1/state/health"
	case "state_education":
		return "/api/v1/state/education"
	default:
		return "/api/v1/state/data"
	}
}

// SupportedDataTypes returns the data types supported by this adapter
func (a *StateAdapter) SupportedDataTypes() []string {
	return []string{
		"state_citizen",
		"state_benefit",
		"state_tax",
		"state_infrastructure",
		"state_health",
		"state_education",
		"state_resource",
	}
}

// StateAdapter implements IntegrationAdapter
var _ IntegrationAdapter = (*StateAdapter)(nil)

// DistrictAdapter handles integration with district-level systems
type DistrictAdapter struct {
	*BaseAdapter
	cfg *config.Config
}

// NewDistrictAdapter creates a new DistrictAdapter
func NewDistrictAdapter(cfg *config.Config) *DistrictAdapter {
	transformer := transformers.NewDataTransformer()
	return &DistrictAdapter{
		BaseAdapter: NewBaseAdapter(cfg, "district", "district-adapter", transformer),
		cfg:         cfg,
	}
}

// Connect establishes connection to district systems
func (a *DistrictAdapter) Connect(ctx context.Context) error {
	return a.HealthCheck(ctx)
}

// Disconnect closes the connection
func (a *DistrictAdapter) Disconnect(ctx context.Context) error {
	return nil
}

// Fetch retrieves data from district systems
func (a *DistrictAdapter) Fetch(ctx context.Context, request *models.AdapterRequest) (*models.AdapterResponse, error) {
	url := a.getEndpoint(request.Endpoint)

	resp, err := a.doRequest(ctx, request.Method, url, request.Payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	response, err := a.parseResponse(resp)
	if err != nil {
		return nil, err
	}

	// Transform data for district-specific formats
	if request.Transform {
		transformedData, err := a.transformer.TransformDistrictData(response.Body, request.DataType)
		if err != nil {
			return nil, fmt.Errorf("district data transformation failed: %w", err)
		}
		response.Body = transformedData
	}

	return response, nil
}

// Push sends data to district systems
func (a *DistrictAdapter) Push(ctx context.Context, message *models.IntegrationMessage) error {
	endpoint := a.getEndpointForMessage(message)

	request := &models.AdapterRequest{
		Method:  http.MethodPost,
		Endpoint: endpoint,
		Payload: message.Payload,
		DataType: message.DataType,
	}

	resp, err := a.Fetch(ctx, request)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("district push failed with status: %d", resp.StatusCode)
	}

	return nil
}

// HealthCheck performs health check on district connection
func (a *DistrictAdapter) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	healthURL := BuildRequestURL(a.config.Adapters.District.Endpoint, "/health", nil)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
	if err != nil {
		return err
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("district health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("district system unhealthy, status: %d", resp.StatusCode)
	}

	return nil
}

// getEndpointForMessage returns the appropriate endpoint for district messages
func (a *DistrictAdapter) getEndpointForMessage(message *models.IntegrationMessage) string {
	switch message.DataType {
	case "district_citizen":
		return "/api/v1/district/citizens"
	case "district_service":
		return "/api/v1/district/services"
	case "district_complaint":
		return "/api/v1/district/complaints"
	case "district_property":
		return "/api/v1/district/property"
	default:
		return "/api/v1/district/data"
	}
}

// SupportedDataTypes returns the data types supported by this adapter
func (a *DistrictAdapter) SupportedDataTypes() []string {
	return []string{
		"district_citizen",
		"district_service",
		"district_complaint",
		"district_property",
		"district_grievance",
		"district_utility",
	}
}

// DistrictAdapter implements IntegrationAdapter
var _ IntegrationAdapter = (*DistrictAdapter)(nil)

// LegacyAdapter handles integration with legacy systems
type LegacyAdapter struct {
	*BaseAdapter
	cfg        *config.Config
	transformer *transformers.DataTransformer
	ftpEnabled bool
}

// NewLegacyAdapter creates a new LegacyAdapter
func NewLegacyAdapter(cfg *config.Config, transformer *transformers.DataTransformer) *LegacyAdapter {
	return &LegacyAdapter{
		BaseAdapter: NewBaseAdapter(cfg, "legacy", "legacy-adapter", transformer),
		cfg:         cfg,
		transformer: transformer,
		ftpEnabled:  cfg.Adapters.Legacy.FTPServer != "",
	}
}

// Connect establishes connection to legacy systems
func (a *LegacyAdapter) Connect(ctx context.Context) error {
	// Create directories if they don't exist
	dirs := []string{
		a.config.Adapters.Legacy.InputDirectory,
		a.config.Adapters.Legacy.OutputDirectory,
		a.config.Adapters.Legacy.ArchiveDirectory,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// If FTP is enabled, connect to FTP server
	if a.ftpEnabled {
		// FTP connection would be established here
		// For now, just verify directory access
	}

	return nil
}

// Disconnect closes the connection
func (a *LegacyAdapter) Disconnect(ctx context.Context) error {
	// Close FTP connection if active
	return nil
}

// Fetch retrieves data from legacy systems
func (a *LegacyAdapter) Fetch(ctx context.Context, request *models.AdapterRequest) (*models.AdapterResponse, error) {
	// Handle file-based legacy data
	if request.SourceType == "file" {
		return a.fetchFromFile(ctx, request)
	}

	// Handle FTP-based legacy data
	if request.SourceType == "ftp" {
		return a.fetchFromFTP(ctx, request)
	}

	// Default to REST API
	url := a.getEndpoint(request.Endpoint)

	resp, err := a.doRequest(ctx, request.Method, url, request.Payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return a.parseResponse(resp)
}

// fetchFromFile retrieves data from local files
func (a *LegacyAdapter) fetchFromFile(ctx context.Context, request *models.AdapterRequest) (*models.AdapterResponse, error) {
	filePath := request.Endpoint

	// Read file
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read legacy file: %w", err)
	}

	// Parse based on format
	var parsedData []map[string]interface{}
	format := a.detectFormat(filePath)

	switch format {
	case "csv":
		parsedData, err = a.parseCSV(data)
	case "xml":
		parsedData, err = a.transformer.ParseXML(data)
	case "json":
		err = json.Unmarshal(data, &parsedData)
	default:
		// Try JSON by default
		err = json.Unmarshal(data, &parsedData)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse legacy data: %w", err)
	}

	// Transform to NEAM format
	transformedData, err := a.transformer.TransformLegacyData(parsedData, request.DataType)
	if err != nil {
		return nil, fmt.Errorf("legacy data transformation failed: %w", err)
	}

	// Archive the processed file
	if err := a.archiveFile(filePath); err != nil {
		// Log warning but don't fail
		fmt.Printf("Warning: failed to archive file: %v\n", err)
	}

	response := &models.AdapterResponse{
		StatusCode: 200,
		Body:       transformedData,
		Timestamp:  time.Now(),
	}

	return response, nil
}

// fetchFromFTP retrieves data from FTP server
func (a *LegacyAdapter) fetchFromFTP(ctx context.Context, request *models.AdapterRequest) (*models.AdapterResponse, error) {
	// FTP fetching would be implemented here
	// For now, return error indicating FTP not implemented
	return nil, fmt.Errorf("FTP fetching not yet implemented")
}

// Push sends data to legacy systems
func (a *LegacyAdapter) Push(ctx context.Context, message *models.IntegrationMessage) error {
	// Convert to legacy format
	legacyFormat, err := a.transformer.ToLegacyFormat(message.Payload, "csv")
	if err != nil {
		return fmt.Errorf("failed to convert to legacy format: %w", err)
	}

	// Write to output directory
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s_%s.csv", message.DataType, message.MessageID, timestamp)
	filePath := filepath.Join(a.config.Adapters.Legacy.OutputDirectory, filename)

	if err := ioutil.WriteFile(filePath, legacyFormat, 0644); err != nil {
		return fmt.Errorf("failed to write legacy file: %w", err)
	}

	return nil
}

// HealthCheck performs health check on legacy systems
func (a *LegacyAdapter) HealthCheck(ctx context.Context) error {
	// Check if directories are accessible
	dirs := []string{
		a.config.Adapters.Legacy.InputDirectory,
		a.config.Adapters.Legacy.OutputDirectory,
	}

	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return fmt.Errorf("legacy directory not accessible: %s", dir)
		}
	}

	return nil
}

// detectFormat detects the file format from extension
func (a *LegacyAdapter) detectFormat(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".csv":
		return "csv"
	case ".xml":
		return "xml"
	case ".json":
		return "json"
	default:
		return "unknown"
	}
}

// parseCSV parses CSV data into a slice of maps
func (a *LegacyAdapter) parseCSV(data []byte) ([]map[string]interface{}, error) {
	reader := csv.NewReader(strings.NewReader(string(data)))
	reader.FieldsPerRecord = -1 // Allow variable number of fields

	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV headers: %w", err)
	}

	var records []map[string]interface{}
	for {
		row, err := reader.Read()
		if err != nil {
			break
		}

		record := make(map[string]interface{})
		for i, header := range headers {
			if i < len(row) {
				record[header] = row[i]
			}
		}
		records = append(records, record)
	}

	return records, nil
}

// archiveFile moves a file to the archive directory
func (a *LegacyAdapter) archiveFile(filePath string) error {
	filename := filepath.Base(filePath)
	timestamp := time.Now().Format("20060102")
	archivePath := filepath.Join(a.config.Adapters.Legacy.ArchiveDirectory, timestamp, filename)

	if err := os.MkdirAll(filepath.Dir(archivePath), 0755); err != nil {
		return err
	}

	return os.Rename(filePath, archivePath)
}

// SupportedDataTypes returns the data types supported by this adapter
func (a *LegacyAdapter) SupportedDataTypes() []string {
	return []string{
		"legacy_citizen",
		"legacy_tax",
		"legacy_property",
		"legacy_benefit",
		"legacy_record",
	}
}

// ProcessFiles processes all files in the input directory
func (a *LegacyAdapter) ProcessFiles(ctx context.Context) error {
	files, err := ioutil.ReadDir(a.config.Adapters.Legacy.InputDirectory)
	if err != nil {
		return fmt.Errorf("failed to read input directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(a.config.Adapters.Legacy.InputDirectory, file.Name())

		request := &models.AdapterRequest{
			SourceType: "file",
			Endpoint:   filePath,
			DataType:   a.detectDataType(file.Name()),
			Transform:  true,
		}

		_, err := a.Fetch(ctx, request)
		if err != nil {
			return fmt.Errorf("failed to process file %s: %w", file.Name(), err)
		}
	}

	return nil
}

// detectDataType detects the data type from filename
func (a *LegacyAdapter) detectDataType(filename string) string {
	filename = strings.ToLower(filename)
	if strings.Contains(filename, "citizen") {
		return "legacy_citizen"
	}
	if strings.Contains(filename, "tax") {
		return "legacy_tax"
	}
	if strings.Contains(filename, "property") {
		return "legacy_property"
	}
	if strings.Contains(filename, "benefit") {
		return "legacy_benefit"
	}
	return "legacy_record"
}

// LegacyAdapter implements IntegrationAdapter
var _ IntegrationAdapter = (*LegacyAdapter)(nil)
