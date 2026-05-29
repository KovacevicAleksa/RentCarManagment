package auth

import (
	"testing"
	"time"
)

func TestTokenGenerateAndParseRoundtrip(t *testing.T) {
	ts := NewTokenService("test-secret", time.Hour)

	token, err := ts.Generate("user-123", "admin@rentcar.com", RoleAdmin)
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if token == "" {
		t.Fatal("Generate returned empty token")
	}

	claims, err := ts.Parse(token)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if claims.UserID != "user-123" {
		t.Errorf("UserID = %q, want %q", claims.UserID, "user-123")
	}
	if claims.Email != "admin@rentcar.com" {
		t.Errorf("Email = %q, want %q", claims.Email, "admin@rentcar.com")
	}
	if claims.Role != RoleAdmin {
		t.Errorf("Role = %q, want %q", claims.Role, RoleAdmin)
	}
}

func TestTokenParseRejectsTamperedToken(t *testing.T) {
	ts := NewTokenService("test-secret", time.Hour)
	token, _ := ts.Generate("user-123", "user@rentcar.com", RoleUser)

	if _, err := ts.Parse(token + "tampered"); err == nil {
		t.Fatal("Parse accepted a tampered token")
	}
}

func TestTokenParseRejectsWrongSecret(t *testing.T) {
	signer := NewTokenService("secret-a", time.Hour)
	verifier := NewTokenService("secret-b", time.Hour)
	token, _ := signer.Generate("user-123", "user@rentcar.com", RoleUser)

	if _, err := verifier.Parse(token); err == nil {
		t.Fatal("Parse accepted a token signed with a different secret")
	}
}

func TestTokenParseRejectsExpiredToken(t *testing.T) {
	ts := NewTokenService("test-secret", -time.Hour)
	token, _ := ts.Generate("user-123", "user@rentcar.com", RoleUser)

	if _, err := ts.Parse(token); err == nil {
		t.Fatal("Parse accepted an expired token")
	}
}
