package character

import (
	"time"

	"github.com/google/uuid"
)

// Character represents a player character in the game
type Character struct {
	ID                   uuid.UUID
	UserID               uuid.UUID
	Name                 string
	SlotNumber           int
	Level                int
	Experience           int64
	ClassType            ClassType
	Race                 Race
	Gender               Gender
	IsDeleted            bool
	DeletedAt            *time.Time
	DeletionScheduledAt  *time.Time
	CreatedAt            time.Time
	UpdatedAt            time.Time
	LastPlayedAt         time.Time
	LastSelectedAt       time.Time
	TotalPlayTime        time.Duration
}

// ClassType represents the character class
type ClassType string

const (
	ClassWarrior  ClassType = "warrior"
	ClassMage     ClassType = "mage"
	ClassRogue    ClassType = "rogue"
	ClassPriest   ClassType = "priest"
	ClassRanger   ClassType = "ranger"
	ClassPaladin  ClassType = "paladin"
	ClassWarlock  ClassType = "warlock"
	ClassDruid    ClassType = "druid"
)

// Race represents the character race
type Race string

const (
	RaceHuman  Race = "human"
	RaceElf    Race = "elf"
	RaceDwarf  Race = "dwarf"
	RaceOrc    Race = "orc"
	RaceGnome  Race = "gnome"
	RaceTroll  Race = "troll"
	RaceUndead Race = "undead"
)

// Gender represents the character gender
type Gender string

const (
	GenderMale   Gender = "male"
	GenderFemale Gender = "female"
	GenderOther  Gender = "other"
)

// NewCharacter creates a new character with default values
func NewCharacter(userID uuid.UUID, name string, slotNumber int, classType ClassType, race Race, gender Gender) *Character {
	now := time.Now()
	return &Character{
		ID:            uuid.New(),
		UserID:        userID,
		Name:          name,
		SlotNumber:    slotNumber,
		Level:         1,
		Experience:    0,
		ClassType:     classType,
		Race:          race,
		Gender:        gender,
		IsDeleted:     false,
		DeletedAt:     nil,
		DeletionScheduledAt: nil,
		CreatedAt:      now,
		UpdatedAt:      now,
		LastPlayedAt:   now,
		LastSelectedAt: now,
		TotalPlayTime:  0,
	}
}

// CanBeRestored checks if a deleted character can be restored
func (c *Character) CanBeRestored() bool {
	if !c.IsDeleted {
		return false
	}
	if c.DeletionScheduledAt == nil {
		return false
	}
	return time.Now().Before(*c.DeletionScheduledAt)
}

// SoftDelete marks the character as deleted with a 30-day recovery period
func (c *Character) SoftDelete() {
	now := time.Now()
	deletionTime := now.Add(30 * 24 * time.Hour)
	c.IsDeleted = true
	c.DeletedAt = &now
	c.DeletionScheduledAt = &deletionTime
	c.UpdatedAt = now
}

// Restore restores a soft-deleted character
func (c *Character) Restore() error {
	if !c.CanBeRestored() {
		return ErrCharacterCannotBeRestored
	}
	c.IsDeleted = false
	c.DeletedAt = nil
	c.DeletionScheduledAt = nil
	c.UpdatedAt = time.Now()
	return nil
}

// UpdatePlayTime updates the character's play time and last played timestamp
func (c *Character) UpdatePlayTime(sessionDuration time.Duration) {
	c.TotalPlayTime += sessionDuration
	c.LastPlayedAt = time.Now()
	c.UpdatedAt = time.Now()
}

// IsValidClass checks if the class type is valid
func IsValidClass(class ClassType) bool {
	switch class {
	case ClassWarrior, ClassMage, ClassRogue, ClassPriest, 
		 ClassRanger, ClassPaladin, ClassWarlock, ClassDruid:
		return true
	default:
		return false
	}
}

// IsValidRace checks if the race is valid
func IsValidRace(race Race) bool {
	switch race {
	case RaceHuman, RaceElf, RaceDwarf, RaceOrc, 
		 RaceGnome, RaceTroll, RaceUndead:
		return true
	default:
		return false
	}
}

// IsValidGender checks if the gender is valid
func IsValidGender(gender Gender) bool {
	switch gender {
	case GenderMale, GenderFemale, GenderOther:
		return true
	default:
		return false
	}
}

// GetExperienceForLevel returns the total experience required for a specific level
func GetExperienceForLevel(level int) int64 {
	if level <= 1 {
		return 0
	}
	// Simple exponential formula for now
	// Can be adjusted based on game balance requirements
	return int64(100 * (level - 1) * (level - 1))
}

// CalculateLevel calculates the level based on total experience
func CalculateLevel(experience int64) int {
	level := 1
	for {
		requiredExp := GetExperienceForLevel(level + 1)
		if experience < requiredExp {
			break
		}
		level++
		if level >= 100 { // Max level cap
			break
		}
	}
	return level
}