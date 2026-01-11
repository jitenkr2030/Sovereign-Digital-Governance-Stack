package swift

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// MTProcessor handles SWIFT MT message parsing
type MTProcessor struct {
	logger      *zap.Logger
	messageDefs map[string]*MTMessageDefinition
}

// MTMessageDefinition defines the structure of an MT message type
type MTMessageDefinition struct {
	Category  string
	Block1    []FieldDefinition
	Block2    []FieldDefinition
	Block3    []FieldDefinition
	Block4    []FieldDefinition
	Block5    []FieldDefinition
}

// FieldDefinition defines a field within a SWIFT block
type FieldDefinition struct {
	Tag        string
	Name       string
	Format     string
	Required   bool
	Repeatable bool
}

// MTMessage represents a parsed SWIFT MT message
type MTMessage struct {
	RawPayload       string            `json:"raw_payload"`
	Block1           *BasicHeaderBlock `json:"block1"`
	Block2           *ApplicationHeaderBlock `json:"block2"`
	Block3           *UserHeaderBlock  `json:"block3"`
	Block4           []Field           `json:"block4"`
	Block5           []Field           `json:"block5"`
	Checksum         string            `json:"checksum"`
	MessageType      string            `json:"message_type"`
	SenderBIC        string            `json:"sender_bic"`
	ReceiverBIC      string            `json:"receiver_bic"`
	TransactionID    string            `json:"transaction_id"`
	Amount           decimal.Decimal   `json:"amount"`
	Currency         string            `json:"currency"`
	SenderAccount    string            `json:"sender_account"`
	ReceiverAccount  string            `json:"receiver_account"`
	ValueDate        time.Time         `json:"value_date"`
	Timestamp        time.Time         `json:"timestamp"`
}

// BasicHeaderBlock contains SWIFT basic header information
type BasicHeaderBlock struct {
	ApplicationID  string `json:"application_id"`
	ServiceID      string `json:"service_id"`
	BIC            string `json:"bic"`
	SessionNumber  string `json:"session_number"`
	SequenceNumber string `json:"sequence_number"`
}

// ApplicationHeaderBlock contains SWIFT application header information
type ApplicationHeaderBlock struct {
	InputOutput    string `json:"input_output"`
	MessageType    string `json:"message_type"`
	ReceiverBIC    string `json:"receiver_bic"`
	DeliveryInfo   string `json:"delivery_info"`
	Priority       string `json:"priority"`
	ObsolescencePeriod string `json:"obsolescence_period"`
}

// UserHeaderBlock contains SWIFT user header information
type UserHeaderBlock struct {
	CorrespondentBank string `json:"correspondent_bank"`
	PaymentInfo       string `json:"payment_info"`
	RemittanceInfo    string `json:"remittance_info"`
	BankOperationCode string `json:"bank_operation_code"`
	CurrencyInstruct  string `json:"currency_instruct"`
	charges           string `json:"charges"`
	 SenderCharges    string `json:"sender_charges"`
	ReceiverCharges  string `json:"receiver_charges"`
}

// Field represents a parsed SWIFT field
type Field struct {
	Tag      string `json:"tag"`
	Name     string `json:"name"`
	Value    string `json:"value"`
	SubFields []SubField `json:"subfields"`
}

// SubField represents a subfield within a SWIFT field
type SubField struct {
	Code string `json:"code"`
	Value string `json:"value"`
}

// NewMTProcessor creates a new SWIFT MT processor
func NewMTProcessor(logger *zap.Logger) *MTProcessor {
	p := &MTProcessor{
		logger:      logger,
		messageDefs: make(map[string]*MTMessageDefinition),
	}

	p.registerMessageDefinitions()
	return p
}

