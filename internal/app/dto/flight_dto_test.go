//go:build unit

package dto

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestSearchCriteria_Validate(t *testing.T) {
	// Initialize validator for tests
	_ = InitValidator()

	validateRequest := func(req SearchCriteria, wantErr bool, wantMsg string) func(t *testing.T) {
		return func(t *testing.T) {
			err := req.Validate()
			if (err != nil) != wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, wantErr)
			}

			if wantErr && err != nil {
				if diff := cmp.Diff(wantMsg, err.Error()); diff != "" {
					t.Fatalf("Validate() error message mismatch (-want +got):\n%s", diff)
				}
			}
		}
	}

	// Helper for pointers
	ptrFloat := func(f float64) *float64 { return &f }
	ptrInt := func(i int) *int { return &i }

	validCriteria := SearchCriteria{
		Origin:        "JKT",
		Destination:   "DPS",
		DepartureDate: "2024-01-01",
		Passengers:    1,
		CabinClass:    "ECONOMY",
	}

	t.Run("valid_criteria", validateRequest(validCriteria, false, ""))

	t.Run("missing_origin", validateRequest(SearchCriteria{
		Destination:   "DPS",
		DepartureDate: "2024-01-01",
		Passengers:    1,
		CabinClass:    "ECONOMY",
	}, true, "origin is a required field"))

	t.Run("invalid_sort_field", validateRequest(SearchCriteria{
		Origin:        "JKT",
		Destination:   "DPS",
		DepartureDate: "2024-01-01",
		Passengers:    1,
		CabinClass:    "ECONOMY",
		SortOption:    &SortOption{Field: "invalid", Order: "asc"},
	}, true, "Invalid sort field invalid"))

	t.Run("invalid_price_range", validateRequest(SearchCriteria{
		Origin:        "JKT",
		Destination:   "DPS",
		DepartureDate: "2024-01-01",
		Passengers:    1,
		CabinClass:    "ECONOMY",
		FilterOption: &FilterOption{
			MinPrice: ptrFloat(1000),
			MaxPrice: ptrFloat(500),
		},
	}, true, "max_price must be greater than min_price"))

	t.Run("invalid_stops_range", validateRequest(SearchCriteria{
		Origin:        "JKT",
		Destination:   "DPS",
		DepartureDate: "2024-01-01",
		Passengers:    1,
		CabinClass:    "ECONOMY",
		FilterOption: &FilterOption{
			MinStops: ptrInt(2),
			MaxStops: ptrInt(1),
		},
	}, true, "max_stops must be greater than min_stops"))

	t.Run("invalid_duration_range", validateRequest(SearchCriteria{
		Origin:        "JKT",
		Destination:   "DPS",
		DepartureDate: "2024-01-01",
		Passengers:    1,
		CabinClass:    "ECONOMY",
		FilterOption: &FilterOption{
			MinDurationMinutes: ptrInt(200),
			MaxDurationMinutes: ptrInt(100),
		},
	}, true, "max_duration_minutes must be greater than min_duration_minutes"))
}

func TestSearchCriteria_Bind(t *testing.T) {
	_ = InitValidator()

	bindRequest := func(req SearchCriteria, wantErr bool) func(t *testing.T) {
		return func(t *testing.T) {
			err := req.Bind(nil)
			if (err != nil) != wantErr {
				t.Fatalf("Bind() error = %v, wantErr %v", err, wantErr)
			}
		}
	}

	validCriteria := SearchCriteria{
		Origin:        "JKT",
		Destination:   "DPS",
		DepartureDate: "2024-01-01",
		Passengers:    1,
		CabinClass:    "ECONOMY",
	}

	t.Run("valid_bind", bindRequest(validCriteria, false))
	t.Run("invalid_bind", bindRequest(SearchCriteria{}, true))
}
