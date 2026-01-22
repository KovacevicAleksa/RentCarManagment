package auth

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.Engine, service *AuthService) {
	authGroup := r.Group("/auth")
	{
		authGroup.POST("/register", RegisterHandler(service))
		authGroup.POST("/login", LoginHandler(service))
		authGroup.POST("/logout", LogoutHandler())
		
		authGroup.GET("/me", AuthMiddleware(), MeHandler())
	}
}