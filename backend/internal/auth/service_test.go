package auth

import (
	"errors"
	"os"
	"testing"

	"github.com/google/uuid"
)

// mockAuthRepo is an in-memory AuthRepo used in tests.
type mockAuthRepo struct {
	users map[string]*User
}

func newMockAuthRepo() *mockAuthRepo {
	return &mockAuthRepo{users: make(map[string]*User)}
}

func (m *mockAuthRepo) FindByEmail(email string) (*User, error) {
	u, ok := m.users[email]
	if !ok {
		return nil, errors.New("user not found")
	}
	return u, nil
}

func (m *mockAuthRepo) CreateUser(user *User) error {
	if user.ID == "" {
		user.ID = uuid.Must(uuid.NewV7()).String()
	}
	m.users[user.Email] = user
	return nil
}

func TestRegister_Success(t *testing.T) {
	svc := NewAuthService(newMockAuthRepo())

	if err := svc.Register("user@example.com", "secret123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	svc := NewAuthService(newMockAuthRepo())

	_ = svc.Register("user@example.com", "secret123")
	err := svc.Register("user@example.com", "secret123")
	if err == nil {
		t.Fatal("expected error for duplicate email, got nil")
	}
}

func TestLogin_Success(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	svc := NewAuthService(newMockAuthRepo())

	_ = svc.Register("user@example.com", "secret123")

	token, userID, err := svc.Login("user@example.com", "secret123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
	if userID == "" {
		t.Fatal("expected non-empty user ID")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	svc := NewAuthService(newMockAuthRepo())

	_ = svc.Register("user@example.com", "secret123")

	_, _, err := svc.Login("user@example.com", "wrongpassword")
	if err == nil {
		t.Fatal("expected error for wrong password, got nil")
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	svc := NewAuthService(newMockAuthRepo())

	_, _, err := svc.Login("nobody@example.com", "password")
	if err == nil {
		t.Fatal("expected error for missing user, got nil")
	}
}
