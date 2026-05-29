package auth

import "testing"

func TestEnsureAdminCreatesAdminWhenMissing(t *testing.T) {
	repo := &stubRepo{}
	svc := newTestService(repo)

	if err := svc.EnsureAdmin("boss@rentcar.com", "password123"); err != nil {
		t.Fatalf("EnsureAdmin returned error: %v", err)
	}
	if len(repo.users) != 1 || repo.users[0].Role != RoleAdmin {
		t.Fatalf("expected one admin user, got %+v", repo.users)
	}
}

func TestEnsureAdminIsIdempotent(t *testing.T) {
	repo := &stubRepo{}
	svc := newTestService(repo)
	_ = svc.EnsureAdmin("boss@rentcar.com", "password123")

	if err := svc.EnsureAdmin("boss@rentcar.com", "password123"); err != nil {
		t.Fatalf("second EnsureAdmin returned error: %v", err)
	}
	if len(repo.users) != 1 {
		t.Errorf("EnsureAdmin created a duplicate, got %d users", len(repo.users))
	}
}

func TestEnsureAdminSkipsWhenCredentialsEmpty(t *testing.T) {
	repo := &stubRepo{}
	svc := newTestService(repo)

	if err := svc.EnsureAdmin("", ""); err != nil {
		t.Fatalf("EnsureAdmin returned error: %v", err)
	}
	if len(repo.users) != 0 {
		t.Error("EnsureAdmin should not create a user when credentials are empty")
	}
}
