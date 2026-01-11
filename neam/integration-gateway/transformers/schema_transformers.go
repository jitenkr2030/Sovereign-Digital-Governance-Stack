package transformers

import (
	"encoding/json"
	"fmt"
	"time"

	"integration-gateway/models"
)

// SchemaTransformer defines the interface for all schema transformations
type SchemaTransformer interface {
	Transform(input interface{}) (*models.CanonicalEvent, error)
	TransformBatch(inputs []interface{}) ([]*models.CanonicalEvent, error)
	GetSupportedSchema() string
}

// MinistryToCanonicalTransformer transforms Ministry of Finance data to canonical format
type MinistryToCanonicalTransformer struct {
	SchemaVersion string
}

func NewMinistryToCanonicalTransformer() *MinistryToCanonicalTransformer {
	return &MinistryToCanonicalTransformer{
		SchemaVersion: "1.0",
	}
}

func (t *MinistryToCanonicalTransformer) GetSupportedSchema() string {
	return "ministry-finance-v1.0"
}

func (t *MinistryToCanonicalTransformer) Transform(input interface{}) (*models.CanonicalEvent, error) {
	data, ok := input.(map[string]interface{})
	if !ok {
		// Try to unmarshal from JSON
		jsonData, err := json.Marshal(input)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal input: %w", err)
		}
		if err := json.Unmarshal(jsonData, &data); err != nil {
			return nil, fmt.Errorf("input is not a valid ministry data structure: %w", err)
		}
	}

	event := &models.CanonicalEvent{
		EventID:       generateEventID(),
		SourceSystem:  "ministry-finance",
		EventType:     t.mapEventType(data),
		Timestamp:     time.Now().UTC(),
		SchemaVersion: t.SchemaVersion,
		Data:          t.extractData(data),
		Metadata:      t.extractMetadata(data),
	}

	return event, nil
}

func (t *MinistryToCanonicalTransformer) TransformBatch(inputs []interface{}) ([]*models.CanonicalEvent, error) {
	events := make([]*models.CanonicalEvent, 0, len(inputs))
	for i, input := range inputs {
		event, err := t.Transform(input)
		if err != nil {
			return nil, fmt.Errorf("failed to transform input at index %d: %w", i, err)
		}
		events = append(events, event)
	}
	return events, nil
}

func (t *MinistryToCanonicalTransformer) mapEventType(data map[string]interface{}) string {
	if eventType, ok := data["event_type"].(string); ok {
		switch eventType {
		case "budget_allocation":
			return "budget.allocated"
		case "expenditure_approval":
			return "expenditure.approved"
		case "revenue_report":
			return "revenue.reported"
		case "fiscal_transfer":
			return "fiscal.transferred"
		default:
			return eventType
		}
	}
	return "unknown"
}

func (t *MinistryToCanonicalTransformer) extractData(data map[string]interface{}) map[string]interface{} {
	extracted := make(map[string]interface{})

	// Extract common financial fields
	if amount, ok := data["amount"].(float64); ok {
		extracted["amount"] = amount
	}
	if currency, ok := data["currency"].(string); ok {
		extracted["currency"] = currency
	}
	if entityID, ok := data["entity_id"].(string); ok {
		extracted["entity_id"] = entityID
	}
	if fiscalYear, ok := data["fiscal_year"].(float64); ok {
		extracted["fiscal_year"] = int(fiscalYear)
	}

	// Additional ministry-specific fields
	if department, ok := data["department"].(string); ok {
		extracted["department"] = department
	}
	if budgetCode, ok := data["budget_code"].(string); ok {
		extracted["budget_code"] = budgetCode
	}

	return extracted
}

func (t *MinistryToCanonicalTransformer) extractMetadata(data map[string]interface{}) map[string]interface{} {
	metadata := make(map[string]interface{})
	metadata["transformer"] = "ministry-to-canonical"
	metadata["transformed_at"] = time.Now().UTC().Format(time.RFC3339)

	// Preserve original identifiers
	if originalID, ok := data["id"].(string); ok {
		metadata["original_event_id"] = originalID
	}

	return metadata
}

// LegacyToCanonicalTransformer transforms legacy system data to canonical format
type LegacyToCanonicalTransformer struct {
	SchemaVersion string
}

func NewLegacyToCanonicalTransformer() *LegacyToCanonicalTransformer {
	return &LegacyToCanonicalTransformer{
		SchemaVersion: "1.0",
	}
}

func (t *LegacyToCanonicalTransformer) GetSupportedSchema() string {
	return "legacy-csv-v1.0"
}

