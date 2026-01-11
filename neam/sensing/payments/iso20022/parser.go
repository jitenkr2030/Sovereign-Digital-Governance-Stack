package iso20022

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"
	"time"

	"github.com/neamplatform/sensing/payments/config"
	"go.uber.org/zap"
)

// Processor handles ISO 20022 message parsing and validation
type Processor struct {
	logger     *zap.Logger
	schemas    map[string]*SchemaDefinition
	validators map[string]MessageValidator
}

// SchemaDefinition represents an ISO 20022 schema definition
type SchemaDefinition struct {
	Name       string
	Namespace  string
	RootElement string
	MessageType string
}

// MessageValidator validates ISO 20022 messages
type MessageValidator interface {
	Validate(msg interface{}) error
}

// NewProcessor creates a new ISO 20022 processor
func NewProcessor(logger *zap.Logger) *Processor {
	p := &Processor{
		logger:     logger,
		schemas:    make(map[string]*SchemaDefinition),
		validators: make(map[string]MessageValidator),
	}

	p.registerSchemas()
	p.registerValidators()

	return p
}

// registerSchemas registers supported ISO 20022 message schemas
func (p *Processor) registerSchemas() {
	p.schemas["pacs.008"] = &SchemaDefinition{
		Name:        "pacs.008",
		Namespace:   "urn:iso:std:iso:20022:tech:xsd:pacs.008.001.10",
		RootElement: "Document",
		MessageType: "pacs.008",
	}

	p.schemas["pacs.004"] = &SchemaDefinition{
		Name:        "pacs.004",
		Namespace:   "urn:iso:std:iso:20022:tech:xsd:pacs.004.001.09",
		RootElement: "Document",
		MessageType: "pacs.004",
	}

	p.schemas["camt.053"] = &SchemaDefinition{
		Name:        "camt.053",
		Namespace:   "urn:iso:std:iso:20022:tech:xsd:camt.053.001.02",
		RootElement: "Document",
		MessageType: "camt.053",
	}

	p.schemas["camt.052"] = &SchemaDefinition{
		Name:        "camt.052",
		Namespace:   "urn:iso:std:iso:20022:tech:xsd:camt.052.001.02",
		RootElement: "Document",
		MessageType: "camt.052",
	}

	p.schemas["pain.001"] = &SchemaDefinition{
		Name:        "pain.001",
		Namespace:   "urn:iso:std:iso:20022:tech:xsd:pain.001.001.09",
		RootElement: "Document",
		MessageType: "pain.001",
	}

	p.schemas["pain.002"] = &SchemaDefinition{
		Name:        "pain.002",
		Namespace:   "urn:iso:std:iso:20022:tech:xsd:pain.002.001.09",
		RootElement: "Document",
		MessageType: "pain.002",
	}
}

// registerValidators registers message validators
func (p *Processor) registerValidators() {
	p.validators["pacs.008"] = &Pacs008Validator{}
	p.validators["pacs.004"] = &Pacs004Validator{}
	p.validators["camt.053"] = &Camt053Validator{}
	p.validators["pain.001"] = &Pain001Validator{}
}

// GetMessageType returns the processor message type
func (p *Processor) GetMessageType() string {
	return "iso20022"
}

// Parse parses ISO 20022 message and returns normalized payment
func (p *Processor) Parse(ctx context.Context, raw []byte) (*NormalizedPayment, error) {
	startTime := time.Now()

	// Detect message type
	msgType, err := detectMessageType(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to detect message type: %w", err)
	}

	// Parse XML
	doc, err := p.parseXML(raw, msgType)
	if err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	// Validate message
	if validator, ok := p.validators[msgType]; ok {
		if err := validator.Validate(doc); err != nil {
			return nil, fmt.Errorf("validation failed for %s: %w", msgType, err)
		}
	}

	// Convert to normalized payment
	normalized, err := p.toNormalizedPayment(doc, msgType, raw, startTime)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize payment: %w", err)
	}

	p.logger.Debug("ISO 20022 message parsed",
		zap.String("message_type", msgType),
		zap.Duration("parsing_time", time.Since(startTime)))

	return normalized, nil
}

// Validate validates an ISO 20022 message
func (p *Processor) Validate(msg interface{}) error {
	// Implement generic validation logic
	return nil
}

