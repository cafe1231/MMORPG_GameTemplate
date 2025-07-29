package character

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mmorpg-template/backend/internal/domain/character"
)

// ErrorCode represents character-specific error codes
type ErrorCode string

const (
	// Character errors
	ErrorCodeCharacterNotFound       ErrorCode = "CHARACTER_NOT_FOUND"
	ErrorCodeCharacterNameTaken      ErrorCode = "CHARACTER_NAME_TAKEN"
	ErrorCodeCharacterLimitReached   ErrorCode = "CHARACTER_LIMIT_REACHED"
	ErrorCodeInvalidCharacterName    ErrorCode = "INVALID_CHARACTER_NAME"
	ErrorCodeCharacterDeleted        ErrorCode = "CHARACTER_DELETED"
	ErrorCodeCharacterCannotRestore  ErrorCode = "CHARACTER_CANNOT_RESTORE"
	ErrorCodeInvalidSlotNumber       ErrorCode = "INVALID_SLOT_NUMBER"
	ErrorCodeSlotOccupied            ErrorCode = "SLOT_OCCUPIED"
	ErrorCodeCharacterBelongsToOther ErrorCode = "CHARACTER_BELONGS_TO_OTHER"
	ErrorCodeCharacterOnline         ErrorCode = "CHARACTER_ONLINE"
	ErrorCodeCharacterInCombat       ErrorCode = "CHARACTER_IN_COMBAT"
	
	// Class/Race/Gender errors
	ErrorCodeInvalidClass  ErrorCode = "INVALID_CLASS"
	ErrorCodeInvalidRace   ErrorCode = "INVALID_RACE"
	ErrorCodeInvalidGender ErrorCode = "INVALID_GENDER"
	
	// Appearance errors
	ErrorCodeInvalidAppearance ErrorCode = "INVALID_APPEARANCE"
	ErrorCodeInvalidFaceType   ErrorCode = "INVALID_FACE_TYPE"
	ErrorCodeInvalidSkinColor  ErrorCode = "INVALID_SKIN_COLOR"
	ErrorCodeInvalidEyeColor   ErrorCode = "INVALID_EYE_COLOR"
	ErrorCodeInvalidHairStyle  ErrorCode = "INVALID_HAIR_STYLE"
	ErrorCodeInvalidHairColor  ErrorCode = "INVALID_HAIR_COLOR"
	ErrorCodeInvalidBodyType   ErrorCode = "INVALID_BODY_TYPE"
	ErrorCodeInvalidHeight     ErrorCode = "INVALID_HEIGHT"
	
	// Stats errors
	ErrorCodeNoStatPoints    ErrorCode = "NO_STAT_POINTS"
	ErrorCodeStatMaxReached  ErrorCode = "STAT_MAX_REACHED"
	ErrorCodeInvalidStatType ErrorCode = "INVALID_STAT_TYPE"
	
	// General errors
	ErrorCodeUnauthorized     ErrorCode = "UNAUTHORIZED"
	ErrorCodeInvalidRequest   ErrorCode = "INVALID_REQUEST"
	ErrorCodeInternalError    ErrorCode = "INTERNAL_ERROR"
	ErrorCodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
	ErrorCodeRateLimited      ErrorCode = "RATE_LIMITED"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
	Timestamp string `json:"timestamp"`
	RequestID string `json:"request_id,omitempty"`
}

