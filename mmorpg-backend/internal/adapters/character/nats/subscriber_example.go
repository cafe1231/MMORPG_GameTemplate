package nats

import (
	"context"
	"encoding/json"
	"fmt"
	
	"github.com/mmorpg-template/backend/internal/domain/character"
	"github.com/mmorpg-template/backend/internal/ports"
	"github.com/mmorpg-template/backend/pkg/logger"
)

// ExampleCharacterEventHandler shows how other services can handle character events
type ExampleCharacterEventHandler struct {
	logger logger.Logger
}

// NewExampleCharacterEventHandler creates a new example event handler
func NewExampleCharacterEventHandler(logger logger.Logger) *ExampleCharacterEventHandler {
	return &ExampleCharacterEventHandler{
		logger: logger,
	}
}

// HandleCharacterCreated handles character created events
func (h *ExampleCharacterEventHandler) HandleCharacterCreated(ctx context.Context, event *character.CharacterCreatedEvent) error {
	h.logger.Infof("Character created: %s (ID: %s, User: %s, Class: %s)",
		event.Name, event.CharacterID, event.UserID, event.ClassType)
	
	// Example: Update user statistics
	// Example: Send welcome message
	// Example: Initialize character in other systems
	
	return nil
}

// HandleCharacterDeleted handles character deleted events
func (h *ExampleCharacterEventHandler) HandleCharacterDeleted(ctx context.Context, event *character.CharacterDeletedEvent) error {
	h.logger.Infof("Character deleted: %s (ID: %s, Soft: %v)",
		event.Name, event.CharacterID, event.SoftDelete)
	
	// Example: Clean up character data in other systems
	// Example: Update guild membership
	// Example: Clear active sessions
	
	return nil
}

// HandleCharacterOnline handles character online events
func (h *ExampleCharacterEventHandler) HandleCharacterOnline(ctx context.Context, event *character.CharacterOnlineEvent) error {
	h.logger.Infof("Character online: %s (ID: %s, World: %s, Zone: %s)",
		event.Name, event.CharacterID, event.WorldID, event.ZoneID)
	
	// Example: Update presence service
	// Example: Notify friends
	// Example: Resume pending activities
	
	return nil
}

// HandleCharacterPositionUpdated handles character position updates
func (h *ExampleCharacterEventHandler) HandleCharacterPositionUpdated(ctx context.Context, event *character.CharacterPositionUpdatedEvent) error {
	if event.NewPosition != nil {
		h.logger.Debugf("Character %s moved to position (%.2f, %.2f, %.2f) via %s",
			event.CharacterID,
			event.NewPosition.PositionX,
			event.NewPosition.PositionY,
			event.NewPosition.PositionZ,
			event.MovementType)
	}
	
	// Example: Update spatial indexing
	// Example: Check proximity events
	// Example: Update minimap for nearby players
	
	return nil
}

// SetupCharacterEventSubscriptions sets up all character event subscriptions
func SetupCharacterEventSubscriptions(
	ctx context.Context,
	mq ports.MessageQueue,
	handler *ExampleCharacterEventHandler,
	logger logger.Logger,
) error {
	subscriber := NewEventSubscriber(mq, logger)
	
	// Subscribe to all character events
	err := subscriber.SubscribeToCharacterEvents(ctx, func(ctx context.Context, eventType character.EventType, data []byte) error {
		switch eventType {
		case character.EventCharacterCreated:
			var event character.CharacterCreatedEvent
			if err := json.Unmarshal(data, &event); err != nil {
				return fmt.Errorf("failed to unmarshal character created event: %w", err)
			}
			return handler.HandleCharacterCreated(ctx, &event)
			
		case character.EventCharacterDeleted:
			var event character.CharacterDeletedEvent
			if err := json.Unmarshal(data, &event); err != nil {
				return fmt.Errorf("failed to unmarshal character deleted event: %w", err)
			}
			return handler.HandleCharacterDeleted(ctx, &event)
			
		case character.EventCharacterOnline:
			var event character.CharacterOnlineEvent
			if err := json.Unmarshal(data, &event); err != nil {
				return fmt.Errorf("failed to unmarshal character online event: %w", err)
			}
			return handler.HandleCharacterOnline(ctx, &event)
			
		case character.EventCharacterPositionUpdated:
			var event character.CharacterPositionUpdatedEvent
			if err := json.Unmarshal(data, &event); err != nil {
				return fmt.Errorf("failed to unmarshal position updated event: %w", err)
			}
			return handler.HandleCharacterPositionUpdated(ctx, &event)
			
		default:
			// Log unhandled events
			logger.Debugf("Unhandled character event type: %s", eventType)
		}
		
		return nil
	})
	
	if err != nil {
		return fmt.Errorf("failed to subscribe to character events: %w", err)
	}
	
	logger.Info("Successfully subscribed to character events")
	return nil
}

// Example of queue-based subscription for load balancing across multiple instances
func SetupQueueBasedSubscription(
	ctx context.Context,
	mq ports.MessageQueue,
	handler *ExampleCharacterEventHandler,
	logger logger.Logger,
) error {
	// Queue subscriptions ensure only one instance processes each event
	queueGroup := "character-event-processors"
	
	// Subscribe to character created events with queue group
	_, err := mq.QueueSubscribe(ctx, string(character.EventCharacterCreated), queueGroup, 
		func(msg *ports.QueueMessage) error {
			var event character.CharacterCreatedEvent
			if err := json.Unmarshal(msg.Data, &event); err != nil {
				return fmt.Errorf("failed to unmarshal: %w", err)
			}
			return handler.HandleCharacterCreated(ctx, &event)
		})
	
	if err != nil {
		return fmt.Errorf("failed to setup queue subscription: %w", err)
	}
	
	logger.Infof("Successfully set up queue-based subscription with group: %s", queueGroup)
	return nil
}