// registerMessageDefinitions registers definitions for supported MT message types
func (p *MTProcessor) registerMessageDefinitions() {
	// MT103 - Customer Credit Transfer
	p.messageDefs["103"] = &MTMessageDefinition{
		Category: "1",
		Block1: []FieldDefinition{
			{Tag: "appid", Name: "Application ID", Format: "a", Required: true},
			{Tag: "servid", Name: "Service ID", Format: "a", Required: true},
			{Tag: "bic", Name: "Logical Terminal", Format: "x", Required: true},
			{Tag: "session", Name: "Session Number", Format: "n", Required: true},
			{Tag: "seqnr", Name: "Sequence Number", Format: "n", Required: true},
		},
		Block2: []FieldDefinition{
			{Tag: "io", Name: "Input/Output", Format: "a", Required: true},
			{Tag: "msgtype", Name: "Message Type", Format: "n", Required: true},
			{Tag: "receiver", Name: "Receiver Address", Format: "x", Required: true},
			{Tag: "delivery", Name: "Delivery Monitoring", Format: "n", Required: false},
			{Tag: "obsolescence", Name: "Obsolescence Period", Format: "n", Required: false},
		},
		Block3: []FieldDefinition{
			{Tag: ":proc", Name: "Processing Priority", Format: "n", Required: false},
			{Tag: ":mid", Name: "Monitoring Input", Format: "a", Required: false},
			{Tag: ":ric", Name: "Receiver Correspondent", Format: "x", Required: false},
			{Tag: ":ses", Name: "Sender's Correspondent", Format: "x", Required: false},
			{Tag: ":boc", Name: "Receiver's Correspondent", Format: "x", Required: false},
			{Tag: ":ref", Name: "Unique Transaction Reference", Format: "c", Required: true},
		},
		Block4: []FieldDefinition{
			{Tag: ":20", Name: "Transaction Reference", Format: "c", Required: true, Repeatable: false},
			{Tag: ":23B", Name: "Bank Operation Code", Format: "c", Required: true, Repeatable: false},
			{Tag: ":23E", Name: "Instruction Code", Format: "c", Required: false, Repeatable: true},
			{Tag: ":26T", Name: "Type of Entry", Format: "c", Required: false, Repeatable: false},
			{Tag: ":32A", Name: "Value Date/Currency/Amount", Format: "c", Required: true, Repeatable: false},
			{Tag: ":33B", Name: "Instructed Amount", Format: "c", Required: false, Repeatable: false},
			{Tag: ":50a", Name: "Ordering Customer", Format: "c", Required: true, Repeatable: true},
			{Tag: ":52a", Name: "Ordering Institution", Format: "c", Required: false, Repeatable: true},
			{Tag: ":53a", Name: "Sender's Correspondent", Format: "c", Required: false, Repeatable: true},
			{Tag: ":56a", Name: "Intermediary Institution", Format: "c", Required: false, Repeatable: true},
			{Tag: ":57a", Name: "Account With Institution", Format: "c", Required: true, Repeatable: true},
			{Tag: ":59a", Name: "Beneficiary Customer", Format: "c", Required: true, Repeatable: true},
			{Tag: ":70", Name: "Remittance Information", Format: "c", Required: false, Repeatable: true},
			{Tag: ":71A", Name: "Sender's Charges", Format: "c", Required: false, Repeatable: false},
			{Tag: ":71F", Name: "Receiver's Charges", Format: "c", Required: false, Repeatable: false},
			{Tag":72", Name: "Bank to Bank Information", Format: "c", Required: false, Repeatable: true},
		},
	}

	// MT202 - General Financial Institution Transfer
	p.messageDefs["202"] = &MTMessageDefinition{
		Category: "2",
		Block1: []FieldDefinition{
			{Tag: "appid", Name: "Application ID", Format: "a", Required: true},
			{Tag: "servid", Name: "Service ID", Format: "a", Required: true},
			{Tag: "bic", Name: "Logical Terminal", Format: "x", Required: true},
			{Tag: "session", Name: "Session Number", Format: "n", Required: true},
			{Tag: "seqnr", Name: "Sequence Number", Format: "n", Required: true},
		},
		Block2: []FieldDefinition{
			{Tag: "io", Name: "Input/Output", Format: "a", Required: true},
			{Tag: "msgtype", Name: "Message Type", Format: "n", Required: true},
			{Tag: "receiver", Name: "Receiver Address", Format: "x", Required: true},
			{Tag: "delivery", Name: "Delivery Monitoring", Format: "n", Required: false},
			{Tag: "obsolescence", Name: "Obsolescence Period", Format: "n", Required: false},
		},
		Block4: []FieldDefinition{
			{Tag: ":20", Name: "Transaction Reference", Format: "c", Required: true, Repeatable: false},
			{Tag: ":21", Name: "Related Reference", Format: "c", Required: false, Repeatable: false},
			{Tag: ":32A", Name: "Value Date/Currency/Amount", Format: "c", Required: true, Repeatable: false},
			{Tag: ":52a", Name: "Ordering Institution", Format: "c", Required: false, Repeatable: true},
			{Tag: ":53a", Name: "Sender's Correspondent", Format: "c", Required: false, Repeatable: true},
			{Tag: ":56a", Name: "Intermediary", Format: "c", Required: false, Repeatable: true},
			{Tag: ":57a", Name: "Account With Institution", Format: "c", Required: false, Repeatable: true},
			{Tag: ":72", Name: "Bank to Bank Information", Format: "c", Required: false, Repeatable: true},
		},
	}

	// MT202COV - General Financial Institution Transfer - Cover
	p.messageDefs["202COV"] = &MTMessageDefinition{
		Category: "2",
		Block4: []FieldDefinition{
			{Tag: ":20", Name: "Transaction Reference", Format: "c", Required: true, Repeatable: false},
			{Tag: ":21", Name: "Related Reference", Format: "c", Required: false, Repeatable: false},
			{Tag: ":22", Name: "Type of Operation", Format: "c", Required: true, Repeatable: false},
			{Tag: ":32A", Name: "Value Date/Currency/Amount", Format: "c", Required: true, Repeatable: false},
			{Tag: ":33A", Name: "Cover Amount", Format: "c", Required: false, Repeatable: false},
			{Tag: ":50a", Name: "Ordering Customer", Format: "c", Required: false, Repeatable: true},
			{Tag: ":52a", Name: "Ordering Institution", Format: "c", Required: false, Repeatable: true},
			{Tag: ":53a", Name: "Sender's Correspondent", Format: "c", Required: false, Repeatable: true},
			{Tag: ":56a", Name: "Intermediary", Format: "c", Required: false, Repeatable: true},
			{Tag: ":57a", Name: "Account With Institution", Format: "c", Required: false, Repeatable: true},
			{Tag: ":58a", Name: "Beneficiary Institution", Format: "c", Required: false, Repeatable: true},
			{Tag: ":72", Name: "Bank to Bank Information", Format: "c", Required: false, Repeatable: true},
		},
	}
}

