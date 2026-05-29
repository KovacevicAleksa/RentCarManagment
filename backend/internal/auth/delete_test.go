package auth

import "testing"

func TestDeleteUserSoftDeletesRegularUser(t *testing.T) {
	repo := &stubRepo{}
	svc := newTestService(repo)
	admin := seedUser(t, svc, repo, "boss@rentcar.com", "password123", RoleAdmin)
	target := seedUser(t, svc, repo, "driver@rentcar.com", "password123", RoleUser)

	if err := svc.DeleteUser(admin, target); err != nil {
		t.Fatalf("DeleteUser returned error: %v", err)
	}
	if u, _ := repo.FindByID(target); u != nil {
		t.Error("deleted user is still present")
	}
}

func TestDeleteUserCannotDeleteSelf(t *testing.T) {
	repo := &stubRepo{}
	svc := newTestService(repo)
	admin := seedUser(t, svc, repo, "boss@rentcar.com", "password123", RoleAdmin)
	// a second admin so the "last admin" guard isn't the one that triggers
	_ = seedUser(t, svc, repo, "boss2@rentcar.com", "password123", RoleAdmin)

	if err := svc.DeleteUser(admin, admin); err == nil {
		t.Fatal("expected error when deleting self, got nil")
	}
	if u, _ := repo.FindByID(admin); u == nil {
		t.Error("self account was deleted despite the guard")
	}
}

func TestDeleteUserCannotDeleteLastAdmin(t *testing.T) {
	repo := &stubRepo{}
	svc := newTestService(repo)
	// only one admin in the system; acting id is someone else
	onlyAdmin := seedUser(t, svc, repo, "boss@rentcar.com", "password123", RoleAdmin)

	if err := svc.DeleteUser("other-acting-id", onlyAdmin); err == nil {
		t.Fatal("expected error when deleting the last admin, got nil")
	}
	if u, _ := repo.FindByID(onlyAdmin); u == nil {
		t.Error("last admin was deleted")
	}
}

func TestDeleteUserCanDeleteAdminWhenOthersExist(t *testing.T) {
	repo := &stubRepo{}
	svc := newTestService(repo)
	acting := seedUser(t, svc, repo, "boss@rentcar.com", "password123", RoleAdmin)
	target := seedUser(t, svc, repo, "boss2@rentcar.com", "password123", RoleAdmin)

	if err := svc.DeleteUser(acting, target); err != nil {
		t.Fatalf("should be able to delete an admin when others remain: %v", err)
	}
	if u, _ := repo.FindByID(target); u != nil {
		t.Error("target admin not deleted")
	}
}

func TestDeleteUserUnknownTarget(t *testing.T) {
	repo := &stubRepo{}
	svc := newTestService(repo)
	admin := seedUser(t, svc, repo, "boss@rentcar.com", "password123", RoleAdmin)

	if err := svc.DeleteUser(admin, "ghost"); err == nil {
		t.Fatal("expected error for unknown target, got nil")
	}
}
