package history

import (
	"github.com/KovacevicAleksa/rentcar/backend/internal/auth"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, service *HistoryService, tokens *auth.TokenService) {
	historyGroup := r.Group("/history")
	historyGroup.Use(auth.AuthMiddleware(tokens))
	{
		historyGroup.GET("/car/:carID", GetCarHistoryHandler(service))
		historyGroup.GET("/car/:carID/latest", GetLatestHandler(service))
		historyGroup.POST("/cleanup", CleanupHandler(service))
	}
}