// GetMessageType returns the processor message type
func (p *MTProcessor) GetMessageType() string {
	return "swift_mt"
}

// Parse parses SWIFT MT message and returns normalized payment
func (p *MTProcessor) Parse(ctx context.Context, raw []byte) (*NormalizedPayment, error) {
	startTime := time.Now()

	// Detect message type
	msgType, err := detectMTMessageType(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to detect MT message type: %w", err)
	}

	// Parse message blocks
	msg, err := p.parseMessage(raw, msgType)
	if err != nil {
		return nil, fmt.Errorf("failed to parse MT message: %w", err)
	}

	// Validate message
	if err := p.validateMessage(msg, msgType); err != nil {
		return nil, fmt.Errorf("validation failed for MT%s: %w", msgType, err)
	}

	// Convert to normalized payment
	normalized, err := p.toNormalizedPayment(msg, msgType, raw, startTime)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize MT message: %w", err)
	}

	p.logger.Debug("SWIFT MT message parsed",
		zap.String("message_type", msgType),
		zap.Duration("parsing_time", time.Since(startTime)))

	return normalized, nil
}

// Validate validates a SWIFT MT message
func (p *MTProcessor) Validate(msg interface{}) error {
	m, ok := msg.(*MTMessage)
	if !ok {
		return fmt.Errorf("invalid message type for MT validation")
	}

	if m.Block1 == nil {
		return fmt.Errorf("missing basic header block")
	}

	if m.Block2 == nil {
		return fmt.Errorf("missing application header block")
	}

	if len(m.Block4) == 0 {
		return fmt.Errorf("missing text block")
	}

	return nil
}

// detectMTMessageType detects the SWIFT MT message type from raw bytes
func detectMTMessageType(raw []byte) (string, error) {
	trimmed := bytes.TrimSpace(raw)

	// Look for block 2 message type pattern
	re := regexp.MustCompile(`\{2:([A-Z])([A-Z0-9]{3})\}`)
	matches := re.FindStringSubmatch(string(trimmed))

	if len(matches) >= 3 {
		msgType := matches[2]
		if _, ok := validMTMessageTypes[msgType]; ok {
			return msgType, nil
		}
	}

	return "", fmt.Errorf("unknown SWIFT MT message type")
}

