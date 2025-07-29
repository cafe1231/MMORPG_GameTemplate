# Character Event System

This document describes the event-driven architecture for the character service in the MMORPG backend.

## Overview

The character service publishes events to NATS JetStream whenever significant character operations occur. This enables other services to react to character changes in real-time without tight coupling.

## Event Types

### Character Lifecycle Events

#### `character.created`
Published when a new character is created.
```json
{
  "event_id": "uuid",
  "event_type": "character.created",
  "character_id": "uuid",
  "user_id": "uuid",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "1.0",
  "name": "PlayerName",
  "class_type": "warrior",
  "race": "human",
  "gender": "male",
  "level": 1,
  "slot_number": 1
}
```

#### `character.deleted`
Published when a character is deleted (soft delete).
```json
{
  "event_id": "uuid",
  "event_type": "character.deleted",
  "character_id": "uuid",
  "user_id": "uuid",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "1.0",
  "name": "PlayerName",
  "delete_reason": "user_requested",
  "soft_delete": true
}
```

#### `character.restored`
Published when a deleted character is restored.
```json
{
  "event_id": "uuid",
  "event_type": "character.restored",
  "character_id": "uuid",
  "user_id": "uuid",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "1.0",
  "name": "PlayerName",
  "restore_reason": "user_requested"
}
```

#### `character.selected`
Published when a player selects a character for gameplay.
```json
{
  "event_id": "uuid",
  "event_type": "character.selected",
  "character_id": "uuid",
  "user_id": "uuid",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "1.0",
  "name": "PlayerName",
  "session_id": "session-uuid",
  "ip": "192.168.1.1"
}
```

### Character State Events

#### `character.online`
Published when a character comes online.
```json
{
  "event_id": "uuid",
  "event_type": "character.online",
  "character_id": "uuid",
  "user_id": "uuid",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "1.0",
  "name": "PlayerName",
  "session_id": "session-uuid",
  "world_id": "world1",
  "zone_id": "zone1"
}
```

#### `character.offline`
Published when a character goes offline.
```json
{
  "event_id": "uuid",
  "event_type": "character.offline",
  "character_id": "uuid",
  "user_id": "uuid",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "1.0",
  "name": "PlayerName",
  "session_id": "session-uuid",
  "online_duration": 3600000000000,
  "reason": "logout"
}
```

### Character Update Events

#### `character.position.updated`
Published when character position changes.
```json
{
  "event_id": "uuid",
  "event_type": "character.position.updated",
  "character_id": "uuid",
  "user_id": "uuid",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "1.0",
  "previous_position": {
    "world_id": "world1",
    "zone_id": "zone1",
    "position_x": 100.0,
    "position_y": 200.0,
    "position_z": 50.0
  },
  "new_position": {
    "world_id": "world1",
    "zone_id": "zone1",
    "position_x": 150.0,
    "position_y": 250.0,
    "position_z": 50.0
  },
  "movement_type": "walk"
}
```

#### `character.stats.updated`
Published when character stats change.
```json
{
  "event_id": "uuid",
  "event_type": "character.stats.updated",
  "character_id": "uuid",
  "user_id": "uuid",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "1.0",
  "update_type": "stat_allocation",
  "previous_stats": {
    "strength": 10,
    "agility": 12,
    "intelligence": 8,
    "wisdom": 7,
    "vitality": 15
  },
  "new_stats": {
    "strength": 11,
    "agility": 12,
    "intelligence": 8,
    "wisdom": 7,
    "vitality": 15
  },
  "changes": {
    "strength": 1
  }
}
```

#### `character.appearance.updated`
Published when character appearance changes.
```json
{
  "event_id": "uuid",
  "event_type": "character.appearance.updated",
  "character_id": "uuid",
  "user_id": "uuid",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "1.0",
  "changed_fields": ["hair_color", "hair_style", "face_type"],
  "reason": "cosmetic_shop"
}
```

