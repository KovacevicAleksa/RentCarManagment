package history

import (
	"errors"
	"log"
	"time"
)

const (
	queueSize     = 50000
	batchSize     = 500
	flushInterval = 500 * time.Millisecond
)

type HistoryService struct {
	repo  HistoryRepo
	queue chan *CarHistory
}

func NewHistoryService(repo HistoryRepo) *HistoryService {
	s := &HistoryService{
		repo:  repo,
		queue: make(chan *CarHistory, queueSize),
	}
	go s.batchWorker()
	return s
}

func (s *HistoryService) SaveTelemetry(carID string, fuel, lat, lon float64) error {
	if carID == "" {
		return errors.New("car_id is required")
	}

	h := &CarHistory{
		CarID:     carID,
		Fuel:      fuel,
		Latitude:  lat,
		Longitude: lon,
	}

	select {
	case s.queue <- h:
		return nil
	default:
		return errors.New("history queue full, telemetry dropped")
	}
}

func (s *HistoryService) batchWorker() {
	buf := make([]*CarHistory, 0, batchSize)
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	flush := func() {
		if len(buf) == 0 {
			return
		}
		if err := s.repo.CreateBatch(buf); err != nil {
			log.Printf("batch insert failed (%d rows): %v", len(buf), err)
		}
		buf = buf[:0]
	}

	for {
		select {
		case h := <-s.queue:
			buf = append(buf, h)
			if len(buf) >= batchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
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
