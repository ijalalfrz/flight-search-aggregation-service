package flight

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/ijalalfrz/flight-search-aggregation-service/internal/app/dto"
	"github.com/stretchr/testify/assert"
)

func TestFilterFlights(t *testing.T) {
	airlineGaruda := "Garuda"
	maxPrice := 1000000.0

	flights := []dto.Flight{
		{
			ID:      "1",
			Airline: dto.Airline{Name: "Garuda Indonesia"},
			Price:   dto.Price{Amount: 800000},
			Stops:   0,
		},
		{
			ID:      "2",
			Airline: dto.Airline{Name: "Lion Air"},
			Price:   dto.Price{Amount: 1200000},
			Stops:   1,
		},
	}

	filterRequest := func(flights []dto.Flight, opts *dto.FilterOption, wantIDs []string) func(t *testing.T) {
		return func(t *testing.T) {
			got := FilterFlights(context.Background(), flights, opts)
			gotIDs := make([]string, len(got))
			for i, f := range got {
				gotIDs[i] = f.ID
			}

			diff := cmp.Diff(wantIDs, gotIDs)
			if diff != "" {
				t.Fatalf("FilterFlights result mismatch (-want +got):\n%s", diff)
			}
		}
	}

	t.Run("nil_filter", filterRequest(flights, nil, []string{"1", "2"}))
	t.Run("filter_by_airline", filterRequest(flights, &dto.FilterOption{Airline: &airlineGaruda}, []string{"1"}))
	t.Run("filter_by_max_price", filterRequest(flights, &dto.FilterOption{MaxPrice: &maxPrice}, []string{"1"}))
	t.Run("no_match", filterRequest(flights, &dto.FilterOption{MaxPrice: func() *float64 { f := 100.0; return &f }()}, []string{}))
}

func TestIsWithinTimeRange_Closure(t *testing.T) {
	timeRangeRequest := func(target, start, end string, want bool) func(t *testing.T) {
		return func(t *testing.T) {
			got := isWithinTimeRange(context.Background(), target, start, end)
			assert.Equal(t, want, got)
		}
	}

	t.Run("within_range", timeRangeRequest("2024-01-01T14:30:00+07:00", "12:00", "16:00", true))
	t.Run("outside_range", timeRangeRequest("2024-01-01T10:00:00+07:00", "12:00", "16:00", false))
	t.Run("invalid_format", timeRangeRequest("invalid", "12:00", "16:00", false))
}
