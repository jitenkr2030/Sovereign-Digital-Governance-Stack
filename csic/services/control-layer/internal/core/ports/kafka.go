package ports

import (
	"context"

	"github.com/csic-platform/internal/core/domain"
	"github.com/segmentio/kafka-go"
)

// KafkaProducer defines the interface for producing messages to Kafka topics.
// This interface provides methods for sending messages to Kafka with various
// delivery guarantees and partitioning strategies.
type KafkaProducer interface {
	// Send sends a single message to a Kafka topic
	Send(ctx context.Context, topic string, message *kafka.Message) error

	// SendBatch sends multiple messages to a Kafka topic
	SendBatch(ctx context.Context, topic string, messages []*kafka.Message) error

	// SendKeyed sends a message with a specific key for partitioning
	SendKeyed(ctx context.Context, topic string, key []byte, value []byte, headers []kafka.Header) error

	// SendOrdered sends messages that must be processed in order
	SendOrdered(ctx context.Context, topic string, messages []*kafka.Message) error

	// SendAsync sends a message asynchronously
	SendAsync(ctx context.Context, topic string, message *kafka.Message, callback func(*kafka.Message, error)) error

	// Flush flushes all pending messages
	Flush(ctx context.Context, timeout time.Duration) error

	// Close closes the producer
	Close() error

	// GetMetrics returns producer metrics
	GetMetrics(ctx context.Context) (*domain.KafkaProducerMetrics, error)
}

// KafkaConsumer defines the interface for consuming messages from Kafka topics.
// This interface provides methods for consuming messages with various
// offset management strategies and delivery guarantees.
type KafkaConsumer interface {
	// Subscribe subscribes to one or more topics
	Subscribe(ctx context.Context, topics []string, rebalanceCallback kafka.RebalanceCallback) error

	// SubscribePattern subscribes to topics matching a pattern
	SubscribePattern(ctx context.Context, pattern string, rebalanceCallback kafka.RebalanceCallback) error

	// Unsubscribe unsubscribes from all topics
	Unsubscribe(ctx context.Context) error

	// Read reads a single message
	Read(ctx context.Context) (*kafka.Message, error)

	// ReadBatch reads a batch of messages
	ReadBatch(ctx context.Context, minBytes, maxBytes int) *kafka.Reader

	// Commit commits the current offset
	Commit(ctx context.Context, message *kafka.Message) error

	// CommitOffsets commits specific offsets
	CommitOffsets(ctx context.Context, offsets []kafka.Offset) error

	// Seek seeks to a specific offset
	Seek(ctx context.Context, partition kafka.Partition, offset int64) error

	// Close closes the consumer
	Close() error

	// GetMetrics returns consumer metrics
	GetMetrics(ctx context.Context) (*domain.KafkaConsumerMetrics, error)

	// Pause pauses consuming from specified partitions
	Pause(partitions []kafka.Partition) error

	// Resume resumes consuming from specified partitions
	Resume(partitions []kafka.Partition) error
}

// KafkaAdminClient defines the interface for Kafka administrative operations.
// This interface provides methods for managing topics, partitions, and
// other Kafka cluster resources.
type KafkaAdminClient interface {
	// CreateTopic creates a new topic
	CreateTopic(ctx context.Context, topic string, numPartitions, replicationFactor int) error

	// CreateTopicWithConfig creates a topic with custom configuration
	CreateTopicWithConfig(ctx context.Context, topic string, config *kafka.TopicConfig) error

	// DeleteTopic deletes a topic
	DeleteTopic(ctx context.Context, topic string) error

	// ListTopics lists all topics
	ListTopics(ctx context.Context) ([]kafka.TopicInfo, error)

	// DescribeTopic describes a topic
	DescribeTopic(ctx context.Context, topic string) (*kafka.TopicInfo, error)

	// ListPartitions lists partitions of a topic
	ListPartitions(ctx context.Context, topic string) ([]kafka.PartitionInfo, error)

	// DescribePartitions describes partitions of a topic
	DescribePartitions(ctx context.Context, topic string) ([]kafka.PartitionInfo, error)

	// CreatePartitions creates additional partitions
	CreatePartitions(ctx context.Context, topic string, count int, assignment [][]int32) error

	// UpdateTopicConfig updates topic configuration
	UpdateTopicConfig(ctx context.Context, topic string, config map[string]string) error

	// GetTopicConfig gets topic configuration
	GetTopicConfig(ctx context.Context, topic string) (map[string]string, error)

	// AlterPartitionReassignments alters partition reassignments
	AlterPartitionReassignments(ctx context.Context, reassignments map[string][]kafka.PartitionAssignments) error

	// ListPartitionReassignments lists partition reassignments
	ListPartitionReassignments(ctx context.Context, topics []string) (map[string][]kafka.PartitionAssignments, error)

	// Close closes the admin client
	Close() error
}

