package auth

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRepository abstracts user persistence so the service can be unit-tested
// against an in-memory stub instead of a real database.
type UserRepository interface {
	CreateUser(user *User) error
	FindByEmail(email string) (*User, error)
	FindByID(id string) (*User, error)
	UpdateUser(user *User) error
	DeleteUser(id string) error
	CountByRole(role string) (int64, error)
	FindAll() ([]User, error)
}

type gormUserRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) UserRepository {
	return &gormUserRepository{db: db}
}

func (r *gormUserRepository) CreateUser(user *User) error {
	return r.db.Create(user).Error
}

// FindByEmail returns (nil, nil) when no user matches, so callers can
// distinguish "not found" from a real database error.
func (r *gormUserRepository) FindByEmail(email string) (*User, error) {
	var user User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *gormUserRepository) FindByID(id string) (*User, error) {
	// A malformed id can never match a UUID primary key; treat it as not found
	// instead of letting Postgres raise an "invalid input syntax" error.
	if _, err := uuid.Parse(id); err != nil {
		return nil, nil
	}
	var user User
	if err := r.db.Where("id = ?", id).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *gormUserRepository) UpdateUser(user *User) error {
	return r.db.Save(user).Error
}

func (r *gormUserRepository) DeleteUser(id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return nil
	}
	return r.db.Where("id = ?", id).Delete(&User{}).Error
}

func (r *gormUserRepository) CountByRole(role string) (int64, error) {
	var n int64
	err := r.db.Model(&User{}).Where("role = ?", role).Count(&n).Error
	return n, err
}

func (r *gormUserRepository) FindAll() ([]User, error) {
	var users []User
	if err := r.db.Order("meta_created_at asc").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
