package auth

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.Engine, service *AuthService, tokens *TokenService) {
	authGroup := r.Group("/auth")
	{
		authGroup.POST("/register", RegisterHandler(service))
		authGroup.POST("/login", LoginHandler(service))
		authGroup.POST("/logout", LogoutHandler())
		authGroup.GET("/me", AuthMiddleware(tokens), MeHandler())
		authGroup.PATCH("/password", AuthMiddleware(tokens), ChangePasswordHandler(service))
		authGroup.PATCH("/email", AuthMiddleware(tokens), ChangeEmailHandler(service, tokens))
	}

	adminGroup := r.Group("/admin", AuthMiddleware(tokens), RequireRole(RoleAdmin))
	{
		adminGroup.POST("/users", AdminCreateUserHandler(service))
		adminGroup.GET("/users", AdminListUsersHandler(service))
		adminGroup.POST("/users/:id/reset-password", AdminResetPasswordHandler(service))
		adminGroup.POST("/users/:id/approve", AdminApproveUserHandler(service))
		adminGroup.DELETE("/users/:id", AdminDeleteUserHandler(service))
	}
}
