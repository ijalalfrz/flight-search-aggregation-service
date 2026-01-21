//go:build unit

package flight

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/ijalalfrz/flight-search-aggregation-service/internal/app/dto"
)

func TestSortFlights_Closure(t *testing.T) {
	flights := []dto.Flight{
		{ID: "1", Price: dto.Price{Amount: 2000}, Score: 0.8},
		{ID: "2", Price: dto.Price{Amount: 1000}, Score: 0.1},
		{ID: "3", Price: dto.Price{Amount: 1500}, Score: 0.5},
	}

	sortRequest := func(flights []dto.Flight, opt *dto.SortOption, wantIDs []string) func(t *testing.T) {
		return func(t *testing.T) {
			// Copy to avoid shared state
			fCopy := make([]dto.Flight, len(flights))
			copy(fCopy, flights)

			got := SortFlights(fCopy, opt)
			gotIDs := make([]string, len(got))
			for i, f := range got {
				gotIDs[i] = f.ID
			}

			diff := cmp.Diff(wantIDs, gotIDs)
			if diff != "" {
				t.Fatalf("SortFlights result mismatch (-want +got):\n%s", diff)
			}
		}
	}

	t.Run("default_sort_best_score_asc", sortRequest(flights, nil, []string{"2", "3", "1"}))
	t.Run("price_asc", sortRequest(flights, &dto.SortOption{Field: "price", Order: "asc"}, []string{"2", "3", "1"}))
	t.Run("price_desc", sortRequest(flights, &dto.SortOption{Field: "price", Order: "desc"}, []string{"1", "3", "2"}))
}
