package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	
	"github.com/google/uuid"
	"github.com/mmorpg-template/backend/internal/domain/character"
	"github.com/mmorpg-template/backend/internal/ports"
	portsCharacter "github.com/mmorpg-template/backend/internal/ports/character"
	"github.com/mmorpg-template/backend/pkg/logger"
)

// EventPublisher implements the character event publisher using NATS
type EventPublisher struct {
	mq            ports.MessageQueue
	logger        logger.Logger
	streamName    string
	deadLetterEnabled bool
}

// NewEventPublisher creates a new NATS event publisher
func NewEventPublisher(mq ports.MessageQueue, logger logger.Logger) *EventPublisher {
	return &EventPublisher{
		mq:               mq,
		logger:           logger,
		streamName:       "CHARACTER_EVENTS",
		deadLetterEnabled: true,
	}
}

// Initialize creates the necessary streams and configurations
func (p *EventPublisher) Initialize(ctx context.Context) error {
	// Create character events stream
	streamConfig := ports.StreamConfig{
		Name: p.streamName,
		Subjects: []string{
			"character.>", // All character events
		},
		Retention:  ports.LimitsPolicy,
		MaxAge:     30 * 24 * time.Hour, // 30 days retention
		MaxBytes:   10 * 1024 * 1024 * 1024, // 10GB max size
		MaxMsgs:    1000000, // 1M messages max
		Replicas:   1,
	}
	
	if err := p.mq.CreateStream(ctx, streamConfig); err != nil {
		// Check if stream already exists
		if _, getErr := p.mq.GetStreamInfo(ctx, p.streamName); getErr != nil {
			return fmt.Errorf("failed to create character events stream: %w", err)
		}
		p.logger.Info("Character events stream already exists")
	} else {
		p.logger.Info("Created character events stream")
	}
	
	// Create dead letter stream if enabled
	if p.deadLetterEnabled {
		dlqConfig := ports.StreamConfig{
			Name: "CHARACTER_EVENTS_DLQ",
			Subjects: []string{
				"character.dlq.>",
			},
			Retention:  ports.LimitsPolicy,
			MaxAge:     90 * 24 * time.Hour, // 90 days retention for DLQ
			MaxBytes:   1 * 1024 * 1024 * 1024, // 1GB max size
			MaxMsgs:    100000, // 100K messages max
			Replicas:   1,
		}
		
		if err := p.mq.CreateStream(ctx, dlqConfig); err != nil {
			if _, getErr := p.mq.GetStreamInfo(ctx, "CHARACTER_EVENTS_DLQ"); getErr != nil {
				p.logger.Warnf("Failed to create dead letter queue stream: %v", err)
			}
		}
	}
	
	return nil
}

// PublishCharacterCreated publishes a character created event
func (p *EventPublisher) PublishCharacterCreated(ctx context.Context, event *character.CharacterCreatedEvent) error {
	event.EventID = uuid.New().String()
	event.Timestamp = time.Now().UTC()
	event.Version = "1.0"
	
	return p.publishEvent(ctx, string(character.EventCharacterCreated), event)
}

// PublishCharacterDeleted publishes a character deleted event
func (p *EventPublisher) PublishCharacterDeleted(ctx context.Context, event *character.CharacterDeletedEvent) error {
	event.EventID = uuid.New().String()
	event.Timestamp = time.Now().UTC()
	event.Version = "1.0"
	
	return p.publishEvent(ctx, string(character.EventCharacterDeleted), event)
}

// PublishCharacterRestored publishes a character restored event
func (p *EventPublisher) PublishCharacterRestored(ctx context.Context, event *character.CharacterRestoredEvent) error {
	event.EventID = uuid.New().String()
	event.Timestamp = time.Now().UTC()
	event.Version = "1.0"
	
	return p.publishEvent(ctx, string(character.EventCharacterRestored), event)
}

// PublishCharacterSelected publishes a character selected event
func (p *EventPublisher) PublishCharacterSelected(ctx context.Context, event *character.CharacterSelectedEvent) error {
	event.EventID = uuid.New().String()
	event.Timestamp = time.Now().UTC()
	event.Version = "1.0"
	
	return p.publishEvent(ctx, string(character.EventCharacterSelected), event)
}

