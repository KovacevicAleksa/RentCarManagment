package carstats

import (
	"math"
	"testing"
	"time"
)

// overheatEpisode drives one full overheat episode: a rising edge above the
// high threshold followed by a drop below the low threshold to re-arm.
func overheatEpisode(s *Service, carID string) {
	s.Observe(carID, 112, false)
	s.Observe(carID, 100, false)
}

type stubRepo struct {
	loaded    []Stats
	upserted  []Stats
	loadErr   error
	upsertErr error

	events         []Event   // pre-seeded, returned by LoadEventsSince
	appended       []Event   // captured from AppendEvents
	loadEventsFrom time.Time // captured "since" argument
	prunedBefore   time.Time // captured DeleteEventsBefore argument
}

func (s *stubRepo) LoadAll() ([]Stats, error) { return s.loaded, s.loadErr }
func (s *stubRepo) UpsertBatch(items []Stats) error {
	if s.upsertErr != nil {
		return s.upsertErr
	}
	s.upserted = items
	return nil
}

func (s *stubRepo) AppendEvents(events []Event) error {
	s.appended = append(s.appended, events...)
	return nil
}

func (s *stubRepo) LoadEventsSince(since time.Time) ([]Event, error) {
	s.loadEventsFrom = since
	var out []Event
	for _, e := range s.events {
		if !e.OccurredAt.Before(since) {
			out = append(out, e)
		}
	}
	return out, nil
}

func (s *stubRepo) LoadEventsByCarSince(carID string, since time.Time) ([]Event, error) {
	var out []Event
	for _, e := range s.events {
		if e.CarID == carID && !e.OccurredAt.Before(since) {
			out = append(out, e)
		}
	}
	return out, nil
}

func (s *stubRepo) DeleteEventsBefore(cutoff time.Time) error {
	s.prunedBefore = cutoff
	return nil
}

// feed replays n readings at the given temperature and returns the last
// reported frequency.
func feed(s *Service, carID string, temp float64, n int) float64 {
	var freq float64
	for i := 0; i < n; i++ {
		freq = s.Observe(carID, temp, false).OverheatFrequency
	}
	return freq
}

// readingsPerHour is how many 5s samples make up one hour of operating time,
// used to drive the frequency equation to clean, assertable values.
const readingsPerHour = 720

func TestOverheatFrequencyIsEventsPerOperatingHour(t *testing.T) {
	s := NewService(&stubRepo{})

	// Three distinct overheating episodes: rising edge then drop below low to
	// re-arm. That is 6 readings.
	for i := 0; i < 3; i++ {
		s.Observe("CAR001", 112, false)
		s.Observe("CAR001", 100, false)
	}
	// Pad with normal readings up to exactly one operating hour.
	freq := feed(s, "CAR001", 70, readingsPerHour-6)

	if math.Abs(freq-3.0) > 1e-9 {
		t.Fatalf("expected 3.0 events/hour, got %v", freq)
	}
}

func TestSustainedOverheatCountsAsOneEpisode(t *testing.T) {
	s := NewService(&stubRepo{})

	// Stays above the high threshold across several readings: one episode.
	s.Observe("CAR001", 112, false)
	s.Observe("CAR001", 115, false)
	s.Observe("CAR001", 111, false) // still above low, not re-armed
	freq := feed(s, "CAR001", 70, readingsPerHour-3)

	if math.Abs(freq-1.0) > 1e-9 {
		t.Fatalf("expected 1.0 event/hour for a single episode, got %v", freq)
	}
}

func TestFrequencyIsZeroDuringWarmup(t *testing.T) {
	s := NewService(&stubRepo{})

	// One event but only a handful of readings (< 10 min of operation).
	freq := s.Observe("CAR001", 112, false).OverheatFrequency

	if freq != 0 {
		t.Fatalf("expected 0 during warmup, got %v", freq)
	}
}

func TestStatsIndependentPerCar(t *testing.T) {
	s := NewService(&stubRepo{})

	s.Observe("CAR001", 112, false) // CAR001: one episode
	s.Observe("CAR001", 100, false)
	f1 := feed(s, "CAR001", 70, readingsPerHour-2)
	f2 := feed(s, "CAR002", 70, readingsPerHour) // CAR002: never overheats

	if math.Abs(f1-1.0) > 1e-9 {
		t.Fatalf("CAR001 expected 1.0, got %v", f1)
	}
	if f2 != 0 {
		t.Fatalf("CAR002 expected 0, got %v", f2)
	}
}

