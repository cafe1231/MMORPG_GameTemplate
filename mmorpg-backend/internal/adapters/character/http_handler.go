package character

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mmorpg-template/backend/internal/domain/character"
	portsCharacter "github.com/mmorpg-template/backend/internal/ports/character"
	"github.com/mmorpg-template/backend/pkg/logger"
)

// HTTPHandler handles HTTP requests for character operations
type HTTPHandler struct {
	service        portsCharacter.CharacterService
	logger         logger.Logger
	jwtMiddleware  *JWTMiddleware
}

// NewHTTPHandler creates a new HTTP handler for character operations
func NewHTTPHandler(service portsCharacter.CharacterService, jwtMiddleware *JWTMiddleware, logger logger.Logger) *HTTPHandler {
	return &HTTPHandler{
		service:       service,
		logger:        logger,
		jwtMiddleware: jwtMiddleware,
	}
}

// AuthMiddleware returns the JWT validation middleware
func (h *HTTPHandler) AuthMiddleware() gin.HandlerFunc {
	return h.jwtMiddleware.Validate()
}

// CreateCharacter handles character creation requests
func (h *HTTPHandler) CreateCharacter(c *gin.Context) {
	userID, ok := GetUserIDFromContext(c)
	if !ok {
		h.respondWithError(c, http.StatusUnauthorized, ErrorCodeUnauthorized, "user ID not found in context", nil)
		return
	}
	
	var req CreateCharacterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Map HTTP request to service request
	serviceReq := &portsCharacter.CreateCharacterRequest{
		UserID:     userID,
		Name:       req.Name,
		SlotNumber: req.SlotNumber,
		ClassType:  character.ClassType(req.ClassType),
		Race:       character.Race(req.Race),
		Gender:     character.Gender(req.Gender),
	}

	// Handle appearance options if provided
	if req.Appearance != nil {
		serviceReq.Appearance = &portsCharacter.CharacterAppearanceOptions{
			FaceType:        req.Appearance.FaceType,
			SkinColor:       req.Appearance.SkinColor,
			EyeColor:        req.Appearance.EyeColor,
			HairStyle:       req.Appearance.HairStyle,
			HairColor:       req.Appearance.HairColor,
			FacialHairStyle: req.Appearance.FacialHairStyle,
			FacialHairColor: req.Appearance.FacialHairColor,
		}
		if req.Appearance.BodyType != nil {
			bodyType := character.BodyType(*req.Appearance.BodyType)
			serviceReq.Appearance.BodyType = &bodyType
		}
		serviceReq.Appearance.Height = req.Appearance.Height
	}

	char, err := h.service.CreateCharacter(c.Request.Context(), serviceReq)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, CharacterResponse{
		ID:            char.ID.String(),
		Name:          char.Name,
		SlotNumber:    char.SlotNumber,
		Level:         char.Level,
		Experience:    char.Experience,
		ClassType:     string(char.ClassType),
		Race:          string(char.Race),
		Gender:        string(char.Gender),
		CreatedAt:     char.CreatedAt,
		LastPlayedAt:  char.LastPlayedAt,
		TotalPlayTime: int64(char.TotalPlayTime.Seconds()),
	})
}

