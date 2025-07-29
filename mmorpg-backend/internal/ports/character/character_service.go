package character

import (
	"context"

	"github.com/mmorpg-template/backend/internal/domain/character"
)

// CharacterService defines the interface for character management operations
type CharacterService interface {
	// Character management
	CreateCharacter(ctx context.Context, req *CreateCharacterRequest) (*character.Character, error)
	GetCharacter(ctx context.Context, characterID string) (*character.Character, error)
	GetCharacterByName(ctx context.Context, name string) (*character.Character, error)
	ListCharactersByUser(ctx context.Context, userID string) ([]*character.Character, error)
	DeleteCharacter(ctx context.Context, characterID string, userID string) error
	RestoreCharacter(ctx context.Context, characterID string, userID string) error
	
	// Character appearance
	GetAppearance(ctx context.Context, characterID string) (*character.Appearance, error)
	UpdateAppearance(ctx context.Context, characterID string, req *UpdateAppearanceRequest) (*character.Appearance, error)
	
	// Character stats
	GetStats(ctx context.Context, characterID string) (*character.Stats, error)
	AllocateStatPoint(ctx context.Context, characterID string, stat string) (*character.Stats, error)
	
	// Character position
	GetPosition(ctx context.Context, characterID string) (*character.Position, error)
	UpdatePosition(ctx context.Context, characterID string, req *UpdatePositionRequest) (*character.Position, error)
	TeleportToSafePosition(ctx context.Context, characterID string) (*character.Position, error)
	
	// Validation
	ValidateCharacterOwnership(ctx context.Context, characterID string, userID string) error
	CanCreateCharacter(ctx context.Context, userID string) (bool, error)
	
	// Gameplay
	SelectCharacter(ctx context.Context, characterID string, userID string, sessionID string) error
}

// CreateCharacterRequest represents a request to create a new character
type CreateCharacterRequest struct {
	UserID     string
	Name       string
	SlotNumber int
	ClassType  character.ClassType
	Race       character.Race
	Gender     character.Gender
	Appearance *CharacterAppearanceOptions
}

// CharacterAppearanceOptions represents optional appearance customization
type CharacterAppearanceOptions struct {
	FaceType        *int
	SkinColor       *string
	EyeColor        *string
	HairStyle       *int
	HairColor       *string
	FacialHairStyle *int
	FacialHairColor *string
	BodyType        *character.BodyType
	Height          *float32
}

// UpdateAppearanceRequest represents a request to update character appearance
type UpdateAppearanceRequest struct {
	FaceType         *int                     `json:"face_type,omitempty"`
	SkinColor        *string                  `json:"skin_color,omitempty"`
	EyeColor         *string                  `json:"eye_color,omitempty"`
	HairStyle        *int                     `json:"hair_style,omitempty"`
	HairColor        *string                  `json:"hair_color,omitempty"`
	FacialHairStyle  *int                     `json:"facial_hair_style,omitempty"`
	FacialHairColor  *string                  `json:"facial_hair_color,omitempty"`
	BodyType         *character.BodyType      `json:"body_type,omitempty"`
	Height           *float32                 `json:"height,omitempty"`
	BodyProportions  *character.BodyProportions `json:"body_proportions,omitempty"`
	Scars            []int                    `json:"scars,omitempty"`
	Tattoos          []int                    `json:"tattoos,omitempty"`
	Accessories      []int                    `json:"accessories,omitempty"`
}

// UpdatePositionRequest represents a request to update character position
type UpdatePositionRequest struct {
	WorldID       *string  `json:"world_id,omitempty"`
	ZoneID        *string  `json:"zone_id,omitempty"`
	MapID         *string  `json:"map_id,omitempty"`
	PositionX     *float64 `json:"position_x,omitempty"`
	PositionY     *float64 `json:"position_y,omitempty"`
	PositionZ     *float64 `json:"position_z,omitempty"`
	RotationPitch *float32 `json:"rotation_pitch,omitempty"`
	RotationYaw   *float32 `json:"rotation_yaw,omitempty"`
	RotationRoll  *float32 `json:"rotation_roll,omitempty"`
	VelocityX     *float32 `json:"velocity_x,omitempty"`
	VelocityY     *float32 `json:"velocity_y,omitempty"`
	VelocityZ     *float32 `json:"velocity_z,omitempty"`
}