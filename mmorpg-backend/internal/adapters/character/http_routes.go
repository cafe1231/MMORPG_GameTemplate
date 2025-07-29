package character

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all character HTTP routes
func (h *HTTPHandler) RegisterRoutes(router *gin.Engine) {
	// Apply auth middleware to all character routes
	characterGroup := router.Group("/api/v1/characters")
	
	// Public routes (no auth required)
	characterGroup.GET("/check-name", h.CheckNameAvailability)
	
	// Protected routes (auth required)
	protected := characterGroup.Group("")
	protected.Use(h.AuthMiddleware())
	{
		// Character management
		protected.GET("", h.ListCharacters)
		protected.POST("", h.CreateCharacter)
		protected.GET("/deleted", h.ListDeletedCharacters)
		
		// Single character operations
		protected.GET("/:id", h.GetCharacter)
		protected.DELETE("/:id", h.DeleteCharacter)
		protected.POST("/:id/restore", h.RestoreCharacter)
		protected.POST("/:id/select", h.SelectCharacter)
		protected.DELETE("/:id/permanent", h.PermanentlyDeleteCharacter)
		
		// Character appearance
		protected.GET("/:id/appearance", h.GetAppearance)
		protected.PUT("/:id/appearance", h.UpdateAppearance)
		
		// Character stats
		protected.GET("/:id/stats", h.GetStats)
		protected.POST("/:id/stats/allocate", h.AllocateStatPoints)
		
		// Character position
		protected.GET("/:id/position", h.GetPosition)
		protected.PUT("/:id/position", h.UpdatePosition)
	}
}