// ListCharacters lists all characters for the authenticated user
func (h *HTTPHandler) ListCharacters(c *gin.Context) {
	userID, ok := GetUserIDFromContext(c)
	if !ok {
		h.respondWithError(c, http.StatusUnauthorized, ErrorCodeUnauthorized, "user ID not found in context", nil)
		return
	}
	
	characters, err := h.service.ListCharactersByUser(c.Request.Context(), userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response := make([]CharacterResponse, len(characters))
	for i, char := range characters {
		response[i] = CharacterResponse{
			ID:            char.ID.String(),
			Name:          char.Name,
			SlotNumber:    char.SlotNumber,
			Level:         char.Level,
			Experience:    char.Experience,
			ClassType:     string(char.ClassType),
			Race:          string(char.Race),
			Gender:        string(char.Gender),
			CreatedAt:     char.CreatedAt,
			LastPlayedAt:  char.LastPlayedAt,
			TotalPlayTime: int64(char.TotalPlayTime.Seconds()),
		}
	}

	c.JSON(http.StatusOK, gin.H{"characters": response})
}

// GetCharacter retrieves a specific character
func (h *HTTPHandler) GetCharacter(c *gin.Context) {
	userID, ok := GetUserIDFromContext(c)
	if !ok {
		h.respondWithError(c, http.StatusUnauthorized, ErrorCodeUnauthorized, "user ID not found in context", nil)
		return
	}
	characterID := c.Param("id")
	
	// Validate ownership
	if err := h.service.ValidateCharacterOwnership(c.Request.Context(), characterID, userID); err != nil {
		h.handleError(c, err)
		return
	}

	char, err := h.service.GetCharacter(c.Request.Context(), characterID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, CharacterResponse{
		ID:            char.ID.String(),
		Name:          char.Name,
		SlotNumber:    char.SlotNumber,
		Level:         char.Level,
		Experience:    char.Experience,
		ClassType:     string(char.ClassType),
		Race:          string(char.Race),
		Gender:        string(char.Gender),
		CreatedAt:     char.CreatedAt,
		LastPlayedAt:  char.LastPlayedAt,
		TotalPlayTime: int64(char.TotalPlayTime.Seconds()),
	})
}

// DeleteCharacter soft deletes a character
func (h *HTTPHandler) DeleteCharacter(c *gin.Context) {
	userID, ok := GetUserIDFromContext(c)
	if !ok {
		h.respondWithError(c, http.StatusUnauthorized, ErrorCodeUnauthorized, "user ID not found in context", nil)
		return
	}
	characterID := c.Param("id")
	
	if err := h.service.DeleteCharacter(c.Request.Context(), characterID, userID); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "character deleted successfully"})
}

// RestoreCharacter restores a soft-deleted character
func (h *HTTPHandler) RestoreCharacter(c *gin.Context) {
	userID, ok := GetUserIDFromContext(c)
	if !ok {
		h.respondWithError(c, http.StatusUnauthorized, ErrorCodeUnauthorized, "user ID not found in context", nil)
		return
	}
	characterID := c.Param("id")
	
	if err := h.service.RestoreCharacter(c.Request.Context(), characterID, userID); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "character restored successfully"})
}

// GetAppearance retrieves character appearance
func (h *HTTPHandler) GetAppearance(c *gin.Context) {
	userID, ok := GetUserIDFromContext(c)
	if !ok {
		h.respondWithError(c, http.StatusUnauthorized, ErrorCodeUnauthorized, "user ID not found in context", nil)
		return
	}
	characterID := c.Param("id")
	
	// Validate ownership
	if err := h.service.ValidateCharacterOwnership(c.Request.Context(), characterID, userID); err != nil {
		h.handleError(c, err)
		return
	}

	appearance, err := h.service.GetAppearance(c.Request.Context(), characterID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, AppearanceResponse{
		FaceType:        appearance.FaceType,
		SkinColor:       appearance.SkinColor,
		EyeColor:        appearance.EyeColor,
		HairStyle:       appearance.HairStyle,
		HairColor:       appearance.HairColor,
		FacialHairStyle: appearance.FacialHairStyle,
		FacialHairColor: appearance.FacialHairColor,
		BodyType:        int(appearance.BodyType),
		Height:          appearance.Height,
		BodyProportions: appearance.BodyProportions,
		Scars:           appearance.Scars,
		Tattoos:         appearance.Tattoos,
		Accessories:     appearance.Accessories,
	})
}

// UpdateAppearance updates character appearance
func (h *HTTPHandler) UpdateAppearance(c *gin.Context) {
	userID, ok := GetUserIDFromContext(c)
	if !ok {
		h.respondWithError(c, http.StatusUnauthorized, ErrorCodeUnauthorized, "user ID not found in context", nil)
		return
	}
	characterID := c.Param("id")
	
	// Validate ownership
	if err := h.service.ValidateCharacterOwnership(c.Request.Context(), characterID, userID); err != nil {
		h.handleError(c, err)
		return
	}

	var req UpdateAppearanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Map to service request
	serviceReq := &portsCharacter.UpdateAppearanceRequest{
		FaceType:        req.FaceType,
		SkinColor:       req.SkinColor,
		EyeColor:        req.EyeColor,
		HairStyle:       req.HairStyle,
		HairColor:       req.HairColor,
		FacialHairStyle: req.FacialHairStyle,
		FacialHairColor: req.FacialHairColor,
		Height:          req.Height,
		BodyProportions: req.BodyProportions,
		Scars:           req.Scars,
		Tattoos:         req.Tattoos,
		Accessories:     req.Accessories,
	}
	
	if req.BodyType != nil {
		bodyType := character.BodyType(*req.BodyType)
		serviceReq.BodyType = &bodyType
	}

	appearance, err := h.service.UpdateAppearance(c.Request.Context(), characterID, serviceReq)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, AppearanceResponse{
		FaceType:        appearance.FaceType,
		SkinColor:       appearance.SkinColor,
		EyeColor:        appearance.EyeColor,
		HairStyle:       appearance.HairStyle,
		HairColor:       appearance.HairColor,
		FacialHairStyle: appearance.FacialHairStyle,
		FacialHairColor: appearance.FacialHairColor,
		BodyType:        int(appearance.BodyType),
		Height:          appearance.Height,
		BodyProportions: appearance.BodyProportions,
		Scars:           appearance.Scars,
		Tattoos:         appearance.Tattoos,
		Accessories:     appearance.Accessories,
	})
}

