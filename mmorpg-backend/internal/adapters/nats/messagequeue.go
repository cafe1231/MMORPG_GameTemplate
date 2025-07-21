package nats

import (
	"context"
	"fmt"
	"time"
	
	"github.com/nats-io/nats.go"
	"github.com/mmorpg-template/backend/internal/ports"
	"github.com/mmorpg-template/backend/pkg/logger"
)

// NATSMessageQueue implements the ports.MessageQueue interface for NATS
type NATSMessageQueue struct {
	conn   *nats.Conn
	js     nats.JetStreamContext
	config *ports.MessageQueueConfig
	log    logger.Logger
}

// NewNATSMessageQueue creates a new NATS message queue adapter
func NewNATSMessageQueue(config *ports.MessageQueueConfig, log logger.Logger) *NATSMessageQueue {
	return &NATSMessageQueue{
		config: config,
		log:    log,
	}
}

// Connect establishes a connection to NATS
func (n *NATSMessageQueue) Connect(ctx context.Context) error {
	opts := []nats.Option{
		nats.Name(n.config.ClientID),
		nats.MaxReconnects(n.config.MaxReconnects),
		nats.ReconnectWait(n.config.ReconnectWait),
		nats.PingInterval(n.config.PingInterval),
		nats.MaxPingsOutstanding(n.config.MaxPingsOut),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			n.log.Warnf("NATS disconnected: %v", err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			n.log.Info("NATS reconnected")
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			n.log.Info("NATS connection closed")
		}),
	}
	
	// Connect to NATS
	conn, err := nats.Connect(n.config.URL, opts...)
	if err != nil {
		return fmt.Errorf("failed to connect to NATS: %w", err)
	}
	
	// Create JetStream context
	js, err := conn.JetStream()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to create JetStream context: %w", err)
	}
	
	n.conn = conn
	n.js = js
	
	n.log.Infof("Connected to NATS at %s", n.config.URL)
	return nil
}

// Close closes the NATS connection
func (n *NATSMessageQueue) Close() error {
	if n.conn != nil {
		n.conn.Close()
	}
	return nil
}

// Publish publishes a message to a subject
func (n *NATSMessageQueue) Publish(ctx context.Context, subject string, data []byte) error {
	if n.conn == nil {
		return ports.ErrMQConnection
	}
	
	return n.conn.Publish(subject, data)
}

// PublishWithReply publishes a message and waits for a reply
func (n *NATSMessageQueue) PublishWithReply(ctx context.Context, subject string, data []byte, timeout time.Duration) ([]byte, error) {
	if n.conn == nil {
		return nil, ports.ErrMQConnection
	}
	
	msg, err := n.conn.RequestWithContext(ctx, subject, data)
	if err != nil {
		if err == nats.ErrTimeout {
			return nil, ports.ErrMQTimeout
		}
		return nil, err
	}
	
	return msg.Data, nil
}

// Subscribe creates a subscription to a subject
func (n *NATSMessageQueue) Subscribe(ctx context.Context, subject string, handler ports.MessageHandler) (ports.Subscription, error) {
	if n.conn == nil {
		return nil, ports.ErrMQConnection
	}
	
	sub, err := n.conn.Subscribe(subject, func(msg *nats.Msg) {
		portMsg := &ports.Message{
			Subject: msg.Subject,
			Data:    msg.Data,
			Reply:   msg.Reply,
			Headers: natsHeadersToMap(msg.Header),
		}
		
		if err := handler(portMsg); err != nil {
			n.log.Errorf("Message handler error: %v", err)
		}
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe: %w", err)
	}
	
	return &natsSubscription{sub: sub}, nil
}

// QueueSubscribe creates a queue subscription to a subject
func (n *NATSMessageQueue) QueueSubscribe(ctx context.Context, subject, queue string, handler ports.MessageHandler) (ports.Subscription, error) {
	if n.conn == nil {
		return nil, ports.ErrMQConnection
	}
	
	sub, err := n.conn.QueueSubscribe(subject, queue, func(msg *nats.Msg) {
		portMsg := &ports.Message{
			Subject: msg.Subject,
			Data:    msg.Data,
			Reply:   msg.Reply,
			Headers: natsHeadersToMap(msg.Header),
		}
		
		if err := handler(portMsg); err != nil {
			n.log.Errorf("Message handler error: %v", err)
		}
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to queue subscribe: %w", err)
	}
	
	return &natsSubscription{sub: sub}, nil
}

// Request sends a request and waits for a reply
func (n *NATSMessageQueue) Request(ctx context.Context, subject string, data []byte, timeout time.Duration) ([]byte, error) {
	return n.PublishWithReply(ctx, subject, data, timeout)
}

// CreateStream creates a new JetStream stream
func (n *NATSMessageQueue) CreateStream(ctx context.Context, config ports.StreamConfig) error {
	if n.js == nil {
		return ports.ErrMQConnection
	}
	
	streamConfig := &nats.StreamConfig{
		Name:       config.Name,
		Subjects:   config.Subjects,
		Retention:  convertRetention(config.Retention),
		MaxAge:     config.MaxAge,
		MaxBytes:   config.MaxBytes,
		MaxMsgs:    config.MaxMsgs,
		MaxMsgSize: config.MaxMsgSize,
		Replicas:   config.Replicas,
		NoAck:      config.NoAck,
	}
	
	_, err := n.js.AddStream(streamConfig)
	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}
	
	n.log.Infof("Created JetStream stream: %s", config.Name)
	return nil
}

// DeleteStream deletes a JetStream stream
func (n *NATSMessageQueue) DeleteStream(ctx context.Context, stream string) error {
	if n.js == nil {
		return ports.ErrMQConnection
	}
	
	return n.js.DeleteStream(stream)
}

// GetStreamInfo gets information about a stream
func (n *NATSMessageQueue) GetStreamInfo(ctx context.Context, stream string) (*ports.StreamInfo, error) {
	if n.js == nil {
		return nil, ports.ErrMQConnection
	}
	
	info, err := n.js.StreamInfo(stream)
	if err != nil {
		return nil, err
	}
	
	return &ports.StreamInfo{
		Name:      info.Config.Name,
		Created:   info.Created,
		Subjects:  info.Config.Subjects,
		Messages:  info.State.Msgs,
		Bytes:     info.State.Bytes,
		FirstSeq:  info.State.FirstSeq,
		LastSeq:   info.State.LastSeq,
		Consumers: info.State.Consumers,
	}, nil
}

// natsSubscription wraps NATS subscription
type natsSubscription struct {
	sub *nats.Subscription
}

func (s *natsSubscription) Unsubscribe() error {
	return s.sub.Unsubscribe()
}

func (s *natsSubscription) Close() error {
	return s.sub.Drain()
}

// Helper functions

func natsHeadersToMap(h nats.Header) map[string]string {
	result := make(map[string]string)
	for k, v := range h {
		if len(v) > 0 {
			result[k] = v[0]
		}
	}
	return result
}

func convertRetention(r ports.RetentionPolicy) nats.RetentionPolicy {
	switch r {
	case ports.LimitsPolicy:
		return nats.LimitsPolicy
	case ports.InterestPolicy:
		return nats.InterestPolicy
	case ports.WorkQueuePolicy:
		return nats.WorkQueuePolicy
	default:
		return nats.LimitsPolicy
	}
}