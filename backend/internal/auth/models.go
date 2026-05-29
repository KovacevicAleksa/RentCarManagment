package auth

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID       string `gorm:"type:uuid;primaryKey"`
	Email    string `gorm:"uniqueIndex"`
	Password string
	Role     string `gorm:"default:user"`
	gorm.Model `gorm:"embedded;embeddedPrefix:meta_"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.Must(uuid.NewV7()).String()
	}
	return nil
}