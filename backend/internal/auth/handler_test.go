package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func newTestRouter(t *testing.T) (*gin.Engine, *AuthService, *TokenService) {
	t.Helper()
	repo := &stubRepo{}
	tokens := NewTokenService("test-secret", time.Hour)
	svc := NewAuthService(repo, tokens)
	r := gin.New()
	RegisterRoutes(r, svc, tokens)
	return r, svc, tokens
}

func cookieFor(tokens *TokenService, id, email, role string) *http.Cookie {
	token, _ := tokens.Generate(id, email, role)
	return &http.Cookie{Name: "token", Value: token}
}

func doJSON(r *gin.Engine, method, path string, body any, cookie *http.Cookie) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if cookie != nil {
		req.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestAdminCreateUserSucceeds(t *testing.T) {
	r, svc, tokens := newTestRouter(t)
	admin := cookieFor(tokens, "a-1", "boss@rentcar.com", RoleAdmin)

	w := doJSON(r, http.MethodPost, "/admin/users", gin.H{
		"email": "new@rentcar.com", "password": "password123", "role": RoleUser,
	}, admin)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d (body: %s)", w.Code, http.StatusCreated, w.Body.String())
	}
	users, _ := svc.ListUsers()
	if len(users) != 1 || users[0].Email != "new@rentcar.com" {
		t.Errorf("user was not created via admin endpoint: %+v", users)
	}
}

func TestAdminCreateUserRequiresAdminRole(t *testing.T) {
	r, _, tokens := newTestRouter(t)
	regular := cookieFor(tokens, "u-1", "driver@rentcar.com", RoleUser)

	w := doJSON(r, http.MethodPost, "/admin/users", gin.H{
		"email": "new@rentcar.com", "password": "password123", "role": RoleUser,
	}, regular)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestAdminCreateUserRejectsInvalidRole(t *testing.T) {
	r, _, tokens := newTestRouter(t)
	admin := cookieFor(tokens, "a-1", "boss@rentcar.com", RoleAdmin)

	w := doJSON(r, http.MethodPost, "/admin/users", gin.H{
		"email": "new@rentcar.com", "password": "password123", "role": "root",
	}, admin)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d (body: %s)", w.Code, http.StatusBadRequest, w.Body.String())
	}
}

func TestAdminCreateUserRejectsAnonymous(t *testing.T) {
	r, _, _ := newTestRouter(t)

	w := doJSON(r, http.MethodPost, "/admin/users", gin.H{
		"email": "new@rentcar.com", "password": "password123", "role": RoleUser,
	}, nil)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAdminListUsersHidesPasswords(t *testing.T) {
	r, svc, tokens := newTestRouter(t)
	_ = svc.CreateUser("a@rentcar.com", "password123", RoleUser)
	admin := cookieFor(tokens, "a-1", "boss@rentcar.com", RoleAdmin)

	w := doJSON(r, http.MethodGet, "/admin/users", nil, admin)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp struct {
		Users []struct {
			Email    string `json:"email"`
			Role     string `json:"role"`
			Password string `json:"password"`
		} `json:"users"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if len(resp.Users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(resp.Users))
	}
	if resp.Users[0].Password != "" {
		t.Error("password hash leaked in admin user listing")
	}
}

func TestMeReturnsRole(t *testing.T) {
	r, _, tokens := newTestRouter(t)
	admin := cookieFor(tokens, "a-1", "boss@rentcar.com", RoleAdmin)

	w := doJSON(r, http.MethodGet, "/auth/me", nil, admin)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["role"] != RoleAdmin {
		t.Errorf("role = %v, want %q", resp["role"], RoleAdmin)
	}
}
