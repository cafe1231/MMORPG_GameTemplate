package ports

import (
	"context"
	"time"
)

// MessageQueue is the interface for message queue operations
// This abstraction supports different messaging systems (NATS, RabbitMQ, Kafka, etc.)
type MessageQueue interface {
	// Connection management
	Connect(ctx context.Context) error
	Close() error
	
	// Publishing
	Publish(ctx context.Context, subject string, data []byte) error
	PublishWithReply(ctx context.Context, subject string, data []byte, timeout time.Duration) ([]byte, error)
	
	// Subscribing
	Subscribe(ctx context.Context, subject string, handler MessageHandler) (QueueSubscription, error)
	QueueSubscribe(ctx context.Context, subject, queue string, handler MessageHandler) (QueueSubscription, error)
	
	// Request-Reply pattern
	Request(ctx context.Context, subject string, data []byte, timeout time.Duration) ([]byte, error)
	
	// Streaming
	CreateStream(ctx context.Context, config StreamConfig) error
	DeleteStream(ctx context.Context, stream string) error
	GetStreamInfo(ctx context.Context, stream string) (*StreamInfo, error)
}

// QueueSubscription represents a message queue subscription
type QueueSubscription interface {
	// Unsubscribe from the subscription
	Unsubscribe() error
	
	// IsValid checks if the subscription is still valid
	IsValid() bool
	
	// Drain drains any pending messages
	Drain() error
}

// MessageHandler handles incoming messages
type MessageHandler func(msg *QueueMessage) error

// QueueMessage represents a message in the queue
type QueueMessage struct {
	Subject  string
	Data     []byte
	Headers  map[string]string
	ReplyTo  string
	Sequence uint64
}

// Reply sends a reply to a message
func (m *QueueMessage) Reply(data []byte) error {
	// Implementation will be provided by the adapter
	return nil
}

// Ack acknowledges message processing
func (m *QueueMessage) Ack() error {
	// Implementation will be provided by the adapter
	return nil
}

// Nack negative acknowledges message processing
func (m *QueueMessage) Nack() error {
	// Implementation will be provided by the adapter
	return nil
}

// StreamConfig defines configuration for a message stream
type StreamConfig struct {
	Name        string
	Subjects    []string
	Retention   RetentionPolicy
	MaxAge      time.Duration
	MaxBytes    int64
	MaxMsgs     int64
	MaxMsgSize  int32
	Replicas    int
	NoAck       bool
}

// RetentionPolicy defines how messages are retained
type RetentionPolicy int

const (
	// LimitsPolicy removes messages when limits are exceeded
	LimitsPolicy RetentionPolicy = iota
	// InterestPolicy removes messages when all consumers have acknowledged
	InterestPolicy
	// WorkQueuePolicy removes messages when consumed
	WorkQueuePolicy
)

// StreamInfo contains information about a stream
type StreamInfo struct {
	Name      string
	Created   time.Time
	Subjects  []string
	Messages  uint64
	Bytes     uint64
	FirstSeq  uint64
	LastSeq   uint64
	Consumers int
}

// Consumer represents a message consumer
type Consumer interface {
	// Fetch messages
	Fetch(batch int, timeout time.Duration) ([]*QueueMessage, error)
	
	// Consumer info
	Info() (*ConsumerInfo, error)
	
	// Close consumer
	Close() error
}

// ConsumerInfo contains information about a consumer
type ConsumerInfo struct {
	Name          string
	Stream        string
	Created       time.Time
	Delivered     uint64
	AckPending    uint64
	RedeliveryCount uint64
}

// ConsumerConfig defines configuration for a consumer
type ConsumerConfig struct {
	Name           string
	Stream         string
	FilterSubject  string
	DeliverPolicy  DeliverPolicy
	AckPolicy      AckPolicy
	MaxDeliver     int
	MaxAckPending  int
	AckWait        time.Duration
}

// DeliverPolicy defines where to start delivering messages
type DeliverPolicy int

const (
	// DeliverAll delivers all messages
	DeliverAll DeliverPolicy = iota
	// DeliverLast delivers only the last message
	DeliverLast
	// DeliverNew delivers only new messages
	DeliverNew
	// DeliverByStartTime delivers messages from a specific time
	DeliverByStartTime
)

// AckPolicy defines how messages should be acknowledged
type AckPolicy int

const (
	// AckExplicit requires explicit acknowledgment
	AckExplicit AckPolicy = iota
	// AckAll acknowledges all messages up to sequence
	AckAll
	// AckNone requires no acknowledgment
	AckNone
)

// MessageQueueConfig holds message queue configuration
type MessageQueueConfig struct {
	URL           string
	ClusterID     string
	ClientID      string
	MaxReconnects int
	ReconnectWait time.Duration
	
	// Connection options
	PingInterval    time.Duration
	MaxPingsOut     int
	
	// Performance options
	MaxPending      int
	SendBufferSize  int
}

// Common message queue errors
var (
	ErrMQConnection     = NewError("MQ_CONNECTION", "Message queue connection error")
	ErrMQPublish        = NewError("MQ_PUBLISH", "Failed to publish message")
	ErrMQSubscribe      = NewError("MQ_SUBSCRIBE", "Failed to subscribe")
	ErrMQTimeout        = NewError("MQ_TIMEOUT", "Message queue operation timeout")
	ErrMQInvalidSubject = NewError("MQ_INVALID_SUBJECT", "Invalid subject")
)