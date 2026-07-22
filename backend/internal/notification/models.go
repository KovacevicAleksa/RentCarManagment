package notification

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	TypeInfo  = "info"
	TypeAlert = "alert"

	SourceAdmin  = "admin"
	SourceSystem = "system"
)

// Notification is a message shown to all registered users. Read state is global
// (shared across users): marking notifications read affects everyone.
type Notification struct {
	ID        string    `gorm:"type:uuid;primaryKey" json:"id"`
	Title     string    `gorm:"type:varchar(200);not null" json:"title"`
	Message   string    `gorm:"type:text;not null" json:"message"`
	Type      string    `gorm:"type:varchar(20);not null;default:info" json:"type"`
	Source    string    `gorm:"type:varchar(20);not null;default:admin" json:"source"`
	CarID     *string   `gorm:"type:varchar(50)" json:"car_id,omitempty"`
	Read      bool      `gorm:"not null;default:false;index" json:"read"`
	CreatedAt time.Time `json:"created_at"`
}

func (n *Notification) BeforeCreate(tx *gorm.DB) error {
	if n.ID == "" {
		n.ID = uuid.Must(uuid.NewV7()).String()
	}
	return nil
}