func TestLoadRestoresCountersAcrossRestart(t *testing.T) {
	repo := &stubRepo{loaded: []Stats{
		{CarID: "CAR001", TotalReadings: readingsPerHour, OverheatEvents: 3},
	}}
	s := NewService(repo)
	if err := s.Load(); err != nil {
		t.Fatal(err)
	}

	// One more normal reading: counters continue from the loaded values.
	freq := s.Observe("CAR001", 70, false).OverheatFrequency

	want := 3.0 / (float64(readingsPerHour+1) * 5 / 3600)
	if math.Abs(freq-want) > 1e-9 {
		t.Fatalf("expected %v after load, got %v", want, freq)
	}
}

func TestTemperatureScoreDropsPerOverheatEpisodeInWindow(t *testing.T) {
	s := NewService(&stubRepo{})

	// Two overheat episodes in the last 30 days: 100 - 2*10 = 80.
	overheatEpisode(s, "CAR001")
	res := s.Observe("CAR001", 112, false)

	if res.OverheatCount != 2 {
		// The second Observe(112) above starts a third episode only if re-armed;
		// here it stays active from the last re-arm so it is a fresh rising edge.
		t.Fatalf("expected 2 episodes in window, got %d", res.OverheatCount)
	}
	if math.Abs(res.TemperatureScore-80) > 1e-9 {
		t.Fatalf("expected temperature score 80, got %v", res.TemperatureScore)
	}
}

func TestTemperatureScoreIsFullWithNoOverheats(t *testing.T) {
	s := NewService(&stubRepo{})

	res := s.Observe("CAR001", 70, false)

	if res.OverheatCount != 0 {
		t.Fatalf("expected 0 episodes, got %d", res.OverheatCount)
	}
	if res.TemperatureScore != 100 {
		t.Fatalf("expected temperature score 100, got %v", res.TemperatureScore)
	}
}

func TestTemperatureScoreFloorsAtZero(t *testing.T) {
	s := NewService(&stubRepo{})

	// 12 episodes would be -20 before clamping.
	for i := 0; i < 12; i++ {
		overheatEpisode(s, "CAR001")
	}
	res := s.Observe("CAR001", 70, false)

	if res.OverheatCount != 12 {
		t.Fatalf("expected 12 episodes, got %d", res.OverheatCount)
	}
	if res.TemperatureScore != 0 {
		t.Fatalf("expected temperature score clamped to 0, got %v", res.TemperatureScore)
	}
}

func TestOverheatEpisodesOutsideWindowAreExcluded(t *testing.T) {
	s := NewService(&stubRepo{})
	now := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	s.now = func() time.Time { return now }

	// Five episodes now, then jump past the 30-day window and add one more.
	for i := 0; i < 5; i++ {
		overheatEpisode(s, "CAR001")
	}
	now = now.Add(31 * 24 * time.Hour)
	overheatEpisode(s, "CAR001")
	res := s.Observe("CAR001", 70, false)

	if res.OverheatCount != 1 {
		t.Fatalf("expected only the 1 recent episode in window, got %d", res.OverheatCount)
	}
	if math.Abs(res.TemperatureScore-90) > 1e-9 {
		t.Fatalf("expected temperature score 90, got %v", res.TemperatureScore)
	}
}

func TestCheckEngineCountsRisingEdgesOnly(t *testing.T) {
	s := NewService(&stubRepo{})

	s.Observe("CAR001", 70, true)        // episode 1: off -> on
	s.Observe("CAR001", 70, true)        // still on, not a new episode
	s.Observe("CAR001", 70, false)       // cleared
	res := s.Observe("CAR001", 70, true) // episode 2: off -> on

	if res.CheckEngineCount != 2 {
		t.Fatalf("expected 2 check-engine episodes, got %d", res.CheckEngineCount)
	}
	if math.Abs(res.CheckEngineScore-90) > 1e-9 {
		t.Fatalf("expected check-engine score 90, got %v", res.CheckEngineScore)
	}
}

func TestCheckEngineScoreIsFullWhenHealthy(t *testing.T) {
	s := NewService(&stubRepo{})

	res := s.Observe("CAR001", 70, false)

	if res.CheckEngineCount != 0 {
		t.Fatalf("expected 0 check-engine episodes, got %d", res.CheckEngineCount)
	}
	if res.CheckEngineScore != 100 {
		t.Fatalf("expected check-engine score 100, got %v", res.CheckEngineScore)
	}
}

