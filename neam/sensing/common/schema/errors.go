package schema

import (
	"encoding/json"
	"fmt"
	"time"
)

// ErrorCategory categorizes validation errors
type ErrorCategory string

const (
	ErrorCategoryRequired    ErrorCategory = "required"
	ErrorCategoryType        ErrorCategory = "type"
	ErrorCategoryFormat      ErrorCategory = "format"
	ErrorCategoryRange       ErrorCategory = "range"
	ErrorCategoryPattern     ErrorCategory = "pattern"
	ErrorCategoryEnum        ErrorCategory = "enum"
	ErrorCategoryCustom      ErrorCategory = "custom"
	ErrorCategoryParse       ErrorCategory = "parse"
	ErrorCategoryConversion  ErrorCategory = "conversion"
	ErrorCategoryUnknown     ErrorCategory = "unknown"
)

// Severity represents the severity level of a validation error
type Severity string

const (
	SeverityCritical   Severity = "critical"
	SeverityWarning    Severity = "warning"
	SeverityInfo       Severity = "info"
)

// ErrorDetail provides detailed information about a validation error
type ErrorDetail struct {
	ID          string        `json:"id"`
	Timestamp   time.Time     `json:"timestamp"`
	SchemaID    string        `json:"schema_id"`
	SchemaVer   string        `json:"schema_version"`
	Field       string        `json:"field"`
	Path        string        `json:"path"`
	Message     string        `json:"message"`
	Category    ErrorCategory `json:"category"`
	Severity    Severity      `json:"severity"`
	Expected    string        `json:"expected,omitempty"`
	Actual      string        `json:"actual,omitempty"`
	Constraint  string        `json:"constraint,omitempty"`
	Context     string        `json:"context,omitempty"`
	RequestID   string        `json:"request_id,omitempty"`
	SourceID    string        `json:"source_id,omitempty"`
	 Remediation string        `json:"remediation,omitempty"`
}

// ErrorReport contains a complete validation error report
type ErrorReport struct {
	ReportID      string        `json:"report_id"`
	Timestamp     time.Time     `json:"timestamp"`
	SchemaID      string        `json:"schema_id"`
	SchemaVersion string        `json:"schema_version"`
	TotalErrors   int           `json:"total_errors"`
	TotalWarnings int           `json:"total_warnings"`
	Errors        []ErrorDetail `json:"errors"`
	Summary       string        `json:"summary"`
}

// NewErrorDetail creates a new error detail from a validation error
func NewErrorDetail(err ValidationError, schemaID, schemaVersion string) *ErrorDetail {
	return &ErrorDetail{
		ID:        generateErrorID(),
		Timestamp: time.Now(),
		SchemaID:  schemaID,
		SchemaVer: schemaVersion,
		Field:     err.Field,
		Message:   err.Message,
		Category:  categorizeError(err),
		Severity:  determineSeverity(err),
		Expected:  err.Expected,
		Actual:    err.Actual,
		Constraint: err.Constraint,
	}
}

// categorizeError categorizes a validation error
func categorizeError(err ValidationError) ErrorCategory {
	switch err.ErrorType {
	case "required":
		return ErrorCategoryRequired
	case "type":
		return ErrorCategoryType
	case "format":
		return ErrorCategoryFormat
	case "minimum", "maximum", "exclusiveMinimum", "exclusiveMaximum",
		"multipleOf", "minLength", "maxLength", "minItems", "maxItems":
		return ErrorCategoryRange
	case "pattern":
		return ErrorCategoryPattern
	case "enum":
		return ErrorCategoryEnum
	case "custom_rule":
		return ErrorCategoryCustom
	case "parse_error":
		return ErrorCategoryParse
	case "conversion_error":
		return ErrorCategoryConversion
	default:
		return ErrorCategoryUnknown
	}
}