func (t *LegacyToCanonicalTransformer) Transform(input interface{}) (*models.CanonicalEvent, error) {
	data, ok := input.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("legacy input must be a map structure")
	}

	event := &models.CanonicalEvent{
		EventID:       generateEventID(),
		SourceSystem:  "legacy-system",
		EventType:     t.mapLegacyType(data),
		Timestamp:     t.parseLegacyTimestamp(data),
		SchemaVersion: t.SchemaVersion,
		Data:          t.extractLegacyData(data),
		Metadata:      t.extractLegacyMetadata(data),
	}

	return event, nil
}

func (t *LegacyToCanonicalTransformer) TransformBatch(inputs []interface{}) ([]*models.CanonicalEvent, error) {
	events := make([]*models.CanonicalEvent, 0, len(inputs))
	for i, input := range inputs {
		event, err := t.Transform(input)
		if err != nil {
			return nil, fmt.Errorf("failed to transform legacy input at index %d: %w", i, err)
		}
		events = append(events, event)
	}
	return events, nil
}

func (t *LegacyToCanonicalTransformer) mapLegacyType(data map[string]interface{}) string {
	if recordType, ok := data["record_type"].(string); ok {
		switch recordType {
		case "TR":
			return "transaction.recorded"
		case "PV":
			return "payment.verified"
		"AU":
			return "audit.logged"
		case "BL":
			return "balance.updated"
		default:
			return "record." + recordType
		}
	}
	return "legacy.record"
}

func (t *LegacyToCanonicalTransformer) parseLegacyTimestamp(data map[string]interface{}) time.Time {
	if timestamp, ok := data["timestamp"].(string); ok {
		// Try common legacy formats
		formats := []string{
			"2006-01-02 15:04:05",
			"2006-01-02",
			"02/01/2006",
			"01-02-2006",
		}
		for _, format := range formats {
			if t, err := time.Parse(format, timestamp); err == nil {
				return t.UTC()
			}
		}
	}
	return time.Now().UTC()
}

func (t *LegacyToCanonicalTransformer) extractLegacyData(data map[string]interface{}) map[string]interface{} {
	extracted := make(map[string]interface{})

	// Standard legacy fields
	if amount, ok := data["amt"].(float64); ok {
		extracted["amount"] = amount
	}
	if accountNum, ok := data["acct_num"].(string); ok {
		extracted["account_number"] = accountNum
	}
	if desc, ok := data["desc"].(string); ok {
		extracted["description"] = desc
	}

	return extracted
}

func (t *LegacyToCanonicalTransformer) extractLegacyMetadata(data map[string]interface{}) map[string]interface{} {
	metadata := make(map[string]interface{})
	metadata["transformer"] = "legacy-to-canonical"
	metadata["transformed_at"] = time.Now().UTC().Format(time.RFC3339)
	metadata["legacy_source"] = "legacy-system-v1"

	return metadata
}

// CentralBankToCanonicalTransformer transforms Central Bank data to canonical format
type CentralBankToCanonicalTransformer struct {
	SchemaVersion string
}

func NewCentralBankToCanonicalTransformer() *CentralBankToCanonicalTransformer {
	return &CentralBankToCanonicalTransformer{
		SchemaVersion: "1.0",
	}
}

func (t *CentralBankToCanonicalTransformer) GetSupportedSchema() string {
	return "central-bank-v1.0"
}

func (t *CentralBankToCanonicalTransformer) Transform(input interface{}) (*models.CanonicalEvent, error) {
	data, ok := input.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("central bank input must be a map structure")
	}

	event := &models.CanonicalEvent{
		EventID:       generateEventID(),
		SourceSystem:  "central-bank",
		EventType:     t.mapCBEventType(data),
		Timestamp:     time.Now().UTC(),
		SchemaVersion: t.SchemaVersion,
		Data:          t.extractCBData(data),
		Metadata:      t.extractCBMetadata(data),
	}

	return event, nil
}

func (t *CentralBankToCanonicalTransformer) TransformBatch(inputs []interface{}) ([]*models.CanonicalEvent, error) {
	events := make([]*models.CanonicalEvent, 0, len(inputs))
	for i, input := range inputs {
		event, err := t.Transform(input)
		if err != nil {
			return nil, fmt.Errorf("failed to transform central bank input at index %d: %w", i, err)
		}
		events = append(events, event)
	}
	return events, nil
}

func (t *CentralBankToCanonicalTransformer) mapCBEventType(data map[string]interface{}) string {
	if msgType, ok := data["message_type"].(string); ok {
		switch msgType {
		case "RTGS":
			return "payment.rtgs"
		case "CLEARING":
			return "payment.clearing"
		case "DIRECT_DEBIT":
			return "payment.direct_debit"
		case "WIRE_TRANSFER":
			return "payment.wire_transfer"
		case "STATISTICS":
			return "banking.statistics"
		case "RESERVE_REQUIREMENT":
			return "banking.reserve_requirement"
		default:
			return "banking." + msgType
		}
	}
	return "banking.unknown"
}

