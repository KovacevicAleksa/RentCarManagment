package carstats

import (
	"sync"
	"time"
)

const (
	// Overheating episode detection uses the same hysteresis band as the
	// notification alarm: a car begins an episode when it crosses engineTempHigh
	// and must drop below engineTempLow before a new episode can be counted.
	engineTempHigh = 110.0
	engineTempLow  = 105.0

	// sampleInterval is the telemetry cadence; total operating time is derived
	// as totalReadings * sampleInterval.
	sampleInterval = 5 * time.Second

	// warmupDuration is the minimum operating time before the frequency is
	// meaningful. Below it the rate is reported as 0 to avoid wild values from a
	// single early episode (e.g. 1 event in the first 3 minutes reading "20/h").
	warmupDuration = 10 * time.Minute
)

type state struct {
	totalReadings  int64
	overheatEvents int64
	maxEngineTemp  float64
	lastOverheatAt time.Time
	active         bool
}

// Service accumulates per-car engine statistics in memory and periodically
// persists them. It is safe for concurrent use from MQTT message handlers.
type Service struct {
	repo Repository
	now  func() time.Time

	mu   sync.Mutex
	cars map[string]*state
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
		now:  time.Now,
		cars: make(map[string]*state),
	}
}

// Load restores accumulated counters from the repository so statistics survive
// a restart. Call once at startup before Observe.
func (s *Service) Load() error {
	loaded, err := s.repo.LoadAll()
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, st := range loaded {
		cs := &state{
			totalReadings:  st.TotalReadings,
			overheatEvents: st.OverheatEvents,
			maxEngineTemp:  st.MaxEngineTemp,
		}
		if st.LastOverheatAt != nil {
			cs.lastOverheatAt = *st.LastOverheatAt
		}
		s.cars[st.CarID] = cs
	}
	return nil
}

// Observe records one telemetry reading for a car and returns the car's current
// overheat frequency (episodes per operating hour).
func (s *Service) Observe(carID string, engineTemp float64) float64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	st := s.cars[carID]
	if st == nil {
		st = &state{}
		s.cars[carID] = st
	}

	st.totalReadings++
	if engineTemp > st.maxEngineTemp {
		st.maxEngineTemp = engineTemp
	}

	switch {
	case engineTemp >= engineTempHigh:
		if !st.active {
			st.overheatEvents++
			st.lastOverheatAt = s.now()
			st.active = true
		}
	case engineTemp < engineTempLow:
		st.active = false
	}

	return frequency(st)
}

// frequency returns overheat episodes per operating hour, or 0 until the car has
// accumulated at least warmupDuration of operating time.
func frequency(st *state) float64 {
	operatingHours := float64(st.totalReadings) * sampleInterval.Seconds() / 3600
	if operatingHours < warmupDuration.Hours() {
		return 0
	}
	return float64(st.overheatEvents) / operatingHours
}

// Flush persists a snapshot of all accumulated statistics.
func (s *Service) Flush() error {
	s.mu.Lock()
	now := s.now()
	items := make([]Stats, 0, len(s.cars))
	for id, st := range s.cars {
		item := Stats{
			CarID:             id,
			TotalReadings:     st.totalReadings,
			OverheatEvents:    st.overheatEvents,
			MaxEngineTemp:     st.maxEngineTemp,
			OverheatFrequency: frequency(st),
			UpdatedAt:         now,
		}
		if !st.lastOverheatAt.IsZero() {
			t := st.lastOverheatAt
			item.LastOverheatAt = &t
		}
		items = append(items, item)
	}
	s.mu.Unlock()

	return s.repo.UpsertBatch(items)
}
