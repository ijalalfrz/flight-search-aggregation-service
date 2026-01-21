package airasia

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/go-redis/redis_rate/v10"
	"github.com/ijalalfrz/flight-search-aggregation-service/internal/app/dto"
	"github.com/ijalalfrz/flight-search-aggregation-service/internal/pkg/flightprovider"
	"github.com/ijalalfrz/flight-search-aggregation-service/internal/pkg/flightprovider/providerutils"
	"github.com/ijalalfrz/flight-search-aggregation-service/internal/pkg/utils"
)

const (
	ProviderName = "AirAsia"
	ProviderCode = "QZ"
)

type Provider struct {
	Name         string
	SearchAPIURL string
	Timeout      time.Duration
	MaxRetries   int
	Limiter      *redis_rate.Limiter
	RateLimitRPS int
}

func NewProvider(config flightprovider.FlightProviderConfig) *Provider {
	return &Provider{
		Name:         ProviderName,
		SearchAPIURL: config.SearchAPIURL,
		Timeout:      config.Timeout,
		MaxRetries:   config.MaxRetries,
		Limiter:      config.Limiter,
		RateLimitRPS: config.RateLimitRPS,
	}
}

// Search will simulate API call to AirAsia flight search API
// it will return 90% success with delay 50-150ms
// and will retry if failed
func (p *Provider) Search(ctx context.Context, criteria dto.SearchCriteria) ([]dto.Flight, error) {
	ctx, cancel := context.WithTimeout(ctx, p.Timeout)
	defer cancel()

	var lastErr error
	for attempt := 0; attempt <= p.MaxRetries; attempt++ {
		// simulate delay 50-150ms
		delay := time.Duration(50+rand.Intn(101)) * time.Millisecond
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled or timeout: %w", ctx.Err())
		}

		// simulate call 90% success, 10% failure
		if rand.Float64() < 0.1 {
			lastErr = providerutils.ErrProviderInternalError
			slog.ErrorContext(ctx, "failed to call airasia flight search API", "attempt",
				attempt+1, "error", lastErr)

			if attempt < p.MaxRetries {
				// Exponential backoff: 200ms * 2^attempt
				backoff := time.Duration(200*(1<<attempt)) * time.Millisecond
				slog.InfoContext(ctx, "retrying with exponential backoff", "backoff",
					backoff, "next_attempt", attempt+2)
				select {
				case <-time.After(backoff):
					continue
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			}
			continue
		}

		// distributed rate limit using leaky bucket
		res, err := p.Limiter.Allow(ctx, fmt.Sprintf("limit:%s", p.Name),
			redis_rate.PerSecond(p.RateLimitRPS))
		if err != nil {
			return nil, fmt.Errorf("failed to rate limit: %w", err)
		}

		if res.Allowed == 0 {
			return nil, providerutils.ErrProviderRateLimitExceeded
		}

		// get response from mock file
		flightData, err := os.ReadFile(p.SearchAPIURL)
		if err != nil {
			return nil, fmt.Errorf("failed to read mock file: %w", err)
		}

		var response SearchFlightResponse
		if err := json.Unmarshal(flightData, &response); err != nil {
			return nil, fmt.Errorf("failed to unmarshal mock file: %w", err)
		}

		// convert to dto.Flight and filter
		flights := p.flightToDTO(response.Flights)

		return providerutils.FilterFlights(flights, criteria), nil
	}

	return nil, fmt.Errorf("failed after retries: %w", providerutils.ErrRetryExceeded)
}

// flightToDTO converts a slice of Flight to a slice of dto.Flight
// it will normalize the data from the provider to the dto.Flight struct
func (p *Provider) flightToDTO(flights []Flight) []dto.Flight {
	results := make([]dto.Flight, len(flights))
	for i, flight := range flights {
		results[i] = dto.Flight{
			ID:       p.generateID(flight.FlightCode, p.Name),
			Provider: p.Name,
			Airline: dto.Airline{
				Name: ProviderName,
				Code: ProviderCode,
			},
			FlightNumber: flight.FlightCode,
			Departure: dto.Departure{
				Airport:   flight.FromAirport,
				Datetime:  flight.DepartTime.Format(time.RFC3339),
				Timestamp: flight.DepartTime.Unix(),
			},
			Arrival: dto.Arrival{
				Airport:   flight.ToAirport,
				Datetime:  flight.ArriveTime.Format(time.RFC3339),
				Timestamp: flight.ArriveTime.Unix(),
			},
			Duration: dto.Duration{
				TotalMinutes: int(flight.DurationHours * 60),
				Formatted:    utils.ConvertHourToDuration(flight.DurationHours),
			},
			Stops: p.getStops(flight),
			Price: dto.Price{
				Amount:    float64(flight.PriceIDR),
				Currency:  "IDR",
				Formatted: utils.FormatRupiah(int64(flight.PriceIDR)),
			},
			AvailableSeats: flight.Seats,
			CabinClass:     flight.CabinClass,
			Aircraft:       nil,
			Amenities:      []string{},
			Baggage:        p.parseBaggage(flight.BaggageNote),
		}
	}

	return results
}

func (p *Provider) getStops(flight Flight) int {
	if flight.DirectFlight {
		return 0
	}

	return len(flight.Stops)
}

func (p *Provider) generateID(code, name string) string {
	return fmt.Sprintf("%s_%s", code, name)
}

func (p *Provider) parseBaggage(note string) dto.Baggage {
	baggage := dto.Baggage{}
	if strings.Contains(note, "Cabin baggage only") {
		baggage.CarryOn = "Cabin baggage only"
	}

	if strings.Contains(note, "checked bags additional fee") {
		baggage.Checked = "Additional fee"
	}

	return baggage
}