// KafkaTopicManager defines the interface for managing Kafka topics.
type KafkaTopicManager interface {
	// EnsureTopicExists ensures a topic exists with specified configuration
	EnsureTopicExists(ctx context.Context, topic string, numPartitions, replicationFactor int) error

	// EnsureTopicExistsWithConfig ensures a topic exists with custom configuration
	EnsureTopicExistsWithConfig(ctx context.Context, topic string, config *kafka.TopicConfig) error

	// GetOrCreateTopic gets an existing topic or creates it
	GetOrCreateTopic(ctx context.Context, topic string, config *kafka.TopicConfig) (*kafka.TopicInfo, error)

	// GetTopicInfo gets information about a topic
	GetTopicInfo(ctx context.Context, topic string) (*kafka.TopicInfo, error)

	// ListTopics lists all topics
	ListTopics(ctx context.Context) ([]string, error)

	// DeleteTopic deletes a topic
	DeleteTopic(ctx context.Context, topic string) error

	// TopicExists checks if a topic exists
	TopicExists(ctx context.Context, topic string) (bool, error)

	// GetTopicPartitions gets the partitions of a topic
	GetTopicPartitions(ctx context.Context, topic string) ([]kafka.PartitionInfo, error)

	// UpdateTopicPartitions updates the number of partitions
	UpdateTopicPartitions(ctx context.Context, topic string, count int) error

	// GetTopicConfig gets the configuration of a topic
	GetTopicConfig(ctx context.Context, topic string) (map[string]string, error)

	// UpdateTopicConfig updates the configuration of a topic
	UpdateTopicConfig(ctx context.Context, topic string, config map[string]string) error
}

// KafkaConsumerGroup defines the interface for Kafka consumer group operations.
type KafkaConsumerGroup interface {
	// Join joins a consumer group
	Join(ctx context.Context, groupID string) error

	// Leave leaves the consumer group
	Leave(ctx context.Context) error

	// SyncGroup synchronizes group membership
	SyncGroup(ctx context.Context, members []kafka.GroupMember) error

	// Heartbeat sends a heartbeat
	Heartbeat(ctx context.Context, generationID string, memberID string) error

	// GetAssignments gets current assignments
	GetAssignments(ctx context.Context) ([]kafka.GroupAssignment, error)
}

// KafkaOffsetManager defines the interface for managing Kafka offsets.
type KafkaOffsetManager interface {
	// GetOffset gets the current offset for a partition
	GetOffset(ctx context.Context, topic string, partition int) (int64, error)

	// SetOffset sets the offset for a partition
	SetOffset(ctx context.Context, topic string, partition int, offset int64) error

	// Commit commits an offset
	Commit(ctx context.Context, topic string, partition int, offset int64) error

	// SeekEarliest seeks to the earliest offset
	SeekEarliest(ctx context.Context, topic string, partition int) error

	// SeekLatest seeks to the latest offset
	SeekLatest(ctx context.Context, topic string, partition int) error

	// SeekOffset seeks to a specific offset
	SeekOffset(ctx context.Context, topic string, partition int, offset int64) error

	// GetCommittedOffset gets the committed offset for a consumer group
	GetCommittedOffset(ctx context.Context, groupID, topic string, partition int) (int64, error)

	// GetWatermark gets the low and high watermarks
	GetWatermark(ctx context.Context, topic string, partition int) (int64, int64, error)
}

// KafkaSchemaRegistry defines the interface for schema registry operations.
type KafkaSchemaRegistry interface {
	// RegisterSchema registers a schema
	RegisterSchema(ctx context.Context, subject string, schema string) (int, error)

	// GetSchema gets a schema by ID
	GetSchema(ctx context.Context, schemaID int) (string, error)

	// GetSchemaBySubject gets the latest schema for a subject
	GetSchemaBySubject(ctx context.Context, subject string) (string, int, error)

	// CheckCompatibility checks schema compatibility
	CheckCompatibility(ctx context.Context, subject string, schema string, version int) (bool, error)

	// GetSubjects gets all subjects
	GetSubjects(ctx context.Context) ([]string, error)

	// DeleteSubject deletes a subject
	DeleteSubject(ctx context.Context, subject string) error
}

// KafkaEventSerializer defines the interface for serializing events.
type KafkaEventSerializer interface {
	// Serialize serializes an event to bytes
	Serialize(ctx context.Context, event domain.Event) ([]byte, error)

	// Deserialize deserializes bytes to an event
	Deserialize(ctx context.Context, data []byte) (domain.Event, error)

	// GetContentType returns the content type
	GetContentType() string
}

// KafkaEventProducer defines the interface for producing domain events.
type KafkaEventProducer interface {
	// Publish publishes a domain event
	Publish(ctx context.Context, event domain.Event) error

	// PublishTo publishes an event to a specific topic
	PublishTo(ctx context.Context, topic string, event domain.Event) error

	// PublishBatch publishes multiple events
	PublishBatch(ctx context.Context, events []domain.Event) error

	// PublishWithKey publishes an event with a key
	PublishWithKey(ctx context.Context, event domain.Event, key []byte) error

	// Flush flushes pending events
	Flush(ctx context.Context) error

	// Close closes the producer
	Close() error
}

