package history

import (
	"errors"
	"testing"
	"time"
)

type stubRepo struct {
	created      []*CarHistory
	lastLimit    int
	lastDays     int
	findErr      error
	createErr    error
	rangeQueried bool
}

func (s *stubRepo) Create(h *CarHistory) error {
	if s.createErr != nil {
		return s.createErr
	}
	s.created = append(s.created, h)
	return nil
}

func (s *stubRepo) FindByCarID(carID string, limit int) ([]CarHistory, error) {
	s.lastLimit = limit
	return nil, s.findErr
}

func (s *stubRepo) FindByCarIDAndTimeRange(carID string, start, end time.Time) ([]CarHistory, error) {
	s.rangeQueried = true
	return nil, s.findErr
}

func (s *stubRepo) GetLatestByCarID(carID string) (*CarHistory, error) {
	return nil, s.findErr
}

func (s *stubRepo) DeleteOlderThan(days int) error {
	s.lastDays = days
	return nil
}

func TestSaveTelemetryRejectsEmptyCarID(t *testing.T) {
	repo := &stubRepo{}
	svc := NewHistoryService(repo)

	if err := svc.SaveTelemetry("", 50, 44.0, 20.0); err == nil {
		t.Fatal("expected error for empty car_id, got nil")
	}
	if len(repo.created) != 0 {
		t.Error("telemetry with empty car_id should not be persisted")
	}
}

func TestSaveTelemetryPersistsFields(t *testing.T) {
	repo := &stubRepo{}
	svc := NewHistoryService(repo)

	if err := svc.SaveTelemetry("CAR001", 42.5, 44.7866, 20.4489); err != nil {
		t.Fatalf("SaveTelemetry returned error: %v", err)
	}
	if len(repo.created) != 1 {
		t.Fatalf("expected 1 record persisted, got %d", len(repo.created))
	}
	got := repo.created[0]
	if got.CarID != "CAR001" || got.Fuel != 42.5 || got.Latitude != 44.7866 || got.Longitude != 20.4489 {
		t.Errorf("persisted record has wrong fields: %+v", got)
	}
}

func TestGetRecentHistoryClampsLimit(t *testing.T) {
	cases := []struct {
		in   int
		want int
	}{
		{in: 0, want: 100},
		{in: -5, want: 100},
		{in: 50, want: 50},
		{in: 5000, want: 1000},
	}
	for _, tc := range cases {
		repo := &stubRepo{}
		svc := NewHistoryService(repo)
		if _, err := svc.GetRecentHistory("CAR001", tc.in); err != nil {
			t.Fatalf("GetRecentHistory(%d) error: %v", tc.in, err)
		}
		if repo.lastLimit != tc.want {
			t.Errorf("limit %d clamped to %d, want %d", tc.in, repo.lastLimit, tc.want)
		}
	}
}

func TestGetHistoryInRangeRejectsInvertedRange(t *testing.T) {
	repo := &stubRepo{}
	svc := NewHistoryService(repo)
	start := time.Date(2026, 5, 29, 12, 0, 0, 0, time.UTC)
	end := start.Add(-time.Hour)

	if _, err := svc.GetHistoryInRange("CAR001", start, end); err == nil {
		t.Fatal("expected error when start is after end, got nil")
	}
	if repo.rangeQueried {
		t.Error("repository should not be queried for an inverted range")
	}
}

func TestCleanupOldRecordsForwardsDays(t *testing.T) {
	repo := &stubRepo{}
	svc := NewHistoryService(repo)

	if err := svc.CleanupOldRecords(30); err != nil {
		t.Fatalf("CleanupOldRecords returned error: %v", err)
	}
	if repo.lastDays != 30 {
		t.Errorf("days forwarded = %d, want 30", repo.lastDays)
	}
}

func TestSaveTelemetryPropagatesRepoError(t *testing.T) {
	repo := &stubRepo{createErr: errors.New("db down")}
	svc := NewHistoryService(repo)

	if err := svc.SaveTelemetry("CAR001", 50, 44.0, 20.0); err == nil {
		t.Fatal("expected repo error to propagate, got nil")
	}
}