// parseXML parses raw bytes into XML document
func (p *Processor) parseXML(raw []byte, msgType string) (interface{}, error) {
	decoder := xml.NewDecoder(bytes.NewReader(raw))

	// Find the root element and parse based on message type
	for {
		token, err := decoder.Token()
		if err != nil {
			return nil, fmt.Errorf("failed to parse XML token: %w", err)
		}

		if token == nil {
			break
		}

		if element, ok := token.(xml.StartElement); ok {
			switch msgType {
			case "pacs.008":
				return p.parsePacs008(decoder, &element)
			case "pacs.004":
				return p.parsePacs004(decoder, &element)
			case "camt.053":
				return p.parseCamt053(decoder, &element)
			case "pain.001":
				return p.parsePain001(decoder, &element)
			default:
				return p.parseGeneric(decoder, &element)
			}
		}
	}

	return nil, fmt.Errorf("could not parse message type: %s", msgType)
}

// Pacs008Document represents ISO 20022 pacs.008 message structure
type Pacs008Document struct {
	XMLName        xml.Name       `xml:"Document"`
	Xmlns          string         `xml:"xmlns,attr"`
	CreditTransfer CreditTransfer `xml:"FIToFICstmrCdtTrf"`
}

// CreditTransfer represents the credit transfer transaction
type CreditTransfer struct {
	GrpHdr       GroupHeader       `xml:"GrpHdr"`
	CdtTrfTxInf  []CreditTransferInfo `xml:"CdtTrfTxInf"`
	SplmtryData  []SupplementaryData `xml:"SplmtryData"`
}

// GroupHeader contains transaction group information
type GroupHeader struct {
	MsgId       string       `xml:"MsgId"`
	CreDtTm     time.Time    `xml:"CreDtTm"`
	NbOfTxs     string       `xml:"NbOfTxs"`
	SttlmInf    SettlementInfo `xml:"SttlmInf"`
	InstgAgt    Agent        `xml:"InstgAgt"`
	InstdAgt    Agent        `xml:"InstdAgt"`
}

// SettlementInfo contains settlement details
type SettlementInfo struct {
	SttlmMtd      string       `xml:"SttlmMtd"`
	SttlmAcct     Account      `xml:"SttlmAcct"`
	SttlmSidePty  Agent        `xml:"SttlmSidePty"`
}

// Agent represents a financial institution agent
type Agent struct {
	FinInstnId FinancialInstitutionIdentification `xml:"FinInstnId"`
}

// FinancialInstitutionIdentification contains BIC and other identification
type FinancialInstitutionIdentification struct {
	BIC     string `xml:"BIC"`
	ClrSysM string `xml:"ClrSysMmbId"`
	Name    string `xml:"Nm"`
}

// Account contains account information
type Account struct {
	Id   string `xml:"Id"`
	Iban string `xml:"IBAN"`
}

// CreditTransferInfo contains individual credit transfer details
type CreditTransferInfo struct {
	PtId               string              `xml:"PmtId"`
	InstrId            string              `xml:"InstrId"`
	EndToEndId         string              `xml:"EndToEndId"`
	TxId               string              `xml:"TxId"`
	CdtrRefInf         CreditorReferenceInfo `xml:"CdtrRefInf"`
	Amt                Amount              `xml:"Amt"`
	CdtrAgt            Agent               `xml:"CdtrAgt"`
	Cdtr               Party               `xml:"Cdtr"`
	CdtrAcct           Account             `xml:"CdtrAcct"`
	DbtrAgt            Agent               `xml:"DbtrAgt"`
	Dbtr               Party               `xml:"Dbtr"`
	DbtrAcct           Account             `xml:"DbtrAcct"`
	Purps              Purpose             `xml:"Purp"`
	RegnReg            RegulatoryReporting `xml:"RgltryRptg"`
	SplmtryData        []SupplementaryData `xml:"SplmtryData"`
}

// Amount represents a monetary amount
type Amount struct {
	InstdAmt     string  `xml:"InstdAmt,attr"`
	Ccy          string  `xml:"Ccy,attr"`
	Value        float64 `xml:",chardata"`
}

// Party represents a payment party
type Party struct {
	Nm   string  `xml:"Nm"`
	PstlAdx PostalAddress `xml:"PstlAdx"`
}

// PostalAddress contains postal address information
type PostalAddress struct {
	StrtNm string `xml:"StrtNm"`
	TwnNm  string `xml:"TwnNm"`
	Ctry   string `xml:"Ctry"`
	PstCd  string `xml:"PstCd"`
}

