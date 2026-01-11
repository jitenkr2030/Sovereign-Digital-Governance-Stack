// Validation Utilities - Data validation for CSIC Platform
// Input validation, sanitization, and schema validation

package validation

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"

	"go.uber.org/zap"
)

// Validator provides validation methods
type Validator struct {
	logger *zap.Logger
}

// NewValidator creates a new Validator
func NewValidator(logger *zap.Logger) *Validator {
	return &Validator{logger: logger}
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}

// ValidationErrors is a collection of validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}
	messages := make([]string, len(e))
	for i, err := range e {
		messages[i] = fmt.Sprintf("%s: %s", err.Field, err.Message)
	}
	return strings.Join(messages, "; ")
}

// IsValid checks if there are no validation errors
func (e ValidationErrors) IsValid() bool {
	return len(e) == 0
}

// ValidatorFunc is a function that validates a field
type ValidatorFunc func(value interface{}) *ValidationError

// FieldValidator validates a specific field
type FieldValidator struct {
	Name       string
	Validators []ValidatorFunc
}

// Validate validates a value against all validators
func (v *FieldValidator) Validate(value interface{}) *ValidationError {
	for _, validator := range v.Validators {
		if err := validator(value); err != nil {
			return err
		}
	}
	return nil
}

// Required validates that a value is not empty
func Required(fieldName string) ValidatorFunc {
	return func(value interface{}) *ValidationError {
		if value == nil {
			return &ValidationError{
				Field:   fieldName,
				Message: "is required",
				Value:   value,
			}
		}
		if str, ok := value.(string); ok {
			if strings.TrimSpace(str) == "" {
				return &ValidationError{
					Field:   fieldName,
					Message: "is required",
					Value:   value,
				}
			}
		}
		if slice, ok := value.([]interface{}); ok {
			if len(slice) == 0 {
				return &ValidationError{
					Field:   fieldName,
					Message: "is required",
					Value:   value,
				}
			}
		}
		return nil
	}
}

// MinLength validates minimum string length
func MinLength(fieldName string, minLen int) ValidatorFunc {
	return func(value interface{}) *ValidationError {
		str, ok := value.(string)
		if !ok {
			return &ValidationError{
				Field:   fieldName,
				Message: "must be a string",
				Value:   value,
			}
		}
		if utf8.RuneCountInString(str) < minLen {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("must be at least %d characters", minLen),
				Value:   value,
			}
		}
		return nil
	}
}

// MaxLength validates maximum string length
func MaxLength(fieldName string, maxLen int) ValidatorFunc {
	return func(value interface{}) *ValidationError {
		str, ok := value.(string)
		if !ok {
			return &ValidationError{
				Field:   fieldName,
				Message: "must be a string",
				Value:   value,
			}
		}
		if utf8.RuneCountInString(str) > maxLen {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("must be at most %d characters", maxLen),
				Value:   value,
			}
		}
		return nil
	}
}

// Email validates email format
func Email(fieldName string) ValidatorFunc {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return func(value interface{}) *ValidationError {
		str, ok := value.(string)
		if !ok {
			return &ValidationError{
				Field:   fieldName,
				Message: "must be a string",
				Value:   value,
			}
		}
		if !emailRegex.MatchString(str) {
			return &ValidationError{
				Field:   fieldName,
				Message: "must be a valid email address",
				Value:   value,
			}
		}
		return nil
	}
}

// UUID validates UUID format
func UUID(fieldName string) ValidatorFunc {
	uuidRegex := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	return func(value interface{}) *ValidationError {
		str, ok := value.(string)
		if !ok {
			return &ValidationError{
				Field:   fieldName,
				Message: "must be a string",
				Value:   value,
			}
		}
		if !uuidRegex.MatchString(str) {
			return &ValidationError{
				Field:   fieldName,
				Message: "must be a valid UUID",
				Value:   value,
			}
		}
		return nil
	}
}

// Numeric validates that a value is numeric
func Numeric(fieldName string) ValidatorFunc {
	return func(value interface{}) *ValidationError {
		switch value.(type) {
		case int, int8, int16, int32, int64:
			return nil
		case float32, float64:
			return nil
		case string:
			str := value.(string)
			if str == "" {
				return nil
			}
			_, err := fmt.Sscanf(str, "%f")
			if err != nil {
				return &ValidationError{
					Field:   fieldName,
					Message: "must be a numeric value",
					Value:   value,
				}
			}
			return nil
		default:
			return &ValidationError{
				Field:   fieldName,
				Message: "must be a numeric value",
				Value:   value,
			}
		}
	}
}

// MinValue validates minimum numeric value
func MinValue(fieldName string, min float64) ValidatorFunc {
	return func(value interface{}) *ValidationError {
		var val float64
		switch v := value.(type) {
		case int:
			val = float64(v)
		case int64:
			val = float64(v)
		case float32:
			val = float64(v)
		case float64:
			val = v
		case string:
			if v == "" {
				return nil
			}
			parsed, err := fmt.Sscanf(v, "%f")
			if err != nil || parsed != 1 {
				return &ValidationError{
					Field:   fieldName,
					Message: "must be a numeric value",
					Value:   value,
				}
			}
			val = parsed
		default:
			return &ValidationError{
				Field:   fieldName,
				Message: "must be a numeric value",
				Value:   value,
			}
		}
		if val < min {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("must be at least %f", min),
				Value:   value,
			}
		}
		return nil
	}
}

