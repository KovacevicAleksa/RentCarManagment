package history

import (
	"sync"
	"testing"
	"time"
)

// mockHistoryRepo is an in-memory HistoryRepo used in tests. It is mutex-guarded
// because the service flushes from a background goroutine while tests read state.
type mockHistoryRepo struct {
	mu      sync.Mutex
	records []CarHistory
}

func (m *mockHistoryRepo) Create(h *CarHistory) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.records = append(m.records, *h)
	return nil
}

func (m *mockHistoryRepo) CreateBatch(items []*CarHistory) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, h := range items {
		m.records = append(m.records, *h)
	}
	return nil
}

func (m *mockHistoryRepo) count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.records)
}

func (m *mockHistoryRepo) FindByCarID(carID string, limit int) ([]CarHistory, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
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
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []CarHistory
	for _, r := range m.records {
		if r.CarID == carID && !r.Timestamp.Before(start) && !r.Timestamp.After(end) {
			result = append(result, r)
		}
	}
	return result, nil
}

func (m *mockHistoryRepo) GetLatestByCarID(carID string) (*CarHistory, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
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

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if repo.count() == 1 {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("expected 1 record, got %d", repo.count())
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
