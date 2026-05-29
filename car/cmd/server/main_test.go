package main

import "testing"

func TestFuelNeverGoesNegative(t *testing.T) {
	sim := NewCarSimulator("CAR001", 44.7866, 20.4489)
	sim.fuelLevel = 0.01

	for i := 0; i < 2000; i++ {
		tel := sim.GetNextTelemetry()
		if tel.FuelLevel < 0 {
			t.Fatalf("fuel level went negative on iteration %d: %f", i, tel.FuelLevel)
		}
	}
}

func TestThrottleStaysWithinBounds(t *testing.T) {
	sim := NewCarSimulator("CAR001", 44.7866, 20.4489)

	for i := 0; i < 2000; i++ {
		tel := sim.GetNextTelemetry()
		if tel.GasThrottle < 0 || tel.GasThrottle > 1 {
			t.Fatalf("throttle out of bounds on iteration %d: %f", i, tel.GasThrottle)
		}
	}
}

func TestTelemetryCarriesCarID(t *testing.T) {
	sim := NewCarSimulator("CAR042", 44.7866, 20.4489)
	if got := sim.GetNextTelemetry().CarID; got != "CAR042" {
		t.Errorf("CarID = %q, want %q", got, "CAR042")
	}
}
