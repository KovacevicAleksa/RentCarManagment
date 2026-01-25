package history

import (
	"errors"
	"time"
)

type HistoryService struct {
	repo *HistoryRepository
}

func NewHistoryService(repo *HistoryRepository) *HistoryService {
	return &HistoryService{repo: repo}
}

func (s *HistoryService) SaveTelemetry(carID string, fuel, lat, lon float64) error {
	if carID == "" {
		return errors.New("car_id is required")
	}

	history := &CarHistory{
		CarID:     carID,
		Fuel:      fuel,
		Latitude:  lat,
		Longitude: lon,
	}

	return s.repo.Create(history)
}

func (s *HistoryService) GetRecentHistory(carID string, limit int) ([]CarHistory, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}
	return s.repo.FindByCarID(carID, limit)
}

func (s *HistoryService) GetHistoryInRange(carID string, start, end time.Time) ([]CarHistory, error) {
	if start.After(end) {
		return nil, errors.New("start time must be before end time")
	}
	return s.repo.FindByCarIDAndTimeRange(carID, start, end)
}

func (s *HistoryService) GetLatest(carID string) (*CarHistory, error) {
	return s.repo.GetLatestByCarID(carID)
}

func (s *HistoryService) CleanupOldRecords(days int) error {
	return s.repo.DeleteOlderThan(days)
}