// CreditorReferenceInfo contains creditor reference
type CreditorReferenceInfo struct {
	Tp  CodeType `xml:"Tp"`
	Ref string   `xml:"Ref"`
}

// CodeType represents a code type
type CodeType struct {
	CdOrPrtry CdOrPrtry `xml:"CdOrPrtry"`
}

// CdOrPrtry contains code or proprietary information
type CdOrPrtry struct {
	Cd string `xml:"Cd"`
}

// Purpose contains transaction purpose
type Purpose struct {
	Cd string `xml:"Cd"`
}

// RegulatoryReporting contains regulatory reporting information
type RegulatoryReporting struct {
	Dt       string           `xml:"Dtls>Authstn>Dt"`
	Authstn  string           `xml:"Dtls>Authstn>AuthstnNr"`
	Ctry     string           `xml:"Dtls>CtrySubDvsn"`
	RltdDtls []RelatedDetails `xml:"Dtls>RltdDtls"`
}

// RelatedDetails contains related details
type RelatedDetails struct {
	Tp  string `xml:"Tp"`
	Dt  string `xml:"Dt"`
	Nb  string `xml:"Nb"`
}

// SupplementaryData contains supplementary data
type SupplementaryData struct {
	Envlp string `xml:"Envlp"`
}

// parsePacs008 parses pacs.008 credit transfer message
func (p *Processor) parsePacs008(decoder *xml.Decoder, startElement *xml.StartElement) (*Pacs008Document, error) {
	var doc Pacs008Document
	if err := decoder.DecodeElement(&doc, startElement); err != nil {
		return nil, fmt.Errorf("failed to decode pacs.008: %w", err)
	}
	return &doc, nil
}

// Pacs004Document represents ISO 20022 pacs.004 message structure
type Pacs004Document struct {
	XMLName xml.Name      `xml:"Document"`
	Xmlns   string        `xml:"xmlns,attr"`
	Return  PaymentReturn `xml:"FIToFIPmtStsRpt"`
}

// PaymentReturn represents a payment return message
type PaymentReturn struct {
	GrpHdr         GroupHeader         `xml:"GrpHdr"`
	TxInfAndSts    []TransactionStatus `xml:"TxInfAndSts"`
	SplmtryData    []SupplementaryData `xml:"SplmtryData"`
}

// TransactionStatus contains transaction status information
type TransactionStatus struct {
	OrgnlGrpInf     OriginalGroupInfo      `xml:"OrgnlGrpInf"`
	OrgnlTxId       string                 `xml:"OrgnlTxId"`
	TxSts           string                 `xml:"TxSts"`
	StsRsnInf       []StatusReasonInfo     `xml:"StsRsnInf"`
	OrgnlAmt        Amount                 `xml:"OrgnlAmt"`
	OrgnlCcy        string                 `xml:"OrgnlCcy"`
	ReqToExctnDt    time.Time              `xml:"ReqToExctnDt"`
	SplmtryData     []SupplementaryData    `xml:"SplmtryData"`
}

// OriginalGroupInfo contains original group information
type OriginalGroupInfo struct {
	OrgnlMsgId     string `xml:"OrgnlMsgId"`
	OrgnlMsgNmId   string `xml:"OrgnlMsgNmId"`
	OrgnlCreDtTm   string `xml:"OrgnlCreDtTm"`
}

// StatusReasonInfo contains status reason information
type StatusReasonInfo struct {
	Rsns   []ReasonCode `xml:"Rsns"`
	Orgtr  Party        `xml:"Orgtr"`
	AddtRsnInf string   `xml:"AddtRsnInf"`
}

// ReasonCode represents a reason code
type ReasonCode struct {
	Cd string `xml:"Cd"`
}

// parsePacs004 parses pacs.004 payment return message
func (p *Processor) parsePacs004(decoder *xml.Decoder, startElement *xml.StartElement) (*Pacs004Document, error) {
	var doc Pacs004Document
	if err := decoder.DecodeElement(&doc, startElement); err != nil {
		return nil, fmt.Errorf("failed to decode pacs.004: %w", err)
	}
	return &doc, nil
}

// Camt053Document represents ISO 20022 camt.053 message structure
type Camt053Document struct {
	XMLName xml.Name     `xml:"Document"`
	Xmlns   string       `xml:"xmlns,attr"`
	Stmt    BankStatement `xml:"BkToCstmrStmt"`
}

