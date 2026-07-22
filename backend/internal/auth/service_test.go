package auth

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// stubRepo is an in-memory UserRepository for unit tests (no real DB).
type stubRepo struct {
	users     []*User
	createErr error
	updateErr error
	seq       int
}

func (s *stubRepo) CreateUser(u *User) error {
	if s.createErr != nil {
		return s.createErr
	}
	if u.ID == "" {
		s.seq++
		u.ID = fmt.Sprintf("stub-id-%d", s.seq)
	}
	s.users = append(s.users, u)
	return nil
}

func (s *stubRepo) FindByEmail(email string) (*User, error) {
	for _, u := range s.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, nil
}

func (s *stubRepo) FindByID(id string) (*User, error) {
	for _, u := range s.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, nil
}

func (s *stubRepo) UpdateUser(u *User) error {
	return s.updateErr
}

func (s *stubRepo) DeleteUser(id string) error {
	for i, u := range s.users {
		if u.ID == id {
			s.users = append(s.users[:i], s.users[i+1:]...)
			return nil
		}
	}
	return nil
}

func (s *stubRepo) CountByRole(role string) (int64, error) {
	var n int64
	for _, u := range s.users {
		if u.Role == role {
			n++
		}
	}
	return n, nil
}

func (s *stubRepo) FindAll() ([]User, error) {
	out := make([]User, 0, len(s.users))
	for _, u := range s.users {
		out = append(out, *u)
	}
	return out, nil
}

func newTestService(repo UserRepository) *AuthService {
	return NewAuthService(repo, NewTokenService("test-secret", time.Hour))
}

func TestRegisterCreatesUserWithUserRole(t *testing.T) {
	repo := &stubRepo{}
	svc := newTestService(repo)

	if err := svc.Register("driver@rentcar.com", "password123"); err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	if len(repo.users) != 1 {
		t.Fatalf("expected 1 user persisted, got %d", len(repo.users))
	}
	created := repo.users[0]
	if created.Role != RoleUser {
		t.Errorf("Role = %q, want %q", created.Role, RoleUser)
	}
	if created.Password == "password123" {
		t.Error("password was stored in plaintext")
	}
	if bcrypt.CompareHashAndPassword([]byte(created.Password), []byte("password123")) != nil {
		t.Error("stored password is not a valid bcrypt hash of the input")
	}
}

func TestRegisterCreatesPendingUser(t *testing.T) {
	repo := &stubRepo{}
	svc := newTestService(repo)

	if err := svc.Register("driver@rentcar.com", "password123"); err != nil {
		t.Fatalf("Register returned error: %v", err)
	}
	if got := repo.users[0].Status; got != StatusPending {
		t.Errorf("self-registered Status = %q, want %q", got, StatusPending)
	}
}

func TestAdminCreateUserIsApproved(t *testing.T) {
	repo := &stubRepo{}
	svc := newTestService(repo)

	if err := svc.CreateUser("driver@rentcar.com", "password123", RoleUser); err != nil {
		t.Fatalf("CreateUser returned error: %v", err)
	}
	if got := repo.users[0].Status; got != StatusApproved {
		t.Errorf("admin-created Status = %q, want %q", got, StatusApproved)
	}
}

func TestRegisterRejectsDuplicateEmail(t *testing.T) {
	repo := &stubRepo{}
	svc := newTestService(repo)
	_ = svc.Register("dup@rentcar.com", "password123")

	err := svc.Register("dup@rentcar.com", "password123")
	if err == nil {
		t.Fatal("expected error on duplicate email, got nil")
	}
	if len(repo.users) != 1 {
		t.Errorf("expected duplicate not to be persisted, got %d users", len(repo.users))
	}
}

func TestCreateUserWithAdminRole(t *testing.T) {
	repo := &stubRepo{}
	svc := newTestService(repo)

	if err := svc.CreateUser("boss@rentcar.com", "password123", RoleAdmin); err != nil {
		t.Fatalf("CreateUser returned error: %v", err)
	}
	if repo.users[0].Role != RoleAdmin {
		t.Errorf("Role = %q, want %q", repo.users[0].Role, RoleAdmin)
	}
}

