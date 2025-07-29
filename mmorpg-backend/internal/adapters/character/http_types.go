package character

import (
	"time"

	"github.com/mmorpg-template/backend/internal/domain/character"
)

// CreateCharacterRequest represents the HTTP request for creating a character
type CreateCharacterRequest struct {
	Name       string                      `json:"name" binding:"required"`
	SlotNumber int                         `json:"slot_number" binding:"required,min=1,max=100"`
	ClassType  string                      `json:"class_type" binding:"required"`
	Race       string                      `json:"race" binding:"required"`
	Gender     string                      `json:"gender" binding:"required"`
	Appearance *CharacterAppearanceOptions `json:"appearance,omitempty"`
}

// CharacterAppearanceOptions represents optional appearance settings
type CharacterAppearanceOptions struct {
	FaceType        *int     `json:"face_type,omitempty"`
	SkinColor       *string  `json:"skin_color,omitempty"`
	EyeColor        *string  `json:"eye_color,omitempty"`
	HairStyle       *int     `json:"hair_style,omitempty"`
	HairColor       *string  `json:"hair_color,omitempty"`
	FacialHairStyle *int     `json:"facial_hair_style,omitempty"`
	FacialHairColor *string  `json:"facial_hair_color,omitempty"`
	BodyType        *int     `json:"body_type,omitempty"`
	Height          *float32 `json:"height,omitempty"`
}

// CharacterResponse represents the HTTP response for character data
type CharacterResponse struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	SlotNumber    int       `json:"slot_number"`
	Level         int       `json:"level"`
	Experience    int64     `json:"experience"`
	ClassType     string    `json:"class_type"`
	Race          string    `json:"race"`
	Gender        string    `json:"gender"`
	CreatedAt     time.Time `json:"created_at"`
	LastPlayedAt  time.Time `json:"last_played_at"`
	TotalPlayTime int64     `json:"total_play_time"` // in seconds
}

// UpdateAppearanceRequest represents the HTTP request for updating appearance
type UpdateAppearanceRequest struct {
	FaceType         *int                        `json:"face_type,omitempty"`
	SkinColor        *string                     `json:"skin_color,omitempty"`
	EyeColor         *string                     `json:"eye_color,omitempty"`
	HairStyle        *int                        `json:"hair_style,omitempty"`
	HairColor        *string                     `json:"hair_color,omitempty"`
	FacialHairStyle  *int                        `json:"facial_hair_style,omitempty"`
	FacialHairColor  *string                     `json:"facial_hair_color,omitempty"`
	BodyType         *int                        `json:"body_type,omitempty"`
	Height           *float32                    `json:"height,omitempty"`
	BodyProportions  *character.BodyProportions  `json:"body_proportions,omitempty"`
	Scars            []int                       `json:"scars,omitempty"`
	Tattoos          []int                       `json:"tattoos,omitempty"`
	Accessories      []int                       `json:"accessories,omitempty"`
}

// AppearanceResponse represents the HTTP response for appearance data
type AppearanceResponse struct {
	FaceType        int                       `json:"face_type"`
	SkinColor       string                    `json:"skin_color"`
	EyeColor        string                    `json:"eye_color"`
	HairStyle       int                       `json:"hair_style"`
	HairColor       string                    `json:"hair_color"`
	FacialHairStyle int                       `json:"facial_hair_style"`
	FacialHairColor string                    `json:"facial_hair_color"`
	BodyType        int                       `json:"body_type"`
	Height          float32                   `json:"height"`
	BodyProportions character.BodyProportions `json:"body_proportions"`
	Scars           []int                     `json:"scars"`
	Tattoos         []int                     `json:"tattoos"`
	Accessories     []int                     `json:"accessories"`
}