// BankStatement represents a bank statement
type BankStatement struct {
	GrpHdr       GroupHeader `xml:"GrpHdr"`
	Stmt         []Statement `xml:"Stmt"`
	SplmtryData  []SupplementaryData `xml:"SplmtryData"`
}

// Statement represents a statement
type Statement struct {
	Id          string      `xml:"Id"`
	ElctrncSeq  string      `xml:"ElctrncSeq"`
	LglSeq      string      `xml:"LglSeq"`
	CreDtTm     time.Time   `xml:"CreDtTm"`
	FrToDt      DateRange   `xml:"FrToDt"`
	Acct        Account     `xml:"Acct"`
	Bal         []Balance   `xml:"Bal"`
	Ntry        []Entry     `xml:"Ntry"`
	Txt         string      `xml:"Txt"`
}

// DateRange represents a date range
type DateRange struct {
	FrDtTm time.Time `xml:"FrDtTm"`
	ToDtTm time.Time `xml:"ToDtTm"`
}

// Balance represents account balance
type Balance struct {
	Tp     CodeType   `xml:"Tp"`
	Amt    Amount     `xml:"Amt"`
	CdtDbt string     `xml:"CdtDbt"`
	Dt     DateTime   `xml:"Dt"`
}

// DateTime represents a date and time
type DateTime struct {
	DateTime time.Time `xml:"DtTm"`
}

// Entry represents a statement entry
type Entry struct {
	NtryRef    string      `xml:"NtryRef"`
	Amt        Amount      `xml:"Amt"`
	CdtDbt     string      `xml:"CdtDbt"`
	Sts        string      `xml:"Sts"`
	BkTxCd     BankTransactionCode `xml:"BkTxCd"`
	Svcs       ServiceLevel `xml:"Svcs"`
	TxDt       time.Time   `xml:"TxDt"`
	AcctSvcrRef string     `xml:"AcctSvcrRef"`
	Chrgs      []Charges   `xml:"Chrgs"`
	NtryDtls   []EntryDetail `xml:"NtryDtls"`
}

// BankTransactionCode represents a bank transaction code
type BankTransactionCode struct {
	Prtry string `xml:"Prtry"`
}

// ServiceLevel represents service level
type ServiceLevel struct {
	Cd string `xml:"Cd"`
}

// Charges represents charges
type Charges struct {
	Amt Amount `xml:"Amt"`
}

// EntryDetail represents entry details
type EntryDetail struct {
	TxDtls     []TransactionDetails `xml:"TxDtls"`
	Refs       string               `xml:"Refs"`
	Amt        Amount               `xml:"Amt"`
	CdtDbt     string               `xml:"CdtDbt"`
	Info       string               `xml:"Info"`
}

// TransactionDetails represents transaction details
type TransactionDetails struct {
	Refs       string  `xml:"Refs"`
	Amt        Amount  `xml:"Amt"`
	CdtDbt     string  `xml:"CdtDbt"`
	Chrgs      []Charges `xml:"Chrgs"`
}

// parseCamt053 parses camt.053 bank statement message
func (p *Processor) parseCamt053(decoder *xml.Decoder, startElement *xml.StartElement) (*Camt053Document, error) {
	var doc Camt053Document
	if err := decoder.DecodeElement(&doc, startElement); err != nil {
		return nil, fmt.Errorf("failed to decode camt.053: %w", err)
	}
	return &doc, nil
}

// Pain001Document represents ISO 20022 pain.001 message structure
type Pain001Document struct {
	XMLName     xml.Name           `xml:"Document"`
	Xmlns       string             `xml:"xmlns,attr"`
	CreditInit  CreditInitiation   `xml:"CstmrCdtTrfInitn"`
}

// CreditInitiation represents a customer credit transfer initiation
type CreditInitiation struct {
	GrpHdr      GroupHeader       `xml:"GrpHdr"`
	PmtInf      []PaymentInfo     `xml:"PmtInf"`
}

// PaymentInfo represents payment information
type PaymentInfo struct {
	PmtInfId       string              `xml:"PmtInfId"`
	PmtMtd         string              `xml:"PmtMtd"`
	PmtTpInf       PaymentTypeInfo     `xml:"PmtTpInf"`
	ReqdExctnDt    time.Time           `xml:"ReqdExctnDt"`
	Dbtr           Party               `xml:"Dbtr"`
	DbtrAcct       Account             `xml:"DbtrAcct"`
	DbtrAgt        Agent               `xml:"DbtrAgt"`
	CdtrAgt        Agent               `xml:"CdtrAgt"`
	Cdtr           Party               `xml:"Cdtr"`
	CdtrAcct       Account             `xml:"CdtrAcct"`
	CdtrRefInf     CreditorReferenceInfo `xml:"CdtrRefInf"`
	Purps          Purpose             `xml:"Purp"`
	ChrgsAcc       Account             `xml:"ChrgsAcc"`
	CdtTrfTxInf    []CreditTransferInfo `xml:"CdtTrfTxInf"`
}

