package schema

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
)

// ProtobufValidator implements Validator interface for Protobuf schema validation
type ProtobufValidator struct {
	schemaID      string
	schemaVersion string
	messageDesc   protoreflect.MessageDescriptor
	config        ValidatorConfig
	logger        *zap.Logger
}

// NewProtobufValidator creates a new Protobuf schema validator
func NewProtobufValidator(schemaData json.RawMessage) (*ProtobufValidator, error) {
	// Parse the schema definition from JSON
	// In a real implementation, you would parse .proto file or use proto registry
	var schemaDef struct {
		MessageName string `json:"message_name"`
		Fields      []struct {
			Name    string `json:"name"`
			Number  int32  `json:"number"`
			Type    string `json:"type"`
			Label   string `json:"label"`
		} `json:"fields"`
	}

	if err := json.Unmarshal(schemaData, &schemaDef); err != nil {
		return nil, fmt.Errorf("failed to parse protobuf schema: %w", err)
	}

	// Create a dynamic message descriptor
	msgDesc := createDynamicMessageDescriptor(schemaDef)

	return &ProtobufValidator{
		schemaID:      "",
		schemaVersion: "1.0.0",
		messageDesc:   msgDesc,
		config: ValidatorConfig{
			Mode:         ValidationModeStrict,
			CacheResults: false,
		},
		logger: nil,
	}, nil
}

// NewProtobufValidatorWithConfig creates a new Protobuf validator with configuration
func NewProtobufValidatorWithConfig(schemaID, schemaVersion string, schemaData json.RawMessage, config ValidatorConfig, logger *zap.Logger) (*ProtobufValidator, error) {
	validator, err := NewProtobufValidator(schemaData)
	if err != nil {
		return nil, err
	}

	validator.schemaID = schemaID
	validator.schemaVersion = schemaVersion
	validator.config = config
	validator.logger = logger

	return validator, nil
}

// Validate validates the provided Protobuf data against the schema
func (v *ProtobufValidator) Validate(ctx context.Context, data json.RawMessage) (*ValidationResult, error) {
	startTime := time.Now()

	var errors []ValidationError

	// First, try to unmarshal as JSON to the proto message
	var jsonData map[string]interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return &ValidationResult{
			Valid:     false,
			Errors:    []ValidationError{{Field: "root", Message: "invalid JSON format", ErrorType: "parse_error", Actual: err.Error()}},
			Timestamp: startTime,
			Duration:  time.Since(startTime),
		}, nil
	}

	// Create a dynamic message instance
	msg := dynamicpb.NewMessage(v.messageDesc)

	// Convert JSON to proto
	if err := convertJSONToProto(jsonData, msg); err != nil {
		return &ValidationResult{
			Valid:     false,
			Errors:    []ValidationError{{Field: "root", Message: "JSON to proto conversion failed", ErrorType: "conversion_error", Actual: err.Error()}},
			Timestamp: startTime,
			Duration:  time.Since(startTime),
		}, nil
	}

	// Validate the proto message
	errors = v.validateProtoMessage(msg)

	// Apply custom rules
	if len(errors) == 0 && len(v.config.CustomRules) > 0 {
		customErrors := v.applyCustomRules(msg)
		errors = append(errors, customErrors...)
	}

	result := &ValidationResult{
		Valid:     len(errors) == 0,
		Errors:    errors,
		Timestamp: startTime,
		Duration:  time.Since(startTime),
	}

	return result, nil
}

// ValidateWithContext validates with additional context
func (v *ProtobufValidator) ValidateWithContext(ctx context.Context, data json.RawMessage, metadata map[string]interface{}) (*ValidationResult, error) {
	result, err := v.Validate(ctx, data)
	if result != nil {
		result.Metadata = metadata
	}
	return result, err
}

// validateProtoMessage validates a proto message against its descriptor
func (v *ProtobufValidator) validateProtoMessage(msg protoreflect.Message) []ValidationError {
	var errors []ValidationError

	// Iterate over all fields
	msgDesc := msg.Descriptor()
	for i := 0; i < msgDesc.Fields().Len(); i++ {
		fieldDesc := msgDesc.Fields().Get(i)
		fieldValue := msg.Get(fieldDesc)

		// Check required fields
		if fieldDesc.HasPresence() && fieldDesc.IsRequired() {
			if fieldValue.IsZero() {
				errors = append(errors, ValidationError{
					Field:     string(fieldDesc.Name()),
					Message:   fmt.Sprintf("required field '%s' is missing", fieldDesc.Name()),
					ErrorType: "required",
				})
				continue
			}
		}

		// Validate based on field type
		fieldErrors := v.validateField(fieldDesc, fieldValue)
		errors = append(errors, fieldErrors...)
	}

	return errors
}

