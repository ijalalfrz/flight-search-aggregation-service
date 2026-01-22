package dto

import (
	"fmt"
	"net/http"

	"github.com/ijalalfrz/flight-search-aggregation-service/internal/pkg/exception"
)

type Flight struct {
	ID             string    `json:"id"`
	Provider       string    `json:"provider"`
	Airline        Airline   `json:"airline"`
	FlightNumber   string    `json:"flight_number"`
	Departure      Departure `json:"departure"`
	Arrival        Arrival   `json:"arrival"`
	Duration       Duration  `json:"duration"`
	Stops          int       `json:"stops"`
	Price          Price     `json:"price"`
	AvailableSeats int       `json:"available_seats"`
	CabinClass     string    `json:"cabin_class"`
	Aircraft       *string   `json:"aircraft"`
	Amenities      []string  `json:"amenities"`
	Baggage        Baggage   `json:"baggage"`
	Score          float64   `json:"score"`
}

type Airline struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type Departure struct {
	Airport   string `json:"airport"`
	City      string `json:"city"`
	Datetime  string `json:"datetime"`
	Timestamp int64  `json:"timestamp"`
}

type Arrival struct {
	Airport   string `json:"airport"`
	City      string `json:"city"`
	Datetime  string `json:"datetime"`
	Timestamp int64  `json:"timestamp"`
}

type Duration struct {
	TotalMinutes int    `json:"total_minutes"`
	Formatted    string `json:"formatted"`
}

type Price struct {
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	Formatted string  `json:"formatted"`
}

type Baggage struct {
	CarryOn string `json:"carry_on"`
	Checked string `json:"checked"`
}

type SearchCriteria struct {
	Origin        string        `json:"origin" validate:"required"`
	Destination   string        `json:"destination" validate:"required"`
	DepartureDate string        `json:"departure_date" validate:"required"`
	Passengers    int           `json:"passengers" validate:"required,min=1,max=10"`
	CabinClass    string        `json:"cabin_class" validate:"required,oneof=economy business first"`
	SortOption    *SortOption   `json:"sort_option,omitempty"`
	FilterOption  *FilterOption `json:"filter_option,omitempty"`
}

func (s *SearchCriteria) Bind(r *http.Request) error {
	if err := s.Validate(); err != nil {
		return fmt.Errorf("error validate request: %w", err)
	}

	return nil
}

func (s *SearchCriteria) Validate() error {
	if err := ValidateSingleError(s); err != nil {
		return exception.ApplicationError{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
		}
	}

	if s.SortOption != nil {
		if !AllowedSortField[s.SortOption.Field] {
			return exception.ApplicationError{
				StatusCode: http.StatusBadRequest,
				Message:    fmt.Sprintf("Invalid sort field %s", s.SortOption.Field),
			}
		}
	}

	if s.FilterOption != nil {
		if s.FilterOption.MinPrice != nil && s.FilterOption.MaxPrice != nil &&
			*s.FilterOption.MaxPrice <= *s.FilterOption.MinPrice {
			return exception.ApplicationError{
				StatusCode: http.StatusBadRequest,
				Message:    "max_price must be greater than min_price",
			}
		}

		if s.FilterOption.MinStops != nil && s.FilterOption.MaxStops != nil &&
			*s.FilterOption.MaxStops <= *s.FilterOption.MinStops {
			return exception.ApplicationError{
				StatusCode: http.StatusBadRequest,
				Message:    "max_stops must be greater than min_stops",
			}
		}

		if s.FilterOption.MinDurationMinutes != nil && s.FilterOption.MaxDurationMinutes != nil &&
			*s.FilterOption.MaxDurationMinutes <= *s.FilterOption.MinDurationMinutes {
			return exception.ApplicationError{
				StatusCode: http.StatusBadRequest,
				Message:    "max_duration_minutes must be greater than min_duration_minutes",
			}
		}
	}

	return nil
}

type FilterOption struct {
	MinPrice           *float64 `json:"min_price,omitempty" validate:"omitempty,numeric,gt=0"`
	MaxPrice           *float64 `json:"max_price,omitempty" validate:"omitempty,numeric,gt=0"`
	MinStops           *int     `json:"min_stops,omitempty" validate:"omitempty,numeric,gte=0"`
	MaxStops           *int     `json:"max_stops,omitempty" validate:"omitempty,numeric,gte=0"`
	Airline            *string  `json:"airline,omitempty"`
	DepartureTimeStart *string  `json:"departure_time_start,omitempty"`
	DepartureTimeEnd   *string  `json:"departure_time_end,omitempty"`
	ArrivalTimeStart   *string  `json:"arrival_time_start,omitempty"`
	ArrivalTimeEnd     *string  `json:"arrival_time_end,omitempty"`
	MinDurationMinutes *int     `json:"min_duration_minutes,omitempty" validate:"omitempty,numeric,gte=0"`
	MaxDurationMinutes *int     `json:"max_duration_minutes,omitempty" validate:"omitempty,numeric,gte=0"`
}

type SortOption struct {
	Field string `json:"field"`
	Order string `json:"order"`
}

type Metadata struct {
	TotalResults       int  `json:"total_results"`
	ProvidersQueried   int  `json:"providers_queried"`
	ProvidersSucceeded int  `json:"providers_succeeded"`
	ProvidersFailed    int  `json:"providers_failed"`
	SearchTimeMs       int  `json:"search_time_ms"`
	CacheHit           bool `json:"cache_hit"`
}

// SearchFlightResponse is the response struct for the search flight endpoint
type SearchFlightResponse struct {
	SearchCriteria SearchCriteria `json:"search_criteria"`
	Metadata       Metadata       `json:"metadata"`
	Flights        []Flight       `json:"flights"`
}
