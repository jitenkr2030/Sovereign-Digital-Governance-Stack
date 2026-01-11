package resilience

import (
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// MessageSerializer handles serialization/deserialization of messages
type MessageSerializer struct {
	logger *zap.Logger
}

// NewMessageSerializer creates a new message serializer
func NewMessageSerializer(logger *zap.Logger) *MessageSerializer {
	return &MessageSerializer{
		logger: logger,
	}
}

// Serialize serializes a DLQ message to JSON
func (s *MessageSerializer) Serialize(msg *DLQMessage) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		s.logger.Error("Failed to serialize DLQ message",
			zap.Error(err),
			zap.String("id", msg.ID))
		return nil, fmt.Errorf("failed to serialize DLQ message: %w", err)
	}
	return data, nil
}

// DeserializeDLQ deserializes a DLQ message from JSON
func (s *MessageSerializer) DeserializeDLQ(data []byte) (*DLQMessage, error) {
	var msg DLQMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		s.logger.Error("Failed to deserialize DLQ message",
			zap.Error(err))
		return nil, fmt.Errorf("failed to deserialize DLQ message: %w", err)
	}
	return &msg, nil
}

// DeserializePayload deserializes the original payload
func (s *MessageSerializer) DeserializePayload(data json.RawMessage, target interface{}) error {
	if err := json.Unmarshal(data, target); err != nil {
		s.logger.Error("Failed to deserialize payload",
			zap.Error(err))
		return fmt.Errorf("failed to deserialize payload: %w", err)
	}
	return nil
}
