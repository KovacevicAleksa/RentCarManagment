package carstats

import (
	"sort"
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

	// reliabilityWindow is the trailing period over which fault episodes count
	// toward the 0-100 reliability score. Episodes older than this are ignored,
	// so a car recovers its rating a month after its last fault.
	reliabilityWindow = 30 * 24 * time.Hour

	// overheatPenalty is the number of points each overheat episode in the
	// window subtracts from the temperature score.
	overheatPenalty = 10.0

	// checkEnginePenalty is the number of points each check-engine episode in
	// the window subtracts from the check-engine score.
	checkEnginePenalty = 5.0

	// temperatureWeight and checkEngineWeight blend the two sub-scores into the
	// overall reliability. Overheating is weighted more heavily than a
	// check-engine lamp because it is the more damaging fault. They sum to 1.
	temperatureWeight = 0.6
	checkEngineWeight = 0.4
)

type state struct {
	totalReadings  int64
	overheatEvents int64
	maxEngineTemp  float64
	lastOverheatAt time.Time
	active         bool

	// overheatTimes holds the timestamp of each overheat episode's rising edge,
	// kept only for episodes still inside the reliability window.
	overheatTimes []time.Time

	// checkEngineActive tracks the check-engine lamp so only its rising edge
	// (off -> on) counts as a new episode. checkEngineTimes holds the rising-edge
	// timestamps still inside the reliability window.
	checkEngineActive bool
	checkEngineTimes  []time.Time
}

// Service accumulates per-car engine statistics in memory and periodically
// persists them. It is safe for concurrent use from MQTT message handlers.
type Service struct {
	repo Repository
	now  func() time.Time

	mu            sync.Mutex
	cars          map[string]*state
	pendingEvents []Event
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
	events, err := s.repo.LoadEventsSince(s.now().Add(-reliabilityWindow))
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
	for _, e := range events {
		cs := s.cars[e.CarID]
		if cs == nil {
			cs = &state{}
			s.cars[e.CarID] = cs
		}
		switch e.Type {
		case EventOverheat:
			cs.overheatTimes = append(cs.overheatTimes, e.OccurredAt)
		case EventCheckEngine:
			cs.checkEngineTimes = append(cs.checkEngineTimes, e.OccurredAt)
		}
	}
	return nil
}

// Result is the derived, per-reading health snapshot returned by Observe and
// attached to the outgoing telemetry broadcast.
type Result struct {
	// OverheatFrequency is the lifetime episodes-per-operating-hour rate.
	OverheatFrequency float64
	// OverheatCount is the number of overheat episodes within the reliability
	// window (trailing 30 days).
	OverheatCount int
	// TemperatureScore is the 0-100 engine-temperature health rating derived
	// from OverheatCount (100 = no overheating in the window).
	TemperatureScore float64
	// CheckEngineCount is the number of check-engine episodes within the
	// reliability window.
	CheckEngineCount int
	// CheckEngineScore is the 0-100 rating derived from CheckEngineCount
	// (100 = the check-engine lamp never lit in the window).
	CheckEngineScore float64
	// Reliability is the overall 0-100 rating for the window, a weighted blend
	// of TemperatureScore and CheckEngineScore.
	Reliability float64
}

