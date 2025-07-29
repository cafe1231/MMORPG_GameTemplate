package character

import (
	"time"
)

// EventType represents the type of character event
type EventType string

const (
	// Character lifecycle events
	EventCharacterCreated   EventType = "character.created"
	EventCharacterDeleted   EventType = "character.deleted"
	EventCharacterRestored  EventType = "character.restored"
	EventCharacterSelected  EventType = "character.selected"
	
	// Character update events
	EventCharacterPositionUpdated   EventType = "character.position.updated"
	EventCharacterStatsUpdated      EventType = "character.stats.updated"
	EventCharacterAppearanceUpdated EventType = "character.appearance.updated"
	EventCharacterLevelUp           EventType = "character.levelup"
	
	// Character state events
	EventCharacterOnline  EventType = "character.online"
	EventCharacterOffline EventType = "character.offline"
)

// BaseEvent contains common fields for all character events
type BaseEvent struct {
	EventID     string    `json:"event_id"`
	EventType   EventType `json:"event_type"`
	CharacterID string    `json:"character_id"`
	UserID      string    `json:"user_id"`
	Timestamp   time.Time `json:"timestamp"`
	Version     string    `json:"version"`
}

// CharacterCreatedEvent is emitted when a new character is created
type CharacterCreatedEvent struct {
	BaseEvent
	Name       string    `json:"name"`
	ClassType  ClassType `json:"class_type"`
	Race       Race      `json:"race"`
	Gender     Gender    `json:"gender"`
	Level      int       `json:"level"`
	SlotNumber int       `json:"slot_number"`
}

// CharacterDeletedEvent is emitted when a character is deleted
type CharacterDeletedEvent struct {
	BaseEvent
	Name         string `json:"name"`
	DeleteReason string `json:"delete_reason,omitempty"`
	SoftDelete   bool   `json:"soft_delete"`
}

// CharacterRestoredEvent is emitted when a deleted character is restored
type CharacterRestoredEvent struct {
	BaseEvent
	Name           string `json:"name"`
	RestoreReason  string `json:"restore_reason,omitempty"`
}

// CharacterSelectedEvent is emitted when a player selects a character for gameplay
type CharacterSelectedEvent struct {
	BaseEvent
	Name      string `json:"name"`
	SessionID string `json:"session_id"`
	IP        string `json:"ip,omitempty"`
}

// CharacterPositionUpdatedEvent is emitted when character position changes
type CharacterPositionUpdatedEvent struct {
	BaseEvent
	PreviousPosition *Position `json:"previous_position,omitempty"`
	NewPosition      *Position `json:"new_position"`
	MovementType     string    `json:"movement_type"` // walk, run, teleport, etc.
}

// CharacterStatsUpdatedEvent is emitted when character stats change
type CharacterStatsUpdatedEvent struct {
	BaseEvent
	UpdateType    string            `json:"update_type"` // level_up, stat_allocation, equipment_change, etc.
	PreviousStats map[string]int    `json:"previous_stats,omitempty"`
	NewStats      map[string]int    `json:"new_stats"`
	Changes       map[string]int    `json:"changes"` // diff between previous and new
}

// CharacterAppearanceUpdatedEvent is emitted when character appearance changes
type CharacterAppearanceUpdatedEvent struct {
	BaseEvent
	ChangedFields []string `json:"changed_fields"`
	Reason        string   `json:"reason,omitempty"` // cosmetic_shop, character_creation, etc.
}

// CharacterLevelUpEvent is emitted when a character gains a level
type CharacterLevelUpEvent struct {
	BaseEvent
	Name          string `json:"name"`
	PreviousLevel int    `json:"previous_level"`
	NewLevel      int    `json:"new_level"`
	Experience    int64  `json:"experience"`
	StatPoints    int    `json:"stat_points_gained"`
	SkillPoints   int    `json:"skill_points_gained"`
}

// CharacterOnlineEvent is emitted when a character comes online
type CharacterOnlineEvent struct {
	BaseEvent
	Name      string `json:"name"`
	SessionID string `json:"session_id"`
	WorldID   string `json:"world_id"`
	ZoneID    string `json:"zone_id"`
}

// CharacterOfflineEvent is emitted when a character goes offline
type CharacterOfflineEvent struct {
	BaseEvent
	Name          string        `json:"name"`
	SessionID     string        `json:"session_id"`
	OnlineDuration time.Duration `json:"online_duration"`
	Reason        string        `json:"reason"` // logout, disconnect, timeout, kick, etc.
}

// EventMetadata contains additional metadata for events
type EventMetadata struct {
	CorrelationID string            `json:"correlation_id,omitempty"`
	CausationID   string            `json:"causation_id,omitempty"`
	UserAgent     string            `json:"user_agent,omitempty"`
	Headers       map[string]string `json:"headers,omitempty"`
}