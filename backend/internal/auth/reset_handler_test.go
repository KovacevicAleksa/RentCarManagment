package auth

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestAdminResetPasswordSucceeds(t *testing.T) {
	r, svc, tokens := newTestRouter(t)
	_ = svc.CreateUser("driver@rentcar.com", "oldpassword", RoleUser)
	target, _ := svc.repo.FindByEmail("driver@rentcar.com")
	admin := cookieFor(tokens, "a-1", "boss@rentcar.com", RoleAdmin)

	w := doJSON(r, http.MethodPost, "/admin/users/"+target.ID+"/reset-password", nil, admin)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body: %s)", w.Code, w.Body.String())
	}

	var resp struct {
		Email     string `json:"email"`
		TempPass  string `json:"temporary_password"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if resp.TempPass == "" {
		t.Fatal("response did not include a temporary_password")
	}

	// the user can authenticate with the returned temporary password
	if _, _, err := svc.Login("driver@rentcar.com", resp.TempPass); err != nil {
		t.Errorf("login with temporary password failed: %v", err)
	}
}

func TestAdminResetPasswordRequiresAdmin(t *testing.T) {
	r, svc, tokens := newTestRouter(t)
	_ = svc.CreateUser("driver@rentcar.com", "oldpassword", RoleUser)
	target, _ := svc.repo.FindByEmail("driver@rentcar.com")
	regular := cookieFor(tokens, "u-1", "driver@rentcar.com", RoleUser)

	w := doJSON(r, http.MethodPost, "/admin/users/"+target.ID+"/reset-password", nil, regular)
	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", w.Code)
	}
}

func TestAdminResetPasswordRequiresAuth(t *testing.T) {
	r, _, _ := newTestRouter(t)
	w := doJSON(r, http.MethodPost, "/admin/users/some-id/reset-password", nil, nil)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestAdminResetPasswordUnknownUser(t *testing.T) {
	r, _, tokens := newTestRouter(t)
	admin := cookieFor(tokens, "a-1", "boss@rentcar.com", RoleAdmin)

	w := doJSON(r, http.MethodPost, "/admin/users/ghost-id/reset-password", nil, admin)
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404 (body: %s)", w.Code, w.Body.String())
	}
}