// determineSeverity determines the severity of a validation error
func determineSeverity(err ValidationError) Severity {
	switch err.ErrorType {
	case "required":
		return SeverityCritical
	case "minimum", "maximum":
		// Critical if out of range
		return SeverityCritical
	case "type":
		return SeverityWarning
	default:
		return SeverityWarning
	}
}

// GenerateErrorReport creates a complete error report from validation results
func GenerateErrorReport(result *ValidationResult, schemaID, schemaVersion string) *ErrorReport {
	report := &ErrorReport{
		ReportID:      generateReportID(),
		Timestamp:     time.Now(),
		SchemaID:      schemaID,
		SchemaVersion: schemaVersion,
		Errors:        make([]ErrorDetail, 0, len(result.Errors)),
	}

	for _, err := range result.Errors {
		detail := NewErrorDetail(err, schemaID, schemaVersion)
		report.Errors = append(report.Errors, *detail)
		
		if detail.Severity == SeverityCritical {
			report.TotalErrors++
		} else {
			report.TotalWarnings++
		}
	}

	report.Summary = generateSummary(report)

	return report
}

// generateErrorID generates a unique error ID
func generateErrorID() string {
	return fmt.Sprintf("err_%d", time.Now().UnixNano())
}

// generateReportID generates a unique report ID
func generateReportID() string {
	return fmt.Sprintf("rpt_%d", time.Now().UnixNano())
}

// generateSummary generates a human-readable summary
func generateSummary(report *ErrorReport) string {
	if len(report.Errors) == 0 {
		return "No validation errors found"
	}

	return fmt.Sprintf("Validation failed with %d error(s) and %d warning(s)",
		report.TotalErrors, report.TotalWarnings)
}

// ToJSON converts the error report to JSON
func (e *ErrorReport) ToJSON() ([]byte, error) {
	return json.MarshalIndent(e, "", "  ")
}

// ToPrettyString returns a formatted string representation
func (e *ErrorReport) ToPrettyString() string {
	lines := []string{
		fmt.Sprintf("Error Report: %s", e.ReportID),
		fmt.Sprintf("Schema: %s (v%s)", e.SchemaID, e.SchemaVersion),
		fmt.Sprintf("Timestamp: %s", e.Timestamp.Format(time.RFC3339)),
		fmt.Sprintf(""),
		fmt.Sprintf("Summary: %s", e.Summary),
		fmt.Sprintf(""),
		fmt.Sprintf("Details:"),
	}

	for i, err := range e.Errors {
		lines = append(lines, fmt.Sprintf("  [%d] %s: %s", i+1, err.Field, err.Message))
		if err.Expected != "" {
			lines = append(lines, fmt.Sprintf("      Expected: %s", err.Expected))
		}
		if err.Actual != "" {
			lines = append(lines, fmt.Sprintf("      Actual: %s", err.Actual))
		}
		lines = append(lines, fmt.Sprintf("      Severity: %s", err.Severity))
		lines = append(lines, "")
	}

	return joinStrings(lines, "\n")
}

// joinStrings joins strings with a separator
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

// AggregateErrors aggregates errors by field
func AggregateErrors(errors []ValidationError) map[string][]ValidationError {
	result := make(map[string][]ValidationError)
	for _, err := range errors {
		result[err.Field] = append(result[err.Field], err)
	}
	return result
}

// AggregateErrorsByCategory aggregates errors by category
func AggregateErrorsByCategory(errors []ValidationError) map[ErrorCategory][]ValidationError {
	result := make(map[ErrorCategory][]ValidationError)
	for _, err := range errors {
		result[categorizeError(err)] = append(result[categorizeError(err)], err)
	}
	return result
}

// GetMostCommonField returns the field with the most errors
func GetMostCommonField(errors []ValidationError) string {
	if len(errors) == 0 {
		return ""
	}

	counts := make(map[string]int)
	maxCount := 0
	mostCommon := ""

	for _, err := range errors {
		counts[err.Field]++
		if counts[err.Field] > maxCount {
			maxCount = counts[err.Field]
			mostCommon = err.Field
		}
	}

	return mostCommon
}