// validMTMessageTypes contains valid MT message types
var validMTMessageTypes = map[string]bool{
	"103": true,
	"202": true,
	"202COV": true,
	"210": true,
	"900": true,
	"910": true,
	"940": true,
	"950": true,
}

// parseMessage parses a SWIFT MT message into structured format
func (p *MTProcessor) parseMessage(raw []byte, msgType string) (*MTMessage, error) {
	msg := &MTMessage{
		RawPayload: string(raw),
		Timestamp:  time.Now(),
	}

	// Parse block 1 (Basic Header)
	if err := p.parseBlock1(raw, msg); err != nil {
		return nil, fmt.Errorf("failed to parse block 1: %w", err)
	}

	// Parse block 2 (Application Header)
	if err := p.parseBlock2(raw, msg); err != nil {
		return nil, fmt.Errorf("failed to parse block 2: %w", err)
	}

	// Parse block 3 (User Header)
	p.parseBlock3(raw, msg)

	// Parse block 4 (Text Block)
	if err := p.parseBlock4(raw, msg); err != nil {
		return nil, fmt.Errorf("failed to parse block 4: %w", err)
	}

	// Parse block 5 (Trailer)
	p.parseBlock5(raw, msg)

	return msg, nil
}

// parseBlock1 parses the basic header block
func (p *MTProcessor) parseBlock1(raw []byte, msg *MTMessage) error {
	re := regexp.MustCompile(`\{1:([A-Z])([A-Z])([A-Z0-9]{12})([0-9]{6})([0-9]{6})\}`)
	matches := re.FindStringSubmatch(string(raw))

	if len(matches) < 6 {
		return fmt.Errorf("invalid block 1 format")
	}

	msg.Block1 = &BasicHeaderBlock{
		ApplicationID:  matches[1],
		ServiceID:      matches[2],
		BIC:            matches[3],
		SessionNumber:  matches[4],
		SequenceNumber: matches[5],
	}

	msg.SenderBIC = matches[3]
	msg.MessageType = matches[2] + "00" // Placeholder

	return nil
}

// parseBlock2 parses the application header block
func (p *MTProcessor) parseBlock2(raw []byte, msg *MTMessage) error {
	// Input block format: {2:I103BANKDEFFOXXXN}
	// Output block format: {2:O103BANKDEFFOXXXN}

	re := regexp.MustCompile(`\{2:([IO])([0-9]{3})([A-Z0-9]{8,11})(N|F)([0-9]{3})?\}`)
	matches := re.FindStringSubmatch(string(raw))

	if len(matches) < 5 {
		return fmt.Errorf("invalid block 2 format")
	}

	ioType := matches[1]
	msgType := matches[2]
	receiverBIC := matches[3]
	priority := matches[4]

	msg.Block2 = &ApplicationHeaderBlock{
		InputOutput:    ioType,
		MessageType:    msgType,
		ReceiverBIC:    receiverBIC,
		Priority:       priority,
	}

	msg.ReceiverBIC = receiverBIC
	msg.MessageType = msgType

	return nil
}

// parseBlock3 parses the user header block
func (p *MTProcessor) parseBlock3(raw []byte, msg *MTMessage) {
	block3Re := regexp.MustCompile(`\{3:([^\}]*)\}`)
	matches := block3Re.FindStringSubmatch(string(raw))

	if len(matches) < 2 {
		return
	}

	content := matches[1]
	msg.Block3 = &UserHeaderBlock{}

	// Parse individual tags
	tagRe := regexp.MustCompile(`:([A-Z0-9]{3}):([^\{]+)`)
	tagMatches := tagRe.FindAllStringSubmatch(content, -1)

	for _, tagMatch := range tagMatches {
		if len(tagMatch) >= 3 {
			tag := tagMatch[1]
			value := strings.TrimSpace(tagMatch[2])

			switch tag {
			case "RIC":
				msg.Block3.CorrespondentBank = value
			case "PID":
				msg.Block3.PaymentInfo = value
			case "ROC":
				msg.Block3.BankOperationCode = value
			case "PDE":
				msg.Block3.RemittanceInfo = value
			case "SSA":
				msg.Block3.SenderCharges = value
			case "SCA":
				msg.Block3.SenderCharges = value
			}
		}
	}
}

