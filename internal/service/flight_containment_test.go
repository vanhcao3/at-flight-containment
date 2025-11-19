package service

import (
	"math"
	"testing"

	config "172.21.5.249/air-trans/at-drone/internal/config"
)

func TestHorizontalDeviationDecreasesWhenReturningToPath(t *testing.T) {
	ms := &MainService{
		SvcConfig: &config.ServiceConfig{
			FlightContainment: config.FlightContainmentConfig{
				Waypoints: []config.FlightContainmentWaypointConfig{
					{Latitude: 21.0, Longitude: 105.0, Altitude: 50},
					{Latitude: 21.0, Longitude: 105.1, Altitude: 50},
				},
				HorizontalDeviationM: 1000,
				AltDeviationM:        1000,
			},
		},
	}

	offsets := []float64{0.0010, 0.0008, 0.0006, 0.0004, 0.0002, 0}
	prev := math.MaxFloat64
	for _, off := range offsets {
		eval, ok := ms.evaluateFlightContainment(21.0+off, 105.05, 50)
		if !ok {
			t.Fatalf("evaluation failed for offset %v", off)
		}
		deviation := math.Abs(eval.horizontalDeviation)
		if deviation > prev+0.5 { // allow small numerical noise
			t.Fatalf("deviation increased while offset shrank: offset=%v deviation=%v prev=%v", off, deviation, prev)
		}
		prev = deviation
	}
}