// validateField validates a single proto field
func (v *ProtobufValidator) validateField(fieldDesc protoreflect.FieldDescriptor, value protoreflect.Value) []ValidationError {
	var errors []ValidationError

	switch fieldDesc.Kind() {
	case protoreflect.StringKind:
		s := value.String()
		if fieldDesc.String().Len() > 0 {
			maxLen := fieldDesc.GetOptions().GetDefault()
			if maxLen > 0 && len(s) > int(maxLen) {
				errors = append(errors, ValidationError{
					Field:     string(fieldDesc.Name()),
					Message:   fmt.Sprintf("string length %d exceeds maximum %d", len(s), int(maxLen)),
					ErrorType: "max_length",
					Constraint: "max_length",
				})
			}
		}
	case protoreflect.BytesKind:
		b := value.Bytes()
		if fieldDesc.Bytes().Len() > 0 {
			maxLen := fieldDesc.GetOptions().GetDefault()
			if maxLen > 0 && len(b) > int(maxLen) {
				errors = append(errors, ValidationError{
					Field:     string(fieldDesc.Name()),
					Message:   fmt.Sprintf("bytes length %d exceeds maximum %d", len(b), int(maxLen)),
					ErrorType: "max_length",
					Constraint: "max_length",
				})
			}
		}
	case protoreflect.Int32Kind, protoreflect.Int64Kind,
		protoreflect.Sint32Kind, protoreflect.Sint64Kind,
		protoreflect.Uint32Kind, protoreflect.Uint64Kind:
		// Check range constraints from field options
		num := value.Int()
		// In a real implementation, you would check against field options for min/max
		_ = num
	case protoreflect.EnumKind:
		// Check if enum value is valid
		enumValue := value.Enum()
		enumDesc := fieldDesc.Enum()
		valid := false
		for i := 0; i < enumDesc.Values().Len(); i++ {
			if enumDesc.Values().Get(i).Number() == enumValue {
				valid = true
				break
			}
		}
		if !valid {
			errors = append(errors, ValidationError{
				Field:     string(fieldDesc.Name()),
				Message:   fmt.Sprintf("invalid enum value: %d", enumValue),
				ErrorType: "enum",
				Expected:  fmt.Sprintf("one of %v", enumDesc.Values()),
			})
		}
	case protoreflect.MessageKind:
		// Recursively validate nested messages
		if !value.Message().IsValid() {
			errors = append(errors, ValidationError{
				Field:     string(fieldDesc.Name()),
				Message:   "nested message is invalid",
				ErrorType: "message",
			})
		}
	case protoreflect.RepeatedKind:
		// Validate repeated fields
		list := value.List()
		for i := 0; i < list.Len(); i++ {
			elem := list.Get(i)
			if fieldDesc.Message() != nil {
				if !elem.IsValid() {
					errors = append(errors, ValidationError{
						Field:     fmt.Sprintf("%s[%d]", fieldDesc.Name(), i),
						Message:   "repeated message element is invalid",
						ErrorType: "message",
					})
				}
			}
		}
	}

	return errors
}

// applyCustomRules applies custom validation rules to a proto message
func (v *ProtobufValidator) applyCustomRules(msg protoreflect.Message) []ValidationError {
	var errors []ValidationError

	for _, rule := range v.config.CustomRules {
		if err := v.evaluateCustomRule(msg, rule); err != nil {
			errors = append(errors, *err)
		}
	}

	return errors
}

// evaluateCustomRule evaluates a custom rule against a proto message
func (v *ProtobufValidator) evaluateCustomRule(msg protoreflect.Message, rule CustomRule) *ValidationError {
	fieldDesc := msg.Descriptor().Fields().ByName(protoreflect.Name(rule.Field))
	if fieldDesc == nil {
		// Field not found in message
		return nil
	}

	value := msg.Get(fieldDesc)

	// Evaluate predicate based on field type
	switch rule.Predicate {
	case "not_empty":
		if value.IsZero() {
			return &ValidationError{
				Field:     rule.Field,
				Message:   fmt.Sprintf("field is required: %s", rule.Message),
				ErrorType: "required",
			}
		}
	case "positive":
		if value.Int() <= 0 {
			return &ValidationError{
				Field:     rule.Field,
				Message:   fmt.Sprintf("field must be positive: %s", rule.Message),
				ErrorType: "constraint",
			}
		}
	case "in_range":
		min := float64(0)
		max := float64(100)
		if pMin, ok := rule.Parameters["min"]; ok {
			fmt.Sscanf(pMin, "%f", &min)
		}
		if pMax, ok := rule.Parameters["max"]; ok {
			fmt.Sscanf(pMax, "%f", &max)
		}
		val := value.Float()
		if val < min || val > max {
			return &ValidationError{
				Field:     rule.Field,
				Message:   fmt.Sprintf("value %f out of range [%f, %f]: %s", val, min, max, rule.Message),
				ErrorType: "range",
			}
		}
	}

	return nil
}

