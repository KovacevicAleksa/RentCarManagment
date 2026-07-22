package notification

import (
	"testing"
	"time"
)

type stubRepo struct {
	created []*Notification
	err     error
}

func (s *stubRepo) Create(n *Notification) error {
	if s.err != nil {
		return s.err
	}
	s.created = append(s.created, n)
	return nil
}
func (s *stubRepo) FindAll(limit int) ([]Notification, error) { return nil, nil }
func (s *stubRepo) MarkAllRead() error                        { return nil }
func (s *stubRepo) CountUnread() (int64, error)               { return 0, nil }

type stubBroadcaster struct{ calls int }

func (b *stubBroadcaster) BroadcastNotification(data any) { b.calls++ }

// newTestService wires the service with a controllable clock.
func newTestService() (*Service, *stubRepo, *stubBroadcaster, *time.Time) {
	repo := &stubRepo{}
	bc := &stubBroadcaster{}
	now := time.Now()
	s := NewService(repo, bc)
	s.now = func() time.Time { return now }
	return s, repo, bc, &now
}

func TestEngineAlarmFiresOnceOnRisingEdge(t *testing.T) {
	s, repo, bc, _ := newTestService()

	s.CheckEngineTemp("CAR001", 112) // crosses high → alarm
	s.CheckEngineTemp("CAR001", 115) // still hot → no new alarm
	s.CheckEngineTemp("CAR001", 111) // still above low → no new alarm

	if len(repo.created) != 1 {
		t.Fatalf("expected exactly 1 notification, got %d", len(repo.created))
	}
	if bc.calls != 1 {
		t.Errorf("expected 1 broadcast, got %d", bc.calls)
	}
	if repo.created[0].Type != TypeAlert || repo.created[0].Source != SourceSystem {
		t.Errorf("unexpected notification meta: %+v", repo.created[0])
	}
	if repo.created[0].CarID == nil || *repo.created[0].CarID != "CAR001" {
		t.Errorf("expected car_id CAR001, got %v", repo.created[0].CarID)
	}
}

func TestEngineAlarmReArmsAfterCooldownAndHysteresis(t *testing.T) {
	s, repo, _, now := newTestService()

	s.CheckEngineTemp("CAR001", 112) // alarm #1
	s.CheckEngineTemp("CAR001", 104) // drops below low → re-armed

	// Heats up again but still within cooldown window → suppressed.
	*now = now.Add(2 * time.Minute)
	s.CheckEngineTemp("CAR001", 113)
	if len(repo.created) != 1 {
		t.Fatalf("alarm within cooldown should be suppressed, got %d", len(repo.created))
	}

	// Cool down, wait past cooldown, heat again → alarm #2.
	s.CheckEngineTemp("CAR001", 100)
	*now = now.Add(6 * time.Minute)
	s.CheckEngineTemp("CAR001", 114)
	if len(repo.created) != 2 {
		t.Fatalf("alarm after cooldown should fire, got %d", len(repo.created))
	}
}

func TestEngineAlarmIndependentPerCar(t *testing.T) {
	s, repo, _, _ := newTestService()

	s.CheckEngineTemp("CAR001", 112)
	s.CheckEngineTemp("CAR002", 113)

	if len(repo.created) != 2 {
		t.Fatalf("each car should alarm independently, got %d", len(repo.created))
	}
}

func TestNormalTemperatureNeverAlarms(t *testing.T) {
	s, repo, _, _ := newTestService()

	for _, temp := range []float64{70, 85, 95, 104, 109.9} {
		s.CheckEngineTemp("CAR001", temp)
	}
	if len(repo.created) != 0 {
		t.Fatalf("temperatures below threshold must not alarm, got %d", len(repo.created))
	}
}