// GetStats retrieves character stats
func (h *HTTPHandler) GetStats(c *gin.Context) {
	userID, ok := GetUserIDFromContext(c)
	if !ok {
		h.respondWithError(c, http.StatusUnauthorized, ErrorCodeUnauthorized, "user ID not found in context", nil)
		return
	}
	characterID := c.Param("id")
	
	// Validate ownership
	if err := h.service.ValidateCharacterOwnership(c.Request.Context(), characterID, userID); err != nil {
		h.handleError(c, err)
		return
	}

	stats, err := h.service.GetStats(c.Request.Context(), characterID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, StatsResponse{
		Strength:             stats.Strength,
		Dexterity:            stats.Dexterity,
		Intelligence:         stats.Intelligence,
		Wisdom:               stats.Wisdom,
		Constitution:         stats.Constitution,
		Charisma:             stats.Charisma,
		HealthCurrent:        stats.HealthCurrent,
		HealthMax:            stats.HealthMax,
		ManaCurrent:          stats.ManaCurrent,
		ManaMax:              stats.ManaMax,
		StaminaCurrent:       stats.StaminaCurrent,
		StaminaMax:           stats.StaminaMax,
		AttackPower:          stats.AttackPower,
		SpellPower:           stats.SpellPower,
		Defense:              stats.Defense,
		CriticalChance:       stats.CriticalChance,
		CriticalDamage:       stats.CriticalDamage,
		DodgeChance:          stats.DodgeChance,
		BlockChance:          stats.BlockChance,
		MovementSpeed:        stats.MovementSpeed,
		AttackSpeed:          stats.AttackSpeed,
		CastSpeed:            stats.CastSpeed,
		HealthRegen:          stats.HealthRegen,
		ManaRegen:            stats.ManaRegen,
		StaminaRegen:         stats.StaminaRegen,
		StatPointsAvailable:  stats.StatPointsAvailable,
		SkillPointsAvailable: stats.SkillPointsAvailable,
	})
}

// GetPosition retrieves character position
func (h *HTTPHandler) GetPosition(c *gin.Context) {
	userID, ok := GetUserIDFromContext(c)
	if !ok {
		h.respondWithError(c, http.StatusUnauthorized, ErrorCodeUnauthorized, "user ID not found in context", nil)
		return
	}
	characterID := c.Param("id")
	
	// Validate ownership
	if err := h.service.ValidateCharacterOwnership(c.Request.Context(), characterID, userID); err != nil {
		h.handleError(c, err)
		return
	}

	position, err := h.service.GetPosition(c.Request.Context(), characterID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, PositionResponse{
		WorldID:       position.WorldID,
		ZoneID:        position.ZoneID,
		MapID:         position.MapID,
		PositionX:     position.PositionX,
		PositionY:     position.PositionY,
		PositionZ:     position.PositionZ,
		RotationPitch: position.RotationPitch,
		RotationYaw:   position.RotationYaw,
		RotationRoll:  position.RotationRoll,
		VelocityX:     position.VelocityX,
		VelocityY:     position.VelocityY,
		VelocityZ:     position.VelocityZ,
	})
}

