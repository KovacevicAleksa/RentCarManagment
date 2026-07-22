package carstats

import (
	"math"
	"testing"
)

type stubRepo struct {
	loaded    []Stats
	upserted  []Stats
	loadErr   error
	upsertErr error
}

func (s *stubRepo) LoadAll() ([]Stats, error) { return s.loaded, s.loadErr }
func (s *stubRepo) UpsertBatch(items []Stats) error {
	if s.upsertErr != nil {
		return s.upsertErr
	}
	s.upserted = items
	return nil
}

// feed replays n readings at the given temperature and returns the last
// reported frequency.
func feed(s *Service, carID string, temp float64, n int) float64 {
	var freq float64
	for i := 0; i < n; i++ {
		freq = s.Observe(carID, temp)
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
		s.Observe("CAR001", 112)
		s.Observe("CAR001", 100)
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
	s.Observe("CAR001", 112)
	s.Observe("CAR001", 115)
	s.Observe("CAR001", 111) // still above low, not re-armed
	freq := feed(s, "CAR001", 70, readingsPerHour-3)

	if math.Abs(freq-1.0) > 1e-9 {
		t.Fatalf("expected 1.0 event/hour for a single episode, got %v", freq)
	}
}

func TestFrequencyIsZeroDuringWarmup(t *testing.T) {
	s := NewService(&stubRepo{})

	// One event but only a handful of readings (< 10 min of operation).
	freq := s.Observe("CAR001", 112)

	if freq != 0 {
		t.Fatalf("expected 0 during warmup, got %v", freq)
	}
}

func TestStatsIndependentPerCar(t *testing.T) {
	s := NewService(&stubRepo{})

	s.Observe("CAR001", 112) // CAR001: one episode
	s.Observe("CAR001", 100)
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
	freq := s.Observe("CAR001", 70)

	want := 3.0 / (float64(readingsPerHour+1) * 5 / 3600)
	if math.Abs(freq-want) > 1e-9 {
		t.Fatalf("expected %v after load, got %v", want, freq)
	}
}

func TestFlushPersistsSnapshot(t *testing.T) {
	repo := &stubRepo{}
	s := NewService(repo)

	s.Observe("CAR001", 112)
	s.Observe("CAR001", 100)

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