// convertJSONToProto converts JSON data to a dynamic proto message
func convertJSONToProto(jsonData map[string]interface{}, msg *dynamicpb.Message) error {
	msgDesc := msg.Descriptor()

	for key, value := range jsonData {
		fieldDesc := msgDesc.Fields().ByName(protoreflect.Name(key))
		if fieldDesc == nil {
			continue // Skip unknown fields
		}

		switch v := value.(type) {
		case bool:
			msg.Set(fieldDesc, protoreflect.ValueOfBool(v))
		case float64:
			setNumericField(msg, fieldDesc, v)
		case string:
			msg.Set(fieldDesc, protoreflect.ValueOfString(v))
		case []interface{}:
			if fieldDesc.IsList() {
				list := msg.Mutable(fieldDesc).List()
				for _, elem := range v {
					list.Append(convertToProtoValue(elem, fieldDesc.Message()))
				}
			} else if fieldDesc.IsMap() {
				// Handle map fields
				mapMsg := msg.Mutable(fieldDesc).Map()
				for _, elem := range v {
					if elemMap, ok := elem.(map[string]interface{}); ok {
						var mapKey protoreflect.Value
						var mapVal protoreflect.Value
						for k, val := range elemMap {
							if kd := fieldDesc.MapKey(); string(kd.Name()) == k {
								mapKey = convertToProtoValue(k, nil)
							} else {
								mapVal = convertToProtoValue(val, fieldDesc.Message().Elements())
							}
						}
						mapMsg.Set(mapKey, mapVal)
					}
				}
			}
		case map[string]interface{}:
			if fieldDesc.Message() != nil {
				nestedMsg := dynamicpb.NewMessage(fieldDesc.Message())
				if err := convertJSONToProto(v, nestedMsg); err == nil {
					msg.Set(fieldDesc, protoreflect.ValueOfMessage(nestedMsg))
				}
			}
		case nil:
			// Skip null values
		}
	}

	return nil
}

// setNumericField sets a numeric field based on its type
func setNumericField(msg *dynamicpb.Message, fieldDesc protoreflect.FieldDescriptor, value float64) {
	switch fieldDesc.Kind() {
	case protoreflect.Int32Kind, protoreflect.Sint32Kind:
		msg.Set(fieldDesc, protoreflect.ValueOfInt32(int32(value)))
	case protoreflect.Int64Kind, protoreflect.Sint64Kind:
		msg.Set(fieldDesc, protoreflect.ValueOfInt64(int64(value)))
	case protoreflect.Uint32Kind:
		msg.Set(fieldDesc, protoreflect.ValueOfUint32(uint32(value)))
	case protoreflect.Uint64Kind:
		msg.Set(fieldDesc, protoreflect.ValueOfUint64(uint64(value)))
	case protoreflect.FloatKind:
		msg.Set(fieldDesc, protoreflect.ValueOfFloat32(float32(value)))
	case protoreflect.DoubleKind:
		msg.Set(fieldDesc, protoreflect.ValueOfFloat64(value))
	default:
		msg.Set(fieldDesc, protoreflect.ValueOfFloat64(value))
	}
}

// convertToProtoValue converts a JSON value to a proto value
func convertToProtoValue(value interface{}, msgDesc protoreflect.MessageDescriptor) protoreflect.Value {
	switch v := value.(type) {
	case bool:
		return protoreflect.ValueOfBool(v)
	case float64:
		return protoreflect.ValueOfFloat64(v)
	case string:
		return protoreflect.ValueOfString(v)
	case []interface{}:
		// Handle array conversion
		list := make([]protoreflect.Value, len(v))
		for i, elem := range v {
			list[i] = convertToProtoValue(elem, msgDesc)
		}
		return protoreflect.ValueOfList(nil) // Simplified
	case map[string]interface{}:
		if msgDesc != nil {
			nestedMsg := dynamicpb.NewMessage(msgDesc)
			convertJSONToProto(v, nestedMsg)
			return protoreflect.ValueOfMessage(nestedMsg)
		}
	}
	return protoreflect.Value{}
}

// createDynamicMessageDescriptor creates a dynamic message descriptor from schema definition
func createDynamicMessageDescriptor(schemaDef interface{}) protoreflect.MessageDescriptor {
	// This is a simplified implementation
	// In a real implementation, you would parse .proto files or use proto registry
	return nil // Placeholder
}

// GetSchemaID returns the schema identifier
func (v *ProtobufValidator) GetSchemaID() string {
	return v.schemaID
}

// GetSchemaVersion returns the schema version
func (v *ProtobufValidator) GetSchemaVersion() string {
	return v.schemaVersion
}

// GetSchemaType returns the type of schema
func (v *ProtobufValidator) GetSchemaType() SchemaType {
	return SchemaTypeProtobuf
}
