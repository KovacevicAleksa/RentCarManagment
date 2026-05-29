package auth

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func seedUser(t *testing.T, svc *AuthService, repo *stubRepo, email, password, role string) string {
	t.Helper()
	if err := svc.CreateUser(email, password, role); err != nil {
		t.Fatalf("seed CreateUser failed: %v", err)
	}
	u, _ := repo.FindByEmail(email)
	if u == nil {
		t.Fatal("seed user not found after create")
	}
	return u.ID
}

func TestChangePasswordSucceeds(t *testing.T) {
	repo := &stubRepo{}
	svc := newTestService(repo)
	id := seedUser(t, svc, repo, "driver@rentcar.com", "oldpassword", RoleUser)

	if err := svc.ChangePassword(id, "oldpassword", "newpassword"); err != nil {
		t.Fatalf("ChangePassword returned error: %v", err)
	}

	u, _ := repo.FindByID(id)
	if bcrypt.CompareHashAndPassword([]byte(u.Password), []byte("newpassword")) != nil {
		t.Error("password was not updated to the new value")
	}
	if bcrypt.CompareHashAndPassword([]byte(u.Password), []byte("oldpassword")) == nil {
		t.Error("old password still works after change")
	}
}

func TestChangePasswordRejectsWrongCurrent(t *testing.T) {
	repo := &stubRepo{}
	svc := newTestService(repo)
	id := seedUser(t, svc, repo, "driver@rentcar.com", "oldpassword", RoleUser)

	if err := svc.ChangePassword(id, "WRONG", "newpassword"); err == nil {
		t.Fatal("expected error for wrong current password, got nil")
	}
	u, _ := repo.FindByID(id)
	if bcrypt.CompareHashAndPassword([]byte(u.Password), []byte("oldpassword")) != nil {
		t.Error("password should be unchanged when current password is wrong")
	}
}

func TestChangePasswordRejectsShortNew(t *testing.T) {
	repo := &stubRepo{}
	svc := newTestService(repo)
	id := seedUser(t, svc, repo, "driver@rentcar.com", "oldpassword", RoleUser)

	if err := svc.ChangePassword(id, "oldpassword", "123"); err == nil {
		t.Fatal("expected error for too-short new password, got nil")
	}
}

func TestChangePasswordUnknownUser(t *testing.T) {
	svc := newTestService(&stubRepo{})
	if err := svc.ChangePassword("ghost", "x", "newpassword"); err == nil {
		t.Fatal("expected error for unknown user, got nil")
	}
}

func TestChangeEmailSucceeds(t *testing.T) {
	repo := &stubRepo{}
	svc := newTestService(repo)
	id := seedUser(t, svc, repo, "old@rentcar.com", "password123", RoleUser)

	updated, err := svc.ChangeEmail(id, "new@rentcar.com", "password123")
	if err != nil {
		t.Fatalf("ChangeEmail returned error: %v", err)
	}
	if updated.Email != "new@rentcar.com" {
		t.Errorf("returned user email = %q, want %q", updated.Email, "new@rentcar.com")
	}
	u, _ := repo.FindByID(id)
	if u.Email != "new@rentcar.com" {
		t.Errorf("persisted email = %q, want %q", u.Email, "new@rentcar.com")
	}
}

func TestChangeEmailRejectsWrongPassword(t *testing.T) {
	repo := &stubRepo{}
	svc := newTestService(repo)
	id := seedUser(t, svc, repo, "old@rentcar.com", "password123", RoleUser)

	if _, err := svc.ChangeEmail(id, "new@rentcar.com", "WRONG"); err == nil {
		t.Fatal("expected error for wrong password, got nil")
	}
}

func TestChangeEmailRejectsDuplicate(t *testing.T) {
	repo := &stubRepo{}
	svc := newTestService(repo)
	_ = seedUser(t, svc, repo, "taken@rentcar.com", "password123", RoleUser)
	id := seedUser(t, svc, repo, "me@rentcar.com", "password123", RoleUser)

	if _, err := svc.ChangeEmail(id, "taken@rentcar.com", "password123"); err == nil {
		t.Fatal("expected error when changing to an already-used email, got nil")
	}
}

func TestChangeEmailToSameEmailIsAllowed(t *testing.T) {
	repo := &stubRepo{}
	svc := newTestService(repo)
	id := seedUser(t, svc, repo, "me@rentcar.com", "password123", RoleUser)

	if _, err := svc.ChangeEmail(id, "me@rentcar.com", "password123"); err != nil {
		t.Fatalf("changing to the same email should be allowed, got: %v", err)
	}
}
