package auth

import (
	"crypto/rand"
	"math/big"

	"golang.org/x/crypto/bcrypt"
)

// passwordAlphabet excludes visually ambiguous characters (0, O, 1, l, I).
const passwordAlphabet = "abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func generatePassword(n int) (string, error) {
	buf := make([]byte, n)
	max := big.NewInt(int64(len(passwordAlphabet)))
	for i := range buf {
		idx, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		buf[i] = passwordAlphabet[idx.Int64()]
	}
	return string(buf), nil
}

// ResetPassword generates a new temporary password for a user, stores its hash,
// and returns the plaintext so an admin can relay it once.
func (s *AuthService) ResetPassword(userID string) (string, error) {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", ErrUserNotFound
	}

	temp, err := generatePassword(12)
	if err != nil {
		return "", err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(temp), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	user.Password = string(hash)
	if err := s.repo.UpdateUser(user); err != nil {
		return "", err
	}
	return temp, nil
}