// PublishCharacterPositionUpdated publishes a character position updated event
func (p *EventPublisher) PublishCharacterPositionUpdated(ctx context.Context, event *character.CharacterPositionUpdatedEvent) error {
	event.EventID = uuid.New().String()
	event.Timestamp = time.Now().UTC()
	event.Version = "1.0"
	
	return p.publishEvent(ctx, string(character.EventCharacterPositionUpdated), event)
}

// PublishCharacterStatsUpdated publishes a character stats updated event
func (p *EventPublisher) PublishCharacterStatsUpdated(ctx context.Context, event *character.CharacterStatsUpdatedEvent) error {
	event.EventID = uuid.New().String()
	event.Timestamp = time.Now().UTC()
	event.Version = "1.0"
	
	return p.publishEvent(ctx, string(character.EventCharacterStatsUpdated), event)
}

// PublishCharacterAppearanceUpdated publishes a character appearance updated event
func (p *EventPublisher) PublishCharacterAppearanceUpdated(ctx context.Context, event *character.CharacterAppearanceUpdatedEvent) error {
	event.EventID = uuid.New().String()
	event.Timestamp = time.Now().UTC()
	event.Version = "1.0"
	
	return p.publishEvent(ctx, string(character.EventCharacterAppearanceUpdated), event)
}

// PublishCharacterLevelUp publishes a character level up event
func (p *EventPublisher) PublishCharacterLevelUp(ctx context.Context, event *character.CharacterLevelUpEvent) error {
	event.EventID = uuid.New().String()
	event.Timestamp = time.Now().UTC()
	event.Version = "1.0"
	
	return p.publishEvent(ctx, string(character.EventCharacterLevelUp), event)
}

// PublishCharacterOnline publishes a character online event
func (p *EventPublisher) PublishCharacterOnline(ctx context.Context, event *character.CharacterOnlineEvent) error {
	event.EventID = uuid.New().String()
	event.Timestamp = time.Now().UTC()
	event.Version = "1.0"
	
	return p.publishEvent(ctx, string(character.EventCharacterOnline), event)
}

// PublishCharacterOffline publishes a character offline event
func (p *EventPublisher) PublishCharacterOffline(ctx context.Context, event *character.CharacterOfflineEvent) error {
	event.EventID = uuid.New().String()
	event.Timestamp = time.Now().UTC()
	event.Version = "1.0"
	
	return p.publishEvent(ctx, string(character.EventCharacterOffline), event)
}

// PublishBatch publishes multiple events in a batch
func (p *EventPublisher) PublishBatch(ctx context.Context, events []interface{}) error {
	for _, event := range events {
		// Type switch to determine event type and subject
		var subject string
		switch e := event.(type) {
		case *character.CharacterCreatedEvent:
			subject = string(character.EventCharacterCreated)
		case *character.CharacterDeletedEvent:
			subject = string(character.EventCharacterDeleted)
		case *character.CharacterRestoredEvent:
			subject = string(character.EventCharacterRestored)
		case *character.CharacterSelectedEvent:
			subject = string(character.EventCharacterSelected)
		case *character.CharacterPositionUpdatedEvent:
			subject = string(character.EventCharacterPositionUpdated)
		case *character.CharacterStatsUpdatedEvent:
			subject = string(character.EventCharacterStatsUpdated)
		case *character.CharacterAppearanceUpdatedEvent:
			subject = string(character.EventCharacterAppearanceUpdated)
		case *character.CharacterLevelUpEvent:
			subject = string(character.EventCharacterLevelUp)
		case *character.CharacterOnlineEvent:
			subject = string(character.EventCharacterOnline)
		case *character.CharacterOfflineEvent:
			subject = string(character.EventCharacterOffline)
		default:
			p.logger.Warnf("Unknown event type in batch: %T", e)
			continue
		}
		
		if err := p.publishEvent(ctx, subject, event); err != nil {
			// Log error but continue with other events
			p.logger.Errorf("Failed to publish event in batch: %v", err)
		}
	}
	
	return nil
}