func (t *CentralBankToCanonicalTransformer) extractCBData(data map[string]interface{}) map[string]interface{} {
	extracted := make(map[string]interface{})

	// Central bank specific fields
	if senderBank, ok := data["sender_bank"].(string); ok {
		extracted["sender_bank"] = senderBank
	}
	if receiverBank, ok := data["receiver_bank"].(string); ok {
		extracted["receiver_bank"] = receiverBank
	}
	if amount, ok := data["transaction_amount"].(float64); ok {
		extracted["amount"] = amount
	}
	if currency, ok := data["currency_code"].(string); ok {
		extracted["currency"] = currency
	}
	if refNum, ok := data["reference_number"].(string); ok {
		extracted["reference_number"] = refNum
	}
	if isUrgent, ok := data["urgent"].(bool); ok {
		extracted["urgent_payment"] = isUrgent
	}

	return extracted
}

func (t *CentralBankToCanonicalTransformer) extractCBMetadata(data map[string]interface{}) map[string]interface{} {
	metadata := make(map[string]interface{})
	metadata["transformer"] = "central-bank-to-canonical"
	metadata["transformed_at"] = time.Now().UTC().Format(time.RFC3339)

	if originalMsgID, ok := data["message_id"].(string); ok {
		metadata["original_message_id"] = originalMsgID
	}

	return metadata
}

// TransformerRegistry manages all available transformers
type TransformerRegistry struct {
	transformers map[string]SchemaTransformer
}

func NewTransformerRegistry() *TransformerRegistry {
	registry := &TransformerRegistry{
		transformers: make(map[string]SchemaTransformer),
	}

	// Register all available transformers
	registry.transformers["ministry-finance-v1.0"] = NewMinistryToCanonicalTransformer()
	registry.transformers["legacy-csv-v1.0"] = NewLegacyToCanonicalTransformer()
	registry.transformers["central-bank-v1.0"] = NewCentralBankToCanonicalTransformer()

	return registry
}

func (r *TransformerRegistry) GetTransformer(schema string) (SchemaTransformer, error) {
	transformer, ok := r.transformers[schema]
	if !ok {
		return nil, fmt.Errorf("no transformer found for schema: %s", schema)
	}
	return transformer, nil
}

func (r *TransformerRegistry) RegisterTransformer(schema string, transformer SchemaTransformer) {
	r.transformers[schema] = transformer
}

func (r *TransformerRegistry) GetSupportedSchemas() []string {
	schemas := make([]string, 0, len(r.transformers))
	for schema := range r.transformers {
		schemas = append(schemas, schema)
	}
	return schemas
}

// CanonicalToOutputTransformer transforms canonical events to specific output formats
type CanonicalToOutputTransformer struct {
	TargetFormat string
}

func NewCanonicalToOutputTransformer(targetFormat string) *CanonicalToOutputTransformer {
	return &CanonicalToOutputTransformer{TargetFormat: targetFormat}
}

func (t *CanonicalToOutputTransformer) Transform(event *models.CanonicalEvent) (interface{}, error) {
	switch t.TargetFormat {
	case "json":
		return t.toJSON(event)
	case "xml":
		return t.toXML(event)
	case "avro":
		return t.toAvro(event)
	default:
		return event, nil
	}
}

func (t *CanonicalToOutputTransformer) toJSON(event *models.CanonicalEvent) (interface{}, error) {
	return json.Marshal(event)
}

func (t *CanonicalToOutputTransformer) toXML(event *models.CanonicalEvent) (interface{}, error) {
	// Simplified XML conversion
	type XMLCanonicalEvent struct {
		EventID       string `xml:"event_id"`
		SourceSystem  string `xml:"source_system"`
		EventType     string `xml:"event_type"`
		Timestamp     string `xml:"timestamp"`
		SchemaVersion string `xml:"schema_version"`
		Data          string `xml:"data"`
	}

	// In a real implementation, proper XML marshaling would be used
	return event, nil
}

func (t *CanonicalToOutputTransformer) toAvro(event *models.CanonicalEvent) (interface{}, error) {
	// In a real implementation, Avro schema marshaling would be used
	return event, nil
}

// Helper function to generate unique event IDs
func generateEventID() string {
	return fmt.Sprintf("%d-%s", time.Now().UnixNano(), randomString(8))
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
		time.Sleep(time.Nanosecond)
	}
	return string(b)
}