// PaymentTypeInfo represents payment type information
type PaymentTypeInfo struct {
	SvcLvl ServiceLevel `xml:"SvcLvl"`
	CtgyPurp string     `xml:"CtgyPurp"`
}

// parsePain001 parses pain.001 credit transfer initiation
func (p *Processor) parsePain001(decoder *xml.Decoder, startElement *xml.StartElement) (*Pain001Document, error) {
	var doc Pain001Document
	if err := decoder.DecodeElement(&doc, startElement); err != nil {
		return nil, fmt.Errorf("failed to decode pain.001: %w", err)
	}
	return &doc, nil
}

// parseGeneric parses a generic ISO 20022 message
func (p *Processor) parseGeneric(decoder *xml.Decoder, startElement *xml.StartElement) (interface{}, error) {
	var generic map[string]interface{}
	if err := decoder.DecodeElement(&generic, startElement); err != nil {
		return nil, fmt.Errorf("failed to decode generic message: %w", err)
	}
	return generic, nil
}

// detectMessageType detects the ISO 20022 message type from raw bytes
func detectMessageType(raw []byte) (string, error) {
	trimmed := bytes.TrimSpace(raw)

	// Check for pacs.008
	if bytes.Contains(trimmed, []byte("FIToFICstmrCdtTrf")) ||
	   bytes.Contains(trimmed, []byte("pacs.008")) {
		return "pacs.008", nil
	}

	// Check for pacs.004
	if bytes.Contains(trimmed, []byte("FIToFIPmtStsRpt")) ||
	   bytes.Contains(trimmed, []byte("pacs.004")) {
		return "pacs.004", nil
	}

	// Check for camt.053
	if bytes.Contains(trimmed, []byte("BkToCstmrStmt")) ||
	   bytes.Contains(trimmed, []byte("camt.053")) {
		return "camt.053", nil
	}

	// Check for camt.052
	if bytes.Contains(trimmed, []byte("BkToCstmrAcctRpt")) ||
	   bytes.Contains(trimmed, []byte("camt.052")) {
		return "camt.052", nil
	}

	// Check for pain.001
	if bytes.Contains(trimmed, []byte("CstmrCdtTrfInitn")) ||
	   bytes.Contains(trimmed, []byte("pain.001")) {
		return "pain.001", nil
	}

	// Check for pain.002
	if bytes.Contains(trimmed, []byte("CstmrPmtStsReq")) ||
	   bytes.Contains(trimmed, []byte("pain.002")) {
		return "pain.002", nil
	}

	return "", fmt.Errorf("unknown ISO 20022 message type")
}