// MaxValue validates maximum numeric value
func MaxValue(fieldName string, max float64) ValidatorFunc {
	return func(value interface{}) *ValidationError {
		var val float64
		switch v := value.(type) {
		case int:
			val = float64(v)
		case int64:
			val = float64(v)
		case float32:
			val = float64(v)
		case float64:
			val = v
		case string:
			if v == "" {
				return nil
			}
			parsed, err := fmt.Sscanf(v, "%f")
			if err != nil || parsed != 1 {
				return &ValidationError{
					Field:   fieldName,
					Message: "must be a numeric value",
					Value:   value,
				}
			}
			val = parsed
		default:
			return &ValidationError{
				Field:   fieldName,
				Message: "must be a numeric value",
				Value:   value,
			}
		}
		if val > max {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("must be at most %f", max),
				Value:   value,
			}
		}
		return nil
	}
}

// InList validates that a value is in a list
func InList(fieldName string, allowed []interface{}) ValidatorFunc {
	allowedSet := make(map[interface{}]bool)
	for _, a := range allowed {
		allowedSet[a] = true
	}
	return func(value interface{}) *ValidationError {
		if !allowedSet[value] {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("must be one of: %v", allowed),
				Value:   value,
			}
		}
		return nil
	}
}

// IPAddress validates IP address format
func IPAddress(fieldName string) ValidatorFunc {
	return func(value interface{}) *ValidationError {
		str, ok := value.(string)
		if !ok {
			return &ValidationError{
				Field:   fieldName,
				Message: "must be a string",
				Value:   value,
			}
		}
		if net.ParseIP(str) == nil {
			return &ValidationError{
				Field:   fieldName,
				Message: "must be a valid IP address",
				Value:   value,
			}
		}
		return nil
	}
}

// URL validates URL format
func URL(fieldName string) ValidatorFunc {
	return func(value interface{}) *ValidationError {
		str, ok := value.(string)
		if !ok {
			return &ValidationError{
				Field:   fieldName,
				Message: "must be a string",
				Value:   value,
			}
		}
		_, err := url.Parse(str)
		if err != nil {
			return &ValidationError{
				Field:   fieldName,
				Message: "must be a valid URL",
				Value:   value,
			}
		}
		return nil
	}
}

// Regex validates against a regular expression
func Regex(fieldName string, pattern string) ValidatorFunc {
	re, err := regexp.Compile(pattern)
	if err != nil {
		panic(fmt.Sprintf("invalid regex pattern: %s", pattern))
	}
	return func(value interface{}) *ValidationError {
		str, ok := value.(string)
		if !ok {
			return &ValidationError{
				Field:   fieldName,
				Message: "must be a string",
				Value:   value,
			}
		}
		if !re.MatchString(str) {
			return &ValidationError{
				Field:   fieldName,
				Message: "does not match required pattern",
				Value:   value,
			}
		}
		return nil
	}
}

// CryptoAddress validates cryptocurrency address format
func CryptoAddress(fieldName string, chain string) ValidatorFunc {
	return func(value interface{}) *ValidationError {
		str, ok := value.(string)
		if !ok {
			return &ValidationError{
				Field:   fieldName,
				Message: "must be a string",
				Value:   value,
			}
		}

		switch strings.ToUpper(chain) {
		case "BTC", "BITCOIN":
			// Bitcoin address validation (simplified)
			if len(str) < 26 || len(str) > 35 {
				return &ValidationError{
					Field:   fieldName,
					Message: "invalid Bitcoin address length",
					Value:   value,
				}
			}
		case "ETH", "ETHEREUM":
			// Ethereum address validation
			if !strings.HasPrefix(str, "0x") || len(str) != 42 {
				return &ValidationError{
					Field:   fieldName,
					Message: "invalid Ethereum address",
					Value:   value,
				}
			}
		}
		return nil
	}
}

// SanitizeString removes potentially dangerous characters
func SanitizeString(str string) string {
	// Remove null bytes
	str = strings.ReplaceAll(str, "\x00", "")
	// Trim whitespace
	str = strings.TrimSpace(str)
	// Limit length
	if len(str) > 10000 {
		str = str[:10000]
	}
	return str
}

// ValidateStruct validates a struct using field validators
func ValidateStruct(validators map[string]*FieldValidator, data map[string]interface{}) ValidationErrors {
	var errors ValidationErrors

	for fieldName, validator := range validators {
		value, exists := data[fieldName]
		if !exists {
			// Check if field is required
			for _, v := range validator.Validators {
				err := v(nil)
				if err != nil != (value == nil) {
					errors = append(errors, *err)
					break
				}
			}
			continue
		}

		if err := validator.Validate(value); err != nil {
			errors = append(errors, *err)
		}
	}

	return errors
}