// KafkaEventConsumer defines the interface for consuming domain events.
type KafkaEventConsumer interface {
	// Subscribe subscribes to event types
	Subscribe(ctx context.Context, eventTypes []string) error

	// SubscribeToTopic subscribes to a topic
	SubscribeToTopic(ctx context.Context, topic string) error

	// Unsubscribe unsubscribes from events
	Unsubscribe(ctx context.Context) error

	// Consume consumes a single event
	Consume(ctx context.Context) (domain.Event, error)

	// ConsumeWithHandler consumes events with a handler
	ConsumeWithHandler(ctx context.Context, handler domain.EventHandler) error

	// Commit commits the current position
	Commit(ctx context.Context) error

	// Close closes the consumer
	Close() error
}

// KafkaEventBus defines the interface for a complete event bus implementation.
type KafkaEventBus interface {
	// KafkaEventProducer
	KafkaEventProducer

	// KafkaEventConsumer
	KafkaEventConsumer

	// RegisterHandler registers an event handler
	RegisterHandler(ctx context.Context, handler domain.EventHandler, eventTypes []string) error

	// UnregisterHandler unregisters an event handler
	UnregisterHandler(ctx context.Context, handlerID string) error

	// GetStats returns event bus statistics
	GetStats(ctx context.Context) (*domain.EventBusStats, error)
}

// KafkaMonitoring defines the interface for Kafka monitoring operations.
type KafkaMonitoring interface {
	// GetBrokers gets information about brokers
	GetBrokers(ctx context.Context) ([]kafka.BrokerInfo, error)

	// GetClusterInfo gets cluster information
	GetClusterInfo(ctx context.Context) (*kafka.ClusterInfo, error)

	// GetController gets the current controller
	GetController(ctx context.Context) (*kafka.BrokerInfo, error)

	// GetLag gets consumer group lag
	GetLag(ctx context.Context, groupID string) ([]kafka.PartitionLag, error)

	// GetConsumerGroups gets all consumer groups
	GetConsumerGroups(ctx context.Context) ([]kafka.GroupInfo, error)

	// GetConsumerGroupInfo gets consumer group information
	GetConsumerGroupInfo(ctx context.Context, groupID string) (*kafka.GroupInfo, error)
}

// KafkaConnection defines the interface for managing Kafka connections.
type KafkaConnection interface {
	// Connect establishes a connection
	Connect(ctx context.Context, brokers []string) error

	// Disconnect disconnects from Kafka
	Disconnect(ctx context.Context) error

	// IsConnected checks if connected
	IsConnected() bool

	// GetBrokers gets connected brokers
	GetBrokers() []string

	// GetConnectedBroker gets the currently connected broker
	GetConnectedBroker() string

	// GetMetadata gets metadata
	GetMetadata(ctx context.Context, request *kafka.MetadataRequest) (*kafka.MetadataResponse, error)
}

// KafkaSecurity defines the interface for Kafka security operations.
type KafkaSecurity interface {
	// Authenticate authenticates with Kafka
	Authenticate(ctx context.Context, mechanism string, credentials map[string]string) error

	// Authorize authorizes an operation
	Authorize(ctx context.Context, resourceType, resourceName, operation string) bool

	// GetACLs gets ACLs for a resource
	GetACLs(ctx context.Context, resourceType, resourceName string) ([]kafka.ACL, error)

	// CreateACL creates an ACL
	CreateACL(ctx context.Context, acl kafka.ACL) error

	// DeleteACL deletes an ACL
	DeleteACL(ctx context.Context, acl kafka.ACL) error

	// ListUserScramCredentials lists SCRAM credentials
	ListUserScramCredentials(ctx context.Context) ([]kafka.ScramCredentialInfo, error)
}

// KafkaConnect defines the interface for Kafka Connect operations.
type KafkaConnect interface {
	// GetConnectors gets all connectors
	GetConnectors(ctx context.Context) ([]kafka.ConnectorInfo, error)

	// CreateConnector creates a connector
	CreateConnector(ctx context.Context, config map[string]interface{}) (*kafka.ConnectorInfo, error)

	// GetConnector gets a connector
	GetConnector(ctx context.Context, name string) (*kafka.ConnectorInfo, error)

	// UpdateConnector updates a connector
	UpdateConnector(ctx context.Context, name string, config map[string]interface{}) (*kafka.ConnectorInfo, error)

	// DeleteConnector deletes a connector
	DeleteConnector(ctx context.Context, name string) error

	// GetConnectorTasks gets connector tasks
	GetConnectorTasks(ctx context.Context, name string) ([]kafka.TaskInfo, error)

	// RestartConnector restarts a connector
	RestartConnector(ctx context.Context, name string, includeTasks, onlyFailed bool) error

	// GetConnectorConfig gets connector configuration
	GetConnectorConfig(ctx context.Context, name string) (map[string]interface{}, error)
}
