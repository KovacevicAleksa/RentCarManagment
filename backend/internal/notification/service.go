package notification

import (
	"fmt"
	"sync"
	"time"
)

const (
	defaultListLimit = 100

	// Engine-overheating alarm thresholds. A car alarms when it crosses
	// engineTempHigh and must drop below engineTempLow (hysteresis) before it
	// can alarm again. alertCooldown additionally caps how often a single car
	// can alarm, so an oscillating sensor cannot spam notifications.
	engineTempHigh = 110.0
	engineTempLow  = 105.0
	alertCooldown  = 5 * time.Minute
)

// Broadcaster pushes a notification to all connected WebSocket clients.
type Broadcaster interface {
	BroadcastNotification(data any)
}

type alarmState struct {
	active    bool
	lastAlert time.Time
}

type Service struct {
	repo        Repository
	broadcaster Broadcaster
	now         func() time.Time

	mu     sync.Mutex
	alarms map[string]alarmState
}

func NewService(repo Repository, broadcaster Broadcaster) *Service {
	return &Service{
		repo:        repo,
		broadcaster: broadcaster,
		now:         time.Now,
		alarms:      make(map[string]alarmState),
	}
}

// Create persists a notification and broadcasts it to connected clients.
func (s *Service) Create(title, message, ntype, source string, carID *string) (*Notification, error) {
	n := &Notification{
		Title:   title,
		Message: message,
		Type:    ntype,
		Source:  source,
		CarID:   carID,
	}
	if err := s.repo.Create(n); err != nil {
		return nil, err
	}
	if s.broadcaster != nil {
		s.broadcaster.BroadcastNotification(n)
	}
	return n, nil
}

func (s *Service) List() ([]Notification, error) {
	return s.repo.FindAll(defaultListLimit)
}

func (s *Service) MarkAllRead() error {
	return s.repo.MarkAllRead()
}

func (s *Service) UnreadCount() (int64, error) {
	return s.repo.CountUnread()
}

// CheckEngineTemp evaluates a car's engine temperature and emits an
// overheating notification on the rising edge, subject to hysteresis and a
// per-car cooldown. It is safe to call concurrently from MQTT message handlers.
func (s *Service) CheckEngineTemp(carID string, engineTemp float64) {
	s.mu.Lock()
	st := s.alarms[carID]
	now := s.now()

	shouldAlert := false
	switch {
	case engineTemp >= engineTempHigh:
		if !st.active && now.Sub(st.lastAlert) >= alertCooldown {
			shouldAlert = true
			st.lastAlert = now
		}
		st.active = true
	case engineTemp < engineTempLow:
		st.active = false
	}
	s.alarms[carID] = st
	s.mu.Unlock()

	if !shouldAlert {
		return
	}

	title := "Pregrevanje motora"
	message := fmt.Sprintf("Vozilo %s: temperatura motora %.1f°C prelazi bezbednu granicu.", carID, engineTemp)
	if _, err := s.Create(title, message, TypeAlert, SourceSystem, &carID); err != nil {
		// Re-arm so a transient persistence error doesn't suppress the next alarm.
		s.mu.Lock()
		st := s.alarms[carID]
		st.lastAlert = time.Time{}
		st.active = false
		s.alarms[carID] = st
		s.mu.Unlock()
	}
}