// StatsResponse represents the HTTP response for character stats
type StatsResponse struct {
	// Primary stats
	Strength     int `json:"strength"`
	Dexterity    int `json:"dexterity"`
	Intelligence int `json:"intelligence"`
	Wisdom       int `json:"wisdom"`
	Constitution int `json:"constitution"`
	Charisma     int `json:"charisma"`
	// Combat stats
	HealthCurrent  int `json:"health_current"`
	HealthMax      int `json:"health_max"`
	ManaCurrent    int `json:"mana_current"`
	ManaMax        int `json:"mana_max"`
	StaminaCurrent int `json:"stamina_current"`
	StaminaMax     int `json:"stamina_max"`
	// Derived stats
	AttackPower    int     `json:"attack_power"`
	SpellPower     int     `json:"spell_power"`
	Defense        int     `json:"defense"`
	CriticalChance float32 `json:"critical_chance"`
	CriticalDamage float32 `json:"critical_damage"`
	DodgeChance    float32 `json:"dodge_chance"`
	BlockChance    float32 `json:"block_chance"`
	// Movement and other stats
	MovementSpeed float32 `json:"movement_speed"`
	AttackSpeed   float32 `json:"attack_speed"`
	CastSpeed     float32 `json:"cast_speed"`
	// Regeneration
	HealthRegen  float32 `json:"health_regen"`
	ManaRegen    float32 `json:"mana_regen"`
	StaminaRegen float32 `json:"stamina_regen"`
	// Points
	StatPointsAvailable  int `json:"stat_points_available"`
	SkillPointsAvailable int `json:"skill_points_available"`
}

// PositionResponse represents the HTTP response for character position
type PositionResponse struct {
	WorldID       string  `json:"world_id"`
	ZoneID        string  `json:"zone_id"`
	MapID         string  `json:"map_id"`
	PositionX     float64 `json:"position_x"`
	PositionY     float64 `json:"position_y"`
	PositionZ     float64 `json:"position_z"`
	RotationPitch float32 `json:"rotation_pitch"`
	RotationYaw   float32 `json:"rotation_yaw"`
	RotationRoll  float32 `json:"rotation_roll"`
	VelocityX     float32 `json:"velocity_x"`
	VelocityY     float32 `json:"velocity_y"`
	VelocityZ     float32 `json:"velocity_z"`
}

// AllocateStatPointRequest represents the HTTP request for allocating a stat point
type AllocateStatPointRequest struct {
	Stat   string `json:"stat" binding:"required,oneof=strength dexterity intelligence wisdom constitution charisma"`
	Points int    `json:"points,omitempty"` // Optional, defaults to 1
}

// DeletedCharacterResponse represents a soft-deleted character
type DeletedCharacterResponse struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	SlotNumber      int       `json:"slot_number"`
	Level           int       `json:"level"`
	ClassType       string    `json:"class_type"`
	Race            string    `json:"race"`
	Gender          string    `json:"gender"`
	DeletedAt       time.Time `json:"deleted_at"`
	CanRestoreUntil time.Time `json:"can_restore_until"`
	DaysRemaining   int       `json:"days_remaining"`
}

// SelectCharacterResponse represents the response for character selection
type SelectCharacterResponse struct {
	Success       bool          `json:"success"`
	CharacterID   string        `json:"character_id"`
	WorldServer   string        `json:"world_server"`
	WorldPort     int           `json:"world_port"`
	SessionToken  string        `json:"session_token"`
	SpawnLocation SpawnLocation `json:"spawn_location"`
}

// SpawnLocation represents where the character will spawn
type SpawnLocation struct {
	WorldID  string  `json:"world_id"`
	ZoneID   string  `json:"zone_id"`
	MapID    string  `json:"map_id"`
	Position Vector3 `json:"position"`
}

// Vector3 represents a 3D position
type Vector3 struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

// PermanentDeleteRequest represents the request for permanent character deletion
type PermanentDeleteRequest struct {
	ConfirmationCode string `json:"confirmation_code" binding:"required"`
	Password         string `json:"password" binding:"required"`
}

// AllocateStatResponse represents the response for stat point allocation
type AllocateStatResponse struct {
	Success         bool                   `json:"success"`
	Stat            string                 `json:"stat"`
	NewValue        int                    `json:"new_value"`
	PointsRemaining int                    `json:"points_remaining"`
	AffectedStats   map[string]interface{} `json:"affected_stats"`
}