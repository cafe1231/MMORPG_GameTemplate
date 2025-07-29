package character

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Appearance represents the visual customization of a character
type Appearance struct {
	ID               uuid.UUID
	CharacterID      uuid.UUID
	FaceType         int
	SkinColor        string
	EyeColor         string
	HairStyle        int
	HairColor        string
	FacialHairStyle  int
	FacialHairColor  string
	BodyType         BodyType
	Height           float32
	BodyProportions  BodyProportions
	Scars            []int
	Tattoos          []int
	Accessories      []int
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// BodyType represents different body builds
type BodyType int

const (
	BodyTypeAthletic BodyType = 1
	BodyTypeMuscular BodyType = 2
	BodyTypeSlim     BodyType = 3
	BodyTypeAverage  BodyType = 4
	BodyTypeHeavy    BodyType = 5
)

// BodyProportions stores detailed body proportion adjustments
type BodyProportions struct {
	ShoulderWidth float32 `json:"shoulder_width"`
	ChestSize     float32 `json:"chest_size"`
	WaistSize     float32 `json:"waist_size"`
	HipSize       float32 `json:"hip_size"`
	ArmLength     float32 `json:"arm_length"`
	LegLength     float32 `json:"leg_length"`
	NeckLength    float32 `json:"neck_length"`
}

// DefaultBodyProportions returns default body proportions
func DefaultBodyProportions() BodyProportions {
	return BodyProportions{
		ShoulderWidth: 1.0,
		ChestSize:     1.0,
		WaistSize:     1.0,
		HipSize:       1.0,
		ArmLength:     1.0,
		LegLength:     1.0,
		NeckLength:    1.0,
	}
}

// NewAppearance creates a new character appearance with defaults
func NewAppearance(characterID uuid.UUID) *Appearance {
	now := time.Now()
	return &Appearance{
		ID:               uuid.New(),
		CharacterID:      characterID,
		FaceType:         1,
		SkinColor:        "#FFD4B2",
		EyeColor:         "#4B8BF5",
		HairStyle:        1,
		HairColor:        "#3B2F2F",
		FacialHairStyle:  0,
		FacialHairColor:  "#3B2F2F",
		BodyType:         BodyTypeAthletic,
		Height:           1.0,
		BodyProportions:  DefaultBodyProportions(),
		Scars:            []int{},
		Tattoos:          []int{},
		Accessories:      []int{},
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

// IsValidHexColor validates if a string is a valid hex color
func IsValidHexColor(color string) bool {
	if len(color) != 7 || color[0] != '#' {
		return false
	}
	for i := 1; i < 7; i++ {
		c := color[i]
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// Validate checks if the appearance values are valid
func (a *Appearance) Validate() error {
	if a.FaceType < 1 || a.FaceType > 20 {
		return ErrInvalidFaceType
	}
	if !IsValidHexColor(a.SkinColor) {
		return ErrInvalidSkinColor
	}
	if !IsValidHexColor(a.EyeColor) {
		return ErrInvalidEyeColor
	}
	if a.HairStyle < 0 || a.HairStyle > 50 {
		return ErrInvalidHairStyle
	}
	if !IsValidHexColor(a.HairColor) {
		return ErrInvalidHairColor
	}
	if a.FacialHairStyle < 0 || a.FacialHairStyle > 20 {
		return ErrInvalidFacialHairStyle
	}
	if a.FacialHairStyle > 0 && !IsValidHexColor(a.FacialHairColor) {
		return ErrInvalidFacialHairColor
	}
	if a.BodyType < 1 || a.BodyType > 5 {
		return ErrInvalidBodyType
	}
	if a.Height < 0.8 || a.Height > 1.2 {
		return ErrInvalidHeight
	}
	return nil
}

// ToJSON converts body proportions to JSON string
func (bp BodyProportions) ToJSON() (string, error) {
	data, err := json.Marshal(bp)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FromJSON parses body proportions from JSON string
func (bp *BodyProportions) FromJSON(data string) error {
	return json.Unmarshal([]byte(data), bp)
}

// ApplyRaceDefaults applies race-specific appearance defaults
func (a *Appearance) ApplyRaceDefaults(race Race) {
	switch race {
	case RaceHuman:
		a.SkinColor = "#FFD4B2"
		a.Height = 1.0
	case RaceElf:
		a.SkinColor = "#FFF0E0"
		a.Height = 1.05
		a.BodyType = BodyTypeSlim
	case RaceDwarf:
		a.SkinColor = "#F4C2A1"
		a.Height = 0.85
		a.BodyType = BodyTypeMuscular
	case RaceOrc:
		a.SkinColor = "#8FBC8F"
		a.Height = 1.1
		a.BodyType = BodyTypeMuscular
	case RaceGnome:
		a.SkinColor = "#FFE4C4"
		a.Height = 0.8
		a.BodyType = BodyTypeSlim
	case RaceTroll:
		a.SkinColor = "#87CEEB"
		a.Height = 1.15
		a.BodyType = BodyTypeAthletic
	case RaceUndead:
		a.SkinColor = "#C0C0C0"
		a.Height = 1.0
		a.BodyType = BodyTypeSlim
	}
}

// ApplyGenderDefaults applies gender-specific appearance defaults
func (a *Appearance) ApplyGenderDefaults(gender Gender) {
	props := a.BodyProportions
	switch gender {
	case GenderMale:
		props.ShoulderWidth = 1.1
		props.ChestSize = 1.0
		props.WaistSize = 0.95
		props.HipSize = 0.9
	case GenderFemale:
		props.ShoulderWidth = 0.9
		props.ChestSize = 1.0
		props.WaistSize = 0.85
		props.HipSize = 1.1
	case GenderOther:
		// Keep neutral proportions
		props = DefaultBodyProportions()
	}
	a.BodyProportions = props
}