#### `character.levelup`
Published when a character gains a level.
```json
{
  "event_id": "uuid",
  "event_type": "character.levelup",
  "character_id": "uuid",
  "user_id": "uuid",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "1.0",
  "name": "PlayerName",
  "previous_level": 10,
  "new_level": 11,
  "experience": 150000,
  "stat_points_gained": 5,
  "skill_points_gained": 2
}
```

## Implementation Details

### Event Publisher

The event publisher is implemented in `internal/adapters/character/nats/publisher.go`:

```go
type EventPublisher struct {
    mq            ports.MessageQueue
    logger        logger.Logger
    streamName    string
    deadLetterEnabled bool
}
```

Features:
- Automatic retry with exponential backoff (3 attempts)
- Dead letter queue for failed events
- Async publishing to not block operations
- Event versioning for compatibility
- JetStream for guaranteed delivery

### Subscribing to Events

Services can subscribe to character events using the event subscriber:

```go
// Subscribe to all character events
subscriber := nats.NewEventSubscriber(mq, logger)
err := subscriber.SubscribeToCharacterEvents(ctx, func(ctx context.Context, eventType character.EventType, data []byte) error {
    // Handle event
    return nil
})

// Subscribe to specific event type
err := subscriber.SubscribeToEvent(ctx, character.EventCharacterCreated, handler)

// Queue subscription for load balancing
mq.QueueSubscribe(ctx, "character.created", "my-service-group", handler)
```

### Stream Configuration

Character events are stored in a JetStream stream with:
- 30-day retention
- 10GB max size
- 1M message limit
- Subject pattern: `character.>`

Dead letter queue stream:
- 90-day retention
- 1GB max size
- 100K message limit
- Subject pattern: `character.dlq.>`

## Usage Examples

### Publishing Events

Events are automatically published by the character service:

```go
// When creating a character
char, err := characterService.CreateCharacter(ctx, request)
// Automatically publishes character.created event

// When updating position
position, err := characterService.UpdatePosition(ctx, characterID, request)
// Automatically publishes character.position.updated event
```

### Consuming Events

Example consumers in other services:

```go
// World service - update spatial index on position change
func HandlePositionUpdate(event *CharacterPositionUpdatedEvent) {
    spatialIndex.UpdateCharacterLocation(
        event.CharacterID,
        event.NewPosition.WorldID,
        event.NewPosition.ZoneID,
        event.NewPosition.PositionX,
        event.NewPosition.PositionY,
        event.NewPosition.PositionZ,
    )
}

// Social service - notify friends when character comes online
func HandleCharacterOnline(event *CharacterOnlineEvent) {
    friends := friendService.GetOnlineFriends(event.CharacterID)
    for _, friend := range friends {
        notificationService.SendFriendOnlineNotification(friend.ID, event.Name)
    }
}

// Analytics service - track character progression
func HandleCharacterLevelUp(event *CharacterLevelUpEvent) {
    analytics.TrackEvent("character.levelup", map[string]interface{}{
        "character_id": event.CharacterID,
        "user_id":      event.UserID,
        "new_level":    event.NewLevel,
        "time_to_level": calculateTimeSincePreviousLevel(event.CharacterID),
    })
}
```

## Best Practices

1. **Event Naming**: Use consistent naming pattern `character.<action>` or `character.<entity>.<action>`

2. **Event Payload**: Include enough data for consumers to process without additional queries

3. **Idempotency**: Design event handlers to be idempotent - safe to process multiple times

4. **Error Handling**: Always handle events gracefully, log errors but don't crash

5. **Performance**: Use queue subscriptions for heavy processing to distribute load

6. **Monitoring**: Track event publishing/consumption metrics

## Testing

Run event publisher tests:
```bash
cd mmorpg-backend
go test ./internal/adapters/character/nats/...
```

## Future Enhancements

1. Event sourcing for character history
2. Event replay capabilities
3. Schema registry for event versioning
4. Cross-region event replication
5. WebSocket forwarding for real-time client updates