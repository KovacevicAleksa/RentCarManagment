package history

import (
	"time"

	"gorm.io/gorm"
)

// HistoryRepository abstracts telemetry persistence so the service can be
// unit-tested against an in-memory stub instead of a real database.
type HistoryRepository interface {
	Create(history *CarHistory) error
	FindByCarID(carID string, limit int) ([]CarHistory, error)
	FindByCarIDAndTimeRange(carID string, start, end time.Time) ([]CarHistory, error)
	GetLatestByCarID(carID string) (*CarHistory, error)
	DeleteOlderThan(days int) error
}

type gormHistoryRepository struct {
	db *gorm.DB
}

func NewHistoryRepository(db *gorm.DB) HistoryRepository {
	return &gormHistoryRepository{db: db}
}

func (r *gormHistoryRepository) Create(history *CarHistory) error {
	return r.db.Create(history).Error
}

func (r *gormHistoryRepository) FindByCarID(carID string, limit int) ([]CarHistory, error) {
	var records []CarHistory
	err := r.db.Where("car_id = ?", carID).
		Order("timestamp DESC").
		Limit(limit).
		Find(&records).Error
	return records, err
}

func (r *gormHistoryRepository) FindByCarIDAndTimeRange(carID string, start, end time.Time) ([]CarHistory, error) {
	var records []CarHistory
	err := r.db.Where("car_id = ? AND timestamp BETWEEN ? AND ?", carID, start, end).
		Order("timestamp ASC").
		Find(&records).Error
	return records, err
}

func (r *gormHistoryRepository) GetLatestByCarID(carID string) (*CarHistory, error) {
	var record CarHistory
	err := r.db.Where("car_id = ?", carID).
		Order("timestamp DESC").
		First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *gormHistoryRepository) DeleteOlderThan(days int) error {
	cutoff := time.Now().AddDate(0, 0, -days)
	return r.db.Where("timestamp < ?", cutoff).Delete(&CarHistory{}).Error
}