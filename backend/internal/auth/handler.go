package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// setTokenCookie writes the auth cookie with a max-age derived from the token TTL.
func setTokenCookie(c *gin.Context, token string, ttl time.Duration) {
	c.SetCookie("token", token, int(ttl/time.Second), "/", "", false, true)
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func RegisterHandler(service *AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := service.Register(req.Email, req.Password); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
	}
}

func LoginHandler(service *AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		token, userID, err := service.Login(req.Email, req.Password)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		c.SetCookie(
			"token",                           
			token,                             
			int(time.Hour*72/time.Second),     
			"/",                               
			"",                                
			false,                             
			true,                              
		)

		c.JSON(http.StatusOK, gin.H{
			"message": "Login successful",
			"user_id": userID,
		})
	}
}

func LogoutHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.SetCookie(
			"token",
			"",
			-1,     
			"/",
			"",
			false,
			true,
		)
		c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
	}
}

func MeHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		email, _ := c.Get("email")
		role, _ := c.Get("role")

		c.JSON(http.StatusOK, gin.H{
			"user_id": userID,
			"email":   email,
			"role":    role,
		})
	}
}

type CreateUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"required"`
}

// UserResponse is the safe representation of a user (never exposes the hash).
type UserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

func toUserResponse(u User) UserResponse {
	return UserResponse{ID: u.ID, Email: u.Email, Role: u.Role}
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
}

type ChangeEmailRequest struct {
	NewEmail        string `json:"new_email" binding:"required,email"`
	CurrentPassword string `json:"current_password" binding:"required"`
}

func ChangePasswordHandler(service *AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ChangePasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userID := c.GetString("user_id")
		if err := service.ChangePassword(userID, req.CurrentPassword, req.NewPassword); err != nil {
			c.JSON(accountErrorStatus(err), gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
	}
}

func ChangeEmailHandler(service *AuthService, tokens *TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ChangeEmailRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userID := c.GetString("user_id")
		user, err := service.ChangeEmail(userID, req.NewEmail, req.CurrentPassword)
		if err != nil {
			c.JSON(accountErrorStatus(err), gin.H{"error": err.Error()})
			return
		}

		// Reissue the token so the email carried in the JWT/cookie stays current.
		token, err := tokens.Generate(user.ID, user.Email, user.Role)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to refresh session"})
			return
		}
		setTokenCookie(c, token, tokens.TTL())

		c.JSON(http.StatusOK, gin.H{"message": "Email updated successfully", "email": user.Email})
	}
}

// accountErrorStatus maps service errors to HTTP status codes for the
// self-service account endpoints.
func accountErrorStatus(err error) int {
	switch {
	case errors.Is(err, ErrUserNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrInvalidCredentials), errors.Is(err, ErrUserExists), errors.Is(err, ErrWeakPassword),
		errors.Is(err, ErrCannotDeleteSelf), errors.Is(err, ErrLastAdmin):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func AdminCreateUserHandler(service *AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := service.CreateUser(req.Email, req.Password, req.Role); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
	}
}

func AdminResetPasswordHandler(service *AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("id")
		temp, err := service.ResetPassword(userID)
		if err != nil {
			c.JSON(accountErrorStatus(err), gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message":            "Password reset successfully",
			"temporary_password": temp,
		})
	}
}

func AdminDeleteUserHandler(service *AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		actingUserID := c.GetString("user_id")
		targetID := c.Param("id")
		if err := service.DeleteUser(actingUserID, targetID); err != nil {
			c.JSON(accountErrorStatus(err), gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
	}
}

func AdminListUsersHandler(service *AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		users, err := service.ListUsers()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list users"})
			return
		}

		out := make([]UserResponse, 0, len(users))
		for _, u := range users {
			out = append(out, toUserResponse(u))
		}
		c.JSON(http.StatusOK, gin.H{"users": out})
	}
}