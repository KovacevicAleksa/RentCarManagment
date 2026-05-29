package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates the JWT cookie and stores the user_id, email and
// role on the request context for downstream handlers.
func AuthMiddleware(tokens *TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("token")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized - no token"})
			return
		}

		claims, err := tokens.Parse(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// RequireRole aborts with 403 unless the context role (set by AuthMiddleware)
// matches the required role. Must run after AuthMiddleware.
func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		current, _ := c.Get("role")
		if current != role {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden - insufficient permissions"})
			return
		}
		c.Next()
	}
}