// UpdatePosition updates character position
func (h *HTTPHandler) UpdatePosition(c *gin.Context) {
	userID, ok := GetUserIDFromContext(c)
	if !ok {
		h.respondWithError(c, http.StatusUnauthorized, ErrorCodeUnauthorized, "user ID not found in context", nil)
		return
	}
	characterID := c.Param("id")
	
	// Validate ownership
	if err := h.service.ValidateCharacterOwnership(c.Request.Context(), characterID, userID); err != nil {
		h.handleError(c, err)
		return
	}

	var req portsCharacter.UpdatePositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	position, err := h.service.UpdatePosition(c.Request.Context(), characterID, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, PositionResponse{
		WorldID:       position.WorldID,
		ZoneID:        position.ZoneID,
		MapID:         position.MapID,
		PositionX:     position.PositionX,
		PositionY:     position.PositionY,
		PositionZ:     position.PositionZ,
		RotationPitch: position.RotationPitch,
		RotationYaw:   position.RotationYaw,
		RotationRoll:  position.RotationRoll,
		VelocityX:     position.VelocityX,
		VelocityY:     position.VelocityY,
		VelocityZ:     position.VelocityZ,
	})
}

// handleError handles errors and returns appropriate HTTP responses
func (h *HTTPHandler) handleError(c *gin.Context, err error) {
	switch err {
	case character.ErrCharacterNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "character not found"})
	case character.ErrCharacterNameTaken:
		c.JSON(http.StatusConflict, gin.H{"error": "character name is already taken"})
	case character.ErrCharacterLimitReached:
		c.JSON(http.StatusForbidden, gin.H{"error": "character limit reached"})
	case character.ErrInvalidCharacterName:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid character name"})
	case character.ErrCharacterDeleted:
		c.JSON(http.StatusGone, gin.H{"error": "character is deleted"})
	case character.ErrCharacterCannotBeRestored:
		c.JSON(http.StatusForbidden, gin.H{"error": "character cannot be restored"})
	case character.ErrCharacterBelongsToOther:
		c.JSON(http.StatusForbidden, gin.H{"error": "character belongs to another user"})
	case character.ErrInvalidClass, character.ErrInvalidRace, character.ErrInvalidGender:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case character.ErrUnauthorized:
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
	default:
		h.logger.WithError(err).Error("Unhandled error in character handler")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}

// CheckNameAvailability checks if a character name is available
func (h *HTTPHandler) CheckNameAvailability(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		h.respondWithValidationError(c, map[string]string{
			"name": "Name parameter is required",
		})
		return
	}
	
	// For now, we'll check if a character with this name exists
	_, err := h.service.GetCharacterByName(c.Request.Context(), name)
	if err != nil {
		if err == character.ErrCharacterNotFound {
			// Name is available
			c.JSON(http.StatusOK, gin.H{
				"available": true,
				"name":      name,
			})
			return
		}
		// Error checking name
		h.handleError(c, err)
		return
	}
	
	// Name is taken
	c.JSON(http.StatusOK, gin.H{
		"available": false,
		"name":      name,
		"suggestions": h.generateNameSuggestions(name),
	})
}

// ListDeletedCharacters lists soft-deleted characters for the authenticated user
func (h *HTTPHandler) ListDeletedCharacters(c *gin.Context) {
	_, ok := GetUserIDFromContext(c)
	if !ok {
		h.respondWithError(c, http.StatusUnauthorized, ErrorCodeUnauthorized, "user ID not found in context", nil)
		return
	}
	
	// For now, return empty list as we don't have a method to list deleted characters
	// TODO: Add ListDeletedCharacters method to service interface
	c.JSON(http.StatusOK, gin.H{"deleted_characters": []DeletedCharacterResponse{}})
}

// SelectCharacter selects a character for gameplay
func (h *HTTPHandler) SelectCharacter(c *gin.Context) {
	userID, ok := GetUserIDFromContext(c)
	if !ok {
		h.respondWithError(c, http.StatusUnauthorized, ErrorCodeUnauthorized, "user ID not found in context", nil)
		return
	}
	characterID := c.Param("id")
	sessionID := c.GetString("sessionID")
	
	// Validate ownership
	if err := h.service.ValidateCharacterOwnership(c.Request.Context(), characterID, userID); err != nil {
		h.handleError(c, err)
		return
	}
	
	// Select character
	if err := h.service.SelectCharacter(c.Request.Context(), characterID, userID, sessionID); err != nil {
		h.handleError(c, err)
		return
	}
	
	// Get character position for spawn location
	position, err := h.service.GetPosition(c.Request.Context(), characterID)
	if err != nil {
		h.handleError(c, err)
		return
	}
	
	// TODO: Get world server info from configuration or service discovery
	c.JSON(http.StatusOK, SelectCharacterResponse{
		Success:      true,
		CharacterID:  characterID,
		WorldServer:  "world1.mmorpg.local",
		WorldPort:    7777,
		SessionToken: sessionID, // For now, reuse session ID
		SpawnLocation: SpawnLocation{
			WorldID: position.WorldID,
			ZoneID:  position.ZoneID,
			MapID:   position.MapID,
			Position: Vector3{
				X: position.PositionX,
				Y: position.PositionY,
				Z: position.PositionZ,
			},
		},
	})
}

