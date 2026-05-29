package auth

import (
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestGeneratePasswordLengthAndCharset(t *testing.T) {
	pw, err := generatePassword(12)
	if err != nil {
		t.Fatalf("generatePassword error: %v", err)
	}
	if len(pw) != 12 {
		t.Errorf("length = %d, want 12", len(pw))
	}
	const ambiguous = "0O1lI"
	for _, c := range pw {
		if !strings.ContainsRune(passwordAlphabet, c) {
			t.Errorf("password contains disallowed char %q", c)
		}
		if strings.ContainsRune(ambiguous, c) {
			t.Errorf("password contains ambiguous char %q", c)
		}
	}
}

func TestGeneratePasswordIsRandom(t *testing.T) {
	a, _ := generatePassword(12)
	b, _ := generatePassword(12)
	if a == b {
		t.Error("two generated passwords are identical (not random)")
	}
}

func TestResetPasswordReturnsWorkingPassword(t *testing.T) {
	repo := &stubRepo{}
	svc := newTestService(repo)
	id := seedUser(t, svc, repo, "driver@rentcar.com", "oldpassword", RoleUser)

	temp, err := svc.ResetPassword(id)
	if err != nil {
		t.Fatalf("ResetPassword error: %v", err)
	}
	if len(temp) < minPasswordLength {
		t.Errorf("temporary password too short: %q", temp)
	}

	u, _ := repo.FindByID(id)
	if bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(temp)) != nil {
		t.Error("returned temporary password does not match stored hash")
	}
	if bcrypt.CompareHashAndPassword([]byte(u.Password), []byte("oldpassword")) == nil {
		t.Error("old password still works after reset")
	}
}

func TestResetPasswordUnknownUser(t *testing.T) {
	svc := newTestService(&stubRepo{})
	if _, err := svc.ResetPassword("ghost"); err == nil {
		t.Fatal("expected error for unknown user, got nil")
	}
}

func TestResetPasswordIsDifferentEachTime(t *testing.T) {
	repo := &stubRepo{}
	svc := newTestService(repo)
	id := seedUser(t, svc, repo, "driver@rentcar.com", "oldpassword", RoleUser)

	first, _ := svc.ResetPassword(id)
	second, _ := svc.ResetPassword(id)
	if first == second {
		t.Error("two consecutive resets produced the same password")
	}
}