func TestCheckEngineScoreFloorsAtZero(t *testing.T) {
	s := NewService(&stubRepo{})

	// 21 episodes would be -5 before clamping.
	for i := 0; i < 21; i++ {
		s.Observe("CAR001", 70, true)
		s.Observe("CAR001", 70, false)
	}
	res := s.Observe("CAR001", 70, false)

	if res.CheckEngineCount != 21 {
		t.Fatalf("expected 21 episodes, got %d", res.CheckEngineCount)
	}
	if res.CheckEngineScore != 0 {
		t.Fatalf("expected check-engine score clamped to 0, got %v", res.CheckEngineScore)
	}
}

func TestCheckEngineEpisodesOutsideWindowAreExcluded(t *testing.T) {
	s := NewService(&stubRepo{})
	now := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	s.now = func() time.Time { return now }

	for i := 0; i < 4; i++ {
		s.Observe("CAR001", 70, true)
		s.Observe("CAR001", 70, false)
	}
	now = now.Add(31 * 24 * time.Hour)
	s.Observe("CAR001", 70, true)
	res := s.Observe("CAR001", 70, false)

	if res.CheckEngineCount != 1 {
		t.Fatalf("expected only the 1 recent check-engine episode, got %d", res.CheckEngineCount)
	}
	if math.Abs(res.CheckEngineScore-95) > 1e-9 {
		t.Fatalf("expected check-engine score 95, got %v", res.CheckEngineScore)
	}
}

func TestEventStatsBucketsByDayWithTotals(t *testing.T) {
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	repo := &stubRepo{events: []Event{
		{CarID: "CAR001", Type: EventOverheat, OccurredAt: now.Add(-2 * time.Hour)},    // 07-22
		{CarID: "CAR001", Type: EventCheckEngine, OccurredAt: now.Add(-3 * time.Hour)}, // 07-22
		{CarID: "CAR001", Type: EventOverheat, OccurredAt: now.Add(-26 * time.Hour)},   // 07-21
		{CarID: "CAR002", Type: EventOverheat, OccurredAt: now.Add(-1 * time.Hour)},    // other car
	}}
	s := NewService(repo)
	s.now = func() time.Time { return now }

	stats, err := s.EventStats("CAR001")
	if err != nil {
		t.Fatal(err)
	}

	if stats.OverheatTotal != 2 {
		t.Errorf("OverheatTotal = %d, want 2", stats.OverheatTotal)
	}
	if stats.CheckEngineTotal != 1 {
		t.Errorf("CheckEngineTotal = %d, want 1", stats.CheckEngineTotal)
	}
	if len(stats.Buckets) != 2 {
		t.Fatalf("expected 2 day buckets, got %d: %+v", len(stats.Buckets), stats.Buckets)
	}
	// Buckets are chronological.
	if stats.Buckets[0].Date != "2026-07-21" || stats.Buckets[1].Date != "2026-07-22" {
		t.Errorf("buckets not sorted by date: %+v", stats.Buckets)
	}
	if stats.Buckets[0].Overheat != 1 || stats.Buckets[0].CheckEngine != 0 {
		t.Errorf("07-21 bucket = %+v, want overheat 1 / check 0", stats.Buckets[0])
	}
	if stats.Buckets[1].Overheat != 1 || stats.Buckets[1].CheckEngine != 1 {
		t.Errorf("07-22 bucket = %+v, want overheat 1 / check 1", stats.Buckets[1])
	}
}

func TestFlushPersistsEpisodeEvents(t *testing.T) {
	repo := &stubRepo{}
	s := NewService(repo)

	overheatEpisode(s, "CAR001")  // one overheat rising edge
	s.Observe("CAR001", 70, true) // one check-engine rising edge

	if err := s.Flush(); err != nil {
		t.Fatal(err)
	}

	if len(repo.appended) != 2 {
		t.Fatalf("expected 2 events appended, got %d: %+v", len(repo.appended), repo.appended)
	}
	var overheat, checkEngine int
	for _, e := range repo.appended {
		if e.CarID != "CAR001" {
			t.Errorf("unexpected car in event: %+v", e)
		}
		switch e.Type {
		case EventOverheat:
			overheat++
		case EventCheckEngine:
			checkEngine++
		default:
			t.Errorf("unexpected event type: %q", e.Type)
		}
	}
	if overheat != 1 || checkEngine != 1 {
		t.Fatalf("expected 1 overheat + 1 check-engine, got %d/%d", overheat, checkEngine)
	}
}

func TestFlushDoesNotReAppendAlreadyPersistedEvents(t *testing.T) {
	repo := &stubRepo{}
	s := NewService(repo)

	overheatEpisode(s, "CAR001")
	if err := s.Flush(); err != nil {
		t.Fatal(err)
	}
	if err := s.Flush(); err != nil { // second flush: nothing new
		t.Fatal(err)
	}

	if len(repo.appended) != 1 {
		t.Fatalf("expected exactly 1 event across two flushes, got %d", len(repo.appended))
	}
}

