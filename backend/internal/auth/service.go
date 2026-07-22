package auth

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidRole        = errors.New("invalid role")
	ErrUserNotFound       = errors.New("user not found")
	ErrWeakPassword       = errors.New("password must be at least 6 characters")
	ErrCannotDeleteSelf   = errors.New("cannot delete your own account")
	ErrLastAdmin          = errors.New("cannot delete the last admin")
	ErrAccountPending     = errors.New("account is pending admin approval")
)

const minPasswordLength = 6

type AuthService struct {
	repo   UserRepository
	tokens *TokenService
}

func NewAuthService(repo UserRepository, tokens *TokenService) *AuthService {
	return &AuthService{repo: repo, tokens: tokens}
}

// Register self-registers a regular user. The account starts pending and
// cannot log in until an admin approves it.
func (s *AuthService) Register(email, password string) error {
	return s.createUser(email, password, RoleUser, StatusPending)
}

// CreateUser creates a user with an explicit role. Used by admins creating
// accounts and by startup admin seeding; such accounts are approved on creation.
func (s *AuthService) CreateUser(email, password, role string) error {
	return s.createUser(email, password, role, StatusApproved)
}

func (s *AuthService) createUser(email, password, role, status string) error {
	if !IsValidRole(role) {
		return ErrInvalidRole
	}

	existing, err := s.repo.FindByEmail(email)
	if err != nil {
		return err
	}
	if existing != nil {
		return ErrUserExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.repo.CreateUser(&User{
		Email:    email,
		Password: string(hash),
		Role:     role,
		Status:   status,
	})
}

func (s *AuthService) Login(email, password string) (string, string, error) {
	user, err := s.repo.FindByEmail(email)
	if err != nil {
		return "", "", err
	}
	if user == nil {
		return "", "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", "", ErrInvalidCredentials
	}

	if user.Status != StatusApproved {
		return "", "", ErrAccountPending
	}

	token, err := s.tokens.Generate(user.ID, user.Email, user.Role)
	if err != nil {
		return "", "", err
	}

	return token, user.ID, nil
}

func (s *AuthService) ListUsers() ([]User, error) {
	return s.repo.FindAll()
}

// ApproveUser marks a pending account as approved so it can log in.
func (s *AuthService) ApproveUser(id string) error {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}
	user.Status = StatusApproved
	return s.repo.UpdateUser(user)
}

// ChangePassword updates a user's password after verifying the current one.
func (s *AuthService) ChangePassword(userID, currentPassword, newPassword string) error {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword)) != nil {
		return ErrInvalidCredentials
	}
	if len(newPassword) < minPasswordLength {
		return ErrWeakPassword
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hash)
	return s.repo.UpdateUser(user)
}

// ChangeEmail updates a user's email after verifying their password and that
// the new address is not already taken by someone else. It returns the updated
// user so the caller can reissue an auth token carrying the new email.
func (s *AuthService) ChangeEmail(userID, newEmail, currentPassword string) (*User, error) {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword)) != nil {
		return nil, ErrInvalidCredentials
	}

	if newEmail != user.Email {
		existing, err := s.repo.FindByEmail(newEmail)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return nil, ErrUserExists
		}
	}

	user.Email = newEmail
	if err := s.repo.UpdateUser(user); err != nil {
		return nil, err
	}
	return user, nil
}

// DeleteUser soft-deletes a user. An admin cannot delete their own account, and
// the last remaining admin cannot be deleted.
func (s *AuthService) DeleteUser(actingUserID, targetID string) error {
	if actingUserID == targetID {
		return ErrCannotDeleteSelf
	}

	target, err := s.repo.FindByID(targetID)
	if err != nil {
		return err
	}
	if target == nil {
		return ErrUserNotFound
	}

	if target.Role == RoleAdmin {
		admins, err := s.repo.CountByRole(RoleAdmin)
		if err != nil {
			return err
		}
		if admins <= 1 {
			return ErrLastAdmin
		}
	}

	return s.repo.DeleteUser(targetID)
}

// EnsureAdmin creates an admin account if one with the given email does not
// already exist. It is a no-op when credentials are empty, so startup seeding
// can be disabled by leaving the env vars unset.
func (s *AuthService) EnsureAdmin(email, password string) error {
	if email == "" || password == "" {
		return nil
	}

	existing, err := s.repo.FindByEmail(email)
	if err != nil {
		return err
	}
	if existing != nil {
		return nil
	}

	return s.CreateUser(email, password, RoleAdmin)
}
