package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestAuthMiddlewareRejectsMissingToken(t *testing.T) {
	tokens := NewTokenService("test-secret", time.Hour)
	r := gin.New()
	r.GET("/protected", AuthMiddleware(tokens), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddlewareSetsClaimsFromValidToken(t *testing.T) {
	tokens := NewTokenService("test-secret", time.Hour)
	token, _ := tokens.Generate("u-1", "boss@rentcar.com", RoleAdmin)

	var gotRole, gotID string
	r := gin.New()
	r.GET("/protected", AuthMiddleware(tokens), func(c *gin.Context) {
		gotRole = getString(c, "role")
		gotID = getString(c, "user_id")
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token})
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if gotRole != RoleAdmin {
		t.Errorf("role = %q, want %q", gotRole, RoleAdmin)
	}
	if gotID != "u-1" {
		t.Errorf("user_id = %q, want %q", gotID, "u-1")
	}
}

func TestRequireRoleAllowsMatchingRole(t *testing.T) {
	r := gin.New()
	r.GET("/admin", injectRole(RoleAdmin), RequireRole(RoleAdmin), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestRequireRoleForbidsWrongRole(t *testing.T) {
	r := gin.New()
	r.GET("/admin", injectRole(RoleUser), RequireRole(RoleAdmin), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestRequireRoleForbidsMissingRole(t *testing.T) {
	r := gin.New()
	r.GET("/admin", RequireRole(RoleAdmin), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func injectRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("role", role)
		c.Next()
	}
}

// c.Get returns (any, bool); the tests above read it as a string via a helper.
func getString(c *gin.Context, key string) string {
	v, ok := c.Get(key)
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}
