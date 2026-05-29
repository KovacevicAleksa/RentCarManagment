package auth

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

func extractTokenCookie(w interface{ Result() *http.Response }) *http.Cookie {
	for _, c := range w.Result().Cookies() {
		if c.Name == "token" {
			return c
		}
	}
	return nil
}

func TestChangePasswordRequiresAuth(t *testing.T) {
	r, _, _ := newTestRouter(t)
	w := doJSON(r, http.MethodPatch, "/auth/password", gin.H{
		"current_password": "x", "new_password": "newpassword",
	}, nil)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestChangePasswordEndToEnd(t *testing.T) {
	r, svc, tokens := newTestRouter(t)
	_ = svc.CreateUser("driver@rentcar.com", "oldpassword", RoleUser)
	u, _ := svc.repo.FindByEmail("driver@rentcar.com")
	cookie := cookieFor(tokens, u.ID, u.Email, u.Role)

	w := doJSON(r, http.MethodPatch, "/auth/password", gin.H{
		"current_password": "oldpassword", "new_password": "newpassword",
	}, cookie)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body: %s)", w.Code, w.Body.String())
	}

	// old password rejected, new password accepted
	if _, _, err := svc.Login("driver@rentcar.com", "oldpassword"); err == nil {
		t.Error("login with old password should fail after change")
	}
	if _, _, err := svc.Login("driver@rentcar.com", "newpassword"); err != nil {
		t.Errorf("login with new password should succeed, got: %v", err)
	}
}

func TestChangePasswordHandlerRejectsWrongCurrent(t *testing.T) {
	r, svc, tokens := newTestRouter(t)
	_ = svc.CreateUser("driver@rentcar.com", "oldpassword", RoleUser)
	u, _ := svc.repo.FindByEmail("driver@rentcar.com")
	cookie := cookieFor(tokens, u.ID, u.Email, u.Role)

	w := doJSON(r, http.MethodPatch, "/auth/password", gin.H{
		"current_password": "WRONG", "new_password": "newpassword",
	}, cookie)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestChangePasswordHandlerRejectsShortNew(t *testing.T) {
	r, svc, tokens := newTestRouter(t)
	_ = svc.CreateUser("driver@rentcar.com", "oldpassword", RoleUser)
	u, _ := svc.repo.FindByEmail("driver@rentcar.com")
	cookie := cookieFor(tokens, u.ID, u.Email, u.Role)

	w := doJSON(r, http.MethodPatch, "/auth/password", gin.H{
		"current_password": "oldpassword", "new_password": "123",
	}, cookie)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestChangeEmailReissuesTokenWithNewEmail(t *testing.T) {
	r, svc, tokens := newTestRouter(t)
	_ = svc.CreateUser("old@rentcar.com", "password123", RoleUser)
	u, _ := svc.repo.FindByEmail("old@rentcar.com")
	cookie := cookieFor(tokens, u.ID, u.Email, u.Role)

	w := doJSON(r, http.MethodPatch, "/auth/email", gin.H{
		"new_email": "new@rentcar.com", "current_password": "password123",
	}, cookie)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body: %s)", w.Code, w.Body.String())
	}

	newCookie := extractTokenCookie(w)
	if newCookie == nil {
		t.Fatal("expected a reissued token cookie")
	}

	// the reissued token must carry the new email
	meW := doJSON(r, http.MethodGet, "/auth/me", nil, newCookie)
	var me map[string]any
	_ = json.Unmarshal(meW.Body.Bytes(), &me)
	if me["email"] != "new@rentcar.com" {
		t.Errorf("/auth/me email = %v, want new@rentcar.com", me["email"])
	}
}

func TestChangeEmailHandlerRejectsDuplicate(t *testing.T) {
	r, svc, tokens := newTestRouter(t)
	_ = svc.CreateUser("taken@rentcar.com", "password123", RoleUser)
	_ = svc.CreateUser("me@rentcar.com", "password123", RoleUser)
	u, _ := svc.repo.FindByEmail("me@rentcar.com")
	cookie := cookieFor(tokens, u.ID, u.Email, u.Role)

	w := doJSON(r, http.MethodPatch, "/auth/email", gin.H{
		"new_email": "taken@rentcar.com", "current_password": "password123",
	}, cookie)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400 (body: %s)", w.Code, w.Body.String())
	}
}

func TestChangeEmailHandlerRejectsWrongPassword(t *testing.T) {
	r, svc, tokens := newTestRouter(t)
	_ = svc.CreateUser("me@rentcar.com", "password123", RoleUser)
	u, _ := svc.repo.FindByEmail("me@rentcar.com")
	cookie := cookieFor(tokens, u.ID, u.Email, u.Role)

	w := doJSON(r, http.MethodPatch, "/auth/email", gin.H{
		"new_email": "new@rentcar.com", "current_password": "WRONG",
	}, cookie)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}
