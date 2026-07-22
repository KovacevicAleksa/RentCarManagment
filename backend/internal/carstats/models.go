package carstats

import "time"

// Stats is the per-car engine-health aggregate. One row per car, upserted
// periodically. Counters accumulate over the car's whole life; OverheatFrequency
// is the derived value (overheat episodes per operating hour).
type Stats struct {
	CarID             string     `gorm:"type:varchar(50);primaryKey" json:"car_id"`
	TotalReadings     int64      `gorm:"not null;default:0" json:"total_readings"`
	OverheatEvents    int64      `gorm:"not null;default:0" json:"overheat_events"`
	MaxEngineTemp     float64    `gorm:"type:decimal(6,2);not null;default:0" json:"max_engine_temp"`
	LastOverheatAt    *time.Time `json:"last_overheat_at,omitempty"`
	OverheatFrequency float64    `gorm:"type:decimal(12,6);not null;default:0" json:"overheat_frequency"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// TableName pins the table to car_stats; GORM would otherwise derive the generic
// name "stats" from the struct.
func (Stats) TableName() string { return "car_stats" }