// Observe records one telemetry reading for a car and returns the car's current
// derived health snapshot.
func (s *Service) Observe(carID string, engineTemp float64, checkEngine bool) Result {
	s.mu.Lock()
	defer s.mu.Unlock()

	st := s.cars[carID]
	if st == nil {
		st = &state{}
		s.cars[carID] = st
	}

	now := s.now()

	st.totalReadings++
	if engineTemp > st.maxEngineTemp {
		st.maxEngineTemp = engineTemp
	}

	switch {
	case engineTemp >= engineTempHigh:
		if !st.active {
			st.overheatEvents++
			st.lastOverheatAt = now
			st.overheatTimes = append(st.overheatTimes, now)
			s.pendingEvents = append(s.pendingEvents, Event{CarID: carID, Type: EventOverheat, OccurredAt: now})
			st.active = true
		}
	case engineTemp < engineTempLow:
		st.active = false
	}

	if checkEngine {
		if !st.checkEngineActive {
			st.checkEngineTimes = append(st.checkEngineTimes, now)
			s.pendingEvents = append(s.pendingEvents, Event{CarID: carID, Type: EventCheckEngine, OccurredAt: now})
			st.checkEngineActive = true
		}
	} else {
		st.checkEngineActive = false
	}

	cutoff := now.Add(-reliabilityWindow)
	var overheatCount, checkEngineCount int
	st.overheatTimes, overheatCount = pruneAndCount(st.overheatTimes, cutoff)
	st.checkEngineTimes, checkEngineCount = pruneAndCount(st.checkEngineTimes, cutoff)

	temperatureScore := clampScore(100 - float64(overheatCount)*overheatPenalty)
	checkEngineScore := clampScore(100 - float64(checkEngineCount)*checkEnginePenalty)

	return Result{
		OverheatFrequency: frequency(st),
		OverheatCount:     overheatCount,
		TemperatureScore:  temperatureScore,
		CheckEngineCount:  checkEngineCount,
		CheckEngineScore:  checkEngineScore,
		Reliability:       clampScore(temperatureWeight*temperatureScore + checkEngineWeight*checkEngineScore),
	}
}

// EventBucket is one day's fault-episode counts for a car.
type EventBucket struct {
	Date        string `json:"date"` // YYYY-MM-DD (UTC)
	Overheat    int    `json:"overheat"`
	CheckEngine int    `json:"check_engine"`
}

// EventStatsResult is the per-car fault history over the reliability window,
// bucketed by day for charting plus window totals.
type EventStatsResult struct {
	CarID            string        `json:"car_id"`
	WindowDays       int           `json:"window_days"`
	OverheatTotal    int           `json:"overheat_total"`
	CheckEngineTotal int           `json:"check_engine_total"`
	Buckets          []EventBucket `json:"buckets"`
}

// EventStats returns the car's overheat and check-engine episodes over the
// reliability window, bucketed by day.
func (s *Service) EventStats(carID string) (EventStatsResult, error) {
	events, err := s.repo.LoadEventsByCarSince(carID, s.now().Add(-reliabilityWindow))
	if err != nil {
		return EventStatsResult{}, err
	}
	return buildEventStats(carID, events), nil
}

// buildEventStats groups events into per-day buckets (UTC) sorted chronologically
// and tallies window totals.
func buildEventStats(carID string, events []Event) EventStatsResult {
	res := EventStatsResult{CarID: carID, WindowDays: int(reliabilityWindow.Hours() / 24)}
	byDay := make(map[string]*EventBucket)
	for _, e := range events {
		day := e.OccurredAt.UTC().Format("2006-01-02")
		b := byDay[day]
		if b == nil {
			b = &EventBucket{Date: day}
			byDay[day] = b
		}
		switch e.Type {
		case EventOverheat:
			b.Overheat++
			res.OverheatTotal++
		case EventCheckEngine:
			b.CheckEngine++
			res.CheckEngineTotal++
		}
	}

	days := make([]string, 0, len(byDay))
	for d := range byDay {
		days = append(days, d)
	}
	sort.Strings(days)
	for _, d := range days {
		res.Buckets = append(res.Buckets, *byDay[d])
	}
	return res
}

// pruneAndCount drops timestamps older than cutoff and returns the retained
// slice together with its length. Order is not assumed.
func pruneAndCount(times []time.Time, cutoff time.Time) ([]time.Time, int) {
	kept := times[:0]
	for _, t := range times {
		if !t.Before(cutoff) {
			kept = append(kept, t)
		}
	}
	return kept, len(kept)
}

// clampScore constrains a raw score to the 0-100 range.
func clampScore(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
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
	events := append([]Event(nil), s.pendingEvents...)
	s.mu.Unlock()

	if err := s.repo.UpsertBatch(items); err != nil {
		return err
	}
	if len(events) > 0 {
		if err := s.repo.AppendEvents(events); err != nil {
			return err
		}
		// Drop only the events just persisted; any appended during the DB call
		// are at the tail and remain buffered for the next flush.
		s.mu.Lock()
		s.pendingEvents = s.pendingEvents[len(events):]
		s.mu.Unlock()
	}
	return s.repo.DeleteEventsBefore(now.Add(-reliabilityWindow))
}
