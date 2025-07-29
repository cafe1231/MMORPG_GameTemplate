package character

import (
	"context"
	
	"github.com/mmorpg-template/backend/internal/domain/character"
)

// EventPublisher defines the interface for publishing character events
type EventPublisher interface {
	// PublishCharacterCreated publishes a character created event
	PublishCharacterCreated(ctx context.Context, event *character.CharacterCreatedEvent) error
	
	// PublishCharacterDeleted publishes a character deleted event
	PublishCharacterDeleted(ctx context.Context, event *character.CharacterDeletedEvent) error
	
	// PublishCharacterRestored publishes a character restored event
	PublishCharacterRestored(ctx context.Context, event *character.CharacterRestoredEvent) error
	
	// PublishCharacterSelected publishes a character selected event
	PublishCharacterSelected(ctx context.Context, event *character.CharacterSelectedEvent) error
	
	// PublishCharacterPositionUpdated publishes a character position updated event
	PublishCharacterPositionUpdated(ctx context.Context, event *character.CharacterPositionUpdatedEvent) error
	
	// PublishCharacterStatsUpdated publishes a character stats updated event
	PublishCharacterStatsUpdated(ctx context.Context, event *character.CharacterStatsUpdatedEvent) error
	
	// PublishCharacterAppearanceUpdated publishes a character appearance updated event
	PublishCharacterAppearanceUpdated(ctx context.Context, event *character.CharacterAppearanceUpdatedEvent) error
	
	// PublishCharacterLevelUp publishes a character level up event
	PublishCharacterLevelUp(ctx context.Context, event *character.CharacterLevelUpEvent) error
	
	// PublishCharacterOnline publishes a character online event
	PublishCharacterOnline(ctx context.Context, event *character.CharacterOnlineEvent) error
	
	// PublishCharacterOffline publishes a character offline event
	PublishCharacterOffline(ctx context.Context, event *character.CharacterOfflineEvent) error
	
	// PublishBatch publishes multiple events in a batch
	PublishBatch(ctx context.Context, events []interface{}) error
}

// EventSubscriber defines the interface for subscribing to character events
type EventSubscriber interface {
	// SubscribeToCharacterEvents subscribes to all character events
	SubscribeToCharacterEvents(ctx context.Context, handler EventHandler) error
	
	// SubscribeToEvent subscribes to a specific event type
	SubscribeToEvent(ctx context.Context, eventType character.EventType, handler EventHandler) error
	
	// SubscribeToUserCharacterEvents subscribes to events for a specific user's characters
	SubscribeToUserCharacterEvents(ctx context.Context, userID string, handler EventHandler) error
	
	// Unsubscribe stops listening to events
	Unsubscribe() error
}

// EventHandler handles incoming character events
type EventHandler func(ctx context.Context, eventType character.EventType, data []byte) error

// EventStore defines the interface for storing and retrieving events
type EventStore interface {
	// StoreEvent stores an event
	StoreEvent(ctx context.Context, event interface{}) error
	
	// GetEventsByCharacter retrieves events for a specific character
	GetEventsByCharacter(ctx context.Context, characterID string, limit int, offset int) ([]interface{}, error)
	
	// GetEventsByUser retrieves events for all characters of a user
	GetEventsByUser(ctx context.Context, userID string, limit int, offset int) ([]interface{}, error)
	
	// GetEventsByType retrieves events of a specific type
	GetEventsByType(ctx context.Context, eventType character.EventType, limit int, offset int) ([]interface{}, error)
}