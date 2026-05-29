package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid or expired token")

type Claims struct {
	UserID string
	Email  string
	Role   string
}

type TokenService struct {
	secret []byte
	ttl    time.Duration
}

func NewTokenService(secret string, ttl time.Duration) *TokenService {
	return &TokenService{secret: []byte(secret), ttl: ttl}
}

func (t *TokenService) TTL() time.Duration {
	return t.ttl
}

func (t *TokenService) Generate(userID, email, role string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"role":    role,
		"exp":     time.Now().Add(t.ttl).Unix(),
	})
	return token.SignedString(t.secret)
}

func (t *TokenService) Parse(tokenString string) (*Claims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return t.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	claims := &Claims{}
	if v, ok := mapClaims["user_id"].(string); ok {
		claims.UserID = v
	}
	if v, ok := mapClaims["email"].(string); ok {
		claims.Email = v
	}
	if v, ok := mapClaims["role"].(string); ok {
		claims.Role = v
	}
	return claims, nil
}
