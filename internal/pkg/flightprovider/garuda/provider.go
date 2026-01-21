package garuda

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"time"

	"github.com/go-redis/redis_rate/v10"
	"github.com/ijalalfrz/flight-search-aggregation-service/internal/app/dto"
	"github.com/ijalalfrz/flight-search-aggregation-service/internal/pkg/flightprovider"
	"github.com/ijalalfrz/flight-search-aggregation-service/internal/pkg/flightprovider/providerutils"
	"github.com/ijalalfrz/flight-search-aggregation-service/internal/pkg/utils"
)

const (
	ProviderName = "Garuda"
	ProviderCode = "GA"
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

// Search will simulate API call to Garuda flight search API
// it will return always success with delay 50-100ms
func (p *Provider) Search(ctx context.Context, criteria dto.SearchCriteria) ([]dto.Flight, error) {
	ctx, cancel := context.WithTimeout(ctx, p.Timeout)
	defer cancel()

	var lastErr error
	for attempt := 0; attempt <= p.MaxRetries; attempt++ {
		// simulate delay 100-300ms
		delay := time.Duration(100+rand.Intn(201)) * time.Millisecond
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled or timeout: %w", ctx.Err())
		}

		// always success not simulate error
		if rand.Float64() < 0 {
			lastErr = providerutils.ErrProviderInternalError
			slog.ErrorContext(ctx, "failed to call garuda flight search API", "attempt", attempt+1, "error", lastErr)

			if attempt < p.MaxRetries {
				// Exponential backoff: 200ms * 2^attempt
				backoff := time.Duration(200*(1<<attempt)) * time.Millisecond
				slog.InfoContext(ctx, "retrying with exponential backoff", "backoff", backoff, "next_attempt", attempt+2)
				select {
				case <-time.After(backoff):
					continue
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			}
			continue
		}

		// rate limit
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

	return nil, fmt.Errorf("failed to get flight data after %d attempts", p.MaxRetries)
}

func (p *Provider) flightToDTO(flights []Flight) []dto.Flight {
	results := make([]dto.Flight, len(flights))
	for i, flight := range flights {
		results[i] = dto.Flight{
			ID:       p.generateID(flight.FlightID, p.Name),
			Provider: p.Name,
			Airline: dto.Airline{
				Name: ProviderName,
				Code: ProviderCode,
			},
			FlightNumber: flight.FlightID,
			Departure: dto.Departure{
				Airport:   flight.Departure.Airport,
				City:      flight.Departure.City,
				Datetime:  flight.Departure.Time.Format(time.RFC3339),
				Timestamp: flight.Departure.Time.Unix(),
			},
			Arrival: dto.Arrival{
				Airport:   flight.Arrival.Airport,
				City:      flight.Arrival.City,
				Datetime:  flight.Arrival.Time.Format(time.RFC3339),
				Timestamp: flight.Arrival.Time.Unix(),
			},
			Duration: dto.Duration{
				TotalMinutes: flight.DurationMinutes,
				Formatted:    utils.ConvertMinutesToDuration(int64(flight.DurationMinutes)),
			},
			Stops: flight.Stops,
			Price: dto.Price{
				Amount:    float64(flight.Price.Amount),
				Currency:  flight.Price.Currency,
				Formatted: utils.FormatRupiah(int64(flight.Price.Amount)),
			},
			AvailableSeats: flight.AvailableSeats,
			CabinClass:     flight.FareClass,
			Aircraft:       &flight.Aircraft,
			Amenities:      flight.Amenities,
			Baggage:        p.parseBaggage(flight.Baggage),
		}
	}
	return results
}

func (p *Provider) generateID(code, name string) string {
	return fmt.Sprintf("%s_%s", code, name)
}

func (p *Provider) parseBaggage(b Baggage) dto.Baggage {
	return dto.Baggage{
		CarryOn: fmt.Sprintf("%dkg", b.CarryOn*7),  // assuming 1 carry on = 7kg
		Checked: fmt.Sprintf("%dkg", b.Checked*15), // assuming 1 checked bag = 15kg
	}
}
