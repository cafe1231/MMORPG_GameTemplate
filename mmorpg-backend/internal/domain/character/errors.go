package character

import "errors"

// Character-related errors
var (
	// Character errors
	ErrCharacterNotFound         = errors.New("character not found")
	ErrCharacterNameTaken        = errors.New("character name is already taken")
	ErrCharacterLimitReached     = errors.New("character limit reached for this account")
	ErrInvalidCharacterName      = errors.New("invalid character name")
	ErrCharacterDeleted          = errors.New("character is deleted")
	ErrCharacterCannotBeRestored = errors.New("character cannot be restored")
	ErrInvalidSlotNumber         = errors.New("invalid slot number")
	ErrSlotOccupied              = errors.New("character slot is already occupied")
	ErrCharacterBelongsToOther   = errors.New("character belongs to another user")
	
	// Class/Race/Gender errors
	ErrInvalidClass  = errors.New("invalid character class")
	ErrInvalidRace   = errors.New("invalid character race")
	ErrInvalidGender = errors.New("invalid character gender")
	
	// Appearance errors
	ErrAppearanceNotFound      = errors.New("character appearance not found")
	ErrInvalidFaceType         = errors.New("invalid face type")
	ErrInvalidSkinColor        = errors.New("invalid skin color")
	ErrInvalidEyeColor         = errors.New("invalid eye color")
	ErrInvalidHairStyle        = errors.New("invalid hair style")
	ErrInvalidHairColor        = errors.New("invalid hair color")
	ErrInvalidFacialHairStyle  = errors.New("invalid facial hair style")
	ErrInvalidFacialHairColor  = errors.New("invalid facial hair color")
	ErrInvalidBodyType         = errors.New("invalid body type")
	ErrInvalidHeight           = errors.New("invalid height value")
	
	// Stats errors
	ErrStatsNotFound          = errors.New("character stats not found")
	ErrNoStatPointsAvailable  = errors.New("no stat points available")
	ErrInvalidStatType        = errors.New("invalid stat type")
	ErrInsufficientMana       = errors.New("insufficient mana")
	ErrInsufficientStamina    = errors.New("insufficient stamina")
	ErrStatMaxReached         = errors.New("stat maximum value reached")
	
	// Position errors
	ErrPositionNotFound    = errors.New("character position not found")
	ErrInvalidRotation     = errors.New("invalid rotation values")
	ErrPositionOutOfBounds = errors.New("position is out of world bounds")
	ErrInvalidWorldID      = errors.New("invalid world ID")
	ErrInvalidZoneID       = errors.New("invalid zone ID")
	ErrCannotTeleport      = errors.New("cannot teleport at this time")
	
	// General errors
	ErrInvalidCharacterID  = errors.New("invalid character ID")
	ErrInvalidUserID       = errors.New("invalid user ID")
	ErrUnauthorized        = errors.New("unauthorized access")
	ErrNoCharacterSelected = errors.New("no character selected")
)