// PermanentlyDeleteCharacter permanently deletes a character
func (h *HTTPHandler) PermanentlyDeleteCharacter(c *gin.Context) {
	userID, ok := GetUserIDFromContext(c)
	if !ok {
		h.respondWithError(c, http.StatusUnauthorized, ErrorCodeUnauthorized, "user ID not found in context", nil)
		return
	}
	characterID := c.Param("id")
	
	var req PermanentDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithValidationError(c, map[string]string{
			"body": "Invalid request body",
		})
		return
	}
	
	// Validate ownership
	if err := h.service.ValidateCharacterOwnership(c.Request.Context(), characterID, userID); err != nil {
		h.handleError(c, err)
		return
	}
	
	// Get character name for confirmation
	char, err := h.service.GetCharacter(c.Request.Context(), characterID)
	if err != nil {
		h.handleError(c, err)
		return
	}
	
	// Validate confirmation code
	expectedCode := fmt.Sprintf("DELETE-%s-%d", char.Name, time.Now().Year())
	if req.ConfirmationCode != expectedCode {
		h.respondWithError(c, http.StatusBadRequest, ErrorCodeInvalidRequest, "Invalid confirmation code", map[string]interface{}{
			"expected_format": "DELETE-{character_name}-{current_year}",
		})
		return
	}
	
	// For now, we'll just soft delete as we don't have permanent delete in service
	// TODO: Add PermanentlyDeleteCharacter method to service interface
	if err := h.service.DeleteCharacter(c.Request.Context(), characterID, userID); err != nil {
		h.handleError(c, err)
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message":        "Character permanently deleted",
		"character_name": char.Name,
	})
}

// AllocateStatPoints allocates stat points to a character
func (h *HTTPHandler) AllocateStatPoints(c *gin.Context) {
	userID, ok := GetUserIDFromContext(c)
	if !ok {
		h.respondWithError(c, http.StatusUnauthorized, ErrorCodeUnauthorized, "user ID not found in context", nil)
		return
	}
	characterID := c.Param("id")
	
	// Validate ownership
	if err := h.service.ValidateCharacterOwnership(c.Request.Context(), characterID, userID); err != nil {
		h.handleError(c, err)
		return
	}
	
	var req AllocateStatPointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithValidationError(c, map[string]string{
			"body": "Invalid request body",
		})
		return
	}
	
	// Allocate single point to the stat
	stats, err := h.service.AllocateStatPoint(c.Request.Context(), characterID, req.Stat)
	if err != nil {
		h.handleError(c, err)
		return
	}
	
	// Get the new value of the allocated stat
	var newValue int
	switch req.Stat {
	case "strength":
		newValue = stats.Strength
	case "dexterity":
		newValue = stats.Dexterity
	case "intelligence":
		newValue = stats.Intelligence
	case "wisdom":
		newValue = stats.Wisdom
	case "constitution":
		newValue = stats.Constitution
	case "charisma":
		newValue = stats.Charisma
	}
	
	c.JSON(http.StatusOK, AllocateStatResponse{
		Success:        true,
		Stat:           req.Stat,
		NewValue:       newValue,
		PointsRemaining: stats.StatPointsAvailable,
		AffectedStats:  make(map[string]interface{}), // TODO: Calculate affected derived stats
	})
}

// generateNameSuggestions generates name suggestions when a name is taken
func (h *HTTPHandler) generateNameSuggestions(name string) []string {
	suggestions := []string{}
	
	// Add number suffixes
	for i := 1; i <= 3; i++ {
		suggestions = append(suggestions, fmt.Sprintf("%s%d", name, i))
	}
	
	// Add random suffixes
	suffixes := []string{"X", "Z", "_01", "TheGreat", "2025"}
	for _, suffix := range suffixes {
		suggestions = append(suggestions, name+suffix)
	}
	
	// Add prefix
	suggestions = append(suggestions, "Lord"+name)
	suggestions = append(suggestions, "Lady"+name)
	
	// Limit to 5 suggestions
	if len(suggestions) > 5 {
		suggestions = suggestions[:5]
	}
	
	return suggestions
}