func TestLoadRehydratesReliabilityWindow(t *testing.T) {
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	repo := &stubRepo{events: []Event{
		{CarID: "CAR001", Type: EventOverheat, OccurredAt: now.Add(-24 * time.Hour)},
		{CarID: "CAR001", Type: EventOverheat, OccurredAt: now.Add(-40 * 24 * time.Hour)}, // outside window
		{CarID: "CAR001", Type: EventCheckEngine, OccurredAt: now.Add(-2 * time.Hour)},
	}}
	s := NewService(repo)
	s.now = func() time.Time { return now }

	if err := s.Load(); err != nil {
		t.Fatal(err)
	}

	if want := now.Add(-reliabilityWindow); !repo.loadEventsFrom.Equal(want) {
		t.Fatalf("Load queried events since %v, want %v", repo.loadEventsFrom, want)
	}

	res := s.Observe("CAR001", 70, false) // normal reading, no new episode

	if res.OverheatCount != 1 {
		t.Fatalf("expected 1 overheat in window after load, got %d", res.OverheatCount)
	}
	if res.CheckEngineCount != 1 {
		t.Fatalf("expected 1 check-engine in window after load, got %d", res.CheckEngineCount)
	}
	// 0.6*90 + 0.4*95 = 92
	if math.Abs(res.Reliability-92) > 1e-9 {
		t.Fatalf("expected reliability 92 after load, got %v", res.Reliability)
	}
}

func TestFlushPrunesEventsOlderThanWindow(t *testing.T) {
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	repo := &stubRepo{}
	s := NewService(repo)
	s.now = func() time.Time { return now }

	if err := s.Flush(); err != nil {
		t.Fatal(err)
	}

	want := now.Add(-reliabilityWindow)
	if !repo.prunedBefore.Equal(want) {
		t.Fatalf("expected prune cutoff %v, got %v", want, repo.prunedBefore)
	}
}

func TestReliabilityWeightsTemperatureAndCheckEngine(t *testing.T) {
	s := NewService(&stubRepo{})

	overheatEpisode(s, "CAR001") // 1 overheat -> temperature score 90
	for i := 0; i < 3; i++ {     // 3 check-engine -> check score 85
		s.Observe("CAR001", 70, true)
		s.Observe("CAR001", 70, false)
	}
	res := s.Observe("CAR001", 70, false)

	// 0.6*90 + 0.4*85 = 88
	if math.Abs(res.Reliability-88) > 1e-9 {
		t.Fatalf("expected reliability 88, got %v", res.Reliability)
	}
}

func TestReliabilityIsFullForHealthyCar(t *testing.T) {
	s := NewService(&stubRepo{})

	res := s.Observe("CAR001", 70, false)

	if res.Reliability != 100 {
		t.Fatalf("expected reliability 100, got %v", res.Reliability)
	}
}

func TestReliabilityIsZeroWhenBothScoresFloored(t *testing.T) {
	s := NewService(&stubRepo{})

	for i := 0; i < 12; i++ {
		overheatEpisode(s, "CAR001")
	}
	for i := 0; i < 21; i++ {
		s.Observe("CAR001", 70, true)
		s.Observe("CAR001", 70, false)
	}
	res := s.Observe("CAR001", 70, false)

	if res.Reliability != 0 {
		t.Fatalf("expected reliability 0, got %v", res.Reliability)
	}
}

func TestFlushPersistsSnapshot(t *testing.T) {
	repo := &stubRepo{}
	s := NewService(repo)

	s.Observe("CAR001", 112, false)
	s.Observe("CAR001", 100, false)

	if err := s.Flush(); err != nil {
		t.Fatal(err)
	}
	if len(repo.upserted) != 1 {
		t.Fatalf("expected 1 stat upserted, got %d", len(repo.upserted))
	}
	got := repo.upserted[0]
	if got.CarID != "CAR001" {
		t.Errorf("wrong car: %v", got.CarID)
	}
	if got.OverheatEvents != 1 {
		t.Errorf("expected 1 event, got %d", got.OverheatEvents)
	}
	if got.TotalReadings != 2 {
		t.Errorf("expected 2 readings, got %d", got.TotalReadings)
	}
	if got.MaxEngineTemp != 112 {
		t.Errorf("expected max temp 112, got %v", got.MaxEngineTemp)
	}
}
