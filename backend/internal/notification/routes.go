package notification

import (
	"github.com/KovacevicAleksa/rentcar/backend/internal/auth"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, service *Service, tokens *auth.TokenService) {
	group := r.Group("/notifications", auth.AuthMiddleware(tokens))
	{
		group.GET("", ListHandler(service))
		group.GET("/unread-count", UnreadCountHandler(service))
		group.POST("/read", MarkReadHandler(service))
	}

	admin := r.Group("/admin/notifications", auth.AuthMiddleware(tokens), auth.RequireRole(auth.RoleAdmin))
	{
		admin.POST("", AdminCreateHandler(service))
	}
}
