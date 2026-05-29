package auth

import (
	"net/http"
	"testing"
)

func TestAdminDeleteUserSucceeds(t *testing.T) {
	r, svc, tokens := newTestRouter(t)
	_ = svc.CreateUser("driver@rentcar.com", "password123", RoleUser)
	target, _ := svc.repo.FindByEmail("driver@rentcar.com")
	admin := cookieFor(tokens, "admin-1", "boss@rentcar.com", RoleAdmin)

	w := doJSON(r, http.MethodDelete, "/admin/users/"+target.ID, nil, admin)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body: %s)", w.Code, w.Body.String())
	}
	if u, _ := svc.repo.FindByID(target.ID); u != nil {
		t.Error("user still present after delete")
	}
}

func TestAdminDeleteUserRequiresAdmin(t *testing.T) {
	r, svc, tokens := newTestRouter(t)
	_ = svc.CreateUser("driver@rentcar.com", "password123", RoleUser)
	target, _ := svc.repo.FindByEmail("driver@rentcar.com")
	regular := cookieFor(tokens, "u-1", "driver@rentcar.com", RoleUser)

	w := doJSON(r, http.MethodDelete, "/admin/users/"+target.ID, nil, regular)
	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", w.Code)
	}
}

func TestAdminDeleteUserRequiresAuth(t *testing.T) {
	r, _, _ := newTestRouter(t)
	w := doJSON(r, http.MethodDelete, "/admin/users/some-id", nil, nil)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestAdminDeleteUserCannotDeleteSelf(t *testing.T) {
	r, svc, tokens := newTestRouter(t)
	_ = svc.CreateUser("boss@rentcar.com", "password123", RoleAdmin)
	_ = svc.CreateUser("boss2@rentcar.com", "password123", RoleAdmin)
	me, _ := svc.repo.FindByEmail("boss@rentcar.com")
	admin := cookieFor(tokens, me.ID, me.Email, RoleAdmin)

	w := doJSON(r, http.MethodDelete, "/admin/users/"+me.ID, nil, admin)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400 (body: %s)", w.Code, w.Body.String())
	}
}

func TestAdminDeleteUserCannotDeleteLastAdmin(t *testing.T) {
	r, svc, tokens := newTestRouter(t)
	_ = svc.CreateUser("boss@rentcar.com", "password123", RoleAdmin)
	onlyAdmin, _ := svc.repo.FindByEmail("boss@rentcar.com")
	acting := cookieFor(tokens, "different-admin-id", "other@rentcar.com", RoleAdmin)

	w := doJSON(r, http.MethodDelete, "/admin/users/"+onlyAdmin.ID, nil, acting)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400 (body: %s)", w.Code, w.Body.String())
	}
}

func TestAdminDeleteUserUnknown(t *testing.T) {
	r, _, tokens := newTestRouter(t)
	admin := cookieFor(tokens, "admin-1", "boss@rentcar.com", RoleAdmin)

	w := doJSON(r, http.MethodDelete, "/admin/users/ghost", nil, admin)
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404 (body: %s)", w.Code, w.Body.String())
	}
}