func TestCreateUserRejectsInvalidRole(t *testing.T) {
	repo := &stubRepo{}
	svc := newTestService(repo)

	err := svc.CreateUser("x@rentcar.com", "password123", "superuser")
	if err == nil {
		t.Fatal("expected error on invalid role, got nil")
	}
	if len(repo.users) != 0 {
		t.Error("invalid-role user should not be persisted")
	}
}

func TestCreateUserRejectsDuplicateEmail(t *testing.T) {
	repo := &stubRepo{}
	svc := newTestService(repo)
	_ = svc.CreateUser("dup@rentcar.com", "password123", RoleUser)

	err := svc.CreateUser("dup@rentcar.com", "password123", RoleAdmin)
	if err == nil {
		t.Fatal("expected error on duplicate email, got nil")
	}
}

func TestLoginRejectsPendingAccount(t *testing.T) {
	repo := &stubRepo{}
	svc := newTestService(repo)
	_ = svc.Register("driver@rentcar.com", "password123") // pending

	_, _, err := svc.Login("driver@rentcar.com", "password123")
	if !errors.Is(err, ErrAccountPending) {
		t.Fatalf("expected ErrAccountPending, got %v", err)
	}
}

func TestApproveUserAllowsLogin(t *testing.T) {
	repo := &stubRepo{}
	svc := newTestService(repo)
	_ = svc.Register("driver@rentcar.com", "password123")
	u, _ := repo.FindByEmail("driver@rentcar.com")

	if err := svc.ApproveUser(u.ID); err != nil {
		t.Fatalf("ApproveUser returned error: %v", err)
	}
	if repo.users[0].Status != StatusApproved {
		t.Errorf("Status after approve = %q, want %q", repo.users[0].Status, StatusApproved)
	}
	if _, _, err := svc.Login("driver@rentcar.com", "password123"); err != nil {
		t.Errorf("login after approval should succeed, got %v", err)
	}
}

func TestApproveUnknownUser(t *testing.T) {
	svc := newTestService(&stubRepo{})
	if err := svc.ApproveUser("stub-id-404"); !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestLoginRejectsWrongPassword(t *testing.T) {
	repo := &stubRepo{}
	svc := newTestService(repo)
	_ = svc.Register("driver@rentcar.com", "correct-password")

	_, _, err := svc.Login("driver@rentcar.com", "wrong-password")
	if err == nil {
		t.Fatal("expected error on wrong password, got nil")
	}
}

func TestLoginUnknownEmail(t *testing.T) {
	svc := newTestService(&stubRepo{})

	_, _, err := svc.Login("ghost@rentcar.com", "whatever")
	if err == nil {
		t.Fatal("expected error on unknown email, got nil")
	}
}

func TestLoginSucceedsAndTokenCarriesRole(t *testing.T) {
	repo := &stubRepo{}
	tokens := NewTokenService("test-secret", time.Hour)
	svc := NewAuthService(repo, tokens)
	_ = svc.CreateUser("boss@rentcar.com", "password123", RoleAdmin)

	token, userID, err := svc.Login("boss@rentcar.com", "password123")
	if err != nil {
		t.Fatalf("Login returned error: %v", err)
	}
	if userID == "" {
		t.Error("Login returned empty userID")
	}

	claims, err := tokens.Parse(token)
	if err != nil {
		t.Fatalf("returned token did not parse: %v", err)
	}
	if claims.Role != RoleAdmin {
		t.Errorf("token Role = %q, want %q", claims.Role, RoleAdmin)
	}
	if claims.Email != "boss@rentcar.com" {
		t.Errorf("token Email = %q, want %q", claims.Email, "boss@rentcar.com")
	}
}

func TestCreateUserPropagatesRepoError(t *testing.T) {
	repo := &stubRepo{createErr: errors.New("db down")}
	svc := newTestService(repo)

	if err := svc.CreateUser("x@rentcar.com", "password123", RoleUser); err == nil {
		t.Fatal("expected repo error to propagate, got nil")
	}
}

func TestListUsers(t *testing.T) {
	repo := &stubRepo{}
	svc := newTestService(repo)
	_ = svc.Register("a@rentcar.com", "password123")
	_ = svc.CreateUser("b@rentcar.com", "password123", RoleAdmin)

	users, err := svc.ListUsers()
	if err != nil {
		t.Fatalf("ListUsers returned error: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}
}