// parseBlock4 parses the text block
func (p *MTProcessor) parseBlock4(raw []byte, msg *MTMessage) error {
	block4Re := regexp.MustCompile(`\{4:([^\}]*)\}`)
	matches := block4Re.FindStringSubmatch(string(raw))

	if len(matches) < 2 {
		return fmt.Errorf("invalid block 4 format")
	}

	content := matches[1]
	msg.Block4 = p.parseBlock4Fields(content, msg.MessageType)

	// Extract key payment information
	p.extractPaymentDetails(msg)

	return nil
}

// parseBlock4Fields parses fields from block 4 content
func (p *MTProcessor) parseBlock4Fields(content string, msgType string) []Field {
	var fields []Field

	// Split by field delimiters, respecting subfield structure
	fieldRe := regexp.MustCompile(`:([A-Z0-9]{2,3})([^:\{]+)`)
	fieldMatches := fieldRe.FindAllStringSubmatch(content, -1)

	for _, fieldMatch := range fieldMatches {
		if len(fieldMatch) >= 3 {
			tag := fieldMatch[1]
			value := strings.TrimSpace(fieldMatch[2])

			field := Field{
				Tag:   tag,
				Name:  getFieldName(tag, msgType),
				Value: value,
			}

			// Parse subfields for certain tags
			if p.hasSubFields(tag) {
				field.SubFields = p.parseSubFields(value)
			}

			fields = append(fields, field)
		}
	}

	return fields
}

// hasSubFields checks if a tag has subfields
func (p *MTProcessor) hasSubFields(tag string) bool {
	subFieldTags := map[string]bool{
		"50a": true, "52a": true, "53a": true, "56a": true,
		"57a": true, "58a": true, "59a": true,
	}
	return subFieldTags[tag]
}

// parseSubFields parses subfields from a field value
func (p *MTProcessor) parseSubFields(value string) []SubField {
	var subFields []SubField

	// Subfields are typically separated by newlines or specific patterns
	lines := strings.Split(value, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) >= 2 {
			code := line[:2]
			subFieldValue := strings.TrimSpace(line[2:])
			subFields = append(subFields, SubField{
				Code:  code,
				Value: subFieldValue,
			})
		}
	}

	return subFields
}

// extractPaymentDetails extracts key payment details from parsed fields
func (p *MTProcessor) extractPaymentDetails(msg *MTMessage) {
	for _, field := range msg.Block4 {
		switch field.Tag {
		case "20":
			msg.TransactionID = strings.TrimSpace(field.Value)
		case "32A":
			// Value Date/Currency/Amount: YYMMDDCCURRENCYAMOUNT
			if len(field.Value) >= 15 {
				dateStr := field.Value[:6]
				currency := field.Value[6:9]
				amount := field.Value[9:]

				// Parse date
				if parsedDate, err := parseSwiftDate(dateStr); err == nil {
					msg.ValueDate = parsedDate
				}

				msg.Currency = currency

				// Parse amount
				if amt, err := strconv.ParseFloat(amount, 64); err == nil {
					msg.Amount = decimal.NewFromFloat(amt)
				}
			}
		case "33B":
			// Instructed amount
			if len(field.Value) >= 3 {
				currency := field.Value[:3]
				amount := field.Value[3:]
				msg.Currency = currency
				if amt, err := strconv.ParseFloat(amount, 64); err == nil {
					msg.Amount = decimal.NewFromFloat(amt)
				}
			}
		case "50A", "50F":
			// Ordering customer - account number
			lines := strings.Split(field.Value, "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "/") {
					msg.SenderAccount = strings.TrimPrefix(line, "/")
					break
				}
			}
		case "52A", "52D":
			// Ordering institution
		case "53A", "53B", "53D":
			// Sender's correspondent
		case "56A", "56C", "56D":
			// Intermediary
		case "57A", "57B", "57C", "57D":
			// Account with institution
			if strings.HasPrefix(field.Value, "//") {
				msg.ReceiverAccount = strings.TrimPrefix(field.Value, "//")
			}
		case "59", "59A", "59F":
			// Beneficiary customer
			lines := strings.Split(field.Value, "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "/") {
					msg.ReceiverAccount = strings.TrimPrefix(line, "/")
					break
				}
			}
		case "70":
			// Remittance information
		case "71A":
			// Sender's charges
		case "71F":
			// Receiver's charges
		case "72":
			// Bank to bank information
		}
	}
}

