package history

import (
	"testing"
	"time"
)

// mockHistoryRepo is an in-memory HistoryRepo used in tests.
type mockHistoryRepo struct {
	records []CarHistory
}

func (m *mockHistoryRepo) Create(h *CarHistory) error {
	m.records = append(m.records, *h)
	return nil
}

func (m *mockHistoryRepo) FindByCarID(carID string, limit int) ([]CarHistory, error) {
	var result []CarHistory
	for _, r := range m.records {
		if r.CarID == carID {
			result = append(result, r)
		}
	}
	if len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (m *mockHistoryRepo) FindByCarIDAndTimeRange(carID string, start, end time.Time) ([]CarHistory, error) {
	var result []CarHistory
	for _, r := range m.records {
		if r.CarID == carID && !r.Timestamp.Before(start) && !r.Timestamp.After(end) {
			result = append(result, r)
		}
	}
	return result, nil
}

func (m *mockHistoryRepo) GetLatestByCarID(carID string) (*CarHistory, error) {
	for i := len(m.records) - 1; i >= 0; i-- {
		if m.records[i].CarID == carID {
			r := m.records[i]
			return &r, nil
		}
	}
	return nil, nil
}

func (m *mockHistoryRepo) DeleteOlderThan(_ int) error { return nil }

func TestSaveTelemetry_EmptyCarID(t *testing.T) {
	svc := NewHistoryService(&mockHistoryRepo{})

	err := svc.SaveTelemetry("", 50.0, 44.8, 20.4)
	if err == nil {
		t.Fatal("expected error for empty car_id, got nil")
	}
}

func TestSaveTelemetry_Success(t *testing.T) {
	repo := &mockHistoryRepo{}
	svc := NewHistoryService(repo)

	if err := svc.SaveTelemetry("CAR001", 75.5, 44.8, 20.4); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(repo.records))
	}
}

func TestGetRecentHistory_LimitClampedToDefault(t *testing.T) {
	svc := NewHistoryService(&mockHistoryRepo{})

	// limit 0 should be clamped to 100 — must not panic
	if _, err := svc.GetRecentHistory("CAR001", 0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetRecentHistory_LimitClampedToMax(t *testing.T) {
	svc := NewHistoryService(&mockHistoryRepo{})

	// limit > 1000 should be clamped to 1000 — must not panic
	if _, err := svc.GetRecentHistory("CAR001", 9999); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetHistoryInRange_InvalidRange(t *testing.T) {
	svc := NewHistoryService(&mockHistoryRepo{})

	end := time.Now()
	start := end.Add(time.Hour) // start is after end

	_, err := svc.GetHistoryInRange("CAR001", start, end)
	if err == nil {
		t.Fatal("expected error when start is after end, got nil")
	}
}

func TestGetHistoryInRange_ValidRange(t *testing.T) {
	svc := NewHistoryService(&mockHistoryRepo{})

	start := time.Now().Add(-time.Hour)
	end := time.Now()

	if _, err := svc.GetHistoryInRange("CAR001", start, end); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
