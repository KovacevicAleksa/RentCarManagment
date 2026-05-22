package history

import (
	"time"

	"gorm.io/gorm"
)

type HistoryRepo interface {
	Create(h *CarHistory) error
	CreateBatch(items []*CarHistory) error
	FindByCarID(carID string, limit int) ([]CarHistory, error)
	FindByCarIDAndTimeRange(carID string, start, end time.Time) ([]CarHistory, error)
	GetLatestByCarID(carID string) (*CarHistory, error)
	DeleteOlderThan(days int) error
}

type HistoryRepository struct {
	db *gorm.DB
}

func NewHistoryRepository(db *gorm.DB) *HistoryRepository {
	return &HistoryRepository{db: db}
}

func (r *HistoryRepository) Create(history *CarHistory) error {
	return r.db.Create(history).Error
}

func (r *HistoryRepository) CreateBatch(items []*CarHistory) error {
	if len(items) == 0 {
		return nil
	}
	return r.db.CreateInBatches(items, 500).Error
}

func (r *HistoryRepository) FindByCarID(carID string, limit int) ([]CarHistory, error) {
	var records []CarHistory
	err := r.db.Where("car_id = ?", carID).
		Order("timestamp DESC").
		Limit(limit).
		Find(&records).Error
	return records, err
}

func (r *HistoryRepository) FindByCarIDAndTimeRange(carID string, start, end time.Time) ([]CarHistory, error) {
	var records []CarHistory
	err := r.db.Where("car_id = ? AND timestamp BETWEEN ? AND ?", carID, start, end).
		Order("timestamp ASC").
		Find(&records).Error
	return records, err
}

func (r *HistoryRepository) GetLatestByCarID(carID string) (*CarHistory, error) {
	var record CarHistory
	err := r.db.Where("car_id = ?", carID).
		Order("timestamp DESC").
		First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *HistoryRepository) DeleteOlderThan(days int) error {
	cutoff := time.Now().AddDate(0, 0, -days)
	return r.db.Where("timestamp < ?", cutoff).Delete(&CarHistory{}).Error
}