// publishEvent is a helper method to publish events with error handling
func (p *EventPublisher) publishEvent(ctx context.Context, subject string, event interface{}) error {
	// Marshal event to JSON
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	
	// Log the event being published
	p.logger.Debugf("Publishing event to subject %s: %s", subject, string(data))
	
	// Publish event with retry logic
	maxRetries := 3
	var lastErr error
	
	for i := 0; i < maxRetries; i++ {
		if err := p.mq.Publish(ctx, subject, data); err != nil {
			lastErr = err
			p.logger.Warnf("Failed to publish event (attempt %d/%d): %v", i+1, maxRetries, err)
			
			// Exponential backoff
			if i < maxRetries-1 {
				time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
			}
			continue
		}
		
		// Success
		p.logger.Debugf("Successfully published event to %s", subject)
		return nil
	}
	
	// All retries failed, publish to dead letter queue if enabled
	if p.deadLetterEnabled && lastErr != nil {
		dlqSubject := fmt.Sprintf("character.dlq.%s", subject)
		dlqData := map[string]interface{}{
			"original_subject": subject,
			"event":           event,
			"error":           lastErr.Error(),
			"timestamp":       time.Now().UTC(),
			"retry_count":     maxRetries,
		}
		
		dlqBytes, _ := json.Marshal(dlqData)
		if dlqErr := p.mq.Publish(ctx, dlqSubject, dlqBytes); dlqErr != nil {
			p.logger.Errorf("Failed to publish to dead letter queue: %v", dlqErr)
		} else {
			p.logger.Warnf("Event published to dead letter queue: %s", dlqSubject)
		}
	}
	
	return fmt.Errorf("failed to publish event after %d retries: %w", maxRetries, lastErr)
}

// EventSubscriber implements the character event subscriber using NATS
type EventSubscriber struct {
	mq       ports.MessageQueue
	logger   logger.Logger
	subscriptions []ports.QueueSubscription
}

// NewEventSubscriber creates a new NATS event subscriber
func NewEventSubscriber(mq ports.MessageQueue, logger logger.Logger) *EventSubscriber {
	return &EventSubscriber{
		mq:     mq,
		logger: logger,
		subscriptions: make([]ports.QueueSubscription, 0),
	}
}

// SubscribeToCharacterEvents subscribes to all character events
func (s *EventSubscriber) SubscribeToCharacterEvents(ctx context.Context, handler portsCharacter.EventHandler) error {
	subject := "character.>"
	
	sub, err := s.mq.Subscribe(ctx, subject, func(msg *ports.QueueMessage) error {
		return handler(ctx, character.EventType(msg.Subject), msg.Data)
	})
	
	if err != nil {
		return fmt.Errorf("failed to subscribe to character events: %w", err)
	}
	
	s.subscriptions = append(s.subscriptions, sub)
	s.logger.Infof("Subscribed to character events: %s", subject)
	return nil
}

// SubscribeToEvent subscribes to a specific event type
func (s *EventSubscriber) SubscribeToEvent(ctx context.Context, eventType character.EventType, handler portsCharacter.EventHandler) error {
	subject := string(eventType)
	
	sub, err := s.mq.Subscribe(ctx, subject, func(msg *ports.QueueMessage) error {
		return handler(ctx, eventType, msg.Data)
	})
	
	if err != nil {
		return fmt.Errorf("failed to subscribe to event %s: %w", eventType, err)
	}
	
	s.subscriptions = append(s.subscriptions, sub)
	s.logger.Infof("Subscribed to event: %s", subject)
	return nil
}

// SubscribeToUserCharacterEvents subscribes to events for a specific user's characters
func (s *EventSubscriber) SubscribeToUserCharacterEvents(ctx context.Context, userID string, handler portsCharacter.EventHandler) error {
	// This would require filtering in the handler since NATS doesn't support
	// filtering by message content natively
	return s.SubscribeToCharacterEvents(ctx, func(ctx context.Context, eventType character.EventType, data []byte) error {
		// Parse event to check user ID
		var baseEvent character.BaseEvent
		if err := json.Unmarshal(data, &baseEvent); err != nil {
			return nil // Skip malformed events
		}
		
		if baseEvent.UserID == userID {
			return handler(ctx, eventType, data)
		}
		
		return nil
	})
}

// Unsubscribe stops listening to events
func (s *EventSubscriber) Unsubscribe() error {
	for _, sub := range s.subscriptions {
		if err := sub.Unsubscribe(); err != nil {
			s.logger.Warnf("Failed to unsubscribe: %v", err)
		}
	}
	s.subscriptions = nil
	return nil
}