package character

import (
	"math"
	"time"

	"github.com/google/uuid"
)

// Position represents a character's location in the game world
type Position struct {
	ID              uuid.UUID
	CharacterID     uuid.UUID
	// World location
	WorldID         string
	ZoneID          string
	MapID           string
	// 3D coordinates
	PositionX       float64
	PositionY       float64
	PositionZ       float64
	// Rotation (Pitch, Yaw, Roll in degrees)
	RotationPitch   float32
	RotationYaw     float32
	RotationRoll    float32
	// Velocity
	VelocityX       float32
	VelocityY       float32
	VelocityZ       float32
	// Instance information
	InstanceID      *uuid.UUID
	InstanceType    *string
	// Safe position for respawn
	SafePositionX   float64
	SafePositionY   float64
	SafePositionZ   float64
	SafeWorldID     string
	SafeZoneID      string
	// Metadata
	LastMovement    time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Vector3 represents a 3D vector
type Vector3 struct {
	X float64
	Y float64
	Z float64
}

// Rotation represents rotation in 3D space
type Rotation struct {
	Pitch float32
	Yaw   float32
	Roll  float32
}

// NewPosition creates a new character position with default starting location
func NewPosition(characterID uuid.UUID) *Position {
	now := time.Now()
	return &Position{
		ID:            uuid.New(),
		CharacterID:   characterID,
		WorldID:       "starter_zone",
		ZoneID:        "tutorial_area",
		MapID:         "main",
		PositionX:     0.0,
		PositionY:     0.0,
		PositionZ:     100.0, // Start slightly above ground
		RotationPitch: 0.0,
		RotationYaw:   0.0,
		RotationRoll:  0.0,
		VelocityX:     0.0,
		VelocityY:     0.0,
		VelocityZ:     0.0,
		InstanceID:    nil,
		InstanceType:  nil,
		SafePositionX: 0.0,
		SafePositionY: 0.0,
		SafePositionZ: 100.0,
		SafeWorldID:   "starter_zone",
		SafeZoneID:    "tutorial_area",
		LastMovement:  now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// GetPosition returns the current position as a Vector3
func (p *Position) GetPosition() Vector3 {
	return Vector3{
		X: p.PositionX,
		Y: p.PositionY,
		Z: p.PositionZ,
	}
}

// SetPosition updates the position
func (p *Position) SetPosition(x, y, z float64) {
	p.PositionX = x
	p.PositionY = y
	p.PositionZ = z
	p.LastMovement = time.Now()
	p.UpdatedAt = time.Now()
}

// GetRotation returns the current rotation
func (p *Position) GetRotation() Rotation {
	return Rotation{
		Pitch: p.RotationPitch,
		Yaw:   p.RotationYaw,
		Roll:  p.RotationRoll,
	}
}

// SetRotation updates the rotation
func (p *Position) SetRotation(pitch, yaw, roll float32) {
	p.RotationPitch = pitch
	p.RotationYaw = yaw
	p.RotationRoll = roll
	p.LastMovement = time.Now()
	p.UpdatedAt = time.Now()
}

// GetVelocity returns the current velocity as a Vector3
func (p *Position) GetVelocity() Vector3 {
	return Vector3{
		X: float64(p.VelocityX),
		Y: float64(p.VelocityY),
		Z: float64(p.VelocityZ),
	}
}

// SetVelocity updates the velocity
func (p *Position) SetVelocity(x, y, z float32) {
	p.VelocityX = x
	p.VelocityY = y
	p.VelocityZ = z
	p.UpdatedAt = time.Now()
}

// SaveSafePosition saves the current position as the safe position
func (p *Position) SaveSafePosition() {
	p.SafePositionX = p.PositionX
	p.SafePositionY = p.PositionY
	p.SafePositionZ = p.PositionZ
	p.SafeWorldID = p.WorldID
	p.SafeZoneID = p.ZoneID
	p.UpdatedAt = time.Now()
}

// TeleportToSafePosition teleports the character to their safe position
func (p *Position) TeleportToSafePosition() {
	p.PositionX = p.SafePositionX
	p.PositionY = p.SafePositionY
	p.PositionZ = p.SafePositionZ
	p.WorldID = p.SafeWorldID
	p.ZoneID = p.SafeZoneID
	p.VelocityX = 0
	p.VelocityY = 0
	p.VelocityZ = 0
	p.InstanceID = nil
	p.InstanceType = nil
	p.LastMovement = time.Now()
	p.UpdatedAt = time.Now()
}

// EnterInstance sets the instance information
func (p *Position) EnterInstance(instanceID uuid.UUID, instanceType string) {
	p.InstanceID = &instanceID
	p.InstanceType = &instanceType
	p.UpdatedAt = time.Now()
}

// LeaveInstance clears the instance information
func (p *Position) LeaveInstance() {
	p.InstanceID = nil
	p.InstanceType = nil
	p.UpdatedAt = time.Now()
}

// IsInInstance checks if the character is in an instance
func (p *Position) IsInInstance() bool {
	return p.InstanceID != nil
}

// DistanceTo calculates the distance to another position
func (p *Position) DistanceTo(other *Position) float64 {
	// Only calculate distance if in the same world/zone/instance
	if p.WorldID != other.WorldID || p.ZoneID != other.ZoneID {
		return math.MaxFloat64
	}
	
	// Check instance match
	if (p.InstanceID == nil && other.InstanceID != nil) ||
		(p.InstanceID != nil && other.InstanceID == nil) ||
		(p.InstanceID != nil && other.InstanceID != nil && *p.InstanceID != *other.InstanceID) {
		return math.MaxFloat64
	}

	dx := p.PositionX - other.PositionX
	dy := p.PositionY - other.PositionY
	dz := p.PositionZ - other.PositionZ
	
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}

// DistanceToPoint calculates the distance to a specific point
func (p *Position) DistanceToPoint(x, y, z float64) float64 {
	dx := p.PositionX - x
	dy := p.PositionY - y
	dz := p.PositionZ - z
	
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}

// IsNearby checks if another position is within a certain distance
func (p *Position) IsNearby(other *Position, maxDistance float64) bool {
	return p.DistanceTo(other) <= maxDistance
}

// Validate checks if the position values are valid
func (p *Position) Validate() error {
	// Check rotation bounds
	if p.RotationPitch < -90 || p.RotationPitch > 90 {
		return ErrInvalidRotation
	}
	if p.RotationYaw < -180 || p.RotationYaw > 180 {
		return ErrInvalidRotation
	}
	if p.RotationRoll < -180 || p.RotationRoll > 180 {
		return ErrInvalidRotation
	}
	
	// Check position bounds (can be adjusted based on world size)
	if p.PositionX < -1000000 || p.PositionX > 1000000 ||
		p.PositionY < -1000000 || p.PositionY > 1000000 ||
		p.PositionZ < -10000 || p.PositionZ > 50000 {
		return ErrPositionOutOfBounds
	}
	
	return nil
}

// ApplyClassStartingPosition sets class-specific starting positions
func (p *Position) ApplyClassStartingPosition(classType ClassType) {
	// Different classes might start in different areas
	switch classType {
	case ClassWarrior, ClassPaladin:
		p.WorldID = "starter_zone"
		p.ZoneID = "warrior_training_grounds"
		p.PositionX = 100.0
		p.PositionY = 50.0
		p.PositionZ = 100.0
	case ClassMage, ClassWarlock:
		p.WorldID = "starter_zone"
		p.ZoneID = "arcane_academy"
		p.PositionX = -100.0
		p.PositionY = -50.0
		p.PositionZ = 150.0
	case ClassRogue:
		p.WorldID = "starter_zone"
		p.ZoneID = "shadow_alley"
		p.PositionX = 200.0
		p.PositionY = -100.0
		p.PositionZ = 80.0
	case ClassPriest, ClassDruid:
		p.WorldID = "starter_zone"
		p.ZoneID = "sacred_grove"
		p.PositionX = -200.0
		p.PositionY = 100.0
		p.PositionZ = 120.0
	case ClassRanger:
		p.WorldID = "starter_zone"
		p.ZoneID = "hunters_lodge"
		p.PositionX = 150.0
		p.PositionY = 150.0
		p.PositionZ = 110.0
	default:
		// Default starting position
		p.WorldID = "starter_zone"
		p.ZoneID = "tutorial_area"
		p.PositionX = 0.0
		p.PositionY = 0.0
		p.PositionZ = 100.0
	}
	
	// Save as safe position
	p.SaveSafePosition()
}