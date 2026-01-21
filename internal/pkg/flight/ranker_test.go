package flight

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/ijalalfrz/flight-search-aggregation-service/internal/app/dto"
)

func TestRankFlights_Closure(t *testing.T) {
	flights := []dto.Flight{
		{
			ID:        "1",
			Price:     dto.Price{Amount: 1000},
			Duration:  dto.Duration{TotalMinutes: 100},
			Stops:     0,
			Amenities: []string{"WiFi"},
		},
		{
			ID:        "2",
			Price:     dto.Price{Amount: 5000},
			Duration:  dto.Duration{TotalMinutes: 500},
			Stops:     2,
			Amenities: []string{},
		},
	}

	rankRequest := func(flights []dto.Flight, wantBestID string) func(t *testing.T) {
		return func(t *testing.T) {
			// Copy to avoid shared state
			fCopy := make([]dto.Flight, len(flights))
			copy(fCopy, flights)

			got := RankFlights(fCopy)

			// Simple check: best flight should have lowest score
			bestScore := 999.0
			var gotBestID string
			for _, f := range got {
				if f.Score < bestScore {
					bestScore = f.Score
					gotBestID = f.ID
				}
			}

			if gotBestID != wantBestID {
				t.Fatalf("RankFlights failed: expected best flight ID %s, got %s", wantBestID, gotBestID)
			}
		}
	}

	t.Run("basic_ranking", rankRequest(flights, "1"))
}

func TestNormalizeValue_Closure(t *testing.T) {
	normalizeRequest := func(val, min, max, want float64) func(t *testing.T) {
		return func(t *testing.T) {
			got := normalizeValue(val, min, max)
			diff := cmp.Diff(want, got)
			if diff != "" {
				t.Fatalf("normalizeValue mismatch (-want +got):\n%s", diff)
			}
		}
	}

	t.Run("mid_value", normalizeRequest(15, 10, 20, 0.5))
	t.Run("min_value", normalizeRequest(10, 10, 20, 0.0))
	t.Run("max_value", normalizeRequest(20, 10, 20, 1.0))
	t.Run("division_by_zero_safety", normalizeRequest(10, 10, 10, 0.0))
}