// parseBlock5 parses the trailer block
func (p *MTProcessor) parseBlock5(raw []byte, msg *MTMessage) {
	// Block 5 contains optional trailer information
	// {5:{CHK:XXXXXX}}
	chkRe := regexp.MustCompile(`\{5:\{CHK:([A-Z0-9]{6})\}\}`)
	matches := chkRe.FindStringSubmatch(string(raw))

	if len(matches) >= 2 {
		msg.Checksum = matches[1]
	}
}

// toNormalizedPayment converts parsed message to normalized payment
func (p *MTProcessor) toNormalizedPayment(msg *MTMessage, msgType string, raw []byte, startTime time.Time) (*NormalizedPayment) {
	// Calculate message hash for idempotency
	hash := sha256.Sum256(raw)

	normalized := &NormalizedPayment{
		ID:             fmt.Sprintf("SWMT-%s-%d", msgType, startTime.UnixNano()),
		MessageHash:    fmt.Sprintf("%x", hash[:16]),
		Format:         "swift_mt",
		MessageType:    fmt.Sprintf("MT%s", msgType),
		TransactionID:  msg.TransactionID,
		Amount:         msg.Amount,
		Currency:       msg.Currency,
		SenderBIC:      msg.SenderBIC,
		ReceiverBIC:    msg.ReceiverBIC,
		SenderAccount:  msg.SenderAccount,
		ReceiverAccount: msg.ReceiverAccount,
		ValueDate:      msg.ValueDate,
		ProcessingDate: startTime,
		Timestamp:      startTime,
		RawPayload:     json.RawMessage(msg.RawPayload),
		Metadata: map[string]interface{}{
			"block1": msg.Block1,
			"block2": msg.Block2,
		},
	}

	return normalized
}

// parseSwiftDate parses a SWIFT date string (YYMMDD)
func parseSwiftDate(dateStr string) (time.Time, error) {
	if len(dateStr) != 6 {
		return time.Time{}, fmt.Errorf("invalid date format: %s", dateStr)
	}

	year := 2000 + parseInt(dateStr[0:2])
	month := parseInt(dateStr[2:4])
	day := parseInt(dateStr[4:6])

	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC), nil
}

// parseInt parses an integer from string
func parseInt(s string) int {
	val, _ := strconv.Atoi(s)
	return val
}

// getFieldName returns the field name for a given tag
func getFieldName(tag string, msgType string) string {
	fieldNames := map[string]string{
		"20":  "Transaction Reference",
		"21":  "Related Reference",
		"23B": "Bank Operation Code",
		"23E": "Instruction Code",
		"26T": "Type of Entry",
		"32A": "Value Date/Currency/Amount",
		"33B": "Instructed Amount",
		"50A": "Ordering Customer (BIC)",
		"50F": "Ordering Customer (Financial Institution)",
		"52A": "Ordering Institution (BIC)",
		"53A": "Sender's Correspondent (BIC)",
		"56A": "Intermediary (BIC)",
		"57A": "Account With Institution (BIC)",
		"58A": "Beneficiary Institution (BIC)",
		"59":  "Beneficiary Customer",
		"70":  "Remittance Information",
		"71A": "Sender's Charges",
		"71F": "Receiver's Charges",
		"72":  "Bank to Bank Information",
	}

	if name, ok := fieldNames[tag]; ok {
		return name
	}
	return fmt.Sprintf("Field %s", tag)
}

// NormalizedPayment is a placeholder to satisfy the interface
type NormalizedPayment struct {
	ID             string                 `json:"id"`
	MessageHash    string                 `json:"message_hash"`
	Format         string                 `json:"format"`
	MessageType    string                 `json:"message_type"`
	TransactionID  string                 `json:"transaction_id"`
	Amount         decimal.Decimal        `json:"amount"`
	Currency       string                 `json:"currency"`
	SenderBIC      string                 `json:"sender_bic"`
	ReceiverBIC    string                 `json:"receiver_bic"`
	SenderAccount  string                 `json:"sender_account"`
	ReceiverAccount string                `json:"receiver_account"`
	ValueDate      time.Time              `json:"value_date"`
	ProcessingDate time.Time              `json:"processing_date"`
	Timestamp      time.Time              `json:"timestamp"`
	RawPayload     json.RawMessage        `json:"raw_payload"`
	Metadata       map[string]interface{} `json:"metadata"`
}
