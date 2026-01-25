package history

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CarHistory struct {
	Timestamp time.Time `gorm:"primaryKey;index:idx_car_time,priority:1" json:"timestamp"`
	CarID     string    `gorm:"type:varchar(50);primaryKey;index:idx_car_time,priority:2" json:"car_id"`
	
	ID        string    `gorm:"type:uuid;not null" json:"id"`
	Fuel      float64   `gorm:"type:decimal(5,2)" json:"fuel"`
	Latitude  float64   `gorm:"type:decimal(10,8)" json:"latitude"`
	Longitude float64   `gorm:"type:decimal(11,8)" json:"longitude"`
	
	gorm.Model `gorm:"embedded;embeddedPrefix:meta_"`
}

func (h *CarHistory) BeforeCreate(tx *gorm.DB) error {
	if h.ID == "" {
		h.ID = uuid.Must(uuid.NewV7()).String()
	}
	if h.Timestamp.IsZero() {
		h.Timestamp = time.Now()
	}
	return nil
}