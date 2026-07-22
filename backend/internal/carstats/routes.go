package carstats

import (
	"github.com/gin-gonic/gin"

	"github.com/KovacevicAleksa/rentcar/backend/internal/auth"
)

func RegisterRoutes(r *gin.Engine, service *Service, tokens *auth.TokenService) {
	g := r.Group("/carstats", auth.AuthMiddleware(tokens))
	{
		g.GET("/car/:carID/events", GetCarEventStatsHandler(service))
	}
}