// ErrorDetail contains error details
type ErrorDetail struct {
	Code    ErrorCode              `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// errorMapping maps domain errors to HTTP status codes and error codes
var errorMapping = map[error]struct {
	status int
	code   ErrorCode
}{
	// Character errors
	character.ErrCharacterNotFound:         {http.StatusNotFound, ErrorCodeCharacterNotFound},
	character.ErrCharacterNameTaken:        {http.StatusConflict, ErrorCodeCharacterNameTaken},
	character.ErrCharacterLimitReached:     {http.StatusForbidden, ErrorCodeCharacterLimitReached},
	character.ErrInvalidCharacterName:      {http.StatusBadRequest, ErrorCodeInvalidCharacterName},
	character.ErrCharacterDeleted:          {http.StatusGone, ErrorCodeCharacterDeleted},
	character.ErrCharacterCannotBeRestored: {http.StatusForbidden, ErrorCodeCharacterCannotRestore},
	character.ErrInvalidSlotNumber:         {http.StatusBadRequest, ErrorCodeInvalidSlotNumber},
	character.ErrSlotOccupied:              {http.StatusConflict, ErrorCodeSlotOccupied},
	character.ErrCharacterBelongsToOther:   {http.StatusForbidden, ErrorCodeCharacterBelongsToOther},
	
	// Class/Race/Gender errors
	character.ErrInvalidClass:  {http.StatusBadRequest, ErrorCodeInvalidClass},
	character.ErrInvalidRace:   {http.StatusBadRequest, ErrorCodeInvalidRace},
	character.ErrInvalidGender: {http.StatusBadRequest, ErrorCodeInvalidGender},
	
	// Appearance errors
	character.ErrInvalidFaceType:        {http.StatusBadRequest, ErrorCodeInvalidFaceType},
	character.ErrInvalidSkinColor:       {http.StatusBadRequest, ErrorCodeInvalidSkinColor},
	character.ErrInvalidEyeColor:        {http.StatusBadRequest, ErrorCodeInvalidEyeColor},
	character.ErrInvalidHairStyle:       {http.StatusBadRequest, ErrorCodeInvalidHairStyle},
	character.ErrInvalidHairColor:       {http.StatusBadRequest, ErrorCodeInvalidHairColor},
	character.ErrInvalidFacialHairStyle: {http.StatusBadRequest, ErrorCodeInvalidAppearance},
	character.ErrInvalidFacialHairColor: {http.StatusBadRequest, ErrorCodeInvalidAppearance},
	character.ErrInvalidBodyType:        {http.StatusBadRequest, ErrorCodeInvalidBodyType},
	character.ErrInvalidHeight:          {http.StatusBadRequest, ErrorCodeInvalidHeight},
	
	// Stats errors
	character.ErrNoStatPointsAvailable: {http.StatusBadRequest, ErrorCodeNoStatPoints},
	character.ErrStatMaxReached:        {http.StatusBadRequest, ErrorCodeStatMaxReached},
	character.ErrInvalidStatType:       {http.StatusBadRequest, ErrorCodeInvalidStatType},
	
	// General errors
	character.ErrUnauthorized: {http.StatusUnauthorized, ErrorCodeUnauthorized},
}

// HandleError processes errors and returns appropriate HTTP responses
func (h *HTTPHandler) HandleError(c *gin.Context, err error) {
	if mapping, ok := errorMapping[err]; ok {
		h.respondWithError(c, mapping.status, mapping.code, err.Error(), nil)
		return
	}
	
	// Default to internal server error
	h.logger.WithError(err).Error("Unhandled error in character handler")
	h.respondWithError(c, http.StatusInternalServerError, ErrorCodeInternalError, "Internal server error", nil)
}

// respondWithError sends a standardized error response
func (h *HTTPHandler) respondWithError(c *gin.Context, status int, code ErrorCode, message string, details map[string]interface{}) {
	response := ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
		Timestamp: time.Now().Format(time.RFC3339),
		RequestID: c.GetString("request_id"),
	}
	
	c.JSON(status, response)
}

// respondWithValidationError sends validation error with field details
func (h *HTTPHandler) respondWithValidationError(c *gin.Context, fieldErrors map[string]string) {
	details := make(map[string]interface{})
	for field, err := range fieldErrors {
		details[field] = err
	}
	
	h.respondWithError(c, http.StatusBadRequest, ErrorCodeInvalidRequest, "Validation failed", details)
}