// toNormalizedPayment converts parsed document to normalized payment
func (p *Processor) toNormalizedPayment(doc interface{}, msgType string, raw []byte, startTime time.Time) (*NormalizedPayment, error) {
	normalized := &NormalizedPayment{
		ID:             generatePaymentID(),
		MessageType:    msgType,
		ProcessingDate: startTime,
		Metadata:       make(map[string]interface{}),
	}

	switch d := doc.(type) {
	case *Pacs008Document:
		normalized.TransactionID = d.CreditTransfer.GrpHdr.MsgId
		normalized.SenderBIC = d.CreditTransfer.GrpHdr.InstgAgt.FinInstnId.BIC
		normalized.ReceiverBIC = d.CreditTransfer.GrpHdr.InstdAgt.FinInstnId.BIC
		if len(d.CreditTransfer.CdtTrfTxInf) > 0 {
			tx := d.CreditTransfer.CdtTrfTxInf[0]
			normalized.Amount = parseAmount(tx.Amt.Value)
			normalized.Currency = tx.Amt.Ccy
			normalized.SenderAccount = tx.DbtrAcct.Iban
			normalized.ReceiverAccount = tx.CdtrAcct.Iban
			normalized.TransactionID = tx.PtId
		}
		normalized.RawPayload = p.extractRawPayload(raw)

	case *Pacs004Document:
		normalized.TransactionID = d.Return.GrpHdr.MsgId
		normalized.SenderBIC = d.Return.GrpHdr.InstgAgt.FinInstnId.BIC
		if len(d.Return.TxInfAndSts) > 0 {
			tx := d.Return.TxInfAndSts[0]
			normalized.Amount = parseAmount(tx.OrgnlAmt.Value)
			normalized.Currency = tx.OrgnlCcy
			normalized.Metadata["status"] = tx.TxSts
		}

	case *Camt053Document:
		normalized.TransactionID = d.Stmt[0].GrpHdr.MsgId
		normalized.RawPayload = p.extractRawPayload(raw)
		normalized.Metadata["statement_id"] = d.Stmt[0].Id

	case *Pain001Document:
		normalized.TransactionID = d.CreditInit.GrpHdr.MsgId
		if len(d.CreditInit.PmtInf) > 0 {
			pmt := d.CreditInit.PmtInf[0]
			normalized.SenderBIC = pmt.DbtrAgt.FinInstnId.BIC
			normalized.ReceiverBIC = pmt.CdtrAgt.FinInstnId.BIC
			normalized.SenderAccount = pmt.DbtrAcct.Iban
			normalized.ReceiverAccount = pmt.CdtrAcct.Iban
		}

	default:
		// Generic handling
		normalized.RawPayload = p.extractRawPayload(raw)
	}

	return normalized, nil
}

// extractRawPayload extracts the relevant portion of the raw payload
func (p *Processor) extractRawPayload(raw []byte) json.RawMessage {
	// Return full payload as JSON for reference
	var prettyJSON map[string]interface{}
	_ = json.Unmarshal(raw, &prettyJSON)
	data, _ := json.Marshal(prettyJSON)
	return data
}

// parseAmount parses a decimal amount string
func parseAmount(value float64) decimal.Decimal {
	return decimal.NewFromFloat(value)
}

// generatePaymentID generates a unique payment ID
func generatePaymentID() string {
	hash := sha256.Sum256([]byte(time.Now().String() + "-" + randomString(16)))
	return fmt.Sprintf("PAY-%x", hash[:8])
}

// randomString generates a random string
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}

// Pacs008Validator validates pacs.008 messages
type Pacs008Validator struct{}

func (v *Pacs008Validator) Validate(msg interface{}) error {
	doc, ok := msg.(*Pacs008Document)
	if !ok {
		return fmt.Errorf("invalid message type for pacs.008 validation")
	}

	if doc.CreditTransfer.GrpHdr.MsgId == "" {
		return fmt.Errorf("missing message ID")
	}

	if doc.CreditTransfer.GrpHdr.CreDtTm.IsZero() {
		return fmt.Errorf("missing creation timestamp")
	}

	for _, tx := range doc.CreditTransfer.CdtTrfTxInf {
		if tx.Amt.Ccy == "" {
			return fmt.Errorf("missing currency code")
		}
	}

	return nil
}

// Pacs004Validator validates pacs.004 messages
type Pacs004Validator struct{}

func (v *Pacs004Validator) Validate(msg interface{}) error {
	doc, ok := msg.(*Pacs004Document)
	if !ok {
		return fmt.Errorf("invalid message type for pacs.004 validation")
	}

	if doc.Return.GrpHdr.MsgId == "" {
		return fmt.Errorf("missing message ID")
	}

	return nil
}

// Camt053Validator validates camt.053 messages
type Camt053Validator struct{}

func (v *Camt053Validator) Validate(msg interface{}) error {
	doc, ok := msg.(*Camt053Document)
	if !ok {
		return fmt.Errorf("invalid message type for camt.053 validation")
	}

	if doc.Stmt[0].GrpHdr.MsgId == "" {
		return fmt.Errorf("missing statement ID")
	}

	return nil
}

// Pain001Validator validates pain.001 messages
type Pain001Validator struct{}

func (v *Pain001Validator) Validate(msg interface{}) error {
	doc, ok := msg.(*Pain001Document)
	if !ok {
		return fmt.Errorf("invalid message type for pain.001 validation")
	}

	if doc.CreditInit.GrpHdr.MsgId == "" {
		return fmt.Errorf("missing message ID")
	}

	return nil
}

// Import for additional dependencies
import (
	"math/big"
)

// Decimal is a custom decimal type for financial calculations
type Decimal big.Int

// String returns the string representation
func (d Decimal) String() string {
	return (*big.Int)(&